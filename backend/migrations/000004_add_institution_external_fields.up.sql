-- Migration 000004: Menambahkan kolom integrasi eksternal API Indonesia pada tabel institutions
ALTER TABLE institutions
	ADD COLUMN IF NOT EXISTS external_id VARCHAR(128),
	ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'MANUAL',
	ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS uq_institutions_external_id ON institutions(external_id) WHERE external_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_institutions_source ON institutions(source);
