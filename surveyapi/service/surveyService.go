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
	"slices"
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

	SubmitSurveyAnswers(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	UpdateSurveyRespondentStatus(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetApprovalHistorySurvey(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetHistoryDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetDetailSurveyKewilayahan(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type surveyService struct {
	manajemenAlurRepo              repository.ManajemenAlurRepo
	templateFormulirPertanyaanRepo repository.TemplateFormulirPertanyaanRepo
	templateUcapanRepo             repository.TemplateUcapanRepo
	surveyRepo                     repository.SurveyRepo
	wilayahRepo                    repository.WilayahRepo
	userRepo                       repository.UserRepo
	fileRepo                       repository.FileRepo
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
			} else if payload.TingkatPelaksanaan == int(enums.RW) {
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

	now := time.Now()

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

	// CHECK APAKAH RESPONDEN BOLEH MELAKUKAN SURVEY
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
	// var statusRespondentSurvey int
	if respondentExistsInSurvey != nil {
		if *respondentExistsInSurvey.Status == 2 {
			return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
		}
		surveyRespondentId = &respondentExistsInSurvey.ID
		// statusRespondentSurvey = *respondentExistsInSurvey.Status
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
		FlowCode:   flowCode,
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

	// CHECK APAKAH RESPONDEN BOLEH MELAKUKAN SURVEY
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
			return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
		}
		// Pastikan Anda membuat fungsi GetAnswersByResponseID di surveyRepo
		// Fungsi ini me-return []models.FieldResponse berdasarkan FormResponseID
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
	flowFieldToOptionMap := make(map[int]int)                   // Pemetaan flow_field_id ke option_id khusus breakdown
	breakdownRawMap := make(map[int]map[int]models.RawNodeData) // nodeKey -> optionId -> rawData

	var formFieldIDs []int
	var flowFieldIDs []int

	totalRequired := 0
	totalOptional := 0
	entryNodeId := 0
	var activeSectionName *string

	rJump := "jump-to"
	rLogic := "logic"

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

	// PART 2 Bangun Kerangka Node & Kalkulasi Progress
	for i, raw := range rawNodes {
		var nodeKey int
		if raw.GroupId != nil && *raw.GroupId != 0 {
			nodeKey = groupMasterMap[*raw.GroupId]
		} else {
			nodeKey = raw.FormFieldId
		}

		flowFieldToNodeKeyMap[raw.FlowFieldId] = nodeKey

		// Catat pemetaan option untuk logic breakdown nanti
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
			step.Routing.Rule = nil // Rule utama dikosongkan karena rute ada di level opsi
			if raw.FormAnswerFieldId != nil {
				// Simpan raw node untuk diekstrak rutenya ke dalam Option di PART 3
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

	// PART 3 Sisipkan Opsi (beserta Rute Breakdown)
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

					// Jika soal breakdown, masukkan Rule dan Target khusus opsi tersebut
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
							// Fallback aman
							optItem.Rule = &rJump
							optItem.IsEnd = true
						}
					}

					step.Questions[i].Options = append(step.Questions[i].Options, optItem)
				}
			}
		}
	}

	// PART 3.5 Sisipkan Advanced Logics
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
					// Jika breakdown, cari opsi spesifik dan letakkan logic ke dalam array opsi tersebut
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
					// Jika bukan breakdown, letakkan di root routing
					step.Routing.Logics = append(step.Routing.Logics, logicItem)
				}
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

	// Convert map pointer kembali ke map value untuk response
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

	// Detached context untuk worker di latar belakang (agar tidak mati jika user cancel HTTP request)
	detachedCtx := context.WithoutCancel(ctx)
	var payload payloads.SubmitSurveyPayload

	// 1. Parsing & Validasi Payload
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
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	// 2. Decode HashIDs (Survey & Section)
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

	// 3. Ambil Detail Survei & Validasi Hak Akses Responden
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

	canRespondentDoSurvey, err := service.surveyRepo.CheckRespondentEligibility(ctx, nil, surveyID, respondentLogin)
	if err != nil {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
	}

	if !canRespondentDoSurvey {
		return utils.SendError(errors.New("Anda tidak memiliki hak akses untuk mengikuti survey ini"), http.StatusUnauthorized)
	}

	// 4. Ambil Blueprint Pertanyaan KHUSUS Section yang Sedang Di-submit
	rawNodes, err := service.manajemenAlurRepo.GetRawNodesForPreview(int(surveyDetail.FlowDetailID), int(sectionID))
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if len(rawNodes) == 0 {
		return utils.SendError(errors.New("Tidak ada pertanyaan pada bagian ini"), http.StatusNotFound)
	}

	// Petakan ID Pertanyaan, Tipe Gambar, dan Soal Wajib (Required)
	var sectionFieldIDs []int64
	imageFieldsMap := make(map[int64]bool)

	requiredQuestionsMap := make(map[int64]string)  // Menyimpan Nama/Label Soal Wajib
	answeredRequiredTracker := make(map[int64]bool) // Menyimpan Status Terjawab (True/False)

	for _, node := range rawNodes {
		fieldID := int64(node.FormFieldId)
		sectionFieldIDs = append(sectionFieldIDs, fieldID)

		if node.Template == "image-template" {
			imageFieldsMap[fieldID] = true
		}

		// 🛡️ TANDAI SOAL WAJIB DARI BLUEPRINT
		if node.IsRequired {
			requiredQuestionsMap[fieldID] = node.Label
			answeredRequiredTracker[fieldID] = false // Set default: Belum dijawab
		}
	}

	// 5. Inisiasi Transaksi Database & Tracker Gambar MinIO
	tx := service.surveyRepo.BeginTransaction()
	if tx.Error != nil {
		return utils.SendError(errors.New("Gagal memulai transaksi database"), http.StatusInternalServerError)
	}

	var successfullyUploadedFiles []string // Tracker untuk aksi kompensasi jika DB rollback
	txCommitted := false

	// Pengawal Eksekusi (Defer): Jika gagal di tengah jalan, bersihkan DB dan File!
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

	// 6. Pastikan Record Induk (Survey Respondent) Hanya Ada 1
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
			return utils.SendError(errors.New("Anda sudah menyelesaikan survey ini"), http.StatusBadRequest)
		}
	}

	// 7. Kumpulkan Berkas Gambar Lama untuk Dibandingkan (Cegah Berkas Yatim)
	oldFilesTracker := make(map[string]bool) // true = akan dihapus dari MinIO
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

		// 💥 WIPE & REPLACE: Hapus total jawaban lama khusus di section ini
		errDelOld := service.surveyRepo.DeleteOldResponsesBySection(tx, respondentExistsInSurvey.ID, sectionFieldIDs)
		if errDelOld != nil {
			return utils.SendError(errors.New("Gagal membersihkan riwayat jawaban lama pada bagian ini"), http.StatusInternalServerError)
		}
	}

	// 8. Pemrosesan Payload Jawaban Baru
	var answersToInsert []models.FieldResponse
	gatewayURL := os.Getenv("API_GATEWAY_URL") + "/view-survey-image/"

	for _, ans := range payload.Answers {
		// 🛡️ PROTEKSI EDGE CASE (Jump-To Logic)
		if ans.ValueString != nil && *ans.ValueString == "[SKIPPED_BY_LOGIC]" {
			if _, isRequired := answeredRequiredTracker[int64(ans.QuestionID)]; isRequired {
				answeredRequiredTracker[int64(ans.QuestionID)] = true // Anggap terjawab
			}

			// 👇 UBAHAN BARU: Kita harus simpan record ini ke DB agar SQL COUNT() di PreviewIndex bisa menghitungnya!
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

			continue // Skip ke soal selanjutnya (jangan diproses masuk ke switch case)
		}

		var answerText string
		var groupID int64 = 0
		if ans.GroupID != nil {
			groupID = int64(*ans.GroupID)
		}

		switch ans.Type {
		case "long-answer":
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

				availableMime := []string{"image/png", "image/jpg", "image/jpeg"}
				availablesExt := []string{".png", ".jpg", ".jpeg"}
				maxSizeInKB := float64(5120)
				var uploadedPaths []string

				for _, image := range ans.ValueImages {
					if strings.HasPrefix(image, "data:image") {
						// Gambar baru berbentuk Base64 -> Upload ke MinIO
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
						// Gambar lama dipertahankan responden -> Bersihkan prefix URL Gateway
						cleanPath := image
						if after, ok0 := strings.CutPrefix(cleanPath, gatewayURL); ok0 {
							cleanPath = after
						}
						uploadedPaths = append(uploadedPaths, cleanPath)

						// Batalkan jadwal penghapusan berkas ini di MinIO
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

		// Jika kosong, skip penyimpanan ke DB
		if strings.TrimSpace(answerText) == "" {
			continue
		}

		// 🛡️ TANDAI SOAL WAJIB SUDAH DIJAWAB
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

	// 9. CEK FINAL VALIDASI WAJIB (REQUIRED)
	var missingLabels []string
	for qID, label := range requiredQuestionsMap {
		if !answeredRequiredTracker[qID] {
			missingLabels = append(missingLabels, label)
		}
	}

	// JIKA ADA ERROR REQUIRED:
	// Return error otomatis memicu blok "defer" di atas untuk me-rollback database
	// dan membuang gambar yang terlanjur terupload pada looping payload tadi.
	if len(missingLabels) > 0 {
		errMsg := fmt.Sprintf("Ada pertanyaan wajib yang belum dijawab: %s", strings.Join(missingLabels, ", "))
		return utils.SendError(errors.New(errMsg), http.StatusBadRequest)
	}

	// 10. Bulk Insert Seluruh Data Baru untuk Section Ini
	if len(answersToInsert) > 0 {
		err = service.surveyRepo.BulkInsertFieldResponses(ctx, tx, answersToInsert)
		if err != nil {
			return utils.SendError(errors.New("Gagal menyimpan jawaban survei, terjadi kesalahan server"), http.StatusInternalServerError)
		}
	}

	// 11. Commit Transaksi Database
	if err := tx.Commit().Error; err != nil {
		return utils.SendError(errors.New("Gagal memfinalisasi data"), http.StatusInternalServerError)
	}
	txCommitted = true

	// 12. Bersihkan Gambar Yatim dari MinIO secara Asinkron
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

	// 1. Parsing & Validasi Payload
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			customErrorMsg := utils.TranslateError(err)
			return utils.SendError(errors.New(customErrorMsg), http.StatusBadRequest)
		}
	}

	// 2. Decode Survey ID dari HashID
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

	// 3. Cek Kredensial dan Eligibility Responden
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

	// 4. Inisiasi Transaksi Database
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

	// 5. Cek Apakah Record Responden Survei Sudah Ada
	respondentExistsInSurvey, err := service.surveyRepo.GetRespondentExistsInSurvey(respondentId, surveyID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	if respondentExistsInSurvey == nil {
		return utils.SendError(errors.New("Anda tidak dapat mengubah status data, ikuti survey terlebih dahulu"), http.StatusBadRequest)
	}

	targetStatus := *payload.Status
	currentStatus := 0
	if respondentExistsInSurvey.Status != nil {
		currentStatus = *respondentExistsInSurvey.Status
	}

	// 5.5 VALIDASI STATE MACHINE (TIDAK BOLEH MUNDUR ATAU REDUDAN)
	if currentStatus == 2 {
		return utils.SendError(errors.New("Survei ini sudah diselesaikan dan datanya telah terkunci"), http.StatusBadRequest)
	}

	if currentStatus == targetStatus {
		// Jika frontend mengirim status yang sama dengan yang ada di DB, kita tidak perlu membuang resource untuk update
		return utils.SendData(nil, "Status pengerjaan berhasil diperbarui")
	}

	// BLOK BARU: Mencegah status mundur (misal dari 1 ke 0)
	if targetStatus < currentStatus {
		return utils.SendError(errors.New("Survei yang sedang dikerjakan tidak dapat dikembalikan ke status belum mulai"), http.StatusBadRequest)
	}

	// 6. VALIDASI KETAT PENYELESAIAN SURVEI
	// Asumsi: targetStatus == 2 adalah status ketika responden menekan tombol "Selesaikan Survei"
	if targetStatus == 2 {
		// Ambil detail survei untuk mendapatkan flow_detail_id
		surveyDetail, err := service.surveyRepo.GetSurveyById(surveyID)
		if err != nil {
			return utils.SendError(errors.New("Gagal mendapatkan detail survei"), http.StatusInternalServerError)
		}

		// Ambil flow detail untuk mengetahui status_section
		flowDetail, err := service.manajemenAlurRepo.GetFlowDetailByID(int(surveyDetail.FlowDetailID))
		if err != nil {
			return utils.SendError(errors.New("Gagal mendapatkan detail alur survei"), http.StatusInternalServerError)
		}

		// Kalkulasi progres jawaban berdasarkan record yang ada
		rawSections, err := service.manajemenAlurRepo.GetPreviewSectionByFlowDetailId(flowDetail.ID, flowDetail.StatusSection, &respondentExistsInSurvey.ID)
		if err != nil {
			return utils.SendError(errors.New("Gagal mengecek progres jawaban survei"), http.StatusInternalServerError)
		}

		// Lakukan iterasi untuk memastikan tidak ada satupun section yang belum selesai
		for _, section := range rawSections {
			if !section.Completed {
				errMsg := fmt.Sprintf("Anda tidak dapat menyelesaikan survei. %s belum selesai dikerjakan.", *section.SectionName)
				return utils.SendError(errors.New(errMsg), http.StatusBadRequest)
			}
		}
	}

	// 7. Eksekusi Update Status jika Lolos Validasi
	err = service.surveyRepo.UpdateStatusSurvey(ctx, tx, respondentExistsInSurvey.ID, targetStatus)
	if err != nil {
		return utils.SendError(errors.New("Gagal memperbarui status survei"), http.StatusInternalServerError)
	}

	// Hanya mencatat log aktivitas jika status BERUBAH menjadi 2 (Selesai)
	if targetStatus == 2 {
		var roleName string
		if respondentLogin.RoleId != nil {
			roleName = string(enums.RoleID(*respondentLogin.RoleId).Label())
		}

		keterangan := respondentLogin.Name + " (" + roleName + ") telah menyelesaikan survey."

		logLevel := strconv.Itoa(int(enums.ROLE_RW))
		notifFor := strconv.Itoa(int(respondentLogin.ID))
		isRead := "false"
		logSurvey := models.LogSurvey{
			RespondentID: respondentId,
			SurveyID:     surveyID,
			Keterangan:   &keterangan,
			LogLevel:     &logLevel,
			NotifFor:     &notifFor,
			IsRead:       &isRead,
		}

		err = service.surveyRepo.MakeLogSurvey(ctx, tx, logSurvey)
		if err != nil {
			return utils.SendError(errors.New("Gagal mencatat log aktivitas survey"), http.StatusInternalServerError)
		}
	}

	// 8. Commit Transaksi
	if err := tx.Commit().Error; err != nil {
		return utils.SendError(errors.New("Gagal memfinalisasi data"), http.StatusInternalServerError)
	}
	txCommitted = true

	// Custom Message berdasarkan status
	msg := "Status pengerjaan berhasil diperbarui"
	if targetStatus == 2 {
		msg = "Survei berhasil diselesaikan!"
	}

	return utils.SendData(nil, msg)
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

func (service *surveyService) GetHistoryDetail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	return utils.SendError(errors.New("Fitur ini belum tersedia"), http.StatusNotImplemented)
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

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	tipe_wilayah, _ := strconv.Atoi(param.Get("tipe_wilayah"))
	kecamatan_idStr, _ := strconv.Atoi(param.Get("kecamatan_id"))
	kelurahan_idStr, _ := strconv.Atoi(param.Get("kelurahan_id"))
	rw_idStr, _ := strconv.Atoi(param.Get("rw_id"))

	kecamatan_id := int64(kecamatan_idStr)
	kelurahan_id := int64(kelurahan_idStr)
	rw_id := int64(rw_idStr)

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

	var metaTotal int
	var metaPage int
	var metaLimit int
	var metaTotalPages int

	var finalData []response.DetailSurveyKewilayahanResponse

	if payload.TipeWilayah != nil {
		switch *payload.TipeWilayah {
		case 0:
			return utils.SendError(errors.New("Tipe wilayah harus diisi"), http.StatusBadRequest)
		case 4:
			if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
			}

			if *payload.KecamatanId == 0 {
				return utils.SendError(errors.New("Kecamatan ID harus diisi untuk tipe wilayah Kecamatan"), http.StatusBadRequest)
			}

			data, err := service.wilayahRepo.GetDaftarKelurahan(ctx, *payload.KecamatanId, payloads.DatatablePayload{
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
				for _, rt := range data.Data {
					extractedKelurahanIds = append(extractedKelurahanIds, rt.ID)
				}

				statusMap, _ := service.surveyRepo.GetStatusKeterisianBulkKelurahan(ctx, int64(survey.ID), extractedKelurahanIds)

				for _, kelurahan := range data.Data {
					newRt := response.DetailSurveyKewilayahanResponse{
						ID:              kelurahan.ID,
						NamaWilayah:     kelurahan.VillageName,
						IsPosibleDetail: true,
					}

					if statusStr, exists := statusMap[kelurahan.ID]; exists {
						val := statusStr
						newRt.Status = val
					}

					finalData = append(finalData, newRt)
				}
			}
		case 3:
			if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
			}

			if *payload.KelurahanId == 0 {
				return utils.SendError(errors.New("Kelurahan ID harus diisi untuk tipe wilayah Kelurahan"), http.StatusBadRequest)
			}

			data, err := service.wilayahRepo.GetDaftarRW(ctx, *payload.KelurahanId, payloads.DatatablePayload{
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
				for _, rt := range data.Data {
					extractedRWIds = append(extractedRWIds, rt.ID)
				}

				statusMap, _ := service.surveyRepo.GetStatusKeterisianBulkRW(ctx, int64(survey.ID), extractedRWIds)

				for _, rw := range data.Data {
					newRt := response.DetailSurveyKewilayahanResponse{
						ID:              rw.ID,
						NamaWilayah:     rw.NamaRw,
						IsPosibleDetail: true,
					}

					if statusStr, exists := statusMap[rw.ID]; exists {
						val := statusStr
						newRt.Status = val
					}

					finalData = append(finalData, newRt)
				}
			}
		case 2:
			if *respondentLogin.RoleId != int64(enums.ROLE_ADMIN) && *respondentLogin.RoleId != int64(enums.ROLE_KECAMATAN) && *respondentLogin.RoleId != int64(enums.ROLE_KELURAHAN) && *respondentLogin.RoleId != int64(enums.ROLE_RW) {
				return utils.SendError(errors.New("Anda tidak memiliki hak akses"), http.StatusUnauthorized)
			}

			if *payload.RWId == 0 {
				return utils.SendError(errors.New("RW ID harus diisi untuk tipe wilayah RW"), http.StatusBadRequest)
			}

			data, err := service.wilayahRepo.GetDaftarRT(ctx, *payload.RWId, payloads.DatatablePayload{
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

				for _, rt := range data.Data {
					newRt := response.DetailSurveyKewilayahanResponse{
						ID:          rt.ID,
						NamaWilayah: rt.NamaRt,
					}

					if statusStr, exists := statusMap[rt.ID]; exists {
						val := statusStr
						if val != "Tidak ada responden" {
							newRt.IsPosibleDetail = true
							newRt.IsPosiblePreviewSurvey = true
						}
						newRt.Status = val
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
