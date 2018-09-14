package segmago

type mem struct {
	RAM [8192]byte
	rom []byte

	CartRAMPagedIn bool
	PageRAMBank    uint16

	Page0Bank uint16
	Page1Bank uint16
	Page2Bank uint16
}

func (m *mem) getPagingControlReg() byte {
	return byteFromBools(
		true,
		true,
		true,
		true,
		m.CartRAMPagedIn,
		m.PageRAMBank == 1,
		true,
		true,
	)
}

func (m *mem) setPagingControlReg(b byte) {
	m.CartRAMPagedIn = b&8 > 0
	m.PageRAMBank = uint16(boolBit(0, b&4 > 0))
	assert(!m.CartRAMPagedIn, "Cart RAM not yet impld")
}

func (emu *emuState) read(addr uint16) byte {
	m := emu.Mem
	if addr < 0x400 {
		return m.rom[addr]
	}
	if addr < 0x4000 {
		bank := m.Page0Bank
		return m.rom[bank*0x4000+addr]
	}
	if addr < 0x8000 {
		bank := m.Page1Bank
		return m.rom[bank*0x4000+(addr-0x4000)]
	}
	if addr < 0xc000 {
		errOut("page2/cartram not impld")
	}
	if addr < 0xe000 {
		return m.RAM[addr-0xc000]
	}
	return m.RAM[addr-0xe000]
}
func (emu *emuState) write(addr uint16, val byte) {
	m := emu.Mem
	if addr < 0x8000 {
		return // rom
	} else if addr < 0xc000 {
		errOut("page2/cartram not impld")
	} else if addr < 0xe000 {
		m.RAM[addr-0xc000] = val
	} else {
		m.RAM[addr-0xe000] = val

		if addr == 0xfffc {
			m.setPagingControlReg(val)
		} else if addr == 0xfffd {
			m.Page0Bank = uint16(val)
		} else if addr == 0xfffe {
			m.Page1Bank = uint16(val)
		} else if addr == 0xffff {
			m.Page2Bank = uint16(val)
		}
	}
}
func (emu *emuState) in(addr uint16) byte {
	addr &= 0xff // sms ignores upper byte
	errOut("IN command not yet impld")
	return 0
}
func (emu *emuState) out(addr uint16, val byte) {
	addr &= 0xff // sms ignores upper byte
	errOut("OUT command not yet impld")
}
