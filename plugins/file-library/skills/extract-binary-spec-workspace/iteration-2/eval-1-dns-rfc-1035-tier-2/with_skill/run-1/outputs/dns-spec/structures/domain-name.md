# DomainName

A domain name on the wire is a sequence of length-prefixed labels, optionally truncated by a 2-octet compression pointer that redirects to an earlier domain name in the same DNS message. Used as `QNAME` in questions, `NAME` in resource records, and inside RDATA payloads (NS, CNAME, PTR, MX EXCHANGE, SOA MNAME/RNAME, MB MADNAME, etc.).

This structure is a **recursive / algorithmic encoding**, not a fixed byte-offset layout — see Encoding below.

## Layout

Variable-length, byte-aligned. No padding. Domain name **may be an odd number of octets** (RFC 1035 §4.1.2).

## Encoding

A domain name is consumed octet-by-octet. The first two bits of each length octet select one of four branches:

| Top 2 bits | Branch | Meaning |
|---|---|---|
| `00` | Label | Lower 6 bits hold the label length L (0..63). If L is 0, this is the root label and the name is complete. Otherwise read L following octets as the label bytes, then continue with the next length octet. |
| `11` | Pointer | Combined with the next octet, forms a 14-bit OFFSET from the start of the message (offset 0 = first byte of header `ID`). The remainder of the name is read by jumping to that offset and continuing decode there. A pointer always terminates the current name. |
| `01` | Reserved | Not defined by RFC 1035; decoders MUST treat as a protocol error (or per current IANA registry) — see Ambiguities. |
| `10` | Reserved | Same as `01`. |

### Label wire form (top 2 bits = 00)

```
 0 1 2 3 4 5 6 7
+-+-+-+-+-+-+-+-+
|0 0|   LEN     |   one length octet, LEN ∈ [0, 63]
+-+-+-+-+-+-+-+-+
|     octet 1   |   label bytes, LEN total
+-+-+-+-+-+-+-+-+
|     octet 2   |
+-+-+-+-+-+-+-+-+
        ...
```

A length octet of `0x00` is the root label and terminates the name.

### Pointer wire form (top 2 bits = 11)

```
 0                   1
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|1 1|         OFFSET            |   2 octets, OFFSET ∈ [0, 16383]
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

OFFSET is a byte offset measured from the start of the **DNS message** (octet 0 = first byte of header `ID`). A pointer always terminates the current name.

### Allowed compositions (RFC 1035 §4.1.4)

A domain name in a message is exactly one of:

1. A sequence of zero or more labels followed by a zero-length octet (uncompressed).
2. A pointer (whole name is by reference).
3. A sequence of one or more labels followed by a pointer (suffix is by reference).

### Decode pseudocode

```
func decodeDomainName(msg []byte, start int) (labels [][]byte, bytesConsumed int, err error) {
    var (
        out          [][]byte
        i            = start
        followed     = false       // have we hopped via a pointer yet?
        firstPtrEnd  = 0           // bytes consumed in the *current* name
        seen         = map[int]bool{} // loop guard for malicious pointer chains
        nameLen      = 0           // total label-octet count, for 255 limit
    )
    for {
        if i >= len(msg) {
            return nil, 0, ErrTruncated
        }
        b := msg[i]
        switch b & 0xC0 {
        case 0x00: // label or root terminator
            length := int(b & 0x3F)
            if !followed {
                firstPtrEnd = i + 1 + length
            }
            i++
            if length == 0 {
                if !followed {
                    return out, firstPtrEnd - start, nil
                }
                return out, firstPtrEnd - start, nil
            }
            if i+length > len(msg) {
                return nil, 0, ErrTruncated
            }
            nameLen += length + 1
            if nameLen > 255 {
                return nil, 0, ErrNameTooLong
            }
            out = append(out, msg[i:i+length])
            i += length
        case 0xC0: // pointer
            if i+1 >= len(msg) {
                return nil, 0, ErrTruncated
            }
            offset := (int(b&0x3F) << 8) | int(msg[i+1])
            if !followed {
                firstPtrEnd = i + 2
                followed = true
            }
            if seen[offset] {
                return nil, 0, ErrPointerLoop
            }
            seen[offset] = true
            i = offset
        default: // 0x40 or 0x80 — reserved
            return nil, 0, ErrReservedLabelType
        }
    }
}
```

`bytesConsumed` is the on-wire length of the name **as it appears at `start`** — the RDATA RDLENGTH and the question-section advance MUST use this value (which counts the pointer itself, not the bytes the pointer redirects to). RFC 1035 §4.1.4: "the length of the compressed name is used in the length calculation, rather than the length of the expanded name".

## Constraints

- Label length L: 0 ≤ L ≤ 63 (top 2 bits zero force the upper bound). RFC 1035 §3.1, §2.3.4.
- Total domain name length on the wire (sum of label-byte octets plus their length octets, including the terminating zero) ≤ 255 octets. RFC 1035 §2.3.4.
- A label of length 0 is reserved for the root and may only appear once, as the terminator.
- Comparisons are case-insensitive (ASCII A–Z vs a–z), but original case SHOULD be preserved on the wire. RFC 1035 §2.3.3.

## Compression scope

Pointers may be used for any domain name in a message **only when the name's class- and type-specific format is known** to the receiver (so it can find the name to compress against). RFC 1035 §3.3 enumerates that NS, SOA, CNAME, and PTR have known formats and so their RDATA domain names may be compressed. RFC 1035 is silent on whether MX EXCHANGE, MB MADNAME, etc. may be compressed; widely deployed implementations compress them — see Ambiguities.

A name being decoded MAY only contain pointers to **earlier** points in the message (offsets less than the current label's offset). Forward pointers and pointer chains that revisit the same offset are protocol errors. (RFC 1035 does not state "earlier" explicitly; this is universal practice for safety against decode loops.)

## Ambiguities

> **Ambiguity:** RFC 1035 §4.1.4 reserves bit patterns `01` and `10` in a length octet for future use but does not say how a decoder must respond. Implementers should treat them as a protocol error, but a permissive decoder could skip the message instead of erroring the whole stream.

> **Ambiguity:** RFC 1035 §3.3 explicitly lists CNAME, NS, SOA, and PTR RDATA as compressible. It is silent on whether the domain names inside MX, MB, MG, MR, MD, MF, MINFO, and HINFO may be compressed. In practice DNS encoders DO compress MX EXCHANGE; some implementations refuse to compress newer types whose format is not known. Decoders MUST always be prepared to decompress regardless.

> **Ambiguity:** RFC 1035 says "the high order two bits of every length octet must be zero" in §3.1, but §4.1.4 then reuses `11` for pointers and reserves `01`/`10`. The reconciliation is that §3.1's text describes a label, not the wire octet's interpretation. Treat the §4.1.4 four-branch dispatch as authoritative.
