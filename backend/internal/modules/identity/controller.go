package identity

import (
	"errors"
	"log"
	"net/http"

	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

// Menginisialisasi instance baru identity controller.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// Menangani permintaan HTTP untuk autentikasi login pengguna dan penerbitan token.
func (ctrl *Controller) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload permintaan login tidak valid", err.Error())
		return
	}

	tokens, profile, err := ctrl.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, "Email atau kata sandi tidak valid")
			return
		}
		if errors.Is(err, ErrAccountInactive) {
			response.Unauthorized(c, "Akun tidak aktif, silakan hubungi administrator")
			return
		}
		log.Printf("Kesalahan internal saat login: %v", err)
		response.InternalError(c, "Terjadi kesalahan pada server saat autentikasi")
		return
	}

	response.Success(c, http.StatusOK, "Login berhasil", gin.H{
		"tokens":  tokens,
		"profile": profile,
	})
}

// Menangani permintaan HTTP untuk pembaruan pasangan token akses menggunakan refresh token.
func (ctrl *Controller) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload refresh token tidak valid", err.Error())
		return
	}

	tokens, err := ctrl.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Refresh token tidak valid atau telah kedaluwarsa")
		return
	}

	response.Success(c, http.StatusOK, "Token berhasil diperbarui", gin.H{
		"tokens": tokens,
	})
}

// Menangani permintaan HTTP untuk mencabut sesi refresh token perangkat saat ini (logout).
func (ctrl *Controller) Logout(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Konteks identitas pengguna tidak valid")
		return
	}

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload logout tidak valid", err.Error())
		return
	}

	if err := ctrl.service.Logout(c.Request.Context(), userID, req.RefreshToken); err != nil {
		log.Printf("Kesalahan internal saat logout: %v", err)
		response.InternalError(c, "Terjadi kesalahan pada server saat logout")
		return
	}

	response.Success(c, http.StatusOK, "Logout berhasil, sesi perangkat dicabut", nil)
}

// Menangani permintaan HTTP untuk mengambil profil pengguna yang sedang terautentikasi.
func (ctrl *Controller) GetMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			response.Unauthorized(c, "Konteks identitas pengguna tidak valid")
			return
		}
		userID = parsed
	default:
		response.Unauthorized(c, "Konteks identitas pengguna tidak valid")
		return
	}

	profile, err := ctrl.service.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "Profil pengguna tidak ditemukan")
			return
		}
		if errors.Is(err, ErrAccountInactive) {
			response.Unauthorized(c, "Akun pengguna tidak aktif")
			return
		}
		log.Printf("Kesalahan internal saat mengambil profil: %v", err)
		response.InternalError(c, "Terjadi kesalahan pada server saat mengambil profil")
		return
	}

	response.Success(c, http.StatusOK, "Profil pengguna berhasil diambil", profile)
}

// Menangani permintaan HTTP untuk mengambil daftar seluruh master role sistem.
func (ctrl *Controller) GetRoles(c *gin.Context) {
	roles, err := ctrl.service.GetRoles(c.Request.Context())
	if err != nil {
		log.Printf("Kesalahan internal saat mengambil daftar peran: %v", err)
		response.InternalError(c, "Terjadi kesalahan pada server saat mengambil daftar peran")
		return
	}

	response.Success(c, http.StatusOK, "Daftar peran berhasil diambil", roles)
}
