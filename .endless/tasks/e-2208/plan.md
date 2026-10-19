# Return the merged config from mergeRootConfigs

## The defect

`mergeRootConfigs` (config_stores.go) folds each layer into `rc` with
`rc = rcMap[typ].Merge(rc)`, then returns `prc = rcMap[dirType]`: the last
layer as loaded, not the merged result. Whenever a project config exists, the
caller gets the project layer alone, and every value meant to fall through from
the CLI layer is lost.

## Approach

Move `Merge` off the `RootConfig` interface and onto the generic constraint
`RootConfigPtr[RC]`, typed in the concrete config:

    type RootConfigPtr[RC any] interface {
        RootConfig
        *RC
        Merge(other *RC) *RC
    }

`RootConfig` keeps `RootConfig()` and `Normalize(...)`, the methods the
non-generic store code (`ensureConfig`, `loadConfigIfExists` and friends) uses.
`mergeRootConfigs` then carries the merge in `PRC` from start to finish and
returns it, so no runtime type conversion exists and a `Merge` with the wrong
types fails to compile. A `nil` from `Merge` is the one runtime failure left;
loading returns an error naming the layer.

The single-layer path is unchanged: one loaded layer is returned as loaded,
and `Merge` is not called. Merge order is unchanged: each later DirType's
config receives the result so far as `other`.

## Tests

In `test/`, using a test config type:

- two layers: a field set only in the CLI layer and one set only in the project
  layer both reach the result; a field set in both takes the project's value;
- three layers, checking that each `Merge` receives the result so far;
- a `Merge` that copies and returns the copy, and one that edits its receiver
  and returns it, giving the same result;
- a `Merge` that returns nil, producing an error;
- one layer, with `Merge` never called.
