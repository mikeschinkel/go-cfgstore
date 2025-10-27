package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
)

type RootConfig interface {
	RootConfig()
	Normalize(dt.Filepath, Options)
}

type RootConfigArgs struct {
	Options Options
}
