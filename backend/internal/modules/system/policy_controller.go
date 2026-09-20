package system

import (
	"log"
	"net/http"
	"strconv"

	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PolicyController struct {
	service *PolicyService
}

// Menginisialisasi instance baru controller policy engine.
func NewPolicyController(service *PolicyService) *PolicyController {
	return &PolicyController{service: service}
}

// Menangani permintaan HTTP untuk membuat kebijakan bisnis baru.
func (ctrl *PolicyController) Create(c *gin.Context) {
	var req CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan kebijakan tidak valid", err.Error())
		return
	}

	p, err := ctrl.service.CreatePolicy(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal membuat kebijakan: %v", err)
		response.SafeBadRequest(c, "Permintaan kebijakan tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Kebijakan berhasil dibuat", p)
}

// Menangani permintaan HTTP untuk mengambil detail satu kebijakan berdasarkan ID.
func (ctrl *PolicyController) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Format ID kebijakan tidak valid")
		return
	}

	p, err := ctrl.service.GetPolicyByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Kebijakan tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Detail kebijakan berhasil diambil", p)
}

// Menangani permintaan HTTP untuk mengambil daftar kebijakan dengan filter dan paginasi.
func (ctrl *PolicyController) List(c *gin.Context) {
	domain := c.Query("domain")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		activeBool := activeStr == "true" || activeStr == "1"
		isActive = &activeBool
	}

	policies, total, err := ctrl.service.ListPolicies(c.Request.Context(), domain, isActive, page, limit)
	if err != nil {
		log.Printf("Gagal mengambil daftar kebijakan: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil daftar kebijakan")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{
		"page":       page,
		"limit":      limit,
		"totalItems": total,
		"totalPages": totalPages,
	}

	response.Success(c, http.StatusOK, "Daftar kebijakan berhasil diambil", policies, meta)
}

// Menangani permintaan HTTP untuk memperbarui kebijakan yang ada.
func (ctrl *PolicyController) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Format ID kebijakan tidak valid")
		return
	}

	var req UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembaruan kebijakan tidak valid", err.Error())
		return
	}

	p, err := ctrl.service.UpdatePolicy(c.Request.Context(), id, &req)
	if err != nil {
		log.Printf("Gagal memperbarui kebijakan: %v", err)
		response.SafeBadRequest(c, "Permintaan kebijakan tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Kebijakan berhasil diperbarui", p)
}

// Menangani permintaan HTTP untuk menghapus kebijakan dari sistem.
func (ctrl *PolicyController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Format ID kebijakan tidak valid")
		return
	}

	if err := ctrl.service.DeletePolicy(c.Request.Context(), id); err != nil {
		log.Printf("Gagal menghapus kebijakan: %v", err)
		response.SafeBadRequest(c, "Permintaan kebijakan tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Kebijakan berhasil dihapus", nil)
}

// Menangani permintaan HTTP untuk mengevaluasi aturan kebijakan aktif terhadap data konteks.
func (ctrl *PolicyController) Evaluate(c *gin.Context) {
	var req PolicyEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload evaluasi kebijakan tidak valid", err.Error())
		return
	}

	result, err := ctrl.service.EvaluatePolicy(c.Request.Context(), req.Domain, req.Context)
	if err != nil {
		log.Printf("Gagal mengevaluasi kebijakan: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengevaluasi kebijakan")
		return
	}

	response.Success(c, http.StatusOK, "Evaluasi kebijakan berhasil", result)
}
