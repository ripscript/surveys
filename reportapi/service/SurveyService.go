package service

import (
	"backend/reportapi/excel"
	"backend/reportapi/models"
	"backend/reportapi/payloads"
	"backend/reportapi/repository"
	"backend/reportapi/utils"
	"backend/siccore/pb"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
)

type SurveyService interface {
	LogSurveys(param url.Values) (*pb.ProxyResponse, error)
	InsertLogSurveys(req map[string]interface{}) (*pb.ProxyResponse, error)
	GetDataSurvey(usr models.JwtCustomClaims, status *string, param url.Values) (*pb.ProxyResponse, error)
	SurveyActivitiesExport(req ExportRequest) (*ExportResult, error)
}

type ExportRequest struct {
	SurveyID  string
	Wilayah   string
	WilayahID string
}

type ExportResult struct {
	Filename string
	Data     []byte
}

type surveyService struct {
	surveyRepo       repository.SurveyRepo
	surveyExportRepo repository.SurveyExportRepo
}

func NewSurveyService(surveyRepo repository.SurveyRepo, surveyExportRepo repository.SurveyExportRepo) SurveyService {
	return &surveyService{surveyRepo, surveyExportRepo}
}

// func hashID(id uint) string {
// 	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", id))))
// }

func (service *surveyService) LogSurveys(param url.Values) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	page, limit, offset, _, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses pagination: %w", err), http.StatusBadRequest)
	}

	log, total, err := service.surveyRepo.LogSurveys(offset, limit, param)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	var totalPages int
	if limit == 1 {
		totalPages = 1
	} else {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}
	meta := map[string]interface{}{
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	}
	response := map[string]interface{}{
		"data": log,
		"meta": meta,
	}

	return utils.SendData(response)
}

func (service *surveyService) InsertLogSurveys(req map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	payload := payloads.Surveys{}

	err := utils.DynamicBind(req, &payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	data := models.Log_Surveys{}

	err = utils.DynamicBind(payload, &data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	user, err := service.surveyRepo.CheckRespondent(data.RespondentID)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	data.Keterangan = user.Name + " (responden) Telah melakukan pengisian survey"
	data.CreatedAt = utils.TimeNow()
	data.UpdatedAt = utils.TimeNow()
	data.IsRead = "false"
	data.LogLevel = "6"

	err = service.surveyRepo.InsertLogSurvey(data)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	return utils.SendData("Log pengisian survey berhasil diinput")
}

func (s *surveyService) GetDataSurvey(usr models.JwtCustomClaims, status *string, param url.Values) (*pb.ProxyResponse, error) {
	page, limit, offset, _, err := utils.SetPagination(param)
	if err != nil {
		return utils.SendError(fmt.Errorf("terjadi kesalahan saat memproses pagination: %w", err), http.StatusBadRequest)
	}

	respondent, err := s.surveyRepo.GetRespondentByID(usr.ID)
	if err != nil {
		return utils.SendError(fmt.Errorf("respondent tidak ditemukan: %w", err), http.StatusInternalServerError)
	}

	kecamatan, err := s.surveyRepo.GetKecamatanByID(respondent.KecamatanID)
	if err != nil {
		return utils.SendError(fmt.Errorf("kecamatan tidak ditemukan: %w", err), http.StatusInternalServerError)
	}

	kelurahan, err := s.surveyRepo.GetKelurahanByID(respondent.KelurahanID)
	if err != nil {
		return utils.SendError(fmt.Errorf("kelurahan tidak ditemukan: %w", err), http.StatusInternalServerError)
	}

	rw, err := s.surveyRepo.GetRwByKelurahanID(kelurahan.ID)
	if err != nil {
		return utils.SendError(fmt.Errorf("RW tidak ditemukan: %w", err), http.StatusInternalServerError)
	}

	roleName := strings.ToLower(respondent.Role.Name)
	var results []models.SurveyResults
	var total int64
	switch roleName {
	case "rt":
		surveys, err := s.surveyRepo.GetSurveysForRT(kecamatan.ID, kelurahan.ID, rw.ID, status)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		results, err = s.buildRTResults(surveys)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

	case "surveyor":
		surveys, err := s.surveyRepo.GetSurveysForSurveyor(usr.RespondentID, status)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		results = s.buildSimpleResults(surveys)

	default:
		surveys, totalCount, err := s.surveyRepo.GetSurveysForOther(respondent, kecamatan.ID, kelurahan.ID, rw.ID, status, offset, limit, param)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}
		results = s.buildSimpleResults(surveys)
		total = totalCount
	}

	var totalPages int
	if limit == 1 {
		totalPages = 1
	} else {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}
	meta := map[string]interface{}{
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	}
	response := map[string]interface{}{
		"data": results,
		"meta": meta,
	}

	return utils.SendData(response)
}

func (s *surveyService) buildRTResults(surveys []models.Surveyss) ([]models.SurveyResults, error) {
	var results []models.SurveyResults

	for _, survey := range surveys {
		flowDetail, err := s.surveyRepo.GetFlowDetailByID(survey.FlowDetailID)
		if err != nil {
			return nil, fmt.Errorf("flow detail tidak ditemukan (survey_id=%d): %w", survey.ID, err)
		}

		flowFields, err := s.surveyRepo.GetFlowFieldsByFlowDetailID(survey.FlowDetailID)
		if err != nil {
			return nil, err
		}

		formFieldIDs := extractFormFieldIDs(flowFields)
		totalPertanyaan, breakdown, err := s.calculateQuestions(survey.FlowDetailID, flowDetail, formFieldIDs)
		if err != nil {
			return nil, err
		}

		encodedID, err := utils.EncryptInt(int(survey.ID))
		if err != nil {
			return nil, err
		}

		results = append(results, models.SurveyResults{
			Name:            survey.Name,
			StartDate:       survey.StartDate,
			EndDate:         survey.EndDate,
			GetFlowDetail:   survey.FlowDetail,
			Status:          survey.Status,
			ListFlowDetail:  survey.FlowDetail,
			TotalPertanyaan: int(totalPertanyaan),
			Breakdown:       breakdown,
			ID:              encodedID,
		})
	}

	return results, nil
}

func (s *surveyService) buildSimpleResults(surveys []models.Surveyss) []models.SurveyResults {
	var results []models.SurveyResults
	for _, survey := range surveys {
		encodedID, err := utils.EncryptInt(int(survey.ID))
		if err != nil {
			return nil
		}
		results = append(results, models.SurveyResults{
			Name:            survey.Name,
			StartDate:       survey.StartDate,
			EndDate:         survey.EndDate,
			GetFlowDetail:   survey.FlowDetail,
			Status:          survey.Status,
			ListFlowDetail:  survey.FlowDetail,
			TotalPertanyaan: 1,
			ID:              encodedID,
		})
	}

	return results
}

func (s *surveyService) calculateQuestions(flowDetailID uint, flowDetail *models.FlowDetails, formFieldIDs []uint) (int64, string, error) {
	if flowDetail.StatusSection == 1 {
		return s.calculateMultipleSection(flowDetailID, formFieldIDs)
	}
	return s.calculateSoloSection(flowDetailID, formFieldIDs)
}

func (s *surveyService) calculateMultipleSection(flowDetailID uint, formFieldIDs []uint) (int64, string, error) {
	breakdownFields, err := s.surveyRepo.GetBreakdownFlowFields(flowDetailID, nil)
	if err != nil {
		return 0, "", err
	}

	seen := make(map[uint]bool)
	var uniqueSections []models.FlowFields
	for _, f := range breakdownFields {
		if f.SectionID != nil && !seen[*f.SectionID] {
			seen[*f.SectionID] = true
			uniqueSections = append(uniqueSections, f)
		}
	}
	countSection := int64(len(uniqueSections))

	var allChildIDs []uint
	for _, section := range uniqueSections {
		firstFlow, err := s.surveyRepo.GetFirstFlowField(flowDetailID, section.SectionID, true)
		if err != nil {
			return 0, "", err
		}
		lastFlow, err := s.surveyRepo.GetLastFlowField(flowDetailID, section.SectionID)
		if err != nil {
			return 0, "", err
		}
		childFields, err := s.surveyRepo.GetFlowFieldsBetween(firstFlow.ID, lastFlow.ID, flowDetailID, section.SectionID)
		if err != nil {
			return 0, "", err
		}
		for _, cf := range childFields {
			allChildIDs = append(allChildIDs, cf.FormFieldID)
		}
	}

	count, err := s.surveyRepo.CountFormFields(formFieldIDs, allChildIDs)
	if err != nil {
		return 0, "", err
	}

	return count + countSection, "multiple", nil
}

func (s *surveyService) calculateSoloSection(flowDetailID uint, formFieldIDs []uint) (int64, string, error) {
	firstBreakdown, err := s.surveyRepo.GetFirstFlowField(flowDetailID, nil, true)
	if err != nil {
		// Tidak ada breakdown — hitung semua
		count, err := s.surveyRepo.CountFormFields(formFieldIDs, nil)
		return count, "solo", err
	}

	lastFlow, err := s.surveyRepo.GetLastFlowField(flowDetailID, nil)
	if err != nil {
		return 0, "", err
	}

	childFields, err := s.surveyRepo.GetFlowFieldsBetween(firstBreakdown.ID, lastFlow.ID, flowDetailID, nil)
	if err != nil {
		return 0, "", err
	}

	excludeIDs := extractFormFieldIDs(childFields)
	count, err := s.surveyRepo.CountFormFields(formFieldIDs, excludeIDs)
	return count, "solo", err
}

func extractFormFieldIDs(fields []models.FlowFields) []uint {
	ids := make([]uint, 0, len(fields))
	for _, f := range fields {
		ids = append(ids, f.FormFieldID)
	}
	return ids
}

func filterRespondentsByID(respondents []models.SurveyRespondents, respondentID uint) []models.SurveyRespondents {
	var filtered []models.SurveyRespondents
	for _, r := range respondents {
		if r.RespondentID == respondentID {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func (s *surveyService) SurveyActivitiesExport(req ExportRequest) (*ExportResult, error) {
	var respondentIDs []uint

	// hd := hashids.NewData()
	// h, err := hashids.NewWithData(hd)
	// if err != nil {
	// 	return nil, fmt.Errorf("Survey ID Tidak Valid Atau Dimanipulasi")
	// }

	// decodedIDs, err := h.DecodeWithError(req.SurveyID)
	// if err != nil || len(decodedIDs) == 0 {
	// 	return nil, fmt.Errorf("Survey ID Tidak Valid Atau Dimanipulasi")
	// }

	surveyId, err := utils.ToInt64(req.SurveyID)
	if err != nil {
		return nil, fmt.Errorf("Data Survey Tidak Valid")
	}

	// decodedWilayahIDs, err := h.DecodeWithError(req.WilayahID)
	// if err != nil || len(decodedIDs) == 0 {
	// 	return nil, fmt.Errorf("Survey ID Tidak Valid Atau Dimanipulasi")
	// }

	wilayahId, err := utils.ToInt64(req.WilayahID)
	if err != nil {
		return nil, fmt.Errorf("Wilayah ID Tidak Valid")
	}

	switch req.Wilayah {
	case "kecamatan":
		kelurahanIDs, err := s.surveyExportRepo.GetKelurahanIDsByKecamatan(uint(wilayahId))
		if err != nil {
			return nil, fmt.Errorf("get kelurahan by kecamatan: %w", err)
		}
		respondentIDs, err = s.surveyExportRepo.GetRespondentsByKelurahanIDs(kelurahanIDs)
		if err != nil {
			return nil, fmt.Errorf("get respondents by kelurahan: %w", err)
		}

	case "kelurahan":
		rwIDs, err := s.surveyExportRepo.GetRwIDsByKelurahan(uint(wilayahId))
		if err != nil {
			return nil, fmt.Errorf("get rw by kelurahan: %w", err)
		}
		respondentIDs, err = s.surveyExportRepo.GetRespondentsByRwIDs(rwIDs)
		if err != nil {
			return nil, fmt.Errorf("get respondents by rw: %w", err)
		}

	case "rw":
		rtIDs, err := s.surveyExportRepo.GetRtIDsByRw(uint(wilayahId))
		if err != nil {
			return nil, fmt.Errorf("get rt by rw: %w", err)
		}
		respondentIDs, err = s.surveyExportRepo.GetRespondentsByRtIDs(rtIDs)
		if err != nil {
			return nil, fmt.Errorf("get respondents by rt: %w", err)
		}
	case "rt":
		rtIDs, err := s.surveyExportRepo.GetRtIDsByRt(uint(wilayahId))
		if err != nil {
			return nil, fmt.Errorf("get rt by id: %w", err)
		}
		respondentIDs, err = s.surveyExportRepo.GetRespondentsByRtIDs(rtIDs)
		if err != nil {
			return nil, fmt.Errorf("get respondents by rt: %w", err)
		}
	default:
	}

	survey, err := s.surveyExportRepo.GetSurveyWithResponses(uint(surveyId), respondentIDs)
	if err != nil {
		return nil, fmt.Errorf("get survey: %w", err)
	}

	var fieldIDs []uint
	for _, f := range survey.FlowDetail.Form.Fields {
		fieldIDs = append(fieldIDs, f.ID)
	}
	answerOptions, err := s.surveyExportRepo.GetFormAnswerOptions(fieldIDs)
	if err != nil {
		return nil, fmt.Errorf("get answer options: %w", err)
	}

	xlsxBytes, err := excel.BuildSurveyExcel(survey, answerOptions)
	if err != nil {
		return nil, fmt.Errorf("build excel: %w", err)
	}

	kecamatanName, kelurahanName, rwName, rtName := "", "", "", ""
	switch req.Wilayah {
	case "kecamatan":
		kec, err := s.surveyExportRepo.GetKecamatanByID(uint(wilayahId))
		if err == nil {
			kecamatanName = kec.SubDistrictName
		}
	case "kelurahan":
		kel, err := s.surveyExportRepo.GetKelurahanByID(uint(wilayahId))
		if err == nil {
			kecamatanName = kel.Kecamatan.SubDistrictName
			kelurahanName = kel.VillageName
		}
	case "rw":
		rw, err := s.surveyExportRepo.GetRwByID(uint(wilayahId))
		if err == nil {
			kecamatanName = rw.Kelurahan.Kecamatan.SubDistrictName
			kelurahanName = rw.Kelurahan.VillageName
			rwName = rw.NamaRw
		}
	case "rt":
		rt, err := s.surveyExportRepo.GetRtByID(uint(wilayahId))
		if err == nil {
			kecamatanName = rt.Rw.Kelurahan.Kecamatan.SubDistrictName
			kelurahanName = rt.Rw.Kelurahan.VillageName
			rwName = rt.Rw.NamaRw
			rtName = rt.NamaRt
		}
	}

	filename := excel.GenerateFilename(survey.Name, req.Wilayah, kecamatanName, kelurahanName, rwName, rtName)

	return &ExportResult{
		Filename: filename,
		Data:     xlsxBytes,
	}, nil
}
