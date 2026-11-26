Found while planning E-2672, 2026-11-25; Mike asked for a bugfix.

How it works today: createLogger in writer_logger.go handles an unwritable log dir by calling dt.CreateTemp(dt.TempDir(), slug+"-*"), which is os.CreateTemp and so creates a file, then sets logDir = dt.DirPath(tmpFile.Name()) and calls logDir.MkdirAll(0755). MkdirAll on a path that is an existing file fails.

Evidence: a standalone program doing os.CreateTemp then os.MkdirAll on its name printed "mkdir <TMPDIR>/gomion-4090341833: not a directory". So when the CLI config logs dir cannot be written, CreateWriterLogger returns ErrFailedLoggerSetup instead of logging to the temp dir its "Logging to ... instead" message promises. That message is also printed before the CreateTemp error is checked.
