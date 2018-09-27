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

	VDP vdp

	ResetPressed bool

	Cycles uint
}

func newState(cart []byte) *emuState {
	state := emuState{
		Mem: mem{
			rom: cart,
		},
	}
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

	return &state
}

func checkCart(cart []byte) {
	hdrLocs := []int{0x1ff0, 0x3ff0, 0x7ff0}
	hdrStart := 0
	for _, addr := range hdrLocs {
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

func (emu *emuState) runCycles(numCycles uint) {
	for i := uint(0); i < numCycles; i++ {
		emu.Cycles++
		emu.VDP.runCycle()
	}
	if emu.VDP.LineInterruptEnable {
		if emu.VDP.LineInterruptPending {
			emu.CPU.IRQ = true
		}
	}
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
	return byteFromBools(
		!emu.Input.Joypad2.Fire,
		!emu.Input.Joypad1.Fire,
		true,
		!emu.ResetPressed,
		!emu.Input.Joypad2.B,
		!emu.Input.Joypad2.A,
		!emu.Input.Joypad2.Right,
		!emu.Input.Joypad2.Left,
	)
}

func (emu *emuState) step() {
	//	fmt.Println(emu.CPU.debugStatusLine())
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
