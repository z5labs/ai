# Resource Record (RR) Format

Source: RFC 1035 §4.1.3 and §3.2.1.

The Answer, Authority, and Additional sections share the same per-entry
format. Each RR has the following layout.

## Wire layout

```
                                1  1  1  1  1  1
  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
/                                               /
/                      NAME                     /   variable length
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      TYPE                     |   2 octets
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     CLASS                     |   2 octets
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      TTL                      |   4 octets
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                   RDLENGTH                    |   2 octets
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     RDATA                     /   RDLENGTH octets
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Fields

### NAME (variable length)
Owner domain name to which this record pertains. Same encoding as QNAME:
length-prefixed labels terminated by a zero byte, with optional pointer
compression per §4.1.4. See `04-domain-names.md`.

### TYPE (16 bits, big-endian)
RR type code. See `05-types-and-classes.md` for the full table. This
value tells the decoder how to interpret RDATA.

### CLASS (16 bits, big-endian)
RR class code (typically `IN` = 1). See `05-types-and-classes.md`.

### TTL (32 bits, big-endian)
Time-to-live. RFC 1035 specifies it as a 32-bit signed integer; only
non-negative values are valid (0 .. 2^31 - 1). A value of 0 means the
record may not be cached.

Implementations commonly treat TTL as a 32-bit unsigned integer on the
wire and clamp the upper half to "do not cache" / treat as zero, but the
strict RFC reading is signed. For an encoder/decoder library, reading and
writing as a `uint32` is acceptable; validation of the high bit is a
separate concern.

### RDLENGTH (16 bits, big-endian)
Unsigned length, in octets, of the RDATA field that follows. RDLENGTH
counts the bytes literally present on the wire — if a domain name in
RDATA is compressed, RDLENGTH counts the compressed length, not the
expanded length (RFC 1035 §4.1.4).

### RDATA (RDLENGTH octets)
Variable-length payload whose interpretation depends on TYPE and CLASS.
See `06-rdata-formats.md` for per-type layouts.

## Encoding

1. Encode NAME (labels + 0, optional compression).
2. Append TYPE (uint16 BE).
3. Append CLASS (uint16 BE).
4. Append TTL (uint32 BE).
5. Reserve 2 bytes for RDLENGTH.
6. Encode RDATA according to TYPE/CLASS.
7. Backfill RDLENGTH with the number of bytes actually written for RDATA.

## Decoding

1. Decode NAME (handle compression pointers).
2. Read TYPE (uint16 BE).
3. Read CLASS (uint16 BE).
4. Read TTL (uint32 BE).
5. Read RDLENGTH (uint16 BE).
6. Read exactly RDLENGTH octets of RDATA.
7. Interpret RDATA per `06-rdata-formats.md` (skip/preserve raw if TYPE
   is unknown).
