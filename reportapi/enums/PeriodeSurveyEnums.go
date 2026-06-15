package enums

type PeriodeSurveyType int32

const (
	TRIWULAN      PeriodeSurveyType = 1
	SEMESTER      PeriodeSurveyType = 2
	TAHUNAN       PeriodeSurveyType = 3
	TANPA_PERIODE PeriodeSurveyType = 4
)

func (t PeriodeSurveyType) IsPeriodeSurveyValid() bool {
	switch t {
	case TRIWULAN, SEMESTER, TAHUNAN, TANPA_PERIODE:
		return true
	}
	return false
}

func IsPeriodeSurveyExist(p PeriodeSurveyType) bool {
	switch p {
	case TRIWULAN, SEMESTER, TAHUNAN, TANPA_PERIODE:
		return true
	}
	return false
}

func PeriodeSurveyToInt64(p PeriodeSurveyType) int64 {
	switch p {
	case TRIWULAN:
		return 1
	case SEMESTER:
		return 2
	case TAHUNAN:
		return 3
	case TANPA_PERIODE:
		return 4
	default:
		return 0
	}
}

func PeriodeSurveyToString(p PeriodeSurveyType) string {
	switch p {
	case TRIWULAN:
		return "1"
	case SEMESTER:
		return "2"
	case TAHUNAN:
		return "3"
	case TANPA_PERIODE:
		return "4"
	default:
		return "0"
	}
}
