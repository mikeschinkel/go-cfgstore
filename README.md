# go-cfgstore

Configuration storage package for Go programs that provides cross-platform support for managing configuration files in multiple locations.

## Overview

`go-cfgstore` simplifies the management of configuration files by providing a unified API for storing and loading configuration data across different directory types (CLI, project, and app-specific). It handles platform-specific directory conventions automatically and supports both JSON and raw byte data.

## Status

This is **beta** software and in development thus **subject to change**. As of December 2025 I am actively working on it for use in several other projects. The state of change is slowingly as I feel I have identified most requirements so I expect to bring to v1.0 in the relatively near future.

If you find value in this project and want to use it, please start a discussion to let me know. If you discover any issues with it, please open an issue or submit a pull request.

## Features

- **Multiple Configuration Locations**: Support for CLI configs (`~/.config/<slug>`), project configs (`<project-dir>/.<slug>`), and app configs (platform-specific user config directory)
- **Cross-Platform**: Automatically handles directory path conventions for macOS, Linux, and Windows
- **Cache Directory Support**: Platform-specific cache directories for shared and app-specific caching needs
- **Project Initialization**: Simplified project config initialization with `InitProjectConfig()` for `init` commands
- **JSON Support**: Built-in JSON serialization/deserialization using Go's JSON v2 with pretty printing
- **Hierarchical Configuration**: Load and merge configuration from multiple locations with precedence rules
- **Flexible Directory Management**: Create subdirectories, override paths, and switch between directory types
- **Type-Safe**: Uses domain types from `go-dt` for compile-time path safety (see [Type-Safe Path Handling](#type-safe-path-handling-with-go-dt))
- **Test-Friendly**: Includes `cstest` package with utilities for testing configuration code
- **Structured Errors**: Uses the `doterr` pattern for rich error context and metadata

## Installation

```bash
go get github.com/mikeschinkel/go-cfgstore
```

## Quick Start

### Basic Usage

```go
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mikeschinkel/go-cfgstore"
)

type AppConfig struct {
	Username string `json:"username"`
	Theme    string `json:"theme"`
}

func main() {
	cfgstore.SetLogger(slog.Default())

	fmt.Println("go-cfgstore Basic Usage Example")
	fmt.Println("================================\n")

	// Create a CLI config store at ~/.config/myapp/config.json
	// NOTE: Each "store" is a single file within a config directory
	store := cfgstore.NewCLIConfigStore("myapp", "config.json")

	// Get the full filepath of the config store file
	fp, err := store.GetFilepath()
	if err != nil {
		panic(err)
	}

	// Display the full filepath
	fmt.Printf("Config store filepath:\n\t%s\n", fp)

	// Check if config exists
	exists := store.Exists()
	fmt.Printf("Config store file exists:\n\t%v\n", exists)

	// Define your configuration
	config := AppConfig{
		Username: "alice",
		Theme:    "dark",
	}

	// Add the values in config to ~/.config/myapp/config.json
	if err := store.SaveJSON(&config); err != nil {
		panic(err)
	}

	// Now load the config into another variable
	var loaded AppConfig
	if err := store.LoadJSON(&loaded); err != nil {
		panic(err)
	}

	// Display the loaded value
	fmt.Printf("Config values:\n\t%#v\n", loaded)

	// Load the config file into a []byte buffer
	content, err := os.ReadFile(string(fp))
	if err != nil {
		panic(err)
	}

	// Finally, display the config file content as a string
	fmt.Printf("JSON Content:\n%s\n", string(content))

	fmt.Println("\nUsage Notes:")
	fmt.Println("- NewCLIConfigStore creates configs in ~/.config/<slug>/")
	fmt.Println("- NewProjectConfigStore creates configs in project root")
	fmt.Println("- SaveJSON/LoadJSON use encoding/json/v2")
	fmt.Println("- ConfigDir() returns the directory path")
	fmt.Println("- Exists() checks if the config file exists")

}
```

### Project Configuration

```go
// Create a project config store (<project-dir>/.myapp/settings.json)
store := cfgstore.NewProjectConfigStore("myapp","settings.json")
```

### Loading Configuration with Precedence

When you need to load configuration from multiple locations with precedence (e.g., CLI defaults + Project overrides), use the convenience functions. Your config struct must implement the `RootConfig` interface (see [RootConfig Interface Implementation](#rootconfig-interface-implementation-pattern)).

```go
// Load from CLI and Project configs, Project takes precedence
config, err := cfgstore.LoadDefaultConfig[AppConfig, *AppConfig](cfgstore.LoadConfigArgs{
    ConfigSlug: "myapp",
    ConfigFile: "config.json",
    Options:    nil,
})
if err != nil {
    panic(err)
}

// Use the merged configuration
fmt.Printf("Username: %s, Theme: %s\n", config.Username, config.Theme)
```

See [API Tiers](#api-tiers) for alternative approaches, [Understanding Normalize()](#understanding-normalize) for the config lifecycle, and [Common Patterns](#common-patterns) for advanced merging patterns.

## API Tiers

`go-cfgstore` provides three tiers of API to match different use cases, from simple convenience functions to maximum control:

### Tier 1: Convenience Functions (Recommended for Most Cases)

Simple, single-call functions for common configuration patterns:

```go
// CLI-only configuration (~/.config/myapp/config.json)
config, err := cfgstore.LoadCLIConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
    ConfigSlug: "myapp",
    ConfigFile: "config.json",
    Options:    myOptions,  // or nil
})

// Project-only configuration (.myapp/config.json)
config, err := cfgstore.LoadProjectConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
    ConfigSlug: "myapp",
    ConfigFile: "config.json",
    Options:    nil,
})

// CLI + Project with precedence (Project overrides CLI)
config, err := cfgstore.LoadDefaultConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
    ConfigSlug: "myapp",
    ConfigFile: "config.json",
    Options:    myOptions,
})
```

**Key Benefits:**
- No need to construct DirTypes arrays
- No store creation boilerplate
- Single function call
- Type-safe with generics

### Tier 2: Flexible LoadConfig

For cases requiring custom DirTypes or DirsProvider:

```go
// Custom directory precedence: App → CLI → Project
config, err := cfgstore.LoadConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
    ConfigSlug: "myapp",
    ConfigFile: "config.json",
    DirTypes: []cfgstore.DirType{
        cfgstore.AppConfigDirType,
        cfgstore.CLIConfigDirType,
        cfgstore.ProjectConfigDirType,
    },
    Options:      myOptions,
    DirsProvider: customDirsProvider,  // optional, for testing
})
```

**Key Benefits:**
- Consolidated parameters (DirTypes specified once)
- Sensible defaults for optional fields
- DirsProvider support for testability
- Flexible for complex precedence rules

**Defaults Applied:**
- `DirTypes`: `[CLIConfigDirType, ProjectConfigDirType]`
- `DirsProvider`: `DefaultDirsProvider()`
- `Options`: `nil` is acceptable

### Tier 3: Low-Level LoadConfigStores

For maximum control and advanced scenarios:

```go
// Pre-create stores for reuse or testing
stores := cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
    DirTypes: []cfgstore.DirType{
        cfgstore.CLIConfigDirType,
        cfgstore.ProjectConfigDirType,
    },
    ConfigStoreArgs: cfgstore.ConfigStoreArgs{
        ConfigSlug:   "myapp",
        RelFilepath:  "config.json",
        DirsProvider: customDirsProvider,
    },
})

// Load with fine-grained control
config, err := cfgstore.LoadConfigStores[MyConfig, *MyConfig](stores, cfgstore.RootConfigArgs{
    DirTypes:     []cfgstore.DirType{cfgstore.CLIConfigDirType, cfgstore.ProjectConfigDirType},
    Options:      myOptions,
    DirsProvider: customDirsProvider,
})
```

**When to Use:**
- Pre-creating stores for reuse across multiple loads
- Testing with custom ConfigStores
- Advanced scenarios requiring store manipulation
- Fine-grained control over store creation and loading

### Understanding DirsProvider

`DirsProvider` is an optional parameter for **testing and custom directory resolution**. In production, you typically don't need it—the package uses sensible platform-specific defaults.

**What it does:**
- Provides functions for resolving directories: `UserHomeDir()`, `UserConfigDir()`, `Getwd()`, `ProjectDir()`
- Allows overriding directory resolution for testing without touching the real filesystem
- Enables dependency injection for test isolation

**When to use:**
- **Testing**: Override directory functions to use temporary test directories
- **Custom environments**: Non-standard directory layouts or containerized environments
- **Most production code**: Omit it entirely (uses `DefaultDirsProvider()`)

**Example (testing):**
```go
// Production code - no DirsProvider needed
config, err := cfgstore.LoadCLIConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
    ConfigSlug: "myapp",
    ConfigFile: "config.json",
})

// Test code - override directories
testProvider := &cfgstore.DirsProvider{
    UserHomeDirFunc: func() (dt.DirPath, error) {
        return "/tmp/test/home", nil
    },
}
config, err := cfgstore.LoadCLIConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
    ConfigSlug:   "myapp",
    ConfigFile:   "config.json",
    DirsProvider: testProvider,  // Override for testing
})
```

See [Testing Support](#testing-support) and [DirsProvider Dependency Injection](#dirsprovider-dependency-injection) for more details.

### Decision Tree: Which API Should I Use?

```
┌─────────────────────────────────────────┐
│ Do you need custom DirsProvider         │
│ (e.g., for testing)?                    │
└─────┬───────────────────────────────────┘
      │
      ├─ No ──────────────────────────────┐
      │                                   │
      │  ┌─────────────────────────────┐  │
      │  │ Which config locations?     │  │
      │  └──┬──────────────────────────┘  │
      │     │                             │
      │     ├─ CLI only ────→ LoadCLIConfig()
      │     │
      │     ├─ Project only ─→ LoadProjectConfig()
      │     │
      │     └─ CLI + Project → LoadDefaultConfig()
      │
      └─ Yes ─────────────────────────────┐
                                          │
         ┌─────────────────────────────┐  │
         │ Do you need custom DirTypes │  │
         │ precedence?                 │  │
         └──┬──────────────────────────┘  │
            │                             │
            ├─ No ─────→ LoadDefaultConfig(LoadConfigArgs{
            │               ConfigSlug: ...,
            │               ConfigFile: ...,
            │               DirsProvider: customProvider,
            │               Options: ...,
            │           })
            │
            ├─ Yes ────→ LoadConfig(LoadConfigArgs{
            │               ConfigSlug: ...,
            │               ConfigFile: ...,
            │               DirTypes: [...],
            │               DirsProvider: customProvider,
            │               Options: ...,
            │           })
            │
            └─ Advanced → LoadConfigStores(stores, args)
               (store reuse, manipulation)
```

## Core Types

### ConfigStore Interface

The main interface for configuration file operations:

```go
type ConfigStore interface {
    // File Operations
    Load() ([]byte, error)
    Save([]byte) error
    LoadJSON(data any, opts ...jsonv2.Options) error
    SaveJSON(data any) error
    Exists() bool

    // Path Operations
    GetFilepath() (dt.Filepath, error)
    GetRelFilepath() dt.RelFilepath
    SetRelFilepath(dt.RelFilepath)
    ConfigDir() (dt.DirPath, error)
    SetConfigDir(dt.DirPath)

    // Directory Management
    EnsureDirs(subdirs []dt.PathSegment) error

    // Configuration
    WithDirType(DirType) ConfigStore
    DirType() DirType
    ConfigSlug() dt.PathSegment
}
```

### Configuration Functions

The following functions use Go generics for type-safe configuration loading. The generic type parameters are:

- **`RC`** (Root Config): Your configuration struct type (e.g., `MyConfig`)
- **`PRC`** (Pointer to Root Config): Pointer type (e.g., `*MyConfig`) that implements `RootConfigPtr[RC]`

This allows the package to work with your custom config types while maintaining type safety. When calling these functions, specify both types explicitly:

```go
// Example: Loading a MyConfig struct
config, err := cfgstore.LoadCLIConfig[MyConfig, *MyConfig](args)
//                                     ^^^^^^^^  ^^^^^^^^^
//                                     RC (struct) PRC (pointer)
```

See [Generic Type Constraints](#generic-type-constraints) for the design rationale.

**Project Initialization:**
```go
// InitProjectConfig initializes a new project configuration.
// Returns ErrConfigAlreadyExists if the config file already exists.
func InitProjectConfig[RC any, PRC RootConfigPtr[RC]](
    configSlug dt.PathSegment,
    configFile dt.RelFilepath,
    opts Options,
) (PRC, error)
```

**Multi-Store Configuration:**
```go
// LoadConfigStores loads and merges configuration from multiple stores.
// Later stores in DirTypes array take precedence over earlier ones.
func LoadConfigStores[RC any, PRC RootConfigPtr[RC]](
    stores *ConfigStores,
    args RootConfigArgs,
) (PRC, error)
```

**Config Directory Helpers:**
```go
// CLIConfigDir returns ~/.config/<slug> directory
func CLIConfigDir(configSlug dt.PathSegment) (dt.DirPath, error)

// AppConfigDir returns platform-specific app config directory
func AppConfigDir(configSlug dt.PathSegment) (dt.DirPath, error)

// ProjectConfigDir returns ./<slug> directory in current working directory
func ProjectConfigDir(configSlug dt.PathSegment) (dt.DirPath, error)
```

### DirType

Configuration directory types determine where config files are stored. Understanding the distinction is crucial for choosing the right storage location. See [Cache Directories](#cache-directories) for information about cache vs. config storage.

#### CLIConfigDir vs. AppConfigDir: When to Use Which?

**CLIConfigDir** (`~/.config/<slug>`):
- **Purpose**: Configuration for **command-line tools** and **developer-focused applications**
- **Location**: Always `~/.config/<slug>` on all platforms (UNIX convention)
- **Visibility**: Easy to find, edit, and version control
- **Use when**: Building CLI tools, developer utilities, or when users need direct file access
- **Examples**: `git` config, `npm` config, command-line database tools

**AppConfigDir** (platform-specific):
- **Purpose**: Configuration for **GUI applications** and **end-user software**
- **Location**: Platform-specific directories managed by the OS:
  - macOS: `~/Library/Application Support/<slug>`
  - Linux: `~/.config/<slug>` (same as CLIConfigDir)
  - Windows: `%APPDATA%\<slug>`
- **Visibility**: OS-managed, follows platform conventions
- **Use when**: Building GUI apps, system services, or following OS integration guidelines
- **Examples**: VS Code, Slack, Chrome (all use platform-specific paths)

**ProjectConfigDir** (`<project-dir>/.<slug>`):
- **Purpose**: Project-specific configuration **within a repository or workspace**
- **Location**: Hidden directory in current working directory
- **Visibility**: Lives alongside project files, can be committed to version control
- **Use when**: Per-project settings that differ from global config
- **Examples**: `.vscode/settings.json`, `.git/config`, ESLint project config

#### Platform-Specific Paths Summary

| DirType | macOS | Linux | Windows |
|---------|-------|-------|---------|
| **CLIConfigDir** | `~/.config/<slug>` | `~/.config/<slug>` | `~/.config/<slug>` |
| **AppConfigDir** | `~/Library/Application Support/<slug>` | `~/.config/<slug>` | `%APPDATA%\<slug>` |
| **ProjectConfigDir** | `<cwd>/.<slug>` | `<cwd>/.<slug>` | `<cwd>\.<slug>` |

#### Decision Guide

```
┌─────────────────────────────────────┐
│ What type of application?           │
└─────┬───────────────────────────────┘
      │
      ├─ CLI tool / Developer utility → CLIConfigDir
      │   (go-cfgstore, database CLIs, build tools)
      │
      ├─ GUI app / End-user software → AppConfigDir
      │   (editors, desktop apps, system services)
      │
      └─ Per-project settings → ProjectConfigDir
          (workspace config, repo-specific settings)
```

**Common Pattern: CLI + Project**
Many developer tools use **both** CLIConfigDir (global defaults) and ProjectConfigDir (project overrides):

```go
// Load global CLI config as defaults, project config overrides
config, err := cfgstore.LoadDefaultConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
    ConfigSlug: "mytool",
    ConfigFile: "config.json",
    // Loads from CLIConfigDir first, then ProjectConfigDir
    // Project settings override CLI settings
})
```

### RootConfig Interface

Interface for application-specific root configuration that requires normalization. See [RootConfig Interface Implementation Pattern](#rootconfig-interface-implementation-pattern) for detailed implementation guidance and [Understanding Normalize()](#understanding-normalize) for the config lifecycle.

```go
type RootConfig interface {
    RootConfig()
    Normalize(NormalizeArgs) error
    Merge(RootConfig) RootConfig
}
```

## Cache Directories

While configuration files store **persistent user preferences and settings**, cache directories store **temporary, regenerable data** that improves performance. The `go-cfgstore` package provides cache directory functions because applications often need both.

### When to Use Cache vs. Config Directories

**Use Config Directories for:**
- User preferences and settings
- Application state that should persist across updates
- Data that users might manually edit (JSON, YAML, etc.)
- Project-specific configuration

**Use Cache Directories for:**
- Downloaded files that can be re-fetched
- Compiled or processed data that can be regenerated
- Temporary build artifacts
- Local copies of remote resources (e.g., Git repos for MCP servers)

### Cache Directory Functions

```go
// GetSharedCacheDir returns platform-specific shared cache directory
// Example: ~/.cache/myapp (Linux), ~/Library/Caches/myapp (macOS)
func GetSharedCacheDir(slug dt.PathSegment, opts ...CacheOptions) (dt.DirPath, error)

// GetAppCacheDir returns platform-specific app-specific cache directory
// Example: ~/.cache/myapp/editor (Linux), ~/Library/Caches/myapp/editor (macOS)
func GetAppCacheDir(slug, appName dt.PathSegment, opts ...CacheOptions) (dt.DirPath, error)
```

### Platform-Specific Cache Locations

| Platform | Shared Cache (`myapp`) | App Cache (`myapp/editor`) |
|----------|------------------------|----------------------------|
| Linux    | `~/.cache/myapp`       | `~/.cache/myapp/editor`    |
| macOS    | `~/Library/Caches/myapp` | `~/Library/Caches/myapp/editor` |
| Windows  | `%LOCALAPPDATA%\myapp` | `%LOCALAPPDATA%\myapp\editor` |

### Example: Caching Remote Git Repos for MCP Servers

A common use case is caching Git repositories that contain content for MCP servers:

```go
package main

import (
    "fmt"
    "os/exec"

    "github.com/mikeschinkel/go-cfgstore"
)

func main() {
    // Get cache directory for your MCP server
    cacheDir, err := cfgstore.GetSharedCacheDir("my-mcp-server")
    if err != nil {
        panic(err)
    }

    repoPath := string(cacheDir) + "/docs-repo"

    // Clone or update the repo
    if !fileExists(repoPath) {
        // First time: clone
        cmd := exec.Command("git", "clone", "https://github.com/example/docs.git", repoPath)
        if err := cmd.Run(); err != nil {
            panic(err)
        }
    } else {
        // Subsequent runs: pull latest
        cmd := exec.Command("git", "-C", repoPath, "pull")
        if err := cmd.Run(); err != nil {
            panic(err)
        }
    }

    fmt.Printf("Docs cached at: %s\n", repoPath)
}
```

### Example: Separating Cache from Config

```go
// Config goes in config directory (persistent user preferences)
configStore := cfgstore.NewCLIConfigStore("myapp", "config.json")
config := AppConfig{
    Username: "alice",
    Theme:    "dark",
}
configStore.SaveJSON(&config)

// Cache goes in cache directory (temporary, regenerable data)
cacheDir, err := cfgstore.GetSharedCacheDir("myapp")
if err != nil {
    panic(err)
}
downloadPath := string(cacheDir) + "/downloads"
// ... download and cache files to downloadPath
```

**Key Distinction:**
- **Config**: `~/.config/myapp/config.json` - User's preferences (survives app updates)
- **Cache**: `~/.cache/myapp/downloads/` - Temporary data (can be deleted anytime)

## Advanced Usage

### Custom Directory Providers

For testing or special scenarios, you can provide custom directory functions:

```go
store := cfgstore.NewConfigStore(cfgstore.CLIConfigDir, cfgstore.ConfigStoreArgs{
    ConfigSlug:  "myapp",
    RelFilepath: "config.json",
    DirsProvider: &cfgstore.DirsProvider{
        UserHomeDirFunc:   customHomeDirFunc,
        UserConfigDirFunc: customConfigDirFunc,
        GetwdFunc:         customGetwdFunc,
        ProjectDirFunc:    customProjectDirFunc,
    },
})
```

### Loading Root Configuration with Precedence

Load and merge configuration from multiple stores (project config overrides CLI config):

```go
// Your config must implement RootConfig interface
type MyRootConfig struct {
    Username string `json:"username"`
    Theme    string `json:"theme"`
}

func (c *MyRootConfig) Normalize(args cfgstore.NormalizeArgs) error {
    if c.Theme == "" {
        c.Theme = "light" // default
    }
    return nil
}
func (c *MyRootConfig) RootConfig() {}

// Create stores
stores := cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
    DirTypes: []cfgstore.DirType{
        cfgstore.CLIConfigDirType,      // ~/.config/myapp/
        cfgstore.ProjectConfigDirType,  // ./.myapp/
    },
    ConfigStoreArgs: cfgstore.ConfigStoreArgs{
        ConfigSlug:  "myapp",
        RelFilepath: "config.json",
    },
})

// Load with precedence (Project overrides CLI)
config, err := cfgstore.LoadConfigStores[MyRootConfig, *MyRootConfig](
    stores,
    cfgstore.RootConfigArgs{
        DirTypes: []cfgstore.DirType{
            cfgstore.CLIConfigDirType,
            cfgstore.ProjectConfigDirType,
        },
        Options: nil, // or your custom options
    },
)
```

The configuration will be loaded with project config taking precedence over CLI config.

### Using Cache Directories

Get platform-specific cache directories for your application:

```go
// Shared cache directory (~/.cache/myapp on Linux, ~/Library/Caches/myapp on macOS)
cacheDir, err := cfgstore.GetSharedCacheDir("myapp")
if err != nil {
    panic(err)
}

// App-specific cache directory
appCacheDir, err := cfgstore.GetAppCacheDir(
    "myapp",
    "editor", // app name
)
if err != nil {
    panic(err)
}
```

### Creating Subdirectories in Config Directory

Use `EnsureDirs` to create subdirectories under your config directory:

```go
store := cfgstore.NewCLIConfigStore(
    "myapp",
    "config.json",
)

// Create subdirectories: ~/.config/myapp/tokens/ and ~/.config/myapp/cache/
err := store.EnsureDirs([]dt.PathSegment{
    "tokens",
    "cache",
})
```

### Subdirectories in Config Files

You can store config files in subdirectories:

```go
store := cfgstore.NewCLIConfigStore(
    "myapp",
    "tokens/user@example.com.json",
)
```

This creates: `~/.config/myapp/tokens/user@example.com.json`

## Common Patterns

### Project Initialization Pattern

For CLI tools with `init` commands, use `InitProjectConfig` to create project configuration:

```go
import (
    "github.com/mikeschinkel/go-cfgstore"
    "github.com/mikeschinkel/go-dt"
)

type MyProjectConfig struct {
    Version string `json:"version"`
    Name    string `json:"name"`
}

func (c *MyProjectConfig) Normalize(args cfgstore.NormalizeArgs) error {
    if c.Version == "" {
        c.Version = "1.0"
    }
    return nil
}

func (c *MyProjectConfig) RootConfig() {}

func initProject() error {
    config, err := cfgstore.InitProjectConfig[MyProjectConfig, *MyProjectConfig](
        "myapp",
        "config.json",
        nil, // options
    )
    if err != nil {
        if errors.Is(err, cfgstore.ErrConfigAlreadyExists) {
            fmt.Println("Project already initialized")
            return nil
        }
        return err
    }

    fmt.Printf("Created project config: %+v\n", config)
    return nil
}
```

### Single/Dual/Triple-Store Configuration

**Single Store** - Project-only configuration:
```go
// Use InitProjectConfig for simple project-only configs
config, err := cfgstore.InitProjectConfig[MyConfig, *MyConfig](
    "myapp",
    "config.json",
    nil,
)
```

**Dual Store** - User/CLI defaults + Project overrides (common pattern):
```go
// Create stores for both CLI and Project configs
configStores := cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
    DirTypes: []cfgstore.DirType{
        cfgstore.CLIConfigDirType,      // User defaults in ~/.config/myapp/
        cfgstore.ProjectConfigDirType,   // Project overrides in ./.myapp/
    },
    ConfigStoreArgs: cfgstore.ConfigStoreArgs{
        ConfigSlug:  "myapp",
        RelFilepath: "config.json",
    },
})

// Load with precedence (Project overrides CLI)
config, err := cfgstore.LoadConfigStores[MyConfig, *MyConfig](
    configStores,
    cfgstore.RootConfigArgs{
        DirTypes: []cfgstore.DirType{
            cfgstore.CLIConfigDirType,
            cfgstore.ProjectConfigDirType,
        },
        Options: myOptions,
    },
)
```

**Triple Store** - Machine + User/CLI + Project (currently untested future use case):
```go
configStores := cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
    DirTypes: []cfgstore.DirType{
        cfgstore.AppConfigDirType,       // System defaults
        cfgstore.CLIConfigDirType,       // User defaults
        cfgstore.ProjectConfigDirType,   // Project overrides
    },
    // ... same as dual store
})
```

### App-Level LoadConfigStores Helper Pattern

For when you need more control with your config stores:

```go
package config

import (
    "github.com/mikeschinkel/go-cfgstore"
    "github.com/mikeschinkel/go-dt"
)

// LoadAppConfig is an app-specific wrapper for LoadRootConfig
// that reduces boilerplate for this application's config loading pattern
func LoadAppConfig[RC any, PRC cfgstore.RootConfigPtr[RC]](
    dirTypes []cfgstore.DirType,
    opts *Options,
    dirsProvider *cfgstore.DirsProvider,
) (PRC, error) {
    configStores := cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
        DirTypes: dirTypes,
        ConfigStoreArgs: cfgstore.ConfigStoreArgs{
            ConfigSlug:   "myapp",  // Your app slug
            RelFilepath:  "config.json",  // Your config file
            DirsProvider: dirsProvider,
        },
    })

    return cfgstore.LoadConfigStores[RC, PRC](configStores, cfgstore.RootConfigArgs{
        DirTypes:     dirTypes,
        Options:      opts,
        DirsProvider: dirsProvider,
    })
}

// Usage in your code:
config, err := LoadAppConfig[MyConfig, *MyConfig](
    []cfgstore.DirType{
        cfgstore.CLIConfigDirType,
        cfgstore.ProjectConfigDirType,
    },
    myOptions,
    nil, // or custom DirsProvider
)
```

### RootConfig Interface Implementation Pattern

Your configuration struct must implement the `RootConfig` interface. The key method is `Normalize()`, which ensures loaded configurations always have complete, valid values.

#### Understanding Normalize()

The `Normalize()` method is called **after loading** a config file but **before** returning it to your application. This ensures:

1. **Default Values**: Missing fields get sensible defaults
2. **Validation**: Invalid configurations are rejected early
3. **Computed Fields**: Derived values can be calculated
4. **Path Resolution**: Relative paths can be made absolute

**The Pattern:**

```go
type MyConfig struct {
    // Your config fields
    Username string `json:"username"`
    Theme    string `json:"theme"`
}

// Normalize applies defaults and validates configuration
func (c *MyConfig) Normalize(args cfgstore.NormalizeArgs) error {
    // 1. Apply defaults for missing values
    if c.Theme == "" {
        c.Theme = "light"  // Default theme
    }

    // 2. Validate required fields
    if c.Username == "" {
        return cfgstore.NewErr(ErrInvalidConfig, "username", "required")
    }

    // 3. (Optional) Parse/transform values if needed
    // For example, you might call Parse functions from go-dt to convert
    // raw strings into domain types with validation

    return nil
}

// RootConfig is a marker method
func (c *MyConfig) RootConfig() {}
```

#### Understanding Merge()

The `Merge()` method enables hierarchical configuration with precedence rules. It's called during `LoadConfigStores()` when loading from multiple directories.

**How Merge Works:**
1. Configs are loaded in DirTypes array order (e.g., CLI first, then Project)
2. For each subsequent config, call `Merge()` on it with the previous config as argument
3. The receiver (caller) typically has higher precedence
4. Return a new merged config

**Example with CLIConfigDirType → ProjectConfigDirType:**
```go
// In merging loop:
rc = projectConfig.Merge(cliConfig)
// projectConfig values typically take precedence over cliConfig values
```

**Implementation Pattern:**
```go
type MyConfig struct {
    Username     string `json:"username"`
    Theme        string `json:"theme"`
    Port         int    `json:"port"`
    GlobalAPIKey string `json:"global_api_key"` // Global setting
}

// Merge combines another config into this one
// The receiver (c) typically takes precedence, but merge logic is field-specific
func (c *MyConfig) Merge(other RootConfig) RootConfig {
    otherCfg := other.(*MyConfig)

    // Start with self
    result := *c

    // Standard case: Receiver wins, fill missing from other
    if result.Username == "" {
        result.Username = otherCfg.Username
    }
    if result.Theme == "" {
        result.Theme = otherCfg.Theme
    }
    if result.Port == 0 {
        result.Port = otherCfg.Port
    }

    // Special case: Global setting - parameter (CLI config) wins over receiver (project)
    // This allows ~/.config/myapp to set a global API key for all projects
    if result.GlobalAPIKey == "" && otherCfg.GlobalAPIKey != "" {
        result.GlobalAPIKey = otherCfg.GlobalAPIKey
    }

    return &result
}
```

**Key Points:**
- The receiver (config calling Merge) **typically** has higher precedence
- The parameter (config passed to Merge) **typically** has lower precedence
- **Usually**, non-zero/empty values from receiver override the parameter
- **However**, merge logic is case-by-case - some fields may make more sense to inherit from the parameter
  - Example: Global settings in `~/.config/myapp/config.json` might override project-specific settings for certain fields
  - The implementer decides which precedence makes sense for each field
- Return a new merged config, don't mutate the original

**Wrapper Pattern:**

If your config struct is defined in another package and you can't modify it, create a wrapper:

```go
import jsonv2 "encoding/json/v2"

// Wrapper pattern when you can't modify the original struct
type MyConfigWrapper struct {
    MyConfig  // Embed the actual config
}

func (w *MyConfigWrapper) Normalize(args cfgstore.NormalizeArgs) error {
    // Delegate to embedded struct's Normalize, or implement here
    return w.MyConfig.Normalize(args)
}

// MarshalJSON delegates to the embedded struct
func (w *MyConfigWrapper) MarshalJSON() ([]byte, error) {
    return jsonv2.Marshal(w.MyConfig)
}

// UnmarshalJSON delegates to the embedded struct
func (w *MyConfigWrapper) UnmarshalJSON(b []byte) error {
    return jsonv2.Unmarshal(b, &w.MyConfig)
}

func (w *MyConfigWrapper) RootConfig() {}
```

**Why the wrapper needs Marshal/Unmarshal methods:**
- JSON operations work on the wrapper type, not the embedded struct
- Without these methods, JSON would try to marshal the wrapper's fields (which includes an embedded struct)
- These methods ensure JSON operations target the actual config struct directly

## Testing Support

The `cstest` package provides utilities for testing:

```go
import "github.com/mikeschinkel/go-cfgstore/cstest"

func TestMyConfig(t *testing.T) {
    // Create test store with temporary directory
    testRoot := dtx.TempTestDir(t)

    store := cfgstore.NewConfigStore(cfgstore.CLIConfigDir, cfgstore.ConfigStoreArgs{
        ConfigSlug:  "testapp",
        RelFilepath: "config.json",
        DirsProvider: cstest.NewTestDirsProvider(&cstest.TestDirsProviderArgs{
            TestRoot: testRoot,
        }),
    })

    // Clean up after test
    t.Cleanup(func() {
        cstest.RemoveAll(t, store)
    })

    // Run tests...
}
```

## Architecture Decisions

This package embodies several intentional design decisions. For detailed rationale, see the `adrs/` directory.

### Type-Safe Path Handling with go-dt

`go-cfgstore` uses the `go-dt` package for **compile-time type safety** with paths and identifiers. While the package internally uses domain types like `dt.Filepath`, `dt.DirPath`, and `dt.PathSegment`, **you don't need to explicitly cast strings** in your code.

**Key Domain Types:**
- **`dt.Filepath`** - Full path including filename (`/home/user/.config/myapp/config.json`)
- **`dt.DirPath`** - Directory path without trailing slash (`/home/user/.config/myapp`)
- **`dt.RelFilepath`** - Relative filepath (`subdir/config.json`)
- **`dt.PathSegment`** - Single path component, no slashes (`myapp`, `config.json`)

**Automatic Conversion:**

Function parameters accept these types, but **Go automatically converts string literals**, so you can write:

```go
// Simple - just use strings
store := cfgstore.NewCLIConfigStore("myapp", "config.json")

// Explicit typing (valid, but unnecessary when using string literals)
store := cfgstore.NewCLIConfigStore(dt.PathSegment("myapp"), dt.RelFilepath("config.json"))
```

Both work identically. Use plain strings for simplicity; explicit types are only needed when:
- Storing paths in variables for reuse
- Working with path manipulation functions from `go-dt`
- Type-checking requires explicit type assertion

**Why use domain types?**
- Prevents passing a full filepath where a directory is expected
- Catches path-related bugs at compile time
- Makes function signatures self-documenting (you know what type of path is expected)

### Embedded doterr Error Handling

Each package embeds its own copy of `doterr.go` (~700 lines) for structured error handling. This is **intentional**, not duplication.

**Why:**
- Provides composable sentinel errors (layer + category) for precise error checking
- Enables key-value metadata instead of brittle format strings (`%s`, `%w`, `%d`)
- Avoids `fmt.Errorf` in favor of structured error construction
- Supports layered error checking: `errors.Is(err, ErrRepo)`, `errors.Is(err, ErrDatabase)`, `errors.Is(err, ErrNotFound)`

**Example:**
```go
err := cfgstore.NewErr(
    cfgstore.ErrFailedToReadFile,
    "filepath", configPath,
    cause, // trailing error
)

// Later, check specific layers:
if errors.Is(err, cfgstore.ErrFailedToReadFile) { /* ... */ }
```

See [ADR 3: Embedded doterr Error Handling Pattern](adrs/003-embedded-doterr-pattern.md) for full rationale.

### Generic Type Constraints

The package uses Go generics with constraints like `[RC any, PRC RootConfigPtr[RC]]` for type-safe configuration loading.

**Trade-offs:**
- ✅ **Benefit:** Compile-time type safety prevents runtime type errors
- ✅ **Benefit:** No type assertions needed in application code
- ⚠️ **Cost:** More verbose function signatures
- ⚠️ **Cost:** Learning curve for developers unfamiliar with generics

**When to use:**
- `InitProjectConfig[RC, PRC]()` - Single config initialization
- `LoadConfigStores[RC, PRC]()` - Multi-config with precedence

See [ADR 4: Generic Type Constraints for Type Safety](adrs/004-generic-type-constraints.md) for full analysis.

### DirsProvider Dependency Injection

The `DirsProvider` pattern enables testability without build tags.

**Why:**
- Allows test code to override directory functions (`UserHomeDir`, `UserConfigDir`, `Getwd`, `ProjectDir`)
- No need for build tags or global state
- Clean dependency injection for test isolation

**Example:**
```go
testProvider := &cfgstore.DirsProvider{
    UserHomeDirFunc: func() (dt.DirPath, error) {
        return "/tmp/test/home", nil
    },
    // ... other overrides
}

store := cfgstore.NewConfigStore(cfgstore.CLIConfigDirType, cfgstore.ConfigStoreArgs{
    ConfigSlug:   "myapp",
    RelFilepath:  "config.json",
    DirsProvider: testProvider,
})
```

See [ADR 1: DirsProvider for Testing](adrs/001-dirs-provider-testing.md) for full rationale.

### TestRoot Naming Convention

Test fixtures use a `TestRoot` naming pattern for consistency.

**Convention:**
- `TestRoot` - Top-level test directory (usually temp directory)
- `TestUserHome` - Mock user home directory within TestRoot
- `TestProjectDir` - Mock project directory within TestRoot

**Why:**
- Consistent naming across test code
- Clear distinction from production paths
- Easier to debug test failures

See [ADR 2: TestRoot Naming Convention](adrs/002-test-root-naming.md) for full details.

### Single vs. Multi-Store API

The package provides both simplified (single-store) and flexible (multi-store) APIs:

- **Single:** `InitProjectConfig()` - For project-only configuration
- **Multi:** `NewConfigStores()` + `LoadConfigStores()` - For config precedence/merging

**When to use which:**
- Use `InitProjectConfig` for CLI tools with `init` commands (simple case)
- Use `LoadConfigStores` when you need CLI defaults + Project overrides (common case)
- Use multi-store for complex precedence: App defaults → CLI config → Project config

See [Common Patterns](#common-patterns) for usage examples.

## Error Handling

The package uses the `doterr` pattern for structured errors with metadata:

```go
err := store.LoadJSON(&config)
if err != nil {
    // Error contains rich context about what failed
    // Examples: ErrFileDoesNotExist, ErrFailedToReadConfigFile, etc.
    log.Printf("Config error: %v", err)
}
```

Common errors:
- `ErrFileDoesNotExist` - Config file not found
- `ErrFailedToReadConfigFile` - File read error
- `ErrFailedToUnmarshalConfigFile` - JSON parsing error
- `ErrConfigDirTypeNotSet` - DirType not specified
- `ErrInvalidConfigDirType` - Invalid DirType value

## Config Directory Cases

### Go Standard Lib on macOS
| Go Func           | Directory                                         |
|-------------------|---------------------------------------------------|
| `UserHomeDir()`   | `/Users/mikeschinkel`                             |
| `Getwd()`         | `/Users/mikeschinkel/Projects/my-project`         |
| `UserConfigDir()` | `/Users/mikeschinkel/Library/Application Support` |
| `UserCacheDir()`  | `/Users/mikeschinkel/Library/Caches`              |

### Production

| macOS         | Directory                   | Linux         | Directory                   | Windows       | Directory                   |
|---------------|-----------------------------|---------------|-----------------------------|---------------|-----------------------------|
| App Config    | `<UserConfigDir>/<s>`       | App Config    | `<UserConfigDir>/<s>`       | App Config    | `<UserConfigDir>\<s>`       |
| CLI Config     | `<UserHomeDir>/.config/<s>` | CLI Config     | `<UserHomeDir>/.config/<s>` | CLI Config     | `<UserHomeDir>\.config\<s>` |
| Project | `<pd>/.<s>`                 | Project | `<pd>/.<s>`                 | Project | `<pd>\.<s>`                 |



### Testing
| macOS         | Directory                                        | Linux         | Directory                         | Windows       | Directory                            |
|---------------|--------------------------------------------------|---------------|-----------------------------------|---------------|--------------------------------------|
| App Config    | `<td>/Users/<u>/Library/Application Support/<s>` | App Config    | `<td>/home/<u>/.config/<s>`       | App Config    | `<td>\Users\<u>\AppData\Roaming\<s>` |
| CLI Config     | `<td>/Users/<u>/.config/<s>`                     | CLI Config     | `<td>/home/<u>/.config/<s>`       | CLI Config     | `<td>\Users\<u>\.config\<s>`         |
| Project | `<td>/Users/<u>/Projects/<p>/.<s>`               | Project | `<td>/home/<u>/projects/<p>/.<s>` | Project | `<td>\Users\<u>\Projects\<p>\.<s>`   |

### Legend

| testDir | projectDir<br>(or `Getwd()`) | username | project | slug  |
|---------|------------------------------|----------|---------|-------|
| `<td>`  | `<pd>`                       | `<u>`    | `<p>`   | `<s>` |

## Dependencies

- `github.com/mikeschinkel/go-dt` - Domain types for type-safe paths and identifiers
- `github.com/mikeschinkel/go-fsfix` - Test fixture utilities
- `github.com/mikeschinkel/go-testutil` - Test utilities

## License

MIT License - Copyright (c) Mike Schinkel

## Contributing

Issues and pull requests are welcome at https://github.com/mikeschinkel/go-cfgstore
