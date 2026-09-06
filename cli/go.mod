module github.com/wow-look-at-my/xml-validator/cli

go 1.26

require (
	github.com/spf13/cobra v1.10.2
	github.com/wow-look-at-my/xml-validator/validator v0.0.0-20260905064311-b54c8b4b3c13 // go-toolchain:auto-branch
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/wow-look-at-my/go-containers v0.0.0-20260826161058-40a3d1ef3d41 // indirect
	github.com/wow-look-at-my/xml-validator/reader v0.0.0-20260905064311-b54c8b4b3c13 // indirect; go-toolchain:auto-branch
)

// go-containers reaches this module through validator, so the require line is
// indirect and cannot carry a branch marker. The replace pins the version that
// actually reaches the build, and tracks the same branch validator does.
replace github.com/wow-look-at-my/go-containers => github.com/wow-look-at-my/go-containers v0.0.0-20260826161058-40a3d1ef3d41 // go-toolchain:auto-branch

// Both ship from this repository.
replace github.com/wow-look-at-my/xml-validator/validator => ../validator

replace github.com/wow-look-at-my/xml-validator/reader => ../reader

require github.com/stretchr/testify v1.11.1

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
