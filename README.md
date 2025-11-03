# go-cfgstore

Configuration storage package for Go programs that provides cross-platform support for managing configuration files in multiple locations.

## Overview

`go-cfgstore` simplifies the management of configuration files by providing a unified API for storing and loading configuration data across different directory types (CLI, project, and app-specific). It handles platform-specific directory conventions automatically and supports both JSON and raw byte data.

## Status

This is **pre-alpha** and in development thus **subject to change**, although I am trying to bring to v1.0 as soon as I feel confident its architecture will not need to change. As of Novemeber 2025 I am actively working on it and using it in current projects.

If you find value in this project and want to use it, please start a discuss to let me know. If you discuver any issues with it, please open an issue or submit a pull request. 

## Features

- **Multiple Configuration Locations**: Support for CLI configs (`~/.config/<slug>`), project configs (`<project-dir>/.<slug>`), and app configs (platform-specific user config directory)
- **Cross-Platform**: Automatically handles directory path conventions for macOS, Linux, and Windows
- **JSON Support**: Built-in JSON serialization/deserialization using Go's JSON v2 with pretty printing
- **Hierarchical Configuration**: Load and merge configuration from multiple locations with precedence rules
- **Type-Safe**: Uses domain types from `go-dt` for compile-time path safety
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
    "github.com/mikeschinkel/go-cfgstore"
    "github.com/mikeschinkel/go-dt"
)

type AppConfig struct {
    Username string `json:"username"`
    Theme    string `json:"theme"`
}

func main() {
    // Create a CLI config store (~/.config/myapp/config.json)
    store := cfgstore.NewCLIConfigStore(
        dt.PathSegment("myapp"),
        dt.RelFilepath("config.json"),
    )

    // Save configuration
    config := AppConfig{
        Username: "alice",
        Theme:    "dark",
    }
    err := store.SaveJSON(&config)
    if err != nil {
        panic(err)
    }

    // Load configuration
    var loaded AppConfig
    err = store.LoadJSON(&loaded)
    if err != nil {
        panic(err)
    }
}
```

### Project Configuration

```go
// Create a project config store (<project-dir>/.myapp/settings.json)
store := cfgstore.NewProjectConfigStore(
    dt.PathSegment("myapp"),
    dt.RelFilepath("settings.json"),
)
```

### Multiple Configuration Stores

```go
// Create stores for both CLI and project configs
stores := cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
    ConfigStoreArgs: cfgstore.ConfigStoreArgs{
        ConfigSlug:  dt.PathSegment("myapp"),
        RelFilepath: dt.RelFilepath("config.json"),
    },
    DirTypes: []cfgstore.DirType{
        cfgstore.CLIConfigDir,
        cfgstore.ProjectConfigDir,
    },
})

// Access specific stores
cliStore := stores.CLIConfigStore()
projectStore := stores.ProjectConfigStore()
```

## Core Types

### ConfigStore Interface

The main interface for configuration file operations:

```go
type ConfigStore interface {
    Load() ([]byte, error)
    Save([]byte) error
    LoadJSON(data any, opts ...jsonv2.Options) error
    SaveJSON(data any) error
    Exists() bool
    GetFilepath() (dt.Filepath, error)
    ConfigDir() (dt.DirPath, error)
    WithDirType(DirType) ConfigStore
    // ... additional methods
}
```

### DirType

Configuration directory types:

- `CLIConfigDir` - User's CLI config directory (`~/.config/<slug>`)
- `ProjectConfigDir` - Project-specific config directory (`<project-dir>/.<slug>`)
- `AppConfigDir` - OS-specific application config directory (e.g., `~/Library/Application Support/<slug>` on macOS)

### RootConfig Interface

Interface for application-specific root configuration that requires normalization:

```go
type RootConfig interface {
    RootConfig()
    Normalize(dt.Filepath, Options) error
    IsNil() bool
}
```

## Advanced Usage

### Custom Directory Providers

For testing or special scenarios, you can provide custom directory functions:

```go
store := cfgstore.NewConfigStore(cfgstore.CLIConfigDir, cfgstore.ConfigStoreArgs{
    ConfigSlug:  dt.PathSegment("myapp"),
    RelFilepath: dt.RelFilepath("config.json"),
    DirsProvider: &cfgstore.DirsProvider{
        UserHomeDirFunc:   customHomeDirFunc,
        UserConfigDirFunc: customConfigDirFunc,
        GetwdFunc:         customGetwdFunc,
        ProjectDirFunc:    customProjectDirFunc,
    },
})
```

### Loading Root Configuration with Precedence

```go
stores := cfgstore.NewConfigStores(/* ... */)

var config MyRootConfig
err := stores.LoadRootConfig(&config, cfgstore.RootConfigArgs{
    DirTypes: []cfgstore.DirType{
        cfgstore.CLIConfigDir,
        cfgstore.ProjectConfigDir,
    },
    Options: myOptions,
})
```

The configuration will be loaded with project config taking precedence over CLI config.

### Subdirectories in Config Files

You can store config files in subdirectories:

```go
store := cfgstore.NewCLIConfigStore(
    dt.PathSegment("myapp"),
    dt.RelFilepath("tokens/user@example.com.json"),
)
```

This creates: `~/.config/myapp/tokens/user@example.com.json`

## Testing Support

The `cstest` package provides utilities for testing:

```go
import "github.com/mikeschinkel/go-cfgstore/cstest"

func TestMyConfig(t *testing.T) {
    // Create test store with temporary directory
    testRoot := dtx.TempTestDir(t)

    store := cfgstore.NewConfigStore(cfgstore.CLIConfigDir, cfgstore.ConfigStoreArgs{
        ConfigSlug:  dt.PathSegment("testapp"),
        RelFilepath: dt.RelFilepath("config.json"),
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
| `Getwd()`         | `/Users/mikeschinkel/Projects/xmlui/localdev`     |
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
