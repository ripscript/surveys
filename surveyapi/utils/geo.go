package utils

func PointInPolygon(lat, lng float64, polygon [][2]float64) bool {
	inside := false
	n := len(polygon)
	j := n - 1

	for i := 0; i < n; i++ {
		latI, lngI := polygon[i][0], polygon[i][1]
		latJ, lngJ := polygon[j][0], polygon[j][1]

		if ((lngI > lng) != (lngJ > lng)) &&
			(lat < (latJ-latI)*(lng-lngI)/(lngJ-lngI)+latI) {
			inside = !inside
		}
		j = i
	}

	return inside
}
