# OS values

Values for the `OS` field of [`../structures/member-header.md`](../structures/member-header.md).
Indicates the type of filesystem on which compression took place; mainly
useful for guessing end-of-line conventions on text input.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | FAT | FAT filesystem (MS-DOS, OS/2, NT/Win32) | RFC 1952 §2.3.1 |
| 1 | AMIGA | Amiga | RFC 1952 §2.3.1 |
| 2 | VMS | VMS or OpenVMS | RFC 1952 §2.3.1 |
| 3 | UNIX | Unix | RFC 1952 §2.3.1 |
| 4 | VM_CMS | VM/CMS | RFC 1952 §2.3.1 |
| 5 | ATARI_TOS | Atari TOS | RFC 1952 §2.3.1 |
| 6 | HPFS | HPFS filesystem (OS/2, NT) | RFC 1952 §2.3.1 |
| 7 | MACINTOSH | Macintosh | RFC 1952 §2.3.1 |
| 8 | Z_SYSTEM | Z-System | RFC 1952 §2.3.1 |
| 9 | CP_M | CP/M | RFC 1952 §2.3.1 |
| 10 | TOPS_20 | TOPS-20 | RFC 1952 §2.3.1 |
| 11 | NTFS | NTFS filesystem (NT) | RFC 1952 §2.3.1 |
| 12 | QDOS | QDOS | RFC 1952 §2.3.1 |
| 13 | ACORN_RISCOS | Acorn RISCOS | RFC 1952 §2.3.1 |
| 255 | UNKNOWN | Unknown | RFC 1952 §2.3.1 |

## Notes

- Values 14 through 254 are unassigned by RFC 1952.
- A compliant encoder that does not know the source OS should write `255`
  (`UNKNOWN`).
- A compliant decoder is **not** required to validate `OS` — it may always
  treat the field as advisory and produce binary output regardless
  (RFC 1952 §2.3.1.2).
