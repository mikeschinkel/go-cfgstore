package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
)

type NormalizeArgs struct {
	DirType    DirType
	SourceFile dt.Filepath
	Options    Options
}
