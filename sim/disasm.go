package sim

import "fmt"

var abiReg = [...]string{
	"zero", "ra", "sp", "gp", "tp", "t0", "t1", "t2",
	"s0", "s1", "a0", "a1", "a2", "a3", "a4", "a5",
	"a6", "a7", "s2", "s3", "s4", "s5", "s6", "s7",
	"s8", "s9", "s10", "s11", "t3", "t4", "t5", "t6",
}

func rn(i uint32) string {
	if i < uint32(len(abiReg)) {
		return abiReg[i]
	}
	return fmt.Sprintf("x%d", i)
}

func Disasm(pc, inst uint32) string {
	op := inst & 0x7F
	rd := (inst >> 7) & 0x1F
	f3 := (inst >> 12) & 0x7
	rs1 := (inst >> 15) & 0x1F
	rs2 := (inst >> 20) & 0x1F
	f7 := (inst >> 25) & 0x7F

	switch op {
	case OpLUI:
		return fmt.Sprintf("lui   %s, 0x%x", rn(rd), uint32(immU(inst)))
	case opAUIPC:
		return fmt.Sprintf("auipc %s, 0x%x", rn(rd), uint32(immU(inst)))
	case opJAL:
		tgt := uint32(int32(pc) + immJ(inst))
		return fmt.Sprintf("jal   %s, 0x%08x", rn(rd), tgt)
	case opJALR:
		return fmt.Sprintf("jalr  %s, %d(%s)", rn(rd), int32(immI(inst)), rn(rs1))
	case opBRANCH:
		tgt := uint32(int32(pc) + immB(inst))
		switch f3 {
		case F3BEQ:
			return fmt.Sprintf("beq   %s, %s, 0x%08x", rn(rs1), rn(rs2), tgt)
		case f3BNE:
			return fmt.Sprintf("bne   %s, %s, 0x%08x", rn(rs1), rn(rs2), tgt)
		case f3BLT:
			return fmt.Sprintf("blt   %s, %s, 0x%08x", rn(rs1), rn(rs2), tgt)
		case f3BGE:
			return fmt.Sprintf("bge   %s, %s, 0x%08x", rn(rs1), rn(rs2), tgt)
		case f3BLTU:
			return fmt.Sprintf("bltu  %s, %s, 0x%08x", rn(rs1), rn(rs2), tgt)
		case f3BGEU:
			return fmt.Sprintf("bgeu  %s, %s, 0x%08x", rn(rs1), rn(rs2), tgt)
		}
	case OpLOAD:
		off := int32(immI(inst))
		switch f3 {
		case f3LB:
			return fmt.Sprintf("lb    %s, %d(%s)", rn(rd), off, rn(rs1))
		case F3LBU:
			return fmt.Sprintf("lbu   %s, %d(%s)", rn(rd), off, rn(rs1))
		case f3LW:
			return fmt.Sprintf("lw    %s, %d(%s)", rn(rd), off, rn(rs1))
		}
	case opSTORE:
		off := int32(immS(inst))
		switch f3 {
		case F3SB:
			return fmt.Sprintf("sb    %s, %d(%s)", rn(rs2), off, rn(rs1))
		case f3SW:
			return fmt.Sprintf("sw    %s, %d(%s)", rn(rs2), off, rn(rs1))
		}
	case OpOPIMM:
		imm := int32(immI(inst))
		switch f3 {
		case F3ADDI:
			return fmt.Sprintf("addi  %s, %s, %d", rn(rd), rn(rs1), imm)
		case f3XORI:
			return fmt.Sprintf("xori  %s, %s, %d", rn(rd), rn(rs1), imm)
		case f3ORI:
			return fmt.Sprintf("ori   %s, %s, %d", rn(rd), rn(rs1), imm)
		case f3ANDI:
			return fmt.Sprintf("andi  %s, %s, %d", rn(rd), rn(rs1), imm)
		case f3SLLI:
			return fmt.Sprintf("slli  %s, %s, %d", rn(rd), rn(rs1), (uint32(imm) & 0x1F))
		case f3SRxI:
			if f7 == funct7SUBSRA {
				return fmt.Sprintf("srai  %s, %s, %d", rn(rd), rn(rs1), (uint32(imm) & 0x1F))
			}
			return fmt.Sprintf("srli  %s, %s, %d", rn(rd), rn(rs1), (uint32(imm) & 0x1F))
		}
	case opOP:
		switch f3 {
		case f3ADD_SUB:
			if f7 == funct7SUBSRA {
				return fmt.Sprintf("sub   %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
			}
			return fmt.Sprintf("add   %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
		case f3XOR:
			return fmt.Sprintf("xor   %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
		case f3OR:
			return fmt.Sprintf("or    %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
		case f3AND:
			return fmt.Sprintf("and   %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
		case f3SLL:
			return fmt.Sprintf("sll   %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
		case f3SRx:
			if f7 == funct7SUBSRA {
				return fmt.Sprintf("sra   %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
			}
			return fmt.Sprintf("srl   %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
		case f3SLT:
			return fmt.Sprintf("slt   %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
		case f3SLTU:
			return fmt.Sprintf("sltu  %s, %s, %s", rn(rd), rn(rs1), rn(rs2))
		}
	case opSYSTEM:
		if inst == 0x00000073 {
			return "ecall"
		}
		return "system"
	}
	return fmt.Sprintf(".word 0x%08x", inst)
}
