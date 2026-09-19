# DCISP v1.0 CODING STANDARDS & STYLE GUIDE
**Platform:** Dagang Creative Intern Solutions Program (DCISP)  
**Languages:** Golang 1.23+ & TypeScript 5.x  
**Frameworks:** Gin (Backend) & Next.js 15+ / React 19 (Frontend)  
**Source of Truth:** `docs/prd/PRD-DCISP-V1.md`, `docs/architecture/FINAL-TECH-STACK-SPEC-DCISP.md`, `docs/design-system/NEO-BRUTALIST-PIXEL-DESIGN-SYSTEM.md`

---

## 1. General Principles

1. **Simplicity Over Cleverness (KISS):** Write boring, readable, maintainable code. The best code is simple and explicit.
2. **You Aren't Gonna Need It (YAGNI):** Do not create abstractions, factories, or generic interfaces for single implementations. Introduce interfaces only when multiple concrete implementations or mock testing require them.
3. **Strict Separation of Concerns:** Handlers handle HTTP; Services handle business rules; Repositories handle database I/O.
4. **Zero Assumptions on Financials:** Always verify double-entry balancing and never use floating-point math for money.

---

## 2. Golang Coding Standards

### 2.1 Package & Identifier Naming
* **Package Names:** Short, lowercase, single-word names without underscores (e.g., `package identity`, `package attendance`, `package database`). Avoid generic names like `common` or `helpers`.
* **Exported Identifiers:** PascalCase (e.g., `ProcessTap`, `NewPostgresDB`, `TokenPair`).
* **Unexported Identifiers:** camelCase (e.g., `isDebounced`, `userRole`, `jwtSecret`).
* **Acronyms:** Uniform casing for acronyms (e.g., `NFCReader`, `UserID`, `JSONBody`, `HTTPClient`, not `NfcReader` or `UserId`).

### 2.2 Error Handling & Wrapping
* Never ignore errors (`_ = doSomething()` is prohibited for fallible operations).
* Always wrap errors with context using `fmt.Errorf("...: %w", err)`:
  ```go
  // GOOD
  user, err := r.repo.FindByID(ctx, id)
  if err != nil {
      return nil, fmt.Errorf("failed to retrieve user profile: %w", err)
  }

  // BAD
  if err != nil {
      return nil, err // Loses contextual information
  }
  ```
* Define sentinel errors for predictable domain outcomes (e.g., `ErrUserNotFound`, `ErrCardUnregistered`, `ErrQuotaExceeded`).

### 2.3 Structs & Interface Design
* Define interfaces at the consumer level, not the producer level.
* Keep interfaces small (1–4 methods).
* Struct tags must be explicit for JSON and Database mapping:
  ```go
  type InternProfile struct {
      ID            uuid.UUID `json:"id" db:"id"`
      BatchCode     string    `json:"batch_code" db:"batch_code"`
      InternshipXP  int       `json:"internship_xp" db:"internship_xp"`
  }
  ```

### 2.4 Logging & Comments
* Log structured JSON with correlation `traceId`.
* Add code comments sparingly: explain *why* non-obvious business logic exists (e.g., *“// BR-009: Unauthorized break incurs -2 XP penalty if resumed after 13:00”*), never narrate what simple code does.

---

## 3. TypeScript & React / Next.js Standards

### 3.1 Strict Typing
* `strict: true` is enforced in `tsconfig.json`.
* **No `any` Policy:** Using `any` is strictly prohibited. Use `unknown` with type guards or explicit generics if dynamic types are unavoidable.
* **Interfaces vs Types:** Use `interface` for object structures that can be extended; use `type` for unions, primitives, and mapped types:
  ```typescript
  // Type union for standard audio events
  export type AudioEventKey = 
    | "card_read" | "processing" | "device_ready" | "check_in" 
    | "work_start" | "break_start" | "break_end" | "check_out" 
    | "card_unregistered" | "card_invalid" | "save_failed" 
    | "offline" | "offline_success" | "too_frequent" | "late";

  export interface UserProfile {
    id: string;
    email: string;
    fullName: string;
    role: string;
    scopes: string[];
    permissions: string[];
  }
  ```

### 3.2 React Components & Hooks
* **Naming:** PascalCase for React component files and functions (e.g., `PrecisionTimer.tsx`, `NeoArcadeButton.tsx`).
* **Hook Naming:** camelCase prefixed with `use` (e.g., `useWorkSession.ts`, `useSessionIntegrity.ts`).
* **Component Size:** Components exceeding 150 lines should be evaluated for decomposition into smaller focused sub-components.
* **Prop Typing:** Explicitly type component props using dedicated interfaces:
  ```typescript
  interface NeoArcadeButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: "primary" | "combat" | "mana" | "neutral";
    children: React.ReactNode;
  }
  ```

### 3.3 Design System Tokens Enforcement (NP-ADS)
* Always utilize predefined Neo-Brutalist utility classes and color tokens:
  - Gold/Yellow: `#F1C812` (Arcade Gold, XP Bar)
  - Blue: `#2D5AB8` (Mana Blue, Quest Headers)
  - Red: `#E13447` (Combat Red, Damage/Penalty)
  - Ink Black: `#000000` (Borders & Hard Drop Shadows)
* Never apply soft blur shadows (`shadow-md`, `shadow-lg` from standard Tailwind); always use hard offset shadows (`shadow-neo-sm`, `shadow-neo-md`, `shadow-neo-lg`).

---

## 4. Comprehensive Naming Conventions

| Item Category | Convention | Pattern Example | Target Domain |
|---|---|---|---|
| **Go Package** | lowercase, single word | `package attendance`, `package finance` | Backend |
| **Go Struct Model** | PascalCase | `type AttendanceEventLog struct` | Backend |
| **Go Interface** | PascalCase | `type AttendanceService interface` | Backend |
| **Go Function / Method** | PascalCase (exported), camelCase (internal) | `ProcessTap()`, `validateSchedule()` | Backend |
| **Go Variable** | camelCase | `userRole`, `bountyAmount` | Backend |
| **React Component** | PascalCase | `PlayerHudCockpit.tsx`, `NeoPixelCard.tsx` | Frontend |
| **Custom Hook** | camelCase (`use*`) | `useWorkSession.ts`, `useLiveEvents.ts` | Frontend |
| **TypeScript Type/Interface** | PascalCase | `type TapResult = { ... }`, `interface UserProfile` | Frontend |
| **API Endpoint URL** | lowercase `kebab-case` with `/api/v1/` prefix | `/api/v1/attendance/terminal-tap` | API |
| **Database Table** | lowercase `snake_case` plural | `attendance_event_logs`, `financial_ledgers` | PostgreSQL |
| **Database Column** | lowercase `snake_case` singular | `planned_contribution_pct`, `device_id` | PostgreSQL |
| **Database Trigger** | `trg_<action>_<table_name>` | `trg_immutable_financial_ledger` | PostgreSQL |
| **CSS Class / Token** | `kebab-case` | `shadow-neo-md`, `border-neo-thick` | Styling |
