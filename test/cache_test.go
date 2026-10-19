package test

import (
	"errors"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
	"github.com/stretchr/testify/assert"
)

func TestGetAppCacheDir_ReturnsUserCacheDirError(t *testing.T) {
	cause := errors.New("no cache dir")
	_, err := cfgstore.GetAppCacheDir("acme", "cli", cfgstore.CacheOptions{
		DirsProvider: &cfgstore.DirsProvider{
			UserCacheDirFunc: func() (dt.DirPath, error) { return "", cause },
		},
	})
	assert.ErrorIs(t, err, cfgstore.ErrFailedGettingUserCacheDir)
	assert.ErrorIs(t, err, cause)
}
