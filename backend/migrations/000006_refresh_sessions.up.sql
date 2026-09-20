-- Migration 000006: Sesi refresh token terotasi dan dapat dicabut (SECURITY.md Section 1.2)
-- Mendukung refresh rotation (token lama diinvalidasi saat refresh) dan revocation
-- seketika oleh pengguna/admin tanpa menunggu masa kedaluwarsa JWT.
CREATE TABLE IF NOT EXISTS refresh_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(128) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_refresh_sessions_token_hash ON refresh_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_sessions_user ON refresh_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_sessions_expiry ON refresh_sessions(expires_at);
