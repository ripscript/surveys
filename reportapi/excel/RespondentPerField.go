package excel

import (
	"backend/reportapi/models"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// buildPerFieldSheet builds "Per Pertanyaan" sheet (Sheet 2).
// Mirrors ResultPerFieldSheet + summary_per_field.blade.php
func buildPerFieldSheet(f *excelize.File, survey *models.SurveyExport, answerOptions map[uint][]models.FormAnswerFieldExport) error {
	const sheetName = "Per Question"
	f.NewSheet(sheetName)

	fields := survey.FlowDetail.Form.Fields
	responses := survey.SurveyDataRespondents
	totalRespondents := len(responses)

	// Collect all form_response IDs
	formResponseIDs := make(map[uint]bool)
	for _, r := range responses {
		formResponseIDs[r.ID] = true
	}

	// ── Styles ──────────────────────────────────────────────────────────────
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 16, Bold: true},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	headerFillStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"15406A"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{
			Vertical: "center", Horizontal: "center", WrapText: true,
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
			Vertical: "center", Horizontal: "left", WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	dataCenterStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical: "center", Horizontal: "center", WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// ── Row 1-2: Title ────────────────────────────────────────────────────────
	f.MergeCell(sheetName, "A1", "D2")
	f.SetCellValue(sheetName, "A1", "HASIL "+strings.ToUpper(survey.Name))
	f.SetCellStyle(sheetName, "A1", "D2", titleStyle)
	f.SetRowHeight(sheetName, 1, 35)

	// ── Row 3: blank ─────────────────────────────────────────────────────────

	// ── Row 4-6: Metadata ────────────────────────────────────────────────────
	f.SetCellValue(sheetName, "A4", "Nama Survey")
	f.SetCellStyle(sheetName, "A4", "A4", boldStyle)
	f.SetCellValue(sheetName, "B4", survey.Name)

	f.SetCellValue(sheetName, "A5", "Tanggal Survey")
	f.SetCellStyle(sheetName, "A5", "A5", boldStyle)
	f.SetCellValue(sheetName, "B5", formatDateRange(survey.StartDate, survey.EndDate))

	f.SetCellValue(sheetName, "A6", "Total Responden")
	f.SetCellStyle(sheetName, "A6", "A6", boldStyle)
	f.SetCellValue(sheetName, "B6", fmt.Sprintf("%d Responden", totalRespondents))

	f.SetColWidth(sheetName, "A", "A", 22)
	f.SetColWidth(sheetName, "B", "B", 40)
	f.SetColWidth(sheetName, "C", "C", 18)
	f.SetColWidth(sheetName, "D", "D", 12)

	// ── Current row cursor ────────────────────────────────────────────────────
	currentRow := 8

	for _, field := range fields {
		// ── Question header block ─────────────────────────────────────────────
		f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow),
			fmt.Sprintf("Pertanyaan %d", field.Sequence))
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(currentRow), "A"+strconv.Itoa(currentRow), headerFillStyle)

		f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), field.Question)
		f.SetCellStyle(sheetName, "B"+strconv.Itoa(currentRow), "D"+strconv.Itoa(currentRow), headerFillStyle)
		f.MergeCell(sheetName, "B"+strconv.Itoa(currentRow), "D"+strconv.Itoa(currentRow))
		f.SetRowHeight(sheetName, currentRow, 30)
		currentRow++

		isOpenEnded := isOpenEndedField(field.Template)

		if isOpenEnded {
			// Count answered respondents for open-ended / image / maps / number
			countAnswered := countFieldAnswers(responses, field.ID)
			pct := 0.0
			if totalRespondents > 0 {
				pct = math.Round((float64(countAnswered)/float64(totalRespondents))*10000) / 100
			}

			f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "")
			f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), "Responden :")
			f.SetCellStyle(sheetName, "B"+strconv.Itoa(currentRow), "B"+strconv.Itoa(currentRow), boldStyle)
			f.SetCellValue(sheetName, "C"+strconv.Itoa(currentRow),
				fmt.Sprintf("'%d/%d", countAnswered, totalRespondents))
			f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow),
				fmt.Sprintf("'%.2f%%", pct))

			for _, col := range []string{"A", "B", "C", "D"} {
				f.SetCellStyle(sheetName, col+strconv.Itoa(currentRow), col+strconv.Itoa(currentRow), dataStyle)
			}
			currentRow++
		} else {
			// Choice-type: show each option with count and percentage
			f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "")
			f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), "Jawaban :")
			f.SetCellStyle(sheetName, "B"+strconv.Itoa(currentRow), "B"+strconv.Itoa(currentRow), boldStyle)
			for _, col := range []string{"A", "B", "C", "D"} {
				f.SetCellStyle(sheetName, col+strconv.Itoa(currentRow), col+strconv.Itoa(currentRow), dataStyle)
			}
			currentRow++

			options := answerOptions[field.ID]
			for _, opt := range options {
				countAnswered := countOptionAnswers(responses, field.ID, opt.ID)
				pct := 0.0
				if totalRespondents > 0 {
					pct = math.Round((float64(countAnswered)/float64(totalRespondents))*10000) / 100
				}

				f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "")
				f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), opt.Option)
				f.SetCellValue(sheetName, "C"+strconv.Itoa(currentRow),
					fmt.Sprintf("'%d/%d", countAnswered, totalRespondents))
				f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow),
					fmt.Sprintf("'%.2f%%", pct))

				f.SetCellStyle(sheetName, "A"+strconv.Itoa(currentRow), "A"+strconv.Itoa(currentRow), dataStyle)
				f.SetCellStyle(sheetName, "B"+strconv.Itoa(currentRow), "B"+strconv.Itoa(currentRow), dataStyle)
				f.SetCellStyle(sheetName, "C"+strconv.Itoa(currentRow), "C"+strconv.Itoa(currentRow), dataCenterStyle)
				f.SetCellStyle(sheetName, "D"+strconv.Itoa(currentRow), "D"+strconv.Itoa(currentRow), dataCenterStyle)
				currentRow++
			}
		}

		// blank separator row between questions
		currentRow++
	}

	return nil
}

// countFieldAnswers counts responses for open-ended/image/maps/number fields
func countFieldAnswers(responses []models.SurveyDataRespondentExport, fieldID uint) int {
	count := 0
	for _, resp := range responses {
		for _, fdr := range resp.FieldDataResponses {
			if fdr.FormFieldID == fieldID && fdr.Answer != "" {
				count++
				break
			}
		}
	}
	return count
}

// countOptionAnswers counts how many respondents chose a specific option
func countOptionAnswers(responses []models.SurveyDataRespondentExport, fieldID uint, optionID uint) int {
	optIDStr := strconv.Itoa(int(optionID))
	count := 0
	for _, resp := range responses {
		for _, fdr := range resp.FieldDataResponses {
			if fdr.FormFieldID == fieldID && fdr.Answer == optIDStr {
				count++
				break
			}
		}
	}
	return count
}
