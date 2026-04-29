# CLASS values

The 16-bit CLASS field of a [`resource-record`](../structures/resource-record.md)
identifies the protocol family of the record.

| Value | Name | Description | Reference |
|---|---|---|---|
| 1 | IN | the Internet (IPv4). The only class in widespread use. | RFC 1035 §3.2.4 |
| 2 | CS | the CSNET class (obsolete). | RFC 1035 §3.2.4 |
| 3 | CH | the CHAOS class. | RFC 1035 §3.2.4 |
| 4 | HS | Hesiod (Dyer 87). | RFC 1035 §3.2.4 |

## Notes

- Encoding is 16-bit unsigned, network byte order.
- Values 0 and 5–65535 are unallocated by RFC 1035 (reserved or
  unassigned).
- Most DNS traffic uses class IN exclusively. CH is occasionally used
  for server identification queries (e.g. `version.bind`).
