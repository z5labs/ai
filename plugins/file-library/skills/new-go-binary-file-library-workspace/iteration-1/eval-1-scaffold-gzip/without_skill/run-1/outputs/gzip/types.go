package gzip

// Magic number identifying a gzip stream (RFC 1952, section 2.3.1).
const (
	Magic1 byte = 0x1f
	Magic2 byte = 0x8b
)

// Compression method identifiers (RFC 1952, section 2.3.1).
const (
	CompressionDeflate byte = 8
)

// Flag bits in the FLG header byte (RFC 1952, section 2.3.1).
const (
	FlagText    byte = 1 << 0 // FTEXT
	FlagHCRC    byte = 1 << 1 // FHCRC
	FlagExtra   byte = 1 << 2 // FEXTRA
	FlagName    byte = 1 << 3 // FNAME
	FlagComment byte = 1 << 4 // FCOMMENT
)

// Operating system identifiers in the OS header byte (RFC 1952, section 2.3.1).
const (
	OSFAT          byte = 0
	OSAmiga        byte = 1
	OSVMS          byte = 2
	OSUnix         byte = 3
	OSVMCMS        byte = 4
	OSAtariTOS     byte = 5
	OSHPFS         byte = 6
	OSMacintosh    byte = 7
	OSZSystem      byte = 8
	OSCPM          byte = 9
	OSTOPS20       byte = 10
	OSNTFS         byte = 11
	OSQDOS         byte = 12
	OSAcornRiscOS  byte = 13
	OSUnknown      byte = 255
)

// ExtraField is an optional sub-field carried in the FEXTRA section.
type ExtraField struct {
	// SubfieldID is the two-byte identifier of this extra sub-field.
	SubfieldID [2]byte
	// Data is the raw bytes of the sub-field payload.
	Data []byte
}

// Header is the parsed gzip member header.
type Header struct {
	// CompressionMethod is the CM byte (typically CompressionDeflate).
	CompressionMethod byte
	// Flags is the FLG byte (combination of Flag* constants).
	Flags byte
	// ModTime is the MTIME field, seconds since Unix epoch (0 if unset).
	ModTime uint32
	// ExtraFlags is the XFL byte.
	ExtraFlags byte
	// OS is the OS byte identifying the source operating system.
	OS byte
	// Extra holds the FEXTRA sub-fields, if FlagExtra is set.
	Extra []ExtraField
	// Name holds the original file name, if FlagName is set.
	Name string
	// Comment holds the file comment, if FlagComment is set.
	Comment string
	// HeaderCRC16 holds the CRC-16 of the header bytes, if FlagHCRC is set.
	HeaderCRC16 uint16
}

// Trailer is the parsed gzip member trailer.
type Trailer struct {
	// CRC32 of the uncompressed data.
	CRC32 uint32
	// ISize is the size of the uncompressed input modulo 2^32.
	ISize uint32
}

// Member is a single gzip member (a complete header + compressed data + trailer).
//
// A gzip stream is one or more concatenated members.
type Member struct {
	Header     Header
	Compressed []byte
	Trailer    Trailer
}

// File is the top-level representation of a parsed gzip stream.
type File struct {
	// Members is the ordered list of gzip members in the stream.
	Members []Member
}
