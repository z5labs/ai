# Domain Names: Labels and Compression

Source: RFC 1035 §3.1 (label encoding) and §4.1.4 (message compression).

Domain names in DNS messages are encoded as a sequence of labels.
Compression allows later occurrences of a name (or trailing labels) to
be replaced by a 2-byte pointer back to an earlier occurrence in the
same message.

## 1. Label encoding (§3.1)

Each label is a single length octet followed by that many data octets:

```
+---------+----------------------+
| LEN (1) | data (LEN octets)    |
+---------+----------------------+
```

A domain name is a concatenation of labels, terminated by the null
label of the root (a single 0x00 length byte).

Constraints:

- The high two bits of every label-length octet are reserved. For a
  literal label, those two bits MUST be `00`. The remaining 6 bits give
  the length, so labels are 0..63 octets.
- A length octet of 0 terminates the name (the root label).
- The high two bits `11` indicate a compression pointer (see below).
- The combinations `10` and `01` are reserved for future use; a
  conforming decoder treats them as an error in current usage.
- The total encoded length of a domain name (sum of all length bytes
  plus all label bytes plus the terminating zero) is at most 255 octets
  (§2.3.4).
- Labels can technically contain any 8-bit values, though preferred
  syntax is letters/digits/hyphen.
- Comparison is case-insensitive (§2.3.3); on-the-wire bytes are
  preserved as-given.

### Example: `F.ISI.ARPA` uncompressed

```
01 'F'  03 'I' 'S' 'I'  04 'A' 'R' 'P' 'A'  00
```
Total length: 1+1 + 1+3 + 1+4 + 1 = 12 octets.

## 2. Message compression (§4.1.4)

To shrink messages, a name (or its trailing labels) may be replaced by
a pointer to an earlier occurrence. A pointer is a 2-octet sequence:

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
| 1  1|                OFFSET                   |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- The first two bits are `11`. This is what distinguishes a pointer
  from a label, since a label's first two bits must be `00`.
- The remaining 14 bits hold OFFSET, an offset (in octets) measured
  from the start of the entire DNS message (octet 0 is the first byte
  of the header's ID field). Therefore OFFSET is in `0 .. 16383`.
- OFFSET must point at the start of a length octet (either a literal
  label-length octet or another pointer).

Combined-byte encoding view: the pointer is two consecutive bytes
`b0 b1` where:

- `b0 & 0xC0 == 0xC0` (top two bits set) marks "this is a pointer".
- `OFFSET = ((b0 & 0x3F) << 8) | b1`.

A name on the wire may take any of these three forms (§4.1.4):

1. A sequence of labels ending in a zero octet (no compression).
2. A pointer alone (entire name is replaced by a back-reference).
3. A sequence of labels ending with a pointer (literal prefix, then
   the rest of the name comes from the pointer target).

## 3. RDLENGTH and compressed names (§4.1.4)

When a domain name lives inside an RDATA field that has its own length
prefix (RDLENGTH), the length counts the **compressed** length on the
wire, not the expanded length. Encoders and decoders must respect that.

## 4. Where compression is allowed (§4.1.4)

Pointers may only be used for names whose format is not class-specific.
For the RR types defined in RFC 1035, compression is permitted for
domain names that appear as the NAME field of any RR and for embedded
domain names in the RDATA of these well-known types: CNAME (CNAME), NS
(NSDNAME), PTR (PTRDNAME), SOA (MNAME, RNAME), MX (EXCHANGE), and the
similar mailbox-bearing types (MB, MD, MF, MG, MR, MINFO). The RFC's
phrasing ("all domain names in the RDATA section of these RRs may be
compressed") in §3.3 covers NS, SOA, CNAME, and PTR explicitly; in
practice MX/MINFO/etc. names also compress when emitted.

For unknown or class-specific RDATA, an encoder MUST NOT use
compression because the receiver may not know to expand pointers there.

A conservative encoder is allowed to never emit pointers at all
(§4.1.4): "Programs are free to avoid using pointers in messages they
generate, although this will reduce datagram capacity, and may cause
truncation. However all programs are required to understand arriving
messages that contain pointers."

## 5. Example from RFC 1035 §4.1.4

The names `F.ISI.ARPA`, `FOO.F.ISI.ARPA`, `ARPA`, and the root encoded
into a single message:

Offset 20:  `F.ISI.ARPA` written out literally:
```
20: 01 'F'
22: 03 'I' 'S' 'I'
26: 04 'A' 'R' 'P' 'A'
30: 00
```

Offset 40:  `FOO.F.ISI.ARPA` reuses `F.ISI.ARPA` from offset 20:
```
40: 03 'F' 'O' 'O'
44: C0 14            <- pointer, OFFSET = 20
```

Offset 64:  `ARPA` reuses the trailing `ARPA` at offset 26:
```
64: C0 1A            <- pointer, OFFSET = 26
```

Offset 92:  the root domain (no labels, just a single zero byte):
```
92: 00
```

## 6. Decoder algorithm (recommended)

```
decode_name(msg, start_offset):
    labels = []
    offset = start_offset
    end_of_name_offset = -1            # offset just after the name on
                                       # the wire (used by the caller
                                       # to advance), set when we follow
                                       # the FIRST pointer.

    while True:
        if offset >= len(msg): error("truncated name")
        b0 = msg[offset]
        type_bits = b0 & 0xC0

        if type_bits == 0x00:
            # literal label
            length = b0 & 0x3F
            offset += 1
            if length == 0:
                # end of name
                if end_of_name_offset == -1:
                    end_of_name_offset = offset
                break
            if offset + length > len(msg): error("truncated label")
            labels.append(msg[offset : offset+length])
            offset += length
            continue

        if type_bits == 0xC0:
            # pointer
            if offset + 1 >= len(msg): error("truncated pointer")
            ptr = ((b0 & 0x3F) << 8) | msg[offset+1]
            if end_of_name_offset == -1:
                end_of_name_offset = offset + 2
            # follow the pointer; protect against loops/forward-jumps
            if ptr >= start_of_followed_chain_min: error("loop/forward")
            offset = ptr
            continue

        # 0x40 or 0x80 — reserved
        error("reserved label type")

    return labels, end_of_name_offset
```

Practical safeguards (RFC is silent on these; required for robustness):

- Cap the number of pointer dereferences (e.g. 128) to bound work.
- Reject pointers that point at or past the current decode position
  (a pointer must reference an earlier offset to break loops).
- After fully decoding, re-check that the total expanded length is
  <= 255 octets per §2.3.4.

## 7. Encoder algorithm (recommended)

- Maintain a dictionary mapping previously-emitted suffixes (label
  sequences, byte-comparison case-insensitive per §2.3.3) to the offset
  at which they were written.
- When asked to emit a name, walk its labels from the leftmost; for
  each suffix `labels[i:]`, if the dictionary contains an offset
  reachable in 14 bits (`< 0x4000`), emit the labels `labels[:i]`
  literally then a 2-byte pointer to that offset, and stop.
- Otherwise, emit all remaining labels literally, then a terminating
  0x00. As you emit, record each suffix-offset pair in the dictionary.
- Conservative encoders MAY simply emit every name uncompressed and
  skip the dictionary entirely.
