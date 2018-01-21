package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	yajirobe "github.com/masaedw/yajirobe/lib"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app   = kingpin.New("yajirobe", "Asset allocation rebalance tool")
	debug = app.Flag("debug", "Enable debug mode").Default("false").Bool()

	show = app.Command("show", "Show your asset allocation").Default()

	logger *zap.Logger
)

func getAllocationTarget() yajirobe.AllocationTarget {
	return yajirobe.AllocationTarget{
		yajirobe.DomesticStocks:      0.01 * 23,
		yajirobe.InternationalStocks: 0.01 * 30,
		yajirobe.EmergingStocks:      0.01 * 13,
		yajirobe.DomesticBonds:       0.01 * 3,
		yajirobe.InternationalBonds:  0.01 * 13,
		yajirobe.EmergingBonds:       0.01 * 3,
		yajirobe.DomesticREIT:        0.01 * 5,
		yajirobe.InternationalREIT:   0.01 * 5,
		yajirobe.Comodity:            0.01 * 5,
	}
}

func main() {
	userID := os.Getenv("SBI_USER_ID")
	userPassword := os.Getenv("SBI_USER_PASSWORD")

	bow, err := yajirobe.SbiLogin(userID, userPassword)
	if err != nil {
		panic(err)
	}

	cache, err := yajirobe.NewFileFundInfoCache()
	if err != nil {
		panic(err)
	}

	s, f, err := yajirobe.SbiScan(bow, cache)
	if err != nil {
		panic(err)
	}

	a := yajirobe.NewAssetAllocation(s, f, getAllocationTarget())

	//renderStocks(s)

	fmt.Println("")

	yajirobe.RenderAllocation(a)
}
