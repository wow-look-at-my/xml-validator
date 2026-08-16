module github.com/wow-look-at-my/xml-validator/writer

go 1.24.7

require (
	github.com/stretchr/testify v1.11.1
	github.com/wow-look-at-my/xml-validator/reader v0.0.0-20260815191334-3096d3b2da43
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/wow-look-at-my/go-containers v0.0.0-20260815235059-bc089f373e68 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// go-containers reaches this module through reader, so the require line is
// indirect and cannot carry a branch marker. The replace pins the version that
// actually reaches the build, and tracks the same branch reader does.
replace github.com/wow-look-at-my/go-containers => github.com/wow-look-at-my/go-containers v0.0.0-20260815235059-bc089f373e68 // go-toolchain:auto-branch

// The reader module ships from this repository.
replace github.com/wow-look-at-my/xml-validator/reader => ../reader
