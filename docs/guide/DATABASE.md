# DCISP v1.0 DATABASE ARCHITECTURE & SCHEMA SPECIFICATION
**Platform:** Dagang Creative Intern Solutions Program (DCISP)  
**Database Engine:** PostgreSQL 16+ (Alpine)  
**ORM / Driver:** `jackc/pgx/v5` (Go)  
**Source of Truth:** `docs/prd/PRD-DCISP-V1.md` (Section 10), `docs/architecture/FINAL-TECH-STACK-SPEC-DCISP.md`, Suite 9 BRDs (`docs/brd/`)

---

## 1. Database Strategy

### 1.1 Primary Database Architecture
* **Database Engine:** PostgreSQL 16+ with extensions `uuid-ossp` and `pg_trgm`.
* **ACID Compliance:** Strict relational consistency with isolation level `READ COMMITTED` by default and `SERIALIZABLE` / row-level locks for concurrent quota allocation and wallet settlements.
* **Naming Conventions:**
  - Table names: `snake_case` plural (e.g., `users`, `attendance_event_logs`, `financial_ledgers`).
  - Column names: `snake_case` singular (e.g., `user_id`, `created_at`, `planned_contribution_pct`).
  - Primary keys: `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`.
  - Foreign keys: `fk_<table>_<referenced_table>`.
  - Indexes: `idx_<table>_<column(s)>`.

### 1.2 Data Type & Monetary Precision Strategy
* **Financial & Monetary Columns:** Strictly stored as `NUMERIC(15, 2)` (fixed-point decimal, representing exact Indonesian Rupiah). Floating-point data types (`FLOAT`, `DOUBLE PRECISION`) are **STRICTLY PROHIBITED** for monetary values to eliminate rounding drift.
* **Percentages:** Stored as `NUMERIC(5, 2)` (e.g., `100.00` to `0.00`).
* **Timestamps:** Stored with timezone (`TIMESTAMPTZ`), saved in UTC and rendered to client in `Asia/Jakarta` (WIB / UTC+7).
* **Dynamic Configuration:** Stored in `JSONB` with `GIN` indexing for runtime flexibility.

### 1.3 Immutability & Append-Only Strategy (BR-023, BR-025, FR-047)
The following tables are structurally **IMMUTABLE** at the database engine level via native PL/pgSQL triggers (`enforce_immutable_records()`):
1. `attendance_event_logs` → trigger `trg_immutable_attendance`
2. `financial_ledgers` → trigger `trg_immutable_financial_ledger`
3. `audit_logs` → trigger `trg_immutable_audit_logs`
4. `xp_transactions` → trigger `trg_immutable_xp_transactions`

Direct `UPDATE` and `DELETE` queries on these tables are blocked and raise database exceptions. Corrections must be recorded as new offsetting/reversal entries (`REVERSAL` or `MANUAL_CORRECTION`).

Additionally, `financial_ledgers` enforces per-transaction double-entry balance at commit time
via deferred constraint trigger `trg_ledger_balanced` (`enforce_balanced_ledger()`,
migration `000005`): any `transaction_id` whose `ΣDebit − ΣCredit ≠ 0` aborts the
transaction (BR-023, BRULE-FIN-004).

### 1.4 Binary Storage Isolation Strategy (BR-024)
Storing binary data (`BYTEA`, `BLOB`) inside PostgreSQL is **STRICTLY PROHIBITED**. The database only stores metadata records inside the `file_metadata` table (DATA-007). The physical files reside in Cloudflare R2 object storage.

---

## 2. Core Entities & Schema Catalog

### 2.1 Identity & RBAC Domain
* **`users`:** Primary user credentials and account status.
  - `id` (UUID PK), `email` (VARCHAR 255 UNIQUE), `password_hash` (VARCHAR 255, Argon2id), `full_name` (VARCHAR 255), `avatar_file_id` (UUID FK), `status` (ENUM: `PENDING`, `ACTIVE`, `SUSPENDED`, `INACTIVE`), `created_at`, `updated_at`.
* **`roles`:** Master system roles (10 standard roles R-01 to R-10).
  - `id` (UUID PK), `name` (VARCHAR 64 UNIQUE), `description` (TEXT), `is_system` (BOOLEAN).
* **`scopes`:** Context access boundaries.
  - `id` (UUID PK), `name` (VARCHAR 64), `scope_type` (ENUM: `SYSTEM`, `WORKFORCE_AND_PEOPLE`, `ASSIGNED_TEAM`, `ASSIGNED_PROJECTS`, `FINANCIAL_DATA`, `SCANNER_ONLY`, `OWN_DATA`, `PUBLIC_DATA`), `context_id` (UUID NULLABLE).
* **`permissions`:** Granular access permissions mapping.
  - `id` (UUID PK), `role_id` (UUID FK), `resource` (VARCHAR 64), `action` (VARCHAR 64), `scope_type` (VARCHAR 64).
* **`user_roles`:** Junction table mapping users to roles and specific scopes.

### 2.2 People & Batches Domain
* **`institutions`:** Educational institutions (universities, vocational schools).
  External catalog integration (migration `000004`): `external_id` (VARCHAR 128,
  e.g. `kampus:pt_001`, unique where not null via `uq_institutions_external_id`),
  `source` (`MANUAL`, `API_KAMPUS`, `API_SEKOLAH`), `synced_at` (TIMESTAMPTZ),
  indexed by `idx_institutions_source`.
* **`holidays`:** Operational/national holiday calendar entries referenced by work schedules
  (`id`, `holiday_calendar_id`, `date`, `name`, `is_national`).
* **`batches`:** Internship cohort groups.
  - `id` (UUID PK), `batch_code` (VARCHAR 64 UNIQUE), `name` (VARCHAR 255), `start_date` (DATE), `end_date` (DATE), `quota` (INTEGER), `status` (ENUM: `DRAFT`, `ACTIVE`, `COMPLETED`, `ARCHIVED`).
* **`interns`:** Active intern lifecycle records.
  - `id` (UUID PK), `user_id` (UUID FK UNIQUE), `batch_id` (UUID FK), `institution_id` (UUID FK), `id_number` (VARCHAR 64), `mentor_id` (UUID FK), `status` (ENUM: `APPLICANT`, `ONBOARDING`, `ACTIVE`, `ON_LEAVE`, `SUSPENDED`, `GRADUATED`, `TERMINATED`), `join_date`, `end_date`, `current_rank_id` (UUID FK), `internship_xp` (INTEGER DEFAULT 0).
* **`alumni`:** Permanent graduate records (BR-003).
  - `id` (UUID PK), `user_id` (UUID FK UNIQUE), `batch_id` (UUID FK), `graduation_date` (DATE), `alumni_xp` (INTEGER DEFAULT 0), `certificate_id` (UUID FK), `is_public_profile` (BOOLEAN).
* **`skills` & `user_skills`:** Standardized skill catalog and proficiency levels (1–5).

### 2.3 Workforce & Attendance Domain
* **`devices`:** Registry of hardware terminal devices (ESP32 / NFC readers / cameras).
* **`work_schedules` (DATA-002):** Schedule specifications, working days, core hours, break windows, and grace periods.
* **`attendance_event_logs` (DATA-001):** Append-only presence event stream.
  - `id` (UUID PK), `user_id` (UUID FK), `device_id` (UUID FK), `event_type` (ENUM: `ARRIVED`, `CHECK_IN`, `WORK_STARTED`, `BREAK_STARTED`, `BREAK_ENDED`, `WORK_RESUMED`, `OVERTIME_STARTED`, `OVERTIME_ENDED`, `CHECK_OUT`), `timestamp` (TIMESTAMPTZ), `method` (ENUM: `NFC`, `QR_CODE`, `ADMIN_SCANNER`, `NFC_OFFLINE_SYNC`, `MANUAL_CORRECTION`), `location_context`, `session_id` (UUID), `idempotency_key` (VARCHAR 64 UNIQUE), `metadata` (JSONB).
* **`work_sessions`:** Productive work session tracker measuring gross, active, and idle durations.
* **`breaks`:** Explicit break tracking records with anomaly flags (`is_anomaly_early`, `is_violation_late`).
* **`overtime_requests`:** Overtime request and approval records.
* **`leave_requests`:** Formal leave applications and doctor certificate links.
* **`attendance_corrections`:** Manual correction adjustments preserving historical logs (BR-025).

### 2.4 Projects, Tasks & Evidence Domain
* **`projects` (DATA-003):** Initiatives published on marketplace.
  - `id` (UUID PK), `title`, `description`, `owner_id` (UUID FK), `visibility` (`INTERN_ONLY`, `PUBLIC`, `PRIVATE`), `required_skills` (UUID[]), `capacity` (INTEGER), `accepted_count` (INTEGER), `deadline` (TIMESTAMPTZ), `bounty_pool` (NUMERIC 15,2), `status` (`DRAFT`, `PUBLISHED`, `IN_PROGRESS`, `COMPLETED`, `CANCELLED`).
* **`project_applications`:** Candidate applications and skill match score.
* **`project_teams`:** Team structure and 3-layer contribution percentages (`planned_contribution_pct`, `actual_contribution_pct`, `final_contribution_pct`, `is_locked`).
* **`milestones`:** Project phase targets with weights (sum = 100%).
* **`tasks`:** Kanban work cards with difficulty weight (1–5) and status.
* **`work_reports` (DATA-004):** Work submissions and accountability reports.
* **`evidence`:** Deliverable artifacts linked to Cloudflare R2 files or Git URLs.

### 2.5 Performance & Gamification Domain
* **`policies`:** Central Unified Policy Engine rules stored as JSONB predicates (FR-045).
* **`xp_rules`:** Configurable point and penalty triggers.
* **`xp_transactions`:** Append-only ledger of experience point mutations partitioned into 3 isolated schemes (`INTERNSHIP_XP`, `PROJECT_XP`, `ALUMNI_CONTRIBUTION` - BR-004).
  *Catatan DOC-02: nama kanonis skema magang adalah `INTERNSHIP_XP` (CHECK constraint
  `xp_rules_xp_scheme_check` / `xp_transactions_scheme_check`); penyebutan `INTERN_XP`
  pada revisi dokumen lama merujuk pada skema yang sama.*
* **`ranks`:** Class tier levels (Novice to Grandmaster).
* **`achievements` & `user_achievements`:** Milestone badge achievements.
* **`performance_evaluations`:** Formal composite supervisor rubric scores (0–100) distinct from Rank (BR-017).

### 2.6 Finance, Taxation & Ledger Domain
* **`rewards` & `reward_claims`:** Multi-component rank promotion reward packages (Cash, XP, Merchandise, Badges, Vouchers) (BR-019).
* **`wallets`:** Personal digital wallet balances (`current_balance NUMERIC(15,2) >= 0.00`).
* **`wallet_transactions` (DATA-006):** Itemized transaction stream with gross, deduction, net, and source origin (BR-022).
  Idempotency guard: `uq_wallet_tx_reference` UNIQUE on `reference_id` (migration `000005`).
* **`payouts`:** Cashout withdrawal requests; idempotency guard `uq_payouts_idempotency`
  UNIQUE on nullable `idempotency_key` (migration `000005`).
* **`reward_claims`:** One ticket per reward-user pair enforced by
  `uq_reward_claims_reward_user` UNIQUE on `(reward_id, user_id)` (migration `000005`).
* **`refresh_sessions`:** Hashed refresh-token sessions for rotation and instant
  revocation (migration `000006`): `user_id`, `token_hash` (SHA-256, UNIQUE via
  `uq_refresh_sessions_token_hash`), `expires_at`, `revoked_at`.
* **`batch_funds`:** Cohort farewell treasury accumulated from batch contributions (BR-021).
* **`deduction_tax_rules` (DATA-005):** Dynamic income tax, project tax, and batch fee rules.
* **`payouts`:** Cashout withdrawal requests to external bank accounts.
* **`financial_ledgers`:** Double-entry accounting ledger entries ($\sum \text{Debit} = \sum \text{Credit}$, append-only) (BR-023).

### 2.7 Documents & Audit Domain
* **`file_metadata` (DATA-007):** Cloudflare R2 object metadata (disk, path_key, filename, mime_type, size_bytes).
* **`certificates`:** Digital graduation scrolls with cryptographic verification hashes.
* **`portfolios`:** Public alumni showcase configurations (`/portfolio/slug`).
* **`audit_logs`:** Immutable security audit event trails (IP, User-Agent, Old State, New State).
* **`system_settings` & `notifications`:** Key-value platform settings and in-app notifications.

---

## 3. Entity Relationships Matrix

```text
User ─────────────┬── (1 : 1) ──► Intern (Active Profile)
                  ├── (1 : 1) ──► Alumni (Permanent Graduate)
                  ├── (1 : 1) ──► Wallet (Personal Coin Pouch)
                  ├── (1 : N) ──► User Roles ──► Roles ──► Permissions
                  ├── (1 : N) ──► Attendance Event Logs (Immutable)
                  ├── (1 : N) ──► Work Sessions ──► Breaks
                  ├── (1 : N) ──► Project Teams ──► Projects
                  ├── (1 : N) ──► Tasks ──► Work Reports ──► Evidence ──► File Metadata (R2)
                  ├── (1 : N) ──► XP Transactions (Immutable)
                  └── (1 : N) ──► Audit Logs (Immutable)

Batch ────────────┬── (1 : N) ──► Interns
                  └── (1 : 1) ──► Batch Fund (Treasury)

Project ──────────┬── (1 : N) ──► Milestones ──► Tasks
                  ├── (1 : N) ──► Project Teams (3-Tier Honor Split)
                  └── (1 : N) ──► Project Applications

Financial Ledger ─┴── (N : 1) ──► Transaction ID (Balanced Debit & Credit Entries)
```

---

## 4. Data Integrity Constraints

| Constraint Name | Type | Enforcement Rule | Target Entity |
|---|---|---|---|
| `chk_wallet_net` | `CHECK` | `net_amount = gross_amount - deduction_amount` | `wallet_transactions` |
| `wallets_current_balance_check` (didokumentasikan sebagai `chk_wallet_balance`) | `CHECK` | `current_balance >= 0.00` (No unauthorized overdraft) | `wallets` |
| `chk_schedule_times` | `CHECK` | `start_time < end_time AND break_start < break_end AND break_start >= start_time AND break_end <= end_time` | `work_schedules` |
| `chk_batch_dates` | `CHECK` | `end_date >= start_date` | `batches` |
| `chk_leave_dates` | `CHECK` | `end_date >= start_date` | `leave_requests` |
| `attendance_event_logs_idempotency_key_key` (didokumentasikan sebagai `uq_idempotency_key`) | `UNIQUE` | Eliminates duplicate hardware/offline ingestion | `attendance_event_logs` |
| `uq_wallet_tx_reference` | `UNIQUE` | Idempotency key per mutasi dompet | `wallet_transactions` |
| `uq_payouts_idempotency` | `UNIQUE` | Idempotency key permohonan payout | `payouts` |
| `uq_reward_claims_reward_user` | `UNIQUE` | Satu tiket per pasangan reward-pengguna | `reward_claims` |
| `uq_refresh_sessions_token_hash` | `UNIQUE` | Satu baris per hash refresh token | `refresh_sessions` |
| `trg_immutable_*` | `TRIGGER` | Raises exception on any `UPDATE` or `DELETE` | `attendance_event_logs`, `financial_ledgers`, `audit_logs`, `xp_transactions` |
| `trg_ledger_balanced` | `CONSTRAINT TRIGGER` (deferred) | Aborts commit bila `ΣDebit − ΣCredit ≠ 0` per `transaction_id` | `financial_ledgers` |

---

## 5. Indexing Strategy

1. **Uniqueness & Foreign Keys:** Automatically indexed by B-tree on all `UUID` primary and foreign keys.
2. **Attendance Time-Series Queries:**
   - `CREATE INDEX idx_attendance_user_time ON attendance_event_logs(user_id, timestamp);`
   - `CREATE INDEX idx_work_sessions_user_date ON work_sessions(user_id, date);`
3. **Marketplace & Tasks Kanban Filtering:**
   - `CREATE INDEX idx_projects_status_vis ON projects(status, visibility);`
   - `CREATE INDEX idx_tasks_project_milestone ON tasks(project_id, milestone_id);`
4. **Ledger & Audit Reconciliation:**
   - `CREATE INDEX idx_financial_ledgers_tx ON financial_ledgers(transaction_id);`
   - `CREATE INDEX idx_audit_logs_event_time ON audit_logs(event_name, timestamp);`
   - `CREATE INDEX idx_wallet_tx_wallet ON wallet_transactions(wallet_id);`
   - `CREATE INDEX idx_payouts_wallet ON payouts(wallet_id);`
   - `CREATE INDEX idx_payouts_status ON payouts(status);`
   - `CREATE INDEX idx_refresh_sessions_user ON refresh_sessions(user_id);`
   - `CREATE INDEX idx_refresh_sessions_expiry ON refresh_sessions(expires_at);`
5. **Full-Text Trigram Search:**
   - `CREATE INDEX idx_users_name_trgm ON users USING gin(full_name gin_trgm_ops);`
   - `CREATE INDEX idx_projects_title_trgm ON projects USING gin(title gin_trgm_ops);`

---

## 6. Data Lifecycle & Retention

* **Permanent Retention (No Purge):** `users`, `interns`, `alumni`, `certificates`, `financial_ledgers`, `audit_logs`, `xp_transactions`. Stored permanently to honor lifetime alumni recognition (BR-003) and accounting auditability.
* **Transient Data Expiration:** Temporary PDF exports in Cloudflare R2 bucket `reports/` expire after **180 days** via R2 lifecycle rules.
