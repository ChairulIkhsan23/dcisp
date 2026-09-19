# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: Intelligence, Analytics & Executive Cockpit Portals
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**Intelligence, Analytics & Executive Cockpit Portals (Dasbor Eksekutif Command Center, Portal Terfokus Peran, Analitik Organisasi Mendalam, dan Kecerdasan Buatan Informatif)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain ini merupakan lapisan penyajian wawasan (*insights*), visualisasi metrik bisnis makro, dan ruang kendali terfokus bagi setiap persona pengguna di platform DCISP:
- Mengonsolidasikan data lintas domain (presensi, proyek, performa, dan keuangan) ke dalam dasbor eksekutif 4 kuadran (*Command Center / The Overworld Map - FR-051*): *TODAY*, *ATTENTION*, *PERFORMANCE*, dan *FINANCE*.
- Menyediakan antarmuka kerja harian layar tunggal yang terfokus bagi peserta magang (*Portal "My Day" / Player HUD - FR-052*) dan penyelia (*Portal "Team Today" / Party Roster - FR-053* serta *Dedicated Supervisor Workspace - FR-054*).
- Menghasilkan laporan analitik agregat organisasi mengenai kepatuhan kedisiplinan, efisiensi penuntasan milestone, dan perputaran kas bounty (*Analytics & Reporting - FR-049*).
- Menyediakan wawasan prediktif kecerdasan buatan (*AI Insights & Activity Intelligence - FR-050*) yang mematuhi prinsip **BR-015** (murni bersifat *informational & advisory*, dilarang mengambil keputusan penolakan/sanksi otomatis).
- Menegakkan navigasi hierarkis terisolasi peran (*Global Sidebar Navigation - FR-055, BR-001, BR-002*).

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Pengguna masuk ke platform, eksekutif memantau dasbor berkala, supervisor memeriksa antrean tindakan mendesak, atau intern mengoperasikan hari kerjanya.
* **Input Bisnis:** Agregasi log presensi harian, status tugas proyek, kurva perolehan XP, saldo mutasi dompet/buku besar, dan event anomali kerja.
* **Proses Utama Bisnis:**
  1. Agregasi metrik organisasi mendekati real-time pada 4 kuadran Command Center.
  2. Pengoperasian hari kerja terpadu pada portal layar tunggal "My Day" (kontrol sesi, pemilih tugas aktif, timer).
  3. Pemantauan radar tim dan peringatan anomali pada portal "Team Today".
  4. Analisis prediktif deteksi dini kejenuhan kerja (*burnout*) dan rekomendasi komposisi tim proyek oleh AI.
  5. Penyaringan struktur menu navigasi samping berdasarkan hak otorisasi dinamis.
* **Keputusan Bisnis yang Dibuat:**
  - Pengambilan tindakan segera atas item antrean pada widget ATTENTION (persetujuan lembur, cuti, koreksi, lamaran).
  - Keputusan strategis manajemen terkait alokasi kuota batch dan efisiensi serapan anggaran proyek.
* **Output Bisnis:** Dasbor pemantauan visual, grafik kurva metrik organisasi, rekomendasi saran AI, dan navigasi peran terisolasi.
* **Kapan Selesai:** Wawasan tersaji secara akurat dan keputusan operasional dapat dieksekusi dengan cepat.
* **Domain Konsumen Output:** Eksekutif PT. ADT, Supervisor, Project Manager, HR Admin, dan Intern.

### 4. Hubungan dengan Domain Lain (Related Domains)
* **Seluruh Domain Bisnis (Upstream Data Source):** Mengonsumsi data agregasi dari Workforce, Projects, Performance, Finance, dan People.

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — Intelligence, Analytics & Portals |
| **Business Domain** | Intelligence, Analytics & UI Portals (Domain 9 & 11 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Sign-Off |
| **Business Owner** | Head of Product & Executive Management (PT. Aplikasi Dagang Teknologi) |
| **Prepared By** | Lead Requirements Engineer & Product Analytics Lead |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 4, 7, 8.11, 8.12, 9, 13) |

---

## 2. Executive Summary

Dokumen ini merinci persyaratan bisnis untuk **Intelligence, Analytics & Executive Cockpit Portals**. Tujuannya adalah mengeliminasi navigasi berlapis yang membingungkan melalui portal-portal khusus terfokus (*My Day, Team Today, Command Center*), menyediakan pelaporan analitik metrik organisasi terpadu (*Section 13*), menghadirkan kecerdasan buatan informatif non-punitif (*BR-015*), serta memastikan menu navigasi terisolasi secara kaku sesuai peran (*BR-001, BR-002*).

---

## 3. Business Context & Problem

* **AS-IS:** Pimpinan perusahaan kesulitan memantau metrik kehadiran dan utilisasi anggaran proyek karena data tersebar di berbagai spreadsheet. Supervisor menghabiskan waktu berpindah-pindah menu hanya untuk memeriksa status presensi dan menyetujui lembur. Peserta magang bingung menentukan apa yang harus dikerjakan hari ini.
* **TO-BE:** Eksekutif memiliki kokpit *Command Center* 4 kuadran; supervisor memiliki *Team Today* untuk memantau anggota tim dalam satu pandangan mata; intern memiliki *My Day* sebagai kokpit layar tunggal harian.

---

## 4. Business Objectives & Scope

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-INT-01** | Ketiadaan visibilitas operasional eksekutif | Tersedianya Command Center 4 kuadran (*Today, Attention, Performance, Finance*) | Pengambilan keputusan cepat & deteksi anomali dini | Waktu resolusi antrean tindakan pada widget ATTENTION menurun |
| **OBJ-INT-02** | Friksi mikromanajemen supervisi | Portal "Team Today" menyajikan status tim real-time dengan kode warna tegas | Efisiensi supervisi & pembinaan tim cepat | 100% anomali kehadiran terdeteksi di dasbor supervisor |
| **OBJ-INT-03** | Kebingungan alur kerja harian peserta | Portal "My Day" menyatukan status presensi, timer kerja, pemilih tugas, dan XP | Produktivitas & fokus kerja peserta magang meningkat | 100% sesi kerja harian dijalankan melalui portal My Day |
| **OBJ-INT-04** | Risiko otomatisasi AI yang merugikan | AI murni menyajikan rekomendasi kecocokan skill & wawasan produktivitas advisory | Kepatuhan etika AI & keadilan keputusan manusia | 0 keputusan sanksi/penolakan yang dieksekusi otomatis oleh AI |

---

## 5. Stakeholders & Business Actors

* **Executive Management / Pimpinan:** Menggunakan Command Center untuk pemantauan makro organisasi.
* **Supervisor (Guild Master):** Menggunakan Team Today & Dedicated Workspace untuk monitoring dan persetujuan.
* **Intern (Adventurer):** Menggunakan portal My Day untuk mengoperasikan seluruh aktivitas kerja harian.
* **Project Manager (Quest Giver):** Memantau velocity milestone dan menggunakan rekomendasi AI skill match.

---

## 6. Business Requirements & Rules

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis | Prioritas | Sumber PRD |
|---|---|---|---|
| **BRQ-INT-001** | Bisnis mewajibkan penyediaan dasbor eksekutif Command Center berbasis 4 kuadran terdedikasi | P1 | FR-051 |
| **BRQ-INT-002** | Bisnis mewajibkan penyediaan portal kerja layar tunggal "My Day" bagi pelaksana tugas | P0 | FR-052, BR-005, BR-007 |
| **BRQ-INT-003** | Bisnis mewajibkan penyediaan portal operasional "Team Today" & Dedicated Supervisor Workspace | P1 | FR-053, FR-054 |
| **BRQ-INT-004** | Bisnis mewajibkan skema navigasi sidebar global merender menu dinamis sesuai permission role | P0 | FR-055, BR-001, BR-002 |
| **BRQ-INT-005** | Bisnis mewajibkan penyediaan analitik visual mendalam atas metrik kehadiran, proyek, dan keuangan | P1 | FR-049, Section 13 |
| **BRQ-INT-006** | Bisnis mewajibkan wawasan prediktif AI murni bersifat informatif non-punitif (advisory only) | P2 | FR-050, BR-015 |

* **BRULE-INT-001 (Scanner Navigation Isolation):** Sidebar operator scanner HANYA menampilkan menu tunggal `Attendance -> Scanner` (*BR-002*).
* **BRULE-INT-002 (Advisory AI Boundary):** Saran AI dilarang memicu penolakan kandidat atau pemotongan poin secara otomatis tanpa persetujuan manusia (*BR-015*).

---

## 7. Traceability Matrix & Acceptance Criteria

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Acceptance Criteria |
|---|---|---|---|
| FR-051 (Command Center) | BRQ-INT-001 | BR-001 | AC-INT-001: Dasbor eksekutif menampilkan headcount live, antrean tindakan, kurva XP, dan ringkasan keuangan |
| FR-052 (My Day Portal) | BRQ-INT-002 | BR-005, BR-007, BR-008 | AC-INT-002: Portal My Day menyajikan tombol kendali sesi kerja, precision timer, pemilih tugas, dan XP counter harian |
| FR-053, FR-054 (Supervisor Portals)| BRQ-INT-003 | BR-007, BR-013 | AC-INT-003: Portal Team Today menampilkan grid status tim (Working, Break, AFK) dan tombol persetujuan lembur/koreksi |
| FR-055 (Sidebar Navigation)| BRQ-INT-004 | BR-001, BR-002 | AC-INT-004: Menu sidebar merender menu terfilter dinamis; Scanner Operator hanya melihat menu scanner |
| FR-049 (Analytics) | BRQ-INT-005 | Section 13 | AC-INT-005: Dasbor analitik menampilkan rasio ketepatan waktu, integritas sesi kerja, dan serapan bounty |
| FR-050 (AI Insights) | BRQ-INT-006 | BR-015 | AC-INT-006: Wawasan AI disajikan sebagai teks rekomendasi tanpa adanya tombol eksekusi sanksi otomatis |

---

## 8. BRD Completion Checklist

- [x] Seluruh kebutuhan Command Center, My Day, Team Today, Supervisor Workspace, Analytics, dan AI terdefinisi lengkap.
- [x] Aturan isolasi navigasi (BR-002) dan batasan etika AI non-punitif (BR-015) terpetakan eksplisit.

---
*Dokumen ini disimpan permanen di `brd/09-BRD-INTELLIGENCE-ANALYTICS.md`.*
