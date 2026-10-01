# Release go-cfgstore's dependency chain

go.mod resolves go-dt and go-dt/dtx through local `replace` directives, so
nobody outside this machine can build a module that imports go-cfgstore at its
current commit.

1. Walk the chain from the leaves: for go-dt, go-dt/dtx (and anything they
   replace in turn), check for uncommitted work, commit, push, and decide the
   version bump (major/minor/patch) from the changes since the last tag.
2. Show Mike each proposed tag and its reason before pushing it; tags on a
   public remote cannot be taken back cleanly.
3. Bump go-cfgstore's requires to the new tags, drop the `replace` lines, run
   `go mod tidy` and the tests, then commit, push and tag go-cfgstore.
4. Write down every step you did by hand, and why: this release is the first
   real case for gomion's dependency-release commands.
