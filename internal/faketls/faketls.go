package faketls

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// FakeTLS mimics real TLS traffic patterns
type FakeTLS struct {
	// Version is the TLS version to mimic
	Version uint16
	// CipherSuites to advertise
	CipherSuites []uint16
}

// NewFakeTLS creates a new FakeTLS handler
func NewFakeTLS() *FakeTLS {
	return &FakeTLS{
		Version:      0x0303, // TLS 1.2
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303}, // AES-GCM
	}
}

// ClientHello generates a realistic TLS ClientHello
func (f *FakeTLS) ClientHello() ([]byte, error) {
	var buf bytes.Buffer

	// Record layer: Handshake (22), TLS 1.2
	buf.WriteByte(0x16)
	buf.WriteByte(0x03)
	buf.WriteByte(0x03)

	// Handshake: ClientHello (1)
	buf.WriteByte(0x01)

	// Handshake length (placeholder)
	helloStart := buf.Len()
	buf.Write([]byte{0x00, 0x00, 0x00})

	// Client version
	binary.Write(&buf, binary.BigEndian, f.Version)

	// Random (32 bytes)
	random := make([]byte, 32)
	rand.Read(random)
	buf.Write(random)

	// Session ID (empty)
	buf.WriteByte(0x00)

	// Cipher suites
	binary.Write(&buf, binary.BigEndian, uint16(len(f.CipherSuites)*2))
	for _, cs := range f.CipherSuites {
		binary.Write(&buf, binary.BigEndian, cs)
	}

	// Compression methods (null)
	buf.WriteByte(0x01)
	buf.WriteByte(0x00)

	// Extensions
	extLenPos := buf.Len()
	buf.Write([]byte{0x00, 0x00})

	// SNI extension
	buf.Write([]byte{0x00, 0x00}) // type: server_name
	sniTypeLenPos := buf.Len()    // position of length field for SNI data
	buf.Write([]byte{0x00, 0x00}) // length placeholder (2 bytes after this)
	buf.WriteByte(0x00)           // name_type: host_name
	buf.Write([]byte{0x00, 0x00}) // name length placeholder (2 bytes)
	nameStrPos := buf.Len()       // position where "example.com" starts
	buf.WriteString("example.com") // SNI name

	// Calculate lengths
	nameLen := len("example.com") // 11 bytes
	sniDataLen := 1 + 2 + nameLen  // name_type(1) + name_len(2) + name_str

	// ALPN extension (for HTTP/2)
	buf.Write([]byte{0x00, 0x10}) // type: alpn
	buf.Write([]byte{0x00, 0x00}) // length placeholder
	buf.Write([]byte{0x00, 0x02}) // protocols length
	buf.WriteString("h2")
	alpnLen := 5 // 2 (type) + 2 (len) + 1 (proto len) + 2 ("h2")

	// Fill lengths
	binary.BigEndian.PutUint16(buf.Bytes()[sniTypeLenPos:], uint16(sniDataLen))
	binary.BigEndian.PutUint16(buf.Bytes()[nameStrPos-2:], uint16(nameLen))
	binary.BigEndian.PutUint16(buf.Bytes()[extLenPos:], uint16(sniDataLen+alpnLen))

	// Fill ALPN length
	binary.BigEndian.PutUint16(buf.Bytes()[buf.Len()-4:], uint16(3)) // 1 + 2 for "h2"

	// ALPN extension (for HTTP/2)
	buf.Write([]byte{0x00, 0x10}) // type: alpn
	buf.Write([]byte{0x00, 0x00}) // length
	buf.Write([]byte{0x00, 0x02}) // protocols length
	buf.WriteString("h2")

	// Fill handshake length
	helloLen := buf.Len() - helloStart - 4
	binary.BigEndian.PutUint32(buf.Bytes()[helloStart:], uint32(helloLen))

	// Fill record length
	recordLen := buf.Len() - 5
	binary.BigEndian.PutUint16(buf.Bytes()[3:], uint16(recordLen))

	return buf.Bytes(), nil
}

// ServerHello generates a TLS ServerHello response
func (f *FakeTLS) ServerHello() ([]byte, error) {
	var buf bytes.Buffer

	// Record layer header
	buf.Write([]byte{0x16, 0x03, 0x03})

	// Handshake
	buf.WriteByte(0x02)

	// Handshake length
	buf.Write([]byte{0x00, 0x00, 0x2c})

	// Server version
	binary.Write(&buf, binary.BigEndian, f.Version)

	// Random
	random := make([]byte, 32)
	rand.Read(random)
	buf.Write(random)

	// Session ID
	buf.WriteByte(0x00)

	// Cipher suite
	binary.Write(&buf, binary.BigEndian, f.CipherSuites[0])

	// Compression
	buf.WriteByte(0x00)

	// Extensions
	buf.Write([]byte{0x00, 0x0c}) // length

	// ALPN
	buf.Write([]byte{0x00, 0x10})
	buf.Write([]byte{0x00, 0x00, 0x02})
	buf.WriteString("h2")

	// Record length
	recordLen := buf.Len() - 5
	binary.BigEndian.PutUint16(buf.Bytes()[3:], uint16(recordLen))

	return buf.Bytes(), nil
}

// TrafficPattern generates realistic traffic timing
type TrafficPattern struct {
	Intervals []time.Duration
	Sizes     []int
	idx       int
}

// NewTrafficPattern creates realistic traffic patterns
func NewTrafficPattern() *TrafficPattern {
	// Common patterns: small bursts, large downloads
	return &TrafficPattern{
		Intervals: []time.Duration{
			10 * time.Millisecond,
			50 * time.Millisecond,
			100 * time.Millisecond,
			200 * time.Millisecond,
		},
		Sizes: []int{
			64,    // ACK
			256,   // Small data
			1024,  // Medium
			1400,  // MTU
		},
	}
}

// Next returns the next interval and size
func (p *TrafficPattern) Next() (time.Duration, int) {
	interval := p.Intervals[p.idx%len(p.Intervals)]
	size := p.Sizes[p.idx%len(p.Sizes)]
	p.idx++
	return interval, size
}

// ObfuscatePacket obfuscates packet to match TLS patterns
func (f *FakeTLS) ObfuscatePacket(data []byte) ([]byte, error) {
	// Add random padding to match TLS record sizes
	if len(data) < 64 {
		padding := make([]byte, 64-len(data))
		rand.Read(padding)
		data = append(data, padding...)
	}

	// Encrypt with AES-GCM (simulated)
	key := sha256.Sum256(data)
	block, err := newAESCipher(key[:])
	if err != nil {
		return nil, err
	}

	iv := make([]byte, block.BlockSize())
	rand.Read(iv)

	encrypted := make([]byte, len(data))
	block.Encrypt(encrypted, data)

	return append(iv, encrypted...), nil
}

func newAESCipher(key []byte) (cipher.Block, error) {
	// Simplified - use crypto/aes in production
	return nil, fmt.Errorf("not implemented")
}

// ReadFullPacket reads a complete TLS record
func ReadFullPacket(r io.Reader) ([]byte, error) {
	// Read header
	header := make([]byte, 5)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	// Check record type
	if header[0] != 0x17 {
		return nil, fmt.Errorf("not application data")
	}

	// Get length
	length := binary.BigEndian.Uint16(header[3:])
	if length > 16384 {
		return nil, fmt.Errorf("record too large")
	}

	// Read data
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return append(header, data...), nil
}
