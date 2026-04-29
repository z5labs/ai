# Opcodes

The 4-bit Opcode field of the [`header`](../structures/header.md) Flags
word identifies the kind of query. Set by the originator of a query and
copied unchanged into the response.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | QUERY | Standard query (look up records by name/type/class). | RFC 1035 §4.1.1 |
| 1 | IQUERY | Inverse query (obsolete). RFC 1035 §6.4 defined it; later RFC 3425 retired it. | RFC 1035 §4.1.1, RFC 3425 |
| 2 | STATUS | Server status request (rarely implemented). | RFC 1035 §4.1.1 |
| 3–15 | Reserved | Not assigned by RFC 1035. Newer RFCs allocate values 4 (Notify, RFC 1996) and 5 (Update, RFC 2136); these are out of scope here. | — |

## Notes

- The Opcode field is 4 bits wide and occupies bits 1–4 of the
  Flags word (MSB-0 numbering across the 16-bit Flags word; equivalent
  to bits 1–4 of byte 2 of the header).
- A receiver that does not recognize the Opcode SHOULD respond with
  RCODE 4 (NOTIMP). See [`rcodes.md`](rcodes.md).
