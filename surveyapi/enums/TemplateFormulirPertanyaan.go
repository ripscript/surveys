package enums

type StatusForm string

const (
	OPEN StatusForm = "open"
)

func (t StatusForm) IsValid() bool {
	switch t {
	case OPEN:
		return true
	}
	return false
}

type QuestionType string

const (
	LONG_ANSWER      QuestionType = "long-answer"
	NUMBER           QuestionType = "number"
	MULTIPLE_CHOICES QuestionType = "multiple-choices"
	IMAGE_TEMPLATE   QuestionType = "image-template"
	MAPS             QuestionType = "maps"
)

func (t QuestionType) IsQuestionTypeValid() bool {
	switch t {
	case LONG_ANSWER, NUMBER, MULTIPLE_CHOICES, IMAGE_TEMPLATE, MAPS:
		return true
	}
	return false
}
