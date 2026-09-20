package finance

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/system"
	"dcisp/backend/internal/shared/eventbus"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo           *Repository
	db             *database.PostgresDB
	auditService   *system.AuditService
	eventBus       *eventbus.EventBus
	contribSource  ContributionSource
	performanceSvc XPMutationProvider
}

// ContributionSource memasok data kontribusi final proyek yang telah disahkan untuk distribusi bounty.
type ContributionSource interface {
	GetDistributionMembers(ctx context.Context, projectID uuid.UUID) (*DistributionInput, error)
}

// DistributionMember adalah satu anggota tim penerima bounty.
type DistributionMember struct {
	UserID   uuid.UUID
	FullName string
	FinalPct float64
	IsLocked bool
	BatchID  uuid.UUID
	HasBatch bool
}

// DistributionInput adalah data proyek siap distribusi bounty.
type DistributionInput struct {
	ProjectID uuid.UUID
	Title     string
	Status    string
	Pool      int64
	Members   []DistributionMember
}

// XPMutationProvider mencatat mutasi XP bonus dari komponen reward non-tunai.
type XPMutationProvider interface {
	RecordBonusXP(ctx context.Context, userID uuid.UUID, points int, referenceEvent string) error
}

// Menginisialisasi instance baru service finance and compensation management.
func NewService(repo *Repository, db *database.PostgresDB, audit *system.AuditService, bus *eventbus.EventBus) *Service {
	s := &Service{
		repo:         repo,
		db:           db,
		auditService: audit,
		eventBus:     bus,
	}
	if bus != nil {
		s.registerEventListeners()
	}
	return s
}

// Menetapkan sumber kontribusi proyek untuk distribusi bounty.
func (s *Service) SetContributionSource(src ContributionSource) {
	s.contribSource = src
}

// Menetapkan penyedia mutasi XP untuk komponen reward BONUS_XP.
func (s *Service) SetXPProvider(p XPMutationProvider) {
	s.performanceSvc = p
}

// Mendaftarkan pendengar event rank promotion untuk penerbitan reward otomatis.
func (s *Service) registerEventListeners() {
	s.eventBus.Subscribe("rank.promoted", func(ctx context.Context, event eventbus.DomainEvent) error {
		userIDStr, _ := event.Payload["user_id"].(string)
		rankIDStr, _ := event.Payload["rank_id"].(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil
		}
		rankID, err := uuid.Parse(rankIDStr)
		if err != nil {
			return nil
		}
		_, _ = s.IssueRankRewards(ctx, userID, rankID)
		return nil
	})
}

// ============================================================================
// 1. DEDUCTION & TAX ENGINE (FR-034, T-062)
// ============================================================================

// Mengevaluasi seluruh aturan deduksi aktif terhadap nilai bruto dan menghasilkan alokasi Gross → Tax → Farewell → Net.
// Seluruh perhitungan memakai integer Rupiah (tanpa float) dengan pembulatan half-up yang deterministik.
func (s *Service) EvaluateDeductions(ctx context.Context, gross int64, onlyWithdrawal bool) (*FlowAllocation, error) {
	if gross < 0 {
		return nil, errors.New("nilai bruto tidak boleh negatif")
	}

	rules, err := s.repo.ListActiveRules(ctx)
	if err != nil {
		return nil, err
	}

	alloc := &FlowAllocation{Gross: gross}
	var taxTotal, farewellTotal, otherTotal int64

	for _, rule := range rules {
		if onlyWithdrawal && rule.Type != RuleWithdrawalTax {
			continue
		}
		if !onlyWithdrawal && rule.Type == RuleWithdrawalTax {
			continue
		}

		var amount int64
		switch rule.RateType {
		case RatePercentage:
			// rate dalam satuan 0.0001 (mis. 5% = 50000): amount = round(gross * rate / 1000000)
			amount = (gross*rule.RateValueUnits + 500000) / 1000000
		case RateFixedAmount:
			// rate dalam satuan 0.0001 Rupiah: bulatkan ke Rupiah penuh
			amount = (rule.RateValueUnits + 5000) / 10000
		default:
			continue
		}

		if rule.MinimumAmount != nil && amount < *rule.MinimumAmount {
			amount = *rule.MinimumAmount
		}
		if rule.MaximumAmount != nil && amount > *rule.MaximumAmount {
			amount = *rule.MaximumAmount
		}
		if amount < 0 {
			amount = 0
		}

		isTax := rule.Type == RuleIncomeTax || rule.Type == RuleProjectTax || rule.Type == RuleGraduationTax || rule.Type == RuleWithdrawalTax
		isFarewell := rule.Type == RuleOtherDeduction

		switch {
		case isTax:
			taxTotal += amount
		case isFarewell:
			farewellTotal += amount
		default:
			otherTotal += amount
		}

		alloc.Deductions = append(alloc.Deductions, DeductionResult{
			RuleID:     rule.ID,
			RuleName:   rule.Name,
			RuleType:   rule.Type,
			Amount:     amount,
			IsTax:      isTax,
			IsFarewell: isFarewell,
		})
	}

	totalDeduction := taxTotal + farewellTotal + otherTotal
	if totalDeduction > gross {
		return nil, fmt.Errorf("total deduksi (%d) melebihi nilai bruto (%d): periksa konfigurasi aturan pajak", totalDeduction, gross)
	}

	alloc.TaxTotal = taxTotal
	alloc.FarewellTotal = farewellTotal
	alloc.OtherTotal = otherTotal
	alloc.Net = gross - totalDeduction
	return alloc, nil
}

// Membuat aturan deduksi baru dengan validasi tipe dan konversi nilai ke satuan presisi integer.
func (s *Service) CreateTaxRule(ctx context.Context, req *CreateTaxRuleRequest) (*DeductionTaxRule, error) {
	ruleType := strings.ToUpper(strings.TrimSpace(req.Type))
	switch ruleType {
	case RuleIncomeTax, RuleProjectTax, RuleGraduationTax, RuleWithdrawalTax, RuleAdminFee, RulePenalty, RuleOtherDeduction:
	default:
		return nil, fmt.Errorf("tipe aturan tidak valid: %s", req.Type)
	}

	rateType := strings.ToUpper(strings.TrimSpace(req.RateType))
	if rateType != RatePercentage && rateType != RateFixedAmount {
		return nil, fmt.Errorf("tipe tarif tidak valid: %s (gunakan PERCENTAGE atau FIXED_AMOUNT)", req.RateType)
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("format tanggal efektif tidak valid (YYYY-MM-DD): %w", err)
	}

	// Konversi nilai tarif float input menjadi satuan integer 0.0001
	rateUnits := int64(req.RateValue*10000 + 0.5)
	if rateUnits < 0 {
		return nil, errors.New("nilai tarif tidak boleh negatif")
	}

	basis := req.CalculationBasis
	if basis == "" {
		basis = "GROSS_AMOUNT"
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule := &DeductionTaxRule{
		ID:               uuid.New(),
		Name:             req.Name,
		Type:             ruleType,
		RateType:         rateType,
		RateValueUnits:   rateUnits,
		CalculationBasis: basis,
		MinimumAmount:    req.MinimumAmount,
		MaximumAmount:    req.MaximumAmount,
		EffectiveDate:    effectiveDate,
		IsActive:         isActive,
	}

	rateText := formatRateUnits(rateUnits)
	if err := s.repo.CreateRule(ctx, rule, rateText); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "CREATE_TAX_RULE",
			ResourceType: "DEDUCTION_TAX_RULE",
			ResourceID:   rule.ID.String(),
			NewState: map[string]interface{}{
				"name":      rule.Name,
				"type":      rule.Type,
				"rate_type": rule.RateType,
			},
		})
	}

	return rule, nil
}

// Memformat satuan integer 0.0001 menjadi representasi desimal untuk penyimpanan NUMERIC.
func formatRateUnits(units int64) string {
	return fmt.Sprintf("%d.%04d", units/10000, units%10000)
}

// ============================================================================
// 2. FLOW RECORDING (T-066 PRIMITIVE)
// ============================================================================

// flowAccounts menentukan akun debit sumber dana untuk suatu alokasi.
type flowAccounts struct {
	debitAccount string
}

// Mencatat satu alokasi Gross → Tax → Farewell → Net ke dompet, ledger, dan kas batch dalam transaksi yang diberikan.
// Seluruh mutasi berhasil atau seluruhnya dibatalkan oleh pemilik transaksi (BRULE-FIN-006).
func (s *Service) recordFlowTx(ctx context.Context, tx pgx.Tx, wallet *Wallet, alloc *FlowAllocation, source, referenceID, debitAccount, referenceTable string, referenceUUID uuid.UUID, narration string, batchID *uuid.UUID, approvedBy *uuid.UUID) (uuid.UUID, error) {
	flowTxID := uuid.New()

	// 1. Mutasi dompet itemized (BRULE-FIN-003)
	walletTx := &WalletTransaction{
		ID:              uuid.New(),
		WalletID:        wallet.ID,
		Source:          source,
		ReferenceID:     referenceID,
		GrossAmount:     alloc.Gross,
		DeductionAmount: alloc.TaxTotal + alloc.FarewellTotal + alloc.OtherTotal,
		NetAmount:       alloc.Net,
		Status:          WalletTxCompleted,
		ApprovedBy:      approvedBy,
	}
	if err := s.repo.CreateWalletTransactionTx(ctx, tx, walletTx); err != nil {
		return uuid.Nil, err
	}

	// 2. Kredit bersih ke dompet
	if err := s.repo.CreditWalletTx(ctx, tx, wallet.ID, alloc.Net); err != nil {
		return uuid.Nil, err
	}

	// 3. Jurnal double-entry berpasangan (BRULE-FIN-004)
	entries := []FinancialLedger{
		{TransactionID: flowTxID, AccountCode: debitAccount, Direction: DirectionDebit, Amount: alloc.Gross, ReferenceTable: referenceTable, ReferenceID: referenceUUID, Narration: narration},
		{TransactionID: flowTxID, AccountCode: AcctMemberLiability, Direction: DirectionCredit, Amount: alloc.Net, ReferenceTable: referenceTable, ReferenceID: referenceUUID, Narration: "Hak kompensasi bersih anggota: " + narration},
	}
	if alloc.TaxTotal > 0 {
		entries = append(entries, FinancialLedger{TransactionID: flowTxID, AccountCode: AcctTaxPayable, Direction: DirectionCredit, Amount: alloc.TaxTotal, ReferenceTable: referenceTable, ReferenceID: referenceUUID, Narration: "Kewajiban pajak atas: " + narration})
	}
	if alloc.FarewellTotal > 0 {
		entries = append(entries, FinancialLedger{TransactionID: flowTxID, AccountCode: AcctBatchFund, Direction: DirectionCredit, Amount: alloc.FarewellTotal, ReferenceTable: referenceTable, ReferenceID: referenceUUID, Narration: "Iuran kas bersama angkatan atas: " + narration})
	}
	if alloc.OtherTotal > 0 {
		entries = append(entries, FinancialLedger{TransactionID: flowTxID, AccountCode: AcctOtherDeductions, Direction: DirectionCredit, Amount: alloc.OtherTotal, ReferenceTable: referenceTable, ReferenceID: referenceUUID, Narration: "Deduksi lain atas: " + narration})
	}
	if err := s.repo.AppendLedgerEntriesTx(ctx, tx, entries); err != nil {
		return uuid.Nil, err
	}

	// 4. Akumulasi kas bersama angkatan (BRULE-FIN-002)
	if alloc.FarewellTotal > 0 && batchID != nil {
		if err := s.repo.AccumulateBatchFundTx(ctx, tx, *batchID, alloc.FarewellTotal); err != nil {
			return uuid.Nil, err
		}
	}

	return flowTxID, nil
}

// ============================================================================
// 3. WALLETS (FR-035, T-063)
// ============================================================================

// Mengambil dompet personal pengguna beserta riwayat mutasi (membuat dompet baru bila belum ada).
func (s *Service) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	w, err := s.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if w != nil {
		return w, nil
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi pembuatan dompet: %w", err)
	}
	defer tx.Rollback(ctx)

	w, err = s.repo.GetOrCreateWalletTx(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal membuat dompet: %w", err)
	}
	return w, nil
}

// Mengambil riwayat mutasi dompet pengguna dengan paginasi.
func (s *Service) ListWalletTransactions(ctx context.Context, userID uuid.UUID, page, limit int) ([]WalletTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	w, err := s.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.ListWalletTransactions(ctx, w.ID, limit, (page-1)*limit)
}

// ============================================================================
// 4. PROJECT BOUNTY DISTRIBUTION (FR-032, T-065)
// ============================================================================

// Mendistribusikan bounty pool proyek ke seluruh anggota tim berdasarkan Final Contribution % secara atomik.
// Total distribusi tepat sama dengan pool; sisa pembulatan dialokasikan deterministik ke anggota dengan kontribusi terbesar.
func (s *Service) DistributeBounty(ctx context.Context, projectID uuid.UUID) (*BountyDistributionResponse, error) {
	if s.contribSource == nil {
		return nil, errors.New("sumber kontribusi proyek belum dikonfigurasi")
	}

	input, err := s.contribSource.GetDistributionMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if input.Status != "COMPLETED" {
		return nil, fmt.Errorf("distribusi bounty mensyaratkan proyek berstatus COMPLETED (status saat ini: %s)", input.Status)
	}
	if len(input.Members) == 0 {
		return nil, errors.New("proyek tidak memiliki anggota tim untuk distribusi")
	}
	for _, m := range input.Members {
		if !m.IsLocked {
			return nil, fmt.Errorf("kontribusi anggota %s belum dikunci supervisor", m.FullName)
		}
	}

	// Urutkan anggota berdasarkan user_id agar penguncian baris selalu berurutan (pencegahan deadlock)
	members := append([]DistributionMember(nil), input.Members...)
	sort.Slice(members, func(i, j int) bool {
		return members[i].UserID.String() < members[j].UserID.String()
	})

	// Hitung gross per anggota dengan floor, lalu alokasikan sisa ke kontribusi terbesar
	type grossCalc struct {
		member DistributionMember
		gross  int64
	}
	calcs := make([]grossCalc, len(members))
	var flooredSum int64
	for i, m := range members {
		pctBP := int64(m.FinalPct*100 + 0.5) // basis points, mis. 40.00% -> 4000
		g := (input.Pool * pctBP) / 10000
		calcs[i] = grossCalc{member: m, gross: g}
		flooredSum += g
	}
	remainder := input.Pool - flooredSum
	if remainder > 0 {
		best := 0
		for i := range calcs {
			if calcs[i].member.FinalPct > calcs[best].member.FinalPct ||
				(calcs[i].member.FinalPct == calcs[best].member.FinalPct && calcs[i].member.UserID.String() < calcs[best].member.UserID.String()) {
				best = i
			}
		}
		calcs[best].gross += remainder
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi distribusi bounty: %w", err)
	}
	defer tx.Rollback(ctx)

	resp := &BountyDistributionResponse{
		ProjectID:    input.ProjectID,
		ProjectTitle: input.Title,
		BountyPool:   input.Pool,
	}

	for _, c := range calcs {
		// Anggota dengan porsi 0% tidak menghasilkan pergerakan dana (tanpa baris jurnal)
		if c.gross == 0 {
			resp.Members = append(resp.Members, MemberDistribution{
				UserID:       c.member.UserID,
				UserFullName: c.member.FullName,
				FinalPct:     c.member.FinalPct,
			})
			continue
		}

		alloc, err := s.EvaluateDeductions(ctx, c.gross, false)
		if err != nil {
			return nil, fmt.Errorf("gagal mengevaluasi deduksi anggota %s: %w", c.member.FullName, err)
		}

		wallet, err := s.repo.GetOrCreateWalletTx(ctx, tx, c.member.UserID)
		if err != nil {
			return nil, err
		}

		var batchID *uuid.UUID
		if c.member.HasBatch {
			b := c.member.BatchID
			batchID = &b
		}

		referenceID := fmt.Sprintf("bounty:%s:%s", input.ProjectID.String(), c.member.UserID.String())
		flowTxID, err := s.recordFlowTx(ctx, tx, wallet, alloc, WalletSourceBounty, referenceID, AcctProjectPool, "projects", input.ProjectID, fmt.Sprintf("Distribusi bounty proyek %s kepada %s", input.Title, c.member.FullName), batchID, nil)
		if err != nil {
			if strings.Contains(err.Error(), "sudah pernah dicatat") {
				resp.Skipped = append(resp.Skipped, c.member.FullName)
				continue
			}
			return nil, err
		}

		resp.Members = append(resp.Members, MemberDistribution{
			UserID:        c.member.UserID,
			UserFullName:  c.member.FullName,
			FinalPct:      c.member.FinalPct,
			Gross:         alloc.Gross,
			TaxTotal:      alloc.TaxTotal,
			FarewellTotal: alloc.FarewellTotal,
			Net:           alloc.Net,
			TransactionID: flowTxID,
		})
		resp.TotalDistributed += alloc.Gross
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit distribusi bounty: %w", err)
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "bounty.distributed",
			AggregateID: input.ProjectID.String(),
			Payload: map[string]interface{}{
				"project_id":        input.ProjectID.String(),
				"total_distributed": resp.TotalDistributed,
				"member_count":      len(resp.Members),
			},
		})
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			EventName:    "DISTRIBUTE_BOUNTY",
			ResourceType: "PROJECT",
			ResourceID:   input.ProjectID.String(),
			NewState: map[string]interface{}{
				"bounty_pool":       resp.BountyPool,
				"total_distributed": resp.TotalDistributed,
			},
		})
	}

	return resp, nil
}

// ============================================================================
// 5. BATCH FUNDS (FR-036, T-068)
// ============================================================================

// Mengambil kas bersama angkatan berdasarkan ID batch.
func (s *Service) GetBatchFund(ctx context.Context, batchID uuid.UUID) (*BatchFund, error) {
	f, err := s.repo.GetBatchFundByBatchID(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, errors.New("kas batch tidak ditemukan")
	}
	return f, nil
}

// Mengambil seluruh daftar kas bersama per angkatan.
func (s *Service) ListBatchFunds(ctx context.Context) ([]BatchFund, error) {
	return s.repo.ListBatchFunds(ctx)
}

// ============================================================================
// 6. PAYOUT ENGINE (FR-038, T-070)
// ============================================================================

// Mengajukan permohonan pencairan dana dengan penahanan saldo atomik (BRULE-FIN-005).
func (s *Service) RequestPayout(ctx context.Context, userID uuid.UUID, req *RequestPayoutRequest) (*Payout, error) {
	if req.Amount <= 0 {
		return nil, errors.New("nominal payout harus lebih besar dari nol")
	}

	// Proteksi klik ganda: kembalikan permohonan existing bila kunci idempotensi sama
	if req.IdempotencyKey != nil && strings.TrimSpace(*req.IdempotencyKey) != "" {
		if existing, err := s.repo.GetPayoutByIdempotencyKey(ctx, strings.TrimSpace(*req.IdempotencyKey)); err == nil && existing != nil {
			return existing, nil
		}
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi payout: %w", err)
	}
	defer tx.Rollback(ctx)

	wallet, err := s.repo.GetOrCreateWalletTx(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	// Tahan saldo secara atomik; gagal bila saldo tidak mencukupi
	if err := s.repo.DebitWalletGuardedTx(ctx, tx, wallet.ID, req.Amount); err != nil {
		return nil, fmt.Errorf("pengajuan payout ditolak: %w", err)
	}

	payout := &Payout{
		ID:                uuid.New(),
		WalletID:          wallet.ID,
		Amount:            req.Amount,
		BankCode:          req.BankCode,
		AccountNumber:     req.AccountNumber,
		AccountHolderName: req.AccountHolderName,
		Status:            PayoutRequested,
		IdempotencyKey:    req.IdempotencyKey,
	}

	if err := s.repo.CreatePayoutTx(ctx, tx, payout); err != nil {
		return nil, err
	}

	// Mutasi penahanan saldo dompet
	holdTx := &WalletTransaction{
		WalletID:        wallet.ID,
		Source:          WalletSourcePayout,
		ReferenceID:     fmt.Sprintf("payout-hold:%s", payout.ID.String()),
		GrossAmount:     0,
		DeductionAmount: req.Amount,
		NetAmount:       -req.Amount,
		Status:          WalletTxCompleted,
	}
	if err := s.repo.CreateWalletTransactionTx(ctx, tx, holdTx); err != nil {
		return nil, err
	}

	// Jurnal penahanan: DEBIT utang anggota, KREDIT utang payout
	holdTxID := uuid.New()
	entries := []FinancialLedger{
		{TransactionID: holdTxID, AccountCode: AcctMemberLiability, Direction: DirectionDebit, Amount: req.Amount, ReferenceTable: "payouts", ReferenceID: payout.ID, Narration: fmt.Sprintf("Penahanan saldo payout %s", payout.ID.String())},
		{TransactionID: holdTxID, AccountCode: AcctPayoutPayable, Direction: DirectionCredit, Amount: req.Amount, ReferenceTable: "payouts", ReferenceID: payout.ID, Narration: fmt.Sprintf("Utang payout %s menunggu persetujuan finance", payout.ID.String())},
	}
	if err := s.repo.AppendLedgerEntriesTx(ctx, tx, entries); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit pengajuan payout: %w", err)
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "payout.requested",
			AggregateID: payout.ID.String(),
			Payload: map[string]interface{}{
				"payout_id": payout.ID.String(),
				"user_id":   userID.String(),
				"amount":    req.Amount,
			},
		})
	}

	return payout, nil
}

// Meninjau permohonan payout oleh finance: APPROVE, REJECT, PROCESSING, SETTLE, FAIL.
func (s *Service) ReviewPayout(ctx context.Context, payoutID uuid.UUID, req *ReviewPayoutRequest, reviewerID uuid.UUID) (*Payout, error) {
	payout, err := s.repo.GetPayoutByID(ctx, payoutID)
	if err != nil {
		return nil, err
	}
	if payout == nil {
		return nil, errors.New("payout tidak ditemukan")
	}

	action := strings.ToUpper(strings.TrimSpace(req.Action))

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi peninjauan payout: %w", err)
	}
	defer tx.Rollback(ctx)

	switch action {
	case "APPROVE":
		if payout.Status != PayoutRequested {
			return nil, fmt.Errorf("payout hanya dapat disetujui dari status REQUESTED (status saat ini: %s)", payout.Status)
		}
		if err := s.repo.UpdatePayoutStatusTx(ctx, tx, payout.ID, PayoutApproved, nil); err != nil {
			return nil, err
		}
		payout.Status = PayoutApproved

	case "REJECT":
		if payout.Status != PayoutRequested && payout.Status != PayoutApproved {
			return nil, fmt.Errorf("payout hanya dapat ditolak dari status REQUESTED atau APPROVED (status saat ini: %s)", payout.Status)
		}
		if err := s.refundPayoutHoldTx(ctx, tx, payout); err != nil {
			return nil, err
		}
		if err := s.repo.UpdatePayoutStatusTx(ctx, tx, payout.ID, PayoutRejected, nil); err != nil {
			return nil, err
		}
		payout.Status = PayoutRejected

	case "PROCESSING":
		if payout.Status != PayoutApproved {
			return nil, fmt.Errorf("payout hanya dapat diproses dari status APPROVED (status saat ini: %s)", payout.Status)
		}
		if err := s.repo.UpdatePayoutStatusTx(ctx, tx, payout.ID, PayoutProcessing, nil); err != nil {
			return nil, err
		}
		payout.Status = PayoutProcessing

	case "SETTLE":
		if payout.Status != PayoutProcessing {
			return nil, fmt.Errorf("payout hanya dapat diselesaikan dari status PROCESSING (status saat ini: %s)", payout.Status)
		}
		if req.SettlementReference == nil || strings.TrimSpace(*req.SettlementReference) == "" {
			return nil, errors.New("referensi settlement transfer bank wajib diisi saat penyelesaian payout")
		}
		settleTxID := uuid.New()
		entries := []FinancialLedger{
			{TransactionID: settleTxID, AccountCode: AcctPayoutPayable, Direction: DirectionDebit, Amount: payout.Amount, ReferenceTable: "payouts", ReferenceID: payout.ID, Narration: fmt.Sprintf("Pelunasan utang payout %s", payout.ID.String())},
			{TransactionID: settleTxID, AccountCode: AcctCash, Direction: DirectionCredit, Amount: payout.Amount, ReferenceTable: "payouts", ReferenceID: payout.ID, Narration: fmt.Sprintf("Transfer bank payout %s ref %s", payout.ID.String(), strings.TrimSpace(*req.SettlementReference))},
		}
		if err := s.repo.AppendLedgerEntriesTx(ctx, tx, entries); err != nil {
			return nil, err
		}
		if err := s.repo.UpdatePayoutStatusTx(ctx, tx, payout.ID, PayoutSettled, req.SettlementReference); err != nil {
			return nil, err
		}
		payout.Status = PayoutSettled
		payout.SettlementReference = req.SettlementReference

	case "FAIL":
		if payout.Status != PayoutProcessing {
			return nil, fmt.Errorf("payout hanya dapat dinyatakan gagal dari status PROCESSING (status saat ini: %s)", payout.Status)
		}
		if err := s.refundPayoutHoldTx(ctx, tx, payout); err != nil {
			return nil, err
		}
		if err := s.repo.UpdatePayoutStatusTx(ctx, tx, payout.ID, PayoutFailed, nil); err != nil {
			return nil, err
		}
		payout.Status = PayoutFailed

	default:
		return nil, errors.New("aksi peninjauan tidak valid (gunakan APPROVE, REJECT, PROCESSING, SETTLE, atau FAIL)")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit peninjauan payout: %w", err)
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, system.MutationAuditEntry{
			UserID:       &reviewerID,
			EventName:    "REVIEW_PAYOUT_" + action,
			ResourceType: "PAYOUT",
			ResourceID:   payout.ID.String(),
			NewState: map[string]interface{}{
				"status": payout.Status,
				"amount": payout.Amount,
			},
		})
	}

	if s.eventBus != nil && (action == "SETTLE" || action == "FAIL") {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "payout.completed",
			AggregateID: payout.ID.String(),
			Payload: map[string]interface{}{
				"payout_id": payout.ID.String(),
				"status":    payout.Status,
				"amount":    payout.Amount,
			},
		})
	}

	return payout, nil
}

// Mengembalikan dana yang ditahan ke dompet pengguna beserta jurnal pembaliknya.
func (s *Service) refundPayoutHoldTx(ctx context.Context, tx pgx.Tx, payout *Payout) error {
	var wallet Wallet
	err := tx.QueryRow(ctx, `SELECT id, user_id, current_balance::bigint FROM wallets WHERE id = $1`, payout.WalletID).Scan(&wallet.ID, &wallet.UserID, &wallet.CurrentBalance)
	if err != nil {
		return fmt.Errorf("gagal mencari dompet payout: %w", err)
	}

	if err := s.repo.CreditWalletTx(ctx, tx, wallet.ID, payout.Amount); err != nil {
		return err
	}

	refundTx := &WalletTransaction{
		WalletID:        wallet.ID,
		Source:          WalletSourcePayout,
		ReferenceID:     fmt.Sprintf("payout-refund:%s", payout.ID.String()),
		GrossAmount:     payout.Amount,
		DeductionAmount: 0,
		NetAmount:       payout.Amount,
		Status:          WalletTxCompleted,
	}
	if err := s.repo.CreateWalletTransactionTx(ctx, tx, refundTx); err != nil {
		return err
	}

	refundLedgerID := uuid.New()
	entries := []FinancialLedger{
		{TransactionID: refundLedgerID, AccountCode: AcctPayoutPayable, Direction: DirectionDebit, Amount: payout.Amount, ReferenceTable: "payouts", ReferenceID: payout.ID, Narration: fmt.Sprintf("Pembatalan utang payout %s", payout.ID.String())},
		{TransactionID: refundLedgerID, AccountCode: AcctMemberLiability, Direction: DirectionCredit, Amount: payout.Amount, ReferenceTable: "payouts", ReferenceID: payout.ID, Narration: fmt.Sprintf("Pengembalian penahanan payout %s ke dompet", payout.ID.String())},
	}
	return s.repo.AppendLedgerEntriesTx(ctx, tx, entries)
}

// ============================================================================
// 7. RANK REWARDS & CLAIMS (FR-031, FR-033, T-067, T-069)
// ============================================================================

// Menerbitkan seluruh paket reward atas kenaikan rank pengguna secara idempoten.
func (s *Service) IssueRankRewards(ctx context.Context, userID, rankID uuid.UUID) ([]RewardClaim, error) {
	rewards, err := s.repo.ListRewards(ctx, &rankID, nil)
	if err != nil {
		return nil, err
	}
	if len(rewards) == 0 {
		return nil, nil
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi penerbitan reward: %w", err)
	}
	defer tx.Rollback(ctx)

	var claims []RewardClaim
	for _, rw := range rewards {
		switch rw.ComponentType {
		case RewardCash:
			if rw.MonetaryValue <= 0 {
				continue
			}
			alloc, err := s.EvaluateDeductions(ctx, rw.MonetaryValue, false)
			if err != nil {
				return nil, fmt.Errorf("gagal mengevaluasi deduksi reward %s: %w", rw.Title, err)
			}
			wallet, err := s.repo.GetOrCreateWalletTx(ctx, tx, userID)
			if err != nil {
				return nil, err
			}
			referenceID := fmt.Sprintf("rank-reward:%s:%s", userID.String(), rw.ID.String())
			if _, err := s.recordFlowTx(ctx, tx, wallet, alloc, WalletSourceReward, referenceID, AcctRewardExpense, "rewards", rw.ID, fmt.Sprintf("Reward tunai kenaikan rank: %s", rw.Title), nil, nil); err != nil {
				if strings.Contains(err.Error(), "sudah pernah dicatat") {
					continue
				}
				return nil, err
			}

		case RewardBonusXP:
			if s.performanceSvc != nil && rw.MonetaryValue > 0 {
				points := int(rw.MonetaryValue)
				ref := fmt.Sprintf("rank-reward-xp:%s:%s", userID.String(), rw.ID.String())
				_ = s.performanceSvc.RecordBonusXP(ctx, userID, points, ref)
			}

		case RewardPhysical, RewardBadge, RewardVoucher, RewardPrivilege:
			claim, _, err := s.repo.IssueRewardClaimTx(ctx, tx, rw.ID, userID)
			if err != nil {
				return nil, err
			}
			claims = append(claims, *claim)

		default:
			continue
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit penerbitan reward: %w", err)
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "rank.reward_issued",
			AggregateID: userID.String(),
			Payload: map[string]interface{}{
				"user_id":     userID.String(),
				"rank_id":     rankID.String(),
				"claim_count": len(claims),
			},
		})
	}

	return claims, nil
}

// Mengajukan klaim atas tiket reward yang berstatus ISSUED oleh pemiliknya sendiri.
func (s *Service) ClaimReward(ctx context.Context, claimID, userID uuid.UUID, req *ClaimRewardRequest) (*RewardClaim, error) {
	claim, err := s.repo.GetClaimByID(ctx, claimID)
	if err != nil {
		return nil, err
	}
	if claim == nil {
		return nil, errors.New("klaim reward tidak ditemukan")
	}
	if claim.UserID != userID {
		return nil, errors.New("akses ditolak: klaim reward bukan milik anda")
	}
	if claim.Status != ClaimIssued {
		return nil, fmt.Errorf("klaim hanya dapat diajukan dari status ISSUED (status saat ini: %s)", claim.Status)
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi klaim: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.UpdateClaimStatusTx(ctx, tx, claimID, ClaimClaimed, req.ShippingAddress, nil); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit klaim reward: %w", err)
	}

	claim.Status = ClaimClaimed
	claim.ShippingAddress = req.ShippingAddress
	return claim, nil
}

// Memproses tiket klaim reward oleh admin: PROCESSING, FULFILLED (dengan resi), atau REJECTED.
func (s *Service) ProcessClaim(ctx context.Context, claimID uuid.UUID, req *ProcessClaimRequest) (*RewardClaim, error) {
	claim, err := s.repo.GetClaimByID(ctx, claimID)
	if err != nil {
		return nil, err
	}
	if claim == nil {
		return nil, errors.New("klaim reward tidak ditemukan")
	}

	action := strings.ToUpper(strings.TrimSpace(req.Action))
	var targetStatus string
	switch action {
	case ClaimProcessing:
		if claim.Status != ClaimIssued && claim.Status != ClaimClaimed {
			return nil, fmt.Errorf("klaim hanya dapat diproses dari status ISSUED atau CLAIMED (status saat ini: %s)", claim.Status)
		}
		targetStatus = ClaimProcessing
	case ClaimFulfilled:
		if claim.Status != ClaimProcessing && claim.Status != ClaimClaimed {
			return nil, fmt.Errorf("klaim hanya dapat diselesaikan dari status PROCESSING atau CLAIMED (status saat ini: %s)", claim.Status)
		}
		targetStatus = ClaimFulfilled
	case ClaimRejected:
		if claim.Status == ClaimFulfilled {
			return nil, errors.New("klaim yang sudah FULFILLED tidak dapat ditolak")
		}
		targetStatus = ClaimRejected
	default:
		return nil, errors.New("aksi pemrosesan tidak valid (gunakan PROCESSING, FULFILLED, atau REJECTED)")
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi pemrosesan klaim: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.UpdateClaimStatusTx(ctx, tx, claimID, targetStatus, nil, req.TrackingNumber); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal melakukan commit pemrosesan klaim: %w", err)
	}

	claim.Status = targetStatus
	claim.TrackingNumber = req.TrackingNumber
	return claim, nil
}

// ============================================================================
// 8. LEDGER REVERSAL (BRULE-FIN-004)
// ============================================================================

// Membuat jurnal pembalik (reversal) atas suatu transaksi ledger tanpa mengubah baris asli.
func (s *Service) ReverseLedgerTransaction(ctx context.Context, transactionID uuid.UUID, reason string) (uuid.UUID, error) {
	entries, err := s.repo.GetLedgerEntriesByTransaction(ctx, transactionID)
	if err != nil {
		return uuid.Nil, err
	}
	if len(entries) == 0 {
		return uuid.Nil, errors.New("transaksi ledger tidak ditemukan")
	}

	reversalID := uuid.New()
	mirror := make([]FinancialLedger, len(entries))
	for i, e := range entries {
		direction := DirectionDebit
		if e.Direction == DirectionDebit {
			direction = DirectionCredit
		}
		mirror[i] = FinancialLedger{
			TransactionID:  reversalID,
			AccountCode:    e.AccountCode,
			Direction:      direction,
			Amount:         e.Amount,
			ReferenceTable: e.ReferenceTable,
			ReferenceID:    e.ReferenceID,
			Narration:      "REVERSAL atas " + transactionID.String() + ": " + reason,
		}
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("gagal memulai transaksi reversal: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.AppendLedgerEntriesTx(ctx, tx, mirror); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("gagal melakukan commit reversal: %w", err)
	}
	return reversalID, nil
}
