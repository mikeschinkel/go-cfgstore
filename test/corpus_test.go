package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
)

// TestFuzzCorpus runs all fuzz corpus files as regression tests
func TestFuzzCorpus(t *testing.T) {
	corpusDir := "testdata/fuzz"

	if _, err := os.Stat(corpusDir); os.IsNotExist(err) {
		// t.Skip("No fuzz corpus found - run fuzzing locally to generate corpus")
		return
	}

	fuzzTests := []string{
		"FuzzNewCLIConfigStore",
		"FuzzNewProjectConfigStore",
	}

	for _, fuzzTest := range fuzzTests {
		t.Run(fuzzTest, func(t *testing.T) {
			testDir := filepath.Join(corpusDir, fuzzTest)

			if _, err := os.Stat(testDir); os.IsNotExist(err) {
				// t.Skipf("No corpus for %s", fuzzTest)
				return
			}

			entries, err := os.ReadDir(testDir)
			if err != nil {
				t.Fatalf("Failed to read corpus directory: %v", err)
			}

			if len(entries) == 0 {
				// t.Skipf("No corpus files for %s", fuzzTest)
				return
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				t.Run(entry.Name(), func(t *testing.T) {
					corpusFile := filepath.Join(testDir, entry.Name())
					data, err := os.ReadFile(corpusFile)
					if err != nil {
						t.Fatalf("Failed to read corpus file: %v", err)
					}

					switch fuzzTest {
					case "FuzzNewCLIConfigStore":
						runNewCLIConfigStoreCorpus(t, data)
					case "FuzzNewProjectConfigStore":
						runNewProjectConfigStoreCorpus(t, data)
					}
				})
			}
		})
	}
}

func runNewCLIConfigStoreCorpus(t *testing.T, data []byte) {
	slug := string(data)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewCLIConfigStore panicked with slug: %q, panic: %v", slug, r)
		}
	}()

	_ = cfgstore.NewCLIConfigStore(dt.PathSegment(slug), "")
}

func runNewProjectConfigStoreCorpus(t *testing.T, data []byte) {
	slug := string(data)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewProjectConfigStore panicked with slug: %q, panic: %v", slug, r)
		}
	}()

	_ = cfgstore.NewProjectConfigStore(dt.PathSegment(slug), "")
}
