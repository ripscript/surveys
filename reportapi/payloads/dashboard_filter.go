package payloads

import "time"

type DashboardFilter struct {
	WilayahLevel int       `query:"wilayah_level" validate:"required,oneof=9 7 5 4 3 2"`
	WilayahID    *int64    `query:"wilayah_id"`
	PeriodStart  time.Time `query:"period_start" validate:"required"`
	PeriodEnd    time.Time `query:"period_end" validate:"required,gtefield=PeriodStart"`
	RTIds        []int64   `query:"rt_ids"`
	RWIds        []int64   `query:"rw_ids"`
	KelurahanIds []int64   `query:"kelurahan_ids"`
	KecamatanIds []int64   `query:"kecamatan_ids"`
}
