# QTypes

16-bit `QTYPE` codes appearing in the question-section `QType` field. `QType` is a superset of `Type`: every value in [`types.md`](types.md) is a valid `QType`, plus the additional values below. Referenced by [`../structures/question.md`](../structures/question.md). RFC 1035 §3.2.3.

All values are big-endian uint16 on the wire.

| Value | Name | Description | Reference |
|---|---|---|---|
| 1–16 | (TYPE values) | Every value in [`types.md`](types.md) | RFC 1035 §3.2.3 |
| 252 | AXFR | Request for transfer of an entire zone | RFC 1035 §3.2.3 |
| 253 | MAILB | Request for mailbox-related records (MB, MG, or MR) | RFC 1035 §3.2.3 |
| 254 | MAILA | Request for mail agent RRs (Obsolete — see MX) | RFC 1035 §3.2.3 |
| 255 | * | Request for all records (a.k.a. ANY) | RFC 1035 §3.2.3 |

## Notes

- AXFR is normally sent over TCP only (RFC 1035 §4.2).
- Out of scope here but defined elsewhere: 251 (IXFR, RFC 1995), 250 (TSIG, RFC 2845), 249 (TKEY, RFC 2930).
- IANA registry (shared with TYPE): <https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4>.
