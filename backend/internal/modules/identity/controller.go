package identity

import (
	"errors"
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
		response.InternalError(c, "Gagal mengautentikasi pengguna", err.Error())
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
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Token berhasil diperbarui", gin.H{
		"tokens": tokens,
	})
}

// Menangani permintaan HTTP untuk mengambil profil pengguna yang sedang terautentikasi.
func (ctrl *Controller) GetMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}

	userID := userIDVal.(uuid.UUID)
	profile, err := ctrl.service.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "Profil pengguna tidak ditemukan")
			return
		}
		response.InternalError(c, "Gagal mengambil profil pengguna", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profil pengguna berhasil diambil", profile)
}

// Menangani permintaan HTTP untuk mengambil daftar seluruh master role sistem.
func (ctrl *Controller) GetRoles(c *gin.Context) {
	roles, err := ctrl.service.GetRoles(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar peran sistem", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Daftar peran berhasil diambil", roles)
}
