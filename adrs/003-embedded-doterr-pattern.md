# Embedded doterr Error Handling Pattern

**Date:** 2025-12-04

**Status:** Accepted

## Context

Go error handling traditionally relies on `fmt.Errorf` with format strings (`%w`, `%s`, `%d`) for wrapping and annotating errors. This approach has several limitations:

1. **Brittle format strings**: Easy to mismatch format specifiers with arguments
2. **Limited metadata**: Difficult to attach structured data to errors
3. **Single-layer sentinels**: Can only check one error condition with `errors.Is()`
4. **Debugging challenges**: Error messages are strings, not structured data
5. **No composability**: Can't easily check "any database error" vs. "database not found error"

### Option 1: Standard library `fmt.Errorf`

```go
err = fmt.Errorf("failed to read config file %s: %w", filepath, cause)
```

**Advantages:**
- Standard library, no dependencies
- Familiar to all Go developers
- Zero learning curve

**Disadvantages:**
- Brittle format strings prone to typos
- Can't extract metadata programmatically
- Single-layer error checking only
- No composable sentinels
- String-based debugging

### Option 2: Third-party error packages

Packages like `pkg/errors`, `go-multierror`, or `cockroachdb/errors` offer richer error handling.

**Advantages:**
- More features than stdlib
- Some support for metadata
- Active maintenance

**Disadvantages:**
- External dependencies that may break
- Varying APIs across packages
- May not fit our exact needs
- Learning curve for contributors

### Option 3: Embedded doterr Package

Embed a ~700-line `doterr.go` file in each package that provides:
- Composable sentinel errors (layer + category)
- Key-value metadata instead of format strings
- Structured error building with `NewErr()` and `WithErr()`
- Multi-layer error checking

```go
err = NewErr(
    ErrFailedToReadFile,
    "filepath", configPath,
    cause,
)
```

**Advantages:**
- Structured metadata extraction for debugging
- Composable sentinels: `errors.Is(err, ErrRepo)`, `errors.Is(err, ErrDatabase)`, `errors.Is(err, ErrNotFound)`
- Type-safe error construction (no format strings)
- Consistent pattern across all packages
- No external dependencies that can break
- Tailored exactly to our needs

**Disadvantages:**
- ~700 lines duplicated per package
- Custom pattern requires documentation
- Learning curve for contributors unfamiliar with pattern

## Decision

Embed `doterr.go` in each package that needs structured error handling.

The benefits of structured errors with composable sentinels outweigh the cost of code duplication. Key factors:

1. **Zero external dependencies**: No risk of upstream breaking changes
2. **Consistency**: Same error pattern across all packages in the ecosystem
3. **Debuggability**: Structured metadata enables better error inspection and logging
4. **Composability**: Multi-layer sentinels enable precise error handling
5. **Type safety**: Compile-time checking instead of runtime format string errors

The ~700 lines of duplicated code is acceptable because:
- Modern storage is cheap
- Consistency across packages is valuable
- No dependency management overhead
- Package-specific modifications are possible if needed

## Consequences

### Positive

- **Structured error metadata**: Errors carry key-value pairs that can be extracted programmatically
- **Composable sentinels**: Check errors at multiple levels (e.g., "any repository error", "any database error", "any not-found error")
- **Type safety**: No format string mismatches at runtime
- **Better debugging**: Errors provide structured context for logging and error tracking
- **No dependency risk**: No external packages that could break or change APIs
- **Consistent patterns**: All packages in ecosystem use same error approach
- **Works with `goto end` pattern**: Enriching errors at function exit point with `WithErr()`

### Negative

- **Code duplication**: ~700 lines of doterr.go in each package
- **Learning curve**: Contributors need to learn NewErr/WithErr pattern instead of fmt.Errorf
- **Documentation burden**: Must document the pattern for external contributors
- **Not idiomatic**: Differs from standard Go error handling (though still uses errors.Is/As)

### Migration Path

Packages can migrate from stdlib errors to doterr incrementally:

1. Copy `doterr.go` into package
2. Define sentinel errors in `errors.go`
3. Gradually replace `fmt.Errorf` with `NewErr` and `WithErr`
4. Add metadata to errors for better debugging
5. Use layered sentinels for composable error checking

## Implementation Notes

### Error Construction Pattern

```go
// errors.go - Define sentinels
var (
    ErrCmd  = errors.New("command")           // layer sentinel
    ErrInit = errors.New("init")              // category sentinel
    ErrFilesAlreadyExist = errors.New("files already exist")
)

// code - Use NewErr with sentinels + metadata + cause
err = NewErr(ErrCmd, ErrInit, ErrFilesAlreadyExist,
    "files", conflicts,
    cause,  // trailing cause
)
```

### Error Enrichment Pattern

```go
func processFile() (err error) {
    var filepath dt.Filepath

    filepath = dt.FilepathJoin(c.Dir, "config.json")
    data, err := filepath.ReadFile()
    if err != nil {
        err = NewErr(ErrFileRead, err)
        goto end
    }

    // ... more operations

end:
    if err != nil {
        // Add function-level context once at exit
        err = WithErr(err, "filepath", filepath)
    }
    return err
}
```

### Multi-Layer Error Checking

```go
// Check specific error
if errors.Is(err, ErrFilesAlreadyExist) {
    // Handle files exist case
}

// Check any init error
if errors.Is(err, ErrInit) {
    // Handle any init error
}

// Check any command error
if errors.Is(err, ErrCmd) {
    // Handle any command error
}
```

## Related Decisions

- The `goto end` pattern (used throughout codebase) works well with `WithErr()` for adding context at function exit
- Sentinel errors should be checked in `go-dt/errors.go` before creating new ones to maximize reuse
- Generic sentinels (ErrNotFound, ErrInvalidInput) are preferred over specific ones (ErrConfigFileNotFound)

## References

- Source: `/Users/mikeschinkel/Projects/go-pkgs/go-doterr/doterr.go`
- Error handling guidance: `~/.claude/CLAUDE-golang-error-handling.md` (private)
- Pattern documentation: See README.md "Architecture Decisions" section
