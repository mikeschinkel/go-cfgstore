# Analysis: choices made in planning (Mike, 2026-11-17)

- **Config reaches flags through a typed ApplyConfig hook in go-cliutil, not a name lookup in go-cfgstore.** Mike wants every config-to-flag mapping compile-checked and the RC to keep typed methods (Verbosity(), Input()). Rejected: a hand-written FlagValue switch, `flag:"name"` struct tags read by a go-cfgstore helper, tags read by go-cliutil, and matching JSON keys to flag names, all of which look values up by name at run time. The hook is go-cliutil work, added to E-2648 as items P46 to P49, and ED-1670 records it.
- **Slug and file come from AppInfo.** appinfo.AppInfo already carries ConfigSlug() and ConfigFile(), so the adapter needs no path arguments.
- **Nil Options to Normalize.** No consumer reads NormalizeArgs.Options.
- **Examples ship in this task** so Mike can judge the API's usability before release.
