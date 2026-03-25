package models

import (
	"time"

	"gorm.io/gorm"
)

type Respondent struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	Name         string         `json:"name" gorm:"type:varchar(70);not null"`
	PhoneNumber  *string        `json:"phone_number" gorm:"type:varchar(191);"`
	Email        *string        `json:"email" gorm:"type:varchar(191);"`
	BLKDID       *int32         `json:"blk_id" gorm:"type:int4;"`
	NIK          *string        `json:"nik" gorm:"type:varchar(191);"`
	KecamatanId  *int64         `json:"kecamatan_id" gorm:"type:int8;"`
	KelurahanId  *int64         `json:"kelurahan_id" gorm:"type:int8;"`
	RWId         *int64         `json:"rw_id" gorm:"type:int8;"`
	RTId         *int64         `json:"rt_id" gorm:"type:int8;"`
	RoleId       *int64         `json:"role_id" gorm:"type:int8;"`
	TanggalLahir *time.Time     `json:"tanggal_lahir" gorm:"type:date;"`
	Alamat       *string        `json:"alamat" gorm:"type:text;"`
	TempatLahir  *string        `json:"tempat_lahir" gorm:"type:varchar(100);"`
	IsBlocked    bool           `json:"is_blocked" gorm:"type:boolean;default:false;"`
	Username     *string        `json:"username" gorm:"type:varchar(255);"`
	CreatedAt    time.Time      `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *Respondent) TableName() string {
	return "respondents"
}
