Filed from gomion's planning of RunCLIArgs (the RunCLI signature change), 2026-11-24, at Mike's request.

How it works today: cfgstore.CreateWriterLogger (writer_logger.go) always builds its own cliutil.Writer from WriterLoggerArgs.IO via cliutil.NewWriter, and puts the log file under CLIConfigDir(args.ConfigSlug) with no DirsProvider, so the log always lands in the real user config dir.

What prompted it:
- Mike wants gomion's in-process CLI tests to inject go-testutil's BufferedWriter (a cliutil.Writer built for tests) directly, rather than only a StdIO over buffers. Since gomion gets its Writer from CreateWriterLogger, that needs a way to hand CreateWriterLogger a ready-made Writer.
- gomion's in-process test runs the CLI against a temp config dir through a DirsProvider. The log ignores that and, with TruncateLog true, would empty gomion's real log file in the user's config dir.

Mike, 2026-11-24: the DirsProvider must be optional.
