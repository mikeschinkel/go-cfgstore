## Approach

- `ProjectConfigStoreExists` (`config_store.go`) builds its store with `NewConfigStore(ProjectConfigDirType, ...)` and a `DirsProvider` copied from `DefaultDirsProvider()` with `ProjectDirFunc` returning `projectDir`, instead of calling `SetConfigDir`. `ConfigDir` then resolves `<projectDir>/.<slug>`.
- Its doc comment says it takes the project directory and reports whether `<projectDir>/.<slug>/<file>` exists.
- `examples/comprehensive`'s `projectExists` step passes the project directory, and its comment about passing the config directory goes away. `main_test.go` keeps expecting `exists: true` and `exists elsewhere: false`.
- A test in `test/` covers the items below with a temp directory.
