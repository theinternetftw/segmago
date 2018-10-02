package segmago

import "fmt"

type mem struct {
	RAM     [8192]byte
	CartRAM [1024 * 32]byte
	rom     []byte

	CartRAMPagedIn bool
	PageRAMBank    uint

	Page0Bank uint
	Page1Bank uint
	Page2Bank uint
}

func (m *mem) setPagingControlReg(b byte) {
	//fmt.Printf("PageCtrl: 0x%02x\n", b)
	m.CartRAMPagedIn = b&8 > 0
	m.PageRAMBank = uint((b & 4) >> 2)
	//fmt.Printf("\tRAM Paged In: %v\n", m.CartRAMPagedIn)
	//fmt.Printf("\tRAM Bank: %v\n", m.PageRAMBank)
	assert(b&0x10 == 0, "cart RAM over internal RAM mode not yet implemented")
}

func (m *mem) wrapROMBankNum(bankNum byte) uint {
	// should always be a power of two - 1
	if len(m.rom) <= 16*1024 {
		return 0
	}
	maxBankNum := byte((len(m.rom) / (16 * 1024)) - 1)
	return uint(bankNum & maxBankNum)
}

func (emu *emuState) read(addr uint16) byte {
	m := &emu.Mem
	var val byte
	if addr < 0x400 {
		val = m.rom[addr]
	} else if addr < 0x4000 {
		bank := m.Page0Bank
		computedAddr := bank*0x4000 + uint(addr)
		//fmt.Println("ROM BANK", bank, "addr", addr, "computed", computedAddr)
		val = m.rom[computedAddr]
	} else if addr < 0x8000 {
		bank := m.Page1Bank
		computedAddr := bank*0x4000 + uint(addr-0x4000)
		//fmt.Println("ROM BANK", bank, "addr", addr, "computed", computedAddr)
		val = m.rom[computedAddr]
	} else if addr < 0xc000 {
		if m.CartRAMPagedIn {
			bank := m.PageRAMBank
			computedAddr := bank*0x4000 + uint(addr-0x8000)
			//fmt.Println("RAM BANK", bank, "addr", addr, "computed", computedAddr)
			val = m.CartRAM[computedAddr]
		} else {
			bank := m.Page2Bank
			computedAddr := bank*0x4000 + uint(addr-0x8000)
			//fmt.Println("ROM BANK", bank, "addr", addr, "computed", computedAddr)
			val = m.rom[computedAddr]
		}
	} else if addr < 0xe000 {
		val = m.RAM[addr-0xc000]
	} else {
		val = m.RAM[addr-0xe000]
	}
	return val
}

func (emu *emuState) write(addr uint16, val byte) {
	m := &emu.Mem
	if addr < 0x8000 {
		// rom
	} else if addr < 0xc000 {
		if m.CartRAMPagedIn {
			bank := m.PageRAMBank
			computedAddr := bank*0x4000 + uint(addr-0x8000)
			m.CartRAM[computedAddr] = val
		}
	} else if addr < 0xe000 {
		m.RAM[addr-0xc000] = val
	} else {
		m.RAM[addr-0xe000] = val

		if addr == 0xfffc {
			m.setPagingControlReg(val)
		} else if addr == 0xfffd {
			m.Page0Bank = m.wrapROMBankNum(val)
			//fmt.Println("set bank0:", m.Page0Bank)
		} else if addr == 0xfffe {
			m.Page1Bank = m.wrapROMBankNum(val)
			//fmt.Println("set bank1:", m.Page1Bank)
		} else if addr == 0xffff {
			m.Page2Bank = m.wrapROMBankNum(val)
			//fmt.Println("set bank2:", m.Page2Bank)
		}
	}
}

func (emu *emuState) in(addr uint16) byte {
	addr &= 0xff // sms ignores upper byte
	var val byte
	if addr < 0x40 {
		val = 0xff // right for SMS2, SMS has weird bus stuff
	} else if addr < 0x80 {
		if addr&1 == 0 {
			val = emu.VDP.readVCounter()
		} else {
			// TODO: add latch, only update value when lightgun pin changes
			val = emu.VDP.readHCounter()
		}
	} else if addr < 0xc0 {
		if addr&1 == 0 {
			val = emu.VDP.readDataPort()
		} else {
			val = emu.VDP.readControlPort()
		}
	} else { // >= 0xc0
		if addr&1 == 0 {
			val = emu.readJoyReg0()
		} else {
			val = emu.readJoyReg1()
		}
	}
	//fmt.Printf("IN: %04x, %02x\n", addr, val)
	//fmt.Printf("got IN: 0x%02x = 0x%02x\n", addr, val)
	return val
}

func (emu *emuState) out(addr uint16, val byte) {
	addr &= 0xff // sms ignores upper byte
	//fmt.Printf("got OUT: 0x%02x, 0x%02x\n", addr, val)
	if addr < 0x40 {
		if addr&1 == 0 {
			errOut(fmt.Sprintf("OUT command not yet impld 0x%02x, 0x%02x", addr, val))
		} else {
			emu.setIOControlReg(val)
		}
	} else if addr < 0x80 {
		if addr&1 == 0 {
			//errOut(fmt.Sprintf("sound not yet impled, got OUT 0x%02x, 0x%02x", addr, val))
		} else {
			//errOut(fmt.Sprintf("sound not yet impled, got OUT 0x%02x, 0x%02x", addr, val))
		}
	} else if addr < 0xc0 {
		if addr&1 == 0 {
			//fmt.Printf("write data port 0x%02x\n", val)
			emu.VDP.writeDataPort(val)
		} else {
			//fmt.Printf("write control port 0x%02x\n", val)
			emu.VDP.writeControlPort(val)
		}
	} else { // >= 0xc0
		// NOP: writes == old SG-3000 keyboard ports that don't matter to sms
	}
}
