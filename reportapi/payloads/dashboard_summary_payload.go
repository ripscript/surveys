package payloads

import "time"

type GetDashboardSummaryRequest struct {
	WilayahLevel int       `query:"wilayah_level" validate:"required,oneof=9 7 5 4 3 2"` // 5=Kecamatan, 4=Kelurahan, 3=RW, 2=RT
	WilayahID    *int64    `query:"wilayah_id"`
	PeriodStart  time.Time `query:"period_start" validate:"required"`
	PeriodEnd    time.Time `query:"period_end" validate:"required,gtefield=PeriodStart"`
}
