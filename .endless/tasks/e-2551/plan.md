# Plan: cfgstore passes a StdIO to the cliutil Writer

Mike's edit, moved from the main checkout onto this branch unchanged (commit 26542f7): `writer_logger.go` adds `IO dt.StdIO` to `WriterLoggerArgs` and passes it in `cliutil.WriterArgs`. With go-cliutil E-2550's nil default, a caller that leaves `IO` unset keeps the process streams.

`go.mod` requires go-cliutil v0.4.0 and go-dt v0.7.0, neither of which has the StdIO changes; the build only works through go.work. Once go-cliutil v0.5.1 resolves, raise the requires to go-cliutil v0.5.1 and go-dt v0.7.1 and update go.sum, so the Release workflow's `GOWORK=off` tests pass.

## After landing (Mike)

From the main checkout, land and push main:

```
endless worktree land E-2551 && git push origin main
```

Also from the main checkout, release v0.4.1 (previous tag v0.4.0). The empty assignments make `gh` ignore the token in the environment and use the keyring PAT:

```
GITHUB_TOKEN= GH_TOKEN= gh workflow run release.yml -f version=v0.4.1 -f description='Adds `IO dt.StdIO` to `cfgstore.WriterLoggerArgs` and passes it to `cliutil.NewWriter`, so applications inject their standard streams. Raises go-cliutil to v0.5.1 and go-dt to v0.7.1.'
```

gomion E-2540 waits for this tag to resolve from the module proxy.
