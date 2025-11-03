# Use DirsProvider Instead of Build Tags for Test Directory Injection

**Date:** 2025-11-02

**Status:** Accepted

## Context

`go-cfgstore` needs to support testing by allowing tests to redirect configuration directories to temporary test locations instead of actual user directories (`~/.config`, `~/Library/Application Support`, etc.). This prevents tests from:
- Polluting real user configuration directories
- Causing flaky tests due to existing config files
- Requiring cleanup of real user directories after test runs

Two approaches were considered:

### Option 1: Build Tags
Use Go build tags to compile different implementations for production vs. test code:

```go
// +build !test

package cfgstore

func getUserHomeDir() (dt.DirPath, error) {
    return dt.UserHomeDir()
}
```

```go
// +build test

package cfgstore

var testHomeDir dt.DirPath

func getUserHomeDir() (dt.DirPath, error) {
    return testHomeDir, nil
}
```

**Advantages:**
- Clean separation between production and test code
- No runtime overhead in production builds
- No test-specific fields in production structs

**Disadvantages:**
- **Forces all test runners to use `-tags test` flag**
- Breaks `go test ./...` without additional flags
- Breaks IDE test runners that don't pass build tags
- Breaks CI/CD pipelines that don't specify build tags
- Poor discoverability - developers may not know they need the tag
- Breaking changes if we add new build tags later

### Option 2: DirsProvider Dependency Injection
Add a `DirsProvider` struct to `ConfigStore` that holds directory functions:

```go
type DirsProvider struct {
    UserHomeDirFunc   DirFunc
    GetwdFunc         DirFunc
    ProjectDirFunc    DirFunc
    UserConfigDirFunc DirFunc
}
```

Production code uses default implementations, test code injects custom implementations.

**Advantages:**
- Tests work with standard `go test ./...` command
- Works with all IDE test runners out of the box
- Explicit dependency injection makes testing clear
- No special flags or setup required
- Backward compatible evolution path

**Disadvantages:**
- Small runtime overhead (struct field, nil check)
- Test-specific field exists in production builds
- Slightly larger struct size

## Decision

Use `DirsProvider` dependency injection instead of build tags.

The ergonomic cost of requiring build tags outweighs the theoretical purity benefits. Most developers expect `go test ./...` to work without additional configuration. Build tags create friction:
- New contributors won't know they need `-tags test`
- CI/CD pipelines need special configuration
- IDE integrations break or require custom setup
- Documentation burden increases

The `DirsProvider` approach trades a small amount of memory (a few pointers in a struct) for significantly better developer experience.

## Consequences

### Positive
- Tests run with standard Go commands: `go test ./...`
- IDE test runners work without configuration
- CI/CD pipelines work without special build tag setup
- Test utilities in `cstest` package are straightforward to use
- Clear dependency injection pattern is self-documenting

### Negative
- Production `ConfigStore` instances carry a `DirsProvider` field that's only used in tests
- Small runtime overhead for nil check and indirection when calling directory functions
- Struct is slightly larger than minimal production-only implementation

### Future Migration Path

If we later decide build tags are worthwhile, we can migrate in a backward-compatible way:

1. Add build-tag-based implementations alongside `DirsProvider`
2. Deprecate but don't remove `DirsProvider` field
3. Make `DirsProvider` a no-op when build tags are active
4. Document migration path for users who want zero-overhead production builds
5. Eventually remove `DirsProvider` in a major version bump

This gives users who care about the overhead an opt-in path while maintaining compatibility for users who prefer convenience.

## Implementation Notes

The `cstest.NewTestDirsProvider()` function creates test-specific directory providers with test root directory support, isolating tests from the real filesystem by staging a filesystem tree in a temporary directory:

```go
dp := cstest.NewTestDirsProvider(&cstest.TestDirsProviderArgs{
    Username:   "testuser",
    TestRoot:   tempDir,
    ConfigSlug: "myapp",
})

cs := cfgstore.NewConfigStore(cfgstore.CLIConfigDir, cfgstore.ConfigStoreArgs{
    ConfigSlug:   dt.PathSegment("myapp"),
    RelFilepath:  dt.RelFilepath("config.json"),
    DirsProvider: dp,
})
```

Production code omits `DirsProvider` and gets default implementations automatically.
