package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestDisassemble(t *testing.T) {
	outputFile := path.Join("testdata", "tmp_output_file")
	outputFileAsm := outputFile + ".asm"
	defer os.Remove(outputFile)
	defer os.Remove(outputFileAsm)
	for _, inputFile := range []string{
		"listing_0037_single_register_mov",
		"listing_0038_many_register_mov",
		"listing_0039_more_movs",
		"listing_0040_challenge_movs",
		"listing_0041_add_sub_cmp_jnz",
		"listing_0043_immediate_movs",
		"listing_0044_register_movs",
		"listing_0046_add_sub_cmp",
		"listing_0047_challenge_flags",
		"listing_0048_ip_register",
		"listing_0049_conditional_jumps",
		"listing_0050_challenge_jumps",
		"listing_0051_memory_mov",
		"listing_0052_memory_add_loop",
		"listing_0053_add_loop_challenge",
	} {
		inputFile = path.Join("testdata", inputFile)
		reassembleAndCompare(t, inputFile, outputFile)
	}
}

func reassembleAndCompare(t *testing.T, inputFile, outputFile string) {
	outputFileAsm := outputFile + ".asm"
	buf, err := ioutil.ReadFile(inputFile)
	ck(err)
	output, err := os.Create(outputFileAsm)
	ck(err)
	defer output.Close()
	Disassemble(output, buf)
	ck(output.Close())
	ck(nasm(outputFileAsm))
	ref, err := ioutil.ReadFile(outputFile)
	ck(err)
	if !bytes.Equal(buf, ref) {
		t.Errorf("Listing %s did not reassemble to expected output", inputFile)
	}
}

func TestSimulate(t *testing.T) {
	for _, tc := range []struct {
		file     string
		expected Registers
	}{
		{"listing_0043_immediate_movs", Registers{1, 2, 3, 4, 5, 6, 7, 8, 24}},
		{"listing_0044_register_movs", Registers{4, 3, 2, 1, 1, 2, 3, 4, 28}},
		{"listing_0046_add_sub_cmp", Registers{
			RegBx:    57602,
			RegCx:    3841,
			RegSp:    998,
			RegIp:    24,
			RegFlags: FlagP | FlagZ,
		}},
		{"listing_0047_challenge_flags", Registers{
			RegBx:    40101,
			RegDx:    10,
			RegSp:    99,
			RegBp:    98,
			RegIp:    44,
			RegFlags: FlagC | FlagA | FlagP | FlagS,
		}},
		{"listing_0048_ip_register", Registers{
			RegBx:    2000,
			RegCx:    64736,
			RegIp:    14,
			RegFlags: FlagC | FlagS,
		}},
		{"listing_0049_conditional_jumps", Registers{
			RegBx:    1030,
			RegIp:    14,
			RegFlags: FlagP | FlagZ,
		}},
		{"listing_0050_challenge_jumps", Registers{
			RegAx:    13,
			RegBx:    65531,
			RegIp:    28,
			RegFlags: FlagC | FlagA | FlagS,
		}},
		{"listing_0051_memory_mov", Registers{
			RegBx: 1,
			RegCx: 2,
			RegDx: 10,
			RegBp: 4,
			RegIp: 48,
		}},
		{"listing_0052_memory_add_loop", Registers{
			RegBx:    6,
			RegCx:    4,
			RegDx:    6,
			RegBp:    1000,
			RegSi:    6,
			RegIp:    35,
			RegFlags: FlagP | FlagZ,
		}},
		{"listing_0053_add_loop_challenge", Registers{
			RegBx:    6,
			RegDx:    6,
			RegBp:    998,
			RegIp:    33,
			RegFlags: FlagP | FlagZ,
		}},
	} {
		buf, err := ioutil.ReadFile(path.Join("testdata", tc.file))
		ck(err)
		regs := Simulate(io.Discard, buf)
		if regs != tc.expected {
			t.Errorf("Listing %s failed, got\n\n%s\nbut expected\n\n%s\n", tc.file, regs.Summary(), tc.expected.Summary())
		}
	}
}
