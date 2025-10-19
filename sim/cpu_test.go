package sim

import "testing"

// --- Minimal encoders (test-only) ---

func encU(op, rd, imm uint32) uint32 {
	return (imm & 0xFFFFF000) | (rd << 7) | op
}
func encI(op, rd, f3, rs1 uint32, imm int32, shiftFunct7 ...uint32) uint32 {
	var top uint32
	if len(shiftFunct7) > 0 {
		top = ((shiftFunct7[0] & 0x7F) << 25) | ((uint32(imm) & 0x1F) << 20)
	} else {
		top = (uint32(imm) & 0xFFF) << 20
	}
	return top | (rs1 << 15) | (f3 << 12) | (rd << 7) | op
}
func encR(rd, f3, rs1, rs2, f7 uint32) uint32 {
	const op = 0x33
	return (f7 << 25) | (rs2 << 20) | (rs1 << 15) | (f3 << 12) | (rd << 7) | op
}
func encS(f3, rs1, rs2 uint32, imm int32) uint32 {
	const op = 0x23
	u := uint32(imm) & 0xFFF
	hi := (u >> 5) & 0x7F
	lo := u & 0x1F
	return (hi << 25) | (rs2 << 20) | (rs1 << 15) | (f3 << 12) | (lo << 7) | op
}
func encB(f3, rs1, rs2 uint32, imm int32) uint32 {
	const op = 0x63
	i := uint32(imm) & 0x1FFF
	return (((i >> 12) & 0x1) << 31) |
		(((i >> 5) & 0x3F) << 25) |
		(rs2 << 20) |
		(rs1 << 15) |
		(f3 << 12) |
		(((i >> 1) & 0xF) << 8) |
		(((i >> 11) & 0x1) << 7) |
		op
}
func encJ(rd uint32, imm int32) uint32 {
	const op = 0x6F
	i := uint32(imm) & 0x1FFFFF
	return (((i >> 20) & 0x1) << 31) |
		(((i >> 1) & 0x3FF) << 21) |
		(((i >> 11) & 0x1) << 20) |
		(((i >> 12) & 0xFF) << 12) |
		(rd << 7) | op
}

// ABI indices (for readability)
const (
	x0 = 0
	a0 = 10
	a1 = 11
	t0 = 5
)

// Helper to push a 32-bit instruction into RAM at an aligned address.
func writeInst(t *testing.T, ram *RAM, addr uint32, inst uint32) {
	t.Helper()
	for i := 0; i < 4; i++ {
		if !ram.Write8(addr+uint32(i), uint8(inst>>(8*i))) {
			t.Fatalf("writeInst OOB at %d", addr+uint32(i))
		}
	}
}

func runToHalt(cpu *CPU, maxSteps int) bool {
	for i := 0; i < maxSteps; i++ {
		if !cpu.Step() {
			return true
		}
	}
	return false
}

func TestCPU_UART_Hello(t *testing.T) {
	ram := NewRAM(64 * 1024)
	uart := NewUART(nil)
	bus := NewBus(ram, uart)
	cpu := NewCPU(bus)

	// Place data string at 0x200.
	msg := []byte("Hi!\n\x00")
	if err := ram.WriteBytes(0x200, msg); err != nil {
		t.Fatalf("WriteBytes msg: %v", err)
	}

	// Program at 0x0000:
	//   lui   a0, UARTBase>>12
	//   addi  a0, a0, (UARTBase & 0xFFF)
	//   addi  a1, zero, 0x200
	// loop:
	//   lbu   t0, 0(a1)
	//   beq   t0, zero, done
	//   sb    t0, 0(a0)
	//   addi  a1, a1, 1
	//   jal   zero, loop
	// done:
	//   ecall
	pc := uint32(0)
	lo := int32(UARTBase & 0xFFF)
	hi := UARTBase & 0xFFFFF000

	insts := []uint32{
		encU(OpLUI, a0, hi),
		encI(opJALR, a0, 0, a0, lo), // reuse jalr encoding trick to add low 12 (or use OP-IMM addi); but use proper ADDI:
	}
	// Replace above with proper ADDI:
	insts[1] = encI(OpOPIMM, a0, F3ADDI, a0, lo)

	insts = append(insts,
		encI(OpOPIMM, a1, F3ADDI, x0, 0x200),
		encI(OpLOAD, t0, F3LBU, a1, 0),
		encB(F3BEQ, t0, x0, 16), // skip 4 insts ahead (16 bytes) to done
		encS(F3SB, a0, t0, 0),
		encI(OpOPIMM, a1, F3ADDI, a1, 1),
		encJ(0, -16), // back to lbu (minus 4 insts)
		0x00000073,   // ecall
	)

	for i, ins := range insts {
		writeInst(t, ram, pc+uint32(i*4), ins)
	}
	cpu.PC = pc

	if !runToHalt(cpu, 1000) {
		t.Fatalf("program did not halt in time")
	}
	if got, want := uart.String(), "Hi!\n"; got != want {
		t.Fatalf("UART output = %q, want %q", got, want)
	}
}

func TestCPU_Arithmetic_StoreToRAM(t *testing.T) {
	ram := NewRAM(4096)
	uart := NewUART(nil)
	bus := NewBus(ram, uart)
	cpu := NewCPU(bus)

	// x1=5; x2=7; x3=x1+x2; store x3 at [0x300]; ecall
	pc := uint32(0)
	insts := []uint32{
		encI(OpOPIMM, 1, F3ADDI, x0, 5),
		encI(OpOPIMM, 2, F3ADDI, x0, 7),
		encR(3, f3ADD_SUB, 1, 2, 0x00),      // add x3, x1, x2
		encI(OpOPIMM, 4, F3ADDI, x0, 0x300), // x4 = 0x300
		encS(f3SW, 4, 3, 0),                 // sw x3, 0(x4)
		0x00000073,                          // ecall
	}
	for i, ins := range insts {
		writeInst(t, ram, pc+uint32(i*4), ins)
	}
	cpu.PC = pc

	if !runToHalt(cpu, 100) {
		t.Fatalf("program did not halt in time")
	}
	// Check RAM[0x300..0x303] == 12 (little-endian)
	var v uint32
	for i := 0; i < 4; i++ {
		b, ok := ram.Read8(0x300 + uint32(i))
		if !ok {
			t.Fatalf("read back failed")
		}
		v |= uint32(b) << (8 * i)
	}
	if v != 12 {
		t.Fatalf("stored value = %d, want 12", v)
	}
}
