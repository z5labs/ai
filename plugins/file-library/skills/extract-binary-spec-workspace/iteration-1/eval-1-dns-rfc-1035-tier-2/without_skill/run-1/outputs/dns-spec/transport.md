# Transport (RFC 1035 Section 4.2)

The DNS protocol allows messages to be sent over either UDP or TCP. The
on-the-wire framing differs between the two, but the message format itself is
identical.

## UDP usage (Section 4.2.1)

- Default port: **53** (both source/destination as appropriate).
- Each DNS message is carried in a single UDP datagram.
- Messages **MUST be 512 octets or fewer**, including the header. Longer
  messages are truncated and the **TC** bit is set in the response header.
- If a server's reply would exceed 512 octets, the server SHOULD return as
  much as fits within the limit, set the TC bit, and let the client retry
  over TCP.

## TCP usage (Section 4.2.2)

- Default port: **53**.
- Messages sent over a TCP connection are prefixed with a **two-octet
  big-endian length field**, giving the length of the message (excluding the
  two-octet length field itself):

  ```
      +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
      |                MESSAGE LENGTH                 |   <- 2 octets, big-endian
      +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
      /                    MESSAGE                    /   <- exactly that many octets
      /                                               /
      +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
  ```

- The 16-bit length field permits messages up to **65,535 octets** over TCP.
- Multiple messages may be exchanged on a single TCP connection, each with
  its own 2-octet length prefix.
- The connection MAY be closed after a response is sent, or MAY be kept open;
  receiving applications must handle either.

## Implications for an encoder/decoder

- The core encoder/decoder operates on a single self-contained DNS message
  byte slice. Wrapping/unwrapping the 2-octet TCP length prefix is the
  responsibility of the transport layer.
- For UDP, the encoder SHOULD verify that emitted messages do not exceed 512
  octets when the UDP transport is in use, or set the TC bit and truncate
  appropriately when emitting a response that is too large.
- The TC bit is a **header flag** that the upper-layer logic sets when it
  decides to truncate; the wire encoder simply emits whatever flags it is
  given.
