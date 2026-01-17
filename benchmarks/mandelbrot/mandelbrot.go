package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"strconv"
)

func main() {
	size := 200
	if len(os.Args) > 1 {
		if s, err := strconv.Atoi(os.Args[1]); err == nil {
			size = s
		}
	}
	size = (size + 7) / 8 * 8
	chunkSize := size / 8
	inv := 2.0 / float64(size)

	xloc := make([][8]float64, chunkSize)
	for i := 0; i < size; i++ {
		xloc[i/8][i%8] = float64(i)*inv - 1.5
	}

	fmt.Printf("P4\n%d %d\n", size, size)

	pixels := make([]byte, size*chunkSize)
	for chunkID := 0; chunkID < size; chunkID++ {
		ci := float64(chunkID)*inv - 1.0
		offset := chunkID * chunkSize
		for i := 0; i < chunkSize; i++ {
			r := mbrot8(&xloc[i], ci)
			if r > 0 {
				pixels[offset+i] = r
			}
		}
	}

	hasher := md5.New()
	hasher.Write(pixels)
	fmt.Printf("%x\n", hasher.Sum(nil))
}

func mbrot8(cr *[8]float64, civ float64) byte {
	var ci, zr, zi, tr, ti, absz, tmp [8]float64
	for i := 0; i < 8; i++ {
		ci[i] = civ
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			add(&zr, &zr, &tmp)
			mul(&tmp, &zi, &tmp)
			add(&tmp, &ci, &zi)

			minus(&tr, &ti, &tmp)
			add(&tmp, cr, &zr)

			mul(&zr, &zr, &tr)
			mul(&zi, &zi, &ti)
		}
		add(&tr, &ti, &absz)
		terminate := true
		for k := 0; k < 8; k++ {
			if absz[k] <= 4.0 {
				terminate = false
				break
			}
		}
		if terminate {
			return 0
		}
	}

	accu := byte(0)
	for i := 0; i < 8; i++ {
		if absz[i] <= 4.0 {
			accu |= byte(0x80) >> i
		}
	}
	return accu
}

func add(a, b, r *[8]float64) {
	for i := 0; i < 8; i++ {
		r[i] = a[i] + b[i]
	}
}

func minus(a, b, r *[8]float64) {
	for i := 0; i < 8; i++ {
		r[i] = a[i] - b[i]
	}
}

func mul(a, b, r *[8]float64) {
	for i := 0; i < 8; i++ {
		r[i] = a[i] * b[i]
	}
}
