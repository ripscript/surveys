package repository

import (
	"backend/reportapi/models"
	"backend/reportapi/request"
	"backend/reportapi/response"
	"backend/reportapi/utils"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type RatingRepo interface {
	InsertRatingTransaction(payload request.RatingRequest, ip string) error
	GetLogRatingByDeviceID(deviceID string) (*models.LogBlockRating, error)
	CheckIfIPBlocked(ip string) bool
	CountRatingByIPToday(ip string) (int64, error)
	GetRatingOverview() (*response.RatingOverviewResponse, error)
	GetListRating(req request.RatingDatatablePayload) ([]models.RatingDatatable, int64, error)
	GetOldRatingByDeviceID(deviceID string) (*response.OldRatingResponse, error)
}

type ratingRepo struct {
	dbSlave  *gorm.DB
	dbMaster *gorm.DB
}

func NewRatingRepo(dbSlave, dbMaster *gorm.DB) RatingRepo {
	defer utils.GeneralRecover()
	return &ratingRepo{dbSlave, dbMaster}
}

func (repository *ratingRepo) InsertRatingTransaction(payload request.RatingRequest, ip string) error {
	defer utils.GeneralRecover()

	tx := repository.dbSlave.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	ratingData := models.Rating{
		Rating: payload.Rating,
		Ulasan: payload.Ulasan,
	}

	if err := tx.Create(&ratingData).Error; err != nil {
		tx.Rollback()
		return err
	}

	logData := models.LogBlockRating{
		RatingID:   ratingData.ID,
		DeviceID:   payload.DeviceId,
		DeviceInfo: &payload.DeviceInfo,
		IsBlocked:  "false",
	}

	if ip != "" {
		logData.IP = &ip
	}

	if err := tx.Create(&logData).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (repository *ratingRepo) GetLogRatingByDeviceID(deviceID string) (*models.LogBlockRating, error) {
	defer utils.GeneralRecover()

	var logRating models.LogBlockRating
	err := repository.dbSlave.Where("device_id = ?", deviceID).First(&logRating).Error
	if err != nil {
		return nil, err
	}

	return &logRating, nil
}

func (repository *ratingRepo) CheckIfIPBlocked(ip string) bool {
	defer utils.GeneralRecover()
	var count int64
	repository.dbSlave.Model(&models.LogBlockRating{}).
		Where("ip = ? AND is_blocked = 'true'", ip).
		Count(&count)

	return count > 0
}

func (repository *ratingRepo) CountRatingByIPToday(ip string) (int64, error) {
	defer utils.GeneralRecover()
	var count int64

	err := repository.dbSlave.Model(&models.LogBlockRating{}).
		Where("ip = ? AND created_at >= NOW() - INTERVAL '1 DAY'", ip).
		Count(&count).Error

	return count, err
}

func (repository *ratingRepo) GetRatingOverview() (*response.RatingOverviewResponse, error) {
	defer utils.GeneralRecover()

	var result response.RatingOverviewResponse

	err := repository.dbSlave.Table("ratings").
		Select(`
			COALESCE(ROUND(AVG(rating)::numeric, 1), 0) AS overview,
			COUNT(id) AS total_review,
			COALESCE(SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END), 0) AS rate_1,
			COALESCE(SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END), 0) AS rate_2,
			COALESCE(SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END), 0) AS rate_3,
			COALESCE(SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END), 0) AS rate_4,
			COALESCE(SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END), 0) AS rate_5
		`).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	result.MaxRating = 5

	return &result, nil
}

func (repository *ratingRepo) GetListRating(req request.RatingDatatablePayload) ([]models.RatingDatatable, int64, error) {
	defer utils.GeneralRecover()
	var data []models.RatingDatatable
	var totalData int64

	db := repository.dbSlave.Model(&models.Rating{})

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		searchStr := strings.TrimSpace(req.Search)
		parsedDate, isDate := utils.TryParseIndonesianDate(req.Search)

		if isDate {
			db = db.Where(`
				ratings.ulasan ILIKE ? OR
				DATE(ratings.created_at) = ? OR
				DATE(ratings.updated_at) = ?
			`, searchTerm, parsedDate, parsedDate)
		} else if len(searchStr) == 4 {
			db = db.Where(`
				ratings.ulasan ILIKE ? OR
				EXTRACT(YEAR FROM ratings.created_at)::TEXT = ? OR
				EXTRACT(YEAR FROM ratings.updated_at)::TEXT = ?
			`, searchTerm, searchStr, searchStr)
		} else {
			db = db.Where(`
				ratings.ulasan ILIKE ?
			`, searchTerm)
		}
	}

	err := db.Count(&totalData).Error
	if err != nil {
		return nil, 0, err
	}

	db = db.Select(`
		ratings.id,
		ratings.rating,
		ratings.ulasan,
		ratings.updated_at,
		ratings.created_at
	`)

	if req.OrderBy != "" {
		finalOrderBy := "ratings.id"
		finalOrderDir := "desc"

		allowedOrderCols := map[string]string{
			"id":         "ratings.id",
			"rating":     "ratings.rating",
			"ulasan":     "ratings.ulasan",
			"created_at": "ratings.created_at",
			"updated_at": "ratings.updated_at",
		}

		if mappedCol, isAllowed := allowedOrderCols[req.OrderBy]; isAllowed {
			finalOrderBy = mappedCol
		}

		if strings.ToLower(req.OrderDir) == "asc" {
			finalOrderDir = "asc"
		}

		db = db.Order(fmt.Sprintf("%s %s", finalOrderBy, finalOrderDir))
	} else {
		db = db.Order("ratings.id desc")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err = db.Limit(limit).Offset(offset).Scan(&data).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range data {
		data[i].No = int64(offset + i + 1)
	}

	return data, totalData, nil
}

func (repository *ratingRepo) GetOldRatingByDeviceID(deviceID string) (*response.OldRatingResponse, error) {
	defer utils.GeneralRecover()

	var result response.OldRatingResponse

	err := repository.dbSlave.Table("log_block_ratings").
		Select("ratings.rating, ratings.ulasan, ratings.created_at").
		Joins("JOIN ratings ON ratings.id = log_block_ratings.rating_id").
		Where("log_block_ratings.device_id = ?", deviceID).
		Order("log_block_ratings.created_at DESC").
		Limit(1).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	if result.Rating == 0 && result.Ulasan == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &result, nil
}
