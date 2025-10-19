package sim

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadELF32_MinimalPTLOAD(t *testing.T) {
	// Build a tiny ELF32-LE RISC-V file in-memory with one PT_LOAD segment.
	eh := elf32Ehdr{}
	eh.Ident[0], eh.Ident[1], eh.Ident[2], eh.Ident[3] = 0x7F, 'E', 'L', 'F'
	eh.Ident[4] = eiCLASS32
	eh.Ident[5] = eiDATA2LSB
	eh.Ident[6] = 1 // version
	eh.Type = etEXEC
	eh.Machine = emRISCV
	eh.Version = 1
	eh.Entry = 0x200
	eh.Ehsize = uint16(binary.Size(eh))
	eh.Phentsize = uint16(binary.Size(elf32Phdr{}))
	eh.Phnum = 1
	eh.Phoff = uint32(eh.Ehsize)

	ph := elf32Phdr{
		Type:   ptLOAD,
		Offset: 0x100, // segment starts at file offset 0x100
		Vaddr:  0x200, // maps into RAM at 0x200
		Filesz: 4,
		Memsz:  8, // last 4 bytes are bss
		Align:  4,
	}

	// Build file bytes: [ELF hdr][phdr][padding .. 0x100][segment bytes]
	buf := new(bytes.Buffer)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(binary.Write(buf, binary.LittleEndian, &eh))
	must(binary.Write(buf, binary.LittleEndian, &ph))
	if pad := int(ph.Offset) - buf.Len(); pad > 0 {
		buf.Write(make([]byte, pad))
	}
	seg := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	buf.Write(seg)

	tmp := t.TempDir()
	path := filepath.Join(tmp, "mini.elf")
	must(os.WriteFile(path, buf.Bytes(), 0o600))

	ram := NewRAM(4096)
	entry, err := LoadELF32(path, ram)
	if err != nil {
		t.Fatalf("LoadELF32: %v", err)
	}
	if entry != 0x200 {
		t.Fatalf("entry=0x%x want 0x200", entry)
	}
	// Check the 4 file bytes
	got := make([]byte, 4)
	for i := 0; i < 4; i++ {
		b, ok := ram.Read8(0x200 + uint32(i))
		if !ok {
			t.Fatalf("ram read fail at 0x%x", 0x200+uint32(i))
		}
		got[i] = b
	}
	if !bytes.Equal(got, seg) {
		t.Fatalf("segment bytes = %v want %v", got, seg)
	}
	// Check the zero-filled tail (bss) at 0x204..0x207
	for i := 0; i < 4; i++ {
		b, ok := ram.Read8(0x204 + uint32(i))
		if !ok || b != 0 {
			t.Fatalf("bss tail not zero at 0x%x (b=%d ok=%v)", 0x204+uint32(i), b, ok)
		}
	}
}
