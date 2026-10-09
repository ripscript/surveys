package service

import (
	"backend/reportapi/enums"
	"backend/reportapi/models"
	"backend/reportapi/payloads"
	"backend/reportapi/repository"
	"backend/reportapi/response"
	"backend/reportapi/utils"
	"backend/siccore/pb"
	"context"
	"errors"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

type DashboardService interface {
	GetSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetSampah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetInfrastruktur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetStuntingVsRTLH(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetTrend(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetComparison(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetHeatmap(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetPeneranganJalan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetTrendCompare(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetSummaryCompare(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetTopWilayahSurvei(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardService struct {
	summaryRepo repository.DashboardSummaryRepository
	wilayahRepo repository.WilayahRepository
	metricRepo  repository.DashboardMetricRepository
	userRepo    repository.UserRepo
}

func NewDashboardService(
	summaryRepo repository.DashboardSummaryRepository,
	wilayahRepo repository.WilayahRepository,
	metricRepo repository.DashboardMetricRepository,
	userRepo repository.UserRepo,
) DashboardService {
	return &dashboardService{
		summaryRepo: summaryRepo,
		wilayahRepo: wilayahRepo,
		metricRepo:  metricRepo,
		userRepo:    userRepo,
	}
}

func intersectAndValidate(requested []int64, allowed []int64, errMsg string) ([]int64, error) {
	allowedSet := make(map[int64]bool, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = true
	}
	result := make([]int64, 0, len(requested))
	for _, id := range requested {
		if !allowedSet[id] {
			return nil, errors.New(errMsg)
		}
		result = append(result, id)
	}
	return result, nil
}

func (s *dashboardService) filterSurveyorRTIDsByWilayah(ctx context.Context, surveyRTIDs []int64, filter payloads.DashboardFilter) ([]int64, error) {
	// Tentukan himpunan RT yang diizinkan dari filter wilayah paling spesifik yang dikirim.
	var allowedRT []int64

	switch {
	case len(filter.RTIds) > 0:
		allowedRT = filter.RTIds

	case len(filter.RWIds) > 0:
		rts, err := s.wilayahRepo.GetRTIDsByRWIDs(ctx, filter.RWIds)
		if err != nil {
			return nil, errors.New("Terjadi kesalahan pada server")
		}
		allowedRT = rts

	case len(filter.KelurahanIds) > 0:
		rwIDs, err := s.wilayahRepo.GetRWIDsByKelurahanIDs(ctx, filter.KelurahanIds)
		if err != nil {
			return nil, errors.New("Terjadi kesalahan pada server")
		}
		rts, err := s.wilayahRepo.GetRTIDsByRWIDs(ctx, rwIDs)
		if err != nil {
			return nil, errors.New("Terjadi kesalahan pada server")
		}
		allowedRT = rts

	case len(filter.KecamatanIds) > 0:
		kelIDs, err := s.wilayahRepo.GetKelurahanIDsByKecamatanIDs(ctx, filter.KecamatanIds)
		if err != nil {
			return nil, errors.New("Terjadi kesalahan pada server")
		}
		rwIDs, err := s.wilayahRepo.GetRWIDsByKelurahanIDs(ctx, kelIDs)
		if err != nil {
			return nil, errors.New("Terjadi kesalahan pada server")
		}
		rts, err := s.wilayahRepo.GetRTIDsByRWIDs(ctx, rwIDs)
		if err != nil {
			return nil, errors.New("Terjadi kesalahan pada server")
		}
		allowedRT = rts

	default:
		// Tidak ada filter wilayah -> pertahankan seluruh RT milik surveyor.
		return surveyRTIDs, nil
	}

	// Interseksi: hanya RT surveyor yang juga ada di himpunan allowedRT.
	allowedSet := make(map[int64]bool, len(allowedRT))
	for _, id := range allowedRT {
		allowedSet[id] = true
	}
	result := make([]int64, 0, len(surveyRTIDs))
	for _, id := range surveyRTIDs {
		if allowedSet[id] {
			result = append(result, id)
		}
	}
	return result, nil
}

func (s *dashboardService) resolveSurveyorSurveyIDs(ctx context.Context, respondentID int64, param url.Values) ([]int64, error) {
	assigned, err := s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, respondentID)
	if err != nil {
		return nil, errors.New("Terjadi kesalahan pada server")
	}
	if len(assigned) == 0 {
		return nil, errors.New("Anda belum ditugaskan pada survey apa pun")
	}

	requested := utils.ParseInt64SliceQueryParam(param.Get("survey_ids"))
	if len(requested) == 0 {
		return assigned, nil
	}

	selected, err := intersectAndValidate(requested, assigned, "survey_ids yang diminta tidak termasuk dalam survey yang ditugaskan kepada Anda")
	if err != nil {
		return nil, err
	}
	return selected, nil
}

func (s *dashboardService) resolveFilteredWilayah(ctx context.Context, filter payloads.DashboardFilter, respondentLogin *models.RespondentModel_1) ([]int64, *response.WilayahCount, response.WilayahInfo, error) {

	var wilayahInfo response.WilayahInfo

	switch filter.WilayahLevel {

	case int(enums.ROLE_RT):
		if filter.WilayahID == nil || respondentLogin.RTId == nil || *filter.WilayahID != *respondentLogin.RTId {
			return nil, nil, wilayahInfo, errors.New("Anda tidak memiliki akses ke wilayah ini")
		}
		info, err := s.wilayahRepo.GetRTInfo(ctx, *filter.WilayahID)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("RT tidak ditemukan")
		}
		wilayahInfo = response.WilayahInfo{Nama: "RT " + info.NamaRt}
		return []int64{*filter.WilayahID}, nil, wilayahInfo, nil

	case int(enums.ROLE_RW):
		if filter.WilayahID == nil || respondentLogin.RWId == nil || *filter.WilayahID != *respondentLogin.RWId {
			return nil, nil, wilayahInfo, errors.New("Anda tidak memiliki akses ke wilayah ini")
		}
		allRT, err := s.wilayahRepo.GetRTIDsUnderRW(ctx, *filter.WilayahID)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}

		finalRT := allRT
		if len(filter.RTIds) > 0 {
			finalRT, err = intersectAndValidate(filter.RTIds, allRT, "rt_ids yang diminta tidak termasuk dalam RW ini")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		totalRT := len(finalRT)
		return finalRT, &response.WilayahCount{TotalRT: &totalRT}, wilayahInfo, nil

	case int(enums.ROLE_KELURAHAN):
		if filter.WilayahID == nil || respondentLogin.KelurahanId == nil || *filter.WilayahID != *respondentLogin.KelurahanId {
			return nil, nil, wilayahInfo, errors.New("Anda tidak memiliki akses ke wilayah ini")
		}
		allRW, err := s.wilayahRepo.GetRWIDsUnderKelurahan(ctx, *filter.WilayahID)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}

		selectedRW := allRW
		if len(filter.RWIds) > 0 {
			selectedRW, err = intersectAndValidate(filter.RWIds, allRW, "rw_ids yang diminta tidak termasuk dalam kelurahan ini")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		allRT, err := s.wilayahRepo.GetRTIDsByRWIDs(ctx, selectedRW)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}

		finalRT := allRT
		if len(filter.RTIds) > 0 {
			finalRT, err = intersectAndValidate(filter.RTIds, allRT, "rt_ids yang diminta tidak termasuk dalam rw yang dipilih")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		totalRW := len(selectedRW)
		totalRT := len(finalRT)
		return finalRT, &response.WilayahCount{TotalRT: &totalRT, TotalRW: &totalRW}, wilayahInfo, nil

	case int(enums.ROLE_KECAMATAN):
		if filter.WilayahID == nil || respondentLogin.KecamatanId == nil || *filter.WilayahID != *respondentLogin.KecamatanId {
			return nil, nil, wilayahInfo, errors.New("Anda tidak memiliki akses ke wilayah ini")
		}
		allKel, err := s.wilayahRepo.GetKelurahanIDsUnderKecamatan(ctx, *filter.WilayahID)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}

		selectedKel := allKel
		if len(filter.KelurahanIds) > 0 {
			selectedKel, err = intersectAndValidate(filter.KelurahanIds, allKel, "kelurahan_ids yang diminta tidak termasuk dalam kecamatan ini")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		allRW, err := s.wilayahRepo.GetRWIDsByKelurahanIDs(ctx, selectedKel)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}

		selectedRW := allRW
		if len(filter.RWIds) > 0 {
			selectedRW, err = intersectAndValidate(filter.RWIds, allRW, "rw_ids yang diminta tidak termasuk dalam kelurahan yang dipilih")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		allRT, err := s.wilayahRepo.GetRTIDsByRWIDs(ctx, selectedRW)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}

		finalRT := allRT
		if len(filter.RTIds) > 0 {
			finalRT, err = intersectAndValidate(filter.RTIds, allRT, "rt_ids yang diminta tidak termasuk dalam rw yang dipilih")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		totalKel := len(selectedKel)
		totalRW := len(selectedRW)
		totalRT := len(finalRT)
		return finalRT, &response.WilayahCount{TotalRT: &totalRT, TotalRW: &totalRW, TotalKelurahan: &totalKel}, wilayahInfo, nil

	case int(enums.ROLE_ADMIN), int(enums.ROLE_WALIKOTA):
		// admin/walikota tidak butuh wilayah_id -- cakupan defaultnya SELURUH
		// kota, lalu dipersempit berjenjang oleh filter yang dikirim
		allKec, err := s.wilayahRepo.GetAllKecamatanIDs(ctx)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}
		selectedKec := allKec
		if len(filter.KecamatanIds) > 0 {
			selectedKec, err = intersectAndValidate(filter.KecamatanIds, allKec, "kecamatan_ids tidak valid")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		allKel, err := s.wilayahRepo.GetKelurahanIDsByKecamatanIDs(ctx, selectedKec)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}
		selectedKel := allKel
		if len(filter.KelurahanIds) > 0 {
			selectedKel, err = intersectAndValidate(filter.KelurahanIds, allKel, "kelurahan_ids yang diminta tidak termasuk dalam kecamatan yang dipilih")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		allRW, err := s.wilayahRepo.GetRWIDsByKelurahanIDs(ctx, selectedKel)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}
		selectedRW := allRW
		if len(filter.RWIds) > 0 {
			selectedRW, err = intersectAndValidate(filter.RWIds, allRW, "rw_ids yang diminta tidak termasuk dalam kelurahan yang dipilih")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		allRT, err := s.wilayahRepo.GetRTIDsByRWIDs(ctx, selectedRW)
		if err != nil {
			return nil, nil, wilayahInfo, errors.New("Terjadi kesalahan pada server")
		}
		finalRT := allRT
		if len(filter.RTIds) > 0 {
			finalRT, err = intersectAndValidate(filter.RTIds, allRT, "rt_ids yang diminta tidak termasuk dalam rw yang dipilih")
			if err != nil {
				return nil, nil, wilayahInfo, err
			}
		}

		totalKec := len(selectedKec)
		totalKel := len(selectedKel)
		totalRW := len(selectedRW)
		totalRT := len(finalRT)
		return finalRT, &response.WilayahCount{
			TotalRT: &totalRT, TotalRW: &totalRW, TotalKelurahan: &totalKel, TotalKecamatan: &totalKec,
		}, wilayahInfo, nil

	default:
		return nil, nil, wilayahInfo, errors.New("Wilayah level tidak valid")
	}
}

func (s *dashboardService) GetSummaryOld(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	period_start := param.Get("period_start")
	period_end := param.Get("period_end")

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if period_end != "" {
		v, err := time.Parse("2006-01-02", period_end)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}

	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}
	if filter.WilayahLevel != int(*respondentLogin.RoleId) {
		return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
	}

	rtIDs, wilayahCount, wilayahInfo, err := s.resolveFilteredWilayah(ctx, filter, respondentLogin)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if len(rtIDs) == 0 {
		return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
	}

	metricCatalog, err := s.metricRepo.List(ctx, "summary")
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	currentByRT, err := s.summaryRepo.GetLatestMetricsByCategory(ctx, rtIDs, "summary", filter.PeriodStart, filter.PeriodEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	prevStart := filter.PeriodStart.AddDate(-1, 0, 0)
	prevEnd := filter.PeriodEnd.AddDate(-1, 0, 0)
	prevByRT, err := s.summaryRepo.GetLatestMetricsByCategory(ctx, rtIDs, "summary", prevStart, prevEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	metrics := make(map[string]response.DashboardMetricValue, len(metricCatalog))
	for _, mc := range metricCatalog {
		var sum, prevSum int64
		for _, rtID := range rtIDs {
			sum += currentByRT[rtID][mc.MetricKey]
			prevSum += prevByRT[rtID][mc.MetricKey]
		}
		delta := sum - prevSum
		metrics[mc.MetricKey] = response.DashboardMetricValue{
			Label: mc.Label, CurrentValue: sum, PreviousValue: prevSum, Delta: delta,
		}
	}

	resp := response.DashboardSummaryResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		WilayahCount: wilayahCount,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		Metrics:      metrics,
	}

	return utils.SendData(resp, "")
}

func (s *dashboardService) GetSummary(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayahLevel, _ := strconv.Atoi(param.Get("wilayah_level"))
	periodStartRaw := param.Get("period_start")
	periodEndRaw := param.Get("period_end")

	now := time.Now()
	periodeStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodeEnd := now

	if (periodStartRaw == "") != (periodEndRaw == "") {
		return utils.SendError(errors.New("period_start dan period_end harus diisi bersamaan"), http.StatusBadRequest)
	}
	if periodStartRaw != "" {
		v, err := time.Parse("2006-01-02", periodStartRaw)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if periodEndRaw != "" {
		v, err := time.Parse("2006-01-02", periodEndRaw)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}
	if periodeEnd.Before(periodeStart) {
		return utils.SendError(errors.New("period_end tidak boleh sebelum period_start"), http.StatusBadRequest)
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayahLevel,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}
	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)
	// SURVEYOR tidak terikat hierarki wilayah; cakupan datanya ditentukan oleh
	// survey yang di-assign ke dirinya (lihat blok khusus surveyor di bawah).
	if role != int(enums.ROLE_ADMIN) && role != int(enums.ROLE_WALIKOTA) && role != int(enums.ROLE_SURVEYOR) && wilayahLevel != role {
		return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
	}

	var (
		rtIDs        []int64
		wilayahCount *response.WilayahCount
		wilayahInfo  response.WilayahInfo
		surveyIDs    []int64 // hanya diisi untuk SURVEYOR
	)

	if role == int(enums.ROLE_SURVEYOR) {
		// SURVEYOR: cakupan ditentukan oleh survey yang di-assign ke dirinya,
		// bukan hierarki wilayah. Ambil survey miliknya, lalu RT yang relevan.
		surveyIDs, err = s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		if len(surveyIDs) == 0 {
			return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
		}

		rtIDs, err = s.wilayahRepo.GetRTIDsBySurveyIDs(ctx, surveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		// Persempit RT surveyor berdasarkan filter wilayah yang dikirim.
		rtIDs, err = s.filterSurveyorRTIDsByWilayah(ctx, rtIDs, filter)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		totalRT := len(rtIDs)
		wilayahCount = &response.WilayahCount{TotalRT: &totalRT}
		wilayahInfo = response.WilayahInfo{Nama: "Survey yang ditugaskan"}
	} else {
		rtIDs, wilayahCount, wilayahInfo, err = s.resolveFilteredWilayah(ctx, filter, respondentLogin)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	if len(rtIDs) == 0 && role != int(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
	}

	metricCatalog, err := s.metricRepo.List(ctx, "summary")
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	prevStart := filter.PeriodStart.AddDate(-1, 0, 0)
	prevEnd := filter.PeriodEnd.AddDate(-1, 0, 0)

	var currentByRT, prevByRT map[int64]map[string]int64
	if role == int(enums.ROLE_SURVEYOR) {
		// Metrik dibatasi HANYA pada survey yang di-assign ke surveyor.
		currentByRT, err = s.summaryRepo.GetLatestMetricsByCategoryForSurveys(ctx, rtIDs, surveyIDs, "summary", filter.PeriodStart, filter.PeriodEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		prevByRT, err = s.summaryRepo.GetLatestMetricsByCategoryForSurveys(ctx, rtIDs, surveyIDs, "summary", prevStart, prevEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	} else {
		currentByRT, err = s.summaryRepo.GetLatestMetricsByCategory(ctx, rtIDs, "summary", filter.PeriodStart, filter.PeriodEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		prevByRT, err = s.summaryRepo.GetLatestMetricsByCategory(ctx, rtIDs, "summary", prevStart, prevEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	}

	metrics := make(map[string]response.DashboardMetricValue, len(metricCatalog))
	for _, mc := range metricCatalog {
		var sum, prevSum int64
		for _, rtID := range rtIDs {
			sum += currentByRT[rtID][mc.MetricKey]
			prevSum += prevByRT[rtID][mc.MetricKey]
		}
		metrics[mc.MetricKey] = response.DashboardMetricValue{
			Label:         mc.Label,
			CurrentValue:  sum,
			PreviousValue: prevSum,
			Delta:         sum - prevSum,
		}
	}

	resp := response.DashboardSummaryResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		WilayahCount: wilayahCount,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		Metrics:      metrics,
	}

	return utils.SendData(resp, "Berhasil mengambil data summary dashboard")
}

func (s *dashboardService) GetSampah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	period_start := param.Get("period_start")
	period_end := param.Get("period_end")

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if period_end != "" {
		v, err := time.Parse("2006-01-02", period_end)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}

	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)
	// SURVEYOR tidak terikat hierarki wilayah; cakupan datanya ditentukan oleh
	// survey yang di-assign ke dirinya.
	if role != int(enums.ROLE_SURVEYOR) && filter.WilayahLevel != role {
		return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
	}

	var (
		rtIDs        []int64
		wilayahCount *response.WilayahCount
		wilayahInfo  response.WilayahInfo
		surveyIDs    []int64 // hanya diisi untuk SURVEYOR
	)

	if role == int(enums.ROLE_SURVEYOR) {
		// SURVEYOR: cakupan RT & data ditentukan oleh survey yang di-assign.
		surveyIDs, err = s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		if len(surveyIDs) == 0 {
			return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
		}

		rtIDs, err = s.wilayahRepo.GetRTIDsBySurveyIDs(ctx, surveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		// Persempit RT surveyor berdasarkan filter wilayah yang dikirim.
		rtIDs, err = s.filterSurveyorRTIDsByWilayah(ctx, rtIDs, filter)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		totalRT := len(rtIDs)
		wilayahCount = &response.WilayahCount{TotalRT: &totalRT}
		wilayahInfo = response.WilayahInfo{Nama: "Survey yang ditugaskan"}
	} else {
		rtIDs, wilayahCount, wilayahInfo, err = s.resolveFilteredWilayah(ctx, filter, respondentLogin)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	if len(rtIDs) == 0 && role != int(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
	}

	metricCatalog, err := s.metricRepo.List(ctx, "sampah")
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	var byRT map[int64]map[string]int64
	if role == int(enums.ROLE_SURVEYOR) {
		// Metrik dibatasi HANYA pada survey yang di-assign ke surveyor.
		byRT, err = s.summaryRepo.GetLatestMetricsByCategoryForSurveys(ctx, rtIDs, surveyIDs, "sampah", filter.PeriodStart, filter.PeriodEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	} else {
		byRT, err = s.summaryRepo.GetLatestMetricsByCategory(ctx, rtIDs, "sampah", filter.PeriodStart, filter.PeriodEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	}

	values := make(map[string]int64, len(metricCatalog))
	var totalUnit int64
	for _, mc := range metricCatalog {
		var sum int64
		for _, rtID := range rtIDs {
			sum += byRT[rtID][mc.MetricKey]
		}
		values[mc.MetricKey] = sum
		totalUnit += sum
	}

	items := make([]response.SampahItem, 0, len(metricCatalog))
	for _, mc := range metricCatalog {
		val := values[mc.MetricKey]
		var pct float64
		if totalUnit > 0 {
			pct = (float64(val) / float64(totalUnit)) * 100
		}
		items = append(items, response.SampahItem{
			MetricKey:  mc.MetricKey,
			Label:      mc.Label,
			Value:      val,
			Percentage: pct,
		})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Value > items[j].Value })

	resp := response.DashboardSampahResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		WilayahCount: wilayahCount,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		TotalUnit:    totalUnit,
		Items:        items,
	}

	return utils.SendData(resp, "")
}

var infraCardDefinitions = []struct {
	Label           string
	NumberMetricKey string
}{
	{"Jalanan Umum", "jalanan_umum_total"},
	{"Jalanan Lingkungan", "jalanan_lingkungan_total"},
	{"Rumah Tidak Layak", "rumah_tidak_layak"},
	{"Septic Tarik Pribadi", "septic_tarik_pribadi"},
	{"MCK Umum", "mck_umum"},
}

func (s *dashboardService) GetInfrastruktur(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	period_start := param.Get("period_start")
	period_end := param.Get("period_end")

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if period_end != "" {
		v, err := time.Parse("2006-01-02", period_end)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}

	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	if filter.WilayahLevel != int(*respondentLogin.RoleId) {
		return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
	}

	rtIDs, wilayahCount, wilayahInfo, err := s.resolveFilteredWilayah(ctx, filter, respondentLogin)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if len(rtIDs) == 0 {
		return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
	}

	infraKeys := make([]string, 0, len(infraCardDefinitions))
	for _, def := range infraCardDefinitions {
		infraKeys = append(infraKeys, def.NumberMetricKey)
	}

	numberByRT, err := s.summaryRepo.GetLatestMetricsByKeys(ctx, rtIDs, infraKeys, filter.PeriodStart, filter.PeriodEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	items := make([]response.InfrastrukturItem, 0, len(infraCardDefinitions))
	for _, def := range infraCardDefinitions {
		var sum int64
		for _, rtID := range rtIDs {
			sum += numberByRT[rtID][def.NumberMetricKey]
		}
		items = append(items, response.InfrastrukturItem{
			Label: def.Label,
			Value: sum,
		})
	}

	resp := response.DashboardInfrastrukturResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		WilayahCount: wilayahCount,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		Items:        items,
	}

	return utils.SendData(resp, "")
}

func (s *dashboardService) GetStuntingVsRTLH(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	period_start := param.Get("period_start")
	period_end := param.Get("period_end")

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if period_end != "" {
		v, err := time.Parse("2006-01-02", period_end)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}
	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}
	if filter.WilayahLevel != int(*respondentLogin.RoleId) {
		return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
	}
	if filter.WilayahLevel == int(enums.ROLE_RT) {
		return utils.SendError(errors.New("chart perbandingan tidak tersedia di level RT"), http.StatusBadRequest)
	}

	var children []repository.WilayahChild
	var wilayahInfo response.WilayahInfo
	var requestedChildIDs []int64

	switch filter.WilayahLevel {
	case int(enums.ROLE_RW):
		if filter.WilayahID == nil || respondentLogin.RWId == nil || *filter.WilayahID != *respondentLogin.RWId {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		children, err = s.wilayahRepo.GetChildrenOfRW(ctx, *filter.WilayahID)
		requestedChildIDs = filter.RTIds

	case int(enums.ROLE_KELURAHAN):
		if filter.WilayahID == nil || respondentLogin.KelurahanId == nil || *filter.WilayahID != *respondentLogin.KelurahanId {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		children, err = s.wilayahRepo.GetChildrenOfKelurahan(ctx, *filter.WilayahID)
		requestedChildIDs = filter.RWIds

	case int(enums.ROLE_KECAMATAN):
		if filter.WilayahID == nil || respondentLogin.KecamatanId == nil || *filter.WilayahID != *respondentLogin.KecamatanId {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		children, err = s.wilayahRepo.GetChildrenOfKecamatan(ctx, *filter.WilayahID)
		requestedChildIDs = filter.KelurahanIds

	case int(enums.ROLE_ADMIN), int(enums.ROLE_WALIKOTA):
		// admin/walikota: tidak butuh wilayah_id, titik = per Kecamatan se-kota,
		// dipersempit berjenjang lewat kecamatan_ids kalau dikirim
		children, err = s.wilayahRepo.GetChildrenOfCity(ctx)
		requestedChildIDs = filter.KecamatanIds

	default:
		return utils.SendError(errors.New("wilayah_level tidak didukung untuk chart ini"), http.StatusBadRequest)
	}
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if len(children) == 0 {
		return utils.SendError(errors.New("wilayah tidak memiliki anak wilayah"), http.StatusNotFound)
	}

	// validasi filter granular: id yang diminta harus termasuk anak langsung dari wilayah_id ini
	if len(requestedChildIDs) > 0 {
		allChildIDs := make([]int64, 0, len(children))
		for _, c := range children {
			allChildIDs = append(allChildIDs, c.ID)
		}
		selectedIDs, err := intersectAndValidate(requestedChildIDs, allChildIDs, "id yang diminta tidak termasuk dalam wilayah ini")
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
		selectedSet := make(map[int64]bool, len(selectedIDs))
		for _, id := range selectedIDs {
			selectedSet[id] = true
		}
		filtered := make([]repository.WilayahChild, 0, len(selectedIDs))
		for _, c := range children {
			if selectedSet[c.ID] {
				filtered = append(filtered, c)
			}
		}
		children = filtered
	}

	allRTIDs := make([]int64, 0)
	for _, c := range children {
		allRTIDs = append(allRTIDs, c.RTIDs...)
	}

	byRT, err := s.summaryRepo.GetLatestMetricsByCategory(ctx, allRTIDs, "summary", filter.PeriodStart, filter.PeriodEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	points := make([]response.ScatterPoint, 0, len(children))
	var sumX, sumY float64
	for _, c := range children {
		var x, y int64
		for _, rtID := range c.RTIDs {
			x += byRT[rtID]["rumah_tidak_layak"]
			y += byRT[rtID]["total_stunting"]
		}
		points = append(points, response.ScatterPoint{Name: c.Label, X: float64(x), Y: float64(y)})
		sumX += float64(x)
		sumY += float64(y)
	}

	thresholdX := sumX / float64(len(points))
	thresholdY := sumY / float64(len(points))

	resp := response.DashboardScatterResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		XAxisLabel:   "Rumah Tidak Layak Huni (Unit)",
		YAxisLabel:   "Stunting (Orang)",
		ThresholdX:   thresholdX,
		ThresholdY:   thresholdY,
		Points:       points,
	}

	return utils.SendData(resp, "")
}

var indoMonths = []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agst", "Sep", "Okt", "Nov", "Des"}

func (s *dashboardService) GetTrend(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	metricKey := param.Get("metric_key")
	if metricKey == "" {
		metricKey = "total_stunting"
	}

	period_start := param.Get("period_start")
	period_end := param.Get("period_end")
	if period_end == "" {
		return utils.SendError(errors.New("period_end wajib diisi"), http.StatusBadRequest)
	}
	referenceDate, err := time.Parse("2006-01-02", period_end)
	if err != nil {
		return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
	}

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	periodeEnd = referenceDate

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}
	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)
	// SURVEYOR tidak terikat hierarki wilayah; cakupan datanya ditentukan oleh
	// survey yang di-assign ke dirinya, sehingga guard level wilayah dilewati.
	if role != int(enums.ROLE_SURVEYOR) {
		if filter.WilayahLevel != role {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		if filter.WilayahLevel != int(enums.ROLE_ADMIN) && filter.WilayahLevel != int(enums.ROLE_WALIKOTA) && filter.WilayahLevel != int(enums.ROLE_RW) {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke level wilayah ini"), http.StatusBadRequest)
		}
	}

	metric, err := s.metricRepo.FindByMetricKey(ctx, metricKey)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if metric == nil {
		return utils.SendError(errors.New("metric_key tidak dikenali"), http.StatusBadRequest)
	}

	var (
		rtIDs       []int64
		wilayahInfo response.WilayahInfo
		surveyIDs   []int64 // hanya diisi untuk SURVEYOR
	)

	if role == int(enums.ROLE_SURVEYOR) {
		surveyIDs, err = s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		if len(surveyIDs) == 0 {
			return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
		}
		rtIDs, err = s.wilayahRepo.GetRTIDsBySurveyIDs(ctx, surveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		wilayahInfo = response.WilayahInfo{Nama: "Survey yang ditugaskan"}
		if len(rtIDs) == 0 {
			return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
		}
	} else {
		rtIDs, _, wilayahInfo, err = s.resolveFilteredWilayah(ctx, filter, respondentLogin)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
		isAdminOrWalikota := filter.WilayahLevel == int(enums.ROLE_ADMIN) || filter.WilayahLevel == int(enums.ROLE_WALIKOTA)
		if !isAdminOrWalikota && len(rtIDs) == 0 {
			return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
		}
	}

	curWindowStart := time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -5, 0)
	curWindowEnd := time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Add(-time.Second)
	prevWindowStart := curWindowStart.AddDate(-1, 0, 0)
	prevWindowEnd := curWindowEnd.AddDate(-1, 0, 0)

	var curRows, prevRows []repository.MonthlyMetricRow
	if role == int(enums.ROLE_SURVEYOR) {
		curRows, err = s.summaryRepo.GetMonthlyMetricSumsForSurveys(ctx, rtIDs, surveyIDs, metricKey, curWindowStart, curWindowEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		prevRows, err = s.summaryRepo.GetMonthlyMetricSumsForSurveys(ctx, rtIDs, surveyIDs, metricKey, prevWindowStart, prevWindowEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	} else {
		curRows, err = s.summaryRepo.GetMonthlyMetricSums(ctx, rtIDs, metricKey, curWindowStart, curWindowEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		prevRows, err = s.summaryRepo.GetMonthlyMetricSums(ctx, rtIDs, metricKey, prevWindowStart, prevWindowEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	}

	curByMonth := make(map[string]int64, len(curRows))
	for _, r := range curRows {
		curByMonth[r.MonthBucket.Format("2006-01")] = r.Total
	}
	prevByMonth := make(map[string]int64, len(prevRows))
	for _, r := range prevRows {
		prevByMonth[r.MonthBucket.Format("2006-01")] = r.Total
	}

	months := make([]response.TrendMonthPoint, 6)
	for i := 0; i < 6; i++ {
		targetMonth := curWindowStart.AddDate(0, i, 0)
		prevTargetMonth := targetMonth.AddDate(-1, 0, 0)

		months[i] = response.TrendMonthPoint{
			MonthLabel: indoMonths[int(targetMonth.Month())-1] + " " + strconv.Itoa(targetMonth.Year()),
			TahunIni:   curByMonth[targetMonth.Format("2006-01")],
			TahunLalu:  prevByMonth[prevTargetMonth.Format("2006-01")],
		}
	}

	resp := response.DashboardTrendResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		MetricKey:    metric.MetricKey,
		Label:        metric.Label,
		Months:       months,
	}

	return utils.SendData(resp, "")
}

func (s *dashboardService) GetComparison(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	period_start := param.Get("period_start")
	period_end := param.Get("period_end")

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if period_end != "" {
		v, err := time.Parse("2006-01-02", period_end)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	metricKeyA := param.Get("metric_key_a")
	if metricKeyA == "" {
		metricKeyA = "total_stunting"
	}
	metricKeyB := param.Get("metric_key_b")
	if metricKeyB == "" {
		metricKeyB = "rumah_tidak_layak"
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}
	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)
	// SURVEYOR tidak terikat hierarki wilayah; perbandingan dikelompokkan
	// PER SURVEY yang di-assign ke dirinya (tiap batang = satu survey).
	if role != int(enums.ROLE_SURVEYOR) {
		if filter.WilayahLevel != role {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		if filter.WilayahLevel == int(enums.ROLE_RT) {
			return utils.SendError(errors.New("chart perbandingan tidak tersedia di level RT"), http.StatusBadRequest)
		}
	}

	metricA, err := s.metricRepo.FindByMetricKey(ctx, metricKeyA)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if metricA == nil {
		return utils.SendError(errors.New("metric_key_a tidak dikenali"), http.StatusBadRequest)
	}
	metricB, err := s.metricRepo.FindByMetricKey(ctx, metricKeyB)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if metricB == nil {
		return utils.SendError(errors.New("metric_key_b tidak dikenali"), http.StatusBadRequest)
	}

	var wilayahInfo response.WilayahInfo

	// =========================================================================
	// SURVEYOR: perbandingan dikelompokkan per survey yang di-assign ke dirinya.
	// =========================================================================
	if role == int(enums.ROLE_SURVEYOR) {
		surveyIDs, err := s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		if len(surveyIDs) == 0 {
			return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
		}

		// nama survey (id DESC) + RT per survey
		surveyOpts, err := s.summaryRepo.ListSurveysLatestFirstByIDs(ctx, surveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		rtBySurvey, err := s.wilayahRepo.GetRTIDsGroupedBySurvey(ctx, surveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}

		// Persempit RT tiap survey berdasarkan filter wilayah yang dikirim.
		for sid, rts := range rtBySurvey {
			filtered, err := s.filterSurveyorRTIDsByWilayah(ctx, rts, filter)
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}
			rtBySurvey[sid] = filtered
		}

		// kumpulkan semua RT untuk query metrik per-survey
		allRTIDs := make([]int64, 0)
		for _, rts := range rtBySurvey {
			allRTIDs = append(allRTIDs, rts...)
		}
		// Catatan: allRTIDs boleh kosong untuk surveyor (filter wilayah tidak
		// beririsan). Jangan error — biarkan hasil menjadi data kosong (nilai 0).

		// data per survey -> rt -> metric (nilai tiap survey TIDAK tercampur)
		perSurvey, err := s.summaryRepo.GetMetricSumsByCategoryPerSurvey(ctx, allRTIDs, surveyIDs, "summary", filter.PeriodStart, filter.PeriodEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}

		items := make([]response.DashboardComparisonItem, 0, len(surveyOpts))
		for _, opt := range surveyOpts {
			var valA, valB int64
			if byRT := perSurvey[opt.SurveyID]; byRT != nil {
				for _, rtID := range rtBySurvey[opt.SurveyID] {
					if m := byRT[rtID]; m != nil {
						valA += m[metricKeyA]
						valB += m[metricKeyB]
					}
				}
			}
			items = append(items, response.DashboardComparisonItem{
				ChildLabel: opt.SurveyName,
				ValueA:     valA,
				ValueB:     valB,
			})
		}

		wilayahInfo = response.WilayahInfo{Nama: "Survey yang ditugaskan"}

		resp := response.DashboardComparisonResponse{
			WilayahLevel: filter.WilayahLevel,
			WilayahInfo:  wilayahInfo,
			DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
			SeriesA:      response.DashboardComparisonSeries{MetricKey: metricA.MetricKey, Label: metricA.Label},
			SeriesB:      response.DashboardComparisonSeries{MetricKey: metricB.MetricKey, Label: metricB.Label},
			Items:        items,
		}
		return utils.SendData(resp, "")
	}

	// =========================================================================
	// Role lain (ADMIN/WALIKOTA/KECAMATAN/KELURAHAN/RW): LOGIKA LAMA, tidak diubah.
	// =========================================================================
	var children []repository.WilayahChild
	var requestedChildIDs []int64

	switch filter.WilayahLevel {
	case int(enums.ROLE_RW):
		if filter.WilayahID == nil || respondentLogin.RWId == nil || *filter.WilayahID != *respondentLogin.RWId {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		children, err = s.wilayahRepo.GetChildrenOfRW(ctx, *filter.WilayahID)
		requestedChildIDs = filter.RTIds

	case int(enums.ROLE_KELURAHAN):
		if filter.WilayahID == nil || respondentLogin.KelurahanId == nil || *filter.WilayahID != *respondentLogin.KelurahanId {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		children, err = s.wilayahRepo.GetChildrenOfKelurahan(ctx, *filter.WilayahID)
		requestedChildIDs = filter.RWIds

	case int(enums.ROLE_KECAMATAN):
		if filter.WilayahID == nil || respondentLogin.KecamatanId == nil || *filter.WilayahID != *respondentLogin.KecamatanId {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		children, err = s.wilayahRepo.GetChildrenOfKecamatan(ctx, *filter.WilayahID)
		requestedChildIDs = filter.KelurahanIds

	case int(enums.ROLE_ADMIN), int(enums.ROLE_WALIKOTA):
		children, err = s.wilayahRepo.GetChildrenOfCity(ctx)
		requestedChildIDs = filter.KecamatanIds

	default:
		return utils.SendError(errors.New("wilayah_level tidak didukung untuk chart ini"), http.StatusBadRequest)
	}
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if len(children) == 0 {
		return utils.SendError(errors.New("wilayah tidak memiliki anak wilayah"), http.StatusNotFound)
	}

	if len(requestedChildIDs) > 0 {
		allChildIDs := make([]int64, 0, len(children))
		for _, c := range children {
			allChildIDs = append(allChildIDs, c.ID)
		}
		selectedIDs, err := intersectAndValidate(requestedChildIDs, allChildIDs, "id yang diminta tidak termasuk dalam wilayah ini")
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
		selectedSet := make(map[int64]bool, len(selectedIDs))
		for _, id := range selectedIDs {
			selectedSet[id] = true
		}
		filtered := make([]repository.WilayahChild, 0, len(selectedIDs))
		for _, c := range children {
			if selectedSet[c.ID] {
				filtered = append(filtered, c)
			}
		}
		children = filtered
	}

	allRTIDs := make([]int64, 0)
	for _, c := range children {
		allRTIDs = append(allRTIDs, c.RTIDs...)
	}

	byRT, err := s.summaryRepo.GetLatestMetricsByCategory(ctx, allRTIDs, "summary", filter.PeriodStart, filter.PeriodEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	items := make([]response.DashboardComparisonItem, 0, len(children))
	for _, c := range children {
		var valA, valB int64
		for _, rtID := range c.RTIDs {
			valA += byRT[rtID][metricKeyA]
			valB += byRT[rtID][metricKeyB]
		}
		items = append(items, response.DashboardComparisonItem{
			ChildLabel: c.Label,
			ValueA:     valA,
			ValueB:     valB,
		})
	}

	resp := response.DashboardComparisonResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		SeriesA:      response.DashboardComparisonSeries{MetricKey: metricA.MetricKey, Label: metricA.Label},
		SeriesB:      response.DashboardComparisonSeries{MetricKey: metricB.MetricKey, Label: metricB.Label},
		Items:        items,
	}

	return utils.SendData(resp, "")
}

func (s *dashboardService) GetHeatmap(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	period_start := param.Get("period_start")
	period_end := param.Get("period_end")

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if period_end != "" {
		v, err := time.Parse("2006-01-02", period_end)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}
	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)

	var wilayahInfo response.WilayahInfo

	// =========================================================================
	// SURVEYOR: heatmap dikelompokkan PER SURVEY yang di-assign ke dirinya.
	// Nilai tiap zona = jumlah responden tervalidasi pada survey tersebut.
	// =========================================================================
	if role == int(enums.ROLE_SURVEYOR) {
		surveyIDs, err := s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		if len(surveyIDs) == 0 {
			return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
		}

		surveyOpts, err := s.summaryRepo.ListSurveysLatestFirstByIDs(ctx, surveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		rtBySurvey, err := s.wilayahRepo.GetRTIDsGroupedBySurvey(ctx, surveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}

		items := make([]response.DashboardHeatmapItem, 0, len(surveyOpts))
		for _, opt := range surveyOpts {
			rts := rtBySurvey[opt.SurveyID]
			var total int64
			if len(rts) > 0 {
				// batasi hitung responden hanya pada survey ini
				countByRT, err := s.summaryRepo.CountValidatedRespondentsByRTForSurveys(ctx, rts, []int64{opt.SurveyID}, filter.PeriodStart, filter.PeriodEnd)
				if err != nil {
					return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
				}
				for _, rtID := range rts {
					total += countByRT[rtID]
				}
			}
			items = append(items, response.DashboardHeatmapItem{
				Label:   opt.SurveyName,
				GeoName: nil,
				Value:   total,
			})
		}

		wilayahInfo = response.WilayahInfo{Nama: "Survey yang ditugaskan"}

		resp := response.DashboardHeatmapResponse{
			WilayahLevel: filter.WilayahLevel,
			WilayahInfo:  wilayahInfo,
			DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
			Items:        items,
		}
		return utils.SendData(resp, "")
	}

	// =========================================================================
	// Role lain: LOGIKA LAMA, tidak diubah.
	// =========================================================================
	if filter.WilayahLevel != role {
		return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
	}

	// heatmap zona cuma untuk 2 role ini
	if filter.WilayahLevel != int(enums.ROLE_KECAMATAN) &&
		filter.WilayahLevel != int(enums.ROLE_ADMIN) &&
		filter.WilayahLevel != int(enums.ROLE_WALIKOTA) {
		return utils.SendError(errors.New("chart heatmap zona hanya tersedia untuk role kecamatan, admin, atau walikota"), http.StatusBadRequest)
	}

	var children []repository.WilayahChild
	var requestedChildIDs []int64

	switch filter.WilayahLevel {
	case int(enums.ROLE_KECAMATAN):
		if filter.WilayahID == nil || respondentLogin.KecamatanId == nil || *filter.WilayahID != *respondentLogin.KecamatanId {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		children, err = s.wilayahRepo.GetChildrenOfKecamatan(ctx, *filter.WilayahID)
		requestedChildIDs = filter.KelurahanIds

	case int(enums.ROLE_ADMIN), int(enums.ROLE_WALIKOTA):
		children, err = s.wilayahRepo.GetChildrenOfCity(ctx)
		requestedChildIDs = filter.KecamatanIds
	}
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if len(children) == 0 {
		return utils.SendError(errors.New("wilayah tidak memiliki anak wilayah"), http.StatusNotFound)
	}

	if len(requestedChildIDs) > 0 {
		allChildIDs := make([]int64, 0, len(children))
		for _, c := range children {
			allChildIDs = append(allChildIDs, c.ID)
		}
		selectedIDs, err := intersectAndValidate(requestedChildIDs, allChildIDs, "id yang diminta tidak termasuk dalam wilayah ini")
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
		selectedSet := make(map[int64]bool, len(selectedIDs))
		for _, id := range selectedIDs {
			selectedSet[id] = true
		}
		filtered := make([]repository.WilayahChild, 0, len(selectedIDs))
		for _, c := range children {
			if selectedSet[c.ID] {
				filtered = append(filtered, c)
			}
		}
		children = filtered
	}

	allRTIDs := make([]int64, 0)
	for _, c := range children {
		allRTIDs = append(allRTIDs, c.RTIDs...)
	}

	countByRT, err := s.summaryRepo.CountValidatedRespondentsByRT(ctx, allRTIDs, filter.PeriodStart, filter.PeriodEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	items := make([]response.DashboardHeatmapItem, 0, len(children))
	for _, c := range children {
		var total int64
		for _, rtID := range c.RTIDs {
			total += countByRT[rtID]
		}
		items = append(items, response.DashboardHeatmapItem{
			Label:   c.Label,
			GeoName: c.GeoName,
			Value:   total,
		})
	}

	resp := response.DashboardHeatmapResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		Items:        items,
	}

	return utils.SendData(resp, "")
}

type PeneranganJalanCardDefinition struct {
	Label          string
	TotalMetricKey string
	RusakMetricKey string
}

var peneranganJalanCardDefinitions = []PeneranganJalanCardDefinition{
	{Label: "Penerangan Jalan Umum (PJU)", TotalMetricKey: "pju_total", RusakMetricKey: "pju_rusak"},
	{Label: "Penerangan Jalan Gang (PJG)", TotalMetricKey: "pjg_total", RusakMetricKey: "pjg_rusak"},
	{Label: "Penerangan Jalan Lingkungan (PJL)", TotalMetricKey: "pjl_total", RusakMetricKey: "pjl_rusak"},
}

func (s *dashboardService) GetPeneranganJalan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	period_start := param.Get("period_start")
	period_end := param.Get("period_end")

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if period_end != "" {
		v, err := time.Parse("2006-01-02", period_end)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}

	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)
	// SURVEYOR tidak terikat hierarki wilayah; cakupan datanya ditentukan oleh
	// survey yang di-assign ke dirinya.
	if role != int(enums.ROLE_SURVEYOR) && filter.WilayahLevel != role {
		return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
	}

	var (
		rtIDs        []int64
		wilayahCount *response.WilayahCount
		wilayahInfo  response.WilayahInfo
		surveyIDs    []int64 // hanya diisi untuk SURVEYOR
	)

	if role == int(enums.ROLE_SURVEYOR) {
		surveyIDs, err = s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		if len(surveyIDs) == 0 {
			return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
		}

		rtIDs, err = s.wilayahRepo.GetRTIDsBySurveyIDs(ctx, surveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		// Persempit RT surveyor berdasarkan filter wilayah yang dikirim.
		rtIDs, err = s.filterSurveyorRTIDsByWilayah(ctx, rtIDs, filter)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		totalRT := len(rtIDs)
		wilayahCount = &response.WilayahCount{TotalRT: &totalRT}
		wilayahInfo = response.WilayahInfo{Nama: "Survey yang ditugaskan"}
	} else {
		rtIDs, wilayahCount, wilayahInfo, err = s.resolveFilteredWilayah(ctx, filter, respondentLogin)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
	}

	if len(rtIDs) == 0 && role != int(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
	}

	pjKeys := make([]string, 0, len(peneranganJalanCardDefinitions)*2)
	for _, def := range peneranganJalanCardDefinitions {
		pjKeys = append(pjKeys, def.TotalMetricKey, def.RusakMetricKey)
	}

	var numberByRT map[int64]map[string]int64
	if role == int(enums.ROLE_SURVEYOR) {
		// Metrik dibatasi HANYA pada survey yang di-assign ke surveyor.
		numberByRT, err = s.summaryRepo.GetLatestMetricsByKeysForSurveys(ctx, rtIDs, surveyIDs, pjKeys, filter.PeriodStart, filter.PeriodEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	} else {
		numberByRT, err = s.summaryRepo.GetLatestMetricsByKeys(ctx, rtIDs, pjKeys, filter.PeriodStart, filter.PeriodEnd)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	}

	items := make([]response.PeneranganJalanItem, 0, len(peneranganJalanCardDefinitions))
	for _, def := range peneranganJalanCardDefinitions {
		var total, rusak int64
		for _, rtID := range rtIDs {
			total += numberByRT[rtID][def.TotalMetricKey]
			rusak += numberByRT[rtID][def.RusakMetricKey]
		}

		baik := total - rusak
		if baik < 0 {
			baik = 0 // jaga-jaga kalau data rusak > total (data tidak konsisten), jangan sampai negatif
		}

		var persenBerfungsi float64
		if total > 0 {
			persenBerfungsi = math.Round((float64(baik)/float64(total))*10000) / 100 // 2 desimal
		}

		items = append(items, response.PeneranganJalanItem{
			Label:           def.Label,
			Total:           total,
			Baik:            baik,
			Rusak:           rusak,
			PersenBerfungsi: persenBerfungsi,
		})
	}

	resp := response.DashboardPeneranganJalanResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		WilayahCount: wilayahCount,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		Items:        items,
	}

	return utils.SendData(resp, "")
}

func (s *dashboardService) GetTrendCompare(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayah_level, _ := strconv.Atoi(param.Get("wilayah_level"))
	metricKey := param.Get("metric_key")
	if metricKey == "" {
		metricKey = "total_stunting"
	}

	period_start := param.Get("period_start")
	period_end := param.Get("period_end")
	if period_end == "" {
		return utils.SendError(errors.New("period_end wajib diisi"), http.StatusBadRequest)
	}
	referenceDate, err := time.Parse("2006-01-02", period_end)
	if err != nil {
		return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
	}

	var periodeStart, periodeEnd time.Time
	if period_start != "" {
		v, err := time.Parse("2006-01-02", period_start)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	periodeEnd = referenceDate

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayah_level,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}
	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)
	// SURVEYOR tidak terikat hierarki wilayah.
	if role != int(enums.ROLE_SURVEYOR) {
		if filter.WilayahLevel != role {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
		}
		if filter.WilayahLevel != int(enums.ROLE_ADMIN) && filter.WilayahLevel != int(enums.ROLE_WALIKOTA) && filter.WilayahLevel != int(enums.ROLE_RW) {
			return utils.SendError(errors.New("Anda tidak memiliki akses ke level wilayah ini"), http.StatusBadRequest)
		}
	}

	metric, err := s.metricRepo.FindByMetricKey(ctx, metricKey)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if metric == nil {
		return utils.SendError(errors.New("metric_key tidak dikenali"), http.StatusBadRequest)
	}

	var (
		rtIDs       []int64
		wilayahInfo response.WilayahInfo
		surveyOpts  []repository.SurveyOptionRow
	)

	if role == int(enums.ROLE_SURVEYOR) {
		assignedSurveyIDs, err := s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		if len(assignedSurveyIDs) == 0 {
			return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
		}
		rtIDs, err = s.wilayahRepo.GetRTIDsBySurveyIDs(ctx, assignedSurveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		wilayahInfo = response.WilayahInfo{Nama: "Survey yang ditugaskan"}
		if len(rtIDs) == 0 {
			return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
		}
		// opsi survey = hanya survey milik surveyor (id DESC)
		surveyOpts, err = s.summaryRepo.ListSurveysLatestFirstByIDs(ctx, assignedSurveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	} else {
		rtIDs, _, wilayahInfo, err = s.resolveFilteredWilayah(ctx, filter, respondentLogin)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}
		isAdminOrWalikota := filter.WilayahLevel == int(enums.ROLE_ADMIN) || filter.WilayahLevel == int(enums.ROLE_WALIKOTA)
		if !isAdminOrWalikota && len(rtIDs) == 0 {
			return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
		}
		surveyOpts, err = s.summaryRepo.ListSurveysLatestFirst(ctx, 0)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	}

	// window 6 bulan terakhir dari period_end
	windowStart := time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -5, 0)
	windowEnd := time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Add(-time.Second)

	// himpunan survey yang diizinkan (untuk surveyor: hanya miliknya)
	allowedSurvey := make(map[int64]bool, len(surveyOpts))
	for _, opt := range surveyOpts {
		allowedSurvey[opt.SurveyID] = true
	}

	latestSurveyID := int64(0)
	if len(surveyOpts) > 0 {
		latestSurveyID = surveyOpts[0].SurveyID
	}

	requestedSurveyIDs := utils.ParseInt64SliceQueryParam(param.Get("survey_ids"))

	var selectedSurveyIDs []int64
	seen := make(map[int64]bool)
	addSurvey := func(id int64) {
		if id != 0 && !seen[id] && allowedSurvey[id] {
			selectedSurveyIDs = append(selectedSurveyIDs, id)
			seen[id] = true
		}
	}

	if len(requestedSurveyIDs) > 0 {
		addSurvey(latestSurveyID)
		for _, id := range requestedSurveyIDs {
			addSurvey(id)
		}
	} else {
		for i := 0; i < len(surveyOpts) && i < 2; i++ {
			addSurvey(surveyOpts[i].SurveyID)
		}
	}

	// label sumbu-X bersama
	monthLabels := make([]string, 6)
	monthKeys := make([]string, 6)
	for i := 0; i < 6; i++ {
		m := windowStart.AddDate(0, i, 0)
		monthLabels[i] = indoMonths[int(m.Month())-1] + " " + strconv.Itoa(m.Year())
		monthKeys[i] = m.Format("2006-01")
	}

	surveyNames, err := s.summaryRepo.GetSurveyNamesByIDs(ctx, selectedSurveyIDs)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	rows, err := s.summaryRepo.GetMonthlyMetricSumsBySurvey(ctx, rtIDs, selectedSurveyIDs, metricKey, windowStart, windowEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	bySurvey := make(map[int64]map[string]int64, len(selectedSurveyIDs))
	for _, r := range rows {
		if bySurvey[r.SurveyID] == nil {
			bySurvey[r.SurveyID] = make(map[string]int64)
		}
		bySurvey[r.SurveyID][r.MonthBucket.Format("2006-01")] = r.Total
	}

	series := make([]response.TrendSurveySeries, 0, len(selectedSurveyIDs))
	for _, sid := range selectedSurveyIDs {
		points := make([]response.TrendComparePoint, 6)
		for i := 0; i < 6; i++ {
			var val int64
			if m := bySurvey[sid]; m != nil {
				val = m[monthKeys[i]]
			}
			points[i] = response.TrendComparePoint{
				MonthLabel: monthLabels[i],
				Value:      val,
			}
		}
		series = append(series, response.TrendSurveySeries{
			SurveyID:   sid,
			SurveyName: surveyNames[sid],
			IsLatest:   sid == latestSurveyID,
			Points:     points,
		})
	}

	resp := response.DashboardTrendCompareResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		MetricKey:    metric.MetricKey,
		Label:        metric.Label,
		MonthLabels:  monthLabels,
		Series:       series,
	}

	return utils.SendData(resp, "")
}

func (s *dashboardService) GetSummaryCompare(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	wilayahLevel, _ := strconv.Atoi(param.Get("wilayah_level"))
	periodStartRaw := param.Get("period_start")
	periodEndRaw := param.Get("period_end")

	now := time.Now()
	periodeStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodeEnd := now

	if (periodStartRaw == "") != (periodEndRaw == "") {
		return utils.SendError(errors.New("period_start dan period_end harus diisi bersamaan"), http.StatusBadRequest)
	}
	if periodStartRaw != "" {
		v, err := time.Parse("2006-01-02", periodStartRaw)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if periodEndRaw != "" {
		v, err := time.Parse("2006-01-02", periodEndRaw)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}
	if periodeEnd.Before(periodeStart) {
		return utils.SendError(errors.New("period_end tidak boleh sebelum period_start"), http.StatusBadRequest)
	}

	filter := payloads.DashboardFilter{
		WilayahLevel: wilayahLevel,
		WilayahID:    utils.ParseInt64QueryParamPointer(param.Get("wilayah_id")),
		PeriodStart:  periodeStart,
		PeriodEnd:    periodeEnd,
		RTIds:        utils.ParseInt64SliceQueryParam(param.Get("rt_ids")),
		RWIds:        utils.ParseInt64SliceQueryParam(param.Get("rw_ids")),
		KelurahanIds: utils.ParseInt64SliceQueryParam(param.Get("kelurahan_ids")),
		KecamatanIds: utils.ParseInt64SliceQueryParam(param.Get("kecamatan_ids")),
	}

	var validate = validator.New()
	if err := validate.Struct(filter); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}
	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)
	// SURVEYOR tidak terikat hierarki wilayah; cakupan datanya ditentukan oleh
	// survey yang di-assign ke dirinya.
	if role != int(enums.ROLE_ADMIN) && role != int(enums.ROLE_WALIKOTA) && role != int(enums.ROLE_SURVEYOR) && wilayahLevel != role {
		return utils.SendError(errors.New("Anda tidak memiliki akses ke wilayah ini"), http.StatusForbidden)
	}

	var (
		rtIDs        []int64
		wilayahCount *response.WilayahCount
		wilayahInfo  response.WilayahInfo
		surveyOpts   []repository.SurveyOptionRow
	)

	if role == int(enums.ROLE_SURVEYOR) {
		// SURVEYOR: cakupan RT ditentukan oleh survey yang di-assign ke dirinya,
		// dan daftar survey yang bisa dibandingkan juga HANYA survey miliknya.
		assignedSurveyIDs, err := s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		if len(assignedSurveyIDs) == 0 {
			return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
		}

		rtIDs, err = s.wilayahRepo.GetRTIDsBySurveyIDs(ctx, assignedSurveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
		// Persempit RT surveyor berdasarkan filter wilayah yang dikirim.
		rtIDs, err = s.filterSurveyorRTIDsByWilayah(ctx, rtIDs, filter)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		totalRT := len(rtIDs)
		wilayahCount = &response.WilayahCount{TotalRT: &totalRT}
		wilayahInfo = response.WilayahInfo{Nama: "Survey yang ditugaskan"}

		// opsi survey = hanya survey milik surveyor (id DESC = terbaru dulu)
		surveyOpts, err = s.summaryRepo.ListSurveysLatestFirstByIDs(ctx, assignedSurveyIDs)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	} else {
		rtIDs, wilayahCount, wilayahInfo, err = s.resolveFilteredWilayah(ctx, filter, respondentLogin)
		if err != nil {
			return utils.SendError(err, http.StatusBadRequest)
		}

		// opsi survey = seluruh survey (id DESC)
		surveyOpts, err = s.summaryRepo.ListSurveysLatestFirst(ctx, 0)
		if err != nil {
			return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
		}
	}

	if len(rtIDs) == 0 && role != int(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("wilayah tidak memiliki data RT"), http.StatusNotFound)
	}

	metricCatalog, err := s.metricRepo.List(ctx, "summary")
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	// himpunan survey yang diizinkan (untuk surveyor: hanya miliknya)
	allowedSurvey := make(map[int64]bool, len(surveyOpts))
	for _, opt := range surveyOpts {
		allowedSurvey[opt.SurveyID] = true
	}

	latestSurveyID := int64(0)
	if len(surveyOpts) > 0 {
		latestSurveyID = surveyOpts[0].SurveyID
	}

	// tentukan survey yang jadi bar (terbaru fixed + pilihan / default 2 terbaru)
	requestedSurveyIDs := utils.ParseInt64SliceQueryParam(param.Get("survey_ids"))
	var selectedSurveyIDs []int64
	seen := make(map[int64]bool)
	addSurvey := func(id int64) {
		if id != 0 && !seen[id] && allowedSurvey[id] {
			selectedSurveyIDs = append(selectedSurveyIDs, id)
			seen[id] = true
		}
	}
	if len(requestedSurveyIDs) > 0 {
		addSurvey(latestSurveyID)
		for _, id := range requestedSurveyIDs {
			addSurvey(id)
		}
	} else {
		for i := 0; i < len(surveyOpts) && i < 2; i++ {
			addSurvey(surveyOpts[i].SurveyID)
		}
	}

	surveyNames, err := s.summaryRepo.GetSurveyNamesByIDs(ctx, selectedSurveyIDs)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	// data per survey -> rt -> metric
	perSurvey, err := s.summaryRepo.GetMetricSumsByCategoryPerSurvey(ctx, rtIDs, selectedSurveyIDs, "summary", filter.PeriodStart, filter.PeriodEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	// series (urutan = selectedSurveyIDs)
	series := make([]response.SummaryCompareSeries, 0, len(selectedSurveyIDs))
	for _, sid := range selectedSurveyIDs {
		series = append(series, response.SummaryCompareSeries{
			SurveyID:   sid,
			SurveyName: surveyNames[sid],
			IsLatest:   sid == latestSurveyID,
		})
	}

	// items (per metrik), nilai tiap survey selaras urutan series
	items := make([]response.SummaryCompareItem, 0, len(metricCatalog))
	for _, mc := range metricCatalog {
		values := make([]response.SummaryCompareMetricValue, 0, len(selectedSurveyIDs))
		for _, sid := range selectedSurveyIDs {
			var sum int64
			if byRT := perSurvey[sid]; byRT != nil {
				for _, rtID := range rtIDs {
					if m := byRT[rtID]; m != nil {
						sum += m[mc.MetricKey]
					}
				}
			}
			values = append(values, response.SummaryCompareMetricValue{
				SurveyID: sid,
				Value:    sum, // default 0
			})
		}
		items = append(items, response.SummaryCompareItem{
			MetricKey: mc.MetricKey,
			Label:     mc.Label,
			Values:    values,
		})
	}

	resp := response.DashboardSummaryCompareResponse{
		WilayahLevel: filter.WilayahLevel,
		WilayahInfo:  wilayahInfo,
		WilayahCount: wilayahCount,
		DataDiambil:  time.Now().Format("02 January 2006 15:04 WIB"),
		Series:       series,
		Items:        items,
	}

	return utils.SendData(resp, "Berhasil mengambil data perbandingan survei")
}

func (s *dashboardService) GetTopWilayahSurvei(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	// ---- parsing param: wilayah_level, period_start, period_end, top ----
	wilayahLevel, _ := strconv.Atoi(param.Get("wilayah_level"))

	periodStartRaw := param.Get("period_start")
	periodEndRaw := param.Get("period_end")

	now := time.Now()
	periodeStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodeEnd := now

	if (periodStartRaw == "") != (periodEndRaw == "") {
		return utils.SendError(errors.New("period_start dan period_end harus diisi bersamaan"), http.StatusBadRequest)
	}
	if periodStartRaw != "" {
		v, err := time.Parse("2006-01-02", periodStartRaw)
		if err != nil {
			return utils.SendError(errors.New("format period_start tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeStart = v
	}
	if periodEndRaw != "" {
		v, err := time.Parse("2006-01-02", periodEndRaw)
		if err != nil {
			return utils.SendError(errors.New("format period_end tidak valid, gunakan format YYYY-MM-DD"), http.StatusBadRequest)
		}
		periodeEnd = v
	}
	if periodeEnd.Before(periodeStart) {
		return utils.SendError(errors.New("period_end tidak boleh sebelum period_start"), http.StatusBadRequest)
	}

	// top = batas jumlah item (0 / kosong = semua)
	topN := 0
	if raw := param.Get("top"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			return utils.SendError(errors.New("parameter top tidak valid"), http.StatusBadRequest)
		}
		topN = v
	}

	// ---- validasi role (endpoint ini khusus surveyor) ----
	respondentLogin, err := s.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent dari UserAPI"), http.StatusInternalServerError)
	}
	if respondentLogin == nil {
		return utils.SendError(errors.New("Data respondent tidak ditemukan"), http.StatusNotFound)
	}
	if respondentLogin.RoleId == nil {
		return utils.SendError(errors.New("Anda tidak memiliki akses"), http.StatusForbidden)
	}

	role := int(*respondentLogin.RoleId)
	if role != int(enums.ROLE_SURVEYOR) {
		return utils.SendError(errors.New("Endpoint ini hanya untuk surveyor"), http.StatusForbidden)
	}
	// wilayah_level dikirim untuk konsistensi FE; untuk surveyor nilainya 8.
	_ = wilayahLevel

	// ---- ambil survey yang di-assign + RT-nya ----
	surveyIDs, err := s.wilayahRepo.GetSurveyIDsBySurveyor(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}
	if len(surveyIDs) == 0 {
		return utils.SendError(errors.New("Anda belum ditugaskan pada survey apa pun"), http.StatusNotFound)
	}

	rtIDs, err := s.wilayahRepo.GetRTIDsBySurveyIDs(ctx, surveyIDs)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	// ---- statistik keterisian per survey (sudah terurut jumlah_data DESC) ----
	stats, err := s.summaryRepo.GetSurveyFillStats(ctx, surveyIDs, rtIDs, periodeStart, periodeEnd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server"), http.StatusInternalServerError)
	}

	// batasi dengan top
	if topN > 0 && len(stats) > topN {
		stats = stats[:topN]
	}

	items := make([]response.DashboardTopWilayahItem, 0, len(stats))
	for _, st := range stats {
		items = append(items, response.DashboardTopWilayahItem{
			KecamatanLabel: st.SurveyName,
			TotalRT:        st.TotalRT,
			JumlahData:     st.JumlahData,
		})
	}

	resp := response.DashboardTopWilayahResponse{
		DataDiambil: time.Now().Format("2006-01-02 15:04:05"),
		Items:       items,
	}
	return utils.SendData(resp, "Berhasil memuat data top wilayah survei")
}
