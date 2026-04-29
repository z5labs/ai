# QClasses

16-bit `QCLASS` codes appearing in the question-section `QClass` field. `QClass` is a superset of `Class`: every value in [`classes.md`](classes.md) is a valid `QClass`, plus the additional value below. Referenced by [`../structures/question.md`](../structures/question.md). RFC 1035 §3.2.5.

All values are big-endian uint16 on the wire.

| Value | Name | Description | Reference |
|---|---|---|---|
| 1–4 | (CLASS values) | Every value in [`classes.md`](classes.md) | RFC 1035 §3.2.5 |
| 255 | * | Any class | RFC 1035 §3.2.5 |

## Notes

- `* (255)` MAY only appear in a question; it is not a valid `Class` for an actual resource record.
- IANA registry (shared with CLASS): <https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-2>.
