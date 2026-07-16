package repository

import (
	"backend/userapi/models"
	"backend/userapi/payloads"
	"backend/userapi/utils"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UsersRepo interface {
	GetUsers(offset int, limit int, param url.Values) ([]models.Respondents, int64, error)
	GetDetailUsers(id int) (models.Respondents, error)
	UpdateUsers(id int, updateData models.UpdateRespondent) error
	DeleteUsers(deletedUsers models.DeleteRespondent) error
	ResetPasswordUsers(id int) error
	UserExport() ([]models.ExportUsers, error)
	StoreUsers(data models.CreateRespondent) error
	CheckEmail(email string) (int64, int64, error)
	CheckPhoneNumber(phoneNumber string) (int64, error)
	CheckNik(nik string) (int64, int64, error)
	GetUserRawById(id int) (*models.UserProfile, error)

	UpdateProfileBundleTx(userID int, respondentID int, isPejabat bool, payload payloads.UpdateProfileBundlePayload, hashedPassword string) error
	CheckDuplicateProfileData(email string, nik *string, phone *string, currentUserID int, currentRespondentID int) error

	UpdateUser(user *models.UserProfile) (*models.UserProfile, error)
	WithTx(tx *gorm.DB) *usersRepo
	GetUserByRespondentId(respondentId int) (*models.UserProfile, error)
	BeginTx() *gorm.DB

	FindByRespondentID(respondentID int) (*models.UserProfile, error)
	UpdateLastLogin(userID int, t time.Time) error
}

type usersRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewUsersRepo(dbSlave, dbMaster *gorm.DB) *usersRepo {
	defer utils.GeneralRecover()
	return &usersRepo{
		dbSlave,
		dbMaster,
	}
}

func (r *usersRepo) GetUsers(offset int, limit int, param url.Values) ([]models.Respondents, int64, error) {
	var usersList []models.Respondents
	var total int64

	search := param.Get("search")
	name := param.Get("name")
	email := param.Get("email")
	nik := param.Get("nik")
	phoneNumber := param.Get("phoneNumber")
	role := param.Get("role")
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	query := r.dbSlave.Preload("KecamatanJoin").Preload("KelurahanJoin").Preload("RwJoin").Preload("RtJoin").Where("deleted_at IS NULL").Where("role_id IN (?)", []int{7, 8})
	if search != "" {
		query = query.Where("LOWER(email) LIKE ? OR LOWER(name) LIKE ? OR LOWER(phone_number) LIKE ?", "%"+strings.ToLower(search)+"%", "%"+strings.ToLower(search)+"%", "%"+strings.ToLower(search)+"%")
	}

	if name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
	}
	if email != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(email)+"%")
	}
	if nik != "" {
		query = query.Where("LOWER(nik) LIKE ?", "%"+strings.ToLower(nik)+"%")
	}
	if phoneNumber != "" {
		query = query.Where("LOWER(phone_number) LIKE ?", "%"+strings.ToLower(phoneNumber)+"%")
	}
	if role != "" {
		var roleId int

		switch role {
		case "Admin":
			roleId = 7
		case "Surveyor":
			roleId = 8
		}

		query = query.Where("role_id = ?", roleId)
	}

	if orderBy != "" {
		orderBy = utils.ToSnakeCase(orderBy)

		finalOrderBy := "respondents.created_at"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":           "respondents.id",
			"name":         "respondents.name",
			"phone_number": "respondents.phone_number",
			"email":        "respondents.email",
		}

		if mappedCol, isAllowed := allowedOrderCols[orderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(orderDir) == "asc" {
			finalOrderDir = "asc"
		}

		query = query.Order(finalOrderBy + " " + finalOrderDir + " NULLS LAST")
	} else {
		query = query.Order("respondents.created_at desc NULLS LAST")
	}

	if err := query.Model(&models.Respondents{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit != 1 {
		if err := query.Offset(offset).Limit(limit).Find(&usersList).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := query.Offset(offset).Find(&usersList).Error; err != nil {
			return nil, 0, err
		}
	}

	for i := range usersList {
		var role string

		if usersList[i].RoleId == 7 {
			role = "Admin"
		} else {
			role = "Surveyor"
		}
		usersList[i].No = offset + i + 1
		usersList[i].Kecamatan = usersList[i].KecamatanJoin.SubDistrictName
		usersList[i].Kelurahan = usersList[i].KelurahanJoin.VillageName
		usersList[i].Rw = usersList[i].RwJoin.NamaRw
		usersList[i].Rt = usersList[i].RtJoin.NamaRt
		usersList[i].Role = role
	}

	return usersList, total, nil
}

func (r *usersRepo) GetDetailUsers(id int) (models.Respondents, error) {
	defer utils.GeneralRecover()
	var usersList models.Respondents

	query := r.dbSlave.Preload("KecamatanJoin").Preload("KelurahanJoin").Preload("RwJoin").Preload("RtJoin").Where("deleted_at IS NULL").Where("id = ?", id)
	if err := query.Find(&usersList).Error; err != nil {
		return usersList, err
	}

	var role string

	if usersList.RoleId == 7 {
		role = "Admin"
	} else {
		role = "Surveyor"
	}

	usersList.Kecamatan = usersList.KecamatanJoin.SubDistrictName
	usersList.Kelurahan = usersList.KelurahanJoin.VillageName
	usersList.Rw = usersList.RwJoin.NamaRw
	usersList.Rt = usersList.RtJoin.NamaRt
	usersList.Role = role

	return usersList, nil
}

func (r *usersRepo) UpdateUsers(id int, updateData models.UpdateRespondent) error {
	defer utils.GeneralRecover()
	var modelsUpdate models.UpdateRespondent
	db := r.dbMaster
	err := db.Model(modelsUpdate).Where("id = ?", id).Updates(updateData).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *usersRepo) DeleteUsers(deletedUsers models.DeleteRespondent) error {
	defer utils.GeneralRecover()
	var modelsDelete models.DeleteRespondent
	db := r.dbMaster
	err := db.Model(modelsDelete).Where("id = ?", deletedUsers.Id).Updates(deletedUsers).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *usersRepo) ResetPasswordUsers(id int) error {
	defer utils.GeneralRecover()

	var users models.Users

	hashedPassword, err := utils.HashPassword("lacirw123")
	if err != nil {
		return err
	}

	err = r.dbMaster.Where("respondent_id = ?", id).First(&users).Error
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": time.Now(),
	}

	err = r.dbMaster.Model(&models.Users{}).
		Where("respondent_id = ?", id).
		Updates(updates).Error
	if err != nil {
		return err
	}

	err = r.dbMaster.Where("user_id = ?", users.ID).
		Delete(&models.LogLogin{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *usersRepo) UserExport() ([]models.ExportUsers, error) {
	defer utils.GeneralRecover()
	var users []models.ExportUsers
	db := r.dbSlave

	err := db.Where("role_id = ? AND deleted_at IS NUlL", 8).Find(&users).Error
	if err != nil {
		return users, err
	}

	return users, nil
}

func (r *usersRepo) CheckEmail(email string) (int64, int64, error) {
	defer utils.GeneralRecover()
	var countRespondent int64
	var countUsers int64
	db := r.dbSlave

	err := db.Model(models.Respondent{}).Where("email = ?", email).Count(&countRespondent).Error
	if err != nil {
		return 0, 0, err
	}
	err = db.Model(models.Users{}).Where("email = ?", email).Count(&countUsers).Error
	if err != nil {
		return 0, 0, err
	}

	return countRespondent, countUsers, nil
}

func (r *usersRepo) CheckPhoneNumber(phoneNumber string) (int64, error) {
	defer utils.GeneralRecover()
	var countRespondent int64
	db := r.dbSlave

	err := db.Model(models.Respondent{}).Where("phone_number = ?", phoneNumber).Count(&countRespondent).Error
	if err != nil {
		return 0, err
	}

	return countRespondent, nil
}

func (r *usersRepo) CheckNik(nik string) (int64, int64, error) {
	defer utils.GeneralRecover()
	var countRespondent int64
	var countUsers int64
	db := r.dbSlave

	err := db.Model(models.Respondent{}).Where("nik = ?", nik).Count(&countRespondent).Error
	if err != nil {
		return 0, 0, err
	}
	err = db.Model(models.Users{}).Where("nik = ?", nik).Count(&countUsers).Error
	if err != nil {
		return 0, 0, err
	}

	return countRespondent, countUsers, nil
}

func (r *usersRepo) StoreUsers(data models.CreateRespondent) error {
	defer utils.GeneralRecover()
	db := r.dbMaster
	tx := db.Begin()

	err := tx.Create(&data).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	hashedPassword, err := utils.HashPassword("lacirw123")
	if err != nil {
		tx.Rollback()
		return err
	}

	randomBytes := make([]byte, 48)
	if _, err := rand.Read(randomBytes); err != nil {
		tx.Rollback()
		return err
	}

	var dataUsers models.StoreUsers
	dataUsers.FirstName = data.Name
	dataUsers.LastName = strings.ReplaceAll(strings.ToLower(data.Name), " ", "-")
	dataUsers.Email = data.Email
	dataUsers.Password = hashedPassword
	dataUsers.EmailToken = base64.URLEncoding.EncodeToString(randomBytes)
	dataUsers.RespondentId = data.Id
	dataUsers.CreatedAt = utils.TimeNow()
	dataUsers.UpdatedAt = utils.TimeNow()
	dataUsers.MustChangePassword = true

	err = tx.Create(&dataUsers).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *usersRepo) GetUserRawById(id int) (*models.UserProfile, error) {
	defer utils.GeneralRecover()

	var user models.UserProfile
	err := r.dbSlave.
		Preload("Respondent").
		Preload("PejabatWilayah").
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	if user.Respondent != nil {
		type wilayahNames struct {
			NamaKecamatan *string
			NamaKelurahan *string
			NamaRw        *string
			NamaRt        *string
		}

		var names wilayahNames

		err = r.dbSlave.
			Table("respondents").
			Select(
				"kecamatans.sub_district_name AS nama_kecamatan",
				"kelurahans.village_name AS nama_kelurahan",
				"data__rws.nama_rw AS nama_rw",
				"data__rts.nama_rt AS nama_rt",
			).
			Joins("LEFT JOIN kecamatans ON kecamatans.id = respondents.kecamatan_id").
			Joins("LEFT JOIN kelurahans ON kelurahans.id = respondents.kelurahan_id").
			Joins("LEFT JOIN data__rws ON data__rws.id = respondents.rw_id").
			Joins("LEFT JOIN data__rts ON data__rts.id = respondents.rt_id").
			Where("respondents.id = ?", user.Respondent.ID).
			Scan(&names).Error

		if err != nil {
			return nil, err
		}

		user.Respondent.NamaKecamatan = names.NamaKecamatan
		user.Respondent.NamaKelurahan = names.NamaKelurahan
		user.Respondent.NamaRw = names.NamaRw
		user.Respondent.NamaRt = names.NamaRt
	}

	return &user, nil
}

func (r *usersRepo) UpdateProfileBundleTx(userID int, respondentID int, isPejabat bool, payload payloads.UpdateProfileBundlePayload, hashedPassword string) error {
	defer utils.GeneralRecover()

	tx := r.dbSlave.Begin()
	if tx.Error != nil {
		return errors.New("gagal memulai transaksi database")
	}

	if hashedPassword != "" {
		if err := tx.Table("users").Where("id = ?", userID).Updates(map[string]interface{}{
			"password":             hashedPassword,
			"must_change_password": false,
		}).Error; err != nil {
			tx.Rollback()
			return errors.New("gagal mengupdate password")
		}
	}

	updateRespondentData := map[string]interface{}{
		"name":          payload.Name,
		"email":         payload.Email,
		"nik":           payload.NIK,
		"alamat":        payload.Alamat,
		"tempat_lahir":  payload.TempatLahir,
		"tanggal_lahir": payload.TanggalLahir,
		"phone_number":  payload.PhoneNumber,
	}
	if err := tx.Table("respondents").Where("id = ?", respondentID).Updates(updateRespondentData).Error; err != nil {
		tx.Rollback()
		return errors.New("gagal mengupdate data responden")
	}

	updateUserData := map[string]interface{}{
		"email": payload.Email,
		"nik":   payload.NIK,
	}
	if err := tx.Table("users").Where("id = ?", userID).Updates(updateUserData).Error; err != nil {
		tx.Rollback()
		return errors.New("gagal mengupdate data user")
	}

	if isPejabat {
		updatePejabatData := map[string]interface{}{
			"no_sk": payload.NoSK,
		}
		if err := tx.Table("pejabat__wilayahs").Where("id_responden = ?", respondentID).Updates(updatePejabatData).Error; err != nil {
			tx.Rollback()
			return errors.New("gagal mengupdate data administratif")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("gagal menyimpan perubahan data")
	}

	return nil
}

func (r *usersRepo) CheckDuplicateProfileData(email string, nik *string, phone *string, currentUserID int, currentRespondentID int) error {
	var count int64

	if email != "" {
		r.dbSlave.Table("users").Where("email = ? AND id != ?", email, currentUserID).Where("deleted_at IS NULL").Count(&count)
		if count > 0 {
			return errors.New("Email sudah digunakan oleh pengguna lain")
		}
	}

	if nik != nil && *nik != "" {
		count = 0
		r.dbSlave.Table("respondents").Where("nik = ? AND id != ?", *nik, currentRespondentID).Where("deleted_at IS NULL").Count(&count)
		if count > 0 {
			return errors.New("NIK sudah terdaftar di sistem")
		}
	}

	// if phone != nil && *phone != "" {
	// 	count = 0
	// 	r.dbSlave.Table("respondents").Where("phone_number = ? AND id != ?", *phone, currentRespondentID).Where("deleted_at IS NULL").Count(&count)
	// 	if count > 0 {
	// 		return errors.New("Nomor telepon sudah terdaftar di sistem")
	// 	}
	// }

	return nil
}

func (repository *usersRepo) UpdateUser(user *models.UserProfile) (*models.UserProfile, error) {
	defer utils.GeneralRecover()
	err := repository.dbMaster.Save(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repository *usersRepo) WithTx(tx *gorm.DB) *usersRepo {
	return &usersRepo{dbMaster: tx}
}

func (repository *usersRepo) GetUserByRespondentId(respondentId int) (*models.UserProfile, error) {
	defer utils.GeneralRecover()
	var user models.UserProfile
	err := repository.dbSlave.Where("respondent_id = ?", respondentId).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *usersRepo) BeginTx() *gorm.DB {
	defer utils.GeneralRecover()
	return r.dbMaster.Begin()
}

func (r *usersRepo) FindByRespondentID(respondentID int) (*models.UserProfile, error) {
	var user models.UserProfile
	err := r.dbSlave.Where("respondent_id = ?", respondentID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *usersRepo) UpdateLastLogin(userID int, t time.Time) error {
	return r.dbMaster.Model(&models.UserProfile{}).
		Where("id = ?", userID).
		Update("last_login", t).Error
}
