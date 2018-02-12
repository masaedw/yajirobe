package yajirobe

import "testing"

func TestRound(t *testing.T) {
	assert := func(expected, n float64) {
		if round(n) != expected {
			t.Errorf(" round(%.2f) expected %.2f but got %.2f", n, expected, round(n))
		}
	}

	assert(-2, -1.6)
	assert(-2, -1.5)
	assert(-1, -1.4)
	assert(-1, -1)
	assert(-1, -0.6)
	assert(0, -0.5)
	assert(0, -0.4)
	assert(0, 0)
	assert(0, .4)
	assert(0, .5)
	assert(1, .6)
	assert(1, 1)
	assert(1, 1.4)
	assert(2, 1.5)
	assert(2, 1.6)
	assert(2, 2)
	assert(2, 2.4)
	assert(2, 2.5)
	assert(3, 2.6)
	assert(3, 3)
}
