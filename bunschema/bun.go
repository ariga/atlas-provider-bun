package bunschema

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"slices"
	"strings"

	"ariga.io/atlas-go-sdk/recordriver"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mssqldialect"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/oracledialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/schema"
	"golang.org/x/tools/go/packages"
)

// Option is a function that configures the Loader.
type Option func(*Loader)

// Loader is the struct that holds the loader configuration.
type Loader struct {
	dialect    string
	delimiter  string
	joinTables []any
}

// New creates a new Loader.
func New(dialect string, opts ...Option) *Loader {
	l := &Loader{
		dialect:   dialect,
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
	case "mysql":
		di = mysqldialect.New()
		recordriver.SetResponse("bun", "SELECT version()", &recordriver.Response{
			Cols: []string{"version()"},
			Data: [][]driver.Value{{"8.0.24"}},
		})

	case "sqlite":
		di = sqlitedialect.New()
		recordriver.SetResponse("bun", "select sqlite_version()", &recordriver.Response{
			Cols: []string{"sqlite_version()"},
			Data: [][]driver.Value{{"3.30.1"}},
		})

	case "mssql":
		di = mssqldialect.New()
		recordriver.SetResponse("bun", "SELECT @@VERSION", &recordriver.Response{
			Cols: []string{"SELECT @@VERSION"},
			Data: [][]driver.Value{{"15.0.2000.58"}},
		})
	case "oracle":
		di = oracledialect.New()
	case "postgres":
		di = pgdialect.New()
	default:
		return "", errors.New("unsupported dialect: " + l.dialect)
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
	if l.dialect == "oracle" {
		for _, t := range tables {
			for _, rel := range t.Relations {
				// Oracle does not support ON UPDATE, but Bun sets it to NO ACTION by default
				// Tracking issue: https://github.com/uptrace/bun/issues/1212
				rel.OnUpdate = ""
				// Oracle supports ON DELETE CASCADE, and SET NULL only, but Bun sets it to NO ACTION by default
				if rel.OnDelete != "CASCADE" && rel.OnDelete != "SET NULL" {
					rel.OnDelete = ""
				}
			}
		}
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
	var buf strings.Builder
	pos, err := l.modelsPos(models...)
	if err != nil {
		return "", fmt.Errorf("failed to get models position: %w", err)
	}
	for _, t := range tables {
		if p, ok := pos[typeID(t.Type)]; ok {
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

func (l *Loader) modelsPos(models ...any) (map[string]string, error) {
	pos := make(map[string]string)
	for _, m := range models {
		p, err := location(m)
		if err != nil {
			return nil, err
		}
		pos[typeID(reflect.TypeOf(m))] = p
	}
	return pos, nil
}

// location uses the go/packages library to find the file and line number
// of a struct definition.
func location(model any) (string, error) {
	// Bun Models are always pointers to structs.
	if err := validate(model); err != nil {
		return "", fmt.Errorf("invalid model %T: %w", model, err)
	}
	typ := reflect.TypeOf(model).Elem()
	pkgPath := typ.PkgPath()
	if pkgPath == "" {
		return "", fmt.Errorf("could not determine package path for struct '%s'", typ.Name())
	}
	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax,
		Fset: fset,
	}, pkgPath)
	if err != nil {
		return "", fmt.Errorf("failed to load package: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return "", fmt.Errorf("errors while loading package %s", pkgPath)
	}
	for _, pkg := range pkgs {
		for i, fileNode := range pkg.Syntax {
			var found *ast.TypeSpec
			ast.Inspect(fileNode, func(n ast.Node) bool {
				if found != nil {
					return false // Stop traversal once found
				}
				typeSpec, ok := n.(*ast.TypeSpec)
				if ok && typeSpec.Name != nil && typeSpec.Name.Name == typ.Name() {
					found = typeSpec
					return false
				}
				return true
			})
			if found != nil {
				return fmt.Sprintf("%s:%d-%d",
					pkg.GoFiles[i],
					fset.Position(found.Pos()).Line,
					fset.Position(found.Type.End()).Line,
				), nil
			}
		}
	}
	return "", fmt.Errorf("struct '%s' not found in package '%s'", typ.Name(), pkgPath)
}

func typeID(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

// validate ensures that the provided model is a pointer to a struct.
// This is required by Bun; otherwise, it will panic at runtime.
func validate(v any) error {
	t := reflect.TypeOf(v)
	if t == nil {
		return fmt.Errorf("model is nil")
	}
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be a pointer, got %s", t.Kind())
	}
	if t.Elem().Kind() != reflect.Struct {
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
