package system

import (
	"log"
	"net/http"

	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type SettingsController struct {
	service *SettingsService
}

// Menginisialisasi instance baru controller pengaturan sistem terpusat.
func NewSettingsController(service *SettingsService) *SettingsController {
	return &SettingsController{service: service}
}

// Menangani permintaan HTTP untuk menetapkan atau memperbarui nilai pengaturan sistem.
func (ctrl *SettingsController) SetSetting(c *gin.Context) {
	var req SetSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pengaturan sistem tidak valid", err.Error())
		return
	}

	setting, err := ctrl.service.SetSetting(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal menetapkan pengaturan sistem: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Pengaturan sistem berhasil disimpan", setting)
}

// Menangani permintaan HTTP untuk mengambil daftar seluruh pengaturan sistem.
func (ctrl *SettingsController) ListSettings(c *gin.Context) {
	settings, err := ctrl.service.ListSettings(c.Request.Context())
	if err != nil {
		log.Printf("Gagal mengambil daftar pengaturan sistem: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil daftar pengaturan")
		return
	}

	response.Success(c, http.StatusOK, "Daftar pengaturan sistem berhasil diambil", settings)
}

// Menangani permintaan HTTP untuk mengambil satu nilai pengaturan sistem berdasarkan kunci.
func (ctrl *SettingsController) GetSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "Kunci pengaturan sistem wajib diisi")
		return
	}

	setting, err := ctrl.service.GetSetting(c.Request.Context(), key)
	if err != nil {
		response.NotFound(c, "Pengaturan sistem tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Pengaturan sistem berhasil diambil", setting)
}

// Menangani permintaan HTTP untuk menghapus pengaturan sistem berdasarkan kunci unik.
func (ctrl *SettingsController) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "Kunci pengaturan sistem wajib diisi")
		return
	}

	if err := ctrl.service.DeleteSetting(c.Request.Context(), key); err != nil {
		log.Printf("Gagal menghapus pengaturan sistem: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Pengaturan sistem berhasil dihapus", nil)
}
