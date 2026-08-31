package response

import "time"

type RatingOverviewResponse struct {
	Overview    float64 `json:"overview" gorm:"column:overview"`
	MaxRating   int     `json:"max_rating" gorm:"-"`
	TotalReview int64   `json:"total_review" gorm:"column:total_review"`
	Rate1       int64   `json:"rate_1" gorm:"column:rate_1"`
	Rate2       int64   `json:"rate_2" gorm:"column:rate_2"`
	Rate3       int64   `json:"rate_3" gorm:"column:rate_3"`
	Rate4       int64   `json:"rate_4" gorm:"column:rate_4"`
	Rate5       int64   `json:"rate_5" gorm:"column:rate_5"`
}

type OldRatingResponse struct {
	Rating    int        `json:"rating"`
	Ulasan    string     `json:"ulasan"`
	CreatedAt *time.Time `json:"created_at"`
}
