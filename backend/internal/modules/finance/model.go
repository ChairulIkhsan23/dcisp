package finance

import (
	"time"

	"github.com/google/uuid"
)

// Bagan akun (Chart of Accounts) resmi platform DCISP.
// Seluruh jurnal double-entry wajib menggunakan kode akun berikut agar agregasi konsisten.
const (
	AcctCash            = "1001-CASH"
	AcctProjectPool     = "1201-PROJECT-BOUNTY-POOL"
	AcctMemberLiability = "2101-LIABILITY-MEMBER"
	AcctPayoutPayable   = "2102-PAYOUT-PAYABLE"
	AcctTaxPayable      = "2201-TAX-PAYABLE"
	AcctBatchFund       = "3101-BATCH-FUND"
	AcctOtherDeductions = "4901-OTHER-DEDUCTIONS"
	AcctRewardExpense   = "5001-REWARD-EXPENSE"
)

// Arah jurnal double-entry.
const (
	DirectionDebit  = "DEBIT"
	DirectionCredit = "CREDIT"
)

// Status mutasi dompet (BRD-05 bagian 16).
const (
	WalletTxPending   = "PENDING"
	WalletTxCompleted = "COMPLETED"
	WalletTxFailed    = "FAILED"
	WalletTxReversed  = "REVERSED"
)

// Sumber mutasi dompet (DATA-006, BRULE-FIN-003).
const (
	WalletSourceBounty = "PROJECT_BOUNTY"
	WalletSourceOT     = "OVERTIME_BONUS"
	WalletSourceReward = "RANK_REWARD"
	WalletSourceTax    = "TAX_DEDUCTION"
	WalletSourceFund   = "BATCH_CONTRIBUTION"
	WalletSourcePayout = "PAYOUT"
	WalletSourceOther  = "OTHER"
)

// Status payout (BRD-05 bagian 16).
const (
	PayoutRequested  = "REQUESTED"
	PayoutApproved   = "APPROVED"
	PayoutProcessing = "PROCESSING"
	PayoutSettled    = "SETTLED"
	PayoutFailed     = "FAILED"
	PayoutRejected   = "REJECTED"
)

// Tipe aturan deduksi (DATA-005).
const (
	RuleIncomeTax      = "INCOME_TAX"
	RuleProjectTax     = "PROJECT_TAX"
	RuleGraduationTax  = "GRADUATION_TAX"
	RuleWithdrawalTax  = "WITHDRAWAL_TAX"
	RuleAdminFee       = "ADMINISTRATIVE_FEE"
	RulePenalty        = "PENALTY"
	RuleOtherDeduction = "OTHER_DEDUCTION"
	RatePercentage     = "PERCENTAGE"
	RateFixedAmount    = "FIXED_AMOUNT"
)

// Tipe komponen reward (FR-031).
const (
	RewardCash      = "CASH"
	RewardBonusXP   = "BONUS_XP"
	RewardPhysical  = "PHYSICAL_ITEM"
	RewardBadge     = "BADGE"
	RewardVoucher   = "VOUCHER"
	RewardPrivilege = "PLATFORM_PRIVILEGE"
)

// Status klaim reward fisik (skema database).
const (
	ClaimIssued     = "ISSUED"
	ClaimClaimed    = "CLAIMED"
	ClaimProcessing = "PROCESSING"
	ClaimFulfilled  = "FULFILLED"
	ClaimRejected   = "REJECTED"
)

// Entitas Dompet Personal (FR-035). Saldo dalam Rupiah penuh (int64, tanpa desimal).
type Wallet struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	CurrentBalance int64     `json:"current_balance" db:"current_balance"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Entitas Mutasi Dompet Itemized (DATA-006, BRULE-FIN-003). Nilai dalam Rupiah penuh.
type WalletTransaction struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	WalletID        uuid.UUID  `json:"wallet_id" db:"wallet_id"`
	Source          string     `json:"source" db:"source"`
	ReferenceID     string     `json:"reference_id" db:"reference_id"`
	GrossAmount     int64      `json:"gross_amount" db:"gross_amount"`
	DeductionAmount int64      `json:"deduction_amount" db:"deduction_amount"`
	NetAmount       int64      `json:"net_amount" db:"net_amount"`
	Status          string     `json:"status" db:"status"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty" db:"approved_by"`
}

// Entitas Baris Jurnal Buku Besar Ganda (FR-037, BRULE-FIN-004). Nilai dalam Rupiah penuh.
type FinancialLedger struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TransactionID  uuid.UUID `json:"transaction_id" db:"transaction_id"`
	AccountCode    string    `json:"account_code" db:"account_code"`
	Direction      string    `json:"direction" db:"direction"`
	Amount         int64     `json:"amount" db:"amount"`
	ReferenceTable string    `json:"reference_table" db:"reference_table"`
	ReferenceID    uuid.UUID `json:"reference_id" db:"reference_id"`
	Narration      string    `json:"narration" db:"narration"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Entitas Aturan Pajak dan Deduksi (DATA-005, FR-034).
type DeductionTaxRule struct {
	ID               uuid.UUID `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Type             string    `json:"type" db:"type"`
	RateType         string    `json:"rate_type" db:"rate_type"`
	RateValueUnits   int64     `json:"rate_value_units" db:"rate_value_units"` // Nilai rate dalam satuan 0.0001 (untuk PERCENTAGE) atau Rupiah (untuk FIXED_AMOUNT)
	CalculationBasis string    `json:"calculation_basis" db:"calculation_basis"`
	MinimumAmount    *int64    `json:"minimum_amount,omitempty" db:"minimum_amount"`
	MaximumAmount    *int64    `json:"maximum_amount,omitempty" db:"maximum_amount"`
	EffectiveDate    time.Time `json:"effective_date" db:"effective_date"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Entitas Kas Bersama Angkatan (FR-036, BRULE-FIN-002). Nilai dalam Rupiah penuh.
type BatchFund struct {
	ID               uuid.UUID `json:"id" db:"id"`
	BatchID          uuid.UUID `json:"batch_id" db:"batch_id"`
	TotalAccumulated int64     `json:"total_accumulated" db:"total_accumulated"`
	CurrentBalance   int64     `json:"current_balance" db:"current_balance"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Entitas Permohonan Pencairan Dana (FR-038). Nilai dalam Rupiah penuh.
type Payout struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	WalletID            uuid.UUID `json:"wallet_id" db:"wallet_id"`
	Amount              int64     `json:"amount" db:"amount"`
	BankCode            string    `json:"bank_code" db:"bank_code"`
	AccountNumber       string    `json:"account_number" db:"account_number"`
	AccountHolderName   string    `json:"account_holder_name" db:"account_holder_name"`
	Status              string    `json:"status" db:"status"`
	SettlementReference *string   `json:"settlement_reference,omitempty" db:"settlement_reference"`
	IdempotencyKey      *string   `json:"idempotency_key,omitempty" db:"idempotency_key"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`

	UserID       uuid.UUID `json:"user_id,omitempty"`
	UserFullName string    `json:"user_full_name,omitempty"`
}

// Entitas Master Hadiah Promosi Rank (FR-031).
type Reward struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	RankID        *uuid.UUID `json:"rank_id,omitempty" db:"rank_id"`
	AchievementID *uuid.UUID `json:"achievement_id,omitempty" db:"achievement_id"`
	ComponentType string     `json:"component_type" db:"component_type"`
	Title         string     `json:"title" db:"title"`
	MonetaryValue int64      `json:"monetary_value" db:"monetary_value"`
	Description   *string    `json:"description,omitempty" db:"description"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`

	RankName *string `json:"rank_name,omitempty"`
}

// Entitas Tiket Klaim Hadiah Fisik (FR-033).
type RewardClaim struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	RewardID        uuid.UUID  `json:"reward_id" db:"reward_id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	Status          string     `json:"status" db:"status"`
	ShippingAddress *string    `json:"shipping_address,omitempty" db:"shipping_address"`
	TrackingNumber  *string    `json:"tracking_number,omitempty" db:"tracking_number"`
	ClaimedAt       *time.Time `json:"claimed_at,omitempty" db:"claimed_at"`
	FulfilledAt     *time.Time `json:"fulfilled_at,omitempty" db:"fulfilled_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	RewardTitle   string `json:"reward_title,omitempty"`
	ComponentType string `json:"component_type,omitempty"`
	UserFullName  string `json:"user_full_name,omitempty"`
}

// Hasil evaluasi mesin deduksi untuk satu nilai bruto (Rupiah penuh).
type DeductionResult struct {
	RuleID     uuid.UUID `json:"rule_id"`
	RuleName   string    `json:"rule_name"`
	RuleType   string    `json:"rule_type"`
	Amount     int64     `json:"amount"`
	IsTax      bool      `json:"is_tax"`
	IsFarewell bool      `json:"is_farewell"`
}

// Ringkasan alokasi Gross → Tax → Farewell → Net (FR-039).
type FlowAllocation struct {
	Gross         int64             `json:"gross"`
	TaxTotal      int64             `json:"tax_total"`
	FarewellTotal int64             `json:"farewell_total"`
	OtherTotal    int64             `json:"other_total"`
	Net           int64             `json:"net"`
	Deductions    []DeductionResult `json:"deductions"`
}

// ============================================================================
// DTOs
// ============================================================================

type CreateTaxRuleRequest struct {
	Name             string  `json:"name" binding:"required"`
	Type             string  `json:"type" binding:"required"`
	RateType         string  `json:"rate_type" binding:"required"`
	RateValue        float64 `json:"rate_value" binding:"required,min=0"`
	CalculationBasis string  `json:"calculation_basis"`
	MinimumAmount    *int64  `json:"minimum_amount"`
	MaximumAmount    *int64  `json:"maximum_amount"`
	EffectiveDate    string  `json:"effective_date" binding:"required"`
	IsActive         *bool   `json:"is_active"`
}

type DistributeBountyRequest struct {
	ProjectID uuid.UUID `json:"project_id" binding:"required"`
}

type MemberDistribution struct {
	UserID        uuid.UUID `json:"user_id"`
	UserFullName  string    `json:"user_full_name"`
	FinalPct      float64   `json:"final_pct"`
	Gross         int64     `json:"gross"`
	TaxTotal      int64     `json:"tax_total"`
	FarewellTotal int64     `json:"farewell_total"`
	Net           int64     `json:"net"`
	TransactionID uuid.UUID `json:"transaction_id"`
}

type BountyDistributionResponse struct {
	ProjectID        uuid.UUID            `json:"project_id"`
	ProjectTitle     string               `json:"project_title"`
	BountyPool       int64                `json:"bounty_pool"`
	TotalDistributed int64                `json:"total_distributed"`
	Members          []MemberDistribution `json:"members"`
	Skipped          []string             `json:"skipped,omitempty"`
}

type RequestPayoutRequest struct {
	Amount            int64   `json:"amount" binding:"required,min=1"`
	BankCode          string  `json:"bank_code" binding:"required"`
	AccountNumber     string  `json:"account_number" binding:"required"`
	AccountHolderName string  `json:"account_holder_name" binding:"required"`
	IdempotencyKey    *string `json:"idempotency_key"`
}

type ReviewPayoutRequest struct {
	Action              string  `json:"action" binding:"required"` // APPROVE, REJECT, PROCESSING, SETTLE, FAIL
	SettlementReference *string `json:"settlement_reference"`
}

type CreateRewardRequest struct {
	RankID        *uuid.UUID `json:"rank_id"`
	AchievementID *uuid.UUID `json:"achievement_id"`
	ComponentType string     `json:"component_type" binding:"required"`
	Title         string     `json:"title" binding:"required"`
	MonetaryValue int64      `json:"monetary_value" binding:"min=0"`
	Description   *string    `json:"description"`
}

type IssueRankRewardRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	RankID uuid.UUID `json:"rank_id" binding:"required"`
}

type ClaimRewardRequest struct {
	ShippingAddress *string `json:"shipping_address"`
}

type ProcessClaimRequest struct {
	Action         string  `json:"action" binding:"required"` // PROCESSING, FULFILLED, REJECTED
	TrackingNumber *string `json:"tracking_number"`
}
