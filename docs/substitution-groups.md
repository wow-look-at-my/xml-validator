# Substitution groups

A global element names a head with `substitutionGroup`, and may then appear anywhere a reference to that head does. The instance is validated against the SUBSTITUTE's own declaration, which is the whole point: the member carries its own type.

Code: `validator/schema_substitution.go` (group building, derivation check, matching) with the match hooks in `matchElement` and `validateAll`.

## What is enforced

- **Substitution is transitive.** A member of a member stands in for the head too, so the members are closed at resolution and each head holds the full list.
- **The member's own declaration validates the instance.** A `circle` standing in for a `shape` is checked against `circle`'s type, including the content its own extension adds.
- **`abstract="true"`** means only substitutes may appear. The abstract element itself is an error wherever it is used, including as the document root.
- **`block="substitution"`** on the head, or `blockDefault` on the schema, refuses substitution. The head accepts only itself, and a member used there is an unexpected element.
- **The member's type must derive from the head's.** The check walks the base chain of complex and simple types alike.
- **A cycle** in a `substitutionGroup` chain, and a `substitutionGroup` naming no global element, are both schema errors.

## Where substitution applies

Only where a content model REFERENCES a global element (`<xs:element ref="head"/>`). A local declaration that happens to share the head's name is a different element and nothing substitutes for it. This matches XSD, and it is also what keeps the feature from quietly changing the meaning of a schema that never asked for it.

In an `xs:all` group a substitute fills the slot of the element it stands in for. The occurrence counts are the particle's. Two members in a slot that allows one is still too many.

## The one place this stays quiet

`checkSubstitutionType` skips the derivation check when it cannot see a whole chain. That is an element with no declared type, or a type named but absent because its schema was imported without a `schemaLocation`. An unresolved type is unknown, not wrong. Rejecting it fails a schema that is fine. A head typed `xs:anyType` accepts any member, which is what that type means.
