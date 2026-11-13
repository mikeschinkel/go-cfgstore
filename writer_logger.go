package cfgstore

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-logutil"
)

var ErrFailedWriterSetup = errors.New("failed writer setup")
var ErrFailedLoggerSetup = errors.New("failed logger setup")

type WriterLoggerArgs struct {
	IO          dt.StdIO
	Quiet       bool
	Verbosity   cliutil.Verbosity
	ConfigSlug  dt.PathSegment
	LogFile     dt.RelFilepath
	TruncateLog bool // Remove the log file before the logger opens it, so the log starts empty.
}

func CreateWriterLogger(args *WriterLoggerArgs) (wr cliutil.WriterLogger, err error) {
	var configDir dt.DirPath
	var logger *slog.Logger
	var logDir dt.DirPath

	writer := cliutil.NewWriter(&cliutil.WriterArgs{
		IO:        args.IO,
		Quiet:     args.Quiet,
		Verbosity: args.Verbosity,
	})

	configDir, err = CLIConfigDir(args.ConfigSlug)
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

func createLogger(logDir dt.DirPath, writer cliutil.Writer, args *WriterLoggerArgs) (logger *slog.Logger, err error) {
	var tmpFile *os.File
	var canWrite bool
	var fp dt.Filepath

	canWrite, _ = logDir.CanWrite()
	if !canWrite {
		tmpDir := dt.TempDir()
		tmpFile, err = dt.CreateTemp(tmpDir, string(args.ConfigSlug)+"-*")
		writer.Errorf("Cannot write to %s. Logging to %s/%s-* instead\n", logDir, tmpDir, args.ConfigSlug)
		if err != nil {
			err = dt.NewErr(dt.ErrFailedtoCreateTempFile, err)
			goto end
		}
		defer dt.CloseOrLog(tmpFile)
		logDir = dt.DirPath(tmpFile.Name())
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
