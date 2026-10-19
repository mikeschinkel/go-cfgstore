#!/usr/bin/env bash
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This script belongs to E-2208's verify suite and records what was true
# when E-2208 landed. Edit it only if you ARE E-2208. If your change
# breaks it, leave it alone — see .endless/tasks/CLAUDE.md.
#
# Checks, as TAP, that E-2208 left the exported API as it was on main (P13)
# and that the docs state the Merge contract (P12, P14, P15).

set -u
export GOEXPERIMENT=jsonv2

n=0
fails=0
ok() { n=$((n + 1)); echo "ok $n - $1"; }
not_ok() { n=$((n + 1)); fails=$((fails + 1)); echo "not ok $n - $1"; }
bail() { echo "Bail out! $1"; exit 2; }
# has reports whether haystack $3 contains needle $2.
has() { if [[ "$3" == *"$2"* ]]; then ok "$1"; else not_ok "$1"; echo "# missing: $2"; fi; }

root=$(git rev-parse --show-toplevel) || bail "not in a git repo"
cd "$root" || bail "cannot cd to $root"
scratch=$(mktemp -d) || bail "mktemp failed"
trap 'rm -rf "$scratch"' EXIT

echo "TAP version 13"

base=$(git merge-base main HEAD) || bail "no merge base with main"
mkdir -p "$scratch/base"
git archive "$base" | tar -x -C "$scratch/base"
api="$ENDLESS_VERIFY_DIR/api.go"
go run "$api" "$scratch/base" > "$scratch/api.base" 2>&1 || bail "api.go failed on main"
go run "$api" "$root" > "$scratch/api.head" 2>&1 || bail "api.go failed on HEAD"
if diff "$scratch/api.base" "$scratch/api.head" > "$scratch/api.diff"; then
    ok "P13: exported declarations match main"
else
    not_ok "P13: exported declarations match main"
    sed 's/^/# /' "$scratch/api.diff"
fi

doc=$(cat root_config.go)
merge_section=$(awk '/^#### Understanding Merge\(\)/{f=1} /^## Testing Support/{f=0} f' README.md)
has "P12: RootConfig.Merge's doc states the receiver's own concrete type" \
    "value of the receiver's own concrete type" "$doc"
has "P12: RootConfig.Merge's doc states the receiver's precedence" \
    "receiver's values take" "$doc"
has "P12: the README states Merge returns the receiver's own concrete type" \
    "as a value of its receiver's own concrete type" "$merge_section"
has "P14: the README's wrapper declares a Merge returning the wrapper" \
    "func (w *MyConfigWrapper) Merge(other cfgstore.RootConfig) cfgstore.RootConfig {" "$merge_section"
for rule in "project-wins (the default)" "project-only" "global-only" "global-wins"; do
    has "P15: the README names the $rule rule" "| $rule |" "$merge_section"
done

echo "1..$n"
[ "$fails" -eq 0 ]
