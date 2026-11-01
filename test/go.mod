module github.com/mikeschinkel/go-cfgstore/test

go 1.27.0

require (
	github.com/google/uuid v1.6.0
	github.com/mikeschinkel/go-cfgstore v0.4.0
	github.com/mikeschinkel/go-cliutil v0.4.0
	github.com/mikeschinkel/go-dt v0.7.0
	github.com/mikeschinkel/go-dt/appinfo v0.4.0
	github.com/mikeschinkel/go-dt/dtx v0.3.0
	github.com/mikeschinkel/go-testutil v0.3.0
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/mikeschinkel/go-logutil v0.2.1 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
)

replace github.com/mikeschinkel/go-cfgstore => ..

replace github.com/mikeschinkel/go-testutil => ../../../go-pkgs/go-testutil
