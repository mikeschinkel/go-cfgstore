package cfgstore

import (
	jsonv2 "encoding/json/v2"

	"github.com/mikeschinkel/go-dt"
)

// RootConfig is the app's root config type. Loading builds one value from every
// layer's file: it calls ApplyDefaults on a new value, then, for each layer in
// order, Merge decodes that layer's file over the value and Normalize runs on the
// result. Before that, Validate runs on each file decoded alone. *RC also needs
// Merge; see RootConfigPtr.
type RootConfig interface {
	RootConfig()

	// ApplyDefaults sets the config's defaults. It runs on a new, empty value
	// before any layer's file is decoded, so a default survives every layer whose
	// file omits its key, and any layer that sets the key replaces it. This is
	// the place for defaults; Normalize is not. ApplyDefaults also runs before
	// go-cfgstore creates a missing CLI config file, so the file holds the
	// defaults.
	ApplyDefaults(DefaultsArgs) error

	// Normalize runs after each layer's Merge, on the value built from every
	// layer decoded so far, with that layer's DirType and SourceFile. It cleans
	// values, such as lowercasing a name or stamping a schema version. It runs
	// once per layer on the same value, so it must be idempotent: running it
	// twice gives the same result as running it once. It must not set defaults,
	// which belong in ApplyDefaults, and must not validate, which belongs in
	// Validate or a field type's FieldValidator.
	Normalize(NormalizeArgs) error

	// Validate checks values that span fields. It runs once per file on that
	// file decoded alone, after go-cfgstore has called Validate on every field
	// implementing FieldValidator, so a file's problems are reported against
	// that file. Loading reports every error from every file at once.
	Validate(ValidateArgs) error
}

// RootConfigPtr constrains the pointer to an app's root config type RC.
//
// Merge combines one layer's file into c, the value built from every earlier
// layer. Its common body is one line:
//
//	func (c *Config) Merge(args cfgstore.MergeArgs[*Config]) error {
//	    return args.Decode(c)
//	}
//
// Decode leaves each field whose key the file omits unchanged, merges maps and
// replaces slices; AppendSlice and ReplaceMap change that per field. A rule the
// decoder cannot express can read the layer's file decoded alone from
// args.Layer and combine it with c by hand.
type RootConfigPtr[RC any] interface {
	RootConfig
	Merge(MergeArgs[*RC]) error
	*RC
}

// DefaultsArgs is what ApplyDefaults receives.
type DefaultsArgs struct {
	Options  Options
	DirTypes []DirType
}

// ValidateArgs is what RootConfig.Validate receives: the file being checked.
type ValidateArgs struct {
	DirType    DirType
	SourceFile dt.Filepath
	Options    Options
}

// FieldValidator is a field type that checks its own value. go-cfgstore calls
// Validate on every field implementing it, in nested structs, slice elements
// and map values too, on each file decoded alone.
//
// A field type should accept any raw value when it is decoded and report the
// problem from Validate: a decode error stops decoding of that file, which hides
// every later problem in it, while Validate errors are all collected.
type FieldValidator interface {
	Validate() error
}

// MergeArgs is what Merge receives for one layer.
type MergeArgs[PRC any] struct {
	DirType    DirType
	SourceFile dt.Filepath
	Options    Options
	data       []byte
	layer      PRC
}

// Decode decodes the layer's file over c with encoding/json/v2. A field whose
// key the file omits is left unchanged, a map's entries merge into c's map, a
// slice replaces c's slice, and a key set to null resets its field to zero.
func (args MergeArgs[PRC]) Decode(c PRC) (err error) {
	err = jsonv2.Unmarshal(args.data, c)
	if err != nil {
		err = NewErr(ErrFailedToUnmarshalConfigFile, err)
	}
	return err
}

// Layer returns the layer's file decoded alone into a new value, for a Merge
// that applies a rule Decode cannot express.
func (args MergeArgs[PRC]) Layer() (PRC, error) {
	return args.layer, nil
}
