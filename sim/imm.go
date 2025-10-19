package sim

// sext sign-extends the low 'bits' bits of v to int32.
func sext(v uint32, bits uint) int32 {
	shift := 32 - bits
	return int32(v<<shift) >> shift
}

func immI(inst uint32) int32 {
	return sext(inst>>20, 12)
}

func immU(inst uint32) int32 {
	// Already in bits 31..12.
	return int32(inst & 0xFFFFF000)
}

func immS(inst uint32) int32 {
	// imm[11:5]=inst[31:25], imm[4:0]=inst[11:7]
	imm := ((inst >> 25) << 5) | ((inst >> 7) & 0x1F)
	return sext(uint32(imm), 12)
}

func immB(inst uint32) int32 {
	// imm[12|10:5|4:1|11] = inst[31|30:25|11:8|7], LSB=0
	u := ((inst>>31)&0x1)<<12 |
		((inst>>25)&0x3F)<<5 |
		((inst>>8)&0xF)<<1 |
		((inst>>7)&0x1)<<11
	return sext(uint32(u), 13)
}

func immJ(inst uint32) int32 {
	// imm[20|10:1|11|19:12] = inst[31|30:21|20|19:12], LSB=0
	u := ((inst>>31)&0x1)<<20 |
		((inst>>21)&0x3FF)<<1 |
		((inst>>20)&0x1)<<11 |
		((inst>>12)&0xFF)<<12
	return sext(uint32(u), 21)
}
