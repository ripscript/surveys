package customValidator

import (
	"backend/masterapi/payloads"
	"regexp"

	"errors"

	"github.com/go-playground/validator/v10"
)

var (
	errStatusMinTwo       = errors.New("metrik multiple-choices wajib punya minimal 2 status")
	errStatusNotAllowed   = errors.New("metrik number/maps tidak boleh punya status")
	errStatusKeyDuplicate = errors.New("status_key duplikat dalam satu request")
)

func CreateDashboardMetricValidator(sl validator.StructLevel) {
	payload := sl.Current().Interface().(payloads.CreateDashboardMetricRequest)
	validateStatusesAgainstTemplate(sl, payload.ExpectedTemplate, payload.Statuses)
}

func UpdateDashboardMetricStatusesValidator(expectedTemplate string, statuses []payloads.DashboardMetricStatusPayload) error {
	return validateStatusRules(expectedTemplate, statuses)
}

func validateStatusesAgainstTemplate(sl validator.StructLevel, expectedTemplate string, statuses []payloads.DashboardMetricStatusPayload) {
	if err := validateStatusRules(expectedTemplate, statuses); err != nil {
		sl.ReportError(statuses, "Statuses", "statuses", "invalidstatuses", err.Error())
	}
}

func validateStatusRules(expectedTemplate string, statuses []payloads.DashboardMetricStatusPayload) error {
	switch expectedTemplate {
	case "multiple-choices":
		if len(statuses) < 2 {
			return errStatusMinTwo
		}
	case "number", "maps":
		if len(statuses) > 0 {
			return errStatusNotAllowed
		}
	}

	seen := make(map[string]bool, len(statuses))
	for _, s := range statuses {
		if seen[s.StatusKey] {
			return errStatusKeyDuplicate
		}
		seen[s.StatusKey] = true
	}
	return nil
}

var metricKeyFormatRegex = regexp.MustCompile(`^[a-z0-9_]+$`)

func RegisterMetricKeyFormat(v *validator.Validate) error {
	return v.RegisterValidation("metric_key_format", func(fl validator.FieldLevel) bool {
		return metricKeyFormatRegex.MatchString(fl.Field().String())
	})
}
