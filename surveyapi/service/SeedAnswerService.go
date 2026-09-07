package service

import (
	"backend/siccore/pb"
	"backend/surveyapi/assets"
	"backend/surveyapi/models"
	"backend/surveyapi/payloads"
	"backend/surveyapi/repository"
	"backend/surveyapi/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
)

type SeedAnswerService interface {
	SeedSurveyAnswer(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type seedAnswerService struct {
	surveyRepo        repository.SurveyRepo
	manajemenAlurRepo repository.ManajemenAlurRepo
	surveyService     SurveyService
}

func NewSeedAnswerService(
	surveyRepo repository.SurveyRepo,
	manajemenAlurRepo repository.ManajemenAlurRepo,
	surveyService SurveyService,
) SeedAnswerService {
	return &seedAnswerService{
		surveyRepo:        surveyRepo,
		manajemenAlurRepo: manajemenAlurRepo,
		surveyService:     surveyService,
	}
}

var (
	seedImagePool = []string{
		"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAASwAAAEsAQMAAABDsxw2AAAAAXNSR0IB2cksfwAAAAlwSFlzAAALEwAACxMBAJqcGAAAAANQTFRFAAD/injSVwAAADxJREFUeJztyjEBAAAMAqDZv/Qq6A83uUo0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdO0vT0NmwEtZkyx0wAAAABJRU5ErkJggg==",
		"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAASwAAAEsAQMAAABDsxw2AAAAAXNSR0IB2cksfwAAAAlwSFlzAAALEwAACxMBAJqcGAAAAANQTFRFAP8ANF7AqAAAADxJREFUeJztyjEBAAAMAqDZv/Qq6A83uUo0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdO0vT0NmwEtZkyx0wAAAABJRU5ErkJggg==",
		"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAASwAAAEsAQMAAABDsxw2AAAAAXNSR0IB2cksfwAAAAlwSFlzAAALEwAACxMBAJqcGAAAAANQTFRF/wAAGeIJNwAAADxJREFUeJztyjEBAAAMAqDZv/Qq6A83uUo0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdM0TdO0vT0NmwEtZkyx0wAAAABJRU5ErkJggg==",
	}
	seedBandungLatMin = -6.9697480
	seedBandungLatMax = -6.8368542
	seedBandungLngMin = 107.5453618
	seedBandungLngMax = 107.7395207

	seedFallbackLat = -6.9218457
	seedFallbackLng = 107.6070833
)

func (s *seedAnswerService) SeedSurveyAnswer(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.SeedSurveyAnswerRequest
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
			return utils.SendError(errors.New(utils.TranslateError(err)), http.StatusBadRequest)
		}
	}

	// 1. Ambil survey & form_id-nya
	survey, err := s.surveyRepo.GetSurveyById(payload.SurveyID)
	if err != nil {
		return utils.SendError(errors.New("Survey tidak ditemukan"), http.StatusNotFound)
	}

	// 2. Ambil semua pertanyaan (semua section sekaligus, sectionID=0 artinya semua)
	rawNodes, err := s.manajemenAlurRepo.GetRawNodesForPreview(int(survey.FlowDetailID), 0)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}
	if len(rawNodes) == 0 {
		return utils.SendError(errors.New("Alur survey tidak punya pertanyaan"), http.StatusNotFound)
	}

	// 3. Group berdasarkan section_id (0 = tanpa section)
	sectionGroups := make(map[int64][]models.RawNodeData)
	for _, node := range rawNodes {
		var secID int64 = 0
		if node.SectionId != nil {
			secID = int64(*node.SectionId)
		}
		sectionGroups[secID] = append(sectionGroups[secID], node)
	}

	// 4. Siapkan hashid encoder (sama persis seperti production)
	hd := hashids.NewData()
	hd.Salt = os.Getenv("HASHID_SALT")
	hd.MinLength = 24
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	surveyCode, err := h.Encode([]int{int(payload.SurveyID)})
	if err != nil {
		return utils.SendError(errors.New("Gagal encode kode survey"), http.StatusInternalServerError)
	}

	fakeUsr := models.JwtCustomClaims{
		RespondentID: payload.RespondentID,
	}

	totalTerisi := 0
	var sectionErrors []string

	// 5. Submit per-section — reuse persis fungsi SubmitSurveyAnswers yang sudah ada
	for sectionID, nodes := range sectionGroups {
		sectionCode, err := h.Encode([]int{int(sectionID)})
		if err != nil {
			sectionErrors = append(sectionErrors, fmt.Sprintf("section %d: gagal encode", sectionID))
			continue
		}

		answers := s.buildAnswersForNodes(nodes)
		if len(answers) == 0 {
			continue
		}

		submitPayload := payloads.SubmitSurveyPayload{Answers: answers}

		reqBytes, _ := json.Marshal(submitPayload)
		var reqMap map[string]interface{}
		json.Unmarshal(reqBytes, &reqMap)

		slugMap := map[string]interface{}{
			"survey_code":  surveyCode,
			"section_code": sectionCode,
		}

		resp, err := s.surveyService.SubmitSurveyAnswers(ctx, reqMap, fakeUsr, url.Values{}, slugMap)
		if err != nil {
			sectionErrors = append(sectionErrors, fmt.Sprintf("section %d: %v", sectionID, err))
			continue
		}
		if resp != nil && !resp.Success {
			sectionErrors = append(sectionErrors, fmt.Sprintf("section %d: %s", sectionID, resp.Message))
			continue
		}

		totalTerisi += len(answers)
	}

	if totalTerisi == 0 {
		errMsg := "Tidak ada jawaban yang berhasil disimpan"
		if len(sectionErrors) > 0 {
			errMsg += ": " + fmt.Sprintf("%v", sectionErrors)
		}
		return utils.SendError(errors.New(errMsg), http.StatusInternalServerError)
	}

	resp := map[string]interface{}{
		"total_field_terisi": totalTerisi,
		"errors":             sectionErrors,
	}

	return utils.SendData(resp, "Seeder berhasil dijalankan")
}

// buildAnswersForNodes: generate jawaban palsu untuk satu section, format persis AnswerPayload asli
func (s *seedAnswerService) buildAnswersForNodes(nodes []models.RawNodeData) []payloads.AnswerPayload {
	var answers []payloads.AnswerPayload

	for _, node := range nodes {
		ans := payloads.AnswerPayload{
			QuestionID: node.FormFieldId,
			Type:       node.Template,
		}

		switch node.Template {
		case "long-answer":
			text := generateLoremIpsum(rand.Intn(15) + 5)
			ans.ValueString = &text

		case "number":
			num := int64(rand.Intn(301)) // 0 - 300
			ans.ValueNumber = &num

		case "multiple-choices":
			optionIDs, err := s.manajemenAlurRepo.GetAnswerOptionsByQuestionID(node.FormFieldId)
			if err != nil || len(optionIDs) == 0 {
				continue // skip pertanyaan ini kalau tidak ada opsi
			}
			selected := optionIDs[rand.Intn(len(optionIDs))]
			ans.ValueOptionID = &selected

		case "maps":
			totalPoints := rand.Intn(15) + 1 // random 1 - 15 titik

			points := make([]payloads.MapCoordinatePayload, 0, totalPoints)
			for i := 0; i < totalPoints; i++ {
				lat, lng := generateRandomPointInBandung()
				points = append(points, payloads.MapCoordinatePayload{Lat: lat, Lng: lng})
			}
			ans.ValueMaps = points

		case "image-template":
			if len(seedImagePool) == 0 {
				continue
			}
			qty := 1
			if node.ImageQuantity != nil {
				if q, err := parseIntSafe(*node.ImageQuantity); err == nil && q > 0 {
					qty = q
				}
			}
			if qty > len(seedImagePool) {
				qty = len(seedImagePool)
			}
			indices := rand.Perm(len(seedImagePool))[:qty]
			var images []string
			for _, idx := range indices {
				images = append(images, seedImagePool[idx])
			}
			ans.ValueImages = images

		default:
			continue
		}

		answers = append(answers, ans)
	}

	return answers
}

// generateRandomPointInBandung: random titik di dalam bounding box, divalidasi harus di dalam polygon asli Kota Bandung
func generateRandomPointInBandung() (float64, float64) {
	maxAttempts := 200
	for i := 0; i < maxAttempts; i++ {
		lat := seedBandungLatMin + rand.Float64()*(seedBandungLatMax-seedBandungLatMin)
		lng := seedBandungLngMin + rand.Float64()*(seedBandungLngMax-seedBandungLngMin)

		if utils.PointInPolygon(lat, lng, assets.BandungPolygon) {
			return lat, lng
		}
	}
	// fallback kalau 200x gagal (harusnya nyaris tidak pernah terjadi)
	return seedFallbackLat, seedFallbackLng
}

func parseIntSafe(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

var loremWords = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur",
	"adipiscing", "elit", "sed", "do", "eiusmod", "tempor",
	"incididunt", "ut", "labore", "et", "dolore", "magna", "aliqua",
	"perlu", "digalakan", "ronda", "malam", "fasilitas", "umum",
	"lingkungan", "kondisi", "cukup", "baik", "perbaikan", "drainase",
}

func generateLoremIpsum(wordCount int) string {
	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = loremWords[rand.Intn(len(loremWords))]
	}
	sentence := ""
	for i, w := range words {
		if i > 0 {
			sentence += " "
		}
		sentence += w
	}
	return capitalizeFirst(sentence) + "."
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] = r[0] - 32
	}
	return string(r)
}
