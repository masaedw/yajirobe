package main

import (
	//"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/browser"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	surf "gopkg.in/headzoo/surf.v1"
)

func sjisToUtf8(str string) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(str), japanese.ShiftJIS.NewDecoder()))
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func sbiLogin(userID, userPassword string) (*browser.Browser, error) {
	bow := surf.NewBrowser()
	err := bow.Open("https://www.sbisec.co.jp/ETGate")
	if err != nil {
		return nil, err
	}

	form, err := bow.Form("[name='form_login']")
	if err != nil {
		return nil, err
	}

	form.Input("user_id", userID)
	form.Input("usrr_passwd", userPassword)
	form.Input("JS_FLG", "1")
	form.Input("BW_FLG", "chrome,56")
	form.Input("ACT_login.x", "12")
	form.Input("ACT_login.y", "12")
	err = form.Submit()
	if err != nil {
		return nil, err
	}

	return bow, nil
}

func sbiScan(bow *browser.Browser) error {
	bow.Click("//a[*[contains(@alt,\"口座管理\")]]")
	bow.Click("//area[@title=\"保有証券\"]")

	stocks := bow.Find("table").FilterFunction(func(_ int, s *goquery.Selection) bool {
		str := s.Find("tr").First().Text()
		return strings.Contains(str, "銘柄")
	})

	fmt.Print("stocks")
	stocks.Each(func(_ int, s *goquery.Selection) {
		fmt.Print(s)
	})

	funds := bow.Find("table").FilterFunction(func(_ int, s *goquery.Selection) bool {
		str := s.Find("tr").First().Text()
		return strings.Contains(str, "投資信託")
	})

	fmt.Print("funds")
	funds.Each(func(_ int, s *goquery.Selection) {
		fmt.Print(s)
	})

	return nil
}

func main() {
	userID := os.Getenv("SBI_USER_ID")
	userPassword := os.Getenv("SBI_UESR_PASSWORD")

	bow, err := sbiLogin(userID, userPassword)
	if err != nil {
		panic(err)
	}

	fmt.Print("Login!")

	sbiScan(bow)
}
