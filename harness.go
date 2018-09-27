package segmago

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
)

// RunZEXTEST emulates just enough of cpm
// to run a comprehensive z80 test
func RunZEXTEST(cart []byte) {
	state := newState(cart)

	// changes for ZEXDOC:

	state.CPU.PC = 0x0100
	state.CPU.SP = 0xdffe

	RAM := make([]byte, 0x10000)
	copy(RAM[0x100:], cart)
	RAM[5] = 0xc9 // ret
	RAM[6] = 0xfe // sp
	RAM[7] = 0xdf // sp

	state.CPU.Read = func(addr uint16) byte {
		return RAM[addr]
	}
	state.CPU.Write = func(addr uint16, val byte) {
		RAM[addr] = val
	}
	state.CPU.RunCycles = func(uint) {
		if state.CPU.PC == 5 {
			switch state.CPU.C {
			case 2:
				fmt.Printf("%c", state.CPU.E)
			case 9:
				ptr := state.CPU.getDE()
				str := ""
				for {
					c := state.CPU.Read(ptr)
					if c == '$' {
						fmt.Printf(str)
						if str == "Tests complete\n\r" {
							os.Exit(0)
						}
						break
					}
					str += fmt.Sprintf("%c", c)
					ptr++
				}
			}
		} else if state.CPU.PC > 0x2800 {
			fmt.Printf("0x%04x\n", state.CPU.PC)
			os.Exit(1)
		}
	}

	for {
		state.Step()
	}
}

type testEvent struct {
	cycleTime uint
	eventType string
	addr      uint16
	data      byte
}
type testEndState struct {
	regs testRegSet

	halted        bool
	endCycleCount uint

	events      []testEvent
	memSettings []memSetting
}

type testRegSet struct {
	AF, BC, DE, HL, AFh, BCh, DEh, HLh, IX, IY, SP, PC, MEMPTR uint16

	I, R, IFF1, IFF2, IM byte
}

type memSetting struct {
	startAddr uint16
	data      []byte
}

type suiteTest struct {
	descript string

	regs testRegSet

	halted    bool
	minCycles uint

	memSettings []memSetting

	endState testEndState
}

func (s suiteTest) String() string {
	return fmt.Sprintf("%v:\n", s.descript) +
		fmt.Sprintf("\t%x\n", s.regs) +
		fmt.Sprintf("\thalted: %v\n", s.halted) +
		fmt.Sprintf("\tminCycles: %v\n", s.minCycles) +
		fmt.Sprintf("\tmemSettings:\n") +
		fmt.Sprintf("\t\t%x\n", s.memSettings) +
		fmt.Sprintf("\tendState:\n") +
		fmt.Sprintf("\t\t%x\n", s.endState.regs) +
		fmt.Sprintf("\t\thalted: %v\n", s.endState.halted) +
		fmt.Sprintf("\t\tendCycleCount: %v\n", s.endState.endCycleCount) +
		fmt.Sprintf("\t\tevents: %#v\n", s.endState.events)
}

func scanInt(s *strScanner) int {
	assert(!s.IsEOF(), fmt.Sprintf("scanInt found eof instead"))
	i, err := strconv.Atoi(s.Next())
	assert(err == nil, fmt.Sprintf("scanInt: %v", err))
	return i
}
func scanHexInt(s *strScanner) int {
	assert(!s.IsEOF(), fmt.Sprintf("scanHexInt found eof instead"))
	i, err := strconv.ParseInt(s.Next(), 16, 32)
	assert(err == nil, fmt.Sprintf("scanInt: %v", err))
	return int(i)
}
func scanUint16(s *strScanner) uint16 {
	i := scanHexInt(s)
	assert(i >= 0 && i <= 0x10000, fmt.Sprintf("scanUnit16: found non uint16 %v!", i))
	return uint16(i)
}
func scanByte(s *strScanner) byte {
	i := scanHexInt(s)
	assert(i >= 0 && i <= 0x100, fmt.Sprintf("scanByte: found non byte %v!", i))
	return byte(i)
}

func scanRegs(s *strScanner) testRegSet {
	r := testRegSet{}
	r.AF = scanUint16(s)
	r.BC = scanUint16(s)
	r.DE = scanUint16(s)
	r.HL = scanUint16(s)
	r.AFh = scanUint16(s)
	r.BCh = scanUint16(s)
	r.DEh = scanUint16(s)
	r.HLh = scanUint16(s)
	r.IX = scanUint16(s)
	r.IY = scanUint16(s)
	r.SP = scanUint16(s)
	r.PC = scanUint16(s)
	r.MEMPTR = scanUint16(s)
	r.I = scanByte(s)
	r.R = scanByte(s)
	r.IFF1 = scanByte(s)
	r.IFF2 = scanByte(s)
	r.IM = scanByte(s)
	return r
}

func scanEvent(s *strScanner) testEvent {
	e := testEvent{}
	e.cycleTime = uint(scanInt(s))
	e.eventType = s.Next()
	e.addr = scanUint16(s)
	if e.eventType != "MC" && e.eventType != "PC" {
		e.data = scanByte(s)
	}
	return e
}
func scanEvents(s *strScanner) []testEvent {
	events := []testEvent{}
	for {
		str := s.Peek()
		if len(str) == 4 {
			break // start of reg list
		}
		events = append(events, scanEvent(s))
	}

	return events
}

func scanMemSetting(s *strScanner) memSetting {
	addr, err := strconv.ParseUint(s.Next(), 16, 16)
	dieIf(err)
	setting := memSetting{startAddr: uint16(addr)}
	for {
		str := s.Next()
		if str == "-1" {
			break
		}
		i, err := strconv.ParseUint(str, 16, 8)
		dieIf(err)
		setting.data = append(setting.data, byte(i))
	}
	return setting
}

func scanMemSettings(s *strScanner) []memSetting {
	settings := []memSetting{}
	for !s.IsEOF() {
		str := s.Peek()
		if str == "-1" {
			s.Next()
			break
		}
		settings = append(settings, scanMemSetting(s))
	}
	return settings
}

type strScanner struct {
	bScan    *bufio.Scanner
	foundEOF bool
	haveWord bool
	word     string
}

func (s *strScanner) IsEOF() bool {
	if !s.haveWord {
		s.Peek()
	}
	return s.foundEOF
}
func (s *strScanner) Next() string {
	if !s.haveWord {
		s.Peek()
	}
	s.haveWord = false
	return s.word
}
func (s *strScanner) Peek() string {
	if !s.haveWord {
		s.foundEOF = !s.bScan.Scan()
		s.word = s.bScan.Text()
		s.haveWord = true
	}
	return s.word
}
func newStrScanner(b []byte) *strScanner {
	r := bytes.NewReader(b)
	s := &strScanner{
		bScan: bufio.NewScanner(r),
	}
	s.bScan.Split(bufio.ScanWords)
	return s
}

func parseSuiteTests(input []byte, expected []byte) []suiteTest {
	tests := []suiteTest{}

	in := newStrScanner(input)
	ex := newStrScanner(expected)

	for {
		test := suiteTest{}
		if in.IsEOF() {
			break
		}
		test.descript = in.Next()

		test.regs = scanRegs(in)
		test.halted = scanInt(in) == 1
		test.minCycles = uint(scanInt(in))

		test.memSettings = scanMemSettings(in)

		nextDescript := in.Peek()

		assert(!ex.IsEOF(), "hit eof early in expected-vals file")
		descript := ex.Next()
		assert(descript == test.descript, fmt.Sprintf("expected desc doesn't match input: %v", descript))

		test.endState.events = scanEvents(ex)
		test.endState.regs = scanRegs(ex)
		test.endState.halted = scanInt(ex) == 1
		test.endState.endCycleCount = uint(scanInt(ex))

		if ex.Peek() != nextDescript {
			test.endState.memSettings = append(test.endState.memSettings, scanMemSetting(ex))
		}
		tests = append(tests, test)
	}
	return tests
}

func setInitialState(s *emuState, test *suiteTest) {
	s.Cycles = 0
	s.CPU.setAF(test.regs.AF)
	s.CPU.setBC(test.regs.BC)
	s.CPU.setDE(test.regs.DE)
	s.CPU.setHL(test.regs.HL)
	s.CPU.setAFh(test.regs.AFh)
	s.CPU.setBCh(test.regs.BCh)
	s.CPU.setDEh(test.regs.DEh)
	s.CPU.setHLh(test.regs.HLh)
	s.CPU.IX = test.regs.IX
	s.CPU.IY = test.regs.IY
	s.CPU.SP = test.regs.SP
	s.CPU.PC = test.regs.PC
	s.CPU.I = test.regs.I
	s.CPU.R = test.regs.R
	s.CPU.InterruptMode = test.regs.IM
	s.CPU.IsHalted = test.halted
	s.CPU.InterruptMasterEnable = test.regs.IFF1 != 0
	s.CPU.InterruptSettingPreNMI = test.regs.IFF2 != 0

	// TODO: load IFF1, IFF2 once we do something with those

	for _, setting := range test.memSettings {
		start := setting.startAddr
		for j, b := range setting.data {
			s.CPU.Write(start+uint16(j), b)
		}
	}
}

func matchUint(testName, fieldName string, output, expected uint) bool {
	if output != expected {
		fmt.Printf("%v: %v: %v should be %v\n", testName, fieldName, output, expected)
		return false
	}
	return true
}
func match16(testName, fieldName string, output, expected uint16) bool {
	if output != expected {
		fmt.Printf("%v: %v: 0x%04x should be 0x%04x\n", testName, fieldName, output, expected)
		return false
	}
	return true
}
func match8(testName, fieldName string, output, expected uint8) bool {
	if output != expected {
		fmt.Printf("%v: %v: 0x%02x should be 0x%02x\n", testName, fieldName, output, expected)
		return false
	}
	return true
}
func matchMem(testName string, s *emuState, settings []memSetting) bool {
	for _, m := range settings {
		for j := range m.data {
			start := m.startAddr
			addr := start + uint16(j)
			output := s.CPU.Read(addr)
			expected := m.data[j]
			if output != expected {
				fmt.Printf("%v: at RAM[0x%04x], 0x%02x should be 0x%02x\n", testName, addr, output, expected)
				return false
			}
		}
	}
	return true
}
func compareEndState(s *emuState, test *suiteTest) bool {
	endRegs := test.endState.regs

	// TODO: understand what unused flag bits should be set to...
	// until then, clear 'em
	s.CPU.F &^= 0x28
	endRegs.AF &^= 0x0028

	return (match16(test.descript, "AF", s.CPU.getAF(), endRegs.AF) &&
		match16(test.descript, "BC", s.CPU.getBC(), endRegs.BC) &&
		match16(test.descript, "DE", s.CPU.getDE(), endRegs.DE) &&
		match16(test.descript, "HL", s.CPU.getHL(), endRegs.HL) &&
		match16(test.descript, "AFh", s.CPU.getAFh(), endRegs.AFh) &&
		match16(test.descript, "BCh", s.CPU.getBCh(), endRegs.BCh) &&
		match16(test.descript, "DEh", s.CPU.getDEh(), endRegs.DEh) &&
		match16(test.descript, "HLh", s.CPU.getHLh(), endRegs.HLh) &&
		match16(test.descript, "SP", s.CPU.SP, endRegs.SP) &&
		match16(test.descript, "PC", s.CPU.PC, endRegs.PC) &&
		matchUint(test.descript, "EndCycles", s.Cycles, test.endState.endCycleCount) &&
		matchMem(test.descript, s, test.endState.memSettings) &&
		match8(test.descript, "IFF1", boolBit(0, s.CPU.InterruptMasterEnable), endRegs.IFF1) &&
		match8(test.descript, "IFF2", boolBit(0, s.CPU.InterruptSettingPreNMI), endRegs.IFF2) &&
		match8(test.descript, "IM", s.CPU.InterruptMode, endRegs.IM))
	// TODO: check I and R once we do something with those
}

var testsToIgnore = []string{

	// FIXME: run these when you understand what's going
	// on with ports
	"db_1", "db_2", "db_3", "db",
	"ed40", "ed48", "ed50", "ed58", "ed60", "ed68",
	"ed70", "ed78",

	// FIXME: probably doing INI/OUTI/IND/OUTD wrong
	"eda2", "eda2_01", "eda2_02", "eda2_03",
	"eda3", "eda3_01", "eda3_02", "eda3_03", "eda3_04", "eda3_05", "eda3_06", "eda3_07", "eda3_08", "eda3_09", "eda3_10", "eda3_11",
	"edaa", "edaa_01", "edaa_02", "edaa_03",
	"edab", "edab_01", "edab_02", "edab_03", "edab_04", "edab_05", "edab_06", "edab_07", "edab_08", "edab_09", "edab_10", "edab_11",
	"edb2", "edb2_1", "edb2_2", "edb2_3",
	"edb3", "edb3_1", "edb3_2", "edb3_3",
	"edba", "edba_1", "edba_2", "edba_3",
	"edbb", "edbb_1", "edbb_2", "edbb_3",
}

func isIgnorableTest(test *suiteTest) bool {
	for _, testName := range testsToIgnore {
		if testName == test.descript {
			return true
		}
	}
	return false
}

// RunTestSuite parses the test input and expected output, then runs all tests found
func RunTestSuite(input []byte, expected []byte) {
	state := newState([]byte{})

	// no surprises
	state.VDP.LineInterruptEnable = false
	state.VDP.FrameInterruptEnable = false

	// changes for Test Suite:

	RAM := make([]byte, 0x10000)
	state.CPU.Read = func(addr uint16) byte {
		return RAM[addr]
	}
	state.CPU.Write = func(addr uint16, val byte) {
		RAM[addr] = val
	}
	state.CPU.In = func(addr uint16) byte {
		return 0
	}
	state.CPU.Out = func(addr uint16, val byte) {
	}

	tests := parseSuiteTests(input, expected)
	assert(len(tests) > 0, "testSuite: no tests found")

	fmt.Println(len(tests), "tests found")

	fmt.Println(len(testsToIgnore), "tests are being skipped")

	failCount := 0
	for i := range tests {
		if isIgnorableTest(&tests[i]) {
			continue
		}
		setInitialState(state, &tests[i])
		for state.Cycles < tests[i].minCycles {
			state.Step()
		}
		if !compareEndState(state, &tests[i]) {
			failCount++
			//fmt.Println("bad test, stopping early")
			//return
		}
		//fmt.Println("finished test", i)
	}
	fmt.Println()
	fmt.Println("Fail count:", failCount)
	fmt.Println("ALL non-skipped TESTS COMPLETE")
}
