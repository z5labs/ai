# AtomOfAtoms

A container atom whose payload is a back-to-back sequence of sub-atoms. Used
by `HEAD`, `DEFN`, `GEOD`, and `DEMS`.

## Encoding

The payload of an atom-of-atoms is just sub-atoms laid end-to-end. There is
no count and no per-sub-atom delimiter; the parent atom's `Size` defines the
total payload length, and the decoder walks sub-atoms by repeatedly reading
each child's 8-byte header and consuming its `Size` bytes.

```
sub-atom 1 (Atom)
sub-atom 2 (Atom)
...
sub-atom N (Atom)   <-- parent payload ends exactly at sub-atom N's tail
```

Decoding pseudocode:

```
remaining := parent.Size - 8           // payload length
while remaining > 0:
    child := readAtom()                // 8-byte header + Size-8 payload
    consumed := child.Size
    if consumed < 8 or consumed > remaining:
        error: malformed atom
    remaining -= consumed
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `[]Atom` | SubAtoms | Sub-atoms walked by `Size`. |

## Nested structures

The set of sub-atom IDs that may appear depends on the parent:

- `HEAD` → `PROP`
- `DEFN` → `TERT`, `OBJT`, `POLY`, `NETW`, `DEMN`
- `GEOD` → `POOL`, `SCAL`, `PO32`, `SC32`
- `DEMS` → `DEMI`, `DEMD` (in pairs, in declaration order)

## Ambiguities

> **Ambiguity:** The spec does not require sub-atoms inside a container atom
> to use only the IDs listed above. A decoder should ignore unknown sub-atoms
> rather than treating them as fatal, to preserve the spec's "atoms with all-
> capital ASCII letters and digits are reserved; other IDs may be used for
> private data" extensibility rule.
