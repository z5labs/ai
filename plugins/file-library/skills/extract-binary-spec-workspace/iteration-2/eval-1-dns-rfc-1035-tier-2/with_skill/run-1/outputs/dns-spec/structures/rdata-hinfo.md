# RDataHINFO

RDATA payload of a `Type=HINFO` resource record: two `<character-string>` fields giving CPU and OS identifiers. RFC 1035 §3.3.2.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                      CPU                      /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                       OS                      /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | CharacterString | CPU | CPU type identifier. See [`character-string.md`](character-string.md). |
| variable | variable | CharacterString | OS | Operating-system type identifier. See [`character-string.md`](character-string.md). |

## Notes

- `RDLength` MUST equal `(1 + len(CPU.Data)) + (1 + len(OS.Data))`.
- Standard values are listed in RFC 1010 (out of scope for this reference; treat both fields as opaque bytes).
