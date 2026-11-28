package cfgstore

import (
	"errors"
	"io/fs"
	"log/slog"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-logutil"
)

var ErrFailedWriterSetup = errors.New("failed writer setup")
var ErrFailedLoggerSetup = errors.New("failed logger setup")

// ErrWriterAndIO reports a WriterLoggerArgs with both Writer and IO set. A
// Writer already has its streams, and a cliutil.Writer cannot be given new ones.
var ErrWriterAndIO = errors.New("writer and IO both provided")

// ErrFailedToCreateTempDir reports a failure creating the temp directory that
// replaces an unwritable log directory.
var ErrFailedToCreateTempDir = errors.New("failed to create temp directory")

// WriterLoggerArgs configures CreateWriterLogger. Quiet and Verbosity are
// pointers because their zero values are real settings: Quiet false prints,
// and Verbosity 0 is cliutil.NoVerbosity, which prints nothing from Printf.
type WriterLoggerArgs struct {
	Writer       cliutil.Writer     // Optional; nil means build one from IO, Quiet and Verbosity.
	IO           dt.StdIO           // Optional; nil means not provided. Must be nil when Writer is set.
	Quiet        *bool              // Optional; nil means not provided.
	Verbosity    *cliutil.Verbosity // Optional; nil means not provided.
	DirsProvider *DirsProvider      // Optional; nil means DefaultDirsProvider().
	ConfigSlug   dt.PathSegment
	LogFile      dt.RelFilepath
	TruncateLog  bool // Remove the log file before the logger opens it, so the log starts empty.
}

func CreateWriterLogger(args *WriterLoggerArgs) (wr cliutil.WriterLogger, err error) {
	var configDir dt.DirPath
	var logger *slog.Logger
	var logDir dt.DirPath

	var writer cliutil.Writer

	writer, err = newWriter(args)
	if err != nil {
		goto end
	}

	configDir, err = CLIConfigDir(args.ConfigSlug, args.DirsProvider)
	if err != nil {
		err = dt.NewErr(
			ErrFailedWriterSetup,
			ErrFailedGettingUserConfigDir,
			err,
		)
		goto end
	}
	logDir = dt.DirPathJoin(configDir, "logs")
	err = logDir.MkdirAll(0755)
	if err != nil {
		goto end
	}

	logger, err = createLogger(logDir, writer, args)
	if err != nil {
		goto end
	}
	wr = cliutil.NewWriterLogger(writer, logger)

end:
	return wr, err
}

// newWriter returns args.Writer with any provided Quiet and Verbosity applied,
// or, with no Writer, a new one over args.IO. A nil Quiet builds a Writer that
// prints, and a nil Verbosity builds one at cliutil.LowVerbosity, the
// --verbosity default.
func newWriter(args *WriterLoggerArgs) (w cliutil.Writer, err error) {
	if args.Writer == nil {
		wa := &cliutil.WriterArgs{
			IO:        args.IO,
			Verbosity: cliutil.LowVerbosity,
		}
		if args.Quiet != nil {
			wa.Quiet = *args.Quiet
		}
		if args.Verbosity != nil {
			wa.Verbosity = *args.Verbosity
		}
		w = cliutil.NewWriter(wa)
		goto end
	}
	if args.IO != nil {
		err = dt.NewErr(ErrFailedWriterSetup, ErrWriterAndIO)
		goto end
	}
	w = args.Writer
	if args.Quiet != nil {
		w = w.WithQuiet(*args.Quiet)
	}
	if args.Verbosity != nil {
		w = w.WithVerbosity(*args.Verbosity)
	}
end:
	return w, err
}

func createLogger(logDir dt.DirPath, writer cliutil.Writer, args *WriterLoggerArgs) (logger *slog.Logger, err error) {
	var canWrite bool
	var fp dt.Filepath
	var tmpDir dt.DirPath

	canWrite, _ = logDir.CanWrite()
	if !canWrite {
		tmpDir, err = dt.MkdirTemp(dt.TempDir(), string(args.ConfigSlug)+"-*")
		if err != nil {
			err = dt.NewErr(ErrFailedToCreateTempDir, err)
			goto end
		}
		writer.Errorf("Cannot write to %s. Logging to %s instead\n", logDir, tmpDir)
		logDir = tmpDir
	}
	err = logDir.MkdirAll(0755)
	if err != nil {
		err = dt.NewErr(dt.ErrFailedToMakeDirectory,
			"log_dir", logDir,
			err)
		goto end
	}
	fp = dt.FilepathJoin(logDir, args.LogFile)
	if args.TruncateLog {
		err = truncateLog(fp)
		if err != nil {
			goto end
		}
	}
	// TODO Add args to logutil.CreateJSONFileLogger so it can open the file with
	//      truncation, in place of truncateLog() removing it first.
	logger, err = logutil.CreateJSONFileLogger(fp)
	if err != nil {
		err = dt.NewErr(dt.ErrFailedtoCreateFile,
			"log_file", fp,
			err,
		)
		goto end
	}
	SetLogger(logger)
end:
	if err != nil {
		err = dt.WithErr(err, ErrFailedLoggerSetup)
	}
	return logger, err
}

// truncateLog removes the log file so the logger starts it empty. A log file
// that does not exist yet has nothing to truncate, so that is not an error.
func truncateLog(fp dt.Filepath) (err error) {
	err = fp.Remove()
	if err == nil {
		goto end
	}
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
		goto end
	}
	err = dt.NewErr(dt.ErrFailedtoRemoveFile, fp.ErrKV(), err)
end:
	return err
}
