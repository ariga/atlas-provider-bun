package bunschema

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"go/ast"
	"go/token"
	"maps"
	"reflect"
	"slices"
	"strings"

	"ariga.io/atlas/sdk/recordriver"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mssqldialect"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/oracledialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/schema"
	"golang.org/x/tools/go/packages"
)

// Dialect is the type for supported dialects.
type Dialect string

const (
	DialectMSSQL    Dialect = "mssql"
	DialectMySQL    Dialect = "mysql"
	DialectOracle   Dialect = "oracle"
	DialectPostgres Dialect = "postgres"
	DialectSQLite   Dialect = "sqlite"
)

// Option is a function that configures the Loader.
type Option func(*Loader)

// Loader is the struct that holds the loader configuration.
type Loader struct {
	dialect    Dialect
	delimiter  string
	joinTables []any
	buildTags  string
}

// New creates a new Loader.
func New(d Dialect, opts ...Option) *Loader {
	l := &Loader{
		dialect:   d,
		delimiter: ";",
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// WithStmtDelimiter sets the delimiter for the output.
// The default delimiter is `;`.
// This is helpful for SQL Server, which uses the GO keyword as a delimiter.
func WithStmtDelimiter(delimiter string) Option {
	return func(l *Loader) {
		l.delimiter = delimiter
	}
}

// WithJoinTable registers the given models as join tables.
func WithJoinTable(models ...any) Option {
	return func(l *Loader) {
		l.joinTables = append(l.joinTables, models...)
	}
}

// WithBuildTags sets the build tags to use when loading packages.
func WithBuildTags(tags string) Option {
	return func(l *Loader) {
		l.buildTags = tags
	}
}

// Load loads the models and returns DDL statements.
func (l *Loader) Load(models ...any) (string, error) {
	for _, m := range models {
		if err := validate(m); err != nil {
			return "", fmt.Errorf("invalid model %T: %w", m, err)
		}
	}
	rc, err := sql.Open("recordriver", "bun")
	if err != nil {
		return "", fmt.Errorf("failed to open database: %w", err)
	}
	defer rc.Close()
	var di schema.Dialect
	switch l.dialect {
	case DialectMySQL:
		di = mysqldialect.New()
		recordriver.SetResponse("bun", "SELECT version()", &recordriver.Response{
			Cols: []string{"version()"},
			Data: [][]driver.Value{{"8.0.24"}},
		})

	case DialectSQLite:
		di = sqlitedialect.New()
		recordriver.SetResponse("bun", "select sqlite_version()", &recordriver.Response{
			Cols: []string{"sqlite_version()"},
			Data: [][]driver.Value{{"3.30.1"}},
		})

	case DialectMSSQL:
		di = mssqldialect.New()
		recordriver.SetResponse("bun", "SELECT @@VERSION", &recordriver.Response{
			Cols: []string{"SELECT @@VERSION"},
			Data: [][]driver.Value{{"15.0.2000.58"}},
		})
	case DialectOracle:
		di = oracledialect.New()
	case DialectPostgres:
		di = pgdialect.New()
	default:
		return "", fmt.Errorf("unsupported dialect: %s", l.dialect)
	}
	db := bun.NewDB(rc, di)
	for _, m := range l.joinTables {
		// Bun requires join tables to be registered before use
		db.RegisterModel(m)
	}
	db.RegisterModel(models...)
	// Sort tables topologically based on their dependencies
	tables, err := topologicalSort(db.Dialect().Tables().All())
	if err != nil {
		return "", fmt.Errorf("failed to sort tables: %w", err)
	}
	// Create tables in dependency order
	for _, t := range tables {
		if _, err := db.NewCreateTable().
			Model(t.ZeroIface).
			WithForeignKeys().
			Exec(context.Background()); err != nil {
			return "", err
		}
	}
	ss, ok := recordriver.Session("bun")
	if !ok {
		return "", fmt.Errorf("failed to read session")
	}
	pos, err := l.tablePos(tables)
	if err != nil {
		return "", fmt.Errorf("failed to get tables position: %w", err)
	}
	var buf strings.Builder
	for _, t := range tables {
		if p, ok := pos(t.ZeroIface); ok {
			if _, err = fmt.Fprintf(&buf, "-- atlas:pos %s[type=table] %s\n", t.Name, p); err != nil {
				return "", err
			}
		}
	}
	// Add another new line to separate the file directives from the statements.
	if _, err = fmt.Fprintln(&buf); err != nil {
		return "", err
	}
	for _, stmt := range ss.Statements {
		if _, err = fmt.Fprintln(&buf, stmt+l.delimiter); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

func (l *Loader) tablePos(tables []*schema.Table) (func(any) (string, bool), error) {
	var (
		pkgTables = make(map[string][]*schema.Table)
		pos       = make(map[string]string)
	)
	for _, t := range tables {
		typ := t.Type
		pkgPath := typ.PkgPath()
		if pkgPath == "" {
			return nil, fmt.Errorf("could not determine package path for struct '%s'", typ.Name())
		}
		pkgTables[pkgPath] = append(pkgTables[pkgPath], t)
	}
	for p, t := range pkgTables {
		pkgPos, err := l.packagePos(p, t)
		if err != nil {
			return nil, err
		}
		maps.Copy(pos, pkgPos)
	}
	return func(m any) (string, bool) {
		val, ok := pos[modelID(m)]
		return val, ok
	}, nil
}

// packagePos finds positions for all structs within a single package.
// It returns a map in the format pkg.struct=location
func (l *Loader) packagePos(pkgPath string, tables []*schema.Table) (map[string]string, error) {
	var (
		tableModel = make(map[string]any, len(tables))
		pos        = make(map[string]string, len(tables))
		found      = make(map[string]bool, len(tables))
		fset       = token.NewFileSet()
	)
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax,
		Fset: fset,
	}
	if l.buildTags != "" {
		cfg.BuildFlags = []string{"-tags=" + l.buildTags}
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	switch {
	case err != nil:
		return nil, fmt.Errorf("failed to load package: %w", err)
	case packages.PrintErrors(pkgs) > 0:
		return nil, fmt.Errorf("errors while loading package %s", pkgPath)
	case len(pkgs) == 0:
		return nil, fmt.Errorf("package %s not found", pkgPath)
	}
	for _, t := range tables {
		tableModel[t.TypeName] = t.ZeroIface
	}
	for _, pkg := range pkgs {
		for i, fnode := range pkg.Syntax {
			ast.Inspect(fnode, func(n ast.Node) bool {
				ts, ok := n.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					return true
				}
				name := ts.Name.Name
				if m, exists := tableModel[name]; exists && !found[name] {
					location := fmt.Sprintf("%s:%d-%d",
						pkg.GoFiles[i],
						fset.Position(ts.Pos()).Line,
						fset.Position(ts.Type.End()).Line,
					)
					pos[modelID(m)] = location
					found[name] = true
				}
				return true
			})
			if len(found) == len(tables) { // all tables found
				break
			}
		}
	}
	for _, t := range tables {
		if !found[t.TypeName] {
			return nil, fmt.Errorf("struct '%s' not found in package '%s'", t.TypeName, pkgPath)
		}
	}
	return pos, nil
}

// modelID returns a unique identifier for a model.
func modelID(m any) string {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

// validate ensures that the provided model is a pointer to a struct.
// This is required by Bun; otherwise, it will panic at runtime.
func validate(v any) error {
	t := reflect.TypeOf(v)
	switch {
	case t == nil:
		return fmt.Errorf("model is nil")
	case t.Kind() != reflect.Ptr:
		return fmt.Errorf("model must be a pointer, got %s", t.Kind())
	case t.Elem().Kind() != reflect.Struct:
		return fmt.Errorf("model must point to a struct, got pointer to %s", t.Elem().Kind())
	}
	return nil
}

// topologicalSort returns tables in dependency order (dependencies first).
func topologicalSort(tables []*schema.Table) ([]*schema.Table, error) {
	// Sort input tables by name for deterministic ordering
	slices.SortFunc(tables, func(a, b *schema.Table) int {
		return strings.Compare(a.Name, b.Name)
	})

	tableMap := make(map[string]*schema.Table, len(tables))
	for _, t := range tables {
		tableMap[t.Name] = t
	}
	visited := make(map[string]bool, len(tables))
	visiting := make(map[string]bool, len(tables)) // tracks tables currently being visited (for cycle detection)
	var result []*schema.Table
	var visit func(t *schema.Table) error
	visit = func(t *schema.Table) error {
		if visited[t.Name] {
			return nil
		}
		if visiting[t.Name] {
			return fmt.Errorf("circular dependency detected at table %s", t.Name)
		}
		visiting[t.Name] = true
		for _, dep := range getTableDependencies(t) {
			if depTable, ok := tableMap[dep]; ok {
				if err := visit(depTable); err != nil {
					return err
				}
			}
		}
		visiting[t.Name] = false
		visited[t.Name] = true
		result = append(result, t)
		return nil
	}
	for _, t := range tables {
		if !visited[t.Name] {
			if err := visit(t); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

// getTableDependencies returns the names of tables that this table depends on
func getTableDependencies(t *schema.Table) []string {
	var deps []string
	depMap := make(map[string]bool)
	for _, rel := range t.Relations {
		if rel.Type == schema.BelongsToRelation {
			// This table has foreign keys pointing to rel.JoinTable
			if !depMap[rel.JoinTable.Name] && rel.JoinTable.Name != t.Name {
				deps = append(deps, rel.JoinTable.Name)
				depMap[rel.JoinTable.Name] = true
			}
		}
	}
	return deps
}
