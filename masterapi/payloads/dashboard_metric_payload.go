package payloads

type DashboardMetricStatusPayload struct {
	ID        int64  `json:"id"`
	StatusKey string `json:"status_key" validate:"required,max=100,metric_key_format"`
	Label     string `json:"label" validate:"required,max=191"`
	Sequence  int    `json:"sequence"`
}

type CreateDashboardMetricRequest struct {
	MetricKey        string                         `json:"metric_key" validate:"required,max=100,metric_key_format"`
	Label            string                         `json:"label" validate:"required,max=191"`
	ExpectedTemplate string                         `json:"expected_template" validate:"required,oneof=number multiple-choices"`
	Category         string                         `json:"category" validate:"required,max=50"`
	Statuses         []DashboardMetricStatusPayload `json:"statuses" validate:"dive"`
}

type UpdateDashboardMetricRequest struct {
	Label    string `json:"label" validate:"required,max=191"`
	Category string `json:"category" validate:"required,max=50"`
	// MetricKey & ExpectedTemplate SENGAJA tidak ada di sini —
	// keduanya immutable setelah create (kontrak dashboard/eksternal).
	Statuses []DashboardMetricStatusPayload `json:"statuses" validate:"dive"`
}

type DashboardMetricCategoryOptionsPayload struct {
	Q          string   `form:"q" query:"q"`
	Page       int      `form:"page" query:"page"`
	Limit      int      `form:"limit" query:"limit"`
	Categories []string `form:"values[]" query:"values[]"`
}

type DashboardMetricOptionsPayload struct {
	Q                string  `form:"q" query:"q"`
	Page             int     `form:"page" query:"page"`
	Limit            int     `form:"limit" query:"limit"`
	IDs              []int64 `form:"id[]" query:"id[]"`
	ExpectedTemplate string  `form:"expected_template" query:"expected_template"`
}

type PemetaanMetrikSurveyDatatablePayload struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
	Status   string `json:"status" validate:"omitempty,oneof=upcoming ongoing finished"`
}
