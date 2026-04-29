# Domain Name Encoding and Compression (RFC 1035 Sections 2.3.4, 3.1, 4.1.4)

Domain names appear in QNAME (question section), NAME (RR header), and inside
many RDATA fields. The same encoding is used everywhere.

## Label encoding (Section 3.1)

A domain name is a sequence of **labels**. Each label on the wire is encoded
as:

```
    +--+--+--+--+--+--+--+--+
    |  length  |  octets... |
    +--+--+--+--+--+--+--+--+
```

- A single **length octet** giving the number of octets in the label.
- Followed by exactly that many octets of label data (the label content).

The full domain name is the concatenation of these length-prefixed labels,
**terminated by a zero-length label** (a single `0x00` byte) that represents
the root.

### Length octet encoding

The length octet has its top two bits used as a tag:

| Top 2 bits | Meaning                                           |
|:---------:|---------------------------------------------------|
| `00`      | A label of length 0-63 octets follows. Lower 6 bits give the length. |
| `11`      | This octet, plus the next, form a 14-bit pointer (compression). |
| `01`      | Reserved (RFC 1035 leaves it for future use; not currently valid). |
| `10`      | Reserved (RFC 1035 leaves it for future use; not currently valid). |

Per RFC 1035 section 4.1.4: "The first two bits are zero. This allows the
label to be read as an ordinary length-prefixed string... The 10 and 01
combinations are reserved for future use."

### Root

The root domain is encoded as the single octet `0x00`.

## Size limits (Section 2.3.4)

The following size limits MUST be enforced by encoders and SHOULD be enforced
by decoders:

| Limit                                              | Maximum |
|----------------------------------------------------|--------:|
| Single label (octets, excluding the length byte)   | 63      |
| Full domain name (octets, including length octets and the terminating zero) | 255 |
| TTL field values                                   | positive `int32` (per RFC 2181 clarification) |
| UDP message payload (without TCP framing)          | 512 (see `transport.md`) |

Because the maximum label length is 63, a non-pointer length octet's value is
always in `[0, 63]` and its top two bits are always `00`. Any length octet
with the top two bits not equal to `00` or `11` is invalid under RFC 1035.

## Label compression (Section 4.1.4)

In order to reduce the size of messages, the domain system uses a
compression scheme which eliminates the repetition of domain names in a
message. In this scheme, an entire domain name or a list of labels at the end
of a domain name is replaced with a **pointer** to a prior occurrence of the
same name.

### Pointer format

A pointer is 2 octets:

```
                                    1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    | 1  1|                OFFSET                   |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- The **first 2 bits** are `1 1` (binary), distinguishing the pointer from a
  label whose length octet has its top 2 bits `0 0`.
- The remaining **14 bits** encode an `OFFSET` (big-endian) from the start of
  the **DNS message** (i.e., from the first byte of the header), pointing to
  another label or sequence of labels.

Equivalently, a pointer is the 16-bit value `0xC000 | offset` on the wire.
The byte sequence `0xC0 0x0C`, for example, is a pointer to byte offset 12
(the start of the first question - immediately after the 12-byte header).

### Compression rules

- A pointer terminates a domain name; once encountered, the decoder follows
  the pointer to read the remaining labels but does not read further bytes
  in-place after the 2-pointer octets for that name.
- A domain name may consist of one of three forms:
  1. A sequence of labels ending in a zero octet.
  2. A pointer.
  3. A sequence of labels ending with a pointer.
- Pointers can only point to a prior occurrence of a name within the same
  message. They MUST point to an earlier offset (lower than the position of
  the pointer itself) to avoid loops; nonetheless, decoders SHOULD still
  guard against pointer loops (typically by capping the number of pointer
  jumps).
- The OFFSET is relative to the start of the message, regardless of which
  section the pointer appears in.

### When compression is used

Per the RFC: "If a domain name is contained in a part of the message subject
to a length field (such as the RDATA section of an RR), and compression is
used, the length of the compressed name is used in the length calculation,
rather than the length of the expanded name."

Programs are free to avoid using pointers in messages they generate, although
this will reduce datagram capacity, and may cause truncation. However, all
programs are required to understand arriving messages that contain pointers.

For example, a compression scheme would be used in NS, SOA, MX, CNAME, PTR,
and similar RRs whose RDATA consists of (or contains) a domain name.

### Decoder algorithm

```
parseName(buf, off):
    name = []
    jumped = false
    cursor = off
    end_of_name = -1
    hops = 0
    while true:
        if hops > MAX_HOPS:        # e.g., 128
            error "compression loop"
        b = buf[cursor]
        if b == 0x00:
            cursor += 1
            if not jumped:
                end_of_name = cursor
            break
        if (b & 0xC0) == 0xC0:     # pointer
            if cursor + 1 >= len(buf):
                error "truncated pointer"
            ptr = ((b & 0x3F) << 8) | buf[cursor+1]
            if not jumped:
                end_of_name = cursor + 2
            cursor = ptr
            jumped = true
            hops += 1
            continue
        if (b & 0xC0) != 0x00:     # 01 or 10 - reserved
            error "invalid label tag"
        n = b & 0x3F               # 0..63
        if n > 63:
            error "label too long"   # cannot occur due to mask, but defensive
        cursor += 1
        if cursor + n > len(buf):
            error "truncated label"
        name.append(buf[cursor : cursor+n])
        cursor += n
    if not jumped:
        end_of_name = cursor
    return name, end_of_name
```

Note that `end_of_name` (the offset just past the in-place encoding) is
returned separately from the cursor used to read labels, because pointers
may take the read cursor far away from the original position.

### Encoder strategy

A simple, RFC-compliant encoder can:
1. Emit each domain name fully expanded (uncompressed), terminated by `0x00`.

A compressing encoder additionally:
1. Maintains a map from suffix-of-name to the byte offset where that suffix
   was previously emitted.
2. Before writing a name, walks its suffixes from longest to shortest. If a
   suffix is found in the map at offset `< 0x4000`, it emits the unmatched
   prefix labels followed by a pointer (`0xC000 | offset`) to that suffix.
3. Otherwise emits the full name and registers each of its suffix offsets
   in the map.

Implementations MUST NOT emit a pointer with offset `>= 0x4000` (the offset
field is only 14 bits). For very long messages where a target name lies past
offset `0x3FFF`, the encoder must fall back to emitting that name expanded.
