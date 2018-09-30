package segmago

import "fmt"

type vdp struct {
	framebuffer         [256 * 240 * 4]byte
	OnSecondControlByte bool
	AddrReg             uint16
	CodeReg             byte

	VRAM [16 * 1024]byte

	SpriteLineBuf [8]uint32
	ColorRAM      [32]byte

	BufferReg byte

	FrameInterruptPending bool
	SpriteOverflow        bool
	SpriteCollision       bool

	DisableScrollForRightSide bool
	DisableScrollForTop       bool
	MaskCol0WithOverscanCol   bool
	LineInterruptEnable       bool
	ShiftSpritesLeft          bool
	RegM4                     bool
	RegM2                     bool
	TurnOffSync               bool

	LineInterruptPending       bool
	LineInterruptCounter       byte
	LineInterruptCounterSetReg byte

	ScreenMode byte

	ScreenX uint16
	ScreenY uint16
}

func (v *vdp) writeDataPort(val byte) {
	switch v.CodeReg {
	case 0, 1, 2:
		v.VRAM[v.AddrReg] = val
	case 3:
		mask := uint16(len(v.ColorRAM) - 1)
		v.ColorRAM[v.AddrReg&mask] = val
	}
	v.AddrReg = (v.AddrReg + 1) & 0x3fff

	v.BufferReg = val
	v.OnSecondControlByte = false
}
func (v *vdp) readDataPort() byte {
	val := v.BufferReg

	v.BufferReg = v.VRAM[v.AddrReg]
	v.AddrReg = (v.AddrReg + 1) & 0x3fff

	v.OnSecondControlByte = false
	return val
}

var activeLineTable = [16]uint16{
	192, 192, 192, 192,
	192, 192, 192, 192,
	192, 192, 192, 224,
	192, 192, 240, 224,
}

func (v *vdp) getOnePastLastActiveLine() uint16 {
	return activeLineTable[v.ScreenMode]
}

func (v *vdp) runCycle() {
	v.ScreenX++
	if v.ScreenX == 228 {
		v.ScreenX = 0
		v.ScreenY++
		onePastLastLine := v.getOnePastLastActiveLine()
		if v.ScreenY <= onePastLastLine {
			v.LineInterruptCounter--
			if v.LineInterruptCounter == 0xff {
				v.LineInterruptCounter = v.LineInterruptCounterSetReg
				v.LineInterruptPending = true
			}
			if v.ScreenY == onePastLastLine {
				v.FrameInterruptPending = true
			}
		} else {
			v.LineInterruptCounter = v.LineInterruptCounterSetReg
		}
	}
}

func (v *vdp) setReg(regNum byte, val byte) {
	if regNum >= 11 {
		return
	}
	switch regNum {
	case 0:
		boolsFromByte(val,
			&v.DisableScrollForRightSide,
			&v.DisableScrollForTop,
			&v.MaskCol0WithOverscanCol,
			&v.LineInterruptEnable,
			&v.ShiftSpritesLeft,
			&v.RegM4,
			&v.RegM2,
			&v.TurnOffSync)
		assert(v.RegM4, "non-mode-4 modes not implemented")
		assert(v.RegM2, "messing with M2 outside of sega doc'd sms spec not implemented")
	default:
		errOut(fmt.Sprintf("set vdp reg, 0x%02x, 0x%02x", regNum, val))
	}
}

func (v *vdp) writeControlPort(val byte) {
	if !v.OnSecondControlByte {
		v.AddrReg &^= 0x00ff
		v.AddrReg |= uint16(val)
		v.OnSecondControlByte = true
	} else {
		v.AddrReg &^= 0xff00
		v.AddrReg |= uint16(val&0x3f) << 8
		v.CodeReg = val >> 6
		v.OnSecondControlByte = false
		switch v.CodeReg {
		case 0:
			v.BufferReg = v.VRAM[v.AddrReg]
			v.AddrReg++
		case 2:
			v.setReg(byte(v.AddrReg>>8&0x0f), byte(v.AddrReg))
		}
	}
}
func (v *vdp) readControlPort() byte {
	val := byteFromBools(
		v.FrameInterruptPending,
		v.SpriteOverflow,
		v.SpriteCollision,
		true,
		true,
		true,
		true,
		true,
	)
	v.OnSecondControlByte = false
	v.LineInterruptPending = false
	v.FrameInterruptPending = false
	return val
}
