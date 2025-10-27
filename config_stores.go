package cfgstore

type ConfigStoreMap map[DirType]ConfigStore

type ConfigStores struct {
	DirTypes []DirType
	StoreMap ConfigStoreMap
}

func NewConfigStores(info AppInfo) (css *ConfigStores) {
	if len(info.DirTypes) == 0 {
		info.DirTypes = []DirType{
			DotConfigDir,
			LocalConfigDir,
		}
	}
	css = &ConfigStores{
		DirTypes: info.DirTypes,
		StoreMap: make(ConfigStoreMap, len(info.DirTypes)),
	}
	for _, dirType := range info.DirTypes {
		css.StoreMap[dirType] = NewConfigStore(info).WithDirType(dirType)
	}
	return css
}

// LastStore returns the store identified by the last element in the DirTypes array
func (stores *ConfigStores) LastStore() (cs ConfigStore) {
	return stores.StoreMap[stores.DirTypes[len(stores.DirTypes)-1]].(*configStore)
}

// LoadRootConfig also specifying the config stores in a map to enable unit testing
func (stores *ConfigStores) LoadRootConfig(rc RootConfig, info AppInfo) (err error) {
	var cs *configStore
	var errs []error

	if len(info.DirTypes) == 0 {
		info.DirTypes = []DirType{
			DotConfigDir,
			LocalConfigDir,
		}
	}

	for _, store := range stores.StoreMap {
		cs := store.(*configStore)
		err = cs.ensureConfig(rc, info.Options)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	// TODO Merge them here instead of just returning LastStore

	cs = stores.LastStore().(*configStore)
	err = cs.loadConfigIfExists(rc, info.Options)
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
