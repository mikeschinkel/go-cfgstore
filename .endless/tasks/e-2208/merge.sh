#!/usr/bin/env bash
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This suite belongs to E-2208 and records what was true when E-2208
# landed. Edit it only if you ARE E-2208. If your change breaks an
# assertion here, leave it alone — see .endless/tasks/CLAUDE.md.
#
# Proves mergeRootConfigs returns the merge of every loaded layer.
# Runs from the repo root of the worktree under test.
#
# go-cfgstore has no _harness.sh, so this is a self-contained script
# emitting TAP. test/go.mod reaches go-testutil through a replace that only
# resolves from the main checkout, and the sibling go-testutil working tree
# may be mid-change, so the test module builds against go-testutil's
# committed HEAD through a scratch -modfile, leaving tracked files alone.

set -u
export GOEXPERIMENT=jsonv2

n=0
fails=0
ok() { n=$((n + 1)); echo "ok $n - $1"; }
not_ok() { n=$((n + 1)); fails=$((fails + 1)); echo "not ok $n - $1"; }
check() { desc=$1; shift; if "$@" >/dev/null 2>&1; then ok "$desc"; else not_ok "$desc"; fi; }
bail() { echo "Bail out! $1"; exit 2; }

root=$(git rev-parse --show-toplevel) || bail "not in a git repo"
cd "$root" || bail "cannot cd to $root"
main_checkout=$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")
testutil_repo="$main_checkout/../go-testutil"

scratch=$(mktemp -d) || bail "mktemp failed"
trap 'rm -rf "$scratch"' EXIT

echo "TAP version 13"

mkdir -p "$scratch/testutil" "$scratch/base"
git -C "$testutil_repo" archive HEAD | tar -x -C "$scratch/testutil" \
    || bail "cannot export go-testutil HEAD from $testutil_repo"
sed -e "s#=> \.\.\$#=> $root#" \
    -e "s#=> \.\./\.\./\.\./go-pkgs/go-testutil#=> $scratch/testutil#" \
    test/go.mod > "$scratch/test.mod"
cp test/go.sum "$scratch/test.sum"

# Fail fast: E-2208's own tests
out="$scratch/go-test.out"
if (cd test && go test -modfile="$scratch/test.mod" -race -count=1 -v \
        -run 'TestLoad_|TestGetAppCacheDir_' .) >"$out" 2>&1; then
    ok "E-2208's tests pass"
else
    not_ok "E-2208's tests pass"
    sed 's/^/# /' "$out"
    echo "1..$n"
    exit 1
fi
passed() { grep -qE "^[[:space:]]*--- PASS: $1 " "$out"; }

check "P1/P2: LoadConfig merges the CLI and project layers, project winning" \
    passed TestLoad_CLIAndProjectLayersMerge/LoadConfig
check "P1/P2: LoadDefaultConfig merges the CLI and project layers, project winning" \
    passed TestLoad_CLIAndProjectLayersMerge/LoadDefaultConfig
check "P1/P2: LoadConfigStores merges the CLI and project layers, project winning" \
    passed TestLoad_CLIAndProjectLayersMerge/LoadConfigStores
check "P3: each later Merge receives the merge of the earlier layers" \
    passed TestLoad_EachMergeReceivesTheMergeOfEarlierLayers
check "P4: a Merge returning another type fails, naming both types" \
    passed TestLoad_MergeReturningAnotherTypeFails
check "P5: an in-place Merge loads the same result as a copying one" \
    passed TestLoad_InPlaceMergeMatchesCopyingMerge
check "P6: a Merge returning a nil RootConfig fails, naming the layer" \
    passed "TestLoad_MergeReturningNilFails/nil_RootConfig"
check "P6: a Merge returning a nil config pointer fails, naming the layer" \
    passed "TestLoad_MergeReturningNilFails/nil_pointer_of_the_config_type"
check "P7: one DirType is returned as loaded, without Merge" \
    passed "TestLoad_SingleLayerIsReturnedWithoutMerge/one_DirType"
check "P7: a missing project file leaves one layer, returned without Merge" \
    passed "TestLoad_SingleLayerIsReturnedWithoutMerge/no_project_config_file"
check "GetAppCacheDir returns the user cache dir error" \
    passed TestGetAppCacheDir_ReturnsUserCacheDirError

# P13: the exported API matches the branch's base on main
base=$(git merge-base main HEAD) || bail "no merge base with main"
git archive "$base" | tar -x -C "$scratch/base"
api="$ENDLESS_VERIFY_DIR/api.go"
go run "$api" "$scratch/base" > "$scratch/api.base" 2>&1
go run "$api" "$root" > "$scratch/api.head" 2>&1
check "P13: the exported API is unchanged from main" \
    diff -q "$scratch/api.base" "$scratch/api.head"

# P12, P14, P15: the docs
merge_section=$(awk '/^#### Understanding Merge\(\)/{f=1} /^## Testing Support/{f=0} f' README.md)
has() { printf '%s\n' "$merge_section" | grep -qF -- "$1"; }
check "P12: RootConfig.Merge's doc states the receiver's own concrete type" \
    grep -qF "value of the receiver's own concrete type" root_config.go
check "P12: RootConfig.Merge's doc states the receiver's precedence" \
    grep -qF "receiver's values take" root_config.go
check "P12: the README states Merge returns the receiver's own concrete type" \
    has "as a value of its receiver's own concrete type"
check "P14: the README's wrapper declares a Merge returning the wrapper" \
    has "func (w *MyConfigWrapper) Merge(other cfgstore.RootConfig) cfgstore.RootConfig {"
for rule in "project-wins (the default)" "project-only" "global-only" "global-wins"; do
    check "P15: the README names the $rule rule" has "| $rule |"
done

echo "1..$n"
[ "$fails" -eq 0 ]
