package main

import "github.com/theinternetftw/segmago"

import "fmt"
import "io/ioutil"
import "os"

func main() {

	assert(len(os.Args) == 2, "usage: ./segmago ROM_FILENAME")

	cart, err := ioutil.ReadFile(os.Args[1])
	dieIf(err)

	emu := segmago.NewEmulator(cart)

	for {
		emu.Step()
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
