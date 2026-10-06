// Package validator checks that a document is well-formed XML, by the rules of
// the version its declaration names, and checks it against. An XSD schema. The
// xml-validator command-line tool runs it, and a Go program can embed it.
//
// [Validate] checks well-formedness alone. [ValidateWithSchemaBytes] and
// [ValidateWithSchemaFile] also enforce a schema, and [ValidateWithSchemaResolver]
// takes a [SchemaResolver] for xs:import hints. [ParseTree], [ParseSchema] and
// [ValidateSchema] let a caller parse once and validate many times.
//
// Every failure is a *[Error], with the line and column of the construct.
package validator
