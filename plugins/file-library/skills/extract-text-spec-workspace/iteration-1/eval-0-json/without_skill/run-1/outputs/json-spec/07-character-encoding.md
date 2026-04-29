# 7. String and Character Issues

Source: RFC 8259, Section 8 ("String and Character Issues").

## 7.1 Character encoding (Section 8.1)

JSON text exchanged between systems that are not part of a closed ecosystem
MUST be encoded using UTF-8 (RFC 3629).

Previous specifications of JSON did not require UTF-8 when transmitting
JSON text, but the vast majority of implementations have chosen UTF-8 to
the extent that it is the only encoding that achieves interoperability.

### Byte order mark (BOM)

- Implementations MUST NOT add a byte order mark (U+FEFF) to the beginning
  of a network-transmitted JSON text.
- In the interest of interoperability, implementations that parse JSON
  texts MAY ignore the presence of a BOM rather than treating it as an
  error.

Implementer guidance for Go: a robust tokenizer can optionally strip a
leading 0xEF 0xBB 0xBF (UTF-8 BOM) from the input before scanning.

## 7.2 Unicode characters (Section 8.2)

When all the strings in a JSON text are composed entirely of Unicode
characters (however escaped), the text is interoperable: all conforming
implementations will agree on the contents of names and string values.

However, the ABNF allows strings to contain bit sequences that cannot
encode Unicode characters (see Section 6.5 of this extract: unpaired
surrogates from `\uXXXX` escapes). Behavior of receivers in that case is
unpredictable; observed effects include differing reported string lengths
and fatal runtime exceptions.

## 7.3 String comparison (Section 8.3)

Software is typically required to test names of object members for
equality. Implementations that:

1. transform the textual representation into sequences of Unicode code
   units, then
2. compare numerically, code unit by code unit,

are interoperable: they will agree on equality or inequality of any two
strings.

Implementations that compare strings with escape sequences left unconverted
may incorrectly find, for example, that `"a\\b"` and `"a\b"` are not
equal even though they encode the same three code points (`a`, `\`, `b`)
and must compare equal.

Implementer guidance for Go: when implementing object lookup or duplicate-
key detection, decode the escapes first and compare the resulting Go
strings. Comparing the raw on-the-wire bytes is incorrect.
