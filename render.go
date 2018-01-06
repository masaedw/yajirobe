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

func renderAllocation(a AssetAllocation, t AllocationTarget) {
	table := tablewriter.NewWriter(os.Stdout)
	p := message.NewPrinter(message.MatchLanguage("en"))

	//table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Class", "Target", "Actual", "Name", "Current", "P/L"})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
	})

	table.Append([]string{"全体", "", "", "", p.Sprintf("%.2f", a.cprice), p.Sprintf("%.1f%%", a.cprice/a.aprice*100-100)})

	for _, class := range AssetClasses {
		detail, de := a.details[class]
		target, te := t[class]

		if !de && !te {
			continue
		}

		if !de {
			row := []string{
				class.String(),                    // Class
				fmt.Sprintf("%.1f%%", target*100), // Target
				"0.0%", // Actual
				"",     // Name
				"0",    // Current
				"-",    // P/L
			}
			table.Append(row)
		} else {
			row := []string{
				class.String(),                                    // Class
				fmt.Sprintf("%.1f%%", target*100),                 // Target
				fmt.Sprintf("%.1f%%", detail.cprice/a.cprice*100), // Actual
				"", // Name
				p.Sprintf("%.2f", detail.cprice),                         // Current
				p.Sprintf("%.1f%%", detail.cprice/detail.aprice*100-100), // P/L
			}
			table.Append(row)
		}
	}

	table.Render()
}
