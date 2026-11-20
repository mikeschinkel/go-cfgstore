package cfgstore

import (
	jsonv2 "encoding/json/v2"
	"errors"

	"github.com/mikeschinkel/go-dt"
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

var ErrNotValidConfigDirsAvailable = errors.New("not valid config dirs available")

type RootConfigArgs struct {
	DirTypes     []DirType
	Options      Options
	DirsProvider *DirsProvider
}

func makeRootConfig[RC any, PRC RootConfigPtr[RC]]() PRC {
	return PRC(new(RC))
}

// configLayer is one layer's file, read and decoded alone by the check pass.
type configLayer[PRC any] struct {
	dirType DirType
	fp      dt.Filepath
	data    []byte
	value   PRC
}

// LoadConfigStores loads root configuration from multiple config stores.
// Later stores in the DirTypes array take precedence over earlier ones.
// This is the low-level function that provides maximum control over the loading process.
//
// Loading runs in two passes:
//
//  1. The check pass reads each layer's file, creating a missing CLI or app
//     file and skipping a missing project file, and decodes it alone. It calls
//     Validate on every field implementing FieldValidator and then
//     RootConfig.Validate, and fails the load with every error from every file.
//  2. The layered pass calls ApplyDefaults on a new value, then for each layer
//     in order calls Merge, which decodes the layer's file over the value, and
//     Normalize. It stops at the first layer whose Merge or Normalize fails.
//
// Use this when you need to:
// - Pre-create ConfigStores for reuse
// - Test with custom ConfigStores
// - Have fine-grained control over store creation and configuration loading
//
// For simpler use cases, consider using LoadConfig, LoadCLIConfig, LoadProjectConfig,
// or LoadDefaultConfig instead.
func LoadConfigStores[RC any, PRC RootConfigPtr[RC]](stores *ConfigStores, args RootConfigArgs) (prc PRC, err error) {
	var layers []configLayer[PRC]

	if len(args.DirTypes) == 0 {
		args.DirTypes = []DirType{
			CLIConfigDirType,
			ProjectConfigDirType,
		}
	}

	layers, err = checkConfigLayers[RC, PRC](stores, args)
	if err != nil {
		goto end
	}
	if len(layers) == 0 {
		err = NewErr(ErrNotValidConfigDirsAvailable)
		goto end
	}
	prc, err = layerConfigs[RC, PRC](layers, args)
end:
	return prc, err
}

// checkConfigLayers is the check pass: it returns each layer whose file exists,
// in args.DirTypes order, or every error found in every file.
func checkConfigLayers[RC any, PRC RootConfigPtr[RC]](stores *ConfigStores, args RootConfigArgs) (layers []configLayer[PRC], err error) {
	var errs []error

	for _, dirType := range args.DirTypes {
		var layer configLayer[PRC]
		var lerr error

		store, ok := stores.StoreMap[dirType]
		if !ok {
			continue
		}
		cs := store.(*configStore)
		if args.DirsProvider != nil {
			cs.dirsProvider = args.DirsProvider
		}
		layer, ok, lerr = readConfigLayer[RC, PRC](cs, dirType, args)
		if lerr != nil {
			// TODO(E-2658): a file that fails to read or decode lands here, which
			// is where detecting, repairing or quarantining a corrupt config file
			// would hook in.
			errs = append(errs, NewErr(
				ErrFailedToEnsureConfig,
				"filepath", layer.fp,
				lerr,
			))
			continue
		}
		if !ok {
			continue
		}
		errs = append(errs, validateConfigLayer(layer, args.Options)...)
		layers = append(layers, layer)
	}
	err = CombineErrs(errs)
	if err != nil {
		layers = nil
	}
	return layers, err
}

// readConfigLayer reads and decodes dirType's file alone. It creates a missing
// file for any layer but the project layer, and reports ok=false for a missing
// project file.
func readConfigLayer[RC any, PRC RootConfigPtr[RC]](cs *configStore, dirType DirType, args RootConfigArgs) (layer configLayer[PRC], ok bool, err error) {
	layer.dirType = dirType
	layer.fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	if !cs.Exists() {
		if dirType == ProjectConfigDirType {
			goto end
		}
		err = cs.createConfig(makeRootConfig[RC, PRC](), dirType, args.Options, args.DirTypes)
		if err != nil {
			goto end
		}
	}
	layer.data, err = cs.Load()
	if err != nil {
		err = NewErr(ErrFailedToReadConfigFile, err)
		goto end
	}
	layer.value = makeRootConfig[RC, PRC]()
	err = jsonv2.Unmarshal(layer.data, layer.value)
	if err != nil {
		err = NewErr(ErrFailedToUnmarshalConfigFile, err)
		goto end
	}
	ok = true
end:
	return layer, ok, err
}

// validateConfigLayer validates a layer's file decoded alone: every
// FieldValidator field, then RootConfig.Validate.
func validateConfigLayer[PRC RootConfig](layer configLayer[PRC], opts Options) (errs []error) {
	errs = validateFields(layer.value, layer.fp)
	err := layer.value.Validate(ValidateArgs{
		DirType:    layer.dirType,
		SourceFile: layer.fp,
		Options:    opts,
	})
	if err != nil {
		errs = append(errs, NewErr(ErrInvalidConfig, "config_file", layer.fp, err))
	}
	return errs
}

// layerConfigs is the layered pass: defaults, then each layer's Merge and
// Normalize, in order.
func layerConfigs[RC any, PRC RootConfigPtr[RC]](layers []configLayer[PRC], args RootConfigArgs) (prc PRC, err error) {
	var layer configLayer[PRC]

	prc = makeRootConfig[RC, PRC]()
	err = prc.ApplyDefaults(DefaultsArgs{
		Options:  args.Options,
		DirTypes: args.DirTypes,
	})
	if err != nil {
		err = NewErr(ErrFailedToApplyDefaults, err)
		goto end
	}
	for _, layer = range layers {
		err = prc.Merge(MergeArgs[*RC]{
			DirType:    layer.dirType,
			SourceFile: layer.fp,
			Options:    args.Options,
			data:       layer.data,
			layer:      layer.value,
		})
		if err != nil {
			err = NewErr(ErrFailedToMergeConfig, err)
			goto end
		}
		err = prc.Normalize(NormalizeArgs{
			DirType:    layer.dirType,
			SourceFile: layer.fp,
			Options:    args.Options,
		})
		if err != nil {
			err = NewErr(ErrFailedToNormalizeConfig, err)
			goto end
		}
	}
end:
	if err != nil {
		err = WithErr(err,
			ErrFailedToLoadConfig,
			"dir_type", layer.dirType,
			"config_file", layer.fp,
		)
		prc = nil
	}
	return prc, err
}
