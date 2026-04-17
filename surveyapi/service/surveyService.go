package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/enums"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/response"
	"backend/surveyapi/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
	"gorm.io/gorm"
)

type SurveyService interface {
	OptionsPeriodeSurvey(usr models.JwtCustomClaims, param url.Values) (*pb.ProxyResponse, error)
	CreateSurvey(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}) (*pb.ProxyResponse, error)
	GetListSurvey(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type surveyService struct {
	manajemenAlurRepo              repository.ManajemenAlurRepo
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo
	templateUcapanRepo             repository.TemplateUcapanRepo
	surveyRepo                     repository.SurveyRepo
	wilayahRepo                    repository.WilayahRepo
	userRepo                       repository.UserRepo
}

func NewSurveyService(
	manajemenAlurRepo repository.ManajemenAlurRepo,
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo,
	templateUcapanRepo repository.TemplateUcapanRepo,
	surveyRepo repository.SurveyRepo,
	wilayahRepo repository.WilayahRepo,
	userRepo repository.UserRepo,
) SurveyService {
	return &surveyService{
		manajemenAlurRepo:              manajemenAlurRepo,
		templateFormulirPertanyaanRepo: templateFormulirPertanyaanRepo,
		templateUcapanRepo:             templateUcapanRepo,
		surveyRepo:                     surveyRepo,
		wilayahRepo:                    wilayahRepo,
		userRepo:                       userRepo,
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
		if !enums.IsWilayahExist(enums.WilayahID(payload.TingkatPelaksanaan)) {
			return utils.SendError(errors.New("Tingkat pelaksanaan survey tidak valid"), http.StatusBadRequest)
		}

		if payload.TingkatPelaksanaan == int(enums.KECAMATAN) {
			if payload.Kecamatan == 0 {
				return utils.SendError(errors.New("Kecamatan harus diisi untuk tingkat pelaksanaan kecamatan"), http.StatusBadRequest)
			}

			_, err = service.wilayahRepo.GetKecamatanById(ctx, int64(payload.Kecamatan))
			if err != nil {
				return utils.SendError(errors.New("Kecamatan tidak ditemukan"), http.StatusBadRequest)
			}
		} else if payload.TingkatPelaksanaan == int(enums.KELURAHAN) {
			if payload.Kecamatan == 0 {
				return utils.SendError(errors.New("Kecamatan harus diisi untuk tingkat pelaksanaan kelurahan"), http.StatusBadRequest)
			}
			if payload.Kelurahan == 0 {
				return utils.SendError(errors.New("Kelurahan harus diisi untuk tingkat pelaksanaan kelurahan"), http.StatusBadRequest)
			}

			_, err = service.wilayahRepo.GetKecamatanById(ctx, int64(payload.Kecamatan))
			if err != nil {
				return utils.SendError(errors.New("Kecamatan tidak ditemukan"), http.StatusBadRequest)
			}

			_, err = service.wilayahRepo.GetKelurahanById(ctx, int64(payload.Kelurahan))
			if err != nil {
				return utils.SendError(errors.New("Kelurahan tidak ditemukan"), http.StatusBadRequest)
			}
		} else if payload.TingkatPelaksanaan == int(enums.RW) {
			if payload.Kecamatan == 0 {
				return utils.SendError(errors.New("Kecamatan harus diisi untuk tingkat pelaksanaan RW"), http.StatusBadRequest)
			}
			if payload.Kelurahan == 0 {
				return utils.SendError(errors.New("Kelurahan harus diisi untuk tingkat pelaksanaan RW"), http.StatusBadRequest)
			}
			if payload.RW == 0 {
				return utils.SendError(errors.New("RW harus diisi untuk tingkat pelaksanaan RW"), http.StatusBadRequest)
			}

			_, err = service.wilayahRepo.GetKecamatanById(ctx, int64(payload.Kecamatan))
			if err != nil {
				return utils.SendError(errors.New("Kecamatan tidak ditemukan"), http.StatusBadRequest)
			}

			_, err = service.wilayahRepo.GetKelurahanById(ctx, int64(payload.Kelurahan))
			if err != nil {
				return utils.SendError(errors.New("Kelurahan tidak ditemukan"), http.StatusBadRequest)
			}

			_, err = service.wilayahRepo.GetRWById(ctx, int64(payload.RW))
			if err != nil {
				return utils.SendError(errors.New("RW tidak ditemukan"), http.StatusBadRequest)
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

	if usr.Role == 5 {
		approvalStatus = "waiting"
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

		tingkatWilayahStr := strconv.FormatInt(int64(payload.TingkatPelaksanaan), 10)

		if payload.RespondenSurvey == 2 {
			surveyWilayah := models.SurveyWilayah{
				TingkatWilayah: tingkatWilayahStr,
				SurveyId:       int64(createdSurvey.ID),
			}

			if payload.TingkatPelaksanaan == int(enums.KECAMATAN) {
				surveyWilayah.KecamatanId = payload.Kecamatan
			}

			if payload.TingkatPelaksanaan == int(enums.KELURAHAN) {
				surveyWilayah.KelurahanId = &payload.Kelurahan
			}

			if payload.TingkatPelaksanaan == int(enums.RW) {
				surveyWilayah.RWId = &payload.RW
			}

			_, err := txRepo.AssignWilayahToSurvey(surveyWilayah)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return utils.SendError(errors.New("gagal membuat survey: "+err.Error()), http.StatusInternalServerError)
	}

	return utils.SendData(nil, "Survey berhasil dibuat!")
}

func (service *surveyService) GetListSurvey(usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.DatatablePayload
	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	if payload.Page <= 0 {
		payload.Page = 1
	}
	if payload.Limit <= 0 {
		payload.Limit = 25
	}

	data, totalData, err := service.surveyRepo.GetListSurvey(payload)
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

		if data[i].Approval == "waiting" {
			if usr.Role == 7 {
				data[i].PosibleApproval = true
			}
		}

		if data[i].CreatedBy == int(usr.ID) {
			data[i].IsMyOwn = true
		} else {
			data[i].IsMyOwn = false
		}

		isUsed := false

		if data[i].IsMyOwn && !isUsed {
			data[i].PosibleUpdate = true
			data[i].PosibleDelete = true
		} else {
			data[i].PosibleUpdate = false
			data[i].PosibleDelete = false
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

	return utils.SendData(result, "Berhasil mengambil list survey")
}
