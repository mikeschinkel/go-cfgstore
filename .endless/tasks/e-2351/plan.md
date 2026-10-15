# Use go-dt's NoFileOrDirErr and sync doterr.go

The uncommitted work on main for this task is in: `util.go`, `config_store.go`, `errors.go`, `doterr.go`, `go.mod`, `go.sum`, `test/go.mod`, `test/go.sum`, `examples/basic_usage/go.mod`, `examples/basic_usage/go.sum`.

## Steps
1. Delete `cfgstore.NoFileOrDirErr` from `util.go` and call `dtx.NoFileOrDirErr` in `util.go` and `config_store.go` (`Load`). This removes an exported function; Mike approved the breaking change.
2. Rename `ErrFileDoesNotExist` to `ErrFileNotExist` in `errors.go` (the WIP spelled it `ErrFileNotExistExist`, a typo) and at its use in `config_store.go`. Mike asked for this rename for consistency.
3. `doterr.go`: re-sync from go-doterr with its `make sync` once go-doterr has committed its own change (the `Sprintf` to `Fprintf` change). Never hand-edit the copy; today it matches go-doterr's uncommitted working tree apart from the package line.
4. Drop the re-indent of the `InitProjectConfig` parameters; gofmt reverts it anyway.
5. Bump go-dt and go-dt/dtx to the tags E-2209 releases, and drop the `replace` lines for go-dt and dtx from `go.mod`.
6. `go mod tidy` in the root, `test/` and `examples/basic_usage/`, then build, vet and run the tests with `GOEXPERIMENT=jsonv2`.

## Blocked by
E-2209: `dtx.NoFileOrDirErr` is committed in go-dt but not in any released dtx tag (`dtx/v0.2.1` lacks it).
