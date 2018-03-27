/* This file contains the implementation of hypo, the Hypothetical
Machine. See README.md for a description of the machinee.  */
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

// inBounds validates whether an address is valid or not.
func inBounds(addr int) bool {
	return addr >= 0 && addr < memSize
}

// boundsCap implements integer bounds capping. The hypo machine
// allows for values in the range [-99999, 99999].
func boundsCap(i int) int {
	if i > 99999 {
		return 99999
	}

	if i < -99999 {
		return -99999
	}

	return i
}

// The Hypo machine has 50 memory addresses
const memSize = 50

var (
	loadErrBadFile  = errors.New("Invalid program file")
	loadErrBadLine  = errors.New("Invalid line in program")
	loadErrBadAddr  = errors.New("Invalid memory address - can't load data there")
	loadErrBadValue = errors.New("Invalid value - couldn't parse")
)

// CPUState indicates whether the Hypo machine can continue operating
// or needs to be reset for various reasons.
type CPUState int

const (
	CPUok      = iota // Program execution may continue
	CPUbadinst = iota // Invalid instruction
	CPUbadaddr = iota // Invalid memory reference
	CPUdivzero = iota // Divide by zero
	CPUhalt    = iota // Halted
)

func (s CPUState) String() string {
	switch s {
	case CPUok:
		return "CPUok"
	case CPUbadaddr:
		return "CPUbadaddr"
	case CPUdivzero:
		return "CPUdivzero"
	case CPUbadinst:
		return "CPUbadinst"
	case CPUhalt:
		return "CPUhalt"
	default:
		return ("Unknown CPU state.")
	}
}

type Instruction struct {
	op   string
	addr int
}

// String ensures that Instruction implements the Stringer interface
// for easy display.
func (i Instruction) String() string {
	return fmt.Sprintf("%s %03d", i.op, i.addr)
}

// The op codes with textual equivalents.
var ops = map[int]string{
	0:  "HLT", // HALT
	1:  "JEQ", // Jump to addr if AC == 0
	2:  "JGT", // Jump to addr if AC > 0
	3:  "JLT", // Jump to addr if AC < 0
	5:  "JMP", // Jump to addr
	6:  "JLE", // Jump to addr if AC <= 0
	7:  "JNE", // Jump to addr if AC != 0
	10: "LAC", // Load addr into AC, the accumulator
	11: "PAC", // Put AC to addr
	12: "LMQ", // Load addr into MQ, the multipler-quotient
	13: "PMQ", // Put MQ to addr
	20: "ADD", // Add addr to AC
	21: "SUB", // Subtract addr from AC
	22: "MUL", // Multiply MQ by the content of addr
	23: "DIV", // Divide MQ by the content of addr. The remainder is in AC.
	30: "GET", // Read input to addr
	31: "PUT", // Output addr
}

// A Getter is a generic function that return an integer value from
// the user.
type Getter func() int

// A default Getter
var Input = func() int {
	fmt.Printf("Enter a numeric value: ")
	f, err := os.Open("/dev/stdin")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var num int
	for {
		if n, err := fmt.Fscanf(f, "%d", &num); n < 1 || err != nil {
			fmt.Println("Error reading input. Try again.")
			continue
		}
		break
	}
	return num
}

// A Putter is a generic function that accepts an integer and
// (presumably) displays it to the user somehow.
type Putter func(int)

// A default Putter
var Output = func(i int) {
	fmt.Printf("% 06d\n", i)
}

// Machine represents all register, memory, state and I/O objects
// required to implement a "Hypothetical Machine".
type Machine struct {
	mem    [memSize]int // Instructions and data aren't distinguishable by anything other than a valid opcode and address when "parsed".
	pc     int          // program counter
	ac     int          // accumulator
	mq     int          // mulitplier quotient
	state  CPUState     // The program should stop
	input  Getter       // Our ears
	output Putter       // Out mouth
	trace  bool         // If true, instructions will be displayed at execution time.
}

// NewMachine returns an initialized machine. For now, no special
// initialization is required, but we provide it as the default way to
// create a machine anyway. It wires up Input and Output for
// i/o. Library users may override those at need.
func NewMachine() *Machine {
	return &Machine{input: Input, output: Output}
}

// getInstruction returns the instruction stored at addr if addr is in
// bounds and represents a valid instruction. If the opcode or target
// address of the instruction is out of bounds, the returned CPUState
// will be set appropriately.
func (h *Machine) getInstruction(addr int) (Instruction, CPUState) {
	if !inBounds(addr) {
		return Instruction{"UNK", 0}, CPUbadinst
	}

	d := h.mem[addr]
	op := d / 1000
	a := d % 1000
	o, ok := ops[op]
	if !ok {
		return Instruction{"UNK", a}, CPUbadinst
	} else if !inBounds(a) {
		return Instruction{o, a}, CPUbadaddr
	}

	return Instruction{o, a}, CPUok
}

func (h *Machine) LoadProgram(r io.Reader) error {
	// Ensure the machine is halted until we signal a clean load below.
	h.state = CPUhalt
	// Reset machine memory so we can read directly into it below.
	h.mem = [memSize]int{}

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		// Code files are instructions with optional, free-form comments
		// following them
		re := regexp.MustCompile("^(\\d+):\\s*(-*\\d+)(\\s.*)*$")
		m := re.FindStringSubmatch(line)
		if m == nil {
			log.Printf("Invalid line: %q", line)
			return loadErrBadLine
		}

		a, err := strconv.Atoi(m[1]) // The address for this instruction to be stored
		if err != nil {
			log.Printf("Invalid memory address: %q", m[1])
			return loadErrBadAddr
		}
		if a < 0 || a >= memSize {
			log.Printf("Out of range memory address: %d", a)
			return loadErrBadAddr
		}

		v, err := strconv.Atoi(m[2])
		if err != nil {
			log.Printf("Invalid data value: %q", m[2])
			return loadErrBadValue
		}
		h.mem[a] = boundsCap(v)
	}

	if err := s.Err(); err != nil {
		log.Printf("LoadProgram Error: %v", err)
		return loadErrBadFile
	}

	// We didn't return an error, so set the CPU state to ok.
	h.state = CPUok
	fmt.Println("Program loaded successfully.")
	return nil
}

// Step executes the instruction in the memory address referenced by
// PC (program counter). If it is valid, the machine state is alerted
// accordingly, otherwise the CPU state is transitioned to an
// appropriate !ok value.
func (h *Machine) Step() {
	i, cs := h.getInstruction(h.pc)
	if cs != CPUok {
		h.state = cs
		return
	}

	if h.trace {
		fmt.Println(i)
	}

	h.pc += 1
	switch i.op {
	case "HLT":
		h.state = CPUhalt
	case "JEQ":
		if h.ac == 0 {
			h.pc = i.addr
		}
	case "JGT":
		if h.ac > 0 {
			h.pc = i.addr
		}
	case "JLT":
		if h.ac < 0 {
			h.pc = i.addr
		}
	case "JMP":
		h.pc = i.addr
	case "JLE":
		if h.ac <= 0 {
			h.pc = i.addr
		}
	case "JNE":
		if h.ac != 0 {
			h.pc = i.addr
		}
	case "LAC":
		h.ac = h.mem[i.addr]
	case "PAC":
		h.mem[i.addr] = h.ac
	case "LMQ":
		h.mq = h.mem[i.addr]
	case "PMQ":
		h.mem[i.addr] = h.mq
	case "ADD":
		h.ac = boundsCap(h.ac + h.mem[i.addr])
	case "SUB":
		h.ac = boundsCap(h.ac - h.mem[i.addr])
	case "MUL":
		h.mq = boundsCap(h.mq * h.mem[i.addr])
	case "DIV":
		if h.mem[i.addr] == 0 {
			h.state = CPUdivzero
			return
		}
		h.ac = h.mq % h.mem[i.addr]
		h.mq = h.mq / h.mem[i.addr]
	case "GET":
		h.mem[i.addr] = boundsCap(h.input())
	case "PUT":
		h.output(h.mem[i.addr])
	default:
		h.state = CPUbadinst
	}
}

// Halted returns true if the system state is such that execution
// cannot continue. Currently this is equivalent to !CPUok.
func (h *Machine) Halted() bool {
	return h.state != CPUok
}

// Reset restores the CPU to initial state (all registers 0, program
// counter 0 and CPU state OK).
func (h *Machine) ResetCPU() {
	h.ac = 0
	h.mq = 0
	h.pc = 0
	h.state = CPUok
	fmt.Println("CPU state reset.")
}

// DumpMem prints memory content to stdout.
func (h *Machine) DumpMem() {
	for i, c := range h.mem {
		fmt.Printf("%02d: % 06d  ", i, c)
		if i%5 == 4 {
			fmt.Printf("\n")
		}
	}
}

// DumpRegs prints register content to stdout.
func (h *Machine) DumpRegs() {
	fmt.Printf("PC: %02d  AC: % 06d  MQ: % 06d\n", h.pc, h.ac, h.mq)
}

// DumpState prints memory, register and cpu state to stdout.
func (h *Machine) DumpState() {
	fmt.Println("Memory:")
	h.DumpMem()
	fmt.Println()
	fmt.Println("Registers:")
	h.DumpRegs()
	fmt.Println()
	fmt.Printf("CPU State: %s\n\n", h.state)
}

// Run executes instructions until the CPU enters a halted state.
func (h *Machine) Run() {
	for {
		h.Step()
		if h.Halted() {
			fmt.Printf("Program terminated with: %q\n", h.state)
			break
		}
	}
}

func (h *Machine) ToggleTrace() {
	h.trace = !h.trace
	fmt.Println("Tracing mode:", h.trace)
}
