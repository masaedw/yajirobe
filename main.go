package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	//"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"golang.org/x/text/encoding/japanese"
	surf "gopkg.in/headzoo/surf.v1"
)

func sjisToUtf8(str string) string {
	s, _ := japanese.ShiftJIS.NewDecoder().String(str)
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
	fmt.Printf("title:%s\n", sjisToUtf8(bow.Title()))
	fmt.Printf("body:\n%s", sjisToUtf8(bow.Body()))
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

	text := sjisToUtf8(bow.Find("font").Text())
	if strings.Contains(text, "WBLE") {
		return nil, errors.New(text)
	}

	nextForm := bow.Find("form").First()
	if nextForm == nil {
		return nil, errors.New("formSwitch not found")
	}

	bow.PostForm(nextForm.AttrOr("action", "url not found"), exportValues(nextForm))

	if !strings.Contains(sjisToUtf8(bow.Find(".tp-box-05").Text()), "最終ログイン:") {
		return nil, errors.New("the SBI User ID or Password failed")
	}

	return bow, nil
}

func sbiScan(bow *browser.Browser) error {
	bow.Click("//a[*[contains(@alt,\"口座管理\")]]")
	bow.Click("//area[@title=\"保有証券\"]")

	stocks := bow.Find("table").FilterFunction(func(_ int, s *goquery.Selection) bool {
		str := sjisToUtf8(s.Find("tr").First().Text())
		return strings.Contains(str, "銘柄")
	})

	fmt.Print("stocks\n")
	stocks.Each(func(_ int, s *goquery.Selection) {
		str := sjisToUtf8(s.Text())
		if str != "" {
			fmt.Println(strings.TrimSpace(str))
		}
	})

	funds := bow.Find("table").FilterFunction(func(_ int, s *goquery.Selection) bool {
		str := sjisToUtf8(s.Find("tr").First().Text())
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

	sbiScan(bow)
}
