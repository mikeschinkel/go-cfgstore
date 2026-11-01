# Add TruncateLog to WriterLoggerArgs

This task's work-in-progress is parked on branch `wip/e2351-e2352`, together with the NoFileOrDirErr task's work and a PLAN.md that belongs to neither. Take only this task's change from it: `git checkout wip/e2351-e2352 -- writer_logger.go`, plus any go.mod hunk it needs.

## Steps
1. Keep the new `TruncateLog bool` field on `WriterLoggerArgs`.
2. Fix the WIP's bug: `fp.Remove()` fails on the first run, when the log file does not exist yet. Treat a not-exist error from the remove as success.
3. The WIP uses `dt.ErrFailedtoRemoveFile`, which go-dt has from v0.6.0. Require go-dt v0.7.0 or later (published).
4. Add a test: with `TruncateLog` and no existing file, the call succeeds; with an existing file, its old content is gone.
5. Leave the TODO about `logutil.CreateJSONFileLogger` taking a truncation option; that belongs to go-logutil.
6. Build, vet and run the tests with `GOWORK=off`.
