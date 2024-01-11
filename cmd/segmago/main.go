package main

import (
	"github.com/theinternetftw/glimmer"
	"github.com/theinternetftw/segmago"
	"github.com/theinternetftw/segmago/profiling"

	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func main() {

	defer profiling.Start().Stop()

	numArgs := len(os.Args)
	assert(numArgs == 2 || numArgs == 3, "usage: ./segmago ROM_FILENAME [BIOS_FILENAME]")
	cartFilename := os.Args[1]

	// TODO: config file instead
	devMode := fileExists("devmode")

	var cart []byte
	if cartFilename != "null" {
		var err error
		cart, err = ioutil.ReadFile(cartFilename)
		dieIf(err)
	}

	bios := []byte{}
	biosFilename := ""
	if numArgs > 2 {
		var err error
		biosFilename = os.Args[2]
		bios, err = ioutil.ReadFile(biosFilename)
		dieIf(err)
	}

	fileMagic := ""
	if len(cart) > 0 {
		fileMagic = string(cart[:4])
	}

	isVGM := strings.HasSuffix(cartFilename, ".vgm") ||
		strings.HasSuffix(cartFilename, ".vgz") ||
		fileMagic == "Vgm "

	var emu segmago.Emulator
	if isVGM {
		emu = segmago.NewVgmPlayer(cart, devMode)
	} else if strings.HasSuffix(cartFilename, ".gg") {
		bios = []byte{} // no bios in gg yet
		emu = segmago.NewEmulatorGG(cart, bios, devMode)
	} else {
		emu = segmago.NewEmulatorSMS(cart, bios, devMode)
	}

	gameName := cartFilename
	if gameName == "null" {
		gameName = biosFilename
	}

	screenW := 256
	screenH := 240

	glimmer.InitDisplayLoop(glimmer.InitDisplayLoopOptions{
		WindowTitle:  "segmago",
		WindowWidth:  screenW * 2,
		WindowHeight: screenH * 2,
		RenderWidth:  screenW,
		RenderHeight: screenH,
		InitCallback: func(sharedState *glimmer.WindowState) {
			startEmu(gameName, sharedState, emu)
		},
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func startEmu(filename string, window *glimmer.WindowState, emu segmago.Emulator) {

	snapshotPrefix := filename + ".snapshot"

	saveFilename := filename + ".sav"
	if saveFile, err := ioutil.ReadFile(saveFilename); err == nil {

		inBuf := bytes.NewBuffer(saveFile)
		gzipReader, err := gzip.NewReader(inBuf)

		var outBuf []byte
		if err == nil {
			outBuf, err = ioutil.ReadAll(gzipReader)
		}

		if err == nil {
			err = emu.SetCartRAM(outBuf)
		}
		if err != nil {
			fmt.Println("could not load savefile", err)
		} else {
			fmt.Println("loaded save!")
		}
	}

	audio, audioErr := glimmer.OpenAudioBuffer(glimmer.OpenAudioBufferOptions{
		OutputBufDuration: 25 * time.Millisecond,
		SamplesPerSecond:  44100,
		BitsPerSample:     16,
		ChannelCount:      2,
	})
	dieIf(audioErr)
	workingAudioBuffer := make([]byte, audio.GetPrevCallbackReadLen())
	audioToGen := audio.GetPrevCallbackReadLen()

	frameTimer := glimmer.MakeFrameTimer()

	snapshotMode := 'x'

	newInput := segmago.Input{}

	lastSaveTime := time.Now()
	lastInputPollTime := time.Now()

	for {
		now := time.Now()

		inputDiff := now.Sub(lastInputPollTime)
		if inputDiff > 8*time.Millisecond {
			numDown := 'x'

			newInput = segmago.Input{}

			window.InputMutex.Lock()
			{
				window.CopyKeyCharArray(newInput.Keys[:])

				cid := func(c glimmer.KeyCode) bool {
					return window.CodeIsDown(c)
				}

				newInput.Joypad1.Up = cid(glimmer.KeyCodeW)
				newInput.Joypad1.Down = cid(glimmer.KeyCodeS)
				newInput.Joypad1.Left = cid(glimmer.KeyCodeA)
				newInput.Joypad1.Right = cid(glimmer.KeyCodeD)
				newInput.Joypad1.A = cid(glimmer.KeyCodeJ)
				newInput.Joypad1.B = cid(glimmer.KeyCodeK)
				newInput.Joypad1.Start = cid(glimmer.KeyCodeY)
			}
			window.InputMutex.Unlock()

			lastInputPollTime = time.Now()

			emu.SetInput(newInput)

			for r := '0'; r <= '9'; r++ {
				if newInput.Keys[r] {
					numDown = r
					break
				}
			}
			if newInput.Keys['m'] {
				snapshotMode = 'm'
			} else if newInput.Keys['l'] {
				snapshotMode = 'l'
			}
			if numDown > '0' && numDown <= '9' {
				snapFilename := snapshotPrefix + string(numDown)
				if snapshotMode == 'm' {
					snapshotMode = 'x'
					numDown = 'x'
					snapshot := emu.MakeSnapshot()
					fmt.Println("writing snap!")
					err := ioutil.WriteFile(snapFilename, snapshot, os.FileMode(0644))
					if err != nil {
						fmt.Println("failed to write snapshot:", err)
					}
				} else if snapshotMode == 'l' {
					snapshotMode = 'x'
					numDown = 'x'
					snapBytes, err := ioutil.ReadFile(snapFilename)
					fmt.Println("loading snap!")
					if err != nil {
						fmt.Println("failed to load snapshot:", err)
						continue
					}
					newEmu, err := emu.LoadSnapshot(snapBytes)
					if err != nil {
						fmt.Println("failed to load snapshot:", err)
						continue
					}
					emu = newEmu
				}
			}
		}

		emu.Step()

		if emu.GetSoundBufferUsed() >= audioToGen {
			if cap(workingAudioBuffer) < audioToGen {
				workingAudioBuffer = make([]byte, audioToGen)
			}
			workingAudioBuffer = workingAudioBuffer[:audioToGen]
			emu.ReadSoundBuffer(workingAudioBuffer)
			audio.Write(workingAudioBuffer)
		}

		if emu.CartRAMModified() {
			if time.Now().Sub(lastSaveTime) > 10*time.Second {
				ram := emu.GetCartRAM()
				if len(ram) > 0 {
					buf := bytes.NewBuffer([]byte{})
					writer := gzip.NewWriter(buf)
					writer.Write(ram)
					writer.Close()

					ioutil.WriteFile(saveFilename, buf.Bytes(), os.FileMode(0644))
					lastSaveTime = time.Now()
					fmt.Println("game saved!")
				}
			}
		}

		if emu.FlipRequested() {
			window.RenderMutex.Lock()
			copy(window.Pix, emu.Framebuffer())
			window.RenderMutex.Unlock()

			audio.WaitForPlaybackIfAhead()

			if emu.InDevMode() {
				frameTimer.PrintStatsEveryXFrames(60 * 5)
			}
		}
	}
}

func dieIf(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func assert(test bool, msg string) {
	if !test {
		fmt.Println(msg)
		os.Exit(1)
	}
}
