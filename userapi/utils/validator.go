package utils

import (
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

func PasswordRuleValidation(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var hasUpper, hasDigit, hasSpecial bool
	special := "@$!%*?&#_+*"

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsDigit(c):
			hasDigit = true
		case strings.ContainsRune(special, c):
			hasSpecial = true
		}
	}

	return hasUpper && hasDigit && hasSpecial
}

func EmailOptionalValidation(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		// Kosong/null -> tidak divalidasi, dianggap valid.
		return true
	}
	// Terisi -> wajib format email yang valid. Gunakan validator email bawaan.
	return validator.New().Var(value, "email") == nil
}
