package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

func evalA(i, j int) int {
	return (i+j)*(i+j+1)/2 + i + 1
}

func times(v, u []float64, n int) {
	for i := 0; i < n; i++ {
		a := 0.0
		for j := 0; j < n; j++ {
			a += u[j] / float64(evalA(i, j))
		}
		v[i] = a
	}
}

func timesTrans(v, u []float64, n int) {
	for i := 0; i < n; i++ {
		a := 0.0
		for j := 0; j < n; j++ {
			a += u[j] / float64(evalA(j, i))
		}
		v[i] = a
	}
}

func aTimesTransp(v, u []float64, n int) {
	x := make([]float64, n)
	times(x, u, n)
	timesTrans(v, x, n)
}

func main() {
	n := 100
	if len(os.Args) > 1 {
		n, _ = strconv.Atoi(os.Args[1])
	}

	u := make([]float64, n)
	v := make([]float64, n)
	for i := range u {
		u[i] = 1
		v[i] = 1
	}

	for range 10 {
		aTimesTransp(v, u, n)
		aTimesTransp(u, v, n)
	}

	var vBv, vv float64
	for i, vi := range v {
		vBv += u[i] * vi
		vv += vi * vi
	}
	fmt.Printf("%0.9f\n", math.Sqrt(vBv/vv))
}
