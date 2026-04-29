# 1. Grammar Overview and Structural Tokens

Source: RFC 8259, Section 2 ("JSON Grammar").

## 1.1 Top-level

A JSON text is a sequence of tokens. The token set is: six structural
characters, strings, numbers, and three literal names (`true`, `false`,
`null`).

A JSON text is a serialized value. (Note: previous specifications constrained
top-level texts to be objects or arrays; RFC 8259 does not. Implementations
that produce only top-level objects/arrays remain interoperable, since all
implementations will accept those.)

```abnf
JSON-text = ws value ws
```

Tokenizer note: the top-level rule means leading and trailing insignificant
whitespace must be consumed before/after the single root value.

## 1.2 Structural characters

There are six structural characters. Each is wrapped in optional whitespace
on both sides in the grammar:

```abnf
begin-array     = ws %x5B ws  ; [ left square bracket
begin-object    = ws %x7B ws  ; { left curly bracket
end-array       = ws %x5D ws  ; ] right square bracket
end-object      = ws %x7D ws  ; } right curly bracket
name-separator  = ws %x3A ws  ; : colon
value-separator = ws %x2C ws  ; , comma
```

| Token             | Char | Code  | Use                             |
|-------------------|------|-------|---------------------------------|
| `begin-array`     | `[`  | U+005B| Start of array                  |
| `end-array`       | `]`  | U+005D| End of array                    |
| `begin-object`    | `{`  | U+007B| Start of object                 |
| `end-object`      | `}`  | U+007D| End of object                   |
| `name-separator`  | `:`  | U+003A| Between member name and value   |
| `value-separator` | `,`  | U+002C| Between members / array elements|

Tokenizer note: because every structural character production embeds `ws` on
both sides, whitespace handling can equivalently be performed once at the
tokenizer layer (skip-whitespace-then-emit-token) rather than re-applying
`ws` at every grammar boundary. The two approaches accept the same language.

## 1.3 Whitespace

Insignificant whitespace is allowed before or after any of the six
structural characters.

```abnf
ws = *(
        %x20 /              ; Space
        %x09 /              ; Horizontal tab
        %x0A /              ; Line feed or New line
        %x0D )              ; Carriage return
```

Only those four bytes count as JSON whitespace. Notably:

- No vertical tab, form feed, or NEL.
- No Unicode whitespace such as U+00A0 (NBSP) or U+2028 (LINE SEPARATOR).
- Whitespace is not permitted inside numbers, strings, or literal names
  (the `ws` rule is only invoked at the structural-character boundaries
  and at the top-level `JSON-text` rule).
