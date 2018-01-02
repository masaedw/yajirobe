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
	s := Stock{}

	nameCode := toUtf8(cells[0].Text())
	re := regexp.MustCompile(`(\S+?)(\d{4})`)
	m := re.FindStringSubmatch(nameCode)
	s.Name = m[1]
	s.Code, _ = strconv.Atoi(m[2])

	s.Amount = int(parseSeparatedInt(toUtf8(cells[1].Text())))

	units := iterateText(cells[2])

	s.AcquisitionUnitPrice = parseSeparatedInt(units[0])
	s.CurrentUnitPrice = parseSeparatedInt(units[1])

	s.AcquisitionPrice = s.AcquisitionUnitPrice * int64(s.Amount)
	s.CurrentPrice = s.CurrentUnitPrice * int64(s.Amount)

	return s
}

func sbiScanFund(row *goquery.Selection) Fund {
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

	return Fund{
		Name:                 name,
		Code:                 code,
		Amount:               int(amount),
		AcquisitionUnitPrice: float64(acquisitionUnitPrice),
		CurrentUnitPrice:     float64(currentUnitPrice),
		AcquisitionPrice:     float64(acquisitionUnitPrice) * float64(amount) / 10000,
		CurrentPrice:         float64(currentUnitPrice) * float64(amount) / 10000,
	}
}

func sbiScan(bow *browser.Browser) error {
	stockFont := filterTextContains(bow.Find("font"), toSjis("銘柄"))
	stockTable := stockFont.ParentsFiltered("table").First()

	stocks := []Stock{}

	for _, tr := range iterate(stockTable.Find("tr"))[1:] {
		stocks = append(stocks, sbiScanStock(tr))
	}
	fmt.Printf("stocks\n")

	for _, stock := range stocks {
		fmt.Printf("%+v\n", stock)
	}

	fundFont := filterTextContains(bow.Find("font"), toSjis("ファンド名"))
	fundTables := iterate(fundFont.Parent().Parent().Parent().Parent())

	funds := []Fund{}

	for _, table := range fundTables {
		for i, tr := range iterate(table.Find("tr")) {
			if i%2 == 0 {
				continue
			}

			funds = append(funds, sbiScanFund(tr))
		}
	}

	fmt.Print("funds\n")
	for _, fund := range funds {
		fmt.Printf("%+v\n", fund)
	}

	return nil
}

func main() {
	userID := os.Getenv("SBI_USER_ID")
	userPassword := os.Getenv("SBI_USER_PASSWORD")

	bow, err := sbiLogin(userID, userPassword)
	if err != nil {
		panic(err)
	}

	fmt.Print("Login!\n")

	sbiAccountPage(bow)
	sbiScan(bow)
}
