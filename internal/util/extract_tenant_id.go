package util

import (
	"github.com/dgrijalva/jwt-go"
)

type JWTClaim struct {
	jwt.StandardClaims
	Claims         jwt.Claims
	CustomTenantID string `json:"custom:tenant_id"`
}

func ExtractTenantID(IDToken string) string {
	claims := &JWTClaim{}
	bearerToken, _ := jwt.ParseWithClaims(IDToken, claims, nil)
	claims = bearerToken.Claims.(*JWTClaim)
	tokenTenantID := claims.CustomTenantID
	return tokenTenantID
}
