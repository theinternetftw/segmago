package main

import (
	"github.com/theinternetftw/segmago"
	"github.com/theinternetftw/segmago/profiling"
	"github.com/theinternetftw/segmago/platform"

	"golang.org/x/mobile/event/key"

	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {

	defer profiling.Start().Stop()

	assert(len(os.Args) == 2, "usage: ./segmago ROM_FILENAME")
	cartFilename := os.Args[1]

	cart, err := ioutil.ReadFile(cartFilename)
	dieIf(err)

	emu := segmago.NewEmulator(cart)

	screenW := 256
	screenH := 240

	platform.InitDisplayLoop("segmago", screenW*2, screenH*2, screenW, screenH, func(sharedState *platform.WindowState) {
		startEmu(cartFilename, sharedState, emu)
	})
}

func startEmu(filename string, window *platform.WindowState, emu segmago.Emulator) {

	// FIXME: settings are for debug right now
	lastFlipTime := time.Now()
	lastInputPollTime := time.Now()

	snapshotPrefix := filename + ".snapshot"

	audio, err := platform.OpenAudioBuffer(4, 4096, 44100, 16, 2)
	workingAudioBuffer := make([]byte, audio.BufferSize())
	dieIf(err)

	timer := time.NewTimer(0)
	<-timer.C

	maxRDiff := time.Duration(0)
	maxFDiff := 0.0
	frameCount := 0

	frametimeGoal := 1.0/60.0

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

		if emu.FlipRequested() {
			window.Mutex.Lock()
			copy(window.Pix, emu.Framebuffer())
			window.RequestDraw()
			window.Mutex.Unlock()

			frameCount++
			if frameCount & 0xff == 0 {
				fmt.Printf("maxRTime %.4f, maxFTime %.4f\n", maxRDiff.Seconds(), maxFDiff)
				maxRDiff = 0
				maxFDiff = 0
			}

			rDiff := time.Now().Sub(lastFlipTime)
			const accuracyProtection = 2*time.Millisecond
			ftGoalAsDuration := time.Duration(frametimeGoal*1000)*time.Millisecond
			maxSleep := ftGoalAsDuration - accuracyProtection
			toSleep := maxSleep - rDiff
			if toSleep > accuracyProtection {
				timer.Reset(toSleep)
				<-timer.C
			}

			fDiff := 0.0
			for fDiff < frametimeGoal-0.0005 { // seems to be about 0.0005 resolution? so leave a bit of play
				fDiff = time.Now().Sub(lastFlipTime).Seconds()
			}
			if rDiff > maxRDiff {
				maxRDiff = rDiff
			}
			if fDiff > maxFDiff {
				maxFDiff = fDiff
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
