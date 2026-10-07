package repository

import (
	"backend/userapi/models"
	"backend/userapi/utils"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// RespondentImportRepo adalah repository khusus untuk fitur import responden.
// Dibuat terpisah agar tidak mengubah RespondentRepo yang sudah ada.
type RespondentImportRepo interface {
	BeginTx() *gorm.DB
	StoreRespondentWithUser(tx *gorm.DB, data models.CreateRespondents) error
	GetKecamatanIDByName(name string) (int, error)
	GetKelurahanIDByName(name string, kecamatanID int) (int, error)
	GetRwIDByNameAndKelurahan(namaRw string, kelurahanID int) (int, error)
	GetRwIDByIDAndKelurahan(id uint64, kelurahanID int) (int, error)
	GetRtIDByNameAndRw(namaRt string, rwID int) (int, error)
	GetRtIDByIDAndRw(id uint64, rwID int) (int, error)
}

type respondentImportRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewRespondentImportRepo(dbSlave, dbMaster *gorm.DB) RespondentImportRepo {
	return &respondentImportRepo{dbSlave: dbSlave, dbMaster: dbMaster}
}

func (r *respondentImportRepo) BeginTx() *gorm.DB {
	return r.dbMaster.Begin()
}

// StoreRespondentWithUser menyimpan respondent + membuat record users
// dengan password default "lacirw123", meniru logika respondentRepo.StoreUsers lama.
func (r *respondentImportRepo) StoreRespondentWithUser(tx *gorm.DB, data models.CreateRespondents) error {
	defer utils.GeneralRecover()

	if err := tx.Create(&data).Error; err != nil {
		return err
	}

	hashedPassword, err := utils.HashPassword("lacirw123")
	if err != nil {
		return err
	}

	var dataUsers models.StoreUsers
	dataUsers.FirstName = data.Name
	dataUsers.LastName = strings.ReplaceAll(strings.ToLower(data.Name), " ", "-")
	dataUsers.Email = data.Email
	dataUsers.Password = hashedPassword
	dataUsers.RespondentId = data.Id
	dataUsers.Nik = data.NIK
	dataUsers.CreatedAt = utils.TimeNow()
	dataUsers.UpdatedAt = utils.TimeNow()
	dataUsers.MustChangePassword = true

	if err := tx.Create(&dataUsers).Error; err != nil {
		return err
	}
	return nil
}

func (r *respondentImportRepo) GetKecamatanIDByName(name string) (int, error) {
	defer utils.GeneralRecover()
	var id int
	err := r.dbSlave.
		Table("kecamatans").
		Select("id").
		Where("LOWER(TRIM(sub_district_name)) = ?", strings.ToLower(strings.TrimSpace(name))).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(&id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("Data Kecamatan tidak ditemukan")
	}
	return id, nil
}

func (r *respondentImportRepo) GetKelurahanIDByName(name string, kecamatanID int) (int, error) {
	defer utils.GeneralRecover()
	var id int
	q := r.dbSlave.
		Table("kelurahans").
		Select("id").
		Where("LOWER(TRIM(village_name)) = ?", strings.ToLower(strings.TrimSpace(name))).
		Where("deleted_at IS NULL")
	if kecamatanID > 0 {
		q = q.Where("sub_district_id = ?", kecamatanID)
	}
	if err := q.Limit(1).Scan(&id).Error; err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("Data Kelurahan tidak ditemukan")
	}
	return id, nil
}

func (r *respondentImportRepo) GetRwIDByNameAndKelurahan(namaRw string, kelurahanID int) (int, error) {
	defer utils.GeneralRecover()
	var id int
	err := r.dbSlave.
		Table("data__rws").
		Select("id").
		Where("TRIM(nama_rw) = ?", strings.TrimSpace(namaRw)).
		Where("kelurahan_id = ?", kelurahanID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(&id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("Data RW tidak ditemukan")
	}
	return id, nil
}

func (r *respondentImportRepo) GetRwIDByIDAndKelurahan(id uint64, kelurahanID int) (int, error) {
	defer utils.GeneralRecover()
	var rwID int
	err := r.dbSlave.
		Table("data__rws").
		Select("id").
		Where("id = ?", id).
		Where("kelurahan_id = ?", kelurahanID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(&rwID).Error
	if err != nil {
		return 0, err
	}
	if rwID == 0 {
		return 0, fmt.Errorf("Data RW tidak ditemukan")
	}
	return rwID, nil
}

func (r *respondentImportRepo) GetRtIDByNameAndRw(namaRt string, rwID int) (int, error) {
	defer utils.GeneralRecover()
	var id int
	err := r.dbSlave.
		Table("data__rts").
		Select("id").
		Where("TRIM(nama_rt) = ?", strings.TrimSpace(namaRt)).
		Where("rw_id = ?", rwID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(&id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("Data RT tidak ditemukan")
	}
	return id, nil
}

func (r *respondentImportRepo) GetRtIDByIDAndRw(id uint64, rwID int) (int, error) {
	defer utils.GeneralRecover()
	var rtID int
	err := r.dbSlave.
		Table("data__rts").
		Select("id").
		Where("id = ?", id).
		Where("rw_id = ?", rwID).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(&rtID).Error
	if err != nil {
		return 0, err
	}
	if rtID == 0 {
		return 0, fmt.Errorf("Data RT tidak ditemukan")
	}
	return rtID, nil
}
