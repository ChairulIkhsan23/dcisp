# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: System Integrity, Unified Policy & Audit Governance
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**System Integrity, Unified Policy & Audit Governance (Tata Kelola Mesin Kebijakan Runtime Terpadu, Pusat Notifikasi Real-Time, dan Jejak Audit Keamanan Append-Only)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain ini merupakan pusat pengendali hukum operasional (*The Creator's Codex*) dan mekanisme kepatuhan sistemik di seluruh platform DCISP:
- Mengendalikan seluruh parameter aturan bisnis variabel (jam jadwal, grace period, matriks sanksi penalti XP, formula deduksi pajak, ambang batas rank) melalui satu mesin kebijakan terpusat (*Unified Policy Engine - FR-045*) tanpa modifikasi kode sumber program (*BR-001, BR-012, BR-020*).
- Menjamin akuntabilitas dan auditabilitas kepatuhan 100% dengan merekam setiap mutasi data sensitif, tindakan administratif, dan peristiwa otorisasi ke dalam buku sejarah audit yang tidak dapat dimodifikasi atau dihapus (*Append-Only Audit Log - FR-047, BR-025*).
- Mengorkestrasi distribusi pengingat dan peringatan mendesak kepada seluruh pemangku kepentingan secara tepat waktu (*Event-Driven Notification Center - FR-046*).

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Terjadinya mutasi data penting/sensitif, perubahan parameter kebijakan oleh Super Admin, peristiwa sistem yang memicu notifikasi, atau audit investigasi keamanan.
* **Input Bisnis:** Definisi kebijakan JSONB (kondisi evaluasi & aksi), payload peristiwa audit (identitas aktor, IP, user agent, state lama, state baru), dan pesan event notifikasi.
* **Proses Utama Bisnis:**
  1. Evaluasi runtime aturan kebijakan organisasi (*Policy Engine Resolver*).
  2. Pencatatan log jejak audit kekal (*append-only audit persistence*).
  3. Pendistribusian notifikasi reaktif dalam aplikasi (*in-app notification dispatching*).
  4. Pengelolaan pengaturan konfigurasi sistem global.
* **Keputusan Bisnis yang Dibuat:**
  - Penetapan versi kebijakan organisasi aktif (jadwal, penalti, pajak).
  - Peringatan darurat saat terdeteksi anomali keamanan sistemik.
* **Output Bisnis:** Aturan bisnis runtime aktif, buku jejak audit abadi, dan pesan notifikasi terkirim.
* **Kapan Selesai:** Kebijakan diterapkan secara konsisten dan seluruh aksi tercatat permanen di audit log.
* **Domain Konsumen Output:** Seluruh domain bisnis lain (Workforce, Projects, Performance, Finance, People).

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — System Policy & Audit Governance |
| **Business Domain** | System & Integrity (Domain 10 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Sign-Off |
| **Business Owner** | Irfan (Super Administrator / Head of Technology, PT. ADT) |
| **Prepared By** | Lead Requirements Engineer & Compliance Analyst |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 8.10, 10.3, 11.6, 12.5, BR-001, BR-025) |

---

## 2. Executive Summary

Dokumen ini merinci kebutuhan bisnis untuk domain **System Integrity, Unified Policy & Audit Governance**. Tujuannya adalah menghilangkan ketergantungan pada logika bisnis yang di-hardcode melalui *Unified Policy Engine* (*BR-001, BR-012, BR-020*), menegakkan transparansi jejak audit mutlak yang tidak dapat dihapus oleh siapa pun (*BR-025, FR-047*), serta mengelola pusat notifikasi terdistribusi (*FR-046*).

---

## 3. Business Context & Problem

* **AS-IS:** Perubahan jam kerja kantor, penyesuaian tarif pajak, atau toleransi keterlambatan mengharuskan developer mengubah source code dan melakukan deployment ulang server. Jejak siapa yang menyetujui koreksi data seringkali hilang atau tidak tercatat lengkap dengan IP address dan stempel waktu server.
* **TO-BE:** Super Admin dapat mengubah aturan runtime melalui antarmuka web, perubahan langsung berlaku seketika di seluruh modul, dan 100% tindakan sensitif tercatat permanen dalam log audit append-only.

---

## 4. Business Objectives & Scope

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-SYS-01** | Kerapuhan hardcoding aturan bisnis | Seluruh parameter jadwal, penalti XP, dan pajak diatur via Unified Policy Engine | Fleksibilitas operasional tanpa downtime deploy | 100% perubahan aturan runtime aktif tanpa redeployment |
| **OBJ-SYS-02** | Kehilangan jejak audit keamanan | Seluruh mutasi data kritis dicatat dalam tabel audit append-only yang dilarang di-update/delete | Kepatuhan audit 100% & keamanan sistem terjamin | 0 record audit yang dapat dimodifikasi atau dihapus |
| **OBJ-SYS-03** | Keterlambatan informasi operasional | Pusat notifikasi real-time menyiarkan peringatan izin, lembur, dan reward | Responsivitas kerja tim meningkat | Rata-rata waktu sampai notifikasi ke pengguna $\le 1\text{ detik}$ |

---

## 5. Stakeholders & Business Actors

* **The Creator (Super Admin):** Penguasa tertinggi yang mengatur master policy engine dan memantau audit log keamanan.
* **Realm Warden (Admin):** Mengelola konfigurasi pengaturan sistem global aplikasi.
* **Seluruh Pengguna Ranah:** Penerima notifikasi sistem dan subjek audit keamanan.

---

## 6. Business Requirements & Rules

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis | Prioritas | Sumber PRD |
|---|---|---|---|
| **BRQ-SYS-001** | Bisnis mewajibkan tersedianya Unified Policy Engine terpusat untuk konfigurasi aturan bisnis runtime | P0 | FR-045, BR-001, BR-012, BR-020 |
| **BRQ-SYS-002** | Bisnis mewajibkan tersedianya Event-Driven Notification Center multi-kanal dalam aplikasi | P1 | FR-046 |
| **BRQ-SYS-003** | Bisnis mewajibkan pencatatan Audit Logging append-only yang memuat identitas, IP, UA, old_state, dan new_state | P0 | FR-047, BR-023, BR-025 |
| **BRQ-SYS-004** | Bisnis mewajibkan pengelolaan konfigurasi pengaturan sistem global secara terpusat | P1 | FR-048 |

* **BRULE-SYS-001 (Zero Hardcoding):** Seluruh formula penalti XP, pajak, dan toleransi jadwal wajib merujuk pada Policy Engine (*BR-001, BR-012, BR-020*).
* **BRULE-SYS-002 (Audit Immutability):** Tabel audit log dilarang memiliki fungsi UPDATE, DELETE, atau TRUNCATE tingkat database (*BR-025, FR-047*).

---

## 7. Traceability Matrix & Acceptance Criteria

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Acceptance Criteria |
|---|---|---|---|
| FR-045 (Policy Engine) | BRQ-SYS-001 | BRULE-SYS-001 | AC-SYS-001: Perubahan penalti keterlambatan pada tabel policy langsung berlaku seketika pada presensi berikutnya |
| FR-046 (Notifications) | BRQ-SYS-002 | — | AC-SYS-002: Pengajuan lembur oleh intern langsung memunculkan notifikasi toast dan icon alert pada supervisor |
| FR-047 (Audit Trail) | BRQ-SYS-003 | BRULE-SYS-002 | AC-SYS-003: Setiap persetujuan koreksi absensi menghasilkan rekaman audit baru yang memuat data sebelum dan sesudah perubahan |

---

## 8. BRD Completion Checklist

- [x] Seluruh kebutuhan Unified Policy, Notification Center, Audit Trail, dan Settings terdefinisi lengkap.
- [x] Aturan Zero Hardcoding (BR-001) dan Audit Immutability (BR-025) terpetakan eksplisit.

---
*Dokumen ini disimpan permanen di `brd/08-BRD-SYSTEM-POLICY-AUDIT-INTEGRITY.md`.*
