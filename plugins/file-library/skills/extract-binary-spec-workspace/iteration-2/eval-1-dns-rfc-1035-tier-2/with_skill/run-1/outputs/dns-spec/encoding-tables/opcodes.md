# Opcodes

4-bit OPCODE field of the DNS header `Flags` word. Set by the originator and copied unchanged into the response. Referenced by [`../structures/header.md`](../structures/header.md). RFC 1035 §4.1.1.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | QUERY | Standard query | RFC 1035 §4.1.1 |
| 1 | IQUERY | Inverse query | RFC 1035 §4.1.1 (obsoleted by RFC 3425 in deployed practice; still in the wire definition) |
| 2 | STATUS | Server status request | RFC 1035 §4.1.1 |
| 3-15 | — | Reserved for future use | RFC 1035 §4.1.1 |

## Notes

- Implementations targeting RFC 1035 only need to implement Opcode 0 (QUERY); 1 and 2 are rarely used in practice.
- Later RFCs assigned 4 (Notify, RFC 1996) and 5 (Update, RFC 2136). Out of scope here.
- IANA registry: <https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-5>.
