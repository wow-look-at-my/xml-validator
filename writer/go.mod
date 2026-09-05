module github.com/wow-look-at-my/xml-validator/writer

go 1.26

require (
	github.com/stretchr/testify v1.11.1
	github.com/wow-look-at-my/xml-validator/reader v0.0.0-20260905020842-b64bac5c026b // go-toolchain:auto-branch
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/wow-look-at-my/go-containers v0.0.0-20260826161058-40a3d1ef3d41 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// go-containers reaches this module through reader, so the require line is
// indirect and cannot carry a branch marker. The replace pins the version that
// actually reaches the build, and tracks the same branch reader does.
replace github.com/wow-look-at-my/go-containers => github.com/wow-look-at-my/go-containers v0.0.0-20260826161058-40a3d1ef3d41 // go-toolchain:auto-branch

// The reader module ships from this repository.
replace github.com/wow-look-at-my/xml-validator/reader => ../reader
