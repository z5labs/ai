# Encoding Rules, Limits, and Edge Cases

Source: RFC 1035 §2.3.2, §2.3.4, §3.1, §4.1.4 (with explicit notes when
something is added beyond the RFC for robustness).

## Byte order (§2.3.2)

- All multi-octet integer fields are big-endian (network byte order).
- Within an octet, bit numbering in the diagrams puts the most
  significant bit on the left (bit 0 = MSB).
- This applies to every uint16/uint32 in the header, question, RR
  header, and RDATA.

## Hard size limits (§2.3.4)

| Object        | Limit                                 |
|---------------|---------------------------------------|
| Label         | 0..63 octets (high 2 bits of length must be 00) |
| Domain name   | total <= 255 octets, including length bytes and terminating 0x00 |
| TTL           | non-negative 32-bit (0 .. 2^31 - 1)   |
| UDP message   | <= 512 octets total (transport-level) |

Additional structural limits (implicit in field widths):

| Object             | Width   | Range          |
|--------------------|---------|----------------|
| ID                 | 16 bits | 0..65535       |
| OPCODE             | 4 bits  | 0..15          |
| RCODE              | 4 bits  | 0..15          |
| QDCOUNT/ANCOUNT/NSCOUNT/ARCOUNT | 16 bits | 0..65535 |
| QTYPE/QCLASS/TYPE/CLASS         | 16 bits | 0..65535 |
| RDLENGTH           | 16 bits | 0..65535       |
| Compression OFFSET | 14 bits | 0..16383       |
| `<character-string>` length | 8 bits | 0..255 |

## Header invariants

- Header is exactly 12 octets at offset 0.
- Z (3 bits) MUST be zero on encode. On decode, follow Postel and
  ignore non-zero Z to remain forward-compatible.
- The four COUNT fields determine how many entries follow in the four
  sections; an encoder MUST set them to match the actual number of
  encoded entries.

## Domain-name invariants (§3.1, §4.1.4)

- Every name terminates with either a 0x00 length byte (literal
  terminator) or a compression pointer (the 2-byte pointer is itself
  the terminator — there is no trailing 0x00 after a pointer).
- Label length byte's top two bits:
  - `00` — literal label, length is the low 6 bits.
  - `11` — pointer; OFFSET is the low 14 bits across the next 2 bytes.
  - `01`, `10` — reserved; reject as a decode error in current usage.
- Pointer OFFSET is measured from the start of the message (octet 0
  of the header).
- A pointer SHOULD point to an offset that comes earlier in the
  message than the pointer itself. (The RFC does not formally forbid
  forward pointers, but a robust decoder should treat them as
  malformed because they admit infinite loops.)
- The expanded total length of any name (sum of literal label bytes
  encountered while following pointers, plus the length bytes, plus
  the terminating 0) MUST stay within 255 octets per §2.3.4.

## RR invariants

- TTL on the wire occupies 4 octets. RFC 1035 calls it a 32-bit
  signed integer; treating it as `uint32` for I/O is fine, but values
  with the high bit set are out-of-spec.
- RDLENGTH counts the bytes literally on the wire, including any
  compression pointer bytes inside RDATA (§4.1.4).
- A decoder MUST advance by exactly RDLENGTH octets after RDLENGTH,
  regardless of whether it understood the TYPE. This guarantees
  forward progress through the message even for unknown RR types.

## Encoder gotchas

- When using compression: before writing each name, check the
  suffix-offset dictionary. After writing, record any newly-emitted
  suffix offsets. Compare label bytes case-insensitively (§2.3.3) but
  preserve original case in the bytes you emit.
- Compression pointers are only safe in domain names whose semantic
  position is well-known to all DNS implementations. For TYPE values
  defined in RFC 1035, pointers are safe in NAME and in the
  domain-name fields of NS/CNAME/PTR/SOA/MX (and the experimental
  mailbox types). For unknown or class-specific RDATA, do NOT
  compress — the receiver may not know which bytes are names.
- `<character-string>` has no compression and cannot exceed 255 data
  octets (length byte is 8 bits).
- After encoding RDATA, backfill the 2-byte RDLENGTH with the actual
  number of bytes written. Do not include RDLENGTH itself.

## Decoder gotchas

- A pointer dereference must not loop. Bound the chain length (e.g.
  128) and require each pointer target to be strictly less than the
  position the pointer was read from.
- Reading a NAME advances the read cursor past the FIRST pointer (if
  any), not past the end of the pointed-to name. The "end of name on
  the wire" is the position immediately after the terminating 0x00
  for an uncompressed name, or immediately after the first pointer's
  two bytes for a compressed name.
- Always validate: total bytes consumed by the message <= input
  buffer length; QDCOUNT/ANCOUNT/NSCOUNT/ARCOUNT entries actually
  present.
- Unknown TYPE: read RDLENGTH octets verbatim. Do not attempt
  decompression on them.
- Reject reserved label-length top bits `01` and `10`.

## Worked example: header bit packing

Suppose a standard query (QR=0, OPCODE=0, RD=1) with one question and
ID 0x1234:

```
ID       = 0x1234
flags    = (0<<15)|(0<<11)|(0<<10)|(0<<9)|(1<<8)|(0<<7)|(0<<4)|0
         = 0x0100
QDCOUNT  = 1
ANCOUNT  = 0
NSCOUNT  = 0
ARCOUNT  = 0

bytes:
12 34   01 00   00 01   00 00   00 00   00 00
```

Suppose a positive answer (QR=1, OPCODE=0, AA=1, RCODE=0, RD=1, RA=1)
with one answer record:

```
flags = (1<<15)|(0<<11)|(1<<10)|(0<<9)|(1<<8)|(1<<7)|(0<<4)|0
      = 0x8580
```

These match the canonical "common DNS flag word" values seen in
typical query/response captures.
