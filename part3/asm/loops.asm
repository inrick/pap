; On Linux x64 ABI:
; rdi = 1st arg,
; rsi = 2nd arg,
; rdx = 3rd arg,
; rcx = 4th arg,
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
global Read_4x2
global Read_8x2
global Read_16x2
global Read_32x2
global ReadSuccessiveSizes
global ReadSuccessiveSizesNonPow2
global ReadStrided_32x2
global WriteTemporal
global WriteNonTemporal

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

Read_4x2:
  xor rax, rax
  align 64
.loop:
  mov r8d, [rsi]
  mov r8d, [rsi + 4]
  add rax, 8
  cmp rax, rdi
  jb .loop
  ret

Read_8x2:
  xor rax, rax
  align 64
.loop:
  mov r8, [rsi]
  mov r8, [rsi + 8]
  add rax, 16
  cmp rax, rdi
  jb .loop
  ret

Read_16x2:
  xor rax, rax
  align 64
.loop:
  vmovdqu xmm0, [rsi]
  vmovdqu xmm1, [rsi + 16]
  add rax, 32
  cmp rax, rdi
  jb .loop
  ret

Read_32x2:
  xor rax, rax
  align 64
.loop:
  vmovdqu ymm0, [rsi]
  vmovdqu ymm1, [rsi + 32]
  add rax, 64
  cmp rax, rdi
  jb .loop
  ret

ReadSuccessiveSizes:
  xor rax, rax
  xor r8, r8
  align 64
.loop:
  mov r9, rsi
  add r9, r8
  vmovdqu ymm0, [r9]
  vmovdqu ymm0, [r9 + 32]
  vmovdqu ymm0, [r9 + 64]
  vmovdqu ymm0, [r9 + 96]
  vmovdqu ymm0, [r9 + 128]
  vmovdqu ymm0, [r9 + 160]
  vmovdqu ymm0, [r9 + 192]
  vmovdqu ymm0, [r9 + 224]
  add rax, 256
  mov r8, rax
  and r8, rdx
  cmp rax, rdi
  jb .loop
  ret

ReadSuccessiveSizesNonPow2:
  ; Input arguments:
  ;
  ; rdi: total bytes to process
  ; rsi: buffer pointer
  ; rdx: chunk size

  xor rax, rax

  align 64
.loop:
  xor r8, r8

.inner:
  mov r9, rsi
  add r9, r8
  vmovdqu ymm0, [r9]
  vmovdqu ymm0, [r9 + 32]
  vmovdqu ymm0, [r9 + 64]
  vmovdqu ymm0, [r9 + 96]
  vmovdqu ymm0, [r9 + 128]
  vmovdqu ymm0, [r9 + 160]
  vmovdqu ymm0, [r9 + 192]
  vmovdqu ymm0, [r9 + 224]
  add r8, 256
  cmp r8, rdx
  jb .inner

  add rax, r8
  cmp rax, rdi
  jb .loop

  ret

ReadStrided_32x2:
  ; Input arguments:
  ;
  ; rdi: total bytes to process
  ; rsi: buffer pointer
  ; rdx: chunk size
  ; rcx: stride

  xor rax, rax

  align 64
.loop:
  mov r8, rdx
  mov r9, rsi

.inner:
  vmovdqu ymm0, [r9]
  vmovdqu ymm0, [r9 + 32]
  add r9, rcx
  sub r8, 64
  jnz .inner

  sub rdi, rdx
  jnz .loop

  ret

WriteTemporal:
  ; Input arguments:
  ;
  ; rdi: input buffer
  ; rsi: output buffer
  ; rdx: read size
  ; rcx: inner read size

  xor rax, rax

  mov r8, rdi ; inner loop ptr

  ; end of input buffer
  mov r9, rdi
  add r9, rcx

  ; end of output buffer
  mov r10, rsi
  add r10, rdx

  align 64
.loop:
  mov r8, rdi

.inner:
  vmovdqu ymm0, [r8]
  vmovdqu [rsi], ymm0
  vmovdqu ymm0, [r8 + 32]
  vmovdqu [rsi + 32], ymm0
  vmovdqu ymm0, [r8 + 64]
  vmovdqu [rsi + 64], ymm0
  vmovdqu ymm0, [r8 + 96]
  vmovdqu [rsi + 96], ymm0
  vmovdqu ymm0, [r8 + 128]
  vmovdqu [rsi + 128], ymm0
  vmovdqu ymm0, [r8 + 160]
  vmovdqu [rsi + 160], ymm0
  vmovdqu ymm0, [r8 + 192]
  vmovdqu [rsi + 192], ymm0
  vmovdqu ymm0, [r8 + 224]
  vmovdqu [rsi + 224], ymm0
  add r8, 256
  add rsi, 256
  jb .inner

  cmp rsi, r10
  jb .loop

  ret

WriteNonTemporal:
  ; Input arguments:
  ;
  ; rdi: input buffer
  ; rsi: output buffer
  ; rdx: read size
  ; rcx: inner read size

  xor rax, rax

  mov r8, rdi ; inner loop ptr

  ; end of input buffer
  mov r9, rdi
  add r9, rcx

  ; end of output buffer
  mov r10, rsi
  add r10, rdx

  align 64
.loop:
  mov r8, rdi

.inner:
  vmovdqu ymm0, [r8]
  vmovntdq [rsi], ymm0
  vmovdqu ymm0, [r8 + 32]
  vmovntdq [rsi + 32], ymm0
  vmovdqu ymm0, [r8 + 64]
  vmovntdq [rsi + 64], ymm0
  vmovdqu ymm0, [r8 + 96]
  vmovntdq [rsi + 96], ymm0
  vmovdqu ymm0, [r8 + 128]
  vmovntdq [rsi + 128], ymm0
  vmovdqu ymm0, [r8 + 160]
  vmovntdq [rsi + 160], ymm0
  vmovdqu ymm0, [r8 + 192]
  vmovntdq [rsi + 192], ymm0
  vmovdqu ymm0, [r8 + 224]
  vmovntdq [rsi + 224], ymm0
  add r8, 256
  add rsi, 256
  jb .inner

  cmp rsi, r10
  jb .loop

  ret
