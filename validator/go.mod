module github.com/wow-look-at-my/xml-validator/validator

go 1.26

require github.com/stretchr/testify v1.11.1

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/wow-look-at-my/go-containers v0.0.0-20260820210621-2e1261867045 // go-toolchain:auto-branch
	github.com/wow-look-at-my/xml-validator/reader v0.0.0-20260823072327-30bfdcf6eb24 // go-toolchain:auto-branch
)

// The reader module ships from this repository, so a build here uses the tree
// it is in rather than a published version of itself.
replace github.com/wow-look-at-my/xml-validator/reader => ../reader
