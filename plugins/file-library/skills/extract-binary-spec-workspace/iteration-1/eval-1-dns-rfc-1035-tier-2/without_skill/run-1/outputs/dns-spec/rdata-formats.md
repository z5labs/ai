# RDATA Formats by TYPE (RFC 1035 Sections 3.3 and 3.4)

This document describes the wire layout of RDATA for every standard RR TYPE
defined by RFC 1035. RDATA is preceded by an `RDLENGTH` field in the
enclosing RR (see `resource-record.md`); the RDATA byte count MUST match
RDLENGTH.

Domain names appearing in RDATA fields are encoded per `domain-names.md` and
MAY use compression pointers.

## Notation

- `<domain-name>` - a domain name as defined in `domain-names.md`.
- `<character-string>` - a length-prefixed string: a single octet giving the
  string's length (0-255), followed by that many octets of data. Used in
  HINFO and TXT.

## A (TYPE=1, Section 3.4.1)

```
                                    1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    ADDRESS                    |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- ADDRESS: a 32-bit Internet address (4 octets, network byte order).
- RDLENGTH = 4.
- Hosts that have multiple Internet addresses will have multiple A records.

## NS (TYPE=2, Section 3.3.11)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   NSDNAME                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- NSDNAME: `<domain-name>` of a name server for the owner domain.
- RDLENGTH = encoded length of NSDNAME (with any compression).

## MD (TYPE=3, Obsolete, Section 3.3.4)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   MADNAME                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MADNAME: `<domain-name>`. Obsolete; servers SHOULD return `Not Implemented`
  (RCODE 4) for queries of this type. Use MX instead.

## MF (TYPE=4, Obsolete, Section 3.3.5)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   MADNAME                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MADNAME: `<domain-name>`. Obsolete; treat like MD.

## CNAME (TYPE=5, Section 3.3.1)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                     CNAME                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- CNAME: `<domain-name>` specifying the canonical or primary name for the
  owner. The owner name is an alias.

## SOA (TYPE=6, Section 3.3.13)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                     MNAME                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                     RNAME                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    SERIAL                     |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    REFRESH                    |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     RETRY                     |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    EXPIRE                     |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    MINIMUM                    |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MNAME: `<domain-name>` of the name server that was the original or primary
  source of data for this zone.
- RNAME: `<domain-name>` mailbox of the person responsible for this zone.
- SERIAL: 32-bit unsigned integer (big-endian). Zone version.
- REFRESH: 32-bit signed integer (big-endian). Time interval before refresh.
- RETRY: 32-bit signed integer (big-endian). Retry interval after a failed refresh.
- EXPIRE: 32-bit signed integer (big-endian). Upper limit before zone is no longer authoritative.
- MINIMUM: 32-bit unsigned integer (big-endian). Minimum TTL field for any RR from this zone.

Note: the four interval fields are described as 32-bit signed in RFC 1035;
practical implementations treat them as unsigned but with values < 2^31.

Total RDATA size = `len(MNAME) + len(RNAME) + 20`.

## MB (TYPE=7, EXPERIMENTAL, Section 3.3.3)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   MADNAME                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MADNAME: `<domain-name>` specifying a host which has the specified mailbox.

## MG (TYPE=8, EXPERIMENTAL, Section 3.3.6)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   MGMNAME                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MGMNAME: `<domain-name>` specifying a mailbox which is a member of the mail
  group specified by the domain name.

## MR (TYPE=9, EXPERIMENTAL, Section 3.3.8)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   NEWNAME                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- NEWNAME: `<domain-name>` specifying a mailbox which is the proper rename of
  the specified mailbox.

## NULL (TYPE=10, EXPERIMENTAL, Section 3.3.10)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                  <anything>                   /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- Anything at all may be in the RDATA field, so long as it is 65535 octets or
  fewer.
- NULL records are not allowed in master files; they are used in experimental
  extensions.

## WKS (TYPE=11, Section 3.4.2)

```
                                    1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                    ADDRESS                    |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |       PROTOCOL        |                       |
    +--+--+--+--+--+--+--+--+                       |
    |                                               |
    /                   <BIT MAP>                   /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- ADDRESS: 32-bit Internet address (4 octets).
- PROTOCOL: 8-bit IP protocol number (e.g., 6=TCP, 17=UDP).
- `<BIT MAP>`: variable-length octet string. Bit `n` of the bit map (with bit
  0 of the first octet being port 0, bit 7 of the first octet being port 7,
  bit 0 of the second octet being port 8, etc.) indicates that the well-known
  service running on TCP/UDP port `n` is supported.
- Trailing zero octets in the bit map MAY be omitted.
- RDLENGTH = 5 + len(BIT MAP).

## PTR (TYPE=12, Section 3.3.12)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   PTRDNAME                    /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- PTRDNAME: `<domain-name>` which points to some location in the domain name
  space. Used most commonly for reverse-DNS lookups (in-addr.arpa).

## HINFO (TYPE=13, Section 3.3.2)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                      CPU                      /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                       OS                      /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- CPU: `<character-string>` (length-prefixed) specifying the CPU type.
- OS: `<character-string>` specifying the operating system.
- RDLENGTH = (1 + len(CPU)) + (1 + len(OS)).

`<character-string>` strings are NOT compressed and do NOT use a terminating
zero - they are simply length-prefixed octet strings of length 0-255.

## MINFO (TYPE=14, EXPERIMENTAL, Section 3.3.7)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   RMAILBX                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   EMAILBX                     /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- RMAILBX: `<domain-name>` mailbox responsible for the mailing list or mailbox.
- EMAILBX: `<domain-name>` mailbox to receive error messages relating to the
  mailing list or mailbox.

## MX (TYPE=15, Section 3.3.9)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                  PREFERENCE                   |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   EXCHANGE                    /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- PREFERENCE: 16-bit unsigned big-endian integer giving the preference (lower
  values preferred) of this RR among others at the same owner.
- EXCHANGE: `<domain-name>` of a host willing to act as a mail exchange for
  the owner name.
- RDLENGTH = 2 + len(EXCHANGE).

## TXT (TYPE=16, Section 3.3.14)

```
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    /                   TXT-DATA                    /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- TXT-DATA: One or more `<character-string>`s (length-prefixed). The TXT RR
  RDATA contains a sequence of one or more such strings concatenated.
- RDLENGTH = sum of `(1 + len(s))` over each string s.

## RDLENGTH considerations for compression

For the RR TYPEs whose RDATA contains a `<domain-name>` (NS, MD, MF, CNAME,
MB, MG, MR, PTR, MINFO, MX, SOA), encoders MAY use compression pointers
inside RDATA. When they do, RDLENGTH MUST reflect the **encoded** length on
the wire (including the 2-byte pointers), not the expanded length.

For HINFO and TXT, the data is `<character-string>` and is NEVER compressed.

For A, WKS, NULL, the data is opaque/binary and is NEVER compressed.
