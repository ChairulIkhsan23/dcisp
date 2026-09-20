-- Migration 000005 (Down): Menghapus penguatan integritas finansial Phase 7
DROP TRIGGER IF EXISTS trg_ledger_balanced ON financial_ledgers;
DROP FUNCTION IF EXISTS enforce_balanced_ledger();
DROP INDEX IF EXISTS uq_wallet_tx_reference;
DROP INDEX IF EXISTS uq_reward_claims_reward_user;
DROP INDEX IF EXISTS uq_payouts_idempotency;
DROP INDEX IF EXISTS idx_wallet_tx_wallet;
DROP INDEX IF EXISTS idx_payouts_wallet;
DROP INDEX IF EXISTS idx_payouts_status;
ALTER TABLE payouts DROP COLUMN IF EXISTS idempotency_key;
