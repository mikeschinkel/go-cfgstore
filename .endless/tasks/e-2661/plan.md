## Approach

### `RootConfig` and its constraint (`root_config.go`, `config_stores.go`)

- `RootConfig` keeps `RootConfig()` and `Normalize(NormalizeArgs) error`, drops `Merge(RootConfig) RootConfig`, and gains `ApplyDefaults(DefaultsArgs) error`.
- `RootConfigPtr[RC]` adds `Merge(MergeArgs[*RC]) error`, so `Merge` is required and typed. `CLIRootConfigPtr[RC]` embeds `RootConfigPtr[RC]` plus `cliutil.Config`.
- `DefaultsArgs` carries `Options` and the `DirTypes` being loaded.
- `MergeArgs[PRC]` carries the layer's `DirType` and `SourceFile`, and has:
  - `Decode(c PRC) error`: decodes the layer's file over `c` with `encoding/json/v2`;
  - `Layer() (PRC, error)`: decodes the layer's file alone into a new value, for a custom merge.

### Loading (`LoadConfigStores`, `mergeRootConfigs`, `mergedRootConfig`, `loadConfigIfExists`)

`LoadConfigStores` creates one value, calls `ApplyDefaults` on it, then for each DirType in order: ensures the CLI file exists (creating it as below), skips a project layer whose file does not exist, calls `Merge` with that layer's `MergeArgs`, then calls `Normalize` with that layer's `NormalizeArgs`. It stops at the first layer whose `Merge` or `Normalize` fails, wrapping the error with that layer's file path. `mergeRootConfigs` and `mergedRootConfig` go away, along with the `dtx.IsZero` test for an absent project layer; with no layer to load it returns `ErrNotValidConfigDirsAvailable` as today.

### Creating a config file (`createConfig`, `initConfig`, `InitProjectConfig`)

- `createConfig` for the CLI layer: new value, `ApplyDefaults`, `Normalize`, `SaveJSON`.
- `InitProjectConfig` saves a file without defaults, then calls `ApplyDefaults` on the value it returns.

### Field types (new `merge_types.go`)

- `AppendSlice[T]`: a later layer's elements are appended.
- `ReplaceMap[K,V]`: a later layer's map replaces the earlier one instead of merging.

Each implements `UnmarshalJSONFrom` and marshals as its underlying type.

### Doc comments

- `Normalize`: runs after each layer on the value built from every layer so far; must be idempotent; must not set defaults.
- `ApplyDefaults`: runs on an empty value before any layer decodes; set defaults here.
- `Merge`: the common body is `return args.Decode(c)`, and `Layer` is for rules the decoder cannot express.
- The failed-load wrap in `LoadConfigStores` gets a comment naming E-2658.

### Tests (`test/`)

`merge_test.go` drops the tests for `Merge` returning nil or another type and for in-place versus copying merges, which no longer apply, and covers the items below. The `cliConfig` in `cliutil_config_test.go` moves to the new methods.

### Examples and README

`examples/basic`, `examples/comprehensive` and `examples/cliutil` move to `ApplyDefaults` and the one-line `Merge`. `examples/comprehensive` sets its theme default in `ApplyDefaults` and shows the CLI layer's theme surviving a project layer that omits it, replacing its comment about avoiding defaults in `Normalize`. The README describes layering, `ApplyDefaults`, `Merge` and the field types.

### Task notes

E-2658 and E-2659 each get a note that a failed load is wrapped in `ErrFailedToEnsureConfig` in `LoadConfigStores`, with the comment there naming E-2658.
