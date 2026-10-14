`configStore.ensureConfig` (`config_store.go`, the `dtx.IsZero(rc)` check) loads the file, and if the
loaded struct is zero it calls `createConfig`, which writes
a fresh file over the existing one.

A file that exists but holds only keys the program's struct does not declare
loads as a zero struct, so it is overwritten and those keys are lost.

Seen during E-2240 in Endless: a verify run wrote
`{"tmux":{"socket_dir":"..."}}` to the user config. A pane then ran an older
installed `endless-go` that had no `Tmux` field, so it loaded the file as zero
and rewrote it, erasing `tmux.socket_dir` partway through the run.

This hits any consumer whenever two binary versions share one config file and
the newer one adds a setting: the older one destroys the file, not just
ignoring the new key. The likely fix is to create a config only when the file
is missing, never when it exists but decodes to zero.
