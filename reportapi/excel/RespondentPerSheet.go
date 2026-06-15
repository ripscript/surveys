package excel

import (
	"backend/reportapi/models"
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// buildPerRespondentSheet builds "Rekap Hasil Per Responden" sheet (Sheet 1).
// Mirrors ResultPerRespondentSheet + summary_per_respondent.blade.php
func buildPerRespondentSheet(f *excelize.File, survey *models.SurveyExport, answerOptions map[uint][]models.FormAnswerFieldExport) error {
	const sheetName = "Per Respondent"
	f.NewSheet(sheetName)

	fields := survey.FlowDetail.Form.Fields
	responses := survey.SurveyDataRespondents

	// ── Styles ──────────────────────────────────────────────────────────────
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 16, Bold: true, Color: "000000"},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center", WrapText: true},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"15406A"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			Horizontal: "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			Horizontal: "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	dataLeftStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			Horizontal: "left",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	metaKeyStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})

	// ── Total columns: No + Respondent + Kecamatan + Kelurahan + RW + RT + fields ──
	totalCols := 6 + len(fields)
	lastColLetter, _ := excelize.ColumnNumberToName(totalCols)

	// ── Row 1-2: Title (merged across all columns) ───────────────────────────
	titleCell := fmt.Sprintf("A1:%s2", lastColLetter)
	f.MergeCell(sheetName, "A1", fmt.Sprintf("%s2", lastColLetter))
	f.SetCellValue(sheetName, "A1", "HASIL "+strings.ToUpper(survey.Name))
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s2", lastColLetter), titleStyle)
	f.SetRowHeight(sheetName, 1, 35)
	_ = titleCell

	// ── Row 3: blank ─────────────────────────────────────────────────────────

	// ── Row 4-6: Metadata ────────────────────────────────────────────────────
	f.MergeCell(sheetName, "A4", "B4")
	f.MergeCell(sheetName, "A5", "B5")
	f.MergeCell(sheetName, "A6", "B6")

	f.SetCellStyle(sheetName, "A4", "B4", metaKeyStyle)
	f.SetCellStyle(sheetName, "A5", "B5", metaKeyStyle)
	f.SetCellStyle(sheetName, "A6", "B6", metaKeyStyle)

	f.SetCellValue(sheetName, "A4", "Nama Survey")
	f.SetCellValue(sheetName, "C4", survey.Name)

	f.SetCellValue(sheetName, "A5", "Tanggal Survey")
	f.SetCellValue(sheetName, "C5", formatDateRange(survey.StartDate, survey.EndDate))

	f.SetCellValue(sheetName, "A6", "Total Responden")
	f.SetCellValue(sheetName, "C6", fmt.Sprintf("%d Responden", len(responses)))

	// ── Row 7: blank ─────────────────────────────────────────────────────────

	// ── Row 8: Table headers ──────────────────────────────────────────────────
	headers := []string{"No.", "Responden", "Kecamatan", "Kelurahan", "RW", "RT"}
	for _, field := range fields {
		headers = append(headers, fmt.Sprintf("%d. %s", field.Sequence, field.Question))
	}

	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := col + "8"
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, 8, 40)

	// Column widths
	colWidths := map[int]float64{1: 6, 2: 30, 3: 20, 4: 20, 5: 8, 6: 8}
	for i := 1; i <= 6; i++ {
		col, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheetName, col, col, colWidths[i])
	}
	for i := range fields {
		col, _ := excelize.ColumnNumberToName(7 + i)
		f.SetColWidth(sheetName, col, col, 25)
	}

	// ── Rows 9+: Data ─────────────────────────────────────────────────────────
	for rowIdx, resp := range responses {
		excelRow := 9 + rowIdx

		// Build a lookup map: fieldID → FieldDataResponse
		fieldAnswerMap := make(map[uint]*models.FieldDataResponseExport)
		for i := range resp.FieldDataResponses {
			fdr := &resp.FieldDataResponses[i]
			fieldAnswerMap[fdr.FormFieldID] = fdr
		}

		// Fixed columns
		rowData := []interface{}{
			rowIdx + 1,
			resp.Respondent.Name,
			safeStr(resp.Respondent.Kecamatan, func(k *models.KecamatanExport) string { return k.SubDistrictName }),
			safeStr(resp.Respondent.Kelurahan, func(k *models.KelurahanExport) string { return k.VillageName }),
			safeStr(resp.Respondent.RW, func(k *models.DataRwExport) string { return k.NamaRw }),
			safeStr(resp.Respondent.RT, func(k *models.DataRtExport) string { return k.NamaRt }),
		}

		// Dynamic field columns
		for _, field := range fields {
			answer := fieldAnswerMap[field.ID]
			rowData = append(rowData, resolveAnswer(field, answer, answerOptions))
		}

		for colIdx, val := range rowData {
			col, _ := excelize.ColumnNumberToName(colIdx + 1)
			cell := col + strconv.Itoa(excelRow)
			f.SetCellValue(sheetName, cell, val)
			// Use left-aligned style for text columns, center for No.
			if colIdx == 0 {
				f.SetCellStyle(sheetName, cell, cell, dataStyle)
			} else {
				f.SetCellStyle(sheetName, cell, cell, dataLeftStyle)
			}
		}
		f.SetRowHeight(sheetName, excelRow, 20)
	}

	return nil
}
