# Operating System (OS byte)

Maps the `OS` byte at offset 9 of the [member header](../structures/header.md)
to the source operating system / file system on which the member was
created. Used for diagnostic purposes; decoders typically preserve the
value but do not act on it.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | FAT | FAT filesystem (MS-DOS, OS/2, NT/Win32) | RFC 1952 §2.3.1 |
| 1 | Amiga | Amiga | RFC 1952 §2.3.1 |
| 2 | VMS | VMS or OpenVMS | RFC 1952 §2.3.1 |
| 3 | Unix | Unix | RFC 1952 §2.3.1 |
| 4 | VM/CMS | VM/CMS | RFC 1952 §2.3.1 |
| 5 | Atari | Atari TOS | RFC 1952 §2.3.1 |
| 6 | HPFS | HPFS filesystem (OS/2, NT) | RFC 1952 §2.3.1 |
| 7 | Macintosh | Macintosh | RFC 1952 §2.3.1 |
| 8 | Z-System | Z-System | RFC 1952 §2.3.1 |
| 9 | CP/M | CP/M | RFC 1952 §2.3.1 |
| 10 | TOPS-20 | TOPS-20 | RFC 1952 §2.3.1 |
| 11 | NTFS | NTFS filesystem (NT) | RFC 1952 §2.3.1 |
| 12 | QDOS | QDOS | RFC 1952 §2.3.1 |
| 13 | Acorn RISCOS | Acorn RISCOS | RFC 1952 §2.3.1 |
| 14..254 | (unassigned) | Reserved for future use | RFC 1952 §2.3.1 |
| 255 | unknown | Unknown — encoder did not record an OS | RFC 1952 §2.3.1 |

## Notes

- `255` is the conventional value when the encoder cannot determine the
  source OS.
- Values 14..254 are unassigned. Decoders SHOULD accept them and
  preserve the value rather than rejecting the member.
- There is no IANA registry; the value space is owned by RFC 1952.
