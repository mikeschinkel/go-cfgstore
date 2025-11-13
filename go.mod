module github.com/mikeschinkel/go-cfgstore

go 1.25.3

require (
	github.com/google/uuid v1.6.0
	github.com/mikeschinkel/go-dt v0.0.0-20251105233453-a7985f775567
	github.com/mikeschinkel/go-dt/appinfo v0.0.0-20251107040413-53a1559d69c5
	github.com/mikeschinkel/go-dt/de v0.0.0-20251107040413-53a1559d69c5
	github.com/mikeschinkel/go-dt/dtx v0.0.0-20251107040413-53a1559d69c5
	github.com/mikeschinkel/go-fsfix v0.1.0
	github.com/mikeschinkel/go-testutil v0.0.0-20251106130119-1afd5a012dc6
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/mikeschinkel/go-cliutil v0.0.0-20251105231813-8ce963ade5dd // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/mikeschinkel/go-cliutil => ../go-cliutil
	github.com/mikeschinkel/go-dt => ../go-dt
	github.com/mikeschinkel/go-dt/appinfo => ../go-dt/appinfo
	github.com/mikeschinkel/go-dt/de => ../go-dt/de
	github.com/mikeschinkel/go-dt/dtx => ../go-dt/dtx
	github.com/mikeschinkel/go-fsfix => ../go-fsfix
	github.com/mikeschinkel/go-testutil => ../go-testutil
)
