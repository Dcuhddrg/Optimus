package utils

import "github.com/golang-jwt/jwt/v5"

const ClaimsContextKey = "claims"

type BaseClaims struct {
	UUID      string   `json:"uuid"`
	ID        uint     `json:"id"`
	Username  string   `json:"username"`
	RoleIDs   []uint   `json:"roleIds"`
	RoleCodes []string `json:"roleCodes"`
}

type CustomClaims struct {
	BaseClaims
	jwt.RegisteredClaims
}
