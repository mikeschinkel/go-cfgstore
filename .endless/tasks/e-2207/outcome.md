Dispositions, one per source item:
- TODO.md 1 (rename CLIConfig to UserConfig): still wanted, filed as E-2513 (maybe).
- TODO.md 2 (IsNil removed from ConfigStore): done; no IsNil remains in the code. Dropped.
- TODO.md 3 (drop Normalize for a post-load step): filed as E-2514 (research, maybe).
- PLAN.md, all parts (XDG, FHS, prefix layouts, open questions, phases): nothing implemented; filed as E-2515 (brainstorm, maybe) with the full PLAN.md text as its context.
Nothing went to the README or to Endless decisions: PLAN.md is a spec for unbuilt work, and no decision in it was settled.
PLAN.md lived only on wip/e2351-e2352; TODO.md was untracked in the main checkout, hidden from git status by .git/info/exclude (an earlier note saying it was absent was wrong). Mike deleted it from main before landing. The TODO.md line is removed from .git/info/exclude.
