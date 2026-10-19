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

### [2026-10-19] Prefer verify.toml manifests for task verify suites now that the runner has landed; installing the harness in a project is E-2349's job, not a task's
E-2208: go-cfgstore had no _harness.sh, so I wrote a verify.toml suite, then on Mike's suggestion copied the harness and switched to verify.sh. Once I pointed out the manifest runner had landed (E-1602/E-1618; E-2206's verify.toml had passed through it), Mike said: 'If yes, I did not realize we could start using them. We should if we can.' Use a gotest check (with a command override when needed) plus sh/tap checks.
- **Project**: go-cfgstore

### [2026-10-19] In GetAppCacheDir Mike fixed a dropped error with named results (cachePath, err), var declarations and = assignments, not just by changing the return
E-2208: lint flagged cache.go's GetAppCacheDir, which wrapped UserCacheDirFunc's error and then returned nil at end:. I changed only nil to err. Mike: 'the first is more than that', and gave his fix: named results (cachePath dt.DirPath, err error), var cacheDir dt.DirPath declared up front, cacheDir, err = ... with =, and return cachePath, err. He said this about that one function; ask before treating it as a general goto-end rule.
- **Project**: go-cfgstore

### [2026-10-19] When another repo's state gets in the way of a task, file the upstream chain with blockers and expect Mike to want the current task blocked too, for urgency
E-2208: go-cfgstore's test module didn't build because go-testutil, go-cliutil and go-dt had uncommitted main changes. I advised against blocking E-2208, since the fix didn't need them. Mike: 'cleaning those up should be fast, and I need urgency for that cleanup which blocking would provide.' He wants tasks filed up the dependency chain until one lands on its own, with blockers, and existing ones pointed out. He said this about a chain of fast cleanups; ask when the cleanup isn't fast.
- **Project**: go-cfgstore

### [2026-10-19] A repo's main-checkout cleanup is one task that lands its own changes, not a task that files a task per change group (ED-1550); only tasks across projects stay separate
E-2208: E-2375, E-2374 and the two I filed (E-2411, E-2412) said to file one task per change group. Mike: 'It seems like process theater to have so many tasks for such as small amount of work. See ED-1550 (but note that tasks that cross projects cannot be combined.)' Plan such cleanups to commit each group on the task's own branch, and split a group out only when it is too large or risky to review with the rest, with Mike's agreement.
- **Project**: go-cfgstore

### [2026-10-19] Every directory with its own go.mod is a separate Go module that is versioned independently; plan, report and release per module, not per repo
E-2208 follow-up: planning the go-dt cleanup task (E-2375), Mike pointed out that go-dt and go-dt/dtx are versioned independently because each has its own go.mod, and: 'anywhere you see a go.mod in a dir, that dir represents a module and needs to be independently versioned.' go-dt has nine such modules. Group changes, version bumps and tags (dir-prefixed, e.g. dtx/v0.2.1) by module.
- **Project**: go-cfgstore
