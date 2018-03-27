package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestGetInstruction(t *testing.T) {
	cases := []struct {
		input int         // Always set address 0 to this value
		want  Instruction // The expected instruction
		state CPUState    // Returned CPUState
	}{
		{0, Instruction{"HLT", 0}, CPUok},
		{1001, Instruction{"JEQ", 1}, CPUok},
		{2001, Instruction{"JGT", 1}, CPUok},
		{3001, Instruction{"JLT", 1}, CPUok},
		{5002, Instruction{"JMP", 2}, CPUok},
		{6049, Instruction{"JLE", 49}, CPUok},
		{7030, Instruction{"JNE", 30}, CPUok},
		{7030, Instruction{"JNE", 30}, CPUok},
		{10001, Instruction{"LAC", 1}, CPUok},
		{11001, Instruction{"PAC", 1}, CPUok},
		{12001, Instruction{"LMQ", 1}, CPUok},
		{13002, Instruction{"PMQ", 2}, CPUok},
		{20013, Instruction{"ADD", 13}, CPUok},
		{21014, Instruction{"SUB", 14}, CPUok},
		{22003, Instruction{"MUL", 3}, CPUok},
		{23022, Instruction{"DIV", 22}, CPUok},
		{30031, Instruction{"GET", 31}, CPUok},
		{31032, Instruction{"PUT", 32}, CPUok},
		{1000 + memSize, Instruction{"JEQ", memSize}, CPUbadaddr},  // Valid opcode, invalid memory address.
		{31000 + memSize, Instruction{"PUT", memSize}, CPUbadaddr}, // Valid opcode, invalid memory address.
		{32000, Instruction{"UNK", 0}, CPUbadinst},                 // Invalid opcode.
		{33000, Instruction{"UNK", 0}, CPUbadinst},                 // Invalid opcode.
	}

	h := NewMachine()
	for i, c := range cases {
		h.mem[0] = c.input
		got, cs := h.getInstruction(0)
		if !reflect.DeepEqual(got, c.want) || cs != c.state {
			t.Errorf("%02d: h.getInstruction(0) = (%q, %q); Wanted (%q, %q)", i, got, cs, c.want, c.state)
		}
	}

	bad := []int{-1, memSize, memSize + 1}
	u := Instruction{"UNK", 0}
	for i, b := range bad {
		if got, cs := h.getInstruction(b); !reflect.DeepEqual(got, u) || cs != CPUbadinst {
			t.Errorf("%02d: h.getInstruction(%d) = (%q, %q); Wanted (%q, %q)", i, b, got, cs, u, CPUbadinst)
		}
	}
}

func TestLoadProgram(t *testing.T) {
	cases := []struct {
		prog      string
		want      [memSize]int
		wantErr   error
		wantState CPUState
	}{
		{
			"0: 31000", // No comment
			[memSize]int{31000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			nil,
			CPUok,
		},
		{
			"0: 31000    ", // trailing whitespace, no formal comment
			[memSize]int{31000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			nil,
			CPUok,
		},
		{
			"0: 31000 // A comment",
			[memSize]int{31000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			nil,
			CPUok,
		},
		{
			"0: 31000 // Ok\n49: 21000 // Bookends",
			[memSize]int{31000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 21000},
			nil,
			CPUok,
		},
		{
			"0: 31000 // Ok\n0: 21000 // Overwritten",
			[memSize]int{21000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			nil,
			CPUok,
		},
		{
			"1: 31000 // Ok\n0: 21000 // Out of order",
			[memSize]int{21000, 31000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			nil,
			CPUok,
		},
		{
			"1: -5 // Ok, negative  numbers",
			[memSize]int{0, -5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			nil,
			CPUok,
		},
		{
			": 31000 // Bad line",
			[memSize]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			loadErrBadLine,
			CPUhalt,
		},
		{
			"1: 31aasdf // Bad line",
			[memSize]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			loadErrBadLine,
			CPUhalt,
		},
		{
			"50: 31000 // Bad address",
			[memSize]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			loadErrBadAddr,
			CPUhalt,
		},
		{
			"a: 31000 // Bad address, makes the line bad",
			[memSize]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			loadErrBadLine,
			CPUhalt,
		},
	}

	for i, c := range cases {
		h := NewMachine()
		err := h.LoadProgram(strings.NewReader(c.prog))
		if err != c.wantErr {
			t.Errorf("%02d: h.LoadProgram(blah) = %v; want %v", i, err, c.wantErr)
		}

		if !reflect.DeepEqual(h.mem, c.want) {
			t.Errorf("%02d: h.mem = %v; want %v", i, h.mem, c.want)
		}

		if h.state != c.wantState {
			t.Errorf("%02d: h.state = %q; want %q", h.state, c.wantState)
		}
	}
}

func TestHLT(t *testing.T) {
	h := NewMachine()
	if h.state != CPUok || h.pc != 0 || h.mem[0] != 0 {
		t.Error("Invalid initial state for machine.")
	}

	// With a default machine, all instruction are HALT, so assuming
	// that a single step will halt the machine.
	h.Step()
	if h.state != CPUhalt {
		t.Errorf("h.state = %q; wanted CPUhalt", h.state)
	}
}

func TestJEQ(t *testing.T) {
	cases := []struct {
		pc   int
		inst int
		ac   int
		want int // Value of PC after test
	}{
		{0, 1010, 0, 10}, // Set PC to 10 because AC = 0
		{0, 1010, 1, 1},  // Move PC to 1 (next instruction) because AC != 0
	}

	for i, c := range cases {
		h := NewMachine()
		h.pc = c.pc
		h.mem[c.pc] = c.inst
		h.ac = c.ac
		h.Step()
		if h.pc != c.want {
			t.Errorf("%02d: Stepped machine across JEQ with ac = %d. pc = %d; wanted %d", i, h.ac, h.pc, c.want)
		}
	}
}

func TestJGT(t *testing.T) {
	cases := []struct {
		pc   int
		inst int
		ac   int
		want int // Value of PC after test
	}{
		{0, 2010, 0, 1},  // Move PC to 1 (next instruction) 10 because AC not > 0
		{0, 2010, 1, 10}, // Set PC to 10 because AC > 0
	}

	for i, c := range cases {
		h := NewMachine()
		h.pc = c.pc
		h.mem[c.pc] = c.inst
		h.ac = c.ac
		h.Step()
		if h.pc != c.want {
			t.Errorf("%02d: Stepped machine across JGT with ac = %d. pc = %d; wanted %d", i, h.ac, h.pc, c.want)
		}
	}
}

func TestJLT(t *testing.T) {
	cases := []struct {
		pc   int
		inst int
		ac   int
		want int // Value of PC after test
	}{
		{0, 3010, 0, 1},   // Move PC to 1 (next instruction) 10 because AC not < 0
		{0, 3010, -1, 10}, // Set PC to 10 because AC < 0
	}

	for i, c := range cases {
		h := NewMachine()
		h.pc = c.pc
		h.mem[c.pc] = c.inst
		h.ac = c.ac
		h.Step()
		if h.pc != c.want {
			t.Errorf("%02d: Stepped machine across JLT with ac = %d. pc = %d; wanted %d", i, h.ac, h.pc, c.want)
		}
	}
}

func TestJMP(t *testing.T) {
	cases := []struct {
		pc   int
		inst int
		want int // Value of PC after test
	}{
		{0, 5010, 10}, // Set PC to 10 unconditionally
		{0, 5020, 20}, // Set PC to 20 unconditionally
	}

	for i, c := range cases {
		h := NewMachine()
		h.pc = c.pc
		h.mem[c.pc] = c.inst
		h.Step()
		if h.pc != c.want {
			t.Errorf("%02d: Stepped machine across JMP. pc = %d; wanted %d", i, h.pc, c.want)
		}
	}
}

func TestJLE(t *testing.T) {
	cases := []struct {
		pc   int
		inst int
		ac   int
		want int // Value of PC after test
	}{
		{0, 6010, 0, 10},  // Set PC to 10, AC <= 0
		{0, 6010, -1, 10}, // Set PC to 10, AC <= 0
		{0, 6020, 1, 1},   // Advance PC to 1, AC > 0
	}

	for i, c := range cases {
		h := NewMachine()
		h.pc = c.pc
		h.mem[c.pc] = c.inst
		h.ac = c.ac
		h.Step()
		if h.pc != c.want {
			t.Errorf("%02d: Stepped machine across JLE. pc = %d; wanted %d", i, h.pc, c.want)
		}
	}
}

func TestJNE(t *testing.T) {
	cases := []struct {
		pc   int
		inst int
		ac   int
		want int // Value of PC after test
	}{
		{0, 7010, 1, 10},  // Set PC to 10, AC != 0
		{0, 7010, -1, 10}, // Set PC to 10, AC != 0
		{0, 7020, 0, 1},   // Advance PC to 1, AC == 0
	}

	for i, c := range cases {
		h := NewMachine()
		h.pc = c.pc
		h.mem[c.pc] = c.inst
		h.ac = c.ac
		h.Step()
		if h.pc != c.want {
			t.Errorf("%02d: Stepped machine across JLE. pc = %d; wanted %d", i, h.pc, c.want)
		}
	}
}

func TestLAC(t *testing.T) {
	cases := []struct {
		inst int
		want int // Value of AC after test
	}{
		{10000, 10000}, // Load self into AC
		{10010, 0},     // Load addr 10 (00000, by default) into AC
	}

	for i, c := range cases {
		h := NewMachine()
		h.mem[h.pc] = c.inst
		h.Step()
		if h.ac != c.want {
			t.Errorf("%02d: Stepped machine across LAC. ac = %05d; wanted %05d", i, h.ac, c.want)
		}
	}
}

func TestPAC(t *testing.T) {
	cases := []struct {
		ac   int
		inst int
	}{
		{5, 11001},     // Store AC to address 1, expected value is 5
		{6, 11000},     // Store AC to address 0, expected value is 6
		{99999, 11000}, // Store AC to address 0, expected value is 99999
	}

	for i, c := range cases {
		h := NewMachine()
		h.ac = c.ac
		h.mem[h.pc] = c.inst
		h.Step()
		if h.mem[c.inst%11000] != c.ac {
			t.Errorf("%02d: Stepped machine across PAC. mem = %05d; wanted %05d", i, h.mem[c.inst%11000], h.ac)
		}
	}
}

func TestLMQ(t *testing.T) {
	cases := []struct {
		inst int
		want int // Value of MQ after test
	}{
		{12000, 12000}, // Load self into MQ
		{12010, 0},     // Load addr 10 (00000, by default) into MQ
	}

	for i, c := range cases {
		h := NewMachine()
		h.mem[h.pc] = c.inst
		h.Step()
		if h.mq != c.want {
			t.Errorf("%02d: Stepped machine across LMQ. ac = %05d; wanted %05d", i, h.mq, c.want)
		}
	}
}

func TestPMQ(t *testing.T) {
	cases := []struct {
		mq   int
		inst int
	}{
		{5, 13001},     // Store MQ to address 1, expected value is 5
		{6, 13000},     // Store MQ to address 0, expected value is 6
		{99999, 13000}, // Store MQ to address 0, expected value is 99999
	}

	for i, c := range cases {
		h := NewMachine()
		h.mq = c.mq
		h.mem[h.pc] = c.inst
		h.Step()
		if h.mem[c.inst%13000] != c.mq {
			t.Errorf("%02d: Stepped machine across PMQ. mem = %05d; wanted %05d", i, h.mem[c.inst%13000], h.mq)
		}
	}
}

func TestADD(t *testing.T) {
	cases := []struct {
		ac   int
		inst int
		want int // AC after the test
	}{
		{0, 20000, 20000},  // Add mem[0] to AC, expect 20000
		{6, 20000, 20006},  // Add mem[0] to AC, expect 20006
		{-6, 20000, 19994}, // Add mem[0] to AC, expect 19994
		{6, 20001, 6},      // Add mem[1] to AC, expect 6
		{-6, 20001, -6},    // Add mem[1] to AC, expect -6
	}

	for i, c := range cases {
		h := NewMachine()
		h.ac = c.ac
		h.mem[h.pc] = c.inst
		h.Step()
		if h.ac != c.want {
			t.Errorf("%02d: Stepped machine across ADD. ac = %05d; wanted %05d", i, h.ac, c.want)
		}
	}
}

func TestSUB(t *testing.T) {
	cases := []struct {
		ac   int
		inst int
		want int // AC after the test
	}{
		{0, 21000, -21000},  // Sub mem[0] from AC, expect -21000
		{6, 21000, -20994},  // Sub mem[0] from AC, expect -20994
		{-6, 21000, -21006}, // Sub mem[0] from AC, expect -21006
		{6, 21001, 6},       // Sub mem[1] from AC, expect 6
		{-6, 21001, -6},     // Sub mem[1] from AC, expect -6
	}

	for i, c := range cases {
		h := NewMachine()
		h.ac = c.ac
		h.mem[h.pc] = c.inst
		h.Step()
		if h.ac != c.want {
			t.Errorf("%02d: Stepped machine across SUB. ac = %05d; wanted %05d", i, h.ac, c.want)
		}
	}
}

func TestMUL(t *testing.T) {
	cases := []struct {
		mq    int
		addr1 int
		inst  int
		want  int // MQ after the test
	}{
		{0, 1, 22001, 0},           // Multiply MQ by mem[1], expect 0
		{1, 1, 22001, 1},           // Multiply MQ by mem[1], expect 1
		{3, 4, 22001, 12},          // Multiply MQ by mem[1], expect 12
		{33333, 4, 22001, 99999},   // Multiply MQ by mem[1], expect (bounded) 99999
		{33333, -4, 22001, -99999}, // Multiply MQ by mem[1], expect (bounded) -99999
	}

	for i, c := range cases {
		h := NewMachine()
		h.mq = c.mq
		h.mem[h.pc] = c.inst
		h.mem[1] = c.addr1
		h.Step()
		if h.mq != c.want {
			t.Errorf("%02d: Stepped machine across MUL. mq = %05d; wanted %05d", i, h.mq, c.want)
		}
	}
}

func TestDIV(t *testing.T) {
	cases := []struct {
		mq     int
		addr1  int
		inst   int
		wantMQ int      // MQ after the test
		wantAC int      // AC after the test
		state  CPUState // CPUState after the test
	}{
		{0, 1, 23001, 0, 0, CPUok},
		{12, 3, 23001, 4, 0, CPUok},
		{12, 9, 23001, 1, 3, CPUok},
		{-12, 9, 23001, -1, -3, CPUok},
		{12, 0, 23001, 12, 0, CPUdivzero},
	}

	for i, c := range cases {
		h := NewMachine()
		h.mq = c.mq
		h.mem[h.pc] = c.inst
		h.mem[1] = c.addr1
		h.Step()
		if h.state != c.state {
			t.Errorf("%02d: Got cpustate = %q; Wanted %q", h.state, c.state)
		}

		if h.mq != c.wantMQ || h.ac != c.wantAC {
			t.Errorf("%02d: Stepped machine across DIV. mq/ac = %05d/%05d; wanted %05d/%05d", i, h.mq, h.ac, c.wantMQ, c.wantAC)
		}
	}
}

func TestGet(t *testing.T) {
	cases := []struct {
		input int
		want  int
	}{
		{42, 42},
		{-1000000, -99999},
		{99999, 99999},
		{-3, -3},
	}

	orig := Input // Store a reference to original Getter
	defer func() {
		Input = orig
	}()

	for i, c := range cases {
		Input = func() int { return c.input }
		h := NewMachine()
		h.mem[h.pc] = 30000 // Read to address 0
		h.Step()
		if h.mem[0] != c.want {
			t.Errorf("%02d: GET = %05d; Wanted %05d", i, h.mem[0], c)
		}
	}
}

func TestPut(t *testing.T) {
	cases := []struct {
		inst int
		want int
	}{
		{31000, 31000},
		{31001, 0},
	}

	orig := Output // Store a reference to original Getter
	defer func() {
		Output = orig
	}()

	for i, c := range cases {
		var got int
		Output = func(i int) { got = i }
		h := NewMachine()
		h.mem[0] = c.inst
		h.Step()
		if got != c.want {
			t.Errorf("%02d: PUT % 06d = %d, want %d", i, h.mem[0], got, c.want)
		}
	}
}
