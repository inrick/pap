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
global Read_x1
global Read_x2
global Read_x3
global Read_x4
global Write_x1
global Write_x2
global Write_x3
global Write_x4

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

Read_x1:
  align 64
.loop:
  mov rax, [rsi]
  sub rdi, 1
  jnle .loop
  ret

Read_x2:
  align 64
.loop:
  mov rax, [rsi]
  mov rax, [rsi]
  sub rdi, 2
  jnle .loop
  ret

Read_x3:
  align 64
.loop:
  mov rax, [rsi]
  mov rax, [rsi]
  mov rax, [rsi]
  sub rdi, 3
  jnle .loop
  ret

Read_x4:
  align 64
.loop:
  mov rax, [rsi]
  mov rax, [rsi]
  mov rax, [rsi]
  mov rax, [rsi]
  sub rdi, 4
  jnle .loop
  ret

Write_x1:
  align 64
.loop:
  mov [rsi], rax
  sub rdi, 1
  jnle .loop
  ret

Write_x2:
  align 64
.loop:
  mov [rsi], rax
  mov [rsi], rax
  sub rdi, 2
  jnle .loop
  ret

Write_x3:
  align 64
.loop:
  mov [rsi], rax
  mov [rsi], rax
  mov [rsi], rax
  sub rdi, 3
  jnle .loop
  ret

Write_x4:
  align 64
.loop:
  mov [rsi], rax
  mov [rsi], rax
  mov [rsi], rax
  mov [rsi], rax
  sub rdi, 4
  jnle .loop
  ret
