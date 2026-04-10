package models

import "time"

type Menu struct {
	ID        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	MenuName  string `json:"menu_name" gorm:"type:varchar(255);not null"`
	Key       string `json:"key"`
	Endpoint  string `json:"Endpoint"`
	Icon      string `json:"icon"`
	Path      string `json:"path" gorm:"type:varchar(255)"`
	Method    string `json:"method" gorm:"type:varchar(10)"`
	ParentID  *int   `json:"parent_id" gorm:"default:null"`
	Parent    *Menu  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;"`
	Children  []Menu `gorm:"foreignKey:ParentID"`
	CreatedBy uint
	UpdatedBy uint
	DeletedBy *uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type MenuPermission struct {
	ID        int        `json:"id" gorm:"primaryKey;autoIncrement"`
	MenuID    int        `gorm:"uniqueIndex:idx_role_menu"`
	Menu      Menu       `gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;"`
	RoleID    int        `gorm:"uniqueIndex:idx_role_menu"`
	CreatedBy uint       `json:"created_by"`
	UpdatedBy uint       `json:"updated_by"`
	DeletedBy *uint      `json:"deleted_by"`
	CreatedAt time.Time  `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"type:timestamp"`
}
