package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Roles     []string  `json:"roles,omitempty"`
	Scopes    []string  `json:"scopes"`
	TokenType string    `json:"type"`
	TokenID   string    `json:"jti,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // in seconds
	TokenType    string `json:"token_type"`
}

// Menghasilkan pasangan token akses JWT berdurasi 15 menit dan token refresh berdurasi 7 hari dengan pemisahan tipe yang ketat.
func GenerateTokenPair(secret string, accessMinutes, refreshDays int, userID uuid.UUID, email, role string, scopes []string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(time.Duration(accessMinutes) * time.Minute)
	refreshExpiry := now.Add(time.Duration(refreshDays) * 24 * time.Hour)

	// 1. Access Token
	accessClaims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		Scopes:    scopes,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID.String(),
			Issuer:    "dcisp-auth-service",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani token akses: %w", err)
	}

	// 2. Refresh Token (membawa TokenID unik untuk rotasi dan revocation)
	refreshClaims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		Scopes:    scopes,
		TokenType: TokenTypeRefresh,
		TokenID:   uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID.String(),
			Issuer:    "dcisp-auth-service",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
		ExpiresIn:    int64(accessMinutes * 60),
		TokenType:    "Bearer",
	}, nil
}

// Memvalidasi keabsahan token JWT dan memastikan tipe token sesuai dengan yang diharapkan (access atau refresh).
func ValidateToken(secret, tokenString, expectedType string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode penandatanganan tidak diharapkan: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		if claims.UserID == uuid.Nil && claims.Subject != "" {
			if parsed, err := uuid.Parse(claims.Subject); err == nil {
				claims.UserID = parsed
			}
		}
		if claims.UserID == uuid.Nil {
			return nil, errors.New("payload token tidak valid: user_id tidak ditemukan")
		}

		if claims.TokenType == "" {
			return nil, errors.New("payload token tidak valid: tipe token tidak ditemukan")
		}

		if expectedType != "" && claims.TokenType != expectedType {
			return nil, fmt.Errorf("tipe token tidak valid, diharapkan '%s' tetapi ditemukan '%s'", expectedType, claims.TokenType)
		}

		return claims, nil
	}

	return nil, errors.New("token tidak valid")
}

// Memvalidasi token refresh sekaligus memastikan klaim TokenID tersedia untuk rotasi sesi.
func ValidateRefreshToken(secret, tokenString string) (*JWTClaims, error) {
	claims, err := ValidateToken(secret, tokenString, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}
	if claims.TokenID == "" {
		return nil, errors.New("refresh token lawas tanpa identitas sesi tidak didukung, silakan login ulang")
	}
	return claims, nil
}
