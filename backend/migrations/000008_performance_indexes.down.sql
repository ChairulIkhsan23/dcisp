-- ============================================================================
-- DCISP PLATFORM v1.0 — MIGRATION 000008: PERFORMANCE INDEXES (DOWN)
-- ============================================================================

-- Drop Indeks yang ditambahkan
DROP INDEX IF EXISTS idx_interns_id_number;
DROP INDEX IF EXISTS idx_breaks_session_start;
DROP INDEX IF EXISTS idx_xp_tx_user_reference;
DROP INDEX IF EXISTS idx_reward_claims_user_created;
DROP INDEX IF EXISTS idx_tasks_assignee_status;
DROP INDEX IF EXISTS idx_evidence_submission;
DROP INDEX IF EXISTS idx_milestones_project_deadline;
DROP INDEX IF EXISTS idx_perf_eval_batch_period;
DROP INDEX IF EXISTS idx_financial_ledgers_account_created;
DROP INDEX IF EXISTS idx_interns_batch_status_created;

-- Kembalikan Indeks duplikat lama
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_attendance_idempotency ON attendance_event_logs(idempotency_key);
