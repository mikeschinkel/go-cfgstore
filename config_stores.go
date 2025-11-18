package cfgstore

type ConfigStoreMap map[DirType]ConfigStore

type RootConfigMap map[DirType]RootConfig

type ConfigStores struct {
	DirTypes []DirType
	StoreMap ConfigStoreMap
	//GetwdFunc func() (dt.DirPath, error)
}

func (stores *ConfigStores) AppConfigStore() (cs ConfigStore) {
	cs, _ = stores.StoreMap[AppConfigDir]
	return cs
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
	DirTypes     []DirType
	DirsProvider *DirsProvider
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
	DirTypes     []DirType
	Options      Options
	DirsProvider *DirsProvider
}

type RootConfigPtr[RC any] interface {
	RootConfig
	*RC
}

func makeRootConfig[RC any, PRC RootConfigPtr[RC]]() PRC {
	return PRC(new(RC))
}

// LoadRootConfig also specifying the config stores in a map to enable unit testing
func LoadRootConfig[RC any, PRC RootConfigPtr[RC]](stores *ConfigStores, args RootConfigArgs) (prc PRC, err error) {
	var cs *configStore
	var errs []error

	if len(args.DirTypes) == 0 {
		args.DirTypes = []DirType{
			CLIConfigDir,
			ProjectConfigDir,
		}
	}

	rcMap := make(map[DirType]PRC, len(args.DirTypes))
	for dirType, store := range stores.StoreMap {
		cs = store.(*configStore)
		if args.DirsProvider != nil {
			cs.dirsProvider = args.DirsProvider
		}
		var tmpPRC PRC
		tmpPRC = makeRootConfig[RC, PRC]()
		err = cs.ensureConfig(tmpPRC, dirType, args.Options)
		if err != nil {
			fp, _ := cs.GetFilepath()
			errs = append(errs, NewErr(
				ErrFailedToEnsureConfig,
				"filepath", fp,
				err,
			))
			continue
		}
		rcMap[dirType] = tmpPRC
	}
	err = CombineErrs(errs)
	if err != nil {
		goto end
	}
	{
		dirType := args.DirTypes[0]
		rc := RootConfig(rcMap[dirType])
		for i := 1; i < len(args.DirTypes); i++ {
			dirType = args.DirTypes[i]
			rc = rcMap[dirType].Merge(rc)
		}
		prc = rcMap[dirType]
	}
end:
	return prc, err
}
