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

### [2026-10-30] Bring a task branch up to main by rebasing, not merging
In E-2208 I merged main into task/2208 to re-check it against E-2412's changes. endless worktree land then refused: the merge commit carried main's .endless/db-ledger/ lines, and a branch commit that modifies the ledger cannot land. Rebasing onto main (git rebase main) replays only the branch's own commits, drops the merge, and leaves the code tree identical; the SHA change sends the task back to unverified, so it needs verifying again.
- **Project**: go-cfgstore

### [2026-10-31] Never suggest dropping a worktree
After E-2206 landed in go-cfgstore, I told Mike to run 'endless worktree drop E-2206' (and E-2208's) as a next step. He said never suggest it: dropping a worktree pulls the floor out from under the session standing in it. The guide only says not to run land or drop without asking; Mike's rule is stronger: don't propose it at all.
- **Project**: go-cfgstore

### [2026-10-31] A session stewards its unspawned tasks
After E-2206 landed, I told Mike the session had nothing left and could be closed. He corrected me: a session's purpose includes stewarding the unspawned tasks in its session status (endless guide sessions, 'Focus, and a task's steward': the steward keeps a task in view so it gets spawned or rejected). A landed claimed task does not end the session's job; check session status for stewarded tasks first.
- **Project**: go-cfgstore

### [2026-10-31] adrs/ is deprecated; decisions live in Endless
While adding plan items to go-cfgstore's E-2207 (convert PLAN.md and TODO.md), I offered 'move it to adrs/ or the README' as a disposition. Mike: adrs/ is deprecated; design records are Endless decisions now, existing adrs/ will be converted later, and no new ADR files get added. He also said an item from PLAN.md or TODO.md is very unlikely to belong in a README.
- **Project**: go-cfgstore

### [2026-10-31] Another project's uncommitted or unreleased work belongs to a session for that project
In E-2209 the plan said to check each dependency for uncommitted work and commit it. I proposed committing and tagging go-logutil's uncommitted work myself. Mike corrected me: work in another project, such as go-logutil, is managed by a task in that project (here E-2376), never by the session releasing a dependent. Delegate it to that project and block my task on its task; do not commit, land or tag that project's in-flight work, even when my own plan appears to say so.
- **Project**: go-cfgstore

### [2026-10-31] A file hidden by .git/info/exclude is invisible to git status; check with ls before claiming it is absent
On E-2207 I told Mike TODO.md did not exist because git status showed nothing. An exclude entry hid it, and removing the entry made it untracked and block the land. Next time: ls the main checkout before saying a file is absent, and treat an exclude line as hiding a file that may still be there.
- **Project**: go-cfgstore

### [2026-10-31] "You" in Mike's questions means Claude in any session, not this one
While reconciling E-2351 in go-cfgstore, Mike asked whether Claude had removed dt.LogOnError without telling him. The session answered that it had not touched go-dt in this session and that Mike had renamed it, because git attributed the commit to Mike. Mike corrected: when he says "you" he means Claude in ANY session, not just the current one. A git author of Mike with no Co-Authored-By trailer does not establish that a person made the change; most commits in these repos come out of Claude sessions. So answer the question he asked: what changed, when, by which commit, and whether a Claude session plausibly did it, rather than disclaiming responsibility for the current session.
- **Project**: go-cfgstore

### [2026-10-31] Dependency bumps mean all deps to latest; plan version numbers are floors, not specs
On E-2351 (go-cfgstore) the plan named go-dt v0.7.0 and dtx v0.3.0. The session pinned exactly those, hit a compile break in go-cliutil v0.3.0, bumped only go-cliutil, and then stopped to file a reconcile question asking whether that one bump was allowed. Mike: he does not care about specific versions in go.mod, he wants ALL versions updated to the most recent, and treating a plan's version detail as authoritative instead of seeing the bigger picture (make the module build and current) turned a few hours of work into days. When a task touches go.mod, run go get -u across every module, tidy, and move on; do not ask whether a dependency bump that the build requires is permitted.
- **Project**: go-cfgstore

### [2026-10-31] Version numbers in a plan are not gospel; bump dependencies to their most recent releases
While doing E-2352 (TruncateLog in go-cfgstore) the plan named go-dt v0.7.0, and I treated that and the other pinned versions as exact requirements, filing reconcile discrepancies and asking Mike to approve each dependency bump the build forced. Mike said this was pedantic: another session wrote those numbers, and the bigger picture is what matters. When a task touches a module's dependencies, update all of them to their most recent releases and get on with the work, rather than asking for permission version by version.
- **Project**: go-cfgstore

### [2026-10-31] An unpublished dependency is a blocker to name, not a limitation to report
On E-2351 (go-cfgstore) the session reported that test/go.mod still reached go-testutil v0.3.0 through a replace because v0.3.0 was unpublished, and attributed publishing it to E-2209 as if that closed the matter. Mike: NO. Tell him which task lands and publishes it is a blocker, and he will land and publish it. When a dependency the task needs is unpublished, find the task that releases it, name it as a blocker (and record the blocked_by relation), so he can act, rather than leaving a replace in place and calling it someone else's scope.
- **Project**: go-cfgstore

### [2026-10-31] A replace that only resolves from main is a blocker to record, not a limitation to route around
On E-2352 (go-cfgstore) test/go.mod reached go-testutil v0.3.0 through a replace that resolves only from the main checkout, because v0.3.0 is unpublished. I built the test module through a scratch -modfile, called it a known limitation in the handoff, and attributed dropping the replace to E-2209. Mike: NO, it should be a blocker. The task that releases go-testutil (E-2511) blocks this task; record it with endless task block so he can land and publish it. A workaround that lets the suite run does not discharge the blocker, and reporting it as someone else's scope hides the thing he needs to act on.
- **Project**: go-cfgstore

### [2026-11-13] A passing user verify is the review approval; don't ask for it again
On go-cfgstore E-2551, plan item P3 said Mike reviews the code before the task leaves unverified. My handoff told him to review P3 separately, on top of the verify command, which confused him. His own endless task verify run is that review: a passing run moves the task out of unverified. When a plan item gates on Mike's review before unverified ends, hand him only the verify command and don't list the review as a separate step.
- **Project**: go-cfgstore
