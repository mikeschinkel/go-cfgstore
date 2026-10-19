#!/usr/bin/env bash
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This script belongs to E-2208's verify suite and records what was true
# when E-2208 landed. Edit it only if you ARE E-2208. If your change
# breaks it, leave it alone — see .endless/tasks/CLAUDE.md.
#
# Runs E-2208's tests in test/ as `go test -json`, for verify.toml's gotest
# check. test/go.mod reaches go-testutil through a replace that only resolves
# from the main checkout, and the sibling go-testutil working tree may be
# mid-change, so the test module builds against go-testutil's committed HEAD
# through a scratch -modfile, leaving tracked files alone.

set -euo pipefail
export GOEXPERIMENT=jsonv2

root=$(git rev-parse --show-toplevel)
main_checkout=$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")
testutil_repo="$main_checkout/../go-testutil"

scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT

mkdir -p "$scratch/testutil"
git -C "$testutil_repo" archive HEAD | tar -x -C "$scratch/testutil"
sed -e "s#=> \.\.\$#=> $root#" \
    -e "s#=> \.\./\.\./\.\./go-pkgs/go-testutil#=> $scratch/testutil#" \
    "$root/test/go.mod" > "$scratch/test.mod"
cp "$root/test/go.sum" "$scratch/test.sum"

cd "$root/test"
go test -json -modfile="$scratch/test.mod" -race -count=1 \
    -run '^(TestLoad_.*|TestGetAppCacheDir_.*)$' .
