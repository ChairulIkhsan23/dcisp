package institutions

import (
	"log"
	"net/http"
	"strconv"

	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Controller menangani endpoint HTTP katalog institusi eksternal API Indonesia.
type Controller struct {
	service *Service
}

// Menginisialisasi controller katalog institusi eksternal.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// Menangani pencarian institusi pendidikan pada Public API API Indonesia (kampus + sekolah).
func (ctrl *Controller) SearchExternalInstitutions(c *gin.Context) {
	query := c.Query("q")
	source := c.DefaultQuery("source", "ALL")
	province := c.Query("province")
	regency := c.Query("regency")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	result, err := ctrl.service.SearchExternalInstitutions(c.Request.Context(), query, source, province, regency, page, perPage)
	if err != nil {
		log.Printf("Gagal mencari institusi eksternal: %v", err)
		response.SafeBadRequest(c, "Permintaan institusi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Hasil pencarian institusi eksternal berhasil diambil dari API Indonesia", result)
}

// Menangani pengambilan detail institusi eksternal berdasarkan sumber dan ID eksternal.
func (ctrl *Controller) GetExternalInstitutionDetail(c *gin.Context) {
	source := c.Param("source")
	externalID := c.Param("external_id")
	if source == "" || externalID == "" {
		response.BadRequest(c, "Parameter sumber dan ID eksternal wajib diisi")
		return
	}

	detail, err := ctrl.service.GetExternalInstitutionDetail(c.Request.Context(), source, externalID)
	if err != nil {
		log.Printf("Gagal mengambil detail institusi eksternal: %v", err)
		response.SafeBadRequest(c, "Permintaan institusi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Detail institusi eksternal berhasil diambil", detail)
}

// Menangani impor satu institusi eksternal dari API Indonesia ke database internal secara idempoten.
func (ctrl *Controller) ImportExternalInstitution(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload impor institusi eksternal tidak valid", err.Error())
		return
	}

	inst, isNew, err := ctrl.service.ImportExternalInstitution(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal mengimpor institusi eksternal: %v", err)
		response.SafeBadRequest(c, "Permintaan institusi tidak dapat diproses", err)
		return
	}

	message := "Institusi eksternal berhasil diimpor ke database internal"
	if !isNew {
		message = "Institusi sudah pernah diimpor sebelumnya (idempoten)"
	}

	response.Success(c, http.StatusCreated, message, inst)
}
