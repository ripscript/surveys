package models

import (
	"time"

	"gorm.io/gorm"
)

type RespondentModel struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	Name         string         `json:"name" gorm:"type:varchar(70);not null"`
	PhoneNumber  *string        `json:"phone_number" gorm:"type:varchar(191);"`
	Email        *string        `json:"email" gorm:"type:varchar(191);"`
	BLKDID       int64          `json:"blk_id" gorm:"type:int8;not null"`
	NIK          *string        `json:"nik" gorm:"type:varchar(191);"`
	KecamatanId  *int64         `json:"kecamatan_id" gorm:"type:int8;"`
	KelurahanId  *int64         `json:"kelurahan_id" gorm:"type:int8;"`
	RWId         *int64         `json:"rw_id" gorm:"type:int8;"`
	RTId         *int64         `json:"rt_id" gorm:"type:int8;"`
	RoleId       int64          `json:"role_id" gorm:"type:int8;not null"`
	TanggalLahir *time.Time     `json:"tanggal_lahir" gorm:"type:date;"`
	Alamat       *string        `json:"alamat" gorm:"type:text;"`
	TempatLahir  *string        `json:"tempat_lahir" gorm:"type:varchar(100);"`
	IsBlocked    string         `json:"is_blocked" gorm:"type:varchar(255);not null;default:'false'"`
	Username     *string        `json:"username" gorm:"type:varchar(255);"`
	CreatedAt    *time.Time     `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt    *time.Time     `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *RespondentModel) TableName() string {
	return "respondents"
}

type RespondentModel_1 struct {
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
	CreatedAt    *time.Time     `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt    *time.Time     `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (u *RespondentModel_1) TableName() string {
	return "respondents"
}

type DetailRespondent struct {
	Email       string `json:"email"`
	ID          int64  `json:"id"`
	Kecamatan   string `json:"kecamatan"`
	Kelurahan   string `json:"kelurahan"`
	NIK         string `json:"nik"`
	Name        string `json:"name"`
	PhoneNumber string `json:"PhoneNumber"`
	RT          string `json:"Rt"`
	RW          string `json:"Rw"`
	Username    string `json:"username"`
	No          int    `json:"no"`
	Role        string `json:"role"`
}
