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
		z.setOpFn(cyclesNotTaken, instLen, func() {}, 0x222222)
	}
}
func (z *z80) jmpRet(cyclesTaken uint, cyclesNotTaken uint, instLen uint16, test bool) {
	if test {
		z.popOp16(cyclesTaken, instLen, z.setPC)
	} else {
		z.setOpFn(cyclesNotTaken, instLen, func() {}, 0x222222)
	}
}

func (z *z80) followBC() byte { return z.Read(z.getBC()) }
func (z *z80) followDE() byte { return z.Read(z.getDE()) }
func (z *z80) followHL() byte { return z.Read(z.getHL()) }
func (z *z80) followSP() byte { return z.Read(z.SP) }
func (z *z80) followPC() byte { return z.Read(z.PC) }

// reminder: flags == zero, addsub, halfcarry, carry
// set all: 0x111111
// clear all: 0x000000
// ignore all: 0x222222

func zFlag(val uint8) uint32 {
	if val == 0 {
		return 0x010000
	}
	return 0
}

func signFlag(val uint8) uint32 {
	if val&0x80 > 0 {
		return 0x100000
	}
	return 0
}

// half carry
func hFlagAdd(val, addend uint8) uint32 {
	// 4th to 5th bit carry
	if int(val&0x0f)+int(addend&0x0f) >= 0x10 {
		return 0x001000
	}
	return 0
}

// half carry
func hFlagAdc(val, addend, fReg uint8) uint32 {
	carry := fReg & 0x01
	// 4th to 5th bit carry
	if int(carry)+int(val&0x0f)+int(addend&0x0f) >= 0x10 {
		return 0x001000
	}
	return 0
}

// half carry 16
func hFlagAdd16(val, addend uint16) uint32 {
	// 12th to 13th bit carry
	if int(val&0x0fff)+int(addend&0x0fff) >= 0x1000 {
		return 0x001000
	}
	return 0
}

// half carry
func hFlagSub(val, subtrahend uint8) uint32 {
	if int(val&0xf)-int(subtrahend&0xf) < 0 {
		return 0x001000
	}
	return 0
}

// half carry
func hFlagSbc(val, subtrahend, fReg uint8) uint32 {
	carry := fReg & 0x01
	if int(val&0xf)-int(subtrahend&0xf)-int(carry) < 0 {
		return 0x001000
	}
	return 0
}

// carry
func cFlagAdd(val, addend uint8) uint32 {
	if int(val)+int(addend) > 0xff {
		return 0x1
	}
	return 0x0
}

// carry
func cFlagAdc(val, addend, fReg uint8) uint32 {
	carry := fReg & 0x01
	if int(carry)+int(val)+int(addend) > 0xff {
		return 0x1
	}
	return 0x0
}

// carry 16
func cFlagAdd16(val, addend uint16) uint32 {
	if int(val)+int(addend) > 0xffff {
		return 0x1
	}
	return 0x0
}

// carry
func cFlagSub(val, subtrahend uint8) uint32 {
	if int(val)-int(subtrahend) < 0 {
		return 0x1
	}
	return 0x0
}
func cFlagSbc(val, subtrahend, fReg uint8) uint32 {
	carry := (fReg >> 4) & 0x01
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
	z.setOpMem16(cycles, instLen, z.SP-2, val, 0x222222)
	z.SP -= 2
}
func (z *z80) popOp16(cycles uint, instLen uint16, setFn func(val uint16)) {
	z.setOpFn(cycles, instLen, func() { setFn(z.read16(z.SP)) }, 0x222222)
	z.SP += 2
}

func (z *z80) incOpReg(reg *byte) {
	val := *reg
	result := val + 1
	hFlag := hFlagAdd(val, 1)
	overFlag := overflagAdd(val, 1, result)
	z.setOp8(4, 1, reg, result, signFlag(result)|zFlag(result)|hFlag|overFlag|0x02)
}

func (z *z80) decOpReg(reg *byte) {
	val := *reg
	result := val - 1
	hFlag := hFlagSub(val, 1)
	overFlag := overflagSub(val, 1, result)
	z.setOp8(4, 1, reg, result, signFlag(result)|zFlag(result)|hFlag|overFlag|0x12)
}

func parityFlag(b byte) uint32 {
	return uint32(^(b>>7^b>>6^b>>5^b>>4^b>>3^b>>2^b>>1^b)&1) << 8
}

func (z *z80) daaOp() {

	newCarryFlag := uint32(0)
	diff := byte(0)
	if z.getSubFlag() {
		if z.getHalfCarryFlag() {
			diff -= 0x06
		}
		if z.getCarryFlag() {
			newCarryFlag = 0x000001
			diff -= 0x60
		}
	} else {
		if z.A&0x0f > 0x09 || z.getHalfCarryFlag() {
			diff += 0x06
		}
		if z.A > 0x99 || z.getCarryFlag() {
			newCarryFlag = 0x000001
			diff += 0x60
		}
	}
	hFlag := hFlagAdd(z.A, diff) // NOTE: is hFlag really affected? Docs are unclear :/
	z.A += diff
	z.setOpFn(4, 1, func() {}, signFlag(z.A)|zFlag(z.A)|hFlag|parityFlag(z.A)|0x20|newCarryFlag)
}
func (z *z80) imeToString() string {
	if z.InterruptMasterEnable {
		return "1"
	}
	return "0"
}

var clearedBytes = []byte{
	0xc3, 0x2a, 0xf9, 0x11, 0x0e, 0xcd, 0xf5,
	0xc5, 0xd5, 0xe5, 0xc9, 0xe1, 0xc1, 0xf1,
	0xd1, 0x21, 0x7e, 0x23, 0xb6, 0xca, 0x2b,
	0x66, 0x6f,
}

func isGoodOpcode(opcode byte) bool {
	for _, b := range clearedBytes {
		if b == opcode {
			return true
		}
	}
	return false
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

	if isGoodOpcode(z.Read(z.PC)) {
		return "* " + outStr
	}
	return outStr
}

func overflagAdd(a, val, result byte) uint32 {
	return uint32((^(a^val)&(a^result))&0x80) << 1
}
func overflagSub(a, val, result byte) uint32 {
	return uint32(((a^val)&(a^result))&0x80) << 1
}
func (z *z80) addOpA(cycles uint, instLen uint16, val byte) {
	result := z.A + val
	hFlag := hFlagAdd(z.A, val)
	cFlag := cFlagAdd(z.A, val)
	overFlag := overflagAdd(z.A, val, result)
	z.setOpA(cycles, instLen, z.A+val, signFlag(result)|zFlag(result)|hFlag|overFlag|cFlag)
}
func (z *z80) adcOpA(cycles uint, instLen uint16, val byte) {
	carry := z.F & 1
	result := z.A + val + carry
	hFlag := hFlagAdc(z.A, val, z.F)
	cFlag := cFlagAdc(z.A, val, z.F)
	overFlag := overflagAdd(z.A, val, result)
	z.setOpA(cycles, instLen, result, signFlag(result)|zFlag(result)|hFlag|overFlag|cFlag)
}
func (z *z80) subOpA(cycles uint, instLen uint16, val byte) {
	result := z.A - val
	sFlag := signFlag(result)
	hFlag := hFlagSub(z.A, val)
	cFlag := cFlagSub(z.A, val)
	z.setOpA(cycles, instLen, result, sFlag|zFlag(result)|hFlag|0x000010|cFlag)
}
func (z *z80) sbcOpA(cycles uint, instLen uint16, val byte) {
	carry := z.F & 1
	result := z.A - val - carry
	hFlag := hFlagSbc(z.A, val, z.F)
	cFlag := cFlagSbc(z.A, val, z.F)
	z.setOpA(cycles, instLen, result, signFlag(result)|zFlag(result)|hFlag|0x000010|cFlag)
}
func (z *z80) andOpA(cycles uint, instLen uint16, val byte) {
	result := z.A & val
	overFlag := overflagAdd(z.A, val, result) // NOTE: is this right?
	z.setOpA(cycles, instLen, result, signFlag(result)|zFlag(result)|0x1000|overFlag)
}
func (z *z80) xorOpA(cycles uint, instLen uint16, val byte) {
	result := z.A ^ val
	z.setOpA(cycles, instLen, result, signFlag(result)|zFlag(result)|parityFlag(result))
}
func (z *z80) orOpA(cycles uint, instLen uint16, val byte) {
	result := z.A | val
	overFlag := overflagAdd(z.A, val, result) // NOTE: is this right?
	z.setOpA(cycles, instLen, result, signFlag(result)|zFlag(result)|overFlag)
}
func (z *z80) cpOp(cycles uint, instLen uint16, val byte) {
	result := z.A - val
	overFlag := overflagSub(z.A, val, result) // NOTE: is this right?
	hFlag := hFlagSub(z.A, val)
	cFlag := cFlagSub(z.A, val)
	z.setOpFn(cycles, instLen, func() {}, signFlag(result)|zFlag(result)|hFlag|overFlag|0x10|cFlag)
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
	fnPtr(cycles, instLen, val, 0x222222)
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

func (z *z80) addOpHL(v1, v2 uint16) {
	z.setOpHL(11, 1, v1+v2, (0x220200 | hFlagAdd16(v1, v2) | cFlagAdd16(v1, v2)))
}

func (z *z80) StepOpcode() {

	z.Steps++

	// TODO: this is from gameboy, does the real z80 do this?
	// this is here to lag behind the request by
	// one instruction.
	if z.MasterEnableRequested {
		z.MasterEnableRequested = false
		z.InterruptMasterEnable = true
	}

	opcode := z.Read(z.PC)

	// simple cases
	if z.stepSimpleOp(opcode) {
		return
	}

	// complex cases
	switch opcode {

	case 0x00: // nop
		z.setOpFn(4, 1, func() {}, 0x222222)
	case 0x01: // ld bc, nn
		z.setOpBC(10, 3, z.read16(z.PC+1), 0x222222)
	case 0x02: // ld (bc), a
		z.setOpMem8(7, 1, z.getBC(), z.A, 0x222222)
	case 0x03: // inc bc
		z.setOpBC(6, 1, z.getBC()+1, 0x222222)
	case 0x04: // inc b
		z.incOpReg(&z.B)
	case 0x05: // dec b
		z.decOpReg(&z.B)
	case 0x06: // ld b, n8
		z.setOpB(7, 2, z.Read(z.PC+1), 0x222222)
	case 0x07: // rlca
		z.rlcaOp()

	case 0x08: // ex af, af'
		z.setOpFn(4, 1, func() {
			xchg(&z.Ah, &z.A)
			xchg(&z.Fh, &z.F)
		}, 0x222222)
	case 0x09: // add hl, bc
		v1, v2 := z.getHL(), z.getBC()
		z.addOpHL(v1, v2)
	case 0x0a: // ld a, (bc)
		z.setOpA(7, 1, z.followBC(), 0x222222)
	case 0x0b: // dec bc
		z.setOpBC(6, 1, z.getBC()-1, 0x222222)
	case 0x0c: // inc c
		z.incOpReg(&z.C)
	case 0x0d: // dec c
		z.decOpReg(&z.C)
	case 0x0e: // ld c, n8
		z.setOpC(7, 2, z.Read(z.PC+1), 0x222222)
	case 0x0f: // rrca
		z.rrcaOp()

	case 0x10: // djnz r8
		z.B--
		z.jmpRel8(13, 8, 2, z.B != 0, int8(z.Read(z.PC+1)))
	case 0x11: // ld de, nn
		z.setOpDE(10, 3, z.read16(z.PC+1), 0x222222)
	case 0x12: // ld (de), a
		z.setOpMem8(7, 1, z.getDE(), z.A, 0x222222)
	case 0x13: // inc de
		z.setOpDE(6, 1, z.getDE()+1, 0x222222)
	case 0x14: // inc d
		z.incOpReg(&z.D)
	case 0x15: // dec d
		z.decOpReg(&z.D)
	case 0x16: // ld d, n8
		z.setOpD(7, 2, z.Read(z.PC+1), 0x222222)
	case 0x17: // rla
		z.rlaOp()

	case 0x18: // jr r8
		z.jmpRel8(12, 12, 2, true, int8(z.Read(z.PC+1)))
	case 0x19: // add hl, de
		v1, v2 := z.getHL(), z.getDE()
		z.addOpHL(v1, v2)
	case 0x1a: // ld a, (de)
		z.setOpA(7, 1, z.followDE(), 0x222222)
	case 0x1b: // dec de
		z.setOpDE(6, 1, z.getDE()-1, 0x222222)
	case 0x1c: // inc e
		z.incOpReg(&z.E)
	case 0x1d: // dec e
		z.decOpReg(&z.E)
	case 0x1e: // ld e, n8
		z.setOpE(7, 2, z.Read(z.PC+1), 0x222222)
	case 0x1f: // rra
		z.rraOp()

	case 0x20: // jr nz, r8
		z.jmpRel8(12, 7, 2, !z.getZeroFlag(), int8(z.Read(z.PC+1)))
	case 0x21: // ld hl, nn
		z.setOpHL(10, 3, z.read16(z.PC+1), 0x222222)
	case 0x22: // ld (nn), hl
		z.setOpMem16(16, 3, z.read16(z.PC+1), z.getHL(), 0x222222)
	case 0x23: // inc hl
		z.setOpHL(6, 1, z.getHL()+1, 0x222222)
	case 0x24: // inc h
		z.incOpReg(&z.H)
	case 0x25: // dec h
		z.decOpReg(&z.H)
	case 0x26: // ld h, d8
		z.setOpH(7, 2, z.Read(z.PC+1), 0x222222)
	case 0x27: // daa
		z.daaOp()

	case 0x28: // jr z, r8
		z.jmpRel8(12, 7, 2, z.getZeroFlag(), int8(z.Read(z.PC+1)))
	case 0x29: // add hl, hl
		v1, v2 := z.getHL(), z.getHL()
		z.addOpHL(v1, v2)
	case 0x2a: // ld hl, (nn)
		addr := z.read16(z.PC + 1)
		z.setOpHL(16, 3, z.read16(addr), 0x222222)
	case 0x2b: // dec hl
		z.setOpHL(6, 1, z.getHL()-1, 0x222222)
	case 0x2c: // inc l
		z.incOpReg(&z.L)
	case 0x2d: // dec l
		z.decOpReg(&z.L)
	case 0x2e: // ld l, d8
		z.setOpL(7, 2, z.Read(z.PC+1), 0x222222)
	case 0x2f: // cpl
		z.setOpA(4, 1, ^z.A, 0x221212)

	case 0x30: // jr z, r8
		z.jmpRel8(12, 7, 2, !z.getCarryFlag(), int8(z.Read(z.PC+1)))
	case 0x31: // ld sp, nn
		z.setOpSP(10, 3, z.read16(z.PC+1), 0x222222)
	case 0x32: // ld (nn), a
		z.setOpMem8(13, 3, z.read16(z.PC+1), z.A, 0x222222)
	case 0x33: // inc sp
		z.setOpSP(6, 1, z.SP+1, 0x222222)
	case 0x34: // inc (hl)
		val := z.followHL()
		result := val + 1
		overFlag := overflagAdd(val, 1, result)
		z.setOpMem8(11, 1, z.getHL(), result, signFlag(result)|zFlag(result)|hFlagAdd(val, 1)|overFlag|0x02)
	case 0x35: // dec (hl)
		val := z.followHL()
		result := val - 1
		overFlag := overflagSub(val, 1, result)
		z.setOpMem8(11, 1, z.getHL(), result, signFlag(result)|zFlag(result)|hFlagSub(val, 1)|overFlag|0x12)
	case 0x36: // ld (hl), n8
		z.setOpMem8(10, 2, z.getHL(), z.Read(z.PC+1), 0x222222)
	case 0x37: // scf
		z.setOpFn(4, 1, func() {}, 0x220201)

	case 0x38: // jr c, r8
		z.jmpRel8(12, 7, 2, z.getCarryFlag(), int8(z.Read(z.PC+1)))
	case 0x39: // add hl, sp
		v1, v2 := z.getHL(), z.SP
		z.addOpHL(v1, v2)
	case 0x3a: // ld a, (nn)
		z.setOpA(13, 3, z.Read(z.read16(z.PC+1)), 0x222222)
	case 0x3b: // dec sp
		z.setOpSP(6, 1, z.SP-1, 0x222222)
	case 0x3c: // inc a
		z.incOpReg(&z.A)
	case 0x3d: // dec a
		z.decOpReg(&z.A)
	case 0x3e: // ld a, n8
		z.setOpA(7, 2, z.Read(z.PC+1), 0x222222)
	case 0x3f: // ccf
		newH := uint32(z.F&1) << 8
		newC := uint32(z.F&1) ^ 1
		z.setOpFn(4, 1, func() {}, 0x220200|newH|newC)

	case 0x70: // ld (hl), b
		z.setOpMem8(7, 1, z.getHL(), z.B, 0x222222)
	case 0x71: // ld (hl), c
		z.setOpMem8(7, 1, z.getHL(), z.C, 0x222222)
	case 0x72: // ld (hl), d
		z.setOpMem8(7, 1, z.getHL(), z.D, 0x222222)
	case 0x73: // ld (hl), e
		z.setOpMem8(7, 1, z.getHL(), z.E, 0x222222)
	case 0x74: // ld (hl), h
		z.setOpMem8(7, 1, z.getHL(), z.H, 0x222222)
	case 0x75: // ld (hl), l
		z.setOpMem8(7, 1, z.getHL(), z.L, 0x222222)
	case 0x76: // halt
		z.setOpFn(4, 1, func() { z.IsHalted = true }, 0x222222)
	case 0x77: // ld (hl), a
		z.setOpMem8(7, 1, z.getHL(), z.A, 0x222222)

	case 0xc0: // ret nz
		z.jmpRet(11, 5, 1, !z.getZeroFlag())
	case 0xc1: // pop bc
		z.popOp16(10, 1, z.setBC)
	case 0xc2: // jp nz, nn
		z.jmpAbs16(10, 10, 3, !z.getZeroFlag(), z.read16(z.PC+1))
	case 0xc3: // jp nn
		z.setOpPC(10, 3, z.read16(z.PC+1), 0x222222)
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
		z.setOpFn(11, 2, func() {
			addr := uint16(z.A)<<8 | uint16(z.Read(z.PC+1))
			z.Out(addr, z.A)
		}, 0x222222)
	case 0xd4: // call nc, nn
		z.jmpCall(17, 10, 3, !z.getCarryFlag(), z.read16(z.PC+1))
	case 0xd5: // push de
		z.pushOp16(11, 1, z.getDE())
	case 0xd6: // sub n8
		z.subOpA(7, 2, z.Read(z.PC+1))
	case 0xd7: // rst 10h
		z.callOp(11, 1, 0x0010)

	case 0xd8: // ret c
		z.jmpRet(10, 5, 1, z.getCarryFlag())
	case 0xd9: // exx
		z.setOpFn(4, 1, func() {
			xchg(&z.Bh, &z.B)
			xchg(&z.Ch, &z.C)
			xchg(&z.Dh, &z.D)
			xchg(&z.Eh, &z.E)
			xchg(&z.Hh, &z.H)
			xchg(&z.Lh, &z.L)
		}, 0x222222)
	case 0xda: // jp c, nn
		z.jmpAbs16(10, 10, 3, z.getCarryFlag(), z.read16(z.PC+1))
	case 0xdb: // in a, (n)
		z.setOpFn(11, 2, func() {
			addr := uint16(z.A)<<8 | uint16(z.Read(z.PC+1))
			z.A = z.In(addr)
		}, 0x222222)
	case 0xdc: // call c, nn
		z.jmpCall(17, 10, 3, z.getCarryFlag(), z.read16(z.PC+1))
	case 0xdd:
		z.stepDDPrefixOpcode()
	case 0xde: // sbc n8
		z.sbcOpA(7, 2, z.Read(z.PC+1))
	case 0xdf: // rst 18h
		z.callOp(11, 1, 0x0018)

	case 0xe0: // ret po
		z.jmpRet(10, 5, 1, !z.getParityOverflowFlag())
	case 0xe1: // pop hl
		z.popOp16(12, 1, z.setHL)
	case 0xe2: // jp po, nn
		z.jmpAbs16(10, 10, 3, !z.getParityOverflowFlag(), z.read16(z.PC+1))
	case 0xe3:
		z.setOpFn(19, 1, func() {
			tmp := z.read16(z.SP)
			z.write16(z.SP, z.getHL())
			z.setHL(tmp)
		}, 0x222222)
	case 0xe4: // call po, nn
		z.jmpCall(17, 10, 3, !z.getParityOverflowFlag(), z.read16(z.PC+1))
	case 0xe5: // push hl
		z.pushOp16(16, 1, z.getHL())
	case 0xe6: // and n8
		z.andOpA(8, 2, z.Read(z.PC+1))
	case 0xe7: // rst 20h
		z.callOp(16, 1, 0x0020)

	case 0xe8: // ret pe
		z.jmpRet(10, 5, 1, z.getParityOverflowFlag())
	case 0xe9: // jp hl (also written jp (hl))
		z.setOpPC(4, 1, z.getHL(), 0x222222)
	case 0xea: // jp pe, nn
		z.jmpAbs16(10, 10, 3, z.getParityOverflowFlag(), z.read16(z.PC+1))
	case 0xeb: // ex de, hl
		z.setOpFn(4, 1, func() {
			xchg(&z.D, &z.H)
			xchg(&z.E, &z.L)
		}, 0x222222)
	case 0xec: // call pe, nn
		z.jmpCall(17, 10, 3, z.getParityOverflowFlag(), z.read16(z.PC+1))
	case 0xed:
		z.stepEDPrefixOpcode()
	case 0xee: // xor n8
		z.xorOpA(8, 2, z.Read(z.PC+1))
	case 0xef: // rst 28h
		z.callOp(16, 1, 0x0028)

	case 0xf0: // ret p
		z.jmpRet(10, 5, 1, !z.getSignFlag())
	case 0xf1: // pop af
		z.popOp16(12, 1, z.setAF)
	case 0xf2: // jp p, nn
		z.jmpAbs16(10, 10, 3, !z.getSignFlag(), z.read16(z.PC+1))
	case 0xf3: // di
		z.setOpFn(4, 1, func() { z.InterruptMasterEnable = false }, 0x222222)
	case 0xf4: // call p, nn
		z.jmpCall(17, 10, 3, !z.getSignFlag(), z.read16(z.PC+1))
	case 0xf5: // push af
		z.pushOp16(16, 1, z.getAF())
	case 0xf6: // or n8
		z.orOpA(8, 2, z.Read(z.PC+1))
	case 0xf7: // rst 30h
		z.callOp(16, 1, 0x0030)

	case 0xf8: // ret m
		z.jmpRet(10, 5, 1, z.getSignFlag())
	case 0xf9: // ld sp, hl
		z.setOpSP(8, 1, z.getHL(), 0x222222)
	case 0xfa: // jp m, nn
		z.jmpAbs16(10, 10, 3, z.getSignFlag(), z.read16(z.PC+1))
	case 0xfb: // ei
		z.setOpFn(4, 1, func() { z.MasterEnableRequested = true }, 0x222222)
	case 0xfc: // call m, nn
		z.jmpCall(17, 10, 3, z.getSignFlag(), z.read16(z.PC+1))
	case 0xfd: // illegal
		z.stepFDPrefixOpcode()
	case 0xfe: // cp a, n8
		z.cpOp(8, 2, z.Read(z.PC+1))
	case 0xff: // rst 38h
		z.callOp(16, 1, 0x0038)

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
		z.bitOp(8, 15, 2, extOpcode, 0)
	case 0x48: // bit 1, R_OR_(HL)
		z.bitOp(8, 15, 2, extOpcode, 1)
	case 0x50: // bit 2, R_OR_(HL)
		z.bitOp(8, 15, 2, extOpcode, 2)
	case 0x58: // bit 3, R_OR_(HL)
		z.bitOp(8, 15, 2, extOpcode, 3)
	case 0x60: // bit 4, R_OR_(HL)
		z.bitOp(8, 15, 2, extOpcode, 4)
	case 0x68: // bit 5, R_OR_(HL)
		z.bitOp(8, 15, 2, extOpcode, 5)
	case 0x70: // bit 6, R_OR_(HL)
		z.bitOp(8, 15, 2, extOpcode, 6)
	case 0x78: // bit 7, R_OR_(HL)
		z.bitOp(8, 15, 2, extOpcode, 7)

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

func (z *z80) rlaOp() {
	result, flags := z.rlOp(z.A)
	z.setOp8(4, 1, &z.A, result, 0x220200|(flags&1))
}
func (z *z80) rlOp(val byte) (byte, uint32) {
	result, carry := (val<<1)|(z.F&1), (val >> 7)
	return result, signFlag(result) | zFlag(result) | parityFlag(result) | uint32(carry)
}

func (z *z80) rraOp() {
	result, flags := z.rrOp(z.A)
	z.setOp8(4, 1, &z.A, result, 0x220200|(flags&1))
}
func (z *z80) rrOp(val byte) (byte, uint32) {
	result, carry := (z.F<<7)|(val>>1), val&1
	return result, signFlag(result) | zFlag(result) | parityFlag(result) | uint32(carry)
}

func (z *z80) rlcaOp() {
	result, flags := z.rlcOp(z.A)
	z.setOp8(4, 1, &z.A, result, 0x220200|(flags&1))
}
func (z *z80) rlcOp(val byte) (byte, uint32) {
	result, carry := (val<<1)|(val>>7), val>>7
	return result, signFlag(result) | zFlag(result) | parityFlag(result) | uint32(carry)
}

func (z *z80) rrcaOp() {
	result, flags := z.rrcOp(z.A)
	z.setOp8(4, 1, &z.A, result, 0x220200|(flags&1))
}
func (z *z80) rrcOp(val byte) (byte, uint32) {
	result, carry := (val<<7)|(val>>1), val&1
	return result, signFlag(result) | zFlag(result) | parityFlag(result) | uint32(carry)
}

func (z *z80) sllOp(val byte) (byte, uint32) {
	result, carry := (val<<1)|1, val>>7
	return result, zFlag(result) | parityFlag(result) | uint32(carry)
}

func (z *z80) srlOp(val byte) (byte, uint32) {
	result, carry := val>>1, val&1
	return result, zFlag(result) | parityFlag(result) | uint32(carry)
}

func (z *z80) slaOp(val byte) (byte, uint32) {
	result, carry := val<<1, val>>7
	return result, signFlag(result) | zFlag(result) | parityFlag(result) | uint32(carry)
}

func (z *z80) sraOp(val byte) (byte, uint32) {
	result, carry := (val&0x80)|(val>>1), val&0x01
	return result, signFlag(result) | zFlag(result) | parityFlag(result) | uint32(carry)
}

func (z *z80) bitOp(cyclesReg uint, cyclesHL uint, instLen uint16, opcode byte, bitNum uint8) {
	cycles, val := z.getCyclesAndValFromOpBits(cyclesReg, cyclesHL, opcode)
	z.setOpFn(cycles, instLen, func() {}, 0x201202|zFlag(val&(1<<bitNum)))
}

func (z *z80) getResOp(bitNum uint) func(byte) (byte, uint32) {
	return func(val byte) (byte, uint32) {
		result := val &^ (1 << bitNum)
		return result, 0x222222
	}
}

func (z *z80) getBitSetOp(bitNum uint8) func(byte) (byte, uint32) {
	return func(val byte) (byte, uint32) {
		result := val | (1 << bitNum)
		return result, 0x222222
	}
}

func (z *z80) runAndUpdatePC(numCycles uint, instLen uint16) {
	z.RunCycles(numCycles)
	z.PC += instLen
}

func (z *z80) stepEDPrefixOpcode() {
	extOpcode := z.Read(z.PC + 1)
	switch extOpcode {
	case 0x40: // in b, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		z.setOpB(12, 2, result, signFlag(result)|zFlag(result)|parityFlag(result)|0x02)
	case 0x41: // out (c), b
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.B)
		z.runAndUpdatePC(12, 2)
	case 0x42: // sbc hl, ss
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x43: // ld (nn), bc
		z.setOpMem16(20, 4, z.read16(z.PC+2), z.getBC(), 0x222222)
	case 0x44: // neg
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x45: // retn
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x46: // im 0
		z.InterruptMode = 0
		z.runAndUpdatePC(8, 2)
	case 0x47: // ld i, a
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x48: // in c, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		z.setOpC(12, 2, result, signFlag(result)|zFlag(result)|parityFlag(result)|0x02)
	case 0x49: // out (C), C
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.C)
		z.runAndUpdatePC(12, 2)
	case 0x4a: // adc hl, bc
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x4b: // ld bc, (nn)
		addr := z.read16(z.PC + 2)
		z.setOpBC(20, 4, z.read16(addr), 0x222222)
	case 0x4c: // neg (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x4d: // reti
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x4e: // im 0/1 (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x4f: // ld r, a
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x50: // in d, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		z.setOpD(12, 2, result, signFlag(result)|zFlag(result)|parityFlag(result)|0x02)
	case 0x51: // out (c), d
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.D)
		z.runAndUpdatePC(12, 2)
	case 0x52: // sbc hl, de
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x53: // ld (nn), de
		z.setOpMem16(20, 4, z.read16(z.PC+2), z.getDE(), 0x222222)
	case 0x54: // neg (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x55: // retn (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x56: // im 1
		z.InterruptMode = 1
		z.runAndUpdatePC(8, 2)
	case 0x57: // ld a, i
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x58: // in e, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		z.setOpE(12, 2, result, signFlag(result)|zFlag(result)|parityFlag(result)|0x02)
	case 0x59: // out (c), e
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.E)
		z.runAndUpdatePC(12, 2)
	case 0x5a: // adc hl, de
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x5b: // ld de, (nn)
		addr := z.read16(z.PC + 2)
		z.setOpDE(20, 4, z.read16(addr), 0x222222)
	case 0x5c: // neg (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x5d: // retn (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x5e: // im 2
		z.InterruptMode = 2
		z.runAndUpdatePC(8, 2)
	case 0x5f: // ld a, r
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x60: // in h, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		z.setOpH(12, 2, result, signFlag(result)|zFlag(result)|parityFlag(result)|0x02)
	case 0x61: // out (c), h
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.H)
		z.runAndUpdatePC(12, 2)
	case 0x62: // sbc hl, hl
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x63: // ld (nn), hl
		z.setOpMem16(20, 4, z.read16(z.PC+2), z.getHL(), 0x222222)
	case 0x64: // neg (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x65: // retn (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x66: // im 0 (UNDOCUMENTED)
		z.InterruptMode = 0
		z.runAndUpdatePC(8, 2)
	case 0x67: // rrd
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x68: // in l, (c)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		z.setOpL(12, 2, result, signFlag(result)|zFlag(result)|parityFlag(result)|0x02)
	case 0x69: // out (c), l
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.L)
		z.runAndUpdatePC(12, 2)
	case 0x6a: // adc hl, hl
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x6b: // ld hl, (nn)
		z.setOpHL(20, 4, z.read16(z.PC+2), 0x222222)
	case 0x6c: // neg (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x6d: // retn (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x6e: // im 0/1 (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x6f: // rld
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x70: // in f, (c) / in (c) (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x71: // out (c), 0 (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x72: // sbc hl, sp
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x73: // ld (nn), sp
		z.setOpMem16(20, 4, z.read16(z.PC+2), z.SP, 0x222222)
	case 0x74: // neg (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x75: // retn (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x76: // im 1 (UNDOCUMENTED)
		z.InterruptMode = 1
		z.runAndUpdatePC(8, 2)

	case 0x78: // in a, (c)
		addr := uint16(z.B)<<8 | uint16(z.C)
		result := z.In(addr)
		z.setOpA(12, 2, result, signFlag(result)|zFlag(result)|parityFlag(result)|0x02)
	case 0x79: // out (c), a
		addr := uint16(z.B)<<8 | uint16(z.C)
		z.Out(addr, z.A)
		z.runAndUpdatePC(12, 2)
	case 0x7a: // adc hl, sp
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x7b: // ld sp, (nn)
		z.setOpSP(20, 4, z.read16(z.PC+2), 0x222222)
	case 0x7c: // neg (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x7d: // retn (UNDOCUMENTED)
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0x7e: // im 2 (UNDOCUMENTED)
		z.InterruptMode = 2
		z.runAndUpdatePC(8, 2)

	case 0xa0: // ldi
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xa1: // cpi
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xa2: // ini
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xa3: // outi
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))

	case 0xa8: // ldd
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xa9: // cpd
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xaa: // ind
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xab: // outd
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xb0: // ldir
		srcAddr, dstAddr := z.getHL(), z.getDE()
		src := z.Read(srcAddr)
		z.Write(dstAddr, src)
		z.setHL(srcAddr + 1)
		z.setDE(dstAddr + 1)
		z.setBC(z.getBC() - 1)
		if z.getBC() == 0 {
			z.runAndUpdatePC(16, 2)
		} else {
			z.RunCycles(21)
		}
		z.setFlags(0x220002)
	case 0xb1: // cpir
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xb2: // inir
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xb3: // otir
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))

	case 0xb8: // lddr
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xb9: // cpdr
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xba: // indr
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))
	case 0xbb: // otdr
		z.Err(fmt.Errorf("ED prefix opcode not implemented: 0x%02x", extOpcode))

	default:
		z.Err(fmt.Errorf("bad 0xed extended opcode: 0x%02x", extOpcode))
		// Alas, this is what the z80 does...
		//fmt.Printf("ignoring bad 0xed extended opcode: 0x%02x", extOpcode)
		z.setOpFn(4, 2, func() {}, 0x222222)
	}
}

func (z *z80) stepDDPrefixOpcode() {
	extOpcode := z.Read(z.PC + 1)
	switch extOpcode {
	case 0xe1: // pop ix
		z.popOp16(14, 2, z.setIX)
	case 0xe5: // push ix
		z.pushOp16(15, 2, z.IX)
	default:
		z.Err(fmt.Errorf("DD prefix not yet implemented 0x%02x", extOpcode))
	}
}

func (z *z80) stepFDPrefixOpcode() {
	extOpcode := z.Read(z.PC + 1)
	switch extOpcode {
	case 0xe1: // pop iy
		z.popOp16(14, 2, z.setIY)
	case 0xe5: // push iy
		z.pushOp16(15, 2, z.IY)
	default:
		z.Err(fmt.Errorf("FD prefix not yet implemented 0x%02x", extOpcode))
	}
}

func (z *z80) Err(msg error) {
	fmt.Println()
	fmt.Println("z80.Err():", msg)
	fmt.Println(z.debugStatusLine())
	fmt.Println()
	panic("z80.Err")
}
