package segmago

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

	InterruptMasterEnable     bool // IFF1
	InterruptEnableNeedsDelay bool

	// z80 irq is more complicated than this,
	// but not for sms!
	IRQ bool

	NMI                    bool
	InterruptSettingPreNMI bool // IFF2

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

func (z *z80) interruptComplete() {
	// TODO: signal to devices that interrupt is complete
}

func (z *z80) handleInterrupts() {
	if z.NMI {
		z.NMI = false
		z.InterruptSettingPreNMI = z.InterruptMasterEnable
		z.InterruptMasterEnable = false
		z.pushOp16(11, 0, z.PC)
		z.PC = 0x0066
	} else if z.IRQ {
		if z.IsHalted {
			z.resumeFromHalt()
		}
		if z.InterruptMasterEnable && !z.InterruptEnableNeedsDelay {
			z.InterruptMasterEnable = false
			z.pushOp16(13, 0, z.PC)
			if z.InterruptMode == 0 {
				// IMPORTANT: not real mode 0 logic
				// Only valid for SMS2 and genesis!
				z.PC = 0x0038
			} else if z.InterruptMode == 1 {
				z.PC = 0x0038
			} else {
				// IMPORTANT: not real mode 0 logic
				// Only valid for SMS2 and genesis!
				z.PC = uint16(z.I)<<8 | 0xff
			}
		}
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
	// 0x00000000 clear all flags
	// 0x11111111 set all flags
	// 0x22222222 leave all flags

	clearSign := ^flags >> 29 & 1
	clearF5 := ^flags >> 21 & 1
	clearZero := ^flags >> 25 & 1
	clearHalfCarry := ^flags >> 17 & 1
	clearF3 := ^flags >> 13 & 1
	clearParityOverflow := ^flags >> 9 & 1
	clearSub := ^flags >> 5 & 1
	clearCarry := ^flags >> 1 & 1

	setSign := flags >> 28 & 1
	setZero := flags >> 24 & 1
	setF5 := flags >> 20 & 1
	setHalfCarry := flags >> 16 & 1
	setF3 := flags >> 12 & 1
	setParityOverflow := flags >> 8 & 1
	setSub := flags >> 4 & 1
	setCarry := flags & 1

	clearBits := clearSign << 7
	clearBits |= clearZero << 6
	clearBits |= clearF5 << 5
	clearBits |= clearHalfCarry << 4
	clearBits |= clearF3 << 3
	clearBits |= clearParityOverflow << 2
	clearBits |= clearSub << 1
	clearBits |= clearCarry

	setBits := setSign << 7
	setBits |= setZero << 6
	setBits |= setF5 << 5
	setBits |= setHalfCarry << 4
	setBits |= setF3 << 3
	setBits |= setParityOverflow << 2
	setBits |= setSub << 1
	setBits |= setCarry

	z.F &^= byte(clearBits)
	z.F |= byte(setBits)
}

func (z *z80) getAF() uint16 { return (uint16(z.A) << 8) | uint16(z.F) }
func (z *z80) getBC() uint16 { return (uint16(z.B) << 8) | uint16(z.C) }
func (z *z80) getDE() uint16 { return (uint16(z.D) << 8) | uint16(z.E) }
func (z *z80) getHL() uint16 { return (uint16(z.H) << 8) | uint16(z.L) }

func (z *z80) setAF(val uint16) { z.A, z.F = byte(val>>8), byte(val) }
func (z *z80) setBC(val uint16) { z.B, z.C = byte(val>>8), byte(val) }
func (z *z80) setDE(val uint16) { z.D, z.E = byte(val>>8), byte(val) }
func (z *z80) setHL(val uint16) { z.H, z.L = byte(val>>8), byte(val) }

func (z *z80) getAFh() uint16 { return (uint16(z.Ah) << 8) | uint16(z.Fh) }
func (z *z80) getBCh() uint16 { return (uint16(z.Bh) << 8) | uint16(z.Ch) }
func (z *z80) getDEh() uint16 { return (uint16(z.Dh) << 8) | uint16(z.Eh) }
func (z *z80) getHLh() uint16 { return (uint16(z.Hh) << 8) | uint16(z.Lh) }

func (z *z80) setAFh(val uint16) { z.Ah, z.Fh = byte(val>>8), byte(val) }
func (z *z80) setBCh(val uint16) { z.Bh, z.Ch = byte(val>>8), byte(val) }
func (z *z80) setDEh(val uint16) { z.Dh, z.Eh = byte(val>>8), byte(val) }
func (z *z80) setHLh(val uint16) { z.Hh, z.Lh = byte(val>>8), byte(val) }

func (z *z80) setSP(val uint16) { z.SP = val }
func (z *z80) setPC(val uint16) { z.PC = val }

func (z *z80) setIX(val uint16) { z.IX = val }
func (z *z80) setIY(val uint16) { z.IY = val }

func (z *z80) init() {
	z.setSP(0xfffe)
	z.setPC(0x0100)
}
