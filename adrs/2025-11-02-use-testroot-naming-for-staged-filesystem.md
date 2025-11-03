# Use TestRoot Naming for Staged Filesystem Directory

**Date:** 2025-11-02

**Status:** Accepted

## Context

The `cstest` package needs a clear, concise name for the temporary directory that serves as the base for staging a filesystem tree during tests. This directory acts as the root of an isolated filesystem hierarchy where tests can create and manipulate configuration files without affecting the real user filesystem.

### The Concept: Staging Directory

When testing filesystem operations, it's common to:
1. Create a temporary directory
2. Build a directory structure within it that mirrors the real filesystem layout
3. Run tests against this staged tree
4. Clean up the temporary structure

This pattern appears in:
- **ISO image building** - Tools like `mkisofs` and `genisoimage` stage files in a directory before packaging them into an ISO
- **Package building** - RPM uses `buildroot`, Debian uses staging directories to construct package contents before creating `.deb`/`.rpm` files
- **Container image building** - Docker stages filesystem layers before committing them
- **Installation systems** - Installers stage files before copying to final destinations

The staged directory serves as a "root" from which all paths are relative, isolating the staging area from the real filesystem.

## Alternatives Considered

### Option 1: `chrootDir` / `ChrootDir`
**Reasoning:** References the Unix `chroot()` system call that changes the root directory for a process.

**Rejected because:**
- **Technically incorrect** - The code doesn't use `chroot()` or create OS-level isolation
- **Implies security boundary** - `chroot` is a security/isolation mechanism; this is just path manipulation
- **Platform-specific** - `chroot()` is Unix-specific; this code works on Windows too
- **Misleading** - Developers familiar with `chroot` jails might expect kernel-level isolation

### Option 2: `baseDir` / `BaseDir`
**Reasoning:** Generic term for the base/root of a path hierarchy.

**Rejected because:**
- **Too generic** - "base directory" could mean many things (working directory, install directory, etc.)
- **No semantic connection to testing** - Doesn't convey this is test-specific
- **Lacks precedent** - Not a recognized term in build/test tooling

### Option 3: `stageDir` / `StageDir`
**Reasoning:** Describes the action (staging files) rather than the structure.

**Rejected because:**
- **Verb-based** - "stage" is an action, not a structural concept
- **Less clear** - "Staging what? For what purpose?"
- **Uncommon in test code** - More common in build/deploy contexts
- **Missing the "root" concept** - Doesn't convey this is the base of a hierarchy

### Option 4: `testRoot` / `TestRoot` ✓
**Reasoning:** Parallels `buildRoot` from package management, clearly indicates test-specific root directory.

**Accepted because:**
- **Established precedent** - RPM's `buildroot` is widely recognized in the build/packaging domain
- **Semantically clear** - "test root" = root directory for tests
- **Concise** - Short, memorable, easy to type
- **Self-documenting** - Name clearly indicates purpose without needing comments
- **Naming consistency** - `TestRoot` (struct field), `testRoot` (local variable), `TestRootFunc` (function)

## Decision

Use `testRoot` / `TestRoot` naming throughout the codebase.

- **Struct fields:** `TestRoot`, `TestRootFunc`
- **Local variables:** `testRoot`
- **Methods:** `GetTestRoot()`, `OmitTestRoot()`, `WithoutTestRoot()`
- **Private fields:** `omitTestRoot`

In documentation and comments, refer to it as:
- **"test root directory"** - Full formal name
- **"staging directory"** - When explaining the concept
- **"staged filesystem"** - When describing what's inside it

## Rationale

The term "test root" immediately communicates:
1. **Purpose** - This is for testing (not production)
2. **Structure** - It's a root directory (base of a hierarchy)
3. **Scope** - It's isolated and temporary

Developers familiar with build systems will recognize the parallel to `buildroot` and immediately understand the pattern:
- `buildroot` = temporary root directory for building packages
- `testRoot` = temporary root directory for testing filesystem operations

This naming works across all platforms (macOS, Linux, Windows) and doesn't carry incorrect implications about OS-level isolation or security boundaries.

## Examples

### Creating a Test Root and Config Store

```go
// Create a temporary directory as the test root
testRoot := dtx.TempTestDir(t)
defer testRoot.RemoveAll()

// Configure the test directory provider
args := &cstest.TestDirsProviderArgs{
    Username:   "testuser",
    ProjectDir: "myproject",
    ConfigSlug: "myapp",
    TestRoot:   testRoot,
}

// Create config store with test directory provider
store := cfgstore.NewConfigStore(cfgstore.CLIConfigDir, cfgstore.ConfigStoreArgs{
    ConfigSlug:   "myapp",
    RelFilepath:  "config.json",
    DirsProvider: cstest.NewTestDirsProvider(args),
})

// Now you can use the store in tests
// Config files will be staged under testRoot instead of real user directories
```

### Staged Filesystem Structure

Given a `testRoot` of `/tmp/test-xyz123`, the staged filesystem might look like:

```
/tmp/test-xyz123/                    # testRoot
├── Users/
│   └── testuser/
│       ├── .config/
│       │   └── myapp/
│       │       └── config.json      # CLI config
│       └── Library/
│           └── Application Support/
│               └── myapp/
│                   └── settings.json # App config
└── Users/
    └── testuser/
        └── Projects/
            └── myproject/
                └── .myapp/
                    └── config.json   # Project config
```

All paths are staged under `testRoot`, isolating tests from real user directories.

## Consequences

### Positive
- Clear, self-documenting code
- Familiar pattern for developers who've worked with build systems
- No misleading implications about system-level features
- Platform-agnostic terminology
- Consistent naming across variable scopes

### Negative
- Developers unfamiliar with `buildroot` might not immediately recognize the pattern
- Slightly longer than `baseDir` or `rootDir`

### Mitigation
Documentation clearly explains the concept and provides examples. The ADR itself serves as reference material for understanding the rationale.

## See Also
- ADR: Use DirsProvider Instead of Build Tags for Test Directory Injection
- Package documentation: `cstest.TestDirsProviderArgs`
