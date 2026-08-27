package payloads

type GetDashboardTrendRequest struct {
	WilayahLevel int    `query:"wilayah_level" validate:"required,oneof=9 7 5 4 3 2"`
	WilayahID    *int64 `query:"wilayah_id"`
	MetricKey    string `query:"metric_key"`
	PeriodEnd    string `query:"period_end" validate:"required"`
}
