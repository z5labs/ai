# Response Codes (RCODE)

The 4-bit RCODE field of the [`header`](../structures/header.md) Flags
word communicates the result of a query. Meaningful only in responses.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | NoError | No error; the response is valid. | RFC 1035 §4.1.1 |
| 1 | FormErr | Format error. The server could not interpret the query. | RFC 1035 §4.1.1 |
| 2 | ServFail | Server failure. The server encountered an internal problem. | RFC 1035 §4.1.1 |
| 3 | NXDomain | Name error. The domain name in the query does not exist. Meaningful only from an authoritative server. | RFC 1035 §4.1.1 |
| 4 | NotImp | Not implemented. The server does not support the requested kind of query. | RFC 1035 §4.1.1 |
| 5 | Refused | The server refuses to perform the operation for policy reasons. | RFC 1035 §4.1.1 |
| 6–15 | Reserved | Not assigned by RFC 1035. | — |

## Notes

- The RCODE field is 4 bits wide and occupies bits 12–15 of the
  Flags word (MSB-0 numbering across the 16-bit Flags word; equivalent
  to bits 4–7 of byte 3 of the header).
- Later RFCs allocate codes 6–10 (YXDomain, YXRRSet, NXRRSet, NotAuth,
  NotZone) and use the EDNS0 OPT pseudo-RR to extend RCODE beyond 4
  bits. These are out of scope of pure RFC 1035 decoding.
