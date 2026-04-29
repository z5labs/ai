# QCLASS values

The 16-bit QCLASS field of a [`question`](../structures/question.md) is
a superset of CLASS — every value in [`classes.md`](classes.md) plus a
single additional wildcard.

| Value | Name | Description | Reference |
|---|---|---|---|
| 1–4 | (CLASS values) | All values from [`classes.md`](classes.md) are valid QCLASSes. | RFC 1035 §3.2.5 |
| 255 | * (ANY) | Match any class. | RFC 1035 §3.2.5 |

## Notes

- Encoding is 16-bit unsigned, network byte order.
- The wildcard is rarely used; nearly all real queries use QCLASS=IN.
