# Generic Type Constraints for Type Safety

**Date:** 2025-12-04

**Status:** Accepted

## Context

Configuration loading functions like `InitProjectConfig()` and `LoadConfigStores()` need to work with user-defined configuration types while enforcing interface requirements. The configuration must implement the `RootConfig` interface for normalization and validation.

### The Problem

We need to:
1. Accept any user-defined configuration struct
2. Ensure it implements the `RootConfig` interface
3. Return a pointer to the configuration
4. Avoid runtime type assertions
5. Provide compile-time type safety

### Option 1: Interface Parameters with Type Assertions

```go
func InitProjectConfig(
    configSlug dt.PathSegment,
    configFile dt.RelFilepath,
    opts Options,
) (RootConfig, error) {
    // Implementation...
}

// Usage requires type assertion
config, err := InitProjectConfig(...)
myConfig := config.(*MyConfig)  // Runtime assertion - can panic!
```

**Advantages:**
- Simple function signature
- Familiar Go 1.x pattern
- No learning curve

**Disadvantages:**
- **Runtime type assertions can panic**
- Caller must know the concrete type
- No compile-time type checking
- Verbose usage code
- Easy to make mistakes

### Option 2: Code Generation

```go
//go:generate go run gen.go

// Generated code:
func InitMyConfig(...) (*MyConfig, error) { ... }
func InitOtherConfig(...) (*OtherConfig, error) { ... }
```

**Advantages:**
- Type-safe without generics
- No runtime type assertions

**Disadvantages:**
- Requires code generation step
- Generated code must be committed or generated in CI
- Harder to debug generated code
- Adds complexity to build process
- Poor IDE support for generated code

### Option 3: Generic Type Constraints (Go 1.18+)

```go
func InitProjectConfig[RC any, PRC RootConfigPtr[RC]](
    configSlug dt.PathSegment,
    configFile dt.RelFilepath,
    opts Options,
) (PRC, error) {
    // Implementation...
}

// Usage is type-safe at compile time
config, err := InitProjectConfig[MyConfig, *MyConfig](...)
// config is *MyConfig, no assertion needed
```

**Advantages:**
- **Compile-time type safety** - no runtime panics
- No type assertions in caller code
- IDE autocomplete works correctly
- Caller gets concrete type directly
- Standard Go 1.18+ feature

**Disadvantages:**
- More verbose function signatures
- Requires Go 1.18+
- Learning curve for developers new to generics
- Type parameters must be explicitly specified

## Decision

Use generic type constraints: `[RC any, PRC RootConfigPtr[RC]]`

The compile-time type safety and improved developer experience outweigh the verbosity cost. Key factors:

1. **No runtime panics**: Type errors caught at compile time
2. **Better IDE support**: Autocomplete knows the concrete type
3. **Clearer intent**: Type parameters make requirements explicit
4. **Standard Go**: Uses language features, no code generation
5. **Future-proof**: Generics are the Go team's recommended approach

The verbose signatures are a one-time cost paid by the library, not by every caller. Callers benefit from type safety without assertions.

## Consequences

### Positive

- **Compile-time type safety**: Invalid types rejected by compiler, not runtime
- **No type assertions**: Callers receive concrete types directly
- **Better IDE experience**: Autocomplete and type hints work correctly
- **Self-documenting**: Type constraints make requirements explicit
- **Refactoring safety**: Renaming types caught by compiler
- **Clear error messages**: Compiler explains type mismatches

### Negative

- **Verbose signatures**: Function declarations are longer
  ```go
  func InitProjectConfig[RC any, PRC RootConfigPtr[RC]](...)
  ```
- **Requires Go 1.18+**: Can't use with older Go versions
- **Learning curve**: Developers unfamiliar with generics need documentation
- **Type parameter noise**: Must specify types at call site
  ```go
  config, err := InitProjectConfig[MyConfig, *MyConfig](...)
  ```

### Migration Path

For developers who prefer simpler APIs, we can provide wrapper functions:

```go
// Simplified wrapper without generics (uses type assertion internally)
func InitProjectConfigSimple(
    cfg RootConfig,
    configSlug dt.PathSegment,
    configFile dt.RelFilepath,
    opts Options,
) error {
    _, err := InitProjectConfig[any, RootConfig](
        configSlug,
        configFile,
        opts,
    )
    return err
}
```

This gives users a choice between type safety (generics) and simplicity (assertions).

## Implementation Notes

### Type Constraint Definition

```go
// RootConfig is the interface all configs must implement
type RootConfig interface {
    RootConfig()
    Normalize(NormalizeArgs) error
    IsNil() bool
}

// RootConfigPtr constrains the pointer type
type RootConfigPtr[RC any] interface {
    RootConfig
    *RC
}
```

### Usage Pattern

```go
// Define your config
type MyConfig struct {
    Username string `json:"username"`
    Theme    string `json:"theme"`
}

func (c *MyConfig) Normalize(args NormalizeArgs) error {
    if c.Theme == "" {
        c.Theme = "light"
    }
    return nil
}

func (c *MyConfig) RootConfig() {}
func (c *MyConfig) IsNil() bool { return c == nil }

// Initialize with type safety
config, err := cfgstore.InitProjectConfig[MyConfig, *MyConfig](
    dt.PathSegment("myapp"),
    dt.RelFilepath("config.json"),
    nil,
)
// config is *MyConfig, not RootConfig
// No type assertion needed
```

### Why Two Type Parameters?

The pattern `[RC any, PRC RootConfigPtr[RC]]` uses two parameters:

- `RC` - The concrete type (e.g., `MyConfig`)
- `PRC` - The pointer to concrete type (e.g., `*MyConfig`)

This is necessary because:
1. Go generics can't express "pointer to T where T implements I"
2. The interface constraint `RootConfigPtr[RC]` enforces the pointer implements `RootConfig`
3. The concrete type `RC` allows creating `new(RC)` internally
4. The return type `PRC` gives callers the concrete pointer type

## Trade-off Analysis

### Verbosity vs. Safety

**Cost:**
- Function signatures: ~40 extra characters
- Call sites: ~20 extra characters for type parameters

**Benefit:**
- Zero runtime type assertion panics
- Compiler catches type errors
- IDE autocomplete works perfectly

**Verdict:** The safety benefits outweigh the verbosity cost. Type errors caught at compile time prevent production bugs.

### Learning Curve vs. Correctness

**Cost:**
- Developers need to understand generic type constraints
- Documentation must explain the pattern
- Initial confusion for Go 1.x developers

**Benefit:**
- Once learned, pattern is consistent everywhere
- Compiler guides developers to correct usage
- Impossible to use incorrectly

**Verdict:** One-time learning cost pays ongoing dividends in correctness.

## Related Decisions

- ADR 3 (doterr pattern): Both patterns prioritize correctness over simplicity
- The RootConfig interface design: Generic constraints require interface methods
- Future API streamlining: May provide simplified wrappers for common cases

## Future Considerations

If type inference improves in future Go versions, we may be able to omit type parameters:

```go
// Future Go (hypothetical):
config, err := cfgstore.InitProjectConfig(...)
// Compiler infers [MyConfig, *MyConfig] from context
```

Until then, explicit type parameters are necessary but acceptable.

## References

- Go Generics Proposal: https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md
- Type constraints tutorial: https://go.dev/doc/tutorial/generics
- Pattern documentation: See README.md "Architecture Decisions" section
