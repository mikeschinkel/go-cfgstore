# go-cfgstore

Run `endless guide` first. Tasks for this repo live in Endless, not in
PLAN.md or TODO.md.

- Build and test with `just` (`just --list`); every Go command needs
  `GOEXPERIMENT=jsonv2`, which the justfile sets.
- `doterr.go` is an embedded copy of go-doterr's `doterr.go`. Never edit it
  here: change go-doterr and re-sync with its `make sync`.
