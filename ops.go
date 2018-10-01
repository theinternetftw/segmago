package segmago

import "fmt"

func (z *z80) setOp8(cycles uint, instLen uint16, reg *uint8, val uint8, flags uint32) {
	z.RunCycles(cycles)
	z.PC += instLen
	z.R += byte(instLen)
	*reg = val
	z.setFlags(flags)
}

func (z *z80) setOpA(cycles uint, instLen uint16, val uint8, flags uint32) {
	z.setOp8(cycles, instLen, &z.A, val, flags)
}
func (z *z80) setOpB(cycles uint, instLen uint16, val uint8, flags uint32) {
	z.setOp8(cycles, instLen, &z.B, val, flags)
}
func (z *z80) setOpC(cycles uint, instLen uint16, val uint8, flags uint32) {
	z.setOp8(cycles, instLen, &z.C, val, flags)
}
func (z *z80) setOpD(cycles uint, instLen uint16, val uint8, flags uint32) {
	z.setOp8(cycles, instLen, &z.D, val, flags)
}
func (z *z80) setOpE(cycles uint, instLen uint16, val uint8, flags uint32) {
	z.setOp8(cycles, instLen, &z.E, val, flags)
}
func (z *z80) setOpL(cycles uint, instLen uint16, val uint8, flags uint32) {
	z.setOp8(cycles, instLen, &z.L, val, flags)
}
func (z *z80) setOpH(cycles uint, instLen uint16, val uint8, flags uint32) {
	z.setOp8(cycles, instLen, &z.H, val, flags)
}

func (z *z80) setOp16(cycles uint, instLen uint16, setFn func(uint16), val uint16, flags uint32) {
	z.RunCycles(cycles)
	z.PC += instLen
	z.R += byte(instLen)
	setFn(val)
	z.setFlags(flags)
}

func (z *z80) setOpHL(cycles uint, instLen uint16, val uint16, flags uint32) {
	z.setOp16(cycles, instLen, z.setHL, val, flags)
}
func (z *z80) setOpBC(cycles uint, instLen uint16, val uint16, flags uint32) {
	z.setOp16(cycles, instLen, z.setBC, val, flags)
}
func (z *z80) setOpDE(cycles uint, instLen uint16, val uint16, flags uint32) {
	z.setOp16(cycles, instLen, z.setDE, val, flags)
}
func (z *z80) setOpSP(cycles uint, instLen uint16, val uint16, flags uint32) {
	z.setOp16(cycles, instLen, z.setSP, val, flags)
}
func (z *z80) setOpPC(cycles uint, instLen uint16, val uint16, flags uint32) {
	z.setOp16(cycles, instLen, z.setPC, val, flags)
}
func (z *z80) setOpIX(cycles uint, instLen uint16, val uint16, flags uint32) {
	z.setOp16(cycles, instLen, z.setIX, val, flags)
}
func (z *z80) setOpIY(cycles uint, instLen uint16, val uint16, flags uint32) {
	z.setOp16(cycles, instLen, z.setIY, val, flags)
}

func (z *z80) setOpMem8(cycles uint, instLen uint16, addr uint16, val uint8, flags uint32) {
	z.RunCycles(cycles)
	z.PC += instLen
	z.R += byte(instLen)
	z.Write(addr, val)
	z.setFlags(flags)
}
func (z *z80) setOpMem16(cycles uint, instLen uint16, addr uint16, val uint16, flags uint32) {
	z.RunCycles(cycles)
	z.PC += instLen
	z.R += byte(instLen)
	z.write16(addr, val)
	z.setFlags(flags)
}

func (z *z80) jmpRel8(cyclesTaken uint, cyclesNotTaken uint, instLen uint16, test bool, relAddr int8) {
	z.PC += instLen
	z.R += byte(instLen)
	if test {
		z.RunCycles(cyclesTaken)
		z.PC = uint16(int(z.PC) + int(relAddr))
	} else {
		z.RunCycles(cyclesNotTaken)
	}
}
func (z *z80) jmpAbs16(cyclesTaken uint, cyclesNotTaken uint, instLen uint16, test bool, addr uint16) {
	z.PC += instLen
	z.R += byte(instLen)
	if test {
		z.RunCycles(cyclesTaken)
		z.PC = addr
	} else {
		z.RunCycles(cyclesNotTaken)
	}
}

func (z *z80) jmpCall(cyclesTaken uint, cyclesNotTaken uint, instLen uint16, test bool, addr uint16) {
	if test {
		z.pushOp16(cyclesTaken, instLen, z.PC+instLen)
		z.PC = addr
	} else {
		z.setOpFn(cyclesNotTaken, instLen, func() {}, 0x22222222)
	}
}
func (z *z80) jmpRet(cyclesTaken uint, cyclesNotTaken uint, instLen uint16, test bool) {
	if test {
		z.popOp16(cyclesTaken, instLen, z.setPC)
	} else {
		z.setOpFn(cyclesNotTaken, instLen, func() {}, 0x22222222)
	}
}

func (z *z80) followBC() byte { return z.Read(z.getBC()) }
func (z *z80) followDE() byte { return z.Read(z.getDE()) }
func (z *z80) followHL() byte { return z.Read(z.getHL()) }
func (z *z80) followSP() byte { return z.Read(z.SP) }
func (z *z80) followPC() byte { return z.Read(z.PC) }

// reminder: flags == sign, zero, f5, halfcarry, f3, parityOverflow, add/sub, carry
// set all: 0x11111111
// clear all: 0x00000000
// ignore all: 0x22222222

var opcodeNamesNormal = [256]string{
	"NOP",
	"LD BC,nn",
	"LD (BC),A",
	"INC BC",
	"INC B",
	"DEC B",
	"LD B,n",
	"RLCA",
	"EX AF,AF",
	"ADD HL,BC",
	"LD A,(BC)",
	"DEC BC",
	"INC C",
	"DEC C",
	"LD C,n",
	"RRCA",
	"DJNZ (PC+e)",
	"LD DE,nn",
	"LD (DE),A",
	"INC DE",
	"INC D",
	"DEC D",
	"LD D,n",
	"RLA",
	"JR (PC+e)",
	"ADD HL,DE",
	"LD A,(DE)",
	"DEC DE",
	"INC E",
	"DEC E",
	"LD E,n",
	"RRA",
	"JR NZ,(PC+e)",
	"LD HL,nn",
	"LD (nn),HL",
	"INC HL",
	"INC H",
	"DEC H",
	"LD H,n",
	"DAA",
	"JR Z,(PC+e)",
	"ADD HL,HL",
	"LD HL,(nn)",
	"DEC HL",
	"INC L",
	"DEC L",
	"LD L,n",
	"CPL",
	"JR NC,(PC+e)",
	"LD SP,nn",
	"LD (nn),A",
	"INC SP",
	"INC (HL)",
	"DEC (HL)",
	"LD (HL),n",
	"SCF",
	"JR C,(PC+e)",
	"ADD HL,SP",
	"LD A,(nn)",
	"DEC SP",
	"INC A",
	"DEC A",
	"LD A,n",
	"CCF",
	"LD B,B",
	"LD B,C",
	"LD B,D",
	"LD B,E",
	"LD B,H",
	"LD B,L",
	"LD B,(HL)",
	"LD B,A",
	"LD C,B",
	"LD C,C",
	"LD C,D",
	"LD C,E",
	"LD C,H",
	"LD C,L",
	"LD C,(HL)",
	"LD C,A",
	"LD D,B",
	"LD D,C",
	"LD D,D",
	"LD D,E",
	"LD D,H",
	"LD D,L",
	"LD D,(HL)",
	"LD D,A",
	"LD E,B",
	"LD E,C",
	"LD E,D",
	"LD E,E",
	"LD E,H",
	"LD E,L",
	"LD E,(HL)",
	"LD E,A",
	"LD H,B",
	"LD H,C",
	"LD H,D",
	"LD H,E",
	"LD H,H",
	"LD H,L",
	"LD H,(HL)",
	"LD H,A",
	"LD L,B",
	"LD L,C",
	"LD L,D",
	"LD L,E",
	"LD L,H",
	"LD L,L",
	"LD L,(HL)",
	"LD L,A",
	"LD (HL),B",
	"LD (HL),C",
	"LD (HL),D",
	"LD (HL),E",
	"LD (HL),H",
	"LD (HL),L",
	"HALT",
	"LD (HL),A",
	"LD A,B",
	"LD A,C",
	"LD A,D",
	"LD A,E",
	"LD A,H",
	"LD A,L",
	"LD A,(HL)",
	"LD A,A",
	"ADD A,B",
	"ADD A,C",
	"ADD A,D",
	"ADD A,E",
	"ADD A,H",
	"ADD A,L",
	"ADD A,(HL)",
	"ADD A,A",
	"ADC A,B",
	"ADC A,C",
	"ADC A,D",
	"ADC A,E",
	"ADC A,H",
	"ADC A,L",
	"ADC A,(HL)",
	"ADC A,A",
	"SUB B",
	"SUB C",
	"SUB D",
	"SUB E",
	"SUB H",
	"SUB L",
	"SUB (HL)",
	"SUB A",
	"SBC A,B",
	"SBC A,C",
	"SBC A,D",
	"SBC A,E",
	"SBC A,H",
	"SBC A,L",
	"SBC A,(HL)",
	"SBC A,A",
	"AND B",
	"AND C",
	"AND D",
	"AND E",
	"AND H",
	"AND L",
	"AND (HL)",
	"AND A",
	"XOR B",
	"XOR C",
	"XOR D",
	"XOR E",
	"XOR H",
	"XOR L",
	"XOR (HL)",
	"XOR A",
	"OR B",
	"OR C",
	"OR D",
	"OR E",
	"OR H",
	"OR L",
	"OR (HL)",
	"OR A",
	"CP B",
	"CP C",
	"CP D",
	"CP E",
	"CP H",
	"CP L",
	"CP (HL)",
	"CP A",
	"RET NZ",
	"POP BC",
	"JP NZ,(nn)",
	"JP (nn)",
	"CALL NZ,(nn)",
	"PUSH BC",
	"ADD A,n",
	"RST 0H",
	"RET Z",
	"RET",
	"JP Z,(nn)",
	"BIT OP PREFIX",
	"CALL Z,(nn)",
	"CALL (nn)",
	"ADC A,n",
	"RST 8H",
	"RET NC",
	"POP DE",
	"JP NC,(nn)",
	"OUT (n),A",
	"CALL NC,(nn)",
	"PUSH DE",
	"SUB n",
	"RST 10H",
	"RET C",
	"EXX",
	"JP C,(nn)",
	"IN A,(n)",
	"CALL C,(nn)",
	"DD PREFIX",
	"SBC A,n",
	"RST 18H",
	"RET PO",
	"POP HL",
	"JP PO,(nn)",
	"EX (SP),HL",
	"CALL PO,(nn)",
	"PUSH HL",
	"AND n",
	"RST 20H",
	"RET PE",
	"JP (HL)",
	"JP PE,(nn)",
	"EX DE,HL",
	"CALL PE,(nn)",
	"EXTENDED OP PREFIX",
	"XOR n",
	"RST 28H",
	"RET P",
	"POP AF",
	"JP P,(nn)",
	"DI",
	"CALL P,(nn)",
	"PUSH AF",
	"OR n",
	"RST 30H",
	"RET M",
	"LD SP,HL",
	"JP M,(nn)",
	"EI",
	"CALL M,(nn)",
	"FD PREFIX",
	"CP n",
	"RST 38H",
}
var opcodeNamesCBPrefix = [256]string{
	"RLC B",
	"RLC C",
	"RLC D",
	"RLC E",
	"RLC H",
	"RLC L",
	"RLC (HL)",
	"RLC A",
	"RRC B",
	"RRC C",
	"RRC D",
	"RRC E",
	"RRC H",
	"RRC L",
	"RRC (HL)",
	"RRC A",
	"RL B",
	"RL C",
	"RL D",
	"RL E",
	"RL H",
	"RL L",
	"RL (HL)",
	"RL A",
	"RR B",
	"RR C",
	"RR D",
	"RR E",
	"RR H",
	"RR L",
	"RR (HL)",
	"RR A",
	"SLA B",
	"SLA C",
	"SLA D",
	"SLA E",
	"SLA H",
	"SLA L",
	"SLA (HL)",
	"SLA A",
	"SRA B",
	"SRA C",
	"SRA D",
	"SRA E",
	"SRA H",
	"SRA L",
	"SRA (HL)",
	"SRA A",
	"SLL B*",
	"SLL C*",
	"SLL D*",
	"SLL E*",
	"SLL H*",
	"SLL L*",
	"SLL (HL)*",
	"SLL A*",
	"SRL B",
	"SRL C",
	"SRL D",
	"SRL E",
	"SRL H",
	"SRL L",
	"SRL (HL)",
	"SRL A",
	"BIT 0,B",
	"BIT 0,C",
	"BIT 0,D",
	"BIT 0,E",
	"BIT 0,H",
	"BIT 0,L",
	"BIT 0,(HL)",
	"BIT 0,A",
	"BIT 1,B",
	"BIT 1,C",
	"BIT 1,D",
	"BIT 1,E",
	"BIT 1,H",
	"BIT 1,L",
	"BIT 1,(HL)",
	"BIT 1,A",
	"BIT 2,B",
	"BIT 2,C",
	"BIT 2,D",
	"BIT 2,E",
	"BIT 2,H",
	"BIT 2,L",
	"BIT 2,(HL)",
	"BIT 2,A",
	"BIT 3,B",
	"BIT 3,C",
	"BIT 3,D",
	"BIT 3,E",
	"BIT 3,H",
	"BIT 3,L",
	"BIT 3,(HL)",
	"BIT 3,A",
	"BIT 4,B",
	"BIT 4,C",
	"BIT 4,D",
	"BIT 4,E",
	"BIT 4,H",
	"BIT 4,L",
	"BIT 4,(HL)",
	"BIT 4,A",
	"BIT 5,B",
	"BIT 5,C",
	"BIT 5,D",
	"BIT 5,E",
	"BIT 5,H",
	"BIT 5,L",
	"BIT 5,(HL)",
	"BIT 5,A",
	"BIT 6,B",
	"BIT 6,C",
	"BIT 6,D",
	"BIT 6,E",
	"BIT 6,H",
	"BIT 6,L",
	"BIT 6,(HL)",
	"BIT 6,A",
	"BIT 7,B",
	"BIT 7,C",
	"BIT 7,D",
	"BIT 7,E",
	"BIT 7,H",
	"BIT 7,L",
	"BIT 7,(HL)",
	"BIT 7,A",
	"RES 0,B",
	"RES 0,C",
	"RES 0,D",
	"RES 0,E",
	"RES 0,H",
	"RES 0,L",
	"RES 0,(HL)",
	"RES 0,A",
	"RES 1,B",
	"RES 1,C",
	"RES 1,D",
	"RES 1,E",
	"RES 1,H",
	"RES 1,L",
	"RES 1,(HL)",
	"RES 1,A",
	"RES 2,B",
	"RES 2,C",
	"RES 2,D",
	"RES 2,E",
	"RES 2,H",
	"RES 2,L",
	"RES 2,(HL)",
	"RES 2,A",
	"RES 3,B",
	"RES 3,C",
	"RES 3,D",
	"RES 3,E",
	"RES 3,H",
	"RES 3,L",
	"RES 3,(HL)",
	"RES 3,A",
	"RES 4,B",
	"RES 4,C",
	"RES 4,D",
	"RES 4,E",
	"RES 4,H",
	"RES 4,L",
	"RES 4,(HL)",
	"RES 4,A",
	"RES 5,B",
	"RES 5,C",
	"RES 5,D",
	"RES 5,E",
	"RES 5,H",
	"RES 5,L",
	"RES 5,(HL)",
	"RES 5,A",
	"RES 6,B",
	"RES 6,C",
	"RES 6,D",
	"RES 6,E",
	"RES 6,H",
	"RES 6,L",
	"RES 6,(HL)",
	"RES 6,A",
	"RES 7,B",
	"RES 7,C",
	"RES 7,D",
	"RES 7,E",
	"RES 7,H",
	"RES 7,L",
	"RES 7,(HL)",
	"RES 7,A",
	"SET 0,B",
	"SET 0,C",
	"SET 0,D",
	"SET 0,E",
	"SET 0,H",
	"SET 0,L",
	"SET 0,(HL)",
	"SET 0,A",
	"SET 1,B",
	"SET 1,C",
	"SET 1,D",
	"SET 1,E",
	"SET 1,H",
	"SET 1,L",
	"SET 1,(HL)",
	"SET 1,A",
	"SET 2,B",
	"SET 2,C",
	"SET 2,D",
	"SET 2,E",
	"SET 2,H",
	"SET 2,L",
	"SET 2,(HL)",
	"SET 2,A",
	"SET 3,B",
	"SET 3,C",
	"SET 3,D",
	"SET 3,E",
	"SET 3,H",
	"SET 3,L",
	"SET 3,(HL)",
	"SET 3,A",
	"SET 4,B",
	"SET 4,C",
	"SET 4,D",
	"SET 4,E",
	"SET 4,H",
	"SET 4,L",
	"SET 4,(HL)",
	"SET 4,A",
	"SET 5,B",
	"SET 5,C",
	"SET 5,D",
	"SET 5,E",
	"SET 5,H",
	"SET 5,L",
	"SET 5,(HL)",
	"SET 5,A",
	"SET 6,B",
	"SET 6,C",
	"SET 6,D",
	"SET 6,E",
	"SET 6,H",
	"SET 6,L",
	"SET 6,(HL)",
	"SET 6,A",
	"SET 7,B",
	"SET 7,C",
	"SET 7,D",
	"SET 7,E",
	"SET 7,H",
	"SET 7,L",
	"SET 7,(HL)",
	"SET 7,A",
}
var opcodeNamesEDPrefix = map[byte]string{
	0x40: "IN B,(C)",
	0x41: "OUT (C),B",
	0x42: "SBC HL,BC",
	0x43: "LD (nn),BC",
	0x44: "NEG",
	0x45: "RETN",
	0x46: "IM 0",
	0x47: "LD I,A",
	0x48: "IN C,(C)",
	0x49: "OUT (C),C",
	0x4A: "ADC HL,BC",
	0x4B: "LD BC,(nn)",
	0x4C: "NEG*",
	0x4D: "RETI",
	0x4E: "IM 0/1*",
	0x4F: "LD R,A",
	0x50: "IN D,(C)",
	0x51: "OUT (C),D",
	0x52: "SBC HL,DE",
	0x53: "LD (nn),DE",
	0x54: "NEG*",
	0x55: "RETN*",
	0x56: "IM 1",
	0x57: "LD A,I",
	0x58: "IN E,(C)",
	0x59: "OUT (C),E",
	0x5A: "ADC HL,DE",
	0x5B: "LD DE,(nn)",
	0x5C: "NEG*",
	0x5D: "RETN*",
	0x5E: "IM 2",
	0x5F: "LD A,R",
	0x60: "IN H,(C)",
	0x61: "OUT (C),H",
	0x62: "SBC HL,HL",
	0x63: "LD (nn),HL",
	0x64: "NEG*",
	0x65: "RETN*",
	0x66: "IM 0*",
	0x67: "RRD",
	0x68: "IN L,(C)",
	0x69: "OUT (C),L",
	0x6A: "ADC HL,HL",
	0x6B: "LD HL,(nn)",
	0x6C: "NEG*",
	0x6D: "RETN*",
	0x6E: "IM 0/1*",
	0x6F: "RLD",
	0x70: "IN F,(C)* / IN (C)*",
	0x71: "OUT (C),0*",
	0x72: "SBC HL,SP",
	0x73: "LD (nn),SP	",
	0x74: "NEG*",
	0x75: "RETN*",
	0x76: "IM 1*",
	0x78: "IN A,(C)",
	0x79: "OUT (C),A",
	0x7A: "ADC HL,SP",
	0x7B: "LD SP,(nn)",
	0x7C: "NEG*",
	0x7D: "RETN*",
	0x7E: "IM 2*",
	0xA0: "LDI",
	0xA1: "CPI",
	0xA2: "INI",
	0xA3: "OUTI",
	0xA8: "LDD",
	0xA9: "CPD",
	0xAA: "IND",
	0xAB: "OUTD",
	0xB0: "LDIR",
	0xB1: "CPIR",
	0xB2: "INIR",
	0xB3: "OTIR",
	0xB8: "LDDR",
	0xB9: "CPDR",
	0xBA: "INDR",
	0xBB: "OTDR",
}

/*
DD09		ADD IX,BC
DD19		ADD IX,DE
DD21	LD IX,nn
DD22	LD (nn),IX
DD23		INC IX
DD24		INC IXH*
DD25		DEC IXH*
DD26 		LD IXH,n*
DD29		ADD IX,IX
DD2A	LD IX,(nn)
DD2B		DEC IX
DD2C		INC IXL*
DD2D		DEC IXL*
DD2E		LD IXL,n*
DD34 d		INC (IX+d)
DD35 d		DEC (IX+d)
DD36 d	LD (IX+d),n
DD39		ADD IX,SP
DD44		LD B,IXH*
DD45		LD B,IXL*
DD46 d		LD B,(IX+d)
DD4C		LD C,IXH*
DD4D		LD C,IXL*
DD4E d		LD C,(IX+d)
DD54		LD D,IXH*
DD55		LD D,IXL*
DD56 d		LD D,(IX+d)
DD5C		LD E,IXH*
DD5D		LD E,IXL*
DD5E d		LD E,(IX+d)
DD60		LD IXH,B*
DD61		LD IXH,C*
DD62		LD IXH,D*
DD63		LD IXH,E*
DD64		LD IXH,IXH*
DD65		LD IXH,IXL*
DD66 d		LD H,(IX+d)
DD67		LD IXH,A*
DD68		LD IXL,B*
DD69		LD IXL,C*
DD6A		LD IXL,D*
DD6B		LD IXL,E*
DD6C		LD IXL,IXH*
DD6D		LD IXL,IXL*
DD6E d		LD L,(IX+d)
DD6F		LD IXL,A*
DD70 d		LD (IX+d),B
DD71 d		LD (IX+d),C
DD72 d		LD (IX+d),D
DD73 d		LD (IX+d),E
DD74 d		LD (IX+d),H
DD75 d		LD (IX+d),L
DD77 d		LD (IX+d),A
DD7C		LD A,IXH*
DD7D		LD A,IXL*
DD7E d		LD A,(IX+d)
DD84		ADD A,IXH*
DD85		ADD A,IXL*
DD86 d		ADD A,(IX+d)
DD8C		ADC A,IXH*
DD8D		ADC A,IXL*
DD8E d		ADC A,(IX+d)
DD94		SUB IXH*
DD95		SUB IXL*
DD96 d		SUB (IX+d)
DD9C		SBC A,IXH*
DD9D		SBC A,IXL*
DD9E d		SBC A,(IX+d)
DDA4		AND IXH*
DDA5		AND IXL*
DDA6 d		AND (IX+d)
DDAC		XOR IXH*
DDAD		XOR IXL*
DDAE d		XOR (IX+d)
DDB4		OR IXH*
DDB5		OR IXL*
DDB6 d		OR (IX+d)
DDBC		CP IXH*
DDBD		CP IXL*
DDBE d		CP (IX+d)
DDE1		POP IX
DDE3		EX (SP),IX
DDE5		PUSH IX
DDE9		JP (IX)
DDF9		LD SP,IX
FD09		ADD IY,BC
FD19		ADD IY,DE
FD21	LD IY,nn
FD22	LD (nn),IY
FD23		INC IY
FD24		INC IYH*
FD25		DEC IYH*
FD26		LD IYH,n*
FD29		ADD IY,IY
FD2A	LD IY,(nn)
FD2B		DEC IY
FD2C		INC IYL*
FD2D		DEC IYL*
FD2E		LD IYL,n*
FD34 d		INC (IY+d)
FD35 d		DEC (IY+d)
FD36 d	LD (IY+d),n
FD39		ADD IY,SP
FD44		LD B,IYH*
FD45		LD B,IYL*
FD46 d		LD B,(IY+d)
FD4C		LD C,IYH*
FD4D		LD C,IYL*
FD4E d		LD C,(IY+d)
FD54		LD D,IYH*
FD55		LD D,IYL*
FD56 d		LD D,(IY+d)
FD5C		LD E,IYH*
FD5D		LD E,IYL*
FD5E d		LD E,(IY+d)
FD60		LD IYH,B*
FD61		LD IYH,C*
FD62		LD IYH,D*
FD63		LD IYH,E*
FD64		LD IYH,IYH*
FD65		LD IYH,IYL*
FD66 d		LD H,(IY+d)
FD67		LD IYH,A*
FD68		LD IYL,B*
FD69		LD IYL,C*
FD6A		LD IYL,D*
FD6B		LD IYL,E*
FD6C		LD IYL,IYH*
FD6D		LD IYL,IYL*
FD6E d		LD L,(IY+d)
FD6F		LD IYL,A*
FD70 d		LD (IY+d),B
FD71 d		LD (IY+d),C
FD72 d		LD (IY+d),D
FD73 d		LD (IY+d),E
FD74 d		LD (IY+d),H
FD75 d		LD (IY+d),L
FD77 d		LD (IY+d),A
FD7C		LD A,IYH*
FD7D		LD A,IYL*
FD7E d		LD A,(IY+d)
FD84		ADD A,IYH*
FD85		ADD A,IYL*
FD86 d		ADD A,(IY+d)
FD8C		ADC A,IYH*
FD8D		ADC A,IYL*
FD8E d		ADC A,(IY+d)
FD94		SUB IYH*
FD95		SUB IYL*
FD96 d		SUB (IY+d)
FD9C		SBC A,IYH*
FD9D		SBC A,IYL*
FD9E d		SBC A,(IY+d)
FDA4		AND IYH*
FDA5		AND IYL*
FDA6 d		AND (IY+d)
FDAC		XOR IYH*
FDAD		XOR IYL*
FDAE d		XOR (IY+d)
FDB4		OR IYH*
FDB5		OR IYL*
FDB6 d		OR (IY+d)
FDBC		CP IYH*
FDBD		CP IYL*
FDBE d		CP (IY+d)
FDCB d 00	LD B,RLC (IY+d)*
FDCB d 01	LD C,RLC (IY+d)*
FDCB d 02	LD D,RLC (IY+d)*
FDCB d 03	LD E,RLC (IY+d)*
FDCB d 04	LD H,RLC (IY+d)*
FDCB d 05	LD L,RLC (IY+d)*
FDCB d 06	RLC (IY+d)
FDCB d 07	LD A,RLC (IY+d)*
FDCB d 08	LD B,RRC (IY+d)*
FDCB d 09	LD C,RRC (IY+d)*
FDCB d 0A	LD D,RRC (IY+d)*
FDCB d 0B	LD E,RRC (IY+d)*
FDCB d 0C	LD H,RRC (IY+d)*
FDCB d 0D	LD L,RRC (IY+d)*
FDCB d 0E	RRC (IY+d)
FDCB d 0F	LD A,RRC (IY+d)*
FDCB d 10	LD B,RL (IY+d)*
FDCB d 11	LD C,RL (IY+d)*
FDCB d 12	LD D,RL (IY+d)*
FDCB d 13	LD E,RL (IY+d)*
FDCB d 14	LD H,RL (IY+d)*
FDCB d 15	LD L,RL (IY+d)*
FDCB d 16	RL (IY+d)
FDCB d 17	LD A,RL (IY+d)*
FDCB d 18	LD B,RR (IY+d)*
FDCB d 19	LD C,RR (IY+d)*
FDCB d 1A	LD D,RR (IY+d)*
FDCB d 1B	LD E,RR (IY+d)*
FDCB d 1C	LD H,RR (IY+d)*
FDCB d 1D	LD L,RR (IY+d)*
FDCB d 1E	RR (IY+d)
FDCB d 1F	LD A,RR (IY+d)*
FDCB d 20	LD B,SLA (IY+d)*
FDCB d 21	LD C,SLA (IY+d)*
FDCB d 22	LD D,SLA (IY+d)*
FDCB d 23	LD E,SLA (IY+d)*
FDCB d 24	LD H,SLA (IY+d)*
FDCB d 25	LD L,SLA (IY+d)*
FDCB d 26	SLA (IY+d)
FDCB d 27	LD A,SLA (IY+d)*
FDCB d 28	LD B,SRA (IY+d)*
FDCB d 29	LD C,SRA (IY+d)*
FDCB d 2A	LD D,SRA (IY+d)*
FDCB d 2B	LD E,SRA (IY+d)*
FDCB d 2C	LD H,SRA (IY+d)*
FDCB d 2D	LD L,SRA (IY+d)*
FDCB d 2E	SRA (IY+d)
FDCB d 2F	LD A,SRA (IY+d)*
FDCB d 30	LD B,SLL (IY+d)*
FDCB d 31	LD C,SLL (IY+d)*
FDCB d 32	LD D,SLL (IY+d)*
FDCB d 33	LD E,SLL (IY+d)*
FDCB d 34	LD H,SLL (IY+d)*
FDCB d 35	LD L,SLL (IY+d)*
FDCB d 36	SLL (IY+d)*
FDCB d 37	LD A,SLL (IY+d)*
FDCB d 38	LD B,SRL (IY+d)*
FDCB d 39	LD C,SRL (IY+d)*
FDCB d 3A	LD D,SRL (IY+d)*
FDCB d 3B	LD E,SRL (IY+d)*
FDCB d 3C	LD H,SRL (IY+d)*
FDCB d 3D	LD L,SRL (IY+d)*
FDCB d 3E	SRL (IY+d)
FDCB d 3F	LD A,SRL (IY+d)*
FDCB d 40	BIT 0,(IY+d)*
FDCB d 41	BIT 0,(IY+d)*
FDCB d 42	BIT 0,(IY+d)*
FDCB d 43	BIT 0,(IY+d)*
FDCB d 44	BIT 0,(IY+d)*
FDCB d 45	BIT 0,(IY+d)*
FDCB d 46	BIT 0,(IY+d)
FDCB d 47	BIT 0,(IY+d)*
FDCB d 48	BIT 1,(IY+d)*
FDCB d 49	BIT 1,(IY+d)*
FDCB d 4A	BIT 1,(IY+d)*
FDCB d 4B	BIT 1,(IY+d)*
FDCB d 4C	BIT 1,(IY+d)*
FDCB d 4D	BIT 1,(IY+d)*
FDCB d 4E	BIT 1,(IY+d)
FDCB d 4F	BIT 1,(IY+d)*
FDCB d 50	BIT 2,(IY+d)*
FDCB d 51	BIT 2,(IY+d)*
FDCB d 52	BIT 2,(IY+d)*
FDCB d 53	BIT 2,(IY+d)*
FDCB d 54	BIT 2,(IY+d)*
FDCB d 55	BIT 2,(IY+d)*
FDCB d 56	BIT 2,(IY+d)
FDCB d 57	BIT 2,(IY+d)*
FDCB d 58	BIT 3,(IY+d)*
FDCB d 59	BIT 3,(IY+d)*
FDCB d 5A	BIT 3,(IY+d)*
FDCB d 5B	BIT 3,(IY+d)*
FDCB d 5C	BIT 3,(IY+d)*
FDCB d 5D	BIT 3,(IY+d)*
FDCB d 5E	BIT 3,(IY+d)
FDCB d 5F	BIT 3,(IY+d)*
FDCB d 60	BIT 4,(IY+d)*
FDCB d 61	BIT 4,(IY+d)*
FDCB d 62	BIT 4,(IY+d)*
FDCB d 63	BIT 4,(IY+d)*
FDCB d 64	BIT 4,(IY+d)*
FDCB d 65	BIT 4,(IY+d)*
FDCB d 66	BIT 4,(IY+d)
FDCB d 67	BIT 4,(IY+d)*
FDCB d 68	BIT 5,(IY+d)*
FDCB d 69	BIT 5,(IY+d)*
FDCB d 6A	BIT 5,(IY+d)*
FDCB d 6B	BIT 5,(IY+d)*
FDCB d 6C	BIT 5,(IY+d)*
FDCB d 6D	BIT 5,(IY+d)*
FDCB d 6E	BIT 5,(IY+d)
FDCB d 6F	BIT 5,(IY+d)*
FDCB d 70	BIT 6,(IY+d)*
FDCB d 71	BIT 6,(IY+d)*
FDCB d 72	BIT 6,(IY+d)*
FDCB d 73	BIT 6,(IY+d)*
FDCB d 74	BIT 6,(IY+d)*
FDCB d 75	BIT 6,(IY+d)*
FDCB d 76	BIT 6,(IY+d)
FDCB d 77	BIT 6,(IY+d)*
FDCB d 78	BIT 7,(IY+d)*
FDCB d 79	BIT 7,(IY+d)*
FDCB d 7A	BIT 7,(IY+d)*
FDCB d 7B	BIT 7,(IY+d)*
FDCB d 7C	BIT 7,(IY+d)*
FDCB d 7D	BIT 7,(IY+d)*
FDCB d 7E	BIT 7,(IY+d)
FDCB d 7F	BIT 7,(IY+d)*
FDCB d 80	LD B,RES 0,(IY+d)*
FDCB d 81	LD C,RES 0,(IY+d)*
FDCB d 82	LD D,RES 0,(IY+d)*
FDCB d 83	LD E,RES 0,(IY+d)*
FDCB d 84	LD H,RES 0,(IY+d)*
FDCB d 85	LD L,RES 0,(IY+d)*
FDCB d 86	RES 0,(IY+d)
FDCB d 87	LD A,RES 0,(IY+d)*
FDCB d 88	LD B,RES 1,(IY+d)*
FDCB d 89	LD C,RES 1,(IY+d)*
FDCB d 8A	LD D,RES 1,(IY+d)*
FDCB d 8B	LD E,RES 1,(IY+d)*
FDCB d 8C	LD H,RES 1,(IY+d)*
FDCB d 8D	LD L,RES 1,(IY+d)*
FDCB d 8E	RES 1,(IY+d)
FDCB d 8F	LD A,RES 1,(IY+d)*
FDCB d 90	LD B,RES 2,(IY+d)*
FDCB d 91	LD C,RES 2,(IY+d)*
FDCB d 92	LD D,RES 2,(IY+d)*
FDCB d 93	LD E,RES 2,(IY+d)*
FDCB d 94	LD H,RES 2,(IY+d)*
FDCB d 95	LD L,RES 2,(IY+d)*
FDCB d 96	RES 2,(IY+d)
FDCB d 97	LD A,RES 2,(IY+d)*
FDCB d 98	LD B,RES 3,(IY+d)*
FDCB d 99	LD C,RES 3,(IY+d)*
FDCB d 9A	LD D,RES 3,(IY+d)*
FDCB d 9B	LD E,RES 3,(IY+d)*
FDCB d 9C	LD H,RES 3,(IY+d)*
FDCB d 9D	LD L,RES 3,(IY+d)*
FDCB d 9E	RES 3,(IY+d)
FDCB d 9F	LD A,RES 3,(IY+d)*
FDCB d A0	LD B,RES 4,(IY+d)*
FDCB d A1	LD C,RES 4,(IY+d)*
FDCB d A2	LD D,RES 4,(IY+d)*
FDCB d A3	LD E,RES 4,(IY+d)*
FDCB d A4	LD H,RES 4,(IY+d)*
FDCB d A5	LD L,RES 4,(IY+d)*
FDCB d A6	RES 4,(IY+d)
FDCB d A7	LD A,RES 4,(IY+d)*
FDCB d A8	LD B,RES 5,(IY+d)*
FDCB d A9	LD C,RES 5,(IY+d)*
FDCB d AA	LD D,RES 5,(IY+d)*
FDCB d AB	LD E,RES 5,(IY+d)*
FDCB d AC	LD H,RES 5,(IY+d)*
FDCB d AD	LD L,RES 5,(IY+d)*
FDCB d AE	RES 5,(IY+d)
FDCB d AF	LD A,RES 5,(IY+d)*
FDCB d B0	LD B,RES 6,(IY+d)*
FDCB d B1	LD C,RES 6,(IY+d)*
FDCB d B2	LD D,RES 6,(IY+d)*
FDCB d B3	LD E,RES 6,(IY+d)*
FDCB d B4	LD H,RES 6,(IY+d)*
FDCB d B5	LD L,RES 6,(IY+d)*
FDCB d B6	RES 6,(IY+d)
FDCB d B7	LD A,RES 6,(IY+d)*
FDCB d B8	LD B,RES 7,(IY+d)*
FDCB d B9	LD C,RES 7,(IY+d)*
FDCB d BA	LD D,RES 7,(IY+d)*
FDCB d BB	LD E,RES 7,(IY+d)*
FDCB d BC	LD H,RES 7,(IY+d)*
FDCB d BD	LD L,RES 7,(IY+d)*
FDCB d BE	RES 7,(IY+d)
FDCB d BF	LD A,RES 7,(IY+d)*
FDCB d C0	LD B,SET 0,(IY+d)*
FDCB d C1	LD C,SET 0,(IY+d)*
FDCB d C2	LD D,SET 0,(IY+d)*
FDCB d C3	LD E,SET 0,(IY+d)*
FDCB d C4	LD H,SET 0,(IY+d)*
FDCB d C5	LD L,SET 0,(IY+d)*
FDCB d C6	SET 0,(IY+d)
FDCB d C7	LD A,SET 0,(IY+d)*
FDCB d C8	LD B,SET 1,(IY+d)*
FDCB d C9	LD C,SET 1,(IY+d)*
FDCB d CA	LD D,SET 1,(IY+d)*
FDCB d CB	LD E,SET 1,(IY+d)*
FDCB d CC	LD H,SET 1,(IY+d)*
FDCB d CD	LD L,SET 1,(IY+d)*
FDCB d CE	SET 1,(IY+d)
FDCB d CF	LD A,SET 1,(IY+d)*
FDCB d D0	LD B,SET 2,(IY+d)*
FDCB d D1	LD C,SET 2,(IY+d)*
FDCB d D2	LD D,SET 2,(IY+d)*
FDCB d D3	LD E,SET 2,(IY+d)*
FDCB d D4	LD H,SET 2,(IY+d)*
FDCB d D5	LD L,SET 2,(IY+d)*
FDCB d D6	SET 2,(IY+d)
FDCB d D7	LD A,SET 2,(IY+d)*
FDCB d D8	LD B,SET 3,(IY+d)*
FDCB d D9	LD C,SET 3,(IY+d)*
FDCB d DA	LD D,SET 3,(IY+d)*
FDCB d DB	LD E,SET 3,(IY+d)*
FDCB d DC	LD H,SET 3,(IY+d)*
FDCB d DD	LD L,SET 3,(IY+d)*
FDCB d DE	SET 3,(IY+d)
FDCB d DF	LD A,SET 3,(IY+d)*
FDCB d E0	LD B,SET 4,(IY+d)*
FDCB d E1	LD C,SET 4,(IY+d)*
FDCB d E2	LD D,SET 4,(IY+d)*
FDCB d E3	LD E,SET 4,(IY+d)*
FDCB d E4	LD H,SET 4,(IY+d)*
FDCB d E5	LD L,SET 4,(IY+d)*
FDCB d E6	SET 4,(IY+d)
FDCB d E7	LD A,SET 4,(IY+d)*
FDCB d E8	LD B,SET 5,(IY+d)*
FDCB d E9	LD C,SET 5,(IY+d)*
FDCB d EA	LD D,SET 5,(IY+d)*
FDCB d EB	LD E,SET 5,(IY+d)*
FDCB d EC	LD H,SET 5,(IY+d)*
FDCB d ED	LD L,SET 5,(IY+d)*
FDCB d EE	SET 5,(IY+d)
FDCB d EF	LD A,SET 5,(IY+d)*
FDCB d F0	LD B,SET 6,(IY+d)*
FDCB d F1	LD C,SET 6,(IY+d)*
FDCB d F2	LD D,SET 6,(IY+d)*
FDCB d F3	LD E,SET 6,(IY+d)*
FDCB d F4	LD H,SET 6,(IY+d)*
FDCB d F5	LD L,SET 6,(IY+d)*
FDCB d F6	SET 6,(IY+d)
FDCB d F7	LD A,SET 6,(IY+d)*
FDCB d F8	LD B,SET 7,(IY+d)*
FDCB d F9	LD C,SET 7,(IY+d)*
FDCB d FA	LD D,SET 7,(IY+d)*
FDCB d FB	LD E,SET 7,(IY+d)*
FDCB d FC	LD H,SET 7,(IY+d)*
FDCB d FD	LD L,SET 7,(IY+d)*
FDCB d FE	SET 7,(IY+d)
FDCB d FF	LD A,SET 7,(IY+d)*
FDE1		POP IY
FDE3		EX (SP),IY
FDE5		PUSH IY
FDE9		JP (IY)
FDF9		LD SP,IY
*/

func zFlag(val uint8) uint32 {
	if val == 0 {
		return 0x01000000
	}
	return 0
}
func zFlag16(val uint16) uint32 {
	if val == 0 {
		return 0x01000000
	}
	return 0
}

func signFlag(val uint8) uint32 {
	if val&0x80 > 0 {
		return 0x10000000
	}
	return 0
}

func signFlag16(val uint16) uint32 {
	if val&0x8000 > 0 {
		return 0x10000000
	}
	return 0
}

func hFlagAdd(val, addend uint8) uint32 {
	// 4th to 5th bit carry
	if int(val&0x0f)+int(addend&0x0f) >= 0x10 {
		return 0x00010000
	}
	return 0
}

func hFlagAdc(val, addend, fReg uint8) uint32 {
	carry := fReg & 0x01
	// 4th to 5th bit carry
	if int(carry)+int(val&0x0f)+int(addend&0x0f) >= 0x10 {
		return 0x00010000
	}
	return 0
}

func hFlagAdc16(val, addend uint16, fReg uint8) uint32 {
	carry := fReg & 0x01
	// 12th to 13th bit carry
	if int(carry)+int(val&0x0fff)+int(addend&0x0fff) >= 0x1000 {
		return 0x00010000
	}
	return 0
}

func hFlagAdd16(val, addend uint16) uint32 {
	// 12th to 13th bit carry
	if int(val&0x0fff)+int(addend&0x0fff) >= 0x1000 {
		return 0x00010000
	}
	return 0
}

func hFlagSub(val, subtrahend uint8) uint32 {
	if int(val&0xf)-int(subtrahend&0xf) < 0 {
		return 0x00010000
	}
	return 0
}

func hFlagSub16(val, subtrahend uint16) uint32 {
	if int(val&0xfff)-int(subtrahend&0xfff) < 0 {
		return 0x00010000
	}
	return 0
}

func hFlagSbc(val, subtrahend, fReg uint8) uint32 {
	carry := fReg & 1
	if int(val&0xf)-int(subtrahend&0xf)-int(carry) < 0 {
		return 0x00010000
	}
	return 0
}

func hFlagSbc16(val, subtrahend uint16, fReg uint8) uint32 {
	carry := fReg & 1
	if int(val&0xfff)-int(subtrahend&0xfff)-int(carry) < 0 {
		return 0x00010000
	}
	return 0
}

func cFlagAdd(val, addend uint8) uint32 {
	if int(val)+int(addend) > 0xff {
		return 0x1
	}
	return 0x0
}

func cFlagAdc(val, addend, fReg uint8) uint32 {
	carry := fReg & 1
	if int(carry)+int(val)+int(addend) > 0xff {
		return 0x1
	}
	return 0x0
}

func cFlagAdc16(val, addend uint16, fReg uint8) uint32 {
	carry := fReg & 1
	if int(carry)+int(val)+int(addend) > 0xffff {
		return 0x1
	}
	return 0x0
}

func cFlagAdd16(val, addend uint16) uint32 {
	if int(val)+int(addend) > 0xffff {
		return 0x1
	}
	return 0x0
}

func cFlagSub(val, subtrahend uint8) uint32 {
	if int(val)-int(subtrahend) < 0 {
		return 0x1
	}
	return 0x0
}
func cFlagSbc(val, subtrahend, fReg uint8) uint32 {
	carry := fReg & 1
	if int(val)-int(subtrahend)-int(carry) < 0 {
		return 0x1
	}
	return 0x0
}
func cFlagSbc16(val, subtrahend uint16, fReg uint8) uint32 {
	carry := fReg & 1
	if int(val)-int(subtrahend)-int(carry) < 0 {
		return 0x1
	}
	return 0x0
}

func (z *z80) setOpFn(cycles uint, instLen uint16, fn func(), flags uint32) {
	z.RunCycles(cycles)
	z.PC += instLen
	z.R += byte(instLen)
	fn()
	z.setFlags(flags)
}

func (z *z80) pushOp16(cycles uint, instLen uint16, val uint16) {
	z.setOpMem16(cycles, instLen, z.SP-2, val, 0x22222222)
	z.SP -= 2
}
func (z *z80) popOp16(cycles uint, instLen uint16, setFn func(val uint16)) {
	z.setOpFn(cycles, instLen, func() { setFn(z.read16(z.SP)) }, 0x22222222)
	z.SP += 2
}

func undocFlags(result byte) uint32 {
	out := uint32(result&0x20) << (3 + 12)
	out |= uint32(result&0x08) << (1 + 8)
	return out
}

func incOp(val byte) (byte, uint32) {
	result := val + 1
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= hFlagAdd(val, 1)
	flags |= overflagAdd(val, 1, result)
	flags |= undocFlags(result)
	flags |= 0x02
	return result, flags
}

func (z *z80) incOpReg(reg *byte) {
	result, flags := incOp(*reg)
	z.setOp8(4, 1, reg, result, flags)
}

func decOp(val byte) (byte, uint32) {
	result := val - 1
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= hFlagSub(val, 1)
	flags |= undocFlags(result)
	flags |= overflagSub(val, 1, result)
	flags |= 0x12
	return result, flags
}

func (z *z80) decOpReg(reg *byte) {
	result, flags := decOp(*reg)
	z.setOp8(4, 1, reg, result, flags)
}

func parityFlag(b byte) uint32 {
	return uint32(^(b>>7^b>>6^b>>5^b>>4^b>>3^b>>2^b>>1^b)&1) << 8
}

func (z *z80) daaOp() {

	newCarryFlag := uint32(0)
	newHFlag := uint32(0)
	if z.getSubFlag() {
		diff := byte(0)
		if z.A&0x0f > 0x09 || z.getHalfCarryFlag() {
			diff += 0x06
		}
		if z.A > 0x99 || z.getCarryFlag() {
			newCarryFlag = 1
			diff += 0x60
		}
		newHFlag = hFlagSub(z.A, diff)
		z.A -= diff
	} else {
		diff := byte(0)
		if z.A&0x0f > 0x09 || z.getHalfCarryFlag() {
			diff += 0x06
		}
		if z.A > 0x99 || z.getCarryFlag() {
			newCarryFlag = 1
			diff += 0x60
		}
		newHFlag = hFlagAdd(z.A, diff)
		z.A += diff
	}

	flags := signFlag(z.A)
	flags |= zFlag(z.A)
	flags |= undocFlags(z.A)
	flags |= newHFlag
	flags |= parityFlag(z.A)
	flags |= 0x20
	flags |= newCarryFlag

	z.setOpFn(4, 1, func() {}, flags)
}
func (z *z80) imeToString() string {
	if z.InterruptMasterEnable {
		return "1"
	}
	return "0"
}

func (z *z80) debugStatusLine() string {

	outStr := fmt.Sprintf("Step:%08d, ", z.Steps) +
		fmt.Sprintf("(*PC)[0:4]:%02x%02x%02x%02x, ", z.Read(z.PC), z.Read(z.PC+1), z.Read(z.PC+2), z.Read(z.PC+3)) +
		fmt.Sprintf("(*SP):%04x, ", z.read16(z.SP)) +
		fmt.Sprintf("[PC:%04x ", z.PC) +
		fmt.Sprintf("SP:%04x ", z.SP) +
		fmt.Sprintf("AF:%04x ", z.getAF()) +
		fmt.Sprintf("BC:%04x ", z.getBC()) +
		fmt.Sprintf("DE:%04x ", z.getDE()) +
		fmt.Sprintf("HL:%04x ", z.getHL()) +
		fmt.Sprintf("IME:%v ", z.imeToString())

	return outStr
}

func overflagAdd16(a, val, result uint16) uint32 {
	return uint32((^(a^val)&(a^result))&0x8000) >> 7
}
func overflagSub16(a, val, result uint16) uint32 {
	return uint32(((a^val)&(a^result))&0x8000) >> 7
}
func overflagAdd(a, val, result byte) uint32 {
	return uint32((^(a^val)&(a^result))&0x80) << 1
}
func overflagSub(a, val, result byte) uint32 {
	return uint32(((a^val)&(a^result))&0x80) << 1
}
func (z *z80) addOpA(cycles uint, instLen uint16, val byte) {
	result := z.A + val
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= hFlagAdd(z.A, val)
	flags |= overflagAdd(z.A, val, result)
	flags |= cFlagAdd(z.A, val)
	z.setOpA(cycles, instLen, z.A+val, flags)
}
func (z *z80) adcOpA(cycles uint, instLen uint16, val byte) {
	carry := z.F & 1
	result := z.A + val + carry
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= hFlagAdc(z.A, val, z.F)
	flags |= overflagAdd(z.A, val, result)
	flags |= cFlagAdc(z.A, val, z.F)
	z.setOpA(cycles, instLen, result, flags)
}
func (z *z80) adcOpHL(cycles uint, instLen uint16, val uint16) {
	hl := z.getHL()
	carry := uint16(z.F & 1)
	result := hl + val + carry
	flags := signFlag16(result)
	flags |= zFlag16(result)
	flags |= undocFlags(byte(result >> 8))
	flags |= hFlagAdc16(hl, val, z.F)
	flags |= overflagAdd16(hl, val, result)
	flags |= cFlagAdc16(hl, val, z.F)
	z.setOpHL(cycles, instLen, result, flags)
}
func (z *z80) subOpA(cycles uint, instLen uint16, val byte) {
	result := z.A - val
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= hFlagSub(z.A, val)
	flags |= overflagSub(z.A, val, result)
	flags |= 0x10
	flags |= cFlagSub(z.A, val)
	z.setOpA(cycles, instLen, result, flags)
}
func (z *z80) sbcOpA(cycles uint, instLen uint16, val byte) {
	carry := z.F & 1
	result := z.A - val - carry
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= hFlagSbc(z.A, val, z.F)
	flags |= overflagSub(z.A, val, result)
	flags |= 0x10
	flags |= cFlagSbc(z.A, val, z.F)
	z.setOpA(cycles, instLen, result, flags)
}
func (z *z80) sbcOpHL(cycles uint, instLen uint16, val uint16) {
	hl := z.getHL()
	carry := uint16(z.F & 1)
	result := hl - val - carry
	flags := signFlag16(result)
	flags |= zFlag16(result)
	flags |= undocFlags(byte(result >> 8))
	flags |= hFlagSbc16(hl, val, z.F)
	flags |= overflagSub16(hl, val, result)
	flags |= 0x10
	flags |= cFlagSbc16(hl, val, z.F)
	z.setOpHL(cycles, instLen, result, flags)
}
func (z *z80) negOpA(cycles uint, instLen uint16) {
	result := 0 - z.A
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= hFlagSub(0, z.A)
	flags |= overflagSub(0, z.A, result)
	flags |= 0x10
	flags |= cFlagSub(0, z.A)
	z.setOpA(cycles, instLen, result, flags)
}
func (z *z80) andOpA(cycles uint, instLen uint16, val byte) {
	result := z.A & val
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= 0x10000
	flags |= parityFlag(result)
	z.setOpA(cycles, instLen, result, flags)
}
func (z *z80) xorOpA(cycles uint, instLen uint16, val byte) {
	result := z.A ^ val
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= parityFlag(result)
	z.setOpA(cycles, instLen, result, flags)
}
func (z *z80) orOpA(cycles uint, instLen uint16, val byte) {
	result := z.A | val
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= parityFlag(result)
	z.setOpA(cycles, instLen, result, flags)
}
func (z *z80) cpOp(cycles uint, instLen uint16, val byte) {
	result := z.A - val
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= overflagSub(z.A, val, result)
	flags |= undocFlags(val)
	flags |= hFlagSub(z.A, val)
	flags |= 0x10
	flags |= cFlagSub(z.A, val)
	z.setOpFn(cycles, instLen, func() {}, flags)
}

func (z *z80) callOp(cycles uint, instLen, callAddr uint16) {
	z.pushOp16(cycles, instLen, z.PC+instLen)
	z.PC = callAddr
}

// NOTE: should be the relevant bits only
func (z *z80) getRegFromOpBits(opBits byte) *byte {
	switch opBits {
	case 0:
		return &z.B
	case 1:
		return &z.C
	case 2:
		return &z.D
	case 3:
		return &z.E
	case 4:
		return &z.H
	case 5:
		return &z.L
	case 6:
		return nil // (hl)
	case 7:
		return &z.A
	}
	panic("getRegFromOpBits: unknown bits passed")
}

func (z *z80) getCyclesAndValFromOpBits(cyclesReg uint, cyclesHL uint, opcode byte) (uint, byte) {
	if reg := z.getRegFromOpBits(opcode & 0x07); reg != nil {
		return cyclesReg, *reg
	}
	return cyclesHL, z.followHL()
}

func (z *z80) loadOp(cyclesReg uint, cyclesHL uint, instLen uint16,
	opcode byte, fnPtr func(uint, uint16, byte, uint32)) {

	cycles, val := z.getCyclesAndValFromOpBits(cyclesReg, cyclesHL, opcode)
	fnPtr(cycles, instLen, val, 0x22222222)
}

func (z *z80) aluOp(cyclesReg uint, cyclesHL uint, instLen uint16,
	opcode byte, fnPtr func(uint, uint16, byte)) {

	cycles, val := z.getCyclesAndValFromOpBits(cyclesReg, cyclesHL, opcode)
	fnPtr(cycles, instLen, val)
}

func (z *z80) stepSimpleOp(opcode byte) bool {
	switch opcode & 0xf8 {
	case 0x40: // ld b, R_OR_(HL)
		z.loadOp(4, 7, 1, opcode, z.setOpB)
	case 0x48: // ld c, R_OR_(HL)
		z.loadOp(4, 7, 1, opcode, z.setOpC)
	case 0x50: // ld d, R_OR_(HL)
		z.loadOp(4, 7, 1, opcode, z.setOpD)
	case 0x58: // ld e, R_OR_(HL)
		z.loadOp(4, 7, 1, opcode, z.setOpE)
	case 0x60: // ld h, R_OR_(HL)
		z.loadOp(4, 7, 1, opcode, z.setOpH)
	case 0x68: // ld l, R_OR_(HL)
		z.loadOp(4, 7, 1, opcode, z.setOpL)

	case 0x78: // ld a, R_OR_(HL)
		z.loadOp(4, 7, 1, opcode, z.setOpA)

	case 0x80: // add R_OR_(HL)
		z.aluOp(4, 7, 1, opcode, z.addOpA)
	case 0x88: // adc R_OR_(HL)
		z.aluOp(4, 7, 1, opcode, z.adcOpA)
	case 0x90: // sub R_OR_(HL)
		z.aluOp(4, 7, 1, opcode, z.subOpA)
	case 0x98: // sbc R_OR_(HL)
		z.aluOp(4, 7, 1, opcode, z.sbcOpA)
	case 0xa0: // and R_OR_(HL)
		z.aluOp(4, 7, 1, opcode, z.andOpA)
	case 0xa8: // xor R_OR_(HL)
		z.aluOp(4, 7, 1, opcode, z.xorOpA)
	case 0xb0: // or R_OR_(HL)
		z.aluOp(4, 7, 1, opcode, z.orOpA)
	case 0xb8: // cp R_OR_(HL)
		z.aluOp(4, 7, 1, opcode, z.cpOp)
	default:
		return false
	}
	return true
}

func xchg(a, b *byte) {
	tmp := *a
	*a = *b
	*b = tmp
}

func (z *z80) Step() {
	z.handleInterrupts()

	if z.IsHalted {
		z.RunCycles(4)
		return
	}

	z.StepOpcode()
}

func (z *z80) resumeFromHalt() {
	z.IsHalted = false
	z.PC++
}

func (z *z80) StepOpcode() {

	z.Steps++

	// NOTE: this is here to lag behind the IE
	// request by one instruction.
	if z.InterruptEnableNeedsDelay {
		z.InterruptEnableNeedsDelay = false
	}

	opcode := z.Read(z.PC)

	// simple cases
	if z.stepSimpleOp(opcode) {
		return
	}

	// complex cases
	switch opcode {

	case 0x00: // nop
		z.setOpFn(4, 1, func() {}, 0x22222222)
	case 0x01: // ld bc, nn
		z.setOpBC(10, 3, z.read16(z.PC+1), 0x22222222)
	case 0x02: // ld (bc), a
		z.setOpMem8(7, 1, z.getBC(), z.A, 0x22222222)
	case 0x03: // inc bc
		z.setOpBC(6, 1, z.getBC()+1, 0x22222222)
	case 0x04: // inc b
		z.incOpReg(&z.B)
	case 0x05: // dec b
		z.decOpReg(&z.B)
	case 0x06: // ld b, n8
		z.setOpB(7, 2, z.Read(z.PC+1), 0x22222222)
	case 0x07: // rlca
		z.rlcaOp()

	case 0x08: // ex af, af'
		z.setOpFn(4, 1, func() {
			xchg(&z.Ah, &z.A)
			xchg(&z.Fh, &z.F)
		}, 0x22222222)
	case 0x09: // add hl, bc
		v1, v2 := z.getHL(), z.getBC()
		result, flags := add16Op(v1, v2)
		z.setOpHL(11, 1, result, flags)
	case 0x0a: // ld a, (bc)
		z.setOpA(7, 1, z.followBC(), 0x22222222)
	case 0x0b: // dec bc
		z.setOpBC(6, 1, z.getBC()-1, 0x22222222)
	case 0x0c: // inc c
		z.incOpReg(&z.C)
	case 0x0d: // dec c
		z.decOpReg(&z.C)
	case 0x0e: // ld c, n8
		z.setOpC(7, 2, z.Read(z.PC+1), 0x22222222)
	case 0x0f: // rrca
		z.rrcaOp()

	case 0x10: // djnz r8
		z.B--
		z.jmpRel8(13, 8, 2, z.B != 0, int8(z.Read(z.PC+1)))
	case 0x11: // ld de, nn
		z.setOpDE(10, 3, z.read16(z.PC+1), 0x22222222)
	case 0x12: // ld (de), a
		z.setOpMem8(7, 1, z.getDE(), z.A, 0x22222222)
	case 0x13: // inc de
		z.setOpDE(6, 1, z.getDE()+1, 0x22222222)
	case 0x14: // inc d
		z.incOpReg(&z.D)
	case 0x15: // dec d
		z.decOpReg(&z.D)
	case 0x16: // ld d, n8
		z.setOpD(7, 2, z.Read(z.PC+1), 0x22222222)
	case 0x17: // rla
		z.rlaOp()

	case 0x18: // jr r8
		z.jmpRel8(12, 12, 2, true, int8(z.Read(z.PC+1)))
	case 0x19: // add hl, de
		v1, v2 := z.getHL(), z.getDE()
		result, flags := add16Op(v1, v2)
		z.setOpHL(11, 1, result, flags)
	case 0x1a: // ld a, (de)
		z.setOpA(7, 1, z.followDE(), 0x22222222)
	case 0x1b: // dec de
		z.setOpDE(6, 1, z.getDE()-1, 0x22222222)
	case 0x1c: // inc e
		z.incOpReg(&z.E)
	case 0x1d: // dec e
		z.decOpReg(&z.E)
	case 0x1e: // ld e, n8
		z.setOpE(7, 2, z.Read(z.PC+1), 0x22222222)
	case 0x1f: // rra
		z.rraOp()

	case 0x20: // jr nz, r8
		z.jmpRel8(12, 7, 2, !z.getZeroFlag(), int8(z.Read(z.PC+1)))
	case 0x21: // ld hl, nn
		z.setOpHL(10, 3, z.read16(z.PC+1), 0x22222222)
	case 0x22: // ld (nn), hl
		z.setOpMem16(16, 3, z.read16(z.PC+1), z.getHL(), 0x22222222)
	case 0x23: // inc hl
		z.setOpHL(6, 1, z.getHL()+1, 0x22222222)
	case 0x24: // inc h
		z.incOpReg(&z.H)
	case 0x25: // dec h
		z.decOpReg(&z.H)
	case 0x26: // ld h, d8
		z.setOpH(7, 2, z.Read(z.PC+1), 0x22222222)
	case 0x27: // daa
		z.daaOp()

	case 0x28: // jr z, r8
		z.jmpRel8(12, 7, 2, z.getZeroFlag(), int8(z.Read(z.PC+1)))
	case 0x29: // add hl, hl
		v1, v2 := z.getHL(), z.getHL()
		result, flags := add16Op(v1, v2)
		z.setOpHL(11, 1, result, flags)
	case 0x2a: // ld hl, (nn)
		addr := z.read16(z.PC + 1)
		z.setOpHL(16, 3, z.read16(addr), 0x22222222)
	case 0x2b: // dec hl
		z.setOpHL(6, 1, z.getHL()-1, 0x22222222)
	case 0x2c: // inc l
		z.incOpReg(&z.L)
	case 0x2d: // dec l
		z.decOpReg(&z.L)
	case 0x2e: // ld l, d8
		z.setOpL(7, 2, z.Read(z.PC+1), 0x22222222)
	case 0x2f: // cpl
		result := ^z.A
		z.setOpA(4, 1, result, 0x22010212|undocFlags(result))

	case 0x30: // jr z, r8
		z.jmpRel8(12, 7, 2, !z.getCarryFlag(), int8(z.Read(z.PC+1)))
	case 0x31: // ld sp, nn
		z.setOpSP(10, 3, z.read16(z.PC+1), 0x22222222)
	case 0x32: // ld (nn), a
		z.setOpMem8(13, 3, z.read16(z.PC+1), z.A, 0x22222222)
	case 0x33: // inc sp
		z.setOpSP(6, 1, z.SP+1, 0x22222222)
	case 0x34: // inc (hl)
		val := z.followHL()
		result := val + 1
		flags := signFlag(result)
		flags |= zFlag(result)
		flags |= undocFlags(result)
		flags |= hFlagAdd(val, 1)
		flags |= overflagAdd(val, 1, result)
		flags |= 0x02
		z.setOpMem8(11, 1, z.getHL(), result, flags)
	case 0x35: // dec (hl)
		val := z.followHL()
		result := val - 1
		flags := signFlag(result)
		flags |= zFlag(result)
		flags |= undocFlags(result)
		flags |= hFlagSub(val, 1)
		flags |= overflagSub(val, 1, result)
		flags |= 0x12
		z.setOpMem8(11, 1, z.getHL(), result, flags)
	case 0x36: // ld (hl), n8
		z.setOpMem8(10, 2, z.getHL(), z.Read(z.PC+1), 0x22222222)
	case 0x37: // scf
		z.setOpFn(4, 1, func() {}, 0x22000201|undocFlags(z.A))

	case 0x38: // jr c, r8
		z.jmpRel8(12, 7, 2, z.getCarryFlag(), int8(z.Read(z.PC+1)))
	case 0x39: // add hl, sp
		v1, v2 := z.getHL(), z.SP
		result, flags := add16Op(v1, v2)
		z.setOpHL(11, 1, result, flags)
	case 0x3a: // ld a, (nn)
		z.setOpA(13, 3, z.Read(z.read16(z.PC+1)), 0x22222222)
	case 0x3b: // dec sp
		z.setOpSP(6, 1, z.SP-1, 0x22222222)
	case 0x3c: // inc a
		z.incOpReg(&z.A)
	case 0x3d: // dec a
		z.decOpReg(&z.A)
	case 0x3e: // ld a, n8
		z.setOpA(7, 2, z.Read(z.PC+1), 0x22222222)
	case 0x3f: // ccf
		newH := uint32(z.F&1) << 16
		newC := uint32(z.F&1) ^ 1
		flags := 0x22000200 | undocFlags(z.A) | newH | newC
		z.setOpFn(4, 1, func() {}, flags)

	case 0x70: // ld (hl), b
		z.setOpMem8(7, 1, z.getHL(), z.B, 0x22222222)
	case 0x71: // ld (hl), c
		z.setOpMem8(7, 1, z.getHL(), z.C, 0x22222222)
	case 0x72: // ld (hl), d
		z.setOpMem8(7, 1, z.getHL(), z.D, 0x22222222)
	case 0x73: // ld (hl), e
		z.setOpMem8(7, 1, z.getHL(), z.E, 0x22222222)
	case 0x74: // ld (hl), h
		z.setOpMem8(7, 1, z.getHL(), z.H, 0x22222222)
	case 0x75: // ld (hl), l
		z.setOpMem8(7, 1, z.getHL(), z.L, 0x22222222)
	case 0x76: // halt
		z.setOpFn(4, 0, func() { z.IsHalted = true }, 0x22222222)
	case 0x77: // ld (hl), a
		z.setOpMem8(7, 1, z.getHL(), z.A, 0x22222222)

	case 0xc0: // ret nz
		z.jmpRet(11, 5, 1, !z.getZeroFlag())
	case 0xc1: // pop bc
		z.popOp16(10, 1, z.setBC)
	case 0xc2: // jp nz, nn
		z.jmpAbs16(10, 10, 3, !z.getZeroFlag(), z.read16(z.PC+1))
	case 0xc3: // jp nn
		z.setOpPC(10, 3, z.read16(z.PC+1), 0x22222222)
	case 0xc4: // call nz, nn
		z.jmpCall(17, 10, 3, !z.getZeroFlag(), z.read16(z.PC+1))
	case 0xc5: // push bc
		z.pushOp16(11, 1, z.getBC())
	case 0xc6: // add a, n8
		z.addOpA(7, 2, z.Read(z.PC+1))
	case 0xc7: // rst 00h
		z.callOp(11, 1, 0x0000)

	case 0xc8: // ret z
		z.jmpRet(11, 5, 1, z.getZeroFlag())
	case 0xc9: // ret
		z.popOp16(10, 1, z.setPC)
	case 0xca: // jp z, nn
		z.jmpAbs16(10, 10, 3, z.getZeroFlag(), z.read16(z.PC+1))
	case 0xcb: // extended opcode prefix
		z.stepCBPrefixOpcode()
	case 0xcc: // call z, nn
		z.jmpCall(17, 10, 3, z.getZeroFlag(), z.read16(z.PC+1))
	case 0xcd: // call nn
		z.callOp(17, 3, z.read16(z.PC+1))
	case 0xce: // adc a, n8
		z.adcOpA(7, 2, z.Read(z.PC+1))
	case 0xcf: // rst 08h
		z.callOp(11, 1, 0x0008)

	case 0xd0: // ret nc
		z.jmpRet(11, 5, 1, !z.getCarryFlag())
	case 0xd1: // pop de
		z.popOp16(10, 1, z.setDE)
	case 0xd2: // jp nc, nn
		z.jmpAbs16(10, 10, 3, !z.getCarryFlag(), z.read16(z.PC+1))
	case 0xd3: // out (n), A
		z.RunCycles(11)
		addr := uint16(z.A)<<8 | uint16(z.Read(z.PC+1))
		z.Out(addr, z.A)
		z.PC += 2
	case 0xd4: // call nc, nn
		z.jmpCall(17, 10, 3, !z.getCarryFlag(), z.read16(z.PC+1))
	case 0xd5: // push de
		z.pushOp16(11, 1, z.getDE())
	case 0xd6: // sub n8
		z.subOpA(7, 2, z.Read(z.PC+1))
	case 0xd7: // rst 10h
		z.callOp(11, 1, 0x0010)

	case 0xd8: // ret c
		z.jmpRet(11, 5, 1, z.getCarryFlag())
	case 0xd9: // exx
		z.setOpFn(4, 1, func() {
			xchg(&z.Bh, &z.B)
			xchg(&z.Ch, &z.C)
			xchg(&z.Dh, &z.D)
			xchg(&z.Eh, &z.E)
			xchg(&z.Hh, &z.H)
			xchg(&z.Lh, &z.L)
		}, 0x22222222)
	case 0xda: // jp c, nn
		z.jmpAbs16(10, 10, 3, z.getCarryFlag(), z.read16(z.PC+1))
	case 0xdb: // in a, (n)
		addr := uint16(z.A)<<8 | uint16(z.Read(z.PC+1))
		z.setOpA(11, 2, z.In(addr), 0x22222222)
	case 0xdc: // call c, nn
		z.jmpCall(17, 10, 3, z.getCarryFlag(), z.read16(z.PC+1))
	case 0xdd:
		z.stepIndexPrefixOpcode(&z.IX)
	case 0xde: // sbc n8
		z.sbcOpA(7, 2, z.Read(z.PC+1))
	case 0xdf: // rst 18h
		z.callOp(11, 1, 0x0018)

	case 0xe0: // ret po
		z.jmpRet(11, 5, 1, !z.getParityOverflowFlag())
	case 0xe1: // pop hl
		z.popOp16(10, 1, z.setHL)
	case 0xe2: // jp po, nn
		z.jmpAbs16(10, 10, 3, !z.getParityOverflowFlag(), z.read16(z.PC+1))
	case 0xe3: // ex (sp), hl
		z.setOpFn(19, 1, func() {
			tmp := z.read16(z.SP)
			z.write16(z.SP, z.getHL())
			z.setHL(tmp)
		}, 0x22222222)
	case 0xe4: // call po, nn
		z.jmpCall(17, 10, 3, !z.getParityOverflowFlag(), z.read16(z.PC+1))
	case 0xe5: // push hl
		z.pushOp16(11, 1, z.getHL())
	case 0xe6: // and n8
		z.andOpA(7, 2, z.Read(z.PC+1))
	case 0xe7: // rst 20h
		z.callOp(11, 1, 0x0020)

	case 0xe8: // ret pe
		z.jmpRet(11, 5, 1, z.getParityOverflowFlag())
	case 0xe9: // jp hl (also written jp (hl))
		z.setOpPC(4, 1, z.getHL(), 0x22222222)
	case 0xea: // jp pe, nn
		z.jmpAbs16(10, 10, 3, z.getParityOverflowFlag(), z.read16(z.PC+1))
	case 0xeb: // ex de, hl
		z.setOpFn(4, 1, func() {
			xchg(&z.D, &z.H)
			xchg(&z.E, &z.L)
		}, 0x22222222)
	case 0xec: // call pe, nn
		z.jmpCall(17, 10, 3, z.getParityOverflowFlag(), z.read16(z.PC+1))
	case 0xed:
		z.stepEDPrefixOpcode()
	case 0xee: // xor n8
		z.xorOpA(7, 2, z.Read(z.PC+1))
	case 0xef: // rst 28h
		z.callOp(11, 1, 0x0028)

	case 0xf0: // ret p
		z.jmpRet(11, 5, 1, !z.getSignFlag())
	case 0xf1: // pop af
		z.popOp16(10, 1, z.setAF)
	case 0xf2: // jp p, nn
		z.jmpAbs16(10, 10, 3, !z.getSignFlag(), z.read16(z.PC+1))
	case 0xf3: // di
		z.setOpFn(4, 1, func() {
			z.InterruptMasterEnable = false
			z.InterruptSettingPreNMI = false
		}, 0x22222222)
	case 0xf4: // call p, nn
		z.jmpCall(17, 10, 3, !z.getSignFlag(), z.read16(z.PC+1))
	case 0xf5: // push af
		z.pushOp16(11, 1, z.getAF())
	case 0xf6: // or n8
		z.orOpA(7, 2, z.Read(z.PC+1))
	case 0xf7: // rst 30h
		z.callOp(11, 1, 0x0030)

	case 0xf8: // ret m
		z.jmpRet(11, 5, 1, z.getSignFlag())
	case 0xf9: // ld sp, hl
		z.setOpSP(6, 1, z.getHL(), 0x22222222)
	case 0xfa: // jp m, nn
		z.jmpAbs16(10, 10, 3, z.getSignFlag(), z.read16(z.PC+1))
	case 0xfb: // ei
		z.setOpFn(4, 1, func() {
			z.InterruptEnableNeedsDelay = true
			z.InterruptMasterEnable = true
			z.InterruptSettingPreNMI = true
		}, 0x22222222)
	case 0xfc: // call m, nn
		z.jmpCall(17, 10, 3, z.getSignFlag(), z.read16(z.PC+1))
	case 0xfd:
		z.stepIndexPrefixOpcode(&z.IY)
	case 0xfe: // cp a, n8
		z.cpOp(7, 2, z.Read(z.PC+1))
	case 0xff: // rst 38h
		z.callOp(11, 1, 0x0038)

	default:
		z.Err(fmt.Errorf("Unknown Opcode: 0x%02x\r\n", opcode))
	}
}

func (z *z80) stepCBPrefixOpcode() {

	extOpcode := z.Read(z.PC + 1)

	switch extOpcode & 0xf8 {

	case 0x00: // rlc R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.rlcOp)
	case 0x08: // rrc R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.rrcOp)
	case 0x10: // rl R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.rlOp)
	case 0x18: // rr R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.rrOp)
	case 0x20: // sla R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.slaOp)
	case 0x28: // sra R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.sraOp)
	case 0x30: // sll R_OR_(HL) (UNDOCUMENTED)
		z.extSetOp(8, 15, 2, extOpcode, z.sllOp)
	case 0x38: // srl R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.srlOp)

	case 0x40: // bit 0, R_OR_(HL)
		z.stdBitOp(8, 12, 2, extOpcode, 0)
	case 0x48: // bit 1, R_OR_(HL)
		z.stdBitOp(8, 12, 2, extOpcode, 1)
	case 0x50: // bit 2, R_OR_(HL)
		z.stdBitOp(8, 12, 2, extOpcode, 2)
	case 0x58: // bit 3, R_OR_(HL)
		z.stdBitOp(8, 12, 2, extOpcode, 3)
	case 0x60: // bit 4, R_OR_(HL)
		z.stdBitOp(8, 12, 2, extOpcode, 4)
	case 0x68: // bit 5, R_OR_(HL)
		z.stdBitOp(8, 12, 2, extOpcode, 5)
	case 0x70: // bit 6, R_OR_(HL)
		z.stdBitOp(8, 12, 2, extOpcode, 6)
	case 0x78: // bit 7, R_OR_(HL)
		z.stdBitOp(8, 12, 2, extOpcode, 7)

	case 0x80: // res 0, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getResOp(0))
	case 0x88: // res 1, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getResOp(1))
	case 0x90: // res 2, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getResOp(2))
	case 0x98: // res 3, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getResOp(3))
	case 0xa0: // res 4, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getResOp(4))
	case 0xa8: // res 5, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getResOp(5))
	case 0xb0: // res 6, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getResOp(6))
	case 0xb8: // res 7, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getResOp(7))

	case 0xc0: // set 0, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getBitSetOp(0))
	case 0xc8: // set 1, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getBitSetOp(1))
	case 0xd0: // set 2, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getBitSetOp(2))
	case 0xd8: // set 3, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getBitSetOp(3))
	case 0xe0: // set 4, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getBitSetOp(4))
	case 0xe8: // set 5, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getBitSetOp(5))
	case 0xf0: // set 6, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getBitSetOp(6))
	case 0xf8: // set 7, R_OR_(HL)
		z.extSetOp(8, 15, 2, extOpcode, z.getBitSetOp(7))
	}
}

func (z *z80) extSetOp(cyclesReg uint, cyclesHL uint, instLen uint16, opcode byte,
	opFn func(val byte) (result byte, flags uint32)) {

	if reg := z.getRegFromOpBits(opcode & 0x07); reg != nil {
		result, flags := opFn(*reg)
		z.setOp8(cyclesReg, instLen, reg, result, flags)
	} else {
		result, flags := opFn(z.followHL())
		z.setOpMem8(cyclesHL, instLen, z.getHL(), result, flags)
	}
}

func (z *z80) indexExtSetOp(cycles uint, instLen uint16, opcode byte,
	indexReg *uint16, opFn func(val byte) (result byte, flags uint32)) {

	addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
	val := z.Read(addr)

	result, flags := opFn(val)

	if reg := z.getRegFromOpBits(opcode & 0x07); reg != nil {
		*reg = result
	}
	z.setOpMem8(cycles, instLen, addr, result, flags)
}

func (z *z80) indexBitOp(cycles uint, instLen uint16, indexReg *uint16, bitNum byte) {
	addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
	val := z.Read(addr)
	flags := bitOp(val, bitNum)
	flags &^= 0x00101000
	flags |= undocFlags(byte(addr >> 8))
	z.setOpFn(cycles, instLen, func() {}, flags)
}

func rotOpFlags(result, carry byte) uint32 {
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= parityFlag(result)
	flags |= uint32(carry)
	return flags
}
func modifyFlagsForRotAOp(flags uint32) uint32 {
	return 0x22000200 | flags&(1|1<<12|1<<20)
}

func (z *z80) rlaOp() {
	result, flags := z.rlOp(z.A)
	newFlags := modifyFlagsForRotAOp(flags)
	z.setOp8(4, 1, &z.A, result, newFlags)
}
func (z *z80) rlOp(val byte) (byte, uint32) {
	result, carry := (val<<1)|(z.F&1), (val >> 7)
	return result, rotOpFlags(result, carry)
}

func (z *z80) rraOp() {
	result, flags := z.rrOp(z.A)
	newFlags := modifyFlagsForRotAOp(flags)
	z.setOp8(4, 1, &z.A, result, newFlags)
}
func (z *z80) rrOp(val byte) (byte, uint32) {
	result, carry := (z.F<<7)|(val>>1), val&1
	return result, rotOpFlags(result, carry)
}

func (z *z80) rlcaOp() {
	result, flags := z.rlcOp(z.A)
	newFlags := modifyFlagsForRotAOp(flags)
	z.setOp8(4, 1, &z.A, result, newFlags)
}
func (z *z80) rlcOp(val byte) (byte, uint32) {
	result, carry := (val<<1)|(val>>7), val>>7
	return result, rotOpFlags(result, carry)
}

func (z *z80) rrcaOp() {
	result, flags := z.rrcOp(z.A)
	newFlags := modifyFlagsForRotAOp(flags)
	z.setOp8(4, 1, &z.A, result, newFlags)
}
func (z *z80) rrcOp(val byte) (byte, uint32) {
	result, carry := (val<<7)|(val>>1), val&1
	return result, rotOpFlags(result, carry)
}

func (z *z80) sllOp(val byte) (byte, uint32) {
	result, carry := (val<<1)|1, val>>7
	return result, rotOpFlags(result, carry)
}

func (z *z80) srlOp(val byte) (byte, uint32) {
	result, carry := val>>1, val&1
	return result, rotOpFlags(result, carry)
}

func (z *z80) slaOp(val byte) (byte, uint32) {
	result, carry := val<<1, val>>7
	return result, rotOpFlags(result, carry)
}

func (z *z80) sraOp(val byte) (byte, uint32) {
	result, carry := (val&0x80)|(val>>1), val&0x01
	return result, rotOpFlags(result, carry)
}

func bitOp(val byte, bitNum byte) uint32 {
	testVal := val & (1 << bitNum)
	flags := signFlag(testVal)
	flags |= zFlag(testVal)
	flags |= undocFlags(testVal)
	flags |= 0x00010000
	flags |= parityFlag(testVal)
	flags |= 0x02
	return flags
}

func (z *z80) stdBitOp(cyclesReg uint, cyclesHL uint, instLen uint16, opcode byte, bitNum byte) {
	cycles, val := z.getCyclesAndValFromOpBits(cyclesReg, cyclesHL, opcode)
	flags := bitOp(val, bitNum)
	z.setOpFn(cycles, instLen, func() {}, flags)
}

func (z *z80) getResOp(bitNum uint) func(byte) (byte, uint32) {
	return func(val byte) (byte, uint32) {
		result := val &^ (1 << bitNum)
		return result, 0x22222222
	}
}

func (z *z80) getBitSetOp(bitNum uint8) func(byte) (byte, uint32) {
	return func(val byte) (byte, uint32) {
		result := val | (1 << bitNum)
		return result, 0x22222222
	}
}

func (z *z80) runAndUpdatePC(numCycles uint, instLen uint16) {
	z.RunCycles(numCycles)
	z.PC += instLen
}

func getFlagsForInOp(result byte) uint32 {
	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= undocFlags(result)
	flags |= parityFlag(result)
	flags |= 0x02
	return flags
}

func (z *z80) stepEDPrefixOpcode() {
	extOpcode := z.Read(z.PC + 1)
	switch extOpcode {
	case 0x40: // in b, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		flags := getFlagsForInOp(result)
		z.setOpB(12, 2, result, flags)
	case 0x41: // out (c), b
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.B)
		z.runAndUpdatePC(12, 2)
	case 0x42: // sbc hl, bc
		z.sbcOpHL(15, 2, z.getBC())
	case 0x43: // ld (nn), bc
		z.setOpMem16(20, 4, z.read16(z.PC+2), z.getBC(), 0x22222222)
	case 0x44: // neg
		z.negOpA(8, 2)
	case 0x45: // retn
		z.InterruptMasterEnable = z.InterruptSettingPreNMI
		z.popOp16(14, 2, z.setPC)
	case 0x46: // im 0
		z.InterruptMode = 0
		z.runAndUpdatePC(8, 2)
	case 0x47: // ld i, a
		z.setOpFn(9, 2, func() { z.I = z.A }, 0x22222222)
	case 0x48: // in c, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		flags := getFlagsForInOp(result)
		z.setOpC(12, 2, result, flags)
	case 0x49: // out (C), C
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.C)
		z.runAndUpdatePC(12, 2)
	case 0x4a: // adc hl, bc
		z.adcOpHL(15, 2, z.getBC())
	case 0x4b: // ld bc, (nn)
		addr := z.read16(z.PC + 2)
		z.setOpBC(20, 4, z.read16(addr), 0x22222222)
	case 0x4c: // neg (UNDOCUMENTED)
		z.negOpA(8, 2)
	case 0x4d: // reti
		z.interruptComplete()
		z.popOp16(14, 2, z.setPC)
	case 0x4e: // im 0 (UNDOCUMENTED)
		z.InterruptMode = 0
		z.runAndUpdatePC(8, 2)
	case 0x4f: // ld r, a
		z.setOpFn(9, 2, func() { z.R = z.A }, 0x22222222)
	case 0x50: // in d, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		flags := getFlagsForInOp(result)
		z.setOpD(12, 2, result, flags)
	case 0x51: // out (c), d
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.D)
		z.runAndUpdatePC(12, 2)
	case 0x52: // sbc hl, de
		z.sbcOpHL(15, 2, z.getDE())
	case 0x53: // ld (nn), de
		z.setOpMem16(20, 4, z.read16(z.PC+2), z.getDE(), 0x22222222)
	case 0x54: // neg (UNDOCUMENTED)
		z.negOpA(8, 2)
	case 0x55: // retn (UNDOCUMENTED)
		z.InterruptMasterEnable = z.InterruptSettingPreNMI
		z.popOp16(14, 2, z.setPC)
	case 0x56: // im 1
		z.InterruptMode = 1
		z.runAndUpdatePC(8, 2)
	case 0x57: // ld a, i
		flags := signFlag(z.I)
		flags |= zFlag(z.I)
		flags |= uint32(boolBit(0, z.InterruptSettingPreNMI)) << 8
		flags |= 0x2
		z.setOpFn(9, 2, func() { z.A = z.I }, flags)
	case 0x58: // in e, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		flags := getFlagsForInOp(result)
		z.setOpE(12, 2, result, flags)
	case 0x59: // out (c), e
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.E)
		z.runAndUpdatePC(12, 2)
	case 0x5a: // adc hl, de
		z.adcOpHL(15, 2, z.getDE())
	case 0x5b: // ld de, (nn)
		addr := z.read16(z.PC + 2)
		z.setOpDE(20, 4, z.read16(addr), 0x22222222)
	case 0x5c: // neg (UNDOCUMENTED)
		z.negOpA(8, 2)
	case 0x5d: // retn (UNDOCUMENTED)
		z.InterruptMasterEnable = z.InterruptSettingPreNMI
		z.popOp16(14, 2, z.setPC)
	case 0x5e: // im 2
		z.InterruptMode = 2
		z.runAndUpdatePC(8, 2)
	case 0x5f: // ld a, r
		flags := signFlag(z.R)
		flags |= zFlag(z.R)
		flags |= uint32(boolBit(0, z.InterruptSettingPreNMI)) << 8
		flags |= 0x2
		z.setOpFn(9, 2, func() { z.A = z.R }, flags)
	case 0x60: // in h, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		flags := getFlagsForInOp(result)
		z.setOpH(12, 2, result, flags)
	case 0x61: // out (c), h
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.H)
		z.runAndUpdatePC(12, 2)
	case 0x62: // sbc hl, hl
		z.sbcOpHL(15, 2, z.getHL())
	case 0x63: // ld (nn), hl
		z.setOpMem16(20, 4, z.read16(z.PC+2), z.getHL(), 0x22222222)
	case 0x64: // neg (UNDOCUMENTED)
		z.negOpA(8, 2)
	case 0x65: // retn (UNDOCUMENTED)
		z.InterruptMasterEnable = z.InterruptSettingPreNMI
		z.popOp16(14, 2, z.setPC)
	case 0x66: // im 0 (UNDOCUMENTED)
		z.InterruptMode = 0
		z.runAndUpdatePC(8, 2)
	case 0x67: // rrd
		val := z.Read(z.getHL())
		outVal := (val >> 4) | (z.A << 4)
		outA := (z.A &^ 0x0f) | (val & 0x0f)
		z.Write(z.getHL(), outVal)
		flags := signFlag(outA)
		flags |= zFlag(outA)
		flags |= undocFlags(outA)
		flags |= parityFlag(outA)
		flags |= 0x2
		z.setOpA(18, 2, outA, flags)
	case 0x68: // in l, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		flags := getFlagsForInOp(result)
		z.setOpL(12, 2, result, flags)
	case 0x69: // out (c), l
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.L)
		z.runAndUpdatePC(12, 2)
	case 0x6a: // adc hl, hl
		z.adcOpHL(15, 2, z.getHL())
	case 0x6b: // ld hl, (nn)
		addr := z.read16(z.PC + 2)
		z.setOpHL(20, 4, z.read16(addr), 0x22222222)
	case 0x6c: // neg (UNDOCUMENTED)
		z.negOpA(8, 2)
	case 0x6d: // retn (UNDOCUMENTED)
		z.InterruptMasterEnable = z.InterruptSettingPreNMI
		z.popOp16(14, 2, z.setPC)
	case 0x6e: // im 0 (UNDOCUMENTED)
		z.InterruptMode = 0
		z.runAndUpdatePC(8, 2)
	case 0x6f: // rld
		val := z.Read(z.getHL())
		outVal := (val << 4) | (z.A & 0x0f)
		outA := (z.A &^ 0x0f) | (val >> 4)
		z.Write(z.getHL(), outVal)
		flags := signFlag(outA)
		flags |= zFlag(outA)
		flags |= undocFlags(outA)
		flags |= parityFlag(outA)
		flags |= 0x2
		z.setOpA(18, 2, outA, flags)
	case 0x70: // in f, (c) / in (c) (UNDOCUMENTED)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		z.setOpFn(12, 2, func() { z.F = result }, 0x22222222)
	case 0x71: // out (c), 0 (UNDOCUMENTED)
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, 0)
		z.runAndUpdatePC(12, 2)
	case 0x72: // sbc hl, sp
		z.sbcOpHL(15, 2, z.SP)
	case 0x73: // ld (nn), sp
		z.setOpMem16(20, 4, z.read16(z.PC+2), z.SP, 0x22222222)
	case 0x74: // neg (UNDOCUMENTED)
		z.negOpA(8, 2)
	case 0x75: // retn (UNDOCUMENTED)
		z.InterruptMasterEnable = z.InterruptSettingPreNMI
		z.popOp16(14, 2, z.setPC)
	case 0x76: // im 1 (UNDOCUMENTED)
		z.InterruptMode = 1
		z.runAndUpdatePC(8, 2)

	case 0x78: // in a, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		flags := getFlagsForInOp(result)
		z.setOpA(12, 2, result, flags)
	case 0x79: // out (c), a
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.A)
		z.runAndUpdatePC(12, 2)
	case 0x7a: // adc hl, sp
		z.adcOpHL(15, 2, z.SP)
	case 0x7b: // ld sp, (nn)
		addr := z.read16(z.PC + 2)
		z.setOpSP(20, 4, z.read16(addr), 0x22222222)
	case 0x7c: // neg (UNDOCUMENTED)
		z.negOpA(8, 2)
	case 0x7d: // retn (UNDOCUMENTED)
		z.InterruptMasterEnable = z.InterruptSettingPreNMI
		z.popOp16(14, 2, z.setPC)
	case 0x7e: // im 2 (UNDOCUMENTED)
		z.InterruptMode = 2
		z.runAndUpdatePC(8, 2)

	case 0xa0: // ldi
		z.setFlags(z.ldBlockOp(1))
		z.runAndUpdatePC(16, 2)
	case 0xa1: // cpi
		z.setFlags(z.cpBlockOp(1))
		z.runAndUpdatePC(16, 2)
	case 0xa2: // ini
		z.setFlags(z.inBlockOp(1))
		z.runAndUpdatePC(16, 2)
	case 0xa3: // outi
		z.setFlags(z.outBlockOp(1))
		z.runAndUpdatePC(16, 2)

	case 0xa8: // ldd
		z.setFlags(z.ldBlockOp(-1))
		z.runAndUpdatePC(16, 2)
	case 0xa9: // cpd
		z.setFlags(z.cpBlockOp(-1))
		z.runAndUpdatePC(16, 2)
	case 0xaa: // ind
		z.setFlags(z.inBlockOp(-1))
		z.runAndUpdatePC(16, 2)
	case 0xab: // outd
		z.setFlags(z.outBlockOp(-1))
		z.runAndUpdatePC(16, 2)

	case 0xb0: // ldir
		z.setFlags(z.ldBlockOp(1))
		if z.getBC() == 0 {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}
	case 0xb1: // cpir
		z.setFlags(z.cpBlockOp(1))
		if z.getBC() == 0 || z.getZeroFlag() {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}
	case 0xb2: // inir
		z.setFlags(z.inBlockOp(1))
		if z.B == 0 {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}
	case 0xb3: // otir
		z.setFlags(z.outBlockOp(1))
		if z.B == 0 {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}

	case 0xb8: // lddr
		z.setFlags(z.ldBlockOp(-1))
		if z.getBC() == 0 {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}
	case 0xb9: // cpdr
		z.setFlags(z.cpBlockOp(-1))
		if z.getBC() == 0 || z.getZeroFlag() {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}
	case 0xba: // indr
		z.setFlags(z.inBlockOp(-1))
		if z.B == 0 {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}
	case 0xbb: // otdr
		z.setFlags(z.outBlockOp(-1))
		if z.B == 0 {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}

	default:
		z.Err(fmt.Errorf("bad 0xed extended opcode: 0x%02x", extOpcode))
		// Alas, this is what the z80 does...
		//fmt.Printf("ignoring bad 0xed extended opcode: 0x%02x", extOpcode)
		z.setOpFn(4, 2, func() {}, 0x22222222)
	}
}

func (z *z80) outBlockOp(dir int16) uint32 {
	val := z.Read(z.getHL())

	newB, flags := decOp(z.B)
	z.B = newB

	addr := uint16(z.B)<<8 | uint16(z.C)
	z.Out(addr, val)

	z.setHL(z.getHL() + uint16(dir))

	return flags
}

func (z *z80) inBlockOp(dir int16) uint32 {
	addr := uint16(z.B)<<8 | uint16(z.C)
	result := z.In(addr)

	z.Write(z.getHL(), result)
	z.setHL(z.getHL() + uint16(dir))

	newB, flags := decOp(z.B)
	z.B = newB
	return flags
}

func (z *z80) ldBlockOp(dir int16) uint32 {
	srcAddr, dstAddr := z.getHL(), z.getDE()
	src := z.Read(srcAddr)
	z.Write(dstAddr, src)
	z.setHL(srcAddr + uint16(dir))
	z.setDE(dstAddr + uint16(dir))
	z.setBC(z.getBC() - 1)

	flags := uint32(boolBit(0, z.getBC() != 0)) << 8
	undocResult := src + z.A
	undocResult &^= 1 << 5
	undocResult |= (undocResult & 2) << 4
	flags |= undocFlags(undocResult)
	flags |= 0x22000002
	return flags
}

func (z *z80) cpBlockOp(dir int16) uint32 {
	srcAddr := z.getHL()
	src := z.Read(srcAddr)
	z.setHL(srcAddr + uint16(dir))
	z.setBC(z.getBC() - 1)

	result := z.A - src

	hFlag := hFlagSub(z.A, src)

	flags := signFlag(result)
	flags |= zFlag(result)
	flags |= hFlag
	flags |= uint32(boolBit(0, z.getBC() != 0)) << 8
	flags |= 0x12

	undocResult := z.A - src - boolBit(0, hFlag != 0)
	undocResult &^= 1 << 5
	undocResult |= (undocResult & 2) << 4
	flags |= undocFlags(undocResult)

	return flags
}

func add16Op(v1, v2 uint16) (uint16, uint32) {
	result := v1 + v2
	flags := uint32(0x22000200)
	flags |= undocFlags(byte(result >> 8))
	flags |= hFlagAdd16(v1, v2)
	flags |= cFlagAdd16(v1, v2)
	return result, flags
}

func (z *z80) stepIndexPrefixOpcode(indexReg *uint16) {
	extOpcode := z.Read(z.PC + 1)
	switch extOpcode {

	case 0x09: // add ix/iy, bc
		v1, v2 := *indexReg, z.getBC()
		result, flags := add16Op(v1, v2)
		z.setOpFn(15, 2, func() { *indexReg = result }, flags)
	case 0x19: // add ix/iy, de
		v1, v2 := *indexReg, z.getDE()
		result, flags := add16Op(v1, v2)
		z.setOpFn(15, 2, func() { *indexReg = result }, flags)
	case 0x21: // ld ix/iy, nn
		val := z.read16(z.PC + 2)
		z.setOpFn(14, 4, func() { *indexReg = val }, 0x22222222)
	case 0x22: // ld (nn), ix/iy
		addr := z.read16(z.PC + 2)
		z.setOpMem16(20, 4, addr, *indexReg, 0x22222222)
	case 0x23: // inc ix/iy
		z.setOpFn(10, 2, func() { *indexReg++ }, 0x22222222)
	case 0x24: // inc ixh/iyh
		hi := byte((*indexReg) >> 8)
		result, flags := incOp(hi)
		*indexReg &^= 0xff00
		*indexReg |= uint16(result) << 8
		z.setOpFn(8, 2, func() {}, flags)
	case 0x25: // dec ixh/iyh
		hi := byte((*indexReg) >> 8)
		result, flags := decOp(hi)
		*indexReg &^= 0xff00
		*indexReg |= uint16(result) << 8
		z.setOpFn(8, 2, func() {}, flags)
	case 0x26: // ld ixh/iyh, n
		val := z.Read(z.PC + 2)
		*indexReg &^= 0xff00
		*indexReg |= uint16(val) << 8
		z.setOpFn(11, 3, func() {}, 0x22222222)
	case 0x29: // add ix/iy, ix/iy
		v1, v2 := *indexReg, *indexReg
		result, flags := add16Op(v1, v2)
		z.setOpFn(15, 2, func() { *indexReg = result }, flags)
	case 0x2a: // ld ix/iy, (nn)
		addr := z.read16(z.PC + 2)
		val := z.read16(addr)
		z.setOpFn(20, 4, func() { *indexReg = val }, 0x22222222)
	case 0x2b: // dec ix/iy
		z.setOpFn(10, 2, func() { *indexReg-- }, 0x22222222)
	case 0x2c: // inc ixl/iyl
		lo := byte(*indexReg)
		result, flags := incOp(lo)
		*indexReg &^= 0x00ff
		*indexReg |= uint16(result)
		z.setOpFn(8, 2, func() {}, flags)
	case 0x2d: // dec ixl/iyl
		lo := byte(*indexReg)
		result, flags := decOp(lo)
		*indexReg &^= 0x00ff
		*indexReg |= uint16(result)
		z.setOpFn(8, 2, func() {}, flags)
	case 0x2e: // ld ixl/iyl, n
		val := z.Read(z.PC + 2)
		*indexReg &^= 0x00ff
		*indexReg |= uint16(val)
		z.setOpFn(11, 3, func() {}, 0x22222222)
	case 0x34: // inc (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		val := z.Read(addr)
		result, flags := incOp(val)
		z.setOpMem8(23, 3, addr, result, flags)
	case 0x35: // dec (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		val := z.Read(addr)
		result, flags := decOp(val)
		z.setOpMem8(23, 3, addr, result, flags)
	case 0x36: // ld (ix/iy + d), n
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		val := z.Read(z.PC + 3)
		z.setOpMem8(19, 4, addr, val, 0x22222222)
	case 0x39: // add ix/iy, sp
		v1, v2 := *indexReg, z.SP
		result, flags := add16Op(v1, v2)
		z.setOpFn(15, 2, func() { *indexReg = result }, flags)
	case 0x44: // ld b, ixh/iyh
		hi := byte((*indexReg) >> 8)
		z.setOpB(8, 2, hi, 0x22222222)
	case 0x45: // ld b, ixl/iyl
		lo := byte(*indexReg)
		z.setOpB(8, 2, lo, 0x22222222)
	case 0x46: // ld b, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpB(19, 3, z.Read(addr), 0x22222222)
	case 0x4c: // ld c, ixh/iyh
		hi := byte((*indexReg) >> 8)
		z.setOpC(8, 2, hi, 0x22222222)
	case 0x4d: // ld c, ixl/iyl
		lo := byte(*indexReg)
		z.setOpC(8, 2, lo, 0x22222222)
	case 0x4e: // ld c, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpC(19, 3, z.Read(addr), 0x22222222)
	case 0x54: // ld d, ixh/iyh
		hi := byte((*indexReg) >> 8)
		z.setOpD(8, 2, hi, 0x22222222)
	case 0x55: // ld d, ixl/iyl
		lo := byte(*indexReg)
		z.setOpD(8, 2, lo, 0x22222222)
	case 0x56: // ld d, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpD(19, 3, z.Read(addr), 0x22222222)
	case 0x5c: // ld e, ixh/iyh
		hi := byte((*indexReg) >> 8)
		z.setOpE(8, 2, hi, 0x22222222)
	case 0x5d: // ld e, ixl/iyl
		lo := byte(*indexReg)
		z.setOpE(8, 2, lo, 0x22222222)
	case 0x5e: // ld e, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpE(19, 3, z.Read(addr), 0x22222222)
	case 0x60: // ld ixh/iyh, b
		*indexReg &^= 0xff00
		*indexReg |= uint16(z.B) << 8
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x61: // ld ixh/iyh, c
		*indexReg &^= 0xff00
		*indexReg |= uint16(z.C) << 8
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x62: // ld ixh/iyh, d
		*indexReg &^= 0xff00
		*indexReg |= uint16(z.D) << 8
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x63: // ld ixh/iyh, e
		*indexReg &^= 0xff00
		*indexReg |= uint16(z.E) << 8
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x64: // ld ixh/iyh, ixh/iyh
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x65: // ld ixh/iyh, ixl/ixl
		loShifted := *indexReg << 8
		*indexReg &^= 0xff00
		*indexReg |= loShifted
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x66: // ld h, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpH(19, 3, z.Read(addr), 0x22222222)
	case 0x67: // ld ixh/iyh, a
		*indexReg &^= 0xff00
		*indexReg |= uint16(z.A) << 8
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x68: // ld ixl/iyl, b
		*indexReg &^= 0x00ff
		*indexReg |= uint16(z.B)
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x69: // ld ixl/iyl, c
		*indexReg &^= 0x00ff
		*indexReg |= uint16(z.C)
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x6a: // ld ixl/iyl, d
		*indexReg &^= 0x00ff
		*indexReg |= uint16(z.D)
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x6b: // ld ixl/iyl, e
		*indexReg &^= 0x00ff
		*indexReg |= uint16(z.E)
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x6c: // ld ixl/iyl, ixh/iyh
		hiShifted := *indexReg >> 8
		*indexReg &^= 0x00ff
		*indexReg |= hiShifted
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x6d: // ld ixl/iyl, ixl/iyl
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x6e: // ld l, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpL(19, 3, z.Read(addr), 0x22222222)
	case 0x6f: // ld ixl/iyl, a
		*indexReg &^= 0x00ff
		*indexReg |= uint16(z.A)
		z.setOpFn(8, 2, func() {}, 0x22222222)
	case 0x70: // ld (ix/iy + d), b
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpMem8(19, 3, addr, z.B, 0x22222222)
	case 0x71: // ld (ix/iy + d), c
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpMem8(19, 3, addr, z.C, 0x22222222)
	case 0x72: // ld (ix/iy + d), d
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpMem8(19, 3, addr, z.D, 0x22222222)
	case 0x73: // ld (ix/iy + d), e
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpMem8(19, 3, addr, z.E, 0x22222222)
	case 0x74: // ld (ix/iy + d), h
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpMem8(19, 3, addr, z.H, 0x22222222)
	case 0x75: // ld (ix/iy + d), l
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpMem8(19, 3, addr, z.L, 0x22222222)

	case 0x77: // ld (ix/iy + d), a
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpMem8(19, 3, addr, z.A, 0x22222222)

	case 0x7c: // ld a, ixh/iyh
		hi := byte((*indexReg) >> 8)
		z.setOpA(8, 2, hi, 0x22222222)
	case 0x7d: // ld d, ixl/iyl
		lo := byte(*indexReg)
		z.setOpA(8, 2, lo, 0x22222222)
	case 0x7e: // ld a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.setOpA(19, 3, z.Read(addr), 0x22222222)

	case 0x84: // add a, ixh
		hi := byte((*indexReg) >> 8)
		z.addOpA(8, 2, hi)
	case 0x85: // add a, ixh
		lo := byte(*indexReg)
		z.addOpA(8, 2, lo)
	case 0x86: // add a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.addOpA(19, 3, z.Read(addr))

	case 0x8c: // adc a, ixh
		hi := byte((*indexReg) >> 8)
		z.adcOpA(8, 2, hi)
	case 0x8d: // adc a, ixl
		lo := byte(*indexReg)
		z.adcOpA(8, 2, lo)
	case 0x8e: // adc a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.adcOpA(19, 3, z.Read(addr))

	case 0x94: // sub a, ixh
		hi := byte((*indexReg) >> 8)
		z.subOpA(8, 2, hi)
	case 0x95: // sub a, ixh
		lo := byte(*indexReg)
		z.subOpA(8, 2, lo)
	case 0x96: // sub a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.subOpA(19, 3, z.Read(addr))

	case 0x9c: // sbc a, ixh
		hi := byte((*indexReg) >> 8)
		z.sbcOpA(8, 2, hi)
	case 0x9d: // sbc a, ixl
		lo := byte(*indexReg)
		z.sbcOpA(8, 2, lo)
	case 0x9e: // sbc a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.sbcOpA(19, 3, z.Read(addr))

	case 0xa4: // and a, ixh
		hi := byte((*indexReg) >> 8)
		z.andOpA(8, 2, hi)
	case 0xa5: // and a, ixh
		lo := byte(*indexReg)
		z.andOpA(8, 2, lo)
	case 0xa6: // and a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.andOpA(19, 3, z.Read(addr))

	case 0xac: // xor a, ixh
		hi := byte((*indexReg) >> 8)
		z.xorOpA(8, 2, hi)
	case 0xad: // xor a, ixl
		lo := byte(*indexReg)
		z.xorOpA(8, 2, lo)
	case 0xae: // xor a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.xorOpA(19, 3, z.Read(addr))

	case 0xb4: // or a, ixh
		hi := byte((*indexReg) >> 8)
		z.orOpA(8, 2, hi)
	case 0xb5: // or a, ixh
		lo := byte(*indexReg)
		z.orOpA(8, 2, lo)
	case 0xb6: // or a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.orOpA(19, 3, z.Read(addr))

	case 0xbc: // cp a, ixh
		hi := byte((*indexReg) >> 8)
		z.cpOp(8, 2, hi)
	case 0xbd: // cp a, ixl
		lo := byte(*indexReg)
		z.cpOp(8, 2, lo)
	case 0xbe: // cp a, (ix/iy + d)
		addr := getDisplacedAddr(*indexReg, z.Read(z.PC+2))
		z.cpOp(19, 3, z.Read(addr))

	case 0xcb:
		z.stepBitsIndexOpcode(indexReg)

	case 0xe1: // pop ix/iy
		z.popOp16(14, 2, func(val uint16) { *indexReg = val })
	case 0xe3: // ex (sp), ix/iy
		z.setOpFn(23, 2, func() {
			tmp := z.read16(z.SP)
			z.write16(z.SP, *indexReg)
			*indexReg = tmp
		}, 0x22222222)
	case 0xe5: // push ix/iy
		z.pushOp16(15, 2, *indexReg)
	case 0xe9: // jp ix/iy (also written jp (ix/iy))
		z.setOpPC(8, 2, *indexReg, 0x22222222)
	case 0xf9: // ld sp, ix/iy
		z.setOpSP(10, 2, *indexReg, 0x22222222)

	default:
		// any other opcode or prefix just gets loaded from scratch
		z.runAndUpdatePC(4, 1)
	}
}

func (z *z80) stepBitsIndexOpcode(indexReg *uint16) {

	extOpcode := z.Read(z.PC + 3) // the disp is +2

	switch extOpcode & 0xf8 {

	case 0x00: // rlc (ix/iy + d)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.rlcOp)
	case 0x08: // rrc (ix/iy + d)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.rrcOp)
	case 0x10: // rl (ix/iy + d)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.rlOp)
	case 0x18: // rr (ix/iy + d)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.rrOp)
	case 0x20: // sla (ix/iy + d)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.slaOp)
	case 0x28: // sra (ix/iy + d)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.sraOp)
	case 0x30: // sll (ix/iy + d)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.sllOp)
	case 0x38: // srl (ix/iy + d)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.srlOp)

	case 0x40: // bit 0, (ix/iy + d)
		z.indexBitOp(20, 4, indexReg, 0)
	case 0x48: // bit 1, (ix/iy + d)
		z.indexBitOp(20, 4, indexReg, 1)
	case 0x50: // bit 2, (ix/iy + d)
		z.indexBitOp(20, 4, indexReg, 2)
	case 0x58: // bit 3, (ix/iy + d)
		z.indexBitOp(20, 4, indexReg, 3)
	case 0x60: // bit 4, (ix/iy + d)
		z.indexBitOp(20, 4, indexReg, 4)
	case 0x68: // bit 5, (ix/iy + d)
		z.indexBitOp(20, 4, indexReg, 5)
	case 0x70: // bit 6, (ix/iy + d)
		z.indexBitOp(20, 4, indexReg, 6)
	case 0x78: // bit 7, (ix/iy + d)
		z.indexBitOp(20, 4, indexReg, 7)

	case 0x80: // res 0, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getResOp(0))
	case 0x88: // res 1, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getResOp(1))
	case 0x90: // res 2, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getResOp(2))
	case 0x98: // res 3, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getResOp(3))
	case 0xa0: // res 4, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getResOp(4))
	case 0xa8: // res 5, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getResOp(5))
	case 0xb0: // res 6, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getResOp(6))
	case 0xb8: // res 7, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getResOp(7))

	case 0xc0: // set 0, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getBitSetOp(0))
	case 0xc8: // set 1, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getBitSetOp(1))
	case 0xd0: // set 2, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getBitSetOp(2))
	case 0xd8: // set 3, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getBitSetOp(3))
	case 0xe0: // set 4, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getBitSetOp(4))
	case 0xe8: // set 5, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getBitSetOp(5))
	case 0xf0: // set 6, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getBitSetOp(6))
	case 0xf8: // set 7, R_OR_(HL)
		z.indexExtSetOp(23, 4, extOpcode, indexReg, z.getBitSetOp(7))

	default:
		z.Err(fmt.Errorf("bits index prefix opcode not yet implemented 0x%02x", extOpcode))
	}
}

func getDisplacedAddr(base uint16, disp byte) uint16 {
	relAddr := int8(disp)
	return uint16(int(base) + int(relAddr))
}

func (z *z80) Err(msg error) {
	fmt.Println()
	fmt.Println("z80.Err():", msg)
	fmt.Println(z.debugStatusLine())
	fmt.Println()
	panic("z80.Err")
}
