Planned with Mike in gomion's E-2546 session, 2026-11-12.

**What exists.** go-cfgstore already has cstest.NewTestDirsProvider, which builds a DirsProvider under a caller-supplied TestRoot and reproduces the running OS's real layout beneath it (on macOS the user config dir is `<root>/Users/<name>/Library/Application Support`). Its own tests exercise it. So the constructor this task asked for exists in substance; the gap is ergonomics (it demands Username, ProjectDir and ConfigSlug) and the two commented-out With methods.

**Options weighed.** (A) use cstest as is; (B) add a flat one-root constructor in package cfgstore and keep cstest beside it; (C) B and delete cstest. I had listed as A's con that paths differ per OS and a test must know the fake username. Mike asked for (D): one implementation with OS-faithful layouts. That con dissolves once tests ask the provider for directories instead of computing paths, which is how a provider is meant to be used. D is A with defaults, so one implementation remains and nothing is lost.

**Why the With methods here.** gomion's gompkg builds a provider with only ProjectDirFunc set at every per-directory store, leaving the other functions nil. DefaultDirsProvider().WithProjectDir(dir) gives the same pinning with the rest filled in.

**No I/O.** Consistent with Mike's constructor pattern: construction is pure; the store creates directories on save.
