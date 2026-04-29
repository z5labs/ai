# TYPE values

The 16-bit TYPE field of a [`resource-record`](../structures/resource-record.md)
identifies the kind of record. The values listed here are the ones
defined by RFC 1035 itself; later RFCs assign many more (AAAA, SRV,
NAPTR, OPT, ...) which are out of scope.

| Value | Name | Description | Reference |
|---|---|---|---|
| 1 | A | Host address (IPv4). RDATA is a 4-byte address. | RFC 1035 §3.4.1 |
| 2 | NS | Authoritative name server. RDATA is a domain name. | RFC 1035 §3.3.11 |
| 3 | MD | Mail destination (obsolete; use MX). RDATA is a domain name. | RFC 1035 §3.3.4 |
| 4 | MF | Mail forwarder (obsolete; use MX). RDATA is a domain name. | RFC 1035 §3.3.5 |
| 5 | CNAME | Canonical name alias. RDATA is a domain name. | RFC 1035 §3.3.1 |
| 6 | SOA | Marks the start of a zone of authority. | RFC 1035 §3.3.13 |
| 7 | MB | Mailbox domain name (experimental). | RFC 1035 §3.3.3 |
| 8 | MG | Mail group member (experimental). | RFC 1035 §3.3.6 |
| 9 | MR | Mail rename domain name (experimental). | RFC 1035 §3.3.8 |
| 10 | NULL | NULL record (experimental). | RFC 1035 §3.3.10 |
| 11 | WKS | Well known service description. | RFC 1035 §3.4.2 |
| 12 | PTR | Domain name pointer. | RFC 1035 §3.3.12 |
| 13 | HINFO | Host information. | RFC 1035 §3.3.2 |
| 14 | MINFO | Mailbox or mail list information. | RFC 1035 §3.3.7 |
| 15 | MX | Mail exchange. | RFC 1035 §3.3.9 |
| 16 | TXT | Text strings. | RFC 1035 §3.3.14 |

## Notes

- The TYPE field is a 16-bit unsigned integer in network byte order.
- Values 0 and 17–65535 are unallocated by RFC 1035; consult the IANA
  "Domain Name System (DNS) Parameters" registry for current assignments.
- A decoder that does not recognize a TYPE MUST still consume RDLENGTH
  bytes of RDATA so subsequent records can be decoded; per RFC 1035
  §4.1.4 it MUST NOT attempt name-decompression inside such RDATA.
