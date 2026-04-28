package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type HealthServer struct{}

type PermissionAction struct {
	V bool `json:"v"`
	C bool `json:"c"`
	U bool `json:"u"`
	D bool `json:"d"`
}

type JwtCustomClaims struct {
	ID           int64                       `json:"id"`
	RespondentID int64                       `json:"respondent_id"`
	Name         string                      `json:"name"`
	Email        string                      `json:"email"`
	Role         int                         `json:"role"`
	Permissions  map[string]PermissionAction `json:"perm"`
	jwt.RegisteredClaims
}
