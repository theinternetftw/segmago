package segmago

import "fmt"

type vdp struct {
	framebuffer         [256 * 240 * 4]byte
	OnSecondControlByte bool
	AddrReg             uint16
	CodeReg             byte

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

	SMSBackdropColor byte

	SpriteLineBuf [8]uint32
	ColorRAM      [32]byte

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

	ScreenMode byte

	ScreenX uint16
	ScreenY uint16

	FlipRequested bool
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

func (v *vdp) incVCounter() {
	onePastLastLine := v.getOnePastLastActiveLine()
	// TODO: fixups for PAL
	if onePastLastLine == 192 {
		if v.VCounterFixupsThisFrame == 0 && v.VCounter == 0xdb {
			v.VCounterFixupsThisFrame++
			v.VCounter = 0xd5
		} else {
			v.VCounter++
		}
	} else if onePastLastLine == 224 {
		if v.VCounterFixupsThisFrame == 0 && v.VCounter == 0xeb {
			v.VCounterFixupsThisFrame++
			v.VCounter = 0xe5
		} else {
			v.VCounter++
		}
	} else { // 240
	}
}
func (v *vdp) atLastVCounter() bool {
	onePastLastLine := v.getOnePastLastActiveLine()
	// TODO: fixups for PAL
	if onePastLastLine == 192 {
		return v.VCounterFixupsThisFrame == 1 && v.VCounter == 0x00
	} else if onePastLastLine == 224 {
		return v.VCounterFixupsThisFrame == 1 && v.VCounter == 0x00
	} else { // 240
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

func (v *vdp) getBGColor(entry uint16, cPlanes byte) byte {
	palSel := entry >> 11 & 1

	pal := v.ColorRAM[palSel*16:]
	col := pal[cPlanes]
	return col
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
	lineHeight := v.getOnePastLastActiveLine()

	numSprites := 0
	for i := uint16(0); i < 64; i++ {
		spriteY := uint16(v.VRAM[base+i]) + 1
		if spriteY == 0xd1 && lineHeight == 192 {
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
		if numSprites == 8 {
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

	lineHeight := v.getOnePastLastActiveLine()
	if lineHeight == 192 {
		for scrolledTileY >= 28 {
			scrolledTileY -= 28
		}
	} else {
		scrolledTileY &= 31
	}

	bgPriority := [256]bool{}
	bgShow := [256]bool{}
	bgCol := [256]byte{}
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
				pX -= 256
				bgCol[pX] = v.ColorRAM[16:][v.SMSBackdropColor]
			} else {
				cPlanes := v.getBGCPlanes(entry, j, (y+scrollY)&7)
				bgShow[pX] = cPlanes != 0
				bgPriority[pX] = entry&0x1000 > 0
				bgCol[pX] = v.getBGColor(entry, cPlanes)
			}
		}
	}

	if v.MaskColumn0WithOverscanCol {
		bdropCol := v.ColorRAM[16:][v.SMSBackdropColor]
		for j := uint16(0); j < 8; j++ {
			bgCol[j] = bdropCol
		}
	}

	spriteListStorage := [8]sprite{}
	spriteList := v.getSpritesForLine(spriteListStorage[:], y)
	spriteHeight := v.getSpriteHeight()

	colors := [256]byte{}
	for x := uint16(0); x < 256; x++ {
		spriteCplanes := byte(0)
		for i := range spriteList {
			spriteX := spriteList[i].x
			if x >= spriteX && x < spriteX+spriteHeight {
				sprite := spriteList[i]
				colX, colY := x-sprite.x, y-sprite.y
				spriteCplanes = v.getSpriteCplanes(sprite, colX, colY)
				if spriteCplanes != 0 {
					break
				}
			}
		}

		if spriteCplanes != 0 && !(bgPriority[x] && bgShow[x]) {
			pal := v.ColorRAM[16:]
			colors[x] = pal[spriteCplanes]
		} else {
			colors[x] = bgCol[x]
		}
	}

	for x := uint16(0); x < 256; x++ {
		r, g, b := v.getRGB(colors[x])
		v.drawColor(x, y, r, g, b)
	}
}

func (v *vdp) runCycle() {
	v.ScreenX++
	if v.ScreenX == 342 {
		onePastLastLine := v.getOnePastLastActiveLine()
		if v.ScreenY < onePastLastLine {
			v.renderScanline(v.ScreenY)
		}
		v.ScreenX = 0
		v.ScreenY++
		v.incVCounter()
		if v.ScreenY <= onePastLastLine {
			v.LineInterruptCounter--
			if v.LineInterruptCounter == 0xff {
				v.LineInterruptCounter = v.LineInterruptCounterSetReg
				v.LineInterruptPending = true
			}
			if v.ScreenY == onePastLastLine {
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
		assert(v.RegM4, "non-mode-4 modes not yet implemented")
		assert(v.RegM2, "M2 not being set is outside of sega doc'd sms spec and not yet implemented")

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
		assert(!v.RegM1, "modes where M1 is set are not yet implemented")
		assert(!v.RegM3, "modes where M3 is set are not yet implemented")

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
		v.SMSBackdropColor = val & 0x0f

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
