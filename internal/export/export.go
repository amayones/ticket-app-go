package export

// Shared export kernel: CSV (stdlib) + PDF (gofpdf) for dashboard summaries.
// Used by features/organizer and features/seller so both exports look alike.
import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

// Table is one titled string table rendered to CSV and PDF.
type Table struct {
	Title   string
	Headers []string
	Rows    [][]string
}

// CSV renders tables as UTF-8 CSV (BOM for Excel) separated by blank lines.
func CSV(tables []Table) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("\xef\xbb\xbf")
	w := csv.NewWriter(&buf)
	for i, t := range tables {
		if i > 0 {
			if err := w.Write([]string{}); err != nil {
				return nil, err
			}
		}
		if err := w.Write([]string{t.Title}); err != nil {
			return nil, err
		}
		if err := w.Write(t.Headers); err != nil {
			return nil, err
		}
		if err := w.WriteAll(t.Rows); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// PDF renders tables into a simple portrait A4 PDF (one table after another).
func PDF(docTitle string, tables []Table) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(docTitle, false)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, docTitle)
	pdf.Ln(14)
	for _, t := range tables {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, t.Title)
		pdf.Ln(10)
		if len(t.Headers) == 0 {
			continue
		}
		widths := colWidths(t)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(240, 240, 240)
		for i, h := range t.Headers {
			pdf.CellFormat(widths[i], 7, h, "1", 0, "", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(255, 255, 255)
		for _, row := range t.Rows {
			if pdf.GetY() > 270 {
				pdf.AddPage()
			}
			for i := range t.Headers {
				v := ""
				if i < len(row) {
					v = row[i]
				}
				pdf.CellFormat(widths[i], 6, v, "1", 0, "", false, 0, "")
			}
			pdf.Ln(-1)
		}
		pdf.Ln(4)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// colWidths splits the 190mm printable width evenly across columns.
func colWidths(t Table) []float64 {
	n := len(t.Headers)
	if n == 0 {
		return nil
	}
	w := make([]float64, n)
	for i := range w {
		w[i] = 190 / float64(n)
	}
	return w
}
