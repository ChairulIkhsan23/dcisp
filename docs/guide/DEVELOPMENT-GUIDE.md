# DCISP v1.0 ENGINEERING DEVELOPMENT GUIDE
**Platform:** Dagang Creative Intern Solutions Program (DCISP)  
**Backend:** Golang 1.23+ (Gin Modular Monolith)  
**Frontend:** Next.js 15+ (App Router, TypeScript)  
**Source of Truth:** `docs/prd/PRD-DCISP-V1.md`, `docs/architecture/FINAL-TECH-STACK-SPEC-DCISP.md`, `docs/design-system/NEO-BRUTALIST-PIXEL-DESIGN-SYSTEM.md`

---

## 1. Backend Engineering Guidelines (Golang + Gin)

### 1.1 Project Structure

The Go backend follows an idiomatic **Modular Monolith** architecture:

```text
backend/
├── cmd/
│   └── api/
│       └── main.go                  # Application entrypoint, dependency wiring, graceful shutdown
├── internal/
│   ├── config/                      # Environment and policy loader
│   ├── database/                    # Database (pgxpool) & Redis connection management
│   ├── middleware/                  # Gin middlewares (Auth, Dynamic RBAC, Audit, Recovery)
│   ├── modules/                     # Domain modules (Self-contained business boundaries)
│   │   ├── identity/                # Auth, Users, Roles, Scopes
│   │   ├── people/                  # Interns, Alumni, Batches, Skills
│   │   ├── attendance/              # NFC Ingestion, Schedule, Sessions, Breaks, Overtime, Leave
│   │   ├── projects/                # Marketplace, Teams, Milestones, Tasks, Submissions, Evidence
│   │   ├── performance/             # XP Engine, Ranks, Evaluations, Top Performer, Badges
│   │   ├── finance/                 # Wallets, Bounty Split, Batch Fund, Taxes, Double-Entry Ledger
│   │   ├── documents/               # Storage Presigned Bridge, ID Cards, Certificates, Portfolios
│   │   └── system/                  # Unified Policy Engine, Settings, Audit Logs
│   └── shared/                      # Cross-module shared utilities
│       ├── response/                # Unified JSON response envelope
│       ├── utils/                   # Password (Argon2id), JWT Token helpers
│       └── eventbus/                # Internal asynchronous domain event emitter
├── pkg/
│   └── worker/                      # Asynq Redis background job handlers
└── tests/                           # Integration and end-to-end API test suites
```

### 1.2 Layer Responsibilities

```text
HTTP Request
     │
     ▼
[ Controller / Handler ]  ──► Validates payload schema, extracts JWT context, returns HTTP status
     │
     ▼
[ Service (Use Case) ]    ──► Executes business logic, enforces state machines, orchestrates transactions
     │
     ▼
[ Repository ]            ──► Pure SQL queries via pgxpool, mapping database rows to structs
     │
     ▼
[ Database / Storage ]
```

#### What Handlers MAY and MAY NOT Do:
* **MAY:** Parse request body/params, validate input syntax (binding/Zod equivalent), call Service methods, format response using `response.Success()` or `response.Error()`.
* **MAY NOT:** Write raw SQL queries, initiate database transactions, execute mathematical calculations for XP/Taxes, or bypass authorization guards.

#### What Services MAY and MAY NOT Do:
* **MAY:** Enforce business rules (e.g., verifying `actual_worked_hours <= approved_window`), manage database transaction boundaries (`pgx.Tx`), emit domain events, calculate formula outputs.
* **MAY NOT:** Access `*gin.Context` directly, read HTTP headers, or return raw HTTP status codes.

#### What Repositories MAY and MAY NOT Do:
* **MAY:** Execute parameterized SQL queries, scan rows into models, handle database connection pools.
* **MAY NOT:** Contain business decisions, validate user permissions, or call external third-party HTTP services.

### 1.3 Golang Coding Practices

#### Context Propagation
Always accept `ctx context.Context` as the first parameter in all repository and service methods. Propagate deadlines and cancellation signals:
```go
func (s *AttendanceService) ProcessTap(ctx context.Context, req TerminalTapRequest) (*TapResult, error) {
    // Pass ctx to repository and database queries
    user, err := s.repo.FindUserByCardUID(ctx, req.CardUID)
    if err != nil {
        return nil, fmt.Errorf("failed to find user by card UID: %w", err)
    }
    ...
}
```

#### Transaction Handling for Financial & Multi-Entity Operations
Wrap all multi-table mutations (such as Bounty Distribution or Batch Creation with Batch Fund) inside an explicit `pgx.Tx`:
```go
tx, err := r.db.Pool.Begin(ctx)
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}
defer tx.Rollback(ctx)

// Execute queries using tx...

if err := tx.Commit(ctx); err != nil {
    return fmt.Errorf("failed to commit transaction: %w", err)
}
```

#### Goroutines & Concurrency Safety
* Never spawn untracked goroutines inside HTTP handlers without passing a detached context or background timeout (`context.Background()`).
* Always recover panics inside background workers to prevent crashes.

---

## 2. Frontend Engineering Guidelines (Next.js 15+ App Router)

### 2.1 Application Directory Structure

```text
frontend/
├── app/
│   ├── (auth)/                      # Public authentication routes (Login)
│   │   └── login/page.tsx
│   ├── (dashboard)/                 # Authenticated application shell
│   │   ├── layout.tsx               # Global Sidebar, Header, Live Timer
│   │   ├── my-day/page.tsx          # Intern Cockpit (Player HUD - FR-052)
│   │   ├── team-today/page.tsx      # Supervisor Cockpit (Party Roster - FR-053)
│   │   ├── command-center/page.tsx  # Executive 4-Quadrant Map (FR-051)
│   │   ├── projects/page.tsx        # Bounty Board / Quest Marketplace
│   │   ├── finance/page.tsx         # Coin Pouch / Wallet Ledger
│   │   └── attendance/scanner/page.tsx # Terminal Scanner View (Gatekeeper)
│   └── layout.tsx                   # Root HTML, fonts, and theme providers
├── components/
│   ├── ui/                          # Reusable Neo-Brutalist & Shadcn primitives
│   │   ├── button.tsx               # NeoArcadeButton with hard drop shadow
│   │   ├── card.tsx                 # NeoPixelCard with 3px solid black border
│   │   └── progress.tsx             # PixelProgressBar with segmented fills
│   ├── cockpits/                    # Domain-specific interactive widgets
│   │   ├── precision-timer.tsx      # Monospaced active session clock
│   │   └── sprite-avatar.tsx        # Dynamic 8-bit character state animator
│   └── navigation/
│       └── dynamic-sidebar.tsx      # RBAC-filtered JRPG Inventory pause menu
├── hooks/
│   ├── use-work-session.ts          # Zustand session timer & break state machine
│   ├── use-session-integrity.ts     # Privacy-first Page Visibility & Idle Hook
│   └── use-live-events.ts           # Server-Sent Events (SSE) listener
└── lib/
    ├── api-client.ts                # Axios/Fetch wrapper with automatic token refresh
    ├── tokens.ts                    # Neo-Brutalist theme color tokens
    └── types/                       # Shared TypeScript interfaces
```

### 2.2 Server vs Client Component Boundaries

```text
┌────────────────────────────────────────────────────────┐
│ SERVER COMPONENT (Default in Next.js App Router)       │
│ • Initial data fetching directly from backend API      │
│ • Static layout structures, metadata, SEO tags         │
│ • Zero JavaScript bundle shipped to browser for layout │
└───────────────────────────┬────────────────────────────┘
                            │ Passes serializable props
                            ▼
┌────────────────────────────────────────────────────────┐
│ CLIENT COMPONENT ('use client')                        │
│ • Precision Monospace Timers (Session & Break counters)│
│ • State controls ([START WORK], [TAKE BREAK], [END])   │
│ • W3C Page Visibility & Idle detection listeners       │
│ • Drag-and-drop Kanban task boards                     │
│ • Modal dialogs, toast notifications, SSE listeners    │
└────────────────────────────────────────────────────────┘
```

#### Rule:
Keep Client Components as leaves at the bottom of the component tree. Wrap interactive widgets (`<PrecisionTimer />`, `<CombatControls />`) in `'use client'` while keeping the containing page a Server Component.

### 2.3 Data Fetching & Mutation Strategy (TanStack Query v5)
* **Queries:** Use TanStack Query (`useQuery`) with descriptive query keys (e.g., `['attendance', 'live']`, `['tasks', projectId]`).
* **Mutations:** Use `useMutation` with optimistic updates for immediate UI feedback on task dragging or timer toggling.
* **Cache Invalidation:** Invalidate relevant query keys on successful mutations (e.g., invalidating `['performance', 'stats']` after completing a task).

### 2.4 Forms & Validation Strategy
* **Library:** `react-hook-form` paired with `@hookform/resolvers/zod`.
* **Validation:** Define Zod schemas that mirror backend validation rules:
```typescript
export const workReportSchema = z.object({
  progress_percentage: z.number().min(0).max(100),
  what_i_did: z.string().min(30, "Laporan aktivitas minimal 30 karakter"),
  evidence_type: z.enum(["GIT_COMMIT", "SCREENSHOT", "URL", "DOCUMENT", "ATTACHMENT"]),
  evidence_url_or_key: z.string().min(1, "Bukti deliverable wajib dilampirkan"),
  problems: z.string().optional(),
  next_actions: z.string().optional(),
});
```

### 2.5 Authentication & Session Management
1. **Login Flow:** Submits email & password $\rightarrow$ Receives `access_token` and `refresh_token` $\rightarrow$ Stores access token in memory/secure cookie and user profile in Zustand.
2. **API Interceptor:** Automatically attaches `Authorization: Bearer <token>` to outgoing requests. If a `401 Unauthorized` response is received, the interceptor pauses requests, calls `/api/v1/auth/refresh`, updates the token, and replays the original request.
3. **Role & Scope Redirection:** If a user with role `SCANNER_OPERATOR` attempts to navigate to any URL other than `/attendance/scanner`, the router immediately redirects them to `/attendance/scanner` (BR-002).
