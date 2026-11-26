## Causes

- `DirsProvider.CLIConfigDirType` starts from `UserConfigDirFunc` on Linux and `UserHomeDirFunc` elsewhere, then appends `DotConfigPathSegment` on every OS. On Linux the user config dir is already `$XDG_CONFIG_HOME` or `$HOME/.config`, so the CLI config lands in `.config/.config/<slug>`.
- `examples/comprehensive`'s `TestRun` (added in E-2650) expects the macOS layout for the CLI line. E-2650's verify ran only on macOS.
- `cstest`'s `TestDirsProviderArgs.RelConfigDir` hard-codes a `Users` segment, while `getTestUserHomeDir` uses `/home` on Linux, so the two disagree there.

## Old Linux path

Mike: there are no Linux installs to worry about, so configs at the old `.config/.config/<slug>` path are neither read nor moved.

## Test expectation

The example test asks `cfgstore.CLIConfigDir` with the demo's `DirsProvider` for the expected directory, so it follows each OS's rule instead of encoding one.
