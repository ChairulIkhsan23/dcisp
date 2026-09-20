package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/integrations/apiindonesia"
	"dcisp/backend/internal/modules/people"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Membuat server mock yang meniru kontrak Public API API Indonesia (kampus + sekolah).
func newMockAPIIndonesiaServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/kampus", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"MISSING_API_KEY","message":"Header x-api-key tidak dikirim"}}`))
			return
		}
		q := r.URL.Query().Get("q")
		items := []map[string]interface{}{
			{
				"id": "pt_001", "name": "INSTITUT TEKNOLOGI BANDUNG", "short_name": "ITB",
				"jenis": "institut", "kelompok": "PTN", "province_id": "32", "regency_id": "3273",
				"address": "Jl. Ganesha No. 10", "postal_code": "40132", "website": "https://itb.ac.id",
				"accreditation": "Unggul", "is_active": 1, "province_name": "JAWA BARAT", "regency_name": "KOTA BANDUNG",
			},
		}
		if q == "kosong-tidak-ada-hasil-xyz" {
			items = []map[string]interface{}{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": items,
			"meta": map[string]interface{}{"total": len(items), "page": 1, "per_page": 20, "total_pages": 1},
		})
	})

	mux.HandleFunc("/api/v1/kampus/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"MISSING_API_KEY","message":"Header x-api-key tidak dikirim"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "pt_001", "name": "INSTITUT TEKNOLOGI BANDUNG", "short_name": "ITB",
				"jenis": "institut", "kelompok": "PTN", "province_id": "32", "regency_id": "3273",
				"address": "Jl. Ganesha No. 10", "postal_code": "40132",
				"phone": "0222500935", "email": "info@itb.ac.id", "website": "https://itb.ac.id",
				"accreditation": "Unggul", "is_active": 1, "province_name": "JAWA BARAT", "regency_name": "KOTA BANDUNG",
			},
		})
	})

	mux.HandleFunc("/api/v1/sekolah", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"MISSING_API_KEY","message":"Header x-api-key tidak dikirim"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"npsn": "20219557", "name": "SMA NEGERI 3 BANDUNG", "jenis": "SMA", "status": "Negeri",
					"province_id": "32", "regency_id": "3273", "address": "Jl. Belitung No. 8",
					"postal_code": "40113", "accreditation": "A",
				},
			},
			"meta": map[string]interface{}{"total": 1, "page": 1, "per_page": 20, "total_pages": 1},
		})
	})

	mux.HandleFunc("/api/v1/sekolah/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"MISSING_API_KEY","message":"Header x-api-key tidak dikirim"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"npsn": "20219557", "name": "SMA NEGERI 3 BANDUNG", "jenis": "SMA", "status": "Negeri",
				"province_id": "32", "regency_id": "3273", "address": "Jl. Belitung No. 8",
				"postal_code": "40113", "accreditation": "A",
			},
		})
	})

	return httptest.NewServer(mux)
}

// Menguji pencarian kampus eksternal, detail, error auth, dan respons kosong melalui mock API Indonesia.
func TestAPIIndonesiaClientKampusAndErrors(t *testing.T) {
	mockServer := newMockAPIIndonesiaServer()
	defer mockServer.Close()

	client := apiindonesia.NewClient(mockServer.URL, "aip_live_test_key")
	ctx := context.Background()

	// 1. Pencarian sukses
	res, err := client.SearchKampus(ctx, apiindonesia.SearchKampusParams{Query: "teknologi", Page: 1, PerPage: 20})
	require.NoError(t, err)
	require.Len(t, res.Items, 1)
	assert.Equal(t, "pt_001", res.Items[0].ID)
	assert.Equal(t, "INSTITUT TEKNOLOGI BANDUNG", res.Items[0].Name)

	// 2. Respons kosong
	emptyRes, err := client.SearchKampus(ctx, apiindonesia.SearchKampusParams{Query: "kosong-tidak-ada-hasil-xyz"})
	require.NoError(t, err)
	assert.Len(t, emptyRes.Items, 0)

	// 3. Detail kampus sukses beserta field kontak aktual API (phone/email)
	detail, err := client.GetKampusDetail(ctx, "pt_001")
	require.NoError(t, err)
	assert.Equal(t, "INSTITUT TEKNOLOGI BANDUNG", detail.Name)
	require.NotNil(t, detail.Phone)
	assert.Equal(t, "0222500935", *detail.Phone)
	require.NotNil(t, detail.Email)
	assert.Equal(t, "info@itb.ac.id", *detail.Email)

	// 4. Kredensial kosong ditolak sebelum request jaringan
	badClient := apiindonesia.NewClient(mockServer.URL, "")
	_, err = badClient.SearchKampus(ctx, apiindonesia.SearchKampusParams{Query: "itb"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API_INDONESIA_KEY")

	// 5. Header auth hilang dari server mock menghasilkan 401 yang dipetakan aman
	noAuthClient := apiindonesia.NewClient(mockServer.URL, "   ")
	_, err = noAuthClient.SearchSekolah(ctx, apiindonesia.SearchSekolahParams{Query: "sma"})
	assert.Error(t, err)
}

// Menguji mapper external DTO menjadi draf institusi internal tanpa membocorkan kredensial.
func TestAPIIndonesiaMapperToInternalDraft(t *testing.T) {
	shortName := "ITB"
	provinceName := "JAWA BARAT"
	regencyName := "KOTA BANDUNG"
	address := "Jl. Ganesha No. 10"
	phone := "0222500935"
	email := "info@itb.ac.id"
	ext := &apiindonesia.Kampus{
		ID: "pt_001", Name: "INSTITUT TEKNOLOGI BANDUNG", ShortName: &shortName,
		Jenis: "institut", Kelompok: "PTN", ProvinceID: "32", RegencyID: "3273",
		Address: &address, Phone: &phone, Email: &email,
		ProvinceName: &provinceName, RegencyName: &regencyName,
	}
	draft := people.MapKampusToInstitutionDraft(ext)
	assert.Equal(t, "INSTITUT TEKNOLOGI BANDUNG", draft.Name)
	require.NotNil(t, draft.ExternalID)
	assert.Equal(t, "kampus:pt_001", *draft.ExternalID)
	require.NotNil(t, draft.Source)
	assert.Equal(t, "API_KAMPUS", *draft.Source)
	assert.Contains(t, *draft.Address, "Jl. Ganesha No. 10")
	require.NotNil(t, draft.Phone)
	assert.Equal(t, "0222500935", *draft.Phone)
	require.NotNil(t, draft.Email)
	assert.Equal(t, "info@itb.ac.id", *draft.Email)

	sekolahExt := &apiindonesia.Sekolah{NPSN: "20219557", Name: "SMA NEGERI 3 BANDUNG", Jenis: "SMA", Status: "Negeri"}
	sekolahDraft := people.MapSekolahToInstitutionDraft(sekolahExt)
	assert.Equal(t, "sekolah:20219557", *sekolahDraft.ExternalID)
}

// Menguji alur impor idempoten institusi eksternal ke database internal dan pencegahan duplikasi.
func TestImportExternalInstitutionIdempotent(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	mockServer := newMockAPIIndonesiaServer()
	defer mockServer.Close()

	repo := people.NewRepository(db)
	service := people.NewService(repo, db, nil, nil)
	client := apiindonesia.NewClient(mockServer.URL, "aip_live_test_key")
	service.SetExternalDependencies(client, nil)
	ctx := context.Background()

	// Pastikan kolom eksternal tersedia (idempoten terhadap migrasi 000004)
	_, _ = db.Pool.Exec(ctx, `ALTER TABLE institutions ADD COLUMN IF NOT EXISTS external_id VARCHAR(128)`)
	_, _ = db.Pool.Exec(ctx, `ALTER TABLE institutions ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'MANUAL'`)
	_, _ = db.Pool.Exec(ctx, `ALTER TABLE institutions ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ`)
	_, _ = db.Pool.Exec(ctx, `DELETE FROM institutions WHERE external_id = 'kampus:pt_001'`)

	// 1. Impor pertama harus membuat record baru (isNew = true)
	inst1, isNew1, err := service.ImportExternalInstitution(ctx, &people.ImportExternalInstitutionRequest{
		Source: "API_KAMPUS", ExternalID: "pt_001",
	})
	require.NoError(t, err)
	assert.True(t, isNew1)
	assert.Equal(t, "INSTITUT TEKNOLOGI BANDUNG", inst1.Name)

	// 2. Impor kedua dengan external_id sama harus idempoten (isNew = false, ID sama)
	inst2, isNew2, err := service.ImportExternalInstitution(ctx, &people.ImportExternalInstitutionRequest{
		Source: "API_KAMPUS", ExternalID: "kampus:pt_001",
	})
	require.NoError(t, err)
	assert.False(t, isNew2)
	assert.Equal(t, inst1.ID, inst2.ID)

	// 3. Pencarian eksternal dengan cache dilewati (tanpa Redis) tetap mengembalikan hasil mock
	searchRes, err := service.SearchExternalInstitutions(ctx, "teknologi", "KAMPUS", "", "", 1, 20)
	require.NoError(t, err)
	require.NotNil(t, searchRes.Kampus)
	assert.Len(t, searchRes.Kampus.Items, 1)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM institutions WHERE id = $1", inst1.ID)
	})

	_ = time.Now()
	_ = fmt.Sprintf("cleanup-%s", uuid.New().String()[:8])
}
