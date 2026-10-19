## Why the merge method moves to the constraint

`Merge(RootConfig) RootConfig` returns the interface, so any type implementing
`RootConfig` is a legal return. Handing the result back as the caller's `PRC`
then needs a runtime conversion (`rc.(PRC)`), which fails for a `Merge` that
returns a different concrete type. xmlui's `svrcfg` is one: it loads
`*RootConfigV1Wrapper`, a struct that embeds `RootConfigV1` and has no `Merge`
of its own, so the promoted `(*RootConfigV1).Merge` returns the inner
`*RootConfigV1`. Requiring `MergeFrom(*RC) *RC` on the constraint makes that a
compile error, and leaves no conversion to fail.

Rejected: keeping `Merge` on `RootConfig` and returning an error when the
conversion fails, or falling back to the last layer. Both leave a wrong `Merge`
to be found at load time; the fallback silently reproduces this bug.

## Consumers

Every local consumer builds against this checkout through a `replace`, so the
change reaches them when it lands on main. The configs loaded through the
generic loader, and so affected: endless `internal/config` (`EndlessConfig`),
gomion `gomcfg`, and xmlui `clicfg`, `mcpcfg` and `svrcfg`. The user accepts
breaking the non-Endless ones, provided their code carries a comment saying how
to fix them. doctool, macmail-ext, go-stash and gru never load through the
generic loader, so the change does not touch them.

gomion's, `clicfg`'s and `mcpcfg`'s `Merge` are stubs that return the CLI
config, and they load the CLI layer alone, so `Merge` never runs for them today.
