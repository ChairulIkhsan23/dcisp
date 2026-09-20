package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/documents"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji pembuatan presigned URL unggah Cloudflare R2, validasi partisi folder, dan batas ukuran berkas.
func TestPresignedUploadAndValidation(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	service := documents.NewStorageService(db, cfg)
	ctx := context.Background()

	// 1. Uji penolakan folder terlarang (path traversal prevention)
	reqInvalidFolder := &documents.PresignedUploadRequest{
		Folder:    "../malicious_folder",
		Filename:  "exploit.png",
		MimeType:  "image/png",
		SizeBytes: 1024,
	}
	_, err = service.GeneratePresignedUpload(ctx, reqInvalidFolder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partisi folder tidak diizinkan")

	// 2. Uji penolakan ukuran berkas melebihi batas 25MB
	reqOversize := &documents.PresignedUploadRequest{
		Folder:    "task-evidence",
		Filename:  "huge_file.zip",
		MimeType:  "application/zip",
		SizeBytes: 30 * 1024 * 1024, // 30MB
	}
	_, err = service.GeneratePresignedUpload(ctx, reqOversize)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "melebihi batas maksimal 25 MB")

	// 3. Uji pembuatan presigned upload URL valid
	reqValid := &documents.PresignedUploadRequest{
		Folder:    "task-evidence",
		Filename:  "screenshot_submission.png",
		MimeType:  "image/png",
		SizeBytes: 2048000,
	}
	res, err := service.GeneratePresignedUpload(ctx, reqValid)
	require.NoError(t, err)
	assert.NotEmpty(t, res.UploadURL)
	assert.NotEmpty(t, res.PathKey)
	assert.Equal(t, 900, res.ExpiresInSeconds)

	// Pastikan URL memuat query parameter AWS SigV4 yang sah
	parsedURL, err := url.Parse(res.UploadURL)
	require.NoError(t, err)
	q := parsedURL.Query()
	assert.Equal(t, "AWS4-HMAC-SHA256", q.Get("X-Amz-Algorithm"))
	assert.NotEmpty(t, q.Get("X-Amz-Signature"))
	assert.Equal(t, "900", q.Get("X-Amz-Expires"))

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM file_metadata WHERE id = $1", res.FileID)
	})

	// 4. Verifikasi rekod metadata tersimpan di database
	meta, err := service.GetFileMetadata(ctx, res.FileID)
	require.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Equal(t, res.PathKey, meta.PathKey)
	assert.Equal(t, "image/png", meta.MimeType)

	// 5. Uji pembuatan presigned download URL
	downloadURL, err := service.GeneratePresignedDownload(ctx, res.FileID, 3600)
	require.NoError(t, err)
	assert.NotEmpty(t, downloadURL)
	assert.Contains(t, downloadURL, "X-Amz-Signature=")
}

// Menguji pengerasan validasi unggahan: allowlist MIME, batas avatar 5MB, sanitasi nama berkas, dan clamp TTL unduhan (F-UPL-01).
func TestUploadValidationHardening(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	service := documents.NewStorageService(db, cfg)
	ctx := context.Background()

	// 1. MIME di luar allowlist ditolak (application/zip)
	_, err = service.GeneratePresignedUpload(ctx, &documents.PresignedUploadRequest{
		Folder: "task-evidence", Filename: "arsip.zip", MimeType: "application/zip", SizeBytes: 1024,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tipe berkas tidak diizinkan")

	// 2. Avatar profil di atas 5MB ditolak meskipun di bawah 25MB
	_, err = service.GeneratePresignedUpload(ctx, &documents.PresignedUploadRequest{
		Folder: "profile", Filename: "foto.png", MimeType: "image/png", SizeBytes: 6 * 1024 * 1024,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "5 MB")

	// 3. Avatar profil 1MB diterima
	avatarRes, err := service.GeneratePresignedUpload(ctx, &documents.PresignedUploadRequest{
		Folder: "profile", Filename: "foto.png", MimeType: "image/png", SizeBytes: 1024 * 1024,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM file_metadata WHERE id = $1", avatarRes.FileID)
	})

	// 4. Nama berkas dengan path traversal ditolak
	_, err = service.GeneratePresignedUpload(ctx, &documents.PresignedUploadRequest{
		Folder: "attachments", Filename: "../../etc/passwd", MimeType: "image/png", SizeBytes: 1024,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "karakter path")

	// 5. TTL unduhan arbitrer di-clamp ke batas server 3600 detik
	clampedURL, err := service.GeneratePresignedDownload(ctx, avatarRes.FileID, 999999)
	require.NoError(t, err)
	assert.Contains(t, clampedURL, "X-Amz-Expires=3600")
}

// Menguji konektivitas langsung ke Cloudflare R2 dengan mengunggah dan mengunduh berkas uji melalui presigned URL.
func TestR2LiveConnection(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	service := documents.NewStorageService(db, cfg)
	ctx := context.Background()

	testContent := []byte("%PDF-1.4 DCISP R2 Live Connection Verification: SUCCESS!")
	req := &documents.PresignedUploadRequest{
		Folder:    "attachments",
		Filename:  "live_check.pdf",
		MimeType:  "application/pdf",
		SizeBytes: int64(len(testContent)),
	}

	res, err := service.GeneratePresignedUpload(ctx, req)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM file_metadata WHERE id = $1", res.FileID)
	})

	// 1. Eksekusi HTTP PUT langsung ke upload_url R2
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, res.UploadURL, bytes.NewReader(testContent))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/pdf")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	require.NoError(t, err, "Gagal melakukan koneksi HTTP ke Cloudflare R2")
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Cloudflare R2 merespon error: %s", string(respBody))

	// 2. Eksekusi HTTP GET untuk memverifikasi file dapat diunduh
	downloadURL, err := service.GeneratePresignedDownload(ctx, res.FileID, 300)
	require.NoError(t, err)

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	require.NoError(t, err)

	getResp, err := client.Do(getReq)
	require.NoError(t, err)
	defer getResp.Body.Close()

	bodyDownloaded, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Equal(t, string(testContent), string(bodyDownloaded))
}
