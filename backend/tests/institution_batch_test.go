package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/people"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji siklus CRUD institusi mitra dan perlindungan penolakan penghapusan jika masih memiliki peserta (BRULE-PEO-005).
func TestInstitutionCRUDAndDeletionGuard(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := people.NewRepository(db)
	service := people.NewService(repo, db, nil, nil)
	ctx := context.Background()

	// 1. Buat institusi baru
	inst, err := service.CreateInstitution(ctx, &people.CreateInstitutionRequest{
		Name:  "Universitas Uji Kemitraan",
		Email: func(s string) *string { return &s }("kemitraan@kampus.ac.id"),
	})
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, inst.ID)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM institutions WHERE id = $1", inst.ID)
	})

	// 2. Ambil detail institusi
	fetched, err := service.GetInstitutionByID(ctx, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, "Universitas Uji Kemitraan", fetched.Name)

	// 3. Perbarui institusi
	updated, err := service.UpdateInstitution(ctx, inst.ID, &people.UpdateInstitutionRequest{
		Name: "Universitas Kemitraan Diperbarui",
	})
	require.NoError(t, err)
	assert.Equal(t, "Universitas Kemitraan Diperbarui", updated.Name)

	// 4. Simulasi keterikatan peserta magang untuk menguji Delete Guard (BRULE-PEO-005)
	testUserID := uuid.New()
	testEmail := "intern_inst_" + testUserID.String()[:8] + "@dcisp.internal"
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Intern Inst Tester', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	testBatchID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO batches (id, batch_code, name, start_date, end_date, quota, status)
		VALUES ($1, $2, 'Batch Inst Tester', '2026-01-01', '2026-06-30', 10, 'ACTIVE')
	`, testBatchID, "BATCH-INST-"+testBatchID.String()[:6])
	require.NoError(t, err)

	testInternID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO interns (id, user_id, batch_id, institution_id, status, join_date, end_date)
		VALUES ($1, $2, $3, $4, 'ACTIVE', '2026-01-01', '2026-06-30')
	`, testInternID, testUserID, testBatchID, inst.ID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM interns WHERE id = $1", testInternID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id = $1", testBatchID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	// Upaya penghapusan harus gagal karena ada peserta magang terhubung
	err = service.DeleteInstitution(ctx, inst.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "masih memiliki")

	// Hapus intern terlebih dahulu, kemudian penghapusan institusi harus berhasil
	_, err = db.Pool.Exec(ctx, "DELETE FROM interns WHERE id = $1", testInternID)
	require.NoError(t, err)

	err = service.DeleteInstitution(ctx, inst.ID)
	require.NoError(t, err)
}

// Menguji pembuatan kohort batch baru beserta inisialisasi kas bersama Batch Fund secara atomik (FR-004, BRULE-PEO-004).
func TestBatchCreationAndFundAutoInit(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := people.NewRepository(db)
	service := people.NewService(repo, db, nil, nil)
	ctx := context.Background()

	code := fmt.Sprintf("BATCH-%d", time.Now().UnixNano()%1000000)
	req := &people.CreateBatchRequest{
		BatchCode: code,
		Name:      "Angkatan Uji Kas Bersama",
		StartDate: "2026-09-01",
		EndDate:   "2027-02-28",
		Quota:     25,
		Status:    people.BatchStatusActive,
	}

	batch, fund, err := service.CreateBatch(ctx, req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, batch.ID)
	assert.NotEqual(t, uuid.Nil, fund.ID)
	assert.Equal(t, batch.ID, fund.BatchID)
	assert.Equal(t, 0.00, fund.CurrentBalance)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batch_funds WHERE batch_id = $1", batch.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id = $1", batch.ID)
	})

	// Verifikasi rekod kas batch tersimpan di database
	fetchedBatch, fetchedFund, err := service.GetBatchByID(ctx, batch.ID)
	require.NoError(t, err)
	assert.Equal(t, code, fetchedBatch.BatchCode)
	assert.NotNil(t, fetchedFund)
	assert.Equal(t, fund.ID, fetchedFund.ID)
}

// Menguji batas kuota kohort batch pada kondisi permintaan konkuren untuk mencegah kelebihan alokasi kuota.
func TestBatchQuotaConcurrencyGuard(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := people.NewRepository(db)
	service := people.NewService(repo, db, nil, nil)
	ctx := context.Background()

	// Buat batch dengan kuota ketat 2 peserta
	code := fmt.Sprintf("BATCH-Q-%d", time.Now().UnixNano()%1000000)
	batch, _, err := service.CreateBatch(ctx, &people.CreateBatchRequest{
		BatchCode: code,
		Name:      "Batch Kuota Ketat",
		StartDate: "2026-09-01",
		EndDate:   "2027-02-28",
		Quota:     2,
		Status:    people.BatchStatusActive,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batch_funds WHERE batch_id = $1", batch.ID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM batches WHERE id = $1", batch.ID)
	})

	// Buat 4 kandidat peserta magang
	var internIDs []uuid.UUID
	var userIDs []uuid.UUID

	for i := 0; i < 4; i++ {
		uID := uuid.New()
		userIDs = append(userIDs, uID)
		uEmail := fmt.Sprintf("candidate_%d_%s@dcisp.internal", i, uID.String()[:8])
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, status)
			VALUES ($1, $2, 'hash_pass', 'Candidate Tester', 'ACTIVE')
		`, uID, uEmail)
		require.NoError(t, err)

		intern, err := service.RegisterIntern(ctx, &people.RegisterInternRequest{
			UserID:   uID,
			BatchID:  batch.ID,
			JoinDate: "2026-09-01",
			EndDate:  "2027-02-28",
		})
		require.NoError(t, err)
		internIDs = append(internIDs, intern.ID)
	}

	t.Cleanup(func() {
		for _, iID := range internIDs {
			_, _ = db.Pool.Exec(context.Background(), "DELETE FROM interns WHERE id = $1", iID)
		}
		for _, uID := range userIDs {
			_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", uID)
		}
	})

	// Transisi 2 peserta pertama ke ONBOARDING lalu ke ACTIVE (sampai kuota penuh)
	for i := 0; i < 2; i++ {
		_, err = service.ChangeInternStatus(ctx, internIDs[i], people.InternStatusOnboarding, "Lolos seleksi")
		require.NoError(t, err)
		_, err = service.ChangeInternStatus(ctx, internIDs[i], people.InternStatusActive, "Aktivasi magang")
		require.NoError(t, err)
	}

	// Peserta ke-3 mencoba aktivasi ke ACTIVE konkuren dengan peserta ke-4
	var wg sync.WaitGroup
	var errorCount int
	var successCount int
	var mu sync.Mutex

	for i := 2; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = service.ChangeInternStatus(ctx, internIDs[idx], people.InternStatusOnboarding, "Lolos seleksi")
			_, err := service.ChangeInternStatus(ctx, internIDs[idx], people.InternStatusActive, "Aktivasi")
			mu.Lock()
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	// Kuota hanya 2, maka seluruh aktivasi tambahan harus ditolak
	assert.Equal(t, 2, errorCount)
	assert.Equal(t, 0, successCount)
}
