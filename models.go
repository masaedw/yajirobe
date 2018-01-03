package main

import (
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

// Fund 投資信託
type Fund struct {
	Name                 string       // 名称
	Code                 string       // 協会コード
	Amount               int          // 保有口数
	FundCategory         FundCategory // カテゴリ
	AcquisitionUnitPrice float64      // 取得単価
	CurrentUnitPrice     float64      // 基準価額
	AcquisitionPrice     float64      // 取得金額
	CurrentPrice         float64      // 評価額
}

// FundCategory カテゴリ
type FundCategory int

const (
	// Other その他
	Other = FundCategory(iota)
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

func (c FundCategory) String() string {
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

func parseFundCategory(s string) FundCategory {
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
