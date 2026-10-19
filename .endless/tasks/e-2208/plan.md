# Return the merged config from mergeRootConfigs

## The defect

`mergeRootConfigs` (config_stores.go) folds each layer into `rc` with
`rc = rcMap[typ].Merge(rc)`, then returns `prc = rcMap[dirType]`: the last
layer as loaded, not the merged result. Whenever a project config exists, the
caller gets the project layer alone, and every value meant to fall through from
the CLI layer is lost.

## Approach

Replace `Merge` on the `RootConfig` interface with `MergeFrom` on the generic
constraint `RootConfigPtr[RC]`, typed in the concrete config:

    type RootConfigPtr[RC any] interface {
        RootConfig
        *RC
        MergeFrom(other *RC) *RC
    }

`RootConfig` keeps `RootConfig()` and `Normalize(...)`, the methods the
non-generic store code (`ensureConfig`, `loadConfigIfExists` and friends) uses.
`mergeRootConfigs` then carries the merge in `PRC` from start to finish and
returns it, so no runtime type conversion exists and a `MergeFrom` with the
wrong types fails to compile. A `nil` from `MergeFrom` is the one runtime failure left;
loading returns an error naming the layer.

The single-layer path is unchanged: one loaded layer is returned as loaded,
and `MergeFrom` is not called. Merge order is unchanged: each later DirType's
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

## Order of landing

The method gets a new name so Endless can adopt it before this lands. A type
cannot carry two methods named `Merge` with different signatures, but an extra
`MergeFrom` is harmless against today's go-cfgstore. So Endless's
`EndlessConfig` gains `MergeFrom` in a task of its own that lands first, and
this task lands after it. Endless builds against this checkout through a
`replace`, so it never sees a go-cfgstore it cannot compile against.
