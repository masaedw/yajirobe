package main

import (
	"fmt"
	"regexp"
)

// Stock 銘柄
type Stock struct {
	Name                 string // 名称
	Code                 int    // 銘柄コード
	Amount               int    // 保有株数
	AcquisitionUnitPrice int64  // 取得単価
	CurrentUnitPrice     int64  // 現在値
	AcquisitionPrice     int64  // 取得金額
	CurrentPrice         int64  // 評価額
}

// ProfitAndLoss 損益
func (s Stock) ProfitAndLoss() int64 {
	return s.CurrentPrice - s.AcquisitionPrice
}

// ProfitAndLossRatio 損益率
func (s Stock) ProfitAndLossRatio() float64 {
	return float64(s.CurrentPrice-s.AcquisitionPrice) / float64(s.AcquisitionPrice)
}

// Fund 投資信託
type Fund struct {
	Name                 string     // 名称
	Code                 string     // 協会コード
	Amount               int        // 保有口数
	AssetClass           AssetClass // アセットクラス
	AcquisitionUnitPrice float64    // 取得単価
	CurrentUnitPrice     float64    // 基準価額
	AcquisitionPrice     float64    // 取得金額
	CurrentPrice         float64    // 評価額
}

// ProfitAndLoss 損益
func (f Fund) ProfitAndLoss() float64 {
	return f.CurrentPrice - f.AcquisitionPrice
}

// ProfitAndLossRatio 損益率
func (f Fund) ProfitAndLossRatio() float64 {
	return (f.CurrentPrice - f.AcquisitionPrice) / f.AcquisitionPrice
}

// AssetClass アセットクラス
type AssetClass int

const (
	// Other その他
	Other = AssetClass(iota)
	// DomesticStocks 国内株式
	DomesticStocks
	// DomesticBonds 国内債券
	DomesticBonds
	// DomesticREIT 国内REIT
	DomesticREIT
	// InternationalStocks 国際株式
	InternationalStocks
	// InternationalBonds 国際債券
	InternationalBonds
	// InternationalREIT 国際REIT
	InternationalREIT
	// EmergingStocks 新興国株式
	EmergingStocks
	// EmergingBonds 新興国債券
	EmergingBonds
	// EmergingREIT 新興国REIT
	EmergingREIT
	// Balance バランス
	Balance
	// Comodity コモディティ
	Comodity
	// HedgeFund ヘッジファンド
	HedgeFund
	// BullBear ブルベア
	BullBear
)

// AssetClasses アセットクラス一覧
var AssetClasses = []AssetClass{
	DomesticStocks,
	InternationalStocks,
	EmergingStocks,
	DomesticBonds,
	InternationalBonds,
	EmergingBonds,
	DomesticREIT,
	InternationalREIT,
	EmergingREIT,
	Balance,
	Comodity,
	HedgeFund,
	BullBear,
	Other,
}

var (
	domesticStocksPattern      = regexp.MustCompile("国内株式")
	domesticBondsPattern       = regexp.MustCompile("国内債券")
	domesticREITPattern        = regexp.MustCompile("国内REIT")
	internationalStocksPattern = regexp.MustCompile("海外株式")
	internationalBondsPattern  = regexp.MustCompile("海外債券")
	internationalREITPattern   = regexp.MustCompile("海外REIT")
	emergingStocksPattern      = regexp.MustCompile("新興国株式")
	emergingBondsPattern       = regexp.MustCompile("新興国債券")
	emergingREITPattern        = regexp.MustCompile("新興国REIT")
	balancePattern             = regexp.MustCompile("バランス")
	comodity                   = regexp.MustCompile("コモディティ")
	hedgeFundPattern           = regexp.MustCompile("ヘッジファンド")
	bullBearPattern            = regexp.MustCompile("ブル・ベア")
)

func (c AssetClass) String() string {
	switch c {
	default:
		return "その他"
	case DomesticStocks:
		return "国内株式"
	case DomesticBonds:
		return "国内債券"
	case DomesticREIT:
		return "国内REIT"
	case InternationalStocks:
		return "海外株式"
	case InternationalBonds:
		return "海外債券"
	case InternationalREIT:
		return "海外REIT"
	case EmergingStocks:
		return "新興国株式"
	case EmergingBonds:
		return "新興国債券"
	case EmergingREIT:
		return "新興国REIT"
	case Balance:
		return "バランス"
	case Comodity:
		return "コモディティ"
	case HedgeFund:
		return "ヘッジファンド"
	case BullBear:
		return "ブルベア"
	}
}

func parseAssetClass(s string) AssetClass {
	switch {
	default:
		return Other
	case domesticStocksPattern.MatchString(s):
		return DomesticStocks
	case domesticBondsPattern.MatchString(s):
		return DomesticBonds
	case domesticREITPattern.MatchString(s):
		return DomesticREIT
	case internationalStocksPattern.MatchString(s):
		return InternationalStocks
	case internationalBondsPattern.MatchString(s):
		return InternationalBonds
	case internationalREITPattern.MatchString(s):
		return InternationalREIT
	case emergingStocksPattern.MatchString(s):
		return EmergingStocks
	case emergingBondsPattern.MatchString(s):
		return EmergingBonds
	case emergingREITPattern.MatchString(s):
		return EmergingStocks
	case balancePattern.MatchString(s):
		return Balance
	case comodity.MatchString(s):
		return Comodity
	case hedgeFundPattern.MatchString(s):
		return HedgeFund
	case bullBearPattern.MatchString(s):
		return BullBear
	}
}

type fundUnited struct {
	Fund
	sources []Fund
}

func newFundUnited(s Fund) *fundUnited {
	return &fundUnited{Fund: s, sources: []Fund{s}}
}

func (lhs *fundUnited) merge(rhs Fund) {
	newAmount := lhs.Amount + rhs.Amount
	aprice := lhs.AcquisitionPrice + rhs.AcquisitionPrice
	cprice := lhs.CurrentPrice + rhs.CurrentPrice
	aunit := aprice / float64(newAmount) * 10000
	cunit := cprice / float64(newAmount) * 10000

	lhs.Amount = newAmount
	lhs.AcquisitionPrice = aprice
	lhs.AcquisitionUnitPrice = aunit
	lhs.CurrentPrice = cprice
	lhs.CurrentUnitPrice = cunit

	lhs.sources = append(lhs.sources, rhs)
}

// AssetClassDetail 資産カテゴリごとの明細
type AssetClassDetail struct {
	class  AssetClass
	aprice float64
	cprice float64
	funds  map[string]*fundUnited
}

func newAssetClassDetail(fu *fundUnited) *AssetClassDetail {
	return &AssetClassDetail{
		class:  fu.AssetClass,
		aprice: fu.AcquisitionPrice,
		cprice: fu.CurrentPrice,
		funds: map[string]*fundUnited{
			fu.Code: fu,
		},
	}
}

func (d *AssetClassDetail) merge(fu *fundUnited) {
	d.aprice += fu.AcquisitionPrice
	d.cprice += fu.CurrentPrice
	d.funds[fu.Code] = fu
}

// AllocationTarget 目標アロケーション
type AllocationTarget map[AssetClass]float64

// AssetAllocation 現在のアセットアロケーション
type AssetAllocation struct {
	aprice  float64
	cprice  float64
	details map[AssetClass]*AssetClassDetail
}

func (a *AssetAllocation) merge(fu *fundUnited) {
	a.aprice += fu.AcquisitionPrice
	a.cprice += fu.CurrentPrice
	d, e := a.details[fu.AssetClass]
	if e {
		d.merge(fu)
	} else {
		a.details[fu.AssetClass] = newAssetClassDetail(fu)
	}
}

func fundsFromETF(stocks []Stock) []Fund {
	etf := map[int]AssetClass{
		1680: InternationalStocks,
	}

	fs := []Fund{}

	for _, s := range stocks {
		c, e := etf[s.Code]
		if e {
			fs = append(fs, Fund{
				Name:                 s.Name,
				Code:                 fmt.Sprint(s.Code),
				Amount:               s.Amount,
				AssetClass:           c,
				AcquisitionUnitPrice: float64(s.AcquisitionUnitPrice) * 10000,
				CurrentUnitPrice:     float64(s.CurrentUnitPrice) * 10000,
				AcquisitionPrice:     float64(s.AcquisitionPrice),
				CurrentPrice:         float64(s.CurrentPrice),
			})
		}
	}

	return fs
}

// NewAssetAllocation アセットアロケーション計算
func NewAssetAllocation(stocks []Stock, funds []Fund) AssetAllocation {
	mergedFunds := map[string]*fundUnited{}

	funds = append([]Fund{}, funds...)
	funds = append(funds, fundsFromETF(stocks)...)

	for _, f := range funds {
		fu, e := mergedFunds[f.Code]
		if e {
			fu.merge(f)
		} else {
			mergedFunds[f.Code] = newFundUnited(f)
		}
	}

	a := AssetAllocation{
		details: map[AssetClass]*AssetClassDetail{},
	}

	for _, fu := range mergedFunds {
		a.merge(fu)
	}

	return a
}
