package models

type CountKecamatan struct {
	ID int `json:"id"`
}

type CountKelurahan struct {
	ID int `json:"id"`
}

type CountRw struct {
	ID int `json:"id"`
}

type CountRt struct {
	ID int `json:"id"`
}

type Surveys struct {
	Total int `json:"total"`
}

type ResponseCounting struct {
	Total int `json:"total"`
}

func (CountKecamatan) TableName() string {
	return "kecamatans"
}

func (CountKelurahan) TableName() string {
	return "kelurahans"
}

func (CountRw) TableName() string {
	return "data__rws"
}

func (CountRt) TableName() string {
	return "data__rts"
}
