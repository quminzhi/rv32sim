// hello.c — minimal user program for the rv32 simulator
#include <stdint.h>

#define UART_BASE 0x10000000u
static volatile uint8_t * const UART = (volatile uint8_t *)UART_BASE;

static inline void uart_putc(char c) {
    *UART = (uint8_t)c;           // MMIO: store to UART data register
}

static void puts_uart(const char *s) {
    while (*s) {
        uart_putc(*s++);
    }
}

int main(void) {
    puts_uart("Hello, RV32!\n");
    return 0; // startup code will ECALL after main returns
}
