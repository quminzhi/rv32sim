package sim

import (
	"fmt"
)

// CPU: minimal RV32I subset with LB/LBU/LW and SB/SW.
// Any ECALL halts (returns false from Step()).
//
// Tip for teaching: set Trace=true to see human-readable instructions
// via the Disasm() helper below.
type CPU struct {
	Reg   [32]uint32
	PC    uint32
	Bus   *Bus
	Trace bool
}

func NewCPU(bus *Bus) *CPU { return &CPU{Bus: bus} }

func (c *CPU) readReg(i uint32) uint32 {
	if i == 0 {
		return 0
	}
	return c.Reg[i]
}

func (c *CPU) writeReg(i, v uint32) {
	if i != 0 {
		c.Reg[i] = v
	}
}

func (c *CPU) fetch() (uint32, bool) { return c.Bus.Read32(c.PC) }

func (c *CPU) trap(msg string) bool {
	fmt.Printf("\n[trap] %s\n", msg)
	return false
}

func (c *CPU) trace(inst uint32) {
	if c.Trace {
		fmt.Printf("%08x: %08x  %s\n", c.PC, inst, Disasm(c.PC, inst))
	}
}

func addPC(pc uint32, off int32) uint32 { return uint32(int32(pc) + off) }

func (c *CPU) Step() bool {
	inst, ok := c.fetch()
	if !ok {
		return c.trap("fetch OOB or unaligned")
	}

	op := inst & 0x7F
	rd := (inst >> 7) & 0x1F
	f3 := (inst >> 12) & 0x7
	rs1 := (inst >> 15) & 0x1F
	rs2 := (inst >> 20) & 0x1F
	f7 := (inst >> 25) & 0x7F

	nextPC := c.PC + 4
	c.trace(inst)

	switch op {

	case OpLUI:
		c.writeReg(rd, uint32(immU(inst)))

	case opAUIPC:
		c.writeReg(rd, addPC(c.PC, immU(inst)))

	case opJAL:
		c.writeReg(rd, c.PC+4)
		nextPC = addPC(c.PC, immJ(inst))

	case opJALR:
		tgt := (c.readReg(rs1) + uint32(immI(inst))) &^ 1
		c.writeReg(rd, c.PC+4)
		nextPC = tgt

	case opBRANCH:
		a := c.readReg(rs1)
		b := c.readReg(rs2)
		off := uint32(immB(inst))
		switch f3 {
		case F3BEQ:
			if a == b {
				nextPC = c.PC + off
			}
		case f3BNE:
			if a != b {
				nextPC = c.PC + off
			}
		case f3BLT:
			if int32(a) < int32(b) {
				nextPC = c.PC + off
			}
		case f3BGE:
			if int32(a) >= int32(b) {
				nextPC = c.PC + off
			}
		case f3BLTU:
			if a < b {
				nextPC = c.PC + off
			}
		case f3BGEU:
			if a >= b {
				nextPC = c.PC + off
			}
		default:
			fmt.Printf("[warn] BRANCH f3=%d\n", f3)
		}

	case OpLOAD:
		base := c.readReg(rs1)
		addr := base + uint32(immI(inst))
		switch f3 {
		case f3LB:
			b, ok := c.Bus.Read8(addr)
			if !ok {
				return c.trap("LB OOB")
			}
			c.writeReg(rd, uint32(int32(int8(b))))
		case F3LBU:
			b, ok := c.Bus.Read8(addr)
			if !ok {
				return c.trap("LBU OOB")
			}
			c.writeReg(rd, uint32(b))
		case f3LW:
			w, ok := c.Bus.Read32(addr)
			if !ok {
				return c.trap("LW OOB or unaligned")
			}
			c.writeReg(rd, w)
		default:
			fmt.Printf("[warn] LOAD f3=%d\n", f3)
		}

	case opSTORE:
		base := c.readReg(rs1)
		addr := base + uint32(immS(inst))
		switch f3 {
		case F3SB:
			v := uint8(c.readReg(rs2))
			if !c.Bus.Write8(addr, v) {
				return c.trap("SB OOB")
			}
		case f3SW:
			v := c.readReg(rs2)
			if !c.Bus.Write32(addr, v) {
				return c.trap("SW OOB or unaligned")
			}
		default:
			fmt.Printf("[warn] STORE f3=%d\n", f3)
		}

	case OpOPIMM:
		a := c.readReg(rs1)
		imm := uint32(immI(inst))
		switch f3 {
		case F3ADDI:
			c.writeReg(rd, a+uint32(int32(imm)))
		case f3XORI:
			c.writeReg(rd, a^imm)
		case f3ORI:
			c.writeReg(rd, a|imm)
		case f3ANDI:
			c.writeReg(rd, a&imm)
		case f3SLLI:
			sh := (imm & 0x1F)
			c.writeReg(rd, a<<sh)
		case f3SRxI:
			sh := (imm & 0x1F)
			if f7 == funct7SUBSRA { // SRAI
				c.writeReg(rd, uint32(int32(a)>>sh))
			} else { // SRLI
				c.writeReg(rd, a>>sh)
			}
		default:
			fmt.Printf("[warn] OP-IMM f3=%d\n", f3)
		}

	case opOP:
		a := c.readReg(rs1)
		b := c.readReg(rs2)
		switch f3 {
		case f3ADD_SUB:
			if f7 == funct7SUBSRA { // SUB
				c.writeReg(rd, a-b)
			} else { // ADD
				c.writeReg(rd, a+b)
			}
		case f3XOR:
			c.writeReg(rd, a^b)
		case f3OR:
			c.writeReg(rd, a|b)
		case f3AND:
			c.writeReg(rd, a&b)
		case f3SLL:
			c.writeReg(rd, a<<(b&0x1F))
		case f3SRx:
			if f7 == funct7SUBSRA { // SRA
				c.writeReg(rd, uint32(int32(a)>>(b&0x1F)))
			} else { // SRL
				c.writeReg(rd, a>>(b&0x1F))
			}
		case f3SLT:
			if int32(a) < int32(b) {
				c.writeReg(rd, 1)
			} else {
				c.writeReg(rd, 0)
			}
		case f3SLTU:
			if a < b {
				c.writeReg(rd, 1)
			} else {
				c.writeReg(rd, 0)
			}
		default:
			fmt.Printf("[warn] OP f3=%d f7=0x%x\n", f3, f7)
		}

	case opSYSTEM:
		// ECALL: halt
		fmt.Println("\n[halt] ECALL")
		return false

	default:
		fmt.Printf("\n[warn] unsupported opcode 0x%x at pc=%08x\n", op, c.PC)
	}

	c.PC = nextPC
	c.Reg[0] = 0 // x0 is hardwired to zero
	return true
}
