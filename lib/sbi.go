package yajirobe

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	surf "gopkg.in/headzoo/surf.v1"
)

func SbiLogin(userID, userPassword string) (*browser.Browser, error) {
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

func sbiScanStock(row *goquery.Selection) *Stock {
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

	return &Stock{
		Name:                 name,
		Code:                 code,
		Amount:               amount,
		AcquisitionUnitPrice: acquisitionUnitPrice,
		CurrentUnitPrice:     currentUnitPrice,
		AcquisitionPrice:     acquisitionPrice,
		CurrentPrice:         currentPrice,
	}
}

func sbiGetFundInfo(bow *browser.Browser, code FundCode) *FundInfo {
	url, _ := url.Parse("https://site0.sbisec.co.jp/marble/fund/detail/achievement.do")
	query := url.Query()
	query.Set("Param6", string(code))
	url.RawQuery = query.Encode()
	bow.Open(url.String())
	categoryHeader := filterTextContains(bow.Find("tr th div p"), toSjis("商品分類"))
	category := categoryHeader.Parent().Parent().Parent().First().Next()
	nameHeader := strings.TrimSpace(toUtf8(bow.Find("h3").Text()))
	names := strings.Split(nameHeader, "－")
	name := names[0]
	if len(names) > 1 {
		name = names[1]
	}
	assetClass := parseAssetClass(strings.TrimSpace(toUtf8(category.Text())))

	return &FundInfo{
		Name:  name,
		Class: assetClass,
		Code:  code,
	}
}

func sbiScanFund(bow *browser.Browser, row *goquery.Selection, cache FundInfoCache) (*Fund, error) {
	cells := iterate(row.Find("td"))

	href := cells[0].Find("a").AttrOr("href", "Fund name not found")
	url, _ := url.Parse(href)
	query := url.Query()
	code := FundCode(query.Get("fund_sec_code"))
	amount := parseSeparatedInt(cells[1].Text())
	units := iterateText(cells[2])
	acquisitionUnitPrice := parseSeparatedInt(units[0])
	currentUnitPrice := parseSeparatedInt(units[1])

	fi, err := cache.Get(code)
	if err != nil && !IsCacheNotExists(err) {
		return nil, err
	}

	name := fi.Name
	class := fi.Class

	if err != nil { // Cacheにヒットしなかった
		fi = sbiGetFundInfo(bow, code)
		if err = cache.Set(fi); err != nil {
			return nil, err
		}
	}

	return &Fund{
		Name:                 name,
		Code:                 code,
		AssetClass:           class,
		Amount:               int(amount),
		AcquisitionUnitPrice: float64(acquisitionUnitPrice),
		CurrentUnitPrice:     float64(currentUnitPrice),
		AcquisitionPrice:     float64(acquisitionUnitPrice) * float64(amount) / 10000,
		CurrentPrice:         float64(currentUnitPrice) * float64(amount) / 10000,
	}, nil
}

func SbiScan(bow *browser.Browser, cache FundInfoCache) ([]*Stock, []*Fund, error) {
	if e := sbiAccountPage(bow); e != nil {
		return nil, nil, e
	}

	stockFont := filterTextContains(bow.Find("font"), toSjis("銘柄"))
	stockTable := stockFont.ParentsFiltered("table").First()

	stocks := []*Stock{}

	for _, tr := range iterate(stockTable.Find("tr"))[1:] {
		stocks = append(stocks, sbiScanStock(tr))
	}

	fundFont := filterTextContains(bow.Find("font"), toSjis("ファンド名"))
	fundTables := iterate(fundFont.Parent().Parent().Parent().Parent())

	funds := []*Fund{}

	for _, table := range fundTables {
		for i, tr := range iterate(table.Find("tr")) {
			if i%2 == 0 {
				continue
			}
			f, e := sbiScanFund(bow, tr, cache)
			if e != nil {
				return nil, nil, e
			}
			funds = append(funds, f)
		}
	}

	return stocks, funds, nil
}
