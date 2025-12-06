module github.com/mikeschinkel/go-cfgstore/test

go 1.25.3

require (
	github.com/google/uuid v1.6.0
	github.com/mikeschinkel/go-cfgstore v0.4.0
	github.com/mikeschinkel/go-dt v0.3.3
	github.com/mikeschinkel/go-dt/appinfo v0.2.1
	github.com/mikeschinkel/go-dt/dtx v0.2.1
	github.com/mikeschinkel/go-testutil v0.2.1
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mikeschinkel/go-cliutil v0.3.0 // indirect
	github.com/mikeschinkel/go-logutil v0.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/mikeschinkel/go-cfgstore => ..
