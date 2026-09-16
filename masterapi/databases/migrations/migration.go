package migrations

import (
	"backend/masterapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.CMSSection{},
		&models.CMSContent{},
		&models.CMSItem{},
		&models.CMSMedia{},
		&models.Artikel{},
		&models.DashboardMetric{},
		&models.DashboardMetricMapping{},
		&models.DashboardOptionStatusMapping{},
		&models.KecamatanModel{},
		&models.KelurahanModel{},
	)
	if err != nil {
		return err
	}

	if db.Migrator().HasColumn(&models.DashboardMetricMapping{}, "answer_option_id") {
		if err := db.Migrator().DropColumn(&models.DashboardMetricMapping{}, "answer_option_id"); err != nil {
			return err
		}
	}
	return nil
}

func AddDashboardPerformanceIndexes(db *gorm.DB) error {
	statements := []string{
		// prioritas tinggi -- dipakai di semua endpoint dashboard
		`CREATE INDEX IF NOT EXISTS idx_field_responses_form_field_created ON field_responses (form_field_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_respondents_rt_id ON respondents (rt_id)`,
		`CREATE INDEX IF NOT EXISTS idx_survey_respondents_respondent_id ON survey_respondents (respondent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dashboard_metric_mappings_form_field_id ON dashboard_metric_mappings (form_field_id)`,

		// prioritas menengah -- filter kondisi
		`CREATE INDEX IF NOT EXISTS idx_surveys_status ON surveys (status)`,
		`CREATE INDEX IF NOT EXISTS idx_survey_respondents_status_approval ON survey_respondents (status_approval)`,
		`CREATE INDEX IF NOT EXISTS idx_dashboard_metrics_category ON dashboard_metrics (category)`,

		// hierarki wilayah -- dipakai scatter/resolve RW-Kelurahan-Kecamatan
		`CREATE INDEX IF NOT EXISTS idx_data_rts_rw_id ON data__rts (rw_id)`,
		`CREATE INDEX IF NOT EXISTS idx_data_rws_kelurahan_id ON data__rws (kelurahan_id)`,
		`CREATE INDEX IF NOT EXISTS idx_kelurahans_sub_district_id ON kelurahans (sub_district_id)`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	// refresh statistik query planner supaya index baru langsung terdeteksi
	analyzeStatements := []string{
		`ANALYZE field_responses`,
		`ANALYZE respondents`,
		`ANALYZE survey_respondents`,
		`ANALYZE surveys`,
	}
	for _, stmt := range analyzeStatements {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	return nil
}
