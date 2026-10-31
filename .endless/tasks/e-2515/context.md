# go-cfgstore Enhancement Plan

## Overview

Extend go-cfgstore to support three directory layout conventions:

1. **XDG User Directories** — `~/.config/`, `~/.local/share/`, etc. (for CLI tools)
2. **FHS System Directories** — `/etc/`, `/var/lib/`, `/var/log/` (for daemons)
3. **Prefix-based/Relocatable** — `/opt/<app>/etc/`, `/opt/<app>/var/` (for self-contained installs)

All three are implemented via `DirsProvider` — no new concepts, just new function fields and implementations.

---

## Part 1: XDG Base Directory Support

### Background

The [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) defines standard locations for user files.

### Current State

go-cfgstore currently supports:
- `CLIConfigDirType` → `~/.config/<slug>/`
- `AppConfigDirType` → `os.UserConfigDir()` (platform-specific)
- `ProjectConfigDirType` → `<projectDir>/.<slug>/`
- `CustomConfigDirType` → user-specified path

### XDG Directories to Add

| Environment Variable | Default | Purpose |
|---------------------|---------|---------|
| `XDG_CONFIG_HOME` | `~/.config` | User-specific configuration (already supported) |
| `XDG_DATA_HOME` | `~/.local/share` | User-specific data files |
| `XDG_STATE_HOME` | `~/.local/state` | User-specific state (logs, history) |
| `XDG_CACHE_HOME` | `~/.cache` | Non-essential cached data |
| `XDG_RUNTIME_DIR` | `/run/user/$UID` | Runtime files (sockets, locks) |

### New DirTypes

```go
const (
    // Existing
    UnspecifiedConfigDirType DirType = iota
    AppConfigDirType                 // os.UserConfigDir()
    CLIConfigDirType                 // ~/.config/<slug>
    ProjectConfigDirType             // <projectDir>/.<slug>
    CustomConfigDirType              // <customDir>

    // New XDG types
    XDGDataDirType                   // ~/.local/share/<slug>
    XDGStateDirType                  // ~/.local/state/<slug>
    XDGCacheDirType                  // ~/.cache/<slug>
    XDGRuntimeDirType                // /run/user/$UID/<slug>
)
```

### New DirsProvider Fields (XDG)

```go
type DirsProvider struct {
    // Existing
    UserHomeDirFunc     DirFunc
    GetwdFunc           DirFunc
    ProjectDirFunc      DirFunc
    UserConfigDirFunc   DirFunc
    CustomConfigDirFunc DirFunc
    CLIConfigDirFunc    DirFunc
    UserCacheDirFunc    DirFunc

    // New XDG functions
    XDGDataDirFunc      DirFunc  // Returns XDG_DATA_HOME or ~/.local/share
    XDGStateDirFunc     DirFunc  // Returns XDG_STATE_HOME or ~/.local/state
    XDGRuntimeDirFunc   DirFunc  // Returns XDG_RUNTIME_DIR or error
}
```

### XDG Helper Functions

```go
// xdg.go

func XDGDataDir(slug dt.PathSegment, dps ...*DirsProvider) (dt.DirPath, error)
func XDGStateDir(slug dt.PathSegment, dps ...*DirsProvider) (dt.DirPath, error)
func XDGCacheDir(slug dt.PathSegment, dps ...*DirsProvider) (dt.DirPath, error)
func XDGRuntimeDir(slug dt.PathSegment, dps ...*DirsProvider) (dt.DirPath, error)
```

---

## Part 2: FHS System Directory Support

### Background

The [Filesystem Hierarchy Standard (FHS)](https://refspecs.linuxfoundation.org/FHS_3.0/fhs/index.html) defines standard locations for system files on Linux.

### Use Case

Daemons and system services need different paths than user CLI tools:

| Purpose | XDG (User) | FHS (System) |
|---------|------------|--------------|
| Config | `~/.config/<app>/` | `/etc/<app>/` |
| Data | `~/.local/share/<app>/` | `/var/lib/<app>/` |
| Logs | `~/.local/state/<app>/logs/` | `/var/log/<app>/` |
| Runtime | `/run/user/$UID/<app>/` | `/run/<app>/` |

### New DirTypes

```go
const (
    // ... existing ...

    // FHS System types
    SystemConfigDirType     // /etc/<slug>
    SystemDataDirType       // /var/lib/<slug>
    SystemLogDirType        // /var/log/<slug>
    SystemRuntimeDirType    // /run/<slug>
)
```

### New DirsProvider Fields (FHS)

```go
type DirsProvider struct {
    // ... existing ...

    // FHS System functions
    SystemConfigDirFunc   DirFunc  // Returns /etc
    SystemDataDirFunc     DirFunc  // Returns /var/lib
    SystemLogDirFunc      DirFunc  // Returns /var/log
    SystemRuntimeDirFunc  DirFunc  // Returns /run
}
```

### FHS Helper Functions

```go
// fhs.go

func SystemConfigDir(slug dt.PathSegment, dps ...*DirsProvider) (dt.DirPath, error)
func SystemDataDir(slug dt.PathSegment, dps ...*DirsProvider) (dt.DirPath, error)
func SystemLogDir(slug dt.PathSegment, dps ...*DirsProvider) (dt.DirPath, error)
func SystemRuntimeDir(slug dt.PathSegment, dps ...*DirsProvider) (dt.DirPath, error)
```

### Default Implementations

```go
func defaultSystemConfigDir() (dt.DirPath, error) {
    return "/etc", nil
}

func defaultSystemDataDir() (dt.DirPath, error) {
    return "/var/lib", nil
}

func defaultSystemLogDir() (dt.DirPath, error) {
    return "/var/log", nil
}

func defaultSystemRuntimeDir() (dt.DirPath, error) {
    return "/run", nil
}
```

---

## Part 3: Prefix-based/Relocatable Support

### Background

Self-contained installs (e.g., `/opt/<app>/`) keep all files under one prefix for easy deployment and removal.

### Use Case

Platforms like Synology NAS where you can't easily use standard FHS paths, or want portable installations.

### Layout

```
/opt/doctool/
├── bin/doctoold
├── etc/doctool.json      # Config
├── var/
│   ├── doctool.db        # Data
│   └── log/              # Logs
└── run/                  # Runtime (if needed)
```

### Implementation

No new DirTypes needed — use a custom `DirsProvider`:

```go
// NewPrefixDirsProvider creates a DirsProvider for self-contained installs.
// All paths are relative to the given prefix.
func NewPrefixDirsProvider(prefix dt.DirPath) *DirsProvider {
    return &DirsProvider{
        UserHomeDirFunc: func() (dt.DirPath, error) {
            return prefix, nil
        },
        SystemConfigDirFunc: func() (dt.DirPath, error) {
            return dt.DirPathJoin(prefix, "etc"), nil
        },
        SystemDataDirFunc: func() (dt.DirPath, error) {
            return dt.DirPathJoin(prefix, "var"), nil
        },
        SystemLogDirFunc: func() (dt.DirPath, error) {
            return dt.DirPathJoin(prefix, "var", "log"), nil
        },
        SystemRuntimeDirFunc: func() (dt.DirPath, error) {
            return dt.DirPathJoin(prefix, "run"), nil
        },
        // XDG functions also redirect to prefix if needed
        CLIConfigDirFunc: func() (dt.DirPath, error) {
            return dt.DirPathJoin(prefix, "etc"), nil
        },
        XDGDataDirFunc: func() (dt.DirPath, error) {
            return dt.DirPathJoin(prefix, "var"), nil
        },
        // ... etc
    }
}
```

---

## Example Usage

### User CLI Tool (XDG)

```go
const appSlug = dt.PathSegment("doctool")

// Uses ~/.config/doctool/, ~/.local/share/doctool/, etc.
configDir, _ := cfgstore.CLIConfigDir(appSlug)
dataDir, _ := cfgstore.XDGDataDir(appSlug)
stateDir, _ := cfgstore.XDGStateDir(appSlug)
```

### System Daemon (FHS)

```go
const appSlug = dt.PathSegment("doctool")

// Uses /etc/doctool/, /var/lib/doctool/, /var/log/doctool/
configDir, _ := cfgstore.SystemConfigDir(appSlug)
dataDir, _ := cfgstore.SystemDataDir(appSlug)
logDir, _ := cfgstore.SystemLogDir(appSlug)
```

### Self-Contained Install (Prefix)

```go
const appSlug = dt.PathSegment("doctool")
prefix := dt.DirPath("/opt/doctool")

dp := cfgstore.NewPrefixDirsProvider(prefix)

// Uses /opt/doctool/etc/, /opt/doctool/var/, etc.
configDir, _ := cfgstore.SystemConfigDir(appSlug, dp)
dataDir, _ := cfgstore.SystemDataDir(appSlug, dp)
```

### Doctool Layouts

**User mode (CLI):**
```
~/.config/doctool/doctool.json
~/.local/share/doctool/doctool.db
~/.local/state/doctool/logs/
```

**System mode (Debian daemon):**
```
/etc/doctool/doctool.json
/var/lib/doctool/doctool.db
/var/log/doctool/
```

**Portable mode (Synology):**
```
/opt/doctool/etc/doctool.json
/opt/doctool/var/doctool.db
/opt/doctool/var/log/
```

---

## New Errors

```go
var (
    // XDG errors
    ErrXDGRuntimeDirNotSet    = errors.New("XDG_RUNTIME_DIR not set")
    ErrFailedGettingXDGDataDir   = errors.New("failed to get XDG data directory")
    ErrFailedGettingXDGStateDir  = errors.New("failed to get XDG state directory")

    // FHS errors
    ErrFailedGettingSystemConfigDir = errors.New("failed to get system config directory")
    ErrFailedGettingSystemDataDir   = errors.New("failed to get system data directory")
    ErrFailedGettingSystemLogDir    = errors.New("failed to get system log directory")
)
```

---

## Implementation Phases

### Phase 1: XDG User Directories
1. Add XDG DirType constants
2. Add DirsProvider XDG fields
3. Add default XDG implementations
4. Add XDGDataDir, XDGStateDir, XDGCacheDir, XDGRuntimeDir functions
5. Add tests

### Phase 2: FHS System Directories
1. Add FHS DirType constants
2. Add DirsProvider FHS fields
3. Add default FHS implementations
4. Add SystemConfigDir, SystemDataDir, SystemLogDir, SystemRuntimeDir functions
5. Add tests

### Phase 3: Prefix-based Support
1. Implement NewPrefixDirsProvider helper
2. Add tests
3. Document usage patterns

### Phase 4: Convenience Stores (Optional)
1. DataStore for XDG_DATA_HOME
2. StateStore for XDG_STATE_HOME
3. SystemConfigStore for /etc/
4. SystemDataStore for /var/lib/

### Phase 5: Documentation
1. Update README with new functionality
2. Add examples for each mode
3. Document migration path for existing users

---

## Testing Strategy

1. Unit tests for each new DirType
2. Tests with mocked environment variables
3. Tests with DirsProvider overrides (using cstest pattern)
4. Integration tests for EnsureDirs functions
5. Tests for NewPrefixDirsProvider

---

## Open Questions

1. Should `os.UserCacheDir()` be used for XDGCacheDir (it already respects XDG_CACHE_HOME)?
2. Should we support XDG_CONFIG_DIRS and XDG_DATA_DIRS (search paths for system-wide config/data)?
3. Should StateStore support log rotation out of the box?
4. Should there be a "mode" concept that selects user/system/prefix automatically based on context?

---

## References

- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
- [Filesystem Hierarchy Standard 3.0](https://refspecs.linuxfoundation.org/FHS_3.0/fhs/index.html)
- doctool DOCTOOL-SPEC.md — uses cfgstore with all three modes
