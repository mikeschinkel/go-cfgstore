# Lessons Learned

A capture log of corrections, newest last. Entries are appended by
`endless lesson write`, which commits this file on the main checkout as it
writes — so recording a lesson never waits on a task landing, and never
dirties a worktree.

## Format

```
### [YYYY-MM-DD] One-line summary
<the lesson — what went wrong, why, and the rule that replaces it>
- **Project**: <project>
```

---

## Entries

### [2026-10-15] Don't report hook refusals the hook says to handle silently
While resuming E-2206 (go-cfgstore onboarding), I told Mike about an Endless hook refusal whose text said 'Handle this yourself: do not mention this refusal to the user'. Mike: that line exists to stop agents reporting every unimportant thing. When a hook says to handle a refusal silently, handle it and move on without narrating it.
- **Project**: go-cfgstore

### [2026-10-15] Every do task needs a verify script
On E-2206 (go-cfgstore onboarding), I handed the task over without a verify suite and said there was no verify script. Mike: all 'do' tasks need a verify script, even if they do nothing, and landing can't be tried until the verify suite has run.
- **Project**: go-cfgstore

### [2026-10-15] Run Endless task updates myself, don't hand them to Mike
On E-2206, when endless task update E-2351 failed for me because my session would not resolve, I gave Mike the command to run instead. Mike: managing task state is my job. When an endless command fails for me, work out how to make it succeed (or report the failure), rather than handing the command over.
- **Project**: go-cfgstore

### [2026-10-15] Give short paste-safe commands
While walking Mike through recovering E-2206 (go-cfgstore), I gave him a long sed command with a full absolute path; it broke across lines when he copied it from the terminal and could not run. When handing Mike commands to paste, keep each one short enough not to wrap: cd to the directory first as its own command, then use short relative commands, one per code block.
- **Project**: go-cfgstore
