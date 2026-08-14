package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                 int32          `gorm:"column:id;primaryKey" json:"id"`
	FirstName          *string        `gorm:"column:first_name" json:"first_name"`
	LastName           *string        `gorm:"column:last_name" json:"last_name"`
	Email              *string        `gorm:"column:email" json:"email"`
	Password           *string        `gorm:"column:password" json:"-"`
	EmailToken         *string        `gorm:"column:email_token" json:"-"`
	RespondentID       *int64         `gorm:"column:respondent_id" json:"respondent_id"`
	RememberToken      *string        `gorm:"column:remember_token" json:"-"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	CreatedAt          *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          *time.Time     `gorm:"column:updated_at" json:"updated_at"`
	NIK                *string        `gorm:"column:nik" json:"nik"`
	LastLogin          *time.Time     `gorm:"column:last_login" json:"last_login"`
	MustChangePassword bool           `gorm:"column:must_change_password;not null;default:false" json:"must_change_password"`
}

func (User) TableName() string {
	return "users"
}
