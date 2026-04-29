# TYPE / QTYPE / CLASS / QCLASS / Opcode / RCODE Values

All numeric values defined by RFC 1035 Section 3.2.2 - 3.2.5 and Section 4.1.1.

## TYPE values (Section 3.2.2)

TYPE fields are used in RR wire format. Each entry is a 16-bit unsigned big-endian
integer.

| Value | Mnemonic | Meaning                                                |
|------:|----------|--------------------------------------------------------|
| 1     | A        | A host address (32-bit IPv4 address).                  |
| 2     | NS       | An authoritative name server.                          |
| 3     | MD       | A mail destination (Obsolete - use MX).                |
| 4     | MF       | A mail forwarder (Obsolete - use MX).                  |
| 5     | CNAME    | The canonical name for an alias.                       |
| 6     | SOA      | Marks the start of a zone of authority.                |
| 7     | MB       | A mailbox domain name (EXPERIMENTAL).                  |
| 8     | MG       | A mail group member (EXPERIMENTAL).                    |
| 9     | MR       | A mail rename domain name (EXPERIMENTAL).              |
| 10    | NULL     | A null RR (EXPERIMENTAL).                              |
| 11    | WKS      | A well known service description.                      |
| 12    | PTR      | A domain name pointer.                                 |
| 13    | HINFO    | Host information.                                      |
| 14    | MINFO    | Mailbox or mail list information.                      |
| 15    | MX       | Mail exchange.                                         |
| 16    | TXT      | Text strings.                                          |

## QTYPE values (Section 3.2.3)

QTYPE is a superset of TYPE. In addition to all TYPE values above, the
following codes are valid only in the QTYPE field of a question:

| Value | Mnemonic | Meaning                                                  |
|------:|----------|----------------------------------------------------------|
| 252   | AXFR     | A request for a transfer of an entire zone.              |
| 253   | MAILB    | A request for mailbox-related records (MB, MG, or MR).   |
| 254   | MAILA    | A request for mail agent RRs (Obsolete - see MX).        |
| 255   | *        | A request for all records.                               |

## CLASS values (Section 3.2.4)

CLASS fields are used in RR wire format. 16-bit unsigned big-endian.

| Value | Mnemonic | Meaning                                              |
|------:|----------|------------------------------------------------------|
| 1     | IN       | The Internet.                                        |
| 2     | CS       | The CSNET class (Obsolete - used only for examples). |
| 3     | CH       | The CHAOS class.                                     |
| 4     | HS       | Hesiod [Dyer 87].                                    |

## QCLASS values (Section 3.2.5)

QCLASS is a superset of CLASS. Valid only in the QCLASS field of a question:

| Value | Mnemonic | Meaning           |
|------:|----------|-------------------|
| 255   | *        | Any class.        |

## Opcode values (Section 4.1.1, in the header flags word)

4-bit field.

| Value | Mnemonic | Meaning                                       |
|------:|----------|-----------------------------------------------|
| 0     | QUERY    | A standard query.                             |
| 1     | IQUERY   | An inverse query.                             |
| 2     | STATUS   | A server status request.                      |
| 3-15  |          | Reserved for future use.                      |

## RCODE values (Section 4.1.1, in the header flags word)

4-bit field.

| Value | Meaning                                                         |
|------:|-----------------------------------------------------------------|
| 0     | No error condition.                                             |
| 1     | Format error - the name server was unable to interpret the query. |
| 2     | Server failure - the name server was unable to process this query due to a problem with the name server. |
| 3     | Name Error - meaningful only for responses from an authoritative name server; signifies that the domain name referenced in the query does not exist. |
| 4     | Not Implemented - the name server does not support the requested kind of query. |
| 5     | Refused - the name server refuses to perform the specified operation for policy reasons. |
| 6-15  | Reserved for future use.                                        |
