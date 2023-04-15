package main

import (
	"bytes"
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
	ins := Decode(buf)
	PrintInstructions(output, ins)
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
		{"listing_0043_immediate_movs", Registers{1, 2, 3, 4, 5, 6, 7, 8}},
		{"listing_0044_register_movs", Registers{4, 3, 2, 1, 1, 2, 3, 4}},
		{"listing_0046_add_sub_cmp", Registers{
			RegBx:    57602,
			RegCx:    3841,
			RegSp:    998,
			RegFlags: FlagP | FlagZ,
		}},
		{"listing_0047_challenge_flags", Registers{
			RegBx:    40101,
			RegDx:    10,
			RegSp:    99,
			RegBp:    98,
			RegFlags: FlagP | FlagS, /* | FlagC | FlagA */
		}},
	} {
		buf, err := ioutil.ReadFile(path.Join("testdata", tc.file))
		ck(err)
		ins := Decode(buf)
		regs := Simulate(ins)
		if regs != tc.expected {
			t.Errorf("Listing %s failed, got\n%s\nbut expected\n\n%s\n", tc.file, &regs, &tc.expected)
		}
	}
}
