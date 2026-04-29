# Resource Record

A single entry in the Answer, Authority, or Additional sections of a DNS
message. ANCOUNT, NSCOUNT, and ARCOUNT in the header give the per-section
counts; all three sections share this exact wire format. Defined in RFC
1035 §4.1.3.

## Byte diagram

```
+-- NAME (variable; same encoding as QNAME) ------------------------+
|                                                                   |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      TYPE                     |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     CLASS                     |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
|                      TTL                      |
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                   RDLENGTH                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     RDATA                     /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | NAME | Owner name — domain name in the same wire form as QNAME, see [`domain-name.md`](domain-name.md). |
| NAME end | 2 | uint16 | TYPE | Resource record type — see [`../encoding-tables/types.md`](../encoding-tables/types.md). |
| NAME end + 2 | 2 | uint16 | CLASS | Record class — see [`../encoding-tables/classes.md`](../encoding-tables/classes.md). |
| NAME end + 4 | 4 | int32 | TTL | Time-to-live in seconds. RFC 1035 says "32 bit" without sign; RFC 2181 §8 clarifies this is a *signed* 32-bit value: only values in `[0, 2^31-1]` are valid; a decoder seeing the high bit set should treat the TTL as zero. |
| NAME end + 8 | 2 | uint16 | RDLENGTH | Length of RDATA in octets. |
| NAME end + 10 | RDLENGTH | []byte | RDATA | Type/class-specific resource data — see the per-type structure files. |

All multi-byte integer fields are big-endian.

## Bit fields

None at the RR-envelope level. Some specific RDATA payloads have their
own bit fields (e.g., [`rdata-wks.md`](rdata-wks.md)).

## Variable-length fields

### NAME

- **Length determination**: terminated by a zero-length root label or by
  a compression pointer (`0b11xxxxxx` first byte; 2 bytes total).
- **Encoding**: see [`domain-name.md`](domain-name.md).
- **Maximum length**: 255 octets in wire form.

### RDATA

- **Length determination**: explicit 16-bit length-prefix (RDLENGTH).
- **Length prefix counts**: only the bytes of RDATA itself; it does
  *not* include itself, NAME, TYPE, CLASS, or TTL.
- **Maximum length**: bounded by RDLENGTH's `uint16` range (65535 octets);
  realistic limits come from message-size limits (512 over UDP, 65535
  over TCP minus framing).
- **Encoding**: type-specific. The structure files listed below define
  the layout for each TYPE defined in RFC 1035. An unknown TYPE/CLASS is
  decoded as RDLENGTH opaque bytes; the decoder MUST consume exactly
  RDLENGTH bytes regardless of whether it understands the type.

## Conditional / optional fields

The RDATA layout depends on the (TYPE, CLASS) pair. For all classes,
this reference covers TYPE-specific layouts in dedicated files:

| TYPE | Structure |
|---|---|
| A (1) | [`rdata-a.md`](rdata-a.md) |
| NS (2) | [`rdata-ns.md`](rdata-ns.md) |
| MD (3), MF (4) | obsolete; same shape as NS but their use is deprecated by RFC 1035. |
| CNAME (5) | same shape as NS — single domain-name. See [`rdata-ns.md`](rdata-ns.md). |
| SOA (6) | [`rdata-soa.md`](rdata-soa.md) |
| MB (7), MG (8), MR (9) | experimental; single domain-name shape (see `rdata-ns.md`). |
| NULL (10) | [`rdata-null.md`](rdata-null.md) |
| WKS (11) | [`rdata-wks.md`](rdata-wks.md) |
| PTR (12) | same shape as NS — single domain-name. See [`rdata-ns.md`](rdata-ns.md). |
| HINFO (13) | [`rdata-hinfo.md`](rdata-hinfo.md) |
| MINFO (14) | [`rdata-minfo.md`](rdata-minfo.md) |
| MX (15) | [`rdata-mx.md`](rdata-mx.md) |
| TXT (16) | [`rdata-txt.md`](rdata-txt.md) |

## Checksums and integrity

None. RFC 1035 does not define any per-RR checksum.

## Padding and alignment

None. Records are byte-packed and concatenated directly when a section
holds multiple RRs.

## Nested structures

- NAME is a [`domain-name`](domain-name.md).
- RDATA may itself contain one or more domain names (for NS, CNAME, PTR,
  SOA, MX, MINFO). When a name appears inside RDATA, compression
  pointers are permitted and target offsets relative to the start of
  the *whole DNS message*, not relative to RDATA. RFC 1035 §4.1.4
  recommends compression only for well-known TYPEs, not for unknown or
  future types — a decoder always trusts RDLENGTH and works in raw
  bytes when the type is unknown.

## Versioning notes

The envelope is unchanged since RFC 1035. New TYPEs are assigned
regularly (AAAA, SRV, NAPTR, OPT, ...) but they all use this same
NAME / TYPE / CLASS / TTL / RDLENGTH / RDATA shape.

## Ambiguities

> **Ambiguity:** RFC 1035 declares TTL as "32 bit unsigned" in §3.2.1
> but the field-table column in §4.1.3 names it "TTL" without a sign.
> RFC 2181 §8 finalizes the interpretation as *signed* with values
> >= 2^31 to be treated as zero. This reference uses `int32` for that
> reason; an implementer who only consults RFC 1035 will see `uint32`.

> **Ambiguity:** RFC 1035 does not specify what to do when RDLENGTH
> exceeds the bytes remaining in the message. A robust decoder rejects
> the message; some lenient implementations truncate RDATA.
