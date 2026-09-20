# DCISP v1.0 GIT BRANCHING & RELEASE STRATEGY
**Platform:** Dagang Creative Intern Solutions Program (DCISP)  
**Model:** Trunk-Based / GitHub Flow Hybrid  
**Source of Truth:** `docs/guide/DEVELOPMENT-GUIDE.md`, `docs/guide/CODING-STANDARDS.md`

---

## 1. Model Percabangan Utama

Proyek DCISP mengadopsi model **Trunk-Based / GitHub Flow Hybrid** dengan dua cabang utama (*Long-Lived Branches*) dan cabang fitur berumur pendek (*Short-Lived Feature Branches*):

```text
       HOTFIX ──────────────────────────────┐
         ▲                                  ▼
         │  (Emergency Fix)                 │
 ┌───────┴──────────────────────────────────┴─────────────────────────┐
 │  [ main ] (Production-Ready / Tagged vX.Y.Z)                        │
 └───────▲──────────────────────────────────▲─────────────────────────┘
         │ (Merge via Release PR)           │
 ┌───────┴──────────────────────────────────┴─────────────────────────┐
 │  [ staging ] (Integration / QA Testing / Preview Deployment)       │
 └───────▲──────────────────────────────────▲─────────────────────────┘
         │ (PR + CI Check)                  │ (PR + CI Check)
 ┌───────┴──────────────────┐      ┌────────┴──────────────────┐
 │ feat/backend/attendance  │      │ feat/frontend/my-day-hud  │
 └──────────────────────────┘      └───────────────────────────┘
```

---

## 2. Struktur Cabang Utama (Long-Lived Branches)

| Nama Branch | Target Lingkungan | Kebijakan Proteksi | Deskripsi & Syarat Penggabungan |
|---|---|---|---|
| **`main`** | **Production** | Protected (Strict PR + Approval + CI Pass) | Berisi kode stabil produksi. Setiap merge wajib disertai tag versi Semantic Versioning (`v1.0.0`, `v1.0.1`). |
| **`staging`** | **Staging / QA** | Protected (PR + Automated CI Pass) | Integrasi seluruh modul sebelum rilis ke production. Digunakan untuk uji integrasi hardware IoT, verifikasi akuntansi, dan review QA. |

---

## 3. Konvensi Penamaan Branch Kerja (Short-Lived Branches)

Branch kerja dibuat dari `staging` dan memiliki masa hidup pendek (1–3 hari).

### Format Penamaan:
```text
<type>/<scope>/<short-description>
```

### Kategori `<type>`:
* **`feat`**: Fitur atau kapabilitas sistem baru.
* **`fix`**: Perbaikan bug pada branch `staging`.
* **`hotfix`**: Perbaikan darurat langsung dari branch `main`.
* **`refactor`**: Restrukturisasi kode tanpa mengubah behavior sistem.
* **`docs`**: Pembaruan dokumentasi (PRD, BRD, guides).
* **`chore`**: Pembaruan dependensi, Docker config, CI/CD script.
* **`test`**: Penambahan automated test suites atau benchmark.

### Kategori `<scope>`:
* `backend` / `api`
* `frontend` / `ui`
* `iot` / `firmware`
* `db` / `migrations`
* `docs`

### Contoh Penamaan Branch:
* `feat/backend/attendance-nfc-tap`
* `feat/frontend/player-hud-cockpit`
* `feat/iot/esp32-audio-cache`
* `feat/db/double-entry-triggers`
* `fix/backend/jwt-refresh-rotation`
* `docs/brd/workforce-update`
* `hotfix/backend/auth-header-panic`

---

## 4. Standar Pesan Commit (Conventional Commits)

Format commit wajib mengikuti konvensi **Conventional Commits**:
```text
<type>(<scope>): <deskripsi singkat dalam bahasa Indonesia / Inggris>
```

### Contoh Commit:
* `feat(attendance): implement terminal-tap endpoint with 30s debounce`
* `feat(frontend): add neo-brutalist precision timer component`
* `feat(iot): add littlefs persistent audio cache and sha256 check`
* `fix(ledger): correct debit-credit balance constraint validation`
* `docs(brd): finalize 02-BRD-WORKFORCE-ATTENDANCE specification`
* `chore(docker): update redis port mapping to 6381`

---

## 5. Alur Kerja Pull Request (PR) & Quality Gates

Setiap penggabungan branch kerja ke `staging` atau `main` wajib melalui **Pull Request**:

```text
[ Developer Branch ] ──► [ Push to Origin ] ──► [ Open PR to Staging ]
                                                        │
                                                        ▼
                                         [ AUTOMATED CI CHECKS ]
                                         1. Backend: go fmt, go test ./...
                                         2. Frontend: npm run type-check, npm run build
                                         3. Code Commenting Standard Check
                                                        │
                                                        ▼ (CI Passed + 1 Approval)
                                         [ SQUASH & MERGE TO STAGING ]
```

### Syarat Wajib Lolos PR (Definition of Done):
1. **Automated CI Tests Pass:**
   * Backend: `go test ./tests/... -v` lolos 100%.
   * Frontend: `npm run type-check` dan `npm run build` bebas error.
2. **Code Commenting Standard:** Setiap function/method baru wajib memiliki 1 kalimat komentar bahasa Indonesia (`// + Kata kerja + objek/tujuan.`).
3. **Immutability Protection:** Tidak ada query `UPDATE`/`DELETE` pada tabel immutable (`attendance_event_logs`, `financial_ledgers`, dll).
4. **Merge Strategy:** Gunakan **Squash and Merge** agar riwayat commit pada branch utama tetap bersih dan linear.

---

## 6. Strategi Rilis & Tagging Versi (SemVer)

Penggabungan dari `staging` ke `main` menandakan rilis resmi:
* **Format:** `v<Major>.<Minor>.<Patch>` (contoh: `v1.0.0`, `v1.0.1`).
* **Tahapan Rilis Proyek:**
  * `v1.0.0-mvp` $\rightarrow$ Core Foundation & Presence IoT (Fase 1–3).
  * `v1.0.0-rc1` $\rightarrow$ Projects, Gamification & Finance (Fase 4–5).
  * `v1.0.0` $\rightarrow$ Production Go-Live (Fase 6–8).

---

## 7. Alur Penanganan Bug Darurat (Hotfix Workflow)

Jika terjadi kendala kritis di production (`main`):
1. Buat branch langsung dari `main`: `hotfix/<scope>/<issue-desc>`.
2. Perbaiki bug dan uji secara lokal/CI.
3. Buka PR ke `main` $\rightarrow$ Merge dan terbitkan tag patch baru (misal: `v1.0.1`).
4. **Backport:** Lakukan *cherry-pick* atau merge kembali hotfix ke branch `staging` agar perubahan tidak hilang di pengembangan selanjutnya.
