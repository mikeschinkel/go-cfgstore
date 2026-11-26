#!/usr/bin/env bash
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This script belongs to E-2351's verify suite and records what was true
# when E-2351 landed. Edit it only if you ARE E-2351. If your change
# breaks it, leave it alone — see .endless/tasks/CLAUDE.md.
#
# P3: doterr.go is identical to go-doterr's doterr.go apart from the
# package clause. go-doterr is found beside the main checkout, as the
# plan's `../go-doterr/doterr.go` assumes.

set -euo pipefail

root=$(git rev-parse --show-toplevel)
main_checkout=$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")
upstream="$main_checkout/../go-doterr/doterr.go"

test -f "$upstream" || { echo "go-doterr's doterr.go not found at $upstream" >&2; exit 2; }
diff -u <(grep -v '^package ' "$root/doterr.go") <(grep -v '^package ' "$upstream")
grep -qx 'package cfgstore' "$root/doterr.go"
echo "doterr.go matches go-doterr apart from its package clause"
