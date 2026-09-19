# DOKUMEN SPESIFIKASI FINAL TECH STACK & ARSITEKTUR IOT ATTENDANCE
# Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0
**Document Version:** 1.0 — Production Ready Architecture  
**Date:** 2026-09-19  
**Source of Truth:** `PRD-DCISP-V1.md`  

---

## 1. EXECUTIVE SUMMARY & ARSITEKTUR UTAMA

Platform DCISP v1.0 mengadopsi arsitektur **Modular Monolith Berkinerja Tinggi** dengan pemisahan domain yang bersih, dirancang untuk performa konkurensi presensi instan ($\le 200\text{ ms}$), kepatuhan akuntansi ganda immutabel, dan integrasi perangkat keras IoT NFC edge-terminal yang tangguh:

```text
[ ARSITEKTUR KESELURUHAN SISTEM DCISP ]

  ┌────────────────────────────────────────────────────────────────────────┐
  │                           HARDWARE EDGE LAYER                          │
  │  [ ACR1552U NFC Reader ] ──(USB CCID)──► [ ESP32-S3 Microcontroller ] │
  │                                                    │                   │
  │                   [ MAX98357A I2S DAC + Speaker ] ◄┘                   │
  │                   (Global Pre-Generated Audio Cache)                   │
  └───────────────────────────────────┬────────────────────────────────────┘
                                      │ HTTPS (mTLS / HMAC-SHA256)
                                      ▼
  ┌────────────────────────────────────────────────────────────────────────┐
  │                         EDGE & REVERSE PROXY                           │
  │        Cloudflare CDN (DDoS, WAF) + Nginx (TLS 1.3 Termination)        │
  └───────────────────────────────────┬────────────────────────────────────┘
                                      │
                 ┌────────────────────┴────────────────────┐
                 │ REST / JSON                             │ Server-Sent Events (SSE)
                 ▼                                         ▼
  ┌────────────────────────────────────────────────────────────────────────┐
  │                       BACKEND API ENGINE (GOLANG)                      │
  │  • Framework: Gin (github.com/gin-gonic/gin)                           │
  │  • Modules: Identity & RBAC, People, Workforce/Attendance, Projects,   │
  │             Performance & XP, Incentives, Finance & Ledger, Documents, │
  │             System Policies, Intelligence Engine                       │
  │  • Background Workers: Asynq (Redis-backed Job Queue)                  │
  │  • Scheduler: Robfig Cron                                              │
  └──────────────┬────────────────────┬────────────────────┬───────────────┘
                 │                    │                    │
                 ▼                    ▼                    ▼
  ┌──────────────────────┐  ┌──────────────────┐  ┌────────────────────────┐
  │  PRIMARY DATABASE    │  │ IN-MEMORY CACHE  │  │ GLOBAL OBJECT STORAGE  │
  │  PostgreSQL 16+      │  │ Redis 7+         │  │ Cloudflare R2          │
  │  • ACID Transactions │  │ • Debouncing 30s │  │ • Task Evidence (R2)   │
  │  • Append-Only Trig. │  │ • Asynq Queues   │  │ • Global Audio Assets  │
  │  • Double-Entry Ldg. │  │ • Rate Limiter   │  │ • Certificates / PDFs  │
  │  • JSONB Policies    │  │ • Policy Cache   │  │ • Metadata-Only in SQL │
  └──────────────────────┘  └──────────────────┘  └────────────────────────┘
                                      ▲
                                      │ REST API / SSR / Client Hydration
  ┌───────────────────────────────────┴────────────────────────────────────┐
  │                     FRONTEND APPLICATION (NEXT.JS)                     │
  │  • Framework: Next.js 15+ (App Router) + TypeScript 5.x                │
  │  • Styling: Tailwind CSS + Shadcn UI (Radix Primitives)                │
  │  • State & Fetching: TanStack Query v5 + Zustand                       │
  │  • Dedicated Cockpits: "My Day", "Team Today", "Command Center",       │
  │                        "Supervisor Workspace", "Terminal View"         │
  └────────────────────────────────────────────────────────────────────────┘
```

---

## 2. REKOMENDASI FINAL TECH STACK TERINTEGRASI

| Komponen Layer | Teknologi Terpilih | Justifikasi & Pemetaan Requirement PRD |
|---|---|---|
| **Backend Language** | **Golang 1.23+** | Kompilasi biner native, memory footprint rendah (<30MB baseline), goroutine concurrency ultra-cepat untuk melayani throughput presensi 50 scans/sec dan latensi $\le 200\text{ ms}$ (FR-008, Section 12.1). |
| **Backend Framework** | **Gin (`github.com/gin-gonic/gin`)** | HTTP router berbasis Radix tree berperforma tinggi, zero allocation routing, middleware modular untuk dynamic RBAC dan JWT validation (FR-001, Section 11.1). |
| **Backend Database Driver / ORM** | **`jackc/pgx/v5` + `sqlc` atau `GORM`** | Eksekusi transaksi PostgreSQL ACID native dengan performa maksimal, prepared statements, connection pooling aman, dan type-safe code generation untuk pembukuan ganda (FR-037). |
| **Backend Queue Worker** | **`hibiken/asynq` + Redis** | Background queue terdistribusi untuk offline sync, notifikasi real-time, rekonsiliasi keterlambatan harian (00:05 WIB), dan dispatching event presensi (Section 11.6, 14.5). |
| **Backend Scheduler** | **`robfig/cron/v3`** | Penjadwalan periodik pemindaian anomali presensi harian dan rotasi policy engine (FR-045). |
| **Frontend Framework** | **Next.js 15+ (App Router) + React 19** | Full-stack web client, Server Components untuk data rendering cepat, Client Components untuk live timer cockpit "My Day", dan zero-configuration route layouts (FR-051–FR-055). |
| **Frontend Language** | **TypeScript 5.x (Strict Mode)** | Type-safety penuh, sinkronisasi skema data dengan payload API Go backend, eliminasi runtime error pada modul finansial. |
| **Frontend UI & Styling** | **Tailwind CSS + Shadcn UI (Radix Primitives)** | Desain sistem berbasis token semantik (*Emerald, Amber, Crimson, Slate/Cyan*), kepatuhan aksesibilitas WCAG 2.1 AA, dan navigasi keyboard penuh (Section 9.3, 12.7). |
| **Frontend State & Fetching** | **Zustand + TanStack Query v5** | Zustand untuk local work session timer & tab visibility tracking; TanStack Query untuk caching server state dan polling/SSE live attendance (FR-011, FR-052). |
| **Primary Database** | **PostgreSQL 16+** | Single Source of Truth relasional ACID, native triggers untuk memblokir UPDATE/DELETE pada tabel immutable (`financial_ledgers`, `attendance_event_logs`, `audit_logs`), JSONB untuk Unified Policy Engine (BR-023, BR-024, FR-045). |
| **In-Memory Cache & Broker**| **Redis 7+** | Cache debouncing presensi 30 detik (`attendance:debounce:<user_id>`), rate limiting, session store, dan message broker BullMQ/Asynq (FR-008, Section 11.2). |
| **Object Storage** | **Cloudflare R2 (S3-Compatible API)** | Repositori aset global audio WAV/MP3, berkas bukti tugas (*Evidence*), foto profil, dan PDF sertifikat. Zero egress fees. Database SQL hanya mencatat metadata file (BR-024, FR-044). |
| **IoT Edge Controller** | **ESP32-S3 (Dual-Core LX7, 8MB PSRAM, 16MB Flash)** | Controller terminal presensi lobi, USB Host interface ke NFC reader, Wi-Fi 802.11 b/g/n, I2S audio driver, LittleFS persistent audio cache & offline queue. |
| **NFC Reader Hardware** | **ACR1552U (USB CCID / UART NFC Reader)** | Pembacaan kartu NFC ISO 14443 Type A/B, MIFARE, NTAG213/215/216 dengan respon cepat (<100ms) dan anti-collision support. |
| **Audio DAC & Output** | **MAX98357A I2S Mono DAC Amp + Speaker 4Ω 3W** | Audio playback digital I2S berdefinisi tinggi, amplifikasi kelas D tanpa distorsi analog. |

---

## 3. ARSITEKTUR IOT TERMINAL PRESENSI NFC PRODUCTION-READY

### 3.1 Diagram Blok Perangkat Keras Terminal (Hardware Topology)

```text
 ┌─────────────────────────────────────────────────────────────────────────────┐
 │                         DCISP IOT TERMINAL UNIT                             │
 │                                                                             │
 │   ┌───────────────────────┐                  ┌──────────────────────────┐   │
 │   │  ACR1552U NFC READER  │                  │  MAX98357A I2S DAC AMP   │   │
 │   │  (ISO14443A / MIFARE) │                  │  + 4Ω 3W Speaker Box     │   │
 │   └──────────┬────────────┘                  └────────────▲─────────────┘   │
 │              │ USB D+/D- (CCID Host)                      │ I2S (BCLK, LRC, │
 │              │                                            │      DIN, SD)   │
 │   ┌──────────▼────────────────────────────────────────────┴─────────────┐   │
 │   │                     ESP32-S3 CONTROLLER UNIT                        │   │
 │   │  • Xtensa Dual-Core 240MHz, 8MB PSRAM, 16MB Flash                   │   │
 │   │  • Storage: LittleFS Partition (8MB: Audio Cache + Offline Queue)   │   │
 │   │  • Networking: Wi-Fi 802.11 b/g/n + HTTPS Client + mTLS / HMAC      │   │
 │   │  • Logic Engine: Debounce, Mode Selector, Audio Player, Sync Worker │   │
 │   └──────────────────────────────┬──────────────────────────────────────┘   │
 │                                  │                                          │
 │   ┌──────────────────────────────▼──────────────────────────────────────┐   │
 │   │  STATUS PERIPHERALS: RGB LED Status Indicator + Mode Toggle Switch  │   │
 │   └─────────────────────────────────────────────────────────────────────┘   │
 └──────────────────────────────────┬──────────────────────────────────────────┘
                                    │ HTTPS (TLS 1.3)
                                    ▼
                         [ Golang Backend (Gin) ]
```

### 3.2 Alur Operasional Tap Kartu NFC (End-to-End Sequence)

```text
User / Intern              ACR1552U Reader          ESP32-S3 Controller             Go Backend (Gin)              Cloudflare R2 / Audio Cache
     │                            │                         │                              │                               │
     ├── 1. Tap Kartu NFC ───────►│                         │                              │                               │
     │                            ├── 2. Baca Card UID ────►│                              │                               │
     │                            │   (Payload Ingestion)   ├── 3. Play Audio 'card_read'  │                               │
     │                            │                         │   (Dari Local LittleFS) ─────┼──────────────────────────────►│ (I2S DAC Playback)
     │                            │                         │                              │                               │
     │                            │                         ├── 4. POST /api/v1/attendance/terminal-tap                     │
     │                            │                         │   (Headers: Device Auth, Body: terminal_id, card_uid, mode)  │
     │                            │                         │─────────────────────────────►│                               │
     │                            │                         │                              ├── 5. Validasi Kartu & Akun    │
     │                            │                         │                              ├── 6. Evaluasi Schedule Engine │
     │                            │                         │                              ├── 7. Tentukan Event & Anomali │
     │                            │                         │                              ├── 8. Simpan ke PostgreSQL     │
     │                            │                         │                              │   (Atomic ACID Transaction)   │
     │                            │                         │                              ├── 9. Dispatch XP & SSE Event  │
     │                            │                         │◄─────────────────────────────┤                               │
     │                            │                         │  10. Return JSON Response    │                               │
     │                            │                         │      { status: "SUCCESS",    │                               │
     │                            │                         │        event: "CHECK_IN",    │                               │
     │                            │                         │        audio_event: "late" / │                               │
     │                            │                         │        "check_in" }          │                               │
     │                            │                         │                              │                               │
     │                            │                         ├── 11. Trigger I2S Playback ──┼──────────────────────────────►│ (Putar Audio Event)
     │◄── 12. Dengar Suara Event (MAX98357A Speaker) ───────┤   (Check LittleFS Cache ->   │                               │
     │                                                      │    Fallback Download R2)     │                               │
```

---

### 3.3 Konfigurasi Event Audio Global & Transkrip Baku (15 Master Audio Events)

Sistem menggunakan **Audio Global Terstandarisasi** yang dipre-generate satu kali (WAV 16-bit 44.1kHz / MP3 mono 64kbps), disimpan permanen pada Cloudflare R2 bucket `audio-assets/global/`, dan dicache secara persisten di partisi LittleFS ESP32-S3:

| Key Event Audio | Transkrip Suara Resmi (Exact Audio Script) | Kondisi Pemicu dari Go Backend |
|---|---|---|
| `card_read` | **“Kartu terbaca. Stand by.”** | Diputar instan oleh ESP32 lokal sesaat setelah UID NFC terbaca reader sebelum respon API kembali. |
| `processing` | **“Sedang Memproses. One moment.”** | Diputar jika request jaringan backend membutuhkan waktu $> 800\text{ ms}$. |
| `device_ready` | **“System ready.”** | Diputar oleh terminal saat proses booting, sync audio, dan koneksi Wi-Fi berhasil terhubung. |
| `check_in` | **“Good Morning! Absensi berhasil.”** | Backend mencatat event `CHECK_IN` tepat waktu ($\le \text{start\_time} + \text{grace\_period}$). |
| `work_start` | **“Selamat bekerja. Have a productive day!.”** | Backend mencatat event `WORK_STARTED` saat intern memulai sesi kerja pertama hari itu. |
| `break_start` | **“Enjoy your break!.”** | Backend mencatat event `BREAK_STARTED` saat jam istirahat dimulai. |
| `break_end` | **“Welcome back! Silakan lanjut bekerja.”** | Backend mencatat event `BREAK_ENDED` dan `WORK_RESUMED` tepat waktu. |
| `check_out` | **“Thank you, See you tomorrow.”** | Backend mencatat event `CHECK_OUT` saat jam kepulangan kantor. |
| `card_unregistered` | **“Kartu belum terdaftar. Contact admin.”** | Backend mendeteksi UID kartu valid secara hardware tetapi belum terasosiasi dengan user aktif (`404 Not Found`). |
| `card_invalid` | **“Kartu tidak valid. Please try again.”** | Payload kartu rusak, checksum salah, atau kartu diblokir administratif (`400 Bad Request`). |
| `save_failed` | **“Absensi gagal disimpan. Please try again.”** | Backend mengalami kegagalan internal database rollback saat persistensi event (`500 Server Error`). |
| `offline` | **“Sistem offline. Check the connection.”** | Terminal gagal terhubung ke Wi-Fi / Gateway dan antrean offline penuh. |
| `offline_success` | **“Absensi berhasil dicatat. Data akan disinkronkan saat koneksi kembali.”** | Jaringan internet backend terputus; ESP32 berhasil menyimpan tap ke antrean LittleFS lokal. |
| `too_frequent` | **“Terlalu cepat. Please wait.”** | Kartu yang sama di-tap ulang dalam jendela debounce 30 detik (pencegahan duplicate scan). |
| `late` | **“Perhatian. Anda terlambat.”** | Backend mencatat event `CHECK_IN` melebihi grace period (memicu penalti XP Level 1-3). |

---

### 3.4 Mode Operasional Terminal (`Terminal Modes`)

ESP32-S3 mendukung 5 mode operasional yang dikonfigurasi melalui remote backend config atau hardware switch:
1. **`CHECK_IN`:** Terminal khusus gerbang masuk. Seluruh tap diproses strictly sebagai peristiwa kedatangan/check-in.
2. **`CHECK_OUT`:** Terminal khusus pintu keluar. Seluruh tap diproses strictly sebagai check-out kepulangan.
3. **`BREAK_START`:** Terminal area kantin/istirahat saat jeda makan siang dimulai.
4. **`BREAK_END`:** Terminal area kantin/pintu kerja saat kembali dari istirahat.
5. **`AUTO` (Default Intelligent Mode):** Backend mengevaluasi konteks secara dinamis:
   - Jika belum check-in hari ini $\rightarrow$ `CHECK_IN`.
   - Jika sedang `WORKING` dan waktu memasuki jendela break (12:00–13:00) $\rightarrow$ `BREAK_START`.
   - Jika sedang `BREAK` $\rightarrow$ `BREAK_END` & `WORK_RESUMED`.
   - Jika sudah bekerja $\ge \text{jadwal reguler}$ atau waktu kepulangan $\ge 17:00$ $\rightarrow$ `CHECK_OUT`.

---

### 3.5 Protokol Offline, Local Ring-Buffer & Idempotent Synchronization

Jika jaringan internet atau backend terputus saat tap kartu terjadi:
1. **Penyimpanan Lokal:** ESP32-S3 menuliskan rekaman ke file antrean `offline_queue.bin` di LittleFS:
   ```c
   struct OfflineRecord {
       char terminal_id[37];
       char card_uid[32];
       uint32_t local_timestamp; // Epoch time RTC
       uint8_t mode;
       char idempotency_key[65]; // SHA-256(terminal_id + card_uid + timestamp)
   };
   ```
2. **Umpan Balik Instan:** ESP32 memutar audio `offline_success` secara lokal dari LittleFS flash.
3. **Protokol Sinkronisasi Otomatis:**
   - Background task di ESP32 memantau status jaringan setiap 10 detik.
   - Saat online kembali, ESP32 mengirim batch payload ke endpoint Go Backend: `POST /api/v1/attendance/sync-offline`.
   - Backend memproses setiap record secara transaksional dengan klausa **Idempotency Guard**:
     ```sql
     INSERT INTO attendance_event_logs (id, user_id, event_type, timestamp, method, device_id, metadata)
     VALUES ($1, $2, $3, $4, 'NFC_OFFLINE_SYNC', $5, $6)
     ON CONFLICT (idempotency_key) DO NOTHING;
     ```
   - Setelah menerima respon `200 OK (all synced)`, ESP32 menghapus record antrean lokal.

---

### 3.6 Siklus Manajemen Perangkat (Heartbeat, Audio Versioning & OTA)

1. **Heartbeat Protocol (Tiap 30 Detik):**
   - ESP32 mengirim telemetry: `POST /api/v1/devices/heartbeat` memuat: `terminal_id`, `firmware_version`, `audio_catalog_version`, `wifi_rssi`, `free_heap`, `offline_queue_len`.
2. **Sinkronisasi Katalog Audio (Checksum Verification):**
   - Backend menyediakan manifest audio pada `GET /api/v1/attendance/audio-catalog`:
     ```json
     {
       "version": "1.0.4",
       "base_url": "https://pub-r2.dcisp.internal/audio-assets/global/",
       "files": {
         "card_read": { "filename": "card_read.mp3", "sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" },
         "check_in": { "filename": "check_in.mp3", "sha256": "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e" }
       }
     }
     ```
   - Saat boot atau menerima update config, ESP32 membandingkan checksum SHA-256 lokal di LittleFS. Jika ada perubahan atau file hilang, ESP32 mendownload file baru langsung dari storage R2 dan menyimpannya secara persisten.
3. **Over-The-Air (OTA) Firmware Updates:**
   - ESP32 mendukung pembaruan firmware terenkripsi secara remote melalui endpoint `/api/v1/devices/ota-update` dengan rollback otomatis jika booting firmware baru gagal (*dual OTA partition table*).

---

## 4. BACKEND ARCHITECTURE (GOLANG + GIN MODULAR MONOLITH)

### 4.1 Struktur Direktori & Pemisahan Domain Backend

```text
backend-go/
├── cmd/
│   └── api/
│       └── main.go                     # Entrypoint server, dependency wiring
├── internal/
│   ├── config/                         # Environment & runtime policy config loader
│   ├── database/                       # PostgreSQL pgx connection pool & migrations
│   ├── middleware/
│   │   ├── auth_jwt.go                 # Stateless JWT extractor & validator
│   │   ├── dynamic_rbac.go             # 5-Tier RBAC & Scope enforcement
│   │   ├── device_auth.go              # HMAC-SHA256 device signature validator
│   │   ├── rate_limiter.go             # Redis sliding window rate limiter
│   │   └── audit_interceptor.go        # Append-only security audit logger
│   ├── modules/
│   │   ├── identity/                   # Users, Auth, Roles, Permissions, Scopes
│   │   ├── people/                     # Interns, Alumni, Batches, Institutions, Skills
│   │   ├── attendance/                 # NFC Ingestion, Schedule Engine, Work Sessions, Breaks, Overtime
│   │   │   ├── controller.go           # Gin handlers (terminal-tap, sync-offline, audio-catalog)
│   │   │   ├── service.go              # Business validation, schedule checks, state transitions
│   │   │   ├── repository.go           # PostgreSQL queries (DATA-001, DATA-002)
│   │   │   └── audio_catalog.go        # Global audio manifests & checksum manager
│   │   ├── projects/                   # Marketplace, Teams, Milestones, Tasks, Evidence, Submissions
│   │   ├── performance/                # XP Engine (3 partitions), Ranks, Formal Evaluation, Top Performer
│   │   ├── incentives/                 # Multi-component Rank Rewards, Bounty Distribution
│   │   ├── finance/                    # Deduction/Tax Engine, Wallets, Batch Fund, Double-Entry Ledger
│   │   ├── documents/                  # Cloudflare R2 Presigned Bridge, ID Cards, Certificates, Portfolios
│   │   ├── system/                     # Unified Policy Engine (JSONB), Settings, Audit Logs
│   │   └── intelligence/               # Aggregated Analytics & AI Informational Engine
│   └── shared/
│       ├── eventbus/                   # Internal async event dispatcher
│       ├── storage/                    # Cloudflare R2 S3-client wrapper
│       └── response/                   # Standardized JSON response envelope
├── pkg/
│   └── worker/                         # Asynq Background Job Processors
└── go.mod
```

---

### 4.2 Alur Logika Backend pada Ingestion Terminal Presensi (Gin Handler)

```go
// internal/modules/attendance/controller.go (Cuplikan Logika Inti)
func (c *AttendanceController) HandleTerminalTap(ctx *gin.Context) {
    var req TerminalTapRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, response.Error(400, "Invalid payload", err.Error()))
        return
    }

    // 1. Debouncing Check (Redis Key: attendance:debounce:<card_uid>, TTL: 30s)
    if isDebounced(req.CardUID) {
        ctx.JSON(http.StatusOK, gin.H{
            "status": "DEBOUNCED",
            "audio_event": "too_frequent",
            "message": "Scan ignored: debouncing active",
        })
        return
    }

    // 2. Process Business Logic in Service Layer
    result, err := c.attendanceService.ProcessTap(ctx.Request.Context(), req)
    if err != nil {
        if errors.Is(err, ErrCardUnregistered) {
            ctx.JSON(http.StatusNotFound, gin.H{"status": "ERROR", "audio_event": "card_unregistered"})
            return
        }
        if errors.Is(err, ErrCardInvalid) {
            ctx.JSON(http.StatusBadRequest, gin.H{"status": "ERROR", "audio_event": "card_invalid"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"status": "FAILED", "audio_event": "save_failed"})
        return
    }

    // 3. Set Debounce in Redis
    setDebounce(req.CardUID, 30*time.Second)

    // 4. Return Success Response with Appropriate Audio Event
    ctx.JSON(http.StatusOK, gin.H{
        "status": "SUCCESS",
        "event_type": result.EventType,
        "is_late": result.IsLate,
        "audio_event": result.AudioEvent, // e.g. "check_in", "late", "break_start", "check_out"
        "data": result.AttendanceRecord,
    })
}
```

---

## 5. FRONTEND ARCHITECTURE (NEXT.JS 15+ APP ROUTER + TYPESCRIPT)

### 5.1 Struktur Direktori Frontend

```text
frontend-next/
├── app/
│   ├── (auth)/
│   │   └── login/page.tsx              # Halaman Login Multi-Peran
│   ├── (dashboard)/
│   │   ├── layout.tsx                  # Global Collapsible Sidebar & Header Live Status
│   │   ├── command-center/page.tsx     # Executive Cockpit 4 Kuadran (FR-051)
│   │   ├── my-day/page.tsx             # Intern Cockpit: Live Session, Task, Break (FR-052)
│   │   ├── team-today/page.tsx         # Supervisor Live Monitoring & Anomalies (FR-053)
│   │   ├── supervisor/workspace/page.tsx # Persetujuan Lembur, Koreksi & Evaluasi (FR-054)
│   │   ├── attendance/
│   │   │   ├── live/page.tsx           # Monitor Kehadiran Kantor
│   │   │   ├── scanner/page.tsx        # View Terminal Pemindai Khusus Operator (BR-002)
│   │   │   ├── corrections/page.tsx    # Pengajuan & Approval Koreksi Presensi
│   │   │   └── overtime/page.tsx       # Pengajuan & Persetujuan Lembur
│   │   ├── projects/
│   │   │   ├── marketplace/page.tsx    # Bursa Proyek Terbuka (FR-016)
│   │   │   └── [id]/tasks/page.tsx     # Papan Kanban Tugas Proyek
│   │   ├── finance/
│   │   │   ├── wallet/page.tsx         # Dompet Personal & Mutasi Itemized
│   │   │   └── ledger/page.tsx         # Audit Buku Kas Ganda Immutabel
│   │   └── system/
│   │       └── policies/page.tsx       # Unified Policy Engine Configurator
│   └── layout.tsx
├── components/
│   ├── ui/                             # Shadcn UI (Buttons, Tables, Modals, Badges)
│   ├── cockpits/
│   │   ├── precision-timer.tsx         # Monospace Timer Aktif Sesi Kerja
│   │   ├── session-integrity-tracker.tsx # Client Hook Page Visibility (BR-010)
│   │   └── dynamic-sidebar.tsx         # Sidebar Dinamis Berbasis RBAC Guard
├── hooks/
│   ├── use-work-session.ts             # State Machine Kontrol Sesi Kerja (Zustand)
│   ├── use-live-attendance.ts          # Server-Sent Events (SSE) Listener
│   └── use-session-integrity.ts        # Tab visibility & Idle Detection (>15 min)
└── lib/
    ├── api-client.ts                   # Fetch/Axios wrapper dengan auto refresh token
    └── types/                          # Shared TypeScript Interfaces dari backend
```

---

## 6. DATABASE SCHEMA & IMMUTABILITY ENFORCEMENTS (POSTGRESQL 16+)

### 6.1 Skema Data Kritis Presensi, Perangkat & Finansial

```sql
-- 1. TERMINAL & DEVICE REGISTRY (FR-013)
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    terminal_identifier VARCHAR(64) UNIQUE NOT NULL,
    device_type VARCHAR(32) NOT NULL, -- 'ESP32_NFC_TERMINAL', 'CAMERA_SCANNER'
    location_name VARCHAR(128) NOT NULL,
    api_key_hash VARCHAR(64) NOT NULL,
    current_mode VARCHAR(32) DEFAULT 'AUTO', -- 'CHECK_IN', 'CHECK_OUT', 'BREAK_START', 'BREAK_END', 'AUTO'
    firmware_version VARCHAR(32) NOT NULL,
    audio_catalog_version VARCHAR(32) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_heartbeat_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 2. ATTENDANCE EVENT LOGS (DATA-001 & BR-006)
CREATE TABLE attendance_event_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    device_id UUID REFERENCES devices(id),
    event_type VARCHAR(32) NOT NULL, -- 'ARRIVED', 'CHECK_IN', 'WORK_STARTED', 'BREAK_STARTED', 'BREAK_ENDED', 'WORK_RESUMED', 'OVERTIME_STARTED', 'OVERTIME_ENDED', 'CHECK_OUT'
    timestamp TIMESTAMPTZ NOT NULL,
    method VARCHAR(32) NOT NULL, -- 'NFC', 'QR_CODE', 'ADMIN_SCANNER', 'NFC_OFFLINE_SYNC', 'MANUAL_CORRECTION'
    location_context VARCHAR(128),
    session_id UUID,
    idempotency_key VARCHAR(64) UNIQUE,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. FINANCIAL DOUBLE-ENTRY LEDGER (DATA-006, FR-037 & BR-023)
CREATE TABLE financial_ledgers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    account_code VARCHAR(32) NOT NULL, -- '1001-CASH', '2001-LIABILITY-INTERN', '2002-TAX-PAYABLE', '3001-BATCH-FUND'
    direction VARCHAR(8) NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    reference_table VARCHAR(64) NOT NULL,
    reference_id UUID NOT NULL,
    narration TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 4. FILE METADATA RECORD (DATA-007 & BR-024)
CREATE TABLE file_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    disk VARCHAR(32) DEFAULT 'r2',
    path_key VARCHAR(512) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(128) NOT NULL,
    size_bytes BIGINT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

### 6.2 PostgreSQL Native Immutability Trigger (Penegakan Append-Only Mutlak)

```sql
-- Trigger Function Memblokir Operasi UPDATE dan DELETE pada Tabel Append-Only
CREATE OR REPLACE FUNCTION enforce_immutable_records()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'DCISP Security Violation: Modification or Deletion of Immutable Records (%) is strictly prohibited by BR-023 / BR-025 / FR-047', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_immutable_attendance
BEFORE UPDATE OR DELETE ON attendance_event_logs
FOR EACH ROW EXECUTE FUNCTION enforce_immutable_records();

CREATE TRIGGER trg_immutable_financial_ledger
BEFORE UPDATE OR DELETE ON financial_ledgers
FOR EACH ROW EXECUTE FUNCTION enforce_immutable_records();

CREATE TRIGGER trg_immutable_audit_logs
BEFORE UPDATE OR DELETE ON audit_logs
FOR EACH ROW EXECUTE FUNCTION enforce_immutable_records();

CREATE TRIGGER trg_immutable_xp_transactions
BEFORE UPDATE OR DELETE ON xp_transactions
FOR EACH ROW EXECUTE FUNCTION enforce_immutable_records();
```

---

## 7. API CONTRACTS & PROTOCOLS

### 7.1 Ingestion Presensi Terminal (`POST /api/v1/attendance/terminal-tap`)
* **Headers:** `X-Device-ID: UUID`, `X-DCISP-Signature: HMAC-SHA256(device_token, body)`
* **Request Payload:**
  ```json
  {
    "card_uid": "04A1B2C3D4E5F6",
    "timestamp": 1774087200,
    "terminal_mode": "AUTO",
    "idempotency_key": "9f83c613c78cd4a123f14bc3a18ef77a641a92e105e60803dd35c4d093257a01"
  }
  ```
* **Response Payload (Success - On Time):**
  ```json
  {
    "success": true,
    "statusCode": 200,
    "status": "SUCCESS",
    "event_type": "CHECK_IN",
    "is_late": false,
    "audio_event": "check_in",
    "message": "Presensi masuk berhasil dicatat tepat waktu",
    "data": {
      "user_name": "Andi Pratama",
      "timestamp": "2026-09-19T08:28:45+07:00",
      "earned_xp": 10
    }
  }
  ```
* **Response Payload (Success - Late):**
  ```json
  {
    "success": true,
    "statusCode": 200,
    "status": "SUCCESS",
    "event_type": "CHECK_IN",
    "is_late": true,
    "audio_event": "late",
    "message": "Presensi masuk tercatat: Terlambat 18 menit",
    "data": {
      "user_name": "Andi Pratama",
      "timestamp": "2026-09-19T08:48:12+07:00",
      "penalty_xp": -2
    }
  }
  ```

---

## 8. MATRIX KEPATUHAN ATURAN BISNIS PRD (BR-001 S/D BR-025)

| ID Aturan Bisnis | Nama Aturan | Status Implementasi Teknis |
|---|---|---|
| **BR-001** | Dynamic RBAC (Zero Hardcoding) | Ditegakkan via database table `roles`, `permissions`, `scopes` + Gin RBAC Middleware. |
| **BR-002** | Scanner Operator Isolation | Scope `scanner_only` dibatasi khusus ke endpoint `/api/v1/attendance/scanner/*`. |
| **BR-003** | Alumni Account Permanence | Akun alumni retained selamanya di database; role bertransisi ke `ALUMNI`. |
| **BR-004** | 3 XP Tracks Isolation | Partisi skema `internship_xp`, `project_xp`, `alumni_xp` di tabel `xp_transactions`. |
| **BR-005** | Work Type Partition | Pemisahan entitas database `daily_works`, `tasks`, `work_reports`. |
| **BR-006** | Central Attendance Ingestion | Terminal IoT, dynamic QR, dan web scanner bermuara pada satu service `attendance.ProcessTap`. |
| **BR-007** | 5-Way Time Differentiation | Perhitungan mandiri durasi: Attendance, Work Session, Active Focus, Break, Overtime. |
| **BR-008** | Break State Engine | Mesin status eksplisit: `WORKING -> BREAK -> WORKING` di table `breaks`. |
| **BR-009** | Break Anomaly Enforcement | Validasi schedule: Istirahat awal = flag anomali; Resume terlambat = penalti -2 XP. |
| **BR-010** | Privacy-First Tracking | Larangan screenshot/keylogger; murni menggunakan W3C Page Visibility API di Next.js client. |
| **BR-011** | Non-Punitive Tab Blur | Tab switching hanya menjeda active focus timer tanpa pengurangan poin XP instan. |
| **BR-012** | Configurable XP Penalties | Matriks penalti dikendalikan dari JSONB `Unified Policy Engine` (FR-045). |
| **BR-013** | Schedule-Governed Overtime | Jam lembur wajib diajukan di muka dan diverifikasi terhadap `approved_window`. |
| **BR-014** | Actual Worked Overtime | Kompensasi lembur dihitung dari jam aktual yang dikerjakan ($\le \text{Approved OT}$). |
| **BR-015** | Informational AI Non-Punitive | Skor skill match murni rekomendasi advisory; larangan auto-reject kandidat. |
| **BR-016** | 3-Tier Contribution Model | Rekonsiliasi 3 lapis: `Planned -> Actual (algoritma) -> Final (disahkan supervisor)`. |
| **BR-017** | Rank vs Performance Score | Pemisahan mutlak: $\text{Rank (Poin XP)} \neq \text{Performance Score (Evaluasi Formal 0-100)}$. |
| **BR-018** | Singular Top Performer per Batch | Tepat SATU juara per kohort batch per periode evaluasi via formula komposit. |
| **BR-019** | Multi-Component Rank Rewards | Paket hadiah promosi rank: Uang Kas + XP + Merchandise Fisik + Badges + Vouchers. |
| **BR-020** | Dynamic Deduction Engine | Seluruh distribusi finansial wajib melewati evaluasi aturan pajak/biaya dinamis. |
| **BR-021** | Farewell Fund Classification | Iuran kas perpisahan disetor ke akun `Batch Fund`, terpisah total dari pajak resmi negara. |
| **BR-022** | Itemized Wallet Ledger | Dompet mencatat mutasi *gross*, deduksi, *net*, dan referensi ID sumber dana secara transparan. |
| **BR-023** | Mandatory Double-Entry Ledger | Buku kas ganda berpasangan ($\sum \text{Debit} = \sum \text{Kredit}$) dilindungi database trigger append-only. |
| **BR-024** | Cloudflare R2 Storage Separation | Berkas fisik disimpan di R2; database PostgreSQL hanya menyimpan metadata berkas. |
| **BR-025** | Audit Preservation on Exception | Koreksi absensi manual membuat event `MANUAL_CORRECTION` baru tanpa menimpa log asli. |

---

## 9. IMPLEMENTATION ROADMAP & CHECKLIST

- [x] **Tech Stack Finalized:** Backend Golang (Gin) + Frontend Next.js (TypeScript) + PostgreSQL 16 + Redis 7 + Cloudflare R2.
- [x] **IoT Terminal Specification:** ESP32-S3 + ACR1552U + MAX98357A + 15 Global Pre-Generated Audio Catalog + Offline Queue.
- [x] **Security & Integrity Strategy:** PostgreSQL Append-Only Triggers + Dynamic 5-Tier RBAC + Privacy-First Page Visibility Hook.
- [x] **PRD Alignment:** 100% selaras dengan 25 Aturan Bisnis (BR-001 s/d BR-025) dan 55 Kebutuhan Fungsional (FR-001 s/d FR-055).

---
*Dokumen spesifikasi arsitektur final ini disimpan secara permanen di `FINAL-TECH-STACK-SPEC-DCISP.md`.*
