# QTYPE values

The 16-bit QTYPE field of a [`question`](../structures/question.md) is a
superset of TYPE. It accepts every value from
[`types.md`](types.md) plus the additional query-only values listed
below.

| Value | Name | Description | Reference |
|---|---|---|---|
| 1–16 | (TYPE values) | All values from [`types.md`](types.md) are valid QTYPEs. | RFC 1035 §3.2.2 |
| 252 | AXFR | Request for a zone transfer. | RFC 1035 §3.2.3, §4.3.5 |
| 253 | MAILB | Request for mailbox-related records (MB, MG, or MR). | RFC 1035 §3.2.3 |
| 254 | MAILA | Request for mail agent RRs (obsolete; superseded by MX). | RFC 1035 §3.2.3 |
| 255 | * (ANY) | Request for all records (any type). | RFC 1035 §3.2.3 |

## Notes

- QTYPE is wider than TYPE *only* in queries. A response RR uses TYPE,
  not QTYPE; a server never echoes 252/253/254/255 in an Answer's TYPE.
- QTYPE 255 ("any") returns whichever records the server happens to
  have cached; it is not a comprehensive cross-server lookup.
- Encoding is the same as TYPE: 16-bit unsigned, network byte order.
