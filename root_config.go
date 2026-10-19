package cfgstore

type RootConfig interface {
	RootConfig()
	Normalize(NormalizeArgs) error
	// Merge combines other, the merge of every earlier layer, into the receiver,
	// a later layer, and returns the merged config. The result must be a non-nil
	// value of the receiver's own concrete type, and the receiver's values take
	// precedence unless a field's own rule says otherwise. Loading fails with
	// ErrFailedToLoadConfig when Merge returns nil or another concrete type.
	Merge(RootConfig) RootConfig
}
