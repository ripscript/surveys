package service

import (
	"backend/reportapi/models"
	"backend/reportapi/repository"
	"backend/reportapi/request"
	"backend/reportapi/utils"
	"backend/siccore/pb"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type RatingService interface {
	SaveRating(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	RatingOverview(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	RatingList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type ratingService struct {
	ratingRepo repository.RatingRepo
}

func NewRatingService(
	ratingRepo repository.RatingRepo,
) RatingService {
	return &ratingService{ratingRepo}
}

func (service *ratingService) SaveRating(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload request.RatingRequest
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	var ipAddress string
	if len(param["ip_address"]) > 0 {
		ipAddress = param["ip_address"][0]
	}

	if ipAddress != "" {
		isBlocked := service.ratingRepo.CheckIfIPBlocked(ipAddress)
		if isBlocked {
			return utils.SendError(errors.New("aktivitas mencurigakan terdeteksi dari jaringan Anda, permintaan ditolak"), http.StatusForbidden)
		}
	}

	logRating, err := service.ratingRepo.GetLogRatingByDeviceID(payload.DeviceId)
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return utils.SendError(errors.New("terjadi kesalahan saat memeriksa log rating, silakan coba lagi"), http.StatusInternalServerError)
	}

	if logRating != nil {
		return utils.SendError(errors.New("Anda hanya diperbolehkan memberikan ulasan aplikasi satu kali"), http.StatusForbidden)
	}

	if ipAddress != "" {
		countIP, err := service.ratingRepo.CountRatingByIPToday(ipAddress)
		if err == nil && countIP >= 5 {
			return utils.SendError(errors.New("batas maksimal pengiriman ulasan dari jaringan ini telah tercapai, coba lagi besok"), http.StatusTooManyRequests)
		}
	}

	err = service.ratingRepo.InsertRatingTransaction(payload, ipAddress)
	if err != nil {
		return utils.SendError(errors.New("terjadi kesalahan saat menyimpan rating, silakan coba lagi"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Berhasil mengirimkan rating dan ulasan")
}

func (service *ratingService) RatingOverview(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	overviewData, err := service.ratingRepo.GetRatingOverview()
	if err != nil {
		return utils.SendError(errors.New("gagal mengambil data overview rating, silakan coba lagi"), http.StatusInternalServerError)
	}

	if overviewData.TotalReview > 0 {
		total := overviewData.TotalReview

		rawCounts := []int64{
			overviewData.Rate1,
			overviewData.Rate2,
			overviewData.Rate3,
			overviewData.Rate4,
			overviewData.Rate5,
		}

		pcts := make([]int64, 5)
		var sumPct int64 = 0

		var maxCount int64 = -1
		var maxCountIdx int = 0

		for i, count := range rawCounts {
			pcts[i] = (count * 100) / total
			sumPct += pcts[i]

			if count > maxCount {
				maxCount = count
				maxCountIdx = i
			}
		}

		if sumPct < 100 {
			pcts[maxCountIdx] += (100 - sumPct)
		}

		overviewData.Rate1 = pcts[0]
		overviewData.Rate2 = pcts[1]
		overviewData.Rate3 = pcts[2]
		overviewData.Rate4 = pcts[3]
		overviewData.Rate5 = pcts[4]
	}

	return utils.SendData(overviewData, "Berhasil mendapatkan data overview rating")
}

func (service *ratingService) RatingList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	payload := request.RatingDatatablePayload{
		Search:   search,
		Page:     page,
		Limit:    limit,
		OrderBy:  orderBy,
		OrderDir: orderDir,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 5
	}

	data, totalData, err := service.ratingRepo.GetListRating(payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	showingFrom := (payload.Page-1)*payload.Limit + 1
	showingTo := showingFrom + len(data) - 1

	if totalData == 0 {
		showingFrom = 0
		showingTo = 0
	}

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total_entries": totalData,
			"current_page":  payload.Page,
			"per_page":      payload.Limit,
			"showing_from":  showingFrom,
			"showing_to":    showingTo,
		},
	}

	return utils.SendData(result, "Berhasil mendapatkan data rating")
}
