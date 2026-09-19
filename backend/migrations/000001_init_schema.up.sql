-- ============================================================================
-- DCISP PLATFORM v1.0 — INITIAL DATABASE SCHEMA MIGRATION (UP)
-- Source of Truth: PRD-DCISP-V1.md & Suite 9 BRDs
-- Database: PostgreSQL 16+
-- ============================================================================

-- 0. EXTENSIONS
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================================================
-- 1. IDENTITY & ACCESS MANAGEMENT (RBAC)
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    avatar_file_id UUID,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('PENDING', 'ACTIVE', 'SUSPENDED', 'INACTIVE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(64) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scopes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(64) NOT NULL,
    scope_type VARCHAR(64) NOT NULL CHECK (scope_type IN ('SYSTEM', 'WORKFORCE_AND_PEOPLE', 'ASSIGNED_TEAM', 'ASSIGNED_PROJECTS', 'FINANCIAL_DATA', 'SCANNER_ONLY', 'OWN_DATA', 'PUBLIC_DATA')),
    context_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    resource VARCHAR(64) NOT NULL,
    action VARCHAR(64) NOT NULL,
    scope_type VARCHAR(64) NOT NULL DEFAULT 'OWN_DATA',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_permission_role_resource_action_scope UNIQUE(role_id, resource, action, scope_type)
);

CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    scope_id UUID REFERENCES scopes(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_role_scope UNIQUE(user_id, role_id, scope_id)
);

-- ============================================================================
-- 2. PEOPLE & LIFECYCLE MANAGEMENT
-- ============================================================================

CREATE TABLE IF NOT EXISTS institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    address TEXT,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_code VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    quota INTEGER NOT NULL DEFAULT 0 CHECK (quota >= 0),
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'ACTIVE', 'COMPLETED', 'ARCHIVED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_batch_dates CHECK (end_date >= start_date)
);

CREATE TABLE IF NOT EXISTS ranks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(64) NOT NULL,
    min_xp INTEGER NOT NULL DEFAULT 0 CHECK (min_xp >= 0),
    level_order INTEGER UNIQUE NOT NULL,
    badge_icon_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS interns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL REFERENCES batches(id) ON DELETE RESTRICT,
    institution_id UUID REFERENCES institutions(id) ON DELETE RESTRICT,
    id_number VARCHAR(64), -- NIM / NISN
    mentor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'APPLICANT' CHECK (status IN ('APPLICANT', 'ONBOARDING', 'ACTIVE', 'ON_LEAVE', 'SUSPENDED', 'GRADUATED', 'TERMINATED')),
    join_date DATE NOT NULL,
    end_date DATE NOT NULL,
    current_rank_id UUID REFERENCES ranks(id) ON DELETE SET NULL,
    internship_xp INTEGER NOT NULL DEFAULT 0 CHECK (internship_xp >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS alumni (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL REFERENCES batches(id) ON DELETE RESTRICT,
    graduation_date DATE NOT NULL,
    alumni_xp INTEGER NOT NULL DEFAULT 0 CHECK (alumni_xp >= 0),
    certificate_id UUID,
    is_public_profile BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(128) UNIQUE NOT NULL,
    category VARCHAR(64) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    proficiency_level INTEGER NOT NULL DEFAULT 1 CHECK (proficiency_level BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_skill UNIQUE(user_id, skill_id)
);

-- ============================================================================
-- 3. WORKFORCE & ATTENDANCE MANAGEMENT
-- ============================================================================

CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    terminal_identifier VARCHAR(64) UNIQUE NOT NULL,
    device_type VARCHAR(32) NOT NULL CHECK (device_type IN ('ESP32_NFC_TERMINAL', 'CAMERA_SCANNER', 'ADMIN_TERMINAL', 'MOBILE_OPERATOR')),
    location_name VARCHAR(128) NOT NULL,
    api_key_hash VARCHAR(128) NOT NULL,
    current_mode VARCHAR(32) NOT NULL DEFAULT 'AUTO' CHECK (current_mode IN ('CHECK_IN', 'CHECK_OUT', 'BREAK_START', 'BREAK_END', 'AUTO')),
    firmware_version VARCHAR(32) NOT NULL DEFAULT '1.0.0',
    audio_catalog_version VARCHAR(32) NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_heartbeat_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS work_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(128) NOT NULL,
    working_days INTEGER[] NOT NULL DEFAULT '{1,2,3,4,5}', -- 1=Senin, 5=Jumat
    start_time TIME NOT NULL DEFAULT '08:30:00',
    end_time TIME NOT NULL DEFAULT '17:00:00',
    break_start TIME NOT NULL DEFAULT '12:00:00',
    break_end TIME NOT NULL DEFAULT '13:00:00',
    grace_period_minutes INTEGER NOT NULL DEFAULT 10 CHECK (grace_period_minutes BETWEEN 0 AND 60),
    overtime_policy_id UUID,
    holiday_calendar_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_schedule_times CHECK (start_time < end_time AND break_start < break_end AND break_start >= start_time AND break_end <= end_time)
);

CREATE TABLE IF NOT EXISTS holidays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    holiday_calendar_id UUID,
    date DATE NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_national BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attendance_event_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    event_type VARCHAR(32) NOT NULL CHECK (event_type IN ('ARRIVED', 'CHECK_IN', 'WORK_STARTED', 'BREAK_STARTED', 'BREAK_ENDED', 'WORK_RESUMED', 'OVERTIME_STARTED', 'OVERTIME_ENDED', 'CHECK_OUT')),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    method VARCHAR(32) NOT NULL CHECK (method IN ('NFC', 'QR_CODE', 'ADMIN_SCANNER', 'NFC_OFFLINE_SYNC', 'MANUAL_CORRECTION')),
    location_context VARCHAR(128),
    session_id UUID,
    idempotency_key VARCHAR(64) UNIQUE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS work_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    gross_duration_seconds INTEGER NOT NULL DEFAULT 0 CHECK (gross_duration_seconds >= 0),
    active_duration_seconds INTEGER NOT NULL DEFAULT 0 CHECK (active_duration_seconds >= 0),
    idle_duration_seconds INTEGER NOT NULL DEFAULT 0 CHECK (idle_duration_seconds >= 0),
    active_task_id UUID,
    status VARCHAR(32) NOT NULL DEFAULT 'WORKING' CHECK (status IN ('WORKING', 'BREAK', 'PAUSED', 'ENDED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS breaks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES work_sessions(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    duration_seconds INTEGER NOT NULL DEFAULT 0 CHECK (duration_seconds >= 0),
    is_anomaly_early BOOLEAN NOT NULL DEFAULT FALSE,
    is_violation_late BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS overtime_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID,
    task_id UUID,
    date DATE NOT NULL,
    requested_start TIME NOT NULL,
    requested_end TIME NOT NULL,
    approved_start TIME,
    approved_end TIME,
    actual_worked_minutes INTEGER NOT NULL DEFAULT 0 CHECK (actual_worked_minutes >= 0),
    reason TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'SUBMITTED' CHECK (status IN ('SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED', 'COMPLETED', 'CANCELLED')),
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    review_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leave_type VARCHAR(32) NOT NULL CHECK (leave_type IN ('SICK', 'ACADEMIC', 'URGENT_PERSONAL', 'OTHER')),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT NOT NULL,
    evidence_file_id UUID,
    status VARCHAR(32) NOT NULL DEFAULT 'SUBMITTED' CHECK (status IN ('SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED', 'CANCELLED')),
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_leave_dates CHECK (end_date >= start_date)
);

CREATE TABLE IF NOT EXISTS attendance_corrections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attendance_event_id UUID REFERENCES attendance_event_logs(id) ON DELETE SET NULL,
    target_date DATE NOT NULL,
    proposed_event_type VARCHAR(32) NOT NULL CHECK (proposed_event_type IN ('ARRIVED', 'CHECK_IN', 'WORK_STARTED', 'BREAK_STARTED', 'BREAK_ENDED', 'WORK_RESUMED', 'OVERTIME_STARTED', 'OVERTIME_ENDED', 'CHECK_OUT')),
    proposed_timestamp TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL,
    evidence_file_id UUID,
    status VARCHAR(32) NOT NULL DEFAULT 'SUBMITTED' CHECK (status IN ('SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REJECTED')),
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 4. STORAGE & CLOUDFLARE R2 METADATA
-- ============================================================================

CREATE TABLE IF NOT EXISTS file_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    disk VARCHAR(32) NOT NULL DEFAULT 'r2',
    path_key VARCHAR(512) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(128) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign keys for files now that file_metadata exists
ALTER TABLE users ADD CONSTRAINT fk_user_avatar FOREIGN KEY (avatar_file_id) REFERENCES file_metadata(id) ON DELETE SET NULL;
ALTER TABLE leave_requests ADD CONSTRAINT fk_leave_evidence FOREIGN KEY (evidence_file_id) REFERENCES file_metadata(id) ON DELETE SET NULL;
ALTER TABLE attendance_corrections ADD CONSTRAINT fk_correction_evidence FOREIGN KEY (evidence_file_id) REFERENCES file_metadata(id) ON DELETE SET NULL;

-- ============================================================================
-- 5. PROJECTS, TASKS & CONTRIBUTION ENGINE
-- ============================================================================

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    visibility VARCHAR(32) NOT NULL DEFAULT 'INTERN_ONLY' CHECK (visibility IN ('INTERN_ONLY', 'PUBLIC', 'PRIVATE')),
    required_skills UUID[] DEFAULT '{}',
    capacity INTEGER NOT NULL DEFAULT 1 CHECK (capacity >= 1),
    accepted_count INTEGER NOT NULL DEFAULT 0 CHECK (accepted_count >= 0 AND accepted_count <= capacity),
    deadline TIMESTAMPTZ NOT NULL,
    bounty_pool NUMERIC(15, 2) NOT NULL DEFAULT 0.00 CHECK (bounty_pool >= 0),
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PUBLISHED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS project_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cover_letter TEXT,
    skill_match_percentage NUMERIC(5, 2) CHECK (skill_match_percentage BETWEEN 0 AND 100),
    status VARCHAR(32) NOT NULL DEFAULT 'APPLIED' CHECK (status IN ('APPLIED', 'UNDER_REVIEW', 'SHORTLISTED', 'ACCEPTED', 'REJECTED', 'WITHDRAWN')),
    applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_project_applicant UNIQUE(project_id, user_id)
);

CREATE TABLE IF NOT EXISTS project_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    project_role VARCHAR(32) NOT NULL DEFAULT 'MEMBER' CHECK (project_role IN ('OWNER', 'MANAGER', 'SUPERVISOR', 'MEMBER')),
    responsibility TEXT,
    planned_contribution_pct NUMERIC(5, 2) NOT NULL DEFAULT 0.00 CHECK (planned_contribution_pct BETWEEN 0 AND 100),
    actual_contribution_pct NUMERIC(5, 2) CHECK (actual_contribution_pct BETWEEN 0 AND 100),
    final_contribution_pct NUMERIC(5, 2) CHECK (final_contribution_pct BETWEEN 0 AND 100),
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_project_team_member UNIQUE(project_id, user_id)
);

CREATE TABLE IF NOT EXISTS milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    deadline TIMESTAMPTZ NOT NULL,
    weight_pct NUMERIC(5, 2) NOT NULL DEFAULT 0.00 CHECK (weight_pct BETWEEN 0 AND 100),
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    milestone_id UUID REFERENCES milestones(id) ON DELETE SET NULL,
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    estimated_hours NUMERIC(4, 1) NOT NULL DEFAULT 1.0 CHECK (estimated_hours >= 0),
    difficulty_weight INTEGER NOT NULL DEFAULT 1 CHECK (difficulty_weight BETWEEN 1 AND 5),
    priority VARCHAR(32) NOT NULL DEFAULT 'MEDIUM' CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    deadline TIMESTAMPTZ NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'TODO' CHECK (status IN ('TODO', 'IN_PROGRESS', 'IN_REVIEW', 'COMPLETED', 'BLOCKED', 'CANCELLED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE work_sessions ADD CONSTRAINT fk_work_session_task FOREIGN KEY (active_task_id) REFERENCES tasks(id) ON DELETE SET NULL;
ALTER TABLE overtime_requests ADD CONSTRAINT fk_overtime_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;
ALTER TABLE overtime_requests ADD CONSTRAINT fk_overtime_task FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS work_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    progress_percentage INTEGER NOT NULL CHECK (progress_percentage BETWEEN 0 AND 100),
    what_i_did TEXT NOT NULL,
    evidence_type VARCHAR(32) NOT NULL CHECK (evidence_type IN ('GIT_COMMIT', 'SCREENSHOT', 'URL', 'DOCUMENT', 'ATTACHMENT')),
    evidence_url_or_key VARCHAR(512) NOT NULL,
    problems TEXT,
    next_actions TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'SUBMITTED' CHECK (status IN ('SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'REVISION_REQUIRED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES work_reports(id) ON DELETE CASCADE,
    file_metadata_id UUID REFERENCES file_metadata(id) ON DELETE SET NULL,
    evidence_type VARCHAR(32) NOT NULL CHECK (evidence_type IN ('GIT_COMMIT', 'SCREENSHOT', 'URL', 'DOCUMENT', 'ATTACHMENT')),
    external_url VARCHAR(512),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 6. PERFORMANCE & GAMIFICATION
-- ============================================================================

CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_domain VARCHAR(64) NOT NULL, -- 'ATTENDANCE', 'XP', 'TAX', 'OVERTIME', 'RANK'
    name VARCHAR(255) NOT NULL,
    effective_date DATE NOT NULL,
    condition_rules JSONB NOT NULL DEFAULT '{}'::jsonb,
    action_definitions JSONB NOT NULL DEFAULT '{}'::jsonb,
    priority_order INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS xp_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID REFERENCES policies(id) ON DELETE SET NULL,
    event_trigger VARCHAR(64) UNIQUE NOT NULL,
    xp_value INTEGER NOT NULL, -- Positif atau negatif
    xp_scheme VARCHAR(32) NOT NULL CHECK (xp_scheme IN ('INTERNSHIP_XP', 'PROJECT_XP', 'ALUMNI_CONTRIBUTION')),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS xp_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    xp_rule_id UUID REFERENCES xp_rules(id) ON DELETE SET NULL,
    scheme VARCHAR(32) NOT NULL CHECK (scheme IN ('INTERNSHIP_XP', 'PROJECT_XP', 'ALUMNI_CONTRIBUTION')),
    points INTEGER NOT NULL,
    running_balance INTEGER NOT NULL,
    reference_event VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    badge_icon_url VARCHAR(512),
    reward_xp INTEGER NOT NULL DEFAULT 0 CHECK (reward_xp >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_achievement UNIQUE(user_id, achievement_id)
);

CREATE TABLE IF NOT EXISTS performance_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    evaluator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    batch_id UUID NOT NULL REFERENCES batches(id) ON DELETE RESTRICT,
    period_type VARCHAR(32) NOT NULL CHECK (period_type IN ('WEEKLY', 'MONTHLY', 'END_OF_BATCH', 'CUSTOM')),
    attendance_score NUMERIC(5, 2) NOT NULL CHECK (attendance_score BETWEEN 0 AND 100),
    task_delivery_score NUMERIC(5, 2) NOT NULL CHECK (task_delivery_score BETWEEN 0 AND 100),
    work_quality_score NUMERIC(5, 2) NOT NULL CHECK (work_quality_score BETWEEN 0 AND 100),
    supervisor_rubric_score NUMERIC(5, 2) NOT NULL CHECK (supervisor_rubric_score BETWEEN 0 AND 100),
    composite_performance_score NUMERIC(5, 2) NOT NULL CHECK (composite_performance_score BETWEEN 0 AND 100),
    is_top_performer BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 7. INCENTIVES, COMPENSATION, FINANCE & DOUBLE-ENTRY LEDGER
-- ============================================================================

CREATE TABLE IF NOT EXISTS rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rank_id UUID REFERENCES ranks(id) ON DELETE SET NULL,
    achievement_id UUID REFERENCES achievements(id) ON DELETE SET NULL,
    component_type VARCHAR(32) NOT NULL CHECK (component_type IN ('CASH', 'BONUS_XP', 'PHYSICAL_ITEM', 'BADGE', 'VOUCHER', 'PLATFORM_PRIVILEGE')),
    title VARCHAR(255) NOT NULL,
    monetary_value NUMERIC(15, 2) NOT NULL DEFAULT 0.00 CHECK (monetary_value >= 0),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reward_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reward_id UUID NOT NULL REFERENCES rewards(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'ISSUED' CHECK (status IN ('ISSUED', 'CLAIMED', 'PROCESSING', 'FULFILLED', 'REJECTED')),
    shipping_address TEXT,
    tracking_number VARCHAR(128),
    claimed_at TIMESTAMPTZ,
    fulfilled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00 CHECK (current_balance >= 0.00),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS batch_funds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID UNIQUE NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    total_accumulated NUMERIC(15, 2) NOT NULL DEFAULT 0.00 CHECK (total_accumulated >= 0),
    current_balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00 CHECK (current_balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS deduction_tax_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL CHECK (type IN ('INCOME_TAX', 'PROJECT_TAX', 'GRADUATION_TAX', 'WITHDRAWAL_TAX', 'ADMINISTRATIVE_FEE', 'PENALTY', 'OTHER_DEDUCTION')),
    rate_type VARCHAR(32) NOT NULL CHECK (rate_type IN ('PERCENTAGE', 'FIXED_AMOUNT')),
    rate_value NUMERIC(10, 4) NOT NULL CHECK (rate_value >= 0),
    calculation_basis VARCHAR(64) NOT NULL DEFAULT 'GROSS_AMOUNT',
    minimum_amount NUMERIC(15, 2),
    maximum_amount NUMERIC(15, 2),
    effective_date DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    source VARCHAR(32) NOT NULL CHECK (source IN ('PROJECT_BOUNTY', 'OVERTIME_BONUS', 'RANK_REWARD', 'TAX_DEDUCTION', 'BATCH_CONTRIBUTION', 'PAYOUT', 'OTHER')),
    reference_id VARCHAR(255) NOT NULL,
    gross_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    deduction_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    net_amount NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'REVERSED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_wallet_net CHECK (net_amount = gross_amount - deduction_amount)
);

CREATE TABLE IF NOT EXISTS payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    bank_code VARCHAR(32) NOT NULL,
    account_number VARCHAR(64) NOT NULL,
    account_holder_name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'REQUESTED' CHECK (status IN ('REQUESTED', 'APPROVED', 'PROCESSING', 'SETTLED', 'FAILED', 'REJECTED')),
    settlement_reference VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS financial_ledgers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    account_code VARCHAR(64) NOT NULL, -- '1001-CASH', '2001-LIABILITY-INTERN', '2002-TAX-PAYABLE', '3001-BATCH-FUND'
    direction VARCHAR(8) NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    reference_table VARCHAR(64) NOT NULL,
    reference_id UUID NOT NULL,
    narration TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 8. SYSTEM, AUDIT, CERTIFICATES & PORTFOLIOS
-- ============================================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_name VARCHAR(128) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id VARCHAR(128) NOT NULL,
    ip_address VARCHAR(64),
    user_agent TEXT,
    old_state JSONB,
    new_state JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    setting_key VARCHAR(128) UNIQUE NOT NULL,
    setting_value TEXT NOT NULL,
    is_encrypted BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    action_url VARCHAR(512),
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    certificate_number VARCHAR(128) UNIQUE NOT NULL,
    template_version VARCHAR(32) NOT NULL DEFAULT '1.0',
    signer_name VARCHAR(255) NOT NULL,
    signer_title VARCHAR(255) NOT NULL,
    file_metadata_id UUID REFERENCES file_metadata(id) ON DELETE RESTRICT,
    issued_at DATE NOT NULL,
    verification_hash VARCHAR(128) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE alumni ADD CONSTRAINT fk_alumni_certificate FOREIGN KEY (certificate_id) REFERENCES certificates(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS portfolios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    public_slug VARCHAR(128) UNIQUE NOT NULL,
    bio TEXT,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    featured_projects UUID[] DEFAULT '{}',
    custom_theme JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 9. PERFORMANCE INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_attendance_user_time ON attendance_event_logs(user_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_attendance_idempotency ON attendance_event_logs(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_work_sessions_user_date ON work_sessions(user_id, date);
CREATE INDEX IF NOT EXISTS idx_projects_status_vis ON projects(status, visibility);
CREATE INDEX IF NOT EXISTS idx_tasks_project_milestone ON tasks(project_id, milestone_id);
CREATE INDEX IF NOT EXISTS idx_xp_tx_user_scheme ON xp_transactions(user_id, scheme);
CREATE INDEX IF NOT EXISTS idx_financial_ledgers_tx ON financial_ledgers(transaction_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_time ON audit_logs(event_name, timestamp);
CREATE INDEX IF NOT EXISTS idx_policies_domain_active ON policies(policy_domain, is_active);

-- ============================================================================
-- 10. NATIVE POSTGRESQL IMMUTABILITY TRIGGERS (APPEND-ONLY ENFORCEMENT)
-- Rules: BR-023, BR-025, FR-047
-- ============================================================================

CREATE OR REPLACE FUNCTION enforce_immutable_records()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'DCISP Security & Integrity Violation: Direct UPDATE or DELETE on immutable table (%) is strictly forbidden by BR-023 / BR-025 / FR-047', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

-- 1. Immutable Attendance Events
DROP TRIGGER IF EXISTS trg_immutable_attendance ON attendance_event_logs;
CREATE TRIGGER trg_immutable_attendance
BEFORE UPDATE OR DELETE ON attendance_event_logs
FOR EACH ROW EXECUTE FUNCTION enforce_immutable_records();

-- 2. Immutable Financial Ledger
DROP TRIGGER IF EXISTS trg_immutable_financial_ledger ON financial_ledgers;
CREATE TRIGGER trg_immutable_financial_ledger
BEFORE UPDATE OR DELETE ON financial_ledgers
FOR EACH ROW EXECUTE FUNCTION enforce_immutable_records();

-- 3. Immutable Security Audit Logs
DROP TRIGGER IF EXISTS trg_immutable_audit_logs ON audit_logs;
CREATE TRIGGER trg_immutable_audit_logs
BEFORE UPDATE OR DELETE ON audit_logs
FOR EACH ROW EXECUTE FUNCTION enforce_immutable_records();

-- 4. Immutable XP Transactions
DROP TRIGGER IF EXISTS trg_immutable_xp_transactions ON xp_transactions;
CREATE TRIGGER trg_immutable_xp_transactions
BEFORE UPDATE OR DELETE ON xp_transactions
FOR EACH ROW EXECUTE FUNCTION enforce_immutable_records();
