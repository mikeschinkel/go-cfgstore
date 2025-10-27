module github.com/mikeschinkel/go-cfgstore

go 1.25.3

require (
	github.com/google/uuid v1.6.0
	github.com/mikeschinkel/go-dt v0.0.0-20251027170931-0f47f0479185
	github.com/mikeschinkel/go-fsfix v0.1.0
	github.com/mikeschinkel/go-testutil v0.0.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mikeschinkel/go-cliutil v0.0.0-20251027170801-82399064d27f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/mikeschinkel/go-cliutil => ../go-cliutil
	github.com/mikeschinkel/go-dt => ../go-dt
	github.com/mikeschinkel/go-fsfix => ../go-fsfix
	github.com/mikeschinkel/go-testutil => ../go-testutil
)
