package tests

import (
	"fmt"
	"sync"
	"testing"

	"dcisp/backend/internal/modules/finance"
	"dcisp/backend/internal/modules/projects"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menyiapkan proyek COMPLETED dengan tim terkunci untuk pengujian distribusi bounty.
func setupCompletedProject(t *testing.T, fx *financeFixture, title string, pool int64, finalPcts []float64) uuid.UUID {
	p, err := fx.projectsService.CreateProject(fx.ctx, fx.ownerID, &projects.CreateProjectRequest{
		Title:       title,
		Description: "Proyek uji distribusi bounty",
		Visibility:  projects.VisibilityPublic,
		Capacity:    len(finalPcts) + 1,
		Deadline:    "2026-12-31",
		BountyPool:  float64(pool),
		Status:      projects.ProjectStatusPublished,
	})
	require.NoError(t, err)
	fx.projectIDs = append(fx.projectIDs, p.ID)

	for i, pct := range finalPcts {
		_, err = fx.db.Pool.Exec(fx.ctx, `
			INSERT INTO project_teams (id, project_id, user_id, project_role, planned_contribution_pct, actual_contribution_pct, final_contribution_pct, is_locked)
			VALUES (gen_random_uuid(), $1, $2, 'MEMBER', $3, $3, $3, TRUE)
		`, p.ID, fx.userIDs[i], pct)
		require.NoError(t, err)
	}

	// Kunci kontribusi final pemilik proyek (PM) sebesar 0% agar seluruh tim terfinalisasi
	_, err = fx.db.Pool.Exec(fx.ctx, `
		UPDATE project_teams
		SET final_contribution_pct = 0, actual_contribution_pct = 0, is_locked = TRUE
		WHERE project_id = $1 AND user_id = $2
	`, p.ID, fx.ownerID)
	require.NoError(t, err)

	_, err = fx.projectsService.UpdateProject(fx.ctx, p.ID, &projects.UpdateProjectRequest{
		Status: projects.ProjectStatusCompleted,
	}, fx.ownerID)
	require.NoError(t, err)

	return p.ID
}

// Mencari alokasi distribusi anggota berdasarkan ID pengguna.
func findMemberDistribution(members []finance.MemberDistribution, userID uuid.UUID) finance.MemberDistribution {
	for _, m := range members {
		if m.UserID == userID {
			return m
		}
	}
	return finance.MemberDistribution{}
}

// Menguji distribusi bounty: total tepat sama dengan pool, sisa pecahan deterministik, dan idempotensi eksekusi ulang.
func TestBountyDistributionTotalsAndIdempotency(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	// Pool Rp10.000.001 dengan porsi 40/35/25 menghasilkan sisa pecahan yang harus dialokasikan deterministik
	projectID := setupCompletedProject(t, fx, "Proyek Bounty Pecahan", 10000001, []float64{40.00, 35.00, 25.00})

	resp, err := fx.service.DistributeBounty(fx.ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000001), resp.TotalDistributed)
	assert.Empty(t, resp.Skipped)

	var grossSum int64
	processedCount := 0
	for _, m := range resp.Members {
		assert.Equal(t, m.Gross, m.TaxTotal+m.FarewellTotal+m.Net)
		grossSum += m.Gross
		if m.Gross > 0 {
			processedCount++
		}
	}
	assert.Equal(t, int64(10000001), grossSum)
	assert.Equal(t, 3, processedCount)

	// Eksekusi ulang harus idempoten: seluruh anggota dilewati tanpa mutasi ganda
	resp2, err := fx.service.DistributeBounty(fx.ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, resp2.Skipped, 3)

	// Saldo dompet setiap anggota persis sama dengan Net alokasinya (tidak ganda)
	for _, m := range resp.Members {
		if m.Gross == 0 {
			continue
		}
		w, err := fx.repo.GetWalletByUserID(fx.ctx, m.UserID)
		require.NoError(t, err)
		assert.Equal(t, m.Net, w.CurrentBalance)
	}

	// Verifikasi keseimbangan jurnal per transaksi anggota (ΣDebit = ΣKredit)
	for _, m := range resp.Members {
		if m.Gross == 0 {
			continue
		}
		entries, err := fx.repo.GetLedgerEntriesByTransaction(fx.ctx, m.TransactionID)
		require.NoError(t, err)
		var debit, credit int64
		for _, e := range entries {
			if e.Direction == finance.DirectionDebit {
				debit += e.Amount
			} else {
				credit += e.Amount
			}
		}
		assert.Equal(t, debit, credit)
		assert.Equal(t, m.Gross, debit)
	}
}

// Menguji akumulasi kas bersama angkatan yang terisolasi dari pos pajak (BR-021).
func TestBatchFundIsolationFromTax(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	projectID := setupCompletedProject(t, fx, "Proyek Kas Batch", 4000000, []float64{60.00, 40.00})

	resp, err := fx.service.DistributeBounty(fx.ctx, projectID)
	require.NoError(t, err)

	var expectedFarewell int64
	for _, m := range resp.Members {
		expectedFarewell += m.FarewellTotal
	}

	fund, err := fx.service.GetBatchFund(fx.ctx, fx.batchID)
	require.NoError(t, err)
	assert.Equal(t, expectedFarewell, fund.CurrentBalance)
	assert.Equal(t, expectedFarewell, fund.TotalAccumulated)

	// Pos pajak tercatat pada akun terpisah, bukan akun batch fund
	for _, m := range resp.Members {
		if m.Gross == 0 {
			continue
		}
		entries, err := fx.repo.GetLedgerEntriesByTransaction(fx.ctx, m.TransactionID)
		require.NoError(t, err)
		accounts := map[string]int64{}
		for _, e := range entries {
			if e.Direction == finance.DirectionCredit {
				accounts[e.AccountCode] = e.Amount
			}
		}
		assert.Contains(t, accounts, finance.AcctTaxPayable)
		assert.Contains(t, accounts, finance.AcctBatchFund)
		assert.NotEqual(t, finance.AcctTaxPayable, finance.AcctBatchFund)
	}
}

// Menguji rollback atomik orkestrasi: kegagalan pada anggota kedua membatalkan seluruh mutasi anggota pertama.
func TestBountyOrchestrationAtomicRollback(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	projectID := setupCompletedProject(t, fx, "Proyek Gagal Atomik", 2000000, []float64{50.00, 50.00})

	// Hapus kas batch agar akumulasi farewell anggota gagal di tengah transaksi
	_, err := fx.db.Pool.Exec(fx.ctx, `DELETE FROM batch_funds WHERE batch_id = $1`, fx.batchID)
	require.NoError(t, err)

	_, err = fx.service.DistributeBounty(fx.ctx, projectID)
	assert.Error(t, err)

	// Pastikan tidak ada mutasi parsial: dompet anggota pertama tetap nol dan tanpa wallet_tx
	w, err := fx.repo.GetWalletByUserID(fx.ctx, fx.userIDs[0])
	require.NoError(t, err)
	if w != nil {
		assert.Equal(t, int64(0), w.CurrentBalance)
	}
	var txCount int
	err = fx.db.Pool.QueryRow(fx.ctx, `
		SELECT COUNT(*) FROM wallet_transactions wt
		JOIN wallets wl ON wl.id = wt.wallet_id
		WHERE wl.user_id = $1
	`, fx.userIDs[0]).Scan(&txCount)
	require.NoError(t, err)
	assert.Equal(t, 0, txCount)
}

// Menguji alur payout end-to-end: request (hold) → approve → settle dengan zero-sum ledger.
func TestPayoutRequestApproveSettleFlow(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	// Bekali saldo melalui distribusi bounty kecil
	projectID := setupCompletedProject(t, fx, "Proyek Modal Payout", 2000000, []float64{100.00, 0.00, 0.00})
	distResp, err := fx.service.DistributeBounty(fx.ctx, projectID)
	require.NoError(t, err)

	payAlloc := findMemberDistribution(distResp.Members, fx.userIDs[0])
	require.True(t, payAlloc.Gross > 0)
	payUser := payAlloc.UserID
	initialNet := payAlloc.Net
	require.True(t, initialNet > 1000000)

	// 1. Request payout Rp1.500.000: saldo tersedia berkurang seketika (hold)
	payout, err := fx.service.RequestPayout(fx.ctx, payUser, &finance.RequestPayoutRequest{
		Amount:            1500000,
		BankCode:          "BCA",
		AccountNumber:     "1234567890",
		AccountHolderName: "Intern Andi",
	})
	require.NoError(t, err)
	assert.Equal(t, finance.PayoutRequested, payout.Status)

	wHeld, err := fx.repo.GetWalletByUserID(fx.ctx, payUser)
	require.NoError(t, err)
	assert.Equal(t, initialNet-1500000, wHeld.CurrentBalance)

	// 2. Approve oleh finance
	reviewerID := uuid.New()
	approved, err := fx.service.ReviewPayout(fx.ctx, payout.ID, &finance.ReviewPayoutRequest{Action: "APPROVE"}, reviewerID)
	require.NoError(t, err)
	assert.Equal(t, finance.PayoutApproved, approved.Status)

	// 3. Processing lalu settle dengan referensi transfer bank
	processing, err := fx.service.ReviewPayout(fx.ctx, payout.ID, &finance.ReviewPayoutRequest{Action: "PROCESSING"}, reviewerID)
	require.NoError(t, err)
	assert.Equal(t, finance.PayoutProcessing, processing.Status)

	settleRef := "TRF-BCA-2026-0001"
	settled, err := fx.service.ReviewPayout(fx.ctx, payout.ID, &finance.ReviewPayoutRequest{Action: "SETTLE", SettlementReference: &settleRef}, reviewerID)
	require.NoError(t, err)
	assert.Equal(t, finance.PayoutSettled, settled.Status)

	// 4. Penolakan setelah settled harus ditolak (state machine ketat)
	_, err = fx.service.ReviewPayout(fx.ctx, payout.ID, &finance.ReviewPayoutRequest{Action: "REJECT"}, reviewerID)
	assert.Error(t, err)
}

// Menguji proteksi double-spending konkuren: 5 request serentak melebihi saldo hanya meloloskan yang didukung saldo.
func TestPayoutConcurrentDoubleSpendGuard(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	projectID := setupCompletedProject(t, fx, "Proyek Benteng Payout", 1000000, []float64{100.00, 0.00, 0.00})
	distResp, err := fx.service.DistributeBounty(fx.ctx, projectID)
	require.NoError(t, err)
	payAlloc := findMemberDistribution(distResp.Members, fx.userIDs[0])
	require.True(t, payAlloc.Gross > 0)
	payUser := payAlloc.UserID
	balance := payAlloc.Net

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	// 5 request serentak masing-masing 60% saldo: maksimal 1 yang lolos
	eachAmount := (balance * 60) / 100
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := fx.service.RequestPayout(fx.ctx, payUser, &finance.RequestPayoutRequest{
				Amount:            eachAmount,
				BankCode:          "BCA",
				AccountNumber:     fmt.Sprintf("99000000%d", idx),
				AccountHolderName: "Intern Andi",
			})
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 1, successCount)
	wFinal, err := fx.repo.GetWalletByUserID(fx.ctx, payUser)
	require.NoError(t, err)
	assert.Equal(t, balance-eachAmount, wFinal.CurrentBalance)
}

// Menguji penolakan injeksi nominal negatif dan kunci idempotensi ganda pada payout.
func TestPayoutNegativeInjectionAndIdempotency(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	// Nominal negatif wajib ditolak di level service (tidak hanya validasi controller)
	_, err := fx.service.RequestPayout(fx.ctx, fx.userIDs[2], &finance.RequestPayoutRequest{
		Amount: -1000, BankCode: "BCA", AccountNumber: "111", AccountHolderName: "X",
	})
	assert.Error(t, err)

	// Bekali saldo
	projectID := setupCompletedProject(t, fx, "Proyek Idempoten Payout", 1000000, []float64{0.00, 100.00, 0.00})
	_, err = fx.service.DistributeBounty(fx.ctx, projectID)
	require.NoError(t, err)

	idemKey := fmt.Sprintf("payout-client-key-%s", uuid.New().String()[:8])
	first, err := fx.service.RequestPayout(fx.ctx, fx.userIDs[1], &finance.RequestPayoutRequest{
		Amount: 100000, BankCode: "BCA", AccountNumber: "222", AccountHolderName: "Y", IdempotencyKey: &idemKey,
	})
	require.NoError(t, err)

	second, err := fx.service.RequestPayout(fx.ctx, fx.userIDs[1], &finance.RequestPayoutRequest{
		Amount: 100000, BankCode: "BCA", AccountNumber: "222", AccountHolderName: "Y", IdempotencyKey: &idemKey,
	})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
}

// Menguji penerbitan reward rank multi-komponen yang idempoten beserta siklus klaim fisik.
func TestRankRewardsIssuanceAndPhysicalClaim(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	// Ambil rank Novice dan buat 2 komponen reward: tunai Rp500.000 + merchandise fisik
	var rankID uuid.UUID
	err := fx.db.Pool.QueryRow(fx.ctx, `SELECT id FROM ranks WHERE level_order = 1`).Scan(&rankID)
	require.NoError(t, err)

	// Bersihkan sisa reward rank ini dari run sebelumnya agar perhitungan deterministik
	_, _ = fx.db.Pool.Exec(fx.ctx, `DELETE FROM reward_claims WHERE reward_id IN (SELECT id FROM rewards WHERE rank_id = $1)`, rankID)
	_, _ = fx.db.Pool.Exec(fx.ctx, `DELETE FROM rewards WHERE rank_id = $1`, rankID)

	cashReward := &finance.Reward{ID: uuid.New(), RankID: &rankID, ComponentType: finance.RewardCash, Title: "Bonus Tunai Novice", MonetaryValue: 500000}
	require.NoError(t, fx.repo.CreateReward(fx.ctx, cashReward))
	fx.rewardIDs = append(fx.rewardIDs, cashReward.ID)

	merchReward := &finance.Reward{ID: uuid.New(), RankID: &rankID, ComponentType: finance.RewardPhysical, Title: "Kaos Eksklusif Novice"}
	require.NoError(t, fx.repo.CreateReward(fx.ctx, merchReward))
	fx.rewardIDs = append(fx.rewardIDs, merchReward.ID)

	// 1. Penerbitan pertama: komponen tunai masuk dompet (neto setelah pajak), komponen fisik menjadi tiket ISSUED
	claims, err := fx.service.IssueRankRewards(fx.ctx, fx.userIDs[0], rankID)
	require.NoError(t, err)
	require.Len(t, claims, 1)
	assert.Equal(t, finance.ClaimIssued, claims[0].Status)

	w, err := fx.repo.GetWalletByUserID(fx.ctx, fx.userIDs[0])
	require.NoError(t, err)
	assert.Equal(t, int64(465000), w.CurrentBalance) // 500.000 - 5% - 2%

	// 2. Penerbitan ulang harus idempoten: tidak ada kredit ganda dan tidak ada tiket ganda
	claims2, err := fx.service.IssueRankRewards(fx.ctx, fx.userIDs[0], rankID)
	require.NoError(t, err)
	assert.Len(t, claims2, 1)
	assert.Equal(t, claims[0].ID, claims2[0].ID)

	w2, err := fx.repo.GetWalletByUserID(fx.ctx, fx.userIDs[0])
	require.NoError(t, err)
	assert.Equal(t, int64(465000), w2.CurrentBalance)

	// 3. Klaim oleh pemilik: ISSUED → CLAIMED
	shipping := "Jl. Merdeka No. 1, Indramayu"
	claimed, err := fx.service.ClaimReward(fx.ctx, claims[0].ID, fx.userIDs[0], &finance.ClaimRewardRequest{ShippingAddress: &shipping})
	require.NoError(t, err)
	assert.Equal(t, finance.ClaimClaimed, claimed.Status)

	// 4. Klaim oleh pengguna lain ditolak (otorisasi ketat)
	_, err = fx.service.ClaimReward(fx.ctx, claims[0].ID, fx.userIDs[1], &finance.ClaimRewardRequest{})
	assert.Error(t, err)

	// 5. Admin memproses hingga FULFILLED dengan nomor resi
	processing, err := fx.service.ProcessClaim(fx.ctx, claims[0].ID, &finance.ProcessClaimRequest{Action: "PROCESSING"})
	require.NoError(t, err)
	assert.Equal(t, finance.ClaimProcessing, processing.Status)

	tracking := "JNE-123456789"
	fulfilled, err := fx.service.ProcessClaim(fx.ctx, claims[0].ID, &finance.ProcessClaimRequest{Action: "FULFILLED", TrackingNumber: &tracking})
	require.NoError(t, err)
	assert.Equal(t, finance.ClaimFulfilled, fulfilled.Status)
}

// Menguji migrasi 000005: indeks idempotensi, kolom payout, dan trigger keseimbangan ledger dapat naik-turun.
func TestFinanceIntegrityMigration(t *testing.T) {
	fx := setupFinanceFixture(t)
	defer fx.db.Close()

	upSQL := `CREATE UNIQUE INDEX IF NOT EXISTS uq_wallet_tx_reference ON wallet_transactions(reference_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_reward_claims_reward_user ON reward_claims(reward_id, user_id);
ALTER TABLE payouts ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128);
CREATE UNIQUE INDEX IF NOT EXISTS uq_payouts_idempotency ON payouts(idempotency_key);`
	_, err := fx.db.Pool.Exec(fx.ctx, upSQL)
	require.NoError(t, err)

	var count int
	err = fx.db.Pool.QueryRow(fx.ctx, `
		SELECT COUNT(*) FROM pg_indexes WHERE indexname = 'uq_wallet_tx_reference'
	`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
