package seeders

import (
	"backend/masterapi/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type statusSeed struct {
	StatusKey string
	Label     string
	Sequence  int
}

type metricSeed struct {
	models.DashboardMetric
	Statuses []statusSeed
}

// SeedDashboardMetrics: seeder untuk metrik dashboard yang SUDAH dipakai
// production dashboard (legacy, hardcode). Semua di-mark IsLocked=true
// karena dashboard masih bergantung pada metric_key ini secara statis —
// tidak boleh terhapus/ter-nonaktifkan lewat menu Manajemen Metrik.
//
// Sync behavior:
//   - metric_key yang ADA di code tapi BELUM ada di DB -> di-insert (+statuses)
//   - metric_key yang SUDAH ada di DB -> TIDAK di-overwrite (field manual via
//     menu Manajemen Metrik tetap aman), tapi statuses-nya tetap di-sync
//   - metric_key yang ADA di DB tapi TIDAK ADA lagi di code -> dihapus,
//     KECUALI masih direferensikan oleh dashboard_metric_mappings (FK block),
//     dalam hal ini di-skip + warning, bukan error
//   - ID yang sudah ada TIDAK PERNAH diubah — sync ini cuma insert/delete,
//     tidak pernah delete-lalu-recreate baris yang masih ada di code
//
// Idempotent: aman dijalankan berkali-kali (misal tiap deploy).
func SeedDashboardMetrics(dbMaster *gorm.DB) error {
	ctx := context.Background()

	metrics := []metricSeed{
		// ===== summary =====
		{DashboardMetric: models.DashboardMetric{MetricKey: "total_rumah", Label: "Total Rumah", ExpectedTemplate: "number", Category: "summary", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "total_kk", Label: "Total KK", ExpectedTemplate: "number", Category: "summary", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "total_penduduk", Label: "Total Penduduk", ExpectedTemplate: "number", Category: "summary", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "total_stunting", Label: "Total Stunting", ExpectedTemplate: "number", Category: "summary", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "total_tidak_stunting", Label: "Total Tidak Stunting", ExpectedTemplate: "number", Category: "summary", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "rumah_tidak_layak", Label: "Rumah Tidak Layak Huni (RTLH)", ExpectedTemplate: "number", Category: "summary", IsLocked: true}},

		// ===== sampah =====
		{DashboardMetric: models.DashboardMetric{MetricKey: "sampah_tps", Label: "TPS", ExpectedTemplate: "number", Category: "sampah", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "sampah_residu", Label: "Residu", ExpectedTemplate: "number", Category: "sampah", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "sampah_organik_aktif", Label: "Organik Aktif", ExpectedTemplate: "number", Category: "sampah", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "sampah_anorganik_aktif", Label: "Anorganik Aktif", ExpectedTemplate: "number", Category: "sampah", IsLocked: true}},

		// ===== infrastruktur =====
		{DashboardMetric: models.DashboardMetric{MetricKey: "jalanan_umum_total", Label: "Jalanan Umum", ExpectedTemplate: "number", Category: "infrastruktur", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "jalanan_umum_rusak", Label: "Jalanan Umum Rusak/Perlu Perbaikan", ExpectedTemplate: "number", Category: "infrastruktur", IsLocked: true}},
		// {
		// 	DashboardMetric: models.DashboardMetric{MetricKey: "apakah_wilayah_telah_memiliki_sarana_dan_fasilitas_pengelolaan_sampah_organik", Label: "Apakah wilayah telah memiliki sarana dan fasilitas pengelolaan sampah Organik", ExpectedTemplate: "multiple-choices", Category: "infrastruktur", IsLocked: true},
		// 	Statuses: []statusSeed{
		// 		{StatusKey: "ya", Label: "Ya", Sequence: 1},
		// 		{StatusKey: "tidak", Label: "Tidak", Sequence: 2},
		// 	},
		// },

		{DashboardMetric: models.DashboardMetric{MetricKey: "jalanan_lingkungan_total", Label: "Jalanan Lingkungan", ExpectedTemplate: "number", Category: "infrastruktur", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "jalanan_lingkungan_rusak", Label: "Jalanan Lingkungan Rusak/Perlu Perbaikan", ExpectedTemplate: "number", Category: "infrastruktur", IsLocked: true}},
		{DashboardMetric: models.DashboardMetric{MetricKey: "rumah_tidak_layak_infrastruktur", Label: "Rumah Tidak Layak", ExpectedTemplate: "number", Category: "infrastruktur", IsLocked: true}},

		{DashboardMetric: models.DashboardMetric{MetricKey: "septic_tarik_pribadi", Label: "Septic Tarik Pribadi", ExpectedTemplate: "number", Category: "infrastruktur", IsLocked: true}},

		{DashboardMetric: models.DashboardMetric{MetricKey: "mck_umum", Label: "MCK Umum", ExpectedTemplate: "number", Category: "infrastruktur", IsLocked: true}},
	}

	return dbMaster.Transaction(func(tx *gorm.DB) error {
		// ===== PART 1: upsert metric yang ada di code =====
		codeMetricKeys := make(map[string]bool, len(metrics))

		for _, seed := range metrics {
			codeMetricKeys[seed.MetricKey] = true

			var existing models.DashboardMetric
			err := tx.WithContext(ctx).
				Where("metric_key = ?", seed.MetricKey).
				First(&existing).Error

			switch {
			case err == nil:
				fmt.Printf("[seeder] skip field, sudah ada: %s\n", seed.MetricKey)
				if err := syncMetricStatuses(tx, ctx, existing.ID, seed.Statuses); err != nil {
					return fmt.Errorf("gagal sync statuses utk metric %s: %w", seed.MetricKey, err)
				}
				continue

			case !errors.Is(err, gorm.ErrRecordNotFound):
				return err
			}

			metric := seed.DashboardMetric
			if err := tx.WithContext(ctx).Create(&metric).Error; err != nil {
				return fmt.Errorf("gagal insert metric %s: %w", seed.MetricKey, err)
			}
			fmt.Printf("[seeder] created: %s\n", seed.MetricKey)

			if err := syncMetricStatuses(tx, ctx, metric.ID, seed.Statuses); err != nil {
				return fmt.Errorf("gagal insert statuses utk metric %s: %w", seed.MetricKey, err)
			}
		}

		// ===== PART 2: hapus metric di DB yang sudah tidak ada di code =====
		var dbMetrics []models.DashboardMetric
		if err := tx.WithContext(ctx).Find(&dbMetrics).Error; err != nil {
			return err
		}

		for _, m := range dbMetrics {
			if codeMetricKeys[m.MetricKey] {
				continue
			}

			var mappingCount int64
			if err := tx.WithContext(ctx).
				Table("dashboard_metric_mappings").
				Where("dashboard_metric_id = ?", m.ID).
				Count(&mappingCount).Error; err != nil {
				return err
			}

			if mappingCount > 0 {
				fmt.Printf("[seeder] WARNING: %s sudah tidak ada di code tapi masih dipakai %d mapping -> TIDAK dihapus\n", m.MetricKey, mappingCount)
				continue
			}

			if err := tx.WithContext(ctx).
				Where("dashboard_metric_id = ?", m.ID).
				Delete(&models.DashboardMetricStatus{}).Error; err != nil {
				return fmt.Errorf("gagal hapus statuses metric %s: %w", m.MetricKey, err)
			}

			if err := tx.WithContext(ctx).Delete(&m).Error; err != nil {
				return fmt.Errorf("gagal hapus metric %s: %w", m.MetricKey, err)
			}
			fmt.Printf("[seeder] deleted (tidak ada lagi di code): %s\n", m.MetricKey)
		}

		return nil
	})
}

// syncMetricStatuses: sync daftar status milik 1 dashboard_metric terhadap
// definisi di code — insert yang baru, hapus yang sudah tidak ada di code
// (kecuali masih dipakai di dashboard_option_status_mappings), dan TIDAK
// overwrite status yang sudah ada (biar ID & kustomisasi manual tetap aman).
func syncMetricStatuses(tx *gorm.DB, ctx context.Context, metricID int64, codeStatuses []statusSeed) error {
	if len(codeStatuses) == 0 {
		return nil
	}

	var dbStatuses []models.DashboardMetricStatus
	if err := tx.WithContext(ctx).
		Where("dashboard_metric_id = ?", metricID).
		Find(&dbStatuses).Error; err != nil {
		return err
	}

	dbByKey := make(map[string]models.DashboardMetricStatus, len(dbStatuses))
	for _, s := range dbStatuses {
		dbByKey[s.StatusKey] = s
	}

	codeKeys := make(map[string]bool, len(codeStatuses))
	toInsert := make([]models.DashboardMetricStatus, 0)

	for _, cs := range codeStatuses {
		codeKeys[cs.StatusKey] = true
		if _, exists := dbByKey[cs.StatusKey]; exists {
			continue
		}
		toInsert = append(toInsert, models.DashboardMetricStatus{
			DashboardMetricID: metricID,
			StatusKey:         cs.StatusKey,
			Label:             cs.Label,
			Sequence:          cs.Sequence,
			IsActive:          true,
		})
	}

	if len(toInsert) > 0 {
		if err := tx.WithContext(ctx).Create(&toInsert).Error; err != nil {
			return err
		}
	}

	for _, s := range dbStatuses {
		if codeKeys[s.StatusKey] {
			continue
		}

		var usage int64
		if err := tx.WithContext(ctx).
			Table("dashboard_option_status_mappings").
			Where("dashboard_metric_status_id = ?", s.ID).
			Count(&usage).Error; err != nil {
			return err
		}
		if usage > 0 {
			fmt.Printf("[seeder] WARNING: status %s (metric_id=%d) sudah tidak ada di code tapi masih dipakai %d option mapping -> TIDAK dihapus\n", s.StatusKey, metricID, usage)
			continue
		}

		if err := tx.WithContext(ctx).Delete(&s).Error; err != nil {
			return err
		}
	}

	return nil
}
