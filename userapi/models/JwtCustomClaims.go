package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type HealthServer struct{}

type JwtCustomClaims struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}
