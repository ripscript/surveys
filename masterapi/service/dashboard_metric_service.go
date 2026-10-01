package service

import (
	"backend/masterapi/customValidator"
	"backend/masterapi/models"
	"backend/masterapi/payloads"
	"backend/masterapi/repository"
	"backend/masterapi/response"
	"backend/masterapi/utils"
	"backend/siccore/pb"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/speps/go-hashids/v2"
	"gorm.io/gorm"
)

type DashboardMetricService interface {
	GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetCategoryOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetStatuses(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
	GetMetricOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)

	GetPemetaanMetrikSurveyList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error)
}

type dashboardMetricService struct {
	dbMaster    *gorm.DB
	repo        repository.DashboardMetricRepository
	statusRepo  repository.DashboardMetricStatusRepository
	mappingRepo repository.DashboardMetricMappingRepository
}

func NewDashboardMetricService(
	dbMaster *gorm.DB,
	repo repository.DashboardMetricRepository,
	statusRepo repository.DashboardMetricStatusRepository,
	mappingRepo repository.DashboardMetricMappingRepository,
) DashboardMetricService {
	return &dashboardMetricService{dbMaster: dbMaster, repo: repo, statusRepo: statusRepo, mappingRepo: mappingRepo}
}

// ---------------------------------------------------------------------
// CREATE — metrik + statuses dalam satu transaksi.
// ---------------------------------------------------------------------
func (s *dashboardMetricService) Create(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.CreateDashboardMetricRequest
	if err := utils.DynamicBind(req, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	validate := validator.New()
	if err := customValidator.RegisterMetricKeyFormat(validate); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	if err := validate.Struct(payload); err != nil {
		for _, verr := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(verr)), http.StatusBadRequest)
		}
	}

	if err := customValidator.UpdateDashboardMetricStatusesValidator(payload.ExpectedTemplate, payload.Statuses); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	existing, err := s.repo.FindByMetricKey(ctx, payload.MetricKey)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	if existing != nil {
		return utils.SendError(fmt.Errorf("metric_key '%s' sudah dipakai", payload.MetricKey), http.StatusBadRequest)
	}

	var created *models.DashboardMetric
	err = s.dbMaster.Transaction(func(tx *gorm.DB) error {
		m := &models.DashboardMetric{
			MetricKey:        payload.MetricKey,
			Label:            payload.Label,
			ExpectedTemplate: payload.ExpectedTemplate,
			Category:         payload.Category,
			IsLocked:         false,
		}
		if err := s.repo.Create(tx, ctx, m); err != nil {
			return err
		}

		statuses := make([]models.DashboardMetricStatus, 0, len(payload.Statuses))
		for _, sp := range payload.Statuses {
			statuses = append(statuses, models.DashboardMetricStatus{
				DashboardMetricID: m.ID,
				StatusKey:         sp.StatusKey,
				Label:             sp.Label,
				Sequence:          sp.Sequence,
				IsActive:          true,
			})
		}
		if err := s.statusRepo.BulkCreate(tx, ctx, statuses); err != nil {
			return err
		}

		created = m
		return nil
	})
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	detail, err := s.repo.FindDetailByID(ctx, created.ID)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	return utils.SendData(toDashboardMetricResponse(detail), "Berhasil menambahkan metrik")
}

// ---------------------------------------------------------------------
// UPDATE — label/category selalu bisa diubah; metric_key/expected_template
// immutable; statuses di-diff (insert baru, update existing, delete-if-unused
// / deactivate-if-used), semuanya dalam satu transaksi.
// ---------------------------------------------------------------------
func (s *dashboardMetricService) Update(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	var payload payloads.UpdateDashboardMetricRequest
	if err := utils.DynamicBind(req, &payload); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}
	validate := validator.New()
	if err := customValidator.RegisterMetricKeyFormat(validate); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if err := validate.Struct(payload); err != nil {
		for _, verr := range err.(validator.ValidationErrors) {
			return utils.SendError(errors.New(utils.TranslateError(verr)), http.StatusBadRequest)
		}
	}

	id, err := parseIDFromSlug(slug)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	existing, err := s.repo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("metrik tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	// expected_template diambil dari data existing (immutable), dipakai untuk validasi statuses
	if err := customValidator.UpdateDashboardMetricStatusesValidator(existing.ExpectedTemplate, payload.Statuses); err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	// is_locked: metric_key & expected_template memang sudah tidak ada di payload update,
	// jadi tidak perlu dicek di sini. Yang perlu dicek untuk is_locked:
	// larangan hapus/nonaktifkan status yang SUDAH ADA (lihat diffStatuses di bawah).

	err = s.dbMaster.Transaction(func(tx *gorm.DB) error {
		existing.Label = payload.Label
		existing.Category = payload.Category
		if err := s.repo.Update(tx, ctx, existing); err != nil {
			return err
		}
		return s.diffAndApplyStatuses(tx, ctx, existing, payload.Statuses)
	})
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	detail, err := s.repo.FindDetailByID(ctx, id)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	return utils.SendData(toDashboardMetricResponse(detail), "Berhasil memperbarui metrik")
}

// diffAndApplyStatuses membandingkan status existing vs payload:
//   - ada di payload dengan ID -> update label/sequence (status_key TIDAK bisa diubah, immutable)
//   - ada di payload tanpa ID  -> insert baru
//   - ada di existing tapi tidak ada di payload -> coba delete; kalau masih
//     dipakai di dashboard_option_status_mappings, deactivate saja (bukan error)
//   - is_locked=true: status existing yang di-remove dari payload TIDAK boleh
//     dihapus/dinonaktifkan sama sekali -> return error (dashboard hardcode
//     butuh status itu tetap ada & aktif)
func (s *dashboardMetricService) diffAndApplyStatuses(tx *gorm.DB, ctx context.Context, metric *models.DashboardMetric, incoming []payloads.DashboardMetricStatusPayload) error {
	incomingByID := make(map[int64]payloads.DashboardMetricStatusPayload)
	for _, sp := range incoming {
		if sp.ID > 0 {
			incomingByID[sp.ID] = sp
		}
	}

	// existing.Statuses sudah ter-preload dari FindDetailByID
	for _, existingStatus := range metric.Statuses {
		if sp, ok := incomingByID[existingStatus.ID]; ok {
			// update label/sequence saja; status_key immutable
			existingStatus.Label = sp.Label
			existingStatus.Sequence = sp.Sequence
			if err := s.statusRepo.Update(tx, ctx, &existingStatus); err != nil {
				return err
			}
			continue
		}

		// tidak ada di payload -> mau dihapus
		if metric.IsLocked {
			return fmt.Errorf("status '%s' tidak dapat dihapus/diubah karena metrik ini terkunci (is_locked)", existingStatus.StatusKey)
		}

		usage, err := s.statusRepo.CountUsage(ctx, existingStatus.ID)
		if err != nil {
			return err
		}
		if usage > 0 {
			if err := s.statusRepo.Deactivate(tx, ctx, existingStatus.ID); err != nil {
				return err
			}
			continue
		}
		if err := s.statusRepo.Delete(tx, ctx, existingStatus.ID); err != nil {
			return err
		}
	}

	// insert status baru (tanpa ID di payload)
	newStatuses := make([]models.DashboardMetricStatus, 0)
	for _, sp := range incoming {
		if sp.ID > 0 {
			continue
		}
		newStatuses = append(newStatuses, models.DashboardMetricStatus{
			DashboardMetricID: metric.ID,
			StatusKey:         sp.StatusKey,
			Label:             sp.Label,
			Sequence:          sp.Sequence,
			IsActive:          true,
		})
	}
	return s.statusRepo.BulkCreate(tx, ctx, newStatuses)
}

// ---------------------------------------------------------------------
// DELETE metrik — diblokir kalau is_locked ATAU masih ada mapping ke form.
// ---------------------------------------------------------------------
func (s *dashboardMetricService) Delete(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	id, err := parseIDFromSlug(slug)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("metrik tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	if m.IsLocked {
		return utils.SendError(errors.New("metrik ini tidak dapat dihapus, karena digunakan di dashboard"), http.StatusConflict)
	}

	mappingCount, err := s.repo.CountMappingsByMetricID(ctx, id)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	if mappingCount > 0 {
		return utils.SendError(errors.New("metrik ini tidak dapat dihapus, karena masih memiliki pemetaan ke form. Hapus pemetaannya terlebih dahulu"), http.StatusBadRequest)
	}

	// statuses ikut terhapus via FK CASCADE (dashboard_metric_statuses.dashboard_metric_id)
	if err := s.repo.Delete(s.dbMaster, ctx, id); err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	return utils.SendData(nil, "berhasil menghapus metric")
}

// ---------------------------------------------------------------------
// GET LIST / DETAIL / CATEGORY OPTIONS / STATUSES
// ---------------------------------------------------------------------
func (service *dashboardMetricService) GetList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	payload := payloads.DatatablePayload{
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
		payload.Limit = 25
	}

	data, totalData, err := service.repo.GetList(ctx, payload)
	if err != nil {
		return utils.SendError(err, http.StatusInternalServerError)
	}

	items := make([]response.DashboardMetricResponse, 0, len(data))
	for _, m := range data {
		items = append(items, toDashboardMetricResponse(&m))
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(payload.Limit)))

	result := map[string]interface{}{
		"data": items,
		"meta": map[string]interface{}{
			"total":      totalData,
			"page":       payload.Page,
			"limit":      payload.Limit,
			"totalPages": totalPages,
		},
	}

	return utils.SendData(result, "Berhasil mengambil list metrik")
}

func (s *dashboardMetricService) Detail(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	id, err := parseIDFromSlug(slug)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	m, err := s.repo.FindDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SendError(errors.New("metrik tidak ditemukan"), http.StatusNotFound)
		}
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}
	return utils.SendData(toDashboardMetricResponse(m), "data metrik berhasil diambil")
}

func (s *dashboardMetricService) GetStatuses(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	id, err := parseIDFromSlug(slug)
	if err != nil {
		return utils.SendError(err, http.StatusBadRequest)
	}

	statusParam := param.Get("is_active")

	var rawIDs []string
	if len(param["id[]"]) > 0 {
		rawIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		rawIDs = param["id"]
	}

	bypassIDs := make(map[int64]bool, len(rawIDs))
	for _, raw := range rawIDs {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			bypassIDs[parsed] = true
		}
	}

	statuses, err := s.statusRepo.FindByMetricID(ctx, id)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	items := make([]response.DashboardMetricStatusResponse, 0, len(statuses))
	for _, st := range statuses {
		matchesFilter := true
		switch statusParam {
		case "true":
			matchesFilter = st.IsActive
		case "false":
			matchesFilter = !st.IsActive
		}

		if !matchesFilter && !bypassIDs[st.ID] {
			continue
		}

		items = append(items, response.DashboardMetricStatusResponse{
			ID: st.ID, StatusKey: st.StatusKey, Label: st.Label, Sequence: st.Sequence, IsActive: st.IsActive,
		})
	}
	return utils.SendData(items, "")
}

func (s *dashboardMetricService) GetCategoryOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var rawValues []string
	if len(param["values[]"]) > 0 {
		rawValues = param["values[]"]
	}

	categories, totalData, err := s.repo.GetCategoryOptions(ctx, param.Get("q"), rawValues, page, limit)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	options := make([]response.StringOptionItem, 0, len(categories))
	for _, c := range categories {
		options = append(options, response.StringOptionItem{Value: c, Label: utils.ToCamelCase(c)})
	}

	result := response.StringOptionsResponse{
		Options: options,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     int64(page*limit) < totalData,
		},
	}
	return utils.SendData(result, "")
}

// ---------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------
func parseIDFromSlug(slug map[string]interface{}) (int64, error) {
	idStr, ok := slug["id"].(string)
	if !ok {
		return 0, errors.New("Invalid ID in slug")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.New("gagal parse string ke int64")
	}
	return id, nil
}

func toDashboardMetricResponse(m *models.DashboardMetric) response.DashboardMetricResponse {
	statuses := make([]response.DashboardMetricStatusResponse, 0, len(m.Statuses))
	for _, st := range m.Statuses {
		statuses = append(statuses, response.DashboardMetricStatusResponse{
			ID: st.ID, StatusKey: st.StatusKey, Label: st.Label, Sequence: st.Sequence, IsActive: st.IsActive,
		})
	}
	return response.DashboardMetricResponse{
		ID:               m.ID,
		MetricKey:        m.MetricKey,
		Label:            m.Label,
		ExpectedTemplate: m.ExpectedTemplate,
		Category:         m.Category,
		IsLocked:         m.IsLocked,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		Statuses:         statuses,
	}
}

func (s *dashboardMetricService) GetMetricOptions(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 1000
	}

	var rawIDs []string
	if len(param["id[]"]) > 0 {
		rawIDs = param["id[]"]
	} else if len(param["id"]) > 0 {
		rawIDs = param["id"]
	}
	var parsedIDs []int64
	for _, raw := range rawIDs {
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
			parsedIDs = append(parsedIDs, id)
		}
	}

	expectedTemplate := param.Get("expected_template")
	if expectedTemplate != "" && expectedTemplate != "number" && expectedTemplate != "multiple-choices" {
		return utils.SendError(errors.New("expected_template tidak valid"), http.StatusBadRequest)
	}

	_req := payloads.DashboardMetricOptionsPayload{
		Q:                param.Get("q"),
		Page:             page,
		Limit:            limit,
		IDs:              parsedIDs,
		ExpectedTemplate: expectedTemplate,
	}

	data, totalData, err := s.repo.GetMetricOptions(ctx, _req)
	if err != nil {
		return utils.SendError(errors.New("Terjadi kesalahan pada server, silahkan coba lagi nanti"), http.StatusInternalServerError)
	}

	currentTotalLoaded := (page-1)*limit + len(data)
	hasMore := int64(currentTotalLoaded) < totalData

	responseData := response.OptionsResponse{
		Options: data,
		Meta: response.PaginationMeta{
			CurrentPage: page,
			PerPage:     limit,
			Total:       totalData,
			HasMore:     hasMore,
		},
	}

	return utils.SendData(responseData, "Berhasil mengambil opsi metrik")
}

func (service *dashboardMetricService) GetPemetaanMetrikSurveyList(ctx context.Context, req map[string]interface{}, usr models.JwtCustomClaims, param url.Values, slug map[string]interface{}) (*pb.ProxyResponse, error) {
	defer utils.GeneralRecover()

	search := param.Get("search")
	page, _ := strconv.Atoi(param.Get("page"))
	limit, _ := strconv.Atoi(param.Get("limit"))
	orderBy := param.Get("order_by")
	orderDir := param.Get("order_dir")

	status := param.Get("status")

	payload := payloads.PemetaanMetrikSurveyDatatablePayload{
		Search:   search,
		Page:     page,
		Limit:    limit,
		OrderBy:  orderBy,
		OrderDir: orderDir,
		Status:   status,
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

	data, totalData, err := service.repo.GetPemetaanMetrikSurveyList(payload)
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
		surveyId := []int{int(data[i].ID)}

		surveyCode, err := h.Encode(surveyId)
		if err != nil {
			return utils.SendError(err, http.StatusInternalServerError)
		}

		data[i].SurveyCode = surveyCode
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
