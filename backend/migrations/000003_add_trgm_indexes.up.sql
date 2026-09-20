-- Migration 000003: Menambahkan indeks trigram GIN untuk pencarian teks penuh pada nama pengguna dan judul proyek
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE INDEX IF NOT EXISTS idx_users_name_trgm ON users USING gin(full_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_projects_title_trgm ON projects USING gin(title gin_trgm_ops);
