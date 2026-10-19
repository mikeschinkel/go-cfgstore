## Why the result needs a runtime conversion

`Merge(RootConfig) RootConfig` returns the interface, because Go has no way for
an interface method to say "the implementing type itself". Go also has no
covariant returns: `func (c *T) Merge(RootConfig) *T` does not satisfy
`RootConfig`. So the compiler accepts any `RootConfig` as `Merge`'s result,
and handing the result back as the caller's `PRC` needs `rc.(PRC)` at runtime.
Go 1.27's generic methods do not change this: interface methods may not declare
type parameters, and a generic method cannot implement an interface method.

A conversion that fails is reported as an error, not papered over. Falling back
to the last layer would silently reproduce this bug. The mistake it reports is
caught by any consumer test that loads two layers.

## Rejected: a typed merge method on the generic constraint

A second method such as `MergeFrom(other *RC) *RC` on `RootConfigPtr` would
make a wrong return type a compile error. It was rejected for the package as a
whole. It duplicates `Merge` under a second name, changes every consumer's
signature, and leaves two merge concepts. Taking `Merge` off `RootConfig` would
also stop the interface, and the exported `RootConfigMap`, from describing a
layer that can be merged.

## Consumers

Every local consumer builds against this checkout through a `replace`, and none
needs a code change to compile.

- endless's `EndlessConfig.Merge` returns a `*EndlessConfig` copy, so it gains
  the CLI layer's values.
- gomion's `gomcfg`, and xmlui's `clicfg` and `mcpcfg`, have stub `Merge`
  methods but load the CLI layer alone, so `Merge` never runs.
- xmlui's `svrcfg` loads `*RootConfigV1Wrapper`, which embeds `RootConfigV1`
  and has no `Merge` of its own. The promoted `(*RootConfigV1).Merge` returns
  the inner `*RootConfigV1`, so a two-layer load of it would fail with the new
  error. None of its callers load two layers today; xmlui's own task gives the
  wrapper its own `Merge`.
- doctool, macmail-ext, go-stash and gru never load through the generic loader.
