# FEXTRA subfield IDs (SI1, SI2)

Maps the 2-byte subfield ID `(SI1, SI2)` inside an
[FEXTRA](../structures/fextra.md) subfield to a registered or
conventional meaning.

By RFC 1952 convention, `SI1` and `SI2` are LATIN-1 characters chosen
to form a mnemonic. `SI2 = 0` is reserved for "random / local use" and
identifies an unregistered subfield ID.

| Value (`SI1` `SI2`) | Mnemonic | Description | Reference |
|---|---|---|---|
| any, `0x00` | (local use) | Reserved for random / local / unregistered subfield IDs — implementations MUST NOT assume any structure | RFC 1952 §2.3.1.1 |
| `0x41 0x70` ("Ap") | Apollo | Apollo file type information | RFC 1952 §2.3.1.1 (informational) |
| `0x52 0x4f` ("RO") | (rsync) | Used by rsync's gzip-compatible variant for record offsets — de-facto, not RFC-registered | de-facto |

## Notes

- The RFC defines a registry maintained by the gzip authors at the
  contact address listed in RFC 1952 §2.3.1.1; in practice the registry
  has not grown beyond a handful of entries and is rarely used.
- There is no IANA registry for gzip subfield IDs.
- A decoder that does not recognize a `(SI1, SI2)` pair MUST still
  consume `4 + LEN` bytes for that subfield (per
  [`../structures/fextra.md`](../structures/fextra.md)) and then move on.
- Encoders writing custom metadata SHOULD use `SI2 = 0x00` to mark the
  subfield as local-use, and SHOULD NOT collide with any registered
  ID above.

> **Ambiguity:** RFC 1952 lists only the "Ap" Apollo example. Other
> values seen in the wild (such as the rsync "RO" entry above) are
> de-facto rather than registered. Treat the table above as a starting
> point — implementations should be tolerant of unknown IDs.
