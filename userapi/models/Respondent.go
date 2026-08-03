package models

import (
	"time"

	"gorm.io/gorm"
)

type ExportUsers struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateRespondent struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	RoleID      int       `json:"role"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRespondent struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	RoleID      int       `json:"role"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
	BlkId       int       `json:"blk_id"`
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
	Kecamatan    string    `json:"kecamatan" gorm:"column:kecamatan_id"`
	Kelurahan    string    `json:"kelurahan" gorm:"column:kelurahan_id"`
	RW           string    `json:"rw" gorm:"column:rw_id"`
	RT           string    `json:"rt" gorm:"column:rt_id"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateRespondents struct {
	Id           int       `json:"id"`
	NIK          string    `json:"nik"`
	Name         string    `json:"name"`
	TempatLahir  string    `json:"place_of_birth"`
	TanggalLahir string    `json:"date_of_birth"`
	Alamat       string    `json:"address"`
	PhoneNumber  string    `json:"phone_number"`
	Email        string    `json:"email"`
	RoleID       int       `json:"role"`
	Kecamatan    *int      `gorm:"column:kecamatan_id" json:"-"`
	Kelurahan    *int      `gorm:"column:kelurahan_id" json:"-"`
	RW           *int      `gorm:"column:rw_id" json:"-"`
	RT           *int      `gorm:"column:rt_id" json:"-"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
	BlkId        int       `json:"blk_id"`
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
	IsBlocked   string `gorm:"is_blocked"`
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

	Kecamatan string    `gorm:"-"`
	Kelurahan string    `gorm:"-"`
	Rw        string    `gorm:"-"`
	Rt        string    `gorm:"-"`
	DeletedAt time.Time `json:"-"`
	Status    string    `json:"status" gorm:"-"`
}

type RespondentOptions struct {
	ID   int    `json:"value"`
	Name string `json:"label"`
}

type RespondentBlock struct {
	No          int    `gorm:"-" json:"no"`
	ID          int    `gorm:"primaryKey"`
	Email       string `gorm:"column:email"`
	Name        string `gorm:"column:name"`
	PhoneNumber string `gorm:"column:phone_number"`
	NIK         string `gorm:"column:nik"`
	IsBlocked   string `json:"-"`
}

type RawRespondents struct {
	ID           int64          `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	Name         string         `json:"name" gorm:"type:varchar(70);not null"`
	PhoneNumber  *string        `json:"phone_number" gorm:"type:varchar(191);"`
	Email        *string        `json:"email" gorm:"type:varchar(191);"`
	BLKID        *int32         `json:"blk_id" gorm:"column:blk_id;type:int4;"`
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

type Surveyor struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SurveyorOptions struct {
	ID   int    `json:"value"`
	Name string `json:"label"`
}

func (u *SurveyorOptions) TableName() string {
	return "respondents"
}

func (u *RespondentOptions) TableName() string {
	return "respondents"
}

func (u *Surveyor) TableName() string {
	return "respondents"
}

func (u *RespondentBlock) TableName() string {
	return "respondents"
}

func (u *ExportUsers) TableName() string {
	return "respondents"
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

func (u *CreateRespondent) TableName() string {
	return "respondents"
}

func (u *UpdateRespondents) TableName() string {
	return "respondents"
}

func (u *CreateRespondents) TableName() string {
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

type RespondentRaw struct {
	ID           int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string     `gorm:"column:name;type:varchar(70);not null" json:"name"`
	PhoneNumber  *string    `gorm:"column:phone_number;type:varchar(191)" json:"phone_number"`
	Email        *string    `gorm:"column:email;type:varchar(191)" json:"email"`
	BlkID        int        `gorm:"column:blk_id;not null" json:"blk_id"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
	NIK          *string    `gorm:"column:nik;type:varchar(191)" json:"nik"`
	KecamatanID  *int64     `gorm:"column:kecamatan_id" json:"kecamatan_id"`
	KelurahanID  *int64     `gorm:"column:kelurahan_id" json:"kelurahan_id"`
	RwID         *int64     `gorm:"column:rw_id" json:"rw_id"`
	RtID         *int64     `gorm:"column:rt_id" json:"rt_id"`
	RoleID       int64      `gorm:"column:role_id;not null" json:"role_id"`
	TanggalLahir *time.Time `gorm:"column:tanggal_lahir;type:date" json:"tanggal_lahir"`
	Alamat       *string    `gorm:"column:alamat;type:text" json:"alamat"`
	TempatLahir  *string    `gorm:"column:tempat_lahir;type:varchar(100)" json:"tempat_lahir"`
	IsBlocked    string     `gorm:"column:is_blocked;type:varchar(255);not null;default:'false'" json:"is_blocked"`
	Username     *string    `gorm:"column:username;type:varchar(255)" json:"username"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

func (RespondentRaw) TableName() string {
	return "respondents"
}

type RespondentModel struct {
	ID           int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string     `gorm:"column:name;type:varchar(70);not null" json:"name"`
	PhoneNumber  *string    `gorm:"column:phone_number;type:varchar(191)" json:"phone_number,omitempty"`
	Email        *string    `gorm:"column:email;type:varchar(191)" json:"email,omitempty"`
	BlkID        int        `gorm:"column:blk_id;not null" json:"blk_id"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`
	NIK          *string    `gorm:"column:nik;type:varchar(191)" json:"nik,omitempty"`
	KecamatanID  *int64     `gorm:"column:kecamatan_id" json:"kecamatan_id,omitempty"`
	KelurahanID  *int64     `gorm:"column:kelurahan_id" json:"kelurahan_id,omitempty"`
	RwID         *int64     `gorm:"column:rw_id" json:"rw_id,omitempty"`
	RtID         *int64     `gorm:"column:rt_id" json:"rt_id,omitempty"`
	RoleID       int64      `gorm:"column:role_id;not null" json:"role_id"`
	TanggalLahir *time.Time `gorm:"column:tanggal_lahir;type:date" json:"tanggal_lahir,omitempty"`
	Alamat       *string    `gorm:"column:alamat;type:text" json:"alamat,omitempty"`
	TempatLahir  *string    `gorm:"column:tempat_lahir;type:varchar(100)" json:"tempat_lahir,omitempty"`
	IsBlocked    string     `gorm:"column:is_blocked;type:varchar(255);not null;default:'false'" json:"is_blocked"`
	Username     *string    `gorm:"column:username;type:varchar(255)" json:"username,omitempty"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	Avatar       *string    `gorm:"column:avatar;type:varchar(255)" json:"avatar,omitempty"`
}

// TableName overrides the default table name used by GORM
func (RespondentModel) TableName() string {
	return "respondents"
}
