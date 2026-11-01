# Create a config only when its file is missing

1. In `configStore.ensureConfig` (`config_store.go`), decide by whether the file exists, not by whether the loaded struct is zero: if `cs.Exists()` is false, call `createConfig`; otherwise load it with `loadConfigIfExists` and never write it, even when the struct comes back zero.
2. Keep `rc == nil` handled as today, without writing over an existing file.
3. Tests in `test/`:
   - an existing config file holding only keys the struct does not declare is left byte-for-byte unchanged after the config is ensured;
   - a missing config file is still created, as before;
   - an existing file with known keys still loads into the struct.
4. Build, vet and run the tests with `GOWORK=off`.
