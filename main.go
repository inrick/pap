package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

const DefaultInputFile = "listing_0050_challenge_jumps"

func main() {
	if err := run(); err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}

func ck(err error) {
	if err != nil {
		panic(err)
	}
}

func run() error {
	log.SetFlags(0)
	var inputFile string
	var simulate bool
	flag.StringVar(&inputFile, "file", DefaultInputFile, "input file to parse")
	flag.BoolVar(&simulate, "exec", false, "simulate execution")
	flag.Parse()

	log.Printf("Processing %q", inputFile)

	inputFile = path.Join("testdata", inputFile)
	if err := nasm(inputFile + ".asm"); err != nil {
		return fmt.Errorf("could not assemble %s: %w", inputFile, err)
	}
	buf, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if !simulate {
		Disassemble(os.Stdout, buf)
	} else {
		_ = Simulate(os.Stdout, buf)
	}

	return nil
}

func nasm(file string) error {
	return exec.Command("nasm", file).Run()
}

func DecodeInstruction(buf []byte, ip int) (in Instruction, advance int) {
	b1, b2 := buf[ip], buf[ip+1]
	o := operation(b1, b2)
	switch {
	case o.kind == KindRmToFromRm:
		D, W := (b1>>1)&1, b1&1
		MOD, REG, RM := b2>>6, (b2>>3)&0b111, b2&0b111
		var dst, src OperandType
		dst, advance = RmOperand(buf, ip, MOD, RM, W)
		src = register(REG, W)
		if D == 1 {
			dst, src = src, dst
		}
		in = Instruction{o.op, FromUnsized(dst, src)}
	case o.kind == KindImmToRm:
		// Immediate to register/memory
		D, W := (b1>>1)&1, b1&1
		MOD, RM := b2>>6, b2&0b111
		var dst, src Operand
		dstOp, offset := RmOperand(buf, ip, MOD, RM, W)
		dst.op = dstOp
		if MOD == 0b11 {
			dst.size = SizeNone
		} else {
			dst.size = SizeFrom(W)
		}
		switch o.op {
		case OpMov:
			src.size = SizeFrom(W)
			src.op = OperandSigned(buf[ip+offset : ip+offset+1+int(W)])
			advance = offset + 1 + int(W)
		default:
			src.size = SizeNone
			if D == 1 {
				src.op = OperandSigned(buf[ip+offset : ip+offset+1])
				advance = offset + 1
			} else {
				src.op = OperandUnsigned(buf[ip+offset : ip+offset+1+int(W)])
				advance = offset + 1 + int(W)
			}
		}
		in = Instruction{o.op, []Operand{dst, src}}
	case o.kind == KindMemToFromAcc && o.op == OpMov:
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
	case o.kind == KindImmToReg && o.op == OpMov:
		// Immediate to register
		W, REG := (b1>>3)&1, b1&0b111
		dst := register(REG, W)
		src := OperandSigned(buf[ip+1 : ip+2+int(W)])
		in = Instruction{o.op, FromUnsized(dst, src)}
		advance = 2 + int(W)
	case o.kind == KindImmToAcc:
		// Immediate to accumulator
		W := b1 & 1
		width := [...]RegisterWidth{WidthLo, WidthFull}[W]
		dst := OperandReg{RegAx, width}
		src := OperandSigned(buf[ip+1 : ip+2+int(W)])
		in = Instruction{o.op, FromUnsized(dst, src)}
		advance = 2 + int(W)
	case o.kind == KindCondJmp:
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

func RmOperand(buf []byte, ip int, MOD, RM, W byte) (OperandType, int) {
	if MOD == 0b11 {
		// Register to register
		return register(RM, W), 2
	}
	var advance int
	var disp OperandDisplacement
	disp.kind = GetDisplacementKind(RM)
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
	return disp, advance
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

func Simulate(w io.Writer, buf []byte) Registers {
	var regs, regsPrev Registers
	//var mem [1 << 20]byte
	for int(regs[RegIp]) < len(buf) {
		in, advance := DecodeInstruction(buf, int(regs[RegIp]))
		regsPrev = regs
		regs[RegIp] += uint16(advance)
		switch in.op {
		case OpMov:
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				imm := uint16(immediate(&regs, in.operands[1]))
				regs[dst.name], _ = applyOp(OpMov, dst.width, regs[dst.name], imm)
			}
		case OpAdd:
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				imm := immediate(&regs, in.operands[1])
				var aflags ArithmeticFlags
				regs[dst.name], aflags = applyOp(OpAdd, dst.width, regs[dst.name], imm)
				regs.ProcessFlags(dst.width, regs[dst.name], aflags)
			}
		case OpSub, OpCmp:
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				imm := immediate(&regs, in.operands[1])
				out, aflags := applyOp(OpSub, dst.width, regs[dst.name], imm)
				regs.ProcessFlags(dst.width, out, aflags)
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
		fmt.Fprintf(w, "%-20s ;", in)
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
	return regs
}

// TODO: Rethink this, and consider how afAdd and afSub can both be shortened
// and treated similarly.
func applyOp(op Op, width RegisterWidth, a, b uint16) (uint16, ArithmeticFlags) {
	var f func(uint16, uint16) uint16
	var af ArithmeticFlags
	switch op {
	case OpAdd:
		f, af = opAdd, afAdd(width, a, b)
	case OpSub:
		f, af = opSub, afSub(width, a, b)
	case OpMov:
		f = opMov
	default:
		panic(op)
	}
	switch width {
	case WidthFull:
		return f(a, b), af
	case WidthLo:
		return f(a&0xff, b) | a&0xff00, af
	case WidthHi:
		return f((a>>8)&0xff, b)<<8 | a&0xff, af
	}
	panic(width)
}

func opMov(a, b uint16) uint16 { return b }
func opAdd(a, b uint16) uint16 { return a + b }
func opSub(a, b uint16) uint16 { return a - b }

func afAdd(w RegisterWidth, a, b uint16) ArithmeticFlags {
	var overflow uint32
	var flags ArithmeticFlags
	switch w {
	case WidthFull:
		overflow = 1<<16 - 1
		ia, ib := int16(a), int16(b)
		if ia < 0 && ib < 0 && 0 < ia+ib || 0 < ia && 0 < ib && ia+ib < 0 {
			flags.O = true
		}
	case WidthLo, WidthHi:
		overflow = 1<<8 - 1
		switch w {
		case WidthLo:
			a, b = a&0xff, b&0xff
		case WidthHi:
			a, b = (a>>8)&0xff, b&0xff
		}
		ia, ib := int8(a), int8(b)
		if ia < 0 && ib < 0 && 0 < ia+ib || 0 < ia && 0 < ib && ia+ib < 0 {
			flags.O = true
		}
	}
	if uint32(a)+uint32(b) > overflow {
		flags.C = true
	}
	if uint8(a&0xf)+uint8(b&0xf) > (1<<4 - 1) {
		flags.A = true
	}
	return flags
}

func afSub(w RegisterWidth, a, b uint16) ArithmeticFlags {
	var overflow uint32
	var flags ArithmeticFlags
	switch w {
	case WidthFull:
		overflow = 1<<16 - 1
		ia, ib := int16(a), int16(b)
		if ia < 0 && 0 < ib && 0 < ia-ib || 0 < ia && ib < 0 && ia-ib < 0 {
			flags.O = true
		}
	case WidthLo, WidthHi:
		overflow = 1<<8 - 1
		switch w {
		case WidthLo:
			a = a & 0xff
		case WidthHi:
			a = (a >> 8) & 0xff
		}
		ia, ib := int16(int8(a)), int16(b)
		if ia < 0 && 0 < ib && 0 < int8(ia-ib) || 0 < ia && ib < 0 && int8(ia-ib) < 0 {
			flags.O = true
		}
	}
	if uint32(a)-uint32(b) > overflow {
		flags.C = true
	}
	if uint8(a&0xf)-uint8(b&0xf) > (1<<4 - 1) {
		flags.A = true
	}
	return flags
}

func immediate(regs *Registers, reg Operand) uint16 {
	switch x := reg.op.(type) {
	case OperandImm:
		return uint16(x)
	case OperandImmU:
		return uint16(x)
	case OperandReg:
		return uint16(regs[x.name])
	default:
		panic(x)
	}
}
