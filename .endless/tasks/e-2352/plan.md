# Add TruncateLog to WriterLoggerArgs

The uncommitted work on main for this task is in `writer_logger.go`.

## Steps
1. Keep the new `TruncateLog bool` field on `WriterLoggerArgs`.
2. Fix the bug in the WIP: `fp.Remove()` fails on the first run, when the log file does not exist yet. Treat a not-exist error from the remove as success.
3. The WIP uses `dt.ErrFailedtoRemoveFile`, which exists only in go-dt v0.6.0 and later. Bump go-dt to the version E-2209 releases.
4. Add a test: with `TruncateLog` and no existing file, the call succeeds; with an existing file, its old content is gone.
5. Leave the TODO about `logutil.CreateJSONFileLogger` taking a truncation option. That belongs to go-logutil.

## Blocked by
E-2209 (go-dt version that has `ErrFailedtoRemoveFile`).
