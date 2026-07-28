package models

import "time"

type ManajemenCMS struct {
	ID             int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Section        string     `json:"section" gorm:"column:section;type:varchar(10);not null"`
	TypeSection    string     `json:"type_section" gorm:"column:type_section;type:varchar(255);not null"`
	SectionText    *string    `json:"section_text" gorm:"column:section_text;type:text"`
	SectionPicture *string    `json:"section_picture" gorm:"column:section_picture;type:text"`
	CreatedAt      *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      *time.Time `json:"updated_at" gorm:"column:updated_at"`
	NameSection    *string    `json:"name_section" gorm:"column:name_section;type:varchar(191)"`
	Status         string     `json:"status" gorm:"column:status;type:varchar(255);not null;default:'true'"`
}

func (ManajemenCMS) TableName() string {
	return "manajemen_cms"
}
