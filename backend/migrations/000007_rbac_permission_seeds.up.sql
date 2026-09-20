-- ============================================================================
-- DCISP PLATFORM v1.0 — RBAC PERMISSION SEEDS FOR STANDARD ROLES (UP)
-- Source of Truth: BR-001, BR-002, BR-003, docs/guide/SECURITY.md Section 2
-- Pola: Izin dievaluasi dinamis dari tabel ini (zero hardcoding di kode aplikasi).
-- Visibilitas marketplace proyectos diatur via resource projects.marketplace:
--   action view          -> INTERN_ONLY + PUBLIC (scope selain PUBLIC_DATA)
--   action view + scope PUBLIC_DATA saja -> PUBLIC saja (BR-003, peran ALUMNI)
--   action view_private  -> ditambah PRIVATE (hanya peran terotorisasi)
-- ============================================================================

-- 5. SEED OPERATIONAL ROLE PERMISSIONS
INSERT INTO permissions (role_id, resource, action, scope_type) VALUES
-- ADMIN: administrator operasional (cakupan SYSTEM pada domain operasional)
('10000000-0000-0000-0000-000000000002', 'people.institutions', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.institutions', 'create', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.institutions', 'update', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.institutions', 'delete', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.batches', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.batches', 'create', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.batches', 'update', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.interns', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.interns', 'create', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.interns', 'update', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.alumni', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.skills', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.skills', 'create', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.skills', 'update', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'people.skills', 'delete', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'projects.marketplace', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'projects.marketplace', 'view_private', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'projects.applications', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'projects.applications', 'review', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'attendance.overtime', 'review', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'attendance.leave', 'review', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'attendance.corrections', 'review', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'performance.evaluations', 'create', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'performance.top_performer', 'create', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.wallets', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.ledger', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.tax_rules', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.tax_rules', 'create', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.bounty', 'distribute', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.payouts', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.payouts', 'review', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.rewards', 'view', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.rewards', 'create', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.rewards', 'issue', 'SYSTEM'),
('10000000-0000-0000-0000-000000000002', 'finance.rewards', 'process', 'SYSTEM'),
-- HR_ADMIN: siklus hidup magang, batch, cuti, top performer
('10000000-0000-0000-0000-000000000003', 'people.institutions', 'view', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.institutions', 'create', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.institutions', 'update', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.batches', 'view', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.batches', 'create', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.batches', 'update', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.interns', 'view', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.interns', 'create', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.interns', 'update', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.alumni', 'view', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'people.skills', 'view', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'attendance.leave', 'review', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'performance.top_performer', 'create', 'WORKFORCE_AND_PEOPLE'),
('10000000-0000-0000-0000-000000000003', 'projects.marketplace', 'view', 'WORKFORCE_AND_PEOPLE'),
-- PROJECT_MANAGER: bursa proyek, lamaran, tim, milestone, tugas, kontribusi
('10000000-0000-0000-0000-000000000004', 'projects.marketplace', 'view', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.marketplace', 'view_private', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.marketplace', 'create', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.marketplace', 'update', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.applications', 'view', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.applications', 'review', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.team', 'update', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.milestones', 'create', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.tasks', 'create', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000004', 'projects.contribution', 'finalize', 'ASSIGNED_PROJECTS'),
-- SUPERVISOR: tim binaan (lembur, cuti, koreksi, evaluasi, kontribusi)
('10000000-0000-0000-0000-000000000005', 'attendance.overtime', 'review', 'ASSIGNED_TEAM'),
('10000000-0000-0000-0000-000000000005', 'attendance.leave', 'review', 'ASSIGNED_TEAM'),
('10000000-0000-0000-0000-000000000005', 'attendance.corrections', 'review', 'ASSIGNED_TEAM'),
('10000000-0000-0000-0000-000000000005', 'performance.evaluations', 'create', 'ASSIGNED_TEAM'),
('10000000-0000-0000-0000-000000000005', 'tasks.submissions', 'review', 'ASSIGNED_TASKS'),
('10000000-0000-0000-0000-000000000005', 'projects.contribution', 'finalize', 'ASSIGNED_PROJECTS'),
('10000000-0000-0000-0000-000000000005', 'projects.marketplace', 'view', 'ASSIGNED_PROJECTS'),
-- REVIEWER: telaah laporan tugas dan bukti deliverable
('10000000-0000-0000-0000-000000000006', 'tasks.submissions', 'review', 'ASSIGNED_TASKS'),
('10000000-0000-0000-0000-000000000006', 'projects.marketplace', 'view', 'ASSIGNED_PROJECTS'),
-- FINANCE: seluruh domain keuangan
('10000000-0000-0000-0000-000000000007', 'finance.wallets', 'view', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.ledger', 'view', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.tax_rules', 'view', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.tax_rules', 'create', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.bounty', 'distribute', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.payouts', 'view', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.payouts', 'review', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.rewards', 'view', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.rewards', 'create', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.rewards', 'issue', 'FINANCIAL_DATA'),
('10000000-0000-0000-0000-000000000007', 'finance.rewards', 'process', 'FINANCIAL_DATA'),
-- INTERN: bursa proyek internal + publik (tanpa PRIVATE)
('10000000-0000-0000-0000-000000000009', 'projects.marketplace', 'view', 'OWN_DATA'),
-- ALUMNI: hanya proyek publik (BR-003)
('10000000-0000-0000-0000-000000000010', 'projects.marketplace', 'view', 'PUBLIC_DATA'),
-- SCANNER_OPERATOR: direktori publik read-only (tanpa PRIVATE, tanpa mutasi)
('10000000-0000-0000-0000-000000000008', 'projects.marketplace', 'view', 'PUBLIC_DATA')
ON CONFLICT (role_id, resource, action, scope_type) DO NOTHING;
