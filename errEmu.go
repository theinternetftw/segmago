package segmago

import (
	"fmt"
	"os"
)

type errEmu struct {
	terminal      dbgTerminal
	screen        [256 * 240 * 4]byte
	flipRequested bool

	devMode bool
}

func (e *errEmu) InDevMode() bool   { return e.devMode }
func (e *errEmu) SetDevMode(b bool) { e.devMode = b }

// NewErrEmu returns an emulator that only shows an error message
func NewErrEmu(msg string) Emulator {
	emu := errEmu{}
	emu.terminal = dbgTerminal{w: 256, h: 240, screen: emu.screen[:]}
	os.Stderr.Write([]byte(msg + "\n"))
	emu.terminal.newline()
	emu.terminal.writeString(msg)
	emu.flipRequested = true
	return &emu
}

func (e *errEmu) IsPAL() bool             { return false }
func (e *errEmu) GetCartRAM() []byte      { return []byte{} }
func (e *errEmu) SetCartRAM([]byte) error { return nil }
func (e *errEmu) CartRAMModified() bool   { return false }
func (e *errEmu) MakeSnapshot() []byte    { return nil }
func (e *errEmu) LoadSnapshot([]byte) (Emulator, error) {
	return nil, fmt.Errorf("snapshots not implemented for errEmu")
}
func (e *errEmu) ReadSoundBuffer(toFill []byte) {}
func (e *errEmu) GetSoundBufferUsed() int       { return 0 }
func (e *errEmu) SetInput(input Input)          {}
func (e *errEmu) Step()                         {}

func (e *errEmu) Framebuffer() []byte { return e.screen[:] }
func (e *errEmu) FlipRequested() bool {
	result := e.flipRequested
	e.flipRequested = false
	return result
}
