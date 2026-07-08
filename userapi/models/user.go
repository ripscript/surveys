package models

import (
	"time"

	"gorm.io/gorm"
)

type UserProfile struct {
	ID                 int            `gorm:"column:id;primaryKey;autoIncrement"                json:"id"`
	FirstName          *string        `gorm:"column:first_name;type:varchar(191)"               json:"first_name"`
	LastName           *string        `gorm:"column:last_name;type:varchar(191)"                json:"last_name"`
	Email              *string        `gorm:"column:email;type:varchar(191)"                    json:"email"`
	RespondentID       *int64         `gorm:"column:respondent_id"                              json:"respondent_id"`
	NIK                *string        `gorm:"column:nik;type:varchar(191)"                      json:"nik"`
	CreatedAt          *time.Time     `gorm:"column:created_at;autoCreateTime"                  json:"created_at"`
	UpdatedAt          *time.Time     `gorm:"column:updated_at;autoUpdateTime"                  json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index"                           json:"deleted_at,omitempty"`
	Password           *string        `gorm:"column:password;type:varchar(255)"                 json:"-"`
	LastLogin          *time.Time     `gorm:"column:last_login"`
	MustChangePassword *bool          `gorm:"column:must_change_password"`

	Respondent     *RespondentProfile     `gorm:"foreignKey:RespondentID" json:"respondent"`
	PejabatWilayah *PejabatWilayahProfile `gorm:"foreignKey:IdResponden;references:RespondentID" json:"pejabat_wilayah"`
}

func (UserProfile) TableName() string {
	return "users"
}

type RespondentProfile struct {
	ID            int            `gorm:"column:id;primaryKey;autoIncrement"                json:"id"`
	Name          string         `gorm:"column:name;type:varchar(70);not null"             json:"name"`
	PhoneNumber   *string        `gorm:"column:phone_number;type:varchar(191)"             json:"phone_number"`
	Email         *string        `gorm:"column:email;type:varchar(191)"                    json:"email"`
	BlkID         int            `gorm:"column:blk_id;not null"                            json:"blk_id"`
	NIK           *string        `gorm:"column:nik;type:varchar(191)"                      json:"nik"`
	KecamatanID   *int64         `gorm:"column:kecamatan_id"                               json:"kecamatan_id"`
	KecamatanCode *string        `json:"kecamatan_code"`
	NamaKecamatan *string        `json:"nama_kecamatan"`
	KelurahanID   *int64         `gorm:"column:kelurahan_id"                               json:"kelurahan_id"`
	KelurahanCode *string        `json:"kelurahan_code"`
	NamaKelurahan *string        `json:"nama_kelurahan"`
	RwID          *int64         `gorm:"column:rw_id"                                      json:"rw_id"`
	RwCode        *string        `json:"rw_code"`
	NamaRw        *string        `json:"nama_rw"`
	RtID          *int64         `gorm:"column:rt_id"                                      json:"rt_id"`
	RtCode        *string        `json:"rt_code"`
	NamaRt        *string        `json:"nama_rt"`
	RoleID        int64          `gorm:"column:role_id;not null"                           json:"role_id"`
	TanggalLahir  *time.Time     `gorm:"column:tanggal_lahir;type:date"                    json:"tanggal_lahir"`
	Alamat        *string        `gorm:"column:alamat;type:text"                           json:"alamat"`
	TempatLahir   *string        `gorm:"column:tempat_lahir;type:varchar(100)"             json:"tempat_lahir"`
	IsBlocked     string         `gorm:"column:is_blocked;type:varchar(255);default:false" json:"is_blocked"`
	Username      *string        `gorm:"column:username;type:varchar(255)"                 json:"username"`
	CreatedAt     *time.Time     `gorm:"column:created_at;autoCreateTime"                  json:"created_at"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at;autoUpdateTime"                  json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"                           json:"deleted_at,omitempty"`
}

func (RespondentProfile) TableName() string {
	return "respondents"
}

type PejabatWilayahProfile struct {
	ID            int        `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	TipeWilayah   *int64     `gorm:"column:tipe_wilayah" json:"tipe_wilayah"`
	IdWilayah     *int64     `gorm:"column:id_wilayah" json:"id_wilayah"`
	PeriodeAwal   *time.Time `gorm:"column:periode_awal;type:date" json:"periode_awal"`
	PeriodeAkhir  *time.Time `gorm:"column:periode_akhir;type:date" json:"periode_akhir"`
	StatusJabatan *bool      `gorm:"column:status_jabat" json:"status_jabatan"`
	IdResponden   *int64     `gorm:"column:id_responden" json:"id_responden"`
	NoSk          *string    `gorm:"column:no_sk" json:"no_sk"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"-"`
}

func (PejabatWilayahProfile) TableName() string {
	return "pejabat__wilayahs"
}

type UserModel struct {
	ID                 int            `gorm:"column:id;primaryKey"`
	FirstName          *string        `gorm:"column:first_name"`
	LastName           *string        `gorm:"column:last_name"`
	Email              *string        `gorm:"column:email"`
	Password           *string        `gorm:"column:password"`
	EmailToken         *string        `gorm:"column:email_token"`
	RespondentID       *int64         `gorm:"column:respondent_id"`
	RememberToken      *string        `gorm:"column:remember_token"`
	Nik                *string        `gorm:"column:nik"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index"`
	CreatedAt          time.Time      `gorm:"column:created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at"`
	LastLogin          *time.Time     `gorm:"column:last_login"`
	MustChangePassword bool           `gorm:"column:must_change_password;not null;default:false"`
}

func (UserModel) TableName() string {
	return "users"
}
