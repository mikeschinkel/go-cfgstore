package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
)

func ProjectConfigFilepath(configSlug dt.PathSegment, configFile dt.RelFilepath, dps ...*DirsProvider) (cfp dt.Filepath, err error) {
	var cd dt.DirPath
	var dp *DirsProvider
	if dps != nil {
		dp = dps[0]
	}
	cd, err = ConfigDir(ProjectConfigDirType, configSlug, dp)
	if err != nil {
		goto end
	}
	cfp = dt.FilepathJoin(cd, configFile)
end:
	return cfp, err
}
