## Approach

- `DirsProvider.CLIConfigDirType` (`config_store.go`): on Linux return `UserConfigDirFunc`'s result unchanged; on other OSes keep `UserHomeDirFunc` plus `DotConfigPathSegment`.
- `cstest/dirs_provider.go`: `getTestCLIConfigDir` follows the same per-OS rule, and `RelConfigDir` derives its path from `getTestCLIConfigDir` instead of hard-coding `Users` and `.config`.
- `examples/comprehensive/main_test.go`: build the expected `CLI:` line from `cfgstore.CLIConfigDir(configSlug, dp)` using the same `DirsProvider` the demo builds, rendered with the demo's `<root>` substitution.
- A test in `test/` covers the Linux and non-Linux CLI dir with a `DirsProvider` whose funcs return temp dirs; the Linux case is checked by running the suite on Linux (CI) since the branch is chosen by `runtime.GOOS`.
