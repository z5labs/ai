# 3. Objects

Source: RFC 8259, Section 4 ("Objects").

## 3.1 Structure

An object is represented as a pair of curly brackets surrounding zero or
more name/value pairs (called **members**). A name is a string. A single
colon comes after each name, separating the name from the value. A single
comma separates a value from a following name.

```abnf
object = begin-object [ member *( value-separator member ) ]
         end-object

member = string name-separator value
```

Allowed forms:

- Empty object: `{}`
- Single member: `{ "k": v }`
- Multiple members: `{ "k1": v1, "k2": v2 }`

Not allowed by the grammar:

- Trailing comma: `{ "k": v, }`
- Missing colon: `{ "k" v }`
- Non-string keys: `{ 1: v }`, `{ k: v }` (unquoted identifier)

Parser note: `member` requires a `string` name (Section 7), then the
`name-separator` (a `:` with optional surrounding whitespace), then any
`value`.

## 3.2 Uniqueness of names

The names within an object SHOULD be unique. (RFC 8259 uses SHOULD, not
MUST: duplicate keys are not strictly forbidden by the grammar.)

An object whose names are all unique is interoperable: all conforming
software receiving that object will agree on the name-value mapping. When
names within an object are not unique, the behavior of receiving software is
unpredictable. Observed behaviors include:

- Reporting only the last name/value pair.
- Reporting an error or failing to parse the object.
- Reporting all name/value pairs, including duplicates.

Implementer guidance: pick one of these behaviors deliberately and document
it. A typical choice for a Go decoder is "last write wins", matching
`encoding/json`.

## 3.3 Member ordering

JSON parsing libraries differ as to whether they expose member ordering to
calling software. Implementations whose behavior does not depend on member
ordering will be interoperable in the sense that they will not be affected
by these differences.

Implementer guidance: if your AST preserves insertion order (e.g. for
round-tripping or canonical printing), document that as an extension.
Consumers should not rely on a specific order being preserved across
implementations.
