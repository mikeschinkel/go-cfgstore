package test

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-testutil"
	"github.com/stretchr/testify/require"
)

const (
	truncateLogSlug dt.PathSegment = "cfgstore-truncate-log-test"
	truncateLogFile dt.RelFilepath = "test.log"
	staleLogLine                   = "{\"msg\":\"stale line from an earlier run\"}\n"
)

// writerLoggerFixture holds a DirsProvider rooted in a temp dir and the log
// file path CreateWriterLogger should use under it.
type writerLoggerFixture struct {
	dp      *cfgstore.DirsProvider
	logFile dt.Filepath
}

// newWriterLoggerFixture returns a fixture whose DirsProvider places every
// directory under a new temp root, and restores the package logger when the
// test ends.
func newWriterLoggerFixture(t *testing.T) *writerLoggerFixture {
	t.Helper()
	root := dt.DirPath(t.TempDir())
	dp := &cfgstore.DirsProvider{
		UserHomeDirFunc:   func() (dt.DirPath, error) { return dt.DirPathJoin(root, "home"), nil },
		UserConfigDirFunc: func() (dt.DirPath, error) { return dt.DirPathJoin(root, "config"), nil },
	}
	dp.CLIConfigDirFunc = dp.CLIConfigDirType
	prev := cfgstore.Logger()
	t.Cleanup(func() { cfgstore.SetLogger(prev) })
	configDir, err := cfgstore.CLIConfigDir(truncateLogSlug, dp)
	require.NoError(t, err)
	return &writerLoggerFixture{
		dp:      dp,
		logFile: dt.FilepathJoin(dt.DirPathJoin(configDir, "logs"), truncateLogFile),
	}
}

// args returns WriterLoggerArgs for the fixture's DirsProvider and log file.
func (f *writerLoggerFixture) args() *cfgstore.WriterLoggerArgs {
	return &cfgstore.WriterLoggerArgs{
		DirsProvider: f.dp,
		ConfigSlug:   truncateLogSlug,
		LogFile:      truncateLogFile,
	}
}

func writeStaleLog(t *testing.T, logFile dt.Filepath) {
	t.Helper()
	require.NoError(t, logFile.Dir().MkdirAll(0755))
	require.NoError(t, logFile.WriteFile([]byte(staleLogLine), 0644))
}

func ptr[T any](v T) *T { return &v }

func TestCreateWriterLogger_TruncateLog(t *testing.T) {
	create := func(t *testing.T, f *writerLoggerFixture, truncate bool) {
		t.Helper()
		args := f.args()
		args.Quiet = ptr(true)
		args.TruncateLog = truncate
		_, err := cfgstore.CreateWriterLogger(args)
		require.NoError(t, err)
	}
	t.Run("no log file yet", func(t *testing.T) {
		f := newWriterLoggerFixture(t)
		create(t, f, true)
		content, err := os.ReadFile(string(f.logFile))
		require.NoError(t, err, "the logger creates the log file")
		require.Empty(t, content)
	})
	t.Run("existing log file is emptied", func(t *testing.T) {
		f := newWriterLoggerFixture(t)
		writeStaleLog(t, f.logFile)
		create(t, f, true)
		content, err := os.ReadFile(string(f.logFile))
		require.NoError(t, err)
		require.Empty(t, content)
	})
	t.Run("existing log file is kept without TruncateLog", func(t *testing.T) {
		f := newWriterLoggerFixture(t)
		writeStaleLog(t, f.logFile)
		create(t, f, false)
		content, err := os.ReadFile(string(f.logFile))
		require.NoError(t, err)
		require.Equal(t, staleLogLine, string(content))
	})
}

func TestCreateWriterLogger_IO(t *testing.T) {
	var out, errOut strings.Builder
	f := newWriterLoggerFixture(t)
	args := f.args()
	args.IO = dt.NewStdIO(&out, &errOut)
	wl, err := cfgstore.CreateWriterLogger(args)
	require.NoError(t, err)
	wl.Printf("to the injected output\n")
	wl.Errorf("to the injected error output\n")
	require.Equal(t, "to the injected output\n", out.String())
	require.Equal(t, "to the injected error output\n", errOut.String())
}

// P3: with no Writer and nil Quiet and Verbosity, the built Writer prints at
// cliutil.LowVerbosity.
func TestCreateWriterLogger_DefaultsWithoutWriter(t *testing.T) {
	var out strings.Builder
	f := newWriterLoggerFixture(t)
	args := f.args()
	args.IO = dt.NewStdIO(&out, &out)
	wl, err := cfgstore.CreateWriterLogger(args)
	require.NoError(t, err)
	wl.Printf("v1\n")
	wl.V2().Printf("v2\n")
	require.Equal(t, "v1\n", out.String())
}

// P4: a supplied Writer is used, with provided settings applied and the
// Writer's own settings kept for those not provided.
func TestCreateWriterLogger_SuppliedWriter(t *testing.T) {
	t.Run("settings not provided keep the Writer's own", func(t *testing.T) {
		bw := testutil.NewBufferedWriter() // HighVerbosity, not quiet
		f := newWriterLoggerFixture(t)
		args := f.args()
		args.Writer = bw
		wl, err := cfgstore.CreateWriterLogger(args)
		require.NoError(t, err)
		wl.Printf("v1\n")
		wl.V3().Printf("v3\n")
		require.Equal(t, "v1\nv3\n", bw.GetPrintOutput())
	})
	t.Run("provided Verbosity is applied", func(t *testing.T) {
		bw := testutil.NewBufferedWriter()
		f := newWriterLoggerFixture(t)
		args := f.args()
		args.Writer = bw
		args.Verbosity = ptr(cliutil.LowVerbosity)
		wl, err := cfgstore.CreateWriterLogger(args)
		require.NoError(t, err)
		wl.Printf("v1\n")
		wl.V2().Printf("v2\n")
		require.Equal(t, "v1\n", bw.GetPrintOutput())
	})
	t.Run("provided Quiet is applied", func(t *testing.T) {
		bw := testutil.NewBufferedWriter()
		f := newWriterLoggerFixture(t)
		args := f.args()
		args.Writer = bw
		args.Quiet = ptr(true)
		wl, err := cfgstore.CreateWriterLogger(args)
		require.NoError(t, err)
		wl.Printf("quiet\n")
		require.Empty(t, bw.GetPrintOutput())
		require.True(t, wl.Quiet())
	})
}

// P5: a provided Verbosity of 0 is accepted and silences Printf.
func TestCreateWriterLogger_NoVerbosity(t *testing.T) {
	t.Run("built Writer", func(t *testing.T) {
		var out strings.Builder
		f := newWriterLoggerFixture(t)
		args := f.args()
		args.IO = dt.NewStdIO(&out, &out)
		args.Verbosity = ptr(cliutil.NoVerbosity)
		wl, err := cfgstore.CreateWriterLogger(args)
		require.NoError(t, err)
		wl.Printf("nothing\n")
		require.Empty(t, out.String())
	})
	t.Run("supplied Writer", func(t *testing.T) {
		bw := testutil.NewBufferedWriter()
		f := newWriterLoggerFixture(t)
		args := f.args()
		args.Writer = bw
		args.Verbosity = ptr(cliutil.NoVerbosity)
		wl, err := cfgstore.CreateWriterLogger(args)
		require.NoError(t, err)
		wl.Printf("nothing\n")
		require.Empty(t, bw.GetPrintOutput())
	})
}

// P6: Writer and IO together is an error, returned before any log file exists.
func TestCreateWriterLogger_WriterAndIO(t *testing.T) {
	var out strings.Builder
	f := newWriterLoggerFixture(t)
	args := f.args()
	args.Writer = testutil.NewBufferedWriter()
	args.IO = dt.NewStdIO(&out, &out)
	_, err := cfgstore.CreateWriterLogger(args)
	require.ErrorIs(t, err, cfgstore.ErrWriterAndIO)
	require.ErrorIs(t, err, cfgstore.ErrFailedWriterSetup)
	_, statErr := f.logFile.Stat()
	require.True(t, errors.Is(statErr, os.ErrNotExist), "no log file: %v", statErr)
}

// P7: the log file lands under the supplied DirsProvider's CLI config dir.
func TestCreateWriterLogger_DirsProvider(t *testing.T) {
	f := newWriterLoggerFixture(t)
	args := f.args()
	args.Quiet = ptr(true)
	_, err := cfgstore.CreateWriterLogger(args)
	require.NoError(t, err)
	_, err = f.logFile.Stat()
	require.NoError(t, err, "log file under the DirsProvider's CLI config dir")
}

// P12: an unwritable logs dir falls back to a new temp directory, which
// exists by the time the 'Logging to ... instead' message names it.
func TestCreateWriterLogger_UnwritableLogDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write to any directory")
	}
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	f := newWriterLoggerFixture(t)
	logDir := f.logFile.Dir()
	require.NoError(t, logDir.MkdirAll(0755))
	require.NoError(t, os.Chmod(string(logDir), 0555))
	t.Cleanup(func() { _ = os.Chmod(string(logDir), 0755) })

	bw := testutil.NewBufferedWriter()
	args := f.args()
	args.Writer = bw
	_, err := cfgstore.CreateWriterLogger(args)
	require.NoError(t, err)

	m := regexp.MustCompile(`Logging to (\S+) instead`).FindStringSubmatch(bw.GetErrorOutput())
	require.NotNil(t, m, "fallback message: %q", bw.GetErrorOutput())
	fallback := m[1]
	require.True(t, strings.HasPrefix(fallback, tmp), "%s under %s", fallback, tmp)
	info, err := os.Stat(fallback)
	require.NoError(t, err)
	require.True(t, info.IsDir())
	_, err = os.Stat(filepath.Join(fallback, string(truncateLogFile)))
	require.NoError(t, err, "log file in the fallback dir")
}
