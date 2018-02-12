package yajirobe

import "testing"

func Test1(t *testing.T) {

	funds := []*Fund{
		&Fund{
			Name:         "TestStock",
			Code:         "1234",
			Amount:       10,
			AssetClass:   EmergingStocks,
			CurrentPrice: 200,
		},
		&Fund{
			Name:         "TestStock",
			Code:         "5678",
			Amount:       10,
			AssetClass:   DomesticStocks,
			CurrentPrice: 200,
		},
		&Fund{
			Name:         "TestStock",
			Code:         "90ab",
			Amount:       10,
			AssetClass:   InternationalStocks,
			CurrentPrice: 600,
		},
	}

	target := AllocationTarget{
		EmergingStocks:      0.25,
		DomesticStocks:      0.30,
		InternationalStocks: 0.45,
	}

	a := NewAssetAllocation([]*Stock{}, funds, target)
	result := a.RebalancingBuy(100)

	assert := func(class AssetClass, expected float64) {
		if result[class] != expected {
			t.Errorf("%v expected %.2f but got %.2f", class, expected, result[class])
		}
	}

	assert(EmergingStocks, 37)
	assert(DomesticStocks, 63)
	assert(InternationalStocks, 0)
}

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
