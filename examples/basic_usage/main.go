package main

import (
	"fmt"
	"log"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
)

type AppConfig struct {
	Username string `json:"username"`
	Theme    string `json:"theme"`
	Debug    bool   `json:"debug"`
}

func main() {
	fmt.Println("go-cfgstore Basic Usage Example")
	fmt.Println("================================\n")

	// Create a CLI config store pointing to ~/.config/myapp-example/config.json
	store := cfgstore.NewCLIConfigStore(dt.PathSegment("myapp-example"), dt.RelFilepath("config.json"))

	// Example configuration
	config := AppConfig{
		Username: "demo-user",
		Theme:    "dark",
		Debug:    false,
	}

	// Save configuration
	err := store.SaveJSON(&config)
	if err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Println("Configuration saved successfully")

	// Get the config directory location
	configDir, err := store.ConfigDir()
	if err != nil {
		log.Fatalf("Failed to get config dir: %v", err)
	}
	fmt.Printf("Config directory: %s\n\n", configDir)

	// Load configuration
	var loaded AppConfig
	err = store.LoadJSON(&loaded)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Loaded config:\n")
	fmt.Printf("  Username: %s\n", loaded.Username)
	fmt.Printf("  Theme:    %s\n", loaded.Theme)
	fmt.Printf("  Debug:    %v\n\n", loaded.Debug)

	// Check if config exists
	exists := store.Exists()
	fmt.Printf("Config exists: %v\n", exists)

	fmt.Println("\nUsage Notes:")
	fmt.Println("- NewCLIConfigStore creates configs in ~/.config/<slug>/")
	fmt.Println("- NewProjectConfigStore creates configs in project root")
	fmt.Println("- SaveJSON/LoadJSON use encoding/json/v2")
	fmt.Println("- ConfigDir() returns the directory path")
	fmt.Println("- Exists() checks if the config file exists")
}
