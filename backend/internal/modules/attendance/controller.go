package attendance

import (
	"errors"
	"log"
	"net/http"

	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

// Menginisialisasi instance baru controller workforce and attendance management.
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// ============================================================================
// 1. TERMINAL TAP & OFFLINE SYNC (FR-008, T-034, T-042)
// ============================================================================

// Menangani permintaan presensi dari terminal pemindai gerbang (NFC/QR) dengan debouncing dan klasifikasi jadwal.
func (ctrl *Controller) HandleTerminalTap(c *gin.Context) {
	var req TerminalTapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload presensi terminal tidak valid", err.Error())
		return
	}

	// deviceID berasal dari middleware DeviceAuth (terminal terautentikasi).
	// Fallback: header X-Device-ID juga diterima sebagai UUID langsung untuk kompatibilitas admin.
	var deviceID *uuid.UUID
	if devVal, exists := c.Get(middleware.ContextDeviceIDKey); exists {
		if devUUID, ok := devVal.(uuid.UUID); ok {
			deviceID = &devUUID
		}
	}
	if deviceID == nil {
		if devIDStr := c.GetHeader("X-Device-ID"); devIDStr != "" {
			if parsed, err := uuid.Parse(devIDStr); err == nil {
				deviceID = &parsed
			}
		}
	}

	result, err := ctrl.service.ProcessTap(c.Request.Context(), deviceID, &req)
	if err != nil {
		if errors.Is(err, ErrCardUnregistered) {
			c.JSON(http.StatusNotFound, gin.H{
				"success":     false,
				"statusCode":  http.StatusNotFound,
				"status":      "ERROR",
				"audio_event": AudioCardUnregistered,
				"message":     err.Error(),
			})
			return
		}
		if errors.Is(err, ErrCardInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":     false,
				"statusCode":  http.StatusBadRequest,
				"status":      "ERROR",
				"audio_event": AudioCardInvalid,
				"message":     err.Error(),
			})
			return
		}
		log.Printf("Kesalahan internal saat memproses presensi: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":     false,
			"statusCode":  http.StatusInternalServerError,
			"status":      "FAILED",
			"audio_event": AudioSaveFailed,
			"message":     "Gagal mencatat presensi pada sistem",
		})
		return
	}

	// Respon sukses atau debounced
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"statusCode":   http.StatusOK,
		"status":       result.Status,
		"event_type":   result.EventType,
		"is_late":      result.IsLate,
		"late_tier":    result.LateTier,
		"late_minutes": result.LateMinutes,
		"audio_event":  result.AudioEvent,
		"message":      result.Message,
		"data":         result.Data,
	})
}

// Menangani sinkronisasi tumpukan rekaman presensi offline dari penyimpanan lokal terminal LittleFS.
func (ctrl *Controller) SyncOffline(c *gin.Context) {
	var req OfflineSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload sinkronisasi offline tidak valid", err.Error())
		return
	}

	res, err := ctrl.service.SyncOffline(c.Request.Context(), req.Records)
	if err != nil {
		log.Printf("Gagal memproses sinkronisasi offline: %v", err)
		response.InternalError(c, "Terjadi kesalahan saat sinkronisasi presensi offline")
		return
	}

	response.Success(c, http.StatusOK, "Sinkronisasi presensi offline berhasil diproses", res)
}

// Menangani pembuatan kode QR presensi dinamis ber-TTL 30 detik untuk identitas peserta.
func (ctrl *Controller) GenerateQR(c *gin.Context) {
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

	res, err := ctrl.service.GenerateDynamicQR(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Gagal membuat QR dinamis: %v", err)
		response.InternalError(c, "Gagal membuat QR code presensi")
		return
	}

	response.Success(c, http.StatusOK, "QR presensi dinamis berhasil dibuat", res)
}

// Menangani permintaan manifes katalog 15 audio master untuk sinkronisasi firmware terminal ESP32.
func (ctrl *Controller) GetAudioCatalog(c *gin.Context) {
	res, err := ctrl.service.GetAudioCatalog(c.Request.Context())
	if err != nil {
		log.Printf("Gagal mengambil katalog audio: %v", err)
		response.InternalError(c, "Gagal mengambil katalog audio terminal")
		return
	}

	response.Success(c, http.StatusOK, "Katalog audio terminal berhasil diambil", res)
}

// ============================================================================
// 2. WORK SESSIONS & BREAKS (FR-009, FR-010, T-036, T-037)
// ============================================================================

// Menangani permintaan memulai sesi kerja harian dari portal My Day.
func (ctrl *Controller) StartWorkSession(c *gin.Context) {
	val, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}
	userID, ok := val.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Konteks pengguna tidak valid")
		return
	}

	var req StartWorkSessionRequest
	_ = c.ShouldBindJSON(&req)

	session, err := ctrl.service.StartWorkSession(c.Request.Context(), userID, req.TaskID)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Sesi kerja harian berhasil dimulai", session)
}

// Menangani aksi jeda atau melanjutkan kerja pada sesi kerja aktif.
func (ctrl *Controller) ProcessBreak(c *gin.Context) {
	val, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}
	userID, ok := val.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Konteks pengguna tidak valid")
		return
	}

	var req BreakActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Aksi istirahat tidak valid (START atau RESUME)", err.Error())
		return
	}

	brk, err := ctrl.service.ProcessBreakAction(c.Request.Context(), userID, req.Action)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	msg := "Periode istirahat berhasil dimulai"
	if req.Action == "RESUME" {
		msg = "Sesi kerja berhasil dilanjutkan kembali"
	}

	response.Success(c, http.StatusOK, msg, brk)
}

// Menangani penutupan sesi kerja harian dari portal My Day.
func (ctrl *Controller) EndWorkSession(c *gin.Context) {
	val, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}
	userID, ok := val.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Konteks pengguna tidak valid")
		return
	}

	session, err := ctrl.service.EndWorkSession(c.Request.Context(), userID)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Sesi kerja harian berhasil diselesaikan", session)
}

// ============================================================================
// 3. OVERTIME, LEAVE & CORRECTIONS (T-038, T-039, T-040)
// ============================================================================

// Menangani pengajuan permohonan lembur oleh peserta magang.
func (ctrl *Controller) RequestOvertime(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	var req CreateOvertimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload permohonan lembur tidak valid", err.Error())
		return
	}

	ot, err := ctrl.service.RequestOvertime(c.Request.Context(), userID, &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Permohonan lembur berhasil diajukan", ot)
}

// Menangani persetujuan atau penolakan permohonan lembur oleh supervisor.
func (ctrl *Controller) ReviewOvertime(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	reviewerID, _ := val.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID lembur tidak valid")
		return
	}

	var req ReviewOvertimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload peninjauan lembur tidak valid", err.Error())
		return
	}

	ot, err := ctrl.service.ReviewOvertime(c.Request.Context(), id, &req, reviewerID)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Permohonan lembur berhasil ditinjau", ot)
}

// Menangani pengajuan permohonan izin cuti sakit atau akademik.
func (ctrl *Controller) RequestLeave(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	var req CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload permohonan cuti tidak valid", err.Error())
		return
	}

	lr, err := ctrl.service.RequestLeave(c.Request.Context(), userID, &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Permohonan cuti berhasil diajukan", lr)
}

// Menangani persetujuan atau penolakan izin cuti oleh supervisor atau HR.
func (ctrl *Controller) ReviewLeave(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	reviewerID, _ := val.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID cuti tidak valid")
		return
	}

	var req ReviewLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload peninjauan cuti tidak valid", err.Error())
		return
	}

	lr, err := ctrl.service.ReviewLeave(c.Request.Context(), id, &req, reviewerID)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Permohonan cuti berhasil ditinjau", lr)
}

// Menangani pengajuan koreksi absensi manual dengan lampiran bukti.
func (ctrl *Controller) RequestCorrection(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	userID, _ := val.(uuid.UUID)

	var req CreateCorrectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload koreksi absensi tidak valid (alasan minimal 20 karakter)", err.Error())
		return
	}

	corr, err := ctrl.service.RequestCorrection(c.Request.Context(), userID, &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Permohonan koreksi absensi berhasil diajukan", corr)
}

// Menangani persetujuan koreksi absensi yang menghasilkan rekaman log baru secara append-only.
func (ctrl *Controller) ReviewCorrection(c *gin.Context) {
	val, _ := c.Get(middleware.ContextUserIDKey)
	reviewerID, _ := val.(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID koreksi tidak valid")
		return
	}

	var req ReviewCorrectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload peninjauan koreksi tidak valid", err.Error())
		return
	}

	corr, err := ctrl.service.ReviewCorrection(c.Request.Context(), id, &req, reviewerID)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusOK, "Permohonan koreksi absensi berhasil ditinjau", corr)
}

// ============================================================================
// 4. DEVICE REGISTRY (FR-013, T-041)
// ============================================================================

// Menangani pendaftaran terminal pemindai gerbang baru oleh administrator.
func (ctrl *Controller) RegisterDevice(c *gin.Context) {
	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Payload pendaftaran terminal tidak valid", err.Error())
		return
	}

	dev, err := ctrl.service.RegisterDevice(c.Request.Context(), &req)
	if err != nil {
		response.SafeBadRequest(c, "Permintaan presensi tidak dapat diproses", err)
		return
	}

	response.Success(c, http.StatusCreated, "Terminal pemindai berhasil didaftarkan", dev)
}

// Menangani permintaan untuk menampilkan daftar seluruh perangkat terminal terdaftar.
func (ctrl *Controller) ListDevices(c *gin.Context) {
	list, err := ctrl.service.repo.ListDevices(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil daftar terminal")
		return
	}
	response.Success(c, http.StatusOK, "Daftar terminal berhasil diambil", list)
}

// Menangani laporan telemetri detak jantung (heartbeat) dari terminal pemindai.
func (ctrl *Controller) DeviceHeartbeat(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Format ID terminal tidak valid")
		return
	}

	var req DeviceHeartbeatRequest
	_ = c.ShouldBindJSON(&req)

	err = ctrl.service.repo.UpdateDeviceHeartbeat(c.Request.Context(), id, req.FirmwareVersion, req.AudioCatalogVersion, req.CurrentMode)
	if err != nil {
		response.InternalError(c, "Gagal memperbarui heartbeat terminal")
		return
	}

	response.Success(c, http.StatusOK, "Heartbeat terminal berhasil diterima", nil)
}
