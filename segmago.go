package segmago

import (
	"fmt"
	"os"
)

type emuState struct {
	CPU    z80
	Mem    mem
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

	return &state
}

func (emu *emuState) runCycles(numCycles uint) {
	for i := uint(0); i < numCycles; i++ {
		emu.Cycles++
	}
}

// Input covers all outside info sent to the Emulator
type Input struct {
}

func (emu *emuState) step() {
	//fmt.Println(emu.CPU.debugStatusLine())
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
