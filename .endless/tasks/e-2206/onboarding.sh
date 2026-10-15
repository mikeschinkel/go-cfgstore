#!/bin/sh
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This suite belongs to E-2206 and records what was true when E-2206
# landed. Edit it only if you ARE E-2206. If your change breaks an
# assertion here, leave it alone — see .endless/tasks/CLAUDE.md.
#
# Proves go-cfgstore carries what a project managed by Endless needs.
# Runs from the repo root of the worktree under test.

n=0
fails=0
ok() { n=$((n + 1)); echo "ok $n - $1"; }
not_ok() { n=$((n + 1)); fails=$((fails + 1)); echo "not ok $n - $1"; }
check() { desc=$1; shift; if "$@" >/dev/null 2>&1; then ok "$desc"; else not_ok "$desc"; fi; }

cd "$(git rev-parse --show-toplevel)" || { echo "Bail out! not in a git repo"; exit 2; }

echo "TAP version 13"

check "config.json is tracked" git ls-files --error-unmatch .endless/config.json
check "config.json names the project go-cfgstore" \
    python3 -c 'import json,sys; sys.exit(json.load(open(".endless/config.json"))["name"] != "go-cfgstore")'
check "config.json disables the minimizer and optimizer" \
    python3 -c 'import json,sys; m=json.load(open(".endless/config.json")).get("minimizer",{}); sys.exit(not (m.get("enabled") is False and m.get("optimizer") is False))'
for entry in .endless/worktrees/ .endless/tmp/ .endless/worktree.json .endless/worktree.lock; do
    check ".gitignore ignores $entry" grep -qxF "$entry" .gitignore
done
check ".endless/tasks/CLAUDE.md is tracked" git ls-files --error-unmatch .endless/tasks/CLAUDE.md
check "CLAUDE.md is tracked" git ls-files --error-unmatch CLAUDE.md
check "CLAUDE.md points agents at endless guide" grep -q 'endless guide' CLAUDE.md
check "justfile is tracked" git ls-files --error-unmatch justfile
check "Makefile is gone" test ! -e Makefile
check ".squire/ is gone" test ! -e .squire
check "just build succeeds" just build

echo "1..$n"
[ "$fails" -eq 0 ]
