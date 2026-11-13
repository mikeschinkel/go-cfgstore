One implementation, OS-faithful, easy to call: cstest.NewTestDirsProvider is the test provider and the only one.

**Make the call one argument wide.** TestDirsProviderArgs keeps its fields, but NewTestDirsProvider fills defaults when a field is empty: Username defaults to `testuser`, ProjectDir to `project` under the fake home. A test that wants only isolation writes `cstest.NewTestDirsProvider(&cstest.TestDirsProviderArgs{TestRoot: root})`.

**Tests ask the provider for paths.** The layout under the root follows the running OS, so a test never computes `<root>/Users/testuser/...` itself; it calls the provider's UserConfigDirFunc, ProjectDirFunc or CLIConfigDirFunc and joins from there. The README example shows that, and it is why the OS-faithful layout costs a test nothing.

**Finish the two methods.** DirsProvider.WithProjectDir(dir) and DirsProvider.WithUserConfigDir(dir) in dirs_provider.go return a copy of the receiver with that one function replaced by one returning dir. Their commented-out sketches go. gomion's per-directory stores, which today build a provider with only ProjectDirFunc set, are the first callers.

**Tests** go in the existing test module (test/), beside dirs_provider_test.go: defaults applied, every function under the root on this OS, With methods copy rather than mutate.
