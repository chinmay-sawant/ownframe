package benchscore

import (
	"fmt"
	"io"
	"math"
	"text/tabwriter"
)

const (
	tableMinWidth = 0
	tableTabWidth = 4
	tablePadding  = 2
)

// Write prints the full scorecard and a second table with the exact failed
// rows. The summary line reports every compared row.
func (r Report) Write(w io.Writer) error {
	mainTable := tabwriter.NewWriter(w, tableMinWidth, tableTabWidth, tablePadding, ' ', 0)
	if _, err := fmt.Fprintln(mainTable, "CLASS\tBENCHMARK\tMETRIC\tREFERENCE\tCANDIDATE\tRATIO\tRESULT"); err != nil {
		return err
	}
	for _, row := range r.Rows {
		if _, err := fmt.Fprintf(mainTable, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.Class,
			row.Benchmark,
			row.Metric,
			referenceValue(&row),
			candidateValue(&row),
			formatRatio(&row),
			row.Status,
		); err != nil {
			return err
		}
	}
	if err := mainTable.Flush(); err != nil {
		return err
	}

	failed := r.FailedRows()
	if len(failed) == 0 {
		_, err := fmt.Fprintf(w, "\n%d rows checked, all within budget\n", len(r.Rows))
		return err
	}

	if _, err := fmt.Fprintln(w, "\nFAILED ROWS"); err != nil {
		return err
	}
	failureTable := tabwriter.NewWriter(w, tableMinWidth, tableTabWidth, tablePadding, ' ', 0)
	if _, err := fmt.Fprintln(failureTable, "CLASS\tBENCHMARK\tMETRIC\tREFERENCE\tCANDIDATE\tRATIO\tREASON"); err != nil {
		return err
	}
	for _, row := range failed {
		if _, err := fmt.Fprintf(failureTable, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.Class,
			row.Benchmark,
			row.Metric,
			referenceValue(&row),
			candidateValue(&row),
			formatRatio(&row),
			row.Reason,
		); err != nil {
			return err
		}
	}
	if err := failureTable.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n%d of %d rows failed\n", len(failed), len(r.Rows))
	return err
}

func referenceValue(row *Row) string {
	if !row.HasReference {
		return "-"
	}
	return formatValue(row.Metric, row.Reference)
}

func candidateValue(row *Row) string {
	if !row.HasCandidate {
		return "-"
	}
	return formatValue(row.Metric, row.Candidate)
}

func formatValue(metric string, value float64) string {
	switch metric {
	case MetricTime:
		return fmt.Sprintf("%.2f ns", value)
	case MetricBytes:
		return fmt.Sprintf("%.0f B", value)
	case MetricAllocs:
		return fmt.Sprintf("%.0f", value)
	default:
		return fmt.Sprintf("%.0f B", value)
	}
}

func formatRatio(row *Row) string {
	if !row.HasReference || !row.HasCandidate {
		return "-"
	}
	if math.IsInf(row.Ratio, 1) {
		return "inf"
	}
	return fmt.Sprintf("%.2fx", row.Ratio)
}
