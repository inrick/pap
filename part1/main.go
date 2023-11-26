package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/bits"
	"os"
	"os/exec"
	"path"
)

const DefaultInputFile = "listing_0055_challenge_rectangle"

func main() {
	if err := run(); err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}

func run() error {
	log.SetFlags(0)
	var inputFile string
	var simulate, assembleInput, dumpMem bool
	flag.StringVar(&inputFile, "file", DefaultInputFile, "input file to parse")
	flag.BoolVar(&simulate, "exec", false, "simulate execution")
	flag.BoolVar(&assembleInput, "assemble", false, "assemble input .asm file with nasm")
	flag.BoolVar(&dumpMem, "dump", false, "dump memory of simulation to mem.data")
	flag.Parse()

	log.Printf("Processing %q", inputFile)

	inputFile = path.Join("testdata", inputFile)
	if assembleInput {
		if err := nasm(inputFile + ".asm"); err != nil {
			return fmt.Errorf("could not assemble %s: %w", inputFile, err)
		}
	}
	buf, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if !simulate {
		Disassemble(os.Stdout, buf)
		return nil
	}

	_, mem := Simulate(os.Stdout, buf)
	if dumpMem {
		f, err := os.Create("mem.data")
		if err != nil {
			return err
		}
		defer f.Close()
		f.Write(mem[:])
	}

	return nil
}

func nasm(file string) error {
	return exec.Command("nasm", file).Run()
}

func DecodeInstruction(buf []byte, ip int) (in Instruction, advance int) {
	b1, b2 := buf[ip], buf[ip+1]
	o := operation(b1, b2)
	switch o.kind {
	case KindRmToFromRm:
		D, W := (b1>>1)&1, b1&1
		MOD, REG, RM := b2>>6, (b2>>3)&0b111, b2&0b111
		var dst, src Operand
		dst, advance = RmOperand(buf, ip, MOD, RM, W)
		src = Operand{SizeNone, register(REG, W)}
		if D == 1 {
			dst, src = src, dst
		}
		in = Instruction{o.op, []Operand{dst, src}}
	case KindImmToRm:
		// Immediate to register/memory
		D, W := (b1>>1)&1, b1&1
		MOD, RM := b2>>6, b2&0b111
		var dst, src Operand
		var offset int
		dst, offset = RmOperand(buf, ip, MOD, RM, W)
		// It's kind of weird putting an explicit size on an immediate, but it is
		// valid in the disassembly. It's just that it's also unnecessary since it
		// can be inferred from the immediate and the other operand. Change
		// src.size below between SizeNone and SizeFrom(W), either way the tests
		// still pass.
		src.size = SizeNone
		if o.op != OpMov && D == 1 {
			src.op = OperandSigned(buf[ip+offset : ip+offset+1])
			advance = offset + 1
		} else {
			src.op = OperandUnsigned(buf[ip+offset : ip+offset+1+int(W)])
			advance = offset + 1 + int(W)
		}
		in = Instruction{o.op, []Operand{dst, src}}
	case KindMemToFromAcc:
		// Memory/accumulator to acumulator/memory
		D, W := (b1>>1)&1, b1&1
		disp := OperandDisplacement{DispEA, OperandSigned(buf[ip+1 : ip+3])}
		width := [...]RegisterWidth{WidthLo, WidthFull}[W]
		reg := OperandReg{RegAx, width}
		var dst, src OperandType
		if D == 0 {
			dst, src = reg, disp
		} else {
			dst, src = disp, reg
		}
		in = Instruction{o.op, FromUnsized(dst, src)}
		advance = 3
	case KindImmToReg:
		// Immediate to register
		W, REG := (b1>>3)&1, b1&0b111
		dst := register(REG, W)
		src := OperandSigned(buf[ip+1 : ip+2+int(W)])
		in = Instruction{o.op, FromUnsized(dst, src)}
		advance = 2 + int(W)
	case KindImmToAcc:
		// Immediate to accumulator
		W := b1 & 1
		width := [...]RegisterWidth{WidthLo, WidthFull}[W]
		dst := OperandReg{RegAx, width}
		src := OperandSigned(buf[ip+1 : ip+2+int(W)])
		in = Instruction{o.op, FromUnsized(dst, src)}
		advance = 2 + int(W)
	case KindRmToSeg, KindSegToRm:
		MOD, SR, RM := b2>>6, (b2>>3)&0b11, b2&0b111
		if (b2>>5)&1 != 0 {
			panic("illegal instruction")
		}
		var dst, src Operand
		dst, advance = RmOperand(buf, ip, MOD, RM, 1)
		src = Operand{SizeNone, Segment(SR)}
		if o.kind == KindRmToSeg {
			dst, src = src, dst
		}
		in = Instruction{o.op, []Operand{dst, src}}
	case KindCondJmp:
		ipInc := OperandSigned(buf[ip+1 : ip+2])
		in = Instruction{o.op, FromUnsized(ipInc)}
		advance = 2
	default:
		panic(o)
	}
	if advance == 0 {
		panic("instruction stream did not advance")
	}
	return in, advance
}

func RmOperand(buf []byte, ip int, MOD, RM, W byte) (Operand, int) {
	if MOD == 0b11 {
		// Register to register
		return Operand{SizeNone, register(RM, W)}, 2
	}
	var advance int
	var disp OperandDisplacement
	disp.kind = DisplacementKind(RM)
	switch MOD {
	case 0b00:
		if RM == 0b110 {
			// Memory-mode, direct address
			disp.kind = DispEA
			disp.imm = OperandImm(OperandUnsigned(buf[ip+2 : ip+4]))
			advance = 4
		} else {
			// Memory-mode, no displacement
			advance = 2
		}
	case 0b01:
		// Memory-mode with 8-bit displacement
		disp.imm = OperandSigned(buf[ip+2 : ip+3])
		advance = 3
	case 0b10:
		// Memory-mode with 16-bit displacement
		disp.imm = OperandSigned(buf[ip+2 : ip+4])
		advance = 4
	}
	return Operand{SizeFrom(W), disp}, advance
}

func Disassemble(w io.Writer, buf []byte) {
	fmt.Fprintln(w, "bits 16")
	fmt.Fprintln(w)
	for ip := 0; ip < len(buf); {
		in, advance := DecodeInstruction(buf, ip)
		ip += advance
		fmt.Fprintln(w, in)
	}
}

func Simulate(w io.Writer, buf []byte) (Registers, *Memory) {
	var regs, regsPrev Registers
	var mem Memory
	for int(regs[RegIp]) < len(buf) {
		in, advance := DecodeInstruction(buf, int(regs[RegIp]))
		regsPrev = regs
		regs[RegIp] += uint16(advance)
		switch in.op {
		case OpMov:
			imm := uint16(immediate(&regs, &mem, in.operands[1]))
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				regs[dst.name], _ = applyOp(OpMov, dst.width, regs[dst.name], imm)
			case OperandDisplacement:
				offset := dispOffset(&regs, dst)
				switch in.operands[0].size {
				case SizeByte:
					mem[offset] = byte(imm)
				case SizeWord:
					mem[offset] = byte(imm)
					mem[offset+1] = byte(imm >> 8)
				default:
					panic(in.operands[0])
				}
			default:
				panic(dst)
			}
		case OpAdd:
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				imm := immediate(&regs, &mem, in.operands[1])
				regs[dst.name], regs[RegFlags] = applyOp(OpAdd, dst.width, regs[dst.name], imm)
			}
		case OpSub, OpCmp:
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				imm := immediate(&regs, &mem, in.operands[1])
				out, flags := applyOp(OpSub, dst.width, regs[dst.name], imm)
				regs[RegFlags] = flags
				// Cmp is implemented like sub but does not write it's result.
				if in.op != OpCmp {
					regs[dst.name] = out
				}
			}
		case OpJe:
			regs.JumpIf(regs.IsSet(FlagZ), in.operands[0].op)
		case OpJl:
			regs.JumpIf(regs.IsSet(FlagS) != regs.IsSet(FlagO), in.operands[0].op)
		case OpJle:
			regs.JumpIf(
				regs.IsSet(FlagZ) || (regs.IsSet(FlagS) != regs.IsSet(FlagO)),
				in.operands[0].op,
			)
		case OpJb:
			regs.JumpIf(regs.IsSet(FlagC), in.operands[0].op)
		case OpJbe:
			regs.JumpIf(regs.IsSet(FlagC|FlagZ), in.operands[0].op)
		case OpJp:
			regs.JumpIf(regs.IsSet(FlagP), in.operands[0].op)
		case OpJo:
			regs.JumpIf(regs.IsSet(FlagO), in.operands[0].op)
		case OpJs:
			regs.JumpIf(regs.IsSet(FlagS), in.operands[0].op)
		case OpJne:
			regs.JumpIf(!regs.IsSet(FlagZ), in.operands[0].op)
		case OpJnl:
			regs.JumpIf(regs.IsSet(FlagS) == regs.IsSet(FlagO), in.operands[0].op)
		case OpJnle:
			regs.JumpIf(
				!regs.IsSet(FlagZ) && (regs.IsSet(FlagS) == regs.IsSet(FlagO)),
				in.operands[0].op,
			)
		case OpJnb:
			regs.JumpIf(!regs.IsSet(FlagC), in.operands[0].op)
		case OpJnbe:
			regs.JumpIf(!regs.IsSet(FlagC) && !regs.IsSet(FlagZ), in.operands[0].op)
		case OpJnp:
			regs.JumpIf(!regs.IsSet(FlagP), in.operands[0].op)
		case OpJno:
			regs.JumpIf(!regs.IsSet(FlagO), in.operands[0].op)
		case OpJns:
			regs.JumpIf(!regs.IsSet(FlagS), in.operands[0].op)
		case OpLoop:
			// Loop instruction decrements cx but does not change any flags.
			regs[RegCx]--
			regs.JumpIf(regs[RegCx] != 0, in.operands[0].op)
		case OpLoopz:
			regs[RegCx]--
			regs.JumpIf(regs[RegCx] != 0 && regs.IsSet(FlagZ), in.operands[0].op)
		case OpLoopnz:
			regs[RegCx]--
			regs.JumpIf(regs[RegCx] != 0 && !regs.IsSet(FlagZ), in.operands[0].op)
		case OpJcxz:
			regs.JumpIf(regs[RegCx] == 0, in.operands[0].op)
		}
		// Print processed instruction
		fmt.Fprintf(w, "%s ;", in)
		// Write out state changes
		for r := RegAx; r < RegFlags; r++ {
			t0, t1 := regsPrev[r], regs[r]
			if t0 != t1 {
				fmt.Fprintf(w, " %s:0x%x->0x%x", OperandReg{r, WidthFull}, t0, t1)
			}
		}
		f0, f1 := regsPrev[RegFlags], regs[RegFlags]
		if f0 != f1 {
			fmt.Fprintf(w, " flags:%s->%s", FlagsString(f0), FlagsString(f1))
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, regs.Summary())
	return regs, &mem
}

// Applies the given operation (OpMov, OpAdd, OpSub) and returns the new
// register value as well as any flags. Note that the returned register value
// is a full register value: when operating on half registers the returned
// value will be the fully updated register value, with only the high or low
// bits modified.
func applyOp(op Op, width RegisterWidth, a, b uint16) (value uint16, flags Flags) {
	switch op {
	case OpMov:
		value = b
	case OpAdd, OpSub:
		value, flags = applyArithmetic(op, width, a, b)
	}
	// If operating at half width, pack the value appropriately.
	switch width {
	case WidthLo:
		value = value&0xff | a&0xff00
	case WidthHi:
		value = value<<8 | a&0xff
	}
	return value, flags
}

// Returns the value as well as any flags created by the operation (OpAdd,
// OpSub). The value of half register operations will be returned as a plain
// value, not packed together in the full register. For example, "add ah, 3"
// will return ah+3 no matter what is in al.
func applyArithmetic(op Op, w RegisterWidth, a, b uint16) (value uint16, flags Flags) {
	var carry uint32
	var signBit uint16
	switch w {
	case WidthFull:
		carry, signBit = 1<<16-1, 1<<15
		flags |= overflowFlag(op, int16(a), int16(b))
	case WidthLo, WidthHi:
		carry, signBit = 1<<8-1, 1<<7
		if w == WidthHi {
			a >>= 8
		}
		a &= 0xff
		flags |= overflowFlag(op, int8(a), int8(b))
	}
	// Calculate value of operation and any remaining flags.
	var valA uint8
	var valC uint32
	switch op {
	case OpAdd:
		value = a + b
		valA = uint8(a&0xf) + uint8(b&0xf)
		valC = uint32(a) + uint32(b)
	case OpSub:
		value = a - b
		valA = uint8(a&0xf) - uint8(b&0xf)
		valC = uint32(a) - uint32(b)
	}
	flags |= boolToInt(valA > 1<<4-1) * FlagA
	flags |= boolToInt(valC > carry) * FlagC
	flags |= boolToInt(value&signBit != 0) * FlagS
	flags |= boolToInt(value == 0) * FlagZ
	// Parity is only calculated on lower byte
	flags |= boolToInt(bits.OnesCount16(value&0xff)%2 == 0) * FlagP
	return value, flags
}

func overflowFlag[T int16 | int8](op Op, a, b T) Flags {
	ia, ib := T(a), T(b)
	if op == OpSub {
		ib = -ib
	}
	if ia < 0 && ib < 0 && 0 < ia+ib || 0 < ia && 0 < ib && ia+ib < 0 {
		return FlagO
	}
	return 0
}

func immediate(regs *Registers, mem *Memory, src Operand) uint16 {
	switch x := src.op.(type) {
	case OperandImm:
		return uint16(x)
	case OperandImmU:
		return uint16(x)
	case OperandReg:
		switch x.width {
		case WidthFull:
			return uint16(regs[x.name])
		case WidthLo:
			return uint16(regs[x.name] & 0xff)
		case WidthHi:
			return uint16((regs[x.name] >> 8) & 0xff)
		}
	case OperandDisplacement:
		offset := dispOffset(regs, x)
		switch src.size {
		case SizeByte:
			return uint16(mem[offset])
		case SizeWord:
			return uint16(mem[offset+1])<<8 | uint16(mem[offset])
		default:
			panic(src.size)
		}
	}
	panic(src)
}

func dispOffset(regs *Registers, d OperandDisplacement) int {
	switch d.kind {
	case DispBxSi:
		return int(regs[RegBx]) + int(regs[RegSi]) + int(d.imm)
	case DispBxDi:
		return int(regs[RegBx]) + int(regs[RegDi]) + int(d.imm)
	case DispBpSi:
		return int(regs[RegBp]) + int(regs[RegSi]) + int(d.imm)
	case DispBpDi:
		return int(regs[RegBp]) + int(regs[RegDi]) + int(d.imm)
	case DispSi:
		return int(regs[RegSi]) + int(d.imm)
	case DispDi:
		return int(regs[RegDi]) + int(d.imm)
	case DispBp:
		return int(regs[RegBp]) + int(d.imm)
	case DispBx:
		return int(regs[RegBx]) + int(d.imm)
	case DispEA:
		return int(uint16(d.imm))
	}
	panic(d)
}
