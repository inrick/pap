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
	flag.StringVar(&inputFile, "file", "listing_0044_register_movs", "input file to parse")
	flag.Parse()

	inputFile = path.Join("testdata", inputFile)
	if err := nasm(inputFile); err != nil {
		return fmt.Errorf("could not assemble %s: %w", inputFile, err)
	}
	buf, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}
	ins := decode(buf)
	PrintInstructions(os.Stdout, ins)
	Simulate(ins)

	return nil
}

func nasm(filename string) error {
	filename = filename + ".asm"
	cmd := exec.Command("nasm", filename)
	return cmd.Run()
}

func decode(buf []byte) []Instruction {
	var ins []Instruction
	for ip := 0; ip < len(buf); {
		b1, b2 := buf[ip], buf[ip+1]
		var advance int
		o := operation(b1, b2)
		switch {
		case o.kind == KindRmToFromRm:
			D, W := (b1>>1)&1, b1&1
			MOD, REG, RM := b2>>6, (b2>>3)&0b111, b2&0b111
			switch MOD {
			case 0b11:
				// Register to register
				reg := OperandReg(register(REG, W))
				regRm := OperandReg(register(RM, W))
				var dst, src OperandType
				if D == 0 {
					dst, src = regRm, reg
				} else {
					dst, src = reg, regRm
				}
				ins = append(ins, Instruction{ip, o.op, FromUnsized(dst, src)})
				advance = 2
			case 0b00, 0b01, 0b10:
				// Memory to/from register
				disp := OperandDisplacement{
					kind: GetDisplacementKind(RM),
				}
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
				reg := register(REG, W)
				var dst, src OperandType
				if D == 0 {
					dst, src = disp, reg
				} else {
					dst, src = reg, disp
				}
				ins = append(ins, Instruction{ip, o.op, FromUnsized(dst, src)})
			default:
				panic(MOD)
			}
		case o.kind == KindImmToRm:
			// Immediate to register/memory
			D, W := (b1>>1)&1, b1&1
			MOD, RM := b2>>6, b2&0b111
			disp := OperandDisplacement{
				kind: GetDisplacementKind(RM),
			}
			var offset int
			switch MOD {
			case 0b00:
				if RM == 0b110 {
					// Memory-mode, direct address
					disp.kind = DispEA
					disp.imm = OperandImm(OperandUnsigned(buf[ip+2 : ip+4]))
					offset = 4
				} else {
					// Memory-mode, no displacement
					offset = 2
				}
			case 0b01:
				// Memory-mode with 8-bit displacement
				disp.imm = OperandSigned(buf[ip+2 : ip+3])
				offset = 3
			case 0b10:
				// Memory-mode with 16-bit displacement
				disp.imm = OperandSigned(buf[ip+2 : ip+4])
				offset = 4
			case 0b11:
				// Register mode
				offset = 2
			}
			var dst, src Operand
			if o.op == OpMov {
				dst.size = SizeNone
				src.size = SizeFrom(W)
				src.op = OperandSigned(buf[ip+offset : ip+offset+1+int(W)])
				advance = offset + 1 + int(W)
			} else {
				if D == 1 {
					src.op = OperandSigned(buf[ip+offset : ip+offset+1])
					advance = offset + 1
				} else {
					src.op = OperandUnsigned(buf[ip+offset : ip+offset+1+int(W)])
					advance = offset + 1 + int(W)
				}
				if MOD != 0b11 {
					dst.size = SizeFrom(W)
				} else {
					dst.size = SizeNone
				}
				src.size = SizeNone
			}
			if MOD == 0b11 {
				dst.op = register(RM, W)
			} else {
				dst.op = disp
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

func Simulate(ins []Instruction) {
	var regs Registers
	for _, in := range ins {
		if in.op == OpMov {
			switch dst := in.operands[0].op.(type) {
			case OperandReg:
				switch src := in.operands[1].op.(type) {
				case OperandImm:
					regs[dst.name] = uint16(src)
				case OperandImmU:
					regs[dst.name] = uint16(src)
				case OperandReg:
					regs[dst.name] = regs[src.name]
				}
			}
		}
	}
	fmt.Printf("%v\n", regs)
}
