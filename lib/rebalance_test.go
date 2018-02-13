package yajirobe

import (
	"fmt"
	"testing"
)

func newFund(c AssetClass, p float64) *Fund {
	return &Fund{
		Name:         "TestFund",
		Code:         FundCode(fmt.Sprintf("%05d", int64(c))),
		Amount:       10,
		AssetClass:   c,
		CurrentPrice: p,
	}
}

type assert func(class AssetClass, expected float64)

func makeAssert(t *testing.T, result map[AssetClass]float64) assert {
	return func(class AssetClass, expected float64) {
		if result[class] != expected {
			t.Errorf("%v expected %.2f but got %.2f", class, expected, result[class])
		}
	}
}

func Test1(t *testing.T) {

	funds := []*Fund{
		newFund(EmergingStocks, 200),
		newFund(DomesticStocks, 200),
		newFund(InternationalStocks, 600),
	}

	target := AllocationTarget{
		EmergingStocks:      0.25,
		DomesticStocks:      0.30,
		InternationalStocks: 0.45,
	}

	a := NewAssetAllocation([]*Stock{}, funds, target)
	result := a.RebalancingBuy(100)

	assert := makeAssert(t, result)

	assert(EmergingStocks, 37)
	assert(DomesticStocks, 63)
	assert(InternationalStocks, 0)
}

func Test2(t *testing.T) {

	funds := []*Fund{
		newFund(EmergingStocks, 200),
		newFund(DomesticStocks, 200),
		newFund(InternationalStocks, 600),
	}

	target := AllocationTarget{
		EmergingStocks:      0.25,
		DomesticStocks:      0.30,
		InternationalStocks: 0.45,
	}

	a := NewAssetAllocation([]*Stock{}, funds, target)
	result := a.RebalancingBuy(300)

	assert := makeAssert(t, result)

	assert(EmergingStocks, 119)
	assert(DomesticStocks, 181)
	assert(InternationalStocks, 0)
}

func Test3(t *testing.T) {

	funds := []*Fund{
		newFund(EmergingStocks, 200),
		newFund(DomesticStocks, 200),
		newFund(InternationalStocks, 600),
	}

	target := AllocationTarget{
		EmergingStocks:      0.25,
		DomesticStocks:      0.30,
		InternationalStocks: 0.45,
	}

	a := NewAssetAllocation([]*Stock{}, funds, target)
	result := a.RebalancingBuy(400)

	assert := makeAssert(t, result)

	assert(EmergingStocks, 150)
	assert(DomesticStocks, 220)
	assert(InternationalStocks, 30)
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
