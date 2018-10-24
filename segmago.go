package segmago

import (
	"bytes"
	"fmt"
	"os"
)

type emuState struct {
	CPU z80
	Mem mem

	Input Input

	VDP     vdp
	SN76489 sn76489

	ResetPressed bool

	THAOutput bool
	THBOutput bool
	TRAOutput bool
	TRBOutput bool

	THAInOutputMode bool
	THBInOutputMode bool
	TRAInOutputMode bool
	TRBInOutputMode bool

	IsDomesticConsole bool

	IoDisabled bool

	IsGameGear            bool
	GameGearExtDataReg    byte
	GameGearExtDirReg     byte
	GameGearSerialSendReg byte
	GameGearSerialCtrlReg byte

	Cycles uint32
}

func (emu *emuState) setMemControlReg(val byte) {
	if emu.IsGameGear {
		// TODO: support enabling the tiny bios in the first k?
	} else {
		if val&0x40 == 0 {
			fmt.Println("set to cart storage")
			emu.Mem.selectedMem = &emu.Mem.CartStorage
		} else if val&0x08 == 0 {
			fmt.Println("set to bios storage")
			emu.Mem.selectedMem = &emu.Mem.BIOSStorage
		} else {
			fmt.Println("set to null storage")
			emu.Mem.selectedMem = &emu.Mem.NullStorage
		}

		if val&0x04 == 0 {
			emu.IoDisabled = false
			fmt.Println("IO Enabled")
		} else {
			emu.IoDisabled = true
			fmt.Println("IO Disabled")
		}
	}
}

func (emu *emuState) setIOControlReg(val byte) {

	var THBInInputMode, TRBInInputMode bool
	var THAInInputMode, TRAInInputMode bool
	var THBOutputTry, TRBOutputTry bool
	var THAOutputTry, TRAOutputTry bool

	boolsFromByte(val,
		&THBOutputTry,
		&TRBOutputTry,
		&THAOutputTry,
		&TRAOutputTry,
		&THBInInputMode,
		&TRBInInputMode,
		&THAInInputMode,
		&TRAInInputMode,
	)

	if emu.THBInOutputMode {
		emu.THBOutput = THBOutputTry
		if emu.THBOutput {
			emu.VDP.updateHCounter()
		}
	}
	if emu.TRBInOutputMode {
		emu.TRBOutput = TRBOutputTry
	}
	if emu.THAInOutputMode {
		emu.THAOutput = THAOutputTry
		if emu.THAOutput {
			emu.VDP.updateHCounter()
		}
	}
	if emu.TRAInOutputMode {
		emu.TRAOutput = TRAOutputTry
	}

	emu.THBInOutputMode = !THBInInputMode
	emu.TRBInOutputMode = !TRBInInputMode
	emu.THAInOutputMode = !THAInInputMode
	emu.TRAInOutputMode = !TRAInInputMode
}

func newState(cart, bios []byte) *emuState {

	// strip a header that is only sometimes seen...
	if len(cart)&0x3fff == 512 {
		fmt.Println("found added header, stripping...")
		cart = cart[512:]
	}

	state := emuState{}

	state.Mem.init(cart, bios)

	state.CPU.Read = state.read
	state.CPU.Write = state.write
	state.CPU.In = state.in
	state.CPU.Out = state.out
	state.CPU.RunCycles = state.runCycles

	checkCart(cart)

	state.Mem.RAM[0] = 0xab
	for i := 1; i < len(state.Mem.RAM); i++ {
		state.Mem.RAM[i] = 0xff
	}
	state.CPU.SP = 0xdfec

	state.SN76489.init()
	state.VDP.init()

	state.GameGearExtDataReg = 0x7f
	state.GameGearExtDirReg = 0xff
	state.GameGearSerialSendReg = 0x00

	return &state
}

func checkCart(cart []byte) {
	hdrLocs := []int{0x1ff0, 0x3ff0, 0x7ff0}
	hdrStart := 0
	for _, addr := range hdrLocs {
		if len(cart) < addr+8 {
			continue
		}
		magic := cart[addr : addr+8]
		if bytes.Equal(magic, []byte("TMR SEGA")) {
			hdrStart = addr
			break
		}
	}
	if hdrStart == 0 {
		fmt.Println("bad cart start!")
	}
}

func (emu *emuState) runCycles(numCycles uint32) {
	for i := uint32(0); i < numCycles; i++ {
		emu.Cycles++
		emu.VDP.runCycle()
		emu.SN76489.runCycle()
	}

	emu.CPU.IRQ = (emu.VDP.LineInterruptEnable && emu.VDP.LineInterruptPending) ||
		(emu.VDP.FrameInterruptEnable && emu.VDP.FrameInterruptPending)
}

// Input covers all outside info sent to the Emulator
type Input struct {
	// Keys is a bool array of keydown state
	Keys [256]bool

	Joypad1 Joypad
	Joypad2 Joypad
}

// Joypad contains gamepad state
type Joypad struct {
	Up    bool
	Down  bool
	Left  bool
	Right bool
	A     bool
	B     bool
	Fire  bool // for lightgun
	Start bool // for Game Gear
}

func (emu *emuState) readJoyReg0() byte {
	return byteFromBools(
		!emu.Input.Joypad2.Down,
		!emu.Input.Joypad2.Up,
		!emu.Input.Joypad1.B,
		!emu.Input.Joypad1.A,
		!emu.Input.Joypad1.Right,
		!emu.Input.Joypad1.Left,
		!emu.Input.Joypad1.Down,
		!emu.Input.Joypad1.Up,
	)
}
func (emu *emuState) readJoyReg1() byte {

	thB := !emu.Input.Joypad2.Fire
	thA := !emu.Input.Joypad1.Fire

	// TODO: both export and domestic console differences
	// (this is export mode, domestic returns 0x00)
	if emu.THBInOutputMode {
		if emu.IsDomesticConsole {
			thB = false
		} else { // export
			thB = emu.THBOutput
		}
	}
	if emu.THAInOutputMode {
		if emu.IsDomesticConsole {
			thB = false
		} else { // export
			thA = emu.THAOutput
		}
	}

	return byteFromBools(
		thB,
		thA,
		true,
		!emu.ResetPressed,
		!emu.Input.Joypad2.B,
		!emu.Input.Joypad2.A,
		!emu.Input.Joypad2.Right,
		!emu.Input.Joypad2.Left,
	)
}

var showDebugStatusLine = false

func (emu *emuState) step() {
	if emu.Input.Keys['`'] {
		showDebugStatusLine = !showDebugStatusLine
	}
	if showDebugStatusLine {
		fmt.Println(emu.CPU.debugStatusLine())
	}
	emu.CPU.Step()
}

func errOut(v ...interface{}) {
	fmt.Println(v...)
	os.Exit(1)
}

func fatalErr(v ...interface{}) {
	fmt.Println(v...)
	panic("fatalErr()")
}

func assert(test bool, msg string) {
	if !test {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func dieIf(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
