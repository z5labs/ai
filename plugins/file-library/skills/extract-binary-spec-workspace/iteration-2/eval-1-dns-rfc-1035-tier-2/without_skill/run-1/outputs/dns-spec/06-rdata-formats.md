# RDATA Formats (per TYPE)

Source: RFC 1035 §3.3 (standard RRs) and §3.4 (Internet-specific RRs).

The wire interpretation of an RR's RDATA is determined by its TYPE
(and, in some forward-looking cases, CLASS). Below are the on-the-wire
layouts for the standard RR types defined in RFC 1035.

Common shorthand used by the RFC:

- `<domain-name>` — A sequence of labels terminated by a zero byte,
  exactly as described in `04-domain-names.md`. Compression pointers
  ARE allowed for these names within the RDATA of types defined here.
- `<character-string>` — A single length octet (0..255) followed by
  that many data octets. May include any 8-bit values; treated as
  binary. The character-string max length is 255 data octets, so the
  max on-wire length is 256 bytes (1-byte length + 255 data).

All multi-byte integers are big-endian (§2.3.2).

---

## A (TYPE 1) — §3.4.1

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ADDRESS                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- ADDRESS: 32-bit IPv4 address, 4 octets, in network byte order.
- RDLENGTH: 4.

---

## NS (TYPE 2) — §3.3.11

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   NSDNAME                     /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- NSDNAME: `<domain-name>` of an authoritative name server for the
  owner. Compression allowed.

---

## MD (TYPE 3) — §3.3.4 (Obsolete)

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   MADNAME                     /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MADNAME: `<domain-name>`. Obsolete — converted to MX with
  preference 0 in master files.

---

## MF (TYPE 4) — §3.3.5 (Obsolete)

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   MADNAME                     /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MADNAME: `<domain-name>`. Obsolete — converted to MX with
  preference 10 in master files.

---

## CNAME (TYPE 5) — §3.3.1

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     CNAME                     /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- CNAME: `<domain-name>` giving the canonical name for the owner.
  Compression allowed.

---

## SOA (TYPE 6) — §3.3.13

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     MNAME                     /     <domain-name>
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     RNAME                     /     <domain-name>
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    SERIAL                     |     uint32
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    REFRESH                    |     int32 (seconds)
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     RETRY                     |     int32 (seconds)
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    EXPIRE                     |     int32 (seconds)
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    MINIMUM                    |     uint32 (seconds)
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MNAME: `<domain-name>` of the original/primary name server for the
  zone. Compression allowed.
- RNAME: `<domain-name>` of the responsible mailbox. Compression
  allowed.
- SERIAL: unsigned 32-bit version number; uses sequence-space
  arithmetic (RFC 1982).
- REFRESH: 32-bit time interval (seconds) before the zone should be
  refreshed by a secondary.
- RETRY: 32-bit time interval (seconds) between failed-refresh retries.
- EXPIRE: 32-bit upper limit (seconds) on how long the zone may be
  considered authoritative without a successful refresh.
- MINIMUM: unsigned 32-bit minimum TTL (seconds) exported from this
  zone.

The RFC describes REFRESH/RETRY/EXPIRE as "32 bit time interval"
without explicit signedness. SERIAL and MINIMUM are explicitly
unsigned 32-bit. Encoders should treat all five as 32-bit fields
written via big-endian uint32; the implementation may surface them as
`int32` or `uint32` per its API choice.

---

## MB (TYPE 7) — §3.3.3 (EXPERIMENTAL)

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   MADNAME                     /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MADNAME: `<domain-name>` of a host that has the specified mailbox.

---

## MG (TYPE 8) — §3.3.6 (EXPERIMENTAL)

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   MGMNAME                     /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- MGMNAME: `<domain-name>` of a mailbox that is a member of the mail
  group named by the owner.

---

## MR (TYPE 9) — §3.3.8 (EXPERIMENTAL)

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   NEWNAME                     /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- NEWNAME: `<domain-name>` giving the proper rename of the specified
  mailbox.

---

## NULL (TYPE 10) — §3.3.10 (EXPERIMENTAL)

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                  <anything>                   /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- Up to 65535 octets of arbitrary bytes.
- Not allowed in master files (irrelevant for wire-only
  encoder/decoder).

---

## WKS (TYPE 11) — §3.4.2

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ADDRESS                    |   uint32 IPv4
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|       PROTOCOL        |                       |   uint8, then
+--+--+--+--+--+--+--+--+                       |   bitmap starts
|                                               |
/                   <BIT MAP>                   /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- ADDRESS: 32-bit IPv4 address (4 octets).
- PROTOCOL: 8-bit IP protocol number (e.g. TCP=6, UDP=17).
- BIT MAP: variable-length bitmap whose length is `RDLENGTH - 5`. The
  bitmap MUST be a multiple of 8 bits long. Bit 0 (MSB of byte 0)
  corresponds to port 0, bit 1 to port 1, etc. Trailing zero bits MAY
  be omitted (a missing bit is implicitly zero).

---

## PTR (TYPE 12) — §3.3.12

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   PTRDNAME                    /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- PTRDNAME: `<domain-name>` pointing to some location in the domain
  name space. Compression allowed.

---

## HINFO (TYPE 13) — §3.3.2

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                      CPU                      /     <character-string>
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                       OS                      /     <character-string>
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- CPU: `<character-string>` for CPU type.
- OS: `<character-string>` for operating system.

Note: `<character-string>` does NOT use compression; it is just a
length-prefixed byte string.

---

## MINFO (TYPE 14) — §3.3.7 (EXPERIMENTAL)

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                    RMAILBX                    /     <domain-name>
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                    EMAILBX                    /     <domain-name>
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- RMAILBX: `<domain-name>` of a mailbox responsible for the mailing
  list/mailbox. Root means "self".
- EMAILBX: `<domain-name>` of a mailbox to receive related error
  messages. Root means "send errors back to the sender".

---

## MX (TYPE 15) — §3.3.9

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                  PREFERENCE                   |     uint16
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   EXCHANGE                    /     <domain-name>
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- PREFERENCE: 16-bit unsigned integer; lower values are preferred.
- EXCHANGE: `<domain-name>` of a mail-exchange host. Compression
  allowed.

---

## TXT (TYPE 16) — §3.3.14

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   TXT-DATA                    /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

- TXT-DATA: one or more `<character-string>`s concatenated. The total
  length is bounded by RDLENGTH; the decoder reads character-strings
  until it has consumed exactly RDLENGTH octets.

---

## Unknown / future TYPEs

For any TYPE not listed above, the decoder MUST treat RDATA as RDLENGTH
opaque octets and MUST NOT attempt domain-name decompression inside
those bytes. An encoder for an unknown TYPE MUST emit raw bytes only,
with no compression pointers (§4.1.4).
