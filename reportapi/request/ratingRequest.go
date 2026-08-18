package request

type RatingRequest struct {
	Rating     int    `json:"rating" validate:"required"`
	Ulasan     string `json:"ulasan" validate:"required"`
	DeviceId   string `json:"device_id" validate:"required"`
	DeviceInfo string `json:"device_info" validate:"required"`
}
