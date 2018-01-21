package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app   = kingpin.New("yajirobe", "Asset allocation rebalance tool")
	debug = app.Flag("debug", "Enable debug mode").Default("false").Bool()

	show = app.Command("show", "Show your asset allocation").Default()

	logger *zap.Logger
)

func getAllocationTarget() AllocationTarget {
	return AllocationTarget{
		DomesticStocks:      0.01 * 23,
		InternationalStocks: 0.01 * 30,
		EmergingStocks:      0.01 * 13,
		DomesticBonds:       0.01 * 3,
		InternationalBonds:  0.01 * 13,
		EmergingBonds:       0.01 * 3,
		DomesticREIT:        0.01 * 5,
		InternationalREIT:   0.01 * 5,
		Comodity:            0.01 * 5,
	}
}

func main() {
	userID := os.Getenv("SBI_USER_ID")
	userPassword := os.Getenv("SBI_USER_PASSWORD")

	bow, err := sbiLogin(userID, userPassword)
	if err != nil {
		panic(err)
	}

	cache, err := NewFileFundInfoCache()
	if err != nil {
		panic(err)
	}

	s, f, err := sbiScan(bow, cache)
	if err != nil {
		panic(err)
	}

	a := NewAssetAllocation(s, f, getAllocationTarget())

	//renderStocks(s)

	fmt.Println("")

	renderAllocation(a)
}
