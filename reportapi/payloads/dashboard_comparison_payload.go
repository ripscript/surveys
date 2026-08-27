package payloads

type GetDashboardComparisonRequest struct {
	DashboardFilter
	MetricKeyA string `query:"metric_key_a" validate:"required"`
	MetricKeyB string `query:"metric_key_b" validate:"required"`
}
