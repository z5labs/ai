# RDataMINFO

RDATA payload of a `Type=MINFO` resource record (EXPERIMENTAL): two domain names identifying mailboxes responsible for and receiving error reports about a mailing list. RFC 1035 §3.3.7.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                    RMAILBX                    /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                    EMAILBX                    /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | RMailBX | Mailbox responsible for the mailing list or mailbox. Domain name; if it names the root, the owner is responsible for itself. See [`domain-name.md`](domain-name.md). |
| variable | variable | DomainName | EMailBX | Mailbox to receive error messages related to the mailing list/mailbox. If it names the root, errors return to the original sender. |

## Notes

- `RDLength` MUST equal the sum of the two on-wire domain-name lengths.
- MINFO is EXPERIMENTAL per RFC 1035 §3.3.7; encoders MAY produce it but most resolvers ignore it.

## Ambiguities

> **Ambiguity:** RFC 1035 §3.3 does not list MINFO among the explicitly compressible types. Compression is widely tolerated in practice; decoders MUST handle compression here regardless.
