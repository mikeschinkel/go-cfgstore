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
