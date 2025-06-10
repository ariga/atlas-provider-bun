package bunschema

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"ariga.io/atlas-go-sdk/recordriver"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mssqldialect"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/oracledialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/schema"
)

// Option is a function that configures the Loader.
type Option func(*Loader)

// Loader is the struct that holds the loader configuration.
type Loader struct {
	dialect   string
	delimiter string
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
	db.RegisterModel(models...)
	for _, m := range models {
		if _, err := db.NewCreateTable().WithForeignKeys().Model(m).Exec(context.Background()); err != nil {
			return "", err
		}
	}
	ss, ok := recordriver.Session("bun")
	if !ok {
		return "", fmt.Errorf("failed to read session")
	}
	var buf strings.Builder
	for _, stmt := range ss.Statements {
		if _, err = fmt.Fprintln(&buf, stmt+l.delimiter); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

// validate ensures that the provided model is a pointer to a struct.
// This is required by Bun; otherwise, it will panic at runtime.
func validate(v any) error {
	t := reflect.TypeOf(v)
	if t == nil {
		return fmt.Errorf("model is nil")
	}
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be 1 a pointer, got %s", t.Kind())
	}
	if t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("model must point to a struct, got pointer to %s", t.Elem().Kind())
	}
	return nil
}
