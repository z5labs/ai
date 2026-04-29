# Classes

16-bit `CLASS` codes appearing in the resource-record `Class` field. Referenced by [`../structures/resource-record.md`](../structures/resource-record.md). RFC 1035 §3.2.4.

All values are big-endian uint16 on the wire.

| Value | Name | Description | Reference |
|---|---|---|---|
| 1 | IN | The Internet | RFC 1035 §3.2.4 |
| 2 | CS | CSNET (Obsolete — present only in obsolete RFCs) | RFC 1035 §3.2.4 |
| 3 | CH | CHAOS | RFC 1035 §3.2.4 |
| 4 | HS | Hesiod | RFC 1035 §3.2.4 |

## Notes

- For a Go DNS implementation focused on Internet records, only `IN` matters in practice; the others are vestigial.
- IANA registry: <https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-2>.
