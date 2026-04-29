# RCodes

4-bit RCODE field of the DNS header `Flags` word. Meaningful only on responses. Referenced by [`../structures/header.md`](../structures/header.md). RFC 1035 §4.1.1.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | NOERROR | No error condition | RFC 1035 §4.1.1 |
| 1 | FORMERR | Format error — server could not interpret the query | RFC 1035 §4.1.1 |
| 2 | SERVFAIL | Server failure — server problem prevented processing | RFC 1035 §4.1.1 |
| 3 | NXDOMAIN | Name Error — referenced domain name does not exist (only meaningful from authoritative servers) | RFC 1035 §4.1.1 |
| 4 | NOTIMP | Not Implemented — server does not support the requested kind of query | RFC 1035 §4.1.1 |
| 5 | REFUSED | Refused — policy refusal | RFC 1035 §4.1.1 |
| 6-15 | — | Reserved for future use | RFC 1035 §4.1.1 |

## Notes

- Later RFCs assigned 6 (YXDOMAIN), 7 (YXRRSET), 8 (NXRRSET), 9 (NOTAUTH), 10 (NOTZONE) — RFC 2136. Out of scope here.
- IANA registry: <https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-6>.
