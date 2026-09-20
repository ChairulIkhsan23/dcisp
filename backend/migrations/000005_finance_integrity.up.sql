-- Migration 000005: Penguatan integritas finansial Phase 7
-- 1. Idempotensi mutasi dompet melalui reference unik
-- 2. Kunci idempotensi pada payout untuk proteksi double-click
-- 3. Trigger deferred yang menolak jurnal ledger tidak seimbang per transaction_id (BR-023)
-- 4. Indeks agregasi untuk riwayat transaksi dan payout

CREATE UNIQUE INDEX IF NOT EXISTS uq_wallet_tx_reference ON wallet_transactions(reference_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_reward_claims_reward_user ON reward_claims(reward_id, user_id);

ALTER TABLE payouts ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128);
CREATE UNIQUE INDEX IF NOT EXISTS uq_payouts_idempotency ON payouts(idempotency_key);

CREATE INDEX IF NOT EXISTS idx_wallet_tx_wallet ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_payouts_wallet ON payouts(wallet_id);
CREATE INDEX IF NOT EXISTS idx_payouts_status ON payouts(status);

-- Fungsi validasi keseimbangan double-entry per transaction_id (BR-023, BRULE-FIN-004)
CREATE OR REPLACE FUNCTION enforce_balanced_ledger()
RETURNS TRIGGER AS $$
DECLARE
    v_balance NUMERIC(15, 2);
BEGIN
    SELECT COALESCE(SUM(CASE WHEN direction = 'DEBIT' THEN amount ELSE -amount END), 0.00)
    INTO v_balance
    FROM financial_ledgers
    WHERE transaction_id = COALESCE(NEW.transaction_id, OLD.transaction_id);

    IF v_balance <> 0.00 THEN
        RAISE EXCEPTION 'DCISP Financial Integrity Violation: jurnal tidak seimbang pada transaction_id % (selisih %)', COALESCE(NEW.transaction_id, OLD.transaction_id), v_balance;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_ledger_balanced ON financial_ledgers;
CREATE CONSTRAINT TRIGGER trg_ledger_balanced
AFTER INSERT ON financial_ledgers
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION enforce_balanced_ledger();
