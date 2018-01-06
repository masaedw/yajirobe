package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	surf "gopkg.in/headzoo/surf.v1"
)

func sbiLogin(userID, userPassword string) (*browser.Browser, error) {
	bow := surf.NewBrowser()
	bow.SetUserAgent(agent.Chrome())

	bow.Open("https://www.sbisec.co.jp/ETGate")

	loginForm, _ := bow.Form("[name='form_login']")

	setForms(loginForm, map[string]string{
		"JS_FLG":          "0",
		"BW_FLG":          "0",
		"_ControlID":      "WPLETlgR001Control",
		"_DataStoreID":    "DSWPLETlgR001Control",
		"_PageID":         "WPLETlgR001Rlgn20",
		"_ActionID":       "login",
		"getFlg":          "on",
		"allPrmFlg":       "on",
		"_ReturnPageInfo": "WPLEThmR001Control/DefaultPID/DefaultAID/DSWPLEThmR001Control",
		"user_id":         userID,
		"user_password":   userPassword,
	})

	loginForm.Submit()

	text := toUtf8(bow.Find("font").Text())
	if strings.Contains(text, "WBLE") {
		// ログイン失敗画面
		return nil, errors.New(text)
	}

	nextForm := bow.Find("form").First()
	if nextForm == nil {
		return nil, errors.New("formSwitch not found")
	}

	// 2回目のPOST
	bow.PostForm(nextForm.AttrOr("action", "url not found"), exportValues(nextForm))

	if !strings.Contains(toUtf8(bow.Find(".tp-box-05").Text()), "最終ログイン:") {
		// ログイン成功時のメッセージが出てなければログイン失敗してる
		return nil, errors.New("the SBI User ID or Password failed")
	}

	return bow, nil
}

func sbiAccountPage(bow *browser.Browser) error {
	s := filterAttrContains(bow.Dom().Find("img"), "alt", toSjis("口座管理"))
	if s == nil || s.Length() == 0 {
		return errors.New("can't find 口座管理")
	}
	url := s.Parent().AttrOr("href", "can't find url")
	url, _ = bow.ResolveStringUrl(url)
	e := bow.Open(url)
	if e != nil {
		return e
	}

	s = filterAttrContains(bow.Dom().Find("area"), "alt", toSjis("保有証券"))
	url, _ = bow.ResolveStringUrl(s.AttrOr("href", "can't find url"))
	e = bow.Open(url)
	if e != nil {
		return e
	}

	return nil
}

func sbiScanStock(row *goquery.Selection) Stock {
	cells := iterate(row.Find("td"))

	nameCode := toUtf8(cells[0].Text())
	re := regexp.MustCompile(`(\S+?)(\d{4})`)
	m := re.FindStringSubmatch(nameCode)

	name := m[1]
	code, _ := strconv.Atoi(m[2])
	amount := int(parseSeparatedInt(toUtf8(cells[1].Text())))
	units := iterateText(cells[2])
	acquisitionUnitPrice := parseSeparatedInt(units[0])
	currentUnitPrice := parseSeparatedInt(units[1])
	acquisitionPrice := acquisitionUnitPrice * int64(amount)
	currentPrice := currentUnitPrice * int64(amount)

	return Stock{
		Name:                 name,
		Code:                 code,
		Amount:               amount,
		AcquisitionUnitPrice: acquisitionUnitPrice,
		CurrentUnitPrice:     currentUnitPrice,
		AcquisitionPrice:     acquisitionPrice,
		CurrentPrice:         currentPrice,
	}
}

func sbiGetFundCategory(bow *browser.Browser, code string) string {
	url, _ := url.Parse("https://site0.sbisec.co.jp/marble/fund/detail/achievement.do")
	query := url.Query()
	query.Set("Param6", code)
	url.RawQuery = query.Encode()
	bow.Open(url.String())
	categoryHeader := filterTextContains(bow.Find("tr th div p"), toSjis("商品分類"))
	category := categoryHeader.Parent().Parent().Parent().First().Next()
	return strings.TrimSpace(toUtf8(category.Text()))
}

func sbiScanFund(bow *browser.Browser, row *goquery.Selection) Fund {
	cells := iterate(row.Find("td"))

	href := cells[0].Find("a").AttrOr("href", "Fund name not found")
	url, _ := url.Parse(href)
	query := url.Query()
	name := toUtf8(query.Get("sec_name"))
	code := query.Get("fund_sec_code")
	amount := parseSeparatedInt(cells[1].Text())
	units := iterateText(cells[2])
	acquisitionUnitPrice := parseSeparatedInt(units[0])
	currentUnitPrice := parseSeparatedInt(units[1])
	category := parseAssetClass(sbiGetFundCategory(bow, code))

	return Fund{
		Name:                 name,
		Code:                 code,
		AssetClass:           category,
		Amount:               int(amount),
		AcquisitionUnitPrice: float64(acquisitionUnitPrice),
		CurrentUnitPrice:     float64(currentUnitPrice),
		AcquisitionPrice:     float64(acquisitionUnitPrice) * float64(amount) / 10000,
		CurrentPrice:         float64(currentUnitPrice) * float64(amount) / 10000,
	}
}

func sbiScan(bow *browser.Browser) ([]Stock, []Fund, error) {
	if e := sbiAccountPage(bow); e != nil {
		return nil, nil, e
	}

	stockFont := filterTextContains(bow.Find("font"), toSjis("銘柄"))
	stockTable := stockFont.ParentsFiltered("table").First()

	stocks := []Stock{}

	for _, tr := range iterate(stockTable.Find("tr"))[1:] {
		stocks = append(stocks, sbiScanStock(tr))
	}

	fundFont := filterTextContains(bow.Find("font"), toSjis("ファンド名"))
	fundTables := iterate(fundFont.Parent().Parent().Parent().Parent())

	funds := []Fund{}

	for _, table := range fundTables {
		for i, tr := range iterate(table.Find("tr")) {
			if i%2 == 0 {
				continue
			}

			funds = append(funds, sbiScanFund(bow, tr))
		}
	}

	return stocks, funds, nil
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

	s, f, err := sbiScan(bow)
	if err != nil {
		panic(err)
	}

	a := NewAssetAllocation(s, f)

	renderStocks(s)

	fmt.Println("")

	renderAllocation(a, getAllocationTarget())
}
