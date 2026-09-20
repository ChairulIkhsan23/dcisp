-- Migration 000006 (Down): Menghapus tabel sesi refresh token
DROP INDEX IF EXISTS uq_refresh_sessions_token_hash;
DROP INDEX IF EXISTS idx_refresh_sessions_user;
DROP INDEX IF EXISTS idx_refresh_sessions_expiry;
DROP TABLE IF EXISTS refresh_sessions;
