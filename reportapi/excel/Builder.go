package excel

import (
	"backend/reportapi/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// indonesianMonth converts a month number to Indonesian month name
var indonesianMonths = map[int]string{
	1:  "Januari",
	2:  "Februari",
	3:  "Maret",
	4:  "April",
	5:  "Mei",
	6:  "Juni",
	7:  "Juli",
	8:  "Agustus",
	9:  "September",
	10: "Oktober",
	11: "November",
	12: "Desember",
}

func formatDateRange(start, end time.Time) string {
	startStr := fmt.Sprintf("%02d %s, %d",
		start.Day(), indonesianMonths[int(start.Month())], start.Year())
	endStr := fmt.Sprintf("%02d %s, %d",
		end.Day(), indonesianMonths[int(end.Month())], end.Year())
	return startStr + " - " + endStr
}

// isOpenEndedField returns true for fields that don't use answer options
func isOpenEndedField(template string) bool {
	switch template {
	case "long-answer", "number", "image-template", "maps":
		return true
	}
	return false
}

// resolveAnswer returns the display string for a field answer
func resolveAnswer(field models.FormFieldExport, answer *models.FieldDataResponseExport, answerOptions map[uint][]models.FormAnswerFieldExport) string {
	if answer == nil {
		if field.Template == "image-template" {
			return ""
		}
		return "-"
	}

	switch field.Template {
	case "long-answer", "number":
		return answer.Answer

	case "image-template":
		if answer.Answer == "" {
			return ""
		}
		// Decode JSON array of image paths
		var images []string
		if err := json.Unmarshal([]byte(answer.Answer), &images); err != nil {
			return "Gambar Tidak Tersedia."
		}
		if len(images) == 0 {
			return "Gambar Tidak Tersedia."
		}
		// Return image URLs separated by newline
		urls := make([]string, 0, len(images))
		for _, img := range images {
			urls = append(urls, fmt.Sprintf("[Gambar] %s", img))
		}
		return strings.Join(urls, "\n")

	case "maps":
		if answer.LocationAddress == "" {
			return "-"
		}
		// Split by " | " and join with newline (like the blade template's <br>)
		parts := strings.Split(answer.LocationAddress, " | ")
		for i, p := range parts {
			parts[i] = "- " + strings.TrimSpace(p)
		}
		return strings.Join(parts, "\n")

	default:
		// Choice-type: find the option label by ID
		opts := answerOptions[field.ID]
		for _, opt := range opts {
			if fmt.Sprintf("%d", opt.ID) == answer.Answer {
				return opt.Option
			}
		}
		if answer.Answer == "" {
			return "Pertanyaan Dilewati"
		}
		return "Pertanyaan Dilewati"
	}
}

// safeStr safely dereferences a pointer and applies a getter, returning "" if nil
func safeStr[T any](ptr *T, getter func(*T) string) string {
	if ptr == nil {
		return ""
	}
	return getter(ptr)
}

// BuildSurveyExcel creates a multi-sheet Excel workbook ([]byte) from a survey.
// Returns the raw .xlsx bytes ready to be written to a file or HTTP response.
func BuildSurveyExcel(survey *models.SurveyExport, answerOptions map[uint][]models.FormAnswerFieldExport) ([]byte, error) {
	f := excelize.NewFile()

	// Remove default "Sheet1"
	f.DeleteSheet("Sheet1")

	if err := buildPerRespondentSheet(f, survey, answerOptions); err != nil {
		return nil, fmt.Errorf("sheet per respondent: %w", err)
	}

	if err := buildPerFieldSheet(f, survey, answerOptions); err != nil {
		return nil, fmt.Errorf("sheet per field: %w", err)
	}

	// Set first sheet as active
	if idx, err := f.GetSheetIndex("Per Respondent"); err == nil {
		f.SetActiveSheet(idx)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write excel buffer: %w", err)
	}
	return buf.Bytes(), nil
}

// GenerateFilename generates the .xlsx filename based on wilayah type, same as Laravel
func GenerateFilename(surveyName string, wilayah string, kecamatanName, kelurahanName, rwName string) string {
	date := time.Now().Format("02-01-2006")
	slug := toSlug(surveyName)

	switch wilayah {
	case "kecamatan":
		return fmt.Sprintf("%s_Hasil_%s_(%s).xlsx", slug, date, kecamatanName)
	case "kelurahan":
		return fmt.Sprintf("%s_Hasil_%s_(%s_%s).xlsx", slug, date, kecamatanName, kelurahanName)
	case "rw":
		return fmt.Sprintf("%s_Hasil_%s_(%s_%s_%s).xlsx", slug, date, kecamatanName, kelurahanName, rwName)
	default:
		return fmt.Sprintf("%s_Hasil_%s.xlsx", slug, date)
	}
}

// toSlug converts a string to a URL/filename-safe slug (equivalent to Laravel str_slug)
func toSlug(s string) string {
	s = strings.ToLower(s)
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			result.WriteRune('-')
		}
	}
	// Collapse multiple dashes
	slug := result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}
