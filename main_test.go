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
	input, err := ioutil.ReadFile(inputFile)
	ck(err)
	output, err := os.Create(outputFileAsm)
	ck(err)
	defer output.Close()
	ins := decode(input)
	PrintInstructions(output, ins)
	ck(output.Close())
	ck(nasm(outputFile))
	ref, err := ioutil.ReadFile(outputFile)
	ck(err)
	if !bytes.Equal(input, ref) {
		t.Errorf("Listing %s did not reassemble to expected output", inputFile)
	}
}
