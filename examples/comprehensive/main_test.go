package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-dt"
)

func TestRun(t *testing.T) {
	// CreateWriterLogger logs under the real ~/.config, so point it at a temp home.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home+"/.config")

	var out bytes.Buffer
	err := run(&out, dt.DirPath(t.TempDir()))
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}
	for _, want := range []string{
		"writer ready",
		"CLI:           <root>/home/.config/cfgstore-comprehensive",
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
