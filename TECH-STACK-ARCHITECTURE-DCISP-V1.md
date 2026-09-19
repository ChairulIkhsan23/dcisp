# TECH STACK & TECHNICAL ARCHITECTURE DESIGN
# Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0
**Document Version:** 1.0  
**Status:** Architecture Blueprint / Ready for Review  
**Date:** 2026-09-19  
**Source of Truth:** `PRD-DCISP-V1.md`  

---

## 01. TECHNICAL REQUIREMENT EXTRACTION

Ekstraksi kebutuhan teknis langsung dari dokumen PRD DCISP v1.0:

| Kategori Kebutuhan | Spesifikasi Kebutuhan Teknis | Referensi PRD |
|---|---|---|
| **Application Type** | Multi-role Enterprise Workforce & Internship Management Platform with Gamification Performance Layer | Section 2.1, 2.3 |
| **Platform** | Responsive Web Application (Progressive Web-ready) untuk desktop workstation dan tablet terminal | Section 5.2, 9.12 |
| **Client Application** | Modern Web Browser (Chrome, Firefox, Edge, Safari) mendukung W3C Page Visibility API & Window Focus API | Section 12.3, 14.4 |
| **Backend** | Modular Backend API Service berbasis RESTful JSON OpenAPI 3.0, stateless JWT auth, dynamic policy resolver | Section 11.1, FR-001, FR-045 |
| **Frontend** | Single-page UI / Cockpit dashboard ("My Day", "Team Today", "Command Center", "Supervisor Workspace") | Section 7.3, FR-051–FR-055 |
| **Mobile Requirement** | Tampilan web mobile responsif esensial (<768px: My Day, QR scanner, notifikasi). Native mobile out-of-scope | Section 5.2, 9.12 |
| **Web Requirement** | Desktop Workstation primary layout (1024px–1920px+), Collapsible Sidebar, Dense data grids, Precision Timers | Section 9.9, 9.12 |
| **Desktop / Hardware** | Integrasi Terminal Presensi Fisik (NFC reader / Barcode QR camera / Dedicated tablet) via Device Gateway | Section 11.2, FR-008, FR-013 |
| **API Requirement** | RESTful HTTP/JSON API, versioning (`/api/v1/`), rate limiting (120 req/min general, 10 req/min auth), latency $\le 200\text{ms}$ scan | Section 11.1, 12.1 |
| **Database Requirement** | Relational Database (RDBMS) ACID-compliant, JSONB support, fixed-point decimal monetary support, append-only triggers | Section 10.1, 10.3, 12.4 |
| **File / Media Storage** | Cloudflare R2 Object Storage (S3-compatible API), Presigned PUT/GET URL, metadata-only di database SQL (NO BLOBs) | BR-024, FR-044, Section 11.4 |
| **Authentication** | JWT Bearer Token (access token 15 min + refresh token rotation), API Key SHA-256 untuk Device Gateway | Section 11.1, 12.2 |
| **Authorization / RBAC** | Dynamic RBAC 5-lapis (`Role -> Permission -> Scope -> Resource -> Action`), zero hardcoded checks, Scanner isolation | BR-001, BR-002, FR-001 |
| **Real-time Requirement** | Ingestion presensi real-time, live timer My Day, real-time alert anomali Team Today, in-app push notification | Section 11.6, 13.9, FR-053 |
| **Background Processing** | Offline attendance batch sync, daily schedule reconciler (missing checkout / mangkir penalty), notification queue | Section 11.2, 11.6, 14.5 |
| **Queue / Job Requirement** | Message broker / Asynchronous Job Queue dengan exponential backoff retry dan Dead-Letter Queue (DLQ) | Section 11.10, 14.5 |
| **Notification** | Event-driven multi-channel Notification Center (in-app notifications persisten, toast visual real-time) | FR-046, Section 11.6 |
| **Search Requirement** | Pencarian data pengguna, skill catalog, proyek bursa marketplace, filter transaksi ledger & log audit | FR-006, FR-016, FR-047 |
| **Reporting Requirement** | Ekspor Laporan Magang PDF / Logbook Cetak (FR-043), Ekspor Audit Log, Ekspor Neraca Saldo Keuangan | FR-043, FR-047, FR-049 |
| **Analytics Requirement** | Dasbor metrik agregat: Punctuality rate, Work session integrity ratio, Bounty burn rate, Rank velocity | Section 13.1–13.10, FR-049 |
| **AI / ML Requirement** | Smart Skill Matching (FR-018) & AI Insights (FR-050) — Non-punitive, informational advisory only | BR-015, OQ-009 |
| **Third-Party Integration**| Git Repositories (GitHub/GitLab PR/commit verification), Cloudflare R2, Payment Disbursement Gateway [TBD] | Section 11.4, 11.5, 11.8 |
| **Logging & Audit Trail** | Structured JSON logging, Immutable Append-Only Audit Log (FR-047) mencakup IP, UA, old_state, new_state | BR-025, FR-047, Section 12.5 |
| **Monitoring & Health** | Health check endpoints (`/health/liveness`, `/health/readiness`), telemetri kueri, queue monitoring | Section 12.5 |
| **Caching Requirement** | In-memory caching untuk debouncing scan 30 detik, rate limiting, sesi terdistribusi, master policy runtime | Section 11.2, 14.5 |
| **Security & Privacy** | OWASP Top 10, Argon2id/BCrypt password hashing, Strict Privacy boundary (NO screenshot, NO keylogger, NO chat spy) | BR-010, BR-011, Section 12.2 |
| **Performance SLA** | Scan endpoint latency $\le 200\text{ms}$ (p95 $\le 500\text{ms}$), 500 concurrent users, 50 scans/sec peak, first load $\le 2.0\text{s}$ | Section 12.1 |
| **Availability & Backup** | Uptime 99.5%, Daily automated full backup (02:00 WIB), Point-in-time WAL logs every 15 min, Off-site storage | Section 12.4, 12.8 |

---

## 02. SYSTEM CHARACTERISTICS

### A. Explicit Requirements (Tertulis Langsung di PRD)
1. **10 Peran Pengguna Formal:** Super Admin, Admin, HR Admin, Project Manager, Supervisor, Reviewer, Finance, Scanner Operator, Intern, Alumni (Section 3.2).
2. **Dynamic RBAC & Unified Policy Engine:** Konfigurasi peran, hak akses, rumus pajak, dan penalti XP runtime tersimpan di database/policy tanpa hardcoding kode (BR-001, FR-045).
3. **Pemisahan 5 Dimensi Waktu:** Attendance Time, Work Session Time, Active Session Time, Break Time, Overtime Time (BR-007).
4. **Pemisahan 3 Skema XP:** Internship XP, Project XP, Alumni Contribution (BR-004).
5. **Partisi 3 Jenis Pekerjaan:** Daily Work, Project Task, Work Report (BR-005).
6. **Alur Kontribusi 3-Lapis:** Planned -> Actual (sistem) -> Final (disahkan supervisor) (BR-016).
7. **Buku Kas Ganda Immutabel:** Double-entry ledger seimbang ($\sum \text{Debit} = \sum \text{Kredit}$, selisih Rp 0, no update/delete) (BR-023, FR-037).
8. **Cloudflare R2 Storage Separation:** Database hanya menyimpan metadata file; objek fisik disimpan di R2 (BR-024, FR-044).
9. **Isolasi Mutlak Scanner Operator:** Hak akses operator terminal dibatasi pada `scope = scanner_only` (BR-002).
10. **Batasan Privasi Mutlak:** Larangan screenshot, keylogger, intip chat, pemindaian proses OS (BR-010).

### B. Logical Deductions (Hasil Deduksi Logis)
1. **Kebutuhan ACID RDBMS Kuat:** Karena formula finansial ($\text{Gross} - \text{Tax} - \text{Farewell} = \text{Net}$) melibatkan banyak entitas (Wallet, Batch Fund, Financial Ledger) secara serentak, database non-relasional murni (NoSQL) tidak cocok dan berisiko menimbulkan *inconsistent balances*.
2. **Kebutuhan In-Memory Cache Cepat (Redis):** Debouncing 30 detik untuk scanner NFC/QR dan rate limiting membutuhkan penyimpanan in-memory yang cepat dan atomic (TTL-based keys) agar tidak membebani database disk I/O.
3. **Kebutuhan Background Job Queue:** Pengiriman notifikasi, kalkulasi rekonsiliasi akhir hari, dan sinkronisasi presensi offline wajib dijalankan di background worker agar API endpoint scan presensi tetap berada di bawah ambang batas latency $\le 200\text{ms}$.
4. **Kebutuhan Client-side Session State:** Portal "My Day" membutuhkan timer presisi berbasis interval lokal dengan sinkronisasi timestamp server periodik untuk mencegah *timer drifting*.

### C. Assumptions (Asumsi Teknis yang Memerlukan Konfirmasi)
1. **Single Region Deployment:** Server di-deploy di wilayah Jakarta/Singapura (UTC+7 / Asia/Jakarta) dengan waktu database tersimpan dalam UTC (ASM-007).
2. **Disbursement Fallback:** Sebelum integrasi payment gateway (OQ-005) disahkan, modul payout menyediakan mekanisme persetujuan manual dengan pencatatan slip transfer bank (RSK-009).
3. **AI Graceful Fallback:** Fitur AI (FR-018, FR-050) dirancang decoupled; sistem tetap 100% fungsional menggunakan algoritma heuristik jika API AI offline atau belum dikonfigurasi (RSK-011).

---

## 03. ARCHITECTURE APPROACH

### Evaluasi Gaya Arsitektur

| Gaya Arsitektur | Kesesuaian PRD | Alasan Penerimaan / Penolakan |
|---|---|---|
| **Microservices** | **DITOLAK** | Menimbulkan kompleksitas operasional berlebih (*overengineering*), transaksi finansial terdistribusi memerlukan Saga/2PC yang rentan inkonsistensi saldo ledger, tim developer kecil terbebani *service overhead*. Beban hanya 500 concurrent users. |
| **Serverless (Pure FaaS)** | **DITOLAK** | *Cold start* mengancam SLA presensi $\le 200\text{ms}$, kesulitan mengelola koneksi database pool untuk transaksi ledger berpasangan yang ketat, dan kompleksitas background workers panjang. |
| **Traditional Monolith (Spaghetti)** | **DITOLAK** | Tidak memisahkan boundary domain modul, menyulitkan isolasi RBAC dinamis dan mempersulit pemeliharaan jangka panjang. |
| **Modular Monolith** | **DIREKOMENDASIKAN (PILIHAN UTAMA)** | Memenuhi seluruh requirement PRD secara tepat: Domain boundary terisolasi bersih (Identity, Workforce, Projects, Performance, Finance, Documents, Intelligence), transaksi ACID database lokal terjamin 100%, deployment tunggal hemat biaya, latensi antar-modul instan (in-memory calls), dan mudah dipecah menjadi microservices di masa depan jika skala membesar. |

### Dampak Pemilihan Modular Monolith:
* **Development:** Produktivitas tinggi, satu repositori, refactoring aman via TypeScript type checking.
* **Deployment Complexity:** Sangat rendah (satu container aplikasi + satu database + satu Redis).
* **Operational Complexity:** Rendah, logging dan tracing terpusat tanpa distributed tracing yang rumit.
* **Scalability:** Mampu menangani hingga puluhan ribu request/menit dengan multi-instance horizontal scaling di balik reverse proxy.
* **Data Consistency:** 100% konsisten melalui database ACID transaction native.

---

## 04. FRONTEND TECHNOLOGY

### Evaluasi Kandidat Frontend

| Kandidat | Kelebihan | Kekurangan | Kesesuaian |
|---|---|---|---|
| **Next.js (App Router)** | Full-stack SSR/SSG, ekosistem React luas, SEO prima. | SSR overhead tidak dibutuhkan untuk internal dashboard, potensi hidrasi mismatch pada live timer "My Day". | Kurang Optimal |
| **Vue 3 / Nuxt 3** | Reaktivitas tajam, ringan, sintaks bersih. | Ekosistem enterprise UI komponen dan library tipe data tabel finansial lebih terbatas dibanding React. | Alternatif Baik |
| **React 18/19 (SPA via Vite) + TypeScript** | Client-side rendering murni, kontrol total atas DOM lifecycle & Page Visibility API, zero hydration mismatch untuk precision timer, rendering tabel padat data sangat cepat, ekosistem library data grid dan form paling matang di industri. | Memerlukan setup routing client terpisah (namun aplikasi ini adalah private operational tool, bukan public SEO-driven site). | **DIREKOMENDASIKAN** |

### Recommended Frontend Stack:
* **Framework:** React (Vite SPA)
* **Language:** TypeScript 5.x (Strict mode)
* **Styling:** Tailwind CSS v3/v4 (Design tokens semantic: Emerald, Amber, Crimson, Slate/Cyan)
* **UI Component System:** Shadcn UI + Radix UI Primitives (Aksesibilitas WCAG 2.1 AA native, keyboard accessible)
* **Icons:** Lucide React
* **Form Handling:** React Hook Form
* **Validation:** Zod (Skema validasi runtime shared dengan backend)
* **State Management:** Zustand (Global user session, active work timer state, sidebar collapse state)
* **Data Fetching & Cache:** TanStack Query v5 (@tanstack/react-query) (Optimistic updates, auto retry, cache invalidation)
* **Table / Data Grid:** TanStack Table v8 (@tanstack/react-table) (Virtualization-ready, sorting, column visibility)
* **Charting:** Recharts (Kurva XP, utilisasi lembur, metrik kehadiran)
* **File Upload:** Uppy / Native Presigned S3 Uploader (Direct upload ke Cloudflare R2 dengan progress bar)
* **Testing:** Vitest + React Testing Library

---

## 05. BACKEND TECHNOLOGY

### Evaluasi Kandidat Backend

| Kandidat | Kelebihan | Kekurangan | Kesesuaian |
|---|---|---|---|
| **Laravel 11 (PHP 8.3)** | Ekosistem sangat lengkap (Auth, Queue, Scheduler, Eloquent), produktivitas CRUD cepat. | Kurang optimal untuk WebSocket/SSE konkurensi tinggi dan memory footprint lebih berat dibanding Go/Node. | Alternatif Baik |
| **Go (Echo / Gin)** | Performa dan konkurensi luar biasa, binary kecil. | Pembangunan domain business logic kompleks (Three-layer contribution, dynamic policy engine) memerlukan banyak boilerplate manual. | Kurang Cepat Dev |
| **Python (FastAPI)** | Bagus untuk integrasi AI dan sintaks ringkas. | Penanganan transaksi ACID enterprise dan background queues multi-worker membutuhkan setup tool eksternal yang terpisah-pisah. | Kurang Optimal |
| **Node.js / NestJS (TypeScript)** | Arsitektur Modular Monolith tingkat enterprise secara native (Modules, Controllers, Services, Guards, Interceptors), berbagi tipe data Zod/TypeScript dengan frontend, asynchronous event handling prima, integrasi BullMQ dan Prisma/TypeORM kelas satu. | Sedikit kurva pembelajaran konsep dependency injection (DI). | **DIREKOMENDASIKAN** |

### Recommended Backend Stack:
* **Language:** TypeScript 5.x (Node.js 20 LTS)
* **Framework:** NestJS 10.x (Fastify atau Express adapter)
* **API Architecture:** RESTful API with OpenAPI / Swagger Auto-generator
* **ORM / Query Builder:** Prisma ORM / Drizzle ORM (Type-safe SQL queries, declarative migration, native PostgreSQL support)
* **Validation:** Zod / Class-Validator + NestJS ValidationPipe
* **Authentication:** Passport.js + `@nestjs/jwt` (Stateless JWT token access 15m + refresh token rotasi di DB)
* **Authorization:** NestJS Custom Guards (`@RequirePermissions()`, `@RequireScope()`) terintegrasi Dynamic Policy Engine
* **Queue & Background Jobs:** BullMQ (berbasis Redis) untuk asynchronous worker
* **Scheduler:** `@nestjs/schedule` (Cron jobs untuk daily schedule reconciler dan periodic checks)
* **Event System:** `@nestjs/event-emitter` (In-memory asynchronous event bus untuk decoupling domain)
* **HTTP Client:** `@nestjs/axios` (Axios wrapper dengan retry interceptor untuk Cloudflare R2 & Git APIs)
* **Logging:** Winston / Pino (Structured JSON logging dengan correlation/trace ID)
* **Testing:** Jest + Supertest (Unit testing business rules & Integration API testing)

---

## 06. DATABASE ARCHITECTURE

### Primary Database: PostgreSQL 16+
PostgreSQL dipilih sebagai **Database Relasional Primer Tunggal** berdasarkan argumentasi teknis PRD:

1. **Integritas Finansial ACID:** Menjamin pembukuan ganda immutabel (`Financial Ledger`) tidak pernah mengalami ketidakseimbangan neraca ($\sum \text{Debit} - \sum \text{Credit} = 0$) melalui *Database Constraints & Atomic Transactions*.
2. **Aturan Append-Only & Immutability:** Penegakan aturan bisnis BR-023, BR-025, dan FR-047 dapat dikunci di level database menggunakan *PostgreSQL Trigger Function* yang memblokir `UPDATE` dan `DELETE` pada tabel `attendance_event_logs`, `audit_logs`, `financial_ledgers`, dan `xp_transactions`.
3. **Penyimpanan Kebijakan Fleksibel (JSONB):** Mendukung kolom `condition_rules` dan `action_definitions` pada tabel `policies` (FR-045) dengan kueri terindeks `GIN (Generalized Inverted Index)`.
4. **Tipe Data Numerik Presisi:** Mendukung kolom `DECIMAL(15, 2)` untuk saldo moneter dan persentase kontribusi tanpa risiko floating-point error.
5. **Pencarian Teks Bawaan (Full-Text Search):** Menggunakan modul native `pg_trgm` dan `tsvector` untuk pencarian nama intern, skill, dan judul proyek tanpa memerlukan cluster Elasticsearch eksternal.

### Database Pendukung (In-Memory): Redis 7+
* **Fungsi:** 
  1. Rate Limiter storage.
  2. Debouncing presensi scanner 30 detik (key: `attendance:debounce:<user_id>`, TTL: 30s).
  3. BullMQ Job Queue persistence.
  4. Cache evaluasi Policy Engine aktif (TTL: 5-15 menit dengan event invalidation).

---

## 07. AUTHENTICATION & AUTHORIZATION

### 1. Mekanisme Otentikasi (Dual Authentication)
* **Web Users (Intern, Supervisor, Admin, Finance, Alumni):**
  - Menggunakan JSON Web Token (JWT) Bearer Token.
  - **Access Token:** Masa berlaku singkat (**15 menit**), memuat `user_id`, `active_role`, dan `scopes`.
  - **Refresh Token:** Masa berlaku **7 hari**, di-hash dan disimpan dalam database (`refresh_tokens`), mendukung *Token Rotation* dan *Revocation*.
  - **Password Security:** Di-hash menggunakan algoritma **Argon2id** (memory cost 64MB, time cost 3 iterations) sesuai Section 12.2.
* **Device Gateway (Terminal Presensi / Scanner Operator):**
  - Menggunakan **API Key Bertanda Tangan (SHA-256 Hash)** yang terdaftar pada tabel `devices` (FR-013).
  - Header Request: `X-Device-ID`, `X-Device-Token`, `X-DCISP-Signature`, `X-Timestamp`.

### 2. Arsitektur Otorisasi Dinamis 5 Lapis (Dynamic RBAC)
Mengimplementasikan hierarki:
$$\text{Role} \longrightarrow \text{Permission} \longrightarrow \text{Scope} \longrightarrow \text{Resource} \longrightarrow \text{Action}$$

```text
HTTP Request
     │
     ▼
[ JWT Auth Guard ] ──(Ekstrak User ID, Role, Scopes)
     │
     ▼
[ Dynamic Permission Guard ] ──(Evaluasi DB/Cache: Role memiliki Permission untuk Action?)
     │
     ▼
[ Scope Context Guard ] ──(Validasi Context ID: Apakah resource milik user / tim / project yang sah?)
     │
     ├──► [ DENIED ] ──► HTTP 403 Forbidden (Audit Log: ACCESS_DENIED_EVENT)
     │
     └──► [ ALLOWED ] ──► Forward ke Controller Logic
```

* **Penegakan Isolasi Scanner Operator (BR-002):** Peran Scanner Operator hanya memiliki izin tunggal `attendance.scanner.scan` dengan scope `scanner_only`. Akses ke endpoint lain otomatis ditolak dengan HTTP 403.

---

## 08. API ARCHITECTURE

### Standar Desain API RESTful
* **Base URL:** `/api/v1`
* **Format Payload:** `application/json` (Encoding: UTF-8)
* **Envelope Standar Respons:**

```json
// Response Sukses
{
  "success": true,
  "statusCode": 200,
  "message": "Work session started successfully",
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 20,
    "totalItems": 100,
    "totalPages": 5
  }
}

// Response Error
{
  "success": false,
  "statusCode": 422,
  "error": "Unprocessable Entity",
  "message": "Final contribution percentage must sum to exactly 100.00%",
  "errors": [
    { "field": "final_contribution_pct", "message": "Total sum is 95.00%" }
  ],
  "traceId": "c8a4f912-32b1-4b72-9b2f-9811abdc1234"
}
```

* **Idempotensi Endpoint Presensi (FR-008):**
  - Header: `X-Idempotency-Key` (dihitung dari `SHA-256(device_id + user_id + event_type + timestamp_minute)`).
  - Mencegah rekaman ganda akibat pengiriman ulang sinyal jaringan.
* **Rate Limiting:**
  - `/api/v1/auth/*`: 10 request / menit per IP.
  - `/api/v1/attendance/events`: 120 request / menit per Device ID.
  - General API: 120 request / menit per IP.

---

## 09. BACKGROUND PROCESSING

### Identifikasi Job Asinkron & Queue Stack (BullMQ + Redis)

```text
[ Controller / Event Emitter ]
             │
             ▼ (Enqueue Job Payload)
     [ Redis / BullMQ ]
             │
   ┌─────────┼────────────────────────┬────────────────────────┐
   ▼         ▼                        ▼                        ▼
[ Queue:  [ Queue:                 [ Queue:                 [ Queue:
  Attend ]  Notifications ]          Fin-Reconcile ]          Documents ]
   │         │                        │                        │
   ▼         ▼                        ▼                        ▼
Offline   In-App Push /            Daily Schedule           PDF Logbook /
Sync      Toast Dispatcher         Anomaly & Penalty        Certificate Gen
Worker    Worker                   Worker (00:05 WIB)       Worker
```

| Nama Antrean (*Queue*) | Tugas / Fungsi Worker | Retry Policy | Failure Handling |
|---|---|---|---|
| `attendance-offline-sync` | Memproses batch event offline dari terminal scanner | 3x retry (exponential backoff) | Log ke Dead-Letter Queue (DLQ), notifikasi IT Admin |
| `notifications-queue` | Mengirim event notifikasi in-app ke pengguna | 3x retry | Disimpan di database, status `failed` |
| `finance-reconcile-queue` | Menghitung penalti keterlambatan/absensi harian | 2x retry | Rollback transaksi, alert Finance & Super Admin |
| `documents-generator-queue` | Merender PDF sertifikat kelulusan & logbook resmi | 2x retry | Simpan status error pada metadata dokumen |

---

## 10. REAL-TIME ARCHITECTURE

### Strategi Real-Time: Polling Ringan + Server-Sent Events (SSE)
* **Kebutuhan Nyata PRD:**
  1. Status kehadiran live pada portal "Team Today" supervisor (FR-053).
  2. Notifikasi peringatan real-time (anomali istirahat, approval masuk) (FR-046).
  3. Live status counter pada dashboard "My Day" (FR-052).
* **Keputusan Arsitektur:**
  - **Server-Sent Events (SSE) via `/api/v1/events/stream`:** Menggunakan koneksi unidirectional HTTP/2 SSE yang jauh lebih hemat resource dan andal di balik reverse proxy dibanding Full-Duplex WebSocket, karena kebutuhan data murni bersifat *server-to-client push* (broadcast alert & status update).
  - **Client Precision Timers (Zustand + Web Worker / setInterval):** Penghitung waktu detik sesi kerja di My Day berjalan lokal di client dengan kalibrasi timestamp server setiap pergantian status (menghindari beban pengiriman event per detik dari server).

---

## 11. FILE & MEDIA STORAGE

### Implementasi Storage Cloudflare R2 (S3-Compatible API)
* **Prinsip Arsitektur (BR-024):** Database PostgreSQL **hanya menyimpan metadata berkas** (`File Metadata Record` — DATA-007). Objek fisik biner dialirkan langsung ke Cloudflare R2.
* **Alur Upload Langsung (Direct Upload via Presigned URL):**
  ```text
  [ Frontend Client ] ──(1) Request Presigned PUT URL + MIME/Size validation──► [ Backend API ]
          │                                                                           │
          │ ◄────────────────(2) Return Presigned S3 PUT URL (TTL: 15 min)────────────┘
          │
          ▼ (3) Direct HTTP PUT binary stream
  [ Cloudflare R2 Bucket ]
          │
          ▼ (4) Upload Success Callback
  [ Frontend Client ] ──(5) Confirm Upload & Save Record──► [ Backend: Persist File Metadata ]
  ```
* **Partisi Direktori Bucket R2:**
  - `profile/` (Avatar pengguna)
  - `id-cards/` (Dokumen ID card digital)
  - `task-evidence/YYYY/MM/` (Tangkapan layar & berkas bukti tugas)
  - `certificates/` (PDF sertifikat kelulusan digital)
  - `logbooks/` (PDF ekspor logbook cetak)
  - `rewards/` (Gambar katalog hadiah)
  - `gallery/` (Media dokumentasi kegiatan)

---

## 12. CACHE & PERFORMANCE

### Strategi Caching Konseptual

| Lapisan Cache | Komponen yang Dicache | TTL | Strategi Invalidation |
|---|---|---|---|
| **Attendance Debounce** | UID kartu / User ID scan presensi | 30 detik | Expire alami (Time-To-Live) |
| **Active Policy Rules** | Aturan runtime Unified Policy Engine | 15 menit | *Event-driven flush* saat Super Admin mengedit policy |
| **User Permissions** | Role permissions & scopes per pengguna | 15 menit | *Invalidation on update* role/permission mapping |
| **API Rate Limiter** | Request counter per IP / Device Token | 60 detik | Sliding window counter Redis |
| **Static Assets** | File frontend build, bundle JS/CSS | 1 tahun | Cloudflare CDN dengan cache busting hash |

---

## 13. SEARCH

### Strategi Search: Native PostgreSQL Full-Text Search (FTS) & Trigram
* **Analisis Kebutuhan:** Pencarian berkisar pada direktori intern (nama, NIM, institusi), katalog skill (nama skill, kategori), bursa proyek (judul, deskripsi), dan filter audit log.
* **Keputusan Teknis:** Cukup menggunakan ekstensi **`pg_trgm`** dan indeks **`GIN`** pada kolom target di PostgreSQL.
* **Justifikasi (Anti-Overengineering):** Volume data pada platform ini tidak memerlukan cluster Elasticsearch / OpenSearch terpisah yang menambah biaya server dan kompleksitas sinkronisasi data.

---

## 14. THIRD-PARTY INTEGRATION

| Nama Integrasi | Fungsi Bisnis | Protokol / Data | Auth Method | Sifat | Penanganan Kegagalan |
|---|---|---|---|---|---|
| **Cloudflare R2** | Penyimpanan objek berkas bukti, foto, sertifikat | AWS S3 SDK / REST | Access Key + Secret | Async upload | Retry 3x, fallback error message |
| **Device Gateway** | Penghubung terminal scanner NFC/QR lobi | HTTPS POST / JSON | API Key + Signature | Sync ingestion | Buffer lokal terminal, sync saat online |
| **Git Repositories (GitHub/GitLab)** | Verifikasi commit URL / PR tugas | HTTPS GET / Regex parsing | Public URL / Webhook | Async verify | Fallback ke review manual oleh reviewer |
| **Payment Gateway** `[TBD — OQ-005]` | Otomasi pencairan dana payout dompet | REST API / Webhooks | Bearer Token / HMAC | Async payout | Circuit breaker, penahanan status PENDING |
| **AI Provider** `[TBD — OQ-009]` | Smart skill match & activity insights | REST JSON API | API Bearer Key | Async advisory | Fallback ke kalkulasi heuristik lokal |

---

## 15. SECURITY ARCHITECTURE

1. **Prinsip Privacy-First Tracking (BR-010, BR-011):** Larangan mutlak perekaman invasif (screenshot, keylogger, chat reading). Pelacakan murni menggunakan `document.visibilityState` dan `window.onblur`.
2. **Kriptografi & Hashing:**
   - Password: **Argon2id**.
   - API Key & Token Hash: **HMAC-SHA256**.
   - Presigned URL Token: **AWS SigV4**.
3. **Immutability Database:** Trigger database memblokir `UPDATE` dan `DELETE` pada tabel audit, log presensi, transaksi XP, dan buku besar finansial.
4. **Proteksi Jaringan & Web:** Enkripsi TLS 1.3 wajib, Security Headers (CSP, HSTS, X-Frame-Options: DENY), proteksi XSS via input sanitization, dan proteksi SQL Injection via ORM parameterized queries.

---

## 16. OBSERVABILITY

* **Structured Logging:** Format JSON standar dengan field `timestamp`, `level`, `traceId`, `userId`, `service`, `message`, `context`.
* **Security Audit Log (FR-047):** Tabel `audit_logs` merekam seluruh peristiwa mutasi data sensitif, login, dan eskalasi hak akses secara append-only.
* **Health Check Probes:** Endpoint `/health/liveness` (status server hidup) dan `/health/readiness` (konektivitas PostgreSQL, Redis, dan R2).
* **Monitoring:** Telemetri utilisasi CPU, memori, antrean BullMQ, dan PostgreSQL slow queries (>500ms).

---

## 17. TESTING STACK

* **Unit Testing (Jest / Vitest):** Cakupan minimal 85% untuk logika bisnis kritis:
  - Formula perhitungan XP Rules Engine (FR-025).
  - Three-Layer Contribution reconciliation (FR-024).
  - Formula Deduction & Tax Engine (FR-034).
  - Keseimbangan pembukuan Double-Entry Ledger (FR-037).
  - Mesin status transisi (Break state, Intern lifecycle, Overtime).
* **Integration Testing (Supertest + Testcontainers / Ephemeral DB):** Pengujian endpoint API end-to-end (Auth -> Attendance Tap -> Session Update -> Ledger Entry).
* **E2E Testing (Playwright):** Pengujian alur kritis terbatas: Alur login, portal "My Day" (Start -> Break -> End Work), dan approval supervisor.

---

## 18. DEVOPS & INFRASTRUCTURE

* **Containerization:** Docker & Docker Compose (Multi-stage build untuk NodeJS NestJS dan Vite Frontend).
* **Reverse Proxy / Web Server:** Nginx (TLS termination, gzip/brotli compression, rate limiting buffer, SSE proxy buffering off).
* **Database Hosting:** Managed PostgreSQL 16 (Connection pooling via PgBouncer).
* **Cache & Broker:** Managed Redis 7 (AOF / RDB persistence).
* **Object Storage:** Cloudflare R2 Bucket.
* **CI/CD Pipeline:** GitHub Actions (Linting, TypeScript Check, Automated Unit/Integration Test, Docker Build & Push).
* **Backup Strategy:** Automated daily PostgreSQL dump (02:00 WIB) + Continuous WAL archiving (RPO: 15 min, RTO: < 4 jam).

---

## 19. RECOMMENDED TECH STACK

| Layer | Technology | Reason | PRD Requirement |
|---|---|---|---|
| **Frontend Framework** | **React 18/19 (Vite SPA)** | Render client murni, kontrol total event fokus/visibilitas browser, zero hydration issue untuk timer harian | FR-052, FR-011, Section 9.9 |
| **Frontend Language** | **TypeScript 5.x** | Type safety end-to-end, mencegah runtime bug pada data moneter | Section 12.6 |
| **UI Components & Styling** | **Tailwind CSS + Shadcn UI (Radix)** | Komponen enterprise aksesibel (WCAG AA), token warna semantik, keyboard accessible | Section 9.3, 12.7 |
| **Frontend State & Fetching** | **Zustand + TanStack Query v5** | Pengelolaan timer lokal efisien, cache invalidation otomatis untuk live data | FR-052, FR-053 |
| **Backend Framework** | **NestJS (Node.js 20 LTS)** | Arsitektur Modular Monolith native, modularitas domain bersih, performa async I/O tinggi | Section 11.1, FR-001 |
| **Backend Language** | **TypeScript 5.x** | Konsistensi bahasa dengan frontend, skema Zod terbagi (shared validation) | Section 12.6 |
| **ORM / Database Access** | **Prisma ORM / Drizzle ORM** | Type-safe queries, migration versioning, integrasi transaksi ACID PostgreSQL | Section 10.3, 12.4 |
| **Primary Database** | **PostgreSQL 16+** | Standar kepatuhan ACID mutlak untuk Double-Entry Ledger, JSONB Policy Engine, immutability triggers | BR-023, FR-037, FR-045 |
| **Cache & Queue Broker** | **Redis 7+ (via BullMQ)** | Debouncing presensi 30s, rate limiting, background worker decoupled | FR-008, Section 11.2, 14.5 |
| **Object Storage** | **Cloudflare R2** | Kompatibel S3, zero egress fee, direct presigned upload, metadata di SQL | BR-024, FR-044 |
| **Authentication** | **Passport JWT + Argon2id** | Stateless JWT bearer token (15m) + refresh token rotation di database | Section 11.1, 12.2 |
| **Authorization** | **Custom Dynamic RBAC Guard** | Evaluasi izin 5-lapis runtime, isolasi penuh Scanner Operator (`scanner_only`) | BR-001, BR-002, FR-001 |
| **Real-time Push** | **Server-Sent Events (SSE)** | Unidirectional push ringan untuk notifikasi live & update presensi | Section 11.6, FR-046 |
| **Logging & Audit** | **Winston/Pino + PostgreSQL Audit Log** | Structured logging JSON + tabel audit log append-only tidak dapat diubah | FR-047, Section 12.5 |
| **Testing** | **Vitest + Jest + Playwright** | Unit test formula bisnis, integration test API, E2E flow kritis | Section 18.3, 18.4 |
| **DevOps / Deploy** | **Docker + Nginx + GitHub Actions** | Portabilitas container, deployment terotomatisasi, SSL TLS 1.3 | Section 12.2, 18.1 |

---

## 20. ALTERNATIVE STACK

### Option A (Recommended): TypeScript Modular Monolith (NestJS + React SPA)
* **Tech:** NestJS + Prisma + PostgreSQL + Redis + React (Vite) + Cloudflare R2
* **Pros:** Single language (TypeScript end-to-end), modular architecture strictly enforces domain boundaries, outstanding real-time async performance, direct presigned S3 support.
* **Cons:** Memerlukan konfigurasi boilerplate NestJS awal.
* **Suitable when:** Membutuhkan arsitektur bersih, type-safe lintas modul, dan performa tinggi untuk event-driven background processing.

### Option B: PHP Monolith (Laravel 11 + Inertia.js + React)
* **Tech:** Laravel 11 + Eloquent + PostgreSQL + Redis + Inertia.js (React) + Cloudflare R2
* **Pros:** Kecepatan pengembangan CRUD sangat tinggi, built-in queue/scheduler/auth scaffolding matang.
* **Cons:** Konsumsi memori lebih tinggi untuk high concurrency scan, penanganan real-time SSE membutuhkan process manager terpisah (Octane/Reverb).
* **Suitable when:** Tim memiliki keahlian dominan di PHP/Laravel dan ingin meminimalkan pemisahan repositori API.

---

## 21. SYSTEM ARCHITECTURE

### Gambaran Arsitektur Tingkat Tinggi

```text
[ CLIENT APPLICATIONS ]
 ├── Desktop Workstation (Intern, PM, Supervisor, Admin, Finance) ──► React SPA (Vite)
 ├── Terminal Pemindai Lobi (Scanner Operator) ──────────────────────► React Scanner View
 └── Hardware Readers (NFC / QR Scanner) ───────────────────────────► Device Gateway Client
                                                                            │
                                                                            ▼ (HTTPS / TLS 1.3)
[ EDGE & REVERSE PROXY ]
 └── Cloudflare CDN & Nginx Reverse Proxy (SSL Termination, Rate Limiting, Static Asset Cache)
                               │
                               ▼ (REST API / SSE Events)
[ BACKEND MODULAR MONOLITH (NestJS) ]
 ├── API Gateway Layer (JWT Auth Middleware, Rate Limiter, Scope & Dynamic RBAC Guard)
 │
 ├── Core Modular Business Domains:
 │    ├── 1. Identity & RBAC Module (Auth, Roles, Dynamic Permissions)
 │    ├── 2. People & Lifecycle Module (Interns, Alumni, Batches, Skills)
 │    ├── 3. Workforce & Attendance Module (Schedules, Attendance Engine, Work Sessions, Breaks)
 │    ├── 4. Projects & Tasks Module (Marketplace, Teams, Milestones, Tasks, Work Reports, Evidence)
 │    ├── 5. Performance & Gamification Module (XP Rules Engine, Ranks, Formal Evaluation, Top Performer)
 │    ├── 6. Incentives & Compensation Module (Rank Rewards, Bounty Allocation, Reward Claims)
 │    ├── 7. Finance & Taxation Module (Deduction Engine, Wallets, Batch Fund, Double-Entry Ledger)
 │    ├── 8. Documents & Storage Module (Presigned URL Bridge, ID Cards, Certificates, Portfolios)
 │    ├── 9. System & Integrity Module (Unified Policy Engine, Audit Logger, Notifications)
 │    └── 10. Intelligence Module (Analytics Aggregator, AI Informational Insights)
 │
 ├── Internal Event Bus (Asynchronous Domain Event Emitter)
 └── Background Worker Engine (BullMQ Queue Processors)
                               │
            ┌──────────────────┼─────────────────────────┐
            ▼                  ▼                         ▼
[ PRIMARY DATABASE ]    [ IN-MEMORY STORE ]    [ OBJECT STORAGE ]
  PostgreSQL 16+          Redis 7+               Cloudflare R2
  (ACID Tables,           (BullMQ Queues,        (Task Evidence,
   Append-Only Ledger,     Debounce Cache,        Certificates,
   JSONB Policies,         Rate Limiters)         ID Cards, Media)
   Trigram FTS Search)
```

---

## 22. MODULE $\rightarrow$ TECHNOLOGY MAPPING

| Modul PRD | Frontend Component | Backend Service / Module | Database Entity / Tables | External Service | Background Job |
|---|---|---|---|---|---|
| **Identity & RBAC** | Role Guard, User Profile Menu | `IdentityModule`, `RbacGuard` | `users`, `roles`, `permissions`, `scopes` | — | — |
| **People & Lifecycle** | Interns List, Batch Manager | `PeopleModule`, `LifecycleService` | `interns`, `alumni`, `batches`, `institutions` | — | Batch Auto-Completion Check |
| **Workforce & Presensi**| My Day Cockpit, Team Today, Scanner | `AttendanceModule`, `WorkSessionService`| `attendance_events`, `work_sessions`, `breaks`, `schedules` | Terminal NFC Reader | Offline Sync Worker, Debounce |
| **Projects & Tasks** | Marketplace, Kanban Task Board | `ProjectsModule`, `TaskService` | `projects`, `project_teams`, `tasks`, `submissions`, `evidence` | Git Repositories (Verify) | Quota Lock Reconciler |
| **Performance & Game** | XP Progress Bar, Rank Badges, Eval Form | `PerformanceModule`, `XpEngine` | `xp_rules`, `xp_transactions`, `ranks`, `evaluations` | — | Daily Streak & XP Processor |
| **Incentives & Bounty** | Reward Catalog, Bounty Allocator | `IncentivesModule`, `BountyService` | `rewards`, `reward_claims`, `project_teams` | — | Claim Fulfillment Dispatcher |
| **Finance & Ledger** | Personal Wallet, Ledger Viewer | `FinanceModule`, `LedgerService` | `wallets`, `wallet_txs`, `batch_funds`, `financial_ledgers`, `tax_rules` | Payout Gateway [TBD] | Payout Webhook Processor |
| **Documents & Storage**| PDF Viewer, Portfolio Showcase | `DocumentsModule`, `StorageService` | `file_metadata`, `certificates`, `portfolios` | Cloudflare R2 | PDF Certificate / Logbook Gen |
| **System & Policies** | Policy Configurator, Audit Table | `SystemModule`, `PolicyEngine` | `policies`, `audit_logs`, `system_settings` | — | Policy Cache Invalidator |
| **Insights & Analytics**| Command Center Charts, AI Summary | `InsightsModule`, `AnalyticsService` | Read queries on transactions & logs | AI Provider [TBD] | Daily Metrik Aggregator |

---

## 23. TECHNICAL DEPENDENCIES

### Blocking Dependencies (Wajib Siap Sebelum Modul Berjalan):
1. `Identity & RBAC` $\rightarrow$ Seluruh modul (Blocking otorisasi).
2. `PostgreSQL ACID Schema` $\rightarrow$ `Financial Ledger & Wallet` (Blocking integritas saldo).
3. `Cloudflare R2 Bucket Credentials` $\rightarrow$ `Evidence & Document Upload` (Blocking penyerahan deliverable).
4. `Redis Instance` $\rightarrow$ `Attendance Event Engine & Background Workers` (Blocking debouncing dan antrean).
5. `Work Schedule Engine` $\rightarrow$ `Attendance Event Engine & XP Rules` (Blocking validasi keterlambatan).

### Optional / Non-Blocking Dependencies:
1. `Payment Disbursement Gateway` $\rightarrow$ `Payout Engine` (Dapat dialihkan ke alur manual bank transfer sementara).
2. `External AI Provider` $\rightarrow$ `Skill Matching & AI Insights` (Fallback otomatis ke algoritma heuristik lokal).
3. `External Git Webhooks` $\rightarrow$ `Evidence Verification` (Dapat diverifikasi via input link manual oleh reviewer).

---

## 24. TECHNICAL RISKS & MITIGATION

| ID | Technical Risk | Penyebab | Dampak | Strategi Mitigasi Teknis |
|---|---|---|---|---|
| **TR-001** | Ketidakseimbangan Jurnal Pembukuan Ganda | Galat pembulatan desimal atau kegagalan atomik | Saldo gantung, inkonsistensi audit finansial | Wajibkan database transaction rollback jika $\sum \text{Debit} - \sum \text{Credit} \neq 0$; gunakan tipe data fixed-point integer (Rupiah utuh). |
| **TR-002** | Lonjakan Antrean Presensi (Scan Throttling) | Ratusan pengguna melakukan tap dalam jendela 5 menit saat jam masuk | Waktu respon API $> 200\text{ms}$, antrean fisik di kantor | Terapkan caching debouncing di Redis, endpoint ingestion super ramping (hanya validasi dasar & persist event log), proses XP/notifikasi secara asinkron di worker. |
| **TR-003** | Terminal Presensi Offline saat Internet Putus | Gangguan ISP di lobi kantor | Presensi gagal dicatat | Terminal Gateway menyimpan payload di SQLite lokal bertanda tangan kriptografis; batch sync dengan flag `buffered_offline` saat online. |
| **TR-004** | Kebocoran Privasi Sesi Kerja | Bug pada script pemantau fokus | Pelanggaran BR-010, resistensi pengguna | Isolasi kode pelacak: dilarang memuat API screenshot, clipboard, atau desktop scanning; audit berkala pada client-side tracking script. |
| **TR-005** | Kegagalan Penyedia AI Eksternal | Downtime / Rate Limit API AI | Fitur matching & insight error | Desain decoupled: jadikan output AI murni *advisory* dan sediakan fallback instan ke pencocokan irisan array lokal tanpa melempar 500 error. |

---

## 25. OVERENGINEERING CHECK

Prinsip: *"Use the simplest architecture that satisfies the requirements."*

* [x] **Apakah Microservices diperlukan?** **TIDAK.** Skala 500 pengguna aktif dan transaksi finansial ACID ganda paling stabil dan hemat dalam *Modular Monolith*.
* [x] **Apakah Elasticsearch/OpenSearch diperlukan?** **TIDAK.** Pencarian teks pada PRD cukup ditangani oleh indeks `pg_trgm` PostgreSQL native.
* [x] **Apakah Kubernetes (K8s) diperlukan?** **TIDAK.** Docker Compose / Single Node Container Orchestration (seperti Docker Swarm atau Kamal) sudah lebih dari cukup.
* [x] **Apakah GraphQL diperlukan?** **TIDAK.** RESTful JSON dengan envelope standar dan query parameters sudah sangat mencukupi kebutuhan layar terfokus.
* [x] **Apakah Full-Duplex WebSockets diperlukan di semua tempat?** **TIDAK.** Server-Sent Events (SSE) untuk push alert + client-side timers lokal sudah memenuhi kebutuhan tanpa overhead koneksi bidirectional.
* [x] **Apakah Database NoSQL terpisah diperlukan?** **TIDAK.** Kolom JSONB PostgreSQL mampu menangani dynamic policy conditions tanpa perlu MongoDB.

---

## 26. TRACEABILITY

| Requirement PRD | Kebutuhan Teknis (*Technical Need*) | Keputusan Teknologi (*Technology Decision*) | Alasan Teknis (*Reason*) |
|---|---|---|---|
| **BR-001 (Dynamic RBAC)** | Evaluasi izin runtime berbasis database tanpa hardcode | NestJS Custom Guards + PostgreSQL Permissions Cache | Memungkinkan perubahan role-permission instan tanpa deploy ulang kode |
| **BR-002 (Scanner Isolation)** | Operator pemindai terisolasi total dari data lain | Scope-based Authorization Filter (`scope = scanner_only`) | Mencegah kebocoran data sensitif ke terminal fisik lobi |
| **BR-010 (Privacy-First Tracking)**| Pemantauan etis non-invasif | W3C Page Visibility API (`document.visibilityState`) + Idle Timer | Menjaga privasi mutlak tanpa software pengintai |
| **BR-023 (Immutable Ledger)** | Pembukuan ganda berpasangan append-only | PostgreSQL Trigger Functions (Block UPDATE/DELETE) + ACID Transactions | Menjamin integritas audit finansial dan kepatuhan akuntansi |
| **BR-024 (R2 Storage Separation)**| Pemisahan berkas fisik dari database | AWS S3 SDK Presigned URL + Cloudflare R2 Bucket | Database SQL tetap ringan dan performan tanpa beban data biner BLOB |
| **FR-008 (Latency $\le 200\text{ms}$)**| Ingestion presensi kilat & anti-duplicate | Redis Debouncing (30s) + Idempotency Key (SHA-256) | Mencegah antrean penumpukan di pintu masuk kantor |
| **FR-045 (Unified Policy Engine)**| Aturan runtime dinamis untuk XP, Pajak, Jadwal | PostgreSQL JSONB Rules + Rule Evaluator Service | Adaptif terhadap perubahan kebijakan organisasi secara langsung |

---

## 27. UNRESOLVED TECHNICAL DECISIONS

| Keputusan Teknis | Informasi yang Belum Tersedia | Mengapa Berdampak Teknis | Pertanyaan Lanjutan yang Direkomendasikan |
|---|---|---|---|
| **OQ-001 & OQ-002 (Sertifikat Digital)** | Template visual, nomor registrasi baku, pejabat penandatangan | Mempengaruhi library rendering PDF (Puppeteer vs PDFKit) dan foreign key signature | Apa format penomoran sertifikat dan apakah tanda tangan berupa gambar atau sertifikat kriptografis? |
| **OQ-003 (XP Rank Thresholds)** | Ambang batas numerik pasti (Rank F s/d S) | Menentukan data awal (*seed data*) pada tabel ranks/policy | Berapa nilai XP minimum untuk setiap tingkatan Rank? |
| **OQ-004 (Formula Bobot Performa)** | Bobot persentase eksak Presensi, Kualitas, Evaluasi, Kontribusi | Menentukan parameter kalkulasi komposit di `PerformanceEvaluationEngine` | Berapa bobot matematis untuk masing-masing pilar evaluasi formal (misal: 30% Presensi, 30% Deliverable, 40% Review)? |
| **OQ-005 (Payment Gateway Provider)** | Vendor disbursement API (Midtrans / Xendit / Oy!) | Menentukan kontrak integrasi webhook dan payload payout | Vendor gateway apa yang akan digunakan untuk pencairan dana dompet? |
| **OQ-009 (AI Model Provider)** | Penyedia model AI (OpenAI / Claude / Local LLM) | Menentukan SDK perutean payload eksternal dan batasan biaya token | Apakah integrasi AI menggunakan OpenAI API, Claude API, atau model open-source lokal? |
| **OQ-014 (Event `WORK_ENDED` Enum)** | Kepastian pencatatan `WORK_ENDED` pada enum presensi | Mempengaruhi konsistensi enum database `Attendance Event Log` | Apakah `WORK_ENDED` resmi ditambahkan ke enum database atau cukup diwakili status sesi kerja? |

---

## 28. FINAL TECH STACK BLUEPRINT

```text
================================================================================
                    DCISP v1.0 TECHNICAL STACK BLUEPRINT
================================================================================

[ FRONTEND ]
  • Framework        : React 18/19 (Vite Single Page Application)
  • Language         : TypeScript 5.x (Strict Type Checking)
  • Styling          : Tailwind CSS + Semantic Design Tokens
  • Components       : Shadcn UI + Radix UI Primitives (WCAG 2.1 AA)
  • State Management : Zustand (Session & Timers) + TanStack Query v5 (Server State)
  • Forms & Validate : React Hook Form + Zod
  • Tables & Charts  : TanStack Table v8 + Recharts

[ BACKEND ]
  • Framework        : NestJS 10.x (Modular Monolith Architecture)
  • Language         : TypeScript 5.x (Node.js 20 LTS)
  • ORM / Data Layer : Prisma ORM / Drizzle ORM
  • Validation       : Zod / Class-Validator
  • Security / Auth  : Passport.js (JWT Bearer 15m + Refresh Tokens) + Argon2id
  • Authorization    : Dynamic 5-Tier RBAC Guard + Unified Policy Resolver

[ DATABASE & STORAGE ]
  • Primary Database : PostgreSQL 16+ (ACID, Append-Only Triggers, JSONB, Trigram FTS)
  • In-Memory Cache  : Redis 7+ (Rate Limiting, Debounce 30s, Policy Cache)
  • Background Queue : BullMQ (Redis-backed async job processors)
  • Object Storage   : Cloudflare R2 (S3-Compatible API, Presigned Direct Uploads)

[ REAL-TIME & OBSERVABILITY ]
  • Real-Time Push   : Server-Sent Events (SSE) for Notifications & Live Updates
  • Logging          : Winston / Pino Structured JSON Logging (Trace/Correlation ID)
  • Audit Logging    : Dedicated Immutable Append-Only PostgreSQL Table
  • Health Monitoring: `/health/liveness` & `/health/readiness` Probes

[ DEVOPS & INFRASTRUCTURE ]
  • Containerization : Docker (Multi-stage builds) & Docker Compose
  • Web Server       : Nginx (Reverse Proxy, TLS 1.3, Rate Limit Buffer)
  • CI/CD Pipeline   : GitHub Actions (Automated Lint, Test, Build, Deploy)
  • Backup Strategy  : Daily Automated pg_dump (02:00 WIB) + 15-min WAL Archiving
================================================================================
```

---

## 29. FINAL ARCHITECTURE DECISION

1. **Architecture Style:** **Modular Monolith** — Pilihan optimal yang menjamin kedaulatan domain terisolasi, kesederhanaan operasional, dan kepatuhan transaksi ACID lokal untuk buku besar keuangan berpasangan.
2. **Frontend Stack:** **React (Vite) + TypeScript + Tailwind CSS + Shadcn UI** — Menghadirkan antarmuka workstation/cockpit yang cepat, presisi, tanpa SSR timer glitch, dan teroptimasi untuk density data tinggi.
3. **Backend Stack:** **NestJS (TypeScript) + Prisma ORM** — Menyediakan struktur arsitektur enterprise modular yang tangguh, asynchronous event bus terintegrasi, dan type-safety shared dengan frontend.
4. **Database:** **PostgreSQL 16+** — Menjadi sumber kebenaran tunggal (*Single Source of Truth*) dengan penegakan trigger *append-only* untuk ledger, log presensi, dan audit trail.
5. **Authentication & Authorization:** **Stateless JWT (15m) + DB Refresh Token Rotation** dipadukan dengan **Dynamic 5-Tier RBAC Guard** dan isolasi khusus Scanner Operator.
6. **API Strategy:** **RESTful JSON OpenAPI 3.0** dengan envelope standar, idempotency header, dan rate limiting ketat.
7. **Background Processing:** **BullMQ + Redis** untuk pemrosesan presensi offline, pengiriman notifikasi, rekonsiliasi keterlambatan harian, dan generator dokumen.
8. **Storage:** **Cloudflare R2** dengan arsitektur presigned upload langsung; database hanya mencatat metadata file.
9. **Cache:** **Redis 7+** murni untuk debouncing 30s presensi, rate limiter, antrean job, dan cache policy runtime.
10. **External Integrations:** Decoupled design dengan graceful fallback untuk AI, Git, dan Payment Gateway.
11. **Testing Strategy:** Piramida pengujian terfokus pada **Unit Testing Logika Bisnis (Jest/Vitest $\ge 85\%$)** dan **Integration Testing API (Supertest)**.
12. **Infrastructure & Deployment:** **Dockerized Containers di balik Nginx Reverse Proxy**, di-deploy secara terotomatisasi via **GitHub Actions CI/CD**.

---

## 30. IMPLEMENTATION READINESS

### Checklist Kesiapan Arsitektur:

#### Ready (Siap untuk Implementasi):
- [x] **Architecture Style:** Modular Monolith (NestJS + React SPA)
- [x] **Frontend Stack:** React (Vite) + TypeScript + Tailwind CSS + Shadcn UI + Zustand + TanStack Query/Table
- [x] **Backend Stack:** NestJS + TypeScript + Prisma ORM + Zod Validation
- [x] **Database Architecture:** PostgreSQL 16+ (ACID, Triggers Append-Only, JSONB Policies)
- [x] **Authentication:** JWT Bearer (15m) + Refresh Token Rotation + Argon2id Password Hash
- [x] **Authorization:** Dynamic RBAC (5-Tier) + Unified Policy Resolver + Scanner Isolation
- [x] **API Strategy:** RESTful JSON OpenAPI 3.0 + Idempotency Key + Envelope Response
- [x] **Storage Strategy:** Cloudflare R2 S3-Compatible + Presigned URL direct upload + Metadata in SQL
- [x] **Cache & Rate Limiting:** Redis 7+ for debouncing 30s & rate limit counters
- [x] **Queue & Workers:** BullMQ for async background jobs (offline sync, notifications, reconcile)
- [x] **Testing Strategy:** Vitest/Jest for Business Rules Unit Tests + Supertest for API Integration Tests
- [x] **Infrastructure & CI/CD:** Docker + Nginx + GitHub Actions

#### Needs Clarification (Menunggu Konfirmasi Stakeholder saat Fitur Spesifik Dibangun):
- [ ] **OQ-001 / OQ-002:** Detail template desain dan pejabat penandatangan Sertifikat Digital (FR-041).
- [ ] **OQ-003:** Nilai pasti ambang batas XP kenaikan Rank F s/d S (FR-026).
- [ ] **OQ-004:** Bobot matematis pasti formula Skor Kinerja Formal (FR-027).
- [ ] **OQ-005:** Pemilihan vendor spesifik Payment / Disbursement Gateway untuk Payout Engine (FR-038).
- [ ] **OQ-009:** Pemilihan vendor API AI Provider untuk Smart Skill Match (FR-018) dan AI Insights (FR-050).
- [ ] **OQ-014:** Keputusan formal status enum event `WORK_ENDED` pada tabel Attendance Event Log.

---
*Dokumen rancangan arsitektur teknis ini disimpan secara permanen di `TECH-STACK-ARCHITECTURE-DCISP-V1.md`.*
