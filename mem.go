package segmago

import "fmt"

type storage struct {
	rom             []byte
	CartRAM         [32 * 1024]byte
	CartRAMModified bool

	CartRAMPagedIn bool
	PageRAMBank    uint32

	Page0Bank uint32
	Page1Bank uint32
	Page2Bank uint32
}

type mem struct {
	RAM [8192]byte

	SelectedMem *storage
	BIOSStorage storage
	CartStorage storage
	NullStorage storage
}

func (m *mem) init(cart, bios []byte) {

	if len(cart) == 0 {
		cart = make([]byte, 16*1024)
	}

	m.BIOSStorage.init(bios)
	m.CartStorage.init(cart)
	m.NullStorage.init(make([]byte, 16*1024))

	if len(bios) > 0 {
		m.SelectedMem = &m.BIOSStorage
	} else {
		m.SelectedMem = &m.CartStorage
	}
}

func (s *storage) init(rom []byte) {
	s.rom = rom
	s.Page0Bank = s.wrapROMBankNum(0)
	s.Page1Bank = s.wrapROMBankNum(1)
	s.Page2Bank = s.wrapROMBankNum(2)
}

func (s *storage) setPagingControlReg(b byte) {
	s.CartRAMPagedIn = b&8 > 0
	s.PageRAMBank = uint32((b & 4) >> 2)
	// if b != 0x80 {
	// 	fmt.Printf("PageCtrl: 0x%02x\n", b)
	// 	fmt.Printf("\tRAM Paged In: %v\n", s.CartRAMPagedIn)
	// 	fmt.Printf("\tRAM Bank: %v\n", s.PageRAMBank)
	// }
	assert(b&0x10 == 0, "cart RAM over internal RAM mode not yet implemented")
}

func (s *storage) wrapROMBankNum(bankNum byte) uint32 {
	// should always be a power of two - 1
	if len(s.rom) <= 16*1024 {
		return 0
	}
	maxBankNum := byte((len(s.rom) / (16 * 1024)) - 1)
	return uint32(bankNum & maxBankNum)
}

func (s *storage) read(addr uint16) byte {
	var val byte
	if addr < 0x400 {
		val = s.rom[addr]
	} else if addr < 0x4000 {
		bank := s.Page0Bank
		computedAddr := bank*0x4000 + uint32(addr)
		//fmt.Println("ROM BANK", bank, "addr", addr, "computed", computedAddr)
		val = s.rom[computedAddr]
	} else if addr < 0x8000 {
		bank := s.Page1Bank
		computedAddr := bank*0x4000 + uint32(addr-0x4000)
		//fmt.Println("ROM BANK", bank, "addr", addr, "computed", computedAddr)
		val = s.rom[computedAddr]
	} else if addr < 0xc000 {
		if s.CartRAMPagedIn {
			bank := s.PageRAMBank
			computedAddr := bank*0x4000 + uint32(addr-0x8000)
			//fmt.Println("CART RAM READ BANK", bank, "addr", addr, "computed", computedAddr)
			val = s.CartRAM[computedAddr]
		} else {
			bank := s.Page2Bank
			computedAddr := bank*0x4000 + uint32(addr-0x8000)
			//fmt.Println("ROM BANK", bank, "addr", addr, "computed", computedAddr)
			val = s.rom[computedAddr]
		}
	} else {
		errOut(fmt.Sprintf("storage.read: passed non-rom addr 0x%04x", addr))
	}
	return val
}

func (s *storage) write(addr uint16, val byte) {
	if addr < 0x8000 {
		// rom
	} else if addr < 0xc000 {
		if s.CartRAMPagedIn {
			bank := s.PageRAMBank
			computedAddr := bank*0x4000 + uint32(addr-0x8000)
			s.CartRAM[computedAddr] = val
			s.CartRAMModified = true
			//fmt.Println("CART RAM WRITE BANK", bank, "addr", addr, "computed", computedAddr, "val", val)
		}
	} else {
		errOut(fmt.Sprintf("storage.write: passed non-rom addr 0x%04x", addr))
	}
}

func (s *storage) ctrlMapper(addr uint16, val byte) {
	if addr == 0xfffc {
		s.setPagingControlReg(val)
	} else if addr == 0xfffd {
		s.Page0Bank = s.wrapROMBankNum(val)
		//fmt.Println("set bank0:", s.Page0Bank)
	} else if addr == 0xfffe {
		s.Page1Bank = s.wrapROMBankNum(val)
		//fmt.Println("set bank1:", s.Page1Bank)
	} else if addr == 0xffff {
		s.Page2Bank = s.wrapROMBankNum(val)
		//fmt.Println("set bank2:", s.Page2Bank)
	}
}

func (emu *emuState) read(addr uint16) byte {
	m := &emu.Mem

	var val byte
	if addr < 0xc000 {
		val = m.SelectedMem.read(addr)
	} else if addr < 0xe000 {
		val = m.RAM[addr-0xc000]
	} else {
		val = m.RAM[addr-0xe000]
	}
	return val
}

func (emu *emuState) write(addr uint16, val byte) {
	m := &emu.Mem
	if addr < 0xc000 {
		m.SelectedMem.write(addr, val)
	} else if addr < 0xe000 {
		m.RAM[addr-0xc000] = val
	} else {
		m.RAM[addr-0xe000] = val
		m.SelectedMem.ctrlMapper(addr, val)
	}
}

func (emu *emuState) in(addr uint16) byte {
	addr &= 0xff // sms ignores upper byte
	var val byte
	if addr < 0x40 {
		if emu.IsGameGear {
			switch addr {
			case 0:
				val = byteFromBools(
					!emu.Input.Joypad1.Start,
					true,  // isExportedConsole
					false, // isPALConsole
					false,
					false,
					false,
					false,
					false,
				)
			case 1:
				return 0x7f
			case 2:
				return 0xff
			case 3:
				return 0x00
			case 4:
				return 0xff
			case 5:
				return 0x00
			case 6:
				return emu.SN76489.StereoMixerReg
			default:
				return 0xff
			}
		} else {
			// val = 0xff // right for SMS2, SMS has weird bus stuff
			val = 0 // more compatible?, some games try to "read" from mem ctrl port
		}
	} else if addr < 0x80 {
		if addr&1 == 0 {
			val = emu.VDP.readVCounter()
		} else {
			val = emu.VDP.readHCounter()
		}
	} else if addr < 0xc0 {
		if addr&1 == 0 {
			val = emu.VDP.readDataPort()
		} else {
			val = emu.VDP.readControlPort()
		}
	} else { // >= 0xc0
		if emu.IsGameGear {
			switch addr {
			case 0xc0, 0xdc:
				val = emu.readJoyReg0()
			case 0xc1, 0xdd:
				val = emu.readJoyReg1()
			default:
				val = 0xff
			}
		} else {
			if emu.IoDisabled {
				val = 0xff
			} else if addr&1 == 0 {
				val = emu.readJoyReg0()
			} else {
				val = emu.readJoyReg1()
			}
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
		if emu.IsGameGear && addr <= 6 {
			if addr == 6 {
				emu.SN76489.StereoMixerReg = val
			}
		} else if addr&1 == 0 {
			emu.setMemControlReg(val)
		} else {
			emu.setIOControlReg(val)
		}
	} else if addr < 0x80 {
		emu.SN76489.sendByte(val)
	} else if addr < 0xc0 {
		if addr&1 == 0 {
			//fmt.Printf("write data port 0x%02x\n", val)
			emu.VDP.writeDataPort(val, emu.IsGameGear)
		} else {
			//fmt.Printf("write control port 0x%02x\n", val)
			emu.VDP.writeControlPort(val)
		}
	} else { // >= 0xc0
		// NOP: writes == old SG-3000 keyboard ports that don't matter to sms
	}
}
