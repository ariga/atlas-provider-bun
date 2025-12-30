package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"ariga.io/atlas-provider-bun/bunschema"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	for _, d := range []string{"mysql", "sqlite", "postgres", "mssql", "oracle"} {
		t.Run(d, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &LoadCmd{
				Path:    "./internal/testdata/models",
				Dialect: bunschema.Dialect(d),
				out:     &buf,
			}
			err := cmd.Run()
			require.NoError(t, err)
			files, err := os.ReadDir("bunschema/testdata")
			require.NoError(t, err)
			cwd, err := os.Getwd()
			require.NoError(t, err)
			for _, file := range files {
				if strings.HasPrefix(file.Name(), d+"_default.sql") {
					content, err := os.ReadFile("bunschema/testdata/" + file.Name())
					require.NoError(t, err)
					bufStr := strings.ReplaceAll(buf.String(), cwd+string(os.PathSeparator), "")
					require.Equal(t, string(content), bufStr)
				}
			}
		})
	}
}

func TestBuildTags(t *testing.T) {
	var buf bytes.Buffer
	cmd := &LoadCmd{
		Path:      "./internal/testdata/buildtags",
		Dialect:   "mysql",
		BuildTags: "buildtag",
		out:       &buf,
	}
	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, buf.String(), "-- atlas:pos untagged_models[type=table]")
	require.Contains(t, buf.String(), "-- atlas:pos tagged_models[type=table]")
	require.Contains(t, buf.String(), "CREATE TABLE `untagged_models`")
	require.Contains(t, buf.String(), "CREATE TABLE `tagged_models`")
}

func TestNonBuildTags(t *testing.T) {
	var buf bytes.Buffer
	cmd := &LoadCmd{
		Path:    "./internal/testdata/buildtags",
		Dialect: "mysql",
		out:     &buf,
	}
	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, buf.String(), "-- atlas:pos untagged_models[type=table]")
	require.Contains(t, buf.String(), "CREATE TABLE `untagged_models`")
	require.NotContains(t, buf.String(), "CREATE TABLE `tagged_models`")
}

func TestM2MStandalone(t *testing.T) {
	for _, d := range []string{"mysql", "sqlite", "postgres", "mssql", "oracle"} {
		t.Run(d, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &LoadCmd{
				Path:    "./internal/testdata/m2m/models",
				Dialect: bunschema.Dialect(d),
				out:     &buf,
			}
			err := cmd.Run()
			require.NoError(t, err)
			cwd, err := os.Getwd()
			require.NoError(t, err)
			content, err := os.ReadFile("bunschema/testdata/" + d + "_m2m.sql")
			require.NoError(t, err)
			bufStr := strings.ReplaceAll(buf.String(), cwd+string(os.PathSeparator), "")
			require.Equal(t, string(content), bufStr, "standalone m2m should match golden file for %s", d)
		})
	}
}
