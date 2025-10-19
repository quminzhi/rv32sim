package sim

import (
	"bytes"
	"io"
)

const (
	// UART occupies a small 256-byte region; we only use DATA at offset 0.
	UARTSize = 0x100
)

// UART is a trivial TX-only UART.
//   - Write8(offset==0) transmits a byte: it is appended to Buf and optionally
//     mirrored to Out (e.g., os.Stdout).
//   - Read8 returns 0 for all offsets (no RX in this teaching model).
type UART struct {
	Out          io.Writer // optional, can be nil
	buf          bytes.Buffer
	lineBuf      bytes.Buffer
	LineBuffered bool
}

func NewUART(w io.Writer) *UART { return &UART{Out: w} }

func (u *UART) Read8(off uint32) (uint8, bool) {
	// No RX implemented; reading DATA returns 0.
	if off < UARTSize {
		return 0, true
	}
	return 0, false
}

func (u *UART) Write8(off uint32, v uint8) bool {
	if off >= UARTSize {
		return false
	}
	_ = u.buf.WriteByte(v)

	if u.Out != nil {
		if u.LineBuffered {
			_ = u.lineBuf.WriteByte(v)
			if v == '\n' {
				_, _ = u.Out.Write(u.lineBuf.Bytes())
				u.lineBuf.Reset()
			}
		} else {
			_, _ = u.Out.Write([]byte{v})
		}
	}
	return true
}

func (u *UART) String() string { return u.buf.String() }
func (u *UART) Reset()         { u.buf.Reset() }
