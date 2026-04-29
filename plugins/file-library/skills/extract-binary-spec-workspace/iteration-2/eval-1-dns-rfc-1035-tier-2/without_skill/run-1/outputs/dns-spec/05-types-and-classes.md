# TYPE / QTYPE and CLASS / QCLASS Codes

Source: RFC 1035 §3.2.2, §3.2.3, §3.2.4, §3.2.5.

All codes are unsigned 16-bit values transmitted in big-endian order.

## TYPE values (used in RR TYPE; valid as QTYPE)

| Mnemonic | Value | Meaning                                              |
|----------|-------|------------------------------------------------------|
| A        | 1     | A host address                                       |
| NS       | 2     | An authoritative name server                         |
| MD       | 3     | A mail destination (Obsolete — use MX)               |
| MF       | 4     | A mail forwarder (Obsolete — use MX)                 |
| CNAME    | 5     | The canonical name for an alias                      |
| SOA      | 6     | Marks the start of a zone of authority               |
| MB       | 7     | A mailbox domain name (EXPERIMENTAL)                 |
| MG       | 8     | A mail group member (EXPERIMENTAL)                   |
| MR       | 9     | A mail rename domain name (EXPERIMENTAL)             |
| NULL     | 10    | A null RR (EXPERIMENTAL)                             |
| WKS      | 11    | A well known service description                     |
| PTR      | 12    | A domain name pointer                                |
| HINFO    | 13    | Host information                                     |
| MINFO    | 14    | Mailbox or mail list information                     |
| MX       | 15    | Mail exchange                                        |
| TXT      | 16    | Text strings                                         |

## QTYPE-only values (valid only in question section)

QTYPE is a superset of TYPE. The following extra codes appear only in
the Question section:

| Mnemonic | Value | Meaning                                              |
|----------|-------|------------------------------------------------------|
| AXFR     | 252   | Request for transfer of an entire zone               |
| MAILB    | 253   | Request for mailbox-related records (MB / MG / MR)   |
| MAILA    | 254   | Request for mail agent RRs (Obsolete — see MX)       |
| *        | 255   | Request for all records (sometimes called ANY)       |

## CLASS values (used in RR CLASS; valid as QCLASS)

| Mnemonic | Value | Meaning                                              |
|----------|-------|------------------------------------------------------|
| IN       | 1     | The Internet                                         |
| CS       | 2     | CSNET (Obsolete; only for examples in obsolete RFCs) |
| CH       | 3     | The CHAOS class                                      |
| HS       | 4     | Hesiod                                               |

## QCLASS-only values (valid only in question section)

QCLASS is a superset of CLASS. The extra code:

| Mnemonic | Value | Meaning                                              |
|----------|-------|------------------------------------------------------|
| *        | 255   | Any class                                            |

## Implementation notes

- Decoders should accept any 16-bit value for TYPE/CLASS and treat
  unknown codes as "opaque RR" (preserve raw RDATA bytes).
- The wildcard mnemonic `*` (value 255) is overloaded between
  QTYPE/QCLASS — disambiguated by which field it appears in.
- Code values not listed are reserved/IANA-allocated post-RFC-1035 and
  are out of scope for this spec.
