package cstest

import (
	"os"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
)

func GetRelConfigDir(cs cfgstore.ConfigStore, args *TestDirsProviderArgs) (rel dt.PathSegments, err error) {
	var dir dt.DirPath

	dir, err = args.WithoutTestRoot(func() (dt.DirPath, error) {
		return cs.ConfigDir()
	})
	if err != nil {
		goto end
	}
	rel, err = relToVolumeRoot(dir)
end:
	return rel, err
}

// relToVolumeRoot returns dir relative to its volume root: "/" on Unix, or
// "C:\" or "\\server\\share\" on Windows.
func relToVolumeRoot(dir dt.DirPath) (rel dt.PathSegments, err error) {
	var vol dt.VolumeName
	var base dt.DirPath

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
		rel = dt.PathSegments(dir)
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
