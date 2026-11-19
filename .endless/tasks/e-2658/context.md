Raised by Mike on 2026-11-19 after go-cliutil's E-2648 landed, while discussing that a broken config file blocks an app's help. When a config file is corrupt and cannot load, the user needs a way to correct it, and go-cliutil apps should get that out of the box rather than each implementing it. go-cfgstore loads and saves config files and knows their paths, layers and defaults, so the store's side of recovery is decided here.

Mike's starting ideas, input to the discussion rather than decisions:
- Rename the bad file to .corrupt[N] or .backup[N] (N optional) and replace it with a default config.
- A file that will not parse differs from one that parses but fails validation; recovery might clean up specific parts instead of replacing the whole file.
- A streaming JSON parser might repair damage such as a missing trailing brace; that lives in a separate maybe-phase research task.

A go-cliutil brainstorm decides how an app exposes this to users and waits on this one.