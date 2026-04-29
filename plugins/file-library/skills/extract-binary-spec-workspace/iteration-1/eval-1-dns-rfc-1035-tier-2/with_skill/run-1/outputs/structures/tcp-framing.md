# TCP Framing

When a DNS message is sent over TCP, it is preceded by a 2-byte length
field that tells the receiver how many bytes of DNS message follow.
Defined in RFC 1035 §4.2.2.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                MessageLength                  |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
/                  DNS message                  /
/             (MessageLength bytes)             /
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 | MessageLength | Number of octets of DNS message that follow this length prefix. Big-endian. |
| 2 | MessageLength | []byte | Message | The DNS message itself, beginning with [`header`](header.md). |

## Bit fields

None.

## Variable-length fields

### Message

- **Length determination**: explicit 2-byte big-endian prefix.
- **Length prefix counts**: only the bytes of the DNS message
  (header + sections); it does *not* include itself.
- **Maximum length**: 65535 octets.
- **Encoding**: a complete DNS message starting with [`header`](header.md).

## Conditional / optional fields

The 2-byte length field is *only* present on TCP. Over UDP, the DNS
message is the entire UDP payload and there is no length prefix.

## Checksums and integrity

None at this layer; TCP supplies its own.

## Padding and alignment

None.

## Nested structures

The framed payload is a complete DNS message — see [`header`](header.md).

## Versioning notes

Unchanged since RFC 1035.

## Ambiguities

> **Ambiguity:** RFC 1035 §4.2.2 says nothing about whether the receiver
> may close the connection after a single message. RFC 7766 (out of
> scope here) clarifies long-lived TCP connection handling.
