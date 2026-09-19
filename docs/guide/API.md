# DCISP v1.0 API SPECIFICATION & CONTRACT
**Platform:** Dagang Creative Intern Solutions Program (DCISP)  
**Backend:** Golang 1.23+ (Gin Framework)  
**Frontend:** Next.js 15+ (App Router, TypeScript)  
**Source of Truth:** `docs/prd/PRD-DCISP-V1.md`, `docs/architecture/FINAL-TECH-STACK-SPEC-DCISP.md`, Suite 9 BRDs (`docs/brd/`)

---

## 1. API Principles

### 1.1 API Style & Protocol
* **Architecture:** RESTful HTTP/JSON Web API compliant with OpenAPI 3.0 standards.
* **Data Transport:** JSON over HTTPS/TLS 1.3 for external clients; Server-Sent Events (SSE) for unidirectional real-time push streams.
* **Content-Type:** `application/json; charset=utf-8` for standard payloads; `multipart/form-data` strictly restricted to direct local uploads (primary upload uses Presigned Cloudflare R2 URLs).

### 1.2 Base URL & Versioning Convention
* **Base URL Pattern:** `/api/v1`
* **Versioning Strategy:** URI path versioning (`/api/v1/`, `/api/v2/`). Breaking changes require a major version bump. Minor backwards-compatible changes are added in-place.

### 1.3 Request & Response Envelope Convention
All API endpoints return a standardized JSON envelope structure:

#### Success Response Envelope
```json
{
  "success": true,
  "statusCode": 200,
  "message": "Operation completed successfully",
  "data": {},
  "meta": {
    "page": 1,
    "limit": 20,
    "totalItems": 100,
    "totalPages": 5
  }
}
```

#### Error Response Envelope
```json
{
  "success": false,
  "statusCode": 400,
  "error": "Bad Request",
  "message": "Validation failed on payload",
  "errors": [
    {
      "field": "email",
      "message": "email must be a valid email address"
    }
  ],
  "traceId": "c8a4f912-32b1-4b72-9b2f-9811abdc1234"
}
```

### 1.4 Authentication Mechanism
* **User Accounts:** Stateless JWT Bearer Token in `Authorization: Bearer <access_token>` header.
  - Access Token TTL: **15 minutes** (contains `user_id`, `email`, `role`, `scopes`).
  - Refresh Token TTL: **7 days** (stored hashed in database for revocation/rotation).
* **Device Gateway (Hardware Terminals):** Header-based authentication:
  - `X-Device-ID: <UUID>`
  - `X-Timestamp: <Unix Timestamp>`
  - `X-DCISP-Signature: HMAC-SHA256(api_key, method + path + timestamp + body)`

### 1.5 Authorization & RBAC
* Dynamic 5-Tier RBAC evaluation on every request:
  $$\text{Role} \longrightarrow \text{Permission} \longrightarrow \text{Scope} \longrightarrow \text{Resource} \longrightarrow \text{Action}$$
* **BR-001 (Zero Hardcoding):** Permissions are resolved dynamically at runtime from PostgreSQL/Redis cache.
* **BR-002 (Scanner Operator Isolation):** Role `SCANNER_OPERATOR` is restricted to `scope = scanner_only` and endpoints under `/api/v1/attendance/scanner/*` and `/api/v1/attendance/terminal-tap`. All other endpoints return `403 Forbidden`.

### 1.6 Pagination, Filtering, Sorting & Searching
* **Pagination Parameters:** `page` (default: 1, min: 1), `limit` (default: 20, max: 100).
* **Sorting Parameters:** `sort_by` (e.g., `created_at`, `timestamp`), `order` (`asc` | `desc`).
* **Filtering Parameters:** Query params using snake_case (e.g., `status=ACTIVE&batch_id=<UUID>`).
* **Searching:** `q` parameter (e.g., `q=john+doe`), evaluated using PostgreSQL `pg_trgm` indexes.

### 1.7 Rate Limiting
* `/api/v1/auth/login`: 10 requests / minute per IP.
* `/api/v1/attendance/terminal-tap`: 120 requests / minute per Device ID.
* General API routes: 120 requests / minute per IP / authenticated user.
* Enforced via Redis sliding window counter; returns `429 Too Many Requests` on breach.

### 1.8 Idempotency & Debouncing
* **NFC/QR Tap Debouncing (BR-008):** 30-second window in Redis per `card_uid`. Duplicate scans return `status: "DEBOUNCED"` with `audio_event: "too_frequent"`.
* **Idempotency Key:** Mandatory for financial payouts and offline attendance sync via `X-Idempotency-Key` header (`SHA-256(device_id + user_id + timestamp_minute)`).

---

## 2. Endpoint Organization

```text
/api/v1
├── /auth               # Identity & Authentication
├── /people             # Interns, Alumni, Batches, Institutions, Skills
├── /attendance         # NFC Ingestion, Schedule, Sessions, Breaks, Overtime, Leave, Corrections
├── /projects           # Marketplace, Quota, Teams, Milestones, Tasks, Submissions, Evidence
├── /performance        # XP Rules, Rank Progression, Evaluations, Top Performer, Badges
├── /finance            # Wallets, Bounty Split, Batch Fund, Taxes, Ledger, Payouts
├── /documents          # ID Cards, Certificates, Portfolios, Reports, Storage Presigned URLs
├── /system             # Policy Engine, Audit Logs, Settings, Notification Center, SSE Stream
└── /intelligence       # Analytics, AI Activity Insights
```

---

## 3. Detailed Endpoint Contracts

### 3.1 Domain 1: Identity & Authentication (`/api/v1/auth`)

#### POST /api/v1/auth/login
* **Purpose:** Authenticates user credentials via Argon2id hash verification and issues JWT token pair.
* **Authentication:** Public (Rate limited: 10 req/min)
* **Request Body:**
  ```json
  {
    "email": "intern@dcisp.internal",
    "password": "SecurePassword2026!"
  }
  ```
* **Validation:** `email` (valid email format, required), `password` (string, min 6 chars, required).
* **Response (200 OK):**
  ```json
  {
    "success": true,
    "statusCode": 200,
    "message": "Login successful",
    "data": {
      "tokens": {
        "access_token": "eyJhbGciOi...",
        "refresh_token": "eyJhbGciOi...",
        "expires_in": 900,
        "token_type": "Bearer"
      },
      "profile": {
        "id": "c8a4f912-32b1-4b72-9b2f-9811abdc1234",
        "email": "intern@dcisp.internal",
        "full_name": "Andi Pratama",
        "status": "ACTIVE",
        "role": "INTERN",
        "scopes": ["OWN_DATA"],
        "permissions": ["attendance.own.log", "work_session.own.manage"]
      }
    }
  }
  ```
* **Possible Errors:** `400 Bad Request` (malformed JSON), `401 Unauthorized` (invalid credentials or inactive account), `429 Too Many Requests`.

#### POST /api/v1/auth/refresh
* **Purpose:** Issues a new access token using a valid, unrevoked refresh token (Token Rotation).
* **Authentication:** Public
* **Request Body:** `{ "refresh_token": "string" }`
* **Response (200 OK):** Returns new `TokenPair`.
* **Possible Errors:** `401 Unauthorized` (expired, tampered, or revoked token).

#### GET /api/v1/auth/me
* **Purpose:** Retrieves authenticated user profile, active roles, scopes, and granted permissions.
* **Authentication:** Required (`Bearer JWT`)
* **Response (200 OK):** Returns `UserProfileResponse`.
* **Possible Errors:** `401 Unauthorized`.

---

### 3.2 Domain 2: People & Lifecycles (`/api/v1/people`, `/api/v1/batches`, `/api/v1/skills`)

#### GET /api/v1/people/interns
* **Purpose:** Lists active and graduated interns with filtering.
* **Authentication:** Required
* **Authorization:** `RequirePermission('intern.manage', 'view', 'WORKFORCE_AND_PEOPLE', 'SYSTEM')`
* **Query Parameters:** `batch_id` (UUID), `status` (`APPLICANT`|`ACTIVE`|`GRADUATED`|`ON_LEAVE`), `q` (search name/NIM), `page`, `limit`.
* **Response (200 OK):** Returns array of intern entities with institution and rank info.

#### POST /api/v1/batches
* **Purpose:** Creates a new intern cohort batch and automatically initializes its `Batch Fund` (BR-021).
* **Authentication:** Required
* **Authorization:** `RequirePermission('batch.manage', 'create', 'WORKFORCE_AND_PEOPLE', 'SYSTEM')`
* **Request Body:**
  ```json
  {
    "batch_code": "BATCH-2026-01",
    "name": "Spring 2026 Cohort",
    "start_date": "2026-01-15",
    "end_date": "2026-06-15",
    "quota": 30
  }
  ```
* **Response (201 Created):** Returns created batch entity and initialized `batch_fund_id`.

---

### 3.3 Domain 3: Workforce & Attendance (`/api/v1/attendance`, `/api/v1/work-sessions`)

#### POST /api/v1/attendance/terminal-tap
* **Purpose:** Ingests NFC card tap / QR scan from hardware terminals with 30s debouncing and schedule evaluation.
* **Authentication:** Device Authentication (`X-Device-ID`, `X-DCISP-Signature`) OR Operator Token.
* **Request Body:**
  ```json
  {
    "card_uid": "04A1B2C3D4E5F6",
    "timestamp": 1774087200,
    "terminal_mode": "AUTO",
    "idempotency_key": "9f83c613c78cd4a123f14bc3a18ef77a641a92e105e60803dd35c4d093257a01"
  }
  ```
* **Response (200 OK — On-Time):**
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
* **Response (200 OK — Late Arrival):**
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
* **Audio Event Values (15 Global Standard Events):**
  `card_read`, `processing`, `device_ready`, `check_in`, `work_start`, `break_start`, `break_end`, `check_out`, `card_unregistered`, `card_invalid`, `save_failed`, `offline`, `offline_success`, `too_frequent`, `late`.

#### POST /api/v1/attendance/sync-offline
* **Purpose:** Bulk synchronizes offline attendance records from ESP32 LittleFS local queue when connectivity is restored.
* **Authentication:** Device Authentication (`X-Device-ID`, `X-DCISP-Signature`)
* **Request Body:** Array of offline scan objects with `idempotency_key`.
* **Response (200 OK):** Returns summary of successfully synced records and duplicate records skipped (`ON CONFLICT DO NOTHING`).

#### GET /api/v1/attendance/audio-catalog
* **Purpose:** Provides SHA-256 checksum manifest of all 15 global audio files for ESP32 firmware synchronization.
* **Authentication:** Public / Device
* **Response (200 OK):** Version, base R2 URL, and mapping of 15 audio files with their SHA-256 hashes.

#### POST /api/v1/work-sessions/start
* **Purpose:** Starts the daily productive work session from portal "My Day" (Event: `WORK_STARTED`).
* **Authentication:** Required (`INTERN`, `ALUMNI`)
* **Request Body:** `{ "task_id": "UUID (optional)" }`
* **Validation:** User must have active `CHECK_IN` record on the current date (BR-007).
* **Response (200 OK):** Created `work_session` entity with timer started.

#### POST /api/v1/work-sessions/break
* **Purpose:** Toggles break mode (`BREAK_STARTED` / `BREAK_ENDED`).
* **Authentication:** Required
* **Request Body:** `{ "action": "START" | "RESUME" }`
* **Response (200 OK):** Updated break record. Enforces Early Break flag and Unauthorized Break penalty (-2 XP) if resuming late (BR-009).

#### POST /api/v1/attendance/overtime/request
* **Purpose:** Submits overtime request before regular schedule ends (FR-012, BR-013).
* **Authentication:** Required (`INTERN`)
* **Request Body:** `{ "project_id": "UUID", "date": "YYYY-MM-DD", "requested_start": "17:00", "requested_end": "19:30", "reason": "text" }`
* **Response (201 Created):** Created `overtime_request` with status `SUBMITTED`.

#### POST /api/v1/attendance/corrections
* **Purpose:** Submits manual attendance correction request with evidence without overwriting original logs (BR-025).
* **Authentication:** Required (`INTERN`)
* **Request Body:** `{ "target_date": "YYYY-MM-DD", "proposed_event_type": "CHECK_OUT", "proposed_timestamp": "ISO8601", "reason": "text (min 20 chars)", "evidence_file_id": "UUID" }`
* **Response (201 Created):** Created `attendance_correction` entity with status `SUBMITTED`.

---

### 3.4 Domain 4: Projects & Tasks (`/api/v1/projects`, `/api/v1/tasks`)

#### GET /api/v1/projects
* **Purpose:** Explores Project Marketplace with visibility filtering.
* **Authentication:** Required
* **Query Parameters:** `visibility` (`INTERN_ONLY`|`PUBLIC`|`PRIVATE`), `status` (`PUBLISHED`|`IN_PROGRESS`), `page`, `limit`.
* **Rules:** Alumni can only view `PUBLIC` projects (BR-003).

#### POST /api/v1/projects/:id/apply
* **Purpose:** Applies for a project. Triggers skill compatibility divination (FR-018) and locks quota atomically upon reaching capacity (FR-017).
* **Authentication:** Required (`INTERN`, `ALUMNI`)
* **Request Body:** `{ "cover_letter": "text" }`
* **Response (201 Created):** Created `project_application` record.

#### POST /api/v1/tasks/:id/submissions
* **Purpose:** Submits work report and attaches evidence artifacts in Cloudflare R2 (DATA-004, FR-022, FR-023).
* **Authentication:** Required (Task Assignee)
* **Request Body:**
  ```json
  {
    "progress_percentage": 100,
    "what_i_did": "Implemented JWT Auth Guard and RBAC Middleware",
    "evidence_type": "GIT_COMMIT",
    "evidence_url_or_key": "https://github.com/company/repo/commit/abc1234",
    "problems": "None encountered",
    "next_actions": "Proceed with Attendance Engine tests"
  }
  ```
* **Response (201 Created):** Task status updated to `IN_REVIEW`.

#### POST /api/v1/projects/:id/contributions/finalize
* **Purpose:** Supervisor confirms Layer 3 Final Contribution % (sum must equal 100.00%) to unlock bounty distribution (FR-024, BR-016).
* **Authentication:** Required (`SUPERVISOR`)
* **Request Body:**
  ```json
  {
    "contributions": [
      { "user_id": "UUID-1", "final_contribution_pct": 40.00 },
      { "user_id": "UUID-2", "final_contribution_pct": 35.00 },
      { "user_id": "UUID-3", "final_contribution_pct": 25.00 }
    ],
    "notes": "Verified against task completion velocity"
  }
  ```
* **Validation:** $\sum \text{final\_contribution\_pct} == 100.00$.
* **Response (200 OK):** Contributions locked (`is_locked = true`).

---

### 3.5 Domain 5: Performance & Gamification (`/api/v1/performance`, `/api/v1/ranks`, `/api/v1/xp`)

#### GET /api/v1/performance/my-stats
* **Purpose:** Retrieves authenticated user's XP balances partitioned into 3 isolated schemes (BR-004), current Rank level, and rank progress.
* **Authentication:** Required
* **Response (200 OK):**
  ```json
  {
    "success": true,
    "statusCode": 200,
    "data": {
      "internship_xp": 850,
      "project_xp": 450,
      "alumni_xp": 0,
      "current_rank": {
        "tier": 3,
        "name": "Knight",
        "min_xp": 750,
        "next_rank_xp": 1500,
        "progress_pct": 13.33
      }
    }
  }
  ```

#### POST /api/v1/performance/evaluations
* **Purpose:** Submits formal supervisor evaluation rubric producing composite `Performance Score` (0–100) distinct from Rank (BR-017, FR-027).
* **Authentication:** Required (`SUPERVISOR`)
* **Request Body:** `{ "user_id": "UUID", "batch_id": "UUID", "period_type": "MONTHLY", "rubric_scores": { "attendance": 95.0, "task_delivery": 90.0, "work_quality": 92.0, "rubric": 90.0 }, "notes": "text" }`
* **Response (201 Created):** Created `performance_evaluation` entity.

---

### 3.6 Domain 6: Incentives & Finance (`/api/v1/finance`, `/api/v1/wallets`, `/api/v1/payouts`)

#### POST /api/v1/finance/bounty/distribute
* **Purpose:** Executes atomic financial flow ($\text{Gross} - \text{Tax} - \text{Farewell} = \text{Net}$) and posts double-entry ledger entries (FR-032, FR-034, FR-037, FR-039, BR-021, BR-023).
* **Authentication:** Required (`FINANCE`, `SUPER_ADMIN`)
* **Request Body:** `{ "project_id": "UUID" }`
* **Response (200 OK):** Returns calculation breakdown, personal wallet credits, batch fund credit, and balanced `transaction_id`.

#### GET /api/v1/finance/wallets/me
* **Purpose:** Retrieves personal wallet balance and itemized transaction history stream (DATA-006, BR-022).
* **Authentication:** Required
* **Response (200 OK):** Balance, currency (`IDR`), and array of transactions with `gross`, `deduction`, `net`, and `source`.

#### POST /api/v1/finance/payouts
* **Purpose:** Requests cashout from personal wallet to external bank account (FR-038).
* **Authentication:** Required
* **Request Body:** `{ "amount": 500000.00, "bank_code": "BCA", "account_number": "1234567890", "account_holder_name": "Andi Pratama" }`
* **Validation:** Amount $\le$ available balance (wallet status holds funds).
* **Response (201 Created):** Created `payout` record with status `REQUESTED`.

---

### 3.7 Domain 7: Documents & Storage (`/api/v1/documents`, `/api/v1/storage`)

#### POST /api/v1/storage/presigned-upload
* **Purpose:** Generates secure Presigned PUT URL for direct file upload to Cloudflare R2 (BR-024, FR-044).
* **Authentication:** Required
* **Request Body:** `{ "folder": "task-evidence", "filename": "screenshot.png", "mime_type": "image/png", "size_bytes": 2048000 }`
* **Response (200 OK):**
  ```json
  {
    "success": true,
    "statusCode": 200,
    "data": {
      "upload_url": "https://<account>.r2.cloudflarestorage.com/dcisp-vault/task-evidence/2026/09/uuid.png?X-Amz-Signature=...",
      "path_key": "task-evidence/2026/09/uuid.png",
      "expires_in_seconds": 900
    }
  }
  ```

#### GET /api/v1/documents/certificates/verify/:hash
* **Purpose:** Public endpoint to verify authenticity of digital graduation scroll.
* **Authentication:** Public
* **Response (200 OK):** Certificate metadata, graduate name, batch name, issuance date, and validity status.

---

### 3.8 Domain 8: System, Policies & Events (`/api/v1/policies`, `/api/v1/events`)

#### GET /api/v1/events/stream
* **Purpose:** Server-Sent Events (SSE) stream for real-time notification push and live attendance status broadcast.
* **Authentication:** Required (`Bearer JWT` via query param `token` or Header)
* **Response:** Continuous `text/event-stream` emitting JSON events (`attendance.scanned`, `notification.alert`, `session.break_warning`).

#### PUT /api/v1/policies/:id
* **Purpose:** Updates runtime rules on Unified Policy Engine without code deployment (BR-001, BR-012, BR-020, FR-045).
* **Authentication:** Required (`SUPER_ADMIN`)
* **Request Body:** `{ "condition_rules": {}, "action_definitions": {}, "is_active": true }`
* **Response (200 OK):** Policy updated and Redis rule cache flushed.

---

## 4. API Rules for Developers

1. **Always Return Standard Envelope:** Never return raw arrays or plain primitives; always encapsulate responses inside `APIResponse`.
2. **Never Hardcode Roles:** Always evaluate authorization through `middleware.RequirePermission(db, resource, action, scopes...)`.
3. **Respect Immutability:** Endpoints dealing with attendance events, audit logs, and financial ledgers must strictly issue `INSERT` statements. Never create `UPDATE` or `DELETE` endpoints for immutable tables.
4. **Presigned Upload Pattern:** Never stream large binary files through the Go API server. Always issue Presigned S3 URLs to Cloudflare R2.
5. **Idempotency on State Changing Ingestions:** Endpoints receiving hardware events or payments must enforce unique `idempotency_key` headers.
