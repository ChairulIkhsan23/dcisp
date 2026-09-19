# DCISP v1.0 TESTING STRATEGY & TEST SPECIFICATION
**Platform:** Dagang Creative Intern Solutions Program (DCISP)  
**Backend:** Golang 1.23+ (`testing`, `testify`, `httptest`)  
**Frontend:** Next.js 15+ / React 19 (`Vitest`, `React Testing Library`, `Playwright`)  
**Source of Truth:** `docs/prd/PRD-DCISP-V1.md` (Section 18.3), `docs/architecture/FINAL-TECH-STACK-SPEC-DCISP.md`, Suite 9 BRDs (`docs/brd/`)

---

## 1. Testing Principles

1. **Test Business Rules First:** Prioritize high unit test coverage ($\ge 85\%$) on core business logic: formula calculations, state machines, financial balance constraints, and authorization scopes.
2. **Zero Regressions on Financials & Attendance:** Financial ledgers and attendance ingestion must have automated integration tests validating debit-credit equality and append-only trigger protection.
3. **Deterministic & Isolated Tests:** Tests must not depend on network connectivity to external third-party services. Use mocks/stubs for Cloudflare R2, Git APIs, and Payment Gateways.

---

## 2. Testing Levels & Tools

```text
┌─────────────────────────────────────────────────────────────┐
│ E2E UI Tests (Playwright)                                    │
│ • Critical Paths: Login, My Day Cockpit, Supervisor Approval │
├─────────────────────────────────────────────────────────────┤
│ API & Integration Tests (Go `httptest` + Vitest RTL)        │
│ • Endpoint routes, Middleware guards, DB Triggers           │
├─────────────────────────────────────────────────────────────┤
│ Unit Tests (Go `testing` + Vitest)                          │
│ • Formula calculations, State Machines, Policy Resolvers    │
└─────────────────────────────────────────────────────────────┘
```

| Testing Level | Scope | Framework / Tooling | Target Coverage |
|---|---|---|---|
| **Backend Unit Tests** | Formula XP, Taxes, Contribution Split, Password Hashing | Go native `testing` + `stretchr/testify` | $\ge 85\%$ |
| **Backend API Integration** | HTTP Handlers, Database Triggers, RBAC Middleware | `net/http/httptest` + PostgreSQL container | $\ge 80\%$ |
| **Frontend Unit & Component**| Precision Timers, Design Tokens, Zod Validation | `Vitest` + `@testing-library/react` | $\ge 75\%$ |
| **End-to-End (E2E)** | Full user workflows (Intern Check-In $\rightarrow$ Work Session) | `Playwright` | Core P0 Flows |

---

## 3. Critical Business Flows Test Catalog

The following 11 critical business flows derived from the PRD and BRDs **MUST** have automated test suites:

### Test Flow 1: Authentication & Argon2id Hashing
* **Test Case:** `TestAuth_Argon2id_HashingAndVerification`
* **Assertion:** Passwords hashed with Argon2id successfully verify against original text; invalid passwords return false.
* **Test Case:** `TestAuth_Login_SuccessAndIssueJWTPair`
* **Assertion:** Valid credentials return 15-minute access token and 7-day refresh token with correct claims.

### Test Flow 2: Dynamic RBAC & Scanner Operator Isolation (BR-001, BR-002)
* **Test Case:** `TestRBAC_ScannerOperator_StrictIsolation`
* **Assertion:** User with role `SCANNER_OPERATOR` receives `200 OK` on `/api/v1/attendance/scanner/view`, but receives `403 Forbidden` on `/api/v1/finance/ledger` or `/api/v1/projects`.
* **Test Case:** `TestRBAC_SuperAdmin_WildcardAccess`
* **Assertion:** Super Admin with permission `*.*` on `SYSTEM` accesses all protected resources.

### Test Flow 3: Work Schedule Punctuality & Grace Period Validation (BR-007, BRULE-WF-001)
* **Test Case:** `TestSchedule_CheckIn_OnTimeWithinGracePeriod`
* **Input:** Schedule 08:30, Grace Period 10 min, Check-in at 08:38.
* **Assertion:** Status is `ON_TIME` and `audio_event` is `check_in`.
* **Test Case:** `TestSchedule_CheckIn_LateAfterGracePeriod`
* **Input:** Check-in at 08:48.
* **Assertion:** Status is `LATE (18 min)` and `audio_event` is `late`.

### Test Flow 4: NFC Scan Debouncing Window (BR-008, BRULE-WF-013)
* **Test Case:** `TestAttendance_NFC_DebouncingWindow30s`
* **Execution:** First tap at $T=0\text{s}$ returns `200 OK`. Second tap at $T=15\text{s}$ returns `status: DEBOUNCED`, `audio_event: too_frequent`, and does not create a duplicate row in `attendance_event_logs`.

### Test Flow 5: Break State Machine & Unauthorized Break Penalty (BR-008, BR-009)
* **Test Case:** `TestBreak_Resume_LatePenalized`
* **Input:** Break window ends at 13:00. Intern resumes work at 13:18.
* **Assertion:** Status `WORKING` resumed, `is_violation_late = true`, and -2 XP penalty is debited to `xp_transactions`.

### Test Flow 6: Overtime Actual Worked Hours vs Approved Window (BR-013, BR-014)
* **Test Case:** `TestOvertime_Calculation_CappedToActualWorked`
* **Input:** Approved window: 2.5 hours (17:00–19:30). Actual worked duration: 1.75 hours (17:00–18:45).
* **Assertion:** Settled overtime duration is strictly 1.75 hours.

### Test Flow 7: Attendance Correction Append-Only Preservation (BR-025)
* **Test Case:** `TestCorrection_Approve_PreservesOriginalLog`
* **Execution:** Supervisor approves correction for a missing checkout.
* **Assertion:** New record `MANUAL_CORRECTION` created; original event log from that day remains unchanged in database.

### Test Flow 8: Three-Layer Contribution Sum Guard (BR-016)
* **Test Case:** `TestContribution_Finalize_SumMustEqual100`
* **Input:** Final contribution percentages: 40.00% + 35.00% + 25.00% = 100.00%.
* **Assertion:** Transaction succeeds and locks record. If sum is 95.00%, returns `422 Unprocessable Entity`.

### Test Flow 9: XP Scheme Partitions & Rank Promotion (BR-004, BR-017)
* **Test Case:** `TestXP_AlumniTrack_IsolatedFromActiveInterns`
* **Assertion:** Alumni task completion credits `ALUMNI_CONTRIBUTION` and does not increment `internship_xp`.
* **Test Case:** `TestRank_Promotion_ThresholdReached`
* **Input:** Intern accumulates 500 XP (threshold for Rank C).
* **Assertion:** Rank promoted from `Novice` to `Apprentice`, promotion audit event emitted.

### Test Flow 10: Double-Entry Financial Ledger Equality (BR-023)
* **Test Case:** `TestLedger_DoubleEntry_ZeroDifference`
* **Execution:** Settle gross bounty of Rp 1,000,000.
* **Assertion:** Total debits in transaction equal total credits ($\sum \text{Debit} - \sum \text{Credit} == 0.00$).

### Test Flow 11: PostgreSQL Trigger Immutability Protection (BR-023, BR-025, FR-047)
* **Test Case:** `TestDatabase_Triggers_BlockUpdateAndDelete`
* **Execution:** Execute `UPDATE` and `DELETE` on `attendance_event_logs` and `financial_ledgers`.
* **Assertion:** Database driver throws PL/pgSQL trigger exception `DCISP Security & Integrity Violation`.

---

## 4. Test Naming & Structure Standards

Use the standard BDD-inspired naming format:
```text
Test<Component>_<Scenario>_<ExpectedBehavior>
```

#### Examples:
* `TestAuth_InvalidPassword_Returns401`
* `TestAttendance_OfflineSync_InsertsIdempotentRecords`
* `TestLedger_UnbalancedTransaction_RollsBack`

---

## 5. Failure & Edge Case Testing Matrix

| Failure Mode | Test Strategy | Expected System Response |
|---|---|---|
| **Database Disconnection** | Simulate down PostgreSQL connection pool | `/health/readiness` returns `503 Service Unavailable` |
| **Duplicate Scan Ingestion** | Send identical payload with same `idempotency_key` | Returns `200 OK` with `ON CONFLICT DO NOTHING` |
| **Overdraft Wallet Cashout** | Request payout exceeding `current_balance` | Rejected with `400 Bad Request: Insufficient Balance` |
| **Expired QR Code Scan** | Present rotating QR token with age $> 30\text{ seconds}$ | Rejected with `401 Unauthorized: Token Expired` |
| **Unregistered NFC Card** | Tap UID not associated with any active user | Returns `404 Not Found` with `audio_event: "card_unregistered"` |
