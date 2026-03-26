package models

import "time"

type Kecamatans struct {
	ID              int        `json:"id"`
	SubDistrictName string     `json:"subDisctictName"`
	SubDistrictSlug string     `json:"subDistrictSlug"`
	KodeWilayah     string     `json:"kodeWilayah"`
	Lat             string     `json:"lat"`
	Long            string     `json:"long"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
}

type KecamatanOptions struct {
	ID              int    `json:"id"`
	SubDistrictName string `json:"subDisctictName"`
	SubDistrictSlug string `json:"subDistrictSlug"`
}

type KelurahanOptions struct {
	ID                int    `json:"id"`
	VillageName       string `json:"villageName"`
	VillageNameSlug   string `json:"villageNameSlug"`
	VillagePostalCode string `json:"villagePostalCode"`
}

type RwOptions struct {
	ID     int    `json:"id"`
	NamaRw string `json:"namaRw"`
}

type RtOptions struct {
	ID     int    `json:"id"`
	NamaRt string `json:"namaRt"`
}

func (u *KecamatanOptions) TableName() string {
	return "kecamatans"
}

func (u *KelurahanOptions) TableName() string {
	return "kelurahans"
}

func (u *RwOptions) TableName() string {
	return "data__rws"
}

func (u *RtOptions) TableName() string {
	return "data__rts"
}
