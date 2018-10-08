package main

import (
	"github.com/theinternetftw/segmago"
	"github.com/theinternetftw/segmago/profiling"
	"github.com/theinternetftw/segmago/platform"

	"golang.org/x/mobile/event/key"

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

	var emu segmago.Emulator
	if strings.HasSuffix(cartFilename, ".gg") {
		bios = []byte{} // no bios in gg yet
		emu = segmago.NewEmulatorGG(cart, bios)
	} else {
		emu = segmago.NewEmulatorSMS(cart, bios)
	}

	gameName := cartFilename
	if gameName == "null" {
		gameName = biosFilename
	}

	screenW := 256
	screenH := 240

	platform.InitDisplayLoop("segmago", screenW*2, screenH*2, screenW, screenH, func(sharedState *platform.WindowState) {
		startEmu(gameName, sharedState, emu)
	})
}

func startEmu(filename string, window *platform.WindowState, emu segmago.Emulator) {

	// FIXME: settings are for debug right now
	lastFlipTime := time.Now()
	lastSaveTime := time.Now()
	lastInputPollTime := time.Now()

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

	audio, err := platform.OpenAudioBuffer(4, 4096, 44100, 16, 2)
	workingAudioBuffer := make([]byte, audio.BufferSize())
	dieIf(err)

	timer := time.NewTimer(0)
	<-timer.C

	maxRDiff := time.Duration(0)
	maxFDiff := time.Duration(0)
	frameCount := 0

	accuracyProtection := 1*time.Millisecond

	frametimeGoal := 16.66*1000*1000*time.Nanosecond
	if emu.IsPAL() {
		frametimeGoal = 20*1000*1000*time.Nanosecond
	}

	snapshotMode := 'x'

	newInput := segmago.Input{}

	for {
		now := time.Now()

		inputDiff := now.Sub(lastInputPollTime)
		if inputDiff > 8*time.Millisecond {
			numDown := 'x'

			newInput = segmago.Input{}

			window.Mutex.Lock()
			{
				window.CopyKeyCharArray(newInput.Keys[:])

				cid := func (c key.Code) bool {
					return window.CodeIsDown(c)
				}

				newInput.Joypad1.Up = cid(key.CodeW)
				newInput.Joypad1.Down = cid(key.CodeS)
				newInput.Joypad1.Left = cid(key.CodeA)
				newInput.Joypad1.Right = cid(key.CodeD)
				newInput.Joypad1.A = cid(key.CodeJ)
				newInput.Joypad1.B = cid(key.CodeK)
				newInput.Joypad1.Start = cid(key.CodeY)
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
				snapFilename := snapshotPrefix+string(numDown)
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
		audio.Write(emu.ReadSoundBuffer(audioBufSlice))

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
			if frameCount & 0xff == 0 {
				fmt.Printf("maxRTime %.4f, maxFTime %.4f ", maxRDiff.Seconds(), maxFDiff.Seconds())
				fmt.Printf("accuracyProtection %.4f\n", accuracyProtection.Seconds())
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
				if fDiff - frametimeGoal > accuracyProtection {
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
