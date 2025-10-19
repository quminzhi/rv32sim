package main

import (
	"fmt"
	"os"

	"rv32sim/sim"
)

// Minimal encoders (duplicated from tests for clarity in the demo)
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

func main() {
	ram := sim.NewRAM(64 * 1024)
	uart := sim.NewUART(os.Stdout)
	uart.LineBuffered = true
	bus := sim.NewBus(ram, uart)
	cpu := sim.NewCPU(bus)
	cpu.Trace = true // show disassembly

	// Place message bytes at 0x200
	msg := []byte("Hello, RV32!\n\x00")
	if err := ram.WriteBytes(0x200, msg); err != nil {
		fmt.Println("write data:", err)
		return
	}

	// Build program at 0x0
	const (
		x0 = 0
		a0 = 10
		a1 = 11
		t0 = 5
	)
	pc := uint32(0)
	lo := int32(sim.UARTBase & 0xFFF)
	hi := sim.UARTBase & 0xFFFFF000
	insts := []uint32{
		encU(sim.OpLUI, a0, hi),                      // (we'll export OpLUI below, or inline 0x37)
		encI(sim.OpOPIMM, a0, sim.F3ADDI, a0, lo),    // addi a0,a0,UART_lo
		encI(sim.OpOPIMM, a1, sim.F3ADDI, x0, 0x200), // a1=&msg
		encI(sim.OpLOAD, t0, sim.F3LBU, a1, 0),       // loop: lbu t0,0(a1)
		encB(sim.F3BEQ, t0, x0, 16),                  // beq t0,zero,done
		encS(sim.F3SB, a0, t0, 0),                    // sb t0,0(a0)
		encI(sim.OpOPIMM, a1, sim.F3ADDI, a1, 1),     // addi a1,a1,1
		encJ(0, -16),                                 // jal zero,loop
		0x00000073,                                   // ecall
	}
	// Note: If you kept Op*/F3* unexported, replace with literal constants as above.

	for i, ins := range insts {
		for j := 0; j < 4; j++ {
			ram.Write8(pc+uint32(i*4+j), uint8(ins>>(8*j)))
		}
	}
	cpu.PC = pc

	// Run to halt
	for steps := 0; steps < 1000; steps++ {
		if !cpu.Step() {
			break
		}
	}
}
