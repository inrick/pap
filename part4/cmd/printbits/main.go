package main

import (
	"fmt"
	"math"
	"part4/mathalt"
)

func main() {
	for _, val := range []struct {
		name   string
		values [][]float64
	}{
		{"SineRadiansC_MFTWP", mathalt.SineRadiansC_MFTWP},
		{"ArcsineRadiansC_MFTWP", mathalt.ArcsineRadiansC_MFTWP},
	} {
		fmt.Printf("var %s = {\n", val.name)
		for _, ff := range val.values {
			fmt.Print("    {")
			for j, f := range ff {
				if j != 0 {
					fmt.Print(", ")
				}
				bits := math.Float64bits(f)
				fmt.Printf("0x%x", bits)
			}
			fmt.Println("},")
		}
		fmt.Println("}")
	}
}
