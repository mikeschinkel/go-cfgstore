package cfgstore

import (
	"errors"

	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/dtx"
)

type ConfigStoreMap map[DirType]ConfigStore

type RootConfigMap map[DirType]RootConfig

type ConfigStores struct {
	DirTypes []DirType
	StoreMap ConfigStoreMap
	//GetwdFunc func() (dt.DirPath, error)
}

func (stores *ConfigStores) AppConfigStore() (cs ConfigStore) {
	return stores.StoreMap[AppConfigDirType]
}
func (stores *ConfigStores) CLIConfigStore() (cs ConfigStore) {
	return stores.StoreMap[CLIConfigDirType]
}
func (stores *ConfigStores) ProjectConfigStore() (cs ConfigStore) {
	return stores.StoreMap[ProjectConfigDirType]
}

type ConfigStoresArgs struct {
	ConfigStoreArgs
	DirTypes     []DirType
	DirsProvider *DirsProvider
}

func NewConfigStores(args ConfigStoresArgs) (css *ConfigStores) {
	if len(args.DirTypes) == 0 {
		args.DirTypes = []DirType{
			CLIConfigDirType,
			ProjectConfigDirType,
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
			CLIConfigDirType,
			ProjectConfigDirType,
		}
	}

	rcMap := make(map[DirType]PRC, len(args.DirTypes))
	for dirType, store := range stores.StoreMap {
		cs = store.(*configStore)
		if args.DirsProvider != nil {
			cs.dirsProvider = args.DirsProvider
		}
		tmpPRC := makeRootConfig[RC, PRC]()
		switch dirType {
		case ProjectConfigDirType:
			err = cs.loadConfigIfExists(tmpPRC, dirType, args.Options)
			if err == nil && (tmpPRC == nil || dtx.IsZero(tmpPRC)) {
				rcMap[dirType] = nil
				continue
			}
		default:
			err = cs.ensureConfig(tmpPRC, dirType, args.Options)
		}
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

	prc, err = mergeRootConfigs[RC, PRC](rcMap, args)

end:
	return prc, err
}

var ErrNotValidConfigDirsAvailable = errors.New("not valid config dirs available")
var ErrDirTypeNotAssignAfterMerge = errors.New("dirType not assigned after merge")

// mergeRootConfigs also specifying the config stores in a map to enable unit testing
func mergeRootConfigs[RC any, PRC RootConfigPtr[RC]](rcMap map[DirType]PRC, args RootConfigArgs) (prc PRC, err error) {

	var rc RootConfig
	var dirType DirType
	var start, cnt int

	// First, count the valid configs
	for _, typ := range args.DirTypes {
		if rcMap[typ] == nil {
			continue
		}
		cnt++
	}

	// Then find the first valid config
	for i, typ := range args.DirTypes {
		if rcMap[typ] == nil {
			continue
		}
		// This is our starting config
		prc = rcMap[typ]
		rc = RootConfig(rcMap[typ])
		// Skip over this config
		start = i + 1
		break
	}
	if rc == nil {
		// If we did not find any valid configs, return an error
		err = NewErr(ErrNotValidConfigDirsAvailable)
		goto end
	}
	if cnt <= 1 {
		// If we only found one valid config this is our prc
		goto end
	}
	// Now merge the second config with the next, until we have merged all. OTOH, if
	// there was only one valid config, this loop will not end up merging anything.
	for i := start; i < len(args.DirTypes); i++ {
		typ := args.DirTypes[i]
		if rcMap[typ] == nil {
			continue
		}
		rc = rcMap[typ].Merge(rc)
		// Capture the key for the last merged config
		dirType = typ
	}
	if dirType == UnspecifiedConfigDirType {
		// This should never happen - indicates a logic bug
		err = NewErr(
			dt.ErrUnexpectedError,
			dt.ErrInternalError,
			ErrDirTypeNotAssignAfterMerge,
			"config_count", cnt,
		)
		goto end
	}
	// The last merged config will be the config we return
	prc = rcMap[dirType]
end:
	return prc, err
}
