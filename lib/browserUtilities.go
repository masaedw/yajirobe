package yajirobe

import (
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/browser"
	"golang.org/x/net/html"
)

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
