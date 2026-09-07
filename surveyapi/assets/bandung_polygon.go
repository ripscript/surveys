package assets

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed bandung_polygon_nomatim.json
var bandungPolygonRaw []byte

type geoJSONPolygon struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// BandungPolygon berisi titik-titik boundary Kota Bandung, format [lat, lng]
var BandungPolygon [][2]float64

func init() {
	var geo geoJSONPolygon
	if err := json.Unmarshal(bandungPolygonRaw, &geo); err != nil {
		log.Printf("WARNING: gagal load bandung_polygon_nomatim.json: %v", err)
		return
	}

	if len(geo.Coordinates) == 0 {
		log.Println("WARNING: bandung_polygon_nomatim.json kosong")
		return
	}

	ring := geo.Coordinates[0]
	BandungPolygon = make([][2]float64, 0, len(ring))
	for _, point := range ring {
		lng := point[0]
		lat := point[1]
		BandungPolygon = append(BandungPolygon, [2]float64{lat, lng})
	}

	log.Printf("Bandung polygon loaded: %d titik", len(BandungPolygon))
}
