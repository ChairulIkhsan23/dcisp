# AI CODING AGENT OPERATIONAL INSTRUCTIONS
**Platform:** Dagang Creative Intern Solutions Program (DCISP)  
**Primary Tech Stack:** Golang 1.23+ (Gin), Next.js 15+ (App Router, TypeScript), PostgreSQL 16, Redis 7, Cloudflare R2  
**Source of Truth:** `docs/prd/PRD-DCISP-V1.md`, `docs/architecture/FINAL-TECH-STACK-SPEC-DCISP.md`, Suite 9 BRDs (`docs/brd/`), `ROADMAP.md`  
**Execution Blueprint:** `ROADMAP.md` (135 tasks, 17 phases, critical path, acceptance criteria)

---

## 1. Mandatory Workflow Before Writing Any Code

Before implementing any feature, bugfix, or refactoring, you **MUST**:
1. **Consult the Roadmap:** Check `ROADMAP.md` for the current Task ID (e.g., `T-001`), its dependencies, prerequisites, and acceptance criteria. Never execute tasks out of dependency order.
2. **Read Requirements:** Read the relevant section of `docs/prd/PRD-DCISP-V1.md` and the corresponding BRD in `docs/brd/`.
3. **Read Technical Contracts:** Review `docs/guide/API.md` for endpoint contracts, `docs/guide/DATABASE.md` for schema constraints, and `docs/guide/CODING-STANDARDS.md` for conventions.
4. **Inspect Existing Implementation:** Read existing files in `backend/` or `frontend/` to understand conventions, imports, and dependencies.
5. **Identify Affected Modules & Scope:** Determine if the task touches Identity, Workforce, Projects, Performance, Finance, Documents, System, or Intelligence.
6. **Check Authorization & Privacy Impact:** Ensure changes adhere to Dynamic RBAC (*BR-001*), Scanner Isolation (*BR-002*), and Privacy-First tracking (*BR-010*).

Workflow sequence:
> **READ → UNDERSTAND → ANALYZE → VERIFY EXISTING CODE → PLAN → IMPLEMENT → TEST → VERIFY AGAIN**

---

## 2. Strict Rules During Implementation

When writing or editing code, you **MUST**:
* **Follow Established Tech Stack:** Use Golang Gin for backend, Next.js 15 (TypeScript) for frontend, PostgreSQL for storage, Redis for debouncing/caching, Asynq for background workers.
* **Follow Coding Standards:** Adhere strictly to `docs/guide/CODING-STANDARDS.md` (idiomatic Go, strict TypeScript, no `any`, Neo-Brutalist NP-ADS design tokens).
* **Code Commenting Standard:** Setiap function, method, handler, service, repository, middleware, hook, dan utility wajib memiliki tepat satu kalimat komentar dalam bahasa Indonesia yang menjelaskan tujuan/tanggung jawabnya (`// + Kata kerja + objek/tujuan.`).
* **API Response & Error Language:** Seluruh pesan respon API (`message`), error validation, dan sentinel errors wajib menggunakan bahasa Indonesia baku.
* **Adhere to Database & API Contracts:** Do not add undocumented columns or change response JSON envelopes without specification alignment.
* **Never Invent Business Rules:** If a requirement or formula is not defined in the PRD, check open questions in `ROADMAP.md` or mark as `TBD`—never invent arbitrary calculations.
* **Zero Invasive Surveillance Code:** Do NOT write code that captures screenshots, records keystrokes, scans desktop processes, or reads private chats (*BR-010*).
* **Zero Hardcoding on Roles & Policies:** Never check roles with static string equality in application code; always evaluate permissions via dynamic RBAC guards (*BR-001*).
* **Maintain Immutability:** Never write `UPDATE` or `DELETE` queries for `attendance_event_logs`, `financial_ledgers`, `audit_logs`, or `xp_transactions` (*BR-023, BR-025, FR-047*).
* **Do Not Add Unnecessary Dependencies:** Rely on the Go standard library, Gin, pgx, redis-go, and Next.js built-in utilities before adding new third-party packages.
* **Update Roadmap Status:** After successfully verifying a task, update its status in `ROADMAP.md` (e.g., from `TODO` to `DONE`).

---

## 3. Protocol for Architectural or Contract Changes

If a requested task requires modifying:
* System architecture or layer boundaries
* Database schema, constraints, or migrations
* Existing API endpoint contracts or JSON envelopes
* Authentication / Authorization mechanisms
* External third-party integrations (Cloudflare R2, ESP32 IoT gateway)

You must **STOP, EXPLAIN THE REQUIRED CHANGE TO THE USER, AND WAIT FOR EXPLICIT APPROVAL** before making changes. Never perform silent architectural shifts.

---

## 4. Verification Checklist Before Finishing a Task

Before marking any task as complete, you **MUST**:
- [ ] 1. **Format Code:** Run `go fmt ./...` (Go) and format TypeScript files.
- [ ] 2. **Check Types & Compile:** Run `go build ./...` (backend) and `tsc --noEmit` (frontend) to ensure zero compilation or type errors.
- [ ] 3. **Run Automated Tests:** Run `go test ./... -v` (backend) and test suites for any modified components.
- [ ] 4. **Check Regressions:** Verify that existing tests continue to pass.
- [ ] 5. **Review Changed Files:** Inspect `git status` and diffs to ensure no unintended files, secrets, or temporary files are left behind.

---

## 5. Standard Final Task Report Format

Every completed task must be summarized using this exact Markdown structure:

```markdown
## Summary
[Brief 1-2 sentence description of what was completed]

## Changed Files
- `path/to/modified_or_created_file.ext`: [Description of changes]

## Implementation Details
- [Key business logic or architectural element implemented]

## Verification & Tests
- `go test ./tests/... -v`: [Test outcome, e.g., PASS (0.78s)]
- [Specific test scenarios verified]

## Unresolved Issues / Follow-up
- [Any TBD items or next logical phase tasks]
```
