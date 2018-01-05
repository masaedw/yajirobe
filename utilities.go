package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/browser"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/japanese"
)

// encoding

func toUtf8(str string) string {
	s, _ := japanese.ShiftJIS.NewDecoder().String(str)
	return s
}

func toSjis(str string) string {
	s, _ := japanese.ShiftJIS.NewEncoder().String(str)
	return s
}

// form manupilation

func setForms(form browser.Submittable, inputs map[string]string) error {
	for k, v := range inputs {
		err := form.Set(k, v)
		if err != nil {
			return err
		}
	}
	return nil
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

// debugging

func printPage(bow browser.Browsable) {
	fmt.Printf("title:%s\n", toUtf8(bow.Title()))
	fmt.Printf("body:\n%s", toUtf8(bow.Body()))
}

func printHTML(s *goquery.Selection) {
	h, _ := s.Html()
	fmt.Println(toUtf8(h))
}

// queries

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

func iterate(s *goquery.Selection) (result []*goquery.Selection) {
	s.Each(func(_ int, i *goquery.Selection) {
		result = append(result, i)
	})

	return
}

func iterateText(s *goquery.Selection) (result []string) {
	for _, node := range s.Contents().Nodes {
		if node.Type == html.TextNode {
			if node.Data != "" {
				result = append(result, node.Data)
			}
		}
	}

	return
}

// strings

func parseSeparatedInt(s string) int64 {
	s = strings.Replace(s, ",", "", -1)
	s = regexp.MustCompile(`\d+`).FindString(s)
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}
