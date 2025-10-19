package sim

// Memory map (teaching-simple)
const (
	UARTBase uint32 = 0x1000_0000
)

// Bus routes byte/word requests to RAM or UART.
//   - Read32/Write32 are little-endian, alignment required for 32-bit.
type Bus struct {
	RAM  *RAM
	UART *UART
}

func NewBus(ram *RAM, uart *UART) *Bus { return &Bus{RAM: ram, UART: uart} }

func (b *Bus) Read8(addr uint32) (uint8, bool) {
	// RAM: 0 .. RAM.Size()-1
	if addr < b.RAM.Size() {
		return b.RAM.Read8(addr)
	}
	// UART region
	if addr >= UARTBase && addr < UARTBase+UARTSize {
		return b.UART.Read8(addr - UARTBase)
	}
	return 0, false
}

func (b *Bus) Write8(addr uint32, v uint8) bool {
	if addr < b.RAM.Size() {
		return b.RAM.Write8(addr, v)
	}
	if addr >= UARTBase && addr < UARTBase+UARTSize {
		return b.UART.Write8(addr-UARTBase, v)
	}
	return false
}

func (b *Bus) Read32(addr uint32) (uint32, bool) {
	if addr&3 != 0 {
		return 0, false // require alignment
	}
	var v uint32
	for i := 0; i < 4; i++ {
		bb, ok := b.Read8(addr + uint32(i))
		if !ok {
			return 0, false
		}
		v |= uint32(bb) << (8 * i) // little-endian
	}
	return v, true
}

func (b *Bus) Write32(addr uint32, v uint32) bool {
	if addr&3 != 0 {
		return false
	}
	for i := 0; i < 4; i++ {
		if !b.Write8(addr+uint32(i), uint8(v>>(8*i))) {
			return false
		}
	}
	return true
}
