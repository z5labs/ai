# Resource Record Format (RFC 1035 Section 4.1.3 and 3.2.1)

The Answer, Authority, and Additional sections all share the same RR wire
format. Each of those sections contains a sequence of zero or more RRs, each
encoded as follows:

## Wire layout

```
                                    1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                                               |
    /                                               /
    /                      NAME                     /
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      TYPE                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     CLASS                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                      TTL                      |
    |                                               |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                   RDLENGTH                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--|
    /                     RDATA                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

Slashes denote variable-length fields. Fixed-size portion (after NAME) is 10
octets: `2 (TYPE) + 2 (CLASS) + 4 (TTL) + 2 (RDLENGTH)`.

## Fields

### NAME (variable)

An owner name, i.e., the name of the node to which this resource record
pertains. Encoded as a domain name (see `domain-names.md`); MAY use
compression pointers.

### TYPE (16 bits)

Two-octet big-endian unsigned integer. Specifies the meaning of the data in
the RDATA field. See `type-class-values.md`.

### CLASS (16 bits)

Two-octet big-endian unsigned integer. Specifies the class of the data in the
RDATA field. See `type-class-values.md`. For Internet data, this is `IN` (1).

### TTL (32 bits)

A **32-bit unsigned integer** that specifies the time interval (in seconds)
that the resource record may be cached before it should be discarded. Zero
values are interpreted to mean that the RR can only be used for the
transaction in progress and should not be cached.

> Note: although the field is 32 bits and most implementations treat it as
> unsigned, RFC 2181 later clarified that the top bit must be zero (i.e., the
> value lies in `[0, 2^31 - 1]`). For raw RFC 1035 conformance, decoders
> should accept any 32-bit value but encoders SHOULD set values within
> `[0, 2^31 - 1]`.

### RDLENGTH (16 bits)

An unsigned 16-bit integer that specifies the length in octets of the RDATA
field.

### RDATA (variable, length = RDLENGTH)

A variable-length string of octets that describes the resource. The format of
this information varies according to the TYPE and CLASS of the resource
record. For example, if the TYPE is `A` and the CLASS is `IN`, the RDATA
field is a 4-octet ARPA Internet address.

See `rdata-formats.md` for per-TYPE RDATA layouts.

## Encoding/decoding notes

- The decoder reads NAME (variable), then 10 fixed bytes
  (TYPE/CLASS/TTL/RDLENGTH), then RDLENGTH octets of RDATA.
- The encoder MUST set RDLENGTH to the exact byte count of RDATA as actually
  emitted (which may be smaller than the "logical" size if compression
  pointers are used inside RDATA, e.g., for NS/CNAME/PTR/MX/SOA/MINFO).
- RDATA may itself contain domain names that use compression pointers. Those
  pointers are interpreted relative to the start of the **enclosing DNS
  message**, not the start of RDATA.
- An RR's total wire size is `len(NAME) + 10 + RDLENGTH`.
