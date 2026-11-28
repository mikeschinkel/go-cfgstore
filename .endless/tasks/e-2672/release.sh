#!/usr/bin/env bash
# Release script for E-2672's post-land item P11: tags go-cfgstore v0.5.0 by
# dispatching the repo's Release workflow (release.yml) on main.
#
# E-2672 changes WriterLoggerArgs.Quiet and .Verbosity to pointers, which
# breaks every caller that sets them, so the release bumps the minor version
# over v0.4.1. The script refuses to run unless v0.4.1 is still the latest
# remote tag, so it cannot tag v0.5.0 over a release made since, and unless
# origin's main contains E-2672's landed commit, since the workflow tags
# origin's main and main is pushed separately.
#
# It runs gh with an emptied environment (only HOME, PATH, USER and TERM), so
# no GH_TOKEN or GITHUB_TOKEN overrides the credentials gh keeps in the macOS
# keychain.
set -euo pipefail

REPO="mikeschinkel/go-cfgstore"
REF="main"
WORKFLOW="release.yml"
EXPECT_LATEST="v0.4.1"
VERSION="v0.5.0"
LANDED_COMMIT="9da328b2a"

# shellcheck disable=SC2016 # the backticks are Markdown, not shell
DESCRIPTION='`WriterLoggerArgs` gains an optional `Writer`, which `CreateWriterLogger` uses in place of building one, and an optional `DirsProvider`, which places the log under that provider'"'"'s CLI config dir. `Quiet` and `Verbosity` become `*bool` and `*cliutil.Verbosity`, a breaking change for callers that set them: nil means not provided, and a provided value is applied to a supplied `Writer` with `WithQuiet` or `WithVerbosity`. Setting both `Writer` and `IO` returns `ErrWriterAndIO`. When the logs dir cannot be written, the log now goes to a new temp directory instead of failing. Layered loads now decode each config file over one value, so `Normalize` defaults in a later layer no longer hide values set in an earlier one. Requires go-cliutil v0.7.0.'

# clean_gh runs gh without the caller's environment.
clean_gh() {
	env -i HOME="$HOME" PATH="$PATH" USER="${USER:-}" TERM="${TERM:-dumb}" gh "$@"
}

# release.yml pastes the description into `echo "..."` in a run step, so the
# shell would treat a backtick as command substitution and drop the code
# spans. Escaping \ " ` and $ makes the shell print them literally.
shell_escape() {
	printf '%s' "$1" | sed 's/[\\"`$]/\\&/g'
}

latest=$(git ls-remote --tags --refs "https://github.com/${REPO}.git" 'v*' |
	sed 's|.*refs/tags/||' |
	grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' |
	sort -V | tail -n 1)
if [[ "$latest" != "$EXPECT_LATEST" ]]; then
	echo "Latest tag on ${REPO} is ${latest:-none}, not ${EXPECT_LATEST}; not releasing ${VERSION}." >&2
	exit 1
fi

remote_main=$(git ls-remote "https://github.com/${REPO}.git" "refs/heads/${REF}" | cut -f1)
if ! git cat-file -e "${remote_main}^{commit}" 2>/dev/null ||
	! git merge-base --is-ancestor "$LANDED_COMMIT" "$remote_main"; then
	echo "origin's ${REF} (${remote_main:0:7}) does not contain E-2672's ${LANDED_COMMIT}; push ${REF} first." >&2
	exit 1
fi

echo "Latest tag: ${latest}"
echo "New tag:    ${VERSION}"
echo "Ref:        ${REF} @ ${remote_main:0:7}"
echo
read -r -p "Dispatch ${WORKFLOW} for ${VERSION}? [y/N] " answer
if [[ "$answer" != [yY] ]]; then
	echo "Not dispatched."
	exit 0
fi

clean_gh workflow run "$WORKFLOW" \
	--repo "$REPO" \
	--ref "$REF" \
	-f version="$VERSION" \
	-f description="$(shell_escape "$DESCRIPTION")"

# gh workflow run does not print the run it started, so wait for it to appear.
sleep 5
run_id=$(clean_gh run list --repo "$REPO" --workflow "$WORKFLOW" \
	--event workflow_dispatch --limit 1 --json databaseId --jq '.[0].databaseId')
clean_gh run watch "$run_id" --repo "$REPO" --exit-status

git ls-remote --tags "https://github.com/${REPO}.git" "$VERSION"
