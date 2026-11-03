package test

import (
	"os"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unknownKeysJSON holds only keys mergeConfig does not declare, so it decodes
// to a zero mergeConfig. An older binary reading a newer binary's config file
// sees exactly this.
const unknownKeysJSON = "{\n  \"tmux\": {\n    \"socket_dir\": \"/tmp/endless-tmux\"\n  }\n}\n"

// ensureFixture is a mergeFixture whose CLI store is the one under test.
type ensureFixture struct {
	*mergeFixture
	store cfgstore.ConfigStore
}

func newEnsureFixture(t *testing.T) *ensureFixture {
	t.Helper()
	f := newMergeFixture(t)
	return &ensureFixture{
		mergeFixture: f,
		store:        f.stores(cfgstore.CLIConfigDirType).CLIConfigStore(),
	}
}

// load runs LoadConfigStores for the CLI layer only, which ensures the config.
func (f *ensureFixture) load(t *testing.T) *mergeConfig {
	t.Helper()
	cfg, err := cfgstore.LoadConfigStores[mergeConfig](
		f.stores(cfgstore.CLIConfigDirType),
		cfgstore.RootConfigArgs{DirTypes: []cfgstore.DirType{cfgstore.CLIConfigDirType}},
	)
	require.NoError(t, err, "loading CLI config")
	return cfg
}

func (f *ensureFixture) filepath(t *testing.T) dt.Filepath {
	t.Helper()
	fp, err := f.store.GetFilepath()
	require.NoError(t, err, "getting config filepath")
	return fp
}

// TestEnsureConfig_LeavesUnknownKeysFileUnchanged covers P1: an existing file
// whose keys the struct does not declare decodes to zero and must not be
// rewritten.
func TestEnsureConfig_LeavesUnknownKeysFileUnchanged(t *testing.T) {
	f := newEnsureFixture(t)
	fp := f.filepath(t)
	require.NoError(t, fp.Dir().MkdirAll(0o755))
	require.NoError(t, os.WriteFile(string(fp), []byte(unknownKeysJSON), 0o644))

	cfg := f.load(t)

	assert.Equal(t, &mergeConfig{}, cfg, "a file with only unknown keys loads as zero")
	got, err := os.ReadFile(string(fp))
	require.NoError(t, err)
	assert.Equal(t, unknownKeysJSON, string(got), "ensuring the config must not rewrite an existing file")
}

// TestEnsureConfig_CreatesMissingFile covers P2: a missing file is still created.
func TestEnsureConfig_CreatesMissingFile(t *testing.T) {
	f := newEnsureFixture(t)
	require.False(t, f.store.Exists(), "config file must start out missing")

	cfg := f.load(t)

	assert.Equal(t, &mergeConfig{}, cfg)
	assert.True(t, f.store.Exists(), "ensuring the config must create a missing file")
}

// TestEnsureConfig_LoadsKnownKeys covers P3: an existing file with known keys
// loads into the struct and is left as it was.
func TestEnsureConfig_LoadsKnownKeys(t *testing.T) {
	f := newEnsureFixture(t)
	f.write(t, cfgstore.CLIConfigDirType, &mergeConfig{CLIOnly: "cli", Both: "both"})
	fp := f.filepath(t)
	before, err := os.ReadFile(string(fp))
	require.NoError(t, err)

	cfg := f.load(t)

	assert.Equal(t, &mergeConfig{CLIOnly: "cli", Both: "both"}, cfg)
	after, err := os.ReadFile(string(fp))
	require.NoError(t, err)
	assert.Equal(t, string(before), string(after), "loading must not rewrite an existing file")
}
