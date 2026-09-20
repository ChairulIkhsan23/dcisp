package people

import (
	"log"
	"net/http"
	"strconv"

	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

// Menginisialisasi instance baru controller people and lifecycle management.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// ============================================================================
// 1. INSTITUTIONS
// ============================================================================

// Menangani permintaan HTTP untuk membuat institusi pendidikan mitra baru.
func (ctrl *Controller) CreateInstitution(c *gin.Context) {
	var req CreateInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan institusi tidak valid", err.Error())
		return
	}

	inst, err := ctrl.service.CreateInstitution(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal membuat institusi: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Institusi mitra berhasil dibuat", inst)
}

// Menangani permintaan HTTP untuk mengambil detail institusi berdasarkan ID.
func (ctrl *Controller) GetInstitutionByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID institusi tidak valid")
		return
	}

	inst, err := ctrl.service.GetInstitutionByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Institusi tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Detail institusi berhasil diambil", inst)
}

// Menangani permintaan HTTP untuk mengambil daftar seluruh institusi mitra.
func (ctrl *Controller) ListInstitutions(c *gin.Context) {
	list, err := ctrl.service.ListInstitutions(c.Request.Context())
	if err != nil {
		log.Printf("Gagal mengambil daftar institusi: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil daftar institusi")
		return
	}

	response.Success(c, http.StatusOK, "Daftar institusi berhasil diambil", list)
}

// Menangani permintaan HTTP untuk memperbarui data institusi pendidikan mitra.
func (ctrl *Controller) UpdateInstitution(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID institusi tidak valid")
		return
	}

	var req UpdateInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembaruan institusi tidak valid", err.Error())
		return
	}

	inst, err := ctrl.service.UpdateInstitution(c.Request.Context(), id, &req)
	if err != nil {
		log.Printf("Gagal memperbarui institusi: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Data institusi berhasil diperbarui", inst)
}

// Menangani permintaan HTTP untuk menghapus institusi mitra dari sistem.
func (ctrl *Controller) DeleteInstitution(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID institusi tidak valid")
		return
	}

	if err := ctrl.service.DeleteInstitution(c.Request.Context(), id); err != nil {
		log.Printf("Gagal menghapus institusi: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Institusi berhasil dihapus", nil)
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
		response.BadRequest(c, err.Error())
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
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Detail institusi eksternal berhasil diambil", detail)
}

// Menangani impor satu institusi eksternal dari API Indonesia ke database internal secara idempoten.
func (ctrl *Controller) ImportExternalInstitution(c *gin.Context) {
	var req ImportExternalInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload impor institusi eksternal tidak valid", err.Error())
		return
	}

	inst, isNew, err := ctrl.service.ImportExternalInstitution(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal mengimpor institusi eksternal: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	message := "Institusi eksternal berhasil diimpor ke database internal"
	if !isNew {
		message = "Institusi sudah pernah diimpor sebelumnya (idempoten)"
	}

	response.Success(c, http.StatusCreated, message, inst)
}

// ============================================================================
// 2. BATCHES
// ============================================================================

// Menangani permintaan HTTP untuk membuat batch baru beserta inisialisasi kas batch.
func (ctrl *Controller) CreateBatch(c *gin.Context) {
	var req CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan batch tidak valid", err.Error())
		return
	}

	batch, fund, err := ctrl.service.CreateBatch(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal membuat batch: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Batch dan kas angkatan berhasil dibuat", gin.H{
		"batch": batch,
		"fund":  fund,
	})
}

// Menangani permintaan HTTP untuk mengambil detail satu batch beserta kas angkatannya.
func (ctrl *Controller) GetBatchByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID batch tidak valid")
		return
	}

	batch, fund, err := ctrl.service.GetBatchByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Batch tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Detail batch berhasil diambil", gin.H{
		"batch": batch,
		"fund":  fund,
	})
}

// Menangani permintaan HTTP untuk menampilkan daftar seluruh batch.
func (ctrl *Controller) ListBatches(c *gin.Context) {
	status := c.Query("status")
	list, err := ctrl.service.ListBatches(c.Request.Context(), status)
	if err != nil {
		log.Printf("Gagal mengambil daftar batch: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil daftar batch")
		return
	}

	response.Success(c, http.StatusOK, "Daftar batch berhasil diambil", list)
}

// Menangani permintaan HTTP untuk memperbarui konfigurasi kohort batch.
func (ctrl *Controller) UpdateBatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID batch tidak valid")
		return
	}

	var req UpdateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembaruan batch tidak valid", err.Error())
		return
	}

	batch, err := ctrl.service.UpdateBatch(c.Request.Context(), id, &req)
	if err != nil {
		log.Printf("Gagal memperbarui batch: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Batch berhasil diperbarui", batch)
}

// ============================================================================
// 3. INTERNS & STATE MACHINE (FR-002)
// ============================================================================

// Menangani permintaan HTTP pendaftaran peserta magang baru dengan status awal APPLICANT.
func (ctrl *Controller) RegisterIntern(c *gin.Context) {
	var req RegisterInternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pendaftaran magang tidak valid", err.Error())
		return
	}

	intern, err := ctrl.service.RegisterIntern(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal mendaftarkan peserta magang: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Pendaftaran peserta magang berhasil", intern)
}

// Menangani permintaan HTTP untuk memproses transisi status siklus hidup peserta magang.
func (ctrl *Controller) ChangeInternStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID peserta magang tidak valid")
		return
	}

	var req ChangeInternStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload perubahan status tidak valid", err.Error())
		return
	}

	reason := ""
	if req.Reason != nil {
		reason = *req.Reason
	}

	intern, err := ctrl.service.ChangeInternStatus(c.Request.Context(), id, req.Status, reason)
	if err != nil {
		log.Printf("Gagal mengubah status peserta magang: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Status peserta magang berhasil diperbarui", intern)
}

// Menangani permintaan HTTP untuk mengambil profil peserta magang berdasarkan ID.
func (ctrl *Controller) GetInternByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID peserta magang tidak valid")
		return
	}

	intern, err := ctrl.service.GetInternByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Peserta magang tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Profil peserta magang berhasil diambil", intern)
}

// Menangani permintaan HTTP untuk mengambil profil peserta magang berdasarkan user ID akun.
func (ctrl *Controller) GetInternByUserID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.BadRequest(c, "Format ID pengguna tidak valid")
		return
	}

	intern, err := ctrl.service.GetInternByUserID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "Peserta magang tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Profil peserta magang berhasil diambil", intern)
}

// ============================================================================
// 4. USER MANAGEMENT (SEARCH, MENTOR ASSIGNMENT, BATCH PLOTTING - T-032)
// ============================================================================

// Menangani permintaan HTTP untuk mencari dan memfilter daftar peserta magang dengan paginasi.
func (ctrl *Controller) SearchInterns(c *gin.Context) {
	search := c.Query("q")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var batchID *uuid.UUID
	if bStr := c.Query("batch_id"); bStr != "" {
		if parsed, err := uuid.Parse(bStr); err == nil {
			batchID = &parsed
		}
	}

	var institutionID *uuid.UUID
	if iStr := c.Query("institution_id"); iStr != "" {
		if parsed, err := uuid.Parse(iStr); err == nil {
			institutionID = &parsed
		}
	}

	interns, total, err := ctrl.service.SearchInterns(c.Request.Context(), search, batchID, institutionID, status, page, limit)
	if err != nil {
		log.Printf("Gagal mencari peserta magang: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mencari peserta magang")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{
		"page":       page,
		"limit":      limit,
		"totalItems": total,
		"totalPages": totalPages,
	}

	response.Success(c, http.StatusOK, "Pencarian peserta magang berhasil", interns, meta)
}

// Menangani permintaan HTTP penugasan pembimbing (mentor) bagi peserta magang.
func (ctrl *Controller) AssignMentor(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID peserta magang tidak valid")
		return
	}

	var req AssignMentorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload penugasan pembimbing tidak valid", err.Error())
		return
	}

	intern, err := ctrl.service.AssignMentor(c.Request.Context(), id, req.MentorID)
	if err != nil {
		log.Printf("Gagal menetapkan pembimbing: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Pembimbing berhasil ditetapkan", intern)
}

// Menangani permintaan HTTP untuk memindahkan atau menetapkan peserta ke batch baru (batch plotting).
func (ctrl *Controller) PlotBatch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID peserta magang tidak valid")
		return
	}

	var req AssignBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload plotting batch tidak valid", err.Error())
		return
	}

	intern, err := ctrl.service.PlotBatch(c.Request.Context(), id, req.BatchID)
	if err != nil {
		log.Printf("Gagal plotting batch: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Plotting batch peserta berhasil", intern)
}

// ============================================================================
// 5. ALUMNI (FR-003, BR-003)
// ============================================================================

// Menangani permintaan HTTP untuk menampilkan direktori alumni dengan paginasi.
func (ctrl *Controller) ListAlumni(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	list, total, err := ctrl.service.ListAlumni(c.Request.Context(), page, limit)
	if err != nil {
		log.Printf("Gagal mengambil direktori alumni: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil direktori alumni")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{
		"page":       page,
		"limit":      limit,
		"totalItems": total,
		"totalPages": totalPages,
	}

	response.Success(c, http.StatusOK, "Direktori alumni berhasil diambil", list, meta)
}

// Menangani permintaan HTTP untuk mengambil profil alumni berdasarkan user ID akun.
func (ctrl *Controller) GetAlumniByUserID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.BadRequest(c, "Format ID pengguna tidak valid")
		return
	}

	alumni, err := ctrl.service.GetAlumniByUserID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "Profil alumni tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Profil alumni berhasil diambil", alumni)
}

// ============================================================================
// 6. SKILLS & USER SKILLS (FR-006)
// ============================================================================

// Menangani permintaan HTTP pembuatan master skill baru dalam katalog.
func (ctrl *Controller) CreateSkill(c *gin.Context) {
	var req CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan skill tidak valid", err.Error())
		return
	}

	skill, err := ctrl.service.CreateSkill(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal membuat master skill: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Master skill berhasil dibuat", skill)
}

// Menangani permintaan HTTP untuk mengambil seluruh daftar katalog skill.
func (ctrl *Controller) ListSkills(c *gin.Context) {
	category := c.Query("category")
	skills, err := ctrl.service.ListSkills(c.Request.Context(), category)
	if err != nil {
		log.Printf("Gagal mengambil katalog skill: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil katalog skill")
		return
	}

	response.Success(c, http.StatusOK, "Katalog skill berhasil diambil", skills)
}

// Menangani permintaan HTTP untuk menetapkan keahlian dan tingkat kemahiran pada pengguna.
func (ctrl *Controller) AssignUserSkill(c *gin.Context) {
	var req AssignUserSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload penugasan skill tidak valid", err.Error())
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Format user_id tidak valid")
		return
	}

	us, err := ctrl.service.AssignUserSkill(c.Request.Context(), userID, &req)
	if err != nil {
		log.Printf("Gagal menetapkan skill pada pengguna: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Keahlian pengguna berhasil disimpan", us)
}

// Menangani permintaan HTTP untuk mengambil daftar keahlian yang dimiliki pengguna.
func (ctrl *Controller) GetUserSkills(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.BadRequest(c, "Format user_id tidak valid")
		return
	}

	skills, err := ctrl.service.GetUserSkills(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Gagal mengambil skill pengguna: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil skill pengguna")
		return
	}

	response.Success(c, http.StatusOK, "Keahlian pengguna berhasil diambil", skills)
}

// Menangani permintaan HTTP untuk menghapus keahlian dari profil pengguna.
func (ctrl *Controller) DeleteUserSkill(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.BadRequest(c, "Format user_id tidak valid")
		return
	}

	skillID, err := uuid.Parse(c.Param("skill_id"))
	if err != nil {
		response.BadRequest(c, "Format skill_id tidak valid")
		return
	}

	if err := ctrl.service.DeleteUserSkill(c.Request.Context(), userID, skillID); err != nil {
		log.Printf("Gagal menghapus skill pengguna: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Keahlian pengguna berhasil dihapus", nil)
}
