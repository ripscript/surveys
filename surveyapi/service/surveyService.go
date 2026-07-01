package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/dto"
	"backend/surveyapi/enums"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type SurveyService interface {
	OptionsPeriodeSurvey(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	CreateSurvey(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ApprovalSurvey(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)

	AvailableSurveyWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	PreviewSurveyIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PreviewSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	SubmitSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateSurveyRespondentStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetApprovalHistorySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetApprovalHistorySurveyPerWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetHistoryDetailPerWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	SurveyResultSectionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	VerifySurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	RejectSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	RejectValidateSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ValidateSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetPublicImageSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ExportExcelSurveyResultsPerRT(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ExportExcelSurveyResultsMassal(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	ResetStatusToVerifySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetAllRejectedQuestions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetRejectedQuestionsBySurveyCode(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	SyncExpiredSurveysStatus(ctx context.Context)
}

type surveyService struct {
	manajemenAlurRepo              repository.ManajemenAlurRepo
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo
	templateUcapanRepo             repository.TemplateUcapanRepo
	surveyRepo                     repository.SurveyRepo
	wilayahRepo                    repository.WilayahRepo
	userRepo                       repository.UserRepo
	fileRepo                       repository.FileRepo

	isSyncing atomic.Bool
}

func NewSurveyService(
	manajemenAlurRepo repository.ManajemenAlurRepo,
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo,
	templateUcapanRepo repository.TemplateUcapanRepo,
	surveyRepo repository.SurveyRepo,
	wilayahRepo repository.WilayahRepo,
	userRepo repository.UserRepo,
	fileRepo repository.FileRepo,
) SurveyService {
	return &surveyService{
		manajemenAlurRepo:              manajemenAlurRepo,
		templateFormulirPertanyaanRepo: templateFormulirPertanyaanRepo,
		templateUcapanRepo:             templateUcapanRepo,
		surveyRepo:                     surveyRepo,
		wilayahRepo:                    wilayahRepo,
		userRepo:                       userRepo,
		fileRepo:                       fileRepo,
	}
}

func (service *surveyService) OptionsPeriodeSurvey(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	triwulanID := enums.PeriodeSurveyToInt64(enums.TRIWULAN)
	triwulanLabel := "Triwulan"

	semesterID := enums.PeriodeSurveyToInt64(enums.SEMESTER)
	semesterLabel := "Semester"

	tahunanID := enums.PeriodeSurveyToInt64(enums.TAHUNAN)
	tahunanLabel := "Tahunan"

	tanpaPeriodeID := enums.PeriodeSurveyToInt64(enums.TANPA_PERIODE)
	tanpaPeriodeLabel := "Tanpa Periode"

	allOptions := []response.OptionItem{
		{ID: &triwulanID, Label: &triwulanLabel},
		{ID: &semesterID, Label: &semesterLabel},
		{ID: &tahunanID, Label: &tahunanLabel},
		{ID: &tanpaPeriodeID, Label: &tanpaPeriodeLabel},
	}

	var rawIDs []string
	if len(param["id[]"]) > 0 {
		rawIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		rawIDs = param["id"]
	}

	var filteredOptions []response.OptionItem
	if len(rawIDs) > 0 {
		idMap := make(map[int64]bool)
		for _, id := range rawIDs {
			idInt, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				continue
			}
			idMap[idInt] = true
		}
		for _, opt := range allOptions {
			if idMap[*opt.ID] {
				filteredOptions = append(filteredOptions, opt)
			}
		}
	} else {
		filteredOptions = allOptions
	}

	q := strings.ToLower(param.Get("q"))
	fmt.Println(q)
	var searchedOptions []response.OptionItem
	if q != "" {
		for _, opt := range filteredOptions {
			if strings.Contains(strings.ToLower(*opt.Label), q) || strings.Contains(strconv.FormatInt(*opt.ID, 10), q) {
				searchedOptions = append(searchedOptions, opt)
			}
		}
	} else {
		searchedOptions = filteredOptions
	}

	totalData := int64(len(searchedOptions))
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	var paginatedOptions []response.OptionItem
	if startIndex < len(searchedOptions) {
		if endIndex > len(searchedOptions) {
			endIndex = len(searchedOptions)
		}
		paginatedOptions = searchedOptions[startIndex:endIndex]
	} else {
		paginatedOptions = []response.OptionItem{}
	}

	currentTotalLoaded := startIndex + len(paginatedOptions)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.OptionsResponse{
		Options: paginatedOptions,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi periode survey")
}

func (service *surveyService) CreateSurvey(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.SurveyRequest

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

	if !enums.IsPeriodeSurveyExist(payload.TanggalPelaksanaanSurvey) {
		return utils.SendError(errors.New("Tanggal pelaksanaan survey tidak valid"), http.StatusBadRequest)
	}

	flowDetailData, err := service.manajemenAlurRepo.GetFlowDetailByCode(payload.Alur)
	if err != nil {
		return utils.SendError(errors.New("Alur survey tidak ditemukan"), http.StatusBadRequest)
	}

	if payload.RespondenSurvey == 2 {
		if payload.TingkatPelaksanaan != nil {
			if !enums.IsWilayahExist(enums.WilayahID(*payload.TingkatPelaksanaan)) {
				return utils.SendError(errors.New("Tingkat pelaksanaan survey tidak valid"), http.StatusBadRequest)
			}
		}

		if *payload.TingkatPelaksanaan == int(enums.KECAMATAN) {
			if len(payload.Kecamatan) == 0 {
				return utils.SendError(errors.New("Kecamatan harus diisi untuk tingkat pelaksanaan kecamatan"), http.StatusBadRequest)
			}

			for _, kecamatanId := range payload.Kecamatan {
				_, err = service.wilayahRepo.GetKecamatanById(ctx, kecamatanId)
				if err != nil {
					return utils.SendError(fmt.Errorf("Kecamatan dengan ID %d tidak ditemukan", kecamatanId), http.StatusBadRequest)
				}

			}

		} else if *payload.TingkatPelaksanaan == int(enums.KELURAHAN) {
			if len(payload.Kecamatan) == 0 {
				return utils.SendError(errors.New("Kecamatan harus diisi untuk tingkat pelaksanaan kelurahan"), http.StatusBadRequest)
			}

			for _, kecamatanId := range payload.Kecamatan {
				_, err = service.wilayahRepo.GetKecamatanById(ctx, kecamatanId)
				if err != nil {
					return utils.SendError(fmt.Errorf("Kecamatan dengan ID %d tidak ditemukan", kecamatanId), http.StatusBadRequest)
				}
			}

			if len(payload.Kelurahan) == 0 {
				return utils.SendError(errors.New("Kelurahan harus diisi untuk tingkat pelaksanaan kelurahan"), http.StatusBadRequest)
			}

			for _, kelurahanId := range payload.Kelurahan {
				_, err = service.wilayahRepo.GetKelurahanById(ctx, kelurahanId)
				if err != nil {
					return utils.SendError(fmt.Errorf("Kelurahan dengan ID %d tidak ditemukan", kelurahanId), http.StatusBadRequest)
				}
			}
		} else if *payload.TingkatPelaksanaan == int(enums.RW) {
			if len(payload.Kecamatan) == 0 {
				return utils.SendError(errors.New("Kecamatan harus diisi untuk tingkat pelaksanaan rw"), http.StatusBadRequest)
			}

			for _, kecamatanId := range payload.Kecamatan {
				_, err = service.wilayahRepo.GetKecamatanById(ctx, kecamatanId)
				if err != nil {
					return utils.SendError(fmt.Errorf("Kecamatan dengan ID %d tidak ditemukan", kecamatanId), http.StatusBadRequest)
				}
			}

			if len(payload.Kelurahan) == 0 {
				return utils.SendError(errors.New("Kelurahan harus diisi untuk tingkat pelaksanaan rw"), http.StatusBadRequest)
			}

			for _, kelurahanId := range payload.Kelurahan {
				_, err = service.wilayahRepo.GetKelurahanById(ctx, kelurahanId)
				if err != nil {
					return utils.SendError(fmt.Errorf("Kelurahan dengan ID %d tidak ditemukan", kelurahanId), http.StatusBadRequest)
				}
			}

			if len(payload.RW) == 0 {
				return utils.SendError(errors.New("RW harus diisi untuk tingkat pelaksanaan rw"), http.StatusBadRequest)
			}

			for _, rwId := range payload.RW {
				_, err = service.wilayahRepo.GetRWById(ctx, rwId)
				if err != nil {
					return utils.SendError(fmt.Errorf("RW dengan ID %d tidak ditemukan", rwId), http.StatusBadRequest)
				}
			}

		}
	}

	surveyorMap := make(map[int64]bool)
	for _, v := range payload.Surveyor {
		if surveyorMap[int64(v)] {
			return utils.SendError(errors.New("Surveyor tidak boleh duplikat"), http.StatusBadRequest)
		}
		surveyorMap[int64(v)] = true

		_, err = service.userRepo.GetRespondentById(ctx, v)
		if err != nil {
			return utils.SendError(fmt.Errorf("Surveyor dengan ID %d tidak ditemukan", v), http.StatusBadRequest)
		}
	}

	startDate, err := time.Parse("2006-01-02 15:04:05", payload.TanggalSurveyDimulai)
	if err != nil {
		return utils.SendError(errors.New("Format tanggal tidak valid"), http.StatusBadRequest)
	}

	endDate, err := time.Parse("2006-01-02 15:04:05", payload.TanggalSurveyBerakhir)
	if err != nil {
		return utils.SendError(errors.New("Format tanggal tidak valid"), http.StatusBadRequest)
	}

	now := time.Now()
	status := "upcoming"
	if now.After(startDate) && now.Before(endDate) {
		status = "ongoing"
	} else if now.After(endDate) {
		status = "finished"
	}

	approvalStatus := "non_approval"

	if usr.Role == int(enums.ROLE_KECAMATAN) {
		approvalStatus = string(enums.STATUS_APPROVAL_SURVEY_WAITING)
	}

	typeSurvey := strconv.FormatInt(enums.PeriodeSurveyToInt64(payload.TanggalPelaksanaanSurvey), 10)

	isSurveyExists, err := service.surveyRepo.IsSurveyExistsByNameLower(payload.NamaSurvey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(errors.New("Gagal memeriksa nama survey lama"), http.StatusInternalServerError)
	}
	if isSurveyExists {
		return utils.SendError(errors.New("Survey dengan nama tersebut sudah ada"), http.StatusBadRequest)
	}

	survey := models.Survey{
		Name:           payload.NamaSurvey,
		FlowDetailID:   int64(flowDetailData.ID),
		StartDate:      startDate,
		EndDate:        endDate,
		IsRepeated:     payload.IsRepeated,
		Type:           typeSurvey,
		Status:         status,
		Deskripsi:      payload.Deskripsi,
		CreatedBy:      usr.ID,
		ApprovalSurvey: approvalStatus,
	}

	err = service.surveyRepo.RunInTransaction(func(txRepo repository.SurveyRepo) error {
		createdSurvey, err := txRepo.CreateSurvey(survey)
		if err != nil {
			return err
		}

		if len(payload.Surveyor) > 0 {
			for _, v := range payload.Surveyor {
				_, err := txRepo.AssignSurveyorToSurvey(int64(v), int64(createdSurvey.ID))
				if err != nil {
					return err
				}
			}
		}

		tingkatWilayahStr := strconv.FormatInt(int64(*payload.TingkatPelaksanaan), 10)

		if payload.RespondenSurvey == 2 {
			if *payload.TingkatPelaksanaan == int(enums.KECAMATAN) {
				for _, kecamatanId := range payload.Kecamatan {
					surveyWilayah := models.SurveyWilayah{
						TingkatWilayah: tingkatWilayahStr,
						SurveyId:       int64(createdSurvey.ID),
						KecamatanId:    kecamatanId,
					}

					_, err := txRepo.AssignWilayahToSurvey(surveyWilayah)
					if err != nil {
						return err
					}

				}
			} else if *payload.TingkatPelaksanaan == int(enums.KELURAHAN) {
				for _, kelurahanId := range payload.Kelurahan {
					kelurahanDetail, err := service.wilayahRepo.GetKelurahanById(ctx, kelurahanId)
					if err != nil {
						return err
					}

					surveyWilayah := models.SurveyWilayah{
						TingkatWilayah: tingkatWilayahStr,
						SurveyId:       int64(createdSurvey.ID),
						KelurahanId:    &kelurahanId,
						KecamatanId:    kelurahanDetail.SubDistrictId,
					}

					_, err = txRepo.AssignWilayahToSurvey(surveyWilayah)
					if err != nil {
						return err
					}
				}
			} else if *payload.TingkatPelaksanaan == int(enums.RW) {
				for _, rwId := range payload.RW {
					rwDetail, err := service.wilayahRepo.GetRWById(ctx, rwId)
					if err != nil {
						return err
					}

					kelurahanDetail, err := service.wilayahRepo.GetKelurahanById(ctx, rwDetail.VillageId)
					if err != nil {
						return err
					}

					surveyWilayah := models.SurveyWilayah{
						TingkatWilayah: tingkatWilayahStr,
						SurveyId:       int64(createdSurvey.ID),
						RWId:           &rwId,
						KelurahanId:    &rwDetail.VillageId,
						KecamatanId:    kelurahanDetail.SubDistrictId,
					}

					_, err = txRepo.AssignWilayahToSurvey(surveyWilayah)
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return utils.SendError(errors.New("gagal membuat survey: "+err.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Survey berhasil dibuat!")
}

func (service *surveyService) GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	if usr.Role != int(enums.ROLE_ADMIN) && usr.Role != int(enums.ROLE_KECAMATAN) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk melihat daftar survey ini"), http.StatusForbidden)
	}

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	status := param.Get("status")
	startDateStr := param.Get("start_date")
	endDateStr := param.Get("end_date")

	surveyDiikutiStr := param.Get("survey_diikuti")
	surveyDiikuti := false
	if surveyDiikutiStr == "true" {
		surveyDiikuti = true
	}

	isApprovalStr := param.Get("is_approval")
	isApproval := false
	if isApprovalStr == "true" {
		isApproval = true
	}

	payload := payloads.SurveyDatatablePayload{
		Search:        search,
		Page:          page,
		Limit:         limit,
		OrderBy:       orderBy,
		OrderDir:      orderDir,
		SurveyDiikuti: surveyDiikuti,
		Status:        status,
		IsApproval:    isApproval,
		StartDate:     startDateStr,
		EndDate:       endDateStr,
	}

	var validate = validator.New()
	err := validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusInternalServerError)
	}

	data, totalData, err := service.surveyRepo.GetListSurvey(usr, respondentLogin, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	now := time.Now()

	for i := range data {
		surveyId := []int{int(data[i].ID)}

		surveyCode, err := h.Encode(surveyId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		data[i].SurveyCode = surveyCode

		if data[i].Approval == string(enums.STATUS_APPROVAL_SURVEY_WAITING) {
			if usr.Role == int(enums.ROLE_ADMIN) {
				data[i].PosibleApproval = true
			}
		}

		if data[i].CreatedBy == int(usr.ID) {
			data[i].IsMyOwn = true
		} else {
			data[i].IsMyOwn = false
		}

		surveyDimulai := utils.ParseToWIB(data[i].SurveyDimulai)
		surveyBerakhir := utils.ParseToWIB(data[i].SurveyBerakhir)

		if surveyDimulai.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_UPCOMING)
		} else if surveyDimulai.Before(now) && surveyBerakhir.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_ONGOING)
		} else if surveyBerakhir.Before(now) {
			data[i].Status = string(enums.STATUS_SURVEY_FINISHED)
		}
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(payload.Limit)))

	result := map[string]interface{}{
		"data": data,
		"meta": map[string]interface{}{
			"total":      totalData,
			"page":       payload.Page,
			"limit":      payload.Limit,
			"totalPages": totalPages,
		},
	}

	return utils.SendData(result, "Berhasil mengambil list survey")
}

func (service *surveyService) GetDetailSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr := slug["survey_code"]
	surveyCode, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode survey tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(surveyCode)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := int64(decodedSurveyIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyId)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailById(survey.FlowDetailID)
	if err != nil {
		return utils.SendError(errors.New("Data alur survey tidak ditemukan"), http.StatusInternalServerError)
	}

	wilayahs, _ := service.surveyRepo.GetSurveyWilayahsBySurveyId(surveyId)
	surveyors, _ := service.surveyRepo.GetSurveyorsBySurveyId(surveyId)

	respondenSurvey := 1
	var tingkatPelaksanaan *int

	kecamatanMap := make(map[int64]bool)
	kelurahanMap := make(map[int64]bool)
	rwMap := make(map[int64]bool)

	kecamatanIDs := make([]int64, 0)
	kelurahanIDs := make([]int64, 0)
	rwIDs := make([]int64, 0)

	if len(wilayahs) > 0 {
		respondenSurvey = 2

		if wilayahs[0].TingkatWilayah != "" {
			tingkatInt, _ := strconv.Atoi(wilayahs[0].TingkatWilayah)
			tingkatPelaksanaan = &tingkatInt
		}

		for _, w := range wilayahs {
			if w.KecamatanId > 0 && !kecamatanMap[w.KecamatanId] {
				kecamatanMap[w.KecamatanId] = true
				kecamatanIDs = append(kecamatanIDs, w.KecamatanId)
			}
			if w.KelurahanId != nil && *w.KelurahanId > 0 && !kelurahanMap[*w.KelurahanId] {
				kelurahanMap[*w.KelurahanId] = true
				kelurahanIDs = append(kelurahanIDs, *w.KelurahanId)
			}
			if w.RWId != nil && *w.RWId > 0 && !rwMap[*w.RWId] {
				rwMap[*w.RWId] = true
				rwIDs = append(rwIDs, *w.RWId)
			}
		}
	}

	surveyorIDs := make([]int64, 0)
	for _, s := range surveyors {
		surveyorIDs = append(surveyorIDs, s.RespondentId)
	}

	periodeVal, _ := strconv.Atoi(survey.Type)
	startDateStr := survey.StartDate.Format("2006-01-02 15:04:05")
	endDateStr := survey.EndDate.Format("2006-01-02 15:04:05")

	response := models.SurveyDetailResponse{
		SurveyCode:               surveyCode,
		NamaSurvey:               survey.Name,
		TanggalPelaksanaanSurvey: int32(periodeVal),
		TanggalSurveyDimulai:     startDateStr,
		TanggalSurveyBerakhir:    endDateStr,
		Deskripsi:                survey.Deskripsi,
		Alur:                     flowDetail.Code,
		RespondenSurvey:          respondenSurvey,
		TingkatPelaksanaan:       tingkatPelaksanaan,
		Kecamatan:                kecamatanIDs,
		Kelurahan:                kelurahanIDs,
		RW:                       rwIDs,
		Surveyor:                 surveyorIDs,
		IsRepeated:               survey.IsRepeated,
	}

	return utils.SendData(response, "Detail survey berhasil diambil")
}

func (service *surveyService) ApprovalSurvey(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	if usr.Role != int(enums.ROLE_ADMIN) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusForbidden)
	}

	var payload payloads.ApprovalSurveyRequest

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

	if payload.Action == string(enums.STATUS_APPROVAL_SURVEY_REJECTED) {
		if payload.Notes == nil {
			return utils.SendError(errors.New("Alasan penolakan harus diisi"), http.StatusBadRequest)
		}

		if strings.TrimSpace(*payload.Notes) == "" {
			return utils.SendError(errors.New("Alasan penolakan harus diisi"), http.StatusBadRequest)
		}
	}

	codeStr := slug["code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	decodedIDs, err := h.DecodeWithError(code)
	if err != nil || len(decodedIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := decodedIDs[0]

	surveyData, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	if surveyData.ApprovalSurvey != string(enums.STATUS_APPROVAL_SURVEY_WAITING) {
		return utils.SendError(errors.New("Survey sudah tidak dalam status menunggu approval"), http.StatusBadRequest)
	}

	switch payload.Action {
	case string(enums.STATUS_APPROVAL_SURVEY_APPROVED):
		surveyData.ApprovalSurvey = string(enums.STATUS_APPROVAL_SURVEY_APPROVED)
	case string(enums.STATUS_APPROVAL_SURVEY_REJECTED):
		surveyData.ApprovalSurvey = string(enums.STATUS_APPROVAL_SURVEY_REJECTED)
		surveyData.AlasanReject = payload.Notes
	}

	err = service.surveyRepo.UpdateSurvey(surveyData)
	if err != nil {
		return utils.SendError(errors.New("Gagal mengupdate status approval survey"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Status survey berhasil diubah")
}

func (service *surveyService) AvailableSurveyWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	statusSurvey := param.Get("status_survey")

	payload := payloads.SurveyWilayahDatatablePayload{
		Search:       search,
		Page:         page,
		Limit:        limit,
		OrderBy:      orderBy,
		OrderDir:     orderDir,
		StatusSurvey: statusSurvey,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin.RoleId != nil && *respondentLogin.RoleId != int64(enums.ROLE_RT) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	data, totalData, err := service.surveyRepo.GetListSurveyWilayah(usr, respondentLogin, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var rtId int
	if respondentLogin != nil {
		if respondentLogin.RTId != nil {
			rtId = int(*respondentLogin.RTId)
		}
	}

	for i := range data {
		surveyId := []int{data[i].ID}
		surveyCode, err := h.Encode(surveyId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		wilayahId := []int{int(rtId)}

		kodeWilayah, err := h.Encode(wilayahId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		data[i].SurveyCode = surveyCode

		now := time.Now()

		surveyDimulai := utils.ParseToWIB(data[i].SurveyDimulai)
		surveyBerakhir := utils.ParseToWIB(data[i].SurveyBerakhir)
		data[i].CodeWilayah = &kodeWilayah

		if surveyDimulai.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_UPCOMING)

			data[i].PosibleProcess = false
			data[i].PosibleDetail = false
			data[i].PosibleHistory = false
		} else if surveyDimulai.Before(now) && surveyBerakhir.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_ONGOING)

			data[i].PosibleProcess = true
			data[i].PosibleDetail = true
			data[i].PosibleHistory = true
		} else if surveyBerakhir.Before(now) {
			data[i].Status = string(enums.STATUS_SURVEY_FINISHED)

			data[i].PosibleProcess = false
			data[i].PosibleDetail = true
			data[i].PosibleHistory = true
		}

		if data[i].CreatedBy == int(usr.ID) {
			data[i].IsMyOwn = true
		} else {
			data[i].IsMyOwn = false
		}
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

	return utils.SendData(result, "Survey wilayah berhasil diambil")
}

// func (service *surveyService) PreviewSurveyIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
// 	defer utils.GeneralRecover()

// 	hd := hashids.NewData()
// 	hd.Salt = os.Getenv("HASHID_SALT")
// 	hd.MinLength = 24

// 	h, err := hashids.NewWithData(hd)
// 	if err != nil {
// 		return utils.SendError(err, http.StatusInternalServerError)
// 	}

// 	codeStr := slug["code"]
// 	code, ok := codeStr.(string)
// 	if !ok {
// 		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
// 	}

// 	decodedIDs, err := h.DecodeWithError(code)
// 	if err != nil || len(decodedIDs) == 0 {
// 		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
// 	}

// 	surveyId := decodedIDs[0]

// 	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
// 	if err != nil {
// 		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
// 	}

// 	respondentId := usr.RespondentID
// 	if respondentId == 0 {
// 		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
// 	}

// 	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentId)
// 	if err != nil {
// 		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
// 	}

// 	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
// 	if err != nil {
// 		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
// 	}

// 	if !canRespondentDoSurvey {
// 		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
// 	}

// 	if survey.ApprovalSurvey == string(enums.STATUS_APPROVAL_SURVEY_WAITING) {
// 		return utils.SendError(errors.New("Survey belum dapat diakses"), http.StatusBadRequest)
// 	}

// 	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(survey.FlowDetailID))
// 	if err != nil {
// 		return utils.SendError(errors.New("alur survey tidak ditemukan"), http.StatusNotFound)
// 	}

// 	respondentExistsInSurvey, err := service.surveyRepo.GetRespondentExistsInSurvey(respondentId, int64(survey.ID))
// 	if err != nil {
// 		if !errors.Is(err, gorm.ErrRecordNotFound) {
// 			return utils.SendError(err, http.StatusInternalServerError)
// 		}
// 	}

// 	var surveyRespondentId *int64
// 	var isRevisiRT bool
// 	fmt.Println(isRevisiRT)

// 	if respondentExistsInSurvey != nil {
// 		if *respondentExistsInSurvey.Status == 2 {
// 			if respondentExistsInSurvey.StatusApproval != nil {
// 				if *respondentExistsInSurvey.StatusApproval != string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RT) {
// 					return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
// 				} else {
// 					isRevisiRT = true
// 				}
// 			} else {
// 				return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
// 			}
// 		}
// 		surveyRespondentId = &respondentExistsInSurvey.ID
// 	} else {
// 		if usr.Role == int(enums.ROLE_RT) {
// 			// dataCreateSurveyRespondent := models.SurveyRespondent{
// 			// 	RespondentID: respondentId,
// 			// 	SurveyID:     int64(survey.ID),
// 			// }
// 			// _, err := service.surveyRepo.CreateSurveyRespondent(nil, dataCreateSurveyRespondent)
// 			// if err != nil {
// 			// 	return utils.SendError(err, http.StatusInternalServerError)
// 			// }
// 			fmt.Println("====================")
// 		}
// 	}

// 	statusSectionStr := flowDetail.StatusSection
// 	statusSectionInt, _ := strconv.Atoi(statusSectionStr)
// 	hasSectionBool := utils.IntToBool(statusSectionInt)

// 	var sections []models.FlowPreviewSection
// 	rawSections, err := service.manajemenAlurRepo.GetPreviewSectionByFlowDetailId(flowDetail.ID, statusSectionStr, surveyRespondentId)
// 	if err != nil {
// 		return utils.SendError(err, http.StatusInternalServerError)
// 	}

// 	for _, v := range rawSections {
// 		sectionId := []int{v.SectionId}
// 		sectionCode, err := h.Encode(sectionId)
// 		if err != nil {
// 			return utils.SendError(err, http.StatusInternalServerError)
// 		}

// 		finalCompletedStatus := v.Completed
// 		if isRevisiRT {
// 			finalCompletedStatus = false
// 		}

// 		sections = append(sections, models.FlowPreviewSection{
// 			SectionCode:               &sectionCode,
// 			SectionName:               v.SectionName,
// 			TotalRequiredQuestions:    v.TotalRequiredQuestions,
// 			TotalOptionalQuestions:    v.TotalOptionalQuestions,
// 			AnsweredRequiredQuestions: v.AnsweredRequiredQuestions,
// 			AnsweredOptionalQuestions: v.AnsweredOptionalQuestions,
// 			Completed:                 finalCompletedStatus,
// 		})
// 	}

// 	flowId := []int{flowDetail.ID}
// 	flowCode, err := h.Encode(flowId)
// 	if err != nil {
// 		return utils.SendError(err, http.StatusInternalServerError)
// 	}

// 	data := models.FlowPreview{
// 		FlowCode:          flowCode,
// 		FlowName:          flowDetail.Name,
// 		HasSection:        hasSectionBool,
// 		Sections:          sections,
// 		SurveyName:        survey.Name,
// 		SurveyDescription: &survey.Deskripsi,
// 	}

// 	return utils.SendData(data, "Berhasil mengambil data untuk preview survey")
// }

func (service *surveyService) PreviewSurveyIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr := slug["code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	decodedIDs, err := h.DecodeWithError(code)
	if err != nil || len(decodedIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := decodedIDs[0]

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentId := usr.RespondentID
	if respondentId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	if survey.ApprovalSurvey == string(enums.STATUS_APPROVAL_SURVEY_WAITING) {
		return utils.SendError(errors.New("Survey belum dapat diakses"), http.StatusBadRequest)
	}

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(survey.FlowDetailID))
	if err != nil {
		return utils.SendError(errors.New("alur survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentExistsInSurvey, err := service.surveyRepo.GetRespondentExistsInSurvey(respondentId, int64(survey.ID))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(err, http.StatusInternalServerError)
		}
	}

	var surveyRespondentId *int64

	if respondentExistsInSurvey != nil {
		if *respondentExistsInSurvey.Status == 2 {
			if respondentExistsInSurvey.StatusApproval != nil {
				if *respondentExistsInSurvey.StatusApproval != string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RT) {
					return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
				}
			} else {
				return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
			}
		}
		surveyRespondentId = &respondentExistsInSurvey.ID
	} else {
		if usr.Role == int(enums.ROLE_RT) {
			// dataCreateSurveyRespondent := models.SurveyRespondent{
			// 	RespondentID: respondentId,
			// 	SurveyID:     int64(survey.ID),
			// }
			// _, err := service.surveyRepo.CreateSurveyRespondent(nil, dataCreateSurveyRespondent)
			// if err != nil {
			// 	return utils.SendError(err, http.StatusInternalServerError)
			// }
			fmt.Println("====================")
		}
	}

	statusSectionStr := flowDetail.StatusSection
	statusSectionInt, _ := strconv.Atoi(statusSectionStr)
	hasSectionBool := utils.IntToBool(statusSectionInt)

	var sections []models.FlowPreviewSection
	rawSections, err := service.manajemenAlurRepo.GetPreviewSectionByFlowDetailId(flowDetail.ID, statusSectionStr, surveyRespondentId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for _, v := range rawSections {
		sectionId := []int{v.SectionId}
		sectionCode, err := h.Encode(sectionId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		sections = append(sections, models.FlowPreviewSection{
			SectionCode:               &sectionCode,
			SectionName:               v.SectionName,
			TotalRequiredQuestions:    v.TotalRequiredQuestions,
			TotalOptionalQuestions:    v.TotalOptionalQuestions,
			AnsweredRequiredQuestions: v.AnsweredRequiredQuestions,
			AnsweredOptionalQuestions: v.AnsweredOptionalQuestions,
			Completed:                 v.Completed,
		})
	}

	flowId := []int{flowDetail.ID}
	flowCode, err := h.Encode(flowId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	data := models.FlowPreview{
		FlowCode:          flowCode,
		FlowName:          flowDetail.Name,
		HasSection:        hasSectionBool,
		Sections:          sections,
		SurveyName:        survey.Name,
		SurveyDescription: &survey.Deskripsi,
	}

	return utils.SendData(data, "Berhasil mengambil data untuk preview survey")
}

func (service *surveyService) PreviewSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	hostUserAPI := os.Getenv("USERAPI_HOST") + ":" + os.Getenv("USERAPI_PORT")

	newSlug := map[string]interface{}{"id": strconv.FormatInt(usr.RespondentID, 10)}

	dataBytes, err := utils.HitBackend(ctx, hostUserAPI, "GET", "/respondent/:id", newSlug, nil)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var respondentData models.DetailRespondent
	if err := json.Unmarshal(dataBytes, &respondentData); err != nil {
		return utils.SendError(errors.New("Gagal memparsing data respondent dari UserAPI"), http.StatusInternalServerError)
	}

	codeStr := slug["survey_code"]
	surveyCode, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode survey tidak valid"), http.StatusBadRequest)
	}

	sectionCodeStr := slug["section_code"]
	sectionCode, ok := sectionCodeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Kode bagian survey tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(surveyCode)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := decodedSurveyIDs[0]

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentId := usr.RespondentID
	if respondentId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	if survey.ApprovalSurvey == string(enums.STATUS_APPROVAL_SURVEY_WAITING) {
		return utils.SendError(errors.New("Survey belum dapat diakses"), http.StatusBadRequest)
	}

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(survey.FlowDetailID))
	if err != nil {
		return utils.SendError(errors.New("Alur survey tidak ditemukan"), http.StatusNotFound)
	}

	sectionID := 0
	if sectionCode != "0" && sectionCode != "" {
		hd2 := hashids.NewData()
		hd2.Salt = os.Getenv("HASHID_SALT")
		hd2.MinLength = 24

		h2, err2 := hashids.NewWithData(hd2)
		if err2 != nil {
			return utils.SendError(err2, http.StatusInternalServerError)
		}

		decodedIDs, err2 := h2.DecodeWithError(sectionCode)
		if err2 != nil || len(decodedIDs) == 0 {
			return utils.SendError(errors.New("Kode bagian alur survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
		}

		sectionID = decodedIDs[0]
	}

	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(flowDetail.ID, sectionID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if len(rawNodes) == 0 {
		return utils.SendError(errors.New("Tidak ada pertanyaan pada alur atau bagian ini"), http.StatusNotFound)
	}

	answerMap := make(map[int]string)
	respondentExistsInSurvey, _ := service.surveyRepo.GetRespondentExistsInSurvey(respondentId, int64(survey.ID))

	if respondentExistsInSurvey != nil {
		if *respondentExistsInSurvey.Status == 2 {
			if respondentExistsInSurvey.StatusApproval != nil {
				if *respondentExistsInSurvey.StatusApproval != string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RT) {
					return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
				}
			} else {
				return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
			}
		}
		answers, errAns := service.surveyRepo.GetAnswersByResponseID(respondentExistsInSurvey.ID)
		if errAns == nil {
			for _, ans := range answers {
				if ans.Answer != nil {
					answerMap[ans.FormFieldID] = *ans.Answer
				}
			}
		}
	}

	blueprintNodes := make(map[int]*response.PreviewAlurSurveyStep)
	groupMasterMap := make(map[int]int)

	flowFieldToNodeKeyMap := make(map[int]int)
	flowFieldToOptionMap := make(map[int]int)
	breakdownRawMap := make(map[int]map[int]models.RawNodeData)

	var formFieldIDs []int
	var flowFieldIDs []int

	totalRequired := 0
	totalOptional := 0
	entryNodeId := 0
	var activeSectionName *string

	rJump := "jump-to"
	rLogic := "logic"

	for _, raw := range rawNodes {
		formFieldIDs = append(formFieldIDs, raw.FormFieldId)
		flowFieldIDs = append(flowFieldIDs, raw.FlowFieldId)

		if raw.GroupId != nil && *raw.GroupId != 0 {
			if _, exists := groupMasterMap[*raw.GroupId]; !exists {
				groupMasterMap[*raw.GroupId] = raw.FormFieldId
			}
		}
	}

	resolveTargetID := func(childID int, groupChildID int) *int {
		if groupChildID != 0 {
			if target, ok := groupMasterMap[groupChildID]; ok {
				val := target
				return &val
			}
		}
		if childID != 0 {
			val := childID
			return &val
		}
		return nil
	}

	for i, raw := range rawNodes {
		var nodeKey int
		if raw.GroupId != nil && *raw.GroupId != 0 {
			nodeKey = groupMasterMap[*raw.GroupId]
		} else {
			nodeKey = raw.FormFieldId
		}

		flowFieldToNodeKeyMap[raw.FlowFieldId] = nodeKey

		if raw.Breakdown && raw.FormAnswerFieldId != nil {
			flowFieldToOptionMap[raw.FlowFieldId] = *raw.FormAnswerFieldId
		}

		if i == 0 {
			entryNodeId = nodeKey
			activeSectionName = raw.SectionName
		}

		step, exists := blueprintNodes[nodeKey]
		if !exists {
			stepType := "single"
			if raw.GroupId != nil && *raw.GroupId != 0 {
				stepType = "group"
			}

			step = &response.PreviewAlurSurveyStep{
				StepType:  stepType,
				GroupId:   raw.GroupId,
				GroupName: raw.GroupName,
				Questions: []response.PreviewAlurSurveyQuestionDetail{},
				Routing: response.PreviewAlurSurveyRoutingDetail{
					Logics: []response.PreviewAlurSurveyAdvancedLogicItem{},
				},
			}
			breakdownRawMap[nodeKey] = make(map[int]models.RawNodeData)
		}

		isDuplicate := false
		for _, q := range step.Questions {
			if q.QuestionId == raw.FormFieldId {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			var expectedImageCount *int
			if raw.ImageQuantity != nil && *raw.ImageQuantity != "" {
				count, errParse := strconv.Atoi(*raw.ImageQuantity)
				if errParse == nil {
					expectedImageCount = &count
				}
			}

			var finalAnswer interface{} = nil
			if ansStr, exists := answerMap[raw.FormFieldId]; exists {
				switch raw.Template {
				case "image-template":
					var jsonArray []string
					if errUnm := json.Unmarshal([]byte(ansStr), &jsonArray); errUnm == nil {
						finalAnswer = jsonArray
						for i := range jsonArray {
							jsonArray[i] = os.Getenv("API_GATEWAY_URL") + "/view-survey-image/" + jsonArray[i]
						}
					} else {
						finalAnswer = ansStr
					}
				case "maps":
					var jsonArray []interface{}
					if errUnm := json.Unmarshal([]byte(ansStr), &jsonArray); errUnm == nil {
						finalAnswer = jsonArray
					} else {
						finalAnswer = ansStr
					}
				default:
					finalAnswer = ansStr
				}
			}

			step.Questions = append(step.Questions, response.PreviewAlurSurveyQuestionDetail{
				QuestionId:         raw.FormFieldId,
				Type:               raw.Template,
				Label:              raw.Label,
				IsRequired:         raw.IsRequired,
				ExpectedImageCount: expectedImageCount,
				Options:            []response.PreviewAlurSurveyOptionItem{},
				Answer:             finalAnswer,
			})

			if raw.IsRequired {
				totalRequired++
			} else {
				totalOptional++
			}
		}

		target := resolveTargetID(raw.ChildId, int(raw.GroupChildId))

		if raw.Breakdown {
			step.Routing.IsBreakdown = true
			step.Routing.Rule = nil
			if raw.FormAnswerFieldId != nil {
				breakdownRawMap[nodeKey][*raw.FormAnswerFieldId] = raw
			}
		} else {
			step.Routing.IsBreakdown = false
			step.Routing.TargetQuestionId = target
			step.Routing.IsEnd = (target == nil)
			if raw.IsAdvancedOption {
				step.Routing.Rule = &rLogic
			} else {
				step.Routing.Rule = &rJump
			}
		}

		blueprintNodes[nodeKey] = step
	}

	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(formFieldIDs)
	for _, opt := range options {
		for _, step := range blueprintNodes {
			for i, q := range step.Questions {
				if q.QuestionId == opt.FormFieldId {

					optItem := response.PreviewAlurSurveyOptionItem{
						ID:     opt.ID,
						Label:  opt.Option,
						Logics: []response.PreviewAlurSurveyAdvancedLogicItem{},
					}

					if step.Routing.IsBreakdown {
						if rawOpt, ok := breakdownRawMap[q.QuestionId][opt.ID]; ok {
							optTarget := resolveTargetID(rawOpt.ChildId, int(rawOpt.GroupChildId))
							optItem.TargetQuestionId = optTarget
							optItem.IsEnd = (optTarget == nil)

							if rawOpt.IsAdvancedOption {
								optItem.Rule = &rLogic
							} else {
								optItem.Rule = &rJump
							}
						} else {
							optItem.Rule = &rJump
							optItem.IsEnd = true
						}
					}

					step.Questions[i].Options = append(step.Questions[i].Options, optItem)
				}
			}
		}
	}

	logics, _ := service.manajemenAlurRepo.GetAdvancedOptionsByFieldIDs(flowFieldIDs)
	for _, logic := range logics {
		nodeKey, valid := flowFieldToNodeKeyMap[logic.FlowFieldId]

		if valid {
			if step, exists := blueprintNodes[nodeKey]; exists {
				var logicTargetPtr *int
				if logic.ChildId != 0 {
					val := logic.ChildId
					logicTargetPtr = &val
				}

				logicItem := response.PreviewAlurSurveyAdvancedLogicItem{
					IfQuestionId:     logic.FormFieldId,
					IfOptionId:       logic.Option,
					TargetQuestionId: logicTargetPtr,
					IsEnd:            logic.ChildId == 0,
				}

				if step.Routing.IsBreakdown {
					optID, hasOptMap := flowFieldToOptionMap[logic.FlowFieldId]
					if hasOptMap {
						for i, q := range step.Questions {
							for j, o := range q.Options {
								if o.ID == optID {
									step.Questions[i].Options[j].Logics = append(step.Questions[i].Options[j].Logics, logicItem)
								}
							}
						}
					}
				} else {
					step.Routing.Logics = append(step.Routing.Logics, logicItem)
				}
			}
		}
	}

	var openingMeta, closingMeta *string

	if flowDetail.OpeningId != 0 {
		if rawOpening, errOp := service.templateUcapanRepo.GetTemplateUcapanById(flowDetail.OpeningId); errOp == nil && rawOpening != nil {
			openingStr := utils.ReplaceStringRespondentVariable(rawOpening.Content, &respondentData)
			openingMeta = &openingStr
		}
	}

	if flowDetail.ClosingId != 0 {
		if rawClosing, errCl := service.templateUcapanRepo.GetTemplateUcapanById(flowDetail.ClosingId); errCl == nil && rawClosing != nil {
			closingStr := utils.ReplaceStringRespondentVariable(rawClosing.Content, &respondentData)
			closingMeta = &closingStr
		}
	}

	hasSectionBool := utils.StringToBool(flowDetail.StatusSection)

	var activeSecCode *string
	if hasSectionBool && sectionID != 0 {
		activeSecCode = &sectionCode
	}

	finalNodes := make(map[int]response.PreviewAlurSurveyStep)
	for k, v := range blueprintNodes {
		finalNodes[k] = *v
	}

	dataResponse := response.PreviewAlurSurveyBlueprintResponse{
		SurveyInfo: response.PreviewAlurSurveyInfo{
			Name:              flowDetail.Name,
			HasSection:        hasSectionBool,
			ActiveSectionCode: activeSecCode,
			ActiveSectionName: activeSectionName,
			TotalRequired:     totalRequired,
			TotalOptional:     totalOptional,
			EntryNodeId:       entryNodeId,
		},
		Opening: openingMeta,
		Closing: closingMeta,
		Nodes:   finalNodes,
	}

	return utils.SendData(dataResponse, "Berhasil memuat Blueprint Preview Survey")
}

func (service *surveyService) SubmitSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	detachedCtx := context.WithoutCancel(ctx)
	var payload payloads.SubmitSurveyPayload

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	err = json.Unmarshal(jsonBytes, &payload)
	if err != nil {
		if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
			// Mendeteksi jika frontend mengirim string ke field number
			if strings.Contains(jsonErr.Field, "value_number") || strings.Contains(jsonErr.Field, "value_option_id") {
				return utils.SendError(fmt.Errorf("Field '%s' harus berupa angka (number), tidak boleh string", jsonErr.Field), http.StatusBadRequest)
			}
		}
		return utils.SendError(fmt.Errorf("Format payload tidak valid: %v", err), http.StatusBadRequest)
	}

	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr := slug["survey_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(code)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	codeSectionStr := slug["section_code"]
	codeSection, ok := codeSectionStr.(string)
	if !ok {
		return utils.SendError(errors.New("Section tidak valid"), http.StatusBadRequest)
	}

	decodedSectionIDs, err := h.DecodeWithError(codeSection)
	if err != nil || len(decodedSectionIDs) == 0 {
		return utils.SendError(errors.New("Kode Section tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	sectionID := int64(decodedSectionIDs[0])

	surveyDetail, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	respondentId := usr.RespondentID
	if respondentId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, surveyID, respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_RT) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(int(surveyDetail.FlowDetailID), int(sectionID))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if len(rawNodes) == 0 {
		return utils.SendError(errors.New("Tidak ada pertanyaan pada bagian ini"), http.StatusNotFound)
	}

	var sectionFieldIDs []int64
	imageFieldsMap := make(map[int64]bool)

	requiredQuestionsMap := make(map[int64]string)
	answeredRequiredTracker := make(map[int64]bool)

	for _, node := range rawNodes {
		fieldID := int64(node.FormFieldId)
		sectionFieldIDs = append(sectionFieldIDs, fieldID)

		if node.Template == "image-template" {
			imageFieldsMap[fieldID] = true
		}

		if node.IsRequired {
			requiredQuestionsMap[fieldID] = node.Label
			answeredRequiredTracker[fieldID] = false
		}
	}

	tx := service.surveyRepo.BeginTransaction()
	if tx.Error != nil {
		return utils.SendError(errors.New("Gagal memulai transaksi database"), http.StatusInternalServerError)
	}

	var successfullyUploadedFiles []string
	txCommitted := false

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			service.fileRepo.DeleteSurveyImageBulk(detachedCtx, successfullyUploadedFiles)
			panic(r)
		} else if !txCommitted {
			tx.Rollback()
			service.fileRepo.DeleteSurveyImageBulk(detachedCtx, successfullyUploadedFiles)
		}
	}()

	respondentExistsInSurvey, err := service.surveyRepo.GetRespondentExistsInSurvey(respondentId, surveyID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if respondentExistsInSurvey == nil {
		dataCreateSurveyRespondent := models.SurveyRespondent{
			RespondentID: respondentId,
			SurveyID:     surveyID,
		}
		createSurveyRespondent, err := service.surveyRepo.CreateSurveyRespondent(tx, dataCreateSurveyRespondent)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		respondentExistsInSurvey = createSurveyRespondent
	} else {
		if *respondentExistsInSurvey.Status == 2 {
			if respondentExistsInSurvey.StatusApproval != nil {
				if *respondentExistsInSurvey.StatusApproval != string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RT) {
					return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
				}
			} else {
				return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
			}
		}
	}

	oldFilesTracker := make(map[string]bool)
	if len(sectionFieldIDs) > 0 {
		oldResponses, errFindOld := service.surveyRepo.GetOldResponsesBySection(tx, respondentExistsInSurvey.ID, sectionFieldIDs)
		if errFindOld == nil {
			for _, resp := range oldResponses {
				if resp.Answer != nil && imageFieldsMap[int64(resp.FormFieldID)] {
					var oldPaths []string
					if errUnm := json.Unmarshal([]byte(*resp.Answer), &oldPaths); errUnm == nil {
						for _, path := range oldPaths {
							oldFilesTracker[path] = true
						}
					}
				}
			}
		}

		errDelOld := service.surveyRepo.DeleteOldResponsesBySection(tx, respondentExistsInSurvey.ID, sectionFieldIDs)
		if errDelOld != nil {
			return utils.SendError(errors.New("Gagal membersihkan riwayat jawaban lama pada bagian ini"), http.StatusInternalServerError)
		}
	}

	var answersToInsert []models.FieldResponse
	gatewayURL := os.Getenv("API_GATEWAY_URL") + "/view-survey-image/"

	for _, ans := range payload.Answers {
		if ans.ValueString != nil && *ans.ValueString == "[SKIPPED_BY_LOGIC]" {
			if _, isRequired := answeredRequiredTracker[int64(ans.QuestionID)]; isRequired {
				answeredRequiredTracker[int64(ans.QuestionID)] = true
			}

			skippedAnswer := "[SKIPPED_BY_LOGIC]"
			var groupID int64 = 0
			if ans.GroupID != nil {
				groupID = int64(*ans.GroupID)
			}

			dbAnswer := models.FieldResponse{
				FormFieldID:    ans.QuestionID,
				FormResponseID: respondentExistsInSurvey.ID,
				GroupID:        groupID,
				Answer:         &skippedAnswer,
			}
			answersToInsert = append(answersToInsert, dbAnswer)

			continue
		}

		isWrongPayload := false
		expectedField := ""

		switch ans.Type {
		case "short-answer", "long-answer", "email", "date", "time", "phone_number":
			if ans.ValueString == nil && (ans.ValueNumber != nil || ans.ValueOptionID != nil || len(ans.ValueMaps) > 0 || len(ans.ValueImages) > 0) {
				isWrongPayload = true
				expectedField = "value_string"
			}
		case "number":
			if ans.ValueNumber == nil && (ans.ValueString != nil || ans.ValueOptionID != nil || len(ans.ValueMaps) > 0 || len(ans.ValueImages) > 0) {
				isWrongPayload = true
				expectedField = "value_number"
			}
		case "multiple-choices", "dropdown", "checkboxes":
			if ans.ValueOptionID == nil && (ans.ValueString != nil || ans.ValueNumber != nil || len(ans.ValueMaps) > 0 || len(ans.ValueImages) > 0) {
				isWrongPayload = true
				expectedField = "value_option_id"
			}
		case "maps":
			if len(ans.ValueMaps) == 0 && (ans.ValueString != nil || ans.ValueNumber != nil || ans.ValueOptionID != nil || len(ans.ValueImages) > 0) {
				isWrongPayload = true
				expectedField = "value_maps"
			}
		case "image-template":
			if len(ans.ValueImages) == 0 && (ans.ValueString != nil || ans.ValueNumber != nil || ans.ValueOptionID != nil || len(ans.ValueMaps) > 0) {
				isWrongPayload = true
				expectedField = "value_images"
			}
		}

		if isWrongPayload {
			return utils.SendError(fmt.Errorf("Pertanyaan ID %d bertipe '%s' salah format, seharusnya mengirimkan '%s'", ans.QuestionID, ans.Type, expectedField), http.StatusBadRequest)
		}

		var answerText string
		var groupID int64 = 0
		if ans.GroupID != nil {
			groupID = int64(*ans.GroupID)
		}

		switch ans.Type {
		case "long-answer", "short-answer":
			if ans.ValueString != nil && strings.TrimSpace(*ans.ValueString) != "" {
				answerText = *ans.ValueString
			}

		case "number":
			if ans.ValueNumber != nil {
				answerText = strconv.FormatInt(*ans.ValueNumber, 10)
			}

		case "multiple-choices":
			if ans.ValueOptionID != nil && *ans.ValueOptionID != 0 {
				answerText = strconv.Itoa(*ans.ValueOptionID)
			}

		case "maps":
			if len(ans.ValueMaps) > 0 {
				mapBytes, _ := json.Marshal(ans.ValueMaps)
				answerText = string(mapBytes)
			}

		case "image-template":
			if len(ans.ValueImages) > 0 {
				formFieldDetail, err := service.templateFormulirPertanyaanRepo.GetFormFieldByID(ans.QuestionID)
				if err != nil || formFieldDetail == nil {
					return utils.SendError(errors.New("Pertanyaan tidak ditemukan atau terjadi kesalahan server"), http.StatusInternalServerError)
				}

				var imageQuantity int
				if formFieldDetail.ImageQuantity != nil && *formFieldDetail.ImageQuantity != "" {
					imageQuantity, err = strconv.Atoi(*formFieldDetail.ImageQuantity)
					if err != nil {
						return utils.SendError(errors.New("Format kuantitas gambar tidak valid"), http.StatusBadRequest)
					}
				}

				if len(ans.ValueImages) != imageQuantity {
					return utils.SendError(errors.New("Jumlah foto tidak sesuai ketentuan"), http.StatusBadRequest)
				}

				availableMime := []string{"image/png", "image/jpg", "image/jpeg", "image/webp"}
				availablesExt := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif"}
				maxSizeInKB := float64(5120)
				var uploadedPaths []string

				for _, image := range ans.ValueImages {
					if strings.HasPrefix(image, "data:") {
						if !strings.HasPrefix(image, "data:image") {
							return utils.SendError(errors.New("Format file tidak didukung. Hanya menerima file gambar (PNG, JPG, WEBP)"), http.StatusBadRequest)
						}

						base64Data, err := utils.ExtractBase64Info(image)
						if err != nil {
							return utils.SendError(errors.New("Gagal memproses gambar"), http.StatusBadRequest)
						}

						if !slices.Contains(availablesExt, base64Data.Extension) || !slices.Contains(availableMime, base64Data.MimeType) {
							return utils.SendError(errors.New("Format file gambar tidak didukung"), http.StatusBadRequest)
						}

						if base64Data.SizeInKB > maxSizeInKB {
							return utils.SendError(errors.New("Ukuran gambar tidak boleh melebihi 5MB"), http.StatusBadRequest)
						}

						path, err := service.fileRepo.UploadSurveyImage(ctx, &image)
						if err != nil {
							return utils.SendError(errors.New("Gagal mengunggah gambar: "+err.Error()), http.StatusInternalServerError)
						}

						if path != nil {
							uploadedPaths = append(uploadedPaths, *path)
							successfullyUploadedFiles = append(successfullyUploadedFiles, *path)
						}

					} else {
						cleanPath := image
						if after, ok0 := strings.CutPrefix(cleanPath, gatewayURL); ok0 {
							cleanPath = after
						}

						ext := strings.ToLower(filepath.Ext(cleanPath))
						if !slices.Contains(availablesExt, ext) {
							return utils.SendError(errors.New("Terdapat format file/path yang tidak valid. Hanya gambar yang diperbolehkan."), http.StatusBadRequest)
						}

						uploadedPaths = append(uploadedPaths, cleanPath)

						if oldFilesTracker[cleanPath] {
							oldFilesTracker[cleanPath] = false
						}
					}
				}

				if len(uploadedPaths) > 0 {
					pathBytes, _ := json.Marshal(uploadedPaths)
					answerText = string(pathBytes)
				}
			}
		}

		if strings.TrimSpace(answerText) == "" {
			continue
		}

		if _, isRequired := answeredRequiredTracker[int64(ans.QuestionID)]; isRequired {
			answeredRequiredTracker[int64(ans.QuestionID)] = true
		}

		dbAnswer := models.FieldResponse{
			FormFieldID:    ans.QuestionID,
			FormResponseID: respondentExistsInSurvey.ID,
			GroupID:        groupID,
		}
		ansCopy := answerText
		dbAnswer.Answer = &ansCopy

		answersToInsert = append(answersToInsert, dbAnswer)
	}

	var missingLabels []string
	for qID, label := range requiredQuestionsMap {
		if !answeredRequiredTracker[qID] {
			missingLabels = append(missingLabels, label)
		}
	}

	if len(missingLabels) > 0 {
		errMsg := fmt.Sprintf("Ada pertanyaan wajib yang belum dijawab: %s", strings.Join(missingLabels, ", "))
		return utils.SendError(errors.New(errMsg), http.StatusBadRequest)
	}

	if len(answersToInsert) > 0 {
		err = service.surveyRepo.BulkInsertFieldResponses(ctx, tx, answersToInsert)
		if err != nil {
			return utils.SendError(errors.New("Gagal menyimpan jawaban survei, terjadi kesalahan server"), http.StatusInternalServerError)
		}
	}

	if respondentExistsInSurvey.StatusApproval != nil &&
		*respondentExistsInSurvey.StatusApproval == string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RT) {
		if errResolve := service.surveyRepo.ResolveFlaggingBySection(ctx, tx, respondentExistsInSurvey.ID, sectionFieldIDs); errResolve != nil {
			return utils.SendError(errors.New("Gagal memperbarui status revisi pertanyaan pada bagian ini"), http.StatusInternalServerError)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(errors.New("Gagal memfinalisasi data"), http.StatusInternalServerError)
	}
	txCommitted = true

	var finalFilesToCleanup []string
	for path, shouldDelete := range oldFilesTracker {
		if shouldDelete {
			finalFilesToCleanup = append(finalFilesToCleanup, path)
		}
	}

	if len(finalFilesToCleanup) > 0 {
		go func(paths []string) {
			service.fileRepo.DeleteSurveyImageBulk(detachedCtx, paths)
		}(finalFilesToCleanup)
	}

	return utils.SendData(nil, "Jawaban survei pada bagian ini berhasil disimpan")
}

func (service *surveyService) UpdateSurveyRespondentStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateSurveyStatusPayload

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		for _, valErr := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(valErr)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr, ok := slug["survey_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Survey code tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(codeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	respondentId := usr.RespondentID
	if respondentId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, surveyID, respondentLogin)
	if err != nil || !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_RT) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	tx := service.surveyRepo.BeginTransaction()
	if tx.Error != nil {
		return utils.SendError(errors.New("Gagal memulai transaksi database"), http.StatusInternalServerError)
	}

	txCommitted := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if !txCommitted {
			tx.Rollback()
		}
	}()

	respondentExistsInSurvey, err := service.surveyRepo.GetRespondentExistsInSurvey(respondentId, surveyID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if respondentExistsInSurvey == nil {
		return utils.SendError(errors.New("Anda tidak dapat mengubah status data, ikuti survey terlebih dahulu"), http.StatusBadRequest)
	}

	targetStatus := *payload.Status
	currentStatus := 0
	var statusApproval *string
	if respondentExistsInSurvey.Status != nil {
		currentStatus = *respondentExistsInSurvey.Status
	}

	if respondentExistsInSurvey.StatusApproval != nil {
		statusApproval = respondentExistsInSurvey.StatusApproval
	}

	if currentStatus == 2 {
		if statusApproval != nil {
			if *statusApproval != string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RT) {
				return utils.SendError(errors.New("Survei ini sudah diselesaikan dan datanya telah terkunci"), http.StatusBadRequest)
			}
		} else {
			return utils.SendError(errors.New("Survei ini sudah diselesaikan dan datanya telah terkunci"), http.StatusBadRequest)
		}

		if targetStatus == 0 || targetStatus == 1 {
			return utils.SendError(errors.New("Survey ini sudah diselesaikan dan datanya telah terkunci"), http.StatusBadRequest)
		}
	}

	if targetStatus == 2 {
		surveyDetail, err := service.surveyRepo.GetSurveyById(surveyID)
		if err != nil {
			return utils.SendError(errors.New("Gagal mendapatkan detail survei"), http.StatusInternalServerError)
		}

		flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(surveyDetail.FlowDetailID))
		if err != nil {
			return utils.SendError(errors.New("Gagal mendapatkan detail alur survei"), http.StatusInternalServerError)
		}

		rawSections, err := service.manajemenAlurRepo.GetPreviewSectionByFlowDetailId(flowDetail.ID, flowDetail.StatusSection, &respondentExistsInSurvey.ID)
		if err != nil {
			return utils.SendError(errors.New("Gagal mengecek progres jawaban survei"), http.StatusInternalServerError)
		}

		for _, section := range rawSections {
			if !section.Completed {
				errMsg := fmt.Sprintf("Anda tidak dapat menyelesaikan survei. %s belum selesai dikerjakan.", *section.SectionName)
				return utils.SendError(errors.New(errMsg), http.StatusBadRequest)
			}
		}

		var nullApproval *string = nil

		errResetInduk := service.surveyRepo.UpdateRespondentApprovalState(ctx, tx, respondentExistsInSurvey.ID, 2, nullApproval, "false")
		if errResetInduk != nil {
			return utils.SendError(errors.New("Gagal mengunci dan mereset status persetujuan survei"), http.StatusInternalServerError)
		}

		errResetFlagging := service.surveyRepo.ResetFlaggingEditStatusTx(ctx, tx, respondentExistsInSurvey.ID)
		if errResetFlagging != nil {
			return utils.SendError(errors.New("Gagal mereset catatan revisi pertanyaan"), http.StatusInternalServerError)
		}
	} else {
		err = service.surveyRepo.UpdateStatusSurvey(ctx, tx, respondentExistsInSurvey.ID, targetStatus)
		if err != nil {
			return utils.SendError(errors.New("Gagal memperbarui status survei"), http.StatusInternalServerError)
		}
	}

	if targetStatus == 2 {
		var roleName string
		if respondentLogin.RoleId != nil {
			roleName = string(enums.RoleID(*respondentLogin.RoleId).Label())
		}

		var keterangan *string

		if respondentExistsInSurvey.StatusApproval != nil {
			if *respondentExistsInSurvey.StatusApproval == string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RT) {
				keterangan = utils.ToPtr(respondentLogin.Name + " (" + roleName + ") telah menyelesaikan revisi.")
			} else {
				keterangan = utils.ToPtr(respondentLogin.Name + " (" + roleName + ") telah menyelesaikan survey.")
			}
		} else {
			keterangan = utils.ToPtr(respondentLogin.Name + " (" + roleName + ") telah menyelesaikan survey.")
		}

		logLevel := strconv.Itoa(int(enums.ROLE_RW))

		var respondentRwId int64

		if respondentLogin.RWId != nil {
			respondentRwId = *respondentLogin.RWId
		}

		respondentRW, err := service.userRepo.GetRespondentByRWId(ctx, respondentRwId)
		if err != nil {
			return utils.SendError(errors.New("Gagal mendapatkan data respondent RW"), http.StatusInternalServerError)
		}

		var notifFor *string
		if respondentRW != nil {
			notifFor = utils.ToPtr(strconv.Itoa(int(respondentRW.ID)))
		}
		isRead := "false"
		logSurvey := models.LogSurvey{
			RespondentID: respondentId,
			SurveyID:     surveyID,
			Keterangan:   keterangan,
			LogLevel:     &logLevel,
			NotifFor:     notifFor,
			IsRead:       &isRead,
		}

		errLog := service.surveyRepo.MakeLogSurvey(ctx, tx, logSurvey)
		if errLog != nil {
			return utils.SendError(errors.New("Gagal mencatat log aktivitas survey"), http.StatusInternalServerError)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return utils.SendError(errors.New("Gagal memfinalisasi data"), http.StatusInternalServerError)
	}
	txCommitted = true

	msg := "Status pengerjaan berhasil diperbarui"
	if targetStatus == 2 {
		msg = "Survei berhasil diselesaikan dan dikirimkan untuk diperiksa!"
	}

	return utils.SendData(nil, msg)
}

func (service *surveyService) GetApprovalHistorySurveyPerWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr := slug["survey_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	decodedIDs, err := h.DecodeWithError(code)
	if err != nil || len(decodedIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := decodedIDs[0]

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	codeWilayahStr := slug["code_wilayah"]
	codeWilayah, ok := codeWilayahStr.(string)
	if !ok {
		return utils.SendError(errors.New("Code wilayah tidak valid"), http.StatusBadRequest)
	}

	decodedWilayahIDs, err := h.DecodeWithError(codeWilayah)
	if err != nil || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Code wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	rtId := decodedWilayahIDs[0]

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, int64(survey.ID), int64(rtId))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Gagal mengambil data responden survey untuk wilayah ini"), http.StatusInternalServerError)
		}
	}

	if surveyRespondent == nil {
		return utils.SendData(nil, "Belum ada responden yang mengisi survey untuk wilayah ini")
	}

	dataLogs, err := service.surveyRepo.GetHistoryApprovalPerWilayah(ctx, nil, int64(survey.ID), int64(rtId))
	if err != nil {
		return utils.SendError(errors.New("Gagal mengambil riwayat persetujuan survey"), http.StatusInternalServerError)
	}

	return utils.SendData(dataLogs, "Riwayat persetujuan survey berhasil diambil")
}

func (service *surveyService) GetHistoryDetailPerWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr := slug["survey_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	decodedIDs, err := h.DecodeWithError(code)
	if err != nil || len(decodedIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := decodedIDs[0]

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	codeWilayahStr := slug["code_wilayah"]
	codeWilayah, ok := codeWilayahStr.(string)
	if !ok {
		return utils.SendError(errors.New("Code wilayah tidak valid"), http.StatusBadRequest)
	}

	decodedWilayahIDs, err := h.DecodeWithError(codeWilayah)
	if err != nil || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Code wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	rtId := decodedWilayahIDs[0]

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, int64(survey.ID), int64(rtId))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Gagal mengambil data responden survey untuk wilayah ini"), http.StatusInternalServerError)
		}
	}

	if surveyRespondent == nil {
		return utils.SendData(nil, "Belum ada responden yang mengisi survey untuk wilayah ini")
	}

	completeHistory, err := service.surveyRepo.GetSurveyCompletionHistory(ctx, int64(survey.ID), int64(surveyRespondent.RespondentID))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			fmt.Println("==1===============================")
			spew.Dump(err)
			return utils.SendError(errors.New("Gagal mengambil riwayat penyelesaian survey"), http.StatusInternalServerError)
		}
	}

	verificationHistory, err := service.surveyRepo.GetSurveyVerificationHistory(ctx, int64(survey.ID), int64(surveyRespondent.RespondentID))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			fmt.Println("==2===============================")
			spew.Dump(err)
			return utils.SendError(errors.New("Gagal mengambil riwayat penyelesaian survey"), http.StatusInternalServerError)
		}
	}

	validationHistory, err := service.surveyRepo.GetSurveyValidationHistory(ctx, int64(survey.ID), int64(surveyRespondent.RespondentID))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			fmt.Println("==3===============================")
			spew.Dump(err)
			return utils.SendError(errors.New("Gagal mengambil riwayat penyelesaian survey"), http.StatusInternalServerError)
		}
	}

	// 1. Siapkan variabel dengan nilai default "-"
	var completeDate, verificationDate, validationDate interface{} = "-", "-", "-"

	// 2. Lakukan pengecekan apakah history tidak nil
	// Sesuaikan kondisi ini dengan kebutuhan Anda (misal jika CreatedAt berupa pointer, bisa ditambahkan pengecekan)
	if completeHistory != nil {
		completeDate = completeHistory.CreatedAt
	}
	if verificationHistory != nil {
		verificationDate = verificationHistory.CreatedAt
	}
	if validationHistory != nil {
		validationDate = validationHistory.CreatedAt
	}

	// 3. Masukkan ke dalam map
	mergeHistory := []map[string]interface{}{
		{
			"no":         1,
			"created_at": completeDate,
			"keterangan": "Sudah Melakukan Pengisian Survey",
		},
		{
			"no":         2,
			"created_at": verificationDate,
			"keterangan": "Sudah Diverifikasi RW",
		},
		{
			"no":         3,
			"created_at": validationDate,
			"keterangan": "Sudah Divalidasi Kelurahan",
		},
	}

	return utils.SendData(mergeHistory, "Riwayat penyelesaian survey berhasil diambil")
}

func (service *surveyService) GetSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	statusSurvey := param.Get("status_survey")

	payload := payloads.SurveyWilayahDatatablePayload{
		Search:       search,
		Page:         page,
		Limit:        limit,
		OrderBy:      orderBy,
		OrderDir:     orderDir,
		StatusSurvey: statusSurvey,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, usr.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	data, totalData, err := service.surveyRepo.GetSurveyKewilayahan(usr, respondentLogin, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for i := range data {
		surveyId := []int{data[i].ID}
		surveyCode, err := h.Encode(surveyId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		data[i].SurveyCode = surveyCode

		now := time.Now()

		surveyDimulai := utils.ParseToWIB(data[i].SurveyDimulai)
		surveyBerakhir := utils.ParseToWIB(data[i].SurveyBerakhir)

		if surveyDimulai.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_UPCOMING)
		} else if surveyDimulai.Before(now) && surveyBerakhir.After(now) {
			data[i].Status = string(enums.STATUS_SURVEY_ONGOING)
		} else if surveyBerakhir.Before(now) {
			data[i].Status = string(enums.STATUS_SURVEY_FINISHED)
		}
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

	return utils.SendData(result, "Survey wilayah berhasil diambil")
}

func (service *surveyService) GetDetailSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	tipe_wilayah, _ := strconv.Atoi(param.Get("tipe_wilayah"))

	kecamatan_idStr := param.Get("kecamatan_code")
	kelurahan_idStr := param.Get("kelurahan_code")
	rw_idStr := param.Get("rw_code")

	var kecamatan_id int64
	var kelurahan_id int64
	var rw_id int64

	if kecamatan_idStr != "" {
		decodedKecamatanIds, err := h.DecodeWithError(kecamatan_idStr)
		if err != nil {
			return utils.SendError(errors.New("Kecamatan tidak valid"), http.StatusBadRequest)
		}

		kecamatan_id = int64(decodedKecamatanIds[0])
	}

	if kelurahan_idStr != "" {
		decodedKelurahanIds, err := h.DecodeWithError(kelurahan_idStr)
		if err != nil {
			fmt.Println("Error decoding kelurahan_idStr:", err)
			return utils.SendError(errors.New("Kelurahan tidak valid"), http.StatusBadRequest)
		}

		kelurahan_id = int64(decodedKelurahanIds[0])
	}

	if rw_idStr != "" {
		decodedRWIds, err := h.DecodeWithError(rw_idStr)
		if err != nil {
			return utils.SendError(errors.New("RW tidak valid"), http.StatusBadRequest)
		}

		rw_id = int64(decodedRWIds[0])
	}

	payload := payloads.DetailSurveyKewilayahanDatatablePayload{
		Search:      search,
		Page:        page,
		Limit:       limit,
		OrderBy:     orderBy,
		OrderDir:    orderDir,
		TipeWilayah: &tipe_wilayah,
		KecamatanId: &kecamatan_id,
		KelurahanId: &kelurahan_id,
		RWId:        &rw_id,
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	respondentId := usr.RespondentID
	if respondentId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	codeStr := slug["survey_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	decodedIDs, err := h.DecodeWithError(code)
	if err != nil || len(decodedIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := decodedIDs[0]

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	var metaTotal int
	var metaPage int
	var metaLimit int
	var metaTotalPages int

	var finalData []response.DetailSurveyKewilayahanResponse

	if payload.TipeWilayah != nil {
		switch *payload.TipeWilayah {
		case 0:
			return utils.SendError(errors.New("Tipe wilayah harus diisi"), http.StatusBadRequest)
		case 5:
			if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
			}

			data, err := service.wilayahRepo.GetDaftarKecamatan(ctx, payloads.DatatablePayload{
				Search:   payload.Search,
				Page:     payload.Page,
				Limit:    payload.Limit,
				OrderBy:  payload.OrderBy,
				OrderDir: payload.OrderDir,
			})
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}

			metaTotal = data.Meta.Total
			metaPage = data.Meta.Page
			metaLimit = data.Meta.Limit
			metaTotalPages = data.Meta.TotalPages

			if len(data.Data) > 0 {
				var extractedKecamatanIds []int64
				for _, kecamatan := range data.Data {
					extractedKecamatanIds = append(extractedKecamatanIds, kecamatan.ID)
				}

				statusMap, _ := service.surveyRepo.GetStatusKeterisianBulkKecamatan(ctx, int64(survey.ID), extractedKecamatanIds, int64(usr.Role))

				for _, kecamatan := range data.Data {
					wilayahId := []int{int(kecamatan.ID)}

					kodeWilayah, err := h.Encode(wilayahId)
					if err != nil {
						return utils.SendError(err, http.StatusInternalServerError)
					}
					newKecamatan := response.DetailSurveyKewilayahanResponse{
						No:              kecamatan.No,
						ID:              kecamatan.ID,
						NamaWilayah:     kecamatan.SubDistrictName,
						IsPosibleDetail: true,
						Code:            kodeWilayah,
					}

					if usr.Role == int(enums.ROLE_ADMIN) {
						newKecamatan.IsPosibleBackAccess = true
					}

					if statusStr, exists := statusMap[kecamatan.ID]; exists {
						val := statusStr
						newKecamatan.Status = val
					}

					finalData = append(finalData, newKecamatan)
				}
			}
		case 4:
			if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
			}

			var KecamatanId int64

			if *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) {
				if *payload.KecamatanId == 0 {
					return utils.SendError(errors.New("Kecamatan Code harus diisi untuk tipe wilayah Kecamatan"), http.StatusBadRequest)
				}
				KecamatanId = *payload.KecamatanId
			} else {
				KecamatanId = *respondentLogin.KecamatanId
			}

			data, err := service.wilayahRepo.GetDaftarKelurahan(ctx, KecamatanId, payloads.DatatablePayload{
				Search:   payload.Search,
				Page:     payload.Page,
				Limit:    payload.Limit,
				OrderBy:  payload.OrderBy,
				OrderDir: payload.OrderDir,
			})
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}

			metaTotal = data.Meta.Total
			metaPage = data.Meta.Page
			metaLimit = data.Meta.Limit
			metaTotalPages = data.Meta.TotalPages

			if len(data.Data) > 0 {
				var extractedKelurahanIds []int64
				for _, kelurahan := range data.Data {
					extractedKelurahanIds = append(extractedKelurahanIds, kelurahan.ID)
				}

				statusMap, _ := service.surveyRepo.GetStatusKeterisianBulkKelurahan(ctx, int64(survey.ID), extractedKelurahanIds, int64(usr.Role))

				for _, kelurahan := range data.Data {
					wilayahId := []int{int(kelurahan.ID)}

					kodeWilayah, err := h.Encode(wilayahId)
					if err != nil {
						return utils.SendError(err, http.StatusInternalServerError)
					}
					newKelurahan := response.DetailSurveyKewilayahanResponse{
						No:              kelurahan.No,
						ID:              kelurahan.ID,
						NamaWilayah:     kelurahan.VillageName,
						IsPosibleDetail: true,
						Code:            kodeWilayah,
					}

					if usr.Role == int(enums.ROLE_ADMIN) {
						newKelurahan.IsPosibleBackAccess = true
					}

					if statusStr, exists := statusMap[kelurahan.ID]; exists {
						val := statusStr
						newKelurahan.Status = val
					}

					finalData = append(finalData, newKelurahan)
				}
			}
		case 3:
			if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
			}

			var kelurahanId int64

			if *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) {
				if *payload.KelurahanId == 0 {
					return utils.SendError(errors.New("Kelurahan Code harus diisi untuk tipe wilayah Kelurahan"), http.StatusBadRequest)
				}
				kelurahanId = *payload.KelurahanId
			} else {
				kelurahanId = *respondentLogin.KelurahanId
			}

			data, err := service.wilayahRepo.GetDaftarRW(ctx, kelurahanId, payloads.DatatablePayload{
				Search:   payload.Search,
				Page:     payload.Page,
				Limit:    payload.Limit,
				OrderBy:  payload.OrderBy,
				OrderDir: payload.OrderDir,
			})
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}

			metaTotal = data.Meta.Total
			metaPage = data.Meta.Page
			metaLimit = data.Meta.Limit
			metaTotalPages = data.Meta.TotalPages

			if len(data.Data) > 0 {
				var extractedRWIds []int64
				for _, rw := range data.Data {
					extractedRWIds = append(extractedRWIds, rw.ID)
				}

				statusMap, _ := service.surveyRepo.GetStatusKeterisianBulkRW(ctx, int64(survey.ID), extractedRWIds)

				for _, rw := range data.Data {
					wilayahId := []int{int(rw.ID)}

					kodeWilayah, err := h.Encode(wilayahId)
					if err != nil {
						return utils.SendError(err, http.StatusInternalServerError)
					}
					newRw := response.DetailSurveyKewilayahanResponse{
						No:              rw.No,
						ID:              rw.ID,
						NamaWilayah:     rw.NamaRw,
						IsPosibleDetail: true,
						Code:            kodeWilayah,
					}

					if usr.Role == int(enums.ROLE_ADMIN) {
						newRw.IsPosibleBackAccess = true
					}

					if statusStr, exists := statusMap[rw.ID]; exists {
						val := statusStr
						newRw.Status = val
					}

					finalData = append(finalData, newRw)
				}
			}
		case 2:
			if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) && *respondentLogin.RoleId != int64(enums.ROLE_RW) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
			}

			var rwId int64

			if *respondentLogin.RoleId != int64(enums.ROLE_RW) {
				if *payload.RWId == 0 {
					return utils.SendError(errors.New("RW Code harus diisi untuk tipe wilayah RW"), http.StatusBadRequest)
				}
				rwId = *payload.RWId
			} else {
				rwId = *respondentLogin.RWId
			}

			data, err := service.wilayahRepo.GetDaftarRT(ctx, rwId, payloads.DatatablePayload{
				Search:   payload.Search,
				Page:     payload.Page,
				Limit:    payload.Limit,
				OrderBy:  payload.OrderBy,
				OrderDir: payload.OrderDir,
			})
			if err != nil {
				return utils.SendError(err, http.StatusInternalServerError)
			}

			metaTotal = data.Meta.Total
			metaPage = data.Meta.Page
			metaLimit = data.Meta.Limit
			metaTotalPages = data.Meta.TotalPages

			if len(data.Data) > 0 {
				var extractedRTIds []int64
				for _, rt := range data.Data {
					extractedRTIds = append(extractedRTIds, rt.ID)
				}

				statusMap, _ := service.surveyRepo.GetStatusKeterisianBulkRT(ctx, int64(survey.ID), extractedRTIds)

				isDoneMap, _ := service.surveyRepo.GetSurveyIsDoneBulkRT(ctx, int64(survey.ID), extractedRTIds)

				for _, rt := range data.Data {
					wilayahId := []int{int(rt.ID)}

					kodeWilayah, err := h.Encode(wilayahId)
					if err != nil {
						return utils.SendError(err, http.StatusInternalServerError)
					}
					newRt := response.DetailSurveyKewilayahanResponse{
						No:          rt.No,
						ID:          rt.ID,
						NamaWilayah: rt.NamaRt,
						Code:        kodeWilayah,
					}

					if usr.Role == int(enums.ROLE_ADMIN) {
						newRt.IsPosibleBackAccess = true
					}

					if statusStr, exists := statusMap[rt.ID]; exists {
						val := statusStr
						if val != "Tidak ada responden" {
							newRt.IsPosibleDetail = true
							newRt.IsPosiblePreviewSurvey = true
						}
						newRt.Status = val
					}

					if isDone, exists := isDoneMap[rt.ID]; exists {
						newRt.SurveyIsDone = isDone
					} else {
						newRt.SurveyIsDone = false // Default jika tidak ada
					}

					finalData = append(finalData, newRt)
				}
			}
		default:
			return utils.SendError(errors.New("Tipe wilayah tidak valid"), http.StatusBadRequest)
		}
	}

	result := map[string]interface{}{
		"data": finalData,
		"meta": map[string]interface{}{
			"total":      metaTotal,
			"page":       metaPage,
			"limit":      metaLimit,
			"totalPages": metaTotalPages,
		},
	}

	return utils.SendData(result, "Survey wilayah berhasil diambil")
}

func (service *surveyService) SurveyResultIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr := slug["survey_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	decodedIDs, err := h.DecodeWithError(code)
	if err != nil || len(decodedIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := decodedIDs[0]

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) && *respondentLogin.RoleId != int64(enums.ROLE_RW) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk melihat hasil survey ini"), http.StatusUnauthorized)
	}

	codeWilayahStr := slug["code"]
	codeWilayah, ok := codeWilayahStr.(string)
	if !ok {
		return utils.SendError(errors.New("Code wilayah tidak valid"), http.StatusBadRequest)
	}

	decodedWilayahIDs, err := h.DecodeWithError(codeWilayah)
	if err != nil || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Code wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	rtId := decodedWilayahIDs[0]

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, int64(survey.ID), int64(rtId))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Gagal mengambil data responden survey untuk wilayah ini"), http.StatusInternalServerError)
		}
	}

	if surveyRespondent == nil {
		return utils.SendData(nil, "Belum ada responden yang mengisi survey untuk wilayah ini")
	}

	respondent, err := service.userRepo.GetRespondentDetailById(ctx, surveyRespondent.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Responden tidak ditemukan"), http.StatusInternalServerError)
	}

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(survey.FlowDetailID))
	if err != nil {
		return utils.SendError(errors.New("Alur survey tidak ditemukan"), http.StatusNotFound)
	}

	var sections []response.SectionDataIndex
	surveyRespondentId := &surveyRespondent.ID

	rawSections, err := service.manajemenAlurRepo.GetPreviewSectionByFlowDetailId(flowDetail.ID, flowDetail.StatusSection, surveyRespondentId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	for _, v := range rawSections {
		sectionId := []int{v.SectionId}

		sectionCode, err := h.Encode(sectionId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		sectionName := ""
		if v.SectionName != nil {
			sectionName = *v.SectionName
		}

		sections = append(sections, response.SectionDataIndex{
			SectionCode: sectionCode,
			SectionName: sectionName,
		})
	}

	respondentId := []int{int(respondent.ID)}

	respondentCode, err := h.Encode(respondentId)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var statusApproval *string
	if surveyRespondent.StatusApproval != nil {
		if *surveyRespondent.StatusApproval == string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RT) {
			val := "Sedang Proses Revisi Oleh RT"
			statusApproval = &val
		} else if *surveyRespondent.StatusApproval == string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_REVISI_RW) {
			val := "Sedang Proses Revisi Oleh RW"
			statusApproval = &val
		} else if *surveyRespondent.StatusApproval == string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_VALIDATED_LURAH) {
			val := "Sudah Divalidasi Oleh Kelurahan"
			statusApproval = &val
		} else if *surveyRespondent.StatusApproval == string(enums.STATUS_APPROVAL_SURVEY_RESPONDENT_VERIFIED_RW) {
			val := "Sudah Diverifikasi Oleh RW"
			statusApproval = &val
		}
	}

	dataResult := response.SurveyResultIndex{
		Respondent: response.RespondentDataIndex{
			Code:      respondentCode,
			Name:      respondent.Name,
			Kecamatan: respondent.Kecamatan,
			Kelurahan: respondent.Kelurahan,
			RW:        respondent.RW,
			RT:        respondent.RT,
			Status:    statusApproval,
		},
		Sections: sections,
	}

	return utils.SendData(dataResult, "Hasil survey berhasil diambil")
}

func (service *surveyService) SurveyResultSectionDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	surveyCode, ok := slug["survey_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Kode survey tidak valid"), http.StatusBadRequest)
	}
	decodedSurveyIDs, err := h.DecodeWithError(surveyCode)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyId := decodedSurveyIDs[0]

	codeWilayah, ok := slug["code"].(string)
	if !ok {
		return utils.SendError(errors.New("Code wilayah tidak valid"), http.StatusBadRequest)
	}
	decodedWilayahIDs, err := h.DecodeWithError(codeWilayah)
	if err != nil || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Code wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	rtId := decodedWilayahIDs[0]

	sectionCode, ok := slug["section_code"].(string)
	if !ok {
		return utils.SendError(errors.New("Kode section tidak valid"), http.StatusBadRequest)
	}
	decodedSectionIDs, err := h.DecodeWithError(sectionCode)
	if err != nil || len(decodedSectionIDs) == 0 {
		return utils.SendError(errors.New("Kode section tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	sectionID := decodedSectionIDs[0]

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) && *respondentLogin.RoleId != int64(enums.ROLE_RW) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk melihat hasil survey ini"), http.StatusUnauthorized)
	}

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, int64(survey.ID), int64(rtId))
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			return utils.SendError(errors.New("Gagal mengambil data responden survey untuk wilayah ini"), http.StatusInternalServerError)
		}
	}
	if surveyRespondent == nil {
		return utils.SendError(errors.New("Belum ada jawaban dari responden untuk wilayah ini"), http.StatusNotFound)
	}

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(survey.FlowDetailID))
	if err != nil {
		return utils.SendError(errors.New("Alur survey tidak ditemukan"), http.StatusNotFound)
	}

	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(flowDetail.ID, sectionID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if len(rawNodes) == 0 {
		return utils.SendError(errors.New("Tidak ada pertanyaan pada bagian ini"), http.StatusNotFound)
	}

	answerMap := make(map[int]string)
	answers, errAns := service.surveyRepo.GetAnswersByResponseID(surveyRespondent.ID)
	if errAns == nil {
		for _, ans := range answers {
			if ans.Answer != nil {
				answerMap[ans.FormFieldID] = *ans.Answer
			}
		}
	}

	var formFieldIDs []int
	tempIDTracker := make(map[int]bool)

	for _, raw := range rawNodes {
		if !tempIDTracker[raw.FormFieldId] {
			formFieldIDs = append(formFieldIDs, raw.FormFieldId)
			tempIDTracker[raw.FormFieldId] = true
		}
	}

	optionsMap := make(map[int][]response.SurveyResultOptionItem)
	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(formFieldIDs)
	for _, opt := range options {
		optionsMap[opt.FormFieldId] = append(optionsMap[opt.FormFieldId], response.SurveyResultOptionItem{
			ID:    opt.ID,
			Label: opt.Option,
		})
	}

	var resultNodes []response.SurveyResultNode
	nodeIndexMap := make(map[int]int)
	groupMasterMap := make(map[int]int)

	processedQuestions := make(map[int]bool)

	for _, raw := range rawNodes {
		if raw.GroupId != nil && *raw.GroupId != 0 {
			if _, exists := groupMasterMap[*raw.GroupId]; !exists {
				groupMasterMap[*raw.GroupId] = raw.FormFieldId
			}
		}
	}

	var activeSectionName string

	for i, raw := range rawNodes {
		if i == 0 && raw.SectionName != nil {
			activeSectionName = *raw.SectionName
		}

		if processedQuestions[raw.FormFieldId] {
			continue
		}
		processedQuestions[raw.FormFieldId] = true

		var nodeKey int
		stepType := "single"
		if raw.GroupId != nil && *raw.GroupId != 0 {
			nodeKey = groupMasterMap[*raw.GroupId]
			stepType = "group"
		} else {
			nodeKey = raw.FormFieldId
		}

		var finalAnswer interface{} = nil
		if ansStr, exists := answerMap[raw.FormFieldId]; exists {
			if ansStr == "[SKIPPED_BY_LOGIC]" {
				finalAnswer = "[SKIPPED_BY_LOGIC]"
			} else {
				switch raw.Template {
				case "image-template":
					var jsonArray []string
					if errUnm := json.Unmarshal([]byte(ansStr), &jsonArray); errUnm == nil {
						for imgIdx := range jsonArray {
							jsonArray[imgIdx] = os.Getenv("API_GATEWAY_URL") + "/view-survey-image/" + jsonArray[imgIdx]
							// jsonArray[imgIdx] = os.Getenv("API_GATEWAY_URL") + "/survey/show-image/" + surveyCode + "/" + codeWilayah + "/" + jsonArray[imgIdx]
						}
						finalAnswer = jsonArray
					} else {
						finalAnswer = ansStr
					}
				case "maps", "checkboxes":
					var jsonArray []interface{}
					if errUnm := json.Unmarshal([]byte(ansStr), &jsonArray); errUnm == nil {
						finalAnswer = jsonArray
					} else {
						finalAnswer = ansStr
					}
				default:
					finalAnswer = ansStr
				}
			}
		}

		nodeOptions := optionsMap[raw.FormFieldId]
		if nodeOptions == nil {
			nodeOptions = []response.SurveyResultOptionItem{}
		}

		questionDetail := response.SurveyResultQuestionDetail{
			QuestionID: raw.FormFieldId,
			Label:      raw.Label,
			Type:       raw.Template,
			IsRequired: raw.IsRequired,
			Options:    nodeOptions,
			Answer:     finalAnswer,
		}

		if idx, exists := nodeIndexMap[nodeKey]; exists {
			resultNodes[idx].Questions = append(resultNodes[idx].Questions, questionDetail)
		} else {
			newNode := response.SurveyResultNode{
				GroupID:   raw.GroupId,
				GroupName: raw.GroupName,
				StepType:  stepType,
				Questions: []response.SurveyResultQuestionDetail{questionDetail},
			}
			resultNodes = append(resultNodes, newNode)
			nodeIndexMap[nodeKey] = len(resultNodes) - 1
		}
	}

	finalResponse := response.SurveyResultSectionDetail{
		SectionName: activeSectionName,
		Nodes:       resultNodes,
	}

	return utils.SendData(finalResponse, "Berhasil memuat detail jawaban survey")
}

func (service *surveyService) VerifySurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	detachedCtx := context.WithoutCancel(ctx)
	var payload payloads.VerifyAllSurveyAnswersPayload

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	err = json.Unmarshal(jsonBytes, &payload)
	if err != nil {
		if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
			if strings.Contains(jsonErr.Field, "value_number") || strings.Contains(jsonErr.Field, "value_option_id") {
				return utils.SendError(fmt.Errorf("Field '%s' harus berupa angka (number), tidak boleh string", jsonErr.Field), http.StatusBadRequest)
			}
		}
		return utils.SendError(fmt.Errorf("Format payload tidak valid: %v", err), http.StatusBadRequest)
	}

	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, valErr := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(valErr)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr, ok1 := slug["survey_code"].(string)
	codeWilayahStr, ok2 := slug["code"].(string)
	if !ok1 || !ok2 {
		return utils.SendError(errors.New("Parameter URL tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(codeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	decodedWilayahIDs, err := h.DecodeWithError(codeWilayahStr)
	if err != nil || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Code wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	rtId := int64(decodedWilayahIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_RW) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, int64(survey.ID), rtId)
	if err != nil || surveyRespondent == nil {
		return utils.SendError(errors.New("Data isian responden tidak ditemukan"), http.StatusNotFound)
	}

	if surveyRespondent.Status == nil || *surveyRespondent.Status != 2 {
		return utils.SendError(errors.New("Responden belum menyelesaikan pengisian survey ini"), http.StatusBadRequest)
	}
	if surveyRespondent.StatusApproval != nil && *surveyRespondent.StatusApproval != "" && *surveyRespondent.StatusApproval != "revisi_rw" {
		return utils.SendError(errors.New("Survey ini sudah diverifikasi sebelumnya"), http.StatusBadRequest)
	}

	respondentSurveyDetail, err := service.userRepo.GetRespondentById(ctx, surveyRespondent.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	if respondentSurveyDetail == nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	respondentKelurahan, err := service.userRepo.GetRespondentByKelurahanId(ctx, *respondentSurveyDetail.KelurahanId)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent Kecamatan"), http.StatusInternalServerError)
	}

	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(int(survey.FlowDetailID), 0)
	if err != nil || len(rawNodes) == 0 {
		return utils.SendError(errors.New("Gagal memuat master pertanyaan untuk survey ini"), http.StatusInternalServerError)
	}

	validQuestionsMap := make(map[int]models.RawNodeData)
	var blueprintFieldIDs []int
	for _, node := range rawNodes {
		validQuestionsMap[node.FormFieldId] = node
		blueprintFieldIDs = append(blueprintFieldIDs, node.FormFieldId)
	}

	validOptionsMap := make(map[int]map[int]bool)
	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(blueprintFieldIDs)
	for _, opt := range options {
		if validOptionsMap[opt.FormFieldId] == nil {
			validOptionsMap[opt.FormFieldId] = make(map[int]bool)
		}
		validOptionsMap[opt.FormFieldId][opt.ID] = true
	}

	for _, ans := range payload.Answers {
		node, exists := validQuestionsMap[ans.QuestionID]
		if !exists {
			return utils.SendError(fmt.Errorf("Pertanyaan ID %d bukan bagian dari survey ini", ans.QuestionID), http.StatusBadRequest)
		}

		if ans.ValueString != nil && *ans.ValueString == "[SKIPPED_BY_LOGIC]" {
			continue
		}

		if node.Template != ans.Type {
			return utils.SendError(fmt.Errorf("Tipe jawaban tidak cocok untuk pertanyaan ID %d", ans.QuestionID), http.StatusBadRequest)
		}

		isWrongPayload := false
		expectedField := ""

		switch ans.Type {
		case "long-answer":
			if ans.ValueString == nil && (ans.ValueNumber != nil || ans.ValueOptionID != nil || len(ans.ValueMaps) > 0 || len(ans.ValueImages) > 0) {
				isWrongPayload = true
				expectedField = "value_string"
			}
		case "number":
			if ans.ValueNumber == nil && (ans.ValueString != nil || ans.ValueOptionID != nil || len(ans.ValueMaps) > 0 || len(ans.ValueImages) > 0) {
				isWrongPayload = true
				expectedField = "value_number"
			}
		case "multiple-choices":
			if ans.ValueOptionID == nil && (ans.ValueString != nil || ans.ValueNumber != nil || len(ans.ValueMaps) > 0 || len(ans.ValueImages) > 0) {
				isWrongPayload = true
				expectedField = "value_option_id"
			}
		case "maps":
			if len(ans.ValueMaps) == 0 && (ans.ValueString != nil || ans.ValueNumber != nil || ans.ValueOptionID != nil || len(ans.ValueImages) > 0) {
				isWrongPayload = true
				expectedField = "value_maps"
			}
		case "image-template":
			if len(ans.ValueImages) == 0 && (ans.ValueString != nil || ans.ValueNumber != nil || ans.ValueOptionID != nil || len(ans.ValueMaps) > 0) {
				isWrongPayload = true
				expectedField = "value_images"
			}
		}

		if isWrongPayload {
			return utils.SendError(fmt.Errorf("Pertanyaan ID %d bertipe '%s' salah format, seharusnya mengirimkan '%s'", ans.QuestionID, ans.Type, expectedField), http.StatusBadRequest)
		}

		switch ans.Type {
		case "multiple-choices":
			if ans.ValueOptionID != nil && *ans.ValueOptionID != 0 {
				if !validOptionsMap[ans.QuestionID][*ans.ValueOptionID] {
					return utils.SendError(fmt.Errorf("Pilihan jawaban tidak sah untuk pertanyaan ID %d", ans.QuestionID), http.StatusBadRequest)
				}
			}
		}
	}

	imageFieldsMap := make(map[int64]bool)
	for _, node := range rawNodes {
		if node.Template == "image-template" {
			imageFieldsMap[int64(node.FormFieldId)] = true
		}
	}

	oldFilesTracker := make(map[string]bool)
	oldResponses, errFindOld := service.surveyRepo.GetAllOldResponsesByRespondent(ctx, int64(surveyRespondent.ID))
	if errFindOld == nil {
		for _, resp := range oldResponses {
			if resp.Answer != nil && imageFieldsMap[int64(resp.FormFieldID)] {
				var oldPaths []string
				if errUnm := json.Unmarshal([]byte(*resp.Answer), &oldPaths); errUnm == nil {
					for _, path := range oldPaths {
						oldFilesTracker[path] = true
					}
				}
			}
		}
	}

	var successfullyUploadedFiles []string
	txCommitted := false

	defer func() {
		if r := recover(); r != nil {
			if len(successfullyUploadedFiles) > 0 {
				_, _ = service.fileRepo.DeleteSurveyImageBulk(detachedCtx, successfullyUploadedFiles)
			}
			panic(r)
		} else if !txCommitted {
			if len(successfullyUploadedFiles) > 0 {
				_, _ = service.fileRepo.DeleteSurveyImageBulk(detachedCtx, successfullyUploadedFiles)
			}
		}
	}()

	err = service.surveyRepo.RunInTransaction(func(txRepo repository.SurveyRepo) error {

		var finalResponses []models.FieldResponse
		gatewayURL := os.Getenv("API_GATEWAY_URL") + "/view-survey-image/"
		availableMime := []string{"image/png", "image/jpg", "image/jpeg", "image/webp"}
		availablesExt := []string{".png", ".jpg", ".jpeg", ".webp", ".jfif"}
		maxSizeInKB := float64(5120)

		for _, ans := range payload.Answers {
			var finalAnswerStr *string

			if ans.ValueString != nil && *ans.ValueString == "[SKIPPED_BY_LOGIC]" {
				finalAnswerStr = ans.ValueString
			} else {
				switch ans.Type {
				case "long-answer":
					if ans.ValueString != nil {
						ansCopy := *ans.ValueString
						finalAnswerStr = &ansCopy
					}
				case "number":
					if ans.ValueNumber != nil {
						strVal := strconv.Itoa(*ans.ValueNumber)
						finalAnswerStr = &strVal
					}
				case "multiple-choices":
					if ans.ValueOptionID != nil {
						strVal := strconv.Itoa(*ans.ValueOptionID)
						finalAnswerStr = &strVal
					}
				case "maps":
					if len(ans.ValueMaps) > 0 {
						jsonBytes, _ := json.Marshal(ans.ValueMaps)
						strVal := string(jsonBytes)
						finalAnswerStr = &strVal
					}
				case "image-template":
					if len(ans.ValueImages) > 0 {
						var uploadedPaths []string
						for _, image := range ans.ValueImages {

							// 3. HARDENING IMAGE VALIDATION & MAGIC BYTES DETECTION
							if strings.HasPrefix(image, "data:") {
								if !strings.HasPrefix(image, "data:image") {
									return errors.New("Format file tidak didukung. Hanya menerima file gambar (PNG, JPG, WEBP)")
								}

								base64Data, errExtract := utils.ExtractBase64Info(image)
								if errExtract != nil {
									return errors.New("Gagal memproses gambar")
								}

								if !slices.Contains(availablesExt, base64Data.Extension) || !slices.Contains(availableMime, base64Data.MimeType) {
									return errors.New("Format file gambar tidak didukung")
								}

								if base64Data.SizeInKB > maxSizeInKB {
									return errors.New("Ukuran gambar tidak boleh melebihi 5MB")
								}

								imgToUpload := image
								path, errUp := service.fileRepo.UploadSurveyImage(ctx, &imgToUpload)

								if errUp != nil || path == nil {
									return errors.New("Gagal mengunggah gambar ke server dokumen")
								}
								uploadedPaths = append(uploadedPaths, *path)
								successfullyUploadedFiles = append(successfullyUploadedFiles, *path)

							} else {
								// Asumsi URL lama / path string biasa
								cleanPath := image
								if after, ok0 := strings.CutPrefix(cleanPath, gatewayURL); ok0 {
									cleanPath = after
								}

								// VALIDASI EKSTENSI (Cegah injeksi path non-gambar)
								ext := strings.ToLower(filepath.Ext(cleanPath))
								if !slices.Contains(availablesExt, ext) {
									return errors.New("Terdapat format file/path yang tidak valid. Hanya gambar yang diperbolehkan.")
								}

								uploadedPaths = append(uploadedPaths, cleanPath)
								if oldFilesTracker[cleanPath] {
									oldFilesTracker[cleanPath] = false
								}
							}
						}
						pathBytes, _ := json.Marshal(uploadedPaths)
						ansCopy := string(pathBytes)
						finalAnswerStr = &ansCopy
					}
				}
			}

			if finalAnswerStr != nil {
				var groupID int64 = 0
				if ans.GroupID != nil {
					groupID = int64(*ans.GroupID)
				}

				finalResponses = append(finalResponses, models.FieldResponse{
					FormResponseID: int64(surveyRespondent.ID),
					FormFieldID:    ans.QuestionID,
					GroupID:        groupID,
					Answer:         finalAnswerStr,
				})
			}
		}

		if errWipe := txRepo.WipeAndReplaceAllFieldResponses(ctx, int64(surveyRespondent.ID), finalResponses); errWipe != nil {
			return errWipe
		}

		if errResetFlags := txRepo.ResetFlaggingEditStatus(ctx, int64(surveyRespondent.ID)); errResetFlags != nil {
			return errResetFlags
		}

		err = txRepo.UpdateRespondentApprovalStatus(ctx, int64(surveyRespondent.ID), 2, "verified_rw", "false")
		if err != nil {
			return errors.New("Gagal memperbarui status approval responden")
		}

		var roleNameUserLogin string
		if respondentLogin.RoleId != nil {
			roleNameUserLogin = string(enums.RoleID(*respondentLogin.RoleId).Label())
		}

		var roleNameRespondentSurvey string
		if respondentSurveyDetail.RoleId != nil {
			roleNameRespondentSurvey = string(enums.RoleID(*respondentSurveyDetail.RoleId).Label())
		}

		logSurveys := []models.LogSurvey{
			{
				RespondentID: surveyRespondent.RespondentID,
				SurveyID:     surveyID,
				Keterangan:   utils.ToPtr(respondentLogin.Name + " (" + roleNameUserLogin + ") telah melakukan verifikasi " + survey.Name + ", dengan nama responden " + respondentSurveyDetail.Name + " (" + roleNameRespondentSurvey + ")."),
				LogLevel:     utils.ToPtr(strconv.Itoa(int(enums.ROLE_KELURAHAN))),
				NotifFor:     utils.ToPtr(strconv.Itoa(int(respondentKelurahan.ID))),
				IsRead:       utils.ToPtr("false"),
			},
		}

		errLog := txRepo.MakeLogSurveyBulk(ctx, nil, logSurveys)
		if errLog != nil {
			return errors.New("Gagal mencatat log aktivitas survey")
		}

		return nil
	})

	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	txCommitted = true

	var finalFilesToCleanup []string
	for path, shouldDelete := range oldFilesTracker {
		if shouldDelete {
			finalFilesToCleanup = append(finalFilesToCleanup, path)
		}
	}
	if len(finalFilesToCleanup) > 0 {
		go func(paths []string) {
			_, _ = service.fileRepo.DeleteSurveyImageBulk(detachedCtx, paths)
		}(finalFilesToCleanup)
	}

	return utils.SendData(nil, "Hasil pengisian survey berhasil diverifikasi oleh RW")
}

func (service *surveyService) RejectSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.RejectAllSurveyAnswersPayload

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	err = json.Unmarshal(jsonBytes, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	var validate = validator.New()
	err = validate.Struct(payload)
	if err != nil {
		for _, valErr := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(valErr)), http.StatusBadRequest)
		}
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, _ := hashids.NewWithData(hd)

	codeStr, ok1 := slug["survey_code"].(string)
	codeWilayahStr, ok2 := slug["code"].(string)
	if !ok1 || !ok2 {
		return utils.SendError(errors.New("Parameter URL tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, _ := h.DecodeWithError(codeStr)
	decodedWilayahIDs, _ := h.DecodeWithError(codeWilayahStr)
	if len(decodedSurveyIDs) == 0 || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Kode URL tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyID := int64(decodedSurveyIDs[0])
	rtId := int64(decodedWilayahIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyID))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) && *respondentLogin.RoleId != int64(enums.ROLE_RW) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, surveyID, rtId)
	if err != nil || surveyRespondent == nil {
		return utils.SendError(errors.New("Data isian responden tidak ditemukan"), http.StatusNotFound)
	}

	if surveyRespondent.Status == nil || *surveyRespondent.Status != 2 {
		return utils.SendError(errors.New("Responden belum menyelesaikan pengisian survey ini"), http.StatusBadRequest)
	}
	if surveyRespondent.StatusApproval != nil && *surveyRespondent.StatusApproval != "" {
		return utils.SendError(errors.New("Survey ini sudah diverifikasi sebelumnya"), http.StatusBadRequest)
	}

	surveyRespondentDetail, err := service.userRepo.GetRespondentById(ctx, surveyRespondent.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	errTx := service.surveyRepo.RunInTransaction(func(txRepo repository.SurveyRepo) error {
		var flags []models.FlaggingEditPertanyaanSurvey

		for _, qID := range payload.FlaggedQuestionIDs {
			flags = append(flags, models.FlaggingEditPertanyaanSurvey{
				SurveyRespondentID: int64(surveyRespondent.ID),
				FormFieldID:        int64(qID),
				IsRevisied:         "true",
			})
		}

		if errFlag := txRepo.ClearAndCreateFlaggingEdit(ctx, surveyRespondent.ID, flags); errFlag != nil {
			return errFlag
		}

		err := txRepo.UpdateRespondentApprovalStatus(ctx, int64(surveyRespondent.ID), 2, "revisi_rt", "true")
		if err != nil {
			return err
		}

		var roleNameUserLogin string
		if respondentLogin.RoleId != nil {
			roleNameUserLogin = string(enums.RoleID(*respondentLogin.RoleId).Label())
		}

		var roleNameRespondentSurvey string
		if surveyRespondentDetail.RoleId != nil {
			roleNameRespondentSurvey = string(enums.RoleID(*surveyRespondentDetail.RoleId).Label())
		}

		logSurveys := []models.LogSurvey{
			{
				RespondentID: surveyRespondent.RespondentID,
				SurveyID:     surveyID,
				Keterangan:   utils.ToPtr(respondentLogin.Name + " (" + roleNameUserLogin + ") melakukan penolakan pada survey " + survey.Name + ", dengan nama responden " + surveyRespondentDetail.Name + " (" + roleNameRespondentSurvey + ")."),
				LogLevel:     utils.ToPtr(strconv.Itoa(int(enums.ROLE_RT))),
				NotifFor:     utils.ToPtr(strconv.Itoa(int(surveyRespondentDetail.ID))),
				IsRead:       utils.ToPtr("false"),
			},
			{
				RespondentID: surveyRespondent.RespondentID,
				SurveyID:     surveyID,
				Keterangan:   utils.ToPtr(respondentLogin.Name + " (" + roleNameUserLogin + ") meminta revisi terhadap " + survey.Name + ", dengan nama responden " + surveyRespondentDetail.Name + " (" + roleNameRespondentSurvey + ")."),
				LogLevel:     utils.ToPtr(strconv.Itoa(int(enums.ROLE_RT))),
				NotifFor:     utils.ToPtr(strconv.Itoa(int(surveyRespondentDetail.ID))),
				IsRead:       utils.ToPtr("false"),
			},
		}

		errLog := txRepo.MakeLogSurveyBulk(ctx, nil, logSurveys)
		if errLog != nil {
			return errors.New("Gagal mencatat log aktivitas survey")
		}

		return nil
	})

	if errTx != nil {
		return utils.SendError(errors.New("Gagal memproses penolakan berkas survey"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Survey berhasil ditolak dan dikembalikan ke responden untuk revisi")
}

func (service *surveyService) RejectValidateSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, _ := hashids.NewWithData(hd)

	codeStr, ok1 := slug["survey_code"].(string)
	codeWilayahStr, ok2 := slug["code"].(string)
	if !ok1 || !ok2 {
		return utils.SendError(errors.New("Parameter URL tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, _ := h.DecodeWithError(codeStr)
	decodedWilayahIDs, _ := h.DecodeWithError(codeWilayahStr)
	if len(decodedSurveyIDs) == 0 || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Kode URL tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyID := int64(decodedSurveyIDs[0])
	rtId := int64(decodedWilayahIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyID))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, surveyID, rtId)
	if err != nil || surveyRespondent == nil {
		return utils.SendError(errors.New("Data isian responden tidak ditemukan"), http.StatusNotFound)
	}

	if surveyRespondent.Status == nil || *surveyRespondent.Status != 2 {
		return utils.SendError(errors.New("Responden belum menyelesaikan pengisian survey ini"), http.StatusBadRequest)
	}

	if surveyRespondent.StatusApproval != nil && *surveyRespondent.StatusApproval != "verified_rw" {
		return utils.SendError(errors.New("Survey ini sudah diverifikasi sebelumnya"), http.StatusBadRequest)
	}

	surveyRespondentDetail, err := service.userRepo.GetRespondentById(ctx, surveyRespondent.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	respondentRW, err := service.userRepo.GetRespondentByRWId(ctx, *surveyRespondentDetail.RWId)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent RW"), http.StatusInternalServerError)
	}

	errTx := service.surveyRepo.RunInTransaction(func(txRepo repository.SurveyRepo) error {
		err := txRepo.UpdateRespondentApprovalStatus(ctx, int64(surveyRespondent.ID), 2, "revisi_rw", "false")
		if err != nil {
			return err
		}

		var roleNameUserLogin string
		if respondentLogin.RoleId != nil {
			roleNameUserLogin = string(enums.RoleID(*respondentLogin.RoleId).Label())
		}

		var roleNameRespondentSurvey string
		if surveyRespondentDetail.RoleId != nil {
			roleNameRespondentSurvey = string(enums.RoleID(*surveyRespondentDetail.RoleId).Label())
		}

		logSurveys := []models.LogSurvey{
			{
				RespondentID: surveyRespondent.RespondentID,
				SurveyID:     surveyID,
				Keterangan:   utils.ToPtr(respondentLogin.Name + " (" + roleNameUserLogin + ") meminta revisi terhadap " + survey.Name + ", dengan nama responden " + surveyRespondentDetail.Name + " (" + roleNameRespondentSurvey + ")."),
				LogLevel:     utils.ToPtr(strconv.Itoa(int(enums.ROLE_RW))),
				NotifFor:     utils.ToPtr(strconv.Itoa(int(respondentRW.ID))),
				IsRead:       utils.ToPtr("false"),
			},
		}

		errLog := txRepo.MakeLogSurveyBulk(ctx, nil, logSurveys)
		if errLog != nil {
			return errors.New("Gagal mencatat log aktivitas survey")
		}

		return nil
	})

	if errTx != nil {
		return utils.SendError(errors.New("Gagal memproses penolakan berkas survey"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Survey berhasil ditolak dan dikembalikan ke RW untuk revisi")
}

func (service *surveyService) ValidateSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr, ok1 := slug["survey_code"].(string)
	codeWilayahStr, ok2 := slug["code"].(string)
	if !ok1 || !ok2 {
		return utils.SendError(errors.New("Parameter URL tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(codeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	decodedWilayahIDs, err := h.DecodeWithError(codeWilayahStr)
	if err != nil || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Code wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	rtId := int64(decodedWilayahIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	if respondentLoginId == 0 {
		return utils.SendError(errors.New("Respondent tidak ditemukan"), http.StatusNotFound)
	}

	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, int64(survey.ID), respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, int64(survey.ID), rtId)
	if err != nil || surveyRespondent == nil {
		return utils.SendError(errors.New("Data isian responden tidak ditemukan"), http.StatusNotFound)
	}

	if surveyRespondent.Status == nil || *surveyRespondent.Status != 2 {
		return utils.SendError(errors.New("Responden belum menyelesaikan pengisian survey ini"), http.StatusBadRequest)
	}
	if surveyRespondent.StatusApproval != nil && *surveyRespondent.StatusApproval != "" && *surveyRespondent.StatusApproval != "verified_rw" {
		return utils.SendError(errors.New("Survey ini sudah diverifikasi sebelumnya"), http.StatusBadRequest)
	}

	respondentSurveyDetail, err := service.userRepo.GetRespondentById(ctx, surveyRespondent.RespondentID)
	if err != nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	if respondentSurveyDetail == nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	respondentKecamatan, err := service.userRepo.GetRespondentByKecamatanId(ctx, *respondentSurveyDetail.KecamatanId)
	if err != nil {
		return utils.SendError(errors.New("Gagal mendapatkan data respondent Kecamatan"), http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	err = service.surveyRepo.RunInTransaction(func(txRepo repository.SurveyRepo) error {
		err = txRepo.UpdateRespondentApprovalStatus(ctx, int64(surveyRespondent.ID), 2, "validated_lurah", "false")
		if err != nil {
			return errors.New("Gagal memperbarui status approval responden")
		}

		var roleNameUserLogin string
		if respondentLogin.RoleId != nil {
			roleNameUserLogin = string(enums.RoleID(*respondentLogin.RoleId).Label())
		}

		var roleNameRespondentSurvey string
		if respondentSurveyDetail.RoleId != nil {
			roleNameRespondentSurvey = string(enums.RoleID(*respondentSurveyDetail.RoleId).Label())
		}

		logSurveys := []models.LogSurvey{
			{
				RespondentID: surveyRespondent.RespondentID,
				SurveyID:     surveyID,
				Keterangan:   utils.ToPtr(respondentLogin.Name + " (" + roleNameUserLogin + ") telah melakukan validasi " + survey.Name + ", dengan nama responden " + respondentSurveyDetail.Name + " (" + roleNameRespondentSurvey + ")."),
				LogLevel:     utils.ToPtr(strconv.Itoa(int(enums.ROLE_KECAMATAN))),
				NotifFor:     utils.ToPtr(strconv.Itoa(int(respondentKecamatan.ID))),
				IsRead:       utils.ToPtr("false"),
			},
		}

		errLog := txRepo.MakeLogSurveyBulk(ctx, nil, logSurveys)
		if errLog != nil {
			return errors.New("Gagal mencatat log aktivitas survey")
		}

		return nil
	})

	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Hasil pengisian survey berhasil divalidasi oleh Kelurahan")
}

func (service *surveyService) GetApprovalHistorySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr := slug["survey_code"]
	code, ok := codeStr.(string)
	if !ok {
		return utils.SendError(errors.New("Survey tidak valid"), http.StatusBadRequest)
	}

	decodedIDs, err := h.DecodeWithError(code)
	if err != nil || len(decodedIDs) == 0 {
		return utils.SendError(errors.New("Survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyId := decodedIDs[0]

	survey, err := service.surveyRepo.GetSurveyById(int64(surveyId))
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	dataLogs, err := service.surveyRepo.GetHistoryApproval(ctx, nil, int64(survey.ID))
	if err != nil {
		return utils.SendError(errors.New("Gagal mengambil riwayat persetujuan survey"), http.StatusInternalServerError)
	}

	return utils.SendData(dataLogs, "Riwayat persetujuan survey berhasil diambil")
}

func (service *surveyService) GetPublicImageSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	codeStr, ok1 := slug["survey_code"].(string)
	codeWilayahStr, ok2 := slug["code_wilayah"].(string)
	filePathStr, ok3 := slug["path"].(string)
	if !ok1 || !ok2 || !ok3 {
		return utils.SendError(errors.New("Parameter URL tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, err := h.DecodeWithError(codeStr)
	if err != nil || len(decodedSurveyIDs) == 0 {
		return utils.SendError(errors.New("Kode survey tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	surveyID := int64(decodedSurveyIDs[0])

	decodedWilayahIDs, err := h.DecodeWithError(codeWilayahStr)
	if err != nil || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Code wilayah tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}
	rtId := int64(decodedWilayahIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil || survey == nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, int64(survey.ID), rtId)
	if err != nil || surveyRespondent == nil {
		return utils.SendError(errors.New("Data isian responden tidak ditemukan"), http.StatusNotFound)
	}

	isFileExist, err := service.surveyRepo.ValidateIsImageExists(ctx, nil, surveyRespondent.ID, filePathStr)
	if err != nil {
		return utils.SendError(errors.New("Gagal memvalidasi keberadaan berkas gambar"), http.StatusInternalServerError)
	}
	if !isFileExist {
		return utils.SendError(errors.New("Berkas gambar tidak ditemukan"), http.StatusNotFound)
	}

	fileBytes, err := service.fileRepo.GetPublicImageSurvey(ctx, &filePathStr)
	if err != nil {
		spew.Dump(err)
		return utils.SendError(errors.New("Gagal mengambil gambar survey"), http.StatusInternalServerError)
	}

	mimeType := http.DetectContentType(fileBytes)

	return utils.SetResponseData(fileBytes, true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}

func (service *surveyService) ExportExcelSurveyResultsPerRT(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, _ := hashids.NewWithData(hd)

	codeStr, ok1 := slug["survey_code"].(string)
	codeWilayahStr, ok2 := slug["code_wilayah"].(string)
	if !ok1 || !ok2 {
		return utils.SendError(errors.New("Parameter URL tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, _ := h.DecodeWithError(codeStr)
	decodedWilayahIDs, _ := h.DecodeWithError(codeWilayahStr)
	if len(decodedSurveyIDs) == 0 || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Kode URL tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyID := int64(decodedSurveyIDs[0])
	rtId := int64(decodedWilayahIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	respondentLoginId := usr.RespondentID
	respondentLogin, err := service.userRepo.GetRespondentById(ctx, respondentLoginId)
	if err != nil || respondentLogin == nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, surveyID, respondentLogin)
	if err != nil || !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses menindak lanjut survey ini"), http.StatusUnauthorized)
	}

	surveyRespondent, err := service.surveyRepo.GetRespondentSurveyByRTId(ctx, surveyID, rtId)
	if err != nil || surveyRespondent == nil {
		return utils.SendError(errors.New("Data isian responden tidak ditemukan"), http.StatusNotFound)
	}

	if surveyRespondent.Status == nil || *surveyRespondent.Status != 2 {
		return utils.SendError(errors.New("Responden belum menyelesaikan pengisian survey ini"), http.StatusBadRequest)
	}

	respondentSurveyDetail, err := service.userRepo.GetRespondentById(ctx, surveyRespondent.RespondentID)
	if err != nil || respondentSurveyDetail == nil {
		return utils.SendError(errors.New("Data responden tidak ditemukan"), http.StatusUnauthorized)
	}

	wilayahInfo, err := service.wilayahRepo.GetRTById(ctx, rtId)
	if err != nil || wilayahInfo == nil {
		return utils.SendError(errors.New("Gagal memuat detail wilayah"), http.StatusInternalServerError)
	}

	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(int(survey.FlowDetailID), 0)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat struktur pertanyaan"), http.StatusInternalServerError)
	}

	var blueprintFieldIDs []int
	for _, node := range rawNodes {
		blueprintFieldIDs = append(blueprintFieldIDs, node.FormFieldId)
	}
	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(blueprintFieldIDs)

	optionMap := make(map[int]string)
	for _, opt := range options {
		if opt.Option != "" {
			optionMap[opt.ID] = opt.Option
		}
	}

	rawJawaban, err := service.surveyRepo.GetRawJawabanForExport(ctx, surveyID, rtId)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat hasil jawaban survei"), http.StatusInternalServerError)
	}

	rekapMap := make(map[int64]*dto.RekapRespondenExcel)
	var orderedRespondentIDs []int64

	for _, raw := range rawJawaban {
		if _, exists := rekapMap[raw.RespondentID]; !exists {
			statusText := "Belum Terisi"
			if raw.Status != nil {
				if *raw.Status == 2 && raw.StatusApproval != nil && *raw.StatusApproval == "revisi_rt" {
					statusText = "Revisi"
				} else if *raw.Status == 2 {
					statusText = "Selesai"
				} else if *raw.Status == 1 {
					statusText = "Draft (Sedang Berjalan)"
				}
			}

			waktuTeks := "-"
			if raw.WaktuSelesai != nil {
				waktuTeks = raw.WaktuSelesai.Format("02 Jan 2006 15:04:05")
			}

			rekapMap[raw.RespondentID] = &dto.RekapRespondenExcel{
				NamaResponden: raw.NamaResponden,
				StatusText:    statusText,
				WaktuSelesai:  waktuTeks,
				JawabanMap:    make(map[int]string),
			}
			orderedRespondentIDs = append(orderedRespondentIDs, raw.RespondentID)
		}

		if raw.Answer != nil {
			rekapMap[raw.RespondentID].JawabanMap[raw.FormFieldID] = *raw.Answer
		}
	}

	f := excelize.NewFile()
	defer f.Close()
	sheetName := "Hasil Survey"
	errRename := f.SetSheetName("Sheet1", sheetName)
	if errRename != nil {
		sheetName = "Sheet1"
	}

	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleLabel, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#15406a"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleDataCenter, _ := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "top"},
	})
	styleDataLeft, _ := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})

	styleDataText, _ := f.NewStyle(&excelize.Style{
		NumFmt:    49,
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})

	f.MergeCell(sheetName, "A1", "C3")
	titleText := fmt.Sprintf("Hasil Jawaban %s - %s",
		survey.Name, respondentSurveyDetail.Name)
	f.SetCellValue(sheetName, "A1", titleText)
	f.SetCellStyle(sheetName, "A1", "C3", styleTitle)

	infos := []struct{ row, label, value string }{
		{"5", "Kecamatan", wilayahInfo.SubDistrictName},
		{"6", "Kelurahan", wilayahInfo.VillageName},
		{"7", "RT", wilayahInfo.NamaRt},
		{"8", "RW", wilayahInfo.NamaRw},
	}
	for _, info := range infos {
		f.MergeCell(sheetName, "A"+info.row, "B"+info.row)
		f.SetCellValue(sheetName, "A"+info.row, info.label)
		f.SetCellValue(sheetName, "C"+info.row, "\u200B"+info.value)
		f.SetCellStyle(sheetName, "A"+info.row, "C"+info.row, styleLabel)
	}

	rowHeader := 10
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowHeader), "NO")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowHeader), "PERTANYAAN")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowHeader), "JAWABAN")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowHeader), fmt.Sprintf("C%d", rowHeader), styleHeader)

	f.SetColWidth(sheetName, "A", "A", 6)
	f.SetColWidth(sheetName, "B", "B", 65)
	f.SetColWidth(sheetName, "C", "C", 55)

	var currentRespondenData *dto.RekapRespondenExcel
	if data, exists := rekapMap[surveyRespondent.RespondentID]; exists {
		currentRespondenData = data
	} else if len(orderedRespondentIDs) > 0 {
		currentRespondenData = rekapMap[orderedRespondentIDs[0]]
	}

	currentRow := 11
	for index, node := range rawNodes {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), index+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), node.Label)

		jawabanText := "-"
		if currentRespondenData != nil {
			if val, ok := currentRespondenData.JawabanMap[node.FormFieldId]; ok {
				jawabanText = val

				if jawabanText != "-" && jawabanText != "[SKIPPED_BY_LOGIC]" {
					switch node.Template {
					case "number":
						jawabanText = "\u200B" + jawabanText
					case "multiple-choices":
						optID, errConv := strconv.Atoi(jawabanText)
						if errConv == nil {
							if labelText, exists := optionMap[optID]; exists {
								jawabanText = labelText
							}
						}
					case "checkboxes":
						var optIDs []float64
						if errUnm := json.Unmarshal([]byte(jawabanText), &optIDs); errUnm == nil {
							var labelTexts []string
							for _, idFloat := range optIDs {
								if label, exists := optionMap[int(idFloat)]; exists {
									labelTexts = append(labelTexts, "- "+label)
								}
							}
							if len(labelTexts) > 0 {
								jawabanText = strings.Join(labelTexts, "\n")
							}
						}
					case "maps":
						var mapsData []struct {
							Lat float64 `json:"lat"`
							Lng float64 `json:"lng"`
						}
						if errUnm := json.Unmarshal([]byte(jawabanText), &mapsData); errUnm == nil && len(mapsData) > 0 {
							jawabanText = fmt.Sprintf("%f,%f", mapsData[0].Lat, mapsData[0].Lng)
						}
					case "image-template":
						var jsonArray []string
						var finalJawaban string
						if errUnm := json.Unmarshal([]byte(jawabanText), &jsonArray); errUnm == nil {
							for imgIdx := range jsonArray {
								jsonArray[imgIdx] = os.Getenv("API_GATEWAY_URL") + "/survey/show-image/" + codeStr + "/" + codeWilayahStr + "/" + jsonArray[imgIdx]

								if imgIdx > 0 {
									finalJawaban += "\n\n"
								}

								finalJawaban += jsonArray[imgIdx]
							}
							jawabanText = finalJawaban
						}
					}
				} else {
					jawabanText = "-"
				}
			}
		}

		f.SetCellStr(sheetName, fmt.Sprintf("C%d", currentRow), jawabanText)

		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), styleDataCenter)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("B%d", currentRow), styleDataLeft)

		if node.Template == "number" {
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("C%d", currentRow), styleDataText)
		} else {
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("C%d", currentRow), styleDataLeft)
		}

		currentRow++
	}

	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return utils.SendError(errors.New("Gagal menyusun file excel"), http.StatusInternalServerError)
	}

	mimeType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	return utils.SetResponseData(buffer.Bytes(), true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}

func (service *surveyService) ExportExcelSurveyResultsMassal(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, _ := hashids.NewWithData(hd)

	codeStr, ok := slug["survey_code"].(string)
	if !ok || codeStr == "" {
		return utils.SendError(errors.New("Kode survey tidak valid"), http.StatusBadRequest)
	}

	levelWilayahStr := param.Get("level")
	codeWilayahStr := param.Get("code_wilayah")

	if levelWilayahStr == "" || codeWilayahStr == "" {
		return utils.SendError(errors.New("Parameter 'level' dan 'code_wilayah' wajib disertakan pada URL"), http.StatusBadRequest)
	}

	levelWilayah, err := strconv.Atoi(levelWilayahStr)
	if err != nil || levelWilayah < 2 || levelWilayah > 5 {
		return utils.SendError(errors.New("Parameter 'level' tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, _ := h.DecodeWithError(codeStr)
	decodedWilayahIDs, _ := h.DecodeWithError(codeWilayahStr)
	if len(decodedSurveyIDs) == 0 || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Kode HashID tidak valid atau termanipulasi"), http.StatusBadRequest)
	}

	surveyID := int64(decodedSurveyIDs[0])
	wilayahID := int64(decodedWilayahIDs[0])

	survey, err := service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(int(survey.FlowDetailID), 0)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat struktur pertanyaan"), http.StatusInternalServerError)
	}

	var blueprintFieldIDs []int
	for _, node := range rawNodes {
		blueprintFieldIDs = append(blueprintFieldIDs, node.FormFieldId)
	}
	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(blueprintFieldIDs)
	optionMap := make(map[int]string)
	for _, opt := range options {
		if opt.Option != "" {
			optionMap[opt.ID] = opt.Option
		}
	}

	rawJawaban, err := service.surveyRepo.GetRawJawabanWilayahForExport(ctx, surveyID, levelWilayah, wilayahID)
	if err != nil {
		return utils.SendError(errors.New("Gagal memuat hasil jawaban survei massal"), http.StatusInternalServerError)
	}

	rekapMap := make(map[int64]*dto.RekapRespondenWilayahExcel)
	var orderedRespondentIDs []int64

	for _, raw := range rawJawaban {
		if _, exists := rekapMap[raw.RespondentID]; !exists {
			waktuTeks := "-"
			if raw.WaktuSelesai != nil {
				waktuTeks = raw.WaktuSelesai.Format("02 Jan 2006 15:04:05")
			}
			rekapMap[raw.RespondentID] = &dto.RekapRespondenWilayahExcel{
				NamaResponden: raw.NamaResponden,
				KecamatanName: raw.KecamatanName,
				KelurahanName: raw.KelurahanName,
				RwName:        raw.RwName,
				RtName:        raw.RtName,
				WaktuSelesai:  waktuTeks,
				JawabanMap:    make(map[int]string),
				RTID:          raw.RTID,
			}
			orderedRespondentIDs = append(orderedRespondentIDs, raw.RespondentID)
		}
		if raw.Answer != nil {
			rekapMap[raw.RespondentID].JawabanMap[raw.FormFieldID] = *raw.Answer
		}
	}

	totalResp := len(orderedRespondentIDs)

	f := excelize.NewFile()
	defer f.Close()

	sheetName1 := "Rekap Hasil Per Responden"
	f.SetSheetName("Sheet1", sheetName1)

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#15406a"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleData, _ := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	styleLabel, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left"},
	})

	styleInfoValue, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left"},
	})

	startDateStr, endDateStr := "-", "-"
	if !survey.StartDate.IsZero() {
		startDateStr = survey.StartDate.Format("02 Jan 2006")
	}
	if !survey.EndDate.IsZero() {
		endDateStr = survey.EndDate.Format("02 Jan 2006")
	}

	f.MergeCell(sheetName1, "A1", "B1")
	f.MergeCell(sheetName1, "A2", "B2")
	f.MergeCell(sheetName1, "A3", "B3")

	f.MergeCell(sheetName1, "C1", "E1")
	f.MergeCell(sheetName1, "C2", "E2")
	f.MergeCell(sheetName1, "C3", "E3")

	f.SetCellValue(sheetName1, "A1", "Nama Survey")
	f.SetCellStyle(sheetName1, "A1", "A1", styleLabel)
	f.SetCellValue(sheetName1, "C1", survey.Name)
	f.SetCellStyle(sheetName1, "C1", "C1", styleInfoValue)
	f.SetCellValue(sheetName1, "A2", "Tanggal Survey")
	f.SetCellStyle(sheetName1, "A2", "A2", styleLabel)
	f.SetCellValue(sheetName1, "C2", fmt.Sprintf("%s - %s", startDateStr, endDateStr))
	f.SetCellStyle(sheetName1, "C2", "C2", styleInfoValue)
	f.SetCellValue(sheetName1, "A3", "Total Responden")
	f.SetCellStyle(sheetName1, "A3", "A3", styleLabel)
	f.SetCellValue(sheetName1, "C3", totalResp)
	f.SetCellStyle(sheetName1, "C3", "C3", styleInfoValue)

	staticHeaders := []string{"NO", "NAMA RESPONDEN", "KECAMATAN", "KELURAHAN", "RW", "RT"}
	colIndex := 1
	for _, h := range staticHeaders {
		colName, _ := excelize.ColumnNumberToName(colIndex)
		f.SetCellValue(sheetName1, fmt.Sprintf("%s5", colName), h)
		f.SetCellStyle(sheetName1, fmt.Sprintf("%s5", colName), fmt.Sprintf("%s5", colName), styleHeader)
		f.SetColWidth(sheetName1, colName, colName, 18)
		colIndex++
	}

	for i, node := range rawNodes {
		colName, _ := excelize.ColumnNumberToName(colIndex)
		f.SetCellValue(sheetName1, fmt.Sprintf("%s5", colName), fmt.Sprintf("%d. %s", i+1, node.Label))
		f.SetCellStyle(sheetName1, fmt.Sprintf("%s5", colName), fmt.Sprintf("%s5", colName), styleHeader)
		f.SetColWidth(sheetName1, colName, colName, 35)
		colIndex++
	}

	currentRow := 6
	no := 1
	for _, respondentID := range orderedRespondentIDs {
		data := rekapMap[respondentID]
		rtCodeRaw := []int{int(data.RTID)}
		rtCode, _ := h.Encode(rtCodeRaw)

		f.SetCellValue(sheetName1, fmt.Sprintf("A%d", currentRow), no)
		f.SetCellValue(sheetName1, fmt.Sprintf("B%d", currentRow), data.NamaResponden)
		f.SetCellValue(sheetName1, fmt.Sprintf("C%d", currentRow), data.KecamatanName)
		f.SetCellValue(sheetName1, fmt.Sprintf("D%d", currentRow), data.KelurahanName)
		f.SetCellValue(sheetName1, fmt.Sprintf("E%d", currentRow), "\u200B"+data.RwName)
		f.SetCellValue(sheetName1, fmt.Sprintf("F%d", currentRow), "\u200B"+data.RtName)

		colIdx := 7
		for _, node := range rawNodes {
			jawabanText := "-"
			if val, ok := data.JawabanMap[node.FormFieldId]; ok {
				jawabanText = val
				if jawabanText != "-" && jawabanText != "[SKIPPED_BY_LOGIC]" {
					switch node.Template {
					case "number":
						jawabanText = "\u200B" + jawabanText
					case "multiple-choices":
						if optID, errConv := strconv.Atoi(jawabanText); errConv == nil {
							if labelText, exists := optionMap[optID]; exists {
								jawabanText = labelText
							}
						}
					case "maps":
						var mapsData []struct {
							Lat float64 `json:"lat"`
							Lng float64 `json:"lng"`
						}
						if errUnm := json.Unmarshal([]byte(jawabanText), &mapsData); errUnm == nil && len(mapsData) > 0 {
							jawabanText = fmt.Sprintf("%f,%f", mapsData[0].Lat, mapsData[0].Lng)
						}
					case "image-template":
						var jsonArray []string
						var finalJawaban string
						if errUnm := json.Unmarshal([]byte(jawabanText), &jsonArray); errUnm == nil {
							for imgIdx := range jsonArray {
								jsonArray[imgIdx] = os.Getenv("API_GATEWAY_URL") + "/survey/show-image/" + codeStr + "/" + rtCode + "/" + jsonArray[imgIdx]
								if imgIdx > 0 {
									finalJawaban += "\n\n"
								}
								finalJawaban += jsonArray[imgIdx]
							}
							jawabanText = finalJawaban
						}
					}
				} else {
					jawabanText = "-"
				}
			}

			colName, _ := excelize.ColumnNumberToName(colIdx)
			f.SetCellValue(sheetName1, fmt.Sprintf("%s%d", colName, currentRow), jawabanText)
			colIdx++
		}

		lastColName, _ := excelize.ColumnNumberToName(colIdx - 1)
		f.SetCellStyle(sheetName1, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("%s%d", lastColName, currentRow), styleData)
		currentRow++
		no++
	}

	sheetStat := "Rekap Hasil Survey Per Kolom"
	f.NewSheet(sheetStat)

	styleStatTitle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	f.SetCellValue(sheetStat, "A1", "Nama Survey")
	f.SetCellValue(sheetStat, "B1", survey.Name)
	f.SetCellValue(sheetStat, "A2", "Tanggal Survey")
	f.SetCellValue(sheetStat, "B2", fmt.Sprintf("%s - %s", startDateStr, endDateStr))
	f.SetCellValue(sheetStat, "A3", "Total Responden")
	f.SetCellValue(sheetStat, "B3", fmt.Sprintf("%d Responden", totalResp))

	f.SetCellStyle(sheetStat, "A1", "A3", styleStatTitle)
	f.SetColWidth(sheetStat, "A", "A", 15)
	f.SetColWidth(sheetStat, "B", "B", 40)
	f.SetColWidth(sheetStat, "C", "C", 10)
	f.SetColWidth(sheetStat, "D", "D", 10)

	currentRowStat := 5

	for i, node := range rawNodes {
		qID := node.FormFieldId

		answeredCount := 0
		optionTallies := make(map[int]int)

		for _, respID := range orderedRespondentIDs {
			if data, ok := rekapMap[respID]; ok {
				if ans, exists := data.JawabanMap[qID]; exists && ans != "-" && ans != "[SKIPPED_BY_LOGIC]" && ans != "" {
					answeredCount++

					switch node.Template {
					case "multiple-choices":
						if optID, errConv := strconv.Atoi(ans); errConv == nil {
							optionTallies[optID]++
						}
					case "checkboxes":
						var optIDs []float64
						if errUnm := json.Unmarshal([]byte(ans), &optIDs); errUnm == nil {
							for _, idFloat := range optIDs {
								optionTallies[int(idFloat)]++
							}
						}
					}
				}
			}
		}

		pct := 0.0
		if totalResp > 0 {
			pct = (float64(answeredCount) / float64(totalResp)) * 100
		}

		f.SetCellValue(sheetStat, fmt.Sprintf("A%d", currentRowStat), fmt.Sprintf("Pertanyaan %d", i+1))
		f.SetCellValue(sheetStat, fmt.Sprintf("B%d", currentRowStat), node.Label)
		f.SetCellStyle(sheetStat, fmt.Sprintf("A%d", currentRowStat), fmt.Sprintf("B%d", currentRowStat), styleStatTitle)
		currentRowStat++

		if node.Template == "multiple-choices" || node.Template == "checkboxes" {
			f.SetCellValue(sheetStat, fmt.Sprintf("B%d", currentRowStat), "Jawaban :")
			currentRowStat++

			for _, opt := range options {
				if opt.FormFieldId == qID {
					optCount := optionTallies[opt.ID]
					optPct := 0.0
					if totalResp > 0 {
						optPct = (float64(optCount) / float64(totalResp)) * 100
					}

					f.SetCellValue(sheetStat, fmt.Sprintf("B%d", currentRowStat), "\u200B"+opt.Option)
					f.SetCellValue(sheetStat, fmt.Sprintf("C%d", currentRowStat), fmt.Sprintf("\u200B%d/%d", optCount, totalResp))
					f.SetCellValue(sheetStat, fmt.Sprintf("D%d", currentRowStat), fmt.Sprintf("\u200B%.0f%%", optPct))
					currentRowStat++
				}
			}
		} else {
			f.SetCellValue(sheetStat, fmt.Sprintf("B%d", currentRowStat), "Responden :")
			f.SetCellValue(sheetStat, fmt.Sprintf("C%d", currentRowStat), fmt.Sprintf("\u200B%d/%d", answeredCount, totalResp))
			f.SetCellValue(sheetStat, fmt.Sprintf("D%d", currentRowStat), fmt.Sprintf("\u200B%.0f%%", pct))
			currentRowStat++
		}

		currentRowStat++
	}

	f.SetActiveSheet(0)

	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		return utils.SendError(errors.New("Gagal menyusun file excel massal"), http.StatusInternalServerError)
	}

	mimeType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	return utils.SetResponseData(buffer.Bytes(), true, "Data File,"+mimeType, http.StatusOK, nil, ""), nil
}

func (service *surveyService) ResetStatusToVerifySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	if usr.Role != int(enums.ROLE_ADMIN) {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusForbidden)
	}

	type payloadStruct struct {
		TipeWilayah int `json:"tipe_wilayah" validate:"required,oneof=2 3 4 5"`
	}

	var payload payloadStruct

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

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, _ := hashids.NewWithData(hd)

	codeStr, ok1 := slug["survey_code"].(string)
	codeWilayahStr, ok2 := slug["code_wilayah"].(string)
	if !ok1 || !ok2 {
		return utils.SendError(errors.New("Parameter URL tidak valid"), http.StatusBadRequest)
	}

	decodedSurveyIDs, _ := h.DecodeWithError(codeStr)
	decodedWilayahIDs, _ := h.DecodeWithError(codeWilayahStr)
	if len(decodedSurveyIDs) == 0 || len(decodedWilayahIDs) == 0 {
		return utils.SendError(errors.New("Kode URL tidak valid atau dimanipulasi"), http.StatusBadRequest)
	}

	surveyID := int64(decodedSurveyIDs[0])
	wilayahID := int64(decodedWilayahIDs[0])

	_, err = service.surveyRepo.GetSurveyById(surveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	var wilayahColumn string
	switch payload.TipeWilayah {
	case 2:
		wilayahColumn = "rt_id"
	case 3:
		wilayahColumn = "rw_id"
	case 4:
		wilayahColumn = "kelurahan_id"
	case 5:
		wilayahColumn = "kecamatan_id"
	default:
		return utils.SendError(errors.New("Tipe wilayah tidak dikenali"), http.StatusBadRequest)
	}

	err = service.surveyRepo.ResetSurveyRespondentStatus(ctx, surveyID, wilayahColumn, wilayahID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mereset status responden, silakan coba lagi"), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Status responden di wilayah tersebut berhasil direset ke mode verifikasi")
}

func (service *surveyService) GetAllRejectedQuestions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	respondentID := usr.RespondentID
	if respondentID == 0 {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	data, err := service.surveyRepo.GetRejectedQuestionsGrouped(ctx, respondentID, nil)
	if err != nil {
		return utils.SendError(errors.New("Gagal mengambil data pertanyaan reject"), http.StatusInternalServerError)
	}

	return utils.SendData(data, "Berhasil mengambil semua data survey reject")
}

func (service *surveyService) GetRejectedQuestionsBySurveyCode(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	respondentID := usr.RespondentID
	if respondentID == 0 {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	surveyCodeStr, ok := slug["survey_code"].(string)
	if !ok || surveyCodeStr == "" {
		return utils.SendError(errors.New("parameter survey_code tidak ditemukan di URL"), http.StatusBadRequest)
	}

	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kendala saat inisiasi hashid"), http.StatusInternalServerError)
	}

	decoded, err := h.DecodeWithError(surveyCodeStr)
	if err != nil || len(decoded) == 0 {
		return utils.SendError(errors.New("Kode survey (survey_code) tidak valid atau korup"), http.StatusBadRequest)
	}

	surveyID := int64(decoded[0])

	data, err := service.surveyRepo.GetRejectedQuestionsGrouped(ctx, respondentID, &surveyID)
	if err != nil {
		return utils.SendError(errors.New("Gagal mengambil data pertanyaan reject"), http.StatusInternalServerError)
	}

	return utils.SendData(data, "Berhasil mengambil data survey reject berdasarkan kode")
}

func (service *surveyService) SyncExpiredSurveysStatus(ctx context.Context) {
	if !service.isSyncing.CompareAndSwap(false, true) {
		fmt.Println("SKIP: Ada proses yang sedang berjalan. Request ini dibuang.")
		return
	}

	defer service.isSyncing.Store(false)
	defer utils.GeneralRecover()

	expiredRepeated, _ := service.surveyRepo.GetExpiredRepeatedSurveys(ctx)
	if len(expiredRepeated) > 0 {
		_ = service.surveyRepo.RunInTransaction(func(txRepo repository.SurveyRepo) error {
			for _, oldSurvey := range expiredRepeated {
				durasi := oldSurvey.EndDate.Sub(oldSurvey.StartDate)
				newSurvey := oldSurvey
				newSurvey.ID = 0
				newSurvey.StartDate = time.Now()
				newSurvey.EndDate = time.Now().Add(durasi)
				newSurvey.CreatedAt = time.Now()
				newSurvey.UpdatedAt = time.Now()
				newSurvey.IsRepeated = utils.BoolToPointer(true)
				createdSurvey, err := txRepo.CreateSurvey(newSurvey)
				if err != nil {
					return err
				}

				oldSurvey.IsRepeated = utils.BoolToPointer(false)
				oldSurvey.Status = "finished"
				oldSurvey.UpdatedAt = time.Now()
				if err := txRepo.UpdateSurvey(&oldSurvey); err != nil {
					return err
				}

				surveyors, _ := service.surveyRepo.GetSurveyorsBySurveyId(int64(oldSurvey.ID))
				for _, v := range surveyors {
					_, err := txRepo.AssignSurveyorToSurvey(int64(v.RespondentId), int64(createdSurvey.ID))
					if err != nil {
						return err
					}
				}
				wilayahs, _ := service.surveyRepo.GetSurveyWilayahsBySurveyId(int64(oldSurvey.ID))
				for _, v := range wilayahs {
					surveyWilayah := models.SurveyWilayah{
						TingkatWilayah: v.TingkatWilayah,
						SurveyId:       int64(createdSurvey.ID),
						KecamatanId:    v.KecamatanId,
						KelurahanId:    v.KelurahanId,
						RWId:           v.RWId,
					}
					_, err := txRepo.AssignWilayahToSurvey(surveyWilayah)
					if err != nil {
						return err
					}
				}
			}

			return nil
		})
	}

	err := service.surveyRepo.MarkExpiredSurveysAsFinished(ctx)
	if err != nil {
		log.Printf("[SyncSurvey] Error saat bulk update status: %v", err)
	}
}
