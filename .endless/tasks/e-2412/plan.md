# Plan: Land or drop uncommitted main changes

The main checkout of go-cfgstore has uncommitted work that predates this task. Other repos may build against that working tree through local `replace` directives. The goal is a clean main, with every change either landed through a task or deliberately dropped by the maintainer.

The changes at filing time (2026-10-19), excluding `.endless/`:

- Modified: test/go.mod, test/go.sum.
- test/go.mod bumps go-dt from v0.3.3 to v0.5.0 to match go-testutil's uncommitted main, which the test module reaches through its local go-testutil replace. With both in their current state the test module does not build from the main checkout (go reports a missing go.sum entry for go-dt).

## Steps

1. **Snapshot without touching main.** From this task's worktree, read the main checkout's full diff against HEAD, staged and unstaged together, including deletions, binaries and untracked non-`.endless/` files. Recreate it on a branch named `wip/<task-id>` cut from main's HEAD, and commit it there as one commit with message "WIP snapshot of main working tree". Then confirm the branch's tree matches main's working tree file for file. Don't use `git stash`; the stash stack is shared across sessions.
2. **Analyze.** Go through the snapshot and work out why each change was made, from the code, nearby commits and other repos that depend on it. Group the changes into coherent units, such as "API rename following go-dt v0.5.0", "new feature X" or "dependency bumps". Flag anything that looks unfinished, accidental, or like a generated or binary artifact.
3. **Report to the user before landing anything.** Show the grouping, what each group is for, and anything you couldn't explain, then let them adjust it and say what to drop.
4. **Land the agreed groups in this task.** Per ED-1550, this is one task, not one per group: commit each kept group to this task's branch as its own commit, its message saying what the group is for, then build and test the repo. Split a group into its own task only when it is too large or risky to review with the rest, and only with the user's agreement.
5. **Check the dependents.** Find sibling repos whose go.mod has a `replace` pointing at this checkout, and build each one against this task's branch, using a scratch GOWORK that points this module at the worktree. A dependent that fails gets a task in its own project, blocked by this one, with the user's agreement.
6. **Hand the user one command to clear main's working tree**, to run after this task lands and they confirm `wip/<task-id>` holds everything. This session doesn't write to the main checkout.

## Done when

Every change in the snapshot has landed through this task, been split out with the user's agreement, or is listed as "drop" with the user's agreement; the dependents build; and main is clean.
