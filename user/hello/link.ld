/* link.ld — single-RAM image at 0x00000000 for rv32 sim teaching */
ENTRY(_start)

MEMORY {
  RAM (rwx) : ORIGIN = 0x00000000, LENGTH = 64K
}

SECTIONS
{
  /* Code & const data */
  . = ORIGIN(RAM);
  .text : ALIGN(4)
  {
    KEEP(*(.text.start))     /* startup (_start) goes first */
    *(.text*)                /* program code */
    *(.rodata*)              /* const data/strings */
  } > RAM

  /* Writable data */
  . = ALIGN(4);
  .data : ALIGN(4)
  {
    *(.sdata*) *(.data*)
  } > RAM

  /* Zero-initialized data (not present in file; loader will zero it) */
  . = ALIGN(4);
  .bss (NOLOAD) : ALIGN(4)
  {
    __bss_start = .;
    *(.sbss*) *(.bss*) *(COMMON)
    __bss_end = .;
  } > RAM

  /* Simple downward-growing stack at the top of RAM */
  . = ALIGN(4);
  __stack_top = ORIGIN(RAM) + LENGTH(RAM);
}
