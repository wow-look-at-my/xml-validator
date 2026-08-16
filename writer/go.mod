module github.com/wow-look-at-my/xml-validator/writer

go 1.24.7

require (
	github.com/stretchr/testify v1.11.1
	github.com/wow-look-at-my/xml-validator/reader v0.0.0-20260815191334-3096d3b2da43
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/wow-look-at-my/go-containers v0.0.0-20260815193622-200150bfb1c8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// The reader module ships from this repository.
replace github.com/wow-look-at-my/xml-validator/reader => ../reader
