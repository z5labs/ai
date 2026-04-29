# RDATA: MINFO (TYPE = 14)

Mailbox Information RDATA payload — two domain names. Defined in RFC
1035 §3.3.7.

## Byte diagram

```
+-- RMAILBX (domain-name) ----------------------+
|                                               |
+-- EMAILBX (domain-name) ----------------------+
|                                               |
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | RMailbox | Domain-name of the mailbox responsible for the mailing list / mailbox specified by the owner name. See [`domain-name.md`](domain-name.md). |
| RMailbox end | variable | []byte | EMailbox | Domain-name of the mailbox to receive error messages related to this list. See [`domain-name.md`](domain-name.md). |

## Bit fields

None.

## Variable-length fields

Two [`domain-name`](domain-name.md) values. Compression is permitted
(RFC 1035 §4.1.4 lists MINFO).

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

Two [`domain-name`](domain-name.md).

## Versioning notes

Listed by RFC 1035 as experimental. Still defined but rarely used.

## Ambiguities

None.
