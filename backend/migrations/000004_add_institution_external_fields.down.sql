-- Migration 000004 (Down): Menghapus kolom integrasi eksternal API Indonesia
DROP INDEX IF EXISTS uq_institutions_external_id;
DROP INDEX IF EXISTS idx_institutions_source;
ALTER TABLE institutions
	DROP COLUMN IF EXISTS synced_at,
	DROP COLUMN IF EXISTS source,
	DROP COLUMN IF EXISTS external_id;
