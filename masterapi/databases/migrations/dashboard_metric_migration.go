package migrations

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// MigrateDashboardMetrics menjalankan seluruh penyesuaian skema dashboard_*
// dalam SATU transaksi. Aman dijalankan berkali-kali (idempotent).
// Tabel legacy (forms, form_fields, form_answer_fields, dll) TIDAK disentuh.
func MigrateDashboardMetrics(dbMaster *gorm.DB) error {
	return dbMaster.Transaction(func(tx *gorm.DB) error {
		steps := []struct {
			name string
			fn   func(*gorm.DB) error
		}{
			{"01_create_dashboard_metric_statuses", step01CreateStatusesTable},
			{"02_rename_is_dashboard_to_is_locked", step02RenameIsDashboard},
			{"03_backfill_statuses_from_status_key", step03BackfillStatuses},
			{"04_add_status_id_column", step04AddStatusIDColumn},
			{"05_populate_status_id_column", step05PopulateStatusID},
			{"06_precheck_data_before_drop", step06PreCheckBeforeDrop},
			{"07_drop_status_key_column", step07DropStatusKeyColumn},
			{"08_add_constraints_and_checks", step08AddConstraints},
			{"09_add_indexes", step09AddIndexes},
		}

		for _, s := range steps {
			if err := s.fn(tx); err != nil {
				return fmt.Errorf("migration step %s gagal: %w", s.name, err)
			}
		}
		return nil
	})
}

// ---------------------------------------------------------------------
// STEP 1: buat tabel dashboard_metric_statuses (tabel baru, aman dibuat
// dengan IF NOT EXISTS karena memang belum pernah ada).
// ---------------------------------------------------------------------
func step01CreateStatusesTable(tx *gorm.DB) error {
	return tx.Exec(`
		CREATE TABLE IF NOT EXISTS dashboard_metric_statuses (
			id                  BIGSERIAL PRIMARY KEY,
			dashboard_metric_id BIGINT NOT NULL REFERENCES dashboard_metrics(id) ON DELETE CASCADE,
			status_key          VARCHAR(100) NOT NULL,
			label               VARCHAR(191) NOT NULL,
			sequence            INTEGER NOT NULL DEFAULT 0,
			is_active           BOOLEAN NOT NULL DEFAULT true,
			created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
			CONSTRAINT uq_dashboard_metric_statuses_metric_key UNIQUE (dashboard_metric_id, status_key),
			CONSTRAINT chk_dashboard_metric_statuses_key_format CHECK (status_key ~ '^[a-z0-9_]+$')
		);
	`).Error
}

// ---------------------------------------------------------------------
// STEP 2: rename is_dashboard -> is_locked. Idempotent: cek dulu apakah
// kolom is_dashboard masih ada (kalau sudah pernah di-rename, skip).
// ---------------------------------------------------------------------
func step02RenameIsDashboard(tx *gorm.DB) error {
	var hasOldCol int64
	if err := tx.Raw(`
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'dashboard_metrics' AND column_name = 'is_dashboard'
	`).Scan(&hasOldCol).Error; err != nil {
		return err
	}

	if hasOldCol > 0 {
		if err := tx.Exec(`ALTER TABLE dashboard_metrics RENAME COLUMN is_dashboard TO is_locked`).Error; err != nil {
			return err
		}
	}

	// Pastikan kolom is_locked ada, NOT NULL, default false, tanpa null tersisa.
	// (Kolom lama is_dashboard bertipe *bool nullable — backfill null jadi false dulu.)
	if err := tx.Exec(`UPDATE dashboard_metrics SET is_locked = false WHERE is_locked IS NULL`).Error; err != nil {
		return err
	}
	if err := tx.Exec(`ALTER TABLE dashboard_metrics ALTER COLUMN is_locked SET DEFAULT false`).Error; err != nil {
		return err
	}
	return tx.Exec(`ALTER TABLE dashboard_metrics ALTER COLUMN is_locked SET NOT NULL`).Error
}

// ---------------------------------------------------------------------
// STEP 3: backfill status_key (text bebas) lama di dashboard_option_status_mappings
// menjadi baris master di dashboard_metric_statuses. Idempotent lewat
// ON CONFLICT DO NOTHING (unique constraint metric_id+status_key).
// Hanya jalan kalau kolom status_key masih ada (belum pernah di-drop).
// ---------------------------------------------------------------------
func step03BackfillStatuses(tx *gorm.DB) error {
	var hasStatusKeyCol int64
	if err := tx.Raw(`
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'dashboard_option_status_mappings' AND column_name = 'status_key'
	`).Scan(&hasStatusKeyCol).Error; err != nil {
		return err
	}
	if hasStatusKeyCol == 0 {
		return nil // sudah pernah di-drop sebelumnya, skip
	}

	return tx.Exec(`
		INSERT INTO dashboard_metric_statuses (dashboard_metric_id, status_key, label, sequence, is_active)
		SELECT DISTINCT
			dmm.dashboard_metric_id,
			dosm.status_key,
			dosm.status_key AS label, -- label sementara = key, admin bisa ubah manual nanti
			0,
			true
		FROM dashboard_option_status_mappings dosm
		JOIN dashboard_metric_mappings dmm ON dmm.id = dosm.dashboard_metric_mapping_id
		WHERE dosm.status_key IS NOT NULL AND dosm.status_key != ''
		ON CONFLICT (dashboard_metric_id, status_key) DO NOTHING
	`).Error
}

// ---------------------------------------------------------------------
// STEP 4: tambah kolom dashboard_metric_status_id (nullable dulu, supaya
// bisa diisi bertahap sebelum jadi NOT NULL).
// ---------------------------------------------------------------------
func step04AddStatusIDColumn(tx *gorm.DB) error {
	var exists int64
	if err := tx.Raw(`
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'dashboard_option_status_mappings' AND column_name = 'dashboard_metric_status_id'
	`).Scan(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	return tx.Exec(`
		ALTER TABLE dashboard_option_status_mappings
		ADD COLUMN dashboard_metric_status_id BIGINT
	`).Error
}

// ---------------------------------------------------------------------
// STEP 5: isi dashboard_metric_status_id berdasarkan status_key lama
// (join lewat mapping -> metric_id, sama seperti backfill di step 3).
// ---------------------------------------------------------------------
func step05PopulateStatusID(tx *gorm.DB) error {
	var hasStatusKeyCol int64
	if err := tx.Raw(`
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'dashboard_option_status_mappings' AND column_name = 'status_key'
	`).Scan(&hasStatusKeyCol).Error; err != nil {
		return err
	}
	if hasStatusKeyCol == 0 {
		return nil
	}

	// dosm (tabel target UPDATE) TIDAK boleh direferensikan di dalam
	// kondisi JOIN pada klausa FROM — hanya boleh muncul di WHERE.
	// dmm & dms di-join tanpa menyentuh dosm sama sekali; pencocokan
	// ke baris dosm yang sedang di-update dipindah semua ke WHERE.
	return tx.Exec(`
		UPDATE dashboard_option_status_mappings dosm
		SET dashboard_metric_status_id = dms.id
		FROM dashboard_metric_mappings dmm
		JOIN dashboard_metric_statuses dms
			ON dms.dashboard_metric_id = dmm.dashboard_metric_id
		WHERE dmm.id = dosm.dashboard_metric_mapping_id
			AND dms.status_key = dosm.status_key
			AND dosm.dashboard_metric_status_id IS NULL
			AND dosm.status_key IS NOT NULL
	`).Error
}

// ---------------------------------------------------------------------
// STEP 6: PRE-CHECK sebelum drop status_key. Kalau ada baris yang gagal
// ter-resolve (dashboard_metric_status_id masih NULL padahal status_key
// terisi), migrasi DIHENTIKAN di sini — jangan lanjut drop kolom lama,
// supaya data tidak hilang diam-diam.
// ---------------------------------------------------------------------
func step06PreCheckBeforeDrop(tx *gorm.DB) error {
	var hasStatusKeyCol int64
	if err := tx.Raw(`
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'dashboard_option_status_mappings' AND column_name = 'status_key'
	`).Scan(&hasStatusKeyCol).Error; err != nil {
		return err
	}
	if hasStatusKeyCol == 0 {
		return nil
	}

	var unresolved int64
	if err := tx.Raw(`
		SELECT count(*) FROM dashboard_option_status_mappings
		WHERE status_key IS NOT NULL AND status_key != '' AND dashboard_metric_status_id IS NULL
	`).Scan(&unresolved).Error; err != nil {
		return err
	}
	if unresolved > 0 {
		return fmt.Errorf("%d baris dashboard_option_status_mappings gagal di-backfill ke dashboard_metric_status_id, cek data manual dulu sebelum lanjut", unresolved)
	}

	// Cek referensial yatim tambahan yang disebut di KT: answer_option_id
	// harus masih ada di form_answer_fields (kalau sudah dihapus permanen
	// akan gagal saat FK CASCADE dipasang nanti — lebih baik ketahuan di sini).
	var orphanAnswerOption int64
	if err := tx.Raw(`
		SELECT count(*) FROM dashboard_option_status_mappings dosm
		LEFT JOIN form_answer_fields faf ON faf.id = dosm.answer_option_id
		WHERE faf.id IS NULL
	`).Scan(&orphanAnswerOption).Error; err != nil {
		return err
	}
	if orphanAnswerOption > 0 {
		return fmt.Errorf("%d baris dashboard_option_status_mappings menunjuk answer_option_id yang sudah tidak ada di form_answer_fields", orphanAnswerOption)
	}

	var orphanFormField int64
	if err := tx.Raw(`
		SELECT count(*) FROM dashboard_metric_mappings dmm
		LEFT JOIN form_fields ff ON ff.id = dmm.form_field_id
		WHERE ff.id IS NULL
	`).Scan(&orphanFormField).Error; err != nil {
		return err
	}
	if orphanFormField > 0 {
		return fmt.Errorf("%d baris dashboard_metric_mappings menunjuk form_field_id yang sudah tidak ada", orphanFormField)
	}

	// Cek metric_key / status_key format sebelum CHECK constraint dipasang.
	var badMetricKey int64
	if err := tx.Raw(`
		SELECT count(*) FROM dashboard_metrics WHERE metric_key !~ '^[a-z0-9_]+$'
	`).Scan(&badMetricKey).Error; err != nil {
		return err
	}
	if badMetricKey > 0 {
		return fmt.Errorf("%d dashboard_metrics.metric_key tidak sesuai format ^[a-z0-9_]+$", badMetricKey)
	}

	return nil
}

// ---------------------------------------------------------------------
// STEP 7: baru sekarang aman drop status_key lama.
// ---------------------------------------------------------------------
func step07DropStatusKeyColumn(tx *gorm.DB) error {
	var hasStatusKeyCol int64
	if err := tx.Raw(`
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'dashboard_option_status_mappings' AND column_name = 'status_key'
	`).Scan(&hasStatusKeyCol).Error; err != nil {
		return err
	}
	if hasStatusKeyCol == 0 {
		return nil
	}
	return tx.Exec(`ALTER TABLE dashboard_option_status_mappings DROP COLUMN status_key`).Error
}

// ---------------------------------------------------------------------
// STEP 8: pasang semua constraint final (NOT NULL, FK, UNIQUE, CHECK).
// Semua pakai DO block idempotent lewat pg_constraint check.
// ---------------------------------------------------------------------
func step08AddConstraints(tx *gorm.DB) error {
	stmts := []string{
		// dashboard_metric_status_id wajib diisi & FK RESTRICT
		`ALTER TABLE dashboard_option_status_mappings ALTER COLUMN dashboard_metric_status_id SET NOT NULL`,

		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_dosm_status'
			) THEN
				ALTER TABLE dashboard_option_status_mappings
				ADD CONSTRAINT fk_dosm_status
				FOREIGN KEY (dashboard_metric_status_id)
				REFERENCES dashboard_metric_statuses(id) ON DELETE RESTRICT;
			END IF;
		END $$;`,

		// UNIQUE(dashboard_metric_mapping_id, answer_option_id) — mungkin sudah ada, cek dulu
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'uq_dosm_mapping_option'
			) THEN
				ALTER TABLE dashboard_option_status_mappings
				ADD CONSTRAINT uq_dosm_mapping_option
				UNIQUE (dashboard_metric_mapping_id, answer_option_id);
			END IF;
		END $$;`,

		// UNIQUE(form_field_id) di dashboard_metric_mappings — satu pertanyaan = satu metrik
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'uq_dmm_form_field'
			) THEN
				ALTER TABLE dashboard_metric_mappings
				ADD CONSTRAINT uq_dmm_form_field
				UNIQUE (form_field_id);
			END IF;
		END $$;`,

		// UNIQUE(dashboard_metric_id, form_id) — pastikan ada (biasanya sudah)
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'uq_dmm_metric_form'
			) THEN
				ALTER TABLE dashboard_metric_mappings
				ADD CONSTRAINT uq_dmm_metric_form
				UNIQUE (dashboard_metric_id, form_id);
			END IF;
		END $$;`,

		// dashboard_metrics.metric_key format check
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_metric_key_format'
			) THEN
				ALTER TABLE dashboard_metrics
				ADD CONSTRAINT chk_metric_key_format
				CHECK (metric_key ~ '^[a-z0-9_]+$');
			END IF;
		END $$;`,

		// dashboard_metrics.expected_template whitelist
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_expected_template'
			) THEN
				ALTER TABLE dashboard_metrics
				ADD CONSTRAINT chk_expected_template
				CHECK (expected_template IN ('number','long-answer','multiple-choices','image-template','maps'));
			END IF;
		END $$;`,
	}

	for _, stmt := range stmts {
		if err := tx.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------
// STEP 9: index tambahan + hapus index redundan (sudah tercakup UNIQUE).
// ---------------------------------------------------------------------
func step09AddIndexes(tx *gorm.DB) error {
	stmts := []string{
		`DROP INDEX IF EXISTS idx_dashboard_metric_mappings_form_field_id`, // redundan dgn UNIQUE(form_field_id)
		`CREATE INDEX IF NOT EXISTS idx_dashboard_metric_statuses_metric_id ON dashboard_metric_statuses(dashboard_metric_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dosm_status_id ON dashboard_option_status_mappings(dashboard_metric_status_id)`,
	}
	for _, stmt := range stmts {
		if err := tx.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

var _ = errors.New
