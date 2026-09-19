# DCISP v1.0 SECURITY & PRIVACY SPECIFICATION
**Platform:** Dagang Creative Intern Solutions Program (DCISP)  
**Security Standard:** OWASP Top 10 & W3C Privacy Standards  
**Source of Truth:** `docs/prd/PRD-DCISP-V1.md` (Section 12.2, 12.3), `docs/architecture/FINAL-TECH-STACK-SPEC-DCISP.md`

---

## 1. Authentication Security

### 1.1 Password Hashing Architecture (Section 12.2)
* **Algorithm:** **Argon2id** (Memory-hard password hashing).
* **Cryptographic Parameters:**
  - Memory Cost: `64 MB` (`65,536 KB`)
  - Time Cost: `3 iterations`
  - Parallelism: `2 threads`
  - Salt Length: `16 bytes` (cryptographically random)
  - Hash Key Length: `32 bytes`
* Passwords must meet minimum complexity: at least 8 characters, containing uppercase, lowercase, numbers, and special characters.

### 1.2 Token Lifecycle & Session Security (Section 11.1)
* **Access Tokens:** Signed with HMAC-SHA256 (`HS256`).
  - Short-lived TTL: **15 minutes**.
  - Payload contains minimal context: `user_id`, `email`, `role`, `scopes`.
* **Refresh Tokens:**
  - TTL: **7 days**.
  - Stored hashed in PostgreSQL database.
  - Implements **Refresh Token Rotation**: each refresh request issues a new token pair and invalidates the previous refresh token.
  - Revocation: Admin or user can revoke active sessions immediately.

### 1.3 Hardware Device Gateway Authentication
* Hardware attendance terminals (ESP32-S3) authenticate via cryptographic signatures:
  $$\text{Signature} = \text{HMAC-SHA256}(\text{device\_token}, \text{HTTP\_Method} + \text{Path} + \text{Timestamp} + \text{Payload})$$
* Replay attack prevention: Backend rejects requests where `|Server_Time - Timestamp| > 60 seconds`.

---

## 2. Authorization & RBAC Isolation

### 2.1 Dynamic 5-Tier RBAC Evaluation (BR-001)
* Every API endpoint resolves access permissions dynamically:
  $$\text{Role} \longrightarrow \text{Permission} \longrightarrow \text{Scope} \longrightarrow \text{Resource} \longrightarrow \text{Action}$$
* Hardcoding role checks in code (e.g., `if user.role == 'ADMIN'`) is strictly forbidden.

### 2.2 Scanner Operator Isolation (BR-002)
* Role `SCANNER_OPERATOR` (Gatekeeper) is structurally isolated to:
  - Scope: `SCANNER_ONLY`
  - Permitted Resources: `attendance.scanner:view`, `attendance.scanner:scan`
* Any attempt by a scanner operator to query projects, user directories, evaluations, or financial ledgers is rejected with `403 Forbidden`.

---

## 3. Privacy-First Tracking Boundaries (BR-010, BR-011)

DCISP enforces strict architectural and legal privacy boundaries. The software is **NOT an employee surveillance tool**:

```text
┌─────────────────────────────────────────────────────────────┐
│ STRICTLY PROHIBITED IN CODEBASE (Zero Exceptions - BR-010)  │
│ ❌ NO Screenshot Capture (Desktop or Browser)               │
│ ❌ NO Keylogger or Keystroke Recording                      │
│ ❌ NO Tab Content or URL History Inspection                 │
│ ❌ NO Desktop Process Scanning                              │
│ ❌ NO Clipboard Reading or Copying                          │
│ ❌ NO Private Chat or Message Inspection                    │
└─────────────────────────────────────────────────────────────┘
```

### Allowed Integrity Monitoring:
* Standard browser events: `document.visibilityState` and `window.onblur/onfocus`.
* Client-side idle timer: Triggers `Idle (AFK)` state only when no mouse/keyboard interaction is detected within the DCISP tab for $> 15\text{ minutes}$.
* **BR-011 (Non-Punitive Tab Blur):** Tab switching does **NOT** automatically deduct XP points.

---

## 4. Input Validation & Attack Prevention

1. **SQL Injection:**
   - 100% of database queries must use parameterized placeholders (`$1`, `$2` via `pgx`).
   - String concatenation in SQL statements is strictly prohibited.
2. **Cross-Site Scripting (XSS):**
   - User inputs sanitized on ingest.
   - Output encoding handled natively by React/Next.js.
3. **Cross-Origin Resource Sharing (CORS):**
   - Configured strictly to trusted frontend origins in production.
4. **Security HTTP Headers:**
   ```text
   Content-Security-Policy: default-src 'self'; img-src 'self' data: https://*.r2.cloudflarestorage.com;
   X-Frame-Options: DENY
   X-Content-Type-Options: nosniff
   Strict-Transport-Security: max-age=31536000; includeSubDomains
   Referrer-Policy: strict-origin-when-cross-origin
   ```

---

## 5. File Upload Security (BR-024)

1. **Direct Presigned S3 Uploads:** Files are uploaded directly from client browser to Cloudflare R2 bucket using temporary Presigned PUT URLs (TTL: 15 minutes).
2. **Backend Validation Prior to Presigned URL Issuance:**
   - Maximum file size: `25 MB` for evidence/documents; `5 MB` for avatars.
   - Allowed MIME types: `image/jpeg`, `image/png`, `image/webp`, `application/pdf`.
3. **Database Isolation:** Database SQL only saves metadata records; binary data in SQL is prohibited.

---

## 6. Secrets & Environment Management

1. **Never Commit Secrets:** `.env` files, API keys, private certificates, and database passwords must be listed in `.gitignore`.
2. **Template Provided:** Commit only `.env.example` with dummy values.
3. **Key Rotation:** JWT secrets and Device API keys must support zero-downtime rotation.

---

## 7. Logging & Audit Standards

### 7.1 What MAY Be Logged:
* Request method, path, HTTP status code, latency.
* Correlation `traceId`, authenticated `user_id`, client IP address, and User-Agent.
* State transitions (e.g., `WORK_SESSION_STARTED`, `FINAL_CONTRIBUTION_CONFIRMED`).

### 7.2 What MUST NEVER Be Logged (Data Redaction):
* Raw passwords or password hashes.
* Plaintext JWT Bearer access/refresh tokens.
* Cloudflare R2 Secret Access Keys or Database passwords.
* Bank account numbers or private identity numbers in plain unredacted logs.
