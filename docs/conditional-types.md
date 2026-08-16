# Conditional type assignment (xs:alternative)

An `xs:alternative` gives an element a type chosen per instance. The
alternatives are tried in schema order; the first whose `test` holds decides the
type, and an alternative with no `test` always holds, which is how a default is
written. When none holds, the element's declared `type` applies.

Code: `validator/schema_alternative.go` (parse, resolve, choose) and
`validator/schema_alternative_expr.go` (the test language). A ref particle
carries the alternatives of the declaration it names, so `<xs:element ref="v"/>`
behaves the same as the global `v`.

## The test language is the subset XSD requires

XSD 1.1 §3.12.6 defines a "required subset" of XPath 2.0 that a conforming
processor **must** accept, and that is what this implements:

```
Test        ::= OrExpr
OrExpr      ::= AndExpr ( 'or' AndExpr )*
AndExpr     ::= BooleanExpr ( 'and' BooleanExpr )*
BooleanExpr ::= '(' OrExpr ')' | BooleanFunction | ValueExpr ( Comparator ValueExpr )?
BooleanFunction ::= QName '(' OrExpr ')'
Comparator  ::= '=' | '!=' | '<' | '<=' | '>' | '>='
ValueExpr   ::= CastExpr | ConstructorFunction
CastExpr    ::= SimpleValue ( 'cast' 'as' QName '?'? )?
SimpleValue ::= AttrName | Literal
AttrName    ::= '@' NameTest
ConstructorFunction ::= QName '(' SimpleValue ')'
```

`and` binds tighter than `or`, as in XPath; parentheses override that. A test
outside this grammar is a hard error at schema-parse time. The spec permits
that: a processor "may but is not required to" support expressions beyond the
required subset, and a test this engine cannot evaluate would pick the wrong
type in silence.

What that rules out here: a function other than `not()` (only `fn:not` and the
constructor functions are required), a path step into element content, and a
cast to a user-defined type.

## How a comparison reads

An attribute is **untyped** during conditional type assignment -- deciding the
type is the whole point, so it is not known yet. The other operand decides how
the comparison reads:

- Against a **numeric literal** or a numeric cast, both sides are numbers, so
  `@a = 1` holds for `a="01"`.
- Against a **quoted literal** or another attribute, both sides are text, so
  `@a = '1'` does not hold for `a="01"`.
- An **absent** attribute makes any comparison false, `!=` included. Use
  `not(@a)` to test for absence.
- A **failed cast** -- `xs:int(@a)` where `a="six"` -- makes the test false
  rather than failing the document. The alternative simply is not selected,
  which is what XSD says about an error in a test.

## Effective boolean value

A bare value as a condition follows XPath's effective boolean value: an absent
attribute is false, an empty string is false, the number 0 is false, and
anything else is true.
