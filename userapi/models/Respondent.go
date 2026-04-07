package models

import (
	"time"

	"gorm.io/gorm"
)

type UpdateRespondent struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	RoleID      int       `json:"role"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateRespondents struct {
	Id           int       `json:"id"`
	NIK          string    `json:"nik"`
	Name         string    `json:"name"`
	TempatLahir  string    `json:"place_of_birth"`
	TanggalLahir string    `json:"date_of_birth"`
	Alamat       string    `json:"address"`
	PhoneNumber  string    `json:"phone_number"`
	Email        string    `json:"email"`
	RoleID       int       `json:"role"`
	Kecamatan    string    `json:"kecamatan"`
	Kelurahan    string    `json:"kelurahan"`
	RW           string    `json:"rw"`
	RT           string    `json:"rt"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type DeleteRespondent struct {
	Id        int       `json:"id"`
	DeletedAt time.Time `json:"updated_at"`
}

type Respondents struct {
	No          int    `gorm:"-" json:"no"`
	ID          int    `gorm:"primaryKey"`
	Email       string `gorm:"column:email"`
	Name        string `gorm:"column:name"`
	PhoneNumber string `gorm:"column:phone_number"`
	Username    string `gorm:"column:username"`
	NIK         string `gorm:"column:nik"`
	RoleId      int    `json:"-"`
	Role        string `json:"role"`

	KecamatanID *int `gorm:"column:kecamatan_id" json:"-"`
	KelurahanID *int `gorm:"column:kelurahan_id" json:"-"`
	RWID        *int `gorm:"column:rw_id" json:"-"`
	RTID        *int `gorm:"column:rt_id" json:"-"`

	KecamatanJoin Kecamatan `gorm:"foreignKey:KecamatanID" json:"-"`
	KelurahanJoin Kelurahan `gorm:"foreignKey:KelurahanID" json:"-"`
	RwJoin        Rw        `gorm:"foreignKey:RWID" json:"-"`
	RtJoin        Rt        `gorm:"foreignKey:RTID" json:"-"`

	Kecamatan string `gorm:"-"`
	Kelurahan string `gorm:"-"`
	Rw        string `gorm:"-"`
	Rt        string `gorm:"-"`
}

type RawRespondents struct {
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

type Kecamatan struct {
	ID              int    `json:"-"`
	SubDistrictName string `json:"subDisctictName"`
	SubDistrictSlug string `json:"subDistrictSlug"`
}

type Kelurahan struct {
	ID                int    `json:"-"`
	VillageName       string `json:"villageName"`
	VillageNameSlug   string `json:"villageNameSlug"`
	VillagePostalCode string `json:"villagePostalCode"`
}

type Rw struct {
	ID     int    `json:"-"`
	NamaRw string `json:"namaRw"`
}

type Rt struct {
	ID     int    `json:"-"`
	NamaRt string `json:"namaRt"`
}

func (u *RawRespondents) TableName() string {
	return "respondents"
}

func (u *DeleteRespondent) TableName() string {
	return "respondents"
}

func (u *UpdateRespondent) TableName() string {
	return "respondents"
}

func (u *UpdateRespondents) TableName() string {
	return "respondents"
}

func (u *Kecamatan) TableName() string {
	return "kecamatans"
}

func (u *Kelurahan) TableName() string {
	return "kelurahans"
}

func (u *Rw) TableName() string {
	return "data__rws"
}

func (u *Rt) TableName() string {
	return "data__rts"
}
