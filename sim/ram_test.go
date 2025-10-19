package sim

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRAM_ReadWrite8_AndWriteBytes(t *testing.T) {
	ram := NewRAM(16)
	// Zero-init
	if b, ok := ram.Read8(0); !ok || b != 0 {
		t.Fatalf("initial Read8 = (%d,%v), want (0,true)", b, ok)
	}
	// In-bounds write+read
	if ok := ram.Write8(5, 0xAB); !ok {
		t.Fatalf("Write8 in-bounds failed")
	}
	if b, ok := ram.Read8(5); !ok || b != 0xAB {
		t.Fatalf("Read8 got %x want %x", b, 0xAB)
	}
	// OOB
	if _, ok := ram.Read8(16); ok {
		t.Fatalf("Read8 out-of-bounds should fail")
	}
	if ok := ram.Write8(16, 1); ok {
		t.Fatalf("Write8 out-of-bounds should fail")
	}
	// WriteBytes exact end
	if err := ram.WriteBytes(14, []byte{1, 2}); err != nil {
		t.Fatalf("WriteBytes exact-end: %v", err)
	}
	// OOB
	if err := ram.WriteBytes(15, []byte{1, 2}); err == nil {
		t.Fatalf("WriteBytes OOB should error")
	}
}

func TestRAM_LoadFlat(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "img.bin")
	img := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	if err := os.WriteFile(path, img, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	ram := NewRAM(8)
	if err := ram.LoadFlat(path, 2); err != nil {
		t.Fatalf("LoadFlat: %v", err)
	}
	got := make([]byte, 4)
	for i := 0; i < 4; i++ {
		b, ok := ram.Read8(uint32(2 + i))
		if !ok {
			t.Fatalf("Read8 failed at %d", i)
		}
		got[i] = b
	}
	if !bytes.Equal(got, img) {
		t.Fatalf("got %v want %v", got, img)
	}
}
