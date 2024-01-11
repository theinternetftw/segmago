package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/theinternetftw/segmago"
)

func main() {

	assert(len(os.Args) == 3, "usage: ./testz80 INPUT_FNAME EXPECTED_FNAME")

	input, err := ioutil.ReadFile(os.Args[1])
	dieIf(err)
	expected, err := ioutil.ReadFile(os.Args[2])
	dieIf(err)

	segmago.RunTestSuite(input, expected)
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
