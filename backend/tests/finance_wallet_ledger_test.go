package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/finance"
	"dcisp/backend/internal/modules/people"
	"dcisp/backend/internal/modules/projects"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type financeFixture struct {
	db              *database.PostgresDB
	repo            *finance.Repository
	service         *finance.Service
	peopleRepo      *people.Repository
	projectsRepo    *projects.Repository
	projectsService *projects.Service
	ctx             context.Context
	ownerID         uuid.UUID
	batchID         uuid.UUID
	userIDs         []uuid.UUID
	projectIDs      []uuid.UUID
	rewardIDs       []uuid.UUID
	tempRuleIDs     []uuid.UUID
}

// Menyiapkan fixture finansial: 1 owner, 1 batch aktif, dan 3 pengguna peserta magang.
func setupFinanceFixture(t *testing.T) *financeFixture {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)

	repo := finance.NewRepository(db)
	service := finance.NewService(repo, db, nil, nil)
	peopleRepo := people.NewRepository(db)
	projectsRepo := projects.NewRepository(db)
	service.SetContributionSource(finance.NewProjectsContributionAdapter(projectsRepo, peopleRepo, db))
	projectsService := projects.NewService(projectsRepo, db, nil, nil, nil)
	ctx := context.Background()

	fx := &financeFixture{
		db: db, repo: repo, service: service,
		peopleRepo: peopleRepo, projectsRepo: projectsRepo,
		projectsService: projectsService, ctx: ctx,
	}

	// Terapkan migrasi 000005 bila belum (idempoten)
	migUp := `CREATE UNIQUE INDEX IF NOT EXISTS uq_wallet_tx_reference ON wallet_transactions(reference_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_reward_claims_reward_user ON reward_claims(reward_id, user_id);
ALTER TABLE payouts ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128);
CREATE UNIQUE INDEX IF NOT EXISTS uq_payouts_idempotency ON payouts(idempotency_key);`
	_, _ = db.Pool.Exec(ctx, migUp)

	fx.ownerID = uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Finance Owner', 'ACTIVE')
	`, fx.ownerID, fmt.Sprintf("fin_owner_%s@dcisp.internal", fx.ownerID.String()[:8]))
	require.NoError(t, err)

	fx.batchID = uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO batches (id, batch_code, name, start_date, end_date, quota, status)
		VALUES ($1, $2, 'Batch Finance Test', '2026-01-01', '2026-12-31', 50, 'ACTIVE')
	`, fx.batchID, fmt.Sprintf("BATCH-FIN-%s", fx.batchID.String()[:6]))
	require.NoError(t, err)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO batch_funds (id, batch_id, total_accumulated, current_balance)
		VALUES (gen_random_uuid(), $1, 0, 0)
	`, fx.batchID)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		uID := uuid.New()
		fx.userIDs = append(fx.userIDs, uID)
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, status)
			VALUES ($1, $2, 'hash_pass', 'Finance Member', 'ACTIVE')
		`, uID, fmt.Sprintf("fin_member_%d_%s@dcisp.internal", i, uID.String()[:8]))
		require.NoError(t, err)

		_, err = db.Pool.Exec(ctx, `
			INSERT INTO interns (id, user_id, batch_id, status, join_date, end_date)
			VALUES (gen_random_uuid(), $1, $2, 'ACTIVE', '2026-01-01', '2026-12-31')
		`, uID, fx.batchID)
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		c := context.Background()
		for _, rID := range fx.tempRuleIDs {
			_, _ = db.Pool.Exec(c, `UPDATE deduction_tax_rules SET is_active = FALSE WHERE id = $1`, rID)
		}
		for _, pID := range fx.projectIDs {
			_, _ = db.Pool.Exec(c, `DELETE FROM project_teams WHERE project_id = $1`, pID)
			_, _ = db.Pool.Exec(c, `DELETE FROM projects WHERE id = $1`, pID)
		}
		for _, rID := range fx.rewardIDs {
			_, _ = db.Pool.Exec(c, `DELETE FROM reward_claims WHERE reward_id = $1`, rID)
			_, _ = db.Pool.Exec(c, `DELETE FROM rewards WHERE id = $1`, rID)
		}
		for _, uID := range fx.userIDs {
			_, _ = db.Pool.Exec(c, `DELETE FROM interns WHERE user_id = $1`, uID)
			_, _ = db.Pool.Exec(c, `DELETE FROM users WHERE id = $1`, uID)
		}
		_, _ = db.Pool.Exec(c, `DELETE FROM batch_funds WHERE batch_id = $1`, fx.batchID)
		_, _ = db.Pool.Exec(c, `DELETE FROM batches WHERE id = $1`, fx.batchID)
		_, _ = db.Pool.Exec(c, `DELETE FROM users WHERE id = $1`, fx.ownerID)
	})

	return fx
}

// Menguji mesin deduksi dinamis: formula Gross-Tax-Farewell-Net sesuai AC-FIN-001 (Rp1.000.000 → pajak Rp50.000, iuran Rp20.000, bersih Rp930.000).
func TestDeductionEngineGrossTaxFarewellNet(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	alloc, err := fx.service.EvaluateDeductions(fx.ctx, 1000000, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1000000), alloc.Gross)
	assert.Equal(t, int64(50000), alloc.TaxTotal)
	assert.Equal(t, int64(20000), alloc.FarewellTotal)
	assert.Equal(t, int64(930000), alloc.Net)
	assert.Equal(t, alloc.Gross, alloc.TaxTotal+alloc.FarewellTotal+alloc.OtherTotal+alloc.Net)

	// Nilai bruto nol menghasilkan alokasi nol tanpa error
	zeroAlloc, err := fx.service.EvaluateDeductions(fx.ctx, 0, false)
	require.NoError(t, err)
	assert.Equal(t, int64(0), zeroAlloc.Net)

	// Nilai bruto negatif ditolak (proteksi injeksi)
	_, err = fx.service.EvaluateDeductions(fx.ctx, -1000, false)
	assert.Error(t, err)
}

// Menguji dompet personal: pembuatan otomatis, kredit, debit terjaga, dan penolakan saldo minus (BRULE-FIN-005).
func TestPersonalWalletNonNegativeBalance(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	w, err := fx.service.GetWalletByUserID(fx.ctx, fx.userIDs[0])
	require.NoError(t, err)
	assert.Equal(t, int64(0), w.CurrentBalance)

	// Kredit manual via transaksi internal untuk pengujian guard
	tx, err := fx.db.Pool.Begin(fx.ctx)
	require.NoError(t, err)
	require.NoError(t, fx.repo.CreditWalletTx(fx.ctx, tx, w.ID, 2000000))
	require.NoError(t, tx.Commit(fx.ctx))

	// Debit melebihi saldo ditolak secara atomik
	tx2, err := fx.db.Pool.Begin(fx.ctx)
	require.NoError(t, err)
	err = fx.repo.DebitWalletGuardedTx(fx.ctx, tx2, w.ID, 2500000)
	assert.Error(t, err)
	tx2.Rollback(fx.ctx)

	// Debit sah berhasil
	tx3, err := fx.db.Pool.Begin(fx.ctx)
	require.NoError(t, err)
	require.NoError(t, fx.repo.DebitWalletGuardedTx(fx.ctx, tx3, w.ID, 500000))
	require.NoError(t, tx3.Commit(fx.ctx))

	wAfter, err := fx.repo.GetWalletByUserID(fx.ctx, fx.userIDs[0])
	require.NoError(t, err)
	assert.Equal(t, int64(1500000), wAfter.CurrentBalance)
}

// Menguji ketahanan dompet terhadap 10 debit konkuren: hanya yang didukung saldo yang lolos tanpa race.
func TestWalletConcurrentDebitGuard(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	w, err := fx.service.GetWalletByUserID(fx.ctx, fx.userIDs[1])
	require.NoError(t, err)

	tx, err := fx.db.Pool.Begin(fx.ctx)
	require.NoError(t, err)
	require.NoError(t, fx.repo.CreditWalletTx(fx.ctx, tx, w.ID, 1000000))
	require.NoError(t, tx.Commit(fx.ctx))

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	// 10 goroutine masing-masing mendebit Rp300.000 dari saldo Rp1.000.000 → tepat 3 yang lolos
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			innerTx, err := fx.db.Pool.Begin(fx.ctx)
			if err != nil {
				return
			}
			defer innerTx.Rollback(fx.ctx)
			if err := fx.repo.DebitWalletGuardedTx(fx.ctx, innerTx, w.ID, 300000); err != nil {
				return
			}
			if err := innerTx.Commit(fx.ctx); err != nil {
				return
			}
			mu.Lock()
			successCount++
			mu.Unlock()
		}()
	}
	wg.Wait()

	assert.Equal(t, 3, successCount)
	wFinal, err := fx.repo.GetWalletByUserID(fx.ctx, fx.userIDs[1])
	require.NoError(t, err)
	assert.Equal(t, int64(100000), wFinal.CurrentBalance)
	assert.True(t, wFinal.CurrentBalance >= 0)
}

// Menguji buku besar double-entry: jurnal seimbang diterima, jurnal timpang ditolak, dan reversal mencerminkan saldo nol.
func TestDoubleEntryLedgerBalanceAndReversal(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	balancedID := uuid.New()
	refID := uuid.New()
	entries := []finance.FinancialLedger{
		{TransactionID: balancedID, AccountCode: finance.AcctCash, Direction: finance.DirectionDebit, Amount: 1000000, ReferenceTable: "tests", ReferenceID: refID, Narration: "Uji debit kas"},
		{TransactionID: balancedID, AccountCode: finance.AcctMemberLiability, Direction: finance.DirectionCredit, Amount: 1000000, ReferenceTable: "tests", ReferenceID: refID, Narration: "Uji kredit utang"},
	}

	tx, err := fx.db.Pool.Begin(fx.ctx)
	require.NoError(t, err)
	require.NoError(t, fx.repo.AppendLedgerEntriesTx(fx.ctx, tx, entries))
	require.NoError(t, tx.Commit(fx.ctx))

	stored, err := fx.repo.GetLedgerEntriesByTransaction(fx.ctx, balancedID)
	require.NoError(t, err)
	assert.Len(t, stored, 2)

	// Jurnal timpang wajib ditolak di level aplikasi
	unbalancedID := uuid.New()
	badEntries := []finance.FinancialLedger{
		{TransactionID: unbalancedID, AccountCode: finance.AcctCash, Direction: finance.DirectionDebit, Amount: 1000000, ReferenceTable: "tests", ReferenceID: refID, Narration: "Debit timpang"},
		{TransactionID: unbalancedID, AccountCode: finance.AcctMemberLiability, Direction: finance.DirectionCredit, Amount: 999999, ReferenceTable: "tests", ReferenceID: refID, Narration: "Kredit timpang"},
	}
	txBad, err := fx.db.Pool.Begin(fx.ctx)
	require.NoError(t, err)
	err = fx.repo.AppendLedgerEntriesTx(fx.ctx, txBad, badEntries)
	assert.Error(t, err)
	txBad.Rollback(fx.ctx)

	// Reversal menghasilkan cerminan sempurna sehingga neto transaksi menjadi nol
	reversalID, err := fx.service.ReverseLedgerTransaction(fx.ctx, balancedID, "Koreksi pengujian")
	require.NoError(t, err)
	mirror, err := fx.repo.GetLedgerEntriesByTransaction(fx.ctx, reversalID)
	require.NoError(t, err)
	assert.Len(t, mirror, 2)
	var net int64
	for _, e := range append(stored, mirror...) {
		if e.Direction == finance.DirectionDebit {
			net += e.Amount
		} else {
			net -= e.Amount
		}
	}
	assert.Equal(t, int64(0), net)
}
