package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
)

type ConfigStoreMap map[DirType]ConfigStore

type ConfigStores struct {
	DirTypes  []DirType
	StoreMap  ConfigStoreMap
	GetwdFunc func() (dt.DirPath, error)
}

func (stores *ConfigStores) CLIConfigStore() (cs ConfigStore) {
	cs, _ = stores.StoreMap[CLIConfigDir]
	return cs
}
func (stores *ConfigStores) ProjectConfigStore() (cs ConfigStore) {
	cs, _ = stores.StoreMap[ProjectConfigDir]
	return cs

}

type ConfigStoresArgs struct {
	ConfigStoreArgs
	DirTypes []DirType
}

func NewConfigStores(args ConfigStoresArgs) (css *ConfigStores) {
	if len(args.DirTypes) == 0 {
		args.DirTypes = []DirType{
			CLIConfigDir,
			ProjectConfigDir,
		}
	}
	css = &ConfigStores{
		DirTypes: args.DirTypes,
		StoreMap: make(ConfigStoreMap, len(args.DirTypes)),
	}
	for _, dirType := range args.DirTypes {
		css.StoreMap[dirType] = NewConfigStore(dirType, args.ConfigStoreArgs)
	}
	return css
}

// LastStore returns the store identified by the last element in the DirTypes array
func (stores *ConfigStores) LastStore() (cs ConfigStore) {
	if len(stores.DirTypes) == 0 {
		panic("cfgstore.ConfigStores.LastStore(): No stores found")
	}
	return stores.StoreMap[stores.DirTypes[len(stores.DirTypes)-1]].(*configStore)
}

// FirstStore returns the store identified by the first element in the DirTypes array
func (stores *ConfigStores) FirstStore() (cs ConfigStore) {
	if len(stores.DirTypes) == 0 {
		panic("cfgstore.ConfigStores.FirstStore(): No stores found")
	}
	return stores.StoreMap[stores.DirTypes[0]].(*configStore)
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
			CLIConfigDir,
			ProjectConfigDir,
		}
	}

	for _, store := range stores.StoreMap {
		cs = store.(*configStore)
		err = cs.ensureConfig(rc, args.Options)
		if err != nil {
			fp, _ := cs.GetFilepath()
			errs = append(errs, NewErr(
				ErrFailedToEnsureConfig,
				"filepath", fp,
				err,
			))
		}
	}
	err = CombineErrs(errs)
	if err != nil {
		goto end
	}

	// TODO Merge them here instead of just returning LastStore

	cs = stores.LastStore().(*configStore)
	err = cs.loadConfigIfExists(rc, args.Options)
	if err != nil {
		err = NewErr(
			ErrFailedToLoadConfig,
			err,
		)
		goto end
	}

end:
	return err
}
