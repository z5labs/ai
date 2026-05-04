# FileHeader

The fixed 12-byte header that opens every uncompressed DSF file.

## Byte diagram

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|                  Cookie ('XPLNEDSF', 8 bytes)                 |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       Version (uint32)                        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 8 | `[8]byte` | Cookie | ASCII `XPLNEDSF` (`0x58 0x50 0x4C 0x4E 0x45 0x44 0x53 0x46`) |
| 8 | 4 | `uint32` | Version | Master file format version. Currently `1`. |
