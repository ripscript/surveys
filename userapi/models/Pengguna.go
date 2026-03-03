package models

import (
	"backend/userapi/utils"
	"time"

	"gorm.io/gorm"
)

type RoleEnum string

type Users struct {
	ID            int64      `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	Name          string     `json:"name" gorm:"type:varchar(100);not null"`
	Email         string     `json:"email" gorm:"type:varchar(255);not null"`
	PhoneNumber   string     `json:"phone_number" gorm:"type:varchar(50);not null"`
	Password      string     `json:"password" gorm:"type:varchar(100);not null"`
	LastLogin     time.Time  `json:"last_login" gorm:"type:timestamp"`
	Role          string     `json:"role" gorm:"type:varchar(50);not null"`
	AccountStatus bool       `json:"account_status" gorm:"not null"`
	IdentityPhoto *string    `json:"identity_photo" gorm:"type:varchar(255);"`
	Avatar        *string    `json:"avatar" gorm:"type:varchar(255);"`
	CreatedBy     int64      `json:"created_by" gorm:"not null"`
	UpdatedBy     int64      `json:"updated_by" gorm:"not null"`
	DeletedBy     *int64     `json:"deleted_by"`
	CreatedAt     time.Time  `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt     *time.Time `json:"deleted_at" gorm:"type:timestamp;"`
}

type UserProfile struct {
	ID            int64           `json:"-" gorm:"primaryKey;autoIncrement;not null"`
	EncID         string          `json:"id" gorm:"-"`
	Name          string          `json:"name" gorm:"type:varchar(100);not null"`
	Email         string          `json:"email" gorm:"type:varchar(255);not null"`
	PhoneNumber   string          `json:"phone_number" gorm:"type:varchar(50);not null"`
	Password      string          `json:"password" gorm:"type:varchar(100);not null"`
	LastLogin     time.Time       `json:"last_login" gorm:"type:timestamp"`
	Role          string          `json:"role" gorm:"type:varchar(50);not null"`
	AccountStatus bool            `json:"account_status" gorm:"not null"`
	IdentityPhoto *string         `json:"identity_photo" gorm:"type:varchar(255);"`
	Avatar        *string         `json:"avatar" gorm:"type:varchar(255);"`
	CreatedBy     int64           `json:"created_by" gorm:"not null"`
	UpdatedBy     int64           `json:"updated_by" gorm:"not null"`
	DeletedBy     *int64          `json:"deleted_by"`
	CreatedAt     time.Time       `json:"created_at" gorm:"type:timestamp;default:now()"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"type:timestamp;default:now()"`
	DeletedAt     *time.Time      `json:"deleted_at" gorm:"type:timestamp;"`
	UserBank      UserBankProfile `json:"userBank" gorm:"foreignKey:UserID;references:ID;"`
}

type UserBankProfile struct {
	ID            int64  `json:"-" gorm:"primaryKey;autoIncrement;not null"`
	UserID        int64  `json:"-" gorm:"not null;index"`
	EncID         string `json:"id"`
	AccountNumber string `json:"accountNumber" gorm:"type:varchar(55)"`
	BankID        int64  `json:"-" gorm:"not null;index"`
	EncBankId     string `json:"bank_id"`
	BankName      string `json:"bank_name" gorm:"-"`
	BankImage     string `json:"bank_image" gorm:"-"`
}

type DeleteUsers struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	DeletedBy *int64    `json:"deleted_by"`
	DeletedAt time.Time `json:"deleted_at" gorm:"type:timestamp;"`
}

type UpdateUsers struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	UpdatedBy   int64     `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserBank struct {
	ID            int64  `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	UserID        int64  `json:"user_id" gorm:"not null;index"`
	User          Users  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	AccountNumber string `json:"accountNumber" gorm:"type:varchar(55)"`
	BankID        int64  `json:"bank_id" gorm:"not null;index"`
}
type ListAdmin struct {
	No            int       `json:"no"`
	ID            int64     `json:"-"`
	EncryptId     string    `json:"id" gorm:"-"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	PhoneNumber   string    `json:"phone_number"`
	CreatedAt     time.Time `json:"created_at"`
	AccountStatus bool      `json:"account_status"`
}

type CompleteProfile struct {
	ID             int64      `json:"id"`
	Role           string     `json:"role"`
	JastiperStatus *string    `json:"jastiper_status"`
	Name           string     `json:"name"`
	PhoneNumber    string     `json:"phoneNumber"`
	IdentityPhoto  string     `json:"identityPhoto"`
	Avatar         string     `json:"avatar"`
	DeletedAt      *time.Time `json:"deleted_at"`
	Address        string     `json:"address"`
	PostalCode     string     `json:"postalCode"`
	ProvinceID     int64      `json:"province_id"`
	CityID         int64      `json:"city_id"`
	Slug           string     `json:"slug"`
}

type UpdatePasswordPenggune struct {
	ID        int64     `json:"id"`
	Password  string    `json:"password"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy int64     `json:"updated_by"`
}

type UpdateActiveUser struct {
	ID            int64     `json:"id"`
	AccountStatus bool      `json:"account_status"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedBy     int64     `json:"updated_by"`
}

type UserDetail struct {
	ID          int64  `json:"-"`
	EncID       string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type DetailAddress struct {
	CustomerName string  `json:"customerName"`
	NoTelp       string  `json:"noTelp"`
	Address      Address `json:"address"`
	PostalCode   string  `json:"postalCode"`
}

type Address struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Detail   string `json:"detail"`
}

type AddressUser struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
	Address     string `json:"address"`
	PostalCode  string `json:"postalCode"`
	ProvinceID  int64  `json:"provinceId"`
	CityID      int64  `json:"city"`
}

type UpdateProfile struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone_number"`
	Avatar        string `json:"avatar"`
	IdentityPhoto string `json:"identity_photo"`
}

func (UserBankProfile *UserBankProfile) AfterFind(tx *gorm.DB) (err error) {
	encID, err := utils.EncryptInt(int(UserBankProfile.ID))
	if err != nil {
		return err
	}
	UserBankProfile.EncID = encID

	encBankID, err := utils.EncryptInt(int(UserBankProfile.BankID))
	if err != nil {
		return err
	}
	UserBankProfile.EncBankId = encBankID

	return nil
}

func (UserDetail *UserDetail) AfterFind(tx *gorm.DB) (err error) {
	encID, err := utils.EncryptInt(int(UserDetail.ID))
	if err != nil {
		return err
	}
	UserDetail.EncID = encID
	return nil
}

func (UserProfile *UserProfile) AfterFind(tx *gorm.DB) (err error) {
	encID, err := utils.EncryptInt(int(UserProfile.ID))
	if err != nil {
		return err
	}
	UserProfile.EncID = encID
	return nil
}

func (ListAdmin *ListAdmin) AfterFind(tx *gorm.DB) (err error) {
	encID, err := utils.EncryptInt(int(ListAdmin.ID))
	if err != nil {
		return err
	}
	ListAdmin.EncryptId = encID
	return nil
}

func (ListPelanggan *ListPelanggan) AfterFind(tx *gorm.DB) (err error) {
	encID, err := utils.EncryptInt(int(ListPelanggan.ID))
	if err != nil {
		return err
	}
	ListPelanggan.EncryptId = encID
	return nil
}

func (u *UpdateProfile) TableName() string {
	return "users"
}

func (u *UpdateActiveUser) TableName() string {
	return "users"
}

func (u *UserDetail) TableName() string {
	return "users"
}

func (u *UpdatePasswordPenggune) TableName() string {
	return "users"
}

func (u *Users) TableName() string {
	return "users"
}

func (u *ListAdmin) TableName() string {
	return "users"
}

func (u *DeleteUsers) TableName() string {
	return "users"
}

func (u *UpdateUsers) TableName() string {
	return "users"
}

func (u *CompleteProfile) TableName() string {
	return "users"
}

func (u *AddressUser) TableName() string {
	return "users"
}

func (u *UserProfile) TableName() string {
	return "users"
}

func (u *UserBankProfile) TableName() string {
	return "user_bank"
}

func (u *UserBank) TableName() string {
	return "user_bank"
}
