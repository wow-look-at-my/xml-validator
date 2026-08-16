# Conditional type assignment (xs:alternative)

An `xs:alternative` gives an element a type chosen per instance. The
alternatives are tried in schema order; the first whose `test` holds decides the
type, and an alternative with no `test` always holds, which is how a default is
written. When none holds, the element's declared `type` applies.

Code: `validator/schema_alternative.go`. A ref particle carries the
alternatives of the declaration it names, so `<xs:element ref="v"/>` behaves the
same as the global `v`.

## The supported test language

```
test ::= term ( ('and' | 'or') term )*
term ::= '@name'
       | '@name' ('=' | '!=') "'literal'"
       | 'not(' term ')'
```

`@name` alone is an attribute-presence test. A prefixed name resolves through
the schema's own namespace declarations; an unprefixed one matches the local
name in any namespace, the same rule the identity-constraint XPaths use and for
the same reason (see `docs/identity-constraints.md`).

Anything else is a hard error at schema-parse time: a function call, an element
or descendant test, a numeric comparison, an unquoted right-hand side. A test
this engine cannot evaluate would otherwise pick the wrong type in silence, and
the wrong type validates the wrong things.

## Mixing and with or

Rejected rather than guessed at. XPath binds `and` tighter than `or`, so
`@a='1' and @b='2' or @c='3'` means `(@a and @b) or @c` -- and a reader should
not have to know that to know what a schema means. Write one operator per test,
or split the alternative in two.

## Comparison against a missing attribute

An absent attribute makes both `=` and `!=` false, which is what XPath does with
an empty sequence on one side. Use `not(@a)` to test for absence.
