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
	"encoding/json"
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
	GetListSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	ApprovalSurvey(ctx context.Context, usr models.JwtCustomClaims, req map[string]interface{}, slug map[string]interface{}) (*pb.ProxyResponse, error)

	AvailableSurveyWilayah(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	PreviewSurveyIndex(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	PreviewSurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	SurveyBundlingSubmit(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
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
			if len(payload.Kecamatan) == 0 {
				return utils.SendError(errors.New("Kecamatan harus diisi untuk tingkat pelaksanaan kecamatan"), http.StatusBadRequest)
			}

			for _, kecamatanId := range payload.Kecamatan {
				_, err = service.wilayahRepo.GetKecamatanById(ctx, kecamatanId)
				if err != nil {
					return utils.SendError(fmt.Errorf("Kecamatan dengan ID %d tidak ditemukan", kecamatanId), http.StatusBadRequest)
				}

			}

		} else if payload.TingkatPelaksanaan == int(enums.KELURAHAN) {
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
		} else if payload.TingkatPelaksanaan == int(enums.RW) {
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
			if payload.TingkatPelaksanaan == int(enums.KECAMATAN) {
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
			} else if payload.TingkatPelaksanaan == int(enums.KELURAHAN) {
				for _, kelurahanId := range payload.Kelurahan {
					surveyWilayah := models.SurveyWilayah{
						TingkatWilayah: tingkatWilayahStr,
						SurveyId:       int64(createdSurvey.ID),
						KelurahanId:    &kelurahanId,
					}

					_, err := txRepo.AssignWilayahToSurvey(surveyWilayah)
					if err != nil {
						return err
					}
				}
			} else if payload.TingkatPelaksanaan == int(enums.RW) {
				for _, rwId := range payload.RW {
					surveyWilayah := models.SurveyWilayah{
						TingkatWilayah: tingkatWilayahStr,
						SurveyId:       int64(createdSurvey.ID),
						RWId:           &rwId,
					}

					_, err := txRepo.AssignWilayahToSurvey(surveyWilayah)
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

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	surveyDiikutiStr := param.Get("survey_diikuti")
	surveyDiikuti := false
	if surveyDiikutiStr == "true" {
		surveyDiikuti = true
	}

	payload := payloads.SurveyDatatablePayload{
		Search:        search,
		Page:          page,
		Limit:         limit,
		OrderBy:       orderBy,
		OrderDir:      orderDir,
		SurveyDiikuti: surveyDiikuti,
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

	for i := range data {
		surveyId := []int{data[i].ID}

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

	var payload payloads.SurveyWilayahDatatablePayload
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

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(survey.FlowDetailID))
	if err != nil {
		return utils.SendError(errors.New("alur survey tidak ditemukan"), http.StatusNotFound)
	}

	statusSectionStr := flowDetail.StatusSection
	statusSectionInt, _ := strconv.Atoi(statusSectionStr)
	hasSectionBool := utils.IntToBool(statusSectionInt)

	var sections []models.FlowPreviewSection
	rawSections, err := service.manajemenAlurRepo.GetPreviewSectionByFlowDetailId(flowDetail.ID, statusSectionStr)
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
			SectionCode:            &sectionCode,
			SectionName:            v.SectionName,
			TotalRequiredQuestions: v.TotalRequiredQuestions,
			TotalOptionalQuestions: v.TotalOptionalQuestions,
		})
	}

	data := models.FlowPreview{
		FlowName:   flowDetail.Name,
		HasSection: hasSectionBool,
		Sections:   sections,
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

	flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(survey.FlowDetailID))
	if err != nil {
		return utils.SendError(errors.New("Alur survey tidak ditemukan"), http.StatusNotFound)
	}

	sectionID := 0
	if sectionCode != "0" && sectionCode != "" {
		hd := hashids.NewData()
		hd.Salt = os.Getenv("HASHID_SALT")
		hd.MinLength = 24

		h, err := hashids.NewWithData(hd)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		decodedIDs, err := h.DecodeWithError(sectionCode)
		if err != nil || len(decodedIDs) == 0 {
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

	blueprintNodes := make(map[int]response.PreviewAlurSurveyStep)
	groupMasterMap := make(map[int]int)

	flowFieldToNodeKeyMap := make(map[int]int)

	var formFieldIDs []int
	var flowFieldIDs []int

	totalRequired := 0
	totalOptional := 0
	entryNodeId := 0
	var activeSectionName *string

	// PART 1 Peta Grup & Pencatatan ID Master
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
				return &target
			}
		}
		if childID != 0 {
			return &childID
		}
		return nil
	}

	// PART 2 Bangun Kerangka Node & Kalkulasi Progress
	for i, raw := range rawNodes {
		var nodeKey int
		if raw.GroupId != nil && *raw.GroupId != 0 {
			nodeKey = groupMasterMap[*raw.GroupId]
		} else {
			nodeKey = raw.FormFieldId
		}

		flowFieldToNodeKeyMap[raw.FlowFieldId] = nodeKey

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

			step = response.PreviewAlurSurveyStep{
				StepType:  stepType,
				GroupId:   raw.GroupId,
				GroupName: raw.GroupName,
				Questions: []response.PreviewAlurSurveyQuestionDetail{},
				Routing: response.PreviewAlurSurveyRoutingDetail{
					BreakdownRoutes: make(map[int]int),
					AdvancedLogics:  []response.PreviewAlurSurveyAdvancedLogicItem{},
				},
			}
		}

		// Masukkan Pertanyaan
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

			step.Questions = append(step.Questions, response.PreviewAlurSurveyQuestionDetail{
				QuestionId:         raw.FormFieldId,
				Type:               raw.Template,
				Label:              raw.Label,
				IsRequired:         raw.IsRequired,
				ExpectedImageCount: expectedImageCount,
				Options:            []response.PreviewAlurSurveyOptionItem{},
			})

			if raw.IsRequired {
				totalRequired++
			} else {
				totalOptional++
			}
		}

		target := resolveTargetID(raw.ChildId, int(raw.GroupChildId))

		if raw.IsAdvancedOption {
			step.Routing.Rule = "logic"
		} else if raw.Breakdown && raw.FormAnswerFieldId != nil {
			step.Routing.Rule = "jump-to"
			if target != nil {
				step.Routing.BreakdownRoutes[*raw.FormAnswerFieldId] = *target
			}
		} else {
			step.Routing.Rule = "jump-to"
			if target == nil {
				step.Routing.IsEnd = true
			} else {
				step.Routing.DefaultNext = target
			}
		}

		blueprintNodes[nodeKey] = step
	}

	// PART 3 Sisipkan Opsi & Advanced Logics
	options, _ := service.manajemenAlurRepo.GetAnswerOptionsByQuestionIDList(formFieldIDs)
	for _, opt := range options {
		for nodeKey, step := range blueprintNodes {
			for i, q := range step.Questions {
				if q.QuestionId == opt.FormFieldId {
					step.Questions[i].Options = append(step.Questions[i].Options, response.PreviewAlurSurveyOptionItem{
						ID:    opt.ID,
						Label: opt.Option,
					})
					blueprintNodes[nodeKey] = step
				}
			}
		}
	}

	logics, _ := service.manajemenAlurRepo.GetAdvancedOptionsByFieldIDs(flowFieldIDs)
	for _, logic := range logics {
		nodeKey, valid := flowFieldToNodeKeyMap[logic.FlowFieldId]

		if valid {
			if step, exists := blueprintNodes[nodeKey]; exists {

				step.Routing.AdvancedLogics = append(step.Routing.AdvancedLogics, response.PreviewAlurSurveyAdvancedLogicItem{
					IfQuestionId:     logic.FormFieldId,
					IfOptionId:       logic.Option,
					TargetQuestionId: logic.ChildId,
				})

				blueprintNodes[nodeKey] = step
			}
		}
	}

	// PART 4 Ambil Konten Opening & Closing
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

	// FINAL Bungkus ke Data Response
	hasSectionBool := utils.StringToBool(flowDetail.StatusSection)

	var activeSecCode *string
	if hasSectionBool && sectionID != 0 {
		activeSecCode = &sectionCode
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
		Nodes:   blueprintNodes,
	}

	return utils.SendData(dataResponse, "Berhasil mengambil data untuk preview survey")
}

func (service *surveyService) SurveyBundlingSubmit(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	return utils.SendData(nil, "Endpoint submit survey berhasil dipanggil")
}
