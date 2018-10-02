package segmago

import "fmt"

// Emulator exposes the public facing fns for an emulation session
type Emulator interface {
	Step()

	Framebuffer() []byte
	FlipRequested() bool

	SetInput(input Input)
	ReadSoundBuffer([]byte) []byte

	MakeSnapshot() []byte
	LoadSnapshot([]byte) (Emulator, error)
}

func (emu *emuState) MakeSnapshot() []byte {
	return []byte{}
}

func (emu *emuState) LoadSnapshot(snapBytes []byte) (Emulator, error) {
	return nil, fmt.Errorf("snapshots not implemented yet")
}

// NewEmulator creates an emulation session
func NewEmulator(cart, bios []byte) Emulator {
	return newState(cart, bios)
}

// ReadSoundBuffer returns a 44100hz * 16bit * 2ch sound buffer.
// A pre-sized buffer must be provided, which is returned resized
// if the buffer was less full than the length requested.
func (emu *emuState) ReadSoundBuffer(toFill []byte) []byte {
	return emu.SN76489.buffer.read(toFill)
}

// Framebuffer returns the current state of the lcd screen
func (emu *emuState) Framebuffer() []byte {
	return emu.VDP.framebuffer[:]
}

func (emu *emuState) SetInput(input Input) {
	emu.Input = input
}

// FlipRequested indicates if a draw request is pending
func (emu *emuState) FlipRequested() bool {
	req := emu.VDP.FlipRequested
	emu.VDP.FlipRequested = false
	return req
}

// Step steps the emulator one instruction
func (emu *emuState) Step() {
	emu.step()
}
