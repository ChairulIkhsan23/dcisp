-- ============================================================================
-- DCISP PLATFORM v1.0 — MIGRATION 000008: PERFORMANCE INDEXES & DEDUPLICATION (UP)
-- Source: Database Performance Audit Report DCISP v1.0
-- ============================================================================

-- 1. Hapus Indeks Duplikat Redundan (F-DB-05, F-DB-06)
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_attendance_idempotency;

-- 2. Indeks Kritis Hot Path Ingestion NFC (F-DB-01 / Q-ATT-11)
CREATE INDEX IF NOT EXISTS idx_interns_id_number 
    ON interns(id_number) WHERE id_number IS NOT NULL;

-- 3. Indeks Sesi Istirahat (F-DB-03 / Q-ATT-17)
CREATE INDEX IF NOT EXISTS idx_breaks_session_start 
    ON breaks(session_id, start_time DESC);

-- 4. Indeks Komposit Idempotensi Mutasi XP (F-DB-04 / Q-PRF-03)
CREATE INDEX IF NOT EXISTS idx_xp_tx_user_reference 
    ON xp_transactions(user_id, reference_event);

-- 5. Indeks Riwayat Klaim Reward User (F-DB-08 / Q-FIN-26)
CREATE INDEX IF NOT EXISTS idx_reward_claims_user_created 
    ON reward_claims(user_id, created_at DESC);

-- 6. Indeks Tugas Berdasarkan Penugasan & Status (F-DB-09 / Q-PRF-22)
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_status 
    ON tasks(assignee_id, status);

-- 7. Indeks Berkas Bukti Deliverable (F-DB-09 / Q-PRJ-33)
CREATE INDEX IF NOT EXISTS idx_evidence_submission 
    ON evidence(submission_id);

-- 8. Indeks Milestone Proyek (F-DB-09 / Q-PRJ-21)
CREATE INDEX IF NOT EXISTS idx_milestones_project_deadline 
    ON milestones(project_id, deadline ASC);

-- 9. Indeks Komposit Penentuan Top Performer per Batch (F-DB-10 / Q-PRF-14)
CREATE INDEX IF NOT EXISTS idx_perf_eval_batch_period 
    ON performance_evaluations(batch_id, period_type, composite_performance_score DESC);

-- 10. Indeks Komposit Filter Audit Akun Financial Ledger (F-DB-10 / Q-FIN-09)
CREATE INDEX IF NOT EXISTS idx_financial_ledgers_account_created 
    ON financial_ledgers(account_code, created_at DESC);

-- 11. Indeks Komposit Paging & Filter Peserta Magang (F-DB-10 / Q-PEO-22)
CREATE INDEX IF NOT EXISTS idx_interns_batch_status_created 
    ON interns(batch_id, status, created_at DESC);
