// Command mygo opens the thousand-row invoice table in MyGo native UI.
//
// Naive form: all 1000 rows render every frame, so this measures the default
// full-build path against the ownframe invoice table.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

const (
	invoiceWidth    = 1000
	invoiceHeight   = 650
	invoiceRows     = 1000
	invoiceSubtitle = "mygo · Go view → GPU, no HTML"
)

// Invoice themes: gray is the zebra stripe (white/gray), light is plain white.
const (
	invoiceThemeGray  = "gray"
	invoiceThemeLight = "light"
)

var (
	invoiceClients  = [...]string{"Acme Corp", "Globex", "Soylent", "Initech", "Umbrella", "Hooli", "Stark Industries", "Wayne Enterprises", "Massive Dynamic", "Cyberdyne"}
	invoiceItems    = [...]string{"Widget A", "Gadget B", "Service C", "License D", "Support E", "Consulting F", "Hosting G", "Audit H"}
	invoiceStatuses = [...]string{"Paid", "Pending", "Overdue"}
)

// invoiceRow is one invoice line. Keep field order identical to the ownframe
// copy: the rng call sequence below must match it exactly.
type invoiceRow struct {
	num        string
	id         string
	client     string
	date       string
	item       string
	qty        int
	unitCents  int
	totalCents int
	status     string
}

// makeInvoiceRows builds n deterministic rows from a fixed seed.
func makeInvoiceRows(n int) []invoiceRow {
	rng := rand.New(rand.NewPCG(1, 1))
	rows := make([]invoiceRow, n)
	for i := range rows {
		client := invoiceClients[rng.IntN(len(invoiceClients))]
		item := invoiceItems[rng.IntN(len(invoiceItems))]
		month := 1 + rng.IntN(12)
		day := 1 + rng.IntN(28)
		qty := 1 + rng.IntN(10)
		unit := 900 + rng.IntN(49101)
		status := invoiceStatuses[rng.IntN(len(invoiceStatuses))]
		rows[i] = invoiceRow{
			num:        "#" + fmt.Sprint(i+1),
			id:         fmt.Sprintf("INV-%04d", 1001+i),
			client:     client,
			date:       fmt.Sprintf("2026-%02d-%02d", month, day),
			item:       item,
			qty:        qty,
			unitCents:  unit,
			totalCents: qty * unit,
			status:     status,
		}
	}
	return rows
}

func formatCents(c int) string {
	return fmt.Sprintf("$%d.%02d", c/100, c%100)
}

func statusColor(status string) ui.Color {
	switch status {
	case "Paid":
		return ui.Hex("#15803d")
	case "Pending":
		return ui.Hex("#b45309")
	default:
		return ui.Hex("#b91c1c")
	}
}

// invoiceApp is the app state the window shows: rows, selection, and theme.
type invoiceApp struct {
	rows     []invoiceRow
	selected int
	theme    string
	query    string
}

func newInvoiceApp() *invoiceApp {
	return &invoiceApp{rows: makeInvoiceRows(invoiceRows), selected: -1, theme: invoiceThemeGray}
}

func (s *invoiceApp) rowBackground(i int) ui.Color {
	if i == s.selected {
		return ui.Hex("#dbeafe")
	}
	if s.theme == invoiceThemeLight {
		return ui.Hex("#ffffff")
	}
	if i%2 == 1 {
		return ui.Hex("#f3f4f6")
	}
	return ui.Hex("#ffffff")
}

func (s *invoiceApp) themeLabel() string {
	if s.theme == invoiceThemeLight {
		return "Theme: light — switch to gray"
	}
	return "Theme: gray — switch to light"
}

func (s *invoiceApp) themeText() string {
	if s.theme == invoiceThemeLight {
		return "Theme: light (plain white rows)"
	}
	return "Theme: gray (alternating white/gray rows)"
}

func (s *invoiceApp) grandTotal() int {
	sum := 0
	for _, r := range s.rows {
		sum += r.totalCents
	}
	return sum
}

func (s *invoiceApp) matchedRows() []int {
	if s.query == "" {
		order := make([]int, len(s.rows))
		for i := range order {
			order[i] = i
		}
		return order
	}
	q := strings.ToLower(s.query)
	var order []int
	for i := range s.rows {
		if strings.Contains(strings.ToLower(s.rows[i].client), q) {
			order = append(order, i)
		}
	}
	return order
}

func (s *invoiceApp) showingText() string {
	return "Showing " + fmt.Sprint(len(s.matchedRows())) + " of " + fmt.Sprint(len(s.rows))
}

// view builds the interface from the state, on the main thread, whenever the
// window needs a frame.
func (s *invoiceApp) view(c *ui.Context) {
	t := c.Theme()
	white := ui.Hex("#ffffff")
	muted := t.TextMuted
	ui.Column(c).Fill().Children(func() {
		ui.Column(c).Gap(4).Padding(12).Children(func() {
			ui.Text(c, "Invoice Table · 1000 rows").FontSize(18).Bold()
			ui.Text(c, invoiceSubtitle).FontSize(12).TextColor(muted)
			ui.TextInput(c, &s.query).Label("Filter by client...")
			ui.Text(c, s.showingText()).FontSize(13).TextColor(muted)
			ui.Text(c, s.selectedText()).FontSize(14)
			ui.Text(c, s.themeText()).FontSize(13).TextColor(muted)
			ui.Text(c, "Grand total: "+formatCents(s.grandTotal())).FontSize(14).Bold()
			if ui.Button(c, s.themeLabel()).Clicked() {
				if s.theme == invoiceThemeGray {
					s.theme = invoiceThemeLight
				} else {
					s.theme = invoiceThemeGray
				}
			}
		})
		ui.Row(c).Gap(12).Padding(8).Background(ui.Hex("#1c1915")).Children(func() {
			ui.Text(c, "Num").FontSize(13).Bold().Width(48).TextColor(white)
			ui.Text(c, "Invoice").FontSize(13).Bold().Width(96).TextColor(white)
			ui.Text(c, "Client").FontSize(13).Bold().Width(140).TextColor(white)
			ui.Text(c, "Date").FontSize(13).Bold().Width(96).TextColor(white)
			ui.Text(c, "Item").FontSize(13).Bold().Width(170).TextColor(white)
			ui.Text(c, "Qty").FontSize(13).Bold().Width(48).TextColor(white)
			ui.Text(c, "Total").FontSize(13).Bold().Width(96).TextColor(white)
			ui.Text(c, "Status").FontSize(13).Bold().Width(90).TextColor(white)
		})
		ui.Scroll(c).Fill().Children(func() {
			for _, i := range s.matchedRows() {
				r := s.rows[i]
				row := ui.Row(c).Gap(12).Padding(8).Background(s.rowBackground(i))
				row.Children(func() {
					ui.Text(c, r.num).FontSize(13).Width(48).TextColor(muted)
					ui.Text(c, r.id).FontSize(13).Width(96)
					ui.Text(c, r.client).FontSize(13).Width(140)
					ui.Text(c, r.date).FontSize(13).Width(96).TextColor(muted)
					ui.Text(c, r.item).FontSize(13).Width(170)
					ui.Text(c, fmt.Sprint(r.qty)).FontSize(13).Width(48)
					ui.Text(c, formatCents(r.totalCents)).FontSize(13).Width(96)
					ui.Text(c, r.status).FontSize(13).Width(90).TextColor(statusColor(r.status))
				})
				if row.Clicked() {
					s.selected = i
				}
			}
		})
	})
}

func (s *invoiceApp) selectedText() string {
	if s.selected < 0 || s.selected >= len(s.rows) {
		return "Selected: none"
	}
	r := s.rows[s.selected]
	return "Selected: " + r.id + " · " + r.client + " · " + formatCents(r.totalCents)
}

func runInvoice() {
	bench := flag.Bool("bench", false, "run headless benchmarks instead of opening a window")
	benchFile := flag.String("bench-file", "", "write benchmark results to this file as well as stdout")
	flag.Parse()

	if *bench {
		if err := runInvoiceBench(*benchFile); err != nil {
			log.Fatal(err)
		}
		return
	}

	s := newInvoiceApp()
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:   "Invoice Table",
			Width:   invoiceWidth,
			Height:  invoiceHeight,
			Content: ui.View(s.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// benchResult is one headless measurement row: wall time per op (headless, so
// wall time is CPU time), derived FPS, and memory deltas.
type benchResult struct {
	name       string
	ops        int
	nsPerOp    float64
	fps        float64
	msPerFrame float64
	bytesPerOp uint64
	allocsOp   uint64
}

// measure runs fn n times and reports wall time plus MemStats deltas.
func measure(name string, n int, fn func() error) (benchResult, error) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	for i := 0; i < n; i++ {
		if err := fn(); err != nil {
			return benchResult{}, err
		}
	}
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	ns := float64(elapsed.Nanoseconds()) / float64(n)
	return benchResult{
		name:       name,
		ops:        n,
		nsPerOp:    ns,
		fps:        1e9 / ns,
		msPerFrame: ns / 1e6,
		bytesPerOp: (after.TotalAlloc - before.TotalAlloc) / uint64(n),
		allocsOp:   (after.Mallocs - before.Mallocs) / uint64(n),
	}, nil
}

// runBench executes the headless benchmark set without opening a window and
// prints one line per row, also writing benchFile when set.
func runInvoiceBench(benchFile string) error {
	rows := make([]benchResult, 0, 6)
	build, err := measure("build-full-view", 20, func() error {
		s := newInvoiceApp()
		s.selected = 500
		tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
		if !tt.HasText("INV-2000") {
			return fmt.Errorf("missing row text")
		}
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, build)
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	targets := []string{"INV-1005", "INV-1006"}
	clicks := 0
	click, err := measure("click-select-reframe", 100, func() error {
		t := targets[clicks%len(targets)]
		clicks++
		return tt.Click(t)
	})
	if err != nil {
		return err
	}
	rows = append(rows, click)
	if s.selected != 5 {
		return fmt.Errorf("selected = %d, want 5", s.selected)
	}
	frame, err := measure("reframe-same-state", 200, func() error {
		tt.Frame()
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, frame)
	cold, err := measure("cold-build", 10, func() error {
		s := newInvoiceApp()
		tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
		if !tt.HasText("INV-2000") {
			return fmt.Errorf("missing row text")
		}
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, cold)
	resize, err := measure("resize", 50, func() error {
		tt.SetSize(1200, 800)
		tt.SetSize(invoiceWidth, invoiceHeight)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, resize)
	if err := tt.Click("Filter by client..."); err != nil {
		return err
	}
	typed := false
	typing, err := measure("typing", 50, func() error {
		if !typed {
			typed = true
			tt.Type("a")
			return nil
		}
		typed = false
		tt.Key(0, ui.KeyBackspace)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, typing)
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < 10; i++ {
		s := newInvoiceApp()
		tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
		if !tt.HasText("INV-2000") {
			return fmt.Errorf("missing row text")
		}
	}
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	out := ""
	for _, r := range rows {
		out += fmt.Sprintf("%-20s ops=%-4d ns/op=%-12.1f ms/frame=%-8.3f fps=%-10.1f B/op=%-10d allocs/op=%d\n",
			r.name, r.ops, r.nsPerOp, r.msPerFrame, r.fps, r.bytesPerOp, r.allocsOp)
	}
	out += fmt.Sprintf("sizes rows=%d grand-total=%s heap-alloc=%d heap-sys=%d\n",
		invoiceRows, formatCents(newInvoiceApp().grandTotal()), mem.HeapAlloc, mem.HeapSys)
	out += fmt.Sprintf("heap views=10-builds heap-retained=%d heap-objects=%d heap-growth=%d growth-objects=%d\n",
		after.HeapAlloc, after.HeapObjects,
		int64(after.HeapAlloc)-int64(before.HeapAlloc), int64(after.HeapObjects)-int64(before.HeapObjects))
	fmt.Print(out)
	if benchFile != "" {
		if err := os.WriteFile(benchFile, []byte(out), 0o600); err != nil {
			return err
		}
	}
	return nil
}
