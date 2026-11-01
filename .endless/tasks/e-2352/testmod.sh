#!/usr/bin/env bash
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This script belongs to E-2352's verify suite and records what was true
# when E-2352 landed. Edit it only if you ARE E-2352. If your change
# breaks it, leave it alone — see .endless/tasks/CLAUDE.md.
#
# Builds, vets or tests the test module with GOWORK=off. test/go.mod
# reaches go-testutil through a replace that only resolves from the main
# checkout, so the module is built against go-testutil's committed HEAD
# through a scratch -modfile, leaving tracked files alone.
#
#   testmod.sh own    # go test -json -race -run TestCreateWriterLogger_TruncateLog
#   testmod.sh test   # go test -json -race ./...   (for the gotest runner)
#   testmod.sh vet    # go build ./... && go vet ./...

set -euo pipefail
export GOWORK=off

mode=${1:?usage: testmod.sh own|test|vet}

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
case "$mode" in
    own)  go test -json -modfile="$scratch/test.mod" -race -count=1 \
              -run '^TestCreateWriterLogger_TruncateLog$' ./... ;;
    test) go test -json -modfile="$scratch/test.mod" -race -count=1 ./... ;;
    vet)  go build -modfile="$scratch/test.mod" ./... && go vet -modfile="$scratch/test.mod" ./... ;;
    *)    echo "unknown mode: $mode" >&2; exit 2 ;;
esac
