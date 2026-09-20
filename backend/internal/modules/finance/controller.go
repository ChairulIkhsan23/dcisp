package finance

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

// Menginisialisasi instance baru controller finance and compensation management.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// ============================================================================
// 1. WALLETS (FR-035)
// ============================================================================

// Menangani permintaan HTTP untuk mengambil dompet personal pengguna yang sedang login.
func (ctrl *Controller) GetMyWallet(c *gin.Context) {
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

	w, err := ctrl.service.GetWalletByUserID(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Gagal mengambil dompet: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil dompet")
		return
	}

	response.Success(c, http.StatusOK, "Dompet personal berhasil diambil", w)
}

// Menangani permintaan HTTP untuk mengambil dompet pengguna tertentu oleh finance.
func (ctrl *Controller) GetUserWallet(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.BadRequest(c, "Format ID pengguna tidak valid")
		return
	}

	w, err := ctrl.service.GetWalletByUserID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "Dompet pengguna tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Dompet pengguna berhasil diambil", w)
}

// Menangani permintaan HTTP untuk menampilkan riwayat mutasi dompet personal.
func (ctrl *Controller) ListMyTransactions(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	txs, total, err := ctrl.service.ListWalletTransactions(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.InternalError(c, "Gagal mengambil riwayat mutasi dompet")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{"page": page, "limit": limit, "totalItems": total, "totalPages": totalPages}

	response.Success(c, http.StatusOK, "Riwayat mutasi dompet berhasil diambil", txs, meta)
}

// ============================================================================
// 2. LEDGER (FR-037)
// ============================================================================

// Menangani permintaan HTTP untuk menampilkan buku besar finansial bagi peran finance.
func (ctrl *Controller) ListLedger(c *gin.Context) {
	accountCode := c.Query("account_code")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	entries, total, err := ctrl.service.repo.ListLedgerEntries(c.Request.Context(), accountCode, limit, offset)
	if err != nil {
		response.InternalError(c, "Gagal mengambil buku besar finansial")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{"page": page, "limit": limit, "totalItems": total, "totalPages": totalPages}

	response.Success(c, http.StatusOK, "Buku besar finansial berhasil diambil", entries, meta)
}

// ============================================================================
// 3. TAX RULES (FR-034)
// ============================================================================

// Menangani permintaan HTTP untuk menampilkan seluruh aturan deduksi aktif dan non-aktif.
func (ctrl *Controller) ListTaxRules(c *gin.Context) {
	rules, err := ctrl.service.repo.ListAllRules(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil aturan deduksi")
		return
	}

	response.Success(c, http.StatusOK, "Daftar aturan deduksi berhasil diambil", rules)
}

// Menangani pembuatan aturan deduksi baru oleh finance.
func (ctrl *Controller) CreateTaxRule(c *gin.Context) {
	var req CreateTaxRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload aturan deduksi tidak valid", err.Error())
		return
	}

	rule, err := ctrl.service.CreateTaxRule(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Gagal membuat aturan deduksi: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Aturan deduksi berhasil dibuat", rule)
}

// ============================================================================
// 4. BOUNTY (FR-032, FR-039)
// ============================================================================

// Menangani distribusi bounty pool proyek ke seluruh anggota tim secara atomik.
func (ctrl *Controller) DistributeBounty(c *gin.Context) {
	var req DistributeBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload distribusi bounty tidak valid", err.Error())
		return
	}

	resp, err := ctrl.service.DistributeBounty(c.Request.Context(), req.ProjectID)
	if err != nil {
		log.Printf("Gagal mendistribusikan bounty: %v", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Distribusi bounty proyek berhasil dieksekusi", resp)
}

// ============================================================================
// 5. BATCH FUNDS (FR-036)
// ============================================================================

// Menangani permintaan HTTP untuk menampilkan seluruh kas bersama per angkatan.
func (ctrl *Controller) ListBatchFunds(c *gin.Context) {
	funds, err := ctrl.service.ListBatchFunds(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar kas batch")
		return
	}

	response.Success(c, http.StatusOK, "Daftar kas bersama angkatan berhasil diambil", funds)
}

// Menangani permintaan HTTP untuk mengambil kas bersama satu angkatan.
func (ctrl *Controller) GetBatchFund(c *gin.Context) {
	batchID, err := uuid.Parse(c.Param("batch_id"))
	if err != nil {
		response.BadRequest(c, "Format ID batch tidak valid")
		return
	}

	fund, err := ctrl.service.GetBatchFund(c.Request.Context(), batchID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Kas bersama angkatan berhasil diambil", fund)
}

// ============================================================================
// 6. PAYOUTS (FR-038)
// ============================================================================

// Menangani pengajuan permohonan pencairan dana oleh pemilik dompet.
func (ctrl *Controller) RequestPayout(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	var req RequestPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pengajuan payout tidak valid", err.Error())
		return
	}
	// Masking nomor rekening pada log: jangan mencatat nomor rekening lengkap
	maskedReq := req
	if len(maskedReq.AccountNumber) > 4 {
		maskedReq.AccountNumber = "****" + maskedReq.AccountNumber[len(maskedReq.AccountNumber)-4:]
	}
	_ = maskedReq

	payout, err := ctrl.service.RequestPayout(c.Request.Context(), userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Samarkan nomor rekening pada respons
	if len(payout.AccountNumber) > 4 {
		payout.AccountNumber = "****" + payout.AccountNumber[len(payout.AccountNumber)-4:]
	}

	response.Success(c, http.StatusCreated, "Permohonan payout berhasil diajukan dan saldo ditahan", payout)
}

// Menangani peninjauan payout oleh finance: APPROVE, REJECT, PROCESSING, SETTLE, FAIL.
func (ctrl *Controller) ReviewPayout(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	reviewerID, _ := val.(uuid.UUID)

	payoutID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID payout tidak valid")
		return
	}

	var req ReviewPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload peninjauan payout tidak valid", err.Error())
		return
	}

	payout, err := ctrl.service.ReviewPayout(c.Request.Context(), payoutID, &req, reviewerID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Permohonan payout berhasil ditinjau", payout)
}

// Menangani permintaan HTTP untuk menampilkan riwayat payout milik pengguna yang login.
func (ctrl *Controller) ListMyPayouts(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	w, err := ctrl.service.GetWalletByUserID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Gagal mengambil dompet pengguna")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	payouts, total, err := ctrl.service.repo.ListPayoutsByWallet(c.Request.Context(), w.ID, limit, offset)
	if err != nil {
		response.InternalError(c, "Gagal mengambil riwayat payout")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{"page": page, "limit": limit, "totalItems": total, "totalPages": totalPages}

	response.Success(c, http.StatusOK, "Riwayat payout berhasil diambil", payouts, meta)
}

// Menangani permintaan HTTP untuk menampilkan antrean seluruh payout bagi finance.
func (ctrl *Controller) ListAllPayouts(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	payouts, total, err := ctrl.service.repo.ListAllPayouts(c.Request.Context(), status, limit, offset)
	if err != nil {
		response.InternalError(c, "Gagal mengambil antrean payout")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{"page": page, "limit": limit, "totalItems": total, "totalPages": totalPages}

	response.Success(c, http.StatusOK, "Antrean payout berhasil diambil", payouts, meta)
}

// ============================================================================
// 7. REWARDS & CLAIMS (FR-031, FR-033)
// ============================================================================

// Menangani permintaan HTTP untuk menampilkan katalog reward.
func (ctrl *Controller) ListRewards(c *gin.Context) {
	rewards, err := ctrl.service.repo.ListRewards(c.Request.Context(), nil, nil)
	if err != nil {
		response.InternalError(c, "Gagal mengambil katalog reward")
		return
	}

	response.Success(c, http.StatusOK, "Katalog reward berhasil diambil", rewards)
}

// Menangani pembuatan master reward baru oleh finance.
func (ctrl *Controller) CreateReward(c *gin.Context) {
	var req CreateRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan reward tidak valid", err.Error())
		return
	}

	reward := &Reward{
		ID:            uuid.New(),
		RankID:        req.RankID,
		AchievementID: req.AchievementID,
		ComponentType: req.ComponentType,
		Title:         req.Title,
		MonetaryValue: req.MonetaryValue,
		Description:   req.Description,
	}

	if err := ctrl.service.repo.CreateReward(c.Request.Context(), reward); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Master reward berhasil dibuat", reward)
}

// Menangani penerbitan paket reward atas kenaikan rank pengguna.
func (ctrl *Controller) IssueRankRewards(c *gin.Context) {
	var req IssueRankRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload penerbitan reward tidak valid", err.Error())
		return
	}

	claims, err := ctrl.service.IssueRankRewards(c.Request.Context(), req.UserID, req.RankID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Paket reward kenaikan rank berhasil diterbitkan", claims)
}

// Menangani pengajuan klaim tiket reward oleh pemiliknya.
func (ctrl *Controller) ClaimReward(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	claimID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID klaim tidak valid")
		return
	}

	var req ClaimRewardRequest
	_ = c.ShouldBindJSON(&req)

	claim, err := ctrl.service.ClaimReward(c.Request.Context(), claimID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Klaim reward berhasil diajukan", claim)
}

// Menangani pemrosesan tiket klaim reward oleh admin.
func (ctrl *Controller) ProcessClaim(c *gin.Context) {
	claimID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID klaim tidak valid")
		return
	}

	var req ProcessClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pemrosesan klaim tidak valid", err.Error())
		return
	}

	claim, err := ctrl.service.ProcessClaim(c.Request.Context(), claimID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Tiket klaim reward berhasil diproses", claim)
}

// Menangani permintaan HTTP untuk menampilkan klaim reward milik pengguna yang login.
func (ctrl *Controller) ListMyClaims(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	claims, err := ctrl.service.repo.ListClaimsByUser(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Gagal mengambil klaim reward")
		return
	}

	response.Success(c, http.StatusOK, "Daftar klaim reward berhasil diambil", claims)
}
