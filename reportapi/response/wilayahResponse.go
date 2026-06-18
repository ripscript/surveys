package response

import (
	"time"

	"gorm.io/gorm"
)

type Meta struct {
	Limit      int `json:"limit"`
	Page       int `json:"page"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type RTDatatableResponse struct {
	Data []RtDatatableResponse `json:"data"`
	Meta Meta                  `json:"meta"`
}

type DetailSurveyKewilayahanDatatableResponse struct {
	Data []DetailSurveyKewilayahanResponse `json:"data"`
	Meta Meta                              `json:"meta"`
}

type DetailSurveyKewilayahanResponse struct {
	No                     int64  `json:"no"`
	ID                     int64  `json:"-"`
	Code                   string `json:"code"`
	NamaWilayah            string `json:"nama_wilayah"`
	Status                 string `json:"status"`
	IsPosibleDetail        bool   `json:"posible_detail"`
	IsPosiblePreviewSurvey bool   `json:"posible_preview_survey"`
}

type RWDatatableResponse struct {
	Data []RwDatatableResponse `json:"data"`
	Meta Meta                  `json:"meta"`
}

type KelurahanDatatableResponse struct {
	Data []KelurahanDatatableResponses `json:"data"`
	Meta Meta                          `json:"meta"`
}

type RtDatatableResponse struct {
	No              int64      `json:"no"`
	ID              int64      `json:"id"`
	KodeWilayah     *string    `json:"kode_wilayah"`
	SubDistrictName string     `json:"sub_district_name"`
	SubDistrictId   int64      `json:"sub_district_id"`
	VillageName     string     `json:"village_name"`
	VillageId       int64      `json:"village_id"`
	NamaRw          string     `json:"nama_rw"`
	RwId            int64      `json:"rw_id"`
	NamaRt          string     `json:"nama_rt"`
	NamaPejabat     *string    `json:"nama_pejabat"`
	PeriodeAwal     *time.Time `json:"periode_awal"`
	PeriodeAkhir    *time.Time `json:"periode_akhir"`
	Lat             *string    `json:"lat"`
	Long            *string    `json:"long"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type RwDatatableResponse struct {
	No              int64          `json:"no"`
	ID              int64          `json:"id"`
	KodeWilayah     *string        `json:"kode_wilayah"`
	SubDistrictName string         `json:"sub_district_name"`
	SubDistrictId   int64          `json:"sub_district_id"`
	VillageName     string         `json:"village_name"`
	VillageId       int64          `json:"village_id"`
	NamaRw          string         `json:"nama_rw"`
	NamaPejabat     *string        `json:"nama_pejabat"`
	PeriodeAwal     *time.Time     `json:"periode_awal"`
	PeriodeAkhir    *time.Time     `json:"periode_akhir"`
	Lat             *string        `json:"lat"`
	Long            *string        `json:"long"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type KelurahanDatatableResponses struct {
	No              int64          `json:"no"`
	ID              int64          `json:"id"`
	SubDistrictName string         `json:"sub_district_name"`
	SubDistrictId   int64          `json:"sub_district_id"`
	KodeWilayah     *string        `json:"kode_wilayah"`
	VillageName     string         `json:"village_name"`
	NamaPejabat     *string        `json:"nama_pejabat"`
	PeriodeAwal     *time.Time     `json:"periode_awal"`
	PeriodeAkhir    *time.Time     `json:"periode_akhir"`
	Lat             *string        `json:"lat"`
	Long            *string        `json:"long"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
