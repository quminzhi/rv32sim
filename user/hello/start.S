/* start.S — minimal startup for rv32i */
    .section .text.start
    .globl _start
_start:
    /* Set stack pointer to top of RAM */
    la      sp, __stack_top

    /* Clear .bss: for(t = __bss_start; t < __bss_end; t += 4) *t = 0; */
    la      t0, __bss_start
    la      t1, __bss_end
1:  beq     t0, t1, 2f
    sw      zero, 0(t0)
    addi    t0, t0, 4
    blt     t0, t1, 1b
2:

    /* Call main() */
    call    main

    /* Clean halt for the simulator (your CPU treats ECALL as halt) */
    ecall
    j       .   /* in case ECALL is ignored, loop here */
