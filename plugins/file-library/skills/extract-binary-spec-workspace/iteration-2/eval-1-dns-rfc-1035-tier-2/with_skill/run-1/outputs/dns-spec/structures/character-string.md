# CharacterString

A length-prefixed string of bytes used inside RDATA payloads (HINFO CPU/OS, TXT TXT-DATA). Defined in RFC 1035 §3.3.

This is a single-prefix variable-length structure; no compression and no nesting.

## Layout

```
+-+-+-+-+-+-+-+-+
|     LEN       |   1 octet, LEN ∈ [0, 255]
+-+-+-+-+-+-+-+-+
|    DATA[0]    |
+-+-+-+-+-+-+-+-+
       ...           LEN bytes total
+-+-+-+-+-+-+-+-+
|   DATA[LEN-1] |
+-+-+-+-+-+-+-+-+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | uint8 | Length | Number of data bytes that follow. |
| 1 | Length | []byte | Data | Opaque payload. RFC 1035 §3.3: "treated as binary information". |

## Variable-length fields

- **Length determination:** length-prefix.
- **Length prefix format:** one octet, exclusive of itself. So total on-wire size of a `CharacterString` is `1 + Length` octets.
- **Maximum length:** RFC 1035 §3.3 says a `<character-string>` "can be up to 256 characters in length (including the length octet)" — i.e., a maximum payload of 255 bytes (the length octet itself encodes 0..255).
- **Encoding:** binary-safe; not required to be ASCII.

## Ambiguities

> **Ambiguity:** RFC 1035 §3.3 says the limit is "256 characters in length (including the length octet)" — equivalent to 255 payload bytes. Some implementations have historically read this as 256 payload bytes by ignoring the parenthetical. The length octet's range (0..255) settles the issue: max payload is 255.
