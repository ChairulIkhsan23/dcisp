# PROJECT EXECUTION ROADMAP — DCISP v1.0

Dokumen ini merupakan panduan eksekusi teknis yang komprehensif, terstruktur, dan berurutan untuk menyelesaikan platform **DCISP (Dagang Creative Intern Solutions Program)** berdasarkan analisis menyeluruh terhadap dokumentasi di `docs/` dan kondisi aktual codebase.

---

## 1. Executive Summary

**DCISP** adalah platform pengelolaan siklus hidup peserta magang berbasis proyek dengan integrasi gamifikasi bergaya RPG retro (Neo-Pixel Adventure Design System / NP-ADS). Stack teknologi utama yang ditetapkan dalam spesifikasi final adalah:
- **Backend:** Golang 1.23+ (Gin Framework), `jackc/pgx/v5`, `go-redis/v9`, `hibiken/asynq`
- **Frontend:** Next.js 15+ (App Router, React 19), Tailwind CSS, Shadcn UI, Zustand, TanStack Query v5
- **Database:** PostgreSQL 16+ (35+ tabel, trigger immutability)
- **Cache & Message Broker:** Redis 7+
- **Object Storage:** Cloudflare R2 (presigned URL)
- **Otentikasi:** JWT HS256 (Access 15m, Refresh 7d), Argon2id untuk hashing password

### Kondisi Aktual Repository saat ini (~12% Phase 1 Foundation):
1. **Dokumentasi (100%):** PRD, 9 BRD, 7 Guide teknis, dan 2 Design System spec telah lengkap dan menjadi acuan tunggal (*single source of truth*).
2. **Database Schema & Seed (95%):** Skema DDL mencakup 48 tabel dengan 4 trigger immutability dan data awal (10 roles, 8 scopes, ranks, XP rules, tax rules, super admin). Terdapat 2 gap minor (missing GIN trigram indexes dan divergensi nama skema XP).
3. **Backend Core (Identity Module, ~70%):** Endpoint login, token refresh, profile, dan roles telah berjalan. Namun, ditemukan 7 bug/kerentanan (termasuk kerentanan pertukaran token type dan string check hardcoded pada middleware RBAC). 8 domain modul lainnya masih kosong (hanya skema DB).
4. **Frontend (0%):** Masih berupa template dasar `create-next-app` dengan Tailwind CSS v4 default. Belum ada dependensi aplikasi utama (Zustand, TanStack Query, Zod, Shadcn UI), belum ada konfigurasi token NP-ADS, dan belum ada halaman/komponen kustom.
5. **Testing & QA (5%):** Hanya terdapat 2 file pengujian integrasi (`auth_test.go`, `rbac_test.go`) tanpa pembersihan data (*cleanup*).

---

## 2. Current Project Status

| Area | Status | Kondisi | Catatan |
|---|---|---|---|
| PRD / BRD Docs | `DONE` | 55 FR, 25 BR, 10 Personas, 14 OQ lengkap | Sumber kebenaran utama |
| Architecture Spec | `DONE` | FINAL spec (Go+Gin, Next.js 15) | Migrasi V1 ke FINAL terdokumentasi |
| Coding Standards | `DONE` | 7 panduan teknis tersedia | Standar penulisan kode & API |
| DB Migration Schema | `DONE` | 48 tabel, 4 trigger immutability, 11 indeks | Butuh migrasi trigram GIN tambahan |
| DB Seed Data | `DONE` | 10 role, 8 scope, 5 rank, 12 aturan XP, 2 aturan pajak | Sesuai PRD |
| Backend Config | `DONE` | Env loader, pgxpool, redis client | Fallback JWT secret hardcoded perlu diperbaiki |
| Backend Identity Module | `PARTIAL` | Login, refresh, profile, roles berjalan | Ada kerentanan token type & status non-aktif |
| Backend RBAC Middleware | `PARTIAL` | Evaluasi izin runtime dari DB | Scanner role di-hardcode (langgar BR-001) |
| Backend Audit Interceptor | `PARTIAL` | Asynchronous insert dasar | Status code respon belum tercatat, rentan data hilang saat shutdown |
| Backend Response Envelope | `DONE` | Standar JSON envelope | Belum ada helper HTTP 409 & 429 |
| Backend Password Security | `DONE` | Argon2id sesuai OWASP | Memory 64MB, 3 iterasi, 2 thread |
| Backend People Module | `TODO` | Skema DB siap, kode 0% | FR-002 s/d FR-006 |
| Backend Workforce Module | `TODO` | Skema DB siap, kode 0% | FR-007 s/d FR-015 |
| Backend Projects Module | `TODO` | Skema DB siap, kode 0% | FR-016 s/d FR-024 |
| Backend Performance Module| `TODO` | Skema DB siap, kode 0% | FR-025 s/d FR-030 |
| Backend Finance Module | `TODO` | Skema DB siap, kode 0% | FR-031 s/d FR-039 |
| Backend Documents Module | `TODO` | Skema DB siap, kode 0% | FR-040 s/d FR-044 |
| Backend System Module | `TODO` | Skema DB siap, kode 0% | FR-045 s/d FR-048 |
| Backend Intelligence Module| `TODO` | Belum ada skema analitik & kode | FR-049 s/d FR-050 |
| Backend Asynq Worker | `TODO` | Belum diimplementasikan | Diperlukan untuk proses async XP & notifikasi |
| Frontend Design Tokens | `TODO` | Masih styling default Next.js | Perlu integrasi token NP-ADS |
| Frontend Dependencies | `TODO` | Belum ada Zustand, TanStack Query, Zod | 6 dependensi kritis belum terpasang |
| Frontend Auth & Routing | `TODO` | Belum ada form login & guard | Route groups `(auth)` & `(dashboard)` belum dibuat |
| Frontend Portals & Pages | `TODO` | Hanya `app/page.tsx` default | My Day, Scanner, Command Center belum ada |
| Cloudflare R2 Integration | `TODO` | Env vars tersedia, kode belum ada | Diperlukan untuk presigned URL bukti & avatar |
| IoT Scanner Integration | `TODO` | Belum ada handler `terminal-tap` | Diperlukan untuk terminal presensi ESP32 |
| Reverse Proxy & Nginx | `TODO` | Belum dikonfigurasi | Kebutuhan deployment & SSE buffering |
| CI/CD Pipeline | `TODO` | Belum dikonfigurasi | GitHub Actions lint, test, build |
| Automated Test Suite | `PARTIAL` | 2 file integrasi auth/rbac | Butuh unit tests dan cleanup database |

---

## 3. Documentation & Requirement Mapping

| Requirement | Sumber Dokumen | Modul Backend/Frontend | Status | Gap Teridentifikasi |
|---|---|---|---|---|
| **FR-001** Dynamic RBAC | BRD-06, PRD §3.1 | `internal/middleware/dynamic_rbac.go` | `PARTIAL` | Isolasi scanner di-hardcode string; lookup DB belum di-cache ke Redis; scope context guard belum lengkap |
| **FR-002** Intern Lifecycle | BRD-01, PRD §3.2 | `internal/modules/people/` | `TODO` | State machine `APPLICANT` → `GRADUATED` belum dibuat |
| **FR-003** Alumni Retention | BRD-01, PRD §3.2 | `internal/modules/people/` | `TODO` | Isolasi XP alumni dan profil publik belum dibuat |
| **FR-004** Batch Management | BRD-01, PRD §3.2 | `internal/modules/people/` | `TODO` | Inisialisasi otomatis entitas Batch Fund belum ada di service |
| **FR-005** Institution Mgmt | BRD-01, PRD §3.2 | `internal/modules/people/` | `TODO` | CRUD institusi mitra dan pencegahan penghapusan |
| **FR-006** Skill Matrix | BRD-01, PRD §3.2 | `internal/modules/people/` | `TODO` | Master skill dan junction level kecakapan peserta |
| **FR-007** Work Schedule | BRD-02, PRD §3.3 | `internal/modules/attendance/` | `TODO` | Perhitungan toleransi keterlambatan (grace period) & kalender libur |
| **FR-008** Attendance Events | BRD-02, PRD §3.3 | `internal/modules/attendance/` | `TODO` | Endpoint `/terminal-tap`, Redis debounce 30 detik, idempotency SHA-256 |
| **FR-009** Work Session | BRD-02, PRD §3.3 | `internal/modules/attendance/` | `TODO` | Pencatatan durasi gross/active/idle terikat tugas |
| **FR-010** Break State Engine| BRD-02, PRD §3.3 | `internal/modules/attendance/` | `TODO` | Deteksi anomali early break & unauthorized break (penalti -2 XP) |
| **FR-011** Session Integrity | BRD-02, PRD §3.3 | `frontend/hooks/use-session-integrity` | `TODO` | Pemantauan `visibilityState` & idle >15 menit tanpa pemotongan XP otomatis |
| **FR-012** Overtime Workflow | BRD-02, PRD §3.3 | `internal/modules/attendance/` | `TODO` | Pengajuan, persetujuan supervisor, kompensasi jam aktual |
| **FR-013** Device Registry | BRD-02, PRD §3.3 | `internal/modules/attendance/` | `TODO` | Manajemen terminal scanner & otentikasi signature HMAC-SHA256 |
| **FR-014** Leave Management | BRD-02, PRD §3.3 | `internal/modules/attendance/` | `TODO` | Pengajuan izin/cuti dengan pembebasan penalti XP |
| **FR-015** Attendance Correct| BRD-02, PRD §3.3 | `internal/modules/attendance/` | `TODO` | Koreksi presensi append-only (rekod `MANUAL_CORRECTION`) |
| **FR-016** Project Market | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Tiga visibilitas (`INTERN_ONLY`, `PUBLIC`, `PRIVATE`) |
| **FR-017** Quota Locking | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Transaksi isolasi penguncian kuota atomik saat penuh |
| **FR-018** Skill Matching | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Skoring kecocokan skill peserta (advisory only, dilarang auto-reject) |
| **FR-019** Project Team | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Penetapan Planned Contribution % (total harus tepat 100%) |
| **FR-020** Milestones | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Pengelolaan bobot fase milestone proyek (sum = 100%) |
| **FR-021** Task Kanban | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Alur status tugas: `TODO` → `IN_PROGRESS` → `IN_REVIEW` → `COMPLETED` |
| **FR-022** Work Submission | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Form laporan kemajuan tugas dan validasi bukti kerja |
| **FR-023** Evidence Vault | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Integrasi presigned URL R2 dan metadata tipe bukti |
| **FR-024** Contribution Eng | BRD-03, PRD §3.4 | `internal/modules/projects/` | `TODO` | Rekonsiliasi 3 layer: Planned → Actual → Final (pengesahan supervisor) |
| **FR-025** XP Rules Engine | BRD-04, PRD §3.5 | `internal/modules/performance/`| `TODO` | 3 partisi XP terisolasi & pemrosesan penalti terkonfigurasi |
| **FR-026** Rank Progression | BRD-04, PRD §3.5 | `internal/modules/performance/`| `TODO` | Tier Novice s/d Grandmaster tanpa penurunan otomatis |
| **FR-027** Performance Eval | BRD-04, PRD §3.5 | `internal/modules/performance/`| `TODO` | Rubrik evaluasi supervisor (independen dari nilai XP) |
| **FR-028** Top Performer | BRD-04, PRD §3.5 | `internal/modules/performance/`| `TODO` | Algoritma penentuan tepat 1 pemenang per batch per periode |
| **FR-029** Achievements | BRD-04, PRD §3.5 | `internal/modules/performance/`| `TODO` | Sistem lencana pencapaian milestone peserta |
| **FR-030** Skill Growth | BRD-04, PRD §3.5 | `internal/modules/performance/`| `TODO` | Visualisasi radar perkembangan keahlian peserta |
| **FR-031** Rank Rewards | BRD-05, PRD §3.6 | `internal/modules/finance/` | `TODO` | Penerbitan reward multi-komponen saat naik rank |
| **FR-032** Bounty Distribute | BRD-05, PRD §3.6 | `internal/modules/finance/` | `TODO` | Kalkulasi: Gross = Pool × Final Contribution % |
| **FR-033** Reward Claims | BRD-05, PRD §3.6 | `internal/modules/finance/` | `TODO` | Pelacakan status klaim merchandise dan reward fisik |
| **FR-034** Deduction / Tax | BRD-05, PRD §3.7 | `internal/modules/finance/` | `TODO` | Evaluasi dinamis potongan PPh dan iuran kas bersama |
| **FR-035** Personal Wallet | BRD-05, PRD §3.7 | `internal/modules/finance/` | `TODO` | Pencatatan mutasi transparan itemized, saldo minimal Rp0 |
| **FR-036** Batch Fund | BRD-05, PRD §3.7 | `internal/modules/finance/` | `TODO` | Akumulasi dana kas perpisahan angkatan (terpisah dari pajak) |
| **FR-037** Financial Ledger | BRD-05, PRD §3.7 | `internal/modules/finance/` | `TODO` | Pembukuan double-entry immutable (ΣDebit = ΣCredit = Rp0) |
| **FR-038** Payout Engine | BRD-05, PRD §3.7 | `internal/modules/finance/` | `TODO` | Alur penarikan dana ke rekening bank / e-wallet |
| **FR-039** Financial Flow | BRD-05, PRD §3.7 | `internal/modules/finance/` | `TODO` | Orkestrasi atomik: Gross - Tax - Farewell = Net Distributable |
| **FR-040** ID Card Engine | BRD-07, PRD §3.8 | `internal/modules/documents/`| `TODO` | Pembuatan kartu tanda pengenal digital (NIM + QR + Foto) |
| **FR-041** Certificate Eng | BRD-07, PRD §3.8 | `internal/modules/documents/`| `TODO` | Sertifikat kelulusan digital ber-hash verifikasi publik |
| **FR-042** Portfolio Builder | BRD-07, PRD §3.8 | `internal/modules/documents/`| `TODO` | Kompilasi laman publik portofolio `/portfolio/:slug` |
| **FR-043** Report Generator | BRD-07, PRD §3.8 | `internal/modules/documents/`| `TODO` | Ekspor dokumen logbook aktivitas format PDF standar |
| **FR-044** R2 Integration | BRD-07, PRD §3.8 | `internal/modules/documents/`| `TODO` | Manajemen presigned URL PUT/GET Cloudflare R2 |
| **FR-045** Policy Engine | BRD-08, PRD §3.9 | `internal/modules/system/` | `TODO` | Mesin aturan bisnis runtime (JSONB) terpusat dengan cache Redis |
| **FR-046** Notifications | BRD-08, PRD §3.9 | `internal/modules/system/` | `TODO` | Notifikasi in-app berbasis event |
| **FR-047** Audit Trail | BRD-08, PRD §3.9 | `internal/modules/system/` | `PARTIAL` | Logging dasar ada; perlu kelengkapan data sebelum/sesudah dan penanganan shutdown |
| **FR-048** System Settings | BRD-08, PRD §3.9 | `internal/modules/system/` | `TODO` | Konfigurasi parameter sistem terpusat |
| **FR-049** Analytics | BRD-09, PRD §3.10| `internal/modules/intelligence/`| `TODO` | Agregasi metrik ketepatan waktu, integritas sesi, serapan anggaran |
| **FR-050** AI Insights | BRD-09, PRD §3.10| `internal/modules/intelligence/`| `TODO` | Wawasan kecerdasan buatan penasihat (advisory only) |
| **FR-051** Command Center | BRD-09, PRD §3.11| `frontend/app/(dashboard)/command-center/` | `TODO` | Dasbor eksekutif 4 kuadran (Today, Attention, Performance, Finance) |
| **FR-052** My Day Portal | BRD-09, PRD §3.11| `frontend/app/(dashboard)/my-day/` | `TODO` | Kokpit harian peserta: tombol sesi kerja, timer presisi, pemilih tugas, bar XP |
| **FR-053** Team Today | BRD-09, PRD §3.11| `frontend/app/(dashboard)/team-today/` | `TODO` | Grid status tim real-time supervisor (Working, Break, AFK) |
| **FR-054** Supervisor Hub | BRD-09, PRD §3.11| `frontend/app/(dashboard)/supervisor/` | `TODO` | Antrean persetujuan lembur, cuti, laporan kerja, dan koreksi |
| **FR-055** Dynamic Sidebar | BRD-09, PRD §3.11| `frontend/components/navigation/` | `TODO` | Navigasi menu dinamis terfilter izin peran (Scanner hanya lihat Scanner) |

---

## 4. Work Breakdown Structure (WBS)

### PHASE 1 — Foundation Hardening & Technical Debt Clearance
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-001 | 1 | Perbaiki kerentanan token type (tambahkan claim `type: access/refresh` dan validasi timbal balik) | `P0 — BLOCKER` | None | S | `TODO` |
| T-002 | 1 | Cegah refresh token untuk user non-aktif pada `FindByEmail` dan `FindByID` | `P0 — BLOCKER` | None | XS | `TODO` |
| T-003 | 1 | Hapus default fallback hardcoded JWT secret (wajib fail-hard jika kosong) | `P0 — BLOCKER` | None | XS | `TODO` |
| T-004 | 1 | Amankan type assertion `userIDVal.(uuid.UUID)` di `controller.go:73` | `P1 — CRITICAL` | None | XS | `TODO` |
| T-005 | 1 | Cegah kebocoran pesan error internal SQL ke client pada response handler | `P1 — CRITICAL` | None | S | `TODO` |
| T-006 | 1 | Perbaiki status body JSON readiness probe saat database/redis down | `P2 — HIGH` | None | XS | `TODO` |
| T-007 | 1 | Refactor dynamic RBAC middleware: hapus string check `SCANNER_OPERATOR`, gunakan evaluasi DB/cache | `P2 — HIGH` | None | M | `TODO` |
| T-008 | 1 | Ganti `context.Background()` menjadi `c.Request.Context()` pada middleware RBAC | `P2 — HIGH` | None | XS | `TODO` |
| T-009 | 1 | Implementasikan caching izin RBAC pada Redis dengan TTL 5 menit | `P2 — HIGH` | None | M | `TODO` |
| T-010 | 1 | Perbaiki `audit_interceptor`: catat status code HTTP dan cegah kehilangan data saat shutdown | `P2 — HIGH` | None | M | `TODO` |
| T-011 | 1 | Bungkus bare errors pada `service.go` menggunakan `fmt.Errorf` bahasa Indonesia | `P3 — MEDIUM` | None | XS | `TODO` |
| T-012 | 1 | Dukung multi-role assignment pada query repository (hapus `LIMIT 1`) | `P3 — MEDIUM` | None | S | `TODO` |
| T-013 | 1 | Tambahkan helper respon `response.Conflict` (409) dan `response.TooManyRequests` (429) | `P3 — MEDIUM` | None | XS | `TODO` |
| T-014 | 1 | Buat file migrasi `000003` untuk 2 indeks trigram GIN (`users.full_name`, `projects.title`) | `P3 — MEDIUM` | None | XS | `TODO` |
| T-015 | 1 | Koreksi versi Go pada `go.mod` menjadi versi valid (1.23) | `P3 — MEDIUM` | None | XS | `TODO` |
| T-016 | 1 | Tambahkan struct tags `db:` eksplisit pada seluruh struct di `modules/identity/model.go` | `P4 — LOW` | None | XS | `TODO` |
| T-017 | 1 | Tambahkan komentar satu kalimat bahasa Indonesia pada seluruh fungsi di `tests/` | `P4 — LOW` | None | XS | `TODO` |
| T-018 | 1 | Tambahkan mekanisme pembersihan data uji (`t.Cleanup` / rollback transaksi) pada pengujian | `P2 — HIGH` | None | S | `TODO` |

### PHASE 2 — Core Infrastructure & Platform Services
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-019 | 2 | Implementasikan modul Unified Policy Engine (FR-045): CRUD aturan JSONB & cache Redis | `P0 — BLOCKER` | T-009 | L | `TODO` |
| T-020 | 2 | Implementasikan service Cloudflare R2 (FR-044): generator presigned URL PUT/GET & metadata | `P1 — CRITICAL` | None | M | `TODO` |
| T-021 | 2 | Setup infrastruktur background worker menggunakan `hibiken/asynq` terhubung ke Redis | `P1 — CRITICAL` | None | M | `TODO` |
| T-022 | 2 | Buat internal async event bus untuk dispatch domain events antar modul | `P1 — CRITICAL` | T-021 | S | `TODO` |
| T-023 | 2 | Implementasikan middleware rate limiting berbasis sliding window Redis (120 req/m, 10 req/m auth) | `P2 — HIGH` | T-009 | M | `TODO` |
| T-024 | 2 | Buat endpoint Server-Sent Events (SSE) `/api/v1/events/stream` untuk notifikasi real-time | `P2 — HIGH` | None | M | `TODO` |
| T-025 | 2 | Bangun layanan Audit Logging terpusat (FR-047) dengan pencatatan mutasi state JSONB | `P2 — HIGH` | T-010 | M | `TODO` |
| T-026 | 2 | Implementasikan CRUD pengaturan sistem terpusat (FR-048) | `P3 — MEDIUM` | None | S | `TODO` |

### PHASE 3 — People & Lifecycle Backend
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-027 | 3 | Bangun modul Intern Lifecycle (FR-002): state machine 7 status dan validasi transisi | `P0 — BLOCKER` | T-019 | L | `TODO` |
| T-028 | 3 | Bangun modul Batch Management (FR-004): CRUD kohort, kuota, dan auto-inisialisasi Batch Fund | `P0 — BLOCKER` | T-027 | M | `TODO` |
| T-029 | 3 | Bangun modul Alumni Lifecycle (FR-003): retensi akun permanen dan isolasi skema poin | `P1 — CRITICAL` | T-027 | M | `TODO` |
| T-030 | 3 | Bangun modul Institusi Mitra (FR-005): CRUD universitas/sekolah dan guard integritas data | `P2 — HIGH` | None | S | `TODO` |
| T-031 | 3 | Bangun modul Skill Matrix (FR-006): katalog keahlian teknis & penetapan level profisiensi | `P2 — HIGH` | T-027 | S | `TODO` |
| T-032 | 3 | Bangun endpoint User Management: pencarian peserta, penugasan pembimbing, dan plotting batch | `P1 — CRITICAL` | T-027, T-028 | M | `TODO` |

### PHASE 4 — Workforce & Attendance Backend
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-033 | 4 | Bangun Work Schedule Engine (FR-007): konfigurasi jam kerja, hari libur, dan toleransi | `P0 — BLOCKER` | T-019 | M | `TODO` |
| T-034 | 4 | Bangun Ingesti Presensi `/attendance/terminal-tap` (FR-008): debounce 30s, idempotency, 9 event | `P0 — BLOCKER` | T-033, T-023 | XL | `TODO` |
| T-035 | 4 | Implementasikan klasifikasi keterlambatan 3 tier berdasarkan jadwal kerja dinamis | `P0 — BLOCKER` | T-033, T-034 | M | `TODO` |
| T-036 | 4 | Bangun Work Session Tracking (FR-009): siklus kerja, durasi gross/active/idle | `P0 — BLOCKER` | T-034 | L | `TODO` |
| T-037 | 4 | Bangun Break State Engine (FR-010): anomali early break dan sanksi unauthorized break | `P0 — BLOCKER` | T-036, T-019 | M | `TODO` |
| T-038 | 4 | Bangun alur Overtime (FR-012): pengajuan lembur, verifikasi supervisor, hitung jam aktual | `P1 — CRITICAL` | T-036 | M | `TODO` |
| T-039 | 4 | Bangun modul Cuti (FR-014): pengajuan cuti sakit/akademik, persetujuan, pembebasan penalti | `P1 — CRITICAL` | T-027 | M | `TODO` |
| T-040 | 4 | Bangun Koreksi Presensi (FR-015): rekod koreksi append-only tanpa mengubah log asli | `P1 — CRITICAL` | T-034 | S | `TODO` |
| T-041 | 4 | Bangun Device Registry (FR-013): otentikasi terminal scanner via HMAC-SHA256 & telemetry | `P2 — HIGH` | None | S | `TODO` |
| T-042 | 4 | Bangun endpoint sinkronisasi offline presensi `/attendance/sync-offline` (ON CONFLICT DO NOTHING) | `P2 — HIGH` | T-034 | M | `TODO` |
| T-043 | 4 | Bangun endpoint QR Dinamis dengan masa berlaku TTL 30 detik | `P2 — HIGH` | T-034 | S | `TODO` |
| T-044 | 4 | Buat endpoint `/attendance/audio-catalog` untuk manifest audio 15 event terminal ESP32 | `P3 — MEDIUM` | T-020 | S | `TODO` |

### PHASE 5 — Projects & Tasks Backend
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-045 | 5 | Bangun Project Marketplace (FR-016): CRUD proyek, 3 visibilitas, filter publik/magang | `P0 — BLOCKER` | T-027, T-031 | L | `TODO` |
| T-046 | 5 | Bangun pendaftaran proyek & kuota (FR-017): transaksi isolasi SERIALIZABLE untuk lock kuota | `P0 — BLOCKER` | T-045 | M | `TODO` |
| T-047 | 5 | Bangun pembentukan tim proyek (FR-019): penetapan peran dan Planned Contribution % (sum=100%) | `P0 — BLOCKER` | T-046 | M | `TODO` |
| T-048 | 5 | Bangun Milestone Management (FR-020): fase proyek dengan validasi total bobot 100% | `P1 — CRITICAL` | T-045 | S | `TODO` |
| T-049 | 5 | Bangun Task Kanban Engine (FR-021): siklus status tugas dan pembobotan kesulitan | `P0 — BLOCKER` | T-047, T-048 | L | `TODO` |
| T-050 | 5 | Bangun Work Report & Submission (FR-022): pelaporan tugas dan alur review pengesahan | `P0 — BLOCKER` | T-049 | M | `TODO` |
| T-051 | 5 | Bangun Evidence Vault (FR-023): validasi tautan git dan penyimpanan metadata bukti ke R2 | `P0 — BLOCKER` | T-050, T-020 | M | `TODO` |
| T-052 | 5 | Bangun Three-Layer Contribution Engine (FR-024): Planned → Actual → Final % (pengesahan supervisor) | `P0 — BLOCKER` | T-047 | L | `TODO` |
| T-053 | 5 | Bangun Skill Matching Engine (FR-018): rekomendasi kecocokan pelamar (advisory only) | `P2 — HIGH` | T-031, T-045 | M | `TODO` |

### PHASE 6 — Performance & Gamification Backend
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-054 | 6 | Bangun XP Rules Engine (FR-025): evaluasi 3 skema XP terisolasi & mutasi append-only | `P0 — BLOCKER` | T-019, T-034 | L | `TODO` |
| T-055 | 6 | Hubungkan event presensi ke worker Asynq untuk perhitungan reward/penalti XP async | `P0 — BLOCKER` | T-054, T-021 | L | `TODO` |
| T-056 | 6 | Bangun Rank Progression (FR-026): evaluasi threshold XP, kenaikan pangkat tanpa auto-demotion | `P0 — BLOCKER` | T-054 | M | `TODO` |
| T-057 | 6 | Hubungkan penyelesaian tugas dan milestone ke perhitungan Project XP | `P1 — CRITICAL` | T-054, T-050 | M | `TODO` |
| T-058 | 6 | Bangun evaluasi performa supervisor (FR-027): rubrik penilaian berkala 0–100 (terpisah dari XP) | `P1 — CRITICAL` | T-027 | M | `TODO` |
| T-059 | 6 | Bangun penentuan Top Performer per Batch (FR-028): algoritma tunggal per periode | `P2 — HIGH` | T-058, T-028 | M | `TODO` |
| T-060 | 6 | Bangun Achievement & Badge System (FR-029): evaluasi syarat unlock lencana | `P2 — HIGH` | T-054 | M | `TODO` |
| T-061 | 6 | Bangun agregasi Skill Growth Matrix (FR-030): histori perkembangan keahlian peserta | `P3 — MEDIUM` | T-031, T-057 | S | `TODO` |

### PHASE 7 — Finance & Compensation Backend
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-062 | 7 | Bangun Deduction & Tax Engine (FR-034): evaluasi dinamis aturan PPh & potongan kas bersama | `P0 — BLOCKER` | T-019 | L | `TODO` |
| T-063 | 7 | Bangun Personal Wallet (FR-035): pembuatan dompet otomatis, mutasi itemized, saldo minimal Rp0 | `P0 — BLOCKER` | T-027 | M | `TODO` |
| T-064 | 7 | Bangun Financial Ledger (FR-037): double-entry immutable (ΣDebit = ΣCredit = Rp0) & jurnal pembalik | `P0 — BLOCKER` | T-063 | L | `TODO` |
| T-065 | 7 | Bangun distribusi bounty proyek (FR-032): alokasi pool berdasarkan Final Contribution % | `P0 — BLOCKER` | T-052, T-062, T-064 | XL | `TODO` |
| T-066 | 7 | Bangun Financial Flow Orchestration (FR-039): atomik Gross - Tax - Farewell = Net Distributable | `P0 — BLOCKER` | T-065 | L | `TODO` |
| T-067 | 7 | Bangun Multi-Component Rank Rewards (FR-031): penerbitan reward saat naik pangkat | `P1 — CRITICAL` | T-056, T-063 | M | `TODO` |
| T-068 | 7 | Bangun pengelolaan kas angkatan / Batch Fund (FR-036): akumulasi terpisah dari rekening pajak | `P1 — CRITICAL` | T-028, T-064 | M | `TODO` |
| T-069 | 7 | Bangun siklus klaim reward fisik / merchandise (FR-033): alur permohonan hingga pengiriman | `P2 — HIGH` | T-067 | S | `TODO` |
| T-070 | 7 | Bangun Payout Engine (FR-038): pengajuan penarikan dana dompet ke rekening bank | `P2 — HIGH` | T-063, T-064 | M | `TODO` |

### PHASE 8 — Frontend Foundation & Design System Setup
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-071 | 8 | Instalasi dependensi frontend: Zustand, TanStack Query v5, react-hook-form, Zod, Shadcn UI | `P0 — BLOCKER` | None | S | `TODO` |
| T-072 | 8 | Konfigurasi design tokens NP-ADS di Tailwind CSS: warna retro, hard drop shadow, border tebal | `P0 — BLOCKER` | T-071 | M | `TODO` |
| T-073 | 8 | Konfigurasi font retro: Press Start 2P, Silkscreen, dan Mulish via `next/font/google` | `P0 — BLOCKER` | None | S | `TODO` |
| T-074 | 8 | Buat arsitektur folder App Router: `(auth)`, `(dashboard)`, `components/ui`, `hooks`, `lib` | `P0 — BLOCKER` | None | S | `TODO` |
| T-075 | 8 | Bangun HTTP API client (`lib/api-client.ts`): auto Bearer token, auto refresh pada 401, replay | `P0 — BLOCKER` | T-071 | M | `TODO` |
| T-076 | 8 | Bangun auth store terpusat menggunakan Zustand: persistensi token & profil pengguna | `P0 — BLOCKER` | T-075 | M | `TODO` |
| T-077 | 8 | Pasang TanStack QueryClientProvider dan state provider pada root layout | `P0 — BLOCKER` | T-071, T-076 | S | `TODO` |
| T-078 | 8 | Bangun komponen UI primitif Neo-Brutalist: NeoArcadeButton, NeoPixelCard, PixelProgressBar, NesDialog | `P1 — CRITICAL` | T-072 | L | `TODO` |
| T-079 | 8 | Bangun sidebar navigasi dinamis terfilter izin peran / RBAC (FR-055) | `P1 — CRITICAL` | T-076 | M | `TODO` |
| T-080 | 8 | Implementasikan middleware Next.js untuk proteksi route dan pengalihan isolasi Scanner Operator | `P1 — CRITICAL` | T-076 | S | `TODO` |

### PHASE 9 — Frontend Auth & Identity Pages
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-081 | 9 | Bangun halaman Login `(auth)/login`: form bertema NES Dialog, validasi Zod, integrasi auth store | `P0 — BLOCKER` | T-078, T-076 | M | `TODO` |
| T-082 | 9 | Bangun halaman profil pengguna: tampilan biodata, kartu identitas adventurer, dan izin aktif | `P2 — HIGH` | T-081 | S | `TODO` |
| T-083 | 9 | Bangun antarmuka manajemen role & permission untuk Super Admin | `P2 — HIGH` | T-081 | M | `TODO` |

### PHASE 10 — Frontend People & Workforce Pages
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-084 | 10 | Bangun halaman daftar & pencarian peserta magang dengan filter status dan batch | `P1 — CRITICAL` | T-078, T-032 | M | `TODO` |
| T-085 | 10 | Bangun laman profil detail peserta magang: status siklus hidup, pembimbing, XP, dan skill | `P1 — CRITICAL` | T-084 | M | `TODO` |
| T-086 | 10 | Bangun antarmuka pengelolaan batch: pembuatan kohort, pemantauan kuota, dan daftar anggota | `P1 — CRITICAL` | T-028 | M | `TODO` |
| T-087 | 10 | Bangun My Day Portal (FR-052): kokpit harian lengkap, tombol sesi kerja, pemilih tugas, bar XP | `P0 — BLOCKER` | T-036, T-078 | XL | `TODO` |
| T-088 | 10 | Bangun hook `useWorkSession`: timer presisi berbasis Web Worker dengan kalibrasi server | `P0 — BLOCKER` | T-076 | L | `TODO` |
| T-089 | 10 | Bangun hook `useSessionIntegrity`: deteksi tab blur / visibility dan status idle >15m | `P1 — CRITICAL` | T-076 | M | `TODO` |
| T-090 | 10 | Bangun antarmuka terminal scanner `/attendance/scanner`: tampilan tap NFC/QR dan respon audio | `P1 — CRITICAL` | T-034, T-078 | L | `TODO` |
| T-091 | 10 | Bangun portal Team Today (FR-053): monitoring status tim real-time supervisor via SSE | `P1 — CRITICAL` | T-024, T-036 | L | `TODO` |
| T-092 | 10 | Bangun antarmuka pengajuan dan persetujuan lembur | `P2 — HIGH` | T-038 | M | `TODO` |
| T-093 | 10 | Bangun antarmuka pengajuan dan persetujuan cuti / izin akademik | `P2 — HIGH` | T-039 | M | `TODO` |
| T-094 | 10 | Bangun antarmuka pengajuan koreksi presensi dan approval supervisor | `P2 — HIGH` | T-040 | S | `TODO` |

### PHASE 11 — Frontend Projects & Tasks Pages
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-095 | 11 | Bangun Project Marketplace (Bounty Board): kartu proyek bertema quest RPG dan filter visibilitas | `P1 — CRITICAL` | T-045, T-078 | L | `TODO` |
| T-096 | 11 | Bangun halaman detail proyek: ringkasan, alokasi tim, milestone, dan daftar tugas | `P1 — CRITICAL` | T-095 | L | `TODO` |
| T-097 | 11 | Bangun Task Kanban Board: drag-and-drop antar kolom status tugas (TODO s/d COMPLETED) | `P1 — CRITICAL` | T-049, T-078 | L | `TODO` |
| T-098 | 11 | Bangun form penyerahan tugas (Work Report) dengan upload bukti presigned R2 | `P1 — CRITICAL` | T-050, T-020 | M | `TODO` |
| T-099 | 11 | Bangun dasbor rekonsiliasi kontribusi 3 layer: Planned vs Actual vs Final % | `P2 — HIGH` | T-052 | M | `TODO` |
| T-100 | 11 | Bangun dialog pendaftaran proyek dan pelacakan status lamaran peserta | `P2 — HIGH` | T-046 | M | `TODO` |

### PHASE 12 — Frontend Performance & Finance Pages
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-101 | 12 | Bangun dasbor XP & Rank: visualisasi 3 partisi XP, badge level Novice s/d Grandmaster | `P1 — CRITICAL` | T-054, T-078 | M | `TODO` |
| T-102 | 12 | Bangun Leaderboard Hall of Heroes: peringkat peserta non-toksik per batch | `P2 — HIGH` | T-054 | M | `TODO` |
| T-103 | 12 | Bangun formulir evaluasi rubrik supervisor berkala | `P2 — HIGH` | T-058 | M | `TODO` |
| T-104 | 12 | Bangun dasbor Dompet Peserta (Coin Pouch): saldo riil, histori transaksi, form pencairan | `P1 — CRITICAL` | T-063, T-078 | M | `TODO` |
| T-105 | 12 | Bangun antarmuka audit pembukuan keuangan (Finance Ledger) untuk peran Finance | `P2 — HIGH` | T-064 | M | `TODO` |
| T-106 | 12 | Bangun antarmuka klaim reward fisik dan pelacakan resi | `P2 — HIGH` | T-069 | S | `TODO` |

### PHASE 13 — Frontend Executive Portals
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-107 | 13 | Bangun Command Center (FR-051): dasbor eksekutif 4 kuadran (Today, Attention, Performance, Finance)| `P1 — CRITICAL` | T-091, T-104 | XL | `TODO` |
| T-108 | 13 | Bangun Supervisor Workspace (FR-054): hub terpusat persetujuan laporan, izin, dan lembur | `P1 — CRITICAL` | T-092, T-093, T-094 | L | `TODO` |
| T-109 | 13 | Bangun Notification Center (FR-046): lonceng notifikasi, toast alert, dan mark-as-read | `P2 — HIGH` | T-024 | M | `TODO` |

### PHASE 14 — Documents & Digital Assets
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-110 | 14 | Bangun generator kartu tanda pengenal digital / Adventurer License (FR-040) | `P2 — HIGH` | T-020 | M | `TODO` |
| T-111 | 14 | Bangun mesin sertifikat kelulusan digital (FR-041) dan verifikasi publik (Menunggu OQ-001, OQ-002)| `P2 — HIGH` | T-020 | M | `BLOCKED` |
| T-112 | 14 | Bangun generator laman portofolio publik `/portfolio/:slug` (FR-042) | `P2 — HIGH` | T-020 | L | `TODO` |
| T-113 | 14 | Bangun generator ekspor PDF logbook aktivitas harian peserta (FR-043) | `P3 — MEDIUM` | T-020 | M | `TODO` |

### PHASE 15 — Testing & Quality Assurance
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-114 | 15 | Tulis unit test komprehensif backend modul Identity & Auth (target coverage ≥85%) | `P1 — CRITICAL` | T-001 s/d T-018 | M | `TODO` |
| T-115 | 15 | Tulis unit test backend modul People (state machine intern, kuota batch) | `P1 — CRITICAL` | T-027, T-028 | M | `TODO` |
| T-116 | 15 | Tulis integration test presensi: 11 skenario kritis dari `docs/guide/TESTING.md` | `P0 — BLOCKER` | T-034 | L | `TODO` |
| T-117 | 15 | Tulis integration test transaksi keuangan: verifikasi keseimbangan ΣDebit = ΣCredit | `P0 — BLOCKER` | T-064 | M | `TODO` |
| T-118 | 15 | Tulis test verifikasi trigger immutability pada 4 tabel append-only | `P1 — CRITICAL` | T-014 | S | `TODO` |
| T-119 | 15 | Tulis test integritas pembagian kontribusi proyek: sum(final_contribution_pct) == 100.00% | `P1 — CRITICAL` | T-052 | S | `TODO` |
| T-120 | 15 | Tulis test partisi 3 skema XP dan promosi rank tanpa auto-demotion | `P1 — CRITICAL` | T-054 | M | `TODO` |
| T-121 | 15 | Tulis component test frontend menggunakan Vitest dan React Testing Library | `P2 — HIGH` | T-078 | M | `TODO` |
| T-122 | 15 | Tulis end-to-end (E2E) test Playwright untuk alur kritis: Login → My Day → Presensi → Sesi Kerja | `P2 — HIGH` | T-087 | L | `TODO` |

### PHASE 16 — Security & Performance Hardening
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-123 | 16 | Konfigurasi CORS produksi yang ketat berdasarkan whitelist domain terpercaya | `P1 — CRITICAL` | None | S | `TODO` |
| T-124 | 16 | Pasang security headers middleware: CSP, HSTS, X-Frame-Options: DENY, nosniff | `P1 — CRITICAL` | None | S | `TODO` |
| T-125 | 16 | Pasang pembatasan ukuran payload request HTTP (max body size) | `P2 — HIGH` | None | XS | `TODO` |
| T-126 | 16 | Tambahkan validasi kompleksitas kata sandi pada saat pendaftaran pengguna | `P2 — HIGH` | None | S | `TODO` |
| T-127 | 16 | Tambahkan logging telemetri query lambat (>500ms) pada PostgreSQL pool | `P3 — MEDIUM` | None | S | `TODO` |
| T-128 | 16 | Optimasi alur endpoint presensi untuk mencapai SLA ≤200ms pada beban 50 tap/detik | `P2 — HIGH` | T-034, T-055 | M | `TODO` |

### PHASE 17 — Deployment & Production Readiness
| ID | Phase | Task | Priority | Dependency | Complexity | Status |
|---|---|---|---|---|---|---|
| T-129 | 17 | Siapkan konfigurasi reverse proxy Nginx: terminasi TLS 1.3, gzip, SSE proxy buffering off | `P1 — CRITICAL` | None | M | `TODO` |
| T-130 | 17 | Buat `Dockerfile` multi-stage build untuk backend Go (ukuran image minimal) | `P1 — CRITICAL` | None | S | `TODO` |
| T-131 | 17 | Buat `Dockerfile` untuk frontend Next.js menggunakan mode standalone | `P1 — CRITICAL` | None | S | `TODO` |
| T-132 | 17 | Perbarui `docker-compose.yml` untuk memuat layanan backend, frontend, dan Nginx | `P1 — CRITICAL` | T-129 s/d T-131 | M | `TODO` |
| T-133 | 17 | Siapkan workflow GitHub Actions untuk linting, type-checking, automated testing, dan build | `P1 — CRITICAL` | None | M | `TODO` |
| T-134 | 17 | Buat skrip pencadangan otomatis harian database (pg_dump) dan pengarsipan log WAL 15 menit | `P2 — HIGH` | None | M | `TODO` |
| T-135 | 17 | Siapkan template `.env.production` yang terdokumentasi lengkap dan aman | `P2 — HIGH` | None | S | `TODO` |

---

## 5. Detailed Task Breakdown (Core Critical Tasks)

### T-001: Fix JWT Token Type Vulnerability
- **Objective:** Mencegah access token digunakan sebagai refresh token dan sebaliknya.
- **Requirement:** `docs/guide/SECURITY.md` §1.2, `docs/brd/06-BRD-IDENTITY-RBAC-GOVERNANCE.md`.
- **Scope:** `backend/internal/shared/utils/token.go`, `backend/internal/middleware/auth_jwt.go`, `backend/internal/modules/identity/service.go`.
- **Dependency:** None.
- **Implementation Expectation:**
  - Tambahkan claim `token_type` (nilai `"access"` atau `"refresh"`) ke dalam struct `JWTClaims`.
  - Fungsi `ValidateToken` menerima parameter `expectedType string` dan menolak token jika tipe tidak sesuai.
  - `AuthJWT` middleware secara eksplisit meminta tipe `"access"`.
  - Service `RefreshToken` secara eksplisit meminta tipe `"refresh"`.
- **Expected Output:** File token dan middleware terbarui dengan proteksi tipe token.
- **Acceptance Criteria:**
  - Permintaan ke endpoint terproteksi menggunakan refresh token menghasilkan respon `401 Unauthorized`.
  - Permintaan ke `/api/v1/auth/refresh` menggunakan access token menghasilkan respon `401 Unauthorized`.
- **Risk:** Token aktif yang sudah diterbitkan akan menjadi tidak valid (dapat diterima karena sistem belum live di produksi).

### T-019: Implement Unified Policy Engine (FR-045)
- **Objective:** Menyediakan mesin konfigurasi aturan bisnis terpusat saat runtime tanpa perlu deploy ulang kode.
- **Requirement:** FR-045, BR-001, BR-012, BR-020, `docs/brd/08-BRD-SYSTEM-POLICY-AUDIT-INTEGRITY.md`.
- **Scope:** `backend/internal/modules/system/policy_*` (model, repository, service, controller).
- **Dependency:** T-009 (Redis caching).
- **Implementation Expectation:**
  - Menyediakan endpoint CRUD untuk entitas `policies` (menyimpan rules dalam kolom `JSONB`).
  - Mengimplementasikan fungsi evaluasi `EvaluatePolicy(domain string, context map[string]any) (RuleResult, error)`.
  - Menyimpan cache hasil evaluasi di Redis (TTL 5–15 menit) dengan mekanisme invalidasi event-driven saat policy diperbarui.
- **Expected Output:** Modul system policy yang dapat di-inject ke modul presensi, XP, dan finansial.
- **Acceptance Criteria:**
  - Perubahan aturan toleransi keterlambatan pada policy via API langsung mengubah respon klasifikasi presensi tanpa restart server.
- **Risk:** Kompleksitas parsing query JSONB; mitigasi awal dengan struktur predikat terdefinisi jelas (*key-value rules*).

### T-034: Implement Attendance Event Ingestion `/terminal-tap` (FR-008)
- **Objective:** Endpoint berkinerja tinggi untuk memproses tap NFC kartu atau pemindaian QR dari terminal IoT.
- **Requirement:** FR-008, BR-002, BR-006, BR-007, BR-025, `docs/guide/API.md`, `docs/architecture/FINAL-TECH-STACK-SPEC-DCISP.md`.
- **Scope:** `backend/internal/modules/attendance/` (controller, service, repository).
- **Dependency:** T-033 (Work Schedule), T-023 (Rate Limiting).
- **Implementation Expectation:**
  - Endpoint `POST /api/v1/attendance/terminal-tap` menerima `device_id`, `card_uid` atau `qr_token`, dan `timestamp`.
  - Debounce 30 detik menggunakan Redis atomic key `attendance:debounce:<user_id>`. Jika duplikat, kembalikan status `DEBOUNCED` dan audio event `too_frequent`.
  - Generate idempotency key `SHA-256(device_id + user_id + event_type + timestamp_minute)`.
  - Simpan event secara immutable ke tabel `attendance_event_logs`.
  - Emit domain event ke Asynq worker untuk perhitungan penalti/reward XP asinkron.
  - Kembalikan nama berkas audio event (dari katalog 15 audio) untuk dimainkan speaker ESP32.
- **Expected Output:** Endpoint presensi yang memenuhi SLA ≤200ms.
- **Acceptance Criteria:**
  - Lulus 11 skenario pengujian presensi pada `docs/guide/TESTING.md`.
  - Upaya duplikasi tap dalam 30 detik tidak menghasilkan baris baru di database.
- **Risk:** Latensi Redis memengaruhi respon terminal; mitigasi dengan koneksi pool Redis persisten dan query non-blocking.

### T-065: Implement Project Bounty Distribution (FR-032)
- **Objective:** Mendistribusikan dana bounty proyek kepada seluruh anggota tim secara adil berdasarkan kontribusi final.
- **Requirement:** FR-032, BR-016, BR-020, BR-021, BR-022, BR-023, `docs/brd/05-BRD-FINANCE-COMPENSATION.md`.
- **Scope:** `backend/internal/modules/finance/bounty_service.go`.
- **Dependency:** T-052 (Three-Layer Contribution), T-062 (Tax Engine), T-064 (Financial Ledger).
- **Implementation Expectation:**
  - Endpoint `POST /api/v1/finance/bounty/distribute`.
  - Hitung `Gross Bounty = Bounty Pool × Final Contribution %` untuk setiap anggota tim.
  - Alirkan ke Deduction Engine untuk menghitung PPh 21 dan Iuran Kas Bersama (Batch Fund).
  - Hitung `Net Distributable = Gross - Tax - Farewell`.
  - Seluruh mutasi dibungkus dalam satu transaksi database PostgreSQL ACID.
  - Masukkan mutasi ke `wallet_transactions` dan catat entri berpasangan di `financial_ledgers` (ΣDebit = ΣCredit).
- **Expected Output:** Layanan distribusi dana proyek otomatis dan transparan.
- **Acceptance Criteria:**
  - Total debet sama dengan total kredit pada ledger (selisih Rp 0,00).
  - Saldo dompet penerima bertambah sesuai nilai bersih.
  - Jika ada kesalahan perhitungan pada salah satu anggota, seluruh mutasi dibatalkan (*rollback*).
- **Risk:** Pembulatan angka pecahan; mitigasi dengan tipe data `NUMERIC(15,2)` dan pembulatan sen ke kas operasional jika ada sisa pecahan.

### T-087: Build My Day Portal (FR-052)
- **Objective:** Membangun antarmuka kokpit harian tunggal bagi peserta magang untuk mengelola seluruh aktivitas hari ini.
- **Requirement:** FR-052, `docs/brd/09-BRD-INTELLIGENCE-ANALYTICS.md`, `docs/design-system/NEO-BRUTALIST-PIXEL-DESIGN-SYSTEM.md`.
- **Scope:** `frontend/app/(dashboard)/my-day/page.tsx`, komponen terkait, dan hook.
- **Dependency:** T-036 (Work Session Backend), T-078 (UI Primitives).
- **Implementation Expectation:**
  - Menampilkan status kehadiran hari ini (Check-In status, jam tiba, klasifikasi keterlambatan).
  - Tombol kontrol status sesi kerja: `START WORK`, `TAKE BREAK`, `RESUME WORK`, `END WORK` dengan feedback tactile ala arcade.
  - Pemilih tugas aktif dari daftar tugas proyek yang ditugaskan kepada peserta.
  - Timer presisi hitung mundur istirahat dan durasi kerja aktif (font Silkscreen retro mono).
  - Pelacak XP harian menggunakan `PixelProgressBar`.
  - Modal ringkasan akhir hari (End of Day Summary) menggunakan gaya `NesDialog`.
- **Expected Output:** Halaman kokpit interaktif berestetika retro pixel modern yang responsif.
- **Acceptance Criteria:**
  - Peserta dapat menyelesaikan alur kerja harian penuh tanpa berpindah halaman.
  - Indikator status visual berubah seketika sesuai aksi tanpa reload halaman.
- **Risk:** Ketidaksesuaian waktu lokal klien dengan waktu server; mitigasi dengan kalibrasi timestamp berkala setiap 5 menit.

---

## 6. Dependency Graph

```text
[PHASE 1: Foundation Hardening]
(T-001 s/d T-018)  <-- Dapat dikerjakan secara paralel (tanpa dependency antar task)
       │
       ▼
[PHASE 2: Core Infrastructure]
T-009 (Redis Cache) ──────► T-019 (Policy Engine) ─────► T-023 (Rate Limiter)
T-021 (Asynq Setup) ──────► T-022 (Event Bus)
T-010 (Audit Fix)   ──────► T-025 (Full Audit Service)
T-020 (R2 Storage)  ──┐
T-024 (SSE Stream)  ──┼──► (Komponen infrastruktur independen)
T-026 (Settings)    ──┘
       │
       ▼
[PHASE 3: People & Lifecycle]
T-019 (Policy Engine) ────► T-027 (Intern Lifecycle) ──┬──► T-028 (Batch Mgmt) ──► T-032 (User Mgmt)
                                                       ├──► T-029 (Alumni Lifecycle)
                                                       └──► T-031 (Skill Matrix)
T-030 (Institusi) ─────────────────────────────────────────► (Mandiri)
       │
       ├───────────────────────────────────────────────┐
       ▼                                               ▼
[PHASE 4: Workforce & Attendance]             [PHASE 5: Projects & Tasks]
T-019 + T-027                                  T-027 + T-031
       │                                               │
       ▼                                               ▼
T-033 (Work Schedule)                         T-045 (Project Marketplace)
       │                                               │
       ▼                                               ▼
T-034 (Attendance Ingestion)                  T-046 (Application & Quota Lock)
       │                                               │
       ├───────────────┬───────────────┐               ▼
       ▼               ▼               ▼       T-047 (Project Team: Planned %)
T-035 (Lateness)  T-036 (Session)  T-040 (Corr)       │
                       │                       ├───────────────┐
                       ▼                       ▼               ▼
                  T-037 (Break Engine)    T-048 (Milestones) T-052 (Contribution)
                       │                       │
                       ▼                       ▼
                  T-038 (Overtime)        T-049 (Task Kanban)
                                               │
                                               ▼
                                          T-050 (Work Report)
                                               │
                                               ▼
                                          T-051 (Evidence Vault R2)
       │                                               │
       └───────────────────────┬───────────────────────┘
                               ▼
            [PHASE 6: Performance & Gamification]
            T-019 + T-034 + T-050
                   │
                   ▼
            T-054 (XP Rules Engine) ──┬──► T-055 (Attendance XP Triggers)
                   │                  ├──► T-056 (Rank Progression)
                   │                  ├──► T-057 (Task XP Triggers)
                   │                  └──► T-060 (Achievements)
            T-058 (Evaluasi Supervisor) ─► T-059 (Top Performer per Batch)
            T-061 (Skill Growth Matrix)
                   │
                   ▼
            [PHASE 7: Finance & Compensation]
            T-052 + T-056 + T-064
                   │
                   ├───────────────────────────────┐
                   ▼                               ▼
            T-062 (Tax Engine)              T-063 (Personal Wallet)
                   │                               │
                   └───────────────┬───────────────┘
                                   ▼
                            T-064 (Financial Ledger)
                                   │
                                   ▼
                            T-065 (Bounty Distribution)
                                   │
                                   ▼
                            T-066 (Financial Flow Orchestration)
                                   │
                                   ├───────────────┬───────────────┐
                                   ▼               ▼               ▼
                            T-067 (Rank Rewards) T-068 (Batch Fund) T-070 (Payout)
                                   │
                                   ▼
                            T-069 (Reward Claims)

══════════════════════════════════════════════════════════════════════════════════
[FRONTEND TRACK — Dikerjakan Paralel dengan Backend setelah Phase 2]
══════════════════════════════════════════════════════════════════════════════════
T-071 (Deps) ──► T-072 (Tokens) ──► T-078 (UI Primitives) ──► T-081 (Login Page)
T-073 (Fonts) ──► T-074 (Folders) ──► T-075 (API Client) ──► T-076 (Auth Store) ──► T-077 (Providers)
                                                                 │
                                                                 ├──► T-079 (Dynamic Sidebar)
                                                                 └──► T-080 (Auth Guard Middleware)
                                                                 │
                                                                 ▼
[Portals & Domain Pages] ────────────────────────► Terhubung ke API masing-masing
T-087 (My Day)          <-- Butuh T-036 (Work Session) & T-078
T-090 (Scanner Page)    <-- Butuh T-034 (Attendance) & T-078
T-095 (Projects Board)  <-- Butuh T-045 (Projects) & T-078
T-097 (Task Kanban)     <-- Butuh T-049 (Tasks) & T-078
T-104 (Wallet Page)     <-- Butuh T-063 (Wallet) & T-078
T-107 (Command Center)  <-- Butuh T-091 & T-104
T-108 (Supervisor Hub)  <-- Butuh T-092, T-093, T-094
```

---

## 7. Critical Path

Jalur kritis yang menentukan durasi tercepat penyelesaian proyek:

```text
T-009 (Redis Cache)
  │
  ▼
T-019 (Policy Engine)
  │
  ▼
T-027 (Intern Lifecycle)
  │
  ▼
T-033 (Work Schedule Engine)
  │
  ▼
T-034 (Attendance Event Ingestion)
  │
  ▼
T-036 (Work Session Tracking)
  │
  ▼
T-047 (Project Team & Planned Contribution)
  │
  ▼
T-049 (Task Kanban Engine)
  │
  ▼
T-050 (Work Report Submission)
  │
  ▼
T-052 (Three-Layer Contribution Engine)
  │
  ▼
T-064 (Financial Ledger Immutable)
  │
  ▼
T-065 (Project Bounty Distribution)
  │
  ▼
T-066 (Financial Flow Orchestration)
  │
  ▼
T-087 (My Day Portal Frontend)
  │
  ▼
T-116 & T-117 (Critical Verification Tests)
```

**Alasan Jalur Ini Menjadi Critical Path:**
1. **Policy Engine (T-019)** mengontrol seluruh formula toleransi, penalti XP, dan persentase pajak runtime. Keterlambatan di sini memblokir modul presensi, gamifikasi, dan finansial.
2. **Attendance Ingestion (T-034)** merupakan gerbang utama seluruh data aktivitas harian peserta magang.
3. **Task & Report Submission (T-049 → T-050)** menjadi dasar perhitungan kontribusi riil peserta.
4. **Three-Layer Contribution (T-052)** menjadi input mutlak bagi kalkulasi pembagian dana proyek (Bounty Distribution).
5. **Double-Entry Ledger (T-064) & Financial Flow (T-066)** adalah modul dengan risiko tertinggi yang memerlukan pembuktian integritas data (ΣDebit = ΣCredit) sebelum platform dapat dinyatakan siap pakai.

---

## 8. Parallel Work Opportunities

Daftar pekerjaan yang tidak memiliki ketergantungan langsung dan dapat dikerjakan secara bersamaan:

1. **Seluruh Perbaikan Phase 1 (T-001 s/d T-018):** Seluruh perbaikan bug dan penataan teknis backend dapat dikerjakan secara independen tanpa saling memblokir.
2. **Infrastruktur Dasar (Phase 2):**
   - `T-020 (R2 Storage Service)` dapat dikerjakan paralel dengan `T-019 (Policy Engine)`.
   - `T-024 (SSE Stream Endpoint)` dapat dikerjakan paralel dengan `T-021 (Asynq Setup)`.
   - `T-026 (System Settings)` dapat dikerjakan kapan saja di Phase 2.
3. **Backend People vs Frontend Foundation (Phase 3 & Phase 8):**
   - Tim backend mengerjakan `T-027 s/d T-032 (People Module)`.
   - Tim frontend secara simultan mengerjakan `T-071 s/d T-080 (Setup Design System, Primitives, Auth Store)`.
4. **Fitur Pendukung Workforce (Phase 4):**
   - `T-041 (Device Registry)` dapat dikerjakan paralel dengan `T-036 (Work Session)`.
   - `T-039 (Leave Management)` dapat dikerjakan paralel dengan `T-034 (Attendance Ingestion)`.
5. **Fitur Pendukung Projects (Phase 5):**
   - `T-048 (Milestone Management)` dapat dikerjakan paralel dengan `T-046 (Application Quota)`.
   - `T-053 (Skill Matching Engine)` dapat dikerjakan paralel dengan `T-049 (Task Kanban)`.
6. **Frontend Domain Pages (Phase 10, 11, 12):**
   - `T-095 (Projects Board)` dan `T-087 (My Day Portal)` dapat dikerjakan secara paralel oleh developer frontend yang berbeda setelah UI primitives (T-078) selesai.
7. **Pengujian & Deployment (Phase 15, 16, 17):**
   - Penulisan skrip deployment Docker/Nginx (`T-129 s/d T-132`) dapat dikerjakan paralel dengan penulisan test suite (`T-114 s/d T-120`).

---

## 9. Milestones

### Milestone 1 — Foundation Hardened & Core Services Ready
- **Target:** Seluruh kerentanan keamanan backend teratasi, konfigurasi policy engine berjalan, dan setup awal frontend lengkap.
- **Kriteria Keberhasilan:**
  - 0 kerentanan pada modul autentikasi (token type enforced).
  - RBAC dievaluasi dari database/Redis (zero hardcoding).
  - Policy Engine berhasil melayani pembacaan aturan dari cache Redis.
  - Frontend memiliki UI primitives Neo-Brutalist dan terhubung ke API login.

### Milestone 2 — People & Workforce Operational
- **Target:** Siklus hidup peserta magang dan sistem presensi/sesi kerja beroperasi penuh.
- **Kriteria Keberhasilan:**
  - Peserta magang dapat di-onboard dan di-plot ke dalam batch.
  - Endpoint presensi `/terminal-tap` berhasil memproses tap kartu dengan debounce 30 detik.
  - Sesi kerja, istirahat, dan anomali break tercatat akurat di database.
  - Halaman *My Day Portal* dan *Scanner Terminal* di frontend berfungsi interaktif.

### Milestone 3 — Projects, Tasks & Contribution Finalized
- **Target:** Manajemen proyek, kanban tugas, pengunggahan bukti kerja, dan rekonsiliasi kontribusi 3-layer selesai.
- **Kriteria Keberhasilan:**
  - Kuota proyek terkunci secara atomik saat kapasitas terpenuhi.
  - Laporan kerja dapat dikirimkan dengan lampiran bukti di Cloudflare R2 dan disetujui reviewer.
  - Formula rekonsiliasi kontribusi menghasilkan total 100,00% yang disahkan supervisor.

### Milestone 4 — Gamification & Financial Integrity Established
- **Target:** Poin XP mengalir otomatis, sistem ranking aktif, dan pembukuan keuangan double-entry beroperasi tanpa selisih.
- **Kriteria Keberhasilan:**
  - Penalti dan reward XP diproses asinkron oleh worker Asynq berdasarkan event kehadiran dan tugas.
  - Distribusi bounty proyek berhasil memotong pajak PPh dan kas bersama secara dinamis.
  - Buku besar `financial_ledgers` berada pada posisi seimbang (ΣDebit = ΣCredit = Rp 0,00).
  - Saldo dompet peserta terupdate akurat dan dapat diajukan penarikan.

### Milestone 5 — Full Application Experience & Portals Complete
- **Target:** Seluruh portal eksekutif (Command Center, Team Today, Supervisor Workspace) dan fitur dokumen selesai.
- **Kriteria Keberhasilan:**
  - Navigasi sidebar dinamis terfilter sempurna sesuai peran masing-masing pengguna.
  - Seluruh halaman frontend bertema Neo-Brutalist NP-ADS responsif dan mendukung dark mode.
  - Supervisor dapat memantau tim secara real-time melalui Server-Sent Events (SSE).

### Milestone 6 — Production Ready & Verified
- **Target:** Pengujian menyeluruh lulus, keamanan diperketat, dan infrastruktur container deployment siap rilis.
- **Kriteria Keberhasilan:**
  - Backend test coverage ≥85%, frontend test coverage ≥75%.
  - 11 skenario kritis pengujian presensi dan keuangan lulus 100%.
  - Security headers dan rate limiting aktif.
  - Docker Compose menjalankan seluruh komponen (App, DB, Redis, Nginx) dengan graceful shutdown.

---

## 10. Risks & Mitigation

| Risiko | Dampak | Probabilitas | Mitigasi |
|---|---|---|---|
| Kerentanan pertukaran token JWT dieksploitasi sebelum diperbaiki | **CRITICAL** | Rendah (belum rilis) | Tempatkan `T-001` sebagai task pertama (`P0 — BLOCKER`) |
| Kompleksitas evaluasi aturan JSONB pada Policy Engine memperlambat request | **HIGH** | Sedang | Gunakan Redis cache agresif (TTL 5–15 menit) dan struktur JSONB yang flat |
| Lonjakan beban pada endpoint presensi (>50 tap/detik) melebihi SLA 200ms | **HIGH** | Sedang | Buat path eksekusi seringan mungkin (validasi dasar + persist event), proses XP/notifikasi secara asinkron via worker |
| Kesalahan pembulatan desimal menyebabkan ketidakseimbangan buku besar finansial | **CRITICAL** | Rendah | Wajib gunakan tipe data `NUMERIC(15,2)`, hindari kalkulasi float, alokasikan sisa pembulatan ke kas penyeimbang |
| Desinkronisasi timer sesi kerja frontend dengan waktu server (>1 detik) | **MEDIUM** | Sedang | Lakukan kalibrasi timestamp berkala dengan server setiap 5 menit via polling ringan |
| Template sertifikat belum ada (OQ-001 & OQ-002) menghambat rilis modul dokumen | **MEDIUM** | Tinggi | Isolasi modul sertifikat (`T-111` di-mark `BLOCKED`), selesaikan dokumen lain (ID Card, Logbook) terlebih dahulu |
| Ketiadaan vendor payment gateway definitif (OQ-005) menghambat fitur payout | **MEDIUM** | Tinggi | Bangun *state machine* pencairan dana secara lengkap dengan implementasi transfer bank manual sementara |
| Kehilangan event audit log saat container server mati mendadak | **HIGH** | Sedang | Implementasikan channel buffer berukuran terukur dan pastikan graceful shutdown menunggu flush log tuntas |
| Pengujian database merusak data lokal karena tidak ada mekanisme cleanup | **LOW** | Tinggi (sudah terjadi) | Segera perbaiki `auth_test.go` dan `rbac_test.go` dengan `t.Cleanup` atau rollback transaksi database |

---

## 11. Open Questions (OQ)

| ID | Topik Pertanyaan | Sumber | Kebutuhan untuk Task | Status Blocking | Dampak terhadap Keputusan Teknis |
|---|---|---|---|---|---|
| **OQ-001** | Template visual dan format layout sertifikat kelulusan | PRD §3.8 | T-111 (Certificate Engine) | **BLOCKING** | Generator PDF sertifikat belum dapat di-coding sebelum desain visual disetujui |
| **OQ-002** | Nama dan jabatan penandatangan sah sertifikat kelulusan | PRD §3.8 | T-111 (Certificate Engine) | **BLOCKING** | Penentuan kolom tanda tangan digital dan struktur metadata verifikasi sertifikat |
| **OQ-003** | Angka pasti ambang batas (threshold) XP per tingkat Rank | PRD §3.5 | T-056 (Rank Progression) | NON-BLOCKING | Gunakan nilai default dari seed data (0, 250, 750, 1500, 3000) yang dapat diubah via Policy |
| **OQ-004** | Bobot matematis pasti komponen penilaian Performance Score | PRD §3.5 | T-058 (Performance Eval) | NON-BLOCKING | Buat skema pembobotan terkonfigurasi pada policy (default: presensi 25%, tugas 35%, kualitas 20%, rubrik 20%) |
| **OQ-005** | Vendor payment/disbursement gateway untuk pencairan dompet | PRD §3.7 | T-070 (Payout Engine) | PARTIAL | Implementasikan alur approval dan status `REQUESTED` → `APPROVED` → `SETTLED` via transfer manual sementara |
| **OQ-006** | Kuota hari jatah cuti resmi peserta magang per periode | PRD §3.3 | T-039 (Leave Management) | NON-BLOCKING | Implementasikan tanpa batasan kaku di kode; simpan kuota cuti sebagai parameter di Policy Engine |
| **OQ-007** | Prosedur dan otorisasi pencairan dana bersama (Batch Fund) | PRD §3.7 | T-068 (Batch Fund) | PARTIAL | Akumulasi dana dapat berjalan; fitur pencairan dana kas menunggu kejelasan SOP pencairan dari manajemen |
| **OQ-009** | Pemilihan provider model AI dan batasan privasi data | PRD §3.10 | T-053, T-049 (AI Insights) | NON-BLOCKING | Fitur AI bersifat penasihat (P2); gunakan algoritma heuristik statistik lokal untuk MVP |
| **OQ-010** | Pengisian aktivitas harian mandiri vs butuh persetujuan harian | PRD §3.4 | T-050 (Work Report) | NON-BLOCKING | Terapkan *self-reporting* harian secara default, persetujuan formal hanya pada tingkat submission tugas |
| **OQ-011** | Tarif nominal rupiah kompensasi lembur per jam | PRD §3.3 | T-038 (Overtime Workflow) | NON-BLOCKING | Simpan tarif lembur per jam di Policy Engine (default: Rp 0 / kompensasi bonus XP dan jam kerja) |
| **OQ-014** | Event `WORK_ENDED` dimasukkan ke enum DB atau via transisi status | PRD §3.3 | T-036 (Work Session) | NON-BLOCKING | Gunakan 9 enum yang ada; event selesai kerja ditangani melalui transisi status sesi kerja pada saat `CHECK_OUT` |

---

## 12. Definition of Done (DoD)

Proyek DCISP v1.0 dinyatakan **Selesai dan Siap Rilis** apabila memenuhi seluruh kriteria berikut:

### Kualitas Fungsional & Bisnis:
1. Seluruh 55 Functional Requirements (P0 dan P1) terimplementasi dan berfungsi sesuai spesifikasi.
2. Seluruh 25 Business Rules (BR-001 s/d BR-025) terpenuhi tanpa pelanggaran.
3. 10 peran pengguna dapat mengakses modul yang sesuai dengan hak aksesnya, dan isolasi peran Scanner Operator berjalan mutlak (hanya menu pemindai).
4. Siklus hidup peserta magang berjalan lengkap dari pendaftaran hingga status alumni dengan data permanen.
5. Siklus presensi fisik (NFC/QR) hingga pembukuan sesi kerja dan deteksi anomali berjalan tanpa cela.
6. Rekonsiliasi kontribusi proyek 3 layer menghasilkan total 100,00% dan pembagian dana bounty terbukti adil secara matematis.
7. Pembukuan double-entry ledger terbukti seimbang (ΣDebit = ΣCredit) untuk seluruh mutasi keuangan.

### Standar Teknis & Koding:
1. Backend ditulis rapi mengikuti pola Handler → Service → Repository tanpa pelanggaran layer boundary.
2. Setiap fungsi, handler, service, repository, middleware, hook, dan utility memiliki tepat satu kalimat komentar bahasa Indonesia (`// + Kata kerja + objek/tujuan.`).
3. Seluruh pesan respon API dan pesan error menggunakan bahasa Indonesia baku.
4. Nol penggunaan tipe `any` pada frontend TypeScript (`strict: true`).
5. Desain antarmuka mematuhi penuh design tokens Neo-Pixel Adventure Design System (NP-ADS).

### Keamanan & Privasi:
1. Password diamankan menggunakan Argon2id (Memory 64MB, 3 iterasi, 2 thread).
2. Token JWT access (15 menit) dan refresh (7 hari) memiliki validasi tipe yang saling menolak.
3. Nol kode surveilans invasif (dilarang tangkapan layar, keylogger, pemindaian aplikasi OS, dan intip percakapan).
4. Empat tabel immutable (`attendance_event_logs`, `financial_ledgers`, `audit_logs`, `xp_transactions`) terproteksi trigger database dari operasi UPDATE dan DELETE.
5. Seluruh input API divalidasi ketat dan bebas dari celah SQL Injection (100% parameterized query via pgx).

### Pengujian & Kinerja:
1. Backend test coverage mencapai minimal 85% pada core business logic.
2. 11 skenario kritis dari `docs/guide/TESTING.md` lulus secara otomatis.
3. Endpoint presensi `/terminal-tap` memiliki latensi p95 ≤ 200ms pada beban 50 tap per detik.
4. Frontend bebas dari error kompilasi (`tsc --noEmit` bersih) dan lulus build produksi.

### Infrastruktur & Deployment:
1. Seluruh layanan dapat dijalankan menggunakan `docker compose up` dalam keadaan terintegrasi (App, Postgres, Redis, Nginx).
2. Mekanisme backup harian dan pengarsipan WAL berjalan secara otomatis.
3. Health check `/health/liveness` dan `/health/readiness` mencerminkan kondisi konektivitas dependensi riil.

---

## 13. Final Execution Order

Urutan pelaksanaan terstruktur dari awal hingga rilis produksi:

```text
PHASE 01 — Foundation Hardening (Technical Debt & Security Clearance)
  ├── T-001  Fix JWT token type vulnerability (access vs refresh separation)
  ├── T-002  Fix terminated user refresh token bug (filter ACTIVE status)
  ├── T-003  Fix hardcoded JWT secret fallback (fail-hard on empty env)
  ├── T-004  Fix unsafe type assertion in controller.go
  ├── T-005  Prevent internal SQL error details from leaking in API responses
  ├── T-006  Fix readiness probe JSON status body when degraded
  ├── T-007  Refactor dynamic RBAC middleware (remove hardcoded scanner string)
  ├── T-008  Use request context in RBAC middleware queries
  ├── T-009  Implement Redis caching for RBAC permissions lookup
  ├── T-010  Harden audit interceptor (status code tracking & graceful shutdown)
  ├── T-011  Wrap bare errors in service.go with Indonesian context
  ├── T-012  Support multi-role assignment in repository queries
  ├── T-013  Add 409 Conflict and 429 TooManyRequests response helpers
  ├── T-014  Add migration 000003 for missing GIN trigram indexes
  ├── T-015  Correct Go version in go.mod to 1.23
  ├── T-016  Add explicit db: struct tags to identity models
  ├── T-017  Add Indonesian single-sentence comments to test files
  └── T-018  Implement test cleanup routines (t.Cleanup / transaction rollback)

PHASE 02 — Core Infrastructure & Platform Services
  ├── T-019  Implement Unified Policy Engine module (FR-045) [CRITICAL PATH]
  ├── T-020  Implement Cloudflare R2 presigned URL service (FR-044)
  ├── T-021  Setup Asynq background worker infrastructure
  ├── T-022  Implement internal async event bus
  ├── T-023  Implement Redis sliding window rate limiting middleware
  ├── T-024  Implement SSE stream endpoint for real-time events
  ├── T-025  Implement centralized structured audit logging service (FR-047)
  └── T-026  Implement system settings CRUD (FR-048)

PHASE 03 — People & Lifecycle Backend
  ├── T-027  Implement Intern Lifecycle state machine & CRUD (FR-002) [CRITICAL PATH]
  ├── T-028  Implement Batch Management & auto Batch Fund init (FR-004) [CRITICAL PATH]
  ├── T-029  Implement Alumni Lifecycle & permanent retention (FR-003)
  ├── T-030  Implement Institution Master Management (FR-005)
  ├── T-031  Implement Skill Matrix & proficiency assignments (FR-006)
  └── T-032  Implement User Management & supervisor assignment endpoints

PHASE 08 — Frontend Foundation & Design System Setup (PARALEL dengan Phase 3)
  ├── T-071  Install frontend dependencies (Zustand, TanStack Query, Zod, Shadcn)
  ├── T-072  Configure NP-ADS design tokens in Tailwind CSS v4
  ├── T-073  Configure fonts (Press Start 2P, Silkscreen, Mulish)
  ├── T-074  Setup App Router directory structure per development guide
  ├── T-075  Implement HTTP API client with auto-refresh interceptor
  ├── T-076  Implement central auth store using Zustand
  ├── T-077  Wire TanStack Query and state providers in root layout
  ├── T-078  Build Neo-Brutalist UI primitives (Buttons, Cards, Progress, Dialog)
  ├── T-079  Build dynamic RBAC-filtered sidebar navigation (FR-055)
  └── T-080  Implement Next.js middleware for route guards & scanner isolation

PHASE 04 — Workforce & Attendance Backend
  ├── T-033  Implement Work Schedule Engine & grace period (FR-007) [CRITICAL PATH]
  ├── T-034  Implement Attendance Event Ingestion `/terminal-tap` (FR-008) [CRITICAL PATH]
  ├── T-035  Implement 3-tier lateness classification logic
  ├── T-036  Implement Work Session Tracking engine (FR-009) [CRITICAL PATH]
  ├── T-037  Implement Break State Engine & unauthorized penalty (FR-010)
  ├── T-038  Implement Overtime request & approval workflow (FR-012)
  ├── T-039  Implement Leave Management workflow (FR-014)
  ├── T-040  Implement Attendance Correction append-only record (FR-015)
  ├── T-041  Implement Device Registry & terminal HMAC authentication (FR-013)
  ├── T-042  Implement bulk offline attendance sync endpoint
  ├── T-043  Implement dynamic QR code generation with 30s TTL
  └── T-044  Implement audio catalog manifest endpoint for ESP32 terminal

PHASE 09 — Frontend Auth & Identity Pages
  ├── T-081  Build Login page `(auth)/login` with NES Dialog styling
  ├── T-082  Build user profile and Adventurer License display
  └── T-083  Build Super Admin role & permission management interface

PHASE 05 — Projects & Tasks Backend
  ├── T-045  Implement Project Marketplace with 3 visibilities (FR-016) [CRITICAL PATH]
  ├── T-046  Implement atomic project application & quota locking (FR-017)
  ├── T-047  Implement Project Team & Planned Contribution % setup (FR-019) [CRITICAL PATH]
  ├── T-048  Implement Milestone Management with 100% weight validation (FR-020)
  ├── T-049  Implement Task Kanban state machine (FR-021) [CRITICAL PATH]
  ├── T-050  Implement Work Report submission & review workflow (FR-022) [CRITICAL PATH]
  ├── T-051  Implement Evidence Vault with R2 presigned upload (FR-023)
  ├── T-052  Implement Three-Layer Contribution Engine (FR-024) [CRITICAL PATH]
  └── T-053  Implement advisory Skill Matching Engine (FR-018)

PHASE 10 — Frontend People & Workforce Pages
  ├── T-084  Build intern directory and search page
  ├── T-085  Build intern detail profile view
  ├── T-086  Build batch management and cohort overview
  ├── T-087  Build My Day Portal cockpit (FR-052) [CRITICAL PATH]
  ├── T-088  Build `useWorkSession` precision timer hook with server calibration
  ├── T-089  Build `useSessionIntegrity` tab-blur tracking hook (FR-011)
  ├── T-090  Build Attendance Scanner terminal interface `/attendance/scanner`
  ├── T-091  Build Team Today live supervisor dashboard (FR-053)
  ├── T-092  Build overtime request & approval interface
  ├── T-093  Build leave application & tracking interface
  └── T-094  Build attendance correction request interface

PHASE 06 — Performance & Gamification Backend
  ├── T-054  Implement XP Rules Engine with 3 isolated schemes (FR-025) [CRITICAL PATH]
  ├── T-055  Wire attendance events to async XP processing workers [CRITICAL PATH]
  ├── T-056  Implement Rank Progression & threshold evaluation (FR-026)
  ├── T-057  Wire task and milestone completion to Project XP triggers
  ├── T-058  Implement supervisor rubric Performance Evaluation (FR-027)
  ├── T-059  Implement singular Top Performer per Batch algorithm (FR-028)
  ├── T-060  Implement Achievement badge unlock system (FR-029)
  └── T-061  Implement Skill Growth Matrix temporal aggregation (FR-030)

PHASE 07 — Finance & Compensation Backend
  ├── T-062  Implement Deduction & Tax Engine (FR-034)
  ├── T-063  Implement Personal Wallet with itemized transaction history (FR-035)
  ├── T-064  Implement Immutable Double-Entry Financial Ledger (FR-037) [CRITICAL PATH]
  ├── T-065  Implement Project Bounty Distribution engine (FR-032) [CRITICAL PATH]
  ├── T-066  Implement Financial Flow Orchestration with ACID rollback (FR-039) [CRITICAL PATH]
  ├── T-067  Implement Multi-Component Rank Rewards issuance (FR-031)
  ├── T-068  Implement Batch Fund accumulation and balance tracking (FR-036)
  ├── T-069  Implement physical reward claims lifecycle (FR-033)
  └── T-070  Implement Payout Engine cashout workflow (FR-038)

PHASE 11 — Frontend Projects & Tasks Pages
  ├── T-095  Build Project Marketplace (Bounty Board) interface
  ├── T-096  Build project detail and milestone tracking page
  ├── T-097  Build drag-and-drop Task Kanban Board
  ├── T-098  Build Work Report submission modal with R2 upload
  ├── T-099  Build Three-Layer Contribution visual dashboard
  └── T-100  Build project application dialogue and status tracker

PHASE 12 — Frontend Performance & Finance Pages
  ├── T-101  Build XP & Rank status dashboard with 3-scheme progress bars
  ├── T-102  Build Hall of Heroes non-toxic leaderboard
  ├── T-103  Build supervisor performance evaluation rubric form
  ├── T-104  Build Personal Wallet (Coin Pouch) interface & payout form
  ├── T-105  Build double-entry ledger audit view for Finance role
  └── T-106  Build reward claim interface and tracking view

PHASE 13 — Frontend Executive Portals
  ├── T-107  Build Command Center 4-quadrant executive dashboard (FR-051)
  ├── T-108  Build Supervisor Workspace approval hub (FR-054)
  └── T-109  Build Notification Center & toast dispatch (FR-046)

PHASE 14 — Documents & Digital Assets
  ├── T-110  Build digital ID card / Adventurer License generator (FR-040)
  ├── T-111  Build digital certificate engine [BLOCKED by OQ-001, OQ-002] (FR-041)
  ├── T-112  Build public portfolio generator `/portfolio/:slug` (FR-042)
  └── T-113  Build academic logbook PDF export generator (FR-043)

PHASE 15 — Testing & Quality Assurance
  ├── T-114  Write backend unit tests for Identity & Auth module (target ≥85%)
  ├── T-115  Write backend unit tests for People & Lifecycle state machine
  ├── T-116  Write integration tests for 11 attendance scenarios [CRITICAL PATH]
  ├── T-117  Write integration tests for double-entry financial balance [CRITICAL PATH]
  ├── T-118  Write immutability trigger verification tests for all 4 append-only tables
  ├── T-119  Write validation tests for 100.00% contribution sum guard
  ├── T-120  Write tests for 3-scheme XP partition and rank promotion
  ├── T-121  Write frontend component tests with Vitest & React Testing Library
  └── T-122  Write Playwright E2E tests for core user workflows

PHASE 16 — Security & Performance Hardening
  ├── T-123  Configure strict CORS origin whitelist for production
  ├── T-124  Implement security headers middleware (CSP, HSTS, X-Frame-Options)
  ├── T-125  Enforce HTTP request body size limits
  ├── T-126  Enforce registration password complexity validation
  ├── T-127  Configure slow database query logging (>500ms)
  └── T-128  Benchmark and optimize attendance ingestion path for ≤200ms latency

PHASE 17 — Deployment & Production Readiness
  ├── T-129  Write production Nginx configuration with TLS 1.3 and SSE support
  ├── T-130  Write multi-stage Dockerfile for Go backend
  ├── T-131  Write standalone mode Dockerfile for Next.js frontend
  ├── T-132  Update docker-compose.yml for full stack orchestration
  ├── T-133  Create GitHub Actions CI workflow (lint, test, build)
  ├── T-134  Create automated PostgreSQL backup & WAL archive script
  └── T-135  Document production `.env.production` configuration template
```

---

## 14. Recommended First Task

### **Task Rekomendasi: T-001 — Fix JWT Token Type Vulnerability**

#### Kenapa task ini harus dikerjakan pertama kali:
1. **Security Blocker Mutlak:** Saat ini access token (15 menit) dan refresh token (7 hari) dapat saling menggantikan karena tidak adanya klaim `type` pada payload token. Celah ini mengekspos seluruh endpoint terproteksi dan mekanisme refresh token.
2. **Nol Ketergantungan (Zero Dependency):** Task ini tidak membutuhkan modul lain dan dapat langsung diuji secara terisolasi.
3. **Mencegah Rework di Modul Berikutnya:** Seluruh modul baru yang akan dibangun (Workforce, Projects, Finance) akan mengonsumsi middleware autentikasi ini. Membiarkan auth cacat saat membangun modul lain akan mengakibatkan refactoring berulang di masa mendatang.
4. **Perubahan Aman:** Sistem belum berada di tahap produksi dengan pengguna aktif, sehingga perubahan format token tidak menimbulkan disrupsi data pengguna riil.

#### Prerequisite:
- Membaca dan memahami `backend/internal/shared/utils/token.go` (102 baris).
- Membaca dan memahami `backend/internal/middleware/auth_jwt.go` (51 baris).
- Membaca dan memahami metode `RefreshToken` pada `backend/internal/modules/identity/service.go` (baris 112–138).

#### Output yang Harus Dihasilkan:
1. Penambahan field `TokenType string` pada struct `JWTClaims` dengan nilai `"access"` atau `"refresh"`.
2. Pembaruan fungsi `GenerateTokenPair` untuk menyuntikkan klaim `token_type` secara eksplisit pada kedua token.
3. Penambahan parameter `expectedType string` pada `ValidateToken`, yang akan menolak token jika `TokenType` tidak cocok.
4. Pembaruan middleware `AuthJWT` untuk memvalidasi bahwa token bertipe `"access"`.
5. Pembaruan service `RefreshToken` untuk memvalidasi bahwa token bertipe `"refresh"`.

#### Acceptance Criteria:
- Request ke endpoint terproteksi (`GET /api/v1/auth/me`) menggunakan refresh token ditolak dengan respon `401 Unauthorized`.
- Request ke endpoint refresh (`POST /api/v1/auth/refresh`) menggunakan access token ditolak dengan respon `401 Unauthorized`.
- Request ke endpoint terproteksi menggunakan access token yang valid berhasil dengan status `200 OK`.
- Request refresh token yang valid berhasil menerbitkan sepasang token baru dengan status `200 OK`.

#### Task yang Dapat Dimulai Setelahnya:
- `T-002`, `T-003`, `T-004`, `T-005` (seluruh perbaikan bug Phase 1 lainnya dapat langsung dieksekusi secara paralel).
- Setelah Phase 1 rampung: `T-019` (Unified Policy Engine) dan `T-071` (Frontend Dependencies Setup).
