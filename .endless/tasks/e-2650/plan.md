# Plan: go-cliutil ConfigFunc adapter in go-cfgstore

Binding decisions: ED-1670 in go-cliutil (go-cliutil imports no config library;
config reaches flags through a typed ApplyConfig hook, FlagValuer for name-based
configs; a typed flag wins). The go-cliutil side (ConfigFunc, ConfigArgs,
ConfigApplier, ConfigAs) is built by go-cliutil E-2648, which this task waits on.

## Approach

### Adapter
- New `cliutil_config.go`:
  `CLIConfigFunc[RC any, PRC CLIRootConfigPtr[RC]](args CLIConfigArgs) cliutil.ConfigFunc`.
  `CLIRootConfigPtr[RC]` requires `RootConfig`, `cliutil.Config` and `*RC`.
- The returned func reads the config slug and file from `ConfigArgs.AppInfo`
  (`ConfigSlug()`, `ConfigFile()`), calls `LoadConfig[RC, PRC]` with the args'
  `DirTypes` and `DirsProvider` (same defaults as `LoadConfig`: CLI then project,
  project winning) and nil Options, and returns the PRC as the `cliutil.Config`.
- `CLIConfigArgs{DirTypes []DirType, DirsProvider *DirsProvider}`, both optional.
- No tags, no reflection, no name-based lookup: apps map config to flags with
  `ApplyConfig` on their Options and commands, reading the RC's typed methods.

### Tests (test/ module)
- A temp `DirsProvider` isolates config. Cases: CLI layer only; a project layer
  that wins over the CLI layer; slug and file taken from AppInfo; a load error
  returned from the ConfigFunc; through `cliutiltest`, a config value applied by
  `ApplyConfig` and a typed flag beating it.

### Examples
- `examples/basic_usage` is renamed `examples/basic`.
- New `examples/comprehensive` shows every go-cfgstore feature: CLI, project and
  app config directories; layered loading and `Merge`; `Normalize`; `Save` and
  `SaveJSON`; `InitProjectConfig`; `ProjectConfigStoreExists`; cache directories;
  a custom `DirsProvider`; the `cstest` helpers; `CreateWriterLogger`.
- New `examples/cliutil`: a go-cliutil app wired to go-cfgstore with
  `CLIConfigFunc`, an RC with typed methods, `ApplyConfig` on its Options and on a
  command, and a `cliutiltest` table showing config applied and a typed flag
  winning.
- Each example is its own module with a `replace` of `../..` (allowed for
  examples).

### Docs and dependencies
- README gains a go-cliutil section showing the one-line wiring and ApplyConfig.
- `go.mod` requires the go-cliutil version released from E-2648 (no `replace`);
  local work uses `go.work`. `CreateWriterLogger` is updated only as far as the
  go-cliutil API change forces it to compile.

### Review
- Mike reviews the code and all three examples in the task worktree before the
  task leaves `unverified`.
