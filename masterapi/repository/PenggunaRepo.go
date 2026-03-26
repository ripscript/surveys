package repository

import (
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/utils"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PenggunaRepo interface {
	ValidasiCredential(payloads.LoginPayload) (models.Users, error)
	ValidasiPengguna(pengguna models.Users, status string, jenis string) error
	EditLastLog(data models.Users) (interface{}, error)
	GetProfile(usr models.JwtCustomClaims) (interface{}, error)
	Add(data models.Users) (models.Users, error)
	CompleteProfile(data models.CompleteProfile) error
	UserBank(data models.UserBank) error
	GetDetail(id int) (models.Users, error)
	ResetPassword(data models.UpdatePasswordPenggune) error
	ActiveUser(data models.UpdateActiveUser) error
	Verification(id int) error
	AddressDetail(userid int) (models.DetailAddress, error)
	GetUserByEmail(email string) (models.Users, error)
	ValidasiPenggunaUpdate(pengguna models.UpdateProfile) error
	UpdateProfile(data models.UpdateProfile) error
	UpdateBankProfile(data models.UserBank) error
}

type penggunaRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewPenggunaRepo(dbSlave, dbMaster *gorm.DB) *penggunaRepo {
	defer utils.GeneralRecover()
	return &penggunaRepo{
		dbSlave,
		dbMaster,
	}
}

func (repository *penggunaRepo) GetUserByEmail(email string) (models.Users, error) {
	defer utils.GeneralRecover()
	var data models.Users
	db := repository.dbSlave

	err := db.Where("email = ?", email).First(&data).Error
	if err != nil {
		return data, err
	}

	return data, nil
}

func (repository *penggunaRepo) Verification(id int) error {
	defer utils.GeneralRecover()

	db := repository.dbMaster
	err := db.Model(&models.Users{}).Where("id = ?", id).Update("account_status", true).Error
	if err != nil {
		return err
	}
	return nil
}

func (repository *penggunaRepo) ActiveUser(data models.UpdateActiveUser) error {
	defer utils.GeneralRecover()
	db := repository.dbMaster
	err := db.Save(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (repository *penggunaRepo) ValidasiCredential(payload payloads.LoginPayload) (models.Users, error) {
	defer utils.GeneralRecover()

	var storedUser models.Users

	db := repository.dbSlave

	if err := db.Where("email = ?", payload.Email).Find(&storedUser).Error; err != nil {
		err = errors.New("Email atau password tidak valid")
		return models.Users{}, err
	}

	if storedUser.ID == 0 {
		err := errors.New("Email atau password tidak valid")
		return models.Users{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(payload.Password)); err != nil {
		err = errors.New("Email atau password tidak valid")
		return models.Users{}, err
	}
	return storedUser, nil
}

func (repository *penggunaRepo) EditLastLog(data models.Users) (interface{}, error) {
	defer utils.GeneralRecover()
	dbMaster := repository.dbMaster

	if data.Email != "" {
		err := dbMaster.Save(&data).Error
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (repository *penggunaRepo) ValidasiPengguna(pengguna models.Users, status string, jenis string) error {
	defer utils.GeneralRecover()
	if status == "add" {
		// validate email
		if repository.ValidEmail(pengguna.Email, 0) {
			err := "Email sudah digunakan, silahkan gunakan email lain!"
			return errors.New(err)
		}
		if jenis != "register" {
			if repository.ValidPhone(pengguna.PhoneNumber, 0) {
				err := "Nomor Telepon Sudah Di gunakan!"
				return errors.New(err)
			}
		}
	} else {
		// validate email
		if repository.ValidEmail(pengguna.Email, pengguna.ID) {
			err := "Email sudah digunakan, silahkan gunakan email lain!"
			return errors.New(err)
		}
		if repository.ValidPhone(pengguna.PhoneNumber, pengguna.ID) {
			err := "Nomor Telepon sudah digunakan, silahkan gunakan nomor lain!"
			return errors.New(err)
		}
	}

	return nil
}

func (repository *penggunaRepo) ValidasiPenggunaUpdate(pengguna models.UpdateProfile) error {
	defer utils.GeneralRecover()

	if repository.ValidEmail(pengguna.Email, pengguna.ID) {
		err := "Email sudah digunakan, silahkan gunakan email lain!"
		return errors.New(err)
	}
	if repository.ValidPhone(pengguna.PhoneNumber, pengguna.ID) {
		err := "Nomor Telepon sudah digunakan, silahkan gunakan nomor lain!"
		return errors.New(err)
	}

	return nil
}

func (repository *penggunaRepo) ValidEmail(email string, id int64) bool {
	defer utils.GeneralRecover()

	db := repository.dbSlave

	var total int64 = 0

	// Validasi email di tabel
	if id != 0 {
		if err := db.Model(&models.Users{}).Where("email = ?", email).Where("id != ?", id).Count(&total).Error; err != nil {
			return false
		}
	} else {
		if err := db.Model(&models.Users{}).Where("email = ?", email).Count(&total).Error; err != nil {
			return false
		}
	}

	fmt.Println("total", total)

	if total > 0 {
		return true
	}
	return false

}

func (repository *penggunaRepo) ValidPhone(phone string, id int64) bool {
	defer utils.GeneralRecover()

	db := repository.dbSlave

	var total int64 = 0

	if id != 0 {
		if err := db.Model(&models.Users{}).Where("phone_number = ?", phone).Where("id != ?", id).Count(&total).Error; err != nil {
			return false
		}
	} else {
		if err := db.Model(&models.Users{}).Where("phone_number = ?", phone).Count(&total).Error; err != nil {
			return false
		}
	}

	fmt.Println("total", total)

	if total > 0 {
		return true
	}
	return false

}

func (repository *penggunaRepo) GetProfile(usr models.JwtCustomClaims) (interface{}, error) {
	defer utils.GeneralRecover()

	dbSlave := repository.dbSlave

	var pengguna models.UserProfile
	if err := dbSlave.Preload("UserBank").First(&pengguna, usr.ID).Error; err != nil {
		return models.UserProfile{}, err
	}

	var bank models.UserBankProfile
	if err := dbSlave.Where("id = ?", pengguna.UserBank.BankID).First(&bank).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return models.UserProfile{}, err
		}
	}

	pengguna.Password = ""
	if bank.ID != 0 {
		pengguna.UserBank.BankName = bank.BankName
		// pengguna.UserBank.BankImage = bank.Path
	}

	return pengguna, nil
}

func (repository *penggunaRepo) Add(data models.Users) (models.Users, error) {
	defer utils.GeneralRecover()
	db := repository.dbMaster

	err := db.Create(&data).Error
	if err != nil {
		return models.Users{}, err
	}

	return data, nil
}

func (repository *penggunaRepo) CompleteProfile(data models.CompleteProfile) error {
	defer utils.GeneralRecover()
	dbMaster := repository.dbMaster
	err := dbMaster.Save(&data).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *penggunaRepo) UserBank(data models.UserBank) error {
	defer utils.GeneralRecover()
	dbMaster := repository.dbMaster
	err := dbMaster.Save(&data).Error
	if err != nil {
		return err
	}

	return nil
}

func (repository *penggunaRepo) GetDetail(id int) (models.Users, error) {
	defer utils.GeneralRecover()
	var data models.Users
	db := repository.dbSlave

	err := db.Where("id = ?", id).First(&data).Error
	if err != nil {
		return data, err
	}

	return data, nil
}

func (repository *penggunaRepo) ResetPassword(data models.UpdatePasswordPenggune) error {
	defer utils.GeneralRecover()
	db := repository.dbMaster
	err := db.Save(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (repository *penggunaRepo) AddressDetail(userid int) (models.DetailAddress, error) {
	defer utils.GeneralRecover()
	var data models.DetailAddress
	var User models.AddressUser
	db := repository.dbSlave

	err := db.Where("id = ?", userid).First(&User).Error
	if err != nil {
		return data, nil
	}

	cityString, err := utils.ToString(User.CityID)
	if err != nil {
		return data, err
	}

	provinceString, err := utils.ToString(User.ProvinceID)
	if err != nil {
		return data, err
	}

	data.CustomerName = User.Name
	data.NoTelp = User.PhoneNumber
	data.PostalCode = User.PostalCode
	data.Address.City = cityString
	data.Address.Detail = User.Address
	data.Address.Province = provinceString

	return data, nil
}

func (r *penggunaRepo) UpdateProfile(data models.UpdateProfile) error {
	defer utils.GeneralRecover()
	db := r.dbMaster
	err := db.Where("id", data.ID).Updates(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *penggunaRepo) UpdateBankProfile(data models.UserBank) error {
	defer utils.GeneralRecover()
	db := r.dbMaster
	err := db.Where("user_id", data.UserID).Updates(&data).Error
	if err != nil {
		return err
	}
	return nil
}
