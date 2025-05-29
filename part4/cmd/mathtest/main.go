package main

import (
	"os"
	"part4/mathalt"
)

func main() {
	var arg string
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {
	case "asin":
		mathalt.TestAsinFunctions()
	case "sin":
		mathalt.TestSinFunctions()
	case "printsin":
		mathalt.PrintSinTaylorCoeffs(32)
	default:
		mathalt.TestFunctions()
	}
}
