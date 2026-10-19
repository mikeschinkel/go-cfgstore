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

## Types that embed a config

A type that embeds a config and declares no `Merge` of its own inherits the
embedded type's, which returns the embedded type rather than the outer one, so
loading it through two or more layers fails the conversion. The README's
Wrapper Pattern shows exactly such a type: `MyConfigWrapper` declares
`RootConfig`, `Normalize` and the JSON methods but no `Merge`.

## Compatibility

No exported signature changes, so no implementer needs a code change to compile.
An implementer whose `Merge` already returns its receiver's type gains the
lower layers' values. One whose `Merge` returns another type now fails at load,
and only when two or more layers load.
