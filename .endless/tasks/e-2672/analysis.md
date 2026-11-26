Planned with Mike from the E-2628 session, 2026-11-25.

**How WriterLoggerArgs describes the Writer.** Once a ready-made Writer can be supplied, the fields that build one (IO, Quiet, Verbosity) must combine with it, and a plain bool cannot say "not provided". Options: (A) WriterLoggerArgs holds only a Writer and drops IO, Quiet and Verbosity, so callers build the Writer with cliutil.NewWriter; (B) keep the fields but make each able to say "not provided", with provided values building the Writer or being applied to a supplied one. Mike chose B.

**Verbosity 0.** cliutil names 0 NoVerbosity, a real level (nothing printed by Printf), so 0 cannot also mean "not provided"; Verbosity becomes a pointer like Quiet. go-cliutil's NewWriter panics on 0 while its flag validation accepts it; Mike, 2026-11-25: 0 stays a real level, folded into E-2673.

**DirsProvider.** Without one, the log lands in the real user config dir and TruncateLog empties it, so an in-process gomion test would wipe gomion's real log. Mike, 2026-11-24: optional.

**Not here.** createLogger's fallback for an unwritable log dir creates a temp file and then uses its path as a directory; E-2676 fixes it.
