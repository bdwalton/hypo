# Hypo - The "Hypothetical Machine"

The machine is based on the implementation for the Commodore 64 which
was part of the [Commodore Public Domain
Series](https://www.c64-wiki.de/wiki/Commodore_64_Software#Lernprogramme). The
original is archived as part of the [C64 Preservation
Project](http://c64preservation.com/files/cbmpd/).

The "Hypothetical Machine" has 50 memory addresses, 2 user registers
and a program counter. Both instructions and data are stored in the
same memory space. Valid values for a memory address are [-99999,
99999].

## Machine Initialization

At initializtion time, the program counter (PC) is set to 0, as are
the AC (accumulator) and MQ (multiplier quotient) registers. All
memory addresses are initilized as 0 which conveniently maps to the
HLT (halt) instruction's opcode (00000).

## Opcodes

Programs in Hypo are written in the form of instructions that are read
into memory. An instruction is a 2 digit opcode followed by a 0 and
then a 2 digit memory address (eg: 000 - 049).  The following opcodes
are available:

*  00xxx: Stop. (HLT)
*  01xxx: Goto xxx if the AC is zero. (JEQ)
*  02xxx: Goto xxx if the AC is positive. (JGT)
*  03xxx: Goto xxx if the AC is negative. (JLT)
*  05xxx: Goto xxx. (JMP)
*  06xxx: Goto xxx if the AC is negative or zero. (JLE)
*  07xxx: Goto xxx if the AC is not zero. (JNE)
*  10xxx: Load the accumulator (AC) with the contents of location
   xxx. (LAC)
*  11xxx: Store the AC to location xxx. (PAC)
*  12xxx: Load the multiplier-quotient (MQ) with the contents of
   xxx. (LMQ)
*  13xxx: Store the MQ to location xxx. (PMQ)
*  20xxx: Add the content of xxx to the AC. (ADD)
*  21xxx: Subtract the contents of xxx from the AC. (SUB)
*  22xxx: Multiply the MQ by the contents of location xxx. (MUL)
*  23xxx: Divide the MQ by the contents of location xxx. The remainder
   is in AC. (DIV)
*  30xxx: Input a value to the location xxx. (GET)
*  31xxx: Output the value in location xxx. (PUT)

All operations bound the results of calculations to valid numeric
values [-99999, 99999].

## CPU States

At machine initialization time, the CPU is set to CPUok which
indicates that computation may occur. The followiing states are
possible:

*  CPUok: OK - computation may occur
*  CPUbadinst: If PC points at a memory address containing an invalid
   instruction, the CPU will enter this state and no further execution
   will occur.
*  CPUbadaddr: If an instruction refers to an invalid memory address,
   the CPU will enter this state and no further execution will occur.
*  CPUdivzero: If an instruction attempts to divide the MQ register by
   zero, the CPU will enter this state and no further execution will
   occur.
*  CPUhalt: If a HLT (opcode 00000) instruction is encountered, the
   CPU will enter this state and no further execution will occur.
   
## Writing Programs

The hypo machine is able to load program files from disk. These are
written in the form:

addr: value // optional comments after value
addr: value // more comments

An addr is numeric and must represent a valid memory address ([0,
49]). This represents the location where the value will be stored. It
is possible to use the same addr value multiple times, with the final
entry taking precedence. This allows for convenient code commenting if
desired.

The values specified must be valid numbers in the range [-99999,
99999].

### A sample program

The following program is an infinite loop that will print 99103
forever.

```
0: 0     // This entry will store a HLT (halt) instruction in the
0: 0     // first memory address and makes for a handy way to comment
0: 0     // the program file as subsequent instructions will overwrite
0: 0     // these entries.
0: 31002 // PUT content of memory address 2
1: 05000 // GOTO 0 (infinite loop)
2: 99103 // The value that address 0 will output.
```

### Included Programs

For demonstration, there are a few sample programs located in the
examples/ directory. A brief description of each is below.

*  simple_quine.hypo: This is the simplest possible quine program that
   can be implemented in hypo machine language.
*  quine.hypo: A much more interesting quine program.
*  fibonacci.hypo: A fibonacci sequence generator that prompts for the
   number of elements to generate and then outputs that many elements.
*  max.hypo: Ask for two numbers and print the larger one. (Negatives
   not handled cleanly.)
