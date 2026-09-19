# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: Performance Evaluation, XP Engine & Gamification System
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**Performance Evaluation, XP Engine & Gamification System (Sistem Evaluasi Kinerja Formal, Mesin Aturan XP Tiga Jalur, Progresi Peringkat, dan Pengakuan Prestasi)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain ini mengelola lapisan motivasi, reputasi, kedisiplinan, dan evaluasi kualitas profesional talenta magang di PT. Aplikasi Dagang Teknologi:
- Mengubah aktivitas harian yang terverifikasi (presensi tepat waktu, sesi kerja fokus, deliverable tugas proyek selesai) menjadi poin pengalaman (*Experience Points / XP*) yang transparan dan terkonfigurasi dinamis.
- Mengatur kenaikan tingkatan peringkat (*Rank Progression*) yang membuka paket hadiah bertingkat.
- Menegakkan pemisahan mutlak antara akumulasi poin gamifikasi (*Rank/XP*) dengan evaluasi mutu profesional formal (*Performance Score* skala 0–100 oleh supervisor) sesuai aturan **BR-017** ($\text{Rank} \neq \text{Performance Score}$).
- Mengorkestrasi pemilihan tepat SATU peserta terbaik (*Top Performer / MVP*) per kohort batch per periode evaluasi (*BR-018*).

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Terjadinya peristiwa operasional (presensi masuk tepat waktu/terlambat, laporan tugas disetujui, evaluasi berkala supervisor tiba, atau berakhirnya periode evaluasi batch).
* **Input Bisnis:** Event presensi dan keterlambatan, deliverable tugas yang disetujui, nilai rubrik evaluasi berkala supervisor (kualitas kerja, integritas, inisiatif, kerja sama), serta ambang batas XP rank pada policy engine.
* **Proses Utama Bisnis:**
  1. Penghitungan poin pengalaman dinamis yang terpartisi ke dalam 3 jalur terpisah (*Internship XP, Project XP, Alumni Contribution* - *BR-004*).
  2. Penegakan matriks sanksi penalti pengurangan XP atas keterlambatan, istirahat tidak sah, mangkir, dan lupa checkout (*BR-012*).
  3. Kenaikan level otomatis (*Rank Promotion*) saat ambang batas XP terpenuhi.
  4. Komputasi skor kinerja komposit formal (*Performance Score*) berbasis formula berbobot.
  5. Penentuan tepat SATU *Top Performer* per batch pada setiap siklus evaluasi.
  6. Pembukaan lencana pencapaian khusus (*Achievement Badges*) dan visualisasi grafik radar keahlian (*Skill Growth Matrix*).
* **Keputusan Bisnis yang Dibuat:**
  - Penetapan promosi tingkatan rank peserta (Tier Novice $\rightarrow$ Grandmaster).
  - Penilaian skor kualitas formal berkala oleh Supervisor.
  - Penetapan 1 peserta magang terbaik (*MVP of the Season*) per batch.
* **Output Bisnis:** Saldo XP terpartisi, status level rank pengguna, dokumen evaluasi kinerja formal, rekam jejak lencana prestasi, dan tiket pembukaan paket reward promosi.
* **Kapan Selesai:** Berakhirnya siklus evaluasi batch dan penetapan Top Performer serta penerbitan predikat kinerja kelulusan.
* **Domain Konsumen Output:**
  - *Domain Incentives & Compensation* (menerima event promosi rank untuk menerbitkan paket reward uang tunai/merchandise).
  - *Domain Documents* (mengonsumsi data capaian rank, evaluasi, dan badge untuk sertifikat kelulusan dan portofolio alumni).
  - *Domain People* (mengonsumsi nilai performa formal sebagai prasyarat kelulusan magang).

### 4. Hubungan dengan Domain Lain (Related Domains)
* **Domain Workforce & Attendance (Upstream Trigger):** Mengirimkan data ketepatan waktu, keterlambatan, dan kepatuhan istirahat.
* **Domain Projects & Tasks (Upstream Trigger):** Mengirimkan penyelesaian tugas dan rating verifikasi deliverable.
* **Domain Incentives & Compensation (Downstream Consumer):** Mengonsumsi event kenaikan rank untuk klaim hadiah multi-komponen.

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — Performance & Gamification |
| **Business Domain** | Performance & Gamification Management (Domain 5 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Business Sign-Off |
| **Business Owner** | HR Head & Program Lead (PT. Aplikasi Dagang Teknologi) |
| **Prepared By** | Lead Requirements Engineer & Business Analyst |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 2, 3, 5, 6.2.5, 8.6, 10) |

---

## 2. Executive Summary

Dokumen ini mengatur persyaratan bisnis untuk **Performance Evaluation, XP Engine & Gamification System**. Tujuannya adalah menghadirkan sistem penghargaan kinerja yang objektif, transparan, dan memotivasi kerja profesional melalui partisi 3 skema XP (*BR-004*), pemisahan tegas antara status rank gamified dengan skor evaluasi formal (*BR-017*), pengakuan tepat satu Top Performer per batch (*BR-018*), serta penegakan etika gamifikasi tanpa mekanik eksploitatif.

---

## 3. Business Context

Sebelum DCISP dibangun:
1. Terjadi kerancuan fatal di mana skor keaktifan/kehadiran disamakan dengan penilaian kualitas kerja teknis.
2. Tidak adanya sistem apresiasi pencapaian bertahap sehingga peserta magang kehilangan motivasi di tengah periode magang.
3. Penetapan peserta terbaik (*Top Performer*) dilakukan secara subjektif tanpa formula komposit berbasis data yang dapat dipertanggungjawabkan.

---

## 4. Business Problem

1. **Kerancuan Metrik Gamifikasi vs Mutu:** Anggapan keliru bahwa peserta dengan XP tertinggi otomatis memiliki mutu kerja teknis terbaik ($\text{Rank} \neq \text{Performance Score}$).
2. **Ketiadaan Transparansi Penalti:** Sanksi keterlambatan tidak memiliki dasar perhitungan yang jelas dan konsisten.
3. **Pencampuran Prestasi Alumni vs Peserta Aktif:** Kontribusi alumni mencemari persaingan papan peringkat peserta magang yang baru belajar.

---

## 5. Business Objectives

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-PRF-01** | Kerancuan evaluasi mutu vs jam terbang | Pemisahan mutlak antara level rank (XP) dengan nilai evaluasi kinerja formal (0–100) | Keadilan penilaian kinerja profesional & akreditasi kelulusan teruji | 100% peserta menerima skor evaluasi berkala formal supervisor |
| **OBJ-PRF-02** | Ketidakadilan persaingan senior vs magang baru | Partisi 3 jalur XP terisolasi (*Internship XP, Project XP, Alumni Contribution*) | Keadilan kompetisi di lingkungan magang aktif | 0 transaksi poin alumni yang mengubah standing ranking magang aktif |
| **OBJ-PRF-03** | Subjektivitas penentuan peserta terbaik | Penetapan tepat SATU Top Performer per batch per periode berbasis formula komposit terukur | Standar keunggulan objektif & pengakuan prestasi berintegritas | Tepat 1 Top Performer disahkan per batch per siklus |

---

## 6. Business Scope

### 6.1 In Scope
1. **XP Rules Engine Dinamis (FR-025, BR-012):** Perhitungan poin positif dan penalti deduksi keterlambatan/mangkir terkonfigurasi runtime.
2. **Partisi Tiga Jalur XP (FR-025, BR-004):** Pemisahan mutlak *Internship XP (Main XP)*, *Project XP (Guild XP)*, dan *Alumni Contribution (Veteran XP)*.
3. **Sistem Progresi Peringkat / Rank (FR-026, BR-017):** Kenaikan tingkatan rank bertingkat (Novice $\rightarrow$ Grandmaster).
4. **Performance Evaluation Engine Formal (FR-027, BR-017):** Penilaian rubrik berkala supervisor (kehadiran, tugas, mutu, kontribusi).
5. **Top Performer per Batch (FR-028, BR-018):** Algoritma penetapan tepat satu pemenang tunggal per kohort batch.
6. **Achievement System (FR-029):** Katalog lencana prestasi milestone (badges).
7. **Skill Growth Matrix (FR-030):** Grafik radar perkembangan kompetensi individu seiring proyek.

### 6.2 Out of Scope
1. **Mekanik Gamifikasi Adiktif / Judi:** Dilarang menghadirkan loot boxes, gacha, roda keberuntungan (gambling), atau mekanik eksploitatif non-profesional.
2. **Evaluasi Kinerja Karyawan Tetap:** Khusus untuk peserta magang dan talenta proyek alumni.

---

## 7. Stakeholders

| Stakeholder | Responsibility | Business Interest | Decision Authority |
|---|---|---|---|
| **Supervisor (Guild Master)** | Mengisi rubrik evaluasi berkala, menilai kualitas deliverable dan perilaku kerja | Kualitas mutu tim dan pembinaan kompetensi | Mengesahkan nilai rubrik evaluasi formal supervisor |
| **HR Admin / Program Lead** | Meninjau performa kohort batch dan mengesahkan penetapan Top Performer | Kepatuhan standar kompetensi kelulusan dan rekrutmen masa depan | Mengesahkan predikat *Top Performer per Batch* |
| **Intern (Adventurer)** | Mengumpulkan XP kedisiplinan dan kualitas tugas, membuka badge prestasi | Kenaikan level rank, pembukaan reward, dan reputasi portofolio | Mengklaim paket reward kenaikan level rank |
| **Alumni (Legendary Hero)** | Mengumpulkan Veteran XP dari proyek publik tanpa mengubah ranking magang aktif | Repositori portofolio kompetensi dan apresiasi kontribusi lepas | Mengelola showcase badge pada profil publik |

---

## 8. Business Actors

| Business Role | System Role Terkait | Tindakan Utama | Batasan Akses Data |
|---|---|---|---|
| **Performance Evaluator** | Supervisor | Mengisi form penilaian kinerja berkala komposit | `scope = assigned_team` |
| **Gamification Arbiter** | HR Admin / Super Admin | Mengatur formula bobot performa, mengesahkan Top Performer | `scope = workforce_and_people` |
| **Quest Hero** | Intern / Alumni | Memperoleh XP dari aktivitas sah, menaikkan rank level | `scope = own_data` |

---

## 9. Current Business Process (AS-IS)

* **AS-IS:** Penilaian magang hanya dilakukan di akhir periode menggunakan lembar penilaian kampus manual tanpa data rekam jejak kehadiran dan tugas harian yang terukur. Skor akhir seringkali bersifat subjektif dan tidak mencerminkan kontribusi nyata selama magang.

---

## 10. Target Business Process (TO-BE)

```text
[ Aktivitas Terverifikasi (Presensi Tepat / Tugas Selesai / Lembur Sah) ]
                                   │
                                   ▼
[ XP Rules Engine Menghitung Poin ke 3 Partisi Terpisah ]
                                   │
                                   ├──► [ Akumulasi Main XP Mencapai Ambang ] ──► [ Promosi Level Rank & Klaim Loot ]
                                   │
[ Periode Evaluasi Berkala Tiba (Bulanan / Akhir Batch) ]
                                   │
                                   ▼
[ Supervisor Mengisi Rubrik Formal (Performance Score: 0–100) ]
                                   │
                                   ▼
[ Komposit: Performance Score + Final Evaluation + Attendance + Project ]
                                   │
                                   ▼
[ Penetapan Tepat SATU Top Performer per Batch ] ──► [ Dipajang di Hall of Heroes ]
```

---

## 11. Business Process Scenarios

### Scenario 1: Perolehan Poin XP & Kenaikan Tingkat Rank
* **Trigger:** Intern menyelesaikan tugas proyek berbobot tinggi yang disetujui reviewer (+50 XP).
* **Precondition:** Intern berada pada Rank D (Level Novice: 470 XP, Ambang Rank C = 500 XP).
* **Actor:** Intern, Sistem XP Engine.
* **Main Flow:**
  1. Reviewer menyetujui tugas $\rightarrow$ XP Rules Engine memproses penambahan +50 XP pada skema `Internship XP`.
  2. Saldo total menjadi 520 XP (melampaui ambang batas 500 XP).
  3. Sistem memicu event promosi rank: status pengguna naik menjadi **Rank C (Apprentice)**.
  4. Antarmuka menampilkan modal *Level Up Fanfare* dan menerbitkan tiket hak klaim paket reward promosi rank.
* **Output:** Status pengguna diperbarui ke Rank C; mutasi XP dicatat pada buku besar `XP Transaction`.

### Scenario 2: Evaluasi Kinerja Formal Supervisor
* **Trigger:** Akhir bulan masa magang tiba.
* **Actor:** Supervisor (Evaluator), Intern.
* **Main Flow:**
  1. Supervisor membuka modul *Evaluations* pada ruang kerja supervisor.
  2. Sistem menyajikan agregasi data objektif bulan tersebut: tingkat ketepatan hadir 96%, tugas selesai tepat waktu 8/8, kualitas deliverable rata-rata 4.8/5.0.
  3. Supervisor mengisi nilai rubrik kualitatif: komunikasi (4/5), inisiatif (5/5), kerja sama tim (4/5).
  4. Performance Evaluation Engine menghitung nilai komposit formal: **Performance Score = 92.5 / 100**.
  5. Nilai formal disimpan dan terpisah dari total akumulasi poin XP gamifikasi pengguna (*BR-017*).
* **Output:** Rekam evaluasi formal disahkan; ringkasan performa terbarui di profil intern.

### Scenario 3: Penetapan Top Performer per Batch
* **Trigger:** Penutupan siklus evaluasi akhir batch magang.
* **Actor:** HR Admin, Sistem Algoritma.
* **Main Flow:**
  1. HR Admin memicu kalkulasi penentuan Top Performer untuk Batch 2026-01.
  2. Sistem merangking seluruh anggota kohort berdasarkan formula komposit: $Performance Score + Final Evaluation + Project Contribution + Attendance Rate$.
  3. Sistem menetapkan tepat **SATU** peserta dengan nilai komposit tertinggi sebagai *Top Performer (MVP of the Season)* (*BR-018*).
  4. Profil pemenang disematkan badge khusus dan dipajang di *Hall of Fame / Hall of Heroes*.
* **Output:** Record Top Performer disahkan; sertifikat penghargaan khusus diterbitkan.

---

## 12. Business Requirements

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis | Rasional Bisnis | Prioritas | Sumber PRD |
|---|---|---|---|---|
| **BRQ-PRF-001** | Bisnis mewajibkan akumulasi poin pengalaman (XP) dipartisi secara mutlak ke dalam 3 jalur terpisah (*Internship XP, Project XP, Alumni Contribution*) | Menjamin keadilan kompetisi dan mencegah pencemaran standing magang aktif oleh alumni | P0 | FR-025, BR-004 |
| **BRQ-PRF-002** | Bisnis mewajibkan seluruh parameter penalti pengurangan XP atas pelanggaran kehadiran dikendalikan secara dinamis via policy engine | Menghindari hardcoding sanksi dan memungkinkan adaptasi kebijakan kedisiplinan | P0 | FR-025, BR-012 |
| **BRQ-PRF-003** | Bisnis mewajibkan pemisahan konseptual dan visual mutlak antara Rank gamifikasi dengan Skor Kinerja Formal (*Performance Score*) | Mencegah kekeliruan fatal antara jam terbang aktivitas dengan kualitas mutu teknis | P0 | FR-026, FR-027, BR-017 |
| **BRQ-PRF-004** | Bisnis mewajibkan pengakuan tepat SATU peserta terbaik (*Top Performer*) per kohort batch per periode evaluasi berbasis data komposit | Memberikan apresiasi prestasi tertinggi yang objektif dan berintegritas | P1 | FR-028, BR-018 |
| **BRQ-PRF-005** | Bisnis mewajibkan penyediaan katalog lencana pencapaian (*Achievement Badges*) untuk mengapresiasi tonggak milestone kerja | Mendorong pencapaian standar kerja unggul di luar tugas rutin harian | P1 | FR-029 |
| **BRQ-PRF-006** | Bisnis mewajibkan tersedianya visualisasi matriks pertumbuhan keahlian individu seiring penyelesaian proyek nyata | Memberikan visibilitas perkembangan kompetensi riil peserta kepada supervisor | P1 | FR-030 |

---

## 13. Business Rules

| ID Aturan | Nama Aturan Bisnis | Kondisi Bisnis | Tindakan Sistem | Pengecualian |
|---|---|---|---|---|
| **BRULE-PRF-001** | *XP Scheme Separation* | Transaksi mutasi XP diproses | Simpan strictly pada kolom skema yang sesuai; transaksi alumni dilarang memotong/menambah `internship_xp` | Tidak ada pengecualian (BR-004) |
| **BRULE-PRF-002** | *Configurable Penalties* | Pelanggaran kehadiran terdeteksi (terlambat, unauthorized break, mangkir, missing checkout) | Potong XP sesuai matriks policy aktif: terlambat 1-15m: -1 XP, 16-30m: -2 XP, >30m: -3 XP, mangkir: -5 XP | Cuti disetujui = 0 penalti XP |
| **BRULE-PRF-003** | *Rank vs Performance Score* | Data reputasi pengguna ditampilkan | $\text{Rank} \neq \text{Performance Score}$. Dilarang menyamakan rank level dengan nilai evaluasi mutu | Tidak ada pengecualian (BR-017) |
| **BRULE-PRF-004** | *Singular Top Performer per Batch* | Pemenang batch dihitung | Tepat SATU orang diakui sebagai Top Performer per batch per periode evaluasi | Tidak ada pemenang ganda (BR-018) |
| **BRULE-PRF-005** | *Floor Limit on Cumulative XP* | Penalti absensi memotong saldo XP | Saldo kumulatif Internship XP untuk penentuan rank tidak boleh bernilai negatif (floor limit = 0 XP) | Pencatatan mutasi riil tetap mencatat nilai minus |

---

## 14. Business Entities

1. **XP Rule:** Master aturan pemicu poin/penalti (event_trigger, xp_value, xp_scheme, description).
2. **XP Transaction:** Buku besar mutasi poin pengalaman append-only (user_id, scheme, points, running_balance, reference_event).
3. **Rank:** Master level peringkat gamifikasi (name, min_xp, level_order, badge_icon_url).
4. **Achievement:** Master lencana prestasi (code, title, description, badge_icon_url, reward_xp).
5. **Performance Evaluation:** Rekaman evaluasi formal komposit (user_id, evaluator_id, batch_id, attendance_score, task_delivery_score, work_quality_score, supervisor_rubric_score, composite_performance_score, is_top_performer).

---

## 15. Data Flow

```text
[ Presensi / Tugas Disetujui ] ──► [ XP Rules Engine ] ──► [ XP Transactions (3 Skema) ]
                                                                     │
                                                                     ▼
                                                          [ Rank Progression System ]
                                                                     │
[ Evaluasi Rubrik Supervisor ] ──► [ Performance Engine ]            ▼
                 │                                        [ Paket Hadiah Terbuka ]
                 ▼
     [ Performance Score (0-100) ]
                 │
                 ▼
[ Komposit Penentuan Top Performer ] ──► [ 1 MVP per Batch di Hall of Fame ]
```

---

## 16. Status & Lifecycle

```text
RANK PROMOTION LIFECYCLE:
[ TIER 1: NOVICE ] ──(250 XP)──► [ TIER 2: APPRENTICE ] ──(750 XP)──► [ TIER 3: KNIGHT ]
                                                                              │
                                                                         (1500 XP)
                                                                              ▼
[ TIER 5: GRANDMASTER ] ◄────────────── (3000 XP) ───────────── [ TIER 4: PALADIN ]
```

---

## 17. Approval & Decision Flow

* **Pengesahan Evaluasi Mutu:** Requester: Sistem (Akhir Periode) $\rightarrow$ Approver: Supervisor Langsung $\rightarrow$ Status evaluasi terkunci.
* **Pengesahan Top Performer:** Requester: Algoritma Komposit $\rightarrow$ Approver: HR Head / Program Lead $\rightarrow$ Status Top Performer disematkan.

---

## 18. Exception & Failure Handling

* **Nilai Evaluasi Belum Lengkap:** Sistem memblokir finalisasi evaluasi akhir batch jika masih ada supervisor yang belum mengisi rubrik penilaian.

---

## 19. Business Reporting

1. **Laporan Distribusi XP & Progresi Rank Kohort:** Kurva kecepatan kenaikan level pengguna per batch dan deteksi anomali keterlambatan rata-rata.
2. **Laporan Rekapitulasi Evaluasi Kinerja Formal:** Daftar nilai komposit seluruh peserta magang untuk dasar penerbitan sertifikat kelulusan.

---

## 20. KPI & Success Metrics

* **XP Growth Velocity (MTR-PRF-01):** Rata-rata akumulasi penambahan XP mingguan per peserta aktif (`Target: TBD`).
* **Evaluation Quality Index (MTR-PRF-02):** Rata-rata skor evaluasi kinerja formal berkala pada skala 0–100 (`Target: TBD`).

---

## 21. Business Integration

* **Ke Domain Incentives & Finance:** Mengirimkan trigger kenaikan rank untuk menerbitkan hak klaim paket reward promosi rank.
* **Ke Domain Documents:** Mengirimkan capaian rank, skor performa, dan status Top Performer untuk sertifikat resmi.

---

## 22. Business Dependencies

* **Upstream:** Workforce Module (presensi) dan Projects Module (deliverable tugas).

---

## 23. Business Constraints

* **Non-Eksploitatif (BR-010):** Dilarang menghadirkan gameplay manipulatif; seluruh poin bersandar pada kerja riil.

---

## 24. Business Assumptions

* Supervisor memiliki kompetensi teknis untuk menilai kualitas artefak deliverable tugas tim binaannya secara adil.

---

## 25. Business Gaps

* Pembobotan matematis eksak antar-komponen dalam formula Performance Score berstatus `[TBD — OQ-004]`.

---

## 26. Ambiguities

* *Rank adalah ukuran jam terbang aktivitas; Performance Score adalah ukuran mutu profesional (BR-017).*

---

## 27. Conflicting Requirements

* *Tidak ada konflik requirement.*

---

## 28. Traceability Matrix

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Entitas Terkait | Acceptance Criteria |
|---|---|---|---|---|
| FR-025 (XP Engine) | BRQ-PRF-001, BRQ-PRF-002 | BRULE-PRF-001, BRULE-PRF-002 | `XP Rule`, `XP Transaction` | AC-PRF-001 |
| FR-026 (Rank Progression)| BRQ-PRF-003 | BRULE-PRF-003, BRULE-PRF-005 | `Rank`, `User` | AC-PRF-002 |
| FR-027 (Performance Eval)| BRQ-PRF-003 | BRULE-PRF-003 | `Performance Evaluation` | AC-PRF-003 |
| FR-028 (Top Performer) | BRQ-PRF-004 | BRULE-PRF-004 | `Performance Evaluation`, `Batch` | AC-PRF-004 |
| FR-029 (Achievements) | BRQ-PRF-005 | — | `Achievement` | AC-PRF-005 |
| FR-030 (Skill Growth) | BRQ-PRF-006 | — | `Skill`, `User` | AC-PRF-006 |

---

## 29. Acceptance Criteria

### AC-PRF-001: Partisi Skema XP Terisolasi
* **Given:** Pengguna berstatus Alumni menyelesaikan tugas proyek publik.
* **When:** Reviewer menyetujui penyerahan tugas (+30 XP).
* **Then:** Sistem mengkreditkan +30 XP ke skema `Alumni Contribution / Veteran XP` tanpa mengubah standing ranking peserta magang aktif.

### AC-PRF-002: Promosi Rank Otomatis Saat Ambang Batas Tercapai
* **Given:** Intern berada pada Rank D (490 XP) dengan ambang batas Rank C adalah 500 XP.
* **When:** Intern memperoleh +20 XP dari ketepatan kehadiran dan tugas.
* **Then:** Sistem menaikkan status menjadi Rank C, mencatat event promosi rank, dan membuka hak klaim paket reward terkait.

### AC-PRF-003: Singularitas Top Performer per Batch
* **Given:** Siklus evaluasi Batch 2026-01 selesai dan 2 peserta memiliki skor yang sangat dekat (95.4 vs 95.1).
* **When:** Algoritma pemeringkat komposit dijalankan.
* **Then:** Sistem menetapkan tepat SATU peserta berperingkat teratas (skor 95.4) sebagai Top Performer dan menyematkan badge MVP ke profilnya.

---

## 30. Open Questions

* `OQ-PRF-01 (Ref OQ-003)`: Nilai ambang batas XP pasti untuk setiap jenjang rank (`TBD`).
* `OQ-PRF-02 (Ref OQ-004)`: Pembobotan matematis eksak komponen formula Performance Score (`TBD`).

---

## 31. BRD Completion Checklist

- [x] Seluruh kebutuhan Performance, XP Engine, dan Gamification terdefinisi lengkap.
- [x] Aturan pemisahan Rank vs Performance Score (BR-017) dan Top Performer tunggal (BR-018) terpetakan eksplisit.

---
*Dokumen ini disimpan permanen di `brd/04-BRD-PERFORMANCE-GAMIFICATION.md`.*
