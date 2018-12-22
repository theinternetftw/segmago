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

	glimmer.InitDisplayLoop("segmago", screenW*2, screenH*2, screenW, screenH, func(sharedState *glimmer.WindowState) {
		startEmu(gameName, sharedState, emu)
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

	audio, err := glimmer.OpenAudioBuffer(1, 8192, 44100, 16, 2)
	workingAudioBuffer := make([]byte, audio.BufferSize())
	dieIf(err)

	timer := time.NewTimer(0)
	<-timer.C

	maxRDiff := time.Duration(0)
	maxFDiff := time.Duration(0)
	frameCount := 0

	accuracyProtection := 2 * time.Millisecond

	frametimeGoal := 16.66 * 1000 * 1000 * time.Nanosecond
	if emu.IsPAL() {
		frametimeGoal = 20 * 1000 * 1000 * time.Nanosecond
	}

	snapshotMode := 'x'

	newInput := segmago.Input{}

	// FIXME: settings are for debug right now
	lastFlipTime := time.Now()
	lastSaveTime := time.Now()
	lastInputPollTime := time.Now()

	for {
		now := time.Now()

		inputDiff := now.Sub(lastInputPollTime)
		if inputDiff > 8*time.Millisecond {
			numDown := 'x'

			newInput = segmago.Input{}

			window.Mutex.Lock()
			{
				window.CopyKeyCharArray(newInput.Keys[:])

				cid := func(c glimmer.KeyCode) bool {
					return window.CodeIsDown(c)
				}

				newInput.Joypad1.Up = cid(glimmer.CodeW)
				newInput.Joypad1.Down = cid(glimmer.CodeS)
				newInput.Joypad1.Left = cid(glimmer.CodeA)
				newInput.Joypad1.Right = cid(glimmer.CodeD)
				newInput.Joypad1.A = cid(glimmer.CodeJ)
				newInput.Joypad1.B = cid(glimmer.CodeK)
				newInput.Joypad1.Start = cid(glimmer.CodeY)
			}
			window.Mutex.Unlock()

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

		bufferAvailable := audio.BufferAvailable()

		audioBufSlice := workingAudioBuffer[:bufferAvailable]
		if frameCount&0xff == 0 {
			if len(audioBufSlice) != 0 {
				//fmt.Println("audio buf size", len(audioBufSlice))
			}
		}
		emu.ReadSoundBuffer(audioBufSlice)
		audio.Write(audioBufSlice)

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
			window.Mutex.Lock()
			copy(window.Pix, emu.Framebuffer())
			window.RequestDraw()
			window.Mutex.Unlock()

			frameCount++
			if frameCount&0xff == 0 {
				if emu.InDevMode() {
					fmt.Printf("maxRTime %.4f, maxFTime %.4f ", maxRDiff.Seconds(), maxFDiff.Seconds())
					fmt.Printf("accuracyProtection %.4f\n", accuracyProtection.Seconds())
				}
				maxRDiff = 0
				maxFDiff = 0
			}

			rDiff := time.Now().Sub(lastFlipTime)
			maxSleep := frametimeGoal - accuracyProtection
			toSleep := maxSleep - rDiff
			if toSleep > accuracyProtection {
				timer.Reset(toSleep)
				<-timer.C
			} else {
				accuracyProtection /= 2
			}

			waitEnds := lastFlipTime.Add(frametimeGoal)
			for waitEnds.Sub(time.Now()) > 0 {
				// spin
			}

			if rDiff > maxRDiff {
				maxRDiff = rDiff
			}

			fDiff := time.Now().Sub(lastFlipTime)
			if fDiff > maxFDiff {
				maxFDiff = fDiff
			}

			if maxSleep > accuracyProtection && fDiff > frametimeGoal {
				if fDiff-frametimeGoal > accuracyProtection {
					accuracyProtection = fDiff - frametimeGoal
				}
			}

			lastFlipTime = time.Now()
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
