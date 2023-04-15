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
	var inputFile string
	flag.StringVar(&inputFile, "file", "listing_0047_challenge_flags", "input file to parse")
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

	ins := Decode(buf)
	PrintInstructions(os.Stdout, ins)
	regs := Simulate(ins)
	fmt.Println(&regs)

	return nil
}

func nasm(file string) error {
	cmd := exec.Command("nasm", file)
	return cmd.Run()
}

func Decode(buf []byte) []Instruction {
	var ins []Instruction
	for ip := 0; ip < len(buf); {
		b1, b2 := buf[ip], buf[ip+1]
		var advance int
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
			ins = append(ins, Instruction{ip, o.op, FromUnsized(dst, src)})
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
			ins = append(ins, Instruction{ip, o.op, []Operand{dst, src}})
		case o.kind == KindMemToFromAcc && o.op == OpMov:
			// Memory/accumulator to acumulator/memory
			D, W := (b1>>1)&1, b1&1
			disp := OperandDisplacement{DispEA, OperandSigned(buf[ip+1 : ip+3])}
			var width RegisterWidth
			if W == 0 {
				width = WidthLo
			} else {
				width = WidthFull
			}
			reg := OperandReg{RegAx, width}
			var dst, src OperandType
			if D == 0 {
				dst, src = reg, disp
			} else {
				dst, src = disp, reg
			}
			ins = append(ins, Instruction{ip, o.op, FromUnsized(dst, src)})
			advance = 3
		case o.kind == KindImmToReg && o.op == OpMov:
			// Immediate to register
			W, REG := (b1>>3)&1, b1&0b111
			dst := register(REG, W)
			src := OperandSigned(buf[ip+1 : ip+2+int(W)])
			ins = append(ins, Instruction{ip, o.op, FromUnsized(dst, src)})
			advance = 2 + int(W)
		case o.kind == KindImmToAcc:
			// Immediate to accumulator
			W := b1 & 1
			var width RegisterWidth
			if W == 0 {
				width = WidthLo
			} else {
				width = WidthFull
			}
			dst := OperandReg{RegAx, width}
			src := OperandSigned(buf[ip+1 : ip+2+int(W)])
			ins = append(ins, Instruction{ip, o.op, FromUnsized(dst, src)})
			advance = 2 + int(W)
		case o.kind == KindCondJmp:
			ipInc := OperandSigned(buf[ip+1 : ip+2])
			ins = append(ins, Instruction{ip, o.op, FromUnsized(ipInc)})
			advance = 2
		default:
			panic(o)
		}
		if advance == 0 {
			panic("instruction stream did not advance")
		}
		ip += advance
	}
	return ins
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

func PrintInstructions(w io.Writer, ins []Instruction) {
	fmt.Fprintln(w, "bits 16")
	fmt.Fprintln(w)
	for _, in := range ins {
		fmt.Fprintf(w, "%s ", in.op)
		for j, o := range in.operands {
			if j > 0 {
				fmt.Fprint(w, ", ")
			}
			if OpJe <= in.op && in.op <= OpJcxz {
				operand := o.op.(OperandImm)
				// Offset is relative to start of instruction and therefore needs +2.
				fmt.Fprintf(w, "$%+d", operand+2)
			} else {
				fmt.Fprintf(w, "%s", o)
			}
		}
		fmt.Fprintln(w)
	}
}

func Simulate(ins []Instruction) Registers {
	var regs Registers
	for _, in := range ins {
		switch in.op {
		case OpMov:
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				imm := uint16(immediate(&regs, in.operands[1]))
				regs[dst.name] = applyOp(opMov, dst.width, regs[dst.name], imm)
			}
		case OpAdd:
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				imm := immediate(&regs, in.operands[1])
				regs[dst.name] = applyOp(opAdd, dst.width, regs[dst.name], imm)
				regs.ProcessFlags(dst.width, regs[dst.name])
			}
		case OpSub, OpCmp:
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				imm := immediate(&regs, in.operands[1])
				out := applyOp(opSub, dst.width, regs[dst.name], imm)
				regs.ProcessFlags(dst.width, out)
				// Cmp is implemented like sub but does not write it's result.
				if in.op != OpCmp {
					regs[dst.name] = out
				}
			}
		}
	}
	return regs
}

func applyOp[T ~uint16](f func(T, T) T, width RegisterWidth, a, b T) T {
	switch width {
	case WidthFull:
		return f(a, b)
	case WidthLo:
		return f(a&0xff, b&0xff) | a&0xff<<8
	case WidthHi:
		return f(a&0xff, b&0xff)<<8 | a&0xff
	}
	panic(width)
}

func opMov(a, b uint16) uint16 { return b }
func opAdd(a, b uint16) uint16 { return a + b }
func opSub(a, b uint16) uint16 { return a - b }

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
