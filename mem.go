package segmago

import "fmt"

type mem struct {
	RAM     [8192]byte
	CartRAM [1024 * 32]byte
	rom     []byte

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
	fmt.Printf("PageCtrl: 0x%02x\n", b)
	m.CartRAMPagedIn = b&8 > 0
	m.PageRAMBank = uint16((b & 4) >> 2)
	//assert(b&0x80 == 0, "dev machine ROM Write mode not implemented")
	assert(b&0x08 == 0, "cart RAM over internal RAM mode not yet implemented")
}

func (emu *emuState) read(addr uint16) byte {
	m := emu.Mem
	var val byte
	if addr < 0x400 {
		val = m.rom[addr]
	} else if addr < 0x4000 {
		bank := m.Page0Bank
		val = m.rom[bank*0x4000+addr]
	} else if addr < 0x8000 {
		bank := m.Page1Bank
		val = m.rom[bank*0x4000+(addr-0x4000)]
	} else if addr < 0xc000 {
		if m.CartRAMPagedIn {
			bank := m.PageRAMBank
			val = m.CartRAM[bank*0x4000+(addr-0x8000)]
		} else {
			bank := m.Page2Bank
			val = m.rom[bank*0x4000+(addr-0x8000)]
		}
	} else if addr < 0xe000 {
		val = m.RAM[addr-0xc000]
	} else {
		val = m.RAM[addr-0xe000]
	}
	return val
}

func (emu *emuState) write(addr uint16, val byte) {
	m := emu.Mem
	if addr < 0x8000 {
		// rom
	} else if addr < 0xc000 {
		if m.CartRAMPagedIn {
			bank := m.PageRAMBank
			m.CartRAM[bank*0x4000+(addr-0x8000)] = val
		} else {
		}
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
