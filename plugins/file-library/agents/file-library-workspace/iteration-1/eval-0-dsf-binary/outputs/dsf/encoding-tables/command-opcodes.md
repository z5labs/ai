# Command opcodes

The 1-byte opcode that opens every command inside the
[`CMDS`](../structures/cmds-atom.md) atom. See
[`../structures/command.md`](../structures/command.md) for the per-opcode
payload framing.

| Value | Name                            | Family       | Reference (spec section) |
|---|---|---|---|
| 0   | (unassigned)                       | —            | — |
| 1   | COORDINATE POOL SELECT             | State        | 2.4.1.1 |
| 2   | JUNCTION OFFSET SELECT             | State        | 2.4.1.2 |
| 3   | SET DEFINITION 8                   | State        | 2.4.1.3 |
| 4   | SET DEFINITION 16                  | State        | 2.4.1.4 |
| 5   | SET DEFINITION 32                  | State        | 2.4.1.5 |
| 6   | SET ROAD SUBTYPE 8                 | State        | 2.4.1.6 |
| 7   | OBJECT                             | Object       | 2.4.2.1 |
| 8   | OBJECT RANGE                       | Object       | 2.4.2.2 |
| 9   | NETWORK CHAINS                     | Network      | 2.4.3.1 |
| 10  | NETWORK CHAINS RANGE               | Network      | 2.4.3.2 |
| 11  | NETWORK CHAIN 32                   | Network      | 2.4.3.3 |
| 12  | POLYGON                            | Polygon      | 2.4.4.1 |
| 13  | POLYGON RANGE                      | Polygon      | 2.4.4.2 |
| 14  | NESTED POLYGON                     | Polygon      | 2.4.4.3 |
| 15  | NESTED POLYGON RANGE               | Polygon      | 2.4.4.4 |
| 16  | TERRAIN PATCH                      | Mesh         | 2.4.5.1 |
| 17  | TERRAIN PATCH FLAGS                | Mesh         | 2.4.5.2 |
| 18  | TERRAIN PATCH FLAGS AND LOD        | Mesh         | 2.4.5.3 |
| 19  | (unassigned)                       | —            | — |
| 20  | (unassigned)                       | —            | — |
| 21  | (unassigned)                       | —            | — |
| 22  | (unassigned)                       | —            | — |
| 23  | PATCH TRIANGLE                     | Mesh         | 2.4.5.4 |
| 24  | TRIANGLE PATCH CROSS-POOL          | Mesh         | 2.4.5.5 |
| 25  | PATCH TRIANGLE RANGE               | Mesh         | 2.4.5.6 |
| 26  | PATCH TRIANGLE STRIP               | Mesh         | 2.4.5.7 |
| 27  | PATCH TRIANGLE STRIP CROSS-POOL    | Mesh         | 2.4.5.8 |
| 28  | PATCH TRIANGLE STRIP RANGE         | Mesh         | 2.4.5.9 |
| 29  | PATCH TRIANGLE FAN                 | Mesh         | 2.4.5.10 |
| 30  | PATCH TRIANGLE FAN CROSS-POOL      | Mesh         | 2.4.5.11 |
| 31  | PATCH TRIANGLE FAN RANGE           | Mesh         | 2.4.5.12 |
| 32  | COMMENT 8                          | Comment      | 2.4.6.1 |
| 33  | COMMENT 16                         | Comment      | 2.4.6.2 |
| 34  | COMMENT 32                         | Comment      | 2.4.6.3 |
| 35–254 | (unassigned)                    | —            | — |
| 255 | RESERVED                           | —            | spec §2.3.5 reserves for future expansion |

## Notes

- The 19–22 gap is documented as a deliberate skip; the spec lists no
  opcode for those values and a decoder should treat them as unassigned
  (rejected as malformed).
- Decoders cannot skip unknown opcodes: there is no length prefix and
  unknown commands are explicitly fatal.
