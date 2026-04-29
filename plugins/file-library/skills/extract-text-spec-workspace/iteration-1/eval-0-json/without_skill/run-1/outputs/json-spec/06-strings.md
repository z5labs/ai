# 6. Strings

Source: RFC 8259, Section 7 ("Strings").

## 6.1 Structure

A string begins and ends with quotation marks (`"`, U+0022). All Unicode
characters may appear inside the quotation marks, except for the characters
that MUST be escaped:

- quotation mark (U+0022)
- reverse solidus / backslash (U+005C)
- the control characters U+0000 through U+001F

```abnf
string = quotation-mark *char quotation-mark

char = unescaped /
    escape (
        %x22 /          ; "    quotation mark  U+0022
        %x5C /          ; \    reverse solidus U+005C
        %x2F /          ; /    solidus         U+002F
        %x62 /          ; b    backspace       U+0008
        %x66 /          ; f    form feed       U+000C
        %x6E /          ; n    line feed       U+000A
        %x72 /          ; r    carriage return U+000D
        %x74 /          ; t    tab             U+0009
        %x75 4HEXDIG )  ; uXXXX                U+XXXX

escape         = %x5C              ; \
quotation-mark = %x22              ; "
unescaped      = %x20-21 / %x23-5B / %x5D-10FFFF
```

The `unescaped` ranges, expressed as exclusions, are: any code point from
U+0020 through U+10FFFF, **except** U+0022 (the quote) and U+005C (the
backslash). Anything below U+0020 (the C0 control range) must be escaped.

## 6.2 The two-character escape sequences

| Sequence | Code point | Name             |
|----------|------------|------------------|
| `\"`     | U+0022     | quotation mark   |
| `\\`     | U+005C     | reverse solidus  |
| `\/`     | U+002F     | solidus          |
| `\b`     | U+0008     | backspace        |
| `\f`     | U+000C     | form feed        |
| `\n`     | U+000A     | line feed        |
| `\r`     | U+000D     | carriage return  |
| `\t`     | U+0009     | tab              |

`\/` is permitted but never required: the forward slash may also appear
unescaped. Printers commonly emit it bare.

The escape set is closed: any other `\X` sequence (e.g. `\a`, `\v`, `\x41`,
`\0`, single-quote `\'`) is invalid JSON.

## 6.3 The `\uXXXX` escape

Any character may be escaped. If the character is in the Basic Multilingual
Plane (U+0000 through U+FFFF), it may be represented as a six-character
sequence: a reverse solidus, then lowercase `u`, then four hexadecimal
digits encoding the code point.

The hexadecimal letters `A`-`F` may be uppercase or lowercase. Note that
the literal `u` is lowercase only; `\U005C` is not valid.

Example: a string containing only a single backslash may be written as
`"\"`, or more compactly as `"\\"`.

## 6.4 Surrogate pairs (non-BMP characters)

To escape an extended character that is not in the BMP, the character is
represented as a 12-character sequence encoding the UTF-16 surrogate pair:
two `\uXXXX` escapes back-to-back.

Example: U+1D11E (G clef) is `"𝄞"`.

The high-surrogate range is U+D800 through U+DBFF; the low-surrogate range
is U+DC00 through U+DFFF. A correct decoder reassembles the pair into a
single non-BMP code point.

## 6.5 Unpaired surrogates

The ABNF allows member names and string values to contain bit sequences
that cannot encode Unicode characters; for example, `"\uDEAD"` (a single
unpaired surrogate). Such instances have been observed (e.g. when a library
truncates a UTF-16 string without checking whether the truncation split a
surrogate pair).

The behavior of software that receives JSON texts containing such values is
unpredictable: implementations may return different string lengths or even
suffer fatal runtime exceptions.

Implementer guidance for Go: Go strings are conventionally UTF-8. Two
common policies for unpaired surrogates:

1. **Reject** during parse (return an error).
2. **Replace** with U+FFFD (REPLACEMENT CHARACTER) and continue.

Whichever policy you pick, document it. Note that fully round-trip-safe
encoders should preserve the original `\uDEAD` literal sequences in the
AST rather than decoding them to a Go string, because UTF-8 cannot
represent unpaired surrogates.

## 6.6 Tokenizer summary

While scanning a string token, after the opening `"`:

- Bytes in `unescaped` (i.e. U+0020-U+0021, U+0023-U+005B, U+005D-U+10FFFF
  in UTF-8 form) are appended verbatim.
- A byte `\` (U+005C) starts an escape; the next byte chooses one of the
  eight short escapes or `u` for `\uXXXX`.
- A byte `"` (U+0022) ends the string.
- A byte < U+0020 inside an unescaped run is an error (raw control
  character).
- Premature end-of-input before the closing `"` is an error.
