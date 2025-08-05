package bunschema_test

import (
	"os"
	"strings"
	"testing"

	"ariga.io/atlas-provider-bun/bunschema"
	m2m "ariga.io/atlas-provider-bun/internal/testdata/m2m/models"
	"ariga.io/atlas-provider-bun/internal/testdata/models"
	"ariga.io/atlas/sdk/recordriver"
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
	requireEqualContent(t, removeCwd(sql), "testdata/mysql_default.sql")
	resetSession()
	l = bunschema.New("mysql", bunschema.WithJoinTable(&m2m.OrderToItem{}))
	sql, err = l.Load(
		(*m2m.Item)(nil),
		(*m2m.Order)(nil),
		(*m2m.OrderToItem)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/mysql_m2m.sql")
}

func TestSQLiteConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("sqlite")
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/sqlite_default.sql")
	resetSession()
	l = bunschema.New("sqlite", bunschema.WithJoinTable(&m2m.OrderToItem{}))
	sql, err = l.Load(
		(*m2m.Item)(nil),
		(*m2m.Order)(nil),
		(*m2m.OrderToItem)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/sqlite_m2m.sql")
}

func TestPostgreSQLConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("postgres")
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/postgres_default.sql")
	resetSession()
	l = bunschema.New("postgres", bunschema.WithJoinTable(&m2m.OrderToItem{}))
	sql, err = l.Load(
		(*m2m.Item)(nil),
		(*m2m.Order)(nil),
		(*m2m.OrderToItem)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/postgres_m2m.sql")
}

func TestSQLServerConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("mssql", bunschema.WithStmtDelimiter("\nGO"))
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/mssql_default.sql")
	resetSession()
	l = bunschema.New("mssql", bunschema.WithStmtDelimiter("\nGO"), bunschema.WithJoinTable(&m2m.OrderToItem{}))
	sql, err = l.Load(
		(*m2m.Item)(nil),
		(*m2m.Order)(nil),
		(*m2m.OrderToItem)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/mssql_m2m.sql")
}

func TestOracleConfig(t *testing.T) {
	resetSession()
	l := bunschema.New("oracle")
	sql, err := l.Load(
		(*models.User)(nil),
		(*models.Story)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/oracle_default.sql")
	resetSession()
	l = bunschema.New("oracle", bunschema.WithJoinTable(&m2m.OrderToItem{}))
	sql, err = l.Load(
		(*m2m.Item)(nil),
		(*m2m.Order)(nil),
		(*m2m.OrderToItem)(nil),
	)
	require.NoError(t, err)
	requireEqualContent(t, removeCwd(sql), "testdata/oracle_m2m.sql")
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

// removeCwd converts absolute paths to relative paths in the test
func removeCwd(text string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return text
	}
	projectRoot := strings.TrimSuffix(cwd, string(os.PathSeparator)+"bunschema")
	result := strings.ReplaceAll(text, projectRoot+string(os.PathSeparator), "")
	return result
}
