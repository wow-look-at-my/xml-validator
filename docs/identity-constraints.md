# Identity constraints

`xs:key`, `xs:keyref`, and `xs:unique` are evaluated. A constraint that this
engine cannot express is a hard error at schema-parse time, so a constraint
either runs or says it cannot. It never passes silently.

Code: `validator/schema_identity.go` (compile, select, evaluate) and
`compileElementConstraints` in `validator/schema_resolve.go` (registration and
`refer` resolution).

## The supported XPath subset

A selector and a field take the subset XSD defines for them:

```
path      ::= alternative ( '|' alternative )*
alternative ::= ( './/' | './' )? step ( '/' step )*   |   '.'
step      ::= NCName | prefix ':' NCName | '*'
field     ::= path, whose last step may be '@name', '@prefix:name', or '@*'
```

Everything else is rejected by name: a predicate (`provider[@id='x']`), a
function (`count(...)`), an axis (`child::provider`), a parent step (`..`), and
an empty step (`a//b`). An attribute step in a selector is rejected too -- a
selector selects elements.

## Two deliberate deviations from XPath 1.0

Both exist because the alternative is a constraint that quietly matches nothing
and therefore can never fail.

1. **An unprefixed name matches its local name in any namespace.** XPath reads
   an unprefixed name as the no-namespace name. Under a schema with a target
   namespace and `elementFormDefault="qualified"`, that selects nothing, and the
   constraint becomes vacuous with no diagnostic. A prefixed name still matches
   its namespace exactly.
2. **Values compare in the field's declared value space, not lexically.** `01`
   and `1` are one `xs:integer` key; `true` and `1` are one `xs:boolean`. Only
   `xs:string` keeps its whitespace -- every other type compares collapsed. The
   declared type comes from the declaration that matched the node during the
   structural pass; a field with no declared type compares as a string.

## Evaluation order and scope

Identity constraints run in a second pass, after the structural and type pass
reports no errors. On a document that does not match the schema, the fields
point at nodes that never matched a declaration, so every constraint would pile
noise on top of the real failure.

The pass walks the document top-down. Entering an element that declares
constraints, it builds one key table per `xs:key` and `xs:unique`, then checks
that element's `xs:keyref` declarations. A keyref resolves against the nearest
enclosing element that evaluated the key it names, which is what lets a keyref
on a child element refer to a key declared on its parent.

## Field rules

- `xs:key` requires every field to select a value. A target missing one is an
  error.
- `xs:unique` and `xs:keyref` skip a target whose field selects nothing: that
  node is not a target at all.
- A field selecting more than one node is an error, per XSD.

## Schema-level rules

- A constraint name is schema-wide. Declaring one twice is an error.
- An `xs:keyref` must name an `xs:key` or `xs:unique` that exists. A keyref
  pointing at nothing used to be a rule that could never fail.
- A `refer` QName resolves through the schema's own prefix declarations, with a
  local-name fallback for a schema that never mentions namespaces.
