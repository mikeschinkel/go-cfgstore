package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
	"github.com/stretchr/testify/require"
)

const (
	truncateLogSlug dt.PathSegment = "cfgstore-truncate-log-test"
	truncateLogFile dt.RelFilepath = "test.log"
	staleLogLine                   = "{\"msg\":\"stale line from an earlier run\"}\n"
)

// writerLoggerSetup points the CLI config dir at a temporary home, restores
// the package logger when the test ends, and returns the log file's path.
func writerLoggerSetup(t *testing.T) dt.Filepath {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	prev := cfgstore.Logger()
	t.Cleanup(func() { cfgstore.SetLogger(prev) })
	configDir, err := cfgstore.CLIConfigDir(truncateLogSlug)
	require.NoError(t, err)
	return dt.FilepathJoin(dt.DirPathJoin(configDir, "logs"), truncateLogFile)
}

func writeStaleLog(t *testing.T, logFile dt.Filepath) {
	t.Helper()
	require.NoError(t, logFile.Dir().MkdirAll(0755))
	require.NoError(t, logFile.WriteFile([]byte(staleLogLine), 0644))
}

func createWriterLogger(t *testing.T, truncate bool) {
	t.Helper()
	_, err := cfgstore.CreateWriterLogger(&cfgstore.WriterLoggerArgs{
		Quiet:       true,
		Verbosity:   cliutil.LowVerbosity,
		ConfigSlug:  truncateLogSlug,
		LogFile:     truncateLogFile,
		TruncateLog: truncate,
	})
	require.NoError(t, err)
}

func TestCreateWriterLogger_TruncateLog(t *testing.T) {
	t.Run("no log file yet", func(t *testing.T) {
		logFile := writerLoggerSetup(t)
		createWriterLogger(t, true)
		content, err := os.ReadFile(string(logFile))
		require.NoError(t, err, "the logger creates the log file")
		require.Empty(t, content)
	})
	t.Run("existing log file is emptied", func(t *testing.T) {
		logFile := writerLoggerSetup(t)
		writeStaleLog(t, logFile)
		createWriterLogger(t, true)
		content, err := os.ReadFile(string(logFile))
		require.NoError(t, err)
		require.Empty(t, content)
	})
	t.Run("existing log file is kept without TruncateLog", func(t *testing.T) {
		logFile := writerLoggerSetup(t)
		writeStaleLog(t, logFile)
		createWriterLogger(t, false)
		content, err := os.ReadFile(string(logFile))
		require.NoError(t, err)
		require.Equal(t, staleLogLine, string(content))
	})
}

func TestCreateWriterLogger_IO(t *testing.T) {
	var out, errOut strings.Builder
	writerLoggerSetup(t)
	wl, err := cfgstore.CreateWriterLogger(&cfgstore.WriterLoggerArgs{
		IO:         dt.NewStdIO(&out, &errOut),
		Verbosity:  cliutil.LowVerbosity,
		ConfigSlug: truncateLogSlug,
		LogFile:    truncateLogFile,
	})
	require.NoError(t, err)
	wl.Printf("to the injected output\n")
	wl.Errorf("to the injected error output\n")
	require.Equal(t, "to the injected output\n", out.String())
	require.Equal(t, "to the injected error output\n", errOut.String())
}
