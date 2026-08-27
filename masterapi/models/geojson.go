package models

import "encoding/json"

type GeoJSON struct {
	Features []Feature `json:"features"`
}

type Feature struct {
	Properties Properties      `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"`
}

type Properties struct {
	NamaKecamatan string `json:"nama_kecamatan"`
	NamaKelurahan string `json:"nama_kelurahan"`
}
