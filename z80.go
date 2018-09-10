package segmago

import "fmt"

type z80 struct {
	PC                     uint16
	SP                     uint16
	A, F, B, C, D, E, H, L byte

	IX, IY uint16

	I, R byte

	// hidden regs
	Ah, Fh, Bh, Ch, Dh, Eh, Hh, Lh byte

	IsHalted bool

	InterruptMode byte

	InterruptMasterEnable bool
	MasterEnableRequested bool

	NMI                    bool
	InterruptSettingPreNMI bool

	Steps  uint
	Cycles uint

	RunCycles func(cycles uint)
	Read      func(addr uint16) byte
	Write     func(addr uint16, val byte)
	In        func(addr uint16) byte
	Out       func(addr uint16, val byte)
}

func (z *z80) read16(addr uint16) uint16 {
	high := uint16(z.Read(addr + 1))
	low := uint16(z.Read(addr))
	return (high << 8) | low
}

func (z *z80) write16(addr uint16, val uint16) {
	z.Write(addr, byte(val))
	z.Write(addr+1, byte(val>>8))
}

func (z *z80) handleInterrupts() {

	var intFlag *bool
	/*
		if PUT_IRQ_HERE {
			intFlag, intAddr = &z.VBlankIRQ, 0x0040
		} else if PUT_OTHER_IRQ_HERE {
			intFlag, intAddr = &z.LCDStatIRQ, 0x0048
		}
	*/

	if z.NMI {
		z.NMI = false
		z.InterruptSettingPreNMI = z.InterruptMasterEnable
		z.InterruptMasterEnable = false
		z.pushOp16(11, 0, z.PC)
		z.PC = 0x0066
	} else if intFlag != nil {
		if z.InterruptMasterEnable {
			if z.InterruptMode == 0 {
				z.Err(fmt.Errorf("Got interrupt in mode 0, which is not yet implemented"))
			} else if z.InterruptMode == 2 {
				z.Err(fmt.Errorf("Got interrupt in mode 2, which is not yet implemented"))
			} else {
				z.InterruptMasterEnable = false
				*intFlag = false
				z.pushOp16(13, 0, z.PC)
				z.PC = 0x0038
			}
		}
		z.IsHalted = false // NOTE: must they be enabled?
	}
}

func (z *z80) getSignFlag() bool { return z.F&0x80 > 0 }
func (z *z80) getZeroFlag() bool { return z.F&0x40 > 0 }

//
func (z *z80) getHalfCarryFlag() bool { return z.F&0x10 > 0 }

//
func (z *z80) getParityOverflowFlag() bool { return z.F&0x04 > 0 }
func (z *z80) getSubFlag() bool            { return z.F&0x02 > 0 }
func (z *z80) getCarryFlag() bool          { return z.F&0x01 > 0 }

func (z *z80) setFlags(flags uint32) {
	// 0x000000 clear all flags
	// 0x111111 set all flags
	// 0x222222 leave all flags

	setSign, clearSign := flags>>20&1, ^flags>>21&1
	setZero, clearZero := flags>>16&1, ^flags>>17&1
	setHalfCarry, clearHalfCarry := flags>>12&1, ^flags>>13&1
	setParityOverflow, clearParityOverflow := flags>>8&1, ^flags>>9&1
	setSub, clearSub := flags>>4&1, ^flags>>5&1
	setCarry, clearCarry := flags&1, ^flags>>1&1

	z.F &^= byte(clearSign<<7 | clearZero<<6 | clearHalfCarry<<4 | clearParityOverflow<<2 | clearSub<<1 | clearCarry)
	z.F |= byte(setSign<<7 | setZero<<6 | setHalfCarry<<4 | setParityOverflow<<2 | setSub<<1 | setCarry)
}

func (z *z80) getAF() uint16 { return (uint16(z.A) << 8) | uint16(z.F) }
func (z *z80) getBC() uint16 { return (uint16(z.B) << 8) | uint16(z.C) }
func (z *z80) getDE() uint16 { return (uint16(z.D) << 8) | uint16(z.E) }
func (z *z80) getHL() uint16 { return (uint16(z.H) << 8) | uint16(z.L) }

func (z *z80) setAF(val uint16) {
	z.A = byte(val >> 8)
	z.F = byte(val) &^ 0x0f
}
func (z *z80) setBC(val uint16) { z.B, z.C = byte(val>>8), byte(val) }
func (z *z80) setDE(val uint16) { z.D, z.E = byte(val>>8), byte(val) }
func (z *z80) setHL(val uint16) { z.H, z.L = byte(val>>8), byte(val) }

func (z *z80) setSP(val uint16) { z.SP = val }
func (z *z80) setPC(val uint16) { z.PC = val }

func (z *z80) setIX(val uint16) { z.IX = val }
func (z *z80) setIY(val uint16) { z.IY = val }

func (z *z80) init() {
	z.setSP(0xfffe)
	z.setPC(0x0100)
}
