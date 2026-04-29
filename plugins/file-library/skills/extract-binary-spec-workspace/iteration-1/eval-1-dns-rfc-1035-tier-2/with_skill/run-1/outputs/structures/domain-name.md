# Domain Name (wire form)

The encoding used wherever a domain name appears on the wire — QNAME of
a question, NAME of a resource record, and any name embedded inside RDATA
(e.g. NS target, CNAME target, MX exchange, SOA MNAME/RNAME). Defined in
RFC 1035 §3.1 (uncompressed labels) and §4.1.4 (compression pointers).

## Byte diagram

A domain name is a sequence of labels followed by either a
zero-length root label (terminator) or a 2-byte compression pointer.

Uncompressed (root-terminated) form:

```
+-----+--------+-----+--------+ ... +-----+
| len |  label | len |  label |     |  0  |
+-----+--------+-----+--------+ ... +-----+
   ^      ^                            ^
   |   1..63 octets                    |
   |                                root label
 1 byte length (0..63)            (length byte = 0)
```

Compressed form (last label replaced by a 2-byte pointer):

```
+-----+--------+-----+--------+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
| len |  label | len |  label | 1| 1|        14-bit OFFSET                    |
+-----+--------+-----+--------+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
                               ^^^^
                              high bits of byte 1 = 11
```

The pointer is two octets read as a big-endian 16-bit value; the top two
bits are masked off (`value & 0x3FFF`) to obtain the offset.

## Field table

A domain name is not a fixed structure but a *sequence*. The two byte
forms that may appear at any position are:

| Discriminator (top 2 bits of next byte) | Form | Length on wire | Effect |
|---|---|---|---|
| `00` | Length-prefixed label | 1 + N (N = bottom 6 bits, 0..63) | If N == 0: root label — terminates the name. If 1 ≤ N ≤ 63: append the next N octets as a label. |
| `01` | Reserved | — | Not used in RFC 1035 (defined as "reserved for future use"). A decoder MUST reject. |
| `10` | Reserved | — | Not used in RFC 1035 ("reserved for future use"). A decoder MUST reject. |
| `11` | Compression pointer | 2 | Big-endian `uint16` masked with `0x3FFF` is an absolute byte offset from the start of the DNS message; resume label decoding at that offset. The pointer ends the current name. |

Per-byte view of a single label introducer:

| Bit(s) (MSB-0 in this byte) | Name | Description |
|---|---|---|
| 0–1 | Tag | `00` = literal label length, `11` = compression pointer, `01`/`10` reserved. |
| 2–7 | LabelLength | When Tag=`00`: number of octets that follow as the label payload. Range 0..63. When Tag=`11`: high 6 bits of the 14-bit pointer offset. |

For a compression pointer, byte 0 holds `11` || HIGH6(offset) and byte 1
holds LOW8(offset).

## Bit fields

See the table above — the top two bits of each name-position byte
discriminate label vs. pointer.

## Variable-length fields

### Label

- **Length determination**: explicit single-byte length prefix (0..63).
  A length of 0 is the root label and terminates the name.
- **Length prefix counts**: only the label payload; it does *not*
  include itself.
- **Maximum length**: 63 octets per label, 255 octets per complete name
  (counting all length bytes, all label payloads, and the terminating
  zero — RFC 1035 §2.3.4).
- **Encoding**: opaque 8-bit octets on the wire. Comparisons between
  names are case-insensitive over ASCII (RFC 1035 §2.3.3) but the bytes
  are preserved verbatim during transport.

### Compression pointer (RFC 1035 §4.1.4)

- **Form**: 2 bytes; the first byte's two high bits are `11`.
- **Offset extraction**: `offset = (byte0 << 8 | byte1) & 0x3FFF`.
- **Offset reference frame**: byte 0 of the *whole DNS message* (the
  first byte of the header's ID field), not the start of RDATA or the
  current structure.
- **Forward references**: forbidden — the pointer must point earlier in
  the message than the pointer itself, to guarantee the chain of
  pointers is acyclic and bounded.
- **Chain of pointers**: a target may itself be another pointer; a
  decoder must follow chains while detecting loops and bounding total
  resolved length to ≤ 255 octets.
- **Where allowed**: in QNAME, in NAME of an RR, and inside RDATA only
  for the well-known TYPEs whose RDATA is defined to contain a domain
  name (NS, CNAME, PTR, SOA, MX, MINFO). RFC 1035 §4.1.4 forbids
  compression in unknown / future TYPE RDATA so that future receivers
  can blindly skip RDLENGTH bytes.

### Decoding algorithm (informative)

```
name = []
seen_pointer = false
position = start_offset
limit = 0  // total label bytes including length bytes
while true:
    b = msg[position]
    tag = b >> 6
    if tag == 0:
        n = b & 0x3F
        if n == 0:
            position += 1
            break  // root label terminates the name
        if position + 1 + n > len(msg) or limit + 1 + n > 255:
            error
        name.append(msg[position+1 : position+1+n])
        position += 1 + n
        limit += 1 + n
    else if tag == 3:  // 0b11
        offset = ((b & 0x3F) << 8) | msg[position+1]
        if not seen_pointer:
            advance external read position to position+2
            seen_pointer = true
        if offset >= original_position:  // forward reference
            error
        position = offset
    else:
        error  // reserved tag (01 or 10)
```

The "external read position" is what the decoder must continue from
once the name is complete; the first pointer encountered fixes that
endpoint at `position+2`.

## Conditional / optional fields

None — a domain name is exactly one of: a sequence of labels ending in
the root label, or a sequence of labels ending in a compression pointer
(possibly zero labels followed directly by a pointer).

## Checksums and integrity

None.

## Padding and alignment

None — all bytes are tightly packed.

## Nested structures

Used by [`question`](question.md) (QNAME), [`resource-record`](resource-record.md)
(NAME), and the domain-name fields inside specific RDATA layouts:
[`rdata-ns.md`](rdata-ns.md), [`rdata-soa.md`](rdata-soa.md),
[`rdata-mx.md`](rdata-mx.md), [`rdata-minfo.md`](rdata-minfo.md).

## Versioning notes

Unchanged since RFC 1035. The two reserved tag values (`01` and `10`)
have never been allocated.

## Ambiguities

> **Ambiguity:** RFC 1035 §4.1.4 forbids forward compression pointers
> implicitly (by stating pointers point to "a prior occurrence of the
> same name"), but does not state explicitly what a decoder should do
> on a forward / cyclic pointer. Robust decoders treat any pointer that
> does not strictly decrease `position` as an error.

> **Ambiguity:** RFC 1035 does not give a maximum number of labels in a
> name, only a 255-octet wire limit and 63-octet label limit. The
> implicit upper bound is 127 labels (a name of 127 single-byte labels
> plus the root terminator = 255 octets).
