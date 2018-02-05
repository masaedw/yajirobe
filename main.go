package main

import (
	"fmt"
	"os"

	"github.com/masaedw/yajirobe/lib"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
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

func errorExit(err error) {
	fmt.Fprintf(os.Stderr, "%+v", err)
	os.Exit(1)
}

func main() {
	userID := os.Getenv("SBI_USER_ID")
	userPassword := os.Getenv("SBI_USER_PASSWORD")

	logger, err := zap.NewDevelopment()
	if err != nil {
		errorExit(err)
	}

	cache, err := yajirobe.NewFileFundInfoCache(logger)
	if err != nil {
		errorExit(err)
	}

	sbi, err := yajirobe.NewSbiScanner(yajirobe.SbiOption{
		UserID:   userID,
		Password: userPassword,
		Logger:   logger,
		Cache:    cache,
	})

	if err != nil {
		errorExit(err)
	}

	s, f, err := sbi.Scan()
	if err != nil {
		errorExit(err)
	}

	a := yajirobe.NewAssetAllocation(s, f, getAllocationTarget())

	fmt.Println("")

	a.Render()
}
