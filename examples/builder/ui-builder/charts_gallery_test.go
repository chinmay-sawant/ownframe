package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestChartSetsSwitch(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		action string
		want   string
	}{
		{action: "tab:chartjs", want: "Chart.js"},
		{action: "tab:shadcnchart", want: "shadcn charts"},
		{action: "tab:recharts", want: "Recharts"},
		{action: "tab:tremor", want: "Tremor"},
		{action: "tab:echarts", want: "ECharts"},
		{action: "tab:apexcharts", want: "ApexCharts"},
		{action: "tab:nivo", want: "Nivo"},
	}
	for _, tc := range cases {
		if err := pagetest.ClickAction(ctx, a.page, tc.action); err != nil {
			t.Fatalf("click %s: %v", tc.action, err)
		}
		if html := a.page.HTML(); !strings.Contains(html, tc.want) {
			t.Fatalf("after %s, HTML missing %q", tc.action, tc.want)
		}
	}
}
