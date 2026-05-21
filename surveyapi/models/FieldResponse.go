package models

import "time"

type FieldResponse struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id"`
	FormFieldID     int        `gorm:"column:form_field_id"`
	FormResponseID  int64      `gorm:"column:form_response_id"`
	Answer          *string    `gorm:"column:answer"`
	Media           *string    `gorm:"column:media"`
	LocationAddress *string    `gorm:"column:location_address"`
	GroupID         int64      `gorm:"column:group_id"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (FieldResponse) TableName() string {
	return "field_responses"
}
