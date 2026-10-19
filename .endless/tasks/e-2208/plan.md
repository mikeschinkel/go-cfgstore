# Return the merged config from mergeRootConfigs

## The defect

`mergeRootConfigs` (config_stores.go) folds each layer into `rc` with
`rc = rcMap[typ].Merge(rc)`, then returns `prc = rcMap[dirType]`: the last
layer as loaded, not the merged result. Whenever a project config exists, the
caller gets the project layer alone, and every value meant to fall through from
the CLI layer is lost.

## Approach

The exported API does not change: no method, type, constant or error is
added, removed or re-signed.

`mergeRootConfigs` returns the merged `rc`, converted back to the caller's
`PRC`. When the conversion fails, because a `Merge` returned a different
concrete type, or when `Merge` returned nil, loading returns an error wrapping
the existing `ErrFailedToLoadConfig`. Its key-values name the DirType whose
`Merge` produced it, the type that was expected, and the type that came back.

The doc comment on `RootConfig.Merge` states the contract the conversion
relies on: `Merge` returns the merged config, as a value of its receiver's own
concrete type, with the receiver's values taking precedence. The README's
"Understanding Merge()" section says the same.

The single-layer path is unchanged: one loaded layer is returned as loaded,
and `Merge` is not called. Merge order is unchanged: each later DirType's
config receives the result so far as `other`.

## Tests

In `test/`, using test config types:

- two layers: a field set only in the CLI layer and one set only in the project
  layer both reach the result; a field set in both takes the project's value;
- three layers, checking that each `Merge` receives the result so far;
- a `Merge` that copies and returns the copy, and one that edits its receiver
  and returns it, giving the same result;
- a `Merge` that returns another concrete type, and one that returns nil, each
  producing the error above;
- one layer, with `Merge` never called.
