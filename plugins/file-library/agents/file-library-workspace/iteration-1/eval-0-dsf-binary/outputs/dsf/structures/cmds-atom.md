# CmdsAtom

Top-level `CMDS` atom. The payload is a series of variable-length commands
packed end to end, with no count and no terminator — the decoder reads
commands until `parent.Size − 8` bytes have been consumed.

Each command consists of a single 1-byte opcode followed by an
opcode-specific payload (see [`command.md`](command.md) for the framing and
[`../encoding-tables/command-opcodes.md`](../encoding-tables/command-opcodes.md)
for the full opcode list).

The number of bytes used by a command is **only** known by interpreting the
opcode — there is no length prefix and no trailing magic, so unknown opcodes
cannot be skipped: the spec says "unknown commands cannot be skipped", which
implies a fatal error.

> **Ambiguity:** The spec also says "Command ID 255 is reserved for future
> expansion". It does not specify what a decoder should do with opcode 255
> in the current version — likely treat as fatal until a future version
> defines a specific length encoding.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `[]Command` | Commands | Tightly packed [`Command`](command.md) records. |

## Variable-length fields

- **Length determination**: per-command (opcode-driven). Decoder must read
  the opcode, then dispatch to a per-opcode reader that consumes its
  remaining payload bytes.
- **Maximum length**: bounded by the parent atom's `Size`.

## Nested structures

See [`command.md`](command.md) for the per-opcode framing and the full
command catalogue.

## Ambiguities

> **Ambiguity:** The spec text in section 2.4.6 ("Comment Commands") says the
> length field for `COMMENT 16` is implied but the source HTML omits its
> explicit `uint16 length` line. The natural reading — and the one X-Plane
> writers use — is `uint16 length` followed by `length` bytes of comment
> data. This reference uses that reading.
