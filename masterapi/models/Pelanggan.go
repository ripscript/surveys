package models

import (
	"backend/masterapi/utils"
	"time"

	"gorm.io/gorm"
)

type UpdatePelanggan struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	UpdatedBy   int64     `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DeletePelanggan struct {
	ID        int64     `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
	DeletedBy int64     `json:"deleted_by"`
}

type ListPelanggan struct {
	ID            int64     `json:"-"`
	EncryptId     string    `json:"id" gorm:"-"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	PhoneNumber   string    `json:"phone_number"`
	CreatedAt     time.Time `json:"created_at"`
	Role          string    `json:"role"`
	AccountStatus bool      `json:"account_status"`
	No            int       `json:"no"`
}

type DetailPelanggan struct {
	ID               int64     `json:"-"`
	EncryptId        string    `json:"id" gorm:"-"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	PhoneNumber      string    `json:"phone_number"`
	CreatedAt        time.Time `json:"created_at"`
	Role             string    `json:"role"`
	AccountStatus    bool      `json:"account_status"`
	No               int       `json:"no"`
	Avatar           string    `json:"avatar"`
	CountEvent       int       `json:"count_event" gorm:"-"`
	CountTransaction int       `json:"count_transaction" gorm:"-"`
}

type Event struct {
	ID int64 `json:"id"`
}

type ListUnverifJastipers struct {
	ID             int64  `json:"-"`
	EncryptId      string `json:"id" gorm:"-"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	PhoneNumber    string `json:"phone_number"`
	Role           string `json:"role"`
	AccountStatus  bool   `json:"account_status"`
	JastiperStatus string `json:"jastiper_status"`
	No             int    `json:"no" gorm:"-"`
}

type Transaction struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement;not null"`
}

func (ListUnverifJastipers *ListUnverifJastipers) AfterFind(tx *gorm.DB) (err error) {
	encID, err := utils.EncryptInt(int(ListUnverifJastipers.ID))
	if err != nil {
		return err
	}
	ListUnverifJastipers.EncryptId = encID
	return nil
}

func (DetailPelanggan *DetailPelanggan) AfterFind(tx *gorm.DB) (err error) {
	encID, err := utils.EncryptInt(int(DetailPelanggan.ID))
	if err != nil {
		return err
	}
	DetailPelanggan.EncryptId = encID
	return nil
}

func (u *UpdatePelanggan) TableName() string {
	return "users"
}

func (u *DeletePelanggan) TableName() string {
	return "users"
}

func (u *ListPelanggan) TableName() string {
	return "users"
}

func (u *ListUnverifJastipers) TableName() string {
	return "users"
}

func (u *DetailPelanggan) TableName() string {
	return "users"
}
