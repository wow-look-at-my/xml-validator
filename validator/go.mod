module github.com/wow-look-at-my/xml-validator/validator

go 1.26

require github.com/stretchr/testify v1.11.1

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/wow-look-at-my/go-containers v0.0.0-20260826161058-40a3d1ef3d41 // go-toolchain:auto-branch
	github.com/wow-look-at-my/xml-validator/reader v0.0.0-20260823201203-7268eddb2e3f // go-toolchain:auto-branch
)

// The reader module ships from this repository, so a build here uses the tree
// it is in rather than a published version of itself.
replace github.com/wow-look-at-my/xml-validator/reader => ../reader
