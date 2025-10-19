package sim

// Opcodes
const (
	OpLUI    = 0x37
	opAUIPC  = 0x17
	opJAL    = 0x6F
	opJALR   = 0x67
	opBRANCH = 0x63
	OpLOAD   = 0x03
	opSTORE  = 0x23
	OpOPIMM  = 0x13
	opOP     = 0x33
	opSYSTEM = 0x73
)

// funct7 markers
const (
	funct7SUBSRA = 0x20 // SUB / SRA / SRAI
)

// funct3 values
const (
	// BRANCH
	F3BEQ  = 0x0
	f3BNE  = 0x1
	f3BLT  = 0x4
	f3BGE  = 0x5
	f3BLTU = 0x6
	f3BGEU = 0x7

	// LOAD
	f3LB  = 0x0
	F3LBU = 0x4
	f3LW  = 0x2

	// STORE
	F3SB = 0x0
	f3SW = 0x2

	// OP-IMM
	F3ADDI = 0x0
	f3XORI = 0x4
	f3ORI  = 0x6
	f3ANDI = 0x7
	f3SLLI = 0x1
	f3SRxI = 0x5 // SRLI / SRAI (via funct7)

	// OP
	f3ADD_SUB = 0x0
	f3XOR     = 0x4
	f3OR      = 0x6
	f3AND     = 0x7
	f3SLL     = 0x1
	f3SRx     = 0x5 // SRL / SRA (via funct7)
	f3SLT     = 0x2
	f3SLTU    = 0x3
)
