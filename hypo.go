/* Binary hypo implements the Hypothetical Machine. See README.md for
details of the machine specification. */
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

var (
	progFile = flag.String("program", "", "Path to the hypo program to run, for use as the default program to be loaded.")
)

type menuAction struct {
	desc   string // The text to diplay.
	action func() // The function to run for the action.
}

func loadProg(h *Machine) {
	fmt.Printf("Program file path (default: %q): ", *progFile)
	r := bufio.NewReader(os.Stdin)
	input, err := r.ReadString('\n')
	if err != nil {
		log.Printf("Error reading program path: %v", err)
		return
	}

	input = input[:len(input)-1]

	var pf *os.File
	switch input {
	case "":
		pf, err = os.Open(*progFile)
	default:
		pf, err = os.Open(input)
	}

	if err != nil {
		fmt.Printf("Error opening program file: %v\n", err)
		return
	}

	h.LoadProgram(pf)
}

func bios(h *Machine) {
	menu := map[string]menuAction{
		"?": menuAction{"display this help text", nil},
		"g": menuAction{"run program to halt state (go!)", h.Run},
		"h": menuAction{"display this help text", nil},
		"l": menuAction{"load program from file", func() { loadProg(h) }},
		"m": menuAction{"display memory", h.DumpMem},
		"q": menuAction{"quit hypo", func() { fmt.Println("Bye!"); os.Exit(0) }},
		"r": menuAction{"dump register contents", h.DumpRegs},
		"s": menuAction{"step program forward by one instruction", h.Step},
		"t": menuAction{"toggle execution tracing", h.ToggleTrace},
		"x": menuAction{"dump all machine state", h.DumpState},
		"z": menuAction{"reboot/reset the CPU state", h.ResetCPU},
	}

	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Hypothetical Machine BiOS (enter h for help)")
		fmt.Printf("Enter command: ")
		input, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Fake up a real "q" entry so we handle eof the same way as a normal
				// exit.
				input = "q\n"
			} else {
				// This will be handled with the unknown case below.
				input = "ijustmashedthekeyboard\n"
			}
		}

		command := input[:len(input)-1]
		if item, ok := menu[command]; ok {
			switch item.action {
			case nil:
				// Show the help text if the menu key has no action
				keys := []string{}
				for k := range menu {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Printf("%s: %s\n", k, menu[k].desc)
				}
			default:
				item.action()
			}
		} else {
			fmt.Println("Command not implemented.")
		}
	}
}

func main() {
	flag.Parse()
	hm := NewMachine()
	bios(hm)
}
