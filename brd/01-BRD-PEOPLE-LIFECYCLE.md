# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: People, Batches & Alumni Lifecycle Management
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**People, Batches & Alumni Lifecycle Management (Manajemen Talenta, Siklus Hidup Peserta, Kelompok Angkatan, dan Komunitas Alumni)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain ini mengatur entitas manusia dan siklus perjalanannya di dalam ekosistem PT. Aplikasi Dagang Teknologi:
- Proses dimulai sejak seorang kandidat mendaftar (*Applicant*), diverifikasi dan dikelompokkan ke dalam kohort angkatan (*Batch*), menjalani masa magang aktif (*Intern*), hingga lulus (*Graduated*) dan bertransisi menjadi aset talenta tetap (*Alumni*).
- Retensi akun alumni permanen (*BR-003*) merupakan kelanjutan langsung dari siklus kelulusan peserta magang, yang mempertahankan identitas, riwayat portofolio, dan hak mengerjakan proyek publik perusahaan.
- Pengelolaan institusi pendidikan mitra (kampus/sekolah) dan katalog keahlian (*Skill Matrix*) adalah fondasi profil talenta yang melekat langsung pada siklus hidup individu.

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Pendaftaran kandidat magang baru, pembentukan kohort batch baru oleh HR, atau transisi kelulusan peserta magang.
* **Input Bisnis:** Formulir pendaftaran pelamar, data identitas (NIM/NISN), data institusi pendidikan, profil keahlian (*skills*), parameter tanggal dan kuota batch, serta persetujuan kelulusan dari HR.
* **Proses Utama Bisnis:**
  1. Seleksi dan transisi status siklus hidup peserta: `APPLICANT → ONBOARDING → ACTIVE → ON_LEAVE → SUSPENDED → GRADUATED → TERMINATED`.
  2. Pengelompokan peserta ke dalam kohort *Batch* dan inisialisasi rekening kas perpisahan angkatan (*Batch Fund*).
  3. Pengelolaan repositori institusi pendidikan mitra (kampus/sekolah) dan perjanjian kemitraan.
  4. Pengelolaan katalog keahlian (*Skill Matrix*) dan profil kompetensi individu.
  5. Retensi akun alumni secara permanen dengan hak akses pasar proyek publik.
* **Keputusan Bisnis yang Dibuat:**
  - Penerimaan/penolakan kandidat pelamar magang ke batch tertentu.
  - Pengesahan kelulusan resmi peserta magang menjadi alumni.
  - Penetapan status disipliner peserta (Suspend / Terminate).
* **Output Bisnis:** Akun pengguna aktif, entitas profil intern/alumni terverifikasi, kohort batch resmi, dan direktori talenta perusahaan.
* **Kapan Selesai:** Akun bertransisi ke status permanen *Alumni* (retensi seumur hidup tanpa batas waktu).
* **Domain Konsumen Output:**
  - *Domain Workforce & Attendance* (mengonsumsi status intern aktif dan penugasan batch untuk presensi).
  - *Domain Projects & Tasks* (mengonsumsi profil skill dan status alumni/intern untuk kelayakan melamar proyek).
  - *Domain Performance & Gamification* (mengonsumsi data batch untuk evaluasi Top Performer).
  - *Domain Finance* (mengonsumsi entitas batch untuk inisialisasi Batch Fund).

### 4. Hubungan dengan Domain Lain (Related Domains)
* **Domain Identity & RBAC (Upstream Dependency):** Menyediakan pembuatan kredensial user akun dan penugasan peran sistem (`INTERN`, `ALUMNI`, `HR_ADMIN`).
* **Domain Workforce & Attendance (Downstream Consumer):** Menggunakan status `ACTIVE` peserta sebagai syarat presensi dan sesi kerja.
* **Domain Projects & Tasks (Downstream Consumer):** Menggunakan profil skill untuk bursa proyek.
* **Domain Performance & Gamification (Downstream Consumer):** Menggunakan batasan batch untuk menentukan juara angkatan.

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — People, Batches & Alumni Lifecycle |
| **Business Domain** | People & Lifecycle Management (Domain 2 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Sign-Off |
| **Business Owner** | HR / Internship Coordinator & Program Lead (PT. Aplikasi Dagang Teknologi) |
| **Prepared By** | Lead Requirements Engineer & Business Analyst |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 2, 3, 5, 6.2.1, 8.3, 10) |

---

## 2. Executive Summary

Dokumen ini mendefinisikan kebutuhan bisnis untuk pengelolaan siklus hidup talenta di PT. Aplikasi Dagang Teknologi, mulai dari penerimaan peserta magang, pengelompokan ke dalam kohort batch, manajemen relasi kampus, pemetaan matriks keahlian, hingga transisi kelulusan menjadi alumni yang mempertahankan akun permanen (*BR-003*). Domain ini memastikan rekam jejak talenta terdokumentasi rapi dan hubungan kerja sama profesional berlanjut pasca-magang.

---

## 3. Business Context

Sebelum platform DCISP dibangun:
1. Data peserta magang tersebar di spreadsheet terpisah dan tidak terhubung ke riwayat proyek nyata.
2. Ketika program magang berakhir, akun peserta dinonaktifkan (*dead accounts*), memutuskan hubungan dengan talenta berbakat yang berpotensi mengerjakan proyek lepas publik perusahaan.
3. Tidak adanya standarisasi data institusi pendidikan asal dan skill matrix peserta yang terverifikasi.

---

## 4. Business Problem

1. **Kehilangan Aset Talenta Pasca-Magang:** Alumni kehilangan akses platform sehingga perusahaan kesulitan merekrut kembali alumni terbaik untuk inisiatif proyek baru.
2. **Fragmentasi Manajemen Kohort:** Ketiadaan isolasi data antar-angkatan menyulitkan evaluasi performa per batch dan tata kelola dana perpisahan angkatan.
3. **Ketiadaan Repositori Skill Terstruktur:** Profil keahlian peserta tidak terpetakan dengan standar kompetensi industri yang terukur.

---

## 5. Business Objectives

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-PEO-01** | Putusnya hubungan kerja sama dengan lulusan magang | 100% peserta yang lulus mempertahankan akun permanen berstatus Alumni | Pipeline talenta lepas siap pakai untuk proyek publik perusahaan | Rasio retensi akun alumni aktif $\ge 90$ hari pasca-lulus |
| **OBJ-PEO-02** | Ketidakteraturan administrasi kohort angkatan | Seluruh peserta magang terkelompokkan ke dalam entitas Batch terisolasi | Evaluasi kohort terstruktur & pengelolaan dana batch rapi | 100% peserta magang aktif terikat ke tepat satu batch |
| **OBJ-PEO-03** | Profil keahlian tidak terstandarisasi | Tersedianya katalog Skill Matrix berjenjang yang mencatat riwayat pemanfaatan skill | Kemudahan pencocokan talenta terhadap kebutuhan proyek | 100% pelamar proyek memiliki profil skill terverifikasi |

---

## 6. Business Scope

### 6.1 In Scope
1. **Manajemen Siklus Hidup Peserta Magang (FR-002):** Pendaftaran kandidat, onboarding, pengaktifan akun, penangguhan, dan kelulusan.
2. **Retensi Akun Alumni Permanen (FR-003, BR-003):** Pemeliharaan akun seumur hidup, hak melamar proyek pasar publik, dan pemisahan standing ranking.
3. **Manajemen Kohort Batch (FR-004):** Pembuatan batch, penetapan kuota, rentang tanggal kalender, dan inisialisasi entitas kas angkatan (*Batch Fund*).
4. **Manajemen Institusi Kampus Mitra (FR-005):** Master data universitas/sekolah, data kontak pembimbing, dan rekapitulasi peserta per kampus.
5. **Katalog Skill Matrix & Profil Keahlian (FR-006):** Repositori master keahlian teknis/non-teknis dengan level kemahiran (Pemula, Menengah, Mahir).

### 6.2 Out of Scope
1. **Perekrutan Karyawan Tetap Korporat Skala Penuh:** Modul ini khusus untuk peserta magang dan talenta proyek alumni, bukan portal karir umum publik bebas.
2. **Kurikulum Akademik Internal Kampus:** Platform tidak mengelola sistem penilaian SKS internal universitas mitra.

---

## 7. Stakeholders

| Stakeholder | Responsibility | Business Interest | Decision Authority |
|---|---|---|---|
| **HR / Internship Admin (Game Master)** | Mengelola batch, memverifikasi pelamar, mengesahkan kelulusan, membina relasi kampus | Kerapihan administrasi, kepatuhan masa magang, dan pipeline talenta | Otorisasi penerimaan pelamar, penetapan batch, dan kelulusan |
| **Intern (Adventurer)** | Mengisi profil diri, memilih keahlian, menjalani masa magang aktif | Kejelasan masa program, pengakuan kompetensi, dan sertifikasi | Mengajukan data profil dan pilihan keahlian pribadi |
| **Alumni (Legendary Hero)** | Memperbarui profil portofolio, menelusuri proyek publik perusahaan | Peluang kerja proyek berkelanjutan dan portofolio profesional terverifikasi | Mengelola profil publik pribadi dan mengajukan lamaran proyek publik |
| **Institusi Kampus Mitra** | Menjalin kerja sama magang, memantau kemajuan mahasiswa | Penyaluran magang terstruktur dan laporan capaian mahasiswa | Penandatanganan MoU kemitraan |

---

## 8. Business Actors

| Business Role | System Role Terkait | Tindakan Utama | Batasan Akses Data |
|---|---|---|---|
| **Talent Manager** | HR Admin / Super Admin | Buat batch, terima peserta, luluskan peserta, kelola institusi | `scope = workforce_and_people` |
| **Intern Candidate** | Applicant / Intern | Mengisi data diri, nomor induk (NIM/NISN), profil skill | `scope = own_data` |
| **Alumni Member** | Alumni | Mengakses direktori proyek publik, mengelola showcase portofolio | `scope = own_data + public_projects` *(BR-003)* |

---

## 9. Current Business Process (AS-IS)

* **AS-IS:** Pendaftaran magang dilakukan via Google Form / email. Data batch dicatat manual di Excel. Saat magang selesai, surat keterangan dibuat manual dan akses akun email/sistem dinonaktifkan sehingga hubungan dengan alumni terputus.

---

## 10. Target Business Process (TO-BE)

```text
[ Applicant Mendaftar & Mengisi Profil Skill ]
                 │
                 ▼
[ HR Meninjau Berkas & Menetapkan ke Batch Tertentu ]
                 │
                 ▼ (Status: ACTIVE, Akun Intern Aktif)
[ Intern Menjalani Program Magang, Tugas & Proyek ]
                 │
                 ▼ (Masa Program Berakhir & Evaluasi Final)
[ HR Mengesahkan Kelulusan (Status: GRADUATED) ]
                 │
                 ▼ (Otomatis Bertransisi Tanpa Hapus Akun)
[ Akun Permanen Alumni Aktif (Peran: ALUMNI) ]
                 │
                 ├──► Akses ke Bursa Proyek Publik (Public Quests)
                 └──► Publikasi Hero's Chronicle (Digital Portfolio)
```

---

## 11. Business Process Scenarios

### Scenario 1: Penerimaan & Aktivasi Peserta Magang Baru
* **Trigger:** HR Admin menyetujui penerimaan kandidat ke dalam Batch 2026-01.
* **Precondition:** Batch 2026-01 berstatus `ACTIVE` dan masih memiliki sisa kuota.
* **Actor:** HR Admin, Intern.
* **Main Flow:**
  1. HR Admin memilih profil Applicant yang lolos seleksi.
  2. HR Admin mengaitkan pelamar ke institusi pendidikan asal dan menetapkan supervisor pembimbing.
  3. HR Admin mengubah status menjadi `ACTIVE`.
  4. Sistem menginisialisasi profil magang, dompet personal, profil XP magang, dan menerbitkan kredensial login.
* **Output:** Akun intern aktif, kuota batch terisi +1, jadwal kerja default diterapkan.

### Scenario 2: Transisi Kelulusan Menjadi Alumni Permanen
* **Trigger:** Peserta magang menyelesaikan seluruh kewajiban program dan masa kalender batch berakhir.
* **Precondition:** Evaluasi kinerja akhir disahkan oleh supervisor.
* **Actor:** HR Admin, Intern (Lulusan).
* **Main Flow:**
  1. HR Admin memproses kelulusan intern $\rightarrow$ Status berubah menjadi `GRADUATED`.
  2. Sistem secara otomatis membuat profil `Alumni` terhubung ke User ID yang sama tanpa menghapus data historis (*BR-003*).
  3. Hak akses peran pengguna dialihkan ke `ALUMNI`.
  4. Sistem membuka akses bursa proyek publik dan mengaktifkan jalur *Alumni XP / Veteran XP*.
* **Output:** Akun bertransisi ke peran Alumni secara permanen; rekam jejak magang terkunci aman.

---

## 12. Business Requirements

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis | Rasional Bisnis | Prioritas | Sumber PRD |
|---|---|---|---|---|
| **BRQ-PEO-001** | Bisnis mewajibkan transisi status peserta magang dikelola melalui mesin status formal | Akuntabilitas tahapan penerimaan hingga kelulusan peserta | P0 | FR-002 |
| **BRQ-PEO-002** | Bisnis mewajibkan akun peserta magang yang lulus dipertahankan secara permanen tanpa batas waktu (*Alumni Retention*) | Mempertahankan jejaring talenta dan kontinuitas kerja sama proyek publik | P0 | FR-003, BR-003 |
| **BRQ-PEO-003** | Bisnis mewajibkan seluruh aktivitas proyek alumni terisolasi dari standing ranking peserta magang aktif | Mencegah ketidakadilan persaingan antara lulusan senior dengan peserta magang baru | P0 | BR-004, Section 8.3 |
| **BRQ-PEO-004** | Bisnis mewajibkan pengelompokan peserta ke dalam kohort batch dengan alokasi rekening kas bersama (*Batch Fund*) otomatis | Menjamin isolasi data angkatan dan transparansi dana perpisahan bersama | P0 | FR-004, BR-021 |
| **BRQ-PEO-005** | Bisnis mewajibkan tersedianya repositori institusi pendidikan mitra | Mempermudah pelaporan kemitraan kampus dan rekapitulasi data mahasiswa | P1 | FR-005 |
| **BRQ-PEO-006** | Bisnis mewajibkan tersedianya katalog Skill Matrix terstandarisasi untuk pencocokan tugas proyek | Memastikan transparansi keahlian peserta dan efektivitas pembentukan tim kerja | P1 | FR-006, BR-015 |

---

## 13. Business Rules

| ID Aturan | Nama Aturan Bisnis | Kondisi Bisnis | Tindakan Sistem | Pengecualian |
|---|---|---|---|---|
| **BRULE-PEO-001** | *Permanent Account Retention* | Intern lulus status `GRADUATED` | Pertahankan akun permanen; ubah peran ke `ALUMNI`; jangan hapus data | Pelanggaran hukum berat (status `TERMINATED`) |
| **BRULE-PEO-002** | *Alumni XP Scheme Isolation* | Alumni menyelesaikan tugas proyek publik | Kreditkan poin ke skema `Alumni Contribution / Veteran XP`; dilarang mengubah `Internship XP` | Tidak ada pengecualian (BR-004) |
| **BRULE-PEO-003** | *Singular Batch Assignment* | Peserta magang didaftarkan ke sistem | Wajib terikat pada tepat 1 kohort batch aktif selama masa magang | Alumni tidak terikat batch aktif lagi |
| **BRULE-PEO-004** | *Batch Fund Initialization* | Batch baru dibuat oleh HR | Otomatis buat entitas `Batch Fund` dengan saldo awal Rp0 | Tidak ada pengecualian |
| **BRULE-PEO-005** | *Institution Deletion Guard* | HR mencoba menghapus data institusi kampus | Tolak penghapusan jika masih ada data peserta/alumni yang terhubung | Institusi tanpa riwayat peserta |

---

## 14. Business Entities

1. **User:** Identitas otentikasi dasar (email, password hash, status akun).
2. **Intern:** Profil peserta magang aktif (NIM/NISN, tanggal mulai/selesai, batch_id, supervisor_id, rank_id, status).
3. **Alumni:** Profil lulusan permanen (graduation_date, alumni_xp, certificate_id, public_profile_flag).
4. **Batch:** Entitas kohort angkatan (batch_code, name, start_date, end_date, quota, status).
5. **Institution:** Data sekolah/universitas mitra (nama, alamat, koordinator, kontak).
6. **Skill:** Master katalog keahlian (nama, kategori, deskripsi).

---

## 15. Data Flow

```text
[ Calon Peserta / Kampus Mitra ] ──► [ Registrasi Profil & Skill Matrix ]
                                                  │
                                                  ▼
                                      [ Penugasan ke Batch oleh HR ]
                                                  │
                                                  ▼
                                     [ Sesi Magang Aktif (Intern) ]
                                                  │
                                                  ▼
                                  [ Pengesahan Kelulusan oleh HR ]
                                                  │
                                                  ▼
                                    [ Akun Permanen Alumni Aktif ]
                                                  │
                                                  ▼
                                    [ Bursa Proyek Publik & Portofolio ]
```

---

## 16. Status & Lifecycle

```text
INTERN LIFECYCLE:
[ APPLICANT ] ──(Diterima HR)──► [ ONBOARDING ] ──(Aktivasi)──► [ ACTIVE ]
       │                                                            │
       └──(Ditolak)──► [ REJECTED ]                                 ├──(Kelulusan)──► [ GRADUATED ] ──► (Alumni)
                                                                    ├──(Cuti)──────► [ ON_LEAVE ]
                                                                    └──(Sanksi)────► [ TERMINATED ]
```

---

## 17. Approval & Decision Flow

* **Penerimaan Pelamar:** Requester: Applicant $\rightarrow$ Approver: HR Admin $\rightarrow$ Action: Tetapkan ke Batch dan ubah status ke `ACTIVE`.
* **Pengesahan Kelulusan:** Requester: Supervisor (Evaluasi Selesai) $\rightarrow$ Approver: HR Admin / Program Lead $\rightarrow$ Action: Ubah status ke `GRADUATED` dan buat record `Alumni`.

---

## 18. Exception & Failure Handling

* **Kuota Batch Penuh:** Sistem memblokir pendaftaran tambahan saat `accepted_count = quota`; pelamar dialihkan ke batch berikutnya.
* **Keluar Dini (Drop Out):** HR mengubah status ke `TERMINATED`; akun dibekukan tanpa hak akses alumni.

---

## 19. Business Reporting

1. **Laporan Demografi & Kemitraan Kampus:** Rekapitulasi jumlah peserta per universitas mitra dan performa rata-rata lulusan per institusi.
2. **Laporan Retensi & Keaktifan Alumni:** Rasio alumni yang aktif mengambil proyek publik pasca-kelulusan.

---

## 20. KPI & Success Metrics

* **Alumni Engagement Rate (MET-05):** Persentase alumni aktif yang melamar/mengerjakan proyek publik (`Target: TBD`).
* **Cohort Completion Rate (Derived KPI):** Persentase peserta magang yang berhasil lulus tepat waktu per batch.

---

## 21. Business Integration

* **Ke Domain Workforce:** Mengirimkan data peserta `ACTIVE` untuk validasi presensi dan jadwal.
* **Ke Domain Projects:** Mengirimkan profil skill untuk verifikasi lamaran proyek bursa.

---

## 22. Business Dependencies

* **Identity Module:** Bergantung mutlak pada user auth dan RBAC engine.

---

## 23. Business Constraints

* **Batasan Retensi Permanen (BR-003):** Akun alumni dilarang dihapus atau dinonaktifkan permanen saat program selesai.

---

## 24. Business Assumptions

* Peserta magang memiliki perangkat komputer/laptop sendiri untuk mengakses platform secara mandiri.

---

## 25. Business Gaps

* Detail field spesifik institusi (MoU, kontak dosen) berstatus `[TBD]`.

---

## 26. Ambiguities

* *Tidak ada ambiguitas kritis pada domain ini.*

---

## 27. Conflicting Requirements

* *Tidak ada konflik requirement.*

---

## 28. Traceability Matrix

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Entitas Terkait | Acceptance Criteria |
|---|---|---|---|---|
| FR-002 (Intern Lifecycle) | BRQ-PEO-001 | BRULE-PEO-001 | `Intern`, `User` | AC-PEO-001 |
| FR-003 (Alumni Retention) | BRQ-PEO-002, BRQ-PEO-003 | BRULE-PEO-001, BRULE-PEO-002 | `Alumni`, `User` | AC-PEO-002 |
| FR-004 (Batch Management) | BRQ-PEO-004 | BRULE-PEO-003, BRULE-PEO-004 | `Batch`, `Batch Fund` | AC-PEO-003 |
| FR-005 (Institution) | BRQ-PEO-005 | BRULE-PEO-005 | `Institution` | AC-PEO-004 |
| FR-006 (Skill Matrix) | BRQ-PEO-006 | — | `Skill`, `User` | AC-PEO-005 |

---

## 29. Acceptance Criteria

### AC-PEO-001: Transisi Status Kelulusan ke Alumni
* **Given:** Intern berstatus `ACTIVE` telah menyelesaikan evaluasi akhir batch.
* **When:** HR Admin menyetujui kelulusan resmi intern.
* **Then:** Status intern berubah menjadi `GRADUATED`, entitas Alumni dibuat secara otomatis, akun tetap dapat login, dan hak akses dialihkan ke peran Alumni.

### AC-PEO-002: Isolasi Poin XP Alumni
* **Given:** Pengguna berstatus Alumni mengerjakan dan menyelesaikan tugas proyek publik.
* **When:** Reviewer menyetujui deliverable tugas tersebut.
* **Then:** Poin pengalaman dicatat pada skema `Alumni Contribution / Veteran XP` tanpa mengubah standing ranking peserta magang aktif kohort berjalan.

### AC-PEO-003: Inisialisasi Otomatis Batch Fund
* **Given:** HR Admin membuat entitas Batch baru (misal: Batch 2026-02).
* **When:** Batch disimpan dan diaktifkan.
* **Then:** Sistem secara otomatis membuat entitas `Batch Fund` terhubung dengan saldo awal Rp0.

---

## 30. Open Questions

* *Tidak ada open question baru di luar katalog master PRD.*

---

## 31. BRD Completion Checklist

- [x] Seluruh kebutuhan People, Batches, dan Alumni terdefinisi lengkap.
- [x] Aturan retensi akun permanen (BR-003) dan isolasi XP alumni (BR-004) terpetakan eksplisit.
- [x] Traceability matrix dan kriteria keberterimaan (AC) tersedia.

---
*Dokumen ini disimpan permanen di `brd/01-BRD-PEOPLE-LIFECYCLE.md`.*
