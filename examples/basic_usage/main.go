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
