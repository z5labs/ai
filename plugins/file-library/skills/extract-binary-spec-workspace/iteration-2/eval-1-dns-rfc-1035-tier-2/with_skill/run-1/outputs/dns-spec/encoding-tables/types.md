# Types

16-bit `TYPE` codes appearing in the resource-record `Type` field. A `Type` is a subset of `QType`. Referenced by [`../structures/resource-record.md`](../structures/resource-record.md). RFC 1035 §3.2.2.

All values are big-endian uint16 on the wire.

| Value | Name | Description | Reference |
|---|---|---|---|
| 1 | A | Host address (IPv4) — see [`../structures/rdata-a.md`](../structures/rdata-a.md) | RFC 1035 §3.4.1 |
| 2 | NS | Authoritative name server — see [`../structures/rdata-ns.md`](../structures/rdata-ns.md) | RFC 1035 §3.3.11 |
| 3 | MD | Mail destination (Obsolete — use MX) — see [`../structures/rdata-md.md`](../structures/rdata-md.md) | RFC 1035 §3.3.4 |
| 4 | MF | Mail forwarder (Obsolete — use MX) — see [`../structures/rdata-mf.md`](../structures/rdata-mf.md) | RFC 1035 §3.3.5 |
| 5 | CNAME | Canonical name for an alias — see [`../structures/rdata-cname.md`](../structures/rdata-cname.md) | RFC 1035 §3.3.1 |
| 6 | SOA | Start of zone of authority — see [`../structures/rdata-soa.md`](../structures/rdata-soa.md) | RFC 1035 §3.3.13 |
| 7 | MB | Mailbox domain name (EXPERIMENTAL) — see [`../structures/rdata-mb.md`](../structures/rdata-mb.md) | RFC 1035 §3.3.3 |
| 8 | MG | Mail-group member (EXPERIMENTAL) — see [`../structures/rdata-mg.md`](../structures/rdata-mg.md) | RFC 1035 §3.3.6 |
| 9 | MR | Mail rename domain name (EXPERIMENTAL) — see [`../structures/rdata-mr.md`](../structures/rdata-mr.md) | RFC 1035 §3.3.8 |
| 10 | NULL | Null RR (EXPERIMENTAL) — see [`../structures/rdata-null.md`](../structures/rdata-null.md) | RFC 1035 §3.3.10 |
| 11 | WKS | Well-known service description — see [`../structures/rdata-wks.md`](../structures/rdata-wks.md) | RFC 1035 §3.4.2 |
| 12 | PTR | Domain-name pointer — see [`../structures/rdata-ptr.md`](../structures/rdata-ptr.md) | RFC 1035 §3.3.12 |
| 13 | HINFO | Host information — see [`../structures/rdata-hinfo.md`](../structures/rdata-hinfo.md) | RFC 1035 §3.3.2 |
| 14 | MINFO | Mailbox / mail list information — see [`../structures/rdata-minfo.md`](../structures/rdata-minfo.md) | RFC 1035 §3.3.7 |
| 15 | MX | Mail exchange — see [`../structures/rdata-mx.md`](../structures/rdata-mx.md) | RFC 1035 §3.3.9 |
| 16 | TXT | Text strings — see [`../structures/rdata-txt.md`](../structures/rdata-txt.md) | RFC 1035 §3.3.14 |

## Notes

- Values 17+ are assigned by IANA in later RFCs (AAAA = 28 / RFC 3596, SRV = 33 / RFC 2782, etc.). Out of scope for RFC 1035.
- An RFC 1035 decoder MUST tolerate unknown `Type` values: read the `RDLength` bytes verbatim and surface the raw payload.
- IANA registry: <https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4>.
