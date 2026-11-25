package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ---- CSV ----

// ExportCSV writes the used range as RFC 4180 CSV using displayed text.
func (s *Sheet) ExportCSV(w io.Writer) error { return s.exportCSV(w, false) }

// ExportCSVFormulas writes the used range with formulas as typed ("=SUM(...)").
func (s *Sheet) ExportCSVFormulas(w io.Writer) error { return s.exportCSV(w, true) }

func (s *Sheet) exportCSV(w io.Writer, formulas bool) error {
	cw := csv.NewWriter(w)
	if len(s.cells) > 0 {
		u := s.UsedRange()
		for row := 0; row <= u.Max.Row; row++ {
			rec := make([]string, u.Max.Col+1)
			for col := range rec {
				if formulas {
					rec[col] = s.Raw(row, col)
				} else {
					rec[col] = s.Display(row, col)
				}
			}
			if err := cw.Write(rec); err != nil {
				return err
			}
		}
	}
	cw.Flush()
	return cw.Error()
}

// ImportCSV reads CSV (or TSV when the first line has tabs and no commas) into a
// new sheet, which becomes the active sheet.
func (w *Workbook) ImportCSV(r io.Reader, name string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	text := strings.TrimPrefix(string(data), "\ufeff")
	cr := csv.NewReader(strings.NewReader(text))
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	first, _, _ := strings.Cut(text, "\n")
	if strings.Contains(first, "\t") && !strings.Contains(first, ",") {
		cr.Comma = '\t'
	}
	recs, err := cr.ReadAll()
	if err != nil {
		return err
	}
	if name == "" {
		name = "Imported"
	}
	s := w.addSheetNamed(name)
	for row, rec := range recs {
		if row >= MaxRows {
			break
		}
		for col, f := range rec {
			if col >= MaxCols {
				break
			}
			if len(f) > 1 && f[0] == '0' && xlAllDigits(f) {
				f = "'" + f
			}
			s.setRawQuiet(CellRef{row, col}, f)
		}
	}
	w.Active = len(w.Sheets) - 1
	w.reloadAll()
	return nil
}

func xlAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

// ---- JSON ----

type xlJCell struct {
	R, C int
	V    string
}

type xlJStyle struct {
	R, C int
	S    Style
}

type xlJNote struct {
	R, C int
	T    string
}

type xlJCond struct {
	Range string
	Rule  CondRule
	Off   bool `json:",omitempty"`
}

type xlJValid struct {
	Range    string
	Kind     string `json:",omitempty"` // "" means list
	List     []string
	Min, Max float64 `json:",omitempty"`
}

type xlJSheet struct {
	Name      string
	TabColor  string `json:",omitempty"`
	Cells     []xlJCell
	Styles    []xlJStyle `json:",omitempty"`
	Comments  []xlJNote  `json:",omitempty"`
	ColW      map[int]int
	RowH      map[int]int
	HiddenR   []int            `json:",omitempty"`
	HiddenC   []int            `json:",omitempty"`
	FreezeR   int              `json:",omitempty"`
	FreezeC   int              `json:",omitempty"`
	FilterHdr int              `json:",omitempty"`
	Filters   map[int][]string `json:",omitempty"`
	Merges    []string         `json:",omitempty"`
	Conds     []xlJCond        `json:",omitempty"`
	Valid     []xlJValid       `json:",omitempty"`
}

type xlJWorkbook struct {
	Version int
	Active  int
	Sheets  []xlJSheet
}

func sortedRefs[T any](m map[CellRef]T) []CellRef {
	refs := make([]CellRef, 0, len(m))
	for r := range m {
		refs = append(refs, r)
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Row != refs[j].Row {
			return refs[i].Row < refs[j].Row
		}
		return refs[i].Col < refs[j].Col
	})
	return refs
}

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// MarshalJSON saves every sheet with formulas, styles, sizes, merges, freeze
// panes, filters, validation and conditional rules.
func (w *Workbook) MarshalJSON() ([]byte, error) {
	out := xlJWorkbook{Version: 1, Active: w.Active}
	for _, s := range w.Sheets {
		js := xlJSheet{Name: s.Name, TabColor: s.TabColor, ColW: s.colW, RowH: s.rowH,
			HiddenR: sortedKeys(s.hiddenRows), HiddenC: sortedKeys(s.hiddenCols),
			FreezeR: s.freezeR, FreezeC: s.freezeC, FilterHdr: s.filterHdr}
		js.Cells = []xlJCell{}
		for _, ref := range sortedRefs(s.cells) {
			js.Cells = append(js.Cells, xlJCell{ref.Row, ref.Col, s.cells[ref].raw})
		}
		for _, ref := range sortedRefs(s.styles) {
			js.Styles = append(js.Styles, xlJStyle{ref.Row, ref.Col, s.styles[ref]})
		}
		for _, ref := range sortedRefs(s.comments) {
			js.Comments = append(js.Comments, xlJNote{ref.Row, ref.Col, s.comments[ref]})
		}
		if len(s.filters) > 0 {
			js.Filters = map[int][]string{}
			for col, set := range s.filters {
				var vals []string
				for v, ok := range set {
					if ok {
						vals = append(vals, v)
					}
				}
				sort.Strings(vals)
				js.Filters[col] = vals
			}
		}
		for _, m := range s.merges {
			js.Merges = append(js.Merges, m.String())
		}
		for _, e := range s.conds {
			js.Conds = append(js.Conds, xlJCond{e.rng.String(), e.rule, e.off})
		}
		for _, v := range s.valid {
			js.Valid = append(js.Valid, xlJValid{Range: v.rng.String(), Kind: v.kind, List: v.list, Min: v.min, Max: v.max})
		}
		out.Sheets = append(out.Sheets, js)
	}
	return json.MarshalIndent(out, "", " ")
}

// UnmarshalWorkbook loads a workbook saved by MarshalJSON.
func UnmarshalWorkbook(b []byte) (*Workbook, error) {
	var in xlJWorkbook
	if err := json.Unmarshal(b, &in); err != nil {
		return nil, err
	}
	if len(in.Sheets) == 0 {
		return nil, fmt.Errorf("workbook has no sheets")
	}
	w := &Workbook{}
	for _, js := range in.Sheets {
		s := newSheet(w, js.Name)
		s.TabColor = js.TabColor
		w.Sheets = append(w.Sheets, s)
		for _, st := range js.Styles {
			if st.R >= 0 && st.R < MaxRows && st.C >= 0 && st.C < MaxCols {
				s.styles[CellRef{st.R, st.C}] = st.S.migrated()
			}
		}
		for _, c := range js.Cells {
			s.setRawQuiet(CellRef{c.R, c.C}, c.V)
		}
		for _, n := range js.Comments {
			s.SetComment(n.R, n.C, n.T)
		}
		for k, v := range js.ColW {
			s.colW[k] = v
		}
		for k, v := range js.RowH {
			s.rowH[k] = v
		}
		for _, r := range js.HiddenR {
			s.hiddenRows[r] = true
		}
		for _, c := range js.HiddenC {
			s.hiddenCols[c] = true
		}
		s.freezeR, s.freezeC, s.filterHdr = js.FreezeR, js.FreezeC, js.FilterHdr
		for col, vals := range js.Filters {
			set := map[string]bool{}
			for _, v := range vals {
				set[v] = true
			}
			s.filters[col] = set
		}
		for _, m := range js.Merges {
			if r, ok := ParseRange(m); ok {
				s.merges = append(s.merges, r)
			}
		}
		for _, c := range js.Conds {
			if r, ok := ParseRange(c.Range); ok {
				s.AddCondFormat(r, c.Rule)
				s.conds[len(s.conds)-1].off = c.Off
			}
		}
		for _, v := range js.Valid {
			r, ok := ParseRange(v.Range)
			switch {
			case !ok:
			case v.Kind == "" || v.Kind == "list":
				s.SetValidation(r, v.List)
			default:
				s.SetValidationBounds(r, v.Kind, v.Min, v.Max)
			}
		}
	}
	w.Active = max(0, min(in.Active, len(w.Sheets)-1))
	w.reloadAll()
	return w, nil
}
