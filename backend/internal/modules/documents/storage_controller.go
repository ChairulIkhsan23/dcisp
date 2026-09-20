package documents

import (
	"log"
	"net/http"
	"strconv"

	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StorageController struct {
	service *StorageService
}

// Menginisialisasi instance baru controller penyimpanan berkas.
func NewStorageController(service *StorageService) *StorageController {
	return &StorageController{service: service}
}

// Menangani permintaan HTTP untuk menghasilkan presigned URL unggah berkas ke Cloudflare R2.
func (ctrl *StorageController) PresignedUpload(c *gin.Context) {
	var req PresignedUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload permohonan unggah presigned tidak valid", err.Error())
		return
	}

	res, err := ctrl.service.GeneratePresignedUpload(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal menghasilkan presigned upload URL: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "URL unggah presigned berhasil dibuat", res)
}

// Menangani permintaan HTTP untuk mengambil metadata informasi berkas berdasarkan ID.
func (ctrl *StorageController) GetMetadata(c *gin.Context) {
	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Format ID berkas tidak valid")
		return
	}

	meta, err := ctrl.service.GetFileMetadata(c.Request.Context(), fileID)
	if err != nil {
		log.Printf("Gagal mengambil metadata berkas: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil metadata berkas")
		return
	}
	if meta == nil {
		response.NotFound(c, "Metadata berkas tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Metadata berkas berhasil diambil", meta)
}

// Menangani permintaan HTTP untuk menghasilkan presigned URL unduh berkas privat.
func (ctrl *StorageController) GetDownloadURL(c *gin.Context) {
	idParam := c.Param("id")
	fileID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Format ID berkas tidak valid")
		return
	}

	expiresIn, _ := strconv.Atoi(c.DefaultQuery("expires_in", "3600"))

	downloadURL, err := ctrl.service.GeneratePresignedDownload(c.Request.Context(), fileID, expiresIn)
	if err != nil {
		log.Printf("Gagal menghasilkan presigned download URL: %v", err)
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "URL unduh berhasil dibuat", gin.H{
		"download_url":       downloadURL,
		"expires_in_seconds": expiresIn,
	})
}
