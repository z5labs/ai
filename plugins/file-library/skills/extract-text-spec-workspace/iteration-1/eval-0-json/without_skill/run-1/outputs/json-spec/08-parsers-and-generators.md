# 8. Parsers and Generators

Source: RFC 8259, Sections 9 ("Parsers") and 10 ("Generators").

## 8.1 Parsers (Section 9)

A JSON parser transforms a JSON text into another representation.

- A parser MUST accept all texts that conform to the JSON grammar.
- A parser MAY accept non-JSON forms or extensions.

Implementations may set limits on:

- the size of texts accepted,
- the maximum depth of nesting,
- the range and precision of numbers,
- the length and character contents of strings.

Implementer guidance for Go:

- A maximum nesting depth is recommended to prevent stack overflow on
  pathological inputs (deeply nested arrays/objects). 1000-10000 is a
  common range.
- A maximum input size guards against memory exhaustion when reading from
  network sources.
- Document any extensions (comments, trailing commas, NaN/Infinity, etc.)
  clearly. They make output non-portable.

## 8.2 Generators (Section 10)

A JSON generator produces JSON text. The resulting text MUST strictly
conform to the JSON grammar.

This is the key constraint for the printer:

- It MUST emit only the seven value kinds.
- It MUST quote string values and use only the escapes defined in
  Section 7 (or the bare unescaped ranges).
- It MUST emit numbers in the form prescribed by Section 6 (no `Infinity`,
  no `NaN`, no leading `+`, no leading zeros on multi-digit ints, no hex,
  no underscores).
- It MUST emit only the lowercase literal names `true`, `false`, `null`.
- It MUST NOT emit a leading BOM on output destined for the network.
- It MUST emit object members separated by exactly one `,` (with optional
  whitespace), and each member as `string : value`.
- It MUST NOT emit a trailing comma after the last member or array
  element.

Implementer guidance for Go printer design:

- Two output modes are conventional: compact (no insignificant whitespace)
  and indented / pretty-printed (newlines + per-depth indent). Both are
  conforming as long as only the four allowed whitespace bytes appear in
  `ws` positions.
- For round-trip fidelity, preserve the original textual form of numbers
  rather than re-formatting from a decoded `float64`. Re-formatting is
  permitted but loses precision and may change the lexical form (e.g.
  `1e2` -> `100`).
- For string output, the minimal-escape strategy is to escape only the
  bytes the grammar requires (`"`, `\`, and U+0000-U+001F). Escaping `/`,
  non-ASCII characters, or U+007F is permitted but not required.
