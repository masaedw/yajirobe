package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	//"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"golang.org/x/text/encoding/japanese"
	surf "gopkg.in/headzoo/surf.v1"
)

func toUtf8(str string) string {
	s, _ := japanese.ShiftJIS.NewDecoder().String(str)
	return s
}

func toSjis(str string) string {
	s, _ := japanese.ShiftJIS.NewEncoder().String(str)
	return s
}

func setForms(form browser.Submittable, inputs map[string]string) error {
	for k, v := range inputs {
		err := form.Set(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func printPage(bow browser.Browsable) {
	fmt.Printf("title:%s\n", toUtf8(bow.Title()))
	fmt.Printf("body:\n%s", toUtf8(bow.Body()))
}

func exportValues(form *goquery.Selection) url.Values {
	data := url.Values{}

	form.Find("input").Each(func(_ int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		data.Add(name, value)
	})

	return data
}

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

func printHTML(s *goquery.Selection) {
	h, _ := s.Html()
	fmt.Println(toUtf8(h))
}

func filterAttrContains(dom *goquery.Selection, attr, text string) *goquery.Selection {
	return dom.FilterFunction(func(_ int, s *goquery.Selection) bool {
		return strings.Contains(s.AttrOr(attr, ""), text)
	})
}

func filterTextContains(dom *goquery.Selection, text string) *goquery.Selection {
	return dom.FilterFunction(func(_ int, s *goquery.Selection) bool {
		return strings.Contains(s.Text(), text)
	})
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

func sbiScan(bow *browser.Browser) error {
	stocksFont := filterTextContains(bow.Find("font"), toSjis("銘柄"))

	stocksTable := stocksFont.ParentsFiltered("table").First()
	stocks := []Stock{}

	fmt.Printf("stocks\n")
	stocksTable.Find("tr").Each(func(i int, tr *goquery.Selection) {
		if i == 0 {
			return
		}

		s := Stock{}
		tr.Find("td").Each(func(i int, td *goquery.Selection) {
			switch i {
			case 0:
				fmt.Println(text)
				fields := strings.Fields(text)
				s.Name = fields[0]
				s.Code = fields[1]
			case 1:
				s.Amount, _ = strconv.Atoi(toUtf8(td.Text()))
			case 2:
				units := strings.Replace(toUtf8(td.Text()), ",", "", -1)
				fields := strings.Fields(units)
				s.AcquisitionUnitPrice, _ = strconv.ParseInt(fields[0], 10, 64)
				s.CurrentUnitPrice, _ = strconv.ParseInt(fields[1], 10, 64)
			}
		})
		stocks = append(stocks, s)
	})

	fmt.Println(stocks)

	funds := bow.Find("table").FilterFunction(func(_ int, s *goquery.Selection) bool {
		str := toUtf8(s.Find("tr").First().Text())
		return strings.Contains(str, "投資信託")
	})

	fmt.Print("funds\n")
	funds.Each(func(_ int, s *goquery.Selection) {
		fmt.Print(s)
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
