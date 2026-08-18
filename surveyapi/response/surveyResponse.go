package response

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type HeatpointResponse struct {
	ID     int64        `json:"id"`
	Answer []Coordinate `json:"answer"`
}
