package models

import "time"

type CMSContent struct {
	ID         int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	SectionID  int        `json:"section_id" gorm:"column:section_id;not null;uniqueIndex:idx_section_key"`
	Key        string     `json:"key" gorm:"column:key;type:varchar(100);not null;uniqueIndex:idx_section_key"`
	ValueText  *string    `json:"value_text" gorm:"column:value_text;type:text"`
	ValueImage *string    `json:"value_image" gorm:"column:value_image;type:text"`
	Status     bool       `json:"status" gorm:"column:status;not null;default:true"`
	CreatedAt  *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	Section *CMSSection `json:"section,omitempty" gorm:"foreignKey:SectionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (CMSContent) TableName() string {
	return "cms_contents"
}
