package main

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

func renderAllocation(a AssetAllocation) {
	table := tablewriter.NewWriter(os.Stdout)
	p := message.NewPrinter(message.MatchLanguage("en"))

	//table.SetAutoMergeCells(true)
	table.SetHeader([]string{
		"Class",
		"Target",
		"Actual",
		"Current",
		"P/L",
	})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	table.Append([]string{
		"全体", // Class
		"",   // Target
		"",   // Actual
		p.Sprintf("%.2f", a.cprice),                    // Current
		p.Sprintf("%.1f%%", a.cprice/a.aprice*100-100), // P/L
	})

	for class, detail := range a.details {
		row := []string{
			class.String(),                                 // Class
			fmt.Sprintf("%.1f%%", detail.targetRatio*100),  // Target
			fmt.Sprintf("%.1f%%", detail.currentRatio*100), // Actual
			p.Sprintf("%.2f", detail.cprice),               // Current
			p.Sprintf("%.1f%%", detail.pl-100),             // P/L
		}
		table.Append(row)
	}

	table.Render()
}
