# Turn PLAN.md and TODO.md into tasks, then delete them

Both files predate Endless's task database.

1. Read both. For each item, decide: already done (verify against the code;
   drop it), still wanted (file a task in go-cfgstore with a description and,
   where the file says how, a plan or analysis), or a design record rather than
   a to-do (move it to `adrs/` or the README).
2. Group items that share a root cause into one task. Lean toward fewer tasks.
3. Show Mike the proposed list before filing.
4. After filing, delete PLAN.md and TODO.md in the same commit that references
   the filed task ids.

5. In the same step, remove the `TODO.md` line from go-cfgstore's
   `.git/info/exclude`. It was added only to keep the file out of commits
   until this task converts it (from E-2349); with the file gone it has no
   purpose.
