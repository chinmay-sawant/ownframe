package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func main() {
	filterFlag := flag.String("filter", "", "Filter scenarios by substring in ID or Name")
	warmCount := flag.Int("warm", 5, "Number of warm iterations to average")
	tableFormat := flag.String("format", "table", "Output format: table, markdown, or csv")
	flag.Parse()

	scenarios := AllScenarios()
	var results []Result

	fmt.Println("================================================================================")
	fmt.Println("OwnFrame Application Performance Benchmark Suite (10 Real-World Scenarios)")
	fmt.Println("================================================================================")
	fmt.Printf("Running %d performance scenarios (warm iterations: %d)...\n\n", len(scenarios), *warmCount)

	for i, s := range scenarios {
		if *filterFlag != "" && !strings.Contains(s.ID, *filterFlag) && !strings.Contains(s.Name, *filterFlag) {
			continue
		}

		fmt.Printf("[%02d/10] Measuring %-42s (%s)... ", i+1, s.Name, s.Level)
		res, err := MeasureScenario(s, *warmCount)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Done (Cold: %.2fms | Warm: %.2fms | Boxes: %d | Ops: %d)\n",
			ms(res.ColdRender), ms(res.WarmRedraw), res.Boxes, res.Ops)
		results = append(results, res)
	}

	fmt.Println()
	switch *tableFormat {
	case "markdown":
		printMarkdownTable(results)
	case "csv":
		printCSV(results)
	default:
		printTerminalTable(results)
	}
}

func ms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}

func kb(bytes uint64) float64 {
	return float64(bytes) / 1024.0
}

func printTerminalTable(results []Result) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tSCENARIO\tLEVEL\tITEMS\tBOXES\tOPS\tCOLD (ms)\tWARM (ms)\tMUTATE (ms)\tRESIZE (ms)\tSTYLE (ms)\tLAYOUT (ms)\tALLOC (KB)")
	fmt.Fprintln(w, "--\t--------\t-----\t-----\t-----\t---\t---------\t---------\t-----------\t-----------\t----------\t-----------\t----------")

	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.1f\n",
			r.ID,
			r.Name,
			r.Level,
			r.ItemCount,
			r.Boxes,
			r.Ops,
			ms(r.ColdRender),
			ms(r.WarmRedraw),
			ms(r.MutationRedraw),
			ms(r.ResizeRelayout),
			ms(r.StyleTime),
			ms(r.LayoutTime),
			kb(r.AllocFrame),
		)
	}
	w.Flush()
}

func printMarkdownTable(results []Result) {
	fmt.Println("| ID | Scenario Name | Level | Items | Boxes | Ops | Cold (ms) | Warm (ms) | Mutate (ms) | Resize (ms) | Layout (ms) | Alloc (KB) |")
	fmt.Println("|---|---|---|---|---|---|---|---|---|---|---|---|")
	for _, r := range results {
		fmt.Printf("| %s | %s | %s | %d | %d | %d | %.2f | %.2f | %.2f | %.2f | %.2f | %.1f |\n",
			r.ID,
			r.Name,
			r.Level,
			r.ItemCount,
			r.Boxes,
			r.Ops,
			ms(r.ColdRender),
			ms(r.WarmRedraw),
			ms(r.MutationRedraw),
			ms(r.ResizeRelayout),
			ms(r.LayoutTime),
			kb(r.AllocFrame),
		)
	}
}

func printCSV(results []Result) {
	fmt.Println("ID,Name,Level,Items,Boxes,Ops,ColdMs,WarmMs,MutateMs,ResizeMs,TemplateMs,StyleMs,LayoutMs,PaintMs,AllocKB")
	for _, r := range results {
		fmt.Printf("%s,\"%s\",%s,%d,%d,%d,%.3f,%.3f,%.3f,%.3f,%.3f,%.3f,%.3f,%.3f,%.2f\n",
			r.ID,
			r.Name,
			r.Level,
			r.ItemCount,
			r.Boxes,
			r.Ops,
			ms(r.ColdRender),
			ms(r.WarmRedraw),
			ms(r.MutationRedraw),
			ms(r.ResizeRelayout),
			ms(r.TemplateTime),
			ms(r.StyleTime),
			ms(r.LayoutTime),
			ms(r.PaintTime),
			kb(r.AllocFrame),
		)
	}
}
