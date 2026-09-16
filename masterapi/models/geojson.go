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

type GeoJSONV2 struct {
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	Features []FeatureV2 `json:"features"`
}

type FeatureV2 struct {
	Type       string          `json:"type"`
	Properties PropertiesV2    `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"` // parse manual sesuai kebutuhan (Polygon/MultiPolygon)
}

type PropertiesV2 struct {
	Nama        string  `json:"NAMA"`
	Kecamatan   string  `json:"KECAMATAN"`
	Alamat      *string `json:"ALAMAT"`
	Email       *string `json:"EMAIL"`
	KepalaKel   *string `json:"KEPALA_KEL"`
	JumlahPen   float64 `json:"JUMLAH_PEN"`
	Kepadatan   float64 `json:"KEPADATAN_"`
	LuasWilayah float64 `json:"LUAS_WILAY"`
	Telepon     float64 `json:"TELEPON"`
	Tahun       int     `json:"TAHUN"`
	SumberData  string  `json:"SUMBER_DAT"`
	// JumlahGabung hanya ada di file kecamatan.geojson (hasil dissolve dari kelurahan).
	// Pakai pointer + omitempty supaya aman dibaca dari file kelurahan.geojson yang tidak punya field ini.
	JumlahGabung *int `json:"JUMLAH_GABUNG,omitempty"`
}
