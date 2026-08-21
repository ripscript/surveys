package customValidator

import (
	"backend/reportapi/payloads"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var alphanumUnderscoreRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func CreateDashboardMetricValidator(sl validator.StructLevel) {
	req := sl.Current().Interface().(payloads.CreateDashboardMetricRequest)

	if req.MetricKey != "" {
		if !alphanumUnderscoreRegex.MatchString(req.MetricKey) {
			sl.ReportError(req.MetricKey, "MetricKey", "metric_key", "alphanum_underscore", "")
		}
	}

	if strings.ToLower(req.Category) == "geografis" {
		if req.ExpectedTemplate != "maps" {
			sl.ReportError(req.ExpectedTemplate, "ExpectedTemplate", "expected_template", "must_be_maps_if_geografis", "")
		}
	}

	if strings.HasPrefix(req.MetricKey, "img_") {
		if req.ExpectedTemplate != "image-template" {
			sl.ReportError(req.ExpectedTemplate, "ExpectedTemplate", "expected_template", "must_be_image_if_img_prefix", "")
		}
	}
}
