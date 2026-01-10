package runtime

import "math"

// Sqrt returns the square root of x.
// Ruby: Math.sqrt(x)
func Sqrt(x float64) float64 {
	return math.Sqrt(x)
}

// Pow returns base raised to exp.
// Ruby: base ** exp or Math.pow(base, exp)
func Pow(base, exp float64) float64 {
	return math.Pow(base, exp)
}

// Ceil returns the least integer value greater than or equal to x.
// Ruby: f.ceil
// Panics if the result is outside the range of int or if x is NaN/Inf.
func Ceil(x float64) int {
	c := math.Ceil(x)
	if c > float64(math.MaxInt) || c < float64(math.MinInt) || math.IsNaN(c) || math.IsInf(c, 0) {
		panic("Ceil: result out of int range")
	}
	return int(c)
}

// Floor returns the greatest integer value less than or equal to x.
// Ruby: f.floor
// Panics if the result is outside the range of int or if x is NaN/Inf.
func Floor(x float64) int {
	f := math.Floor(x)
	if f > float64(math.MaxInt) || f < float64(math.MinInt) || math.IsNaN(f) || math.IsInf(f, 0) {
		panic("Floor: result out of int range")
	}
	return int(f)
}

// Round returns the nearest integer, rounding half away from zero.
// Ruby: f.round
// Panics if the result is outside the range of int or if x is NaN/Inf.
func Round(x float64) int {
	r := math.Round(x)
	if r > float64(math.MaxInt) || r < float64(math.MinInt) || math.IsNaN(r) || math.IsInf(r, 0) {
		panic("Round: result out of int range")
	}
	return int(r)
}
