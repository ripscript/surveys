package models

import (
	"time"
)

type DataGeneralTemplate struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	TableName_   string    `gorm:"column:table_name;type:varchar(191)" json:"table_name"`
	ColumnName   string    `gorm:"column:column_name;type:varchar(191)" json:"column_name"`
	VariableName string    `gorm:"column:variable_name;type:varchar(191)" json:"variable_name"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u *DataGeneralTemplate) TableName() string {
	return "data_general_templates"
}
