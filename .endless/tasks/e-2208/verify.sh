#!/usr/bin/env bash
# ── DO NOT EDIT ─────────────────────────────────────────────────────
# This suite belongs to E-2208 and records what was true when E-2208
# landed. Edit it only if you ARE E-2208. If your change breaks an
# assertion here, leave it alone — see .endless/tasks/CLAUDE.md.
source "$(dirname "${BASH_SOURCE[0]}")/../_harness.sh"

set -u
export GOEXPERIMENT=jsonv2

# Proves mergeRootConfigs returns the merge of every loaded layer.
#
# test/go.mod reaches go-testutil through a replace that only resolves from
# the main checkout, and the sibling go-testutil working tree may be
# mid-change, so the test module builds against go-testutil's committed HEAD
# through a scratch -modfile, leaving tracked files alone.

root=$(git rev-parse --show-toplevel) || setup_error "not in a git repo"
cd "$root" || setup_error "cannot cd to $root"
main_checkout=$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")
testutil_repo="$main_checkout/../go-testutil"

scratch=$(mktemp -d) || setup_error "mktemp failed"
trap 'rm -rf "$scratch"' EXIT

mkdir -p "$scratch/testutil" "$scratch/base"
git -C "$testutil_repo" archive HEAD | tar -x -C "$scratch/testutil" \
    || setup_error "cannot export go-testutil HEAD from $testutil_repo"
sed -e "s#=> \.\.\$#=> $root#" \
    -e "s#=> \.\./\.\./\.\./go-pkgs/go-testutil#=> $scratch/testutil#" \
    test/go.mod > "$scratch/test.mod"
cp test/go.sum "$scratch/test.sum"

section "E-2208's tests (fail fast)"
out="$scratch/go-test.out"
if (cd test && go test -modfile="$scratch/test.mod" -race -count=1 -v \
        -run 'TestLoad_|TestGetAppCacheDir_' .) >"$out" 2>&1; then
    report_pass "E-2208's tests pass"
else
    report_fail "E-2208's tests pass" "go test exits 0" "$(tail -40 "$out")"
    summary
fi

# check_test reports whether go test logged a PASS for the named (sub)test.
check_test() {
    if grep -qE "^[[:space:]]*--- PASS: $2 " "$out"; then
        report_pass "$1"
    else
        report_fail "$1" "--- PASS: $2" "$(grep -E -- "--- (FAIL|SKIP): $2 " "$out" || echo "not run")"
    fi
}

section "Loading merges every layer"
check_test "P1/P2: LoadConfig merges the CLI and project layers, project winning" \
    TestLoad_CLIAndProjectLayersMerge/LoadConfig
check_test "P1/P2: LoadDefaultConfig merges the CLI and project layers, project winning" \
    TestLoad_CLIAndProjectLayersMerge/LoadDefaultConfig
check_test "P1/P2: LoadConfigStores merges the CLI and project layers, project winning" \
    TestLoad_CLIAndProjectLayersMerge/LoadConfigStores
check_test "P3: each later Merge receives the merge of the earlier layers" \
    TestLoad_EachMergeReceivesTheMergeOfEarlierLayers
check_test "P5: an in-place Merge loads the same result as a copying one" \
    TestLoad_InPlaceMergeMatchesCopyingMerge
check_test "P7: one DirType is returned as loaded, without Merge" \
    TestLoad_SingleLayerIsReturnedWithoutMerge/one_DirType
check_test "P7: a missing project file leaves one layer, returned without Merge" \
    TestLoad_SingleLayerIsReturnedWithoutMerge/no_project_config_file

section "A Merge that breaks the contract fails loading"
check_test "P4: a Merge returning another type fails, naming both types" \
    TestLoad_MergeReturningAnotherTypeFails
check_test "P6: a Merge returning a nil RootConfig fails, naming the layer" \
    TestLoad_MergeReturningNilFails/nil_RootConfig
check_test "P6: a Merge returning a nil config pointer fails, naming the layer" \
    TestLoad_MergeReturningNilFails/nil_pointer_of_the_config_type

section "Folded-in lint fix"
check_test "P16: GetAppCacheDir returns the user cache dir error" \
    TestGetAppCacheDir_ReturnsUserCacheDirError

section "The exported API is unchanged"
base=$(git merge-base main HEAD) || setup_error "no merge base with main"
git archive "$base" | tar -x -C "$scratch/base"
api="$ENDLESS_VERIFY_DIR/api.go"
go run "$api" "$scratch/base" > "$scratch/api.base" 2>&1 || setup_error "api.go failed on main"
go run "$api" "$root" > "$scratch/api.head" 2>&1 || setup_error "api.go failed on HEAD"
assert_eq "P13: exported declarations match main" \
    "" "$(diff "$scratch/api.base" "$scratch/api.head")"

section "Docs"
merge_section=$(awk '/^#### Understanding Merge\(\)/{f=1} /^## Testing Support/{f=0} f' README.md)
assert_contains "P12: RootConfig.Merge's doc states the receiver's own concrete type" \
    "value of the receiver's own concrete type" "$(cat root_config.go)"
assert_contains "P12: RootConfig.Merge's doc states the receiver's precedence" \
    "receiver's values take" "$(cat root_config.go)"
assert_contains "P12: the README states Merge returns the receiver's own concrete type" \
    "as a value of its receiver's own concrete type" "$merge_section"
assert_contains "P14: the README's wrapper declares a Merge returning the wrapper" \
    "func (w *MyConfigWrapper) Merge(other cfgstore.RootConfig) cfgstore.RootConfig {" "$merge_section"
for rule in "project-wins (the default)" "project-only" "global-only" "global-wins"; do
    assert_contains "P15: the README names the $rule rule" "| $rule |" "$merge_section"
done

summary
