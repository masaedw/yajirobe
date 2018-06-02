package main

import (
	"fmt"
	"os"

	"github.com/masaedw/yajirobe/lib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/text/message"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app   = kingpin.New("yajirobe", "Asset allocation rebalance tool")
	debug = app.Flag("debug", "Enable debug mode").Default("false").Bool()

	show = app.Command("show", "Show your asset allocation").Default()

	buy       = app.Command("buy", "Calculate re-balancing buy")
	buyAmount = buy.Arg("amount", "amount").Required().Int64()

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

func createLogger() {
	var err error

	if *debug {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		logger, err = config.Build()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		errorExit(err)
	}
}

func main() {
	userID := os.Getenv("SBI_USER_ID")
	password := os.Getenv("SBI_USER_PASSWORD")

	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	createLogger()

	cache, err := yajirobe.NewFileCache(logger)
	if err != nil {
		errorExit(err)
	}

	sbi, err := yajirobe.NewSbiScanner(yajirobe.SbiOption{
		UserID:   userID,
		Password: password,
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

	switch command {
	case show.FullCommand():
		a.Render()

	case buy.FullCommand():
		result := a.RebalancingBuy(float64(*buyAmount))
		a.Render()

		p := message.NewPrinter(message.MatchLanguage("en"))

		for _, c := range yajirobe.AssetClasses {
			if v, e := result[c]; e {
				p.Printf("%v\t%10.0f\n", c, v)
			}
		}
	}
}
