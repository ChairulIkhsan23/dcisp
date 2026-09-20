package system

import (
	"log"
	"net/http"
	"strconv"

	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditController struct {
	service *AuditService
}

// Menginisialisasi instance baru controller audit logging.
func NewAuditController(service *AuditService) *AuditController {
	return &AuditController{service: service}
}

// Menangani permintaan HTTP untuk menampilkan daftar jejak audit sistem dengan filter dan paginasi.
func (ctrl *AuditController) List(c *gin.Context) {
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var userID *uuid.UUID
	if uStr := c.Query("user_id"); uStr != "" {
		if parsed, err := uuid.Parse(uStr); err == nil {
			userID = &parsed
		}
	}

	logs, total, err := ctrl.service.ListAuditLogs(c.Request.Context(), resourceType, resourceID, userID, page, limit)
	if err != nil {
		log.Printf("Gagal mengambil log audit: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil log audit")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{
		"page":       page,
		"limit":      limit,
		"totalItems": total,
		"totalPages": totalPages,
	}

	response.Success(c, http.StatusOK, "Daftar jejak audit berhasil diambil", logs, meta)
}

// Menangani permintaan HTTP untuk mengambil detail satu rekaman log audit berdasarkan ID.
func (ctrl *AuditController) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Format ID log audit tidak valid")
		return
	}

	record, err := ctrl.service.GetAuditLogByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("Gagal mengambil detail log audit: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil detail log audit")
		return
	}
	if record == nil {
		response.NotFound(c, "Log audit tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Detail log audit berhasil diambil", record)
}
