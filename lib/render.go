package yajirobe

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/text/message"
)

func renderStocks(sx []Stock) {
	table := tablewriter.NewWriter(os.Stdout)
	p := message.NewPrinter(message.MatchLanguage("en"))

	table.SetHeader([]string{"Name", "Current", "P/L"})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	for _, s := range sx {
		table.Append([]string{s.Name, p.Sprintf("%d", s.CurrentPrice), p.Sprintf("%.1f%%", s.ProfitAndLossRatio()*100)})
	}
	table.Render()
}

// Render 画面に書き出す
func (a *AssetAllocation) Render() {
	table := tablewriter.NewWriter(os.Stdout)
	p := message.NewPrinter(message.MatchLanguage("en"))

	//table.SetAutoMergeCells(true)
	table.SetHeader([]string{
		"Class",
		"Target",
		"Actual",
		"Current",
		"Diff",
		"P/L",
	})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	table.Append([]string{
		"全体", // Class
		"",   // Target
		"",   // Actual
		p.Sprintf("%.2f", a.cprice), // Current
		"", // Diff
		p.Sprintf("%.1f%%", a.cprice/a.aprice*100-100), // P/L
	})

	for _, class := range AssetClasses {
		detail, e := a.details[class]
		if !e {
			continue
		}

		row := []string{
			class.String(),                                 // Class
			fmt.Sprintf("%.1f%%", detail.targetRatio*100),  // Target
			fmt.Sprintf("%.1f%%", detail.currentRatio*100), // Actual
			p.Sprintf("%.2f", detail.cprice),               // Current
			p.Sprintf("%.0f", detail.diffPrice),            // Diff
			p.Sprintf("%.1f%%", detail.pl*100),             // P/L
		}
		table.Append(row)
	}

	table.Render()
}
