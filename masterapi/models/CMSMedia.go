package models

import "time"

type CMSMedia struct {
	ID        int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	SectionID int        `json:"section_id" gorm:"column:section_id;not null;index"`
	ImageURL  string     `json:"image_url" gorm:"column:image_url;type:text;not null"`
	Caption   *string    `json:"caption" gorm:"column:caption;type:varchar(255)"`
	ItemOrder int        `json:"item_order" gorm:"column:item_order;not null;default:0"`
	Status    bool       `json:"status" gorm:"column:status;not null;default:true"`
	CreatedAt *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	Section *CMSSection `json:"section,omitempty" gorm:"foreignKey:SectionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (CMSMedia) TableName() string {
	return "cms_media"
}
