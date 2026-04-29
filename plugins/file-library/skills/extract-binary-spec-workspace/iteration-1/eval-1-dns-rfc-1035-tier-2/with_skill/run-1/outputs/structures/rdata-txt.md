# RDATA: TXT (TYPE = 16)

The Text RDATA payload — one or more `<character-string>` values
concatenated. Defined in RFC 1035 §3.3.14.

## Byte diagram

```
+-- character-string -+ +-- character-string -+ ... +-- character-string -+
| len |  payload     | | len |  payload     |     | len |  payload     |
+-----+--------------+ +-----+--------------+     +-----+--------------+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | Strings | One or more `<character-string>` values. The total wire length is given by RDLENGTH; strings are read sequentially until that many bytes have been consumed. |

Each string is encoded as documented in [`character-string.md`](character-string.md).

## Bit fields

None.

## Variable-length fields

### Strings

- **Length determination**: outer container is RDLENGTH; each contained
  string carries its own 1-byte length prefix.
- **Maximum length per string**: 255 octets.
- **Minimum**: at least one `<character-string>` per RFC 1035 §3.3.14.
- **Encoding**: opaque bytes. Application-level conventions (e.g. SPF,
  DKIM, key-value pairs) are layered on top by other RFCs.

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

Sequence of [`character-string`](character-string.md).

## Versioning notes

Unchanged.

## Ambiguities

> **Ambiguity:** RFC 1035 says "One or more <character-string>s" but
> doesn't specify a delimiter or how to interpret multi-string TXT
> records. Application protocols define their own concatenation rules
> (e.g. SPF concatenates with no separator, RFC 7208).
