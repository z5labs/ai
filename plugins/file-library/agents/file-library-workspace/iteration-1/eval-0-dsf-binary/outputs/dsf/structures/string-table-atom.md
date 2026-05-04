# StringTableAtom

A `[]byte` payload of NUL-terminated C-strings packed back-to-back. Used by
`PROP`, `TERT`, `OBJT`, `POLY`, `NETW`, `DEMN`.

## Encoding

The payload is `Size − 8` bytes long. Each string is followed by a single
`0x00` terminator. **The final string has its terminator too** (so an empty
string at the end is a single `0x00` byte). There is no count field — the
number of strings is determined by walking the buffer to its end.

The order of strings is significant; readers number them from `0` upward.
For `PROP`, the strings are paired (name, value, name, value, …) and the
total string count must be even.

Decoding pseudocode:

```
strings := []string{}
i := 0
while i < len(payload):
    j := indexOfByte(payload[i:], 0x00)
    if j < 0:
        error: unterminated final string
    strings = append(strings, payload[i:i+j])
    i = i + j + 1
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `[]byte` | Strings | NUL-terminated strings, including a NUL after the last string. |

## Variable-length fields

- **Length determination**: end-of-payload sentinel; each string ends at the
  next `0x00`.
- **Encoding**: ASCII (the spec uses ASCII for property names; values are
  free-form bytes; conventionally treated as UTF-8 for X-Plane property values).
- **Maximum length**: bounded only by the parent atom's `Size`.

## Ambiguities

> **Ambiguity:** The spec says "an empty string may be encoded via a single
> null character" and "a null character on the final string is necessary". A
> trailing single `0x00` is therefore both "empty final string" and "tail
> NUL of the previous non-empty string", which makes a payload of `…X00 00`
> ambiguous between `["…X", ""]` and a malformed extra terminator. The
> conventional decoder reading: every `0x00` ends the current string and
> starts a new one (so `…X00 00` decodes to `["…X", ""]`); writers must
> not emit consecutive `0x00`s unless they intentionally mean "an empty
> string follows".

> **Ambiguity:** The spec is silent on string encoding. Treat as bytes for
> raw I/O; reasonable convention is UTF-8 for X-Plane-defined `sim/*`
> properties.
