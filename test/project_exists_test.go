package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
	"github.com/stretchr/testify/require"
)

// TestProjectConfigStoreExists covers ProjectConfigStoreExists given a project
// directory: it finds <projectDir>/.<slug>/<file> and nothing else.
func TestProjectConfigStoreExists(t *testing.T) {
	const slug dt.PathSegment = "myapp"
	const file dt.RelFilepath = "config.json"

	write := func(t *testing.T, path string) {
		t.Helper()
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644))
	}

	t.Run("finds the config in the project's slug dir", func(t *testing.T) {
		projectDir := t.TempDir()
		write(t, filepath.Join(projectDir, "."+string(slug), string(file)))
		require.True(t, cfgstore.ProjectConfigStoreExists(dt.DirPath(projectDir), slug, file))
	})

	t.Run("ignores a config directly in the project dir", func(t *testing.T) {
		projectDir := t.TempDir()
		write(t, filepath.Join(projectDir, string(file)))
		require.False(t, cfgstore.ProjectConfigStoreExists(dt.DirPath(projectDir), slug, file))
	})

	t.Run("given the slug dir instead of the project dir, finds nothing", func(t *testing.T) {
		projectDir := t.TempDir()
		slugDir := filepath.Join(projectDir, "."+string(slug))
		write(t, filepath.Join(slugDir, string(file)))
		require.False(t, cfgstore.ProjectConfigStoreExists(dt.DirPath(slugDir), slug, file))
	})
}
