package sim

import "testing"

func TestUART_WriteAndReadViaBus(t *testing.T) {
	ram := NewRAM(1024)
	uart := NewUART(nil)
	bus := NewBus(ram, uart)

	// Write "A", "B", "C" to UART DATA (offset 0)
	if !bus.Write8(UARTBase, 'A') || !bus.Write8(UARTBase, 'B') || !bus.Write8(UARTBase, 'C') {
		t.Fatalf("Write8 to UART failed")
	}
	if uart.String() != "ABC" {
		t.Fatalf("UART buffer = %q want %q", uart.String(), "ABC")
	}

	// Read non-failing (returns 0, true)
	if b, ok := bus.Read8(UARTBase); !ok || b != 0 {
		t.Fatalf("UART Read8 = (%d,%v), want (0,true)", b, ok)
	}
}
