package payloads

type DashboardMetricCategoryOptionsPayload struct {
	Q          string   `form:"q" query:"q"`
	Page       int      `form:"page" query:"page"`
	Limit      int      `form:"limit" query:"limit"`
	Categories []string `form:"values[]" query:"values[]"`
}
