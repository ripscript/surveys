package seeders

import (
	"errors"

	"backend/masterapi/models"
	"backend/masterapi/utils"

	"gorm.io/gorm"
)

func CMSSectionSeed(db *gorm.DB) error {
	sections := []models.CMSSection{
		{Slug: "hero", Name: "Hero Banner", Type: models.SectionTypeContent, SectionOrder: 100, IsRepeatable: false, IsSystem: true, IsEditable: utils.BoolToPointer(true), Status: true},
		{Slug: "sop", Name: "Standar Operasional Prosedur", Description: utils.StringToPointer("Menjamin pengelolaan Laci RW berjalan tertib, transparan, dan akuntabel dalam menyimpan,\n mendistribusikan, serta mengarsipkan dokumen/pertanyaan/informasi."), Type: models.SectionTypeItems, SectionOrder: 200, IsRepeatable: true, IsSystem: true, IsEditable: utils.BoolToPointer(true), Status: true},
		{Slug: "fakta-menarik", Name: "Fakta Menarik Kota Bandung", Description: utils.StringToPointer("Hal-Hal Seru yang Mungkin Belum Kamu Tahu dan Lebih Dekat dengan Kota Bandung"), Type: models.SectionTypeItems, SectionOrder: 300, IsRepeatable: true, IsSystem: true, IsEditable: utils.BoolToPointer(true), Status: true},
		{Slug: "peta-sebaran", Name: "Peta Sebaran Data Kota Bandung", Description: utils.StringToPointer("Menampilkan distribusi data berdasarkan wilayah kecamatan dan kelurahan di Kota Bandung."), Type: models.SectionTypeMap, SectionOrder: 400, IsRepeatable: false, IsSystem: true, IsEditable: utils.BoolToPointer(false), Status: true},
		{Slug: "tabel-kecamatan", Name: "Tabel Data Kecamatan Kota Bandung", Type: models.SectionTypeTable, SectionOrder: 500, IsRepeatable: false, IsSystem: true, IsEditable: utils.BoolToPointer(false), Status: true},
		{Slug: "footer", Name: "Footer", Type: models.SectionTypeContent, SectionOrder: 999999, IsRepeatable: false, IsSystem: true, IsEditable: utils.BoolToPointer(true), Status: true},
	}

	var keepSlugs []string

	for _, section := range sections {
		keepSlugs = append(keepSlugs, section.Slug)

		var existing models.CMSSection
		err := db.Where("slug = ?", section.Slug).First(&existing).Error
		isNew := errorsIsNotFound(err)

		if err != nil && !isNew {
			return err
		}

		if isNew {
			if err := db.Create(&section).Error; err != nil {
				return err
			}
			existing = section
		} else {
			// PERBAIKAN: Selalu sinkronkan SectionOrder dengan yang ada di code!
			// Type tetap tidak diubah karena bisa merusak integritas relasi data jika isinya sudah beda
			existing.SectionOrder = section.SectionOrder

			// field lain ikut disync:
			existing.Name = section.Name
			existing.IsRepeatable = section.IsRepeatable
			existing.Description = section.Description
			existing.IsEditable = section.IsEditable
			existing.IsSystem = section.IsSystem
			existing.Status = section.Status

			if err := db.Save(&existing).Error; err != nil {
				return err
			}
		}

		switch section.Slug {
		case "hero":
			if err := seedHeroContent(db, existing.ID); err != nil {
				return err
			}
		case "sop":
			if err := seedSopItems(db, existing.ID); err != nil {
				return err
			}
		case "fakta-menarik":
			if err := seedFaktaItems(db, existing.ID); err != nil {
				return err
			}
		case "footer":
			if err := seedFooterContent(db, existing.ID); err != nil {
				return err
			}
		}
		// peta-sebaran & tabel-kecamatan tidak butuh seedContent/Item
		// karena datanya ambil langsung dari tabel wilayah/kecamatan, bukan dari CMS
	}

	// ===== Sync penghapusan section starter lama =====
	// Hapus section is_system=true yang slug-nya sudah tidak ada lagi di kode
	// (misal karena rename/hapus di atas). Section dinamis (is_system=false)
	// TIDAK disentuh sama sekali oleh seeder — itu murni milik admin.
	if err := db.Where("is_system = ? AND slug NOT IN ?", true, keepSlugs).
		Delete(&models.CMSSection{}).Error; err != nil {
		return err
	}

	return nil
}

// ===== Helper sync untuk CMSContent (key-value per section) =====
func upsertContents(db *gorm.DB, sectionID int, contents []models.CMSContent) error {
	var keepKeys []string

	for _, c := range contents {
		keepKeys = append(keepKeys, c.Key)

		var existing models.CMSContent
		err := db.Where("section_id = ? AND key = ?", c.SectionID, c.Key).First(&existing).Error

		if errorsIsNotFound(err) {
			if err := db.Create(&c).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}

		existing.ValueText = c.ValueText
		existing.Status = c.Status
		if err := db.Save(&existing).Error; err != nil {
			return err
		}
	}

	// hapus content lama di section ini yang key-nya sudah tidak ada di code
	return db.Where("section_id = ? AND key NOT IN ?", sectionID, keepKeys).
		Delete(&models.CMSContent{}).Error
}

// ===== Helper sync untuk CMSItem (berdasarkan section_id + item_order) =====
func upsertItems(db *gorm.DB, sectionID int, items []models.CMSItem) error {
	var keepOrders []int

	for _, item := range items {
		keepOrders = append(keepOrders, item.ItemOrder)

		var existing models.CMSItem
		err := db.Where("section_id = ? AND item_order = ?", item.SectionID, item.ItemOrder).First(&existing).Error

		if errorsIsNotFound(err) {
			if err := db.Create(&item).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}

		existing.Title = item.Title
		existing.Description = item.Description
		existing.Category = item.Category
		existing.Image = item.Image
		existing.Status = item.Status
		if err := db.Save(&existing).Error; err != nil {
			return err
		}
	}

	// hapus item lama di section ini yang item_order-nya sudah tidak ada di code
	return db.Where("section_id = ? AND item_order NOT IN ?", sectionID, keepOrders).
		Delete(&models.CMSItem{}).Error
}

func seedHeroContent(db *gorm.DB, sectionID int) error {
	description := "Portal Layanan Catatan Informasi Rukun Warga. Hadir untuk memenuhi kebutuhan pendataan menuju gerbang satu data secara sederhana, portal digitalisasi data kewilayahan kota Bandungs"
	buttonText := "Download Daftar Pertanyaan"
	buttonLink := ""

	contents := []models.CMSContent{
		{SectionID: sectionID, Key: "description", ValueText: &description, Status: true},
		{SectionID: sectionID, Key: "button_text", ValueText: &buttonText, Status: true},
		{SectionID: sectionID, Key: "button_link", ValueText: &buttonLink, Status: true},
	}

	return upsertContents(db, sectionID, contents)
}

func seedSopItems(db *gorm.DB, sectionID int) error {
	desc1 := "Tahap verifikasi kebutuhan data warga di tingkat RT dan RW secara faktual."
	desc2 := "Petugas menginput data kependudukan ke sistem Laci RW dengan standar keamanan tinggi."
	desc3 := "Data divalidasi dan dikirim ke pusat Satu Data Bandung untuk integrasi layanan publik."

	items := []models.CMSItem{
		{SectionID: sectionID, ItemOrder: 1, Title: "Identifikasi Kebutuhan", Description: &desc1, Status: true},
		{SectionID: sectionID, ItemOrder: 2, Title: "Input Data LaCI RW", Description: &desc2, Status: true},
		{SectionID: sectionID, ItemOrder: 3, Title: "Output & Validasi", Description: &desc3, Status: true},
	}

	return upsertItems(db, sectionID, items)
}

func seedFaktaItems(db *gorm.DB, sectionID int) error {
	desc1 := "Technische Hoogeschool te Bandoeng (THS) di Kota Bandung, didirikan pada 3 Juli 1920, adalah perguruan tinggi teknik pertama di Hindia Belanda (sekarang Indonesia). Institusi ini berlokasi di Jalan Ganesha No. 10, yang saat ini dikenal sebagai Institut Teknologi Bandung (ITB), dan didirikan untuk memenuhi kebutuhan tenaga teknik. Kota Bandung tempat berdirinya Perguruan Tinggi Teknik Pertama di Hindia Belanda."
	category1 := "Sejarah"
	image1 := ""

	items := []models.CMSItem{
		{
			SectionID:   sectionID,
			ItemOrder:   1,
			Title:       "Bandung, Pelopor Pendidikan Teknik Pertama Di Hindia Belanda",
			Description: &desc1,
			Category:    &category1,
			Image:       &image1,
			Status:      true,
		},
	}

	return upsertItems(db, sectionID, items)
}

func seedFooterContent(db *gorm.DB, sectionID int) error {
	orgName := "Badan Perencanaan, Pembangunan, Penelitian dan Pengembangan Kota Bandung"
	address := "Jalan Aceh No. 36 Bandung Indonesia"
	phone := "(022)4222315"
	facebook := ""
	twitter := ""
	instagram := ""
	youtube := ""
	copyright := "2024 © Badan Perencanaan, Pembangunan, Penelitian dan Pengembangan Kota Bandung"

	contents := []models.CMSContent{
		{SectionID: sectionID, Key: "org_name", ValueText: &orgName, Status: true},
		{SectionID: sectionID, Key: "address", ValueText: &address, Status: true},
		{SectionID: sectionID, Key: "phone", ValueText: &phone, Status: true},
		{SectionID: sectionID, Key: "social_facebook", ValueText: &facebook, Status: true},
		{SectionID: sectionID, Key: "social_twitter", ValueText: &twitter, Status: true},
		{SectionID: sectionID, Key: "social_instagram", ValueText: &instagram, Status: true},
		{SectionID: sectionID, Key: "social_youtube", ValueText: &youtube, Status: true},
		{SectionID: sectionID, Key: "copyright", ValueText: &copyright, Status: true},
	}

	return upsertContents(db, sectionID, contents)
}

// errorsIsNotFound pakai errors.Is supaya tetap benar meski error di-wrap
// oleh layer lain (misal middleware/hook GORM custom di masa depan).
func errorsIsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
