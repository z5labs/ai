# Terrain patch flags

Bit-packed `uint8` carried by the `TERRAIN PATCH FLAGS` (opcode 17) and
`TERRAIN PATCH FLAGS AND LOD` (opcode 18) commands. See
[`../structures/command.md`](../structures/command.md).

| Bit value | Bit position | Name      | Description |
|---|---|---|---|
| 0x01    | bit 0  | PHYSICAL    | If set, the patch participates in collision detection. If cleared, it is drawn but not collision-checked. |
| 0x02    | bit 1  | OVERLAY     | If set, the patch is drawn over another patch with z-buffer precautions. |
| 0xFC    | bits 2–7 | (reserved) | Must be 0. |

## Notes

The interpretation of the OVERLAY flag depends on the terrain type
referenced by the current definition.
