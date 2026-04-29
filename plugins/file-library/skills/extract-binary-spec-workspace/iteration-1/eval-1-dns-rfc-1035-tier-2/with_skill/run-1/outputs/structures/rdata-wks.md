# RDATA: WKS (TYPE = 11)

Well Known Services RDATA payload — IPv4 address, IP protocol number,
and a bitmap of which TCP/UDP services are offered. Defined in RFC 1035
§3.4.2.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                            ADDRESS  (IPv4, network byte order)                                |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|  PROTOCOL  |
+------------+
|       <BIT MAP>  (variable length)         /
/                                            /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 4 | [4]byte | Address | IPv4 address in network order. |
| 4 | 1 | uint8 | Protocol | IP protocol number (e.g. 6 = TCP, 17 = UDP) per the IANA Assigned Numbers / IP protocol registry. |
| 5 | RDLENGTH - 5 | []byte | Bitmap | Service bitmap; see [Bit fields](#bit-fields). |

The enclosing RR's RDLENGTH equals 5 plus the bitmap length in bytes.

## Bit fields

The bitmap has one bit per service port. Bit numbering is MSB-0 within
each byte: bit 0 of byte 0 corresponds to port 0, bit 7 of byte 0 to
port 7, bit 0 of byte 1 to port 8, etc.

| Bit position (across all bitmap bytes) | Meaning |
|---|---|
| `n` (zero-based) | The service running on port `n` of the listed protocol is offered by this host if the bit is 1, absent if 0. |

If a bitmap of length `N` bytes is supplied, ports `0 .. 8N-1` are
described; ports beyond that are implicitly absent.

| Byte index | Bit (MSB-0) | Port |
|---|---|---|
| 0 | 0 | 0 |
| 0 | 1 | 1 |
| 0 | 7 | 7 |
| 1 | 0 | 8 |
| ... | ... | ... |

## Variable-length fields

### Bitmap

- **Length determination**: implicit from RDLENGTH; the bitmap occupies
  RDLENGTH - 5 octets.
- **Length prefix counts**: not applicable; no internal prefix.
- **Maximum length**: bounded by RDLENGTH.
- **Encoding**: packed bits, MSB-first within each byte.

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None. The bitmap is right-truncated rather than zero-padded; trailing
zero bytes that supply no information SHOULD be omitted (RFC 1035
§3.4.2 says the bitmap "must be a multiple of 8 bits" — i.e. a whole
number of octets — but trailing all-zero octets are unnecessary).

## Nested structures

None.

## Versioning notes

WKS is rarely used in modern DNS; defined as-is in RFC 1035.

## Ambiguities

> **Ambiguity:** RFC 1035 says the bitmap "must be a multiple of 8
> bits". This precludes a partial trailing byte but permits any number
> of trailing zero bytes; whether to trim them is left to the encoder.
