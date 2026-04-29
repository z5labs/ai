# 2. Values and Literal Names

Source: RFC 8259, Section 3 ("Values").

## 2.1 The seven value kinds

A JSON value MUST be one of:

- an **object** (Section 4)
- an **array** (Section 5)
- a **number** (Section 6)
- a **string** (Section 7)
- one of the three literal names: `false`, `null`, `true`

```abnf
value = false / null / true / object / array / number / string
```

Parser note: the first non-whitespace byte is sufficient to dispatch to one
of these alternatives:

| First byte         | Value kind |
|--------------------|------------|
| `{` (U+007B)       | object     |
| `[` (U+005B)       | array      |
| `"` (U+0022)       | string     |
| `-` (U+002D), `0`-`9` | number  |
| `t` (U+0074)       | `true`     |
| `f` (U+0066)       | `false`    |
| `n` (U+006E)       | `null`     |

## 2.2 Literal names

The literal names MUST be lowercase. No other literal names are allowed.

```abnf
false = %x66.61.6c.73.65   ; false
null  = %x6e.75.6c.6c      ; null
true  = %x74.72.75.65      ; true
```

Tokenizer note: each literal is a fixed byte sequence. Implementations
typically match the keyword bytes exactly and then reject if the next byte
would extend an identifier (e.g. `truex` is invalid because the `x` is not a
valid continuation in any production starting with `t`).

Variant casings such as `True`, `TRUE`, `False`, `FALSE`, `Null`, `NULL` are
not valid JSON.
