package finance

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.PostgresDB
}

// Menginisialisasi instance baru repository finance and compensation management.
func NewRepository(db *database.PostgresDB) *Repository {
	return &Repository{db: db}
}

// Mengurai nilai desimal NUMERIC menjadi satuan integer 0.0001 untuk perhitungan presisi tanpa float.
func parseRateUnits(text string) (int64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, errors.New("nilai tarif kosong")
	}
	negative := false
	if strings.HasPrefix(text, "-") {
		negative = true
		text = text[1:]
	}
	parts := strings.SplitN(text, ".", 2)
	intPart, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("format tarif tidak valid: %s", text)
	}
	var fracPart int64
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 4 {
			frac = frac[:4]
		}
		for len(frac) < 4 {
			frac += "0"
		}
		fracPart, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("format desimal tarif tidak valid: %s", text)
		}
	}
	units := intPart*10000 + fracPart
	if negative {
		units = -units
	}
	return units, nil
}

// ============================================================================
// 1. WALLETS (FR-035, T-063)
// ============================================================================

// Mengambil atau membuat dompet personal pengguna secara atomik dalam transaksi.
func (r *Repository) GetOrCreateWalletTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*Wallet, error) {
	var w Wallet
	err := tx.QueryRow(ctx, `
		SELECT id, user_id, current_balance::bigint, created_at, updated_at
		FROM wallets
		WHERE user_id = $1
	`, userID).Scan(&w.ID, &w.UserID, &w.CurrentBalance, &w.CreatedAt, &w.UpdatedAt)
	if err == nil {
		return &w, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("gagal mencari dompet pengguna: %w", err)
	}

	w.ID = uuid.New()
	w.UserID = userID
	w.CurrentBalance = 0
	err = tx.QueryRow(ctx, `
		INSERT INTO wallets (id, user_id, current_balance, created_at, updated_at)
		VALUES ($1, $2, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, current_balance::bigint, created_at, updated_at
	`, w.ID, userID).Scan(&w.ID, &w.UserID, &w.CurrentBalance, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat dompet pengguna: %w", err)
	}
	return &w, nil
}

// Mengambil dompet personal berdasarkan ID pengguna.
func (r *Repository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	var w Wallet
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, current_balance::bigint, created_at, updated_at
		FROM wallets
		WHERE user_id = $1
	`, userID).Scan(&w.ID, &w.UserID, &w.CurrentBalance, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari dompet pengguna: %w", err)
	}
	return &w, nil
}

// Menambah saldo dompet secara atomik dalam transaksi.
func (r *Repository) CreditWalletTx(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, amount int64) error {
	if amount < 0 {
		return errors.New("nominal kredit dompet tidak boleh negatif")
	}
	cmdTag, err := tx.Exec(ctx, `
		UPDATE wallets
		SET current_balance = current_balance + $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, walletID, amount)
	if err != nil {
		return fmt.Errorf("gagal mengkredit dompet: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("dompet tidak ditemukan")
	}
	return nil
}

// Mengurangi saldo dompet secara atomik dengan penjagaan saldo tidak boleh minus (BRULE-FIN-005).
func (r *Repository) DebitWalletGuardedTx(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return errors.New("nominal debit dompet harus lebih besar dari nol")
	}
	cmdTag, err := tx.Exec(ctx, `
		UPDATE wallets
		SET current_balance = current_balance - $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND current_balance >= $2
	`, walletID, amount)
	if err != nil {
		return fmt.Errorf("gagal mendebit dompet: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("saldo dompet tidak mencukupi atau dompet tidak ditemukan")
	}
	return nil
}

// Mencatat mutasi itemized dompet dengan kunci idempotensi (BRULE-FIN-003).
func (r *Repository) CreateWalletTransactionTx(ctx context.Context, tx pgx.Tx, t *WalletTransaction) error {
	if t.NetAmount != t.GrossAmount-t.DeductionAmount {
		return fmt.Errorf("mutasi dompet tidak konsisten: net (%d) harus sama dengan gross (%d) dikurangi deduksi (%d)", t.NetAmount, t.GrossAmount, t.DeductionAmount)
	}
	query := `
		INSERT INTO wallet_transactions (id, wallet_id, source, reference_id, gross_amount, deduction_amount, net_amount, status, created_at, approved_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, $9)
		ON CONFLICT (reference_id) DO NOTHING
	`
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	cmdTag, err := tx.Exec(ctx, query, t.ID, t.WalletID, t.Source, t.ReferenceID, t.GrossAmount, t.DeductionAmount, t.NetAmount, t.Status, t.ApprovedBy)
	if err != nil {
		return fmt.Errorf("gagal mencatat mutasi dompet: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("mutasi dengan reference_id yang sama sudah pernah dicatat (idempoten)")
	}
	return nil
}

// Mengambil riwayat mutasi dompet dengan paginasi.
func (r *Repository) ListWalletTransactions(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]WalletTransaction, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM wallet_transactions WHERE wallet_id = $1`, walletID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung mutasi dompet: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, wallet_id, source, reference_id, gross_amount::bigint, deduction_amount::bigint, net_amount::bigint, status, created_at, approved_by
		FROM wallet_transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, walletID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil riwayat mutasi dompet: %w", err)
	}
	defer rows.Close()

	var list []WalletTransaction
	for rows.Next() {
		var t WalletTransaction
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Source, &t.ReferenceID, &t.GrossAmount, &t.DeductionAmount, &t.NetAmount, &t.Status, &t.CreatedAt, &t.ApprovedBy); err == nil {
			list = append(list, t)
		}
	}
	return list, total, nil
}

// ============================================================================
// 2. FINANCIAL LEDGER (FR-037, T-064)
// ============================================================================

// Menulis sekumpulan baris jurnal berpasangan dalam satu transaksi atomik (BRULE-FIN-004).
func (r *Repository) AppendLedgerEntriesTx(ctx context.Context, tx pgx.Tx, entries []FinancialLedger) error {
	if len(entries) < 2 {
		return errors.New("jurnal double-entry wajib memiliki minimal dua baris berpasangan")
	}
	var debitSum, creditSum int64
	txID := entries[0].TransactionID
	for _, e := range entries {
		if e.TransactionID != txID {
			return errors.New("seluruh baris jurnal harus berbagi transaction_id yang sama")
		}
		if e.Amount <= 0 {
			return errors.New("nominal jurnal harus lebih besar dari nol")
		}
		switch e.Direction {
		case DirectionDebit:
			debitSum += e.Amount
		case DirectionCredit:
			creditSum += e.Amount
		default:
			return fmt.Errorf("arah jurnal tidak valid: %s", e.Direction)
		}
	}
	if debitSum != creditSum {
		return fmt.Errorf("jurnal tidak seimbang: total debit (%d) tidak sama dengan total kredit (%d)", debitSum, creditSum)
	}

	for i := range entries {
		e := &entries[i]
		if e.ID == uuid.Nil {
			e.ID = uuid.New()
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO financial_ledgers (id, transaction_id, account_code, direction, amount, reference_table, reference_id, narration, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
		`, e.ID, e.TransactionID, e.AccountCode, e.Direction, e.Amount, e.ReferenceTable, e.ReferenceID, e.Narration)
		if err != nil {
			return fmt.Errorf("gagal menulis baris jurnal: %w", err)
		}
	}
	return nil
}

// Mengambil seluruh baris jurnal berdasarkan ID transaksi untuk verifikasi keseimbangan.
func (r *Repository) GetLedgerEntriesByTransaction(ctx context.Context, txID uuid.UUID) ([]FinancialLedger, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, transaction_id, account_code, direction, amount::bigint, reference_table, reference_id, narration, created_at
		FROM financial_ledgers
		WHERE transaction_id = $1
		ORDER BY created_at ASC
	`, txID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil jurnal transaksi: %w", err)
	}
	defer rows.Close()

	var list []FinancialLedger
	for rows.Next() {
		var e FinancialLedger
		if err := rows.Scan(&e.ID, &e.TransactionID, &e.AccountCode, &e.Direction, &e.Amount, &e.ReferenceTable, &e.ReferenceID, &e.Narration, &e.CreatedAt); err == nil {
			list = append(list, e)
		}
	}
	return list, nil
}

// Mengambil daftar jurnal buku besar dengan filter akun dan paginasi untuk audit finansial.
func (r *Repository) ListLedgerEntries(ctx context.Context, accountCode string, limit, offset int) ([]FinancialLedger, int, error) {
	baseQuery := `FROM financial_ledgers WHERE 1=1`
	var args []interface{}
	if accountCode != "" {
		baseQuery += " AND account_code = $1"
		args = append(args, accountCode)
	}

	var total int
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) "+baseQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung jurnal: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, transaction_id, account_code, direction, amount::bigint, reference_table, reference_id, narration, created_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar jurnal: %w", err)
	}
	defer rows.Close()

	var list []FinancialLedger
	for rows.Next() {
		var e FinancialLedger
		if err := rows.Scan(&e.ID, &e.TransactionID, &e.AccountCode, &e.Direction, &e.Amount, &e.ReferenceTable, &e.ReferenceID, &e.Narration, &e.CreatedAt); err == nil {
			list = append(list, e)
		}
	}
	return list, total, nil
}

// ============================================================================
// 3. DEDUCTION & TAX RULES (FR-034, T-062)
// ============================================================================

// Mengambil seluruh aturan deduksi aktif yang berlaku (BRULE-FIN-001).
func (r *Repository) ListActiveRules(ctx context.Context) ([]DeductionTaxRule, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, type, rate_type, rate_value::text, calculation_basis, minimum_amount::bigint, maximum_amount::bigint, effective_date, is_active, created_at, updated_at
		FROM deduction_tax_rules
		WHERE is_active = TRUE
		ORDER BY effective_date ASC, created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil aturan deduksi aktif: %w", err)
	}
	defer rows.Close()

	var list []DeductionTaxRule
	for rows.Next() {
		var rule DeductionTaxRule
		var rateText string
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Type, &rule.RateType, &rateText, &rule.CalculationBasis, &rule.MinimumAmount, &rule.MaximumAmount, &rule.EffectiveDate, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("gagal memindai aturan deduksi: %w", err)
		}
		units, err := parseRateUnits(rateText)
		if err != nil {
			return nil, err
		}
		rule.RateValueUnits = units
		list = append(list, rule)
	}
	return list, nil
}

// Mengambil seluruh aturan deduksi termasuk yang non-aktif untuk administrasi finance.
func (r *Repository) ListAllRules(ctx context.Context) ([]DeductionTaxRule, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, type, rate_type, rate_value::text, calculation_basis, minimum_amount::bigint, maximum_amount::bigint, effective_date, is_active, created_at, updated_at
		FROM deduction_tax_rules
		ORDER BY effective_date DESC, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar aturan deduksi: %w", err)
	}
	defer rows.Close()

	var list []DeductionTaxRule
	for rows.Next() {
		var rule DeductionTaxRule
		var rateText string
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Type, &rule.RateType, &rateText, &rule.CalculationBasis, &rule.MinimumAmount, &rule.MaximumAmount, &rule.EffectiveDate, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("gagal memindai aturan deduksi: %w", err)
		}
		units, err := parseRateUnits(rateText)
		if err != nil {
			return nil, err
		}
		rule.RateValueUnits = units
		list = append(list, rule)
	}
	return list, nil
}

// Menyimpan aturan deduksi baru ke dalam database.
func (r *Repository) CreateRule(ctx context.Context, rule *DeductionTaxRule, rateValueText string) error {
	query := `
		INSERT INTO deduction_tax_rules (id, name, type, rate_type, rate_value, calculation_basis, minimum_amount, maximum_amount, effective_date, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::numeric, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, rule.ID, rule.Name, rule.Type, rule.RateType, rateValueText, rule.CalculationBasis, rule.MinimumAmount, rule.MaximumAmount, rule.EffectiveDate, rule.IsActive)
	if err != nil {
		return fmt.Errorf("gagal menyimpan aturan deduksi: %w", err)
	}
	return nil
}

// Mengambil aturan deduksi berdasarkan ID.
func (r *Repository) GetRuleByID(ctx context.Context, id uuid.UUID) (*DeductionTaxRule, error) {
	var rule DeductionTaxRule
	var rateText string
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, type, rate_type, rate_value::text, calculation_basis, minimum_amount::bigint, maximum_amount::bigint, effective_date, is_active, created_at, updated_at
		FROM deduction_tax_rules
		WHERE id = $1
	`, id).Scan(&rule.ID, &rule.Name, &rule.Type, &rule.RateType, &rateText, &rule.CalculationBasis, &rule.MinimumAmount, &rule.MaximumAmount, &rule.EffectiveDate, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari aturan deduksi: %w", err)
	}
	units, err := parseRateUnits(rateText)
	if err != nil {
		return nil, err
	}
	rule.RateValueUnits = units
	return &rule, nil
}

// ============================================================================
// 4. BATCH FUNDS (FR-036, T-068)
// ============================================================================

// Mengambil kas bersama angkatan berdasarkan ID batch.
func (r *Repository) GetBatchFundByBatchID(ctx context.Context, batchID uuid.UUID) (*BatchFund, error) {
	var f BatchFund
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, batch_id, total_accumulated::bigint, current_balance::bigint, created_at, updated_at
		FROM batch_funds
		WHERE batch_id = $1
	`, batchID).Scan(&f.ID, &f.BatchID, &f.TotalAccumulated, &f.CurrentBalance, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari kas batch: %w", err)
	}
	return &f, nil
}

// Menambah akumulasi dan saldo kas bersama angkatan secara atomik dalam transaksi.
func (r *Repository) AccumulateBatchFundTx(ctx context.Context, tx pgx.Tx, batchID uuid.UUID, amount int64) error {
	if amount < 0 {
		return errors.New("nominal akumulasi kas batch tidak boleh negatif")
	}
	if amount == 0 {
		return nil
	}
	cmdTag, err := tx.Exec(ctx, `
		UPDATE batch_funds
		SET total_accumulated = total_accumulated + $2, current_balance = current_balance + $2, updated_at = CURRENT_TIMESTAMP
		WHERE batch_id = $1
	`, batchID, amount)
	if err != nil {
		return fmt.Errorf("gagal mengakumulasi kas batch: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("kas batch tidak ditemukan")
	}
	return nil
}

// Mengambil seluruh daftar kas bersama per angkatan.
func (r *Repository) ListBatchFunds(ctx context.Context) ([]BatchFund, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, batch_id, total_accumulated::bigint, current_balance::bigint, created_at, updated_at
		FROM batch_funds
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar kas batch: %w", err)
	}
	defer rows.Close()

	var list []BatchFund
	for rows.Next() {
		var f BatchFund
		if err := rows.Scan(&f.ID, &f.BatchID, &f.TotalAccumulated, &f.CurrentBalance, &f.CreatedAt, &f.UpdatedAt); err == nil {
			list = append(list, f)
		}
	}
	return list, nil
}

// ============================================================================
// 5. PAYOUTS (FR-038, T-070)
// ============================================================================

// Membuat permohonan pencairan dana baru dalam transaksi.
func (r *Repository) CreatePayoutTx(ctx context.Context, tx pgx.Tx, p *Payout) error {
	query := `
		INSERT INTO payouts (id, wallet_id, amount, bank_code, account_number, account_holder_name, status, settlement_reference, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (idempotency_key) DO NOTHING
	`
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	cmdTag, err := tx.Exec(ctx, query, p.ID, p.WalletID, p.Amount, p.BankCode, p.AccountNumber, p.AccountHolderName, p.Status, p.SettlementReference, p.IdempotencyKey)
	if err != nil {
		return fmt.Errorf("gagal membuat permohonan payout: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("permohonan payout dengan kunci idempotensi yang sama sudah pernah dibuat")
	}
	return nil
}

// Mengambil permohonan payout berdasarkan ID beserta identitas pemilik dompet.
func (r *Repository) GetPayoutByID(ctx context.Context, id uuid.UUID) (*Payout, error) {
	var p Payout
	err := r.db.Pool.QueryRow(ctx, `
		SELECT p.id, p.wallet_id, p.amount::bigint, p.bank_code, p.account_number, p.account_holder_name, p.status, p.settlement_reference, p.idempotency_key, p.created_at, p.updated_at,
		       w.user_id, u.full_name
		FROM payouts p
		JOIN wallets w ON w.id = p.wallet_id
		JOIN users u ON u.id = w.user_id
		WHERE p.id = $1
	`, id).Scan(
		&p.ID, &p.WalletID, &p.Amount, &p.BankCode, &p.AccountNumber, &p.AccountHolderName, &p.Status, &p.SettlementReference, &p.IdempotencyKey, &p.CreatedAt, &p.UpdatedAt,
		&p.UserID, &p.UserFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari payout: %w", err)
	}
	return &p, nil
}

// Mengambil permohonan payout berdasarkan kunci idempotensi untuk proteksi klik ganda.
func (r *Repository) GetPayoutByIdempotencyKey(ctx context.Context, key string) (*Payout, error) {
	var p Payout
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, wallet_id, amount::bigint, bank_code, account_number, account_holder_name, status, settlement_reference, idempotency_key, created_at, updated_at
		FROM payouts
		WHERE idempotency_key = $1
	`, key).Scan(
		&p.ID, &p.WalletID, &p.Amount, &p.BankCode, &p.AccountNumber, &p.AccountHolderName, &p.Status, &p.SettlementReference, &p.IdempotencyKey, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari payout berdasarkan kunci idempotensi: %w", err)
	}
	return &p, nil
}

// Memperbarui status permohonan payout dalam transaksi.
func (r *Repository) UpdatePayoutStatusTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string, settlementRef *string) error {
	cmdTag, err := tx.Exec(ctx, `
		UPDATE payouts
		SET status = $2, settlement_reference = COALESCE($3, settlement_reference), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, id, status, settlementRef)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status payout: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("payout tidak ditemukan")
	}
	return nil
}

// Mengambil daftar payout milik sebuah dompet dengan paginasi.
func (r *Repository) ListPayoutsByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]Payout, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM payouts WHERE wallet_id = $1`, walletID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung payout: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, wallet_id, amount::bigint, bank_code, account_number, account_holder_name, status, settlement_reference, idempotency_key, created_at, updated_at
		FROM payouts
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, walletID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar payout: %w", err)
	}
	defer rows.Close()

	var list []Payout
	for rows.Next() {
		var p Payout
		if err := rows.Scan(&p.ID, &p.WalletID, &p.Amount, &p.BankCode, &p.AccountNumber, &p.AccountHolderName, &p.Status, &p.SettlementReference, &p.IdempotencyKey, &p.CreatedAt, &p.UpdatedAt); err == nil {
			list = append(list, p)
		}
	}
	return list, total, nil
}

// Mengambil seluruh antrean payout untuk peninjauan finance dengan filter status opsional.
func (r *Repository) ListAllPayouts(ctx context.Context, status string, limit, offset int) ([]Payout, int, error) {
	baseQuery := `
		FROM payouts p
		JOIN wallets w ON w.id = p.wallet_id
		JOIN users u ON u.id = w.user_id
		WHERE 1=1
	`
	var args []interface{}
	if status != "" {
		baseQuery += " AND p.status = $1"
		args = append(args, status)
	}

	var total int
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) "+baseQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung antrean payout: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.wallet_id, p.amount::bigint, p.bank_code, p.account_number, p.account_holder_name, p.status, p.settlement_reference, p.idempotency_key, p.created_at, p.updated_at,
		       w.user_id, u.full_name
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil antrean payout: %w", err)
	}
	defer rows.Close()

	var list []Payout
	for rows.Next() {
		var p Payout
		if err := rows.Scan(
			&p.ID, &p.WalletID, &p.Amount, &p.BankCode, &p.AccountNumber, &p.AccountHolderName, &p.Status, &p.SettlementReference, &p.IdempotencyKey, &p.CreatedAt, &p.UpdatedAt,
			&p.UserID, &p.UserFullName,
		); err == nil {
			list = append(list, p)
		}
	}
	return list, total, nil
}

// ============================================================================
// 6. REWARDS & CLAIMS (FR-031, FR-033, T-067, T-069)
// ============================================================================

// Menyimpan master hadiah baru ke dalam katalog rewards.
func (r *Repository) CreateReward(ctx context.Context, reward *Reward) error {
	query := `
		INSERT INTO rewards (id, rank_id, achievement_id, component_type, title, monetary_value, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
	`
	if reward.ID == uuid.Nil {
		reward.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, reward.ID, reward.RankID, reward.AchievementID, reward.ComponentType, reward.Title, reward.MonetaryValue, reward.Description)
	if err != nil {
		return fmt.Errorf("gagal menyimpan master reward: %w", err)
	}
	return nil
}

// Mengambil daftar katalog reward dengan filter rank atau achievement opsional.
func (r *Repository) ListRewards(ctx context.Context, rankID, achievementID *uuid.UUID) ([]Reward, error) {
	baseQuery := `
		SELECT r.id, r.rank_id, r.achievement_id, r.component_type, r.title, r.monetary_value::bigint, r.description, r.created_at,
		       rk.name
		FROM rewards r
		LEFT JOIN ranks rk ON rk.id = r.rank_id
		WHERE 1=1
	`
	var args []interface{}
	if rankID != nil {
		baseQuery += fmt.Sprintf(" AND r.rank_id = $%d", len(args)+1)
		args = append(args, *rankID)
	}
	if achievementID != nil {
		baseQuery += fmt.Sprintf(" AND r.achievement_id = $%d", len(args)+1)
		args = append(args, *achievementID)
	}
	baseQuery += " ORDER BY r.created_at ASC"

	rows, err := r.db.Pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil katalog reward: %w", err)
	}
	defer rows.Close()

	var list []Reward
	for rows.Next() {
		var rw Reward
		if err := rows.Scan(&rw.ID, &rw.RankID, &rw.AchievementID, &rw.ComponentType, &rw.Title, &rw.MonetaryValue, &rw.Description, &rw.CreatedAt, &rw.RankName); err == nil {
			list = append(list, rw)
		}
	}
	return list, nil
}

// Mengambil master reward berdasarkan ID.
func (r *Repository) GetRewardByID(ctx context.Context, id uuid.UUID) (*Reward, error) {
	var rw Reward
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, rank_id, achievement_id, component_type, title, monetary_value::bigint, description, created_at
		FROM rewards
		WHERE id = $1
	`, id).Scan(&rw.ID, &rw.RankID, &rw.AchievementID, &rw.ComponentType, &rw.Title, &rw.MonetaryValue, &rw.Description, &rw.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari reward: %w", err)
	}
	return &rw, nil
}

// Menerbitkan tiket klaim hadiah dengan status ISSUED secara idempoten per pasangan reward dan pengguna.
func (r *Repository) IssueRewardClaimTx(ctx context.Context, tx pgx.Tx, rewardID, userID uuid.UUID) (*RewardClaim, bool, error) {
	var existing RewardClaim
	err := tx.QueryRow(ctx, `
		SELECT id, reward_id, user_id, status, shipping_address, tracking_number, claimed_at, fulfilled_at, created_at, updated_at
		FROM reward_claims
		WHERE reward_id = $1 AND user_id = $2
	`, rewardID, userID).Scan(
		&existing.ID, &existing.RewardID, &existing.UserID, &existing.Status, &existing.ShippingAddress, &existing.TrackingNumber, &existing.ClaimedAt, &existing.FulfilledAt, &existing.CreatedAt, &existing.UpdatedAt,
	)
	if err == nil {
		return &existing, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, fmt.Errorf("gagal memeriksa klaim reward: %w", err)
	}

	claim := &RewardClaim{
		ID:       uuid.New(),
		RewardID: rewardID,
		UserID:   userID,
		Status:   ClaimIssued,
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO reward_claims (id, reward_id, user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (reward_id, user_id) DO NOTHING
	`, claim.ID, claim.RewardID, claim.UserID, claim.Status)
	if err != nil {
		return nil, false, fmt.Errorf("gagal menerbitkan klaim reward: %w", err)
	}
	return claim, true, nil
}

// Mengambil tiket klaim hadiah berdasarkan ID beserta detail reward dan pemiliknya.
func (r *Repository) GetClaimByID(ctx context.Context, id uuid.UUID) (*RewardClaim, error) {
	var c RewardClaim
	err := r.db.Pool.QueryRow(ctx, `
		SELECT c.id, c.reward_id, c.user_id, c.status, c.shipping_address, c.tracking_number, c.claimed_at, c.fulfilled_at, c.created_at, c.updated_at,
		       r.title, r.component_type, u.full_name
		FROM reward_claims c
		JOIN rewards r ON r.id = c.reward_id
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1
	`, id).Scan(
		&c.ID, &c.RewardID, &c.UserID, &c.Status, &c.ShippingAddress, &c.TrackingNumber, &c.ClaimedAt, &c.FulfilledAt, &c.CreatedAt, &c.UpdatedAt,
		&c.RewardTitle, &c.ComponentType, &c.UserFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari klaim reward: %w", err)
	}
	return &c, nil
}

// Mengambil daftar klaim hadiah milik pengguna tertentu.
func (r *Repository) ListClaimsByUser(ctx context.Context, userID uuid.UUID) ([]RewardClaim, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT c.id, c.reward_id, c.user_id, c.status, c.shipping_address, c.tracking_number, c.claimed_at, c.fulfilled_at, c.created_at, c.updated_at,
		       r.title, r.component_type, u.full_name
		FROM reward_claims c
		JOIN rewards r ON r.id = c.reward_id
		JOIN users u ON u.id = c.user_id
		WHERE c.user_id = $1
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil klaim pengguna: %w", err)
	}
	defer rows.Close()

	var list []RewardClaim
	for rows.Next() {
		var c RewardClaim
		if err := rows.Scan(
			&c.ID, &c.RewardID, &c.UserID, &c.Status, &c.ShippingAddress, &c.TrackingNumber, &c.ClaimedAt, &c.FulfilledAt, &c.CreatedAt, &c.UpdatedAt,
			&c.RewardTitle, &c.ComponentType, &c.UserFullName,
		); err == nil {
			list = append(list, c)
		}
	}
	return list, nil
}

// Memperbarui status tiket klaim hadiah dalam transaksi.
func (r *Repository) UpdateClaimStatusTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string, shippingAddress, trackingNumber *string) error {
	query := `
		UPDATE reward_claims
		SET status = $2::text,
		    shipping_address = COALESCE($3::text, shipping_address),
		    tracking_number = COALESCE($4::text, tracking_number),
		    claimed_at = COALESCE($5::timestamptz, claimed_at),
		    fulfilled_at = COALESCE($6::timestamptz, fulfilled_at),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	var claimedAt, fulfilledAt *time.Time
	now := time.Now().UTC()
	switch status {
	case ClaimClaimed:
		claimedAt = &now
	case ClaimFulfilled:
		fulfilledAt = &now
	}
	cmdTag, err := tx.Exec(ctx, query, id, status, shippingAddress, trackingNumber, claimedAt, fulfilledAt)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status klaim: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("klaim reward tidak ditemukan")
	}
	return nil
}
