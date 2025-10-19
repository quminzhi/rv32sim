# -----------------------------------------
# Teaching Makefile: build a tiny RV32I program
# - Compiles and links: ./user/hello/{start.S, hello.c, link.ld}
# - Outputs to:         ./build/hello/{hello.elf, hello.bin, hello.lst, hello.map}
# - Toolchain:          riscv64-unknown-elf-*
# -----------------------------------------

# --- Toolchain prefix (override on the command line if needed) ---
# Example: make RISCV_PREFIX=riscv-none-elf
RISCV_PREFIX ?= riscv64-unknown-elf

CC      := $(RISCV_PREFIX)-gcc
OBJCOPY := $(RISCV_PREFIX)-objcopy
OBJDUMP := $(RISCV_PREFIX)-objdump
SIZE    := $(RISCV_PREFIX)-size

# --- Architecture/ABI (RV32I base ISA, no compressed, no mul/div) ---
ARCH    := rv32i
ABI     := ilp32

# --- Paths ---
HELLO_DIR := user/hello
BUILD_DIR := build/hello

# --- Inputs / Outputs ---
SOURCES  := $(HELLO_DIR)/start.S $(HELLO_DIR)/hello.c
LINKERLD := $(HELLO_DIR)/link.ld
ELF      := $(BUILD_DIR)/hello.elf
BIN      := $(BUILD_DIR)/hello.bin
LST      := $(BUILD_DIR)/hello.lst
MAP      := $(BUILD_DIR)/hello.map

# --- Flags (teaching friendly: no libc, freestanding, small-data disabled) ---
CFLAGS  := -march=$(ARCH) -mabi=$(ABI) \
           -O2 -g \
           -ffreestanding -fno-builtin -fno-pic -fno-plt \
           -msmall-data-limit=0 \
           -Wall -Wextra
# Notes:
# -ffreestanding/-fno-builtin : we’re bare-metal; no implicit libc/builtins
# -msmall-data-limit=0        : avoid GP/sdata model (we don’t init gp)
# -fno-pic/-fno-plt           : no position-independent code in this tiny bare-metal demo

LDFLAGS := -nostdlib -nostartfiles \
           -Wl,-T,$(LINKERLD) \
           -Wl,--gc-sections \
           -Wl,-Map,$(MAP) \
           -Wl,--build-id=none
# Notes:
# -nostdlib/-nostartfiles : we provide our own startup (start.S)
# --gc-sections           : drop unused code/data
# Map file                : great for teaching where things landed in memory

OBJDUMP_FLAGS := -d -S -M no-aliases     # mixed source/asm, expand pseudoinstrs

# --- Default target: build everything helpful for class ---
.PHONY: all run
all: $(ELF) $(BIN) $(LST) size

run:
	go run ./cmd/runelf -trace

# Build directory
$(BUILD_DIR):
	@mkdir -p $@

# Link the final ELF
$(ELF): $(SOURCES) $(LINKERLD) | $(BUILD_DIR)
	$(CC) $(CFLAGS) $(LDFLAGS) $(SOURCES) -o $@
	@echo "ELF built: $@"

# Flat binary (ROM/RAM image) extracted from ELF
$(BIN): $(ELF)
	$(OBJCOPY) -O binary $< $@
	@echo "BIN built: $@"

# Annotated disassembly listing (great for teaching)
$(LST): $(ELF)
	$(OBJDUMP) $(OBJDUMP_FLAGS) $< > $@
	@echo "Listing: $@"

# Print memory usage (sections, hex & decimal)
.PHONY: size
size: $(ELF)
	$(SIZE) -A -x $<
	@echo "Map file: $(MAP)"

# Run Go tests (your simulator/unit tests)
.PHONY: test
test:
	go test ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	@echo "Cleaned $(BUILD_DIR)"

# Convenience: rebuild from scratch
.PHONY: rebuild
rebuild: clean all
