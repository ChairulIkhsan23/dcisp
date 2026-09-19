# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: Identity, Authentication & Dynamic RBAC Governance
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**Identity, Authentication & Dynamic RBAC Governance (Tata Kelola Identitas Pengguna, Otentikasi Keamanan, dan Otorisasi Multi-Peran Dinamis 5-Lapis)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain ini merupakan fondasi keamanan, kedaulatan akses data, dan kepatuhan tata kelola di seluruh ekosistem DCISP:
- Mengatur identitas tunggal pengguna yang dapat mengemban 10 peran sistem (*Super Admin, Admin, HR Admin, Project Manager, Supervisor, Reviewer, Finance, Scanner Operator, Intern, Alumni*).
- Menegakkan arsitektur otorisasi berjenjang 5 lapis: $\text{Role} \rightarrow \text{Permission} \rightarrow \text{Scope} \rightarrow \text{Resource} \rightarrow \text{Action}$.
- Menerapkan aturan bisnis krusial **BR-001** (*Dynamic RBAC without Hardcoding*) di mana relasi hak akses dikonfigurasi runtime melalui database/kebijakan tanpa modifikasi kode sumber.
- Menegakkan isolasi mutlak peran *Scanner Operator (Gatekeeper)* pada gerbang presensi sesuai **BR-002** (`scope = scanner_only`).

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Percobaan login pengguna, panggilan API yang memerlukan otorisasi, penerbitan kredensial baru, atau perubahan konfigurasi hak akses oleh The Creator (Super Admin).
* **Input Bisnis:** Kredensial akun (email & password), payload request API dengan token otorisasi, penugasan peran (*role mapping*), dan matriks izin (*permission-scope assignment*).
* **Proses Utama Bisnis:**
  1. Verifikasi kredensial pengguna dan penerbitan token sesi akses singkat (15 menit) dengan rotasi refresh token.
  2. Evaluasi otorisasi runtime pada gateway terhadap aksi, sumber daya, dan batasan scope konteks pengguna.
  3. Isolasi ketat akun terminal pemindai gerbang (*Scanner Operator*).
  4. Pencatatan log audit keamanan atas setiap kegagalan login dan percobaan eskalasi hak akses ilegal.
* **Keputusan Bisnis yang Dibuat:**
  - Pengesahan validitas otentikasi identitas pengguna.
  - Keputusan otorisasi: Akses Diizinkan (*Allow*) vs Akses Ditolak (*Deny / HTTP 403 Forbidden*).
  - Penetapan penugasan peran (*Role Assignment*) dan cakupan wewenang (*Scope Context*).
* **Output Bisnis:** Sesi kerja aman terverifikasi, konteks identitas pada setiap transaksi bisnis, dan log audit keamanan identitas.
* **Kapan Selesai:** Sesi pengguna berakhir (*logout* / kedaluwarsa) atau hak akses selesai dievaluasi pada transaksi.
* **Domain Konsumen Output:** Seluruh domain bisnis lain (Workforce, Projects, Performance, Finance, Documents, Intelligence, System).

### 4. Hubungan dengan Domain Lain (Related Domains)
* **Seluruh Domain Bisnis (Downstream Consumer):** Seluruh modul bisnis bergantung pada domain ini untuk memastikan aktor yang mengeksekusi aksi memiliki wewenang sah.

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — Identity & Dynamic RBAC Governance |
| **Business Domain** | Identity & Access Management (Domain 1 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Sign-Off |
| **Business Owner** | Irfan (Super Administrator / Head of Technology, PT. ADT) |
| **Prepared By** | Lead Requirements Engineer & Security Analyst |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 3, 8.2, 10.3, 11.1, 12.2) |

---

## 2. Executive Summary

Dokumen Kebutuhan Bisnis (BRD) ini merinci persyaratan tata kelola identitas, keamanan otentikasi, dan otorisasi dinamis berjenjang untuk platform DCISP. Tujuannya adalah melindungi data rahasia perusahaan, mencegah eskalasi hak akses ilegal, dan menerapkan pemisahan tugas (*Separation of Concerns*) berbasis 10 peran dan 8 batasan scope data tanpa hardcoding logika izin di kode sumber (*BR-001, BR-002*).

---

## 3. Business Context

Sebelum platform DCISP dibangun:
1. Pengecekan izin seringkali di-hardcode seperti `if (role == 'admin')`, menyulitkan penyesuaian wewenang saat struktur organisasi berkembang.
2. Petugas pemindai presensi di lobi seringkali memiliki akun dengan hak akses terlalu luas, membuka risiko kebocoran data nilai evaluasi dan nominal kompensasi finansial magang.

---

## 4. Business Problem

1. **Kerapuhan Logika Otorisasi Statis:** Ketergantungan pada pengecekan nama role statis dalam kode program mengharuskan deployment ulang kode setiap kali terjadi perubahan izin.
2. **Risiko Kebocoran Data pada Terminal Fisik:** Operator pemindai di gerbang lobi tidak terisolasi, berpotensi melihat data pribadi dan keuangan perusahaan.

---

## 5. Business Objectives

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-IDN-01** | Kerapuhan otorisasi statis | 100% hak akses peran dikonfigurasi dinamis di database/policy tanpa deploy ulang | Fleksibilitas tata kelola organisasi & kecepatan adaptasi | Perubahan wewenang berlaku instan pada request API berikutnya |
| **OBJ-IDN-02** | Risiko kebocoran data terminal gerbang | Isolasi mutlak peran Scanner Operator hanya pada fungsi pemindaian | Perlindungan kerahasiaan data internal perusahaan | 100% percobaan akses modul non-scanner oleh operator ditolak (403) |
| **OBJ-IDN-03** | Pembobolan sesi kredensial | Otentikasi stateless berumur pendek (15m) dengan rotasi refresh token di DB | Keamanan data tingkat tinggi & kepatuhan standar industri | 0 insiden token hijacking tidak terdeteksi |

---

## 6. Business Scope

### 6.1 In Scope
1. **Dynamic RBAC & Permission Engine (FR-001, BR-001):** Hierarki 5 lapis (`Role -> Permission -> Scope -> Resource -> Action`).
2. **Manajemen 10 Peran Standar:** Super Admin, Admin, HR Admin, Project Manager, Supervisor, Reviewer, Finance, Scanner Operator, Intern, Alumni.
3. **Isolasi 8 Cakupan Konteks (Scopes):** `system`, `workforce_and_people`, `assigned_team`, `assigned_projects`, `financial_data`, `scanner_only`, `own_data`, `public_data`.
4. **Isolasi Operator Scanner (BR-002):** Pembatasan peran pemindai presensi strictly pada `scope = scanner_only`.
5. **Keamanan Sesi & Otentikasi Ganda:** Token JWT + Argon2id hash password untuk web pengguna; API Key SHA-256 bertanda tangan untuk Device Gateway.

### 6.2 Out of Scope
1. **Single Sign-On (SSO) Eksternal Publik Bebas:** Platform menggunakan sistem identitas terkurasi internal PT. ADT.

---

## 7. Stakeholders & Business Actors

| Stakeholder / Actor | Tanggung Jawab Bisnis | Kepentingan Bisnis | Otoritas Keputusan |
|---|---|---|---|
| **The Creator (Super Admin)** | Mengonfigurasi master roles, permissions, scopes, dan rotasi kunci keamanan | Integritas dan kedaulatan keamanan sistem | Pengesahan konfigurasi master RBAC |
| **Realm Warden (Admin)** | Mengelola penugasan peran ke pengguna dan status akun | Kelancaran operasional akun pengguna | Penugasan role operasional |
| **Gatekeeper (Scanner Operator)**| Memindai kartu fisik di gerbang masuk kantor | Kecepatan pemrosesan presensi lobi | Tidak memiliki otoritas manajerial data |
| **Seluruh Pengguna Ranah** | Menjaga kerahasiaan kredensial login akun pribadi | Keamanan akun dan privasi data pribadi | Mengelola kata sandi pribadi |

---

## 8. Target Business Process & Scenarios

### Scenario: Evaluasi Otorisasi Dinamis pada Panggilan Transaksi
* **Trigger:** Supervisor mencoba menyetujui permohonan koreksi absensi intern.
* **Precondition:** Supervisor terotentikasi dengan token aktif.
* **Main Flow:**
  1. Request masuk ke API Gateway dengan Bearer Token.
  2. Middleware otorisasi mengekstrak data identitas dan scope supervisor.
  3. Dynamic Permission Engine memeriksa tabel permission di database/cache: apakah role `SUPERVISOR` memiliki izin `attendance.correction.approve` pada scope `assigned_team`?
  4. Context Guard memvalidasi bahwa intern yang dikoreksi benar-benar merupakan anggota tim bimbingan supervisor tersebut.
  5. Akses disahkan (`ALLOW`), transaksi dieksekusi, dan audit log mencatat aksi.

---

## 9. Business Requirements & Rules

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis | Rasional Bisnis | Prioritas | Sumber PRD |
|---|---|---|---|---|
| **BRQ-IDN-001** | Bisnis mewajibkan roles dan permissions dapat dikonfigurasi di database/policy tanpa hardcoding kode | Menghilangkan kerapuhan kode dan memfasilitasi perubahan wewenang runtime | P0 | FR-001, BR-001 |
| **BRQ-IDN-002** | Bisnis mewajibkan peran Scanner Operator diisolasi mutlak hanya pada fungsi pemindai (`scope = scanner_only`) | Mencegah kebocoran data rahasia peserta dan keuangan pada terminal lobi | P0 | FR-001, BR-002 |
| **BRQ-IDN-003** | Bisnis mewajibkan penggunaan hashing kata sandi modern (Argon2id) dan token akses berumur singkat (15 menit) | Menjaga keamanan akun dari ancaman pembobolan kredensial | P0 | Section 11.1, 12.2 |

---

## 10. Traceability Matrix & Acceptance Criteria

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Acceptance Criteria |
|---|---|---|---|
| FR-001 (Dynamic RBAC) | BRQ-IDN-001 | BRULE-IDN-001 (Zero Hardcoding) | AC-IDN-001: Perubahan permission di runtime langsung berlaku tanpa restart server |
| FR-001, BR-002 (Scanner Isolation)| BRQ-IDN-002 | BRULE-IDN-002 (Scanner Boundary) | AC-IDN-002: Operator pemindai ditolak (403) saat mengakses endpoint proyek/finansial |

---

## 11. BRD Completion Checklist

- [x] Seluruh kebutuhan domain Identity, Authentication, dan Dynamic RBAC terdefinisi lengkap.
- [x] Aturan Dynamic RBAC (BR-001) dan Isolasi Scanner (BR-002) terpetakan eksplisit.

---
*Dokumen ini disimpan permanen di `brd/06-BRD-IDENTITY-RBAC-GOVERNANCE.md`.*
