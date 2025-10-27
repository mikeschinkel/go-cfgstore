package cfgstore

import (
	"github.com/mikeschinkel/go-dt/appinfo"
)

type ConfigStoreMap map[DirType]ConfigStore

type ConfigStores struct {
	DirTypes []DirType
	StoreMap ConfigStoreMap
}
type ConfigStoreArgs struct {
	appinfo.AppInfo
	DirTypes []DirType
}

func NewConfigStores(args ConfigStoreArgs) (css *ConfigStores) {
	if len(args.DirTypes) == 0 {
		args.DirTypes = []DirType{
			DotConfigDir,
			LocalConfigDir,
		}
	}
	css = &ConfigStores{
		DirTypes: args.DirTypes,
		StoreMap: make(ConfigStoreMap, len(args.DirTypes)),
	}
	for _, dirType := range args.DirTypes {
		css.StoreMap[dirType] = NewConfigStore(args).WithDirType(dirType)
	}
	return css
}

// LastStore returns the store identified by the last element in the DirTypes array
func (stores *ConfigStores) LastStore() (cs ConfigStore) {
	return stores.StoreMap[stores.DirTypes[len(stores.DirTypes)-1]].(*configStore)
}

type RootConfigArgs struct {
	DirTypes []DirType
	Options  Options
}

// LoadRootConfig also specifying the config stores in a map to enable unit testing
func (stores *ConfigStores) LoadRootConfig(rc RootConfig, args RootConfigArgs) (err error) {
	var cs *configStore
	var errs []error

	if len(args.DirTypes) == 0 {
		args.DirTypes = []DirType{
			DotConfigDir,
			LocalConfigDir,
		}
	}

	for _, store := range stores.StoreMap {
		cs := store.(*configStore)
		err = cs.ensureConfig(rc, args.Options)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	// TODO Merge them here instead of just returning LastStore

	cs = stores.LastStore().(*configStore)
	err = cs.loadConfigIfExists(rc, args.Options)
	if err != nil {
		goto end
	}

end:
	if err != nil {
		fp, _ := cs.GetFilepath()
		err = WithErr(err,
			"filepath", fp,
		)
	}
	return err
}
