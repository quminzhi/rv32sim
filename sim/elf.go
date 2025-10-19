package sim

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// Minimal constants for ELF32 little-endian, RISC-V
const (
	elfMAG0    = 0x7F
	elfMAG1    = 'E'
	elfMAG2    = 'L'
	elfMAG3    = 'F'
	eiCLASS32  = 1 // 32-bit
	eiDATA2LSB = 1 // little-endian
	etEXEC     = 2 // executable
	emRISCV    = 243
	ptLOAD     = 1
)

type elf32Ehdr struct {
	Ident     [16]byte
	Type      uint16
	Machine   uint16
	Version   uint32
	Entry     uint32
	Phoff     uint32
	Shoff     uint32
	Flags     uint32
	Ehsize    uint16
	Phentsize uint16
	Phnum     uint16
	Shentsize uint16
	Shnum     uint16
	Shstrndx  uint16
}

type elf32Phdr struct {
	Type   uint32
	Offset uint32
	Vaddr  uint32
	Paddr  uint32
	Filesz uint32
	Memsz  uint32
	Flags  uint32
	Align  uint32
}

// LoadELF32 loads a minimal RV32 little-endian ELF into RAM and returns the entry PC.
// It copies all PT_LOAD segments to RAM at vaddr, and zero-fills any tail (bss).
func LoadELF32(path string, ram *RAM) (entry uint32, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	r := bytes.NewReader(raw)

	var hdr elf32Ehdr
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return 0, fmt.Errorf("read ELF header: %w", err)
	}
	if hdr.Ident[0] != elfMAG0 || hdr.Ident[1] != elfMAG1 || hdr.Ident[2] != elfMAG2 || hdr.Ident[3] != elfMAG3 {
		return 0, errors.New("not an ELF file")
	}
	if hdr.Ident[4] != eiCLASS32 {
		return 0, errors.New("not ELF32")
	}
	if hdr.Ident[5] != eiDATA2LSB {
		return 0, errors.New("not little-endian ELF")
	}
	if hdr.Machine != emRISCV {
		return 0, fmt.Errorf("unexpected machine %d (need RISC-V)", hdr.Machine)
	}
	if hdr.Ehsize != uint16(binary.Size(hdr)) {
		return 0, fmt.Errorf("unexpected ehsize %d", hdr.Ehsize)
	}
	if hdr.Phentsize != uint16(binary.Size(elf32Phdr{})) {
		return 0, fmt.Errorf("unexpected phentsize %d", hdr.Phentsize)
	}

	// Iterate program headers.
	for i := 0; i < int(hdr.Phnum); i++ {
		off := int64(hdr.Phoff) + int64(i)*int64(hdr.Phentsize)
		if _, err := r.Seek(off, 0); err != nil {
			return 0, fmt.Errorf("seek phdr: %w", err)
		}
		var ph elf32Phdr
		if err := binary.Read(r, binary.LittleEndian, &ph); err != nil {
			return 0, fmt.Errorf("read phdr: %w", err)
		}
		if ph.Type != ptLOAD {
			continue
		}
		// Copy the file part, if any.
		if ph.Filesz > 0 {
			if int(ph.Offset+ph.Filesz) > len(raw) {
				return 0, fmt.Errorf("segment %d exceeds file size", i)
			}
			seg := raw[ph.Offset : ph.Offset+ph.Filesz]
			// Bounds check against RAM space
			if ph.Vaddr+ph.Filesz > ram.Size() {
				return 0, fmt.Errorf("segment %d out of RAM bounds (vaddr=0x%x, size=%d)", i, ph.Vaddr, ph.Filesz)
			}
			if err := ram.WriteBytes(ph.Vaddr, seg); err != nil {
				return 0, fmt.Errorf("write segment %d: %w", i, err)
			}
		}
		// Zero-fill tail (bss region inside segment).
		if ph.Memsz > ph.Filesz {
			start := ph.Vaddr + ph.Filesz
			n := ph.Memsz - ph.Filesz
			if start+n > ram.Size() {
				return 0, fmt.Errorf("bss tail for segment %d exceeds RAM", i)
			}
			zero := make([]byte, n)
			if err := ram.WriteBytes(start, zero); err != nil {
				return 0, fmt.Errorf("zero tail: %w", err)
			}
		}
	}

	return hdr.Entry, nil
}
