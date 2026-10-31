#!/usr/bin/env bash
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This script belongs to E-2483's verify suite and records what was true
# when E-2483 landed. Edit it only if you ARE E-2483. If your change
# breaks it, leave it alone — see .endless/tasks/CLAUDE.md.
#
# P2: runs examples/basic_usage from the branch's merge-base with main
# (the pre-change source) and from the branch, each against a fresh HOME,
# and fails on any difference in output. The module is copied to scratch
# with its replace made absolute, so no tracked file is touched.

set -euo pipefail
export GOWORK=off

root=$(git rev-parse --show-toplevel)
base=$(git -C "$root" merge-base main HEAD)

scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT

mkdir -p "$scratch/mod"
cp -R "$root/examples/basic_usage/." "$scratch/mod/"
sed -i '' -e "s#=> \.\./\.\.\$#=> $root#" "$scratch/mod/go.mod"

run_example() {  # $1 = label; stdout+stderr of one run against a fresh HOME
    rm -rf "$scratch/home"
    mkdir -p "$scratch/home"
    (cd "$scratch/mod" && HOME="$scratch/home" XDG_CONFIG_HOME="$scratch/home/.config" go run . 2>&1) \
        > "$scratch/$1.txt"
}

git -C "$root" show "$base:examples/basic_usage/main.go" > "$scratch/mod/main.go"
run_example before
cp "$root/examples/basic_usage/main.go" "$scratch/mod/main.go"
run_example after

if ! diff -u "$scratch/before.txt" "$scratch/after.txt"; then
    echo "example output differs between $base and HEAD" >&2
    exit 1
fi
echo "example output identical across $(wc -l < "$scratch/after.txt" | tr -d ' ') lines"
