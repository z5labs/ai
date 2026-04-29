# RDATA: HINFO (TYPE = 13)

Host Information RDATA payload — two `<character-string>` values, CPU
and OS. Defined in RFC 1035 §3.3.2.

## Byte diagram

```
+-- CPU character-string ----+ +-- OS character-string ------+
| len |  CPU payload         | | len |  OS payload          |
+-----+----------------------+ +-----+----------------------+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | CPU | `<character-string>` naming the CPU type. See [`character-string.md`](character-string.md). |
| CPU end | variable | []byte | OS | `<character-string>` naming the operating system. See [`character-string.md`](character-string.md). |

The enclosing RR's RDLENGTH equals 1 + len(CPU) + 1 + len(OS).

## Bit fields

None.

## Variable-length fields

Two [`character-string`](character-string.md) values back-to-back.

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

Two [`character-string`](character-string.md).

## Versioning notes

Unchanged. HINFO is still defined but rarely used today.

## Ambiguities

None.
