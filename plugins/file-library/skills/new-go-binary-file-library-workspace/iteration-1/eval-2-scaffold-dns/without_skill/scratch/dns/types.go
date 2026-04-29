package dns

// Message is the top-level AST node representing a single DNS wire-format
// message (header + question + answer/authority/additional sections).
//
// Concrete fields will be populated once the implement-binary-file-library
// agent fills in RFC 1035 types.
type Message struct {
	Header      Header
	Questions   []Question
	Answers     []ResourceRecord
	Authorities []ResourceRecord
	Additionals []ResourceRecord
}

// Header is the 12-byte DNS message header.
type Header struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

// Question represents a single entry in the question section of a DNS
// message.
type Question struct {
	Name  Name
	Type  uint16
	Class uint16
}

// ResourceRecord represents a single resource record in any of the
// answer/authority/additional sections of a DNS message.
type ResourceRecord struct {
	Name  Name
	Type  uint16
	Class uint16
	TTL   uint32
	RData []byte
}

// Name is a DNS domain name represented as a sequence of labels.
type Name struct {
	Labels []string
}
