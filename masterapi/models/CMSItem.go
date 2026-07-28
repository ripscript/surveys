package models

import (
	"time"
)

type CMSItem struct {
	ID          int        `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	SectionID   int        `json:"section_id" gorm:"column:section_id;not null;index"`
	ItemOrder   int        `json:"item_order" gorm:"column:item_order;not null;default:0"`
	Title       string     `json:"title" gorm:"column:title;type:varchar(191);not null"`
	Description *string    `json:"description" gorm:"column:description;type:text"`
	Icon        *string    `json:"icon" gorm:"column:icon;type:text"`
	Image       *string    `json:"image" gorm:"column:image;type:text"`
	Category    *string    `json:"category" gorm:"column:category;type:varchar(100)"`
	LinkURL     *string    `json:"link_url" gorm:"column:link_url;type:text"`
	Status      bool       `json:"status" gorm:"column:status;not null;default:true"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	Section *CMSSection `json:"section,omitempty" gorm:"foreignKey:SectionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (CMSItem) TableName() string {
	return "cms_items"
}
