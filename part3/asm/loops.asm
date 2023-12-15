; On Linux x64 ABI:
; rdi = 1st arg,
; rsi = 2nd arg,
; rax = 1st return register.
;
; Reference:
; https://gitlab.com/x86-psABIs/x86-64-ABI/-/jobs/artifacts/master/raw/x86-64-ABI/abi.pdf?job=build

global MovAllBytes
global Nop3x1AllBytes
global CmpAllBytes
global DecAllBytes
global Nop1x3AllBytes
global Nop1x9AllBytes

section .text

MovAllBytes:
  xor rax, rax
.loop:
  mov [rsi + rax], al
  inc rax
  cmp rax, rdi
  jb .loop
  ret

Nop3x1AllBytes:
  xor rax, rax
.loop:
  db 0x0f, 0x1f, 0x00 ; This is apparently a 3-byte NOP
  inc rax
  cmp rax, rdi
  jb .loop
  ret

CmpAllBytes:
  xor rax, rax
.loop:
  inc rax
  cmp rax, rdi
  jb .loop
  ret

DecAllBytes:
.loop:
  dec rdi
  jnz .loop
  ret

Nop1x3AllBytes:
  xor rax, rax
.loop:
  nop
  nop
  nop
  inc rax
  cmp rax, rdi
  jb .loop
  ret

Nop1x9AllBytes:
  xor rax, rax
.loop:
  nop
  nop
  nop
  nop
  nop
  nop
  nop
  nop
  nop
  inc rax
  cmp rax, rdi
  jb .loop
  ret
