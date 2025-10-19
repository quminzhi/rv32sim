// cmd/runelf/main.go
//
// Tiny ELF runner for the teaching RV32 simulator.
// - Loads an ELF (built by user/hello) into RAM via sim.LoadELF32
// - Runs the CPU until it halts (ECALL) or a step limit is reached
// - Prints the UART output *after* execution to avoid interleaving
//
// NOTE: The import path "rv32sim/sim" assumes your go.mod has:  module rv32sim
//       If your module is named differently, change the import below accordingly.

package main

import (
	"flag"
	"fmt"
	"os"

	"rv32sim/sim"
)

func main() {
	elfPath := flag.String("elf", "build/hello/hello.elf", "path to ELF file to run")
	ramKB := flag.Uint("ramkb", 64, "RAM size in KiB")
	steps := flag.Int("steps", 500000, "max instructions to execute before giving up")
	trace := flag.Bool("trace", false, "enable CPU trace (disassembly) to stderr")
	flag.Parse()

	ram := sim.NewRAM(uint64(*ramKB) * 1024)

	// Buffer UART output during execution to avoid mixing with trace.
	// After the run, we print uart.String() in one go.
	uart := sim.NewUART(nil)

	bus := sim.NewBus(ram, uart)
	cpu := sim.NewCPU(bus)

	// Load ELF → copy PT_LOAD segments to RAM, zero bss, return entry PC.
	entry, err := sim.LoadELF32(*elfPath, ram)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ELF load error: %v\n", err)
		os.Exit(1)
	}
	cpu.PC = entry

	// Optional disassembly trace; recommend stderr to keep output clean.
	if *trace {
		cpu.Trace = true
		// If your CPU type exposes TraceOut, you can do:
		// cpu.TraceOut = os.Stderr
	}

	// Run until ECALL (Step returns false) or we hit the step limit.
	halted := false
	for i := 0; i < *steps; i++ {
		if !cpu.Step() {
			halted = true
			break
		}
	}

	if !halted {
		fmt.Fprintf(os.Stderr, "program did not halt within %d steps\n", *steps)
		os.Exit(2)
	}

	// Print UART output cleanly after the run.
	// This avoids interleaving with trace lines.
	out := uart.String()
	if len(out) > 0 && out[len(out)-1] != '\n' {
		// be nice: end with a newline for terminal readability
		out += "\n"
	}
	fmt.Print(out)
}
