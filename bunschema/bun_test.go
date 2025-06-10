package bunschema_test

import (
	"os"
	"testing"

	"ariga.io/atlas-go-sdk/recordriver"
	"ariga.io/atlas-provider-bun/bunschema"
	"ariga.io/atlas-provider-bun/internal/testdata/models"
	"github.com/stretchr/testify/require"
)

func TestMySQLConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("mysql")
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mysql_default.sql")
}

func TestSQLiteConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("sqlite")
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/sqlite_default.sql")
}

func TestPostgreSQLConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("postgres")
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/postgres_default.sql")
}

func TestSQLServerConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("mssql")
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/mssql_default.sql")
}

func TestOracleConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("oracle")
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, sql, "testdata/oracle_default.sql")
}

func resetSession() {
	sess, ok := recordriver.Session("bun")
	if ok {
		sess.Statements = nil
	}
}

func requireEqualContent(t *testing.T, actual, fileName string) {
	buf, err := os.ReadFile(fileName)
	require.NoError(t, err)
	require.Equal(t, string(buf), actual)
}
