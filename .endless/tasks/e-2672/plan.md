**Fields.** In writer_logger.go, WriterLoggerArgs becomes:

```go
type WriterLoggerArgs struct {
	Writer       cliutil.Writer     // optional; nil means build one from IO, Quiet and Verbosity
	IO           dt.StdIO           // optional; nil means not provided
	Quiet        *bool              // optional; nil means not provided
	Verbosity    *cliutil.Verbosity // optional; nil means not provided
	DirsProvider *DirsProvider      // optional; nil means DefaultDirsProvider()
	ConfigSlug   dt.PathSegment
	LogFile      dt.RelFilepath
	TruncateLog  bool
}
```

Quiet and Verbosity are pointers because their zero values are real settings: Quiet false is "print", and Verbosity 0 is cliutil.NoVerbosity, which prints nothing from Printf (Mike, 2026-11-25: 0 stays a real level).

**Building the Writer.** With no Writer, CreateWriterLogger calls cliutil.NewWriter with IO, *Quiet (false when nil) and *Verbosity (cliutil.LowVerbosity when nil, the --verbosity default). With a Writer, it applies each provided setting, `w = w.WithQuiet(*args.Quiet)` and `w = w.WithVerbosity(*args.Verbosity)`, and leaves the others as the Writer has them. That needs E-2673's With methods, hence blocked by E-2673. IO together with a Writer is an error: a Writer already has its streams and cliutil.Writer cannot be given new ones, so the call returns ErrWriterAndIO, wrapped in ErrFailedWriterSetup, before touching the log dir.

**Log location.** CreateWriterLogger calls CLIConfigDir(args.ConfigSlug, args.DirsProvider); CLIConfigDir already treats a nil provider as DefaultDirsProvider().

**Callers.** examples/comprehensive and the tests in test/ move to the pointer fields. gomion consumes this change only through a tagged go-cfgstore release (ED-1639; Mike, 2026-11-25), so after landing, this task tags the next go-cfgstore minor version; the field type changes break callers.

**Tests** in test/writer_logger_test.go, each with a DirsProvider rooted under t.TempDir():
- no Writer, nil Quiet and Verbosity: output at verbosity 1 appears and V2 output does not;
- a supplied testutil-style Writer with Quiet and Verbosity provided: the settings take effect; with them nil, the Writer's own settings stand;
- Verbosity 0 provided: Printf writes nothing;
- IO and Writer together: ErrWriterAndIO and no log file;
- the log file lands under the temp root's CLI config dir.
