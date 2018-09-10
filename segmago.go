package segmago

import (
	"fmt"
	"os"
)

type emuState struct {
	CPU z80
	Mem mem
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

	return &state
}

// NewStateForZEXDOC emulates just enough of cpm
// to run a comprehensive z80 test
func NewStateForZEXDOC(cart []byte) Emulator {
	state := newState(cart)

	// changes for ZEXDOC:

	state.CPU.PC = 0x0100
	state.CPU.SP = 0xdffe

	RAM := make([]byte, 0x10000)
	copy(RAM[0x100:], cart)
	RAM[5] = 0xc9 // ret
	RAM[6] = 0xfe // sp
	RAM[7] = 0xdf // sp

	state.CPU.Read = func(addr uint16) byte {
		return RAM[addr]
	}
	state.CPU.Write = func(addr uint16, val byte) {
		RAM[addr] = val
	}

	return &testWrap{*state}
}

type testWrap struct {
	emuState
}

func (w *testWrap) Step() {
	if w.CPU.PC > 0x8000 {
		fmt.Printf("0x%04x\n", w.CPU.PC)
		os.Exit(1)
	}
	if w.CPU.PC == 5 {
		switch w.CPU.C {
		case 2:
			fmt.Printf("%c", w.CPU.E)
		case 9:
			ptr := w.CPU.getDE()
			for {
				c := w.CPU.Read(ptr)
				if c == '$' {
					break
				}
				fmt.Printf("%c", c)
				ptr++
			}
		}
	}
	w.emuState.Step()
}

func (emu *emuState) runCycles(numCycles uint) {
}

// Input covers all outside info sent to the Emulator
type Input struct {
}

func (emu *emuState) step() {
	fmt.Println(emu.CPU.debugStatusLine())
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
