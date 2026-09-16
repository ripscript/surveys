package seeders

import (
	"backend/masterapi/models"
	"backend/masterapi/utils"

	"gorm.io/gorm"
)

// DashboardMetricSeed adalah katalog metric_key yang dipakai dashboard
// (level RT s.d. Kecamatan, rollup lintas level pakai metric_key yang SAMA --
// RW/Kelurahan/Kecamatan cuma agregasi dari RT, bukan metric baru).
//
// TIDAK termasuk di sini: Total RT/Total RW (hitungan administratif, bukan
// jawaban survey), skor kesejahteraan/ranking/zona kondisi/insight/trend
// (computed/derived, sengaja di-skip dulu sesuai keputusan scope).
func DashboardMetricSeed(db *gorm.DB) error {
	metrics := []models.DashboardMetric{
		// ===== summary =====
		{MetricKey: "total_rumah", Label: "Total Rumah", ExpectedTemplate: "number", Category: "summary"},
		{MetricKey: "total_kk", Label: "Total KK", ExpectedTemplate: "number", Category: "summary"},
		{MetricKey: "total_penduduk", Label: "Total Penduduk", ExpectedTemplate: "number", Category: "summary"},
		{MetricKey: "total_stunting", Label: "Total Stunting", ExpectedTemplate: "number", Category: "summary"},
		{MetricKey: "total_tidak_stunting", Label: "Total Tidak Stunting", ExpectedTemplate: "number", Category: "summary"},
		{MetricKey: "rumah_tidak_layak", Label: "Rumah Tidak Layak Huni (RTLH)", ExpectedTemplate: "number", Category: "summary"},

		// ===== sampah =====
		{MetricKey: "sampah_tps", Label: "TPS", ExpectedTemplate: "number", Category: "sampah"},
		{MetricKey: "sampah_residu", Label: "Residu", ExpectedTemplate: "number", Category: "sampah"},
		{MetricKey: "sampah_organik_aktif", Label: "Organik Aktif", ExpectedTemplate: "number", Category: "sampah"},
		{MetricKey: "sampah_anorganik_aktif", Label: "Anorganik Aktif", ExpectedTemplate: "number", Category: "sampah"},

		// ===== infrastruktur =====
		{MetricKey: "jalanan_umum_total", Label: "Jalanan Umum", ExpectedTemplate: "number", Category: "infrastruktur"},
		{MetricKey: "jalanan_umum_rusak", Label: "Jalanan Umum Rusak/Perlu Perbaikan", ExpectedTemplate: "number", Category: "infrastruktur"},
		{MetricKey: "jalanan_umum_kondisi", Label: "Kondisi Dominan Jalanan Umum", ExpectedTemplate: "multiple-choices", Category: "infrastruktur"},

		{MetricKey: "jalanan_lingkungan_total", Label: "Jalanan Lingkungan", ExpectedTemplate: "number", Category: "infrastruktur"},
		{MetricKey: "jalanan_lingkungan_rusak", Label: "Jalanan Lingkungan Rusak/Perlu Perbaikan", ExpectedTemplate: "number", Category: "infrastruktur"},
		{MetricKey: "jalanan_lingkungan_kondisi", Label: "Kondisi Dominan Jalanan Lingkungan", ExpectedTemplate: "multiple-choices", Category: "infrastruktur"},
		{MetricKey: "rumah_tidak_layak_infrastruktur", Label: "Rumah Tidak Layak", ExpectedTemplate: "number", Category: "infrastruktur"},
		{MetricKey: "rumah_tidak_layak_kondisi_infrastruktur", Label: "Kondisi Dominan Rumah Tidak Layak", ExpectedTemplate: "multiple-choices", Category: "infrastruktur"},

		{MetricKey: "septic_tarik_pribadi", Label: "Septic Tarik Pribadi", ExpectedTemplate: "number", Category: "infrastruktur"},
		{MetricKey: "septic_kondisi", Label: "Kondisi Dominan Septic Tarik Pribadi", ExpectedTemplate: "multiple-choices", Category: "infrastruktur"},

		{MetricKey: "mck_umum", Label: "MCK Umum", ExpectedTemplate: "number", Category: "infrastruktur"},
		{MetricKey: "mck_kondisi", Label: "Kondisi Dominan MCK Umum", ExpectedTemplate: "multiple-choices", Category: "infrastruktur"},
	}

	var keepKeys []string

	for _, metric := range metrics {
		keepKeys = append(keepKeys, metric.MetricKey)

		var existing models.DashboardMetric
		err := db.Where("metric_key = ?", metric.MetricKey).First(&existing).Error
		isNew := errorsIsNotFound(err)

		if err != nil && !isNew {
			return err
		}

		if isNew {
			if err := db.Create(&metric).Error; err != nil {
				return err
			}
			continue
		}

		// sinkronkan label/template/category dengan yang ada di code --
		// metric_key sendiri TIDAK pernah diubah lewat seeder (kontrak stabil,
		// selaras aturan yang sama di service: metric_key immutable setelah create)
		existing.Label = metric.Label
		existing.ExpectedTemplate = metric.ExpectedTemplate
		existing.Category = metric.Category
		existing.IsDashboard = utils.BoolToPointer(true)
		if err := db.Save(&existing).Error; err != nil {
			return err
		}
	}

	// Sync penghapusan: hapus dashboard_metrics yang metric_key-nya sudah tidak
	// ada lagi di code. Kalau masih ada dashboard_metric_mappings yang
	// mereferensikan metric ini, DB akan menolak DELETE (FK constraint,
	// sengaja tanpa ON DELETE CASCADE) -- seeder akan gagal dan itu memang
	// diinginkan, supaya gak diam-diam menghapus metric yang masih dipakai.
	if err := db.Where("metric_key NOT IN ?", keepKeys).
		Delete(&models.DashboardMetric{}).Error; err != nil {
		return err
	}

	return nil
}
