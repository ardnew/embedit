package embedit

import (
	"errors"
	"strings"

	"github.com/ardnew/embedit/history"
)

// CLI is the primary, package-exported instance of CommandLine.
var CLI CommandLine

// Error values returned by methods of type CommandLine.
var (
	ErrReceiver      = errors.New("nil receiver")
	ErrWriteCallback = errors.New("invalid write callback")
	ErrReadCallback  = errors.New("invalid read callback")
)

// Config defines the configuration parameters of a CommandLine.
type Config struct {
	WriteByte func(byte) error
	ReadByte  func() (byte, error)
}

// ControlSequence represents a Control Sequence Introducer (CSI) ANSI sequence.
type ControlSequence struct {
	Byte [16]byte
}

// Set writes the given parameter, intermediate, and final byte code to the
// receiver buffer, following the required 'ESC[' prefix.
//
//	| For Control Sequence Introducer, or CSI, commands, the ESC [ is followed
//	| by any number (including none) of "parameter bytes" in the range 0x30–0x3F
//	| (ASCII 0–9:;<=>?), then by any number of "intermediate bytes" in the range
//	| 0x20–0x2F (ASCII space and !"#$%&'()*+,-./), then finally by a single
//	| "final byte" in the range 0x40–0x7E (ASCII @A–Z[\]^_`a–z{|}~).
//
// Source: https://en.wikipedia.org/wiki/ANSI_escape_code#CSI_(Control_Sequence_Introducer)_sequences
func (cs *ControlSequence) Set(param []byte, inter []byte, code byte) (n int) {
	if cs == nil {
		return 0
	}
	const maxLen = len(cs.Byte) - 1
	cs.Byte[0] = 0x1B // ESC
	cs.Byte[1] = '['
	c := 2
	// Always zero out the remaining bytes starting after the last valid byte.
	defer func() {
		n = c
		for ; c <= maxLen; c++ {
			cs.Byte[c] = 0
		}
	}()
	for i := 0; i < len(param) && c < maxLen; i++ {
		// Special case: convert single digit parameter bytes to ASCII encoding
		if param[i] < 10 {
			param[i] += '0'
		}
		if param[i] < 0x30 || 0x3F < param[i] {
			return
		}
		cs.Byte[c] = param[i]
		c++
	}
	for i := 0; i < len(inter) && c < maxLen; i++ {
		if inter[i] < 0x20 || 0x2F < inter[i] {
			return
		}
		cs.Byte[c] = inter[i]
		c++
	}
	if c < maxLen && (0x40 <= code && code <= 0x7E) {
		cs.Byte[c] = code
		c++
	}
	return
}

func (cs *ControlSequence) String() string {
	if cs == nil {
		return "<nil>"
	}
	if cs.Byte[0] != 0x1B || cs.Byte[1] != '[' {
		return "<undef>" // Invalid sequence
	}
	var subseq byte
	var sb strings.Builder
	sb.WriteString(`ESC[`)
RANGE:
	for _, c := range cs.Byte[2:] {
		switch {
		case 0x30 <= c && c <= 0x3F: // Parameter bytes
			switch subseq {
			case 0:
				sb.WriteRune(' ')
				subseq++
				fallthrough
			case 1:
				sb.WriteRune(rune(c))
			default:
				break RANGE
			}

		case 0x20 <= c && c <= 0x2F: // Intermediate bytes
			if subseq < 2 { // Previous subsequence was prefix or parameter bytes
				sb.WriteRune(' ')
				subseq = 2
			}
			sb.WriteRune(rune(c))

		case 0x40 <= c && c <= 0x7E: // Code byte
			// There can (and must) only be 1 code byte,
			//   and it is always the final byte of the sequence.
			sb.WriteRune(' ')
			sb.WriteRune(rune(c))
			return sb.String()
		}
	}
	return "<invalid>"
}

// CommandLine contains the state and configuration of a line-buffered terminal
// interface.
type CommandLine struct {
	config Config
	history.History
	csi ControlSequence
}

// Configure initializes the receiver's state and configuration.
func (cl *CommandLine) Configure(c Config) {
	cl.config = c
	cl.History.Init()
}

// CSI returns a slice of bytes representing a Control Sequence Introducer (CSI)
// ANSI sequence.
// It overwrites and slices the receiver's byte array instead of allocating one.
func (cl *CommandLine) CSI(code byte, param []byte) []byte {
	n := cl.csi.Set(param, []byte{}, code)
	return cl.csi.Byte[:n]
}

// WriteByte writes the given byte to the receiver's output callback.
func (cl *CommandLine) WriteByte(b byte) (err error) {
	if cl == nil {
		return ErrReceiver
	}
	if cl.config.WriteByte == nil {
		return ErrWriteCallback
	}
	return cl.config.WriteByte(b)
}

func (cl *CommandLine) Write(p []byte) (n int, err error) {
	if cl == nil {
		return 0, ErrReceiver
	}
	for _, b := range p {
		if err = cl.WriteByte(b); err != nil {
			return
		}
		n++
	}
	return
}

// ReadByte writes the given byte to the receiver's output callback.
func (cl *CommandLine) ReadByte() (b byte, err error) {
	if cl == nil {
		return 0, ErrReceiver
	}
	if cl.config.ReadByte == nil {
		return 0, ErrReadCallback
	}
	return cl.config.ReadByte()
}

func (cl *CommandLine) Read(p []byte) (n int, err error) {
	if cl == nil {
		return 0, ErrReceiver
	}
	for n = 0; n < len(p); n++ {
		var b byte
		b, err = cl.config.ReadByte()
		if err != nil {
			return
		}
		p[n] = b
	}
	return
}

func (cl *CommandLine) String() string {
	if cl == nil {
		return "<nil>"
	}
	var sb strings.Builder
	sb.WriteRune('{')
	sb.WriteString("History:")
	sb.WriteString(cl.History.String())
	sb.WriteRune('}')
	return sb.String()
}
