package test

import (
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
)

// FuzzNewCLIConfigStore tests creating CLI config stores with various slugs
func FuzzNewCLIConfigStore(f *testing.F) {
	// Seed with valid and edge-case slugs
	f.Add("myapp")
	f.Add("my-app")
	f.Add("my_app")
	f.Add("app123")
	f.Add("a")
	f.Add("very-long-application-name-with-many-parts")
	f.Add("")
	f.Add(" ")
	f.Add("app with spaces")
	f.Add("app/with/slashes")
	f.Add("app\\with\\backslashes")
	f.Add("app\nwith\nnewlines")
	f.Add("../../../etc/passwd")
	f.Add("~/.ssh")

	f.Fuzz(func(t *testing.T, slug string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewCLIConfigStore panicked with slug %q: %v", slug, r)
			}
		}()

		// Convert string to PathSegment and create store
		// Just verify it doesn't panic - some slugs may be invalid and that's OK
		_ = cfgstore.NewCLIConfigStore(dt.PathSegment(slug), "")
	})
}

// FuzzNewProjectConfigStore tests creating project config stores with various slugs
func FuzzNewProjectConfigStore(f *testing.F) {
	f.Add("myapp")
	f.Add("my-app")
	f.Add("project123")
	f.Add("")
	f.Add(" ")
	f.Add("project with spaces")
	f.Add("../parent")
	f.Add("/absolute/path")

	f.Fuzz(func(t *testing.T, slug string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewProjectConfigStore panicked with slug %q: %v", slug, r)
			}
		}()

		_ = cfgstore.NewProjectConfigStore(dt.PathSegment(slug), "")
	})
}
