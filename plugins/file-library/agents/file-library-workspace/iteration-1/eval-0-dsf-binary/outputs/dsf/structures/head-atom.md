# HeadAtom

Top-level `HEAD` atom (ID = ASCII `'HEAD'` read little-endian, i.e.
`0x44 0x41 0x45 0x48` on the wire). Currently carries exactly one sub-atom,
`PROP`.

## Encoding

`HEAD` is an [atom-of-atoms](atom-of-atoms.md). Its payload is a sequence of
sub-atoms; today the only defined child is [`PROP`](prop-atom.md).

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `AtomOfAtoms` | SubAtoms | Today contains exactly one [`PROP`](prop-atom.md). |
