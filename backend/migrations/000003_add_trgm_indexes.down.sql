-- Migration 000003 (Down): Menghapus indeks trigram GIN
DROP INDEX IF EXISTS idx_users_name_trgm;
DROP INDEX IF EXISTS idx_projects_title_trgm;
