package models

import "time"

type SectionType string

const (
	SectionTypeContent SectionType = "content"
	SectionTypeItems   SectionType = "items"
	SectionTypeMap     SectionType = "map"
	SectionTypeTable   SectionType = "table"

	SectionTypeText    SectionType = "text"
	SectionTypePicture SectionType = "picture"
)

type CMSSection struct {
	ID           int         `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Slug         string      `json:"slug" gorm:"column:slug;type:varchar(50);not null;uniqueIndex"`
	Name         string      `json:"name" gorm:"column:name;type:varchar(191);not null"`
	Description  *string     `json:"description" gorm:"column:description;type:text"`
	Type         SectionType `json:"type" gorm:"column:type;type:varchar(20);not null;default:'content'"`
	SectionOrder int         `json:"section_order" gorm:"column:section_order;not null;default:0"`
	IsRepeatable bool        `json:"is_repeatable" gorm:"column:is_repeatable;not null;default:false"`
	IsSystem     bool        `json:"is_system" gorm:"column:is_system;not null;default:false;index"`
	IsEditable   *bool       `json:"is_editable" gorm:"column:is_editable;not null;default:true"`
	Status       bool        `json:"status" gorm:"column:status;not null;default:true"`
	CreatedAt    *time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    *time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	Contents []CMSContent `json:"contents,omitempty" gorm:"foreignKey:SectionID;references:ID"`
	Items    []CMSItem    `json:"items,omitempty" gorm:"foreignKey:SectionID;references:ID"`
	Media    []CMSMedia   `json:"media,omitempty" gorm:"foreignKey:SectionID;references:ID"`
}

func (CMSSection) TableName() string {
	return "cms_sections"
}

type CMSSectionDatatable struct {
	ID           int         `json:"id"`
	Slug         string      `json:"slug"`
	Name         string      `json:"name"`
	Description  *string     `json:"description"`
	Type         SectionType `json:"type"`
	SectionOrder int         `json:"section_order"`
	IsRepeatable bool        `json:"is_repeatable"`
	IsSystem     bool        `json:"is_system"`
	IsEditable   bool        `json:"is_editable"`
	Status       bool        `json:"status"`
	CreatedAt    *time.Time  `json:"created_at"`
	UpdatedAt    *time.Time  `json:"updated_at"`

	PosibleDelete          bool `json:"posible_delete"`
	PosibleChangeName      bool `json:"posible_change_name"`
	PosibleUpdate          bool `json:"posible_update"`
	PosibleChangeOrderUp   bool `json:"posible_change_order_up"`
	PosibleChangeOrderDown bool `json:"posible_change_order_down"`
}
