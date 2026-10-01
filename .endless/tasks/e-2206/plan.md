# Bring go-cfgstore up to current Endless usage

go-cfgstore is the first project other than Endless itself to be managed by
Endless. Every place it needs something Endless did not provide is a PRODUCT
defect: file it in the `endless` project, don't patch around it here.

## Steps

1. Run `endless guide` from the repo. Compare what this repo has with what a
   newly registered project should have: `.endless/config.json` (valid,
   committed, sensible `tracking`), `.endless/db-ledger/`, the `.gitignore`
   entries Endless needs, and an agent-instructions file (CLAUDE.md or
   AGENTS.md) that points at `endless guide` and states the repo's own rules.
2. Exercise the loop end to end on a throwaway task in this project: `task
   claim` creates a worktree here, the session binds, a commit lands with
   `endless worktree land` (this repo builds with its Makefile, not `just`).
   Record each step that failed or needed an Endless-specific workaround.
3. Remove or relocate anything left from pre-Endless habits, except PLAN.md and
   TODO.md, which have their own task.
4. File each friction point as a task in `endless` (link it here with
   `--relates-to`), with what happened and what a new user on another machine
   would have hit.

## Done when

A spawned session in go-cfgstore can claim, work and land a task with no
Endless-specific instructions beyond `endless guide`, or every gap that
prevents it is filed in `endless`.
