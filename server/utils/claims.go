package utils

import "github.com/golang-jwt/jwt/v5"

type BaseClaims struct {
	UUID   string
	ID     uint
	RoleID uint
}

type CustomClaims struct {
	BaseClaims
	jwt.RegisteredClaims
}
