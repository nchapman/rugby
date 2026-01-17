package main

import (
	"fmt"
	"os"
	"strconv"
)

func fannkuch(n int) (int, int) {
	perm := make([]int, n)
	perm1 := make([]int, n)
	count := make([]int, n)

	for i := 0; i < n; i++ {
		perm1[i] = i
	}

	maxFlips := 0
	checksum := 0
	r := n
	sign := true

	for {
		// Generate next permutation
		for r != 1 {
			count[r-1] = r
			r--
		}

		// Copy and count flips
		copy(perm, perm1)
		flips := 0
		for perm[0] != 0 {
			k := perm[0]
			// Reverse perm[0..k]
			for i, j := 0, k; i < j; i, j = i+1, j-1 {
				perm[i], perm[j] = perm[j], perm[i]
			}
			flips++
		}

		if flips > maxFlips {
			maxFlips = flips
		}
		if sign {
			checksum += flips
		} else {
			checksum -= flips
		}

		// Next permutation
		for {
			if r == n {
				return checksum, maxFlips
			}
			p0 := perm1[0]
			for i := 0; i < r; i++ {
				perm1[i] = perm1[i+1]
			}
			perm1[r] = p0

			count[r]--
			if count[r] > 0 {
				break
			}
			r++
		}
		sign = !sign
	}
}

func main() {
	n := 7
	if len(os.Args) > 1 {
		n, _ = strconv.Atoi(os.Args[1])
	}

	checksum, maxFlips := fannkuch(n)
	fmt.Printf("%d\nPfannkuchen(%d) = %d\n", checksum, n, maxFlips)
}
