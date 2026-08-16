module github.com/wow-look-at-my/xml-validator/validator

go 1.24.7

require github.com/stretchr/testify v1.11.1

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/wow-look-at-my/go-containers v0.0.0-20260815235059-bc089f373e68 // go-toolchain:auto-branch
	github.com/wow-look-at-my/xml-validator/reader v0.0.0-20260816042650-2d019af64af6
)

// The reader module ships from this repository, so a build here uses the tree
// it is in rather than a published version of itself.
replace github.com/wow-look-at-my/xml-validator/reader => ../reader
