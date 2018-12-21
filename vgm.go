package segmago

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type vgmPlayer struct {
	SN76489 sn76489

	Hdr vgmHeader
	GD3 gd3

	CurrentSong      byte
	CurrentSongLen   time.Duration
	CurrentSongStart time.Time

	PlaybackComplete bool

	NumSongs byte

	CmdStream []byte
	CmdPC     uint32

	SamplesToWait     uint16
	StartOfSampleWait uint64

	Paused         bool
	PauseStartTime time.Time

	DbgTerminal dbgTerminal
	DbgScreen   [256 * 240 * 4]byte

	LastFlipCycles uint64
	Cycles         uint64
}

func (vp *vgmPlayer) IsPAL() bool           { return false }
func (vp *vgmPlayer) GetCartRAM() []byte    { return nil }
func (vp *vgmPlayer) CartRAMModified() bool { return false }
func (vp *vgmPlayer) SetCartRAM(ram []byte) error {
	return fmt.Errorf("saves not implemented for VGMs")
}
func (vp *vgmPlayer) MakeSnapshot() []byte { return nil }
func (vp *vgmPlayer) LoadSnapshot(snapBytes []byte) (Emulator, error) {
	return nil, fmt.Errorf("snapshots not implemented for VGMs")
}

type vgmHeader struct {
	Magic [4]byte

	EOFOffset uint32
	Version   uint32

	SNClock     uint32
	YM2413Clock uint32

	GD3Offset uint32

	TotalSamples   uint32
	LoopOffset     uint32
	LoopNumSamples uint32
	TVRate         uint32

	SNFeedback      uint16
	SNShiftRegWidth byte
	SNFlags         byte

	YM2612Clock   uint32
	YM2151Clock   uint32
	VGMDataOffset uint32
}

type gd3Header struct {
	Magic   [4]byte
	Version uint32
	Length  uint32
}

type gd3 struct {
	Hdr              gd3Header
	TrackName        string
	GameName         string
	SystemName       string
	TrackAuthor      string
	ReleaseDate      string
	ConversionAuthor string
	Notes            string
}

func getNullWStr(bytes []byte) (string, []byte) {
	out := []byte{}
	var i int
	for i = 0; i < len(bytes); i += 2 {
		if bytes[i] == 0 && bytes[i+1] == 0 {
			break
		} else if bytes[i+1] != 0 {
			out = append(out, '?')
		} else {
			out = append(out, bytes[i])
		}
	}
	return string(out), bytes[i:]
}

func parseGd3(bytes []byte) (gd3, error) {
	result := gd3{}
	err := readStructLE(bytes, &result.Hdr)
	if err != nil {
		return result, err
	}
	if result.Hdr.Length > uint32(len(bytes)) {
		return result, fmt.Errorf("bad length in gd3 header: %v", result.Hdr.Length)
	}
	data := bytes[12:]
	result.TrackName, data = getNullWStr(data)
	getNullWStr(data) // japanese track name
	result.GameName, data = getNullWStr(data)
	getNullWStr(data) // japanese game name
	result.SystemName, data = getNullWStr(data)
	getNullWStr(data) // japanese system name
	result.TrackAuthor, data = getNullWStr(data)
	getNullWStr(data) // japanese track author
	result.ReleaseDate, data = getNullWStr(data)
	result.ConversionAuthor, data = getNullWStr(data)
	result.Notes, data = getNullWStr(data)

	return result, nil
}

func (hdr *vgmHeader) isNTSC() bool {
	return true
}

func parseVgm(vgm []byte) (vgmHeader, []byte, error) {
	hdr := vgmHeader{}
	err := readStructLE(vgm, &hdr)
	hdr.EOFOffset += 0x04
	if hdr.GD3Offset != 0 {
		hdr.GD3Offset += 0x14
	}
	hdr.LoopOffset += 0x1c
	if hdr.VGMDataOffset == 0 {
		hdr.VGMDataOffset = 0x40
	} else {
		hdr.VGMDataOffset += 0x34
	}
	return hdr, vgm[hdr.VGMDataOffset:], err
}

func readStructLE(structBytes []byte, iface interface{}) error {
	return binary.Read(bytes.NewReader(structBytes), binary.LittleEndian, iface)
}

func hasVgmMagic(vgm []byte) bool {
	return string(vgm[:4]) == "Vgm "
}

// NewVgmPlayer creates an vgmPlayer session
func NewVgmPlayer(vgm []byte) Emulator {

	fmt.Println("VGM TIME!")

	var hdr vgmHeader
	var data []byte
	var err error
	if !hasVgmMagic(vgm) {
		var reader io.Reader
		reader, err = gzip.NewReader(bytes.NewReader(vgm))
		if err == nil {
			vgm, err = ioutil.ReadAll(reader)
			if !hasVgmMagic(vgm) {
				err = fmt.Errorf("was passed gzip file, but not a vgm gzip")
			}
		} else {
			err = fmt.Errorf("implement vgm7z here")
		}
	}
	if err == nil {
		hdr, data, err = parseVgm(vgm)
	}

	fmt.Printf("vgm version: %08x\n", hdr.Version)

	if err != nil {
		return NewErrEmu(fmt.Sprintf("vgm player error\n%s", err.Error()))
	}

	vp := vgmPlayer{
		Hdr:       hdr,
		CmdStream: data,
		NumSongs:  1,
	}
	if vp.Hdr.GD3Offset != 0 {
		if gd3, err := parseGd3(vgm[vp.Hdr.GD3Offset:]); err == nil {
			vp.GD3 = gd3
		} else {
			fmt.Println("gd3 err:", err)
		}
	}
	vp.SN76489.init()

	vp.DbgTerminal = dbgTerminal{w: 256, h: 240, screen: vp.DbgScreen[:]}

	fmt.Println("loop offset:", vp.Hdr.LoopOffset)
	fmt.Println("loop #samples:", vp.Hdr.LoopNumSamples)
	fmt.Println("rate:", vp.Hdr.TVRate)

	// TODO: fix the half-second-of-noise bug that requires this mitigation
	// NOTE: it's in the games too! initial state bug?
	//vp.SamplesToWait = 44100

	vp.initTune(0)

	vp.updateScreen()

	return &vp
}

func (vp *vgmPlayer) initTune(songNum byte) {
	vp.CurrentSong = songNum
	vp.CurrentSongStart = time.Now()
	vp.PlaybackComplete = false
	vp.CmdPC = 0
}

func (vp *vgmPlayer) updateScreen() {

	vp.DbgTerminal.clearScreen()

	vp.DbgTerminal.setXMargin(1)
	vp.DbgTerminal.setPos(1, 1)
	vp.DbgTerminal.writeString("VGM Player\n")
	vp.DbgTerminal.newline()
	vp.DbgTerminal.writeString(vp.GD3.TrackName + "\n")
	vp.DbgTerminal.writeString(vp.GD3.TrackAuthor + "\n")
	vp.DbgTerminal.writeString(vp.GD3.ReleaseDate + "\n")

	nowTime := int(time.Now().Sub(vp.CurrentSongStart).Seconds())
	nowTimeStr := fmt.Sprintf("%02d:%02d", nowTime/60, nowTime%60)
	vp.DbgTerminal.writeString(fmt.Sprintf("%s", nowTimeStr))

	if vp.Paused {
		vp.DbgTerminal.writeString(" *PAUSED*\n")
	} else {
		vp.DbgTerminal.newline()
	}
}

var lastInput time.Time

func (vp *vgmPlayer) prevSong() {
	if vp.CurrentSong > 0 {
		vp.CurrentSong--
		vp.initTune(vp.CurrentSong)
		vp.updateScreen()
	}
}
func (vp *vgmPlayer) nextSong() {
	if vp.CurrentSong < vp.NumSongs-1 {
		vp.CurrentSong++
		vp.initTune(vp.CurrentSong)
		vp.updateScreen()
	}
}
func (vp *vgmPlayer) togglePause() {
	vp.Paused = !vp.Paused
	if vp.Paused {
		vp.PauseStartTime = time.Now()
	} else {
		vp.CurrentSongStart = vp.CurrentSongStart.Add(time.Now().Sub(vp.PauseStartTime))
	}
	vp.updateScreen()
}

func (vp *vgmPlayer) SetInput(input Input) {
	now := time.Now()
	if now.Sub(lastInput).Seconds() > 0.20 {
		if input.Joypad1.Left {
			vp.prevSong()
			lastInput = now
		}
		if input.Joypad1.Right {
			vp.nextSong()
			lastInput = now
		}
		if input.Joypad1.Start {
			vp.togglePause()
			lastInput = now
		}
	}
}

var lastScreenUpdate time.Time

func (vp *vgmPlayer) getCmdStreamByte() byte {
	b := vp.CmdStream[vp.CmdPC]
	vp.CmdPC++
	return b
}
func (vp *vgmPlayer) getCmdStreamWord() uint16 {
	lo := vp.CmdStream[vp.CmdPC]
	vp.CmdPC++
	hi := vp.CmdStream[vp.CmdPC]
	vp.CmdPC++
	return uint16(lo) | uint16(hi)<<8
}
func (vp *vgmPlayer) stepCmd() {
	cmd := vp.getCmdStreamByte()
	switch cmd {
	case 0x4f:
		arg := vp.getCmdStreamByte()
		vp.SN76489.StereoMixerReg = arg
	case 0x50:
		arg := vp.getCmdStreamByte()
		vp.SN76489.sendByte(arg)
	case 0x61:
		arg := vp.getCmdStreamWord()
		vp.SamplesToWait = arg
		vp.StartOfSampleWait = vp.Cycles
	case 0x62:
		vp.SamplesToWait = 735
		vp.StartOfSampleWait = vp.Cycles
	case 0x63:
		vp.SamplesToWait = 882
		vp.StartOfSampleWait = vp.Cycles
	case 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f:
		vp.SamplesToWait = uint16(cmd) - 0x70
		vp.StartOfSampleWait = vp.Cycles
	case 0x66:
		vp.PlaybackComplete = true
	default:
		fmt.Printf("unknown cmd 0x%02x\n", cmd)
		os.Exit(1)
	}
}

func (vp *vgmPlayer) Step() {
	if !vp.Paused {

		now := time.Now()
		if now.Sub(lastScreenUpdate) >= 100*time.Millisecond {
			lastScreenUpdate = now
			vp.updateScreen()
		}
		if vp.CurrentSongLen > 0 && now.Sub(vp.CurrentSongStart) >= vp.CurrentSongLen {
			if vp.CurrentSong < vp.NumSongs-1 {
				vp.nextSong()
			} else {
				vp.initTune(0)
				if !vp.Paused {
					vp.togglePause()
				}
			}
		}
		if vp.PlaybackComplete {
			if vp.Hdr.LoopNumSamples != 0 {
				vp.CmdPC = vp.Hdr.LoopOffset - vp.Hdr.VGMDataOffset
				vp.PlaybackComplete = false
			} else {
				vp.initTune(0)
			}
		}

		if vp.SamplesToWait == 0 {
			vp.stepCmd()
			vp.SN76489.runCycle()
			vp.Cycles++
		} else {
			for i := int32(0); i < vp.SN76489.ClocksPerSample; i++ {
				vp.SN76489.runCycle()
				vp.Cycles++
			}
			vp.SamplesToWait--
		}
	}
}

func (vp *vgmPlayer) ReadSoundBuffer(toFill []byte) {
	if vp.Paused {
		for i := range toFill {
			toFill[i] = 0
		}
	} else {
		vp.SN76489.readSoundBuffer(toFill)
	}
}

func (vp *vgmPlayer) Framebuffer() []byte {
	return vp.DbgScreen[:]
}

func (vp *vgmPlayer) FlipRequested() bool {
	const cyclesPerFrame = 3579545 / 60
	if vp.Cycles-vp.LastFlipCycles >= cyclesPerFrame {
		vp.LastFlipCycles = vp.Cycles
		return true
	}
	return false
}
