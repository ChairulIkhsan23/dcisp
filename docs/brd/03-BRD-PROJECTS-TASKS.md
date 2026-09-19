# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: Projects, Tasks, Evidence & Contribution Management
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**Projects, Tasks, Evidence & Contribution Management (Manajemen Bursa Proyek, Penugasan Tugas, Artefak Bukti Kerja, dan Rekonsiliasi Kontribusi 3-Lapis)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain ini merupakan inti operasional kerja nyata (*project-based work*) pada platform DCISP:
- Proyek diterbitkan di bursa (*Project Marketplace*), direkrut melalui alur lamaran dengan batasan kuota tim, dipecah menjadi fase pencapaian (*Milestones*) dan unit pekerjaan terkecil (*Tasks*).
- Pelaksanaan tugas mewajibkan penyerahan laporan pertanggungjawaban (*Work Report & Submission*) yang disertai bukti deliverable terverifikasi (*Evidence Management*) yang tersimpan di Cloudflare R2 Vault.
- Evaluasi penyelesaian tugas bermuara langsung pada perhitungan kontribusi tim berkeadilan melalui model rekonsiliasi tiga lapis (*Three-Layer Contribution Engine*: Planned $\rightarrow$ Actual $\rightarrow$ Final %) yang menjadi dasar mutlak bagi pembagian hak imbalan finansial (*Bounty*).

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Kebutuhan inisiatif proyek baru dari Project Manager, pembukaan lowongan anggota tim, atau penugasan tugas kerja harian.
* **Input Bisnis:** Spesifikasi proyek (judul, deskripsi, skill, kuota, tenggat waktu, alokasi bounty pool, tingkat visibilitas), formulir lamaran kandidat, kartu tugas kanban, laporan kemajuan kerja (*what_i_did*, *problems*, *next_actions*, *progress %*), dan berkas bukti kerja (tautan Git commit/PR, screenshot, dokumen).
* **Proses Utama Bisnis:**
  1. Publikasi proyek bursa dengan 3 tingkat visibilitas (`INTERN_ONLY`, `PUBLIC`, `PRIVATE`).
  2. Seleksi pelamar, perhitungan kecocokan skill informatif (*Skill Matching Divination*), dan penguncian kuota party atomik.
  3. Pembentukan struktur tim proyek dan penetapan kesepakatan awal kontribusi (*Planned Contribution %*).
  4. Manajemen deliverable proyek berjenjang: $Project \rightarrow Milestone \rightarrow Task \rightarrow Assignment \rightarrow Submission$.
  5. Penelaahan dan verifikasi bukti artefak deliverable oleh *Reviewer (The Oracle)*.
  6. Rekonsiliasi kontribusi 3-lapis (*Planned* kesepakatan $\rightarrow$ *Actual* komputasi sistem $\rightarrow$ *Final* pengesahan supervisor).
* **Keputusan Bisnis yang Dibuat:**
  - Penerimaan/penolakan pelamar proyek oleh Project Manager.
  - Pengesahan kelayakan hasil tugas: *Approved* vs *Revision Required*.
  - Penetapan persentase *Final Contribution %* anggota tim proyek oleh Supervisor (total wajib 100,00%).
* **Output Bisnis:** Status deliverable proyek selesai, repositori bukti kerja terverifikasi di Cloudflare R2, laporan kerja harian, dan persentase kontribusi akhir yang terkunci (*locked*).
* **Kapan Selesai:** Seluruh milestone proyek tuntas disetujui, proyek berstatus `COMPLETED`, dan persentase kontribusi akhir disahkan.
* **Domain Konsumen Output:**
  - *Domain Finance* (menerima Final Contribution % untuk pembagian porsi Gross Bounty proyek).
  - *Domain Performance & Gamification* (menerima data penyelesaian tugas untuk kredit Project XP dan rating mutu).
  - *Domain Documents* (mengonsumsi deliverable proyek terverifikasi untuk menyusun Portofolio Digital).

### 4. Hubungan dengan Domain Lain (Related Domains)
* **Domain People & Skills (Upstream Dependency):** Menyediakan direktori pengguna aktif, status intern/alumni, dan katalog profil keahlian (*Skill Matrix*).
* **Domain Workforce & Attendance (Collaborative Context):** Membedakan jam sesi kerja umum (*Daily Work*) dengan jam pengerjaan tugas proyek terstruktur (*Project Tasks*).
* **Domain Performance & Gamification (Downstream Consumer):** Menerima deliverable yang disetujui sebagai pemicu penambahan poin XP dan evaluasi kinerja supervisor.
* **Domain Finance (Downstream Consumer):** Menerima persentase kontribusi akhir sebagai pembagi nilai *Bounty Pool*.

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — Projects, Tasks & Contribution Management |
| **Business Domain** | Projects & Tasks Management (Domain 4 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Business Sign-Off |
| **Business Owner** | Dimas (Project Manager Lead) & Gita (Reviewer Lead, PT. ADT) |
| **Prepared By** | Lead Requirements Engineer & Business Analyst |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 2, 3, 5, 6.2.4, 8.5, 10) |

---

## 2. Executive Summary

Dokumen ini mendefinisikan persyaratan bisnis untuk **Projects, Tasks, Evidence & Contribution Management**. Tujuannya adalah mengelola siklus bursa proyek transparan, seleksi pelamar tanpa bias, eksekusi tugas berbasis kanban, verifikasi bukti kerja di Cloudflare R2 (*BR-024*), serta penentuan bagi hasil kontribusi tiga lapis yang adil (*BR-016*).

---

## 3. Business Context

Sebelum DCISP dibangun:
1. Penugasan proyek dilakukan secara informal melalui aplikasi perpesanan tanpa pelacakan kapasitas dan milestone.
2. Penyerahan tugas (*submission*) tidak memiliki repositori bukti terpusat; seringkali deliverable tidak dapat diverifikasi keasliannya.
3. Alokasi insentif proyek (*bounty*) ditentukan secara subjektif, memicu konflik internal dan kecemburuan antar-anggota tim.

---

## 4. Business Problem

1. **Ketidakjelasan Kontribusi Proyek:** Anggota tim yang bekerja intensif menerima imbalan yang sama dengan anggota pasif karena tidak adanya metrik kontribusi riil.
2. **Ketiadaan Repositori Bukti Deliverable:** Bukti kerja tercecer dan database seringkali terbebani penyimpanan berkas biner secara langsung.
3. **Pencampuran Jenis Pekerjaan:** Pekerjaan rutin kantor (*Daily Work*) bercampur baur dengan unit tugas proyek terstruktur (*Project Tasks*).

---

## 5. Business Objectives

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-PRJ-01** | Sengketa pembagian insentif proyek | Alokasi bounty bersandar pada rekonsiliasi 3 lapis (*Planned $\rightarrow$ Actual $\rightarrow$ Final %*) | Keadilan kompensasi & motivasi kerja tim tinggi | 100% proyek diselesaikan dengan Final Contribution % disahkan (sum = 100%) |
| **OBJ-PRJ-02** | Hilangnya artefak bukti hasil kerja | 100% penyelesaian tugas wajib melampirkan bukti fisik di R2 / tautan Git commit | Keabsahan hasil kerja terjamin & siap audit | 0% tugas berstatus *Approved* tanpa artefak bukti sah |
| **OBJ-PRJ-03** | Rekrutmen tim proyek yang tidak terukur | Papan bursa proyek dengan visibilitas transparan dan penguncian kuota atomik | Optimalisasi utilisasi keahlian peserta magang/alumni | Rasio keterisian kuota tim proyek terkelola 100% |

---

## 6. Business Scope

### 6.1 In Scope
1. **Bursa Proyek / Project Marketplace (FR-016):** Publikasi proyek dengan atribut lengkap, bounty pool, dan 3 level visibilitas (`INTERN_ONLY`, `PUBLIC`, `PRIVATE`).
2. **Alur Lamaran & Penguncian Kuota (FR-017):** Formulir lamaran, status seleksi, dan penutupan otomatis saat kuota tim terpenuhi.
3. **Skill Matching Engine Informatif (FR-018, BR-015):** Rekomendasi kesesuaian skill pelamar yang murni bersifat advisory (dilarang auto-reject).
4. **Manajemen Tim Proyek (FR-019):** Struktur peran tim (*Owner, Manager, Supervisor, Member*) dan penetapan *Planned Contribution %*.
5. **Manajemen Milestone (FR-020):** Fase pencapaian proyek dengan bobot terukur (sum = 100%).
6. **Manajemen Tugas Kanban (FR-021):** Siklus tugas: `TODO → IN_PROGRESS → IN_REVIEW → COMPLETED → BLOCKED`.
7. **Pelaporan Kerja & Penyerahan Tugas (FR-022, DATA-004):** Pengisian laporan kemajuan kerja (*what_i_did*, *problems*, *next_actions*).
8. **Pengelolaan Bukti Deliverable (FR-023, BR-024):** Validasi tautan Git commit/PR dan penyimpanan objek fisik di Cloudflare R2.
9. **Three-Layer Contribution Engine (FR-024, BR-016):** Rekonsiliasi kontribusi 3 lapis untuk pembagian hak bounty.

### 6.2 Out of Scope
1. **Hosting Repositori Git Internal:** Platform tidak menyediakan server Git sendiri; hanya mengintegrasikan verifikasi URL/webhook GitHub/GitLab.
2. **Penyimpanan Berkas Biner di Database SQL:** Dilarang menyimpan data BLOB di SQL (*BR-024*).

---

## 7. Stakeholders

| Stakeholder | Responsibility | Business Interest | Decision Authority |
|---|---|---|---|
| **Project Manager (Quest Giver)** | Membuat proyek, menetapkan kuota, memilih anggota tim, menyusun milestone | Kecepatan delivery proyek, kualitas tim, kepatuhan tenggat | Menerbitkan proyek, menerima pelamar, menetapkan alokasi Planned % |
| **Supervisor (Guild Master)** | Mengawasi eksekusi tim, mengulas deviasi kontribusi, mengesahkan Final % | Keadilan beban kerja tim dan pembinaan kompetensi | Mengesahkan *Final Contribution %* anggota tim proyek |
| **Reviewer (The Oracle)** | Memeriksa laporan kerja dan memvalidasi keaslian artefak bukti deliverable | Kepatuhan standar mutu teknis deliverable | Menyetujui (*Approve*) atau meminta revisi (*Revision Required*) tugas |
| **Intern / Alumni (Party Member)** | Mengerjakan tugas, mengunggah bukti kerja, mengisi laporan kemajuan | Kejelasan target, perolehan XP, dan kepastian pembagian bounty | Menyerahkan laporan tugas dan bukti deliverable |

---

## 8. Business Actors

| Business Role | System Role Terkait | Tindakan Utama | Batasan Akses Data |
|---|---|---|---|
| **Project Owner / PM** | Project Manager / Admin | Menerbitkan proyek, menyusun milestone & tugas, alokasi bounty | `scope = assigned_projects` |
| **Technical Reviewer** | Reviewer / Supervisor | Menelaah laporan tugas dan bukti commit/screenshot | `scope = assigned_tasks` |
| **Project Member** | Intern / Alumni | Melamar proyek, mengerjakan tugas, submit bukti deliverable | `scope = own_data + project` |

---

## 9. Current Business Process (AS-IS)

* **AS-IS:** Penugasan proyek dilakukan via WhatsApp / Trello tanpa integrasi ke jam kerja. Pelaporan progres disampaikan lisan saat rapat mingguan. Pembagian insentif dibagi rata secara informal tanpa dasar perhitungan data kerja aktual.

---

## 10. Target Business Process (TO-BE)

```text
[ PM Menerbitkan Proyek di Marketplace ] ──► [ Kandidat Melamar (Skill Match %) ]
                                                          │
                                                          ▼
                                            [ PM Menerima Pelamar & Kunci Kuota ]
                                                          │
                                                          ▼
                                            [ Penetapan Planned Contribution % ]
                                                          │
                                                          ▼
                                            [ Eksekusi Tugas & Milestone Kanban ]
                                                          │
                                                          ▼
                                            [ Submit Work Report & Evidence di R2 ]
                                                          │
                                                          ▼
                                            [ Reviewer Memvalidasi (Approved) ]
                                                          │
                                                          ▼
                                            [ Three-Layer Contribution Engine ]
                                            (Planned -> Actual -> Final % Disahkan)
                                                          │
                                                          ▼
                                            [ Penguncian Alokasi Bounty ke Finance ]
```

---

## 11. Business Process Scenarios

### Scenario 1: Rekrutmen Anggota Tim & Penguncian Kuota
* **Trigger:** PM mempublikasikan proyek baru berstatus `PUBLISHED` dengan kuota 3 orang.
* **Actor:** Intern / Alumni (Pelamar), PM (Penyetujui).
* **Main Flow:**
  1. Pelamar melihat proyek di marketplace dan mengajukan lamaran beserta surat motivasi.
  2. Sistem menampilkan skor kesesuaian keahlian (*Skill Matching %*) murni sebagai informasi (*BR-015*).
  3. PM menyetujui pelamar terpilih hingga jumlah yang diterima mencapai 3 orang.
  4. Sistem secara otomatis mengunci slot lamaran (*Party Full*) dan membentuk struktur `Project Team`.
  5. PM menetapkan kesepakatan awal: Anggota A = 40%, B = 35%, C = 25% (*Planned Contribution %*).

### Scenario 2: Penyerahan & Verifikasi Bukti Tugas
* **Trigger:** Anggota tim menyelesaikan tugas teknis (Side Quest).
* **Actor:** Anggota Tim, Reviewer.
* **Main Flow:**
  1. Anggota tim membuka tugas di portal My Day, mengisi form *Work Report* (progres 100%, deskripsi pekerjaan, kendala).
  2. Anggota tim melampirkan URL commit Git atau mengunggah screenshot bukti pengujian ke Cloudflare R2 Vault.
  3. Status tugas beralih ke `IN_REVIEW`.
  4. Reviewer memeriksa bukti artefak, menguji fungsi, dan memberikan status `APPROVED`.
  5. Sistem mencatat penyelesaian tugas ke dalam histori kontribusi riil (*Actual Contribution*).

### Scenario 3: Rekonsiliasi Kontribusi 3-Lapis & Finalisasi Bounty
* **Trigger:** Seluruh milestone proyek tuntas berstatus `COMPLETED`.
* **Actor:** Supervisor, PM.
* **Main Flow:**
  1. Sistem mengalkulasi *Layer 2 (Actual Contribution %)* berbasis kumpulan data: jumlah tugas selesai, bobot kesulitan, dan ketepatan tenggat waktu (misal: A = 45%, B = 30%, C = 25%).
  2. Supervisor membuka antarmuka perbandingan visual (*Planned vs Actual*).
  3. Supervisor mengonfirmasi nilai akhir (*Layer 3: Final Contribution %*) dan memasukkan justifikasi catatan evaluasi.
  4. Sistem memverifikasi bahwa total Final Contribution % berjumlah tepat 100,00%, mengunci data (*LOCKED*), dan menyalurkannya ke modul Finance untuk pembagian bounty.

---

## 12. Business Requirements

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis | Rasional Bisnis | Prioritas | Sumber PRD |
|---|---|---|---|---|
| **BRQ-PRJ-001** | Bisnis mewajibkan bursa proyek mendukung 3 level visibilitas (`INTERN_ONLY`, `PUBLIC`, `PRIVATE`) | Pengaturan hak akses proyek internal vs proyek publik alumni | P0 | FR-016, BR-003 |
| **BRQ-PRJ-002** | Bisnis mewajibkan kuota pelamar proyek dikunci secara atomik saat kapasitas terpenuhi | Mencegah kelebihan kapasitas anggota tim dalam suatu inisiatif proyek | P0 | FR-017 |
| **BRQ-PRJ-003** | Bisnis mewajibkan rekomendasi AI Skill Matching murni bersifat informatif bagi pengambil keputusan | Mencegah diskriminasi atau penolakan otomatis oleh sistem tanpa ulasan manusia | P2 | FR-018, BR-015 |
| **BRQ-PRJ-004** | Bisnis mewajibkan pemecahan proyek ke dalam hierarki tugas terstruktur (*Milestone & Tasks*) | Memastikan pengukuran kemajuan proyek dapat dipantau secara bertahap | P0 | FR-020, FR-021 |
| **BRQ-PRJ-005** | Bisnis mewajibkan setiap penyerahan tugas menyertakan laporan kerja dan artefak bukti sah di Cloudflare R2 | Akuntabilitas deliverable hasil kerja nyata dan optimalisasi penyimpanan | P0 | FR-022, FR-023, BR-024 |
| **BRQ-PRJ-006** | Bisnis mewajibkan pembagian bounty proyek bersandar pada rekonsiliasi kontribusi 3-lapis yang disahkan supervisor | Menjamin keadilan pembagian hasil kerja berbasis kontribusi riil | P0 | FR-024, BR-016 |

---

## 13. Business Rules

| ID Aturan | Nama Aturan Bisnis | Kondisi Bisnis | Tindakan Sistem | Pengecualian |
|---|---|---|---|---|
| **BRULE-PRJ-001** | *Work Type Partition* | Aktivitas kerja dicatat pengguna | Pisahkan strictly: Daily Work $\neq$ Project Task $\neq$ Work Report | Tidak ada pengecualian (BR-005) |
| **BRULE-PRJ-002** | *Three-Layer Balance Guard* | Persentase kontribusi diatur pada layer apa pun | Penjumlahan persentase seluruh anggota tim wajib tepat 100,00% | Proyek individu (1 orang = 100%) |
| **BRULE-PRJ-003** | *Task Completion Verification* | Anggota tim menandai tugas selesai | Dilarang status `COMPLETED` sebelum laporan dan bukti disetujui Reviewer | Tidak ada pengecualian |
| **BRULE-PRJ-004** | *R2 Object Storage Separation* | Bukti kerja berkas diunggah | Simpan berkas fisik di Cloudflare R2; simpan metadata saja di database | Tidak ada (dilarang BLOB di DB, BR-024) |
| **BRULE-PRJ-005** | *Non-Punitive AI Matching* | Algoritma menghitung skor kecocokan skill | Sajikan sebagai persentase advisory; dilarang auto-reject pelamar | Tidak ada (BR-015) |

---

## 14. Business Entities

1. **Project (DATA-003):** Entitas proyek master (title, description, owner_id, visibility, required_skills, capacity, accepted_count, deadline, bounty_pool, status).
2. **Project Application:** Berkas lamaran pelamar (project_id, user_id, cover_letter, match_percentage, status).
3. **Project Team:** Struktur anggota tim (project_role, planned_contribution_pct, actual_contribution_pct, final_contribution_pct, is_locked).
4. **Milestone:** Fase target perantara proyek (title, deadline, weight_pct, status).
5. **Task:** Unit pekerjaan spesifik (title, estimated_hours, difficulty_weight, priority, deadline, status).
6. **Work Report & Submission (DATA-004):** Laporan deliverable (progress_percentage, what_i_did, evidence_type, evidence_url_or_key, problems, next_actions, status).
7. **Evidence:** Artefak bukti deliverable (submission_id, file_metadata_id, evidence_type, external_url).

---

## 15. Data Flow

```text
[ Project Manager ] ──► [ Publikasi Proyek & Bounty Pool ]
                                   │
[ Pelamar ] ──────────► [ Pengajuan Lamaran ] ──► [ Penguncian Tim (Planned %) ]
                                                               │
                                                               ▼
[ Pelaksana ] ────────► [ Eksekusi Task ] ──► [ Submit Report & Evidence R2 ]
                                                               │
                                                               ▼
[ Reviewer ] ─────────► [ Verifikasi Evidence (Approved) ]
                                                               │
                                                               ▼
[ Supervisor ] ───────► [ Pengesahan Final Contribution % (sum=100%) ]
                                                               │
                                                               ▼
                                                  [ Alirkan ke Modul Finance ]
```

---

## 16. Status & Lifecycle

```text
1. PROJECT LIFECYCLE:
   [ DRAFT ] ──► [ PUBLISHED ] ──► [ IN_PROGRESS ] ──► [ COMPLETED ]
        │             │                 │
        └──► [ CANCELLED ] ◄────────────┘

2. TASK LIFECYCLE:
   [ TODO ] ──► [ IN_PROGRESS ] ──► [ IN_REVIEW ] ──► [ COMPLETED ]
                      ▲                   │
                      └── [ REVISION ] ◄──┘
```

---

## 17. Approval & Decision Flow

* **Penerimaan Pelamar:** Requester: Pelamar $\rightarrow$ Approver: Project Manager $\rightarrow$ Slot tim terkunci.
* **Kelayakan Tugas:** Requester: Pelaksana $\rightarrow$ Approver: Reviewer $\rightarrow$ Status `APPROVED`.
* **Pengesahan Kontribusi Akhir:** Requester: Sistem $\rightarrow$ Approver: Supervisor $\rightarrow$ Status kontribusi `LOCKED`.

---

## 18. Exception & Failure Handling

* **Tugas Melewati Deadline:** Sistem menandai flag *Overdue*; memicu pengingat otomatis dan penurunan skor ketepatan waktu di Layer 2 (Actual Contribution).
* **Unggah Bukti Gagal:** Presigned URL ber-TTL 15 menit; jika gagal jaringan, klien melakukan retry upload langsung ke Cloudflare R2 tanpa membebani server backend.

---

## 19. Business Reporting

1. **Laporan Kecepatan Proyek (*Milestone Velocity Report*):** Persentase milestone selesai tepat waktu per proyek dan beban tugas tim.
2. **Laporan Audit Kontribusi Proyek:** Rekam perbandingan Planned vs Actual vs Final Contribution % untuk evaluasi transparansi pembagian bounty.

---

## 20. KPI & Success Metrics

* **Milestone On-Time Delivery Rate (MET-03):** Persentase milestone yang diserahkan sebelum/tepat deadline (`Target: TBD`).
* **Task Cycle Time (MTR-PRJ-02):** Rata-rata durasi tugas dari status *In Progress* hingga *Approved* (`Target: TBD`).

---

## 21. Business Integration

* **Ke Domain Finance:** Mengirimkan *Final Contribution %* untuk dievaluasi oleh *Deduction & Tax Engine* dan pembagian porsi *Gross Bounty*.
* **Ke Domain Performance:** Mengirimkan penyelesaian tugas untuk kredit *Project XP* dan rating mutu.

---

## 22. Business Dependencies

* **Upstream:** People & Skills Module (data user & skill).
* **Downstream:** Finance Module (bounty distribution).

---

## 23. Business Constraints

* **Pemisahan Storage R2 (BR-024):** Database dilarang menyimpan data biner; seluruh bukti tugas disimpan di Cloudflare R2.
* **Total Kontribusi Wajib 100% (BR-016):** Total persentase kontribusi seluruh anggota tim wajib tepat 100,00%.

---

## 24. Business Assumptions

* Peserta memiliki akses repositori Git publik/organisasi (GitHub/GitLab) untuk menautkan bukti commit pengerjaan kode.

---

## 25. Business Gaps

* Penentuan batasan format berkas bukti tambahan dan opsi integrasi webhook Git otomatis berstatus `[TBD]`.

---

## 26. Ambiguities

* *Nomenklatur visibilitas disepakati baku: `INTERN_ONLY`, `PUBLIC`, `PRIVATE` (OQ-012).*

---

## 27. Conflicting Requirements

* *Tidak ada konflik requirement.*

---

## 28. Traceability Matrix

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Entitas Terkait | Acceptance Criteria |
|---|---|---|---|---|
| FR-016 (Marketplace) | BRQ-PRJ-001 | BRULE-PRJ-001 | `Project` | AC-PRJ-001 |
| FR-017 (Application & Quota) | BRQ-PRJ-002 | — | `Project Application` | AC-PRJ-002 |
| FR-018 (Skill Matching) | BRQ-PRJ-003 | BRULE-PRJ-005 | `Skill`, `Project` | AC-PRJ-003 |
| FR-019 (Team Mgmt) | BRQ-PRJ-004 | BRULE-PRJ-002 | `Project Team` | AC-PRJ-004 |
| FR-021 (Task Mgmt) | BRQ-PRJ-004 | BRULE-PRJ-003 | `Task` | AC-PRJ-005 |
| FR-022 (Work Report) | BRQ-PRJ-005 | BRULE-PRJ-001 | `Work Report & Submission` | AC-PRJ-006 |
| FR-023 (Evidence Vault) | BRQ-PRJ-005 | BRULE-PRJ-004 | `Evidence`, `File Metadata` | AC-PRJ-007 |
| FR-024 (3-Layer Contribution)| BRQ-PRJ-006 | BRULE-PRJ-002 | `Project Team` | AC-PRJ-008 |

---

## 29. Acceptance Criteria

### AC-PRJ-001: Filter Visibilitas Proyek Marketplace
* **Given:** Terdapat proyek berstatus `PUBLISHED` dengan visibilitas `INTERN_ONLY`.
* **When:** Pengguna berstatus Alumni membuka bursa proyek.
* **Then:** Proyek tersebut tidak muncul dalam daftar bursa alumni dan akses langsung via URL ditolak (HTTP 403).

### AC-PRJ-002: Penguncian Kuota Tim Otomatis
* **Given:** Proyek dengan sisa kuota 1 orang (`capacity = 4`, `accepted_count = 3`).
* **When:** PM menyetujui 1 pelamar baru menjadi `ACCEPTED`.
* **Then:** Kuota penuh (`accepted_count = 4`), tombol lamaran terkunci otomatis, dan pelamar dimasukkan ke `Project Team`.

### AC-PRJ-003: Rekonsiliasi Kontribusi Tiga Lapis
* **Given:** Proyek selesai dengan data Layer 2 (Actual) Anggota A = 45% dan Anggota B = 55%.
* **When:** Supervisor mengesahkan proporsi tersebut sebagai Final Contribution %.
* **Then:** Sistem memvalidasi total sum = 100,00%, mengunci data persentase, dan mengalirkannya ke mesin pembagian bounty finansial.

---

## 30. Open Questions

* `OQ-PRJ-01`: Standardisasi integrasi webhook Git otomatis vs input URL commit manual.

---

## 31. BRD Completion Checklist

- [x] Seluruh kebutuhan domain Projects, Tasks, Evidence, dan Contribution terdefinisi lengkap.
- [x] Aturan Three-Layer Contribution (BR-016) dan R2 Storage Separation (BR-024) terpetakan eksplisit.

---
*Dokumen ini disimpan permanen di `brd/03-BRD-PROJECTS-TASKS.md`.*
