package performance

import (
	"log"
	"net/http"
	"strconv"

	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

// Menginisialisasi instance baru controller performance and gamification management.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// ============================================================================
// 1. XP BALANCES & TRANSACTIONS (FR-025, FR-026)
// ============================================================================

// Menangani permintaan HTTP untuk mengambil saldo XP 3 jalur dan status rank pengguna yang sedang login.
func (ctrl *Controller) GetMyStats(c *gin.Context) {
	val, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}
	userID, ok := val.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Konteks identitas pengguna tidak valid")
		return
	}

	stats, err := ctrl.service.GetXPBalance(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Gagal mengambil statistik XP: %v", err)
		response.InternalError(c, "Gagal mengambil data statistik XP")
		return
	}

	response.Success(c, http.StatusOK, "Statistik XP berhasil diambil", stats)
}

// Menangani permintaan HTTP untuk mengambil saldo XP dan rank pengguna tertentu oleh supervisor atau admin.
func (ctrl *Controller) GetUserStats(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.BadRequest(c, "Format ID pengguna tidak valid")
		return
	}

	stats, err := ctrl.service.GetXPBalance(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "Statistik XP pengguna tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Statistik XP pengguna berhasil diambil", stats)
}

// Menangani pencatatan mutasi poin pengalaman secara append-only.
func (ctrl *Controller) MutateXP(c *gin.Context) {
	var req RecordXPMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload mutasi XP tidak valid", err.Error())
		return
	}

	txRecord, err := ctrl.service.RecordXPMutation(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal memproses mutasi XP: %v", err)
		response.SafeBadRequest(c, "Permintaan data performa tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Mutasi XP berhasil dicatat ke buku besar", txRecord)
}

// Menangani permintaan HTTP untuk menampilkan riwayat transaksi mutasi XP pengguna.
func (ctrl *Controller) GetXPHistory(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	scheme := c.Query("scheme")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	txs, total, err := ctrl.service.repo.ListUserTransactions(c.Request.Context(), userID, scheme, limit, offset)
	if err != nil {
		response.InternalError(c, "Gagal mengambil riwayat transaksi XP")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{
		"page":       page,
		"limit":      limit,
		"totalItems": total,
		"totalPages": totalPages,
	}

	response.Success(c, http.StatusOK, "Riwayat transaksi XP berhasil diambil", txs, meta)
}

// Menangani permintaan HTTP untuk menampilkan seluruh daftar tingkatan rank gamifikasi.
func (ctrl *Controller) ListRanks(c *gin.Context) {
	ranks, err := ctrl.service.repo.ListRanks(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar rank")
		return
	}
	response.Success(c, http.StatusOK, "Daftar rank berhasil diambil", ranks)
}

// ============================================================================
// 2. PERFORMANCE EVALUATIONS & TOP PERFORMER (FR-027, FR-028)
// ============================================================================

// Menangani pembuatan dokumen evaluasi kinerja formal supervisor berbasis rubrik 0–100.
func (ctrl *Controller) CreateEvaluation(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	evaluatorID, _ := val.(uuid.UUID)

	var req CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload evaluasi kinerja tidak valid", err.Error())
		return
	}

	eval, err := ctrl.service.CreateEvaluation(c.Request.Context(), evaluatorID, &req)
	if err != nil {
		log.Printf("Gagal membuat evaluasi performa: %v", err)
		response.SafeBadRequest(c, "Permintaan data performa tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Evaluasi kinerja formal berhasil disahkan", eval)
}

// Menangani permintaan HTTP untuk mengambil detail satu dokumen evaluasi kinerja berdasarkan ID.
func (ctrl *Controller) GetEvaluationByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID evaluasi tidak valid")
		return
	}

	eval, err := ctrl.service.GetEvaluationByID(c.Request.Context(), id)
	if err != nil {
		response.SafeNotFound(c, "Data tidak ditemukan", err)
		return
	}

	response.Success(c, http.StatusOK, "Detail evaluasi kinerja berhasil diambil", eval)
}

// Menangani penetapan tepat SATU Top Performer per kohort batch per periode evaluasi.
func (ctrl *Controller) DetermineTopPerformer(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("batch_id"))
	if err != nil {
		response.BadRequest(c, "Format ID batch tidak valid")
		return
	}

	periodType := c.DefaultQuery("period_type", PeriodEndOfBatch)

	res, err := ctrl.service.DetermineTopPerformer(c.Request.Context(), batchID, periodType)
	if err != nil {
		log.Printf("Gagal menetapkan top performer: %v", err)
		response.SafeBadRequest(c, "Permintaan data performa tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Top Performer berhasil ditetapkan", res)
}

// ============================================================================
// 3. ACHIEVEMENTS & BADGES (FR-029)
// ============================================================================

// Menangani pembuatan master lencana prestasi baru dalam katalog.
func (ctrl *Controller) CreateAchievement(c *gin.Context) {
	var req CreateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan achievement tidak valid", err.Error())
		return
	}

	ach, err := ctrl.service.CreateAchievement(c.Request.Context(), &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan data performa tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Master achievement berhasil dibuat", ach)
}

// Menangani permintaan HTTP untuk menampilkan seluruh katalog lencana prestasi.
func (ctrl *Controller) ListAchievements(c *gin.Context) {
	list, err := ctrl.service.repo.ListAchievements(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil katalog achievement")
		return
	}
	response.Success(c, http.StatusOK, "Katalog achievement berhasil diambil", list)
}

// Menangani permintaan HTTP untuk menampilkan daftar lencana prestasi yang dimiliki pengguna saat ini.
func (ctrl *Controller) GetMyAchievements(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	userAchs, err := ctrl.service.GetUserAchievements(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Gagal mengambil achievement pengguna")
		return
	}

	response.Success(c, http.StatusOK, "Lencana prestasi berhasil diambil", userAchs)
}

// Menangani pembukaan lencana prestasi pengguna secara idempoten.
func (ctrl *Controller) UnlockAchievement(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Kode achievement wajib diisi")
		return
	}

	isNew, err := ctrl.service.UnlockAchievement(c.Request.Context(), userID, code)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan data performa tidak dapat diproses", err)
		return
	}

	msg := "Achievement berhasil dibuka"
	if !isNew {
		msg = "Achievement sudah pernah dibuka sebelumnya (idempoten)"
	}

	response.Success(c, http.StatusOK, msg, gin.H{
		"unlocked": true,
		"is_new":   isNew,
	})
}

// ============================================================================
// 4. SKILL GROWTH MATRIX (FR-030, T-061)
// ============================================================================

// Menangani permintaan HTTP untuk mengambil visualisasi matriks pertumbuhan keahlian pengguna.
func (ctrl *Controller) GetMySkillGrowth(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	matrix, err := ctrl.service.GetSkillGrowthMatrix(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Gagal mengambil matriks pertumbuhan keahlian")
		return
	}

	response.Success(c, http.StatusOK, "Matriks pertumbuhan keahlian berhasil diambil", matrix)
}
