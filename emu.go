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

	GetCartRAM() []byte
	CartRAMModified() bool
	SetCartRAM(ram []byte) error

	IsPAL() bool
}

func (emu *emuState) MakeSnapshot() []byte {
	return emu.makeSnapshot()
}

func (emu *emuState) LoadSnapshot(snapBytes []byte) (Emulator, error) {
	return emu.loadSnapshot(snapBytes)
}

// NewEmulatorSMS creates a Sega Master System emulation session
func NewEmulatorSMS(cart, bios []byte) Emulator {
	return newState(cart, bios)
}

// NewEmulatorGG creates a Game Gear emulation session
func NewEmulatorGG(cart, bios []byte) Emulator {
	state := newState(cart, bios)
	state.IsGameGear = true
	return state
}

// GetCartRAM returns the current state of external RAM
func (emu *emuState) GetCartRAM() []byte {
	return emu.Mem.CartStorage.CartRAM[:]
}

// CartRAMActive returns if local RAM has been modified, and resets it to false
func (emu *emuState) CartRAMModified() bool {
	val := emu.Mem.CartStorage.CartRAMModified
	emu.Mem.CartStorage.CartRAMModified = false
	return val
}

// SetCartRAM attempts to set the RAM, returning error if size not correct
func (emu *emuState) SetCartRAM(ram []byte) error {
	if len(emu.Mem.CartStorage.CartRAM) == len(ram) {
		copy(emu.Mem.CartStorage.CartRAM[:], ram)
		return nil
	}
	// TODO: better checks (e.g. real format, cart checksum, etc.)
	return fmt.Errorf("ram size mismatch")
}

func (emu *emuState) IsPAL() bool {
	return emu.VDP.TVType == tvPAL
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
