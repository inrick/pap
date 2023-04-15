package main

import (
	"fmt"
	"math/bits"
	"strings"
)

type Op uint32

const (
	OpMov Op = iota
	OpAdd
	OpSub
	OpCmp
	OpJe
	OpJl
	OpJle
	OpJb
	OpJbe
	OpJp
	OpJo
	OpJs
	OpJne
	OpJnl
	OpJnle
	OpJnb
	OpJnbe
	OpJnp
	OpJno
	OpJns
	OpLoop
	OpLoopz
	OpLoopnz
	OpJcxz
)

var opStrs = [...]string{
	"mov",
	"add",
	"sub",
	"cmp",
	"je",
	"jl",
	"jle",
	"jb",
	"jbe",
	"jp",
	"jo",
	"js",
	"jne",
	"jnl",
	"jnle",
	"jnb",
	"jnbe",
	"jnp",
	"jno",
	"jns",
	"loop",
	"loopz",
	"loopnz",
	"jcxz",
}

func (o Op) String() string {
	return opStrs[o]
}

type OpKind uint32

const (
	KindRmToFromRm OpKind = iota
	KindMemToFromAcc
	KindImmToRm
	KindImmToReg
	KindImmToAcc
	KindCondJmp
)

type OpDescr struct {
	kind OpKind
	op   Op
}

func operation(b1, b2 byte) OpDescr {
	switch b1 >> 2 {
	case 0:
		return OpDescr{KindRmToFromRm, OpAdd}
	case 0b000001:
		return OpDescr{KindImmToAcc, OpAdd}
	case 0b100010:
		return OpDescr{KindRmToFromRm, OpMov}
	case 0b110001:
		return OpDescr{KindImmToRm, OpMov}
	case 0b101000:
		return OpDescr{KindMemToFromAcc, OpMov}
	case 0b001010:
		return OpDescr{KindRmToFromRm, OpSub}
	case 0b001011:
		return OpDescr{KindImmToAcc, OpSub}
	case 0b001110:
		return OpDescr{KindRmToFromRm, OpCmp}
	case 0b001111:
		return OpDescr{KindImmToAcc, OpCmp}
	case 0b100000:
		switch (b2 >> 3) & 0b111 {
		case 0b000:
			return OpDescr{KindImmToRm, OpAdd}
		case 0b101:
			return OpDescr{KindImmToRm, OpSub}
		case 0b111:
			return OpDescr{KindImmToRm, OpCmp}
		}
	default:
		if b1>>4 == 0b1011 {
			return OpDescr{KindImmToReg, OpMov}
		} else {
			switch b1 {
			case 0b01110100:
				return OpDescr{KindCondJmp, OpJe}
			case 0b01111100:
				return OpDescr{KindCondJmp, OpJl}
			case 0b01111110:
				return OpDescr{KindCondJmp, OpJle}
			case 0b01110010:
				return OpDescr{KindCondJmp, OpJb}
			case 0b01110110:
				return OpDescr{KindCondJmp, OpJbe}
			case 0b01111010:
				return OpDescr{KindCondJmp, OpJp}
			case 0b01110000:
				return OpDescr{KindCondJmp, OpJo}
			case 0b01111000:
				return OpDescr{KindCondJmp, OpJs}
			case 0b01110101:
				return OpDescr{KindCondJmp, OpJne}
			case 0b01111101:
				return OpDescr{KindCondJmp, OpJnl}
			case 0b01111111:
				return OpDescr{KindCondJmp, OpJnle}
			case 0b01110011:
				return OpDescr{KindCondJmp, OpJnb}
			case 0b01110111:
				return OpDescr{KindCondJmp, OpJnbe}
			case 0b01111011:
				return OpDescr{KindCondJmp, OpJnp}
			case 0b01110001:
				return OpDescr{KindCondJmp, OpJno}
			case 0b01111001:
				return OpDescr{KindCondJmp, OpJns}
			case 0b11100010:
				return OpDescr{KindCondJmp, OpLoop}
			case 0b11100001:
				return OpDescr{KindCondJmp, OpLoopz}
			case 0b11100000:
				return OpDescr{KindCondJmp, OpLoopnz}
			case 0b11100011:
				return OpDescr{KindCondJmp, OpJcxz}
			}
		}
	}
	panic(fmt.Sprintf("unimplemented instruction: %08b %08b", b1, b2))
}

type Register uint32

const (
	RegAx Register = iota
	RegBx
	RegCx
	RegDx
	RegSp
	RegBp
	RegSi
	RegDi
	RegFlags
	RegCount
)

// Flags are described on page 22 of the manual.
const (
	FlagC uint16 = 1      // Carry
	FlagP uint16 = 1 << 2 // Parity
	FlagA uint16 = 1 << 4 // Auxiliary Carry
	FlagZ uint16 = 1 << 6 // Zero
	FlagS uint16 = 1 << 7 // Sign
)

func FlagString(f uint16) string {
	switch f {
	case FlagC:
		return "C"
	case FlagP:
		return "P"
	case FlagA:
		return "A"
	case FlagZ:
		return "Z"
	case FlagS:
		return "S"
	}
	panic(nil)
}

var regFlags = [...]uint16{
	FlagC, FlagP, FlagA, FlagZ, FlagS,
}

type RegisterWidth uint32

const (
	WidthFull RegisterWidth = iota
	WidthLo
	WidthHi
)

type Registers [RegCount]uint16

func (rr *Registers) IsSet(flag uint16) bool {
	return rr[RegFlags]&flag > 0
}

func (rr *Registers) Flag(cond bool, flag uint16) {
	if cond {
		rr[RegFlags] |= flag
	} else {
		rr[RegFlags] &= ^flag
	}
}

// ProcessFlags toggles all the implemented flags based on the given width and
// value. Note that it is not possible to just take the target register and
// pick the value from there, since some operations (cmp) do not write the
// value back to the registers.
func (rr *Registers) ProcessFlags(width RegisterWidth, value uint16) {
	switch width {
	case WidthFull:
	case WidthLo:
		value = (value & 0xff) << 8
	case WidthHi:
		value = value & 0xff
	}
	rr.Flag(value>>15 > 0, FlagS)
	rr.Flag(bits.OnesCount16(value)%2 == 0, FlagP)
	rr.Flag(value == 0, FlagZ)
}

func (rr *Registers) String() string {
	var sb strings.Builder
	for i := RegAx; i <= RegDi; i++ {
		fmt.Fprintf(&sb, "%s=%d\n", OperandReg{i, WidthFull}, rr[i])
	}
	fmt.Fprintf(&sb, "Flags=")
	for _, flag := range regFlags {
		if rr.IsSet(flag) {
			fmt.Fprintf(&sb, "%s", FlagString(flag))
		}
	}
	fmt.Fprintln(&sb)
	return sb.String()
}

type Operand struct {
	size SizeMark
	op   OperandType
}

type (
	OperandType interface{ OperandType() }
	OperandReg  struct {
		name  Register
		width RegisterWidth
	}
	OperandImm          int16
	OperandImmU         uint16
	OperandDisplacement struct {
		kind DisplacementKind
		imm  OperandImm
	}
)

func (_ OperandReg) OperandType()          {}
func (_ OperandImm) OperandType()          {}
func (_ OperandImmU) OperandType()         {}
func (_ OperandDisplacement) OperandType() {}

type SizeMark uint32

const (
	SizeNone SizeMark = iota
	SizeByte
	SizeWord
)

var sizeMarkStrs = [...]string{
	"", "byte", "word",
}

func FromUnsized(ops ...OperandType) []Operand {
	oo := make([]Operand, len(ops))
	for i := range oo {
		oo[i] = Operand{SizeNone, ops[i]}
	}
	return oo
}

func (o Operand) String() string {
	var sb strings.Builder
	if o.size != SizeNone {
		fmt.Fprintf(&sb, "%s ", o.size)
	}
	fmt.Fprintf(&sb, "%v", o.op)
	return sb.String()
}

func (s SizeMark) String() string {
	return sizeMarkStrs[s]
}

func SizeFrom(w byte) SizeMark {
	return [...]SizeMark{SizeByte, SizeWord}[w]
}

var regStrsFull = [...][3]string{
	{"ax", "al", "ah"},
	{"bx", "bl", "bh"},
	{"cx", "cl", "ch"},
	{"dx", "dl", "dh"},
	{"sp", "", ""},
	{"bp", "", ""},
	{"si", "", ""},
	{"di", "", ""},
}

func (r OperandReg) String() string {
	return regStrsFull[r.name][r.width]
}

// Note the unexpected order of the first four registers: AX, CX, DX, BX.
var regVal = [...]OperandReg{
	{RegAx, WidthLo}, {RegCx, WidthLo}, {RegDx, WidthLo}, {RegBx, WidthLo},
	{RegAx, WidthHi}, {RegCx, WidthHi}, {RegDx, WidthHi}, {RegBx, WidthHi},
	{RegAx, WidthFull}, {RegCx, WidthFull}, {RegDx, WidthFull}, {RegBx, WidthFull},
	{RegSp, WidthFull}, {RegBp, WidthFull}, {RegSi, WidthFull}, {RegDi, WidthFull},
}

func register(reg, w byte) OperandReg {
	return regVal[w<<3|reg]
}

type DisplacementKind uint32

const (
	DispBxSi DisplacementKind = iota
	DispBxDi
	DispBpSi
	DispBpDi
	DispSi
	DispDi
	DispBp
	DispBx
	DispEA
)

var dispKindStrs = [...]string{
	"bx+si", "bx+di", "bp+si", "bp+di", "si", "di", "bp", "bx",
}

func GetDisplacementKind(rm byte) DisplacementKind {
	return DisplacementKind(rm)
}

func (d OperandDisplacement) String() string {
	if d.kind == DispEA {
		return fmt.Sprintf("[%d]", uint16(d.imm))
	}
	return fmt.Sprintf("[%s%+d]", dispKindStrs[d.kind], d.imm)
}

func OperandSigned(bb []byte) OperandImm {
	var i int16
	switch len(bb) {
	case 1:
		i = int16(int8(bb[0]))
	case 2:
		i = int16(uint16(bb[0]) | uint16(bb[1])<<8)
	default:
		panic(len(bb))
	}
	return OperandImm(i)
}

func OperandUnsigned(bb []byte) OperandImmU {
	var u uint16
	switch len(bb) {
	case 1:
		u = uint16(bb[0])
	case 2:
		u = uint16(uint16(bb[0]) | uint16(bb[1])<<8)
	default:
		panic(len(bb))
	}
	return OperandImmU(u)
}

type Instruction struct {
	ip       int
	op       Op
	operands []Operand
}
