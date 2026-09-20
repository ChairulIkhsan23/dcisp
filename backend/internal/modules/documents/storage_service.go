package documents

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StorageService struct {
	db  *database.PostgresDB
	cfg *config.Config
}

// Menginisialisasi instance baru service penyimpanan Cloudflare R2 dan metadata berkas.
func NewStorageService(db *database.PostgresDB, cfg *config.Config) *StorageService {
	return &StorageService{db: db, cfg: cfg}
}

// Menghasilkan tanda tangan HMAC-SHA256 berdasarkan kunci dan data masukan.
func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// Menghitung kunci penandatanganan AWS Signature Version 4 untuk Cloudflare R2.
func getSigningKey(secretKey, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	return kSigning
}

// Menghasilkan Presigned PUT URL aman untuk unggah langsung ke Cloudflare R2 dan menyimpan rekod metadatanya.
func (s *StorageService) GeneratePresignedUpload(ctx context.Context, req *PresignedUploadRequest) (*PresignedUploadResponse, error) {
	// Validasi ukuran berkas (Maksimal 25MB sesuai SECURITY.md)
	const maxSizeBytes = 25 * 1024 * 1024
	if req.SizeBytes > maxSizeBytes {
		return nil, fmt.Errorf("ukuran berkas melebihi batas maksimal 25 MB: %d bytes", req.SizeBytes)
	}

	// Validasi folder partisi (FR-044)
	allowedFolders := map[string]bool{
		"task-evidence": true,
		"profile":       true,
		"certificates":  true,
		"id-cards":      true,
		"logbooks":      true,
		"rewards":       true,
		"gallery":       true,
		"reports":       true,
		"attachments":   true,
		"project":       true,
	}
	cleanFolder := strings.Trim(req.Folder, "/\\")
	if !allowedFolders[cleanFolder] {
		return nil, fmt.Errorf("partisi folder tidak diizinkan: %s", req.Folder)
	}

	fileID := uuid.New()
	ext := path.Ext(req.Filename)
	now := time.Now().UTC()
	pathKey := fmt.Sprintf("%s/%s/%s%s", cleanFolder, now.Format("2006/01"), fileID.String(), ext)

	// Buat rekod metadata berkas pada PostgreSQL
	query := `
		INSERT INTO file_metadata (id, disk, path_key, filename, mime_type, size_bytes, metadata, created_at)
		VALUES ($1, 'r2', $2, $3, $4, $5, '{}'::jsonb, CURRENT_TIMESTAMP)
	`
	_, err := s.db.Pool.Exec(ctx, query, fileID, pathKey, req.Filename, req.MimeType, req.SizeBytes)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan metadata berkas: %w", err)
	}

	// Buat presigned PUT URL SigV4
	expiresIn := 900 // 15 menit
	uploadURL, err := s.signURL("PUT", pathKey, expiresIn, now)
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani URL unggah presigned: %w", err)
	}

	return &PresignedUploadResponse{
		FileID:           fileID,
		UploadURL:        uploadURL,
		PathKey:          pathKey,
		ExpiresInSeconds: expiresIn,
	}, nil
}

// Menghasilkan Presigned GET URL untuk unduh aman berkas privat dari Cloudflare R2.
func (s *StorageService) GeneratePresignedDownload(ctx context.Context, fileID uuid.UUID, expiresInSeconds int) (string, error) {
	var pathKey string
	err := s.db.Pool.QueryRow(ctx, "SELECT path_key FROM file_metadata WHERE id = $1", fileID).Scan(&pathKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("metadata berkas tidak ditemukan")
		}
		return "", fmt.Errorf("gagal mengambil path berkas: %w", err)
	}

	if expiresInSeconds <= 0 {
		expiresInSeconds = 3600 // 60 menit default
	}

	return s.signURL("GET", pathKey, expiresInSeconds, time.Now().UTC())
}

// Menandatangani URL HTTP dengan algoritma AWS SigV4 untuk Cloudflare R2 S3 API.
func (s *StorageService) signURL(httpMethod, pathKey string, expiresInSeconds int, t time.Time) (string, error) {
	accountID := s.cfg.R2AccountID
	accessKey := s.cfg.R2AccessKeyID
	secretKey := s.cfg.R2SecretAccessKey
	bucketName := s.cfg.R2BucketName

	host := fmt.Sprintf("%s.r2.cloudflarestorage.com", accountID)
	region := "auto"
	service := "s3"

	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	credential := fmt.Sprintf("%s/%s/%s/%s/aws4_request", accessKey, dateStamp, region, service)

	canonicalURI := fmt.Sprintf("/%s/%s", bucketName, pathKey)
	canonicalHeaders := fmt.Sprintf("host:%s\n", host)
	signedHeaders := "host"

	queryParams := url.Values{}
	queryParams.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	queryParams.Set("X-Amz-Credential", credential)
	queryParams.Set("X-Amz-Date", amzDate)
	queryParams.Set("X-Amz-Expires", fmt.Sprintf("%d", expiresInSeconds))
	queryParams.Set("X-Amz-SignedHeaders", signedHeaders)

	canonicalQueryString := queryParams.Encode()

	payloadHash := "UNSIGNED-PAYLOAD"
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)

	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate,
		fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service),
		hex.EncodeToString(canonicalRequestHash[:]),
	)

	signingKey := getSigningKey(secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	finalURL := fmt.Sprintf("https://%s%s?%s&X-Amz-Signature=%s",
		host,
		canonicalURI,
		canonicalQueryString,
		signature,
	)

	return finalURL, nil
}

// Mengambil metadata informasi berkas dari database berdasarkan identitas unik.
func (s *StorageService) GetFileMetadata(ctx context.Context, fileID uuid.UUID) (*FileMetadata, error) {
	query := `
		SELECT id, disk, path_key, filename, mime_type, size_bytes, metadata, created_at
		FROM file_metadata
		WHERE id = $1
	`
	var m FileMetadata
	err := s.db.Pool.QueryRow(ctx, query, fileID).Scan(&m.ID, &m.Disk, &m.PathKey, &m.Filename, &m.MimeType, &m.SizeBytes, &m.Metadata, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari metadata berkas: %w", err)
	}
	return &m, nil
}
