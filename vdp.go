package segmago

import "fmt"

type vdp struct {
	framebuffer         [256 * 240 * 4]byte
	OnSecondControlByte bool
	AddrReg             uint16
	CodeReg             byte

	ModeHeight uint16

	VRAM [16 * 1024]byte

	TMS9918NameTableAddr uint16
	SMSNameTableAddr     uint16
	SMSNameTableMaskBit  uint16

	TMS9918SpriteAttrTableAddr uint16
	SMSSpriteAttrTableAddr     uint16
	SMSSpriteAttrTableMaskBit  uint16

	TMS9918SpriteTileTableAddr uint16
	SMSSpriteTileTableAddr     uint16
	SMSSpriteTileTableMaskBit  uint16

	TMS9918ColortableAddr uint16
	TMS9918TileAddr       uint16

	ScrollX uint16
	ScrollY uint16

	SMSBackdropCplane byte

	SpriteLineBuf [8]uint32
	ColorRAM      [64]byte
	GGColorLatch  byte

	BufferReg byte

	FrameInterruptPending bool
	SpriteOverflow        bool
	SpriteCollision       bool

	DisableVertScrollForRightSide bool
	DisableHorizScrollForTop      bool
	MaskColumn0WithOverscanCol    bool
	LineInterruptEnable           bool
	ShiftSpritesLeft              bool
	RegM4                         bool
	RegM2                         bool
	TurnOffSync                   bool

	DisplayEnable        bool
	FrameInterruptEnable bool
	RegM1                bool
	RegM3                bool
	LargeSprites         bool
	StretchedSprites     bool

	LineInterruptPending       bool
	LineInterruptCounter       byte
	LineInterruptCounterSetReg byte

	VCounter                byte
	VCounterFixupsThisFrame byte

	ScreenX uint16
	ScreenY uint16

	FlipRequested bool

	IsGameGear bool
}

func (v *vdp) writeDataPort(val byte, isGameGear bool) {
	switch v.CodeReg {
	case 0, 1, 2:
		v.VRAM[v.AddrReg] = val
	case 3:
		if isGameGear {
			v.IsGameGear = true
			if v.AddrReg&1 == 0 {
				v.GGColorLatch = val
			} else {
				v.ColorRAM[(v.AddrReg-1)&63] = v.GGColorLatch
				v.ColorRAM[v.AddrReg&63] = val
			}
		} else {
			v.ColorRAM[v.AddrReg&31] = val
		}
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

func (v *vdp) incVCounter() {
	// TODO: fixups for PAL
	switch v.ModeHeight {
	case 192:
		if v.VCounterFixupsThisFrame == 0 && v.VCounter == 0xdb {
			v.VCounterFixupsThisFrame++
			v.VCounter = 0xd5
		} else {
			v.VCounter++
		}
	case 224:
		if v.VCounterFixupsThisFrame == 0 && v.VCounter == 0xeb {
			v.VCounterFixupsThisFrame++
			v.VCounter = 0xe5
		} else {
			v.VCounter++
		}
	}
}

func (v *vdp) atLastVCounter() bool {
	// TODO: fixups for PAL
	switch v.ModeHeight {
	case 192:
		return v.VCounterFixupsThisFrame == 1 && v.VCounter == 0x00
	case 224:
		return v.VCounterFixupsThisFrame == 1 && v.VCounter == 0x00
	}
	return false
}

func (v *vdp) readVCounter() byte {
	return v.VCounter
}
func (v *vdp) readHCounter() byte {
	return byte(v.ScreenX >> 1)
}

func (v *vdp) drawColor(x, y uint16, r, g, b byte) {
	base := int(y)*256*4 + int(x)*4
	v.framebuffer[base+0] = r
	v.framebuffer[base+1] = g
	v.framebuffer[base+2] = b
	v.framebuffer[base+3] = 0xff
}

func (v *vdp) getNameTableEntry(tileX, tileY uint16) uint16 {
	baseAddr := v.SMSNameTableAddr
	addr := baseAddr | (tileY << 6) | tileX<<1
	entry := uint16(v.VRAM[addr]) | uint16(v.VRAM[addr+1])<<8
	return entry
}

func (v *vdp) getPatternCplanes(patternNum, x, y uint16) byte {
	pattern := v.VRAM[patternNum*32:]
	line := pattern[4*y:]

	bit0 := line[0] >> (7 - x) & 1
	bit1 := line[1] >> (7 - x) & 1
	bit2 := line[2] >> (7 - x) & 1
	bit3 := line[3] >> (7 - x) & 1

	palIdx := bit0 | bit1<<1 | bit2<<2 | bit3<<3
	return palIdx
}

func (v *vdp) getBGCPlanes(entry, x, y uint16) byte {
	// TODO: xscrollFine, yscrollFine

	patternNum := entry & 0x1ff
	hFlip := entry>>9&1 > 0
	vFlip := entry>>10&1 > 0

	if hFlip {
		x = 7 - x
	}
	if vFlip {
		y = 7 - y
	}

	palIdx := v.getPatternCplanes(patternNum, x, y)
	return palIdx
}

func (v *vdp) getRGB(vdpCol byte) (byte, byte, byte) {
	r := (vdpCol & 3) << 6
	g := (vdpCol >> 2 & 3) << 6
	b := (vdpCol >> 4 & 3) << 6
	return r, g, b
}

func (v *vdp) ggGetRGB(ggCol uint16) (byte, byte, byte) {
	r := byte(ggCol&0x0f) << 4
	g := byte(ggCol & 0xf0)
	b := byte((ggCol >> 4) & 0xf0)
	return r, g, b
}

type sprite struct {
	x, y       uint16
	patternNum uint16
}

func (v *vdp) getSpriteHeight() uint16 {
	spriteHeight := uint16(8)
	if v.LargeSprites {
		spriteHeight *= 2
	}
	if v.StretchedSprites {
		spriteHeight *= 2
	}
	return spriteHeight
}

func (v *vdp) getSpritesForLine(slist []sprite, y uint16) []sprite {
	base := v.SMSSpriteAttrTableAddr

	slist = slist[:0]

	spriteHeight := v.getSpriteHeight()

	numSprites := 0
	for i := uint16(0); i < 64; i++ {
		spriteY := uint16(v.VRAM[base+i]) + 1
		if spriteY == 0xd1 && v.ModeHeight == 192 {
			break
		}
		if y >= spriteY && y < spriteY+spriteHeight {
			spriteX := uint16(v.VRAM[base+0x80+i*2])
			patternNum := uint16(v.VRAM[base+0x80+i*2+1])
			slist = append(slist, sprite{
				x:          spriteX,
				y:          spriteY,
				patternNum: patternNum,
			})
			numSprites++
		}
		if numSprites == 9 {
			v.SpriteOverflow = true
			numSprites = 8
			break
		}
	}

	return slist[:numSprites]
}

func (v *vdp) getSpriteCplanes(sp sprite, x, y uint16) byte {
	tileBase := v.SMSSpriteTileTableAddr

	patternNum := sp.patternNum
	if tileBase > 0 {
		patternNum += 0x100
	}

	if v.StretchedSprites {
		x /= 2
		y /= 2
	}
	if v.LargeSprites {
		if y >= 8 {
			patternNum |= 1
			y -= 8
		} else {
			patternNum &^= 1
		}
	}

	palIdx := v.getPatternCplanes(patternNum, x, y)
	return palIdx
}

func (v *vdp) renderScanline(y uint16) {

	scrollY := v.ScrollY
	scrollX := v.ScrollX

	if v.DisableHorizScrollForTop && y < 16 {
		scrollX = 0
	}

	tileY := y / 8
	scrolledTileY := (y + scrollY) / 8

	if v.ModeHeight == 192 {
		for scrolledTileY >= 28 {
			scrolledTileY -= 28
		}
	} else {
		scrolledTileY &= 31
	}

	bgPriority := [256]bool{}
	bgCplanes := [256]byte{}
	bgPalettes := [256]byte{}
	for i := uint16(0); i < 32; i++ {

		tileX := (i - scrollX/8) & 31

		var entry uint16
		if v.DisableVertScrollForRightSide && i >= 24 {
			entry = v.getNameTableEntry(tileX, tileY)
		} else {
			entry = v.getNameTableEntry(tileX, scrolledTileY)
		}

		for j := uint16(0); j < 8; j++ {
			// TODO: set color to backdrop on pX wrap
			pX := i*8 + j + scrollX&7
			if pX >= 256 {
				bgCplanes[pX-256] = v.SMSBackdropCplane
				bgPalettes[pX-256] = 1
			} else {
				bgCplanes[pX] = v.getBGCPlanes(entry, j, (y+scrollY)&7)
				bgPriority[pX] = bgCplanes[pX] != 0 && entry&0x1000 > 0
				bgPalettes[pX] = byte((entry >> 11) & 1)
			}
		}
	}

	spriteListStorage := [9]sprite{} // extra is used for overflow check
	spriteList := v.getSpritesForLine(spriteListStorage[:], y)
	spriteHeight := v.getSpriteHeight()

	cPlanes := [256]byte{}
	palettes := [256]byte{}
	if v.DisplayEnable {
		for x := uint16(0); x < 256; x++ {
			spriteCplanes := byte(0)
			for i := range spriteList {
				spriteX := spriteList[i].x
				if v.ShiftSpritesLeft {
					spriteX -= 8
				}
				if x >= spriteX && x < spriteX+spriteHeight {
					sprite := spriteList[i]
					colX, colY := x-sprite.x, y-sprite.y
					cPlanes := v.getSpriteCplanes(sprite, colX, colY)
					if spriteCplanes != 0 && cPlanes != 0 {
						v.SpriteCollision = true
						break
					}
					spriteCplanes = cPlanes
				}
			}

			if spriteCplanes != 0 && !bgPriority[x] {
				cPlanes[x] = spriteCplanes
				palettes[x] = 1
			} else {
				cPlanes[x] = bgCplanes[x]
				palettes[x] = bgPalettes[x]
			}
		}

		if v.MaskColumn0WithOverscanCol {
			for j := uint16(0); j < 8; j++ {
				cPlanes[j] = v.SMSBackdropCplane
				palettes[j] = 1
			}
		}

		if v.IsGameGear {
			if y >= 3*8 && y < 3*8+18*8 {
				for x := uint16(6 * 8); x < 256-6*8; x++ {
					pal := v.ColorRAM[32*palettes[x]:]
					color := uint16(pal[cPlanes[x]*2])
					color |= uint16(pal[cPlanes[x]*2+1]) << 8
					r, g, b := v.ggGetRGB(color)
					v.drawColor(x, y, r, g, b)
				}
			}
		} else {
			for x := uint16(0); x < 256; x++ {
				pal := v.ColorRAM[16*palettes[x]:]
				color := pal[cPlanes[x]]
				r, g, b := v.getRGB(color)
				v.drawColor(x, y, r, g, b)
			}
		}

	} else {
		for x := uint16(0); x < 256; x++ {
			v.drawColor(x, y, 0, 0, 0)
		}
	}

}

func (v *vdp) init() {
	v.ModeHeight = 192
	v.RegM2 = true
	v.RegM4 = true
	v.LineInterruptEnable = true
	v.FrameInterruptEnable = true
}

func (v *vdp) runCycle() {

	v.ScreenX++
	if v.ScreenX == 342 {
		if v.ScreenY < v.ModeHeight {
			v.renderScanline(v.ScreenY)
		}
		v.ScreenX = 0
		v.ScreenY++
		v.incVCounter()
		if v.ScreenY <= v.ModeHeight {
			v.LineInterruptCounter--
			if v.LineInterruptCounter == 0xff {
				v.LineInterruptCounter = v.LineInterruptCounterSetReg
				v.LineInterruptPending = true
			}
			if v.ScreenY == v.ModeHeight {
				v.FrameInterruptPending = true
				// TODO: match up vCounter and frame to border/blanking
			}
		} else {
			v.LineInterruptCounter = v.LineInterruptCounterSetReg
		}
		if v.atLastVCounter() {
			v.ScreenY = 0
			v.VCounter = 0
			v.VCounterFixupsThisFrame = 0
			v.FlipRequested = true
		}
	}
}

func (v *vdp) updateMode() {
	assert(v.RegM4, "non-mode-4 modes not yet implemented")
	assert(v.RegM2, "M2 not being set is outside of sega doc'd sms spec and not yet implemented")
	assert(!v.RegM1, "modes where M1 is set are not yet implemented")
	assert(!v.RegM3, "modes where M3 is set are not yet implemented")
}

func (v *vdp) setReg(regNum byte, val byte) {
	if regNum >= 11 {
		fmt.Println("big reg num", regNum)
		return
	}
	switch regNum {
	case 0:
		boolsFromByte(val,
			&v.DisableVertScrollForRightSide,
			&v.DisableHorizScrollForTop,
			&v.MaskColumn0WithOverscanCol,
			&v.LineInterruptEnable,
			&v.ShiftSpritesLeft,
			&v.RegM4,
			&v.RegM2,
			&v.TurnOffSync)
		v.updateMode()

	case 1:
		boolsFromByte(val,
			nil,
			&v.DisplayEnable,
			&v.FrameInterruptEnable,
			&v.RegM1,
			&v.RegM3,
			nil,
			&v.LargeSprites,
			&v.StretchedSprites)
		v.updateMode()

	case 2:
		v.TMS9918NameTableAddr = uint16(val & 0x0f)
		v.SMSNameTableAddr = uint16((val>>1)&0x07) << 11
		v.SMSNameTableMaskBit = uint16(val & 1) // TODO use?

	case 3:
		// FIXME: noting here in case I want to implement sg-1000 mode
		// that i'm unsure on the masking/shifting in this and the other
		// tms99918-only tables.

		v.TMS9918ColortableAddr = uint16(val)

	case 4:
		// TODO: bottom 3 bits should be 1 in sms mode. enforce this?
		v.TMS9918TileAddr = uint16(val & 0x3f)

	case 5:
		v.TMS9918SpriteAttrTableAddr = uint16(val & 0x7f)
		v.SMSSpriteAttrTableAddr = uint16((val>>1)&0x3f) << 8
		v.SMSSpriteAttrTableMaskBit = uint16(val & 1)

	case 6:
		v.TMS9918SpriteTileTableAddr = uint16(val & 0x07)
		v.SMSSpriteTileTableAddr = uint16((val>>2)&1) << 13
		v.SMSSpriteTileTableMaskBit = uint16(val & 1)

	case 7:
		v.SMSBackdropCplane = val & 0x0f

	case 8:
		v.ScrollX = uint16(val)

	case 9:
		v.ScrollY = uint16(val)

	case 10:
		v.LineInterruptCounterSetReg = val
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
