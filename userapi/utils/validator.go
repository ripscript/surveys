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
