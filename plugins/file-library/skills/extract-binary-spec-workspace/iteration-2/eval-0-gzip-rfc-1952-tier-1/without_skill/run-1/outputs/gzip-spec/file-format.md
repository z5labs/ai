# File format (RFC 1952 Section 2.2)

A gzip file is a concatenation of one or more **members** (each member is
an independent compressed data set).

- Members appear one after another in the file.
- There is **no** additional information before, between, or after the
  members — no global header, no separator, no global trailer.
- A decoder must therefore be prepared to read another member immediately
  after the trailer of the previous one, until end of input.

The structure of a single member is defined in `member-layout.md` and
detailed in `member-header.md`, `flg-bits.md`, `optional-fields.md`, and
`trailer.md`.
