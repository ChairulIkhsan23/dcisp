package tests

import (
	"context"
	"net/url"
	"testing"

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
