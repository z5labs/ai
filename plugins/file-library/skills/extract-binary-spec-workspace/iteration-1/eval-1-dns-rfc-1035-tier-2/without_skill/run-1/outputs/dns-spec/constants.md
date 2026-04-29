# Constants Quick Reference

All numeric constants needed to implement a DNS encoder/decoder against RFC
1035 Section 4. Provided in one place for convenient reference / Go code
generation.

## Sizes (octets)

| Constant                              | Value                  |
|---------------------------------------|-----------------------:|
| Header size                           | 12                     |
| Fixed RR fields after NAME (TYPE+CLASS+TTL+RDLENGTH) | 10      |
| Question fixed fields after QNAME (QTYPE+QCLASS) | 4           |
| Max single-label length (octets)      | 63                     |
| Max domain-name length (octets, encoded with length octets and root)  | 255 |
| Max UDP DNS message size (no EDNS)    | 512                    |
| Max TCP DNS message size              | 65535 (2-octet length) |
| TCP length prefix size                | 2                      |
| Compression pointer size              | 2                      |
| Max compression pointer offset        | 0x3FFF (14 bits)       |

## Header flag bit masks (within the 16-bit flags word)

| Mask     | Field   |
|----------|---------|
| 0x8000   | QR      |
| 0x7800   | Opcode (4 bits) |
| 0x0400   | AA      |
| 0x0200   | TC      |
| 0x0100   | RD      |
| 0x0080   | RA      |
| 0x0070   | Z (3 bits, reserved, must be 0) |
| 0x000F   | RCODE (4 bits) |

## Header flag shift counts

| Shift | Field   |
|------:|---------|
| 15    | QR      |
| 11    | Opcode  |
| 10    | AA      |
| 9     | TC      |
| 8     | RD      |
| 7     | RA      |
| 4     | Z       |
| 0     | RCODE   |

## Opcodes

| Mnemonic | Value |
|----------|------:|
| QUERY    | 0     |
| IQUERY   | 1     |
| STATUS   | 2     |

## RCODEs

| Mnemonic        | Value |
|-----------------|------:|
| NoError         | 0     |
| FormatError     | 1     |
| ServerFailure   | 2     |
| NameError       | 3     |
| NotImplemented  | 4     |
| Refused         | 5     |

## TYPEs

| Mnemonic | Value |
|----------|------:|
| A        | 1     |
| NS       | 2     |
| MD       | 3 (Obsolete) |
| MF       | 4 (Obsolete) |
| CNAME    | 5     |
| SOA      | 6     |
| MB       | 7 (Experimental) |
| MG       | 8 (Experimental) |
| MR       | 9 (Experimental) |
| NULL     | 10 (Experimental) |
| WKS      | 11    |
| PTR      | 12    |
| HINFO    | 13    |
| MINFO    | 14 (Experimental) |
| MX       | 15    |
| TXT      | 16    |

## QTYPE-only additions

| Mnemonic | Value |
|----------|------:|
| AXFR     | 252   |
| MAILB    | 253   |
| MAILA    | 254 (Obsolete) |
| ANY (`*`) | 255  |

## CLASSes

| Mnemonic | Value |
|----------|------:|
| IN       | 1     |
| CS       | 2 (Obsolete) |
| CH       | 3     |
| HS       | 4     |

## QCLASS-only additions

| Mnemonic | Value |
|----------|------:|
| ANY (`*`) | 255  |

## Compression pointer encoding

- A label-length octet whose top two bits are `00` introduces a literal label
  of length 0-63.
- A two-octet sequence whose top two bits are `11` is a compression pointer.
  The remaining 14 bits are the offset (big-endian) into the message at which
  to continue reading the name.
- A label-length octet with top two bits `01` or `10` is **reserved** and
  MUST be treated as a parse error.
- Pointer 16-bit value on the wire: `0xC000 | offset`, where
  `0 <= offset <= 0x3FFF`.

## End-of-name marker

- A single octet `0x00` terminates a non-pointer-terminated domain name (the
  zero-length label = root).
