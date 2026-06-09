package service

import (
	"backend/reportapi/models"
	"backend/reportapi/repository"
	"backend/siccore/pb"
	"backend/userapi/utils"
	"crypto/md5"
	"fmt"
	"math"
	"net/http"
	"net/url"
)

type SurveyService interface {
	SurveyActivities(slug map[string]interface{}) (*pb.ProxyResponse, error)
	LogSurveys(param url.Values) (*pb.ProxyResponse, error)
}

type surveyService struct {
	surveyRepo repository.SurveyRepo
}

func NewSurveyService(surveyRepo repository.SurveyRepo) SurveyService {
	return &surveyService{surveyRepo}
}

func hashID(id uint) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", id))))
}

func (service *surveyService) SurveyActivities(slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()
	roleName, _ := slug["role_name"].(string)
	respondentID, _ := slug["respondent_id"].(uint)
	kecamatanID, _ := slug["kecamatan_id"].(uint)
	kelurahanID, _ := slug["kelurahan_id"].(uint)
	rwID, _ := slug["rw_id"].(uint)

	var statusPtr *string
	if s, ok := slug["status"].(string); ok && s != "" {
		statusPtr = &s
	}

	wilayah := models.WilayahFilter{
		KecamatanID: kecamatanID,
		KelurahanID: kelurahanID,
		RwID:        rwID,
	}

	var results []models.SurveyResult

	switch roleName {
	case "rt":
		surveys, err := service.surveyRepo.GetSurveysByWilayahAndStatus(models.GetSurveyParams{
			Status:        statusPtr,
			WilayahFilter: wilayah,
		})
		if err != nil {
			return utils.SendData(nil)
		}

		for _, s := range surveys {
			totalQ, breakdown, err := service.computeFormQuestion(s.FlowDetailID)
			if err != nil {
				continue
			}
			results = append(results, models.SurveyResult{
				Name:              s.Name,
				SurveyRespondents: s.SurveyRespondents,
				StartDate:         s.StartDate,
				EndDate:           s.EndDate,
				GetFlowDetail:     s.GetFlowDetail,
				Status:            s.Status,
				ListFlowDetail:    s.ListFlowDetail,
				TotalPertanyaan:   totalQ,
				Breakdown:         breakdown,
				ID:                hashID(s.ID),
			})
		}

	case "surveyor":
		surveys, err := service.surveyRepo.GetSurveysBySurveyor(respondentID, statusPtr)
		if err != nil {
			return utils.SendData(nil)
		}
		for _, s := range surveys {
			results = append(results, models.SurveyResult{
				Name:              s.Name,
				SurveyRespondents: s.SurveyRespondents,
				StartDate:         s.StartDate,
				EndDate:           s.EndDate,
				GetFlowDetail:     s.GetFlowDetail,
				Status:            s.Status,
				ListFlowDetail:    s.ListFlowDetail,
				TotalPertanyaan:   1,
				ID:                hashID(s.ID),
			})
		}

	default:
		roleFilter := models.DataRespondentFilter{
			Role:        roleName,
			KecamatanID: kecamatanID,
			KelurahanID: kelurahanID,
			RwID:        rwID,
		}
		surveys, err := service.surveyRepo.GetSurveysByDataRespondents(roleFilter, wilayah, statusPtr)
		if err != nil {
			return utils.SendData(nil)
		}
		for _, s := range surveys {
			results = append(results, models.SurveyResult{
				Name:              s.Name,
				SurveyRespondents: s.SurveyDataRespondents,
				StartDate:         s.StartDate,
				EndDate:           s.EndDate,
				GetFlowDetail:     s.GetFlowDetail,
				Status:            s.Status,
				ListFlowDetail:    s.ListFlowDetail,
				TotalPertanyaan:   1,
				ID:                hashID(s.ID),
			})
		}
	}

	return utils.SendData(results)
}

func (service *surveyService) computeFormQuestion(flowDetailID uint) (int64, string, error) {
	flowDetail, err := service.surveyRepo.GetFlowDetail(flowDetailID)
	if err != nil {
		return 0, "", err
	}

	allFields, err := service.surveyRepo.GetFlowFields(flowDetailID)
	if err != nil {
		return 0, "", err
	}

	var allFormFieldIDs []uint
	for _, f := range allFields {
		allFormFieldIDs = append(allFormFieldIDs, f.FormFieldID)
	}

	if flowDetail.StatusSection == 1 {
		sectionMap := map[uint]bool{}
		var sections []uint
		for _, f := range allFields {
			if f.Breakdown && !sectionMap[f.SectionID] {
				sectionMap[f.SectionID] = true
				sections = append(sections, f.SectionID)
			}
		}
		countSection := int64(len(sections))

		var excludeIDs []uint
		for _, sectionID := range sections {
			sectionFields, err := service.surveyRepo.GetFlowFieldsInSection(flowDetailID, sectionID)
			if err != nil {
				continue
			}
			var firstBreakdownIdx, lastIdx int = -1, len(sectionFields) - 1
			for i, f := range sectionFields {
				if f.Breakdown && firstBreakdownIdx == -1 {
					firstBreakdownIdx = i
				}
			}
			if firstBreakdownIdx >= 0 {
				for i := firstBreakdownIdx; i <= lastIdx; i++ {
					excludeIDs = append(excludeIDs, sectionFields[i].FormFieldID)
				}
			}
		}

		count, err := service.surveyRepo.CountFormQuestions(allFormFieldIDs, excludeIDs)
		if err != nil {
			return 0, "", err
		}
		return count + countSection, "multiple", nil

	} else {
		var firstBreakdownIdx int = -1
		for i, f := range allFields {
			if f.Breakdown {
				firstBreakdownIdx = i
				break
			}
		}
		if firstBreakdownIdx >= 0 {
			var excludeIDs []uint
			for i := firstBreakdownIdx; i < len(allFields); i++ {
				excludeIDs = append(excludeIDs, allFields[i].FormFieldID)
			}
			count, err := service.surveyRepo.CountFormQuestions(allFormFieldIDs, excludeIDs)
			if err != nil {
				return 0, "", err
			}
			return count, "solo", nil
		}

		count, err := service.surveyRepo.CountFormQuestions(allFormFieldIDs, nil)
		if err != nil {
			return 0, "", err
		}
		return count, "solo", nil
	}
}

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
