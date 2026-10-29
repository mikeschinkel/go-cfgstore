#!/usr/bin/env bash
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This script belongs to E-2412's verify suite and records what was true
# when E-2412 landed. Edit it only if you ARE E-2412. If your change
# breaks it, leave it alone — see .endless/tasks/CLAUDE.md.
#
# Checks, as TAP, that wip/e-2412 holds main's snapshot (P1), that each
# kept group is its own commit (P4), and that the repo is on Go 1.27 with
# no jsonv2 experiment and builds, vets and lints clean (P10, P12–P14),
# and meets the go-pkgs package baseline (P15).

set -u
unset GOEXPERIMENT
export GOWORK=off

n=0
fails=0
ok() { n=$((n + 1)); echo "ok $n - $1"; }
not_ok() { n=$((n + 1)); fails=$((fails + 1)); echo "not ok $n - $1"; }
bail() { echo "Bail out! $1"; exit 2; }
# check runs "$@" and reports it under label $1, echoing output on failure.
check() {
    local label=$1; shift
    local out
    if out=$("$@" 2>&1); then ok "$label"; else not_ok "$label"; printf '%s\n' "$out" | tail -20 | sed 's/^/# /'; fi
}
eq() { if [[ "$2" == "$3" ]]; then ok "$1"; else not_ok "$1"; echo "# want: $2"; echo "# got:  $3"; fi; }

root=$(git rev-parse --show-toplevel) || bail "not in a git repo"
cd "$root" || bail "cannot cd to $root"

echo "TAP version 13"

# P1: the snapshot is one commit on main's HEAD at the time, touching
# exactly the three files main had modified.
eq "P1: wip/e-2412 is one WIP commit" "WIP snapshot of main working tree" \
    "$(git log -1 --format=%s wip/e-2412 2>&1)"
eq "P1: the snapshot touches cache.go, test/go.mod and test/go.sum" \
    "cache.go test/go.mod test/go.sum" \
    "$(git diff --name-only wip/e-2412~1 wip/e-2412 2>&1 | tr '\n' ' ' | sed 's/ $//')"

# P4: each kept group is its own E-2412 commit.
base=$(git merge-base main HEAD) || bail "no merge base with main"
subjects=$(git log --format=%s "$base"..HEAD)
for s in "return GetAppCacheDir's error instead of nil" \
         "bump test module to go-dt v0.6.1 to match go-testutil" \
         "require Go 1.27 and drop the jsonv2 experiment" \
         "re-sync doterr.go from go-doterr" \
         "bring go-cfgstore to the go-pkgs package baseline"; do
    if grep -qxF "E-2412: $s" <<<"$subjects"; then ok "P4: commit \"$s\""; else not_ok "P4: commit \"$s\""; fi
done

# P12: every module requires Go 1.27 and nothing sets the experiment.
for m in go.mod test/go.mod examples/basic_usage/go.mod; do
    eq "P12: $m says go 1.27.0" "go 1.27.0" "$(grep '^go ' "$m")"
done
if grep -rn 'GOEXPERIMENT' justfile .github >/dev/null; then
    not_ok "P12: justfile and workflows do not set GOEXPERIMENT"
else
    ok "P12: justfile and workflows do not set GOEXPERIMENT"
fi
eq "P12: CLAUDE.md says the modules require Go 1.27" "1" "$(grep -c 'The modules require Go 1.27' CLAUDE.md)"
eq "test/go.mod requires go-dt v0.6.1" "github.com/mikeschinkel/go-dt v0.6.1" \
    "$(grep -o 'github.com/mikeschinkel/go-dt v[0-9.]*' test/go.mod)"

# P10: every module builds and the root vets clean on Go 1.27.
check "P10: root module builds" go build ./...
check "P10: root module vets clean" go vet ./...
check "P10: examples/basic_usage builds" bash -c 'cd examples/basic_usage && go build -o /dev/null .'

# P13, P14: the pinned linter runs clean, doterr.go included.
eq "P13: justfile pins golangci-lint v2.15.0" "1" "$(grep -c 'golangci-lint@v2.15.0' justfile)"
scratch=$(mktemp -d) || bail "mktemp failed"
trap 'rm -rf "$scratch"' EXIT
GOBIN="$scratch/bin" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 \
    || bail "cannot install golangci-lint v2.15.0"
lint="$scratch/bin/golangci-lint"
check "P14: golangci-lint passes on the root module" "$lint" run ./... --timeout=5m
check "P15: golangci-lint passes on examples/basic_usage" \
    bash -c "cd examples/basic_usage && '$lint' run ./... --timeout=5m"

# test/go.mod's go-testutil replace resolves only from the main checkout,
# so lint the test module through a scratch go.work aimed at go-testutil's
# committed HEAD.
main_checkout=$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")
mkdir -p "$scratch/testutil"
git -C "$main_checkout/../go-testutil" archive HEAD | tar -x -C "$scratch/testutil" \
    || bail "cannot read go-testutil's HEAD"
printf 'go 1.27.0\n\nuse %s/test\n\nreplace github.com/mikeschinkel/go-testutil => %s/testutil\n' \
    "$root" "$scratch" > "$scratch/go.work"
check "P15: golangci-lint passes on the test module" \
    bash -c "cd test && GOWORK='$scratch/go.work' '$lint' run ./... --timeout=5m"

# P15: the rest of the package baseline.
eq "P15: release notes exclude .endless" "1" \
    "$(grep -c "git log .*':(exclude).endless'" .github/workflows/release.yml)"
eq "P15: /bin/ is gitignored" "1" "$(grep -cx '/bin/' .gitignore)"
eq "P15: nothing under bin/ is tracked" "" "$(git ls-files bin)"
eq "P15: the example's binary is gitignored" "1" "$(grep -cx '/examples/basic_usage/basic_usage' .gitignore)"
for p in .endless/worktrees/ .endless/tmp/ .endless/logs/ .endless/worktree.json .endless/worktree.lock; do
    eq "P15: $p is gitignored" "1" "$(grep -cxF "$p" .gitignore)"
done

echo "1..$n"
[[ $fails -eq 0 ]]
