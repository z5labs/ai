package tlv

// Type identifies the kind of value carried by a Record.
//
// The concrete numeric width (e.g. uint8, uint16, uint32) is up to the
// caller; this package treats Type as an opaque identifier.
type Type uint16

// Record is a single Type-Length-Value entry.
//
// Length is not stored explicitly — it is always len(Value) when encoding.
// Decoders populate Value with a freshly allocated copy of the wire bytes.
type Record struct {
	Type  Type
	Value []byte
}

// Len returns the length of the record's value payload in bytes.
func (r Record) Len() int {
	return len(r.Value)
}
