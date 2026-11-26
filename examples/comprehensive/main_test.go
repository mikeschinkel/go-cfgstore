package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
)

func TestRun(t *testing.T) {
	// CreateWriterLogger logs under the real ~/.config, so point it at a temp home.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home+"/.config")

	root := dt.DirPath(t.TempDir())

	// The CLI config dir differs by OS, so ask cfgstore for it.
	cliDir, err := cfgstore.CLIConfigDir(configSlug, newDirsProvider(root))
	if err != nil {
		t.Fatalf("CLIConfigDir: %v", err)
	}
	cliRel, err := cliDir.Rel(root)
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}

	var out bytes.Buffer
	err = run(&out, root)
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}
	for _, want := range []string{
		"writer ready",
		"CLI:           <root>/" + string(cliRel) + "\n",
		"Project:       <root>/project/.cfgstore-comprehensive",
		"App:           <root>/appconfig/cfgstore-comprehensive",
		"project file:  <root>/project/.cfgstore-comprehensive/config.json",
		"merged: theme=dark editor=helix layers=[cli project]",
		"cli only:     theme=dark editor=vim",
		"project only: theme= editor=helix",
		"exists: true",
		"exists elsewhere: false",
		"created:       <root>/newproject/.cfgstore-comprehensive/config.json",
		"second call: config already exists",
		"app cache:     <root>/cache/cfgstore-comprehensive/cli",
		"fresh CLI config created: layers=[cli]",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}
