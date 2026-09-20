package projects

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

// Menginisialisasi instance baru controller projects and tasks management.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// ============================================================================
// 1. PROJECTS (FR-016)
// ============================================================================

// Menangani permintaan HTTP pembuatan proyek baru di marketplace.
func (ctrl *Controller) CreateProject(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	ownerID, _ := val.(uuid.UUID)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan proyek tidak valid", err.Error())
		return
	}

	p, err := ctrl.service.CreateProject(c.Request.Context(), ownerID, &req)
	if err != nil {
		log.Printf("Gagal membuat proyek: %v", err)
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Proyek berhasil diterbitkan", p)
}

// Menangani permintaan HTTP untuk menampilkan bursa proyek dengan penyaringan visibilitas berbasis izin dinamis.
func (ctrl *Controller) ListProjects(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	status := c.Query("status")
	search := c.Query("q")
	vis := c.Query("visibility")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	projects, total, err := ctrl.service.ListProjects(c.Request.Context(), userID, status, search, vis, page, limit)
	if err != nil {
		log.Printf("Gagal mengambil bursa proyek: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat mengambil bursa proyek")
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := gin.H{
		"page":       page,
		"limit":      limit,
		"totalItems": total,
		"totalPages": totalPages,
	}

	response.Success(c, http.StatusOK, "Bursa proyek berhasil diambil", projects, meta)
}

// Menangani permintaan HTTP untuk mengambil detail satu proyek setelah verifikasi visibilitas.
func (ctrl *Controller) GetProjectByID(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	p, err := ctrl.service.GetProjectByID(c.Request.Context(), id, userID)
	if err != nil {
		response.SafeNotFound(c, "Data tidak ditemukan", err)
		return
	}

	response.Success(c, http.StatusOK, "Detail proyek berhasil diambil", p)
}

// Menangani permintaan HTTP untuk memperbarui konfigurasi proyek.
func (ctrl *Controller) UpdateProject(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembaruan proyek tidak valid", err.Error())
		return
	}

	p, err := ctrl.service.UpdateProject(c.Request.Context(), id, &req, userID)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Proyek berhasil diperbarui", p)
}

// ============================================================================
// 2. APPLICATIONS & QUOTA (FR-017, T-046)
// ============================================================================

// Menangani permintaan HTTP pengajuan lamaran pada proyek dengan kalkulasi skor kecocokan keahlian informatif.
func (ctrl *Controller) ApplyProject(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	var req ApplyProjectRequest
	_ = c.ShouldBindJSON(&req)

	app, err := ctrl.service.ApplyProject(c.Request.Context(), projectID, userID, &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Lamaran proyek berhasil diajukan", app)
}

// Menangani permintaan HTTP untuk menampilkan seluruh daftar pelamar pada proyek.
func (ctrl *Controller) ListApplications(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	apps, err := ctrl.service.repo.ListApplicationsByProject(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar pelamar proyek")
		return
	}

	response.Success(c, http.StatusOK, "Daftar pelamar proyek berhasil diambil", apps)
}

// Menangani persetujuan atau penolakan lamaran proyek dengan penguncian kuota tim secara atomik.
func (ctrl *Controller) ReviewApplication(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID lamaran tidak valid")
		return
	}

	var req ReviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload peninjauan lamaran tidak valid", err.Error())
		return
	}

	app, err := ctrl.service.ReviewApplication(c.Request.Context(), appID, req.Action)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Lamaran proyek berhasil ditinjau", app)
}

// ============================================================================
// 3. TEAMS & THREE-LAYER CONTRIBUTION (FR-019, FR-024, T-047, T-052)
// ============================================================================

// Menangani permintaan HTTP untuk menampilkan seluruh anggota tim pada proyek.
func (ctrl *Controller) ListTeamMembers(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	members, err := ctrl.service.repo.ListTeamMembers(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar anggota tim proyek")
		return
	}

	response.Success(c, http.StatusOK, "Daftar anggota tim proyek berhasil diambil", members)
}

// Menangani penetapan kesepakatan awal persentase Planned Contribution anggota tim (sum=100%).
func (ctrl *Controller) FinalizePlannedContribution(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	var req FinalizePlannedContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload penetapan planned contribution tidak valid", err.Error())
		return
	}

	if err := ctrl.service.FinalizePlannedContribution(c.Request.Context(), projectID, &req); err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Kesepakatan Planned Contribution berhasil disahkan (total 100.00%)", nil)
}

// Menangani komputasi persentase Actual Contribution (Layer 2) berbasis bobot tugas selesai.
func (ctrl *Controller) CalculateActualContribution(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	members, err := ctrl.service.CalculateActualContribution(c.Request.Context(), projectID)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Kalkulasi Actual Contribution berhasil dihitung", members)
}

// Menangani pengesahan dan penguncian Final Contribution (Layer 3) oleh Supervisor (sum=100%).
func (ctrl *Controller) FinalizeContribution(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	supervisorID, _ := val.(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	var req FinalizeContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pengesahan final contribution tidak valid", err.Error())
		return
	}

	if err := ctrl.service.FinalizeContribution(c.Request.Context(), projectID, &req, supervisorID); err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Final Contribution % berhasil disahkan dan dikunci (sum=100.00%)", nil)
}

// ============================================================================
// 4. MILESTONES (FR-020, T-048)
// ============================================================================

// Menangani pembuatan milestone target fase proyek baru.
func (ctrl *Controller) CreateMilestone(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	var req CreateMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan milestone tidak valid", err.Error())
		return
	}

	m, err := ctrl.service.CreateMilestone(c.Request.Context(), projectID, &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Milestone proyek berhasil dibuat", m)
}

// Menangani permintaan HTTP untuk menampilkan seluruh milestone proyek.
func (ctrl *Controller) ListMilestones(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	list, err := ctrl.service.repo.ListMilestones(c.Request.Context(), projectID)
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar milestone proyek")
		return
	}

	response.Success(c, http.StatusOK, "Daftar milestone proyek berhasil diambil", list)
}

// ============================================================================
// 5. TASKS KANBAN (FR-021, T-049)
// ============================================================================

// Menangani pembuatan tugas kanban baru pada proyek.
func (ctrl *Controller) CreateTask(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pembuatan tugas tidak valid", err.Error())
		return
	}

	task, err := ctrl.service.CreateTask(c.Request.Context(), projectID, &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Tugas kanban berhasil dibuat", task)
}

// Menangani permintaan HTTP untuk menampilkan daftar tugas kanban dengan filter.
func (ctrl *Controller) ListTasks(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID proyek tidak valid")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	var milestoneID *uuid.UUID
	if mStr := c.Query("milestone_id"); mStr != "" {
		if parsed, err := uuid.Parse(mStr); err == nil {
			milestoneID = &parsed
		}
	}

	var assigneeID *uuid.UUID
	if aStr := c.Query("assignee_id"); aStr != "" {
		if parsed, err := uuid.Parse(aStr); err == nil {
			assigneeID = &parsed
		}
	}

	offset := (page - 1) * limit
	tasks, total, err := ctrl.service.repo.ListTasks(c.Request.Context(), projectID, milestoneID, assigneeID, status, limit, offset)
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar tugas kanban")
		return
	}

	meta := gin.H{
		"page":       page,
		"limit":      limit,
		"totalItems": total,
		"totalPages": (total + limit - 1) / limit,
	}

	response.Success(c, http.StatusOK, "Daftar tugas kanban berhasil diambil", tasks, meta)
}

// Menangani permintaan HTTP untuk mengambil detail satu tugas kanban.
func (ctrl *Controller) GetTaskByID(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID tugas tidak valid")
		return
	}

	task, err := ctrl.service.repo.GetTaskByID(c.Request.Context(), taskID)
	if err != nil {
		response.NotFound(c, "Tugas tidak ditemukan")
		return
	}

	response.Success(c, http.StatusOK, "Detail tugas berhasil diambil", task)
}

// Menangani transisi status tugas pada papan kanban sesuai mesin status.
func (ctrl *Controller) ChangeTaskStatus(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID tugas tidak valid")
		return
	}

	var req ChangeTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload perubahan status tugas tidak valid", err.Error())
		return
	}

	task, err := ctrl.service.ChangeTaskStatus(c.Request.Context(), taskID, req.Status)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Status tugas berhasil diperbarui", task)
}

// ============================================================================
// 6. WORK REPORTS & EVIDENCE (FR-022, FR-023, T-050, T-051)
// ============================================================================

// Menangani penyerahan laporan kemajuan tugas beserta lampiran artefak bukti deliverable.
func (ctrl *Controller) SubmitWorkReport(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID tugas tidak valid")
		return
	}

	var req SubmitWorkReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload laporan pekerjaan tidak valid (deskripsi minimal 20 karakter)", err.Error())
		return
	}

	rep, err := ctrl.service.SubmitWorkReport(c.Request.Context(), taskID, userID, &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Laporan pekerjaan tugas berhasil diserahkan dan status tugas beralih ke IN_REVIEW", rep)
}

// Menangani penelaahan laporan pekerjaan tugas oleh reviewer (APPROVE atau REVISION_REQUIRED).
func (ctrl *Controller) ReviewWorkReport(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID laporan tidak valid")
		return
	}

	var req ReviewWorkReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload peninjauan laporan tidak valid", err.Error())
		return
	}

	rep, err := ctrl.service.ReviewWorkReport(c.Request.Context(), reportID, req.Action)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan proyek tidak dapat diproses", err)
		return
	}

	msg := "Laporan pekerjaan tugas berhasil disetujui (status tugas COMPLETED)"
	if req.Action == "REVISION_REQUIRED" {
		msg = "Permintaan revisi berhasil diajukan (status tugas kembali ke IN_PROGRESS)"
	}

	response.Success(c, http.StatusOK, msg, rep)
}
