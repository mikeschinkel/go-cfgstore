# Fix mergeRootConfigs returning the unmerged project layer

## The defect

`mergeRootConfigs` (config_stores.go) folds each layer into `rc` with
`rc = rcMap[typ].Merge(rc)`, then returns `prc = rcMap[dirType]`, which is the
last layer as loaded, not the merged result. When a project config file
exists, the caller gets the project layer alone, and every value meant to fall
through from the CLI layer is lost. Endless confirmed it with a probe: with a
project `config.json` present, `config.Load` returned `tracking=""` and
`tmux.session_name=""`, although both were set in the user config.

## Fix

Return the merged `rc` (as the `PRC` type), not the receiver as loaded. Add a
test in go-cfgstore: CLI layer sets a field, project layer sets a different
one, and the result carries both, with project precedence on overlap.

## Then

Releasing the fix needs the dependency chain released first (its own task).
Once go-cfgstore is tagged, Endless bumps it and re-checks that user-level
`tracking`, `checks` and `tmux.session_name` reach a project that has its own
config file. File that check as an `endless` task when this lands.
