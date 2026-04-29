# Message

The top-level DNS protocol data unit. Every DNS communication — query, response, notification — is one Message. RFC 1035 §4.1.

## Byte diagram

```
+---------------------+
|       Header        |   12 octets, fixed
+---------------------+
|      Question       |   QDCOUNT entries (0 or more)
+---------------------+
|       Answer        |   ANCOUNT resource records
+---------------------+
|     Authority       |   NSCOUNT resource records
+---------------------+
|     Additional      |   ARCOUNT resource records
+---------------------+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 12 | Header | Header | Fixed-size header with section counts and flags. See [`header.md`](header.md). |
| 12 | variable | []Question | Question | `Header.QDCOUNT` entries; usually 1 in queries, 0 or 1 in responses. See [`question.md`](question.md). |
| variable | variable | []ResourceRecord | Answer | `Header.ANCOUNT` resource records. See [`resource-record.md`](resource-record.md). |
| variable | variable | []ResourceRecord | Authority | `Header.NSCOUNT` resource records pointing toward an authoritative name server. |
| variable | variable | []ResourceRecord | Additional | `Header.ARCOUNT` resource records carrying related but not strictly answering data. |

## Variable-length fields

- **Length determination:** Each section's record count comes from the corresponding header field (QDCOUNT, ANCOUNT, NSCOUNT, ARCOUNT). There is no whole-message length prefix on UDP. Over TCP, the message is preceded by an external 2-octet big-endian length field — see Transport.
- **Encoding:** Records are concatenated with no padding or separators; the decoder advances by the size each record consumes.

## Transport

| Transport | Framing | Size limit |
|---|---|---|
| UDP | Bare message in a datagram. | 512 octets max payload (RFC 1035 §4.2.1). Longer responses are truncated and the header `TC` flag is set. |
| TCP | Message preceded by a **2-octet big-endian length** field that gives the message length, **excluding the 2 length octets**. (RFC 1035 §4.2.2.) | Bound by the 16-bit length, so up to 65535 octets per message. |

## Nested structures

- [`header.md`](header.md) — the 12-byte header (always first).
- [`question.md`](question.md) — question-section entry.
- [`resource-record.md`](resource-record.md) — answer/authority/additional entry format.

## Compression context

Domain-name compression pointers (§4.1.4) reference offsets **from the start of this Message** — offset 0 is the first byte of `Header.ID`. The encoder must therefore know the message's base address while emitting names; the decoder must keep the message buffer accessible while it consumes domain names.

## Ambiguities

> **Ambiguity:** RFC 1035 §4.2.1 caps UDP messages at 512 octets but is silent on what happens when the response would still be too large after omitting records. The header `TC` bit signals truncation; resolvers retry over TCP. EDNS0 (RFC 6891) later raised this limit, but is out of scope for RFC 1035.
