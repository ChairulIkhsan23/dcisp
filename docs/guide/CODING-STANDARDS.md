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

## 5. Code Commenting Standards (Wajib Satu Kalimat Bahasa Indonesia)

Setiap function, method, handler, service, repository method, custom hook, middleware, dan utility yang memiliki logic atau responsibility **wajib memiliki tepat satu kalimat komentar dalam bahasa Indonesia** tepat sebelum deklarasinya.

### 5.1 Format Baku
```text
// + Kata kerja + objek/tujuan.
```

### 5.2 Contoh Penerapan
* **Handler:**
  ```go
  // Menangani permintaan HTTP untuk autentikasi login pengguna dan menerbitkan token akses JWT.
  func (ctrl *Controller) Login(c *gin.Context) { ... }
  ```
* **Service:**
  ```go
  // Memverifikasi kredensial email dan password serta menghasilkan pasangan token akses dan refresh.
  func (s *Service) Login(ctx context.Context, email, password string) (*utils.TokenPair, *UserProfileResponse, error) { ... }
  ```
* **Repository:**
  ```go
  // Mengambil data pengguna aktif dari database berdasarkan alamat email.
  func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) { ... }
  ```
* **Middleware:**
  ```go
  // Memvalidasi token JWT Bearer pada header request dan menyematkan konteks identitas pengguna ke dalam request context.
  func AuthJWT(cfg *config.Config) gin.HandlerFunc { ... }
  ```
* **TypeScript Hook / Action:**
  ```typescript
  // Mengelola timer aktif sesi kerja harian dan transisi status istirahat di sisi klien.
  export function useWorkSession() { ... }
  ```

### 5.3 Checklist Komentar
- [x] Tepat satu kalimat dan diakhiri tanda titik.
- [x] Menggunakan bahasa Indonesia profesional dan alami.
- [x] Menjawab *"Function ini digunakan untuk melakukan apa?"*.
- [x] Tidak menjelaskan detail implementasi internal baris per baris.

