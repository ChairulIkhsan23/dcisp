-- ============================================================================
-- DCISP PLATFORM v1.0 — INITIAL MASTER DATA SEED (UP)
-- Source of Truth: PRD-DCISP-V1.md Section 3.2, 3.5, 8.4, 8.6, 8.8
-- ============================================================================

-- 1. SEED 10 MASTER ROLES (R-01 s/d R-10)
INSERT INTO roles (id, name, description, is_system) VALUES
('10000000-0000-0000-0000-000000000001', 'SUPER_ADMIN', 'The Creator: Administrator tertinggi dengan wewenang sistem dan policy engine penuh', TRUE),
('10000000-0000-0000-0000-000000000002', 'ADMIN', 'Realm Warden: Administrator operasional harian ranah sistem', TRUE),
('10000000-0000-0000-0000-000000000003', 'HR_ADMIN', 'Game Master - Operations: Penanggung jawab siklus hidup magang, batch, dan kelulusan', TRUE),
('10000000-0000-0000-0000-000000000004', 'PROJECT_MANAGER', 'Quest Giver: Pengelola proyek, bursa misi, dan alokasi bounty pool', TRUE),
('10000000-0000-0000-0000-000000000005', 'SUPERVISOR', 'Guild Master / Party Captain: Pembimbing lapangan tim, approver lembur, koreksi, dan evaluasi', TRUE),
('10000000-0000-0000-0000-000000000006', 'REVIEWER', 'The Oracle: Penilai mutu deliverable tugas dan validasi bukti artefak', TRUE),
('10000000-0000-0000-0000-000000000007', 'FINANCE', 'The Merchant / Vault Keeper: Pengelola buku besar immutabel, pajak, batch fund, dan payout', TRUE),
('10000000-0000-0000-0000-000000000008', 'SCANNER_OPERATOR', 'Gatekeeper: Operator terminal fisik pemindai presensi di gerbang kantor (scope = scanner_only)', TRUE),
('10000000-0000-0000-0000-000000000009', 'INTERN', 'Adventurer / Hero Initiate: Peserta magang aktif pelaksana tugas dan petualang leveling XP', TRUE),
('10000000-0000-0000-0000-000000000010', 'ALUMNI', 'Legendary Hero: Lulusan magang dengan retensi akun permanen dan hak proyek publik', TRUE)
ON CONFLICT (id) DO NOTHING;

-- 2. SEED 8 MASTER SCOPES (Section 3.4)
INSERT INTO scopes (id, name, scope_type) VALUES
('20000000-0000-0000-0000-000000000001', 'System Scope', 'SYSTEM'),
('20000000-0000-0000-0000-000000000002', 'Workforce & People Scope', 'WORKFORCE_AND_PEOPLE'),
('20000000-0000-0000-0000-000000000003', 'Assigned Team Scope', 'ASSIGNED_TEAM'),
('20000000-0000-0000-0000-000000000004', 'Assigned Projects Scope', 'ASSIGNED_PROJECTS'),
('20000000-0000-0000-0000-000000000005', 'Financial Data Scope', 'FINANCIAL_DATA'),
('20000000-0000-0000-0000-000000000006', 'Scanner Only Scope', 'SCANNER_ONLY'),
('20000000-0000-0000-0000-000000000007', 'Own Data Scope', 'OWN_DATA'),
('20000000-0000-0000-0000-000000000008', 'Public Data Scope', 'PUBLIC_DATA')
ON CONFLICT (id) DO NOTHING;

-- 3. SEED SCANNER OPERATOR ISOLATION PERMISSION (BR-002)
INSERT INTO permissions (role_id, resource, action, scope_type) VALUES
('10000000-0000-0000-0000-000000000008', 'attendance.scanner', 'view', 'SCANNER_ONLY'),
('10000000-0000-0000-0000-000000000008', 'attendance.scanner', 'scan', 'SCANNER_ONLY')
ON CONFLICT (role_id, resource, action, scope_type) DO NOTHING;

-- 4. SEED SUPER ADMIN MASTER PERMISSIONS (FULL ACCESS)
INSERT INTO permissions (role_id, resource, action, scope_type) VALUES
('10000000-0000-0000-0000-000000000001', '*', '*', 'SYSTEM')
ON CONFLICT (role_id, resource, action, scope_type) DO NOTHING;

-- 5. SEED 5 CLASS TIERS / RANKS (Section 8.6)
INSERT INTO ranks (id, name, min_xp, level_order, badge_icon_url) VALUES
('30000000-0000-0000-0000-000000000001', 'Tier 1: Novice', 0, 1, '/badges/rank_novice.png'),
('30000000-0000-0000-0000-000000000002', 'Tier 2: Apprentice', 250, 2, '/badges/rank_apprentice.png'),
('30000000-0000-0000-0000-000000000003', 'Tier 3: Knight', 750, 3, '/badges/rank_knight.png'),
('30000000-0000-0000-0000-000000000004', 'Tier 4: Paladin', 1500, 4, '/badges/rank_paladin.png'),
('30000000-0000-0000-0000-000000000005', 'Tier 5: Grandmaster', 3000, 5, '/badges/rank_grandmaster.png')
ON CONFLICT (id) DO NOTHING;

-- 6. SEED DEFAULT WORK SCHEDULE (FR-007, DATA-002)
-- Schedule: Senin-Jumat (1-5), 08:30-17:00, Break 12:00-13:00, Grace Period 10m
INSERT INTO work_schedules (id, name, working_days, start_time, end_time, break_start, break_end, grace_period_minutes) VALUES
('40000000-0000-0000-0000-000000000001', 'Jadwal Standar Kantor PT. ADT (Senin-Jumat)', '{1,2,3,4,5}', '08:30:00', '17:00:00', '12:00:00', '13:00:00', 10)
ON CONFLICT (id) DO NOTHING;

-- 7. SEED DEFAULT XP RULES (FR-025, BR-012)
INSERT INTO xp_rules (id, event_trigger, xp_value, xp_scheme, description) VALUES
('50000000-0000-0000-0000-000000000001', 'CHECK_IN_ON_TIME', 10, 'INTERNSHIP_XP', 'Presensi kedatangan tepat waktu di kantor'),
('50000000-0000-0000-0000-000000000002', 'LATE_TIER_1', -1, 'INTERNSHIP_XP', 'Keterlambatan 1 - 15 menit'),
('50000000-0000-0000-0000-000000000003', 'LATE_TIER_2', -2, 'INTERNSHIP_XP', 'Keterlambatan 16 - 30 menit'),
('50000000-0000-0000-0000-000000000004', 'LATE_TIER_3', -3, 'INTERNSHIP_XP', 'Keterlambatan > 30 menit'),
('50000000-0000-0000-0000-000000000005', 'UNAUTHORIZED_BREAK', -2, 'INTERNSHIP_XP', 'Kembali dari istirahat melebihi toleransi jadwal'),
('50000000-0000-0000-0000-000000000006', 'MISSING_CHECKOUT', -2, 'INTERNSHIP_XP', 'Lupa melakukan presensi kepulangan checkout'),
('50000000-0000-0000-0000-000000000007', 'ABSENT_UNAUTHORIZED', -5, 'INTERNSHIP_XP', 'Mangkir tanpa keterangan sah'),
('50000000-0000-0000-0000-000000000008', 'LEAVE_APPROVED', 0, 'INTERNSHIP_XP', 'Cuti resmi yang telah disetujui (0 Penalti)'),
('50000000-0000-0000-0000-000000000009', 'TASK_COMPLETED_LOW', 15, 'PROJECT_XP', 'Penyelesaian tugas proyek tingkat kesulitan rendah'),
('50000000-0000-0000-0000-000000000010', 'TASK_COMPLETED_MED', 30, 'PROJECT_XP', 'Penyelesaian tugas proyek tingkat kesulitan menengah'),
('50000000-0000-0000-0000-000000000011', 'TASK_COMPLETED_HIGH', 50, 'PROJECT_XP', 'Penyelesaian tugas proyek tingkat kesulitan tinggi'),
('50000000-0000-0000-0000-000000000012', 'ALUMNI_TASK_COMPLETED', 30, 'ALUMNI_CONTRIBUTION', 'Penyelesaian tugas proyek publik oleh alumni')
ON CONFLICT (id) DO NOTHING;

-- 8. SEED DEDUCTION & TAX RULES (FR-034, BR-020, BR-021)
INSERT INTO deduction_tax_rules (id, name, type, rate_type, rate_value, calculation_basis, effective_date, is_active) VALUES
('60000000-0000-0000-0000-000000000001', 'PPh 21 Kompensasi Magang & Lepas', 'INCOME_TAX', 'PERCENTAGE', 5.0000, 'GROSS_AMOUNT', '2026-01-01', TRUE),
('60000000-0000-0000-0000-000000000002', 'Iuran Kas Bersama Angkatan (Batch Fund)', 'OTHER_DEDUCTION', 'PERCENTAGE', 2.0000, 'GROSS_AMOUNT', '2026-01-01', TRUE)
ON CONFLICT (id) DO NOTHING;

-- 9. SEED DEFAULT SYSTEM POLICIES (FR-045)
INSERT INTO policies (id, policy_domain, name, effective_date, condition_rules, action_definitions, priority_order, is_active) VALUES
('70000000-0000-0000-0000-000000000001', 'ATTENDANCE', 'Kebijakan Toleransi Kehadiran Standar', '2026-01-01', 
 '{"grace_period_minutes": 10, "max_daily_hours": 8}'::jsonb, 
 '{"on_time_xp": 10, "late_tier_1_xp": -1, "late_tier_2_xp": -2, "late_tier_3_xp": -3}'::jsonb, 
 1, TRUE),
('70000000-0000-0000-0000-000000000002', 'FINANCE', 'Kebijakan Bagi Hasil & Pajak Standar', '2026-01-01', 
 '{"income_tax_pct": 5.0, "batch_fund_pct": 2.0}'::jsonb, 
 '{"formula": "GROSS - (GROSS * (income_tax_pct/100)) - (GROSS * (batch_fund_pct/100))"}'::jsonb, 
 1, TRUE)
ON CONFLICT (id) DO NOTHING;

-- 10. SEED INITIAL SUPER ADMIN USER (Password: 'SuperAdminPass2026!')
-- Hash: Argon2id generated password for Super Admin
INSERT INTO users (id, email, password_hash, full_name, status) VALUES
('00000000-0000-0000-0000-000000000001', 'superadmin@dcisp.internal', '$argon2id$v=19$m=65536,t=3,p=2$YmFzZXNhbHQxMjM0NTY3OA$9vG9w4D8iZ2D1uL+ZkWXl6o9Wk2jV9a8S7d6F5g4H3I', 'The Creator (Super Administrator)', 'ACTIVE')
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_roles (user_id, role_id, scope_id) VALUES
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001')
ON CONFLICT (user_id, role_id, scope_id) DO NOTHING;
