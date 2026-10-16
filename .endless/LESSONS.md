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

### [2026-10-15] Work through an agreed list without asking
After E-2206 landed, I listed ten Endless gaps from the recovery and asked Mike 'Shall I work through the list now?'. He had already asked for the gaps to be closed (find a task, fold in, or file), so asking again only stalled the work; he asked why I waited. When Mike has asked for a body of work and I've just enumerated it, start on it.
- **Project**: go-cfgstore

### [2026-10-15] File the feature Mike described, not a near reading
While filing E-2357 from E-2206, Mike asked that a session's task stay in the status graph, dimmed, once it lands. I titled and summarized it as 'keep a landed task in the graph, dimmed', which read as reverting the earlier task that removes terminal tasks from the graph. What he wanted: dim it only once it has landed or otherwise stopped being a blocker. When filing a feature Mike describes, state the condition he named, especially where a looser wording would undo earlier work he wanted.
- **Project**: go-cfgstore

### [2026-10-15] Fold findings before filing them
After E-2206 landed, Mike asked me to close the gaps it surfaced by landing, folding or filing. I filed nine new tasks, and he pointed me to ED-1550: filing is the exception, several symptoms of one cause are one task, and a finding goes into an open task that owns its area. Six of the nine belonged in E-2063, E-2291, E-2349 or a sibling, and I superseded them into those. When closing a batch of findings, group them by cause and search for owning tasks first; file only what nothing owns.
- **Project**: go-cfgstore
