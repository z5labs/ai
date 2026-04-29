# Extra subfield IDs (SI1, SI2)

IDs registered for the `(SI1, SI2)` pair of [`../structures/extra-subfield.md`](../structures/extra-subfield.md).
At the time RFC 1952 was published, only one ID had been registered.

| Value | Name | Description | Reference |
|---|---|---|---|
| `(0x41, 0x70)` ('A','P') | APOLLO | Apollo file type information | RFC 1952 §2.3.1.1 |

## Notes

- Subfield IDs with `SI2 == 0` are reserved for future use (RFC 1952 §2.3.1.1).
- The registry of `(SI1, SI2)` pairs is maintained out-of-band; RFC 1952
  references Jean-Loup Gailly as the registrar at publication time.
- A decoder encountering an unrecognized `(SI1, SI2)` should still parse
  `LEN` and skip exactly that many bytes — the framing is well-defined
  even when the payload's meaning is not.
