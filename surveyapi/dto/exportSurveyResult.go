package dto

import "time"

type ExportWilayahDTO struct {
	NamaRT        string `gorm:"column:nama_rt"`
	NamaRW        string `gorm:"column:nama_rw"`
	NamaKelurahan string `gorm:"column:village_name"`
	NamaKecamatan string `gorm:"column:sub_district_name"`
}

type ExportRawJawabanDTO struct {
	RespondentID   int64      `gorm:"column:respondent_id"`
	NamaResponden  string     `gorm:"column:nama_responden"`
	Status         *int       `gorm:"column:status"`
	StatusApproval *string    `gorm:"column:status_approval"`
	WaktuSelesai   *time.Time `gorm:"column:waktu_selesai"`
	FormFieldID    int        `gorm:"column:form_field_id"`
	Answer         *string    `gorm:"column:answer"`
}

type RekapRespondenExcel struct {
	NamaResponden string
	StatusText    string
	WaktuSelesai  string
	JawabanMap    map[int]string
}

// ============
type ExportRawJawabanWilayah struct {
	RespondentID   int64
	NamaResponden  string
	KecamatanName  string
	KelurahanName  string
	RwName         string
	RtName         string
	RTID           int64
	Status         *int16
	StatusApproval *string
	WaktuSelesai   *time.Time
	FormFieldID    int
	Answer         *string
}

type RekapRespondenWilayahExcel struct {
	NamaResponden string
	KecamatanName string
	KelurahanName string
	RwName        string
	RtName        string
	StatusText    string
	WaktuSelesai  string
	RTID          int64
	JawabanMap    map[int]string
}
