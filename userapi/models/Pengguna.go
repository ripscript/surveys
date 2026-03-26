package models

import (
	"time"

	"gorm.io/gorm"
)

type RoleEnum string

type Users struct {
	ID        int       `gorm:"primaryKey"`
	Password  string    `gorm:"column:password"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LogLogin struct {
	ID     int `gorm:"primaryKey"`
	UserID int `json:"user_id"`
}

type Respondent struct {
	ID          int            `gorm:"primaryKey"`
	RoleID      int            `gorm:"column:role_id"`
	Email       string         `gorm:"column:email"`
	Username    string         `gorm:"column:username"`
	NIK         string         `gorm:"column:nik"`
	KecamatanID *int           `gorm:"column:kecamatan_id"`
	KelurahanID *int           `gorm:"column:kelurahan_id"`
	RWID        *int           `gorm:"column:rw_id"`
	RTID        *int           `gorm:"column:rt_id"`
	IsBlocked   bool           `gorm:"column:is_blocked"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
	ID            int            `gorm:"primaryKey"`
	RespondentID  int            `gorm:"column:respondent_id"`
	Name          string         `gorm:"column:name"`
	Email         string         `gorm:"column:email"`
	Password      string         `gorm:"column:password"`
	Role          int            `gorm:"column:role"`
	AccountStatus bool           `gorm:"column:account_status"`
	LastLogin     time.Time      `gorm:"column:last_login"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type LogBlockLogin struct {
	ID             int  `gorm:"primaryKey"`
	UserCredential int  `gorm:"column:user_credential"`
	IsBlocked      bool `gorm:"column:is_blocked"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PejabatWilayah struct {
	ID          int `gorm:"primaryKey"`
	IDResponden int `gorm:"column:id_responden"`
	StatusJabat int `gorm:"column:status_jabat"`
}

func (PejabatWilayah) TableName() string {
	return "pejabat__wilayahs"
}

func (User) TableName() string {
	return "users"
}

func (Respondent) TableName() string {
	return "respondents"
}

func (LogBlockLogin) TableName() string {
	return "log_block_logins"
}

func (LogLogin) TableName() string {
	return "log__logins"
}
