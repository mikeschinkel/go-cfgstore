## Cause

`ProjectConfigStoreExists` calls `SetConfigDir(projectDir)`, which sets the store's config directory directly and so skips `ConfigDir`, the one place that appends `.<slug>` for `ProjectConfigDirType`. The function therefore checks `<projectDir>/<file>`.

## Callers

A search of the endless, gomion and go-pkgs checkouts finds no caller in Endless or Gomion, the consumers Mike said matter. The only caller is `examples/comprehensive`, which passes `<project>/.<slug>` to get a correct answer and says so in a comment. Changing the behavior to match the parameter name breaks no consumer that matters, so the fix changes behavior rather than renaming the parameter to document the current behavior. Renaming would leave every caller writing the `.<slug>` rule a second time.

## Approach choice

Giving the store a `DirsProvider` whose `ProjectDirFunc` returns `projectDir` lets `ConfigDir` apply `.<slug>`, so the rule stays in one place. Joining `"." + slug` inside `ProjectConfigStoreExists` would copy it.
