package cstest

import (
	"os"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
)

func RemoveAll(t *testing.T, cs cfgstore.ConfigStore) {
	dir, err := cs.ConfigDir()
	if err != nil {
		t.Errorf("Failed tp get config dir: %v", err)
		goto end
	}
	err = dir.RemoveAll()
	if err != nil {
		t.Errorf("Failed to remove config dir: %v", err)
	}
end:
	return
}

func GetRelConfigDir(cs cfgstore.ConfigStore, args *TestDirsProviderArgs) (rel dt.PathSegments, err error) {
	var vol dt.VolumeName
	var dir, base dt.DirPath

	dir, err = args.WithoutTestRoot(func() (dt.DirPath, error) {
		return cs.ConfigDir()
	})

	if err != nil {
		goto end
	}
	if len(dir) == 0 {
		goto end
	}
	dir = dir.Clean()
	if len(dir) == 0 {
		goto end
	}
	if dir[0] == '/' {
		rel = dt.PathSegments(dir[1:])
		goto end
	}
	vol = dir.VolumeName() // "C:" or "\\server\\share"
	if vol == "" {
		goto end
	}
	base = dt.DirPath(string(vol) + string(os.PathSeparator)) // "C:\" or "\\server\\share\"
	rel, err = dir.Rel(base)                                  // strip the volume root
	if err != nil {
		goto end
	}
	if rel == "." { // exactly the root
		rel = ""
		goto end
	}
end:
	return rel, err
}
