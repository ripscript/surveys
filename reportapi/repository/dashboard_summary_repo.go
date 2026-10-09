package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type MetricFieldMapping struct {
	MetricKey   string
	Label       string
	FormFieldID int64
}

type DashboardSummaryRepository interface {
	GetLatestNumberMetrics(ctx context.Context, rtIDs []int64, mappings []MetricFieldMapping, periodStart, periodEnd time.Time) (map[int64]map[string]int64, error)
	GetLatestMetricsByCategory(ctx context.Context, rtIDs []int64, category string, periodStart, periodEnd time.Time) (map[int64]map[string]int64, error)
	GetLatestMetricsByCategoryForSurveys(ctx context.Context, rtIDs []int64, surveyIDs []int64, category string, periodStart, periodEnd time.Time) (map[int64]map[string]int64, error)

	GetLatestStatusByCategory(ctx context.Context, rtIDs []int64, category string, periodStart, periodEnd time.Time) (map[int64]map[string]string, error)
	GetMonthlyMetricSums(ctx context.Context, rtIDs []int64, metricKey string, rangeStart, rangeEnd time.Time) ([]MonthlyMetricRow, error)

	CountValidatedRespondentsByRT(ctx context.Context, rtIDs []int64, periodStart, periodEnd time.Time) (map[int64]int64, error)

	GetMonthlyMetricSumsBySurvey(ctx context.Context, rtIDs []int64, surveyIDs []int64, metricKey string, rangeStart, rangeEnd time.Time) ([]MonthlyMetricBySurveyRow, error)
	ListSurveyIDsWithMetricData(ctx context.Context, rtIDs []int64, metricKey string, rangeStart, rangeEnd time.Time) ([]SurveyOptionRow, error)
	ListSurveysLatestFirst(ctx context.Context, limit int) ([]SurveyOptionRow, error)
	GetSurveyNamesByIDs(ctx context.Context, surveyIDs []int64) (map[int64]string, error)

	GetMetricSumsByCategoryPerSurvey(ctx context.Context, rtIDs []int64, surveyIDs []int64, category string, periodStart, periodEnd time.Time) (map[int64]map[int64]map[string]int64, error)

	GetLatestMetricsByKeys(ctx context.Context, rtIDs []int64, metricKeys []string, periodStart, periodEnd time.Time) (map[int64]map[string]int64, error)

	ListSurveysLatestFirstByIDs(ctx context.Context, surveyIDs []int64) ([]SurveyOptionRow, error)
	GetLatestMetricsByKeysForSurveys(ctx context.Context, rtIDs []int64, surveyIDs []int64, metricKeys []string, periodStart, periodEnd time.Time) (map[int64]map[string]int64, error)

	GetMonthlyMetricSumsForSurveys(ctx context.Context, rtIDs []int64, surveyIDs []int64, metricKey string, rangeStart, rangeEnd time.Time) ([]MonthlyMetricRow, error)

	CountValidatedRespondentsByRTForSurveys(ctx context.Context, rtIDs []int64, surveyIDs []int64, periodStart, periodEnd time.Time) (map[int64]int64, error)

	GetSurveyFillStats(ctx context.Context, surveyIDs []int64, rtIDs []int64, periodStart, periodEnd time.Time) ([]SurveyFillStatRow, error)
}

type dashboardSummaryRepository struct {
	dbSlave *gorm.DB
}

func NewDashboardSummaryRepository(dbSlave *gorm.DB) DashboardSummaryRepository {
	return &dashboardSummaryRepository{dbSlave: dbSlave}
}

type latestFieldRow struct {
	RTID        int64
	FormFieldID int64
	Answer      string
}

func (r *dashboardSummaryRepository) GetLatestNumberMetrics(
	ctx context.Context,
	rtIDs []int64,
	mappings []MetricFieldMapping,
	periodStart, periodEnd time.Time,
) (map[int64]map[string]int64, error) {
	result := make(map[int64]map[string]int64, len(rtIDs))
	for _, id := range rtIDs {
		result[id] = make(map[string]int64)
	}
	if len(rtIDs) == 0 || len(mappings) == 0 {
		return result, nil
	}

	formFieldIDs := make([]int64, 0, len(mappings))
	fieldToMetric := make(map[int64]string, len(mappings))
	for _, m := range mappings {
		formFieldIDs = append(formFieldIDs, m.FormFieldID)
		fieldToMetric[m.FormFieldID] = m.MetricKey
	}

	// DISTINCT ON (respondents.rt_id, field_responses.form_field_id) mengambil
	// 1 baris terbaru (ORDER BY created_at DESC) per kombinasi RT + pertanyaan --
	// inilah implementasi "latest submission wins" yang sudah disepakati,
	// dijalankan langsung di DB (bukan diambil semua lalu difilter di Go) supaya
	// tetap efisien walau 1 RT punya banyak field_responses/periode.
	var rows []latestFieldRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (resp.rt_id, fr.form_field_id)
			resp.rt_id        AS rt_id,
			fr.form_field_id  AS form_field_id,
			fr.answer          AS answer
		FROM field_responses fr
		JOIN survey_respondents sr ON sr.id = fr.form_response_id
		JOIN respondents resp      ON resp.id = sr.respondent_id
		JOIN surveys s              ON s.id = sr.survey_id
		WHERE resp.rt_id IN (?)
		  AND fr.form_field_id IN (?)
		  AND s.status = 'finished'
		  AND sr.status_approval = 'validated_lurah'
		  AND fr.created_at BETWEEN ? AND ?
		ORDER BY resp.rt_id, fr.form_field_id, fr.created_at DESC
	`, rtIDs, formFieldIDs, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		metricKey, ok := fieldToMetric[row.FormFieldID]
		if !ok {
			continue
		}
		val, convErr := parseNumberAnswer(row.Answer)
		if convErr != nil {
			continue // jawaban tidak valid sebagai number, skip diam-diam (bisa di-log)
		}
		result[row.RTID][metricKey] = val
	}

	return result, nil
}

type latestMetricRow struct {
	RTID      int64
	MetricKey string
	Answer    string
}

func (r *dashboardSummaryRepository) GetLatestMetricsByCategory(
	ctx context.Context,
	rtIDs []int64,
	category string,
	periodStart, periodEnd time.Time,
) (map[int64]map[string]int64, error) {
	result := make(map[int64]map[string]int64, len(rtIDs))
	for _, id := range rtIDs {
		result[id] = make(map[string]int64)
	}
	if len(rtIDs) == 0 {
		return result, nil
	}

	// DISTINCT ON (rt_id, metric_key) = 1 baris terbaru per (RT, metric) --
	// tetap "latest submission wins" walau field_field_id-nya beda antar RT
	// (form_field_id -> mapping -> metric_key, bukan sebaliknya), sehingga
	// RT yang masih pakai form versi lama TETAP kehitung dengan benar
	// selama form lamanya juga sudah pernah di-mapping ke metric ini.
	var rows []latestMetricRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (resp.rt_id, dm.metric_key)
			resp.rt_id     AS rt_id,
			dm.metric_key  AS metric_key,
			fr.answer       AS answer
		FROM field_responses fr
		JOIN survey_respondents sr        ON sr.id = fr.form_response_id
		JOIN respondents resp             ON resp.id = sr.respondent_id
		JOIN surveys s                     ON s.id = sr.survey_id
		JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
		JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
		WHERE resp.rt_id IN (?)
			AND dm.category = ?
			AND sr.status_approval = 'validated_lurah'
			AND fr.created_at BETWEEN ? AND ?
		ORDER BY resp.rt_id, dm.metric_key, fr.created_at DESC
	`, rtIDs, category, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		val, convErr := parseNumberAnswer(row.Answer)
		if convErr != nil {
			continue
		}
		result[row.RTID][row.MetricKey] = val
	}

	return result, nil
}

func (r *dashboardSummaryRepository) GetLatestMetricsByCategoryForSurveys(
	ctx context.Context,
	rtIDs []int64,
	surveyIDs []int64,
	category string,
	periodStart, periodEnd time.Time,
) (map[int64]map[string]int64, error) {
	result := make(map[int64]map[string]int64, len(rtIDs))
	for _, id := range rtIDs {
		result[id] = make(map[string]int64)
	}
	if len(rtIDs) == 0 || len(surveyIDs) == 0 {
		return result, nil
	}

	var rows []latestMetricRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (resp.rt_id, dm.metric_key)
			resp.rt_id     AS rt_id,
			dm.metric_key  AS metric_key,
			fr.answer       AS answer
		FROM field_responses fr
		JOIN survey_respondents sr        ON sr.id = fr.form_response_id
		JOIN respondents resp             ON resp.id = sr.respondent_id
		JOIN surveys s                     ON s.id = sr.survey_id
		JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
		JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
		WHERE resp.rt_id IN (?)
			AND sr.survey_id IN (?)
			AND dm.category = ?
			AND sr.status_approval = 'validated_lurah'
			AND fr.created_at BETWEEN ? AND ?
		ORDER BY resp.rt_id, dm.metric_key, fr.created_at DESC
	`, rtIDs, surveyIDs, category, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		val, convErr := parseNumberAnswer(row.Answer)
		if convErr != nil {
			continue
		}
		result[row.RTID][row.MetricKey] = val
	}

	return result, nil
}

func parseNumberAnswer(answer string) (int64, error) {
	var v int64
	_, err := fmt.Sscan(answer, &v)
	return v, err
}

type latestStatusRow struct {
	RTID      int64  `gorm:"column:rt_id"`
	MetricKey string `gorm:"column:metric_key"`
	StatusKey string `gorm:"column:status_key"`
}

func (r *dashboardSummaryRepository) GetLatestStatusByCategory(
	ctx context.Context,
	rtIDs []int64,
	category string,
	periodStart, periodEnd time.Time,
) (map[int64]map[string]string, error) {
	result := make(map[int64]map[string]string, len(rtIDs))
	for _, id := range rtIDs {
		result[id] = make(map[string]string)
	}
	if len(rtIDs) == 0 {
		return result, nil
	}

	var rows []latestStatusRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (resp.rt_id, dm.metric_key)
			resp.rt_id      AS rt_id,
			dm.metric_key   AS metric_key,
			dosm.status_key AS status_key
		FROM field_responses fr
		JOIN survey_respondents sr        ON sr.id = fr.form_response_id
		JOIN respondents resp             ON resp.id = sr.respondent_id
		JOIN surveys s                     ON s.id = sr.survey_id
		JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
		JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
		JOIN dashboard_option_status_mappings dosm
			ON dosm.dashboard_metric_mapping_id = dmm.id
			AND dosm.answer_option_id = fr.answer::bigint
		WHERE resp.rt_id IN (?)
			AND dm.category = ?
			AND dm.expected_template = 'multiple-choices'
			AND s.status = 'finished'
			AND sr.status_approval = 'validated_lurah'
			AND fr.created_at BETWEEN ? AND ?
		ORDER BY resp.rt_id, dm.metric_key, fr.created_at DESC
	`, rtIDs, category, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.RTID][row.MetricKey] = row.StatusKey
	}

	return result, nil
}

type MonthlyMetricRow struct {
	MonthBucket time.Time `gorm:"column:month_bucket"`
	Total       int64     `gorm:"column:total"`
}

func (r *dashboardSummaryRepository) GetMonthlyMetricSums(ctx context.Context, rtIDs []int64, metricKey string, rangeStart, rangeEnd time.Time) ([]MonthlyMetricRow, error) {
	rtFilter := "TRUE"
	args := []interface{}{}
	if len(rtIDs) > 0 {
		rtFilter = "resp.rt_id IN (?)"
		args = append(args, rtIDs)
	}
	args = append(args, metricKey, rangeStart, rangeEnd)

	sql := fmt.Sprintf(`
		WITH latest_per_rt_month AS (
			SELECT DISTINCT ON (resp.rt_id, date_trunc('month', fr.created_at))
				resp.rt_id                          AS rt_id,
				date_trunc('month', fr.created_at)  AS month_bucket,
				fr.answer                            AS answer
			FROM field_responses fr
			JOIN survey_respondents sr        ON sr.id = fr.form_response_id
			JOIN respondents resp             ON resp.id = sr.respondent_id
			JOIN surveys s                     ON s.id = sr.survey_id
			JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
			JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
			WHERE %s
				AND dm.metric_key = ?
				AND sr.status_approval = 'validated_lurah'
				AND fr.created_at BETWEEN ? AND ?
			ORDER BY resp.rt_id, date_trunc('month', fr.created_at), fr.created_at DESC
		)
		SELECT
			month_bucket,
			COALESCE(SUM(NULLIF(answer, '')::bigint), 0) AS total
		FROM latest_per_rt_month
		GROUP BY month_bucket
		ORDER BY month_bucket
	`, rtFilter)

	var rows []MonthlyMetricRow
	err := r.dbSlave.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

type validatedRespondentRow struct {
	RTID  int64 `gorm:"column:rt_id"`
	Total int64 `gorm:"column:total"`
}

func (r *dashboardSummaryRepository) CountValidatedRespondentsByRT(ctx context.Context, rtIDs []int64, periodStart, periodEnd time.Time) (map[int64]int64, error) {
	result := make(map[int64]int64, len(rtIDs))
	for _, id := range rtIDs {
		result[id] = 0
	}
	if len(rtIDs) == 0 {
		return result, nil
	}

	var rows []validatedRespondentRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT resp.rt_id AS rt_id, COUNT(DISTINCT resp.id) AS total
		FROM survey_respondents sr
		JOIN respondents resp ON resp.id = sr.respondent_id
		WHERE resp.rt_id IN (?)
			AND sr.status = 2
			AND sr.status_approval = 'validated_lurah'
			AND sr.created_at BETWEEN ? AND ?
		GROUP BY resp.rt_id
	`, rtIDs, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.RTID] = row.Total
	}
	return result, nil
}

// ============================================================================
// TREND COMPARE — fungsi baru untuk membandingkan antar-survey pada chart trend.
// Semua di bawah ini tidak mengubah query/relasi yang sudah ada.
// ============================================================================

// MonthlyMetricBySurveyRow mirip MonthlyMetricRow tapi membawa dimensi survey_id.
type MonthlyMetricBySurveyRow struct {
	SurveyID    int64     `gorm:"column:survey_id"`
	MonthBucket time.Time `gorm:"column:month_bucket"`
	Total       int64     `gorm:"column:total"`
}

// SurveyOptionRow dipakai untuk opsi dropdown & pemilihan "survey terbaru".
type SurveyOptionRow struct {
	SurveyID   int64  `gorm:"column:survey_id"`
	SurveyName string `gorm:"column:survey_name"`
}

func (r *dashboardSummaryRepository) GetMonthlyMetricSumsBySurvey(
	ctx context.Context,
	rtIDs []int64,
	surveyIDs []int64,
	metricKey string,
	rangeStart, rangeEnd time.Time,
) ([]MonthlyMetricBySurveyRow, error) {
	if len(surveyIDs) == 0 {
		return []MonthlyMetricBySurveyRow{}, nil
	}

	rtFilter := "TRUE"
	args := []interface{}{}
	if len(rtIDs) > 0 {
		rtFilter = "resp.rt_id IN (?)"
		args = append(args, rtIDs)
	}
	// survey_ids selalu ada (sudah divalidasi len>0 di atas)
	args = append(args, surveyIDs, metricKey, rangeStart, rangeEnd)

	// Perhatikan DISTINCT ON kini memasukkan sr.survey_id sebagai dimensi,
	// supaya data tiap survey TIDAK saling menimpa walau RT & bulannya sama.
	sql := fmt.Sprintf(`
		WITH latest_per_survey_rt_month AS (
			SELECT DISTINCT ON (sr.survey_id, resp.rt_id, date_trunc('month', fr.created_at))
				sr.survey_id                        AS survey_id,
				resp.rt_id                          AS rt_id,
				date_trunc('month', fr.created_at)  AS month_bucket,
				fr.answer                            AS answer
			FROM field_responses fr
			JOIN survey_respondents sr         ON sr.id = fr.form_response_id
			JOIN respondents resp              ON resp.id = sr.respondent_id
			JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
			JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
			WHERE %s
				AND sr.survey_id IN (?)
				AND dm.metric_key = ?
				AND sr.status_approval = 'validated_lurah'
				AND fr.created_at BETWEEN ? AND ?
			ORDER BY sr.survey_id, resp.rt_id, date_trunc('month', fr.created_at), fr.created_at DESC
		)
		SELECT
			survey_id,
			month_bucket,
			COALESCE(SUM(NULLIF(answer, '')::bigint), 0) AS total
		FROM latest_per_survey_rt_month
		GROUP BY survey_id, month_bucket
		ORDER BY survey_id, month_bucket
	`, rtFilter)

	var rows []MonthlyMetricBySurveyRow
	if err := r.dbSlave.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *dashboardSummaryRepository) ListSurveyIDsWithMetricData(
	ctx context.Context,
	rtIDs []int64,
	metricKey string,
	rangeStart, rangeEnd time.Time,
) ([]SurveyOptionRow, error) {
	rtFilter := "TRUE"
	args := []interface{}{}
	if len(rtIDs) > 0 {
		rtFilter = "resp.rt_id IN (?)"
		args = append(args, rtIDs)
	}
	args = append(args, metricKey, rangeStart, rangeEnd)

	// DISTINCT survey yang punya data metric_key tervalidasi di wilayah ini,
	// urut id DESC = terbaru dulu.
	sql := fmt.Sprintf(`
		SELECT DISTINCT s.id AS survey_id, s.name AS survey_name
		FROM field_responses fr
		JOIN survey_respondents sr         ON sr.id = fr.form_response_id
		JOIN respondents resp              ON resp.id = sr.respondent_id
		JOIN surveys s                      ON s.id = sr.survey_id
		JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
		JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
		WHERE %s
			AND dm.metric_key = ?
			AND sr.status_approval = 'validated_lurah'
			AND fr.created_at BETWEEN ? AND ?
		ORDER BY s.id DESC
	`, rtFilter)

	var rows []SurveyOptionRow
	if err := r.dbSlave.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *dashboardSummaryRepository) GetSurveyNamesByIDs(ctx context.Context, surveyIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(surveyIDs))
	if len(surveyIDs) == 0 {
		return result, nil
	}

	type nameRow struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	var rows []nameRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT id, name FROM surveys WHERE id IN (?)
	`, surveyIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ID] = row.Name
	}
	return result, nil
}

func (r *dashboardSummaryRepository) ListSurveysLatestFirst(ctx context.Context, limit int) ([]SurveyOptionRow, error) {
	q := r.dbSlave.WithContext(ctx).
		Table("surveys").
		Select("id AS survey_id, name AS survey_name").
		Order("id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}

	var rows []SurveyOptionRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type latestMetricPerSurveyRow struct {
	SurveyID  int64  `gorm:"column:survey_id"`
	RTID      int64  `gorm:"column:rt_id"`
	MetricKey string `gorm:"column:metric_key"`
	Answer    string `gorm:"column:answer"`
}

// GetMetricSumsByCategoryPerSurvey — sama logic "latest submission wins" seperti
// GetLatestMetricsByCategory, tapi DISTINCT ON-nya menambahkan sr.survey_id
// sebagai dimensi sehingga tiap survey dihitung terpisah.
func (r *dashboardSummaryRepository) GetMetricSumsByCategoryPerSurvey(
	ctx context.Context,
	rtIDs []int64,
	surveyIDs []int64,
	category string,
	periodStart, periodEnd time.Time,
) (map[int64]map[int64]map[string]int64, error) {
	result := make(map[int64]map[int64]map[string]int64, len(surveyIDs))
	// pre-init supaya survey yang dipilih tetap ada key-nya walau datanya kosong
	for _, sid := range surveyIDs {
		result[sid] = make(map[int64]map[string]int64, len(rtIDs))
		for _, rtID := range rtIDs {
			result[sid][rtID] = make(map[string]int64)
		}
	}
	if len(rtIDs) == 0 || len(surveyIDs) == 0 {
		return result, nil
	}

	var rows []latestMetricPerSurveyRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (sr.survey_id, resp.rt_id, dm.metric_key)
			sr.survey_id   AS survey_id,
			resp.rt_id     AS rt_id,
			dm.metric_key  AS metric_key,
			fr.answer       AS answer
		FROM field_responses fr
		JOIN survey_respondents sr        ON sr.id = fr.form_response_id
		JOIN respondents resp             ON resp.id = sr.respondent_id
		JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
		JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
		WHERE resp.rt_id IN (?)
			AND sr.survey_id IN (?)
			AND dm.category = ?
			AND sr.status_approval = 'validated_lurah'
			AND fr.created_at BETWEEN ? AND ?
		ORDER BY sr.survey_id, resp.rt_id, dm.metric_key, fr.created_at DESC
	`, rtIDs, surveyIDs, category, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		val, convErr := parseNumberAnswer(row.Answer)
		if convErr != nil {
			continue
		}
		if result[row.SurveyID] == nil {
			result[row.SurveyID] = make(map[int64]map[string]int64)
		}
		if result[row.SurveyID][row.RTID] == nil {
			result[row.SurveyID][row.RTID] = make(map[string]int64)
		}
		result[row.SurveyID][row.RTID][row.MetricKey] = val
	}

	return result, nil
}

func (r *dashboardSummaryRepository) GetLatestMetricsByKeys(
	ctx context.Context,
	rtIDs []int64,
	metricKeys []string,
	periodStart, periodEnd time.Time,
) (map[int64]map[string]int64, error) {
	result := make(map[int64]map[string]int64, len(rtIDs))
	for _, id := range rtIDs {
		result[id] = make(map[string]int64)
	}
	if len(rtIDs) == 0 || len(metricKeys) == 0 {
		return result, nil
	}

	var rows []latestMetricRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (resp.rt_id, dm.metric_key)
			resp.rt_id     AS rt_id,
			dm.metric_key  AS metric_key,
			fr.answer       AS answer
		FROM field_responses fr
		JOIN survey_respondents sr        ON sr.id = fr.form_response_id
		JOIN respondents resp             ON resp.id = sr.respondent_id
		JOIN surveys s                     ON s.id = sr.survey_id
		JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
		JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
		WHERE resp.rt_id IN (?)
			AND dm.metric_key IN (?)
			AND sr.status_approval = 'validated_lurah'
			AND fr.created_at BETWEEN ? AND ?
		ORDER BY resp.rt_id, dm.metric_key, fr.created_at DESC
	`, rtIDs, metricKeys, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		val, convErr := parseNumberAnswer(row.Answer)
		if convErr != nil {
			continue
		}
		result[row.RTID][row.MetricKey] = val
	}
	return result, nil
}

func (r *dashboardSummaryRepository) ListSurveysLatestFirstByIDs(ctx context.Context, surveyIDs []int64) ([]SurveyOptionRow, error) {
	var rows []SurveyOptionRow
	if len(surveyIDs) == 0 {
		return rows, nil
	}
	err := r.dbSlave.WithContext(ctx).
		Table("surveys").
		Select("id AS survey_id, name AS survey_name").
		Where("id IN ?", surveyIDs).
		Order("id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *dashboardSummaryRepository) GetLatestMetricsByKeysForSurveys(
	ctx context.Context,
	rtIDs []int64,
	surveyIDs []int64,
	metricKeys []string,
	periodStart, periodEnd time.Time,
) (map[int64]map[string]int64, error) {
	result := make(map[int64]map[string]int64, len(rtIDs))
	for _, id := range rtIDs {
		result[id] = make(map[string]int64)
	}
	if len(rtIDs) == 0 || len(metricKeys) == 0 || len(surveyIDs) == 0 {
		return result, nil
	}

	var rows []latestMetricRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (resp.rt_id, dm.metric_key)
			resp.rt_id     AS rt_id,
			dm.metric_key  AS metric_key,
			fr.answer       AS answer
		FROM field_responses fr
		JOIN survey_respondents sr        ON sr.id = fr.form_response_id
		JOIN respondents resp             ON resp.id = sr.respondent_id
		JOIN surveys s                     ON s.id = sr.survey_id
		JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
		JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
		WHERE resp.rt_id IN (?)
			AND sr.survey_id IN (?)
			AND dm.metric_key IN (?)
			AND sr.status_approval = 'validated_lurah'
			AND fr.created_at BETWEEN ? AND ?
		ORDER BY resp.rt_id, dm.metric_key, fr.created_at DESC
	`, rtIDs, surveyIDs, metricKeys, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		val, convErr := parseNumberAnswer(row.Answer)
		if convErr != nil {
			continue
		}
		result[row.RTID][row.MetricKey] = val
	}
	return result, nil
}

func (r *dashboardSummaryRepository) GetMonthlyMetricSumsForSurveys(
	ctx context.Context,
	rtIDs []int64,
	surveyIDs []int64,
	metricKey string,
	rangeStart, rangeEnd time.Time,
) ([]MonthlyMetricRow, error) {
	if len(surveyIDs) == 0 {
		return []MonthlyMetricRow{}, nil
	}

	rtFilter := "TRUE"
	args := []interface{}{}
	if len(rtIDs) > 0 {
		rtFilter = "resp.rt_id IN (?)"
		args = append(args, rtIDs)
	}
	args = append(args, surveyIDs, metricKey, rangeStart, rangeEnd)

	sql := fmt.Sprintf(`
		WITH latest_per_rt_month AS (
			SELECT DISTINCT ON (resp.rt_id, date_trunc('month', fr.created_at))
				resp.rt_id                          AS rt_id,
				date_trunc('month', fr.created_at)  AS month_bucket,
				fr.answer                            AS answer
			FROM field_responses fr
			JOIN survey_respondents sr        ON sr.id = fr.form_response_id
			JOIN respondents resp             ON resp.id = sr.respondent_id
			JOIN surveys s                     ON s.id = sr.survey_id
			JOIN dashboard_metric_mappings dmm ON dmm.form_field_id = fr.form_field_id
			JOIN dashboard_metrics dm          ON dm.id = dmm.dashboard_metric_id
			WHERE %s
				AND sr.survey_id IN (?)
				AND dm.metric_key = ?
				AND sr.status_approval = 'validated_lurah'
				AND fr.created_at BETWEEN ? AND ?
			ORDER BY resp.rt_id, date_trunc('month', fr.created_at), fr.created_at DESC
		)
		SELECT
			month_bucket,
			COALESCE(SUM(NULLIF(answer, '')::bigint), 0) AS total
		FROM latest_per_rt_month
		GROUP BY month_bucket
		ORDER BY month_bucket
	`, rtFilter)

	var rows []MonthlyMetricRow
	err := r.dbSlave.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *dashboardSummaryRepository) CountValidatedRespondentsByRTForSurveys(ctx context.Context, rtIDs []int64, surveyIDs []int64, periodStart, periodEnd time.Time) (map[int64]int64, error) {
	result := make(map[int64]int64, len(rtIDs))
	for _, id := range rtIDs {
		result[id] = 0
	}
	if len(rtIDs) == 0 || len(surveyIDs) == 0 {
		return result, nil
	}

	var rows []validatedRespondentRow
	err := r.dbSlave.WithContext(ctx).Raw(`
		SELECT resp.rt_id AS rt_id, COUNT(DISTINCT resp.id) AS total
		FROM survey_respondents sr
		JOIN respondents resp ON resp.id = sr.respondent_id
		WHERE resp.rt_id IN (?)
			AND sr.survey_id IN (?)
			AND sr.status = 2
			AND sr.status_approval = 'validated_lurah'
			AND sr.created_at BETWEEN ? AND ?
		GROUP BY resp.rt_id
	`, rtIDs, surveyIDs, periodStart, periodEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.RTID] = row.Total
	}
	return result, nil
}

type SurveyFillStatRow struct {
	SurveyID   int64  `gorm:"column:survey_id"`
	SurveyName string `gorm:"column:survey_name"`
	TotalRT    int64  `gorm:"column:total_rt"`
	JumlahData int64  `gorm:"column:jumlah_data"`
}

// GetSurveyFillStats — per survey: jumlah RT berbeda yang mengisi (tervalidasi)
// dan jumlah data/responden tervalidasi. Survey tanpa data tetap muncul
// (total_rt = 0, jumlah_data = 0), sesuai tampilan kartu "0 RT • 0 Data".
// Diurut dari jumlah_data terbanyak.
func (r *dashboardSummaryRepository) GetSurveyFillStats(
	ctx context.Context,
	surveyIDs []int64,
	rtIDs []int64,
	periodStart, periodEnd time.Time,
) ([]SurveyFillStatRow, error) {
	if len(surveyIDs) == 0 {
		return []SurveyFillStatRow{}, nil
	}

	rtJoinFilter := ""
	args := []interface{}{periodStart, periodEnd}
	if len(rtIDs) > 0 {
		rtJoinFilter = "AND resp.rt_id IN (?)"
		args = append(args, rtIDs)
	}
	args = append(args, surveyIDs)

	sql := fmt.Sprintf(`
		SELECT
			s.id   AS survey_id,
			s.name AS survey_name,
			COUNT(DISTINCT resp.rt_id) AS total_rt,
			COUNT(DISTINCT resp.id)    AS jumlah_data
		FROM surveys s
		LEFT JOIN survey_respondents sr
			ON sr.survey_id = s.id
			AND sr.status = 2
			AND sr.status_approval = 'validated_lurah'
			AND sr.created_at BETWEEN ? AND ?
		LEFT JOIN respondents resp
			ON resp.id = sr.respondent_id
			AND resp.rt_id IS NOT NULL
			%s
		WHERE s.id IN (?)
		GROUP BY s.id, s.name
		ORDER BY jumlah_data DESC, s.id DESC
	`, rtJoinFilter)

	var rows []SurveyFillStatRow
	if err := r.dbSlave.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
