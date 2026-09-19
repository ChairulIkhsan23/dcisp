# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: Workforce & Attendance Management (Manajemen Kehadiran & Sesi Kerja)
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**Workforce & Attendance Management (Manajemen Kehadiran, Sesi Kerja, dan Integritas Waktu Kerja)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain *Workforce & Attendance Management* merepresentasikan satu siklus operasional utuh yang mengatur keberadaan fisik, alokasi jam kerja produktif, kepatuhan jadwal, dan penegakan disiplin harian peserta kerja/magang di lingkungan PT. Aplikasi Dagang Teknologi.

Domain ini tidak dapat dipisahkan secara logis karena:
- Pencatatan kehadiran fisik (*attendance presence*) merupakan prasyarat (*precondition*) mutlak untuk memulai sesi kerja (*work session*).
- Status istirahat (*break state*) dan kerja lembur (*overtime*) merupakan sub-transisi dari sesi kerja aktif yang terikat langsung pada jadwal kerja acuan (*work schedule*).
- Anomali ketidakhadiran, istirahat tidak sah, dan keterlambatan bermuara pada satu mekanisme rekonsiliasi formal (*attendance correction & exception management*).
- Output dari domain ini adalah data jam kerja terverifikasi dan log kehadiran yang menjadi dasar masukan bagi domain lain (Gamifikasi XP, Evaluasi Performa, dan Kompensasi Lembur).

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Kedatangan peserta di tempat kerja atau kebutuhan pengajuan jadwal kerja/izin/lembur.
* **Input Bisnis:** Kredensial identitas (kartu fisik NFC / kode QR dinamis), penekanan tombol kendali sesi kerja di portal, formulir permohonan lembur, permohonan cuti, formulir koreksi presensi.
* **Proses Utama Bisnis:**
  1. Validasi keberadaan fisik terhadap jadwal kerja resmi, toleransi keterlambatan (*grace period*), dan kalender libur.
  2. Pengukuran dan pemisahan 5 dimensi waktu kerja (*Attendance, Work Session, Active, Break, Overtime*).
  3. Pemantauan integritas fokus sesi kerja berbasis browser non-invasif (*Privacy-First*).
  4. Peninjauan dan persetujuan pengajuan lembur, izin cuti, dan koreksi data kehadiran oleh pembimbing (*Supervisor*).
* **Keputusan Bisnis yang Dibuat:**
  - Penetapan status ketepatan waktu: `ON_TIME` vs `LATE` (Tingkat 1, 2, atau 3).
  - Penetapan status anomali istirahat: `Early Break Anomaly` vs `Unauthorized Break Violation`.
  - Keputusan persetujuan/penolakan terhadap lembur, cuti, dan koreksi kehadiran.
* **Output Bisnis:** Rekaman log kehadiran append-only yang sah, ringkasan jam kerja produktif harian, data anomali kehadiran terverifikasi, dan mutasi saldo cuti.
* **Kapan Selesai:** Sesi kerja ditutup (*End Work*), laporan harian diserahkan, dan check-out fisik berhasil tercatat pada terminal gerbang.
* **Domain Konsumen Output:** 
  - *Domain Performance & Gamification* (menerima trigger perolehan XP dan penalti keterlambatan).
  - *Domain Projects & Tasks* (menerima alokasi waktu kerja aktif untuk penugasan tugas).
  - *Domain Finance* (menerima data jam lembur aktual yang disetujui untuk kompensasi).

### 4. Hubungan dengan Domain Lain (Related Domains)
* **Domain Identity & RBAC (Upstream Dependency):** Menyediakan identitas terotentikasi, relasi peran pengguna, dan izin akses scanner operator.
* **Domain People & Batches (Upstream Dependency):** Menyediakan status aktif peserta magang (*Intern*), pengelompokan kohort (*Batch*), dan penugasan *Supervisor*.
* **Domain Performance & Gamification (Downstream Consumer):** Mengonsumsi data keterlambatan, istirahat berlebih, dan mangkir untuk dievaluasi oleh *XP Rules Engine*.
* **Domain Finance (Downstream Consumer):** Mengonsumsi durasi lembur aktual untuk kalkulasi bonus lembur.

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — Workforce & Attendance Management |
| **Business Domain** | Workforce & Attendance Management (Domain 3 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Business Sign-Off |
| **Business Owner** | Reihan (Program Lead / Operations Owner, PT. Aplikasi Dagang Teknologi) |
| **Prepared By** | Lead Requirements Engineer & Business Analyst |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 2, 3, 5, 6, 8.4, 10, 11.2, 12.3) |

---

## 2. Executive Summary

Dokumen Kebutuhan Bisnis (BRD) ini merinci persyaratan operasional, aturan bisnis, dan tata kelola untuk domain **Workforce & Attendance Management** pada platform Dagang Creative Intern Solutions Program (DCISP). 

Tujuan utama domain ini adalah membangun sistem pencatatan dan pengelolaan waktu kerja magang yang transparan, akuntabel, dan berintegritas tinggi dengan menerapkan pemisahan presisi 5 dimensi waktu kerja (*5-Way Time Differentiation*), pemantauan sesi kerja yang menjaga privasi pekerja (*Privacy-First Tracking*), penyerapan kehadiran terpusat multi-metode (NFC dan QR Code dinamis), serta alur pengecualian administratif (lembur, cuti, dan koreksi presensi) yang memelihara jejak audit permanen (*append-only*).

---

## 3. Business Context

PT. Aplikasi Dagang Teknologi mengelola program magang berbasis proyek (*project-based internship*). Sebelum sistem DCISP dirancang, pengelolaan kehadiran dan jam kerja menghadapi tantangan struktural:
1. Kehadiran dicatat secara terpisah dari jam kerja nyata, sehingga tidak ada visibilitas apakah peserta benar-benar bekerja produktif atau sekadar hadir di kantor.
2. Pengajuan kerja lembur dan izin cuti seringkali informal tanpa rekonsiliasi terhadap jam kehadiran aktual di jadwal kerja.
3. Koreksi kesalahan presensi dilakukan secara manual dengan mengubah database langsung, merusak jejak audit kepatuhan (*audit trail*).
4. Aturan jam kerja, toleransi keterlambatan, dan sanksi penalti kehadiran diatur secara kaku tanpa fleksibilitas kebijakan organisasi.

DCISP memposisikan domain *Workforce & Attendance* sebagai fondasi integritas disiplin kerja yang menghubungkan kehadiran fisik harian dengan pencapaian performa profesional peserta magang.

---

## 4. Business Problem

1. **Ketiadaan Batas Tegas Jam Kerja:** Tidak adanya pemisahan antara waktu kehadiran di kantor (*Attendance Time*), durasi sesi kerja resmi (*Work Session Time*), jam interaksi aktif (*Active Time*), jeda istirahat (*Break Time*), dan jam lembur (*Overtime*).
2. **Manipulasi & Sengketa Durasi Kerja:** Rentannya pencatatan jam lembur yang diajukan secara mentah tanpa verifikasi terhadap jam kerja aktual yang disetujui dalam jadwal.
3. **Kerentanan Pengawasan Invasif:** Risiko pemantauan kerja yang melanggar privasi individu (seperti penggunaan keylogger/screenshot diam-diam) yang merusak kepercayaan peserta kerja.
4. **Kehilangan Jejak Audit pada Koreksi:** Praktik penimpaan data kehadiran saat terjadi kesalahan mesin pembaca, menghilangkan catatan historis asli.

---

## 5. Business Objectives

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-WF-01** | Manipulasi jam kehadiran & kerancuan waktu kerja | Seluruh waktu kerja terdistribusi ke dalam 5 dimensi independen secara presisi | Keadilan evaluasi beban kerja & akurasi data operasional | *Work Session Integrity Ratio* terukur akurat pada 100% peserta |
| **OBJ-WF-02** | Fragmentasi kanal absensi kantor | 100% metode kehadiran (NFC, QR Code, Scanner Admin) bermuara pada satu mesin validasi sentral | Menghilangkan duplikasi data & inkonsistensi status | Rasio keberhasilan ingesti presensi tanpa duplikasi scan |
| **OBJ-WF-03** | Pembengkakan klaim lembur sepihak | Kompensasi lembur dihitung murni dari jam aktual dalam jendela waktu yang disetujui supervisor | Efisiensi anggaran insentif & transparansi lembur | 0% klaim lembur tanpa persetujuan jadwal di muka |
| **OBJ-WF-04** | Pelanggaran privasi pekerja | Pemantauan aktivitas kerja berfokus pada visibilitas tab browser tanpa software pengintai invasif | Kepatuhan hukum privasi & peningkatan martabat kerja | 0 insiden surveilans invasif (screenshot/keylogger) |
| **OBJ-WF-05** | Perusakan data saat koreksi absensi manual | Koreksi presensi dicatat sebagai penyesuaian baru tanpa menimpa log asli | Akuntabilitas kepatuhan audit organisasi 100% | 100% log historis asli tetap utuh saat koreksi disetujui |

---

## 6. Business Scope

### 6.1 In Scope (Lingkup Bisnis Domain Ini)
1. **Tata Kelola Jadwal Kerja (*Work Schedule*):** Konfigurasi hari kerja aktif, jam masuk, jam pulang, jendela istirahat, toleransi keterlambatan (*grace period*), dan kalender libur resmi (FR-007).
2. **Penyerapan Log Kehadiran Multi-Metode (*Central Ingestion*):** Penerimaan event kehadiran via Kartu NFC fisik, QR Code dinamis ber-TTL 30 detik, dan terminal scanner operator (FR-008).
3. **Manajemen Sesi Kerja (*Work Session Tracking*):** Pengukuran durasi kotor (*Gross*), durasi aktif (*Active*), dan durasi idle (*Idle*) melalui portal kerja harian (FR-009).
4. **Manajemen Status Istirahat (*Break State Engine*):** Mesin status istirahat, deteksi anomali istirahat awal (*Early Break*), dan penegakan sanksi keterlambatan istirahat (*Unauthorized Break*) (FR-010).
5. **Integritas Sesi Ramah Privasi (*Session Integrity Tracking*):** Pemantauan fokus tab browser dan deteksi diam/tidak aktif >15 menit tanpa perekam layar atau keylogger (FR-011).
6. **Tata Kelola Permohonan Lembur (*Overtime Workflow*):** Pengajuan lembur di muka, peninjauan supervisor, dan kalkulasi kompensasi berbasis jam kerja aktual (FR-012).
7. **Pendaftaran Perangkat Pemindai (*Device Registry*):** Inventarisasi terminal fisik dan isolasi akses operator pemindai (FR-013, BR-002).
8. **Tata Kelola Izin & Cuti (*Leave Management*):** Pengajuan cuti sakit/akademik, persetujuan pembimbing, dan pembebasan penalti kehadiran (FR-014).
9. **Tata Kelola Koreksi Absensi (*Attendance Correction*):** Pengajuan perbaikan data presensi yang salah/gagal dengan preservasi jejak audit asli (FR-015, BR-025).

### 6.2 Out of Scope (Di Luar Lingkup Bisnis Domain Ini)
1. **Pengadaan & Perakitan Perangkat Keras Fisik:** Pembelian fisik alat baca NFC, instalasi kabel reader, dan perakitan casing terminal di kantor (hanya menyediakan kontrak data Device Gateway).
2. **Kalkulasi Nilai Poin Gamifikasi (XP Engine):** Perhitungan penambahan/pengurangan poin XP atas kehadiran diproses di luar domain ini (*Domain Performance & Gamification*).
3. **Pembayaran Gaji Bulanan Karyawan Tetap (*Corporate Payroll*):** Domain ini hanya mengelola jam kehadiran dan kompensasi lembur magang, bukan payroll korporat skala penuh.
4. **Metode Presensi Biometrik Lanjutan (Fase Masa Depan):** Verifikasi wajah (*Face Recognition*) dan pelacakan GPS seluler (*Geofencing*) dicatat sebagai *Future Scope*.

---

## 7. Stakeholders

| Stakeholder | Responsibility | Business Interest | Decision Authority |
|---|---|---|---|
| **Program Lead / Operations Owner (Reihan)** | Menetapkan kebijakan jam kerja, batas toleransi, dan kalender operasional magang | Efisiensi program, kepatuhan jadwal, dan integritas data kerja | Pengesahan master policy jadwal kerja dan toleransi |
| **HR / Internship Coordinator** | Mengelola penetapan jadwal kerja batch, memproses izin cuti, dan meninjau kepatuhan kehadiran | Ketertiban administrasi peserta magang dan validitas surat izin kampus/dokter | Persetujuan cuti dan eskalasi koreksi absensi batch |
| **Supervisor Lapangan (Guild Master)** | Membimbing tim, memantau kehadiran harian via portal "Team Today", menyetujui lembur & koreksi | Produktivitas tim, penyelesaian tugas tepat waktu, dan disiplin kerja | Persetujuan lembur (*Overtime*), koreksi kehadiran, dan izin harian |
| **Intern (Adventurer)** | Melakukan presensi masuk/pulang, mengoperasikan sesi kerja/istirahat, mengajukan lembur/cuti | Transparansi jam kerja, keadilan evaluasi, dan kepastian rekam jejak | Mengajukan permohonan lembur, cuti, dan koreksi data diri |
| **Scanner Operator (Gatekeeper)** | Mengoperasikan terminal scanner fisik di lobi kantor saat jam kedatangan/kepulangan | Kelancaran proses tap kartu tanpa antrean penumpukan di pintu masuk | Melaporkan kendala kartu tidak terbaca ke administrator |
| **Finance Administrator** | Meninjau akumulasi jam lembur aktual yang disetujui untuk pencairan insentif | Akurasi pembayaran bonus lembur sesuai bukti kehadiran yang sah | Validasi pembayaran kompensasi lembur |

---

## 8. Business Actors

### 8.1 Pemetaan Business Role vs System / RBAC Role

| Business Role | Definisi Aktivitas Bisnis | System / RBAC Role Terkait | Cakupan Akses Data (*Scope*) |
|---|---|---|---|
| **Workforce Admin** | Mengonfigurasi master jadwal kerja, kalender libur, toleransi, dan registrasi scanner | Super Admin, Admin, HR Admin | `scope = workforce_and_people` / `system` |
| **Field Supervisor** | Memantau kehadiran tim secara live, menyetujui lembur, memvalidasi koreksi absensi | Supervisor | `scope = assigned_team` |
| **Intern Worker** | Menjalankan presensi harian, mengontrol timer sesi kerja di portal, melapor kendala | Intern | `scope = own_data` |
| **Alumni Contributor** | Menjalankan sesi kerja saat mengerjakan proyek publik (tanpa absensi kantor harian) | Alumni | `scope = own_data + public_projects` |
| **Terminal Attendant** | Memfasilitasi tap kartu NFC / scan QR fisik di gerbang masuk kantor | Scanner Operator | `scope = scanner_only` *(Isolasi Ketat BR-002)* |

---

## 9. Current Business Process (AS-IS)

Berdasarkan Bagian 2.1 PRD DCISP v1.0, kondisi operasional saat ini (*AS-IS*) adalah:
1. **Pencatatan Kehadiran Manual/Terfragmentasi:** Peserta magang mencatat kehadiran pada buku tamu fisik atau form online terpisah yang tidak terhubung dengan jadwal kerja.
2. **Ketiadaan Visibilitas Sesi Kerja:** Perusahaan tidak memiliki instrumen untuk membedakan antara waktu kedatangan di gedung dengan waktu mulai bekerja aktif di depan komputer.
3. **Pengajuan Lembur Pasca-Fakta (*Post-Facto*):** Peserta magang seringkali bekerja hingga malam tanpa persetujuan tertulis di muka, lalu menuntut kompensasi berdasarkan estimasi jam kasar.
4. **Manipulasi Koreksi Absensi:** Saat peserta lupa mencatat kehadiran, koreksi dilakukan secara lisan kepada admin yang kemudian menimpa data mentah tanpa jejak audit siapa yang mengubah dan apa bukti pendukungnya.

---

## 10. Target Business Process (TO-BE)

Proses bisnis target mengintegrasikan seluruh interaksi kehadiran dan sesi kerja ke dalam alur terstandarisasi:

```text
[ Tiba di Gerbang Kantor ]
       │
       ▼
[ 1. PRESENTASI IDENTITAS (NFC / QR) ]
       │
       ├──► Validasi Kartu & Debouncing (30 Detik)
       ├──► Pencocokan Jadwal Kerja & Toleransi (Grace Period 10 Menit)
       ▼
[ 2. PENCATATAN EVENT 'CHECK_IN' (Append-Only) ]
       │
       ├── IF: Jam Datang <= Jadwal + 10m ──► Status: ON_TIME (Audio: "Good Morning!...")
       └── IF: Jam Datang >  Jadwal + 10m ──► Status: LATE (Audio: "Perhatian. Anda terlambat.")
       │
       ▼
[ 3. BUKA PORTAL "MY DAY" & START WORK ]
       │
       ├── Intern menekan [START WORK] ──► Status Sesi: WORKING (Event: WORK_STARTED)
       ├── Pemilih Tugas Aktif (Mengaitkan durasi dengan Task Proyek atau Daily Work)
       ▼
[ 4. EKSEKUSI SESI KERJA & PEMANTAUAN INTEGRITAS ]
       │
       ├── Pemantauan Tab Visibility & Focus (Tanpa screenshot/keylogger)
       └── IF: Idle > 15 Menit ──► Sesi ditandai IDLE (AFK Prompt: "Masih bekerja?")
       │
       ▼
[ 5. JEDA ISTIRAHAT (BREAK MANAGEMENT) ]
       │
       ├── Pukul 12:00: Intern menekan [TAKE BREAK] ──► Status: BREAK (Event: BREAK_STARTED)
       │    └── IF: Mulai sebelum 12:00 ──► Flag: Early Break Anomaly
       ├── Pukul 12:55: Intern menekan [RESUME WORK] ──► Status: WORKING (Event: BREAK_ENDED)
       │    └── IF: Kembali > 13:00 ──► Flag: Unauthorized Break Violation (Penalti Deduksi)
       │
       ▼
[ 6. ALUR LEMBUR (JIKA MEMBUTUHKAN JAM KERJA EKSTRA) ]
       │
       ├── Pengajuan Form Lembur di Muka ──► Supervisor Meninjau & Menyetujui Jendela Waktu
       └── Pelaksanaan Lembur ──► Sistem mengukur jam aktual di dalam jendela yang disetujui
       │
       ▼
[ 7. PENYELESAIAN SESI KERJA & CHECK-OUT FISIK ]
       │
       ├── Intern menekan [END WORK] ──► Pengisian ringkasan laporan kerja harian
       └── Tap kartu keluar di Monolit Gerbang ──► Event: CHECK_OUT (Menutup agregasi harian)
```

---

## 11. Business Process Scenarios

### Scenario 1: Presensi Kedatangan Normal & Sesi Kerja Harian
* **Trigger:** Intern tiba di gerbang kantor pada pagi hari kerja.
* **Precondition:** Intern berstatus `ACTIVE`, terdaftar pada batch aktif, dan jadwal kerja hari tersebut telah disahkan.
* **Actor:** Intern, Scanner Operator (opsional jika mandiri).
* **Input:** Tap kartu fisik NFC atau pemindaian QR code dinamis di terminal lobi.
* **Main Flow:**
  1. Terminal membaca data kredensial dan mengirimkan ke sistem.
  2. Sistem memvalidasi bahwa scan ini bukan duplikasi dalam 30 detik terakhir (*debouncing*).
  3. Sistem membandingkan waktu server terhadap jadwal kerja (contoh: jadwal 08:30, kedatangan 08:28).
  4. Sistem mencatat event `CHECK_IN` bertanda `ON_TIME` pada log kehadiran permanen.
  5. Terminal memutar audio konfirmasi kehadiran sukses: *“Good Morning! Absensi berhasil.”*
  6. Intern menuju meja kerja, membuka laptop, masuk ke portal "My Day", dan menekan tombol `[START WORK]`.
  7. Sistem mencatat event `WORK_STARTED`, memulai penghitung waktu sesi kerja aktif, dan mengaitkan sesi dengan tugas harian yang dipilih.
* **Business Decision:** Kehadiran disahkan sebagai tepat waktu; hak memulai sesi kerja di portal "My Day" diaktifkan.
* **Output:** Record `CHECK_IN` dan `WORK_STARTED` tersimpan; dashboard live supervisor terbarui.
* **Postcondition:** Status intern pada sistem adalah `WORKING`.

### Scenario 2: Penanganan Keterlambatan Kedatangan (*Late Arrival*)
* **Trigger:** Intern melakukan check-in pada pukul 08:48 (Jadwal masuk: 08:30, Grace Period: 10 menit).
* **Precondition:** Jadwal kerja aktif dengan batas toleransi maksimal 08:40.
* **Actor:** Intern.
* **Input:** Tap kartu NFC pada terminal presensi.
* **Main Flow:**
  1. Sistem memvalidasi bahwa waktu kedatangan (08:48) melewati ambang batas toleransi (08:40).
  2. Sistem menghitung durasi keterlambatan bersih yaitu 18 menit.
  3. Sistem mencatat event `CHECK_IN` dengan klasifikasi status `LATE (Kategori: 16–30 Menit)`.
  4. Terminal memutar audio peringatan: *“Perhatian. Anda terlambat.”*
  5. Sistem meneruskan data keterlambatan ke modul performa untuk pengenaan penalti kehadiran.
* **Business Decision:** Kedatangan ditandai terlambat tingkat 2; flag anomali dicatat pada dashboard supervisor.
* **Output:** Record kehadiran berstatus `LATE` dengan durasi 18 menit.
* **Postcondition:** Status hadir aktif, notifikasi peringatan disiplin terkirim ke intern.

### Scenario 3: Alur Permohonan & Pelaksanaan Lembur (*Overtime*)
* **Trigger:** Intern membutuhkan waktu kerja tambahan untuk menyelesaikan rilis deliverable proyek.
* **Precondition:** Intern memiliki tugas proyek aktif dengan tenggat waktu mendesak.
* **Actor:** Intern (Pemohon), Supervisor (Penyetujui).
* **Input:** Formulir pengajuan lembur memuat tanggal, estimasi jendela waktu (17:00–19:30), justifikasi, dan referensi tugas.
* **Main Flow:**
  1. Intern mengisi formulir pengajuan lembur pada portal sebelum jam kerja normal berakhir (pukul 16:00).
  2. Permohonan masuk ke antrean *Approval Center* pada portal "Team Today" supervisor.
  3. Supervisor meninjau urgensi tugas dan menyetujui jendela lembur resmi pukul 17:00–19:30.
  4. Pukul 17:00, sistem secara otomatis mendaftarkan jendela lembur aktif.
  5. Intern bekerja hingga pukul 19:00, lalu menekan `[END WORK]` dan melakukan tap `CHECK_OUT` pada terminal gerbang.
  6. Sistem menghitung durasi lembur aktual yang terlaksana yaitu 2,0 jam (17:00–19:00), bukan 2,5 jam yang diajukan.
* **Business Decision:** Durasi lembur yang disahkan untuk kompensasi adalah 2,0 jam (mengacu pada jam kerja riil dalam batas jendela yang disetujui).
* **Output:** Record lembur berstatus `COMPLETED` dengan durasi terkompensasi 2,0 jam.
* **Postcondition:** Data lembur terverifikasi diteruskan ke bagian keuangan untuk kompensasi.

### Scenario 4: Rekonsiliasi Koreksi Absensi Manual (*Attendance Correction*)
* **Trigger:** Intern lupa melakukan check-out pada sore hari sebelumnya karena terburu-buru.
* **Precondition:** Sistem mendeteksi anomali *Missing Checkout* pada penutupan hari dan mengenakan penalti.
* **Actor:** Intern (Pemohon), Supervisor (Penyetujui).
* **Input:** Pengajuan koreksi (tanggal kejadian, waktu pulang riil 17:05, bukti foto/surat pendukung, alasan minimal 20 karakter).
* **Main Flow:**
  1. Intern mengajukan permohonan koreksi absensi melalui menu *Corrections*.
  2. Supervisor menerima notifikasi dan memeriksa bukti lampiran serta log aktivitas tugas intern pada jam tersebut.
  3. Supervisor menyetujui (*APPROVE*) pengajuan koreksi.
  4. Sistem membuat record event penyesuaian baru bertipe `MANUAL_CORRECTION` yang mereferensikan hari tersebut.
  5. Sistem mempertahankan log asli tanpa menghapusnya, menghitung ulang total durasi kerja hari tersebut, dan membatalkan penalti *Missing Checkout*.
* **Business Decision:** Data jam kerja hari tersebut disahkan lengkap; penalti absensi dicabut secara sah.
* **Output:** Record `Attendance Correction` berstatus `APPROVED`; jejak audit mencatat identitas supervisor dan stempel waktu persetujuan.
* **Postcondition:** Rekap jam kerja dan status disiplin intern terpulihkan.

---

## 12. Business Requirements

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis (*Business Requirement*) | Rasional Bisnis (*Business Rationale*) | Prioritas | Sumber PRD |
|---|---|---|---|---|
| **BRQ-WF-001** | Bisnis mewajibkan seluruh jadwal kerja, jam masuk, jam pulang, jendela istirahat, toleransi keterlambatan, dan hari libur dapat dikonfigurasi secara dinamis | Fleksibilitas kebijakan jam kerja antar-batch tanpa perlu merombak sistem | P0 | FR-007, BR-001 |
| **BRQ-WF-002** | Bisnis mewajibkan seluruh metode presensi fisik (NFC, QR Code dinamis) bermuara pada satu mesin validasi terpusat | Mencegah fragmentasi data kehadiran dan memusatkan penegakan aturan jadwal | P0 | FR-008, BR-006 |
| **BRQ-WF-003** | Bisnis mewajibkan sistem membedakan secara tegas 5 dimensi waktu kerja (*Attendance, Work Session, Active, Break, Overtime*) | Menjamin keadilan pengukuran beban kerja riil dan mengeliminasi manipulasi presensi | P0 | BR-007, Section 6.5 |
| **BRQ-WF-004** | Bisnis mewajibkan sesi kerja harian terikat pada jenis pekerjaan yang jelas (*Daily Work* atau *Project Task*) | Memastikan akuntabilitas jam kerja terhadap deliverable proyek perusahaan | P0 | FR-009, BR-005 |
| **BRQ-WF-005** | Bisnis mewajibkan penegakan status istirahat eksplisit dengan pemantauan anomali istirahat awal dan sanksi istirahat melebihi batas | Menjaga disiplin jam kerja kantor dan keadilan bagi peserta yang patuh jadwal | P0 | FR-010, BR-008, BR-009 |
| **BRQ-WF-006** | Bisnis mewajibkan pemantauan keaktifan kerja berfokus pada status browser (fokus/blur/idle) tanpa alat pengawasan invasif (screenshot/keylogger) | Menghormati privasi dan martabat pekerja sesuai standar etika dan hukum | P0 | FR-011, BR-010, BR-011 |
| **BRQ-WF-007** | Bisnis mewajibkan seluruh kerja lembur memiliki persetujuan supervisor di muka dan kompensasi dihitung dari jam kerja aktual | Mencegah pembengkakan biaya insentif sepihak dan memastikan efektivitas lembur | P0 | FR-012, BR-013, BR-014 |
| **BRQ-WF-008** | Bisnis mewajibkan operator terminal scanner diisolasi secara ketat hanya pada fungsi pemindaian presensi gerbang | Mencegah kebocoran data rahasia perusahaan dan data personal peserta di area umum | P0 | FR-001, FR-013, BR-002 |
| **BRQ-WF-009** | Bisnis mewajibkan pengajuan cuti resmi yang disetujui membebaskan peserta dari kewajiban hadir dan penalti absensi (0 Penalti) | Memberikan perlindungan hak izin sakit dan keperluan akademik bagi mahasiswa magang | P1 | FR-014, BR-012 |
| **BRQ-WF-010** | Bisnis mewajibkan koreksi absensi manual mempertahankan log historis asli secara utuh (*append-only correction*) | Menjamin integritas audit kepatuhan organisasi dan mencegah manipulasi riwayat | P0 | FR-015, BR-025 |
| **BRQ-WF-011** | Bisnis mewajibkan tersedianya mekanisme pencegahan pemindaian ganda (*debouncing*) minimal 30 detik pada setiap titik tap presensi | Menghindari pencatatan transaksi ganda akibat kartu yang tertempel berulang kali | P0 | FR-008, Section 11.2 |
| **BRQ-WF-012** | Bisnis mewajibkan terminal presensi dapat mencatat kehadiran saat jaringan offline dan menyinkronkannya kembali saat online | Menjamin kelancaran operasional masuk kantor saat terjadi gangguan internet | P0 | Section 11.2, RSK-005 |

---

## 13. Business Rules

| ID Aturan | Nama Aturan Bisnis | Kondisi Bisnis (*Condition*) | Tindakan Sistem (*Action*) | Pengecualian (*Exception*) |
|---|---|---|---|---|
| **BRULE-WF-001** | *Dynamic Schedule & Grace Period* | Check-in diterima $\le \text{start\_time} + \text{grace\_period}$ (default 10 menit) | Tandai kehadiran sebagai `ON_TIME` | Cuti yang telah disetujui membebaskan kewajiban check-in |
| **BRULE-WF-002** | *Late Arrival Penalty Tiering* | Check-in diterima $> \text{start\_time} + \text{grace\_period}$ | Klasifikasikan keterlambatan: Tingkat 1 (1–15m), Tingkat 2 (16–30m), Tingkat 3 (>30m) | Hari libur resmi kalender membebaskan keterlambatan |
| **BRULE-WF-003** | *5-Way Time Independence* | Sesi kerja dan presensi berlangsung | Sistem menghitung 5 metrik waktu secara mandiri: $\text{Attend} \neq \text{Work} \neq \text{Active} \neq \text{Break} \neq \text{Overtime}$ | Tidak ada pengecualian |
| **BRULE-WF-004** | *Work Session Precondition* | Intern menekan tombol `[START WORK]` | Wajib memiliki record `CHECK_IN` kehadiran sah pada hari yang sama | Alumni yang mengerjakan proyek publik remote |
| **BRULE-WF-005** | *Break Anomaly Enforcement* | Intern memulai break sebelum jadwal (misal < 12:00) | Tandai record dengan flag `Early Break Anomaly` | Izin khusus dari supervisor |
| **BRULE-WF-006** | *Unauthorized Break Violation* | Intern menekan `[RESUME WORK]` melebihi akhir jadwal istirahat (> 13:00) | Tandai pelanggaran `Unauthorized Break` dan kenakan penalti kehadiran (-2 XP) | Penundaan istirahat resmi yang disetujui pembimbing |
| **BRULE-WF-007** | *Privacy-First Activity Monitoring* | Sesi kerja aktif berjalan di browser pengguna | Pantau status `visibilityState` dan `blur`; jika idle $>15\text{ menit}$, tandai status `Idle` | Dilarang keras merekam layar, keylogger, intip chat |
| **BRULE-WF-008** | *Non-Punitive Tab Switching* | Pengguna berpindah tab browser saat sesi kerja | Kurangi penghitung durasi aktif tanpa mengenakan pemotongan poin langsung | Tidak ada |
| **BRULE-WF-009** | *Pre-Approved Overtime* | Jam kerja berada di luar batas jadwal normal ($> 17:00$) | Wajib memiliki record pengajuan lembur yang berstatus `APPROVED` oleh supervisor | Skenario darurat server yang disahkan Super Admin |
| **BRULE-WF-010** | *Actual Worked Overtime Calculation* | Eksekusi lembur selesai dicatat | Durasi lembur terkompensasi dihitung: $\text{Actual OT} \le \text{Approved Window}$ | Jika jam kerja riil melebihi window, kelebihan waktu tidak dihitung |
| **BRULE-WF-011** | *Scanner Operator Boundary* | Pengguna dengan peran Scanner Operator melakukan login | Sistem mengisolasi akses secara ketat hanya pada tampilan input scanner (`scope = scanner_only`) | Tidak ada pengecualian (BR-002) |
| **BRULE-WF-012** | *Audit Trail on Attendance Correction* | Pengajuan koreksi absensi disetujui supervisor | Buat record penyesuaian baru `MANUAL_CORRECTION` tanpa mengubah atau menghapus data log asli | Tidak ada pengecualian (BR-025) |
| **BRULE-WF-013** | *Scan Debouncing Window* | Kartu/QR yang sama dipindai berulang dalam rentang waktu $\le 30\text{ detik}$ | Abaikan pemindaian kedua sebagai duplikasi dan jangan simpan log ganda | Pemindaian setelah 30 detik dianggap event baru |
| **BRULE-WF-014** | *Dynamic Rotating QR Code TTL* | Pemindaian kehadiran menggunakan QR Code aplikasi | String QR Code wajib memiliki masa kedaluwarsa maksimal 30 detik | Screenshot QR yang kedaluwarsa otomatis ditolak |

---

## 14. Business Entities

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                 BUSINESS ENTITIES: WORKFORCE & ATTENDANCE                   │
├──────────────────────────┬──────────────────────────────────────────────────┤
│ Work Schedule (DATA-002) │ Definisi jam kerja acuan, break, toleransi, libur│
│ Attendance Event (DATA-001) Catatan peristiwa kehadiran fisik append-only  │
│ Work Session             │ Sesi pelacakan jam kerja aktif & tugas terkait   │
│ Break                    │ Rekaman rentang istirahat dan status anomali     │
│ Overtime Request         │ Dokumen pengajuan dan pengesahan jam lembur      │
│ Leave Request            │ Dokumen pengajuan izin cuti sakit/akademik       │
│ Attendance Correction    │ Dokumen rekonsiliasi koreksi absensi manual      │
│ Device Terminal          │ Identitas terminal pemindai resmi di gerbang     │
└──────────────────────────┴──────────────────────────────────────────────────┘
```

---

## 15. Data Flow

```text
[ Terminal Reader (NFC/QR) ] 
       │ (1. Credential Ingestion)
       ▼
[ Central Attendance Ingestion Engine ]
       │ (2. Validate Schedule & Debounce 30s)
       ▼
[ Attendance Event Log (DATA-001) ] ──(Append-Only Persistence)
       │
       ├──► (3a. Sync Real-time Status) ──► [ Dashboard "Team Today" Supervisor ]
       │
       ├──► (3b. Enable Work Controls) ──► [ Portal "My Day" Intern ]
       │                                           │
       │                                           ▼ (Start Work / Break / End Work)
       │                                   [ Work Session Tracking Engine ]
       │                                           │
       │                                           ▼
       │                                   [ Work Session & Break Logs ]
       │
       └──► (3c. Trigger Evaluasi Poin) ──► [ Downstream: XP Rules Engine ]
```

---

## 16. Status & Lifecycle

```text
1. WORK SESSION LIFECYCLE:
   [ IDLE ] ──(START WORK)──► [ WORKING ] ◄─────────────────┐
                                  │                         │
                                  ├──(TAKE BREAK)           │
                                  ▼                         │
                              [ BREAK ] ──(RESUME WORK)─────┘
                                  │
                                  └──(END WORK)──► [ ENDED ]

2. OVERTIME REQUEST LIFECYCLE:
   [ SUBMITTED ] ──► [ UNDER_REVIEW ] ──► [ REJECTED ]
                            │
                            └──► [ APPROVED ] ──(Kerja Terlaksana)──► [ COMPLETED ]

3. ATTENDANCE CORRECTION LIFECYCLE:
   [ SUBMITTED ] ──► [ UNDER_REVIEW ] ──► [ REJECTED ]
                            │
                            └──► [ APPROVED ] ──► [ RECALCULATED ]

4. LEAVE REQUEST LIFECYCLE:
   [ SUBMITTED ] ──► [ UNDER_REVIEW ] ──► [ REJECTED ]
                            │
                            └──► [ APPROVED ] ──► [ ACTIVE_LEAVE ]
```

---

## 17. Approval & Decision Flow

* **Persetujuan Lembur:** Intern $\rightarrow$ Supervisor $\rightarrow$ Otorisasi jendela jam lembur.
* **Koreksi Absensi:** Intern $\rightarrow$ Supervisor / HR $\rightarrow$ Penerbitan record `MANUAL_CORRECTION`.
* **Izin Cuti:** Intern $\rightarrow$ HR Admin / Supervisor $\rightarrow$ Pembebasan penalti absensi 0 XP.

---

## 18. Exception & Failure Handling

* **Jaringan Offline:** Terminal menyimpan ke antrean lokal LittleFS + audio *offline_success* $\rightarrow$ Sync otomatis saat online.
* **Scan Ganda:** Debouncing 30 detik mengabaikan scan kedua + audio *too_frequent*.
* **Kartu Tidak Terdaftar:** Ditolak + audio *card_unregistered*.

---

## 19. Business Reporting

1. **Rekapitulasi Kehadiran Bulanan:** Punctuality rate, keterlambatan, anomali istirahat, rekap cuti.
2. **Utilisasi Lembur Tim:** Total jam diajukan vs jam aktual riil yang disetujui.
3. **Audit Koreksi Presensi:** Jejak koreksi manual dan supervisor yang menyetujui.

---

## 20. KPI & Success Metrics

* **Punctuality Rate (MET-01):** Persentase kehadiran tepat waktu (`Target: TBD`).
* **Work Session Integrity Index (MET-02):** Rasio waktu aktif terhadap sesi kotor (`Target: TBD`).
* **System Ingestion Success (MET-06):** Rasio tap sukses tanpa galat sistemik (`Target: TBD`).

---

## 21. Business Integration

* **Ke Domain Projects:** Mengirimkan durasi kerja sesi harian untuk task tracking.
* **Ke Domain Performance:** Mengirimkan data keterlambatan/anomali untuk mutasi XP.
* **Ke Domain Finance:** Mengirimkan total jam lembur aktual yang disahkan untuk bonus.

---

## 22. Business Dependencies

* Bergantung upstream pada **Identity & RBAC** (Auth & Scanner Isolation) dan **People & Batches** (Status Intern Aktif).

---

## 23. Business Constraints

* **Privasi Mutlak (BR-010):** Dilarang screenshot, keylogger, intip chat, scan proses OS.
* **Immutability (BR-025):** Riwayat log presensi append-only (no update/delete).
* **Isolasi Scanner (BR-002):** Operator terminal dibatasi `scope = scanner_only`.

---

## 24. Business Assumptions

* Terminal pemindai fisik NFC/QR beroperasi di pintu masuk kantor PT. ADT.

---

## 25. Business Gaps

* Kuota jatah cuti magang dan standar tarif nominal lembur berstatus `[TBD]`.

---

## 26. Ambiguities

* *Daily Work vs Task Submission terpisah secara mandiri (BR-005).*

---

## 27. Conflicting Requirements

* *Event WORK_ENDED dikelola pada level siklus sesi kerja, sedangkan log presensi ditutup CHECK_OUT (OQ-014).*

---

## 28. Traceability Matrix

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Entitas Terkait | Acceptance Criteria |
|---|---|---|---|---|
| FR-007 (Schedule) | BRQ-WF-001 | BRULE-WF-001, BRULE-WF-002 | `Work Schedule` | AC-WF-001 |
| FR-008 (Ingestion) | BRQ-WF-002, BRQ-WF-011 | BRULE-WF-001, BRULE-WF-013 | `Attendance Event Log` | AC-WF-002, AC-WF-003 |
| FR-009 (Session) | BRQ-WF-003, BRQ-WF-004 | BRULE-WF-003, BRULE-WF-004 | `Work Session` | AC-WF-004 |
| FR-010 (Break) | BRQ-WF-005 | BRULE-WF-005, BRULE-WF-006 | `Break` | AC-WF-005 |
| FR-011 (Integrity) | BRQ-WF-006 | BRULE-WF-007, BRULE-WF-008 | `Work Session` | AC-WF-006 |
| FR-012 (Overtime) | BRQ-WF-007 | BRULE-WF-009, BRULE-WF-010 | `Overtime Request` | AC-WF-007 |
| FR-013 (Device) | BRQ-WF-008 | BRULE-WF-011 | `Device` | AC-WF-008 |
| FR-014 (Leave) | BRQ-WF-009 | BRULE-WF-001 | `Leave Request` | AC-WF-009 |
| FR-015 (Correction)| BRQ-WF-010 | BRULE-WF-012 | `Attendance Correction` | AC-WF-010 |

---

## 29. Acceptance Criteria

### AC-WF-001: Validasi Ketepatan Waktu Terhadap Grace Period
* **Given:** Jadwal kerja 08:30 dengan grace period 10 menit (maksimal 08:40).
* **When:** Intern tap kartu pukul 08:38 $\rightarrow$ Status `ON_TIME` + Audio *check_in*.
* **When (Terlambat):** Intern tap kartu pukul 08:42 $\rightarrow$ Status `LATE (12 Menit)` + Audio *late*.

### AC-WF-002: Pencegahan Pemindaian Ganda (Debouncing 30s)
* **Given:** Intern telah berhasil tap kartu check-in.
* **When:** Kartu yang sama di-tap kembali dalam selang 15 detik.
* **Then:** Sistem mengabaikan tap kedua, memutar audio *too_frequent*, dan tidak mencatat log ganda.

### AC-WF-003: Ketahanan Presensi Offline Terminal
* **Given:** Jaringan internet terminal gerbang terputus.
* **When:** Intern menempelkan kartu NFC yang valid.
* **Then:** Terminal menyimpan tap ke antrean lokal, memutar audio *offline_success*, dan menyinkronkan data secara otomatis saat koneksi pulih.

---

## 30. Open Questions

* `OQ-WF-01`: Kuota hari jatah izin cuti magang (`TBD`).
* `OQ-WF-02`: Standar tarif nominal per jam lembur (`TBD`).

---

## 31. BRD Completion Checklist

- [x] Semua kebutuhan domain Workforce & Attendance teridentifikasi lengkap.
- [x] 25 Aturan Bisnis dan 10 Acceptance Criteria terpetakan eksplisit.

---
*Dokumen ini disimpan permanen di `brd/02-BRD-WORKFORCE-ATTENDANCE.md`.*
