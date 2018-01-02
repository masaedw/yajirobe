package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	//"time"

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

	s.Amount, _ = strconv.Atoi(toUtf8(cells[1].Text()))

	units := iterateText(cells[2])

	fmt.Println(units)
	s.AcquisitionUnitPrice = parseCurrencyValue(units[0])
	s.CurrentUnitPrice = parseCurrencyValue(units[1])

	s.AcquisitionPrice = s.AcquisitionUnitPrice * int64(s.Amount)
	s.CurrentPrice = s.CurrentUnitPrice * int64(s.Amount)

	return s
}

func sbiScan(bow *browser.Browser) error {
	stocksFont := filterTextContains(bow.Find("font"), toSjis("銘柄"))
	stocksTable := stocksFont.ParentsFiltered("table").First()

	stocks := []Stock{}

	for _, tr := range iterate(stocksTable.Find("tr"))[1:] {
		stocks = append(stocks, sbiScanStock(tr))
	}
	fmt.Printf("stocks\n")

	for _, stock := range stocks {
		fmt.Printf("%#v\n", stock)
	}

	fundsFont := filterTextContains(bow.Find("font"), toSjis("投資信託"))
	fmt.Println(fundsFont.Length())
	fundsTable := fundsFont.ParentsFiltered("table")
	fmt.Println(fundsTable.Length())

	fmt.Print("funds\n")
	fundsTable.Each(func(_ int, s *goquery.Selection) {
		//fmt.Println(s)
	})

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
