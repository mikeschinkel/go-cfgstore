# Use go-dt's NoFileOrDirErr and sync doterr.go

This task's work-in-progress is parked on branch `wip/e2351-e2352`, together with the TruncateLog task's work and a PLAN.md that belongs to neither. Take only this task's files from it: `git checkout wip/e2351-e2352 -- util.go config_store.go errors.go doterr.go go.mod go.sum test/go.mod test/go.sum examples/basic_usage/go.mod examples/basic_usage/go.sum`, then review each hunk; config_store.go and go.mod may also carry the other task's hunks.

## Steps
1. Delete `cfgstore.NoFileOrDirErr` from `util.go` and call `dtx.NoFileOrDirErr` in `util.go` and in `config_store.go` (`Load`). This removes an exported function; the owner approved the breaking change.
2. Rename `ErrFileDoesNotExist` to `ErrFileNotExist` in `errors.go` and at its uses (the WIP spelled it `ErrFileNotExistExist`, a typo). The owner asked for this rename for consistency.
3. `doterr.go`: re-sync from go-doterr with its `make sync`; go-doterr's `Fprintf` change is committed. Never hand-edit the copy.
4. Drop the WIP's re-indent of the `InitProjectConfig` parameters; gofmt reverts it anyway.
5. Require go-dt v0.7.0 and go-dt/dtx dtx/v0.3.0 (or later published tags; dtx/v0.3.0 is the first with `NoFileOrDirErr`) in the root, `test/` and `examples/basic_usage/` go.mod files, and drop the root go.mod's `replace` lines for go-dt and dtx.
6. `go mod tidy` in each of those modules, then build, vet and run the tests with `GOWORK=off`. Go 1.27 needs no `GOEXPERIMENT=jsonv2`.
