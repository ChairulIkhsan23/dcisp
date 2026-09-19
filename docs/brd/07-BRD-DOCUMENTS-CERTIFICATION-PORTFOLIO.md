# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: Documents, Certification, Portfolio & Digital Assets
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**Documents, Certification, Portfolio & Digital Assets (Penerbitan Dokumen Resmi, Kartu Identitas Digital, Mesin Sertifikat Kelulusan Berverifikasi, Pembangun Portofolio Publik Alumni, dan Penyimpanan Cloudflare R2)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain ini mengatur seluruh penerbitan artefak formal, dokumen legalitas, dan representasi reputasi digital peserta:
- Mengubah rekam jejak kerja terverifikasi menjadi aset portofolio profesional permanen (*Digital Portfolio Builder / Hero’s Chronicle*) yang mendukung karir jangka panjang alumni sesuai **BR-003**.
- Menerbitkan dokumen identitas resmi ber-QR (*Adventurer License / Digital ID Card*) untuk akses fisik gerbang kantor.
- Menerbitkan sertifikat kelulusan digital resmi (*Legendary Scroll / Certificate Engine*) dengan segel hash verifikasi publik.
- Mengompilasi laporan berkala magang dan logbook kegiatan (*Adventurer’s Journal*) ke dalam format PDF standar akademik.
- Menerapkan aturan penyimpanan mutlak **BR-024** di mana seluruh berkas biner fisik disimpan di Cloudflare R2, sedangkan database SQL hanya menyimpan metadata berkas (DATA-007).

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Aktivasi peserta magang baru (terbit ID Card), kelulusan peserta (terbit Sertifikat), kebutuhan ekspor logbook akademik, atau publikasi portofolio oleh alumni.
* **Input Bisnis:** Data profil pengguna, foto resmi, rekam jejak proyek & tugas yang disetujui, stempel tanda tangan digital pejabat, dan metadata berkas fisik R2.
* **Proses Utama Bisnis:**
  1. Pembuatan kartu identitas digital (ID Card) dengan kode QR/NFC terenkripsi.
  2. Penerbitan sertifikat kelulusan digital ber-QR verifikasi publik (*Certificate Engine*).
  3. Kompilasi otomatis portofolio publik alumni berbasis deliverable terverifikasi (*Portfolio Builder*).
  4. Generator laporan dokumen resmi PDF (logbook kegiatan).
  5. Pengelolaan penyimpanan berkas terisolasi di Cloudflare R2 via presigned URL.
* **Keputusan Bisnis yang Dibuat:**
  - Pengesahan penerbitan sertifikat kelulusan oleh pejabat berwenang.
  - Pengaturan visibilitas publik/privat pada portofolio alumni.
* **Output Bisnis:** Kartu ID Card digital/cetak, sertifikat kelulusan PDF dengan URL validasi publik, laman portofolio digital dinamis (`/portfolio/slug`), dan dokumen laporan PDF resmi.
* **Kapan Selesai:** Dokumen resmi diterbitkan, disimpan permanen di R2 Vault, dan dapat diverifikasi publik.
* **Domain Konsumen Output:** Publik eksternal / Rekruter (melihat portofolio & validasi sertifikat), Institusi Pendidikan Kampus (menerima logbook PDF).

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — Documents, Certification & Portfolio |
| **Business Domain** | Documents & Storage (Domain 8 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Sign-Off |
| **Business Owner** | HR Head & Operations Lead (PT. Aplikasi Dagang Teknologi) |
| **Prepared By** | Lead Requirements Engineer & Document Systems Analyst |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 8.9, 10.3, 11.4, BR-024) |

---

## 2. Executive Summary

Dokumen ini merinci kebutuhan bisnis untuk domain **Documents, Certification, Portfolio & Digital Assets**. Tujuannya adalah mengotomatiskan penerbitan kartu identitas digital, sertifikat kelulusan berverifikasi QR publik, laporan logbook akademik resmi, serta showcase portofolio digital alumni dengan mematuhi aturan penyimpanan Cloudflare R2 tanpa membebani database utama (*BR-024*).

---

## 3. Business Context & Problem

* **AS-IS:** Kartu magang dan sertifikat dibuat manual menggunakan software desain grafis satu per satu. Validasi keaslian sertifikat oleh pihak luar sulit dibuktikan. Alumni tidak memiliki repositori portofolio otomatis dari hasil kerja magang mereka.
* **TO-BE:** Sistem menerbitkan dokumen secara otomatis, menyematkan kode verifikasi publik yang dapat dipindai siapa saja, dan mengompilasi portofolio alumni secara instan dari deliverable yang disetujui.

---

## 4. Business Objectives & Scope

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-DOC-01** | Pemalsuan sertifikat kelulusan | Seluruh sertifikat memiliki kode registrasi unik & QR verifikasi publik | Kredibilitas institusi & keabsahan legalitas lulusan | 100% sertifikat kelulusan dapat divalidasi keasliannya secara online |
| **OBJ-DOC-02** | Kehilangan bukti portofolio alumni | Portofolio digital terkompilasi otomatis dari proyek & tugas nyata | Nilai jual talenta alumni di industri meningkat | 100% lulusan memiliki laman portofolio digital siap bagikan |
| **OBJ-DOC-03** | Pembengkakan ukuran database | Seluruh file fisik biner dialirkan ke Cloudflare R2; database hanya simpan metadata | Efisiensi infrastruktur & performa database cepat | 0 file biner (BLOB) tersimpan di database relasional |

---

## 5. Stakeholders & Business Actors

* **HR Admin / Penandatangan Sah:** Mengesahkan penerbitan sertifikat kelulusan resmi.
* **Intern / Alumni:** Mengunduh kartu identitas, membagikan link portofolio publik, mencetak logbook akademik.
* **Pihak Ketiga / Rekruter Industri:** Memindai QR sertifikat untuk memvalidasi keaslian kelulusan kandidat.

---

## 6. Business Requirements & Rules

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis | Prioritas | Sumber PRD |
|---|---|---|---|
| **BRQ-DOC-001** | Bisnis mewajibkan pembuatan Kartu Identitas Digital (ID Card) ber-QR/NFC otomatis saat peserta aktif | P1 | FR-040 |
| **BRQ-DOC-002** | Bisnis mewajibkan penerbitan Sertifikat Kelulusan Digital resmi dengan QR verifikasi publik saat lulus | P1 | FR-041, OQ-001 |
| **BRQ-DOC-003** | Bisnis mewajibkan penyediaan Digital Portfolio Builder bagi alumni berbasis deliverable nyata | P1 | FR-042, BR-003 |
| **BRQ-DOC-004** | Bisnis mewajibkan ekspor laporan formal magang dan logbook kegiatan ke format PDF standar | P1 | FR-043 |
| **BRQ-DOC-005** | Bisnis mewajibkan pemisahan penyimpanan berkas fisik di Cloudflare R2 (database hanya simpan metadata) | P0 | FR-044, BR-024 |

* **BRULE-DOC-001 (R2 Separation):** Database SQL dilarang menyimpan data biner; seluruh file diunggah langsung ke R2 via presigned URL (*BR-024*).
* **BRULE-DOC-002 (Permanent Retention):** Dokumen sertifikat dan portofolio alumni disimpan secara permanen seumur hidup (*BR-003*).

---

## 7. Traceability Matrix & Acceptance Criteria

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Acceptance Criteria |
|---|---|---|---|
| FR-040 (ID Card) | BRQ-DOC-001 | — | AC-DOC-001: ID Card digital terbit otomatis memuat foto, NIM, dan QR code presensi |
| FR-041 (Certificate) | BRQ-DOC-002 | OQ-001, OQ-002 | AC-DOC-002: Sertifikat kelulusan memiliki QR yang jika dipindai mengarah ke URL validasi resmi |
| FR-042 (Portfolio) | BRQ-DOC-003 | BR-003 | AC-DOC-003: Laman `/portfolio/slug` menampilkan proyek, peran, badge, dan evidence terverifikasi |
| FR-044 (R2 Storage) | BRQ-DOC-005 | BR-024 | AC-DOC-004: File bukti tugas terunggah ke Cloudflare R2 dan record metadata tersimpan di SQL |

---

## 8. BRD Completion Checklist

- [x] Seluruh kebutuhan Documents, Certification, Portfolio, dan Storage terdefinisi lengkap.
- [x] Aturan pemisahan penyimpanan Cloudflare R2 (BR-024) terpetakan eksplisit.

---
*Dokumen ini disimpan permanen di `brd/07-BRD-DOCUMENTS-CERTIFICATION-PORTFOLIO.md`.*
