package migrations

import (
	"backend/userapi/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	db.Exec("ALTER TABLE respondents DROP CONSTRAINT IF EXISTS respondents_nik_unique;")
	db.Exec("DROP INDEX IF EXISTS respondents_nik_unique;")

	err := db.AutoMigrate(&models.UserModel{}, &models.RespondentModel{}, &models.Menu{}, &models.MenuPermission{}, &models.Survey{})
	if err != nil {
		return err
	}
	return nil
}
