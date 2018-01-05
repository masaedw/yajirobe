package main

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/text/message"
)

func renderStocks(sx []Stock) {
	table := tablewriter.NewWriter(os.Stdout)
	p := message.NewPrinter(message.MatchLanguage("en"))

	table.SetHeader([]string{"Name", "Current", "P/L Ratio"})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	for _, s := range sx {
		table.Append([]string{s.Name, p.Sprintf("%d", s.CurrentPrice), p.Sprintf("%.0f%%", s.ProfitAndLossRatio()*100)})
	}
	table.Render()
}

func renderFunds(a assetAllocation) {
	table := tablewriter.NewWriter(os.Stdout)
	p := message.NewPrinter(message.MatchLanguage("en"))

	table.SetHeader([]string{"Category", "Name", "Current", "P/L Ratio"})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})
	for k, d := range a.details {
		for _, f := range d.funds {
			table.Append([]string{p.Sprint(k), f.Name, p.Sprintf("%.2f", f.CurrentPrice), p.Sprintf("%.0f%%", f.ProfitAndLossRatio()*100)})
		}
	}
	table.Render()
}
