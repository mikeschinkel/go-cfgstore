package cfgstore

type RootConfig interface {
	RootConfig()
	Normalize(NormalizeArgs) error
}
