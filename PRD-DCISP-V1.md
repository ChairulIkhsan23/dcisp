# PRODUCT REQUIREMENTS DOCUMENT
# Dagang Creative Intern Solutions Program (DCISP)

---

## 1. Document Control

### 1.1 Informasi Dokumen

| Atribut | Detail |
|---|---|
| **Nama Produk** | Dagang Creative Intern Solutions Program (DCISP) |
| **Judul Dokumen** | Product Requirements Document (PRD) — Platform DCISP |
| **Versi Dokumen** | 1.0 (Audit-Verified & Standardized) |
| **Status Dokumen** | Final Draft / Ready for Engineering Review |
| **Tanggal Terbit** | 2026-09-19 |
| **Penulis Utama** | Chairul Ikhsan (Product Analyst & Requirements Engineer) |
| **Pemilik Produk (Owner)** | PT. Aplikasi Dagang Teknologi |
| **Sumber Dokumen** | Catatan Kebutuhan Pengguna, Spesifikasi Sistem DCISP v1.0, & Dokumen Audit |
| **Klasifikasi Akses** | Internal Proprietary — PT. Aplikasi Dagang Teknologi |

### 1.2 Riwayat Revisi Dokumen

| Versi | Tanggal | Penulis | Deskripsi Perubahan | Status Validasi |
|---|---|---|---|---|
| **1.0-D** | 2026-09-19 | Chairul Ikhsan | Draf awal kebutuhan fungsional, aturan bisnis, dan katalog entitas data | Draft |
| **1.0-A** | 2026-09-19 | Tim Audit Produk | Audit kepatuhan 100% preservasi logika bisnis, pencegahan invensi liar, penandaan OQ-001 s/d OQ-013 | Lulus Audit |
| **1.0** | 2026-09-19 | Senior Product Team | Standardisasi 18 Bagian PRD formal, penyempurnaan profil 10 peran, siklus hidup, matriks izin, penambahan edge cases, states, DoD, dan Appendices A-C | Disetujui |

### 1.3 Pengesahan Dokumen (Approvals)

| Peran | Nama | Jabatan | Tanggal | Status |
|---|---|---|---|---|
| **Product Owner** | `[TBD]` | Head of Product, PT. Aplikasi Dagang Teknologi | `[TBD]` | Pending Review |
| **Technical Lead** | `[TBD]` | Principal Software Architect | `[TBD]` | Pending Review |
| **Business Stakeholder** | Reihan | Program Lead / Operations Owner | `[TBD]` | Pending Review |

---

## 2. Product Context

### 2.1 Background (Latar Belakang)

DCISP dikembangkan oleh PT. Aplikasi Dagang Teknologi untuk menjawab kebutuhan pengelolaan program magang dan talenta muda berbasis proyek (*project-based workforce*). Sebelum inisiatif DCISP, manajemen magang di lingkungan perusahaan menghadapi tantangan operasional:

1. **Fragmentasi Operasional:** Kehadiran dicatat secara manual/terpisah, penugasan proyek dilakukan secara informal melalui aplikasi perpesanan, dan pelaporan progres kerja tidak memiliki repositori bukti terpusat.
2. **Ketiadaan Visibilitas Waktu Kerja:** Tidak adanya pemisahan antara waktu kehadiran di kantor (*attendance time*) dengan durasi kerja nyata (*work session*), istirahat (*break*), dan kerja lembur (*overtime*).
3. **Ketidakjelasan Kontribusi Proyek:** Alokasi insentif/bounty proyek tidak mencerminkan kontribusi nyata anggota tim, melainkan kesepakatan informal yang rentan sengketa.
4. **Kehilangan Aset Talenta Pasca-Magang:** Akun peserta magang dinonaktifkan setelah program berakhir (*dead accounts*), memutus hubungan dengan alumni bertalenta yang sebenarnya dapat diberdayakan untuk proyek-proyek publik perusahaan.
5. **Kerapuhan Logika Bisnis:** Aturan kehadiran, penalti, pajak, dan deduksi diatur secara manual atau di-hardcode, menyulitkan adaptasi kebijakan organisasi.

Evolusi paradigma produk DCISP mengalami tiga fase:
- **Fase 1 (Konsep Awal):** *"Internship Management + Gamification"* — alat pencatatan magang dasar dengan skor game.
- **Fase 2 (Evolusi):** *"Work & Internship Performance Management Platform"* — perluasan ke arah pelacakan kinerja terstruktur.
- **Fase 3 (Konsep Final Pasca-Audit):** *"Platform pengelolaan individu dan project-based work yang memiliki internship lifecycle sebagai core business dan gamification sebagai performance layer."*

Platform berfungsi sebagai catatan kerja digital yang tidak dapat diubah (*immutable digital work record*), menghubungkan kehadiran granular dengan deliverable proyek, kontribusi terverifikasi, reward multi-tingkat, dan kompensasi finansial yang patuh hukum.

### 2.2 Problem Statement (Pernyataan Masalah)

Permasalahan bisnis yang diselesaikan DCISP dikelompokkan berdasarkan domain fungsional:

1. **Domain People:** Siklus transisi kandidat magang menjadi peserta aktif hingga menjadi alumni tidak terkelola dalam satu siklus hidup kontinu. Data institusi asal dan skill matrix peserta tidak terdokumentasi rapi.
2. **Domain Attendance:** Ingestion kehadiran dari berbagai kanal fisik (kartu NFC, QR code dinamis, pemindai admin) tidak bermuara pada satu mesin validasi jadwal sentral.
3. **Domain Work Session & Breaks:** Tidak adanya batas tegas antara sesi kerja aktif, waktu idle/jeda, dan jam istirahat. Hal ini sering menimbulkan manipulasi durasi kerja atau sebaliknya, ketidakadilan bagi peserta yang bekerja intensif.
4. **Domain Overtime:** Permintaan lembur berlangsung tanpa persetujuan terdokumentasi di muka, serta kompensasi lembur sering dihitung dari durasi pengajuan mentah, bukan dari jam kerja aktual yang disetujui dalam jadwal.
5. **Domain Project & Tasks:** Proyek tidak memiliki mekanisme publikasi terstruktur (marketplace internal/publik) dan pembatasan kapasitas/kuota pelamar yang transparan. Penugasan tugas (*tasks*) sering tidak terhubung ke milestone proyek.
6. **Domain Work Reports & Evidence:** Pelaporan kerja harian (*Daily Work*) bercampur baur dengan penyerahan tugas proyek (*Project Task Submission*). Bukti kerja (*evidence*) tersebar tanpa validasi tautan repositori git, berkas dokumen, atau tangkapan layar terverifikasi.
7. **Domain Performance & Gamification:** Terjadi kerancuan fatal di mana skor rank gamifikasi disamakan dengan skor evaluasi kinerja profesional ($	ext{Rank} 
eq 	ext{Performance Score}$). Pengakuan Top Performer tidak memiliki formula komposit yang baku.
8. **Domain Compensation & Finance:** Perhitungan bagi hasil proyek (*bounty*) tidak memiliki tahapan audit transparan (Planned vs Actual vs Final). Pemotongan pajak dan kas perpisahan batch (*Farewell Contribution*) bercampur baur dan tidak memiliki pembukuan ganda (*double-entry ledger*) yang tidak dapat diubah (*immutable*).
9. **Domain Alumni:** Alumni kehilangan akses platform, sehingga potensi keterlibatan alumni dalam proyek publik perusahaan menjadi terhambat.

### 2.3 Product Vision (Visi Produk)

> *"Platform pengelolaan individu dan project-based work yang memiliki internship lifecycle sebagai core business dan gamification sebagai performance layer."*

DCISP memposisikan program magang sebagai **perjalanan progresi profesional** (*professional progression journey*), di mana setiap aktivitas kerja nyata diubah menjadi rekam jejak terverifikasi, poin pengalaman (XP), pengakuan prestasi (Rank/Achievement), dan kompensasi finansial yang adil.

### 2.4 Value Proposition (Proposisi Nilai Stakeholder)

DCISP memberikan proposisi nilai spesifik bagi 8 pemangku kepentingan utama:

| Pemangku Kepentingan | Proposisi Nilai Utama |
|---|---|
| **1. Intern (Peserta Magang)** | Transparansi penuh atas sesi kerja harian melalui portal *"My Day"*, kejelasan target tugas dan milestone, akumulasi XP dan kenaikan rank yang objektif, kepastian kompensasi bounty berbasis kontribusi nyata, serta dompet digital pribadi dengan jejak audit itemized. |
| **2. Alumni** | Retensi akun permanen pasca-kelulusan, akses eksklusif ke marketplace proyek publik perusahaan, hak memperoleh bounty dan Project XP tanpa memengaruhi rank magang aktif, serta pembentukan portofolio digital profesional terverifikasi secara otomatis. |
| **3. Supervisor** | Visibilitas real-time kehadiran dan aktivitas tim melalui portal *"Team Today"*, alur persetujuan lembur dan koreksi presensi yang terstruktur, instrumen evaluasi kinerja berkala, serta kendali penentuan persentase *Final Contribution* bounty tim secara berkeadilan. |
| **4. Project Manager** | Kemudahan menerbitkan proyek di marketplace dengan kriteria skill dan kuota kapasitas terukur, visibilitas pelamar berbakat melalui *Skill Matching Engine*, serta pemantauan delivery milestone dan tugas secara terpusat. |
| **5. HR / Internship Admin** | Pengelolaan menyeluruh siklus hidup peserta (Applicant $	o$ Intern $	o$ Alumni), pengelompokan kohort batch, manajemen relasi institusi pendidikan, tata kelola kuota cuti, dan otomatisasi pengakuan tepat SATU *Top Performer* per batch. |
| **6. Finance Administrator** | Orkestrasi otomatis distribusi dana: $	ext{Gross Bounty} - 	ext{Tax/Deductions} - 	ext{Farewell Contribution} = 	ext{Net Distributable}$. Buku kas ganda (*double-entry ledger*) yang immutabel, pemisahan tegas antara pajak negara dengan kas perpisahan angkatan (*Batch Fund*), dan tata kelola pencairan dana (*payout*). |
| **7. Management / Executive (PT. ADT)** | Dashboard eksekutif (*Command Center*) dengan metrik real-time kehadiran, efisiensi operasional proyek, utilisasi anggaran, kepatuhan regulasi kerja/pajak di Indonesia, dan pipeline akuisisi talenta masa depan. |
| **8. Scanner Operator / Admin Lapangan** | Antarmuka pemindai kehadiran yang terisolasi khusus (`scope = scanner_only`), cepat, stabil, dan minim distraksi untuk memproses kartu NFC dan QR code di terminal fisik kantor. |

### 2.5 Product Principles (Prinsip Produk)

DCISP dirancang dengan 8 prinsip dasar yang mengikat seluruh domain arsitektur:

1. **Privacy-First Tracking (BR-010, BR-011):** Pemantauan integritas sesi kerja berfokus pada status browser (fokus, blur, idle). Sistem **TIDAK PERNAH** melakukan pengawasan invasif: tidak ada tangkapan layar diam-diam (*screenshot*), tidak ada perekaman ketukan tombol (*keylogger*), tidak ada inspeksi konten tab, tidak ada pemindaian aplikasi desktop, tidak ada perekaman clipboard, dan tidak ada pembacaan obrolan pribadi. Perpindahan tab tidak memicu penalti otomatis.
2. **Dynamic Policy, Zero Hardcoding (BR-001, BR-012, BR-020):** Seluruh parameter operasional — izin peran, ambang penalti XP, formula deduksi pajak, jadwal kerja, toleransi keterlambatan — dikendalikan melalui *Unified Policy Engine* (FR-045). Tidak ada logika bisnis krusial yang dikunci mati di dalam kode program (*zero hardcoded checks*).
3. **Gamification as Performance Layer, Not a Toy (BR-017):** Gamifikasi adalah instrumen pendorong produktivitas profesional, bukan permainan anak-anak. Rank mencerminkan akumulasi progresi gamified; *Performance Score* mencerminkan evaluasi kualitas kerja formal oleh atasan ($	ext{Rank} 
eq 	ext{Performance Score}$).
4. **Immutable Financial Records (BR-022, BR-023):** Setiap peristiwa moneter (alokasi bounty, pemotongan pajak, setoran dana batch, penarikan dompet) wajib dicatat dalam buku besar berpasangan (*double-entry ledger*) yang bersifat kekal (*append-only*, tidak dapat diedit atau dihapus).
5. **Audit Trail Preservation on Exception (BR-025):** Tindakan koreksi manual (seperti koreksi kehadiran atau penyesuaian kontribusi) tidak boleh menimpa log peristiwa historis. Data asli tetap utuh di log audit bersama rekam jejak siapa yang mengubah, kapan, dan alasannya.
6. **Informational AI, Non-Punitive Automation (BR-015):** Rekomendasi kecerdasan buatan (*Skill Matching*) murni bersifat informatif dan membantu pengambil keputusan manusia. Sistem dilarang melakukan penolakan kandidat secara otomatis tanpa intervensi peninjau.
7. **Strict Separation of Concerns (BR-004, BR-005, BR-021):**
   - 3 Jenis Pekerjaan terpisah: *Daily Work* $
eq$ *Project Task* $
eq$ *Work Report*.
   - 3 Jalur XP terpisah: *Internship XP* $
eq$ *Project XP* $
eq$ *Alumni Contribution*.
   - Pemisahan Dana: Dana Perpisahan (*Farewell Contribution*) adalah *Batch Contribution*, terpisah dari *Pajak Resmi*.
8. **Cloudflare R2 Storage Separation (BR-024):** Database SQL hanya menyimpan metadata berkas (`disk`, `path_key`, `mime`, `size`, `metadata`). Penyimpanan berkas biner secara langsung di database (BLOB) sangat dilarang; berkas fisik disimpan di Cloudflare R2.

---

## 3. Users & Roles

### 3.1 Personas (Profil Persona)

1. **Andi (Peserta Magang / Intern):** Mahasiswa tingkat akhir jurusan Teknik Informatika yang membutuhkan wadah magang profesional terstruktur, ingin melacak jam kerjanya dengan adil, mengumpulkan XP untuk naik rank, serta memperoleh insentif proyek nyata.
2. **Budi (Alumni Magang):** Lulusan magang DCISP batch sebelumnya yang telah bekerja paruh waktu, ingin tetap terhubung dengan ekosistem PT. ADT, mengerjakan proyek lepas (*public projects*) berbayar, dan memamerkan portofolio digital terverifikasi.
3. **Citra (Supervisor Lapangan):** Senior Engineer yang membimbing 5-8 peserta magang, membutuhkan alat monitoring kehadiran real-time tanpa repot memeriksa satu per satu, mengulas laporan tugas, menyetujui lembur, dan menilai kontribusi tim secara objektif.
4. **Dimas (Project Manager):** Manajer produk yang mengelola beberapa inisiatif proyek perusahaan, membutuhkan wadah untuk menerbitkan kebutuhan proyek, menyaring anggota berdasarkan kecocokan skill, dan menetapkan alokasi bounty pool.
5. **Eka (HR / Internship Coordinator):** Koordinator program magang yang mengurus penerimaan batch baru, administrasi surat menyurat ke kampus, validasi kehadiran formal, perizinan cuti, hingga penentuan Top Performer di akhir periode.
6. **Faisal (Finance Administrator):** Staf keuangan yang bertanggung jawab atas kepatuhan pajak penghasilan atas insentif magang, pemotongan dana kas angkatan, validasi permohonan pencairan dompet digital, dan rekonsiliasi buku kas.
7. **Reihan (Scanner Operator / Resepsionis):** Petugas operasional di lobi kantor yang bertugas mengawasi terminal presensi, memindai kartu fisik NFC atau QR code peserta magang saat tiba di kantor dengan cepat tanpa akses ke data rahasia perusahaan.
8. **Gita (Reviewer Tugas):** Quality Assurance lead yang fokus memvalidasi kualitas submission tugas teknis, memeriksa tautan PR/commit git, dan memberikan persetujuan atau permintaan revisi.
9. **Hendra (System Administrator):** Administrator IT yang mengelola akun pengguna, konfigurasi keamanan, rotasi kredensial API, monitoring penyimpanan R2, dan audit log sistem.
10. **Irfan (Super Administrator):** Pimpinan tertinggi teknologi dengan hak veto sistem penuh, penentu master policy engine, dan pengelola hak akses multi-peran.

### 3.2 Profil Peran Lengkap (Detailed Role Specifications)

Berikut adalah spesifikasi formal 10 peran pengguna pada platform DCISP:

#### R-01: SUPER ADMIN
- **Role:** Super Admin
- **Description:** Administrator tingkat tertinggi dengan kedaulatan penuh atas seluruh konfigurasi arsitektur, kebijakan, dan keamanan platform.
- **Primary Objective:** Menjamin integritas sistem, ketersediaan layanan, tata kelola keamanan, dan penerapan kebijakan organisasi secara menyeluruh.
- **Scope:** `scope = system` (Seluruh modul, seluruh domain, seluruh data organisasi).
- **Key Responsibilities:** Mengonfigurasi master *Unified Policy Engine*, mengelola perizinan peran dinamis (RBAC), memantau log audit keamanan, dan memelihara pengaturan global platform.
- **Permissions:** Akses penuh (`*.*`) mencakup seluruh aksi: `view`, `create`, `edit`, `delete`, `approve`, `export`, `override`, `configure`.
- **Restrictions:** Dibatasi oleh audit logging immutabel (setiap aksinya tetap tercatat dan tidak dapat dihapus).

#### R-02: ADMIN
- **Role:** Admin
- **Description:** Operator administratif umum yang menjalankan tata kelola operasional sistem harian.
- **Primary Objective:** Menjaga kelancaran operasional administratif multi-modul di luar konfigurasi sistem inti.
- **Scope:** `scope = system` (Operasional umum sistem; batasan spesifik terhadap Super Admin didefinisikan pada OQ-013).
- **Key Responsibilities:** Mengelola data pengguna, membantu penanganan isu operasional, mengelola repositori berkas publik, dan mencetak laporan dokumen formal.
- **Permissions:** Akses administratif standar: `users.manage`, `attendance.view_all`, `projects.view_all`, `finance.view_summary`, `documents.generate`.
- **Restrictions:** Tidak dapat mengubah konfigurasi policy engine tingkat root dan tidak dapat memanipulasi role Super Admin.

#### R-03: HR / INTERNSHIP ADMIN
- **Role:** HR / Internship Admin
- **Description:** Penanggung jawab operasional program magang dan manajemen sumber daya manusia.
- **Primary Objective:** Mengelola seluruh siklus hidup peserta magang dari penerimaan hingga kelulusan secara teratur dan akuntabel.
- **Scope:** `scope = workforce_and_people` (Data seluruh peserta magang, batch, jadwal, cuti, dan institusi pendidikan).
- **Key Responsibilities:** Mengelola siklus Applicant $	o$ Intern $	o$ Alumni, membuat batch magang, menetapkan jadwal kerja standar, memproses izin cuti (FR-014), dan menetapkan Top Performer per batch (FR-028).
- **Permissions:** `intern.manage`, `batch.manage`, `institution.manage`, `attendance.records.view`, `leave.approve`, `top_performer.manage`, `id_card.generate`, `certificate.manage`.
- **Restrictions:** Tidak memiliki akses langsung untuk mengubah transaksi keuangan (*finance write*) atau memanipulasi kode tugas proyek.

#### R-04: PROJECT MANAGER
- **Role:** Project Manager
- **Description:** Pengelola inisiatif proyek dan arsitek delivery pekerjaan berbasis proyek.
- **Primary Objective:** Memastikan proyek terdefinisi dengan jelas, memiliki tim yang kompeten, dan selesai sesuai batas waktu dan anggaran.
- **Scope:** `scope = assigned_projects` (Proyek yang dibuat atau ditugaskan kepadanya).
- **Key Responsibilities:** Menerbitkan proyek ke Marketplace (FR-016), meninjau aplikasi pelamar, membentuk tim proyek (FR-019), membuat milestone dan tugas, serta menetapkan bounty pool proyek.
- **Permissions:** `project.create`, `project.edit`, `project.application.review`, `milestone.manage`, `task.create`, `task.assign`, `bounty.set_pool`.
- **Restrictions:** Tidak dapat menyetujui koreksi kehadiran atau mengubah saldo dompet kas pribadi secara sepihak.

#### R-05: SUPERVISOR
- **Role:** Supervisor
- **Description:** Pemimpin teknis langsung di lapangan yang membawahi dan membimbing sejumlah peserta magang.
- **Primary Objective:** Memastikan kedisiplinan kerja harian tim, mengawal kualitas deliverable, dan memberikan evaluasi kinerja berkala yang adil.
- **Scope:** `scope = assigned_team` (Terbatas hanya pada anggota tim/peserta magang yang berada di bawah supervisinya).
- **Key Responsibilities:** Memantau dashboard *"Team Today"* (FR-053), menyetujui pengajuan lembur (FR-012), menyetujui koreksi kehadiran (FR-015), mengevaluasi kinerja berkala (FR-027), dan mengonfirmasi *Final Contribution %* (FR-024).
- **Permissions:** `team.attendance.view`, `overtime.approve`, `attendance_correction.approve`, `performance.evaluate`, `contribution.finalize`, `work_report.review`.
- **Restrictions:** Tidak memiliki akses melihat data peserta magang di luar tim supervisinya tanpa izin khusus.

#### R-06: REVIEWER
- **Role:** Reviewer
- **Description:** Penilai kualitas teknis independen atau spesialis domain terhadap tugas dan bukti kerja.
- **Primary Objective:** Menjaga standar mutu dan keabsahan setiap deliverable tugas sebelum dinyatakan selesai.
- **Scope:** `scope = assigned_tasks` (Tugas dan laporan kerja yang ditugaskan kepadanya untuk diulas).
- **Key Responsibilities:** Meninjau laporan kerja harian dan penyerahan tugas (*submission*), memvalidasi tautan repositori git atau berkas bukti (FR-023), memberikan status *Approved* atau *Revision Required*.
- **Permissions:** `submission.review`, `submission.approve`, `submission.reject`, `evidence.verify`, `work_report.view`.
- **Restrictions:** Tidak memiliki wewenang mengubah struktur tim proyek, mengubah bobot bounty, atau mengelola absensi.

#### R-07: FINANCE
- **Role:** Finance
- **Description:** Administrator keuangan dan kepatuhan perpajakan perusahaan.
- **Primary Objective:** Menjamin integritas finansial platform, akurasi perhitungan deduksi pajak/kas batch, dan keamanan distribusi kompensasi.
- **Scope:** `scope = financial_data` (Data transaksi moneter, buku besar/ledger, dompet pengguna, kas batch, dan payout).
- **Key Responsibilities:** Mengonfigurasi formula mesin deduksi & pajak (FR-034), mengelola buku besar immutabel (FR-037), memantau akumulasi kas angkatan (*Batch Fund*), dan memvalidasi pencairan dana (*payout engine* FR-038).
- **Permissions:** `finance.ledger.view`, `finance.tax.configure`, `finance.batch_fund.manage`, `wallet.payout.approve`, `finance.export`.
- **Restrictions:** Dilarang keras melakukan manipulasi log kehadiran, mengubah skor kinerja magang, atau menghapus entri buku besar (*no delete/update on ledger*).

#### R-08: SCANNER OPERATOR
- **Role:** Scanner Operator
- **Description:** Operator pemindai perangkat presensi fisik di area gerbang/lobi kantor (contoh: Reihan).
- **Primary Objective:** Memfasilitasi proses presensi fisik (tap NFC / scan QR) berjalan cepat, lancar, dan tanpa kendala antrean.
- **Scope:** `scope = scanner_only` (Terisolasi strictly pada terminal pemindai presensi).
- **Key Responsibilities:** Mengoperasikan aplikasi scanner pada terminal kehadiran, memantau konfirmasi keberhasilan tap kartu, dan melaporkan kendala kartu tidak terbaca.
- **Permissions:** `attendance.scanner.view`, `attendance.scanner.scan`.
- **Restrictions:** Secara eksplisit **DILARANG** mengakses modul data peserta (`intern.manage`), keuangan (`finance.view`), konfigurasi (`settings.manage`), manajemen tugas, maupun laporan evaluasi *(BR-002)*.

#### R-09: INTERN
- **Role:** Intern
- **Description:** Peserta aktif yang sedang menjalani program magang di PT. ADT.
- **Primary Objective:** Menyelesaikan tugas magang, berpartisipasi dalam proyek, meningkatkan keahlian profesional, dan membangun rekam jejak kerja terverifikasi.
- **Scope:** `scope = own_data` (Data pribadi, kehadiran pribadi, proyek yang diikuti, tugas pribadi, dompet pribadi).
- **Key Responsibilities:** Melakukan presensi masuk/pulang harian, mengoperasikan sesi kerja dan istirahat di portal *"My Day"* (FR-052), mengajukan lembur/cuti/koreksi, mengerjakan tugas proyek, mengirimkan work report dan evidence, serta mengklaim reward rank.
- **Permissions:** `attendance.own.log`, `work_session.own.manage`, `project.marketplace.view`, `project.apply`, `task.own.submit`, `reward.claim`, `wallet.own.view`.
- **Restrictions:** Tidak dapat mengakses data kehadiran atau rincian dompet peserta lain, tidak dapat menyetujui laporannya sendiri.

#### R-10: ALUMNI
- **Role:** Alumni
- **Description:** Talenta yang telah menyelesaikan program magang secara resmi dan mempertahankan akun aktif di platform.
- **Primary Objective:** Berkolaborasi dalam proyek-proyek publik perusahaan, mengasah portofolio profesional, dan memperoleh pendapatan insentif/bounty.
- **Scope:** `scope = own_data + public_projects` (Data pribadi, riwayat portofolio, dompet pribadi, dan proyek marketplace berstatus publik).
- **Key Responsibilities:** Menjelajahi marketplace proyek publik, mengajukan lamaran proyek, mengerjakan tugas proyek yang diterima, mengirimkan bukti kerja, menerima bounty, dan mengelola profil portofolio digital (FR-003, FR-042).
- **Permissions:** `project.public.view`, `project.public.apply`, `task.own.submit`, `wallet.own.view`, `portfolio.manage`, `alumni_xp.earn`.
- **Restrictions:** Tidak dapat masuk ke dalam perhitungan *Top Performer per Batch* peserta magang aktif, tidak memiliki akses ke proyek internal bertanda `INTERN_ONLY`, dan aktivitasnya tidak mengubah rank magang aktif *(BR-004)*.

### 3.3 Arsitektur Izin (Permissions Architecture)

Sistem DCISP mengadopsi struktur otorisasi formal berjenjang:
$$	ext{Role} \longrightarrow 	ext{Permission} \longrightarrow 	ext{Scope} \longrightarrow 	ext{Resource} \longrightarrow 	ext{Action}$$

- **Konfigurasi Dinamis (BR-001):** Hubungan antara peran (*role*) dan izin (*permission*) disimpan dalam basis data dan dikelola melalui antarmuka administratif. Tidak ada pemeriksaan izin yang di-hardcode seperti `if (user.role == 'ADMIN')` pada kode sumber; sistem memeriksa izin granular seperti `if (user.hasPermission('attendance.correction.approve', context))`.
- **Katalog Aksi Sumber Daya (Actions):** `view`, `create`, `edit`, `delete`, `approve`, `reject`, `submit`, `review`, `scan`, `export`, `claim`.

### 3.4 Definisi Ruang Lingkup (Scopes)

| Kode Scope | Definisi Akses Data | Contoh Penggunaan |
|---|---|---|
| `system` | Akses menyeluruh tanpa batas hierarki organisasi | Super Admin, Admin |
| `workforce_and_people` | Akses data seluruh peserta magang, batch, jadwal, dan absensi | HR / Internship Admin |
| `assigned_team` | Akses terbatas hanya pada data peserta yang dibimbing langsung | Supervisor |
| `assigned_projects` | Akses terbatas pada proyek di mana pengguna menjadi PM/anggota | Project Manager |
| `financial_data` | Akses data transaksi moneter, ledger, dompet, dan pajak | Finance Administrator |
| `scanner_only` | Akses tampilan terminal pemindai presensi kamera/NFC | Scanner Operator *(BR-002)* |
| `own_data` | Akses terbatas hanya pada data milik entitas pengguna sendiri | Intern, Alumni |
| `public_data` | Akses informasi yang dipublikasikan terbuka (misal proyek publik) | Alumni, Pengguna Tamu Terdaftar |

### 3.5 Matriks Akses Peran Berdasarkan Modul (Role-Based Access Matrix)

| Modul Navigasi | Super Admin | Admin | HR Admin | Project Mgr | Supervisor | Reviewer | Finance | Scanner Op | Intern | Alumni |
|---|---|---|---|---|---|---|---|---|---|---|
| **Command Center** | Full | Full | Full | PM Scope | Team Scope | — | Fin Scope | — | My Day | My Day |
| **People — Interns** | Full | Full | Full | Read | Team Read | — | — | — | — | — |
| **People — Batches** | Full | Full | Full | Read | Read | — | — | — | — | — |
| **Attendance — Live** | Full | Full | Full | — | Team Read | — | — | — | Own Read | — |
| **Attendance — Scanner**| Full | Full | Full | — | — | — | — | Scan Only | — | — |
| **Attendance — Correction**| Full | Full | Full | — | Team Appr | — | — | — | Own Submit | — |
| **Attendance — Overtime**| Full | Full | Full | — | Team Appr | — | — | — | Own Submit | — |
| **Attendance — Leave** | Full | Full | Full | — | Team Read | — | — | — | Own Submit | — |
| **Projects — Market** | Full | Full | Read | Manage | Read | Read | — | — | View/Apply | Public View/Apply |
| **Projects — Tasks** | Full | Full | Read | Full | Team Manage | Review | — | — | Own Submit | Own Submit |
| **Performance — Eval** | Full | Full | Full | Read | Team Eval | Review | — | — | Own View | — |
| **Performance — XP/Rank**| Full | Full | Full | Read | Read | Read | — | — | Own View | Own View |
| **Rewards — Claims** | Full | Full | Read | — | — | — | Process | — | Own Claim | Own Claim |
| **Finance — Wallets** | Full | Read | — | — | — | — | Manage | — | Own View | Own View |
| **Finance — Ledger** | Full | Read | — | — | — | — | Manage | — | — | — |
| **Documents — Reports** | Full | Full | Full | Read | Team Read | — | — | — | Own Gen | Own Gen |
| **System — Settings** | Full | Read | — | — | — | — | — | — | — | — |

*Keterangan: Full = Create, Read, Update, Delete, Approve; Team = Khusus anggota tim yang dibimbing; Own = Khusus rekam data pengguna sendiri; Scan Only = Tampilan input reader tanpa visibilitas data lain.*

---

## 4. Goals & Success Metrics

### 4.1 Product Goals (Tujuan Produk)
1. **Penyatuan Siklus Hidup Tunggal:** Menyatukan pengelolaan kehadiran, jam kerja aktif, delivery proyek, evaluasi kinerja, gamifikasi, kompensasi, perpajakan, hingga portal alumni dalam satu arsitektur terpadu.
2. **Integritas Waktu Kerja Granular:** Menerapkan pemisahan presisi 5 dimensi waktu kerja (*Attendance, Work Session, Active, Break, Overtime*) untuk mengeliminasi manipulasi data presensi.
3. **Kompensasi Berbasis Kontribusi Riil:** Mengotomatiskan perhitungan alokasi bounty proyek melalui alur 3-lapis (*Planned $	o$ Actual $	o$ Final*) dan orkestrasi pemotongan pajak/kas batch yang patuh regulasi.
4. **Retensi Hubungan Talenta Alumni:** Menyediakan akses berkelanjutan bagi alumni untuk berkontribusi pada proyek perusahaan dan memamerkan portofolio digital terverifikasi.

### 4.2 Business Goals (Tujuan Bisnis PT. ADT)
1. **Efisiensi Operasional Program Magang:** Mengurangi beban administrasi manual tim HR, Supervisor, dan Finance dalam rekapitulasi jam kerja, evaluasi berkala, dan perhitungan insentif.
2. **Kepatuhan Pajak & Audit Finansial:** Memastikan pemotongan pajak penghasilan atas insentif proyek/magang tercatat akurat dan transparan melalui buku besar ganda immutabel.
3. **Peningkatan Kualitas Deliverable Proyek:** Mendorong penyelesaian deliverable proyek perusahaan secara tepat waktu dan berkualitas melalui insentif bounty dan pengakuan prestasi berbasis XP/Rank.
4. **Talent Pipeline Berkualitas Tinggi:** Mengidentifikasi talenta berkinerja terbaik (*Top Performer*) berbasis data objektif untuk rekrutmen karyawan tetap di masa mendatang.

### 4.3 User Goals (Tujuan Pengguna)
- **Intern:** Menikmati pengalaman magang yang jelas, terpantau, dihargai dengan adil, tidak dieksploitasi, serta mendapatkan insentif dan sertifikat/portofolio nyata.
- **Supervisor:** Mengurangi friksi mikromanajemen kehadiran, memantau progres kerja secara visual, dan memiliki instrumen terstandar dalam membina tim.
- **Finance:** Menghindari kekeliruan perhitungan bagi hasil proyek, memastikan saldo batch fund terpisah dari pajak, dan memiliki jejak audit keuangan yang kokoh.

### 4.4 Success Metrics (Metrik Keberhasilan Produk)

> **Status Tata Kelola:** Sesuai prinsip *No Invention*, target angka absolut (KPI numerik) belum ditetapkan secara formal dalam dokumen sumber dan berstatus `TBD`. Tabel di bawah menyajikan **Proposed Candidate Metrics** yang direkomendasikan untuk disahkan oleh Product Owner:

| ID Metrik | Nama Metrik yang Diusulkan | Definisi Operasional | Status Target | Pemilik Metrik |
|---|---|---|---|---|
| **MET-01** | Attendance Punctuality Rate | Persentase kehadiran tepat waktu terhadap total jadwal kerja aktif | `TBD` | HR Admin |
| **MET-02** | Work Session Integrity Index | Rasio durasi kerja aktif (*Active Time*) terhadap total durasi sesi kerja (*Gross Session*) | `TBD` | Supervisor Lead |
| **MET-03** | Project Milestone On-Time Delivery | Persentase milestone proyek yang diserahkan sebelum atau tepat batas tenggat | `TBD` | Project Manager |
| **MET-04** | Bounty Settlement Turnaround Time | Durasi waktu dari persetujuan deliverable proyek hingga dana masuk ke dompet pribadi | `TBD` | Finance Lead |
| **MET-05** | Alumni Project Engagement Rate | Persentase alumni aktif yang melamar atau mengerjakan proyek publik perusahaan | `TBD` | Program Lead |
| **MET-06** | System Availability & Ingestion Success | Rasio keberhasilan pencatatan tap presensi fisik tanpa kegagalan sistemik | `TBD` | Tech Lead |

### 4.5 Non-Goals (Batasan Bukan-Tujuan Produk)

Untuk menjaga fokus pengembangan dan kepatuhan etika privasi, DCISP secara eksplisit **BUKAN** merupakan:
1. **Bukan Perangkat Pengawasan Invasif (*Not an Invasive Employee Surveillance Tool*):** Sistem dilarang mengambil tangkapan layar desktop (*screenshot*), merekam penekanan tombol (*keylogger*), menginspeksi isi percakapan pribadi, atau melacak aktivitas di luar tab aplikasi resmi *(BR-010)*.
2. **Bukan Pengganti Sistem Payroll Perusahaan Eksternal:** DCISP mengelola insentif magang, reward, bounty proyek, dan kas perpisahan batch; DCISP bukan software payroll penggajian karyawan tetap korporat skala penuh.
3. **Bukan Game Hiburan Mandiri (*Not a Standalone Video Game*):** Gamifikasi dihadirkan murni sebagai *performance layer* pendorong disiplin dan motivasi kerja, bukan untuk menghadirkan gameplay fantasi yang mengaburkan tanggung jawab kerja profesional.
4. **Bukan Marketplace Freelance Terbuka Bebas:** Marketplace proyek DCISP ditujukan terbatas untuk lingkungan internal perusahaan (peserta magang aktif) dan proyek publik khusus alumni terverifikasi; bukan platform freelance bebas untuk publik umum tanpa kurasi.

---

## 5. Scope

### 5.1 In Scope (Lingkup yang Dikerjakan)

Pengembangan DCISP diklasifikasikan ke dalam tiga tingkat prioritas implementasi:

#### Prioritas P0 (Wajib Ada / Core MVP)
- RBAC & Unified Permission Engine berbasis basis data dinamis (FR-001, BR-001).
- Manajemen Siklus Hidup Peserta Magang (Applicant $	o$ Intern $	o$ Alumni) (FR-002).
- Manajemen Kelompok Batch & Dana Kas Angkatan (FR-004, FR-036).
- Work Schedule Engine dengan kalender libur dan toleransi keterlambatan (FR-007).
- Central Attendance Event Engine dengan multi-method ingestion QR & NFC (FR-008, BR-006).
- Work Session Tracking & Break State Engine dengan penegakan anomali (FR-009, FR-010, BR-008).
- Session Integrity Tracking berbasis visibilitas browser tanpa pelanggaran privasi (FR-011, BR-010).
- Alur Pengajuan & Persetujuan Overtime terikat jam kerja aktual (FR-012, BR-013, BR-014).
- Alur Pengajuan & Rekonsiliasi Koreksi Presensi dengan pemeliharaan jejak audit (FR-015, BR-025).
- Project Marketplace dengan visibilitas granular (*Intern Only, Public, Private*) (FR-016).
- Project Application & Quota Workflow (FR-017).
- Manajemen Tim Proyek & Milestone Management (FR-019, FR-020).
- Task Management, Work Report harian, dan Evidence Management terverifikasi (FR-021, FR-022, FR-023).
- Three-Layer Contribution Engine (*Planned $	o$ Actual $	o$ Final*) (FR-024, BR-016).
- XP Rules Engine dinamis dengan partisi 3 skema XP terpisah (FR-025, BR-004, BR-012).
- Rank Progression System bertingkat terpisah dari skor evaluasi formal (FR-026, BR-017).
- Performance Evaluation Engine formal oleh supervisor (FR-027, BR-017).
- Multi-Component Rank Rewards (Tunai, Merchandise, Badge, Voucher) (FR-031, BR-019).
- Project Bounty Distribution terikat persentase kontribusi akhir (FR-032).
- Deduction & Tax Engine dinamis mengikuti regulasi perpajakan (FR-034, BR-020, BR-021).
- Personal Wallet dengan pencatatan mutasi itemized multi-sumber (FR-035, BR-022).
- Financial Ledger berpasangan yang tidak dapat diubah (*immutable double-entry*) (FR-037, BR-023).
- Financial Flow Orchestration ($	ext{Gross} - 	ext{Tax} - 	ext{Farewell} = 	ext{Net}$) (FR-039).
- Integrasi Penyimpanan Berkas Cloudflare R2 dengan pemisahan metadata di database (FR-044, BR-024).
- Unified Policy Engine terpusat untuk konfigurasi aturan runtime (FR-045).
- Audit Logging & Security Trail kekal (*append-only*) (FR-047).

#### Prioritas P1 (Sebaiknya Ada / Tahap Penguatan)
- Pengakuan tepat SATU Top Performer per batch per periode evaluasi (FR-028, BR-018).
- Skill Matrix Repositori & Profil Keahlian Pengguna (FR-006).
- Retensi Akun Alumni & Partisipasi Proyek Publik Berkelanjutan (FR-003, BR-003).
- Manajemen Relasi Institusi Pendidikan / Kampus (FR-005).
- Device Registry inventaris terminal presensi NFC/Scanner (FR-013).
- Manajemen Pengajuan & Persetujuan Cuti Peserta (FR-014).
- Achievement System pengumpulan badge pencapaian (FR-029).
- Reward Claim Lifecycle Management (FR-033).
- Payout Engine pencairan saldo dompet digital pribadi (FR-038).
- Generator Kartu Identitas Digital (ID Card) ber-QR/NFC (FR-040).
- Digital Portfolio Builder alumni otomatis berbasis deliverable terverifikasi (FR-042).
- Event-Driven Notification Center multi-kanal (FR-046).
- Command Center Executive Dashboard (FR-051).
- Portal Harian Peserta "My Day" (FR-052).
- Portal Operasional Supervisor "Team Today" & Dedicated Supervisor Workspace (FR-053, FR-054).
- Global Sidebar Navigation terisolasi peran (FR-055).

#### Prioritas P2 (Bagus Jika Ada / Tahap Kecerdasan Lanjutan)
- Smart Skill Matching Engine berbasis skor kesesuaian keahlian (FR-018, BR-015).
- Matriks Pertumbuhan Keahlian Pengguna seiring waktu (FR-030).
- Certificate Engine dengan QR verifikasi dan tanda tangan digital (FR-041, OQ-001).
- Generator Dokumen Laporan Magang & Logbook Formal (FR-043).
- Modul Media & Galeri Dokumentasi Kegiatan (FR-044, OQ-008).
- Modul Analitik & Pelaporan Agregat Lanjutan (FR-049).
- AI Insights & Activity Intelligence pendeteksi pola anomali (FR-050, OQ-009).

### 5.2 Out of Scope (Di Luar Lingkup Pengembangan)

> **Catatan Audit Sumber:** Dokumen sumber PRD awal menyatakan `[TIDAK DISEBUTKAN DALAM SUMBER]`. Mengacu pada prinsip kehati-hatian (*Derived Inference with Explicit Marking*), batas ruang lingkup di luar platform disepakati sebagai berikut:

1. `[Inferred / Out of Scope]` **Sistem Penggajian Karyawan Tetap (Corporate Payroll):** Pengelolaan gaji pokok, tunjangan hari raya (THR), dan asuransi BPJS ketenagakerjaan karyawan tetap tidak ditangani oleh platform ini.
2. `[Inferred / Out of Scope]` **Integrasi Perangkat Keras Biometrik Fisik:** Pengadaan, perakitan fisik reader, dan instalasi firmware mesin pembaca kartu NFC/RFID di kantor; platform hanya menyediakan Device Gateway API dan registry.
3. `[Inferred / Out of Scope]` **Aplikasi Mobile Native (iOS / Android Native):** Fase awal DCISP dibangun berbasis web responsif (Progressive Web-ready), bukan aplikasi native toko aplikasi.
4. `[Inferred / Out of Scope]` **Situs Pemasaran Publik Bebas (Public Marketing Site):** Portal promosi korporat umum berada di luar cakupan aplikasi operasional DCISP.

### 5.3 Future Scope (Pengembangan Masa Depan)

Item-item yang secara eksplisit dicatat sebagai visi masa depan pada dokumen sumber:
1. **Metode Presensi Biometrik Lanjutan (FR-008):** Pengenalan wajah (*Face Recognition*), verifikasi geolokasi (*Geofencing Mobile*), dan verifikasi sidik jari perangkat keras.
2. **Kecerdasan Buatan Prediktif Penuh (FR-050):** Model AI prediktif untuk mendeteksi dini kejenuhan kerja peserta (*burnout detection*), rekomendasi kurikulum magang terpersonalisasi, dan pemetaan otomatis tugas terhadap spesialisasi industri.

---

## 6. Product / Business Process

### 6.1 Alur Proses Bisnis End-to-End (End-to-End Business Flow)

Alur bisnis DCISP mengorkestrasi aktivitas dari pendaftaran talenta hingga kolaborasi pasca-magang sebagai alumni:

```text
                       [ PEOPLE ]
                           │
                    Intern / Alumni
                           │
                           ▼
                     [ ATTENDANCE ]
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
      QR / NFC / Admin             Work Schedule
             │                           │
             └─────────────┬─────────────┘
                           ▼
                    [ WORK SESSION ]
                           │
             ┌─────────────┼─────────────┐
             ▼             ▼             ▼
         Work Work       Break       Overtime
             │                           │
             │                       Approval
             │                           │
             └─────────────┬─────────────┘
                           ▼
                       [ PROJECT ]
                           │
                           ▼
                     [ APPLICATION ]
                           │
                           ▼
                      [ ASSIGNMENT ]
                           │
                           ▼
                        [ TASK ]
                           │
                           ▼
                     [ WORK REPORT ]
                           │
                           ▼
                       [ EVIDENCE ]
                           │
                           ▼
                        [ REVIEW ]
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
      [ PERFORMANCE ]                 [ XP ]
             │                           │
             └─────────────┬─────────────┘
                           ▼
                        [ RANK ]
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
         [ REWARD ]                  [ BOUNTY ]
             │                           │
             ▼                           ▼
       Reward Package          [ TAX / DEDUCTION ENGINE ]
                                         │
                                         ▼
                               [ FAREWELL CONTRIBUTION ]
                                         │
                               ┌─────────┴─────────┐
                               ▼                   ▼
                           [ PERSONAL          [ BATCH
                            WALLET ]            FUND ]
                               │
                               ▼
                           [ PAYOUT ]
                               │
                               ▼
                           [ ALUMNI ]
                               │
                     ┌─────────┴─────────┐
                     ▼                   ▼
             [ PUBLIC PROJECT       [ PORTFOLIO ]
               MARKETPLACE ]
```

### 6.2 Siklus Hidup Inti (Core Business Lifecycles)

#### 6.2.1 Internship Lifecycle (Siklus Hidup Peserta Magang)
```text
Registration (Applicant mendaftar dengan data profil & institusi)
  ↓
Selection (HR / Supervisor meninjau kecocokan berkas & skill)
  ↓
Acceptance (Kandidat diterima, ditetapkan ke Batch tertentu, & akun Intern aktif)
  ↓
Active Internship (Menjalankan presensi harian, proyek, tugas, akumulasi XP & evaluasi berkala)
  ↓
Completion (Evaluasi akhir batch, penerbitan sertifikat kelulusan & kartu alumni)
  ↓
Alumni (Akun bertransisi ke peran Alumni secara permanen dengan hak proyek publik)
```

#### 6.2.2 Attendance Lifecycle (Siklus Hidup Presensi)
```text
Card / QR Presentation (Peserta melakukan tap kartu NFC atau scan QR dinamis pada reader)
  ↓
Identify User (Device Gateway mengirim data credential ke Attendance Event Engine)
  ↓
Validate Schedule (Sistem mencocokkan timestamp terhadap Work Schedule, kalender libur, & grace period)
  ↓
Validate Event (Pemeriksaan urutan logis event: misal CHECK_IN harus didahului atau berurutan dengan ARRIVED)
  ↓
Record Attendance (Menyimpan log transaksi immutabel dengan atribut method, device, & metadata)
  ↓
Update Timeline & Trigger (Memperbarui status Live Attendance, sinkronisasi widget portal, & trigger aturan XP)
```

#### 6.2.3 Work Session Lifecycle (Siklus Hidup Sesi Kerja)
```text
Check-in (Kehadiran terkonfirmasi di kantor / sistem)
  ↓
Start Work (Intern menekan tombol "Start Work" pada portal My Day; event WORK_STARTED tercatat)
  ↓
Active Session (Pekerjaan aktif berlangsung dengan pemantauan visibilitas window/tab yang menjaga privasi)
  ↓
Break Window (Intern menekan jeda / jadwal istirahat tiba; event BREAK_STARTED memicu status BREAK)
  ↓
Resume Work (Intern kembali aktif; event BREAK_ENDED & WORK_RESUMED dicatat; validasi anomali keterlambatan)
  ↓
End Work (Batas jadwal berakhir atau pengguna menekan "End Work"; ringkasan harian diserahkan)
  ↓
Check-out (Pemindaian keluar kantor; event CHECK_OUT menutup agregasi waktu harian & memicu kalkulasi XP)
```

#### 6.2.4 Project Lifecycle (Siklus Hidup Proyek)
```text
Draft (Project Manager menyusun rincian proyek, skill yang dibutuhkan, kuota, & alokasi bounty pool)
  ↓
Published (Proyek diterbitkan ke Marketplace dengan visibilitas: Intern Only, Public, atau Private)
  ↓
Application (Kandidat Intern/Alumni mengajukan lamaran; sistem menampilkan indikator Skill Matching %)
  ↓
Review & Shortlisting (PM meninjau portofolio/skill pelamar dan menyusun daftar kandidat terpilih)
  ↓
Approved / Team Formed (Pelamar diterima ke dalam struktur Project Team dengan kesepakatan Planned Contribution %)
  ↓
In Progress (Eksekusi tugas, milestone delivery, penyerahan work report, dan verifikasi evidence)
  ↓
Completed (Seluruh milestone tuntas, Final Contribution % disahkan, & distribusi porsi bounty dieksekusi)
  ↓
Archived (Proyek diarsipkan; rekam jejak otomatis dikompilasi ke dalam Portofolio Digital anggota)
```

#### 6.2.5 Performance & Gamification Lifecycle (Siklus Hidup Kinerja)
```text
Daily Work & Project Activity (Pekerjaan harian, presensi tepat waktu, penyelesaian tugas proyek)
  ↓
Contribution Verification (Verifikasi bukti kerja oleh reviewer & pengesahan kontribusi oleh supervisor)
  ↓
Formal Quality Evaluation (Supervisor mengisi evaluasi berkala menghasilkan formal Performance Score)
  ↓
XP Scoring Engine (Sistem memproses kalkulasi XP dinamis ke dalam 3 partisi: Internship XP, Project XP, Alumni XP)
  ↓
Rank Progression (Akumulasi XP menaikkan tingkatan Rank; Rank != Performance Score)
  ↓
Reward & Recognition (Promosi rank menerbitkan paket reward multi-komponen; pengakuan 1 Top Performer per batch)
```

#### 6.2.6 Finance & Compensation Lifecycle (Siklus Hidup Keuangan)
```text
Bounty Pool Allocation (Alokasi dana imbalan saat proyek selesai disahkan)
  ↓
Contribution Splitting (Perhitungan porsi kotor anggota tim berdasarkan Final Contribution %)
  ↓
Deduction & Tax Engine (Perhitungan pemotongan pajak penghasilan resmi & iuran kas perpisahan Batch Fund)
  ↓
Wallet Crediting (Dana bersih Net Distributable dikreditkan ke Personal Wallet pengguna)
  ↓
Double-Entry Ledger (Pencatatan mutasi kredit/debit secara permanen pada buku besar berpasangan immutabel)
  ↓
Payout Processing (Pengguna mengajukan penarikan dana; bagian Finance memvalidasi & mengeksekusi transfer bank)
```

### 6.3 Mesin Status Transisi (State Machines)

Sistem DCISP mengoperasikan 12 mesin status transisi eksplisit:

#### 1. Intern Lifecycle State Machine
```text
[ Applicant ] ──(Lolos Seleksi)──> [ Intern ] ──(Selesai Program / Lulus)──> [ Alumni ]
      │                                │
      └──(Ditolak)──> [ Rejected ]      └──(Keluar Dini)──> [ Terminated ]
```

#### 2. Project Application State Machine
```text
[ Applied ] ──(Ditelaah PM)──> [ Under Review ] ──(Masuk Nominasi)──> [ Shortlisted ]
     │                                │                                    │
     ├──(Dibatalkan Pelamar)          ├──(Ditolak PM)                      ├──(Diterima Bergabung)
     ▼                                ▼                                    ▼
[ Withdrawn ]                   [ Rejected ]                         [ Accepted ]
```

#### 3. Work Session State Machine
```text
[ Started ] ──(Fokus Bekerja)──> [ Active ] ──(Jeda/Tidak Aktif > 15m)──> [ Idle ]
     ▲                             │    ▲                                    │
     │                             │    └────────(Kembali Fokus)─────────────┘
     │                             ▼
[ Resumed ] <──(Selesai Istirahat)── [ Paused (Break) ]
     │
     └──(Sesi Kerja Selesai)──> [ Ended ]
```

#### 4. Break State Machine (BR-008, BR-009)
```text
[ WORKING ] ──(Mulai Istirahat / Sesuai Jadwal)──> [ BREAK ] ──(Selesai Istirahat Tepat Waktu)──> [ WORKING ]
     │                                                │
     └──(Istirahat Sebelum Jam Jadwal)                └──(Kembali Terlambat > Toleransi)
                     ▼                                                ▼
           [ Flag: Early Break ]                           [ Flag: Unauthorized Break ]
           (Peringatan Anomali)                            (Memicu Penalti Deduksi XP)
```

#### 5. Attendance Event Lifecycle (FR-008 & OQ-014)
```text
[ ARRIVED ] ──> [ CHECK_IN ] ──> [ WORK_STARTED ] ──> [ BREAK_STARTED ] ──> [ BREAK_ENDED ]
                                                                                  │
                                                                                  ▼
[ CHECK_OUT ] <── [ OVERTIME_ENDED ] <── [ OVERTIME_STARTED ] <── [ WORK_RESUMED ]
```
> **Catatan Inkonsistensi (OQ-014):** Pada WF-001 alur kerja menyertakan event `WORK_ENDED` saat batas jadwal kerja reguler tercapai, namun pada FR-008 dan DATA-001 jenis event `WORK_ENDED` tidak tercantum di dalam katalog enum standar. Ditandai sebagai `[INCONSISTENCY — PERLU RESOLUSI]`.

#### 6. Overtime State Machine (FR-012, BR-013, BR-014)
```text
[ Submitted ] ──(Ulasan Supervisor)──> [ Reviewed ] ──(Ditolak)──> [ Rejected ]
                                            │
                                            └──(Disetujui Jendela Jadwal)──> [ Approved ]
                                                                                  │
                                                                                  ▼
[ Completed / Settled ] <──(Kalkulasi Jam Aktual)── [ Actual Work Recorded ] <──┘
```

#### 7. Attendance Correction State Machine (FR-015, BR-025)
```text
[ Submitted ] ──(Pemeriksaan Bukti)──> [ Under Review ] ──(Ditolak)──> [ Rejected ]
                                             │
                                             └──(Disetujui)──> [ Approved ]
                                                                   │
                                                                   ▼
                                                       [ Recalculation Triggered ]
                                                       (Jejak Log Asli Tetap Utuh)
```

#### 8. Project State Machine (DATA-003)
```text
[ DRAFT ] ──(Publikasi ke Pasar)──> [ PUBLISHED ] ──(Tim Terbentuk & Dimulai)──> [ IN_PROGRESS ]
    │                                     │                                            │
    └──(Dihapus)                          └──(Dibatalkan)                              ├──(Selesai)
           ▼                                     ▼                                     ▼
     [ CANCELLED ]                         [ CANCELLED ]                         [ COMPLETED ]
```

#### 9. Work Report & Submission State Machine (DATA-004)
```text
[ SUBMITTED ] ──(Penelaahan Reviewer)──> [ UNDER_REVIEW ] ──(Perlu Perbaikan)──> [ REVISION_REQUIRED ]
                                                │                                       │
                                                │                                       └──(Ajukan Ulang)
                                                ▼                                                │
                                          [ APPROVED ] <─────────────────────────────────────────┘
```

#### 10. Wallet Transaction State Machine (DATA-006)
```text
[ PENDING ] ──(Validasi Buku Besar Berhasil)──> [ COMPLETED ]
     │
     ├──(Kegagalan Sistem / Dana Tidak Cukup)──> [ FAILED ]
     │
     └──(Pembatalan / Koreksi Keuangan)───────> [ REVERSED ]
```

#### 11. Reward Claim State Machine (FR-033 `[Inferred]`)
```text
[ AVAILABLE ] ──(Intern Mengajukan Klaim)──> [ CLAIM_REQUESTED ] ──(Verifikasi Finance/Admin)──> [ PROCESSING ]
                                                                                                        │
                                                                                                        ▼
                                                                                                  [ FULFILLED ]
```

#### 12. Leave Request State Machine (FR-014 `[Inferred]`)
```text
[ SUBMITTED ] ──(Peninjauan HR / Supervisor)──> [ UNDER_REVIEW ] ──(Ditolak)──> [ REJECTED ]
                                                      │
                                                      └──(Disetujui)──> [ APPROVED ] (0 Penalti XP)
```

### 6.4 Tiga Jenis Pekerjaan (Work Types Partition)

Sistem menerapkan pemisahan ketat terhadap tiga jenis aktivitas kerja *(BR-005)*:
1. **Daily Work (Pekerjaan Harian):** Aktivitas kerja umum selama jam kantor (misal: standup meeting harian, merapikan dokumentasi, troubleshooting bug minor, riset teknologi). Aktivitas ini **TIDAK SELALU** merupakan bagian dari tugas proyek marketplace.
2. **Project Task (Tugas Proyek Terstruktur):** Unit deliverable resmi yang berada di bawah hierarki proyek: $	ext{Project} 	o 	ext{Milestone} 	o 	ext{Task} 	o 	ext{Assignment} 	o 	ext{Submission}$. Tugas ini terhubung dengan perhitungan kontribusi proyek dan insentif bounty.
3. **Work Report (Laporan Kerja):** Pelaporan akuntabilitas formal yang wajib diisi pengguna pada akhir hari atau saat menyerahkan penyelesaian tugas teknis.

### 6.5 Dimensi Waktu Granular (5-Way Time Differentiation)

Untuk memastikan keadilan evaluasi dan mencegah kerancuan metrik kerja, sistem mendefinisikan lima metrik waktu yang saling independen *(BR-007)*:
$$	ext{Attendance Time} 
eq 	ext{Work Session Time} 
eq 	ext{Active Session Time} 
eq 	ext{Break Time} 
eq 	ext{Overtime Time}$$

| Dimensi Waktu | Definisi Operasional | Sumber Pencatatan |
|---|---|---|
| **Attendance Time** | Rentang waktu kotor sejak check-in fisik pertama hingga check-out fisik akhir di kantor | Log `CHECK_IN` hingga `CHECK_OUT` |
| **Work Session Time** | Total durasi sesi kerja resmi sejak pengguna menekan "Start Work" | Log `WORK_STARTED` hingga Sesi Berakhir |
| **Active Session Time** | Waktu nyata saat pengguna berinteraksi aktif pada aplikasi/tugas (di luar jeda idle) | Pelacak fokus/integritas sesi browser |
| **Break Time** | Durasi pengguna berada dalam mode istirahat terjadwal maupun permohonan jeda | Log `BREAK_STARTED` hingga `BREAK_ENDED` |
| **Overtime Time** | Durasi jam kerja aktual yang disetujui di luar jam kerja reguler | Log `OVERTIME_STARTED` hingga `OVERTIME_ENDED` |

### 6.6 Perjalanan Pengguna (User Journeys)

#### 6.6.1 Intern Daily Journey
```text
1. Tiba di Kantor:
   - Intern melakukan tap kartu NFC atau scan QR dinamis pada reader resepsionis.
   - Status terkonfirmasi hadir (Event: CHECK_IN); indikator waktu kedatangan tervalidasi vs toleransi jadwal.
2. Membuka Portal "My Day":
   - Intern membuka laptop, masuk ke aplikasi, melihat widget "My Day" yang menampilkan jadwal hari ini.
   - Intern menekan tombol aksi [START WORK] (Event: WORK_STARTED).
3. Melaksanakan Tugas:
   - Intern memilih tugas aktif dari daftar "Today's Tasks" (kombinasi Daily Work & Project Tasks).
   - Sistem mencatat durasi sesi aktif secara etis (tanpa screenshot).
4. Sesi Istirahat:
   - Pukul 12:00, Intern menekan tombol [BREAK] (Event: BREAK_STARTED); timer hitung mundur 60 menit aktif.
   - Pukul 12:55, Intern menekan tombol [RESUME] (Event: BREAK_ENDED & WORK_RESUMED); presensi istirahat patuh jadwal.
5. Menyerahkan Hasil Kerja & Check-out:
   - Pukul 16:45, Intern menyelesaikan tugas proyek, melampirkan tautan commit Git / berkas bukti pada form Work Report.
   - Pukul 17:00, Intern menekan tombol [END WORK], meninjau ringkasan XP harian yang diperoleh.
   - Melakukan tap kartu keluar pada scanner terminal kantor (Event: CHECK_OUT).
```

#### 6.6.2 Supervisor Daily Journey
```text
1. Membuka Portal "Team Today":
   - Supervisor membuka portal operasional tim, memantau daftar anggota: siapa yang Hadir, Aktif, Istirahat, Lembur, atau Absen.
2. Memeriksa Peringatan Anomali Real-Time:
   - Menerima indikator peringatan: 1 anggota terlambat > 15 menit, 1 pengajuan koreksi presensi masuk.
3. Meninjau & Menyetujui Pengajuan:
   - Membuka Approval Center: menyetujui permohonan Overtime seorang anggota tim yang mengajukan perpanjangan kerja untuk rilis proyek.
   - Meninjau pengajuan koreksi presensi karena kegagalan jaringan scanner pagi hari, lalu menyetujuinya.
4. Menilai Kualitas Deliverable:
   - Meninjau penyerahan milestone proyek tim; memeriksa laporan dan bukti kerja yang telah lolos verifikasi reviewer.
   - Memberikan skor evaluasi kualitas berkala pada modul Performance Evaluation.
   - Menetapkan porsi Final Contribution % anggota tim sebelum klaim pembagian bounty.
```

#### 6.6.3 Project Manager Journey
```text
1. Inisiasi Proyek Baru:
   - PM membuat proyek baru di Project Marketplace: menentukan judul, deskripsi, tanggal batas akhir (deadline), kuota tim (misal 4 orang), dan total alokasi Bounty Pool (misal Rp 10.000.000).
   - Menentukan kriteria keahlian (Required Skills: React, Node.js, UI Design).
   - Menetapkan tingkat visibilitas: "Intern Only" atau "Public".
2. Menyeleksi Pelamar:
   - Membuka daftar lamaran yang masuk; sistem menampilkan indikator Skill Matching % pelamar.
   - Mengubah status pelamar: dari Under Review, Shortlisted, hingga Accepted.
3. Mengatur Milestone & Monitoring:
   - Memecah proyek menjadi 3 Milestone utama dengan bobot kontribusi terukur.
   - Memantau progres penyelesaian tugas dan kecepatan kerja tim secara berkala hingga proyek berstatus Completed.
```

#### 6.6.4 HR & Finance Journeys
- **HR Administrator:** Membuat kohort batch baru, mendaftarkan peserta magang baru dari institusi kampus mitra, mengesahkan kalender kerja dan toleransi keterlambatan, memproses pengajuan cuti, serta menjalankan kalkulasi penentuan SATU Top Performer di akhir siklus evaluasi batch.
- **Finance Administrator:** Memantau akumulasi dana kas perpisahan angkatan (*Batch Fund*), meninjau konfigurasi tarif persentase pemotongan pajak penghasilan, memvalidasi permohonan pencairan saldo dompet (*payout*), dan melakukan audit silang transaksi melalui buku besar immutabel.

### 6.7 Aliran Data Antar Sistem (System Data Flow)

```text
[ Terminal Reader / Scanner ] ──(Credential Ingestion)──> [ Attendance Event Engine ]
                                                                  │
                                       ┌──────────────────────────┴──────────────────────────┐
                                       ▼                                                     ▼
                            [ Work Schedule Engine ]                              [ Work Session Engine ]
                                       │                                                     │
                                       ▼                                                     ▼
                           (Validasi Keterlambatan)                               (Gross / Active Durasi)
                                       │                                                     │
                                       └──────────────────────────┬──────────────────────────┘
                                                                  ▼
                                                       [ Unified Policy Engine ]
                                                                  │
                                       ┌──────────────────────────┴──────────────────────────┐
                                       ▼                                                     ▼
                             [ XP Rules Engine ]                                [ Notification Center ]
                                       │                                                     │
                                       ▼                                                     ▼
                            [ Rank & Rewards Engine ]                              [ User Push Alert ]
                                       │
                                       ▼
                       [ Performance Evaluation Engine ]
                                       │
  [ Project & Tasks ] ──> [ Contribution Engine ]
                                       │
                                       ▼
                          [ Deduction & Tax Engine ]
                                       │
             ┌─────────────────────────┴─────────────────────────┐
             ▼                                                   ▼
   [ Personal Wallet ]                                    [ Batch Fund ]
             │                                                   │
             └─────────────────────────┬─────────────────────────┘
                                       ▼
                         [ Financial Double-Entry Ledger ]
                                       │
                                       ▼
                             [ Cloudflare R2 Store ]
                           (Bukti & Dokumen Audit)
```

---

## 7. Information Architecture

### 7.1 Arsitektur Domain Sistem (System Domain Architecture)

Arsitektur DCISP dibangun di atas 10 domain fungsional modular yang saling terhubung:
```text
DCISP PLATFORM ARCHITECTURE
├── 1. IDENTITY      : Manajemen pengguna, peran dinamis, hak izin, ruang lingkup (scopes), sesi otentikasi.
├── 2. PEOPLE        : Pengelolaan profil peserta magang, alumni, kelompok kohort batch, kampus/institusi, matriks skill.
├── 3. WORKFORCE     : Jadwal kerja, penyerapan log presensi (NFC/QR), pelacak sesi kerja aktif, istirahat, lembur, dan cuti.
├── 4. PROJECTS      : Marketplace proyek, lamaran kerja proyek, pembentukan tim, milestone, tugas, laporan kerja, bukti deliverable.
├── 5. PERFORMANCE   : Mesin perhitungan XP, sistem progresi rank, evaluasi kualitas formal, pencapaian badge, Top Performer.
├── 6. INCENTIVES    : Paket reward rank multi-komponen, alokasi porsi bounty proyek, tata kelola klaim hadiah.
├── 7. FINANCE       : Dompet digital pribadi, kas perpisahan batch fund, mesin deduksi pajak, pencairan dana, buku kas ganda immutabel.
├── 8. DOCUMENTS     : Penerbitan kartu ID ber-QR/NFC, sertifikat digital kelulusan, portofolio digital terverifikasi, logbook resmi.
├── 9. INTELLIGENCE  : Analitik organisasi, pelaporan agregat, activity intelligence, rekomendasi kecerdasan buatan.
└── 10. SYSTEM       : Konfigurasi master policy engine, pusat notifikasi event, pencatatan log audit keamanan, storage Cloudflare R2.
```

### 7.2 Struktur Peta Situs (Sitemap Hierarchy)

```text
DCISP SITEMAP
├── Command Center (Dashboard Eksekutif & Admin)
│
├── People
│   ├── Interns (Daftar & Detail Profil Peserta Magang)
│   ├── Alumni (Direktori Akun & Portofolio Alumni)
│   ├── Batches (Manajemen Kohort Angkatan & Alokasi Dana)
│   ├── Institutions (Daftar Universitas / Sekolah Mitra)
│   └── Skills (Katalog Keahlian & Standar Kemahiran)
│
├── Attendance
│   ├── Live Attendance (Pemantauan Kehadiran Real-Time Kantor)
│   ├── Scanner (Antarmuka Pemindai Khusus Operator Terminal)
│   ├── Attendance Records (Riwayat Lengkap Log Presensi)
│   ├── Work Sessions (Daftar Rekam Sesi Kerja & Durasi Aktif)
│   ├── Corrections (Pusat Pengajuan & Persetujuan Koreksi Presensi)
│   ├── Leave (Manajemen Pengajuan & Saldo Cuti)
│   └── Overtime (Pusat Persetujuan & Rekonsiliasi Jam Lembur)
│
├── Projects
│   ├── Project Marketplace (Papan Bursa Proyek Terbuka & Filter)
│   ├── My Projects (Daftar Proyek yang Sedang Diikuti Pengguna)
│   ├── Applications (Manajemen Lamaran Proyek Masuk & Seleksi)
│   ├── Projects (Manajemen Master Proyek PM)
│   ├── Tasks (Papan Manajemen Tugas / Kanban Board)
│   ├── Milestones (Pelacakan Target Fase Proyek)
│   ├── Submissions (Penyerahan Hasil Tugas & Ulasan Reviewer)
│   └── Reports (Laporan Kerja Harian & Mingguan)
│
├── Performance
│   ├── Performance Overview (Ringkasan Metrik Kinerja Organisasi)
│   ├── Evaluations (Form & Riwayat Penilaian Kinerja Formal)
│   ├── XP (Riwayat Transaksi & Perolehan Poin Pengalaman)
│   ├── Ranks (Tingkatan Status Progresi Karier Gamified)
│   ├── Achievements (Katalog Lencana Prestasi Terbuka & Terkunci)
│   ├── Leaderboard (Papan Peringkat Kompetisi Sehat Global)
│   └── Top Performers (Hall of Fame SATU Juara per Batch)
│
├── Rewards
│   ├── Rank Rewards (Katalog Paket Hadiah Tingkat Rank)
│   ├── Reward Claims (Pusat Pengajuan Klaim Hadiah Pengguna)
│   └── Reward History (Riwayat Distribusi Hadiah Terpenuhi)
│
├── Finance
│   ├── Wallets (Ringkasan Dompet Pribadi Seluruh Pengguna)
│   ├── Project Bounty (Alokasi & Pembagian Kas Imbalan Proyek)
│   ├── Batch Fund (Akumulasi & Penggunaan Kas Perpisahan Angkatan)
│   ├── Tax & Deductions (Konfigurasi Tarif Pajak & Pemotongan Resmi)
│   ├── Payouts (Antrean Permohonan Pencairan Dana & Eksekusi)
│   └── Ledger (Buku Kas Ganda Immutabel Seluruh Peristiwa Moneter)
│
├── Documents
│   ├── ID Cards (Generator & Cetak Kartu Identitas Digital)
│   ├── Certificates (Penerbitan Sertifikat Kelulusan Resmi)
│   ├── Reports (Ekspor Laporan Magang PDF / Logbook Cetak)
│   └── Portfolio (Pembangun Portofolio Publik Alumni Otomatis)
│
├── Media
│   └── Gallery (Dokumentasi Foto/Video Kegiatan & Karya Peserta)
│
├── Insights
│   ├── Analytics (Dasbor Analitik Metrik Operasional Mendalam)
│   ├── AI Insights (Deteksi Pola Anomali & Rekomendasi Pintar)
│   └── Reports (Ekspor Laporan Kinerja Bisnis Berkala)
│
└── System
    ├── Users (Manajemen Kredensial Akun & Status Pengguna)
    ├── Roles & Permissions (Konfigurasi RBAC Dinamis & Izin Granular)
    ├── Notifications (Log Notifikasi Terkirim & Pengaturan Kanal)
    ├── Audit Logs (Buku Jejak Audit Keamanan Sistem Append-Only)
    ├── Security (Pengaturan Otentikasi, Sesi, & Batasan Akses)
    ├── Storage (Monitoring Utilisasi Bucket Cloudflare R2)
    └── Settings (Konfigurasi Global Aplikasi & Parameter Master)
```

### 7.3 Pola Navigasi Antarmuka (Navigation Patterns)

Platform DCISP menggunakan pola navigasi terpadu berbasis antarmuka modern:
1. **Global Collapsible Sidebar:** Terletak di sebelah kiri layar, menyajikan seluruh menu modular dalam hierarki vertikal yang dapat diciutkan (*collapsed*) untuk memaksimalkan ruang kerja.
2. **Top Contextual Header:** Menampilkan rekam status sesi kerja aktif (Live Timer), XP Counter terkini, Rank Badge pengguna, tombol cepat jeda istirahat, tombol lonceng notifikasi real-time, dan menu profil.
3. **Portal Khusus Terfokus (Dedicated Portals):**
   - **Command Center (FR-051):** Tampilan dasbor tingkat tinggi untuk eksekutif/admin dengan 4 widget utama: *TODAY*, *ATTENTION*, *PERFORMANCE*, dan *FINANCE*.
   - **Portal "My Day" (FR-052):** Antarmuka layar tunggal khusus peserta magang/alumni untuk mengoperasikan hari kerjanya: tombol aksi `[START WORK]`/`[END WORK]`, status check-in, timer sesi aktif, pemilih tugas aktif, dan timer hitung mundur istirahat.
   - **Portal "Team Today" (FR-053):** Dasbor operasional langsung bagi supervisor untuk memantau anggota tim bimbingannya secara real-time tanpa navigasi berlapis.
   - **Supervisor Dedicated Portal (FR-054):** Ruang kerja manajemen terpadu bagi supervisor untuk memproses approval lembur, koreksi absensi, evaluasi tugas, dan pengesahan porsi kontribusi.
   - **Scanner Terminal View (FR-001, BR-002):** Antarmuka pemindai presensi berlayar penuh khusus operator terminal (Reihan) yang hanya menampilkan viewport kamera pembaca QR / detektor NFC dan konfirmasi hasil tap terakhir.

### 7.4 Pemetaan Domain Arsitektur ke Modul Navigasi

| Modul Navigasi Utama | Domain Arsitektur Terkait | Sub-modul Fungsional Utama |
|---|---|---|
| **Command Center** | Cross-Domain Reporting | Ringkasan Eksekutif, Widget Atensi Tindakan, Agregasi Keuangan & Kehadiran |
| **People** | PEOPLE | Interns, Alumni, Batches, Institutions, Skills Matrix |
| **Attendance** | WORKFORCE | Live Attendance, Scanner View, Logs, Work Sessions, Corrections, Leave, Overtime |
| **Projects** | PROJECTS | Marketplace, My Projects, Applications, Project Master, Tasks, Milestones, Submissions |
| **Performance** | PERFORMANCE | Performance Overview, Formal Evaluations, XP History, Ranks, Achievements, Leaderboard |
| **Rewards** | INCENTIVES | Rank Reward Catalog, Reward Claim Requests, Fulfillment History |
| **Finance** | FINANCE | User Wallets, Project Bounties, Batch Fund, Tax Engine Rules, Payout Queue, Ledger |
| **Documents** | DOCUMENTS | ID Card Generator, Certificate Engine, PDF Reports, Digital Portfolio Builder |
| **Media** | DOCUMENTS (Storage) | Activity Gallery & Project Showcase |
| **Insights** | INTELLIGENCE | Advanced Analytics, AI Anomaly Insights, Intelligence Reports |
| **System** | IDENTITY & SYSTEM | Users, Dynamic RBAC, Notifications, Immutable Audit Logs, Storage R2, Global Settings |

### 7.5 Matriks Visibilitas Menu Navigasi Berbasis Peran

| Menu Sidebar | Super Admin | Admin | HR Admin | Project Mgr | Supervisor | Reviewer | Finance | Scanner Op | Intern | Alumni |
|---|---|---|---|---|---|---|---|---|---|---|
| **Command Center** | Visible | Visible | Visible | Visible (PM) | Visible (Team) | Hidden | Visible (Fin) | Hidden | My Day Only | My Day Only |
| **People** | Visible | Visible | Visible | Read Only | Team Read | Hidden | Hidden | Hidden | Hidden | Hidden |
| **Attendance** | Visible | Visible | Visible | Hidden | Team Only | Hidden | Hidden | Scanner Only | Own Records | Hidden |
| **Projects** | Visible | Visible | Read Only | Manage | Team Manage | Review Tasks | Hidden | Hidden | View / Apply | Public Market |
| **Performance** | Visible | Visible | Visible | Read Only | Team Eval | Review | Hidden | Hidden | Own XP/Rank | Own Portf/XP |
| **Rewards** | Visible | Visible | Read Only | Hidden | Hidden | Hidden | Process Claims | Hidden | Own Claim | Own Claim |
| **Finance** | Visible | Read Only | Hidden | Hidden | Hidden | Hidden | Manage | Hidden | Own Wallet | Own Wallet |
| **Documents** | Visible | Visible | Visible | Read Only | Team Read | Hidden | Hidden | Hidden | Own Docs | Own Portfolio |
| **Media** | Visible | Visible | Read Only | Read Only | Read Only | Read Only | Read Only | Hidden | Visible | Visible |
| **Insights** | Visible | Visible | Read Only | Read Only | Team Read | Hidden | Read Only | Hidden | Hidden | Hidden |
| **System** | Visible | Read Only | Hidden | Hidden | Hidden | Hidden | Hidden | Hidden | Hidden | Hidden |

> **Catatan Kepatuhan Otorisasi:** Sesuai aturan bisnis BR-001 dan BR-002, seluruh hak visibilitas menu di atas dikendalikan secara dinamis melalui data permission yang tersimpan di basis data, bukan melalui pengkondisian statis di kode front-end. Scanner Operator strictly terisolasi hanya pada menu `Attendance -> Scanner` *(BR-002)*.

---

## 8. Kebutuhan Fungsional (Feature Requirements)

### 8.1 Taksonomi & Ikhtisar Fitur
Seluruh kebutuhan fungsional DCISP dikelompokkan ke dalam 11 domain kapabilitas modular. Setiap fitur berprioritas **P0** didokumentasikan menggunakan spesifikasi formal lengkap 16-elemen. Fitur berprioritas **P1** dan **P2** didokumentasikan dalam format ringkas terstruktur dengan tetap menjaga integritas aturan bisnis.

| ID Fitur | Nama Fitur | Prioritas | Domain | Pemetaan Aturan Bisnis |
|---|---|---|---|---|
| FR-001 | RBAC & Permission Engine | P0 | IDENTITY & RBAC | BR-001, BR-002 |
| FR-002 | Manajemen Siklus Hidup Intern | P0 | PEOPLE | BR-003, BR-004 |
| FR-003 | Siklus Hidup Alumni & Retensi Akun | P0 | PEOPLE | BR-003, BR-004 |
| FR-004 | Manajemen Batch | P0 | PEOPLE | BR-018, BR-021 |
| FR-005 | Manajemen Institusi | P1 | PEOPLE | `[TBD]` |
| FR-006 | Skill Matrix & Profil Skill | P1 | PEOPLE | BR-015 |
| FR-007 | Work Schedule Engine | P0 | WORKFORCE | BR-007, BR-013 |
| FR-008 | Attendance Event Engine & Multi-Method Ingestion | P0 | WORKFORCE | BR-006, BR-007 |
| FR-009 | Work Session Tracking | P0 | WORKFORCE | BR-005, BR-007 |
| FR-010 | Break State Engine | P0 | WORKFORCE | BR-008, BR-009 |
| FR-011 | Session Integrity Tracking (Menjaga Privasi) | P0 | WORKFORCE | BR-010, BR-011 |
| FR-012 | Alur Permintaan & Persetujuan Overtime | P0 | WORKFORCE | BR-013, BR-014 |
| FR-013 | Device Registry | P1 | WORKFORCE | BR-002 |
| FR-014 | Leave Management | P1 | WORKFORCE | BR-012 |
| FR-015 | Attendance Correction & Exception Management | P0 | WORKFORCE | BR-025 |
| FR-016 | Project Marketplace | P0 | PROJECTS & TASKS | BR-003, BR-005 |
| FR-017 | Project Application & Quota Workflow | P0 | PROJECTS & TASKS | BR-015 |
| FR-018 | Skill Matching Engine | P2 | PROJECTS & TASKS | BR-015 |
| FR-019 | Project Team Management | P0 | PROJECTS & TASKS | BR-016 |
| FR-020 | Milestone Management | P1 | PROJECTS & TASKS | BR-005 |
| FR-021 | Task Management | P0 | PROJECTS & TASKS | BR-005 |
| FR-022 | Work Report / Submission | P0 | PROJECTS & TASKS | BR-005 |
| FR-023 | Evidence Management | P0 | PROJECTS & TASKS | BR-024 |
| FR-024 | Three-Layer Contribution Engine | P0 | PROJECTS & TASKS | BR-016 |
| FR-025 | XP Rules Engine | P0 | PERFORMANCE & GAMIFICATION | BR-004, BR-012 |
| FR-026 | Rank Progression System | P0 | PERFORMANCE & GAMIFICATION | BR-017, BR-019 |
| FR-027 | Performance Evaluation Engine | P0 | PERFORMANCE & GAMIFICATION | BR-017, BR-018 |
| FR-028 | Top Performer per Batch | P1 | PERFORMANCE & GAMIFICATION | BR-018 |
| FR-029 | Achievement System | P1 | PERFORMANCE & GAMIFICATION | BR-004 |
| FR-030 | Skill Growth Matrix | P1 | PERFORMANCE & GAMIFICATION | BR-015 |
| FR-031 | Multi-Component Rank Rewards | P0 | INCENTIVES & COMPENSATION | BR-019 |
| FR-032 | Project Bounty Distribution | P0 | INCENTIVES & COMPENSATION | BR-016, BR-020 |
| FR-033 | Reward Claim Management | P1 | INCENTIVES & COMPENSATION | BR-019 |
| FR-034 | Deduction & Tax Engine | P0 | FINANCE & TAXATION | BR-020, BR-021 |
| FR-035 | Personal Wallet & Source Tracking | P0 | FINANCE & TAXATION | BR-022, BR-023 |
| FR-036 | Batch Fund | P1 | FINANCE & TAXATION | BR-021 |
| FR-037 | Financial Ledger | P0 | FINANCE & TAXATION | BR-022, BR-023 |
| FR-038 | Payout Engine | P1 | FINANCE & TAXATION | BR-022 |
| FR-039 | Financial Flow Orchestration | P0 | FINANCE & TAXATION | BR-020, BR-021, BR-022 |
| FR-040 | ID Card Generation | P1 | DOCUMENTS & STORAGE | `[TBD]` |
| FR-041 | Certificate Engine | P1 | DOCUMENTS & STORAGE | `[PERLU KONFIRMASI — OQ-001, OQ-002]` |
| FR-042 | Digital Portfolio Builder | P1 | DOCUMENTS & STORAGE | BR-003 |
| FR-043 | Document Reports Generator | P1 | DOCUMENTS & STORAGE | `[TBD]` |
| FR-044 | Integrasi Cloudflare R2 Object Storage | P0 | DOCUMENTS & STORAGE | BR-024 |
| FR-045 | Unified Policy Engine | P0 | SYSTEM & INTEGRITY | BR-001, BR-012, BR-020 |
| FR-046 | Event-Driven Notification Center | P1 | SYSTEM & INTEGRITY | `[TBD]` |
| FR-047 | Audit Logging & Security | P0 | SYSTEM & INTEGRITY | BR-023, BR-025 |
| FR-048 | System Settings & Configurations | P1 | SYSTEM & INTEGRITY | `[TBD]` |
| FR-049 | Analytics & Reporting | P1 | INTELLIGENCE & INSIGHTS | `[TBD]` |
| FR-050 | AI Insights & Activity Intelligence | P2 | INTELLIGENCE & INSIGHTS | `[TBD — OQ-009]` |
| FR-051 | Command Center (Dashboard Admin / Eksekutif) | P1 | USER INTERFACE & PORTALS | BR-001 |
| FR-052 | Portal "My Day" (Tampilan Harian Intern & Alumni) | P0 | USER INTERFACE & PORTALS | BR-005, BR-007, BR-008 |
| FR-053 | Portal "Team Today" (Tampilan Supervisor) | P1 | USER INTERFACE & PORTALS | BR-007, BR-009 |
| FR-054 | Portal Khusus Supervisor | P1 | USER INTERFACE & PORTALS | BR-013, BR-016 |
| FR-055 | Skema Navigasi Sidebar Global | P0 | USER INTERFACE & PORTALS | BR-001, BR-002 |

---

### 8.2 IDENTITY & RBAC

#### FR-001: RBAC & Permission Engine
- **Feature ID:** FR-001
- **Feature Name:** RBAC & Permission Engine
- **Priority:** P0
- **Actor:** Super Admin, Seluruh Peran Pengguna (Admin, HR Admin, Project Manager, Supervisor, Reviewer, Finance, Scanner Operator, Intern, Alumni)
- **Purpose:** Menyediakan mekanisme otorisasi multi-peran granular berbasis data dinamis tanpa logika peran yang di-hardcode dalam kode sumber aplikasi.
- **Preconditions:** Sistem terinisialisasi dengan basis data RBAC dan token otentikasi JWT pengguna valid.
- **Postconditions:** Keputusan otorisasi (diizinkan atau ditolak) ditegakkan dan dicatat dalam konteks request/audit log.
- **User Story:** Sebagai Super Admin, saya ingin mengonfigurasi peran, hak akses (permissions), dan cakupan konteks (scopes) secara terpusat melalui database/kebijakan, sehingga setiap aktor dalam platform hanya dapat mengakses data dan menjalankan aksi yang sesuai dengan kewenangannya.
- **User Flow:** Request masuk ke API Gateway -> Middleware mengekstrak token dan scope -> Policy engine memvalidasi permission terhadap endpoint/aksi -> Akses diteruskan atau dihentikan dengan HTTP 403.
- **Functional Requirements:**
  1. Mendukung hierarki otorisasi 5 lapis: `Role → Permissions → Scope → Resources → Actions`.
  2. Mendukung 10 peran standar sistem: Super Admin, Admin, HR Admin, Project Manager, Supervisor, Reviewer, Finance, Scanner Operator, Intern, Alumni.
  3. Mendukung isolasi scope kontekstual: `global`, `organization`, `batch`, `project`, `team`, `scanner_only`, `own`.
  4. Menyediakan evaluasi runtime kebijakan hak akses pada setiap panggilan API gateway dan modul backend.
  5. Menegakkan isolasi ketat bagi peran Scanner Operator: cakupan dibatasi hanya pada antarmuka pemindai dan riwayat pemindaian terkini tanpa akses ke modul lain.
- **Business Rules:**
  - BR-001: Roles dan permissions harus dapat dikonfigurasi di database/policy, tidak pernah di-hardcode.
  - BR-002: Scanner Operator Isolation: Role Attendance Operator hanya menerima izin scanner view/scan (`scope = scanner_only`).
- **Validation:**
  - Kombinasi resource + action + scope harus unik dalam tabel permission.
  - Penugasan peran ke pengguna wajib menyertakan scope identifier yang valid jika scope bukan `global`.
- **Error States:**
  - `401 Unauthorized`: Token otentikasi tidak ditemukan, kedaluwarsa, atau tidak valid.
  - `403 Forbidden`: Pengguna memiliki sesi aktif namun tidak memiliki permission atau scope yang memadai untuk resource/action yang diminta.
- **Empty States:**
  - Jika pengguna belum diberikan role apa pun, antarmuka menampilkan status penugasan tertunda (Pending Assignment) dengan opsi kontak administrator.
- **Acceptance Criteria:**
  - **Given** pengguna terotentikasi dengan peran Scanner Operator, **When** pengguna mencoba mengakses endpoint `/api/v1/projects` atau `/api/v1/finance`, **Then** sistem menolak akses dengan kode HTTP `403 Forbidden`.
  - **Given** Super Admin memperbarui hak akses role Supervisor di runtime, **When** Supervisor melakukan request API berikutnya, **Then** hak akses baru langsung berlaku tanpa perlu restart aplikasi atau deployment kode.
- **Dependencies:** Layanan Database RBAC, Middleware Autentikasi JWT.
- **Audit Events:** `ROLE_CREATED`, `ROLE_UPDATED`, `PERMISSION_ASSIGNED`, `USER_ROLE_MAPPED`, `ACCESS_DENIED_EVENT`.
- **Notifications:** Notifikasi keamanan ke Super Admin saat terjadi anomali eskalasi hak akses atau akses tidak sah berulang.
- **Data Entities:** `User`, `Role`, `Permission`, `Scope`, `Audit Log`.

---

### 8.3 PEOPLE & LIFECYCLE MANAGEMENT

#### FR-002: Manajemen Siklus Hidup Intern
- **Feature ID:** FR-002
- **Feature Name:** Manajemen Siklus Hidup Intern
- **Priority:** P0
- **Actor:** HR Admin, Super Admin, Intern
- **Purpose:** Mengelola transisi status peserta magang secara formal dari kandidat hingga kelulusan menjadi alumni dengan riwayat lengkap.
- **Preconditions:** Data pelamar magang (Applicant) telah terverifikasi oleh HR dan batch magang aktif tersedia.
- **Postconditions:** Status pengguna bertransisi menjadi Intern aktif dengan hak akses modul magang dan jadwal kerja.
- **User Story:** Sebagai HR Admin, saya ingin mendaftarkan, mengaktifkan, menugaskan batch, memantau, dan meluluskan intern, sehingga rekam jejak formal peserta magang tercatat secara utuh dan terhubung ke modul operasional lainnya.
- **User Flow:** HR meninjau profil kandidat -> HR menetapkan ke Batch tertentu -> Mengubah status menjadi Intern -> Kredensial akun dikirimkan -> Peserta dapat masuk ke portal My Day.
- **Functional Requirements:**
  1. Mengelola transisi status intern mengikuti mesin status: `APPLICANT → ONBOARDING → ACTIVE → ON_LEAVE → SUSPENDED → GRADUATED → TERMINATED`.
  2. Mengaitkan profil intern dengan data institusi asal, nomor induk (NIM/NISN), batch penugasan, dan supervisor penanggung jawab.
  3. Menginisialisasi personal wallet, profil XP (Internship XP), dan jadwal kerja default saat intern mencapai status `ACTIVE`.
  4. Mentransisikan status intern menjadi `GRADUATED` dan secara otomatis memicu pembuatan entitas `Alumni` tanpa menghapus akun pengguna.
- **Business Rules:**
  - BR-003: Retensi akun permanen: Akun peserta magang yang lulus tidak boleh dihapus atau dinonaktifkan secara permanen.
- **Validation:**
  - Wajib memiliki data batch dan institusi yang valid sebelum status diubah menjadi `ACTIVE`.
  - Tanggal mulai magang tidak boleh lebih besar dari tanggal perkiraan selesai magang.
- **Error States:**
  - `422 Unprocessable Entity`: Transisi status tidak sah (misal dari `APPLICANT` langsung ke `GRADUATED`).
  - `409 Conflict`: Intern sudah terdaftar aktif pada batch lain yang berjalan bersamaan.
- **Empty States:**
  - Direktori intern menampilkan ilustrasi informatif "Belum ada data peserta magang pada batch terpilih" dengan tombol aksi tambah/impor data.
- **Acceptance Criteria:**
  - **Given** intern dengan status `ACTIVE` telah menyelesaikan seluruh kewajiban program, **When** HR Admin menyetujui kelulusan, **Then** status intern berubah menjadi `GRADUATED`, entitas Alumni dibuat, dan hak akses akun dialihkan ke peran Alumni.
- **Dependencies:** FR-001 (RBAC), FR-004 (Batch Management), FR-005 (Institution Management).
- **Audit Events:** `INTERN_REGISTERED`, `INTERN_STATUS_TRANSITIONED`, `INTERN_BATCH_ASSIGNED`, `INTERN_GRADUATED`.
- **Notifications:** Notifikasi selamat datang ke Intern baru, notifikasi kelulusan ke Intern dan Supervisor.
- **Data Entities:** `User`, `Intern`, `Alumni`, `Batch`, `Institution`, `Wallet`.

---

#### FR-003: Siklus Hidup Alumni & Retensi Akun
- **Feature ID:** FR-003
- **Feature Name:** Siklus Hidup Alumni & Retensi Akun
- **Priority:** P0
- **Actor:** Alumni, Project Manager, Admin
- **Purpose:** Mempertahankan akun pengguna yang telah lulus tanpa batas waktu agar dapat berpartisipasi dalam proyek publik, menghasilkan bounty, dan membangun portofolio profesional.
- **Preconditions:** Peserta magang telah menyelesaikan masa magang resmi dan evaluasi akhir disahkan.
- **Postconditions:** Peran bertransisi menjadi Alumni secara permanen dengan hak akses marketplace proyek publik.
- **User Story:** Sebagai Alumni, saya ingin tetap dapat login ke platform menggunakan kredensial saya, melihat serta melamar proyek publik, dan menerima kompensasi bounty, sehingga hubungan profesional dengan Dagang Creative tetap berlanjut pasca-magang.
- **User Flow:** HR/Sistem memproses kelulusan -> Akun dialihkan ke peran Alumni -> Hak akses dibatasi ke marketplace publik, portofolio, dan dompet -> Akun tetap aktif tanpa batas waktu.
- **Functional Requirements:**
  1. Menjaga persistensi akun alumni tanpa batas waktu kedaluwarsa akun (akun permanen aktif).
  2. Mengizinkan Alumni melihat marketplace proyek publik, mengajukan lamaran proyek publik, mengerjakan tugas yang ditugaskan, dan mengunggah laporan kerja.
  3. Memisahkan pembukuan XP alumni ke dalam jalur khusus `Alumni Contribution` / `Project XP`.
  4. Mengisolasi leaderboard dan rank: XP atau Rank yang diperoleh alumni dilarang memengaruhi atau mengubah standing rank peserta magang aktif.
  5. Memberikan akses ke builder portofolio digital dan riwayat pencapaian selama dan setelah magang.
- **Business Rules:**
  - BR-003: Alumni Account Permanence: Alumni tetap memiliki akun dan dapat mengerjakan proyek di marketplace publik.
  - BR-004: XP Scheme Separation: 3 jalur XP berbeda (Internship XP, Project XP, Alumni Contribution). XP Alumni tidak mengubah rank magang aktif.
- **Validation:**
  - Alumni hanya dapat melamar proyek dengan tingkat visibilitas `PUBLIC`.
  - Upaya alumni mengakses fitur absensi harian kantor aktif (`Live Attendance` harian intern) harus diblokir oleh otorisasi role.
- **Error States:**
  - `403 Forbidden`: Alumni mencoba melamar proyek berlabel `INTERN_ONLY` atau `PRIVATE` (tanpa undangan).
- **Empty States:**
  - Daftar proyek publik kosong menampilkan pesan: "Saat ini belum ada proyek publik yang tersedia untuk alumni. Silakan periksa kembali nanti."
- **Acceptance Criteria:**
  - **Given** pengguna berstatus Alumni terotentikasi, **When** alumni menyelesaikan task pada proyek publik yang disetujui, **Then** sistem mencatat XP pada skema `Alumni Contribution` dan mengkreditkan bounty ke wallet pribadi tanpa mengubah klasifikasi ranking magang aktif batch berjalan.
- **Dependencies:** FR-001 (RBAC), FR-002 (Intern Lifecycle), FR-016 (Project Marketplace), FR-025 (XP Rules Engine).
- **Audit Events:** `ALUMNI_LOGIN`, `ALUMNI_PROJECT_APPLIED`, `ALUMNI_XP_CREDITED`, `ALUMNI_PORTFOLIO_EXPORTED`.
- **Notifications:** Notifikasi proyek publik baru yang relevan dengan keahlian alumni, notifikasi penerimaan aplikasi proyek.
- **Data Entities:** `User`, `Alumni`, `Project`, `Task`, `XP Transaction`, `Portfolio`.

---

#### FR-004: Manajemen Batch
- **Feature ID:** FR-004
- **Feature Name:** Manajemen Batch
- **Priority:** P0
- **Actor:** HR Admin, Super Admin, Finance
- **Purpose:** Mengelompokkan peserta magang ke dalam kohort terstruktur untuk tata kelola jadwal, evaluasi performa kohort, dan pengelolaan Batch Fund.
- **Preconditions:** Super Admin atau HR Admin memiliki otorisasi pembuatan kohort baru.
- **Postconditions:** Entitas Batch aktif terbentuk dengan alokasi Batch Fund terisolasi dan kalender program.
- **User Story:** Sebagai HR Admin, saya ingin membuat dan mengelola kohort batch dengan rentang tanggal dan aturan spesifik, sehingga evaluasi performa Top Performer dan pengelolaan dana kebersamaan (Farewell Fund) dapat terisolasi per kohort.
- **User Flow:** HR memasukkan nama batch, tanggal mulai, tanggal selesai, dan target kuota -> Sistem membuat record Batch -> Batch siap menerima alokasi peserta.
- **Functional Requirements:**
  1. Membuat entitas batch baru dengan atribut: kode batch, nama, tanggal mulai, tanggal selesai, kapasitas kuota, dan status (`DRAFT`, `ACTIVE`, `COMPLETED`, `ARCHIVED`).
  2. Menghubungkan peserta magang ke tepat satu batch aktif selama masa magang.
  3. Menginisialisasi rekening penampung `Batch Fund` otomatis untuk setiap batch baru.
  4. Menyediakan kalkulasi agregat kohort untuk menentukan tepat satu `Top Performer per Batch` pada setiap siklus evaluasi.
- **Business Rules:**
  - BR-018: Singular Top Performer per Batch: Tepat SATU Top Performer diakui per kohort batch per periode evaluasi.
  - BR-021: Farewell Fund Classification: Alokasi dana perpisahan diisolasi per entitas Batch Fund.
- **Validation:**
  - Kode batch harus unik (misal: `BATCH-2026-01`).
  - Tanggal selesai batch harus setelah tanggal mulai batch.
- **Error States:**
  - `409 Conflict`: Kode batch duplikat.
  - `422 Unprocessable Entity`: Upaya menutup batch saat masih ada intern dengan status aktif yang belum ditransisikan.
- **Empty States:**
  - Tampilan daftar batch kosong jika sistem baru diinisialisasi dengan panduan pembentukan batch perdana.
- **Acceptance Criteria:**
  - **Given** batch baru dibuat oleh HR Admin, **When** batch disimpan, **Then** sistem otomatis membuat entitas `Batch Fund` terkait dengan saldo awal Rp0 dan status batch menjadi `DRAFT`.
- **Dependencies:** FR-001 (RBAC).
- **Audit Events:** `BATCH_CREATED`, `BATCH_ACTIVATED`, `BATCH_COMPLETED`, `BATCH_FUND_INITIALIZED`.
- **Notifications:** Notifikasi aktivasi batch kepada supervisor dan admin operasional.
- **Data Entities:** `Batch`, `Intern`, `Batch Fund`, `Performance Evaluation`.

---

#### FR-005: Manajemen Institusi (Prioritas: P1)
- **Aktor:** HR Admin, Admin.
- **Tujuan:** Mengelola data master sekolah/universitas mitra magang untuk pelaporan kemitraan dan analitik distribusi peserta.
- **Kebutuhan Utama:** CRUD master institusi (Nama, Alamat, Kontak Koordinator/Dosen Pembimbing, MoU/Perjanjian Kemitraan). Pengelompokan data intern berdasarkan institusi asal.
- **Aturan Bisnis & Catatan:** Detail field spesifik institusi bertanda `[DETAIL FIELD TIDAK DISEBUTKAN DALAM SUMBER / TBD]`. Menghapus institusi dilarang jika masih memiliki intern terhubung.
- **Entitas:** `Institution`, `Intern`.

---

#### FR-006: Skill Matrix & Profil Skill (Prioritas: P1)
- **Aktor:** Intern, Alumni, Project Manager, Admin.
- **Tujuan:** Menyediakan repositori keahlian terstandarisasi untuk pencocokan proyek dan pemantauan perkembangan kompetensi peserta magang.
- **Kebutuhan Utama:** Master data skill berjenjang (kategori, nama keahlian, tingkat kemahiran: Pemula, Menengah, Mahir). Pengguna dapat memilih skill pada profil pribadi; sistem mencatat riwayat pemanfaatan skill pada tugas proyek yang terselesaikan.
- **Aturan Bisnis & Catatan:** Digunakan sebagai data masukan bagi mesin pencocokan proyek (FR-018). Bersifat advisory bagi penilai (BR-015).
- **Entitas:** `Skill`, `User`, `Project`, `Task`.

---

### 8.4 WORKFORCE & ATTENDANCE MANAGEMENT

#### FR-007: Work Schedule Engine
- **Feature ID:** FR-007
- **Feature Name:** Work Schedule Engine
- **Priority:** P0
- **Actor:** HR Admin, Super Admin, System Engine
- **Purpose:** Menentukan jadwal kerja acuan yang menjadi dasar perhitungan keterlambatan, pulang awal, jam lembur, dan pelanggaran kehadiran lainnya.
- **Preconditions:** Kebijakan jam kerja kantor dan kalender hari libur resmi telah dikonfigurasi.
- **Postconditions:** Jadwal kerja harian aktif menjadi acuan resmi perhitungan keterlambatan dan kepatuhan istirahat.
- **User Story:** Sebagai HR Admin, saya ingin menetapkan jadwal kerja harian, jendela istirahat terjadwal, batas toleransi (grace period), dan kalender libur, sehingga sistem dapat memvalidasi log kehadiran secara otomatis dan objektif.
- **User Flow:** Admin menentukan hari kerja, jam mulai (08:30), jam selesai (17:00), jeda istirahat (12:00-13:00), dan grace period (10 menit) -> Jadwal diterapkan ke batch/organisasi.
- **Functional Requirements:**
  1. Mendefinisikan konfigurasi jadwal kerja dengan atribut: nama jadwal, hari kerja aktif (array integer 1-5), jam mulai (`start_time`), jam selesai (`end_time`), awal istirahat (`break_start`), akhir istirahat (`break_end`), batas toleransi terlambat (`grace_period_minutes`, default: 10 menit), tautan kebijakan lembur, dan kalender libur.
  2. Menyediakan kalkulasi status keterlambatan: `ON_TIME` jika kedatangan $\le \text{start\_time} + \text{grace\_period}$; `LATE` jika kedatangan $> \text{start\_time} + \text{grace\_period}$.
  3. Mengklasifikasikan durasi keterlambatan ke dalam kategori penalti: 1–15 menit, 16–30 menit, dan >30 menit.
  4. Menyediakan validasi kalender libur resmi: hari libur membebaskan kewajiban check-in tanpa pinalti absensi.
- **Business Rules:**
  - BR-007: 5-Way Time Differentiation: Jadwal kerja menjadi baseline untuk membedakan Attendance Time, Work Session Time, Active Time, Break Time, dan Overtime.
  - BR-013: Schedule-Governed Overtime: Jam kerja di luar jadwal inti wajib memiliki persetujuan lembur formal.
- **Validation:**
  - `start_time` harus lebih awal dari `end_time`.
  - `break_start` dan `break_end` harus berada di antara `start_time` dan `end_time`.
  - `grace_period_minutes` minimal 0 menit dan maksimal 60 menit.
- **Error States:**
  - `422 Unprocessable Entity`: Rentang waktu jadwal tumpang tindih atau konfigurasi hari kerja tidak valid.
- **Empty States:**
  - Jika jadwal belum dikonfigurasi, sistem memuat jadwal default perusahaan (Senin-Jumat, 08:30–17:00, Break 12:00–13:00, Grace Period 10 menit).
- **Acceptance Criteria:**
  - **Given** jadwal kerja dengan `start_time = 08:30` dan `grace_period_minutes = 10`, **When** intern melakukan check-in pada pukul 08:39, **Then** sistem menandai kehadiran sebagai `ON_TIME`.
  - **Given** jadwal kerja yang sama, **When** intern melakukan check-in pada pukul 08:42, **Then** sistem menandai kehadiran sebagai `LATE` (keterlambatan 12 menit) dan memicu evaluasi penalti XP level 1.
- **Dependencies:** FR-045 (Unified Policy Engine).
- **Audit Events:** `SCHEDULE_CREATED`, `SCHEDULE_UPDATED`, `HOLIDAY_ADDED`.
- **Notifications:** Notifikasi pembaruan jadwal kerja ke seluruh pengguna terhubung.
- **Data Entities:** `Work Schedule` (DATA-002), `Holiday`, `Policy`.

---

#### FR-008: Attendance Event Engine & Multi-Method Ingestion
- **Feature ID:** FR-008
- **Feature Name:** Attendance Event Engine & Multi-Method Ingestion
- **Priority:** P0
- **Actor:** Intern, Scanner Operator, Supervisor, Admin, Hardware Scanner
- **Purpose:** Menelan (ingest), memvalidasi, dan mencatat setiap peristiwa kehadiran dari berbagai kanal masukan ke dalam satu sumber kebenaran data yang tidak dapat diubah (append-only).
- **Preconditions:** Terminal scanner aktif, perangkat terdaftar di Device Registry, dan pengguna memiliki kartu NFC atau QR Code valid.
- **Postconditions:** Record Attendance Event Log tersimpan secara append-only dan memicu sinkronisasi kehadiran live.
- **User Story:** Sebagai Intern atau Scanner Operator, saya ingin merekam peristiwa check-in/out melalui QR Code atau kartu NFC fisik pada terminal scanner, sehingga catatan kehadiran saya tersimpan secara real-time dan terverifikasi.
- **User Flow:** Pengguna menempelkan kartu NFC / memindai QR pada reader -> Device Gateway mengirim payload bertanda tangan -> Event Engine memvalidasi jadwal -> Event tercatat -> Notifikasi sukses.
- **Functional Requirements:**
  1. Menerima data log kehadiran melalui ingestion multi-metode: QR Code dinamis, Kartu NFC, Admin/Operator Scanner, dan Pengajuan Koreksi Manual yang disetujui.
  2. Menyediakan kesiapan arsitektur integrasi masa depan: Face Recognition, Geolocation, dan Device Verification.
  3. Mendukung tipe event log resmi: `ARRIVED`, `CHECK_IN`, `WORK_STARTED`, `BREAK_STARTED`, `BREAK_ENDED`, `WORK_RESUMED`, `OVERTIME_STARTED`, `OVERTIME_ENDED`, `CHECK_OUT`.  
     *(Catatan inkonsistensi: Penanganan event `WORK_ENDED` dari alur WF-001 tercatat pada OQ-014; jika jadwal selesai tercapai tanpa lembur, status transisi menuju checkout)*.
  4. Merekam payload event record lengkap: `id` (UUID), `user_id` (UUID), `event_type` (Enum), `timestamp` (with timezone), `method` (Enum), `device_id` (String), `location_context` (String/JSON), `session_id` (UUID, nullable), `metadata` (JSONB).
  5. Memastikan prinsip idempotensi dan pencegahan pemindaian ganda (duplicate scan debouncing) dalam rentang toleransi waktu (misal: 30 detik).
- **Business Rules:**
  - BR-006: Central Attendance Ingestion: Semua metode kehadiran mengarah ke SATU Attendance Engine terpusat.
  - BR-007: 5-Way Time Differentiation: Attendance event menandai awal/akhir keberadaan di tempat kerja.
- **Validation:**
  - Pengguna harus berstatus `ACTIVE`.
  - Token QR Code harus memiliki tanda tangan kriptografis valid dan belum kedaluwarsa (TTL berlaku).
  - Kartu NFC harus terdaftar pada registry perangkat pengguna terkait.
- **Error States:**
  - `400 Bad Request`: Format payload event tidak sesuai skema.
  - `401 Unauthorized`: Token QR kedaluwarsa atau signature tidak valid.
  - `404 Not Found`: UID kartu NFC tidak terdaftar pada akun mana pun.
  - `409 Conflict`: Percobaan pemindaian ganda dalam jendela debouncing terdeteksi.
- **Empty States:**
  - Riwayat kehadiran hari ini menampilkan status "Belum ada rekaman kehadiran hari ini" dengan tombol scan atau info jadwal.
- **Acceptance Criteria:**
  - **Given** kartu NFC valid milik intern ditempelkan pada reader scanner operator, **When** payload diterima Attendance Engine, **Then** satu record baru tersimpan di tabel `Attendance Event Log` dengan event `CHECK_IN`, metode `NFC`, dan timestamp server saat itu.
- **Dependencies:** FR-001 (RBAC), FR-007 (Work Schedule Engine), FR-013 (Device Registry).
- **Audit Events:** `ATTENDANCE_EVENT_INGESTED`, `ATTENDANCE_SCAN_FAILED`, `DUPLICATE_SCAN_REJECTED`.
- **Notifications:** Notifikasi konfirmasi check-in berhasil ke aplikasi intern; notifikasi ke supervisor jika intern tidak hadir tanpa kabar.
- **Data Entities:** `Attendance Event Log` (DATA-001), `Device`, `User`, `Work Schedule`.

---

#### FR-009: Work Session Tracking
- **Feature ID:** FR-009
- **Feature Name:** Work Session Tracking
- **Priority:** P0
- **Actor:** Intern, Alumni, System Engine
- **Purpose:** Mengukur durasi kerja aktual dan membedakan antara keberadaan fisik (attendance) dengan eksekusi kerja nyata (work session).
- **Preconditions:** Peserta telah melakukan CHECK_IN pada hari yang sama.
- **Postconditions:** Sesi kerja resmi dimulai (WORK_STARTED) dan timer durasi sesi aktif berjalan.
- **User Story:** Sebagai Intern, saya ingin memulai, menjeda, dan mengakhiri sesi kerja saya melalui portal "My Day", sehingga durasi jam kerja produktif saya tercatat secara transparan dan akurat.
- **User Flow:** Intern membuka portal My Day -> Menekan tombol [START WORK] -> Sistem mencatat timestamp mulai -> Sesi aktif dengan pemantauan visibilitas tab etis.
- **Functional Requirements:**
  1. Mengelola transisi status sesi kerja: `IDLE → WORKING → BREAK → RESUMED → ENDED`.
  2. Menghitung dan menyimpan 3 metrik durasi waktu terpisah:
     - **Gross Session Duration:** Total rentang waktu dari mulai kerja hingga sesi selesai.
     - **Active Session Duration:** Waktu aktual bekerja di luar waktu istirahat dan idle tidak wajar.
     - **Idle Duration:** Waktu di mana pengguna terdeteksi tidak aktif melebihi batas toleransi.
  3. Mengaitkan sesi kerja dengan tugas proyek (`Task`) atau pekerjaan rutin harian (`Daily Work`) yang sedang dikerjakan.
  4. Menegakkan aturan bahwa Active Session Duration bukan merupakan skor produktivitas absolut, melainkan indikator ketersediaan kerja.
- **Business Rules:**
  - BR-005: Work Type Partition: Daily Work, Project Tasks, dan Work Reports adalah entitas terpisah.
  - BR-007: 5-Way Time Differentiation: Work Session Duration $\neq$ Attendance Presence Duration.
  - Formula: $\text{Active Session} \neq \text{Productivity Score}$.
- **Validation:**
  - Sesi kerja hanya dapat dimulai (`WORK_STARTED`) setelah pengguna melakukan `CHECK_IN` kehadiran pada hari yang sama.
  - Tidak boleh ada lebih dari satu sesi kerja aktif berjalan bersamaan untuk satu pengguna.
- **Error States:**
  - `400 Bad Request`: Mencoba memulai sesi kerja sebelum melakukan check-in kehadiran.
  - `409 Conflict`: Sesi kerja sebelumnya masih berstatus aktif dan belum ditutup.
- **Empty States:**
  - Tampilan sesi kerja pada widget My Day menampilkan kartu kosong: "Anda belum memulai sesi kerja hari ini. Klik tombol [START WORK] untuk mulai."
- **Acceptance Criteria:**
  - **Given** intern telah check-in, **When** intern menekan tombol [START WORK] pada My Day, **Then** sesi kerja baru dibuat dengan status `WORKING`, dan event `WORK_STARTED` tercatat di ledger kehadiran.
- **Dependencies:** FR-008 (Attendance Event Engine), FR-052 (Portal My Day).
- **Audit Events:** `WORK_SESSION_STARTED`, `WORK_SESSION_PAUSED`, `WORK_SESSION_RESUMED`, `WORK_SESSION_ENDED`.
- **Notifications:** Pengingat otomatis di browser/aplikasi jika intern telah check-in lebih dari 30 menit namun belum menekan [START WORK].
- **Data Entities:** `Work Session`, `Attendance Event Log` (DATA-001), `Task`, `User`.

---

#### FR-010: Break State Engine
- **Feature ID:** FR-010
- **Feature Name:** Break State Engine
- **Priority:** P0
- **Actor:** Intern, Alumni, System Engine
- **Purpose:** Menerapkan manajemen status istirahat eksplisit untuk menjaga akurasi perhitungan jam kerja produktif dan mendeteksi anomali istirahat.
- **Preconditions:** Sesi kerja sedang berstatus ACTIVE pada portal My Day.
- **Postconditions:** Status kerja bertransisi menjadi BREAK; timer istirahat berjalan dan active timer berhenti sementara.
- **User Story:** Sebagai Intern, saya ingin mengambil jeda istirahat dengan menekan tombol istirahat dan melanjutkannya kembali, sehingga jam istirahat saya tidak terhitung sebagai jam kerja aktif dan sesuai dengan jadwal yang ditentukan.
- **User Flow:** Intern menekan tombol [BREAK] -> Status berubah ke BREAK -> Hitung mundur jeda aktif -> Intern menekan [RESUME] -> Status kembali ke WORKING.
- **Functional Requirements:**
  1. Menerapkan mesin status istirahat eksplisit: `WORKING → BREAK → WORKING`.
  2. Merekam event `BREAK_STARTED` saat intern memulai jeda dan `BREAK_ENDED` bersamaan dengan `WORK_RESUMED` saat kembali bekerja.
  3. Membandingkan waktu istirahat aktual terhadap jendela istirahat terjadwal (`break_start` dan `break_end` pada Work Schedule).
  4. Mendeteksi dan menandai dua anomali istirahat utama:
     - **Early Break Anomaly:** Istirahat dimulai sebelum jendela terjadwal (misal: pukul 11:42 vs jadwal 12:00) $\rightarrow$ sistem menandai flag anomali.
     - **Late Resume Violation (Unauthorized Break):** Melanjutkan kerja melampaui batas akhir jendela istirahat (misal: pukul 13:18 vs jadwal 13:00) $\rightarrow$ sistem mencatat pelanggaran unauthorized break dan memicu evaluasi penalti XP.
- **Business Rules:**
  - BR-008: Work Session State Machine: Break adalah status transisi eksplisit dalam sesi kerja.
  - BR-009: Break Anomaly Enforcement: Early break memicu flag anomali; resume terlambat memicu penalti pelanggaran absensi.
- **Validation:**
  - Aksi istirahat hanya dapat dilakukan jika sesi kerja saat ini berada pada status `WORKING`.
  - Durasi istirahat tidak boleh negatif.
- **Error States:**
  - `400 Bad Request`: Mencoba memulai break saat sesi kerja tidak berstatus `WORKING`.
- **Empty States:**
  - Timer break pada UI menampilkan status "Tidak dalam masa istirahat" saat pengguna sedang aktif bekerja.
- **Acceptance Criteria:**
  - **Given** intern sedang dalam status `WORKING` dan waktu menunjukkan pukul 13:18 (jadwal istirahat berakhir 13:00), **When** intern menekan tombol [RESUME WORK], **Then** sistem mencatat `BREAK_ENDED`, menghitung kelebihan istirahat 18 menit, menandai flag `UNAUTHORIZED_BREAK`, dan menerapkan penalti -2 XP.
- **Dependencies:** FR-007 (Work Schedule Engine), FR-009 (Work Session Tracking), FR-025 (XP Rules Engine).
- **Audit Events:** `BREAK_STARTED_EVENT`, `BREAK_ENDED_EVENT`, `BREAK_ANOMALY_DETECTED`, `UNAUTHORIZED_BREAK_PENALIZED`.
- **Notifications:** Notifikasi peringatan 5 menit sebelum batas waktu istirahat berakhir terkirim ke intern.
- **Data Entities:** `Break`, `Work Session`, `Attendance Event Log` (DATA-001), `XP Transaction`.

---

#### FR-011: Session Integrity Tracking (Menjaga Privasi)
- **Feature ID:** FR-011
- **Feature Name:** Session Integrity Tracking (Menjaga Privasi)
- **Priority:** P0
- **Actor:** Intern, Alumni, System Client Engine
- **Purpose:** Memantau integritas keaktifan sesi kerja berbasis browser secara non-intrusif tanpa melanggar privasi pengguna.
- **Preconditions:** Sesi kerja aktif berjalan pada peramban web pengguna.
- **Postconditions:** Data metrik durasi fokus/idle diperbarui secara berkala tanpa pelanggaran privasi pengguna.
- **User Story:** Sebagai Pengembang Sistem dan Intern, saya ingin sistem memantau fokus tab browser saya secara objektif saat sesi kerja berlangsung tanpa menggunakan software pengintai yang melanggar privasi saya.
- **User Flow:** Pengguna memindahkan fokus peramban -> Event visibilitychange/blur mencatat jeda fokus -> Jika idle > 15 menit, sistem menandai status idle.
- **Functional Requirements:**
  1. Menangkap status visibilitas tab dan fokus jendela browser selama sesi kerja aktif: `Started`, `Active`, `Tab Hidden`, `Window Blur`, `Idle`, `Resumed`, `Ended`.
  2. Menerapkan ambang batas konfigurasi integritas sesi:
     - Deteksi **Idle Timer:** Pengguna tidak melakukan interaksi mouse/keyboard selama >15 menit $\rightarrow$ status sesi ditandai sebagai `Idle`.
     - Peringatan **Tab Inactive:** Pengguna berpindah tab melebihi ambang batas tertentu ($X$ kali dalam satu jam) $\rightarrow$ sistem memicu peringatan halus pada antarmuka.
     - **Anomali Berulang:** Durasi idle yang ekstrim dan berulang dicatat sebagai anomali sesi kerja untuk tinjauan supervisor.
  3. Menegakkan batasan privasi mutlak pada arsitektur sistem:
     - DILARANG merekam layar (NO screenshot).
     - DILARANG merekam ketukan papan tik (NO keylogger).
     - DILARANG menginspeksi isi konten atau URL tab lain (NO tab content inspection).
     - DILARANG memindai aplikasi latar belakang atau desktop pengguna (NO desktop process scanning).
     - DILARANG menyalin isi papan klip (NO clipboard recording).
     - DILARANG membaca percakapan pesan instan pribadi pengguna (NO chat reading).
  4. Menegakkan aturan bahwa perpindahan tab (tab switching/blur) tidak secara otomatis memotong XP secara langsung.
- **Business Rules:**
  - BR-010: Privacy-First Tracking: Activity tracking hanya memantau fokus/visibilitas window dan idle browser. Alat surveilans invasif dilarang mutlak.
  - BR-011: Non-Punitive Tab Blur: Tab blur/switching tidak secara otomatis menimbulkan deduksi XP instan.
- **Validation:**
  - Script pelacak hanya aktif saat status sesi kerja bernilai `WORKING`. Saat status `BREAK` atau `ENDED`, seluruh pendengar event browser dinonaktifkan.
- **Error States:**
  - `400 Bad Request`: Paket heartbeat integritas sesi yang dikirim memiliki format aneh atau timestamp tidak sinkron.
- **Empty States:**
  - Log anomali kosong menampilkan indikator: "Integritas sesi berjalan dengan baik, tidak ada anomali terdeteksi."
- **Acceptance Criteria:**
  - **Given** intern sedang bekerja dalam sesi aktif, **When** intern berpindah ke tab referensi kerja lain selama 10 menit, **Then** sistem mencatat durasi visibilitas berkurang namun TIDAK memotong XP pengguna secara otomatis.
- **Dependencies:** FR-009 (Work Session Tracking).
- **Audit Events:** `INTEGRITY_IDLE_TRIGGERED`, `INTEGRITY_ANOMALY_FLAGGED`.
- **Notifications:** Dialog konfirmasi aktivitas muncul di layar pengguna saat idle mencapai 15 menit: "Apakah Anda masih bekerja?".
- **Data Entities:** `Work Session`, `Audit Log`.

---

#### FR-012: Alur Permintaan & Persetujuan Overtime
- **Feature ID:** FR-012
- **Feature Name:** Alur Permintaan & Persetujuan Overtime
- **Priority:** P0
- **Actor:** Intern, Supervisor, Admin, Finance
- **Purpose:** Mengelola pengajuan, peninjauan, dan eksekusi jam lembur resmi yang terikat pada jadwal kerja dan proyek.
- **Preconditions:** Peserta magang memiliki kebutuhan kerja tambahan di luar jam reguler untuk tugas/proyek spesifik.
- **Postconditions:** Jendela lembur terdaftar pada jadwal kerja dan jam aktual yang dikerjakan diperhitungkan untuk bonus kompensasi/XP.
- **User Story:** Sebagai Intern, saya ingin mengajukan izin lembur dengan alasan dan tugas yang jelas, dan sebagai Supervisor, saya ingin meninjau pengajuan tersebut, sehingga jam lembur yang disetujui dapat dieksekusi dan menghasilkan bonus kompensasi yang sah.
- **User Flow:** Intern mengisi form pengajuan OT (tanggal, jam, alasan, tugas) -> Supervisor meninjau di Team Today -> Supervisor menyetujui/menolak -> Intern bekerja pada jendela disetujui -> Sistem menghitung jam aktual.
- **Functional Requirements:**
  1. Menyediakan formulir pengajuan lembur bagi intern dengan payload: Tanggal lembur, Jendela waktu yang diminta (`requested_start`, `requested_end`), Justifikasi/Alasan tertulis, dan Referensi Proyek/Tugas (`project_id`/`task_id`).
  2. Menyediakan antarmuka peninjauan persetujuan lembur bagi Supervisor pada portal Team Today: aksi `APPROVE` atau `REJECT` disertai catatan peninjauan.
  3. Mendaftarkan jendela lembur yang disetujui (`approved_window`) ke dalam Work Schedule Engine.
  4. Menangkap pelaksanaan lembur aktual pengguna melalui event `OVERTIME_STARTED` dan `OVERTIME_ENDED`.
  5. Menghitung durasi lembur terkompensasi berdasarkan jam lembur aktual yang dikerjakan di dalam jendela yang telah disetujui ($\text{Actual Worked OT} \le \text{Approved OT}$).
  6. Memicu pendistribusian Bonus XP dan kompensasi finansial (jika ada kebijakan tarif berlaku) melalui Deduction Engine ke Personal Wallet.
- **Business Rules:**
  - BR-013: Schedule-Governed Overtime: Overtime wajib disetujui supervisor sebelum dilaksanakan dan dibatasi jendela jadwal.
  - BR-014: Actual Worked Overtime: Kompensasi dihitung dari jam aktual yang dikerjakan dalam jendela disetujui, bukan dari durasi pengajuan mentah.
- **Validation:**
  - Pengajuan lembur harus dilakukan sebelum jendela waktu lembur dimulai (tidak berlaku retroaktif, kecuali melalui alur koreksi khusus).
  - Waktu mulai lembur harus $\ge \text{end\_time}$ dari jadwal kerja normal hari tersebut.
- **Error States:**
  - `422 Unprocessable Entity`: Pengajuan lembur di masa lampau tanpa otorisasi admin.
  - `400 Bad Request`: Eksekusi lembur tanpa adanya record approval yang sah.
- **Empty States:**
  - Daftar permohonan lembur supervisor menampilkan keterangan: "Tidak ada permohonan lembur yang memerlukan tindakan saat ini."
- **Acceptance Criteria:**
  - **Given** intern mengajukan lembur 2 jam (17:00–19:00) dan disetujui Supervisor, **When** intern bekerja aktual selama 1,5 jam (17:00–18:30) lalu melakukan checkout, **Then** sistem menghitung kompensasi lembur untuk 1,5 jam kerja riil dan memberikan bonus XP terkait.
- **Dependencies:** FR-007 (Work Schedule Engine), FR-008 (Attendance Event Engine), FR-025 (XP Rules Engine), FR-034 (Deduction Engine).
- **Audit Events:** `OVERTIME_REQUESTED`, `OVERTIME_APPROVED`, `OVERTIME_REJECTED`, `OVERTIME_COMPENSATED`.
- **Notifications:** Notifikasi ke Supervisor saat lembur diajukan; notifikasi hasil keputusan (Disetujui/Ditolak) ke Intern.
- **Data Entities:** `Overtime Request`, `Attendance Event Log` (DATA-001), `Work Schedule`, `Wallet Transaction`.

---

#### FR-013: Device Registry (Prioritas: P1)
- **Aktor:** Admin, Super Admin, Scanner Operator.
- **Tujuan:** Mengelola daftar terminal perangkat pembaca kehadiran resmi (hardware NFC scanner, dedicated camera tablet, browser operator) untuk mencegah pemalsuan identitas check-in.
- **Kebutuhan Utama:** Pendaftaran identitas perangkat (`device_id`, MAC/IMEI/UUID, tipe perangkat, lokasi fisik, status aktif). Penerbitan API key/device token khusus yang dirotasi secara berkala. Pemblokiran perangkat yang terindikasi mencurigakan.
- **Aturan Bisnis & Catatan:** Mengikat peran Scanner Operator (BR-002).
- **Entitas:** `Device`, `Attendance Event Log` (DATA-001).

---

#### FR-014: Leave Management (Prioritas: P1)
- **Aktor:** Intern, Supervisor, HR Admin.
- **Tujuan:** Memfasilitasi pengajuan dan persetujuan izin/cuti magang (sakit, izin akademik, keperluan mendesak).
- **Kebutuhan Utama:** Pengajuan cuti (rentang tanggal, tipe cuti, bukti surat keterangan dokter/dokumen institusi, alasan). Alur persetujuan oleh Supervisor atau HR Admin. Status cuti yang disetujui secara otomatis mengecualikan intern dari penalti absensi harian.
- **Aturan Bisnis & Catatan:** Cuti yang disetujui tidak dikenakan penalti kehadiran (0 XP). Kebijakan kuota cuti bertanda `[TBD — OQ-006]`.
- **Entitas:** `Leave Request`, `User`, `Attendance Event Log`.

---

#### FR-015: Attendance Correction & Exception Management
- **Feature ID:** FR-015
- **Feature Name:** Attendance Correction & Exception Management
- **Priority:** P0
- **Actor:** Intern, Supervisor, Admin
- **Purpose:** Menyediakan mekanisme formal untuk merevisi kesalahan data kehadiran (lupa checkout, kegagalan pemindai NFC, kendala koneksi) tanpa merusak keaslian jejak audit sistem.
- **Preconditions:** Terdapat kegagalan pencatatan presensi (lupa checkout, kartu error, jaringan offline) yang dilaporkan pengguna.
- **Postconditions:** Record koreksi disahkan, sistem menghitung ulang durasi sesi dan penalti XP, jejak audit asli tetap utuh.
- **User Story:** Sebagai Intern, saya ingin mengajukan koreksi kehadiran dengan melampirkan bukti dan alasan ketika saya lupa checkout atau scanner bermasalah, sehingga jam kerja dan XP saya dapat disesuaikan kembali setelah disetujui.
- **User Flow:** Pengguna mengajukan permohonan koreksi dengan alasan dan bukti -> Supervisor meninjau di Approval Center -> Supervisor menyetujui -> Sistem mengeksekusi penyesuaian -> Audit log mencatat riwayat lengkap.
- **Functional Requirements:**
  1. Menyediakan formulir pengajuan koreksi kehadiran dengan payload: Tanggal kejadian, Tipe event yang dikoreksi (misal: penambahan `CHECK_OUT`), Waktu koreksi yang diajukan, Alasan koreksi, dan Berkas bukti pendukung.
  2. Menyediakan antarmuka peninjauan bagi Supervisor dan HR Admin pada Approval Center.
  3. Menerapkan kalkulasi ulang otomatis pada durasi sesi kerja, pelanggaran absensi, dan transaksi XP setelah pengajuan koreksi disetujui (`APPROVED`).
  4. Menjaga jejak audit permanen: sistem DILARANG menghapus atau menimpa rekaman event asli di database, melainkan mencatat entitas penyesuaian baru (`Attendance Correction Record`) yang mereferensikan event asal.
- **Business Rules:**
  - BR-025: Audit Trail Preservation on Corrections: Koreksi kehadiran manual harus mempertahankan log event asli secara utuh.
- **Validation:**
  - Waktu koreksi yang diajukan harus berada dalam batas yang logis terhadap jadwal kerja dan event kehadiran yang telah tercatat sebelumnya pada hari tersebut.
  - Alasan tertulis wajib diisi minimal 20 karakter dan wajib menyertakan bukti lampiran.
- **Error States:**
  - `422 Unprocessable Entity`: Mengajukan koreksi untuk tanggal di masa depan atau data yang sudah pernah dikoreksi dan disetujui sebelumnya.
- **Empty States:**
  - Antarmuka riwayat koreksi menampilkan informasi: "Belum ada permohonan koreksi kehadiran yang diajukan."
- **Acceptance Criteria:**
  - **Given** intern lupa melakukan checkout pada pukul 17:00, **When** permohonan koreksi checkout pukul 17:00 disetujui Supervisor, **Then** sistem mencatat event koreksi baru berlabel `MANUAL_CORRECTION`, membatalkan penalti Missing Checkout (-2 XP), dan menghitung ulang durasi kerja hari tersebut tanpa mengubah event log historis yang ada.
- **Dependencies:** FR-008 (Attendance Event Engine), FR-025 (XP Rules Engine), FR-047 (Audit Logging).
- **Audit Events:** `ATTENDANCE_CORRECTION_SUBMITTED`, `ATTENDANCE_CORRECTION_APPROVED`, `ATTENDANCE_CORRECTION_REJECTED`, `RECALCULATION_TRIGGERED`.
- **Notifications:** Notifikasi permohonan koreksi baru ke Supervisor; notifikasi status persetujuan ke Intern.
- **Data Entities:** `Attendance Correction`, `Attendance Event Log` (DATA-001), `Audit Log`, `XP Transaction`.

---

### 8.5 PROJECTS & TASKS

#### FR-016: Project Marketplace
- **Feature ID:** FR-016
- **Feature Name:** Project Marketplace
- **Priority:** P0
- **Actor:** Project Owner (Admin, PM, Supervisor terotorisasi), Intern, Alumni
- **Purpose:** Menyediakan bursa proyek kerja terpusat di mana pemilik proyek dapat memublikasikan inisiatif pekerjaan dan kandidat (intern/alumni) dapat menelusuri serta melamar proyek.
- **Preconditions:** Project Manager memiliki inisiatif proyek dengan deskripsi, skill, dan anggaran bounty yang jelas.
- **Postconditions:** Proyek terbit di marketplace dengan visibilitas yang sesuai dan siap menerima lamaran.
- **User Story:** Sebagai Project Owner, saya ingin memublikasikan proyek dengan rincian keterampilan yang dibutuhkan, kapasitas kuota, tenggat waktu, dan besaran bounty pool, sehingga anggota tim yang kompeten dapat ditemukan dan ditugaskan.
- **User Flow:** PM mengisi spesifikasi proyek -> Memilih visibilitas (Intern Only, Public, Private) -> Mempublikasikan proyek -> Proyek muncul di katalog bursa.
- **Functional Requirements:**
  1. Memungkinkan pemilik proyek membuat dan menerbitkan proyek dengan atribut lengkap: Judul (`title`), Deskripsi (`description`), Pemilik (`owner_id`), Tingkat Visibilitas (`visibility`), Kebutuhan Skill (`required_skills`), Kapasitas Kuota Tim (`capacity`), Tenggat Waktu (`deadline`), Alokasi Bounty Pool (`bounty_pool`), dan Status Proyek.
  2. Mendukung 3 tingkat visibilitas proyek secara ketat:
     - `INTERN_ONLY`: Hanya terlihat dan dapat dilamar oleh peserta magang aktif.
     - `PUBLIC`: Terlihat dan dapat dilamar oleh peserta magang aktif dan Alumni.
     - `PRIVATE`: Hanya dapat dilihat dan diakses oleh pengguna yang diundang secara eksplisit.
  3. Mengelola siklus status proyek: `DRAFT → PUBLISHED → IN_PROGRESS → COMPLETED → CANCELLED`.
  4. Menampilkan indikator keterisian kuota secara real-time (misal: `1/5`).
- **Business Rules:**
  - BR-003: Alumni berhak melihat dan melamar proyek dengan visibilitas `PUBLIC`.
  - BR-005: Daily Work, Project Tasks, dan Work Reports adalah entitas terpisah. Proyek mewadahi kumpulan Task terstruktur.
- **Validation:**
  - Nilai `bounty_pool` tidak boleh bernilai negatif.
  - `deadline` proyek harus berada di masa depan pada saat dipublikasikan.
  - Kapasitas kuota tim minimal 1 orang.
- **Error States:**
  - `400 Bad Request`: Publikasi proyek gagal karena field wajib belum lengkap.
  - `403 Forbidden`: Intern mencoba mengakses proyek `PRIVATE` di mana dirinya tidak diundang.
- **Empty States:**
  - Halaman marketplace menampilkan pesan: "Belum ada proyek yang dipublikasikan saat ini. Silakan hubungi Project Manager Anda."
- **Acceptance Criteria:**
  - **Given** Project Manager memublikasikan proyek baru bertanda `PUBLIC` dengan kuota 3 orang dan bounty pool Rp1.500.000, **When** Intern atau Alumni membuka marketplace, **Then** proyek tersebut tampil pada daftar eksplorasi dan tombol lamaran aktif selama kuota belum terpenuhi.
- **Dependencies:** FR-001 (RBAC), FR-006 (Skill Matrix).
- **Audit Events:** `PROJECT_CREATED`, `PROJECT_PUBLISHED`, `PROJECT_STATUS_CHANGED`.
- **Notifications:** Notifikasi proyek baru tersiar ke seluruh pengguna yang memiliki keahlian relevan.
- **Data Entities:** `Project` (DATA-003), `User`, `Skill`.

---

#### FR-017: Project Application & Quota Workflow
- **Feature ID:** FR-017
- **Feature Name:** Project Application & Quota Workflow
- **Priority:** P0
- **Actor:** Intern, Alumni, Project Owner, Reviewer
- **Purpose:** Mengatur alur seleksi pendaftaran anggota tim proyek mulai dari pengajuan lamaran hingga konfirmasi penerimaan dan penguncian kuota tim.
- **Preconditions:** Proyek berstatus PUBLISHED dan masih memiliki kuota kapasitas pelamar yang tersedia.
- **Postconditions:** Lamaran tercatat dengan status Applied dan menunggu penelaahan PM.
- **User Story:** Sebagai Intern atau Alumni, saya ingin melamar proyek yang sesuai dengan keahlian saya, dan sebagai Reviewer, saya ingin menyeleksi serta menyetujui pelamar terpilih hingga kuota tim terpenuhi.
- **User Flow:** Kandidat melihat detail proyek -> Menekan [Lamar Proyek] -> Menuliskan pengantar singkat -> Sistem menghitung Skill Match % -> PM meninjau dan menerima/menolak.
- **Functional Requirements:**
  1. Menyediakan formulir lamaran proyek bagi kandidat dengan menyertakan surat motivasi/catatan lamaran dan portofolio keahlian terkait.
  2. Menerapkan siklus status lamaran resmi: `APPLIED → UNDER_REVIEW → SHORTLISTED → ACCEPTED → REJECTED → WITHDRAWN`.
  3. Memperbarui jumlah anggota yang diterima (`accepted_count`) secara atomik ketika pelamar berpindah status ke `ACCEPTED`.
  4. Secara otomatis menonaktifkan tombol lamaran dan menolak pengajuan baru ketika `accepted_count` mencapai batas `capacity` kuota proyek.
  5. Menghubungkan pelamar yang telah `ACCEPTED` ke dalam entitas tim proyek (`Project Team`).
- **Business Rules:**
  - BR-015: Penolakan otomatis berbasis algoritma dilarang; keputusan akhir penolakan/penerimaan wajib dilakukan oleh penilai manusia (Reviewer/PM).
- **Validation:**
  - Pelamar tidak dapat melamar proyek yang sama dua kali jika lamaran sebelumnya masih berjalan (`APPLIED` / `UNDER_REVIEW`).
  - Pelamar tidak dapat melamar proyek yang kuotanya telah penuh atau statusnya bukan `PUBLISHED`.
- **Error States:**
  - `409 Conflict`: Kuota tim telah penuh saat pelamar menekan tombol konfirmasi penerimaan.
  - `400 Bad Request`: Melamar proyek yang telah dibatalkan atau selesai.
- **Empty States:**
  - Tab lamaran pada profil pengguna menampilkan: "Anda belum pernah melamar proyek apa pun. Kunjungi Marketplace Proyek."
- **Acceptance Criteria:**
  - **Given** proyek dengan sisa kuota 1 pelamar (`capacity = 5`, `accepted_count = 4`), **When** Reviewer menyetujui satu kandidat menjadi `ACCEPTED`, **Then** `accepted_count` menjadi 5, sistem otomatis mengunci status lamaran proyek menjadi penuh, dan pelamar dimasukkan ke dalam tim proyek.
- **Dependencies:** FR-016 (Project Marketplace), FR-019 (Project Team Management).
- **Audit Events:** `PROJECT_APPLICATION_SUBMITTED`, `APPLICATION_STATUS_UPDATED`, `PROJECT_QUOTA_REACHED`.
- **Notifications:** Notifikasi penerimaan/penolakan lamaran ke pelamar; notifikasi pengajuan baru ke Project Owner.
- **Data Entities:** `Project Application`, `Project` (DATA-003), `Project Team`, `User`.

---

#### FR-018: Skill Matching Engine (Prioritas: P2)
- **Aktor:** Reviewer, Project Owner, Sistem Algoritma.
- **Tujuan:** Menghitung persentase kecocokan kompetensi pelamar terhadap syarat keahlian proyek untuk membantu penilai mengambil keputusan penugasan.
- **Kebutuhan Utama:** Algoritma pembanding berbasis kemiripan vektor/irisan keahlian: membandingkan profil skill pelamar terhadap daftar `required_skills` proyek dan menghasilkan nilai presentase kecocokan (misal: 85% Match). Menampilkan rincian skill yang cocok dan yang belum dimiliki.
- **Aturan Bisnis & Catatan:** Sesuai BR-015: Output persentase kecocokan murni bersifat informatif/advisory bagi Reviewer. Penolakan kandidat secara otomatis oleh sistem sangat dilarang.
- **Entitas:** `Skill`, `Project`, `User`, `Project Application`.

---

#### FR-019: Project Team Management
- **Feature ID:** FR-019
- **Feature Name:** Project Team Management
- **Priority:** P0
- **Actor:** Project Manager, Project Owner, Supervisor, Anggota Tim Proyek
- **Purpose:** Mengelola struktur keanggotaan, penugasan peran fungsional, dan alokasi proporsi kontribusi anggota dalam tim proyek.
- **Preconditions:** Proyek memiliki anggota yang telah diterima melalui alur lamaran resmi.
- **Postconditions:** Struktur tim proyek terbentuk dengan peran, tanggung jawab, dan Planned Contribution %.
- **User Story:** Sebagai Project Manager, saya ingin menetapkan peran, tanggung jawab, dan proporsi kontribusi terencana (Planned Contribution %) bagi setiap anggota tim proyek, sehingga pembagian beban kerja dan hak bounty terdefinisi sejak awal.
- **User Flow:** PM menetapkan peran anggota tim (Lead, Developer, Designer) -> Menetapkan porsi Planned Contribution % (misal A: 40%, B: 35%, C: 25%) -> Mengesahkan tim.
- **Functional Requirements:**
  1. Membentuk struktur hierarki tim proyek yang terdiri dari: Pemilik Proyek (`Owner`), Pengelola (`Manager`), Penyelia (`Supervisor`), dan Anggota Tim Pelaksana (`Team Members`).
  2. Menyimpan atribut per anggota tim: Peran spesifik dalam proyek, Tanggung jawab utama, Daftar penugasan tugas (`Assigned Tasks`), Persentase Kontribusi Terencana (`Planned Contribution %`), Skor Performa Proyek, dan Hak Pembagian Bounty.
  3. Memvalidasi bahwa total akumulasi `Planned Contribution %` dari seluruh anggota tim harus berjumlah tepat 100%.
  4. Menjadi fondasi bagi kalkulasi Three-Layer Contribution Engine (FR-024).
- **Business Rules:**
  - BR-016: 3-Tier Contribution Model: Planned Contribution merupakan lapisan dasar kesepakatan awal proporsi kerja sebelum diuji terhadap realisasi kerja aktual.
- **Validation:**
  - Anggota tim yang ditambahkan harus terdaftar sebagai pengguna aktif di platform.
  - Penjumlahan seluruh `Planned Contribution %` wajib sama dengan 100,00% sebelum milestone pertama dapat dimulai.
- **Error States:**
  - `422 Unprocessable Entity`: Total alokasi persentase kontribusi terencana kurang atau lebih dari 100%.
- **Empty States:**
  - Tab struktur tim menampilkan informasi: "Tim proyek belum dibentuk. Silakan terima pelamar dari daftar aplikasi proyek."
- **Acceptance Criteria:**
  - **Given** proyek dengan 3 anggota tim, **When** Project Manager mengatur alokasi kontribusi: Anggota A = 40%, Anggota B = 35%, Anggota C = 25%, **Then** sistem menyimpan konfigurasi tim sebagai `Planned Contribution` yang valid.
- **Dependencies:** FR-016 (Project Marketplace), FR-017 (Project Application).
- **Audit Events:** `PROJECT_TEAM_FORMED`, `TEAM_MEMBER_ADDED`, `PLANNED_CONTRIBUTION_SET`.
- **Notifications:** Notifikasi penugasan tim dan detail alokasi peran kepada anggota tim yang bersangkutan.
- **Data Entities:** `Project Team`, `Project` (DATA-003), `User`.

---

#### FR-020: Milestone Management (Prioritas: P1)
- **Aktor:** Project Manager, Supervisor, Anggota Tim.
- **Tujuan:** Memecah pekerjaan proyek besar ke dalam fase-fase pencapaian perantara yang terukur dengan bobot dan tenggat waktu bertahap.
- **Kebutuhan Utama:** CRUD milestone (Judul, Deskripsi, Tenggat Waktu, Bobot Persentase Proyek, Status: `PENDING`, `IN_PROGRESS`, `COMPLETED`). Setiap milestone membawahi sekumpulan tugas pekerjaan (`Tasks`). Penyelesaian milestone berkontribusi terhadap kalkulasi kontribusi aktual tim.
- **Aturan Bisnis & Catatan:** Menjadi jembatan antara Project dan Task. Total bobot seluruh milestone dalam satu proyek wajib 100%.
- **Entitas:** `Milestone`, `Project`, `Task`.

---

#### FR-021: Task Management
- **Feature ID:** FR-021
- **Feature Name:** Task Management
- **Priority:** P0
- **Actor:** Project Manager, Supervisor, Intern, Alumni
- **Purpose:** Menyediakan unit pekerjaan spesifik yang ditugaskan kepada anggota tim dengan kriteria deliverable dan bobot kontribusi yang terdefinisi jelas.
- **Preconditions:** Proyek aktif dan milestone telah ditetapkan oleh Project Manager.
- **Postconditions:** Tugas kerja spesifik terdaftar pada papan tugas dan siap dikerjakan oleh anggota yang ditugaskan.
- **User Story:** Sebagai Anggota Tim, saya ingin melihat tugas yang ditugaskan kepada saya, memperbarui statusnya, dan menautkannya ke laporan kerja harian saya, sehingga progres penyelesaian proyek terpantau dengan jelas.
- **User Flow:** PM/Lead membuat kartu tugas baru -> Menentukan judul, deskripsi, bobot kontribusi, deadline, dan penugasan -> Anggota melihat tugas di My Day.
- **Functional Requirements:**
  1. Membuat, mengedit, dan menetapkan tugas dengan atribut: Judul tugas, Deskripsi teknis, Tautan milestone induk, Penanggung jawab (`assignee_id`), Estimasi jam kerja, Tingkat kesulitan/bobot tugas, Tenggat waktu, dan Prioritas.
  2. Menerapkan siklus status tugas resmi: `TODO → IN_PROGRESS → IN_REVIEW → COMPLETED → BLOCKED → CANCELLED`.
  3. Mengaitkan sesi kerja aktif intern dengan tugas tertentu yang sedang dikerjakan (`active_task_id`).
  4. Mencegah penyelesaian tugas secara sepihak oleh pelaksana: status `COMPLETED` hanya dapat diberikan setelah pengajuan laporan kerja dan bukti (`Submission & Evidence`) diverifikasi dan disetujui oleh Reviewer/Supervisor.
- **Business Rules:**
  - BR-005: Work Type Partition: Daily Work, Project Tasks, dan Work Reports adalah entitas terpisah.
- **Validation:**
  - Penanggung jawab tugas wajib merupakan anggota aktif dari tim proyek bersangkutan.
  - Tenggat waktu tugas tidak boleh melampaui tenggat waktu proyek induknya.
- **Error States:**
  - `400 Bad Request`: Mencoba memindahkan tugas ke `IN_REVIEW` tanpa melampirkan Work Report atau Evidence.
- **Empty States:**
  - Papan kanban atau daftar tugas menampilkan pesan: "Belum ada tugas yang ditugaskan kepada Anda pada proyek ini."
- **Acceptance Criteria:**
  - **Given** tugas dengan status `IN_PROGRESS`, **When** anggota tim mengirimkan laporan kerja penyelesaian beserta tautan bukti, **Then** status tugas berubah menjadi `IN_REVIEW` dan masuk ke antrean verifikasi reviewer.
- **Dependencies:** FR-016 (Project Marketplace), FR-019 (Project Team Management), FR-020 (Milestone Management).
- **Audit Events:** `TASK_CREATED`, `TASK_ASSIGNED`, `TASK_STATUS_UPDATED`.
- **Notifications:** Notifikasi penugasan tugas baru kepada pelaksana; notifikasi keterlambatan jika tugas melewati batas tenggat.
- **Data Entities:** `Task`, `Project Team`, `Milestone`, `User`.

---

#### FR-022: Work Report / Submission
- **Feature ID:** FR-022
- **Feature Name:** Work Report / Submission
- **Priority:** P0
- **Actor:** Intern, Alumni, Reviewer, Supervisor
- **Purpose:** Merekam laporan pertanggungjawaban penyelesaian pekerjaan berbasis tugas atau harian beserta kemajuan progres dan kendala yang dihadapi.
- **Preconditions:** Peserta magang telah menyelesaikan sebagian atau seluruh pengerjaan tugas proyek.
- **Postconditions:** Laporan kerja resmi tersimpan dan masuk ke antrean ulasan Reviewer.
- **User Story:** Sebagai Pelaksana Tugas, saya ingin mengisi formulir laporan kerja harian yang terstruktur dan mengirimkannya ke penyelia, sehingga akuntabilitas jam kerja dan pencapaian tugas saya tercatat secara sah.
- **User Flow:** Intern membuka tugas -> Mengisi form Work Report (progress %, deskripsi kerja, kendala, rencana berikut) -> Melampirkan evidence -> Menyerahkan laporan.
- **Functional Requirements:**
  1. Menyediakan formulir pelaporan kerja dengan skema data resmi (DATA-004):
     - `id` (UUID)
     - `task_id` (UUID, referensi tugas proyek atau tugas rutin)
     - `user_id` (UUID)
     - `date` (Date pelaporan)
     - `progress_percentage` (Integer 0–100%)
     - `what_i_did` (Text, uraian aktivitas pekerjaan riil)
     - `evidence_type` (Enum: `GIT_COMMIT`, `SCREENSHOT`, `URL`, `DOCUMENT`, `ATTACHMENT`)
     - `evidence_url_or_key` (String, tautan eksternal atau path Cloudflare R2)
     - `problems` (Text, kendala yang dihadapi di lapangan)
     - `next_actions` (Text, rencana aksi hari kerja berikutnya)
     - `status` (Enum: `SUBMITTED`, `UNDER_REVIEW`, `APPROVED`, `REVISION_REQUIRED`)
  2. Mendukung alur peninjauan: Reviewer dapat menyetujui (`APPROVED`) atau meminta revisi perbaikan (`REVISION_REQUIRED`) dengan menyertakan catatan umpan balik.
  3. Pengajuan laporan yang disetujui memicu pencatatan XP kualitas dan mengalir ke perhitungan kontribusi riil (Actual Contribution).
- **Business Rules:**
  - BR-005: Work Type Partition: Daily Work, Project Tasks, dan Work Reports adalah entitas terpisah.
- **Validation:**
  - Field `what_i_did` wajib diisi minimal 30 karakter.
  - `progress_percentage` harus bernilai antara 0 hingga 100.
  - Wajib menyertakan minimal 1 bukti (`Evidence`) yang valid.
- **Error States:**
  - `422 Unprocessable Entity`: Field bukti tidak valid atau persentase progres menurun tanpa justifikasi revisi.
- **Empty States:**
  - Riwayat pengiriman menampilkan: "Belum ada laporan kerja yang dikirimkan untuk tugas ini."
- **Acceptance Criteria:**
  - **Given** intern mengisi laporan tugas dengan progres 100% dan melampirkan tautan commit Git, **When** intern menekan kirim, **Then** sistem menyimpan record dengan status `SUBMITTED` dan memberitahukan Reviewer terkait.
- **Dependencies:** FR-021 (Task Management), FR-023 (Evidence Management).
- **Audit Events:** `WORK_REPORT_SUBMITTED`, `WORK_REPORT_APPROVED`, `WORK_REPORT_REVISION_REQUESTED`.
- **Notifications:** Notifikasi pengajuan laporan kerja baru ke Supervisor; notifikasi status persetujuan ke Intern.
- **Data Entities:** `Work Report & Submission` (DATA-004), `Task`, `User`, `Evidence`.

---

#### FR-023: Evidence Management
- **Feature ID:** FR-023
- **Feature Name:** Evidence Management
- **Priority:** P0
- **Actor:** Intern, Alumni, Reviewer, Sistem Storage
- **Purpose:** Mengelola, memvalidasi, dan menyimpan artefak bukti pekerjaan yang dapat dipertanggungjawabkan untuk setiap penyelesaian tugas proyek atau laporan kerja.
- **Preconditions:** Pengguna memiliki tautan commit Git, PR, dokumen, atau berkas tangkapan layar hasil kerja nyata.
- **Postconditions:** Metadata bukti tersimpan di database dan berkas fisik aman di bucket Cloudflare R2.
- **User Story:** Sebagai Pelaksana Tugas, saya ingin mengunggah tangkapan layar hasil kerja, dokumen pendukung, atau menautkan tautan commit Git sebagai bukti sah pekerjaan saya, sehingga hasil kerja saya dapat diverifikasi keasliannya oleh penilai.
- **User Flow:** Pengguna melampirkan URL commit GitHub / mengunggah berkas -> Sistem memvalidasi URL/format berkas -> Mengunggah ke R2 melalui presigned URL -> Menyimpan metadata.
- **Functional Requirements:**
  1. Mendukung 5 kategori tipe bukti verifikasi:
     - `GIT_COMMIT`: Tautan commit atau Pull Request pada repositori Git resmi (GitHub/GitLab).
     - `SCREENSHOT`: Tangkapan layar artefak kerja visual yang diunggah ke storage.
     - `URL`: Tautan staging live, figma, atau dashboard produksi eksternal.
     - `DOCUMENT`: Berkas laporan formal (PDF, Word, spreadsheet).
     - `ATTACHMENT`: Berkas arsip atau aset pelengkap lainnya.
  2. Mengintegrasikan penyimpanan berkas biner secara eksklusif ke Cloudflare R2 Object Storage (FR-044).
  3. Menegakkan aturan integritas berkas: database hanya menyimpan metadata berkas (DATA-007), tidak menyimpan data biner.
  4. Menyediakan penampil pratinjau bukti terintegrasi bagi Reviewer pada saat proses peninjauan laporan.
- **Business Rules:**
  - BR-024: Cloudflare R2 Storage Separation: Database hanya menyimpan metadata file; objek fisik berada di Cloudflare R2.
- **Validation:**
  - Berkas unggahan dibatasi ukuran maksimalnya (misal: 25 MB per dokumen/gambar) dan tipe MIME yang diizinkan.
  - Bukti berupa URL Git wajib lolos validasi format regex URL repositori resmi.
- **Error States:**
  - `413 Payload Too Large`: Ukuran berkas bukti melebihi batas yang dikonfigurasi.
  - `415 Unsupported Media Type`: Format berkas yang diunggah tidak didukung kebijakan sistem.
- **Empty States:**
  - Panel bukti menampilkan keterangan: "Belum ada berkas bukti yang dilampirkan."
- **Acceptance Criteria:**
  - **Given** pelaksana mengunggah berkas screenshot bukti pengujian, **When** proses upload selesai, **Then** berkas tersimpan di Cloudflare R2 dengan key `task-evidence/YYYY/MM/{uuid}.png` dan referensi record metadata tersimpan di database.
- **Dependencies:** FR-044 (Cloudflare R2 Storage), FR-022 (Work Report / Submission).
- **Audit Events:** `EVIDENCE_UPLOADED`, `EVIDENCE_DELETED`, `EVIDENCE_VERIFIED`.
- **Notifications:** Tidak ada notifikasi langsung; terhubung pada notifikasi submission FR-022.
- **Data Entities:** `Evidence`, `File Metadata` (DATA-007), `Work Report & Submission` (DATA-004).

---

#### FR-024: Three-Layer Contribution Engine
- **Feature ID:** FR-024
- **Feature Name:** Three-Layer Contribution Engine
- **Priority:** P0
- **Actor:** Supervisor, Project Manager, Sistem Algoritma
- **Purpose:** Menghitung dan menetapkan proporsi kontribusi akhir setiap anggota tim proyek secara adil dan transparan melalui rekonsiliasi tiga lapisan perhitungan.
- **Preconditions:** Proyek telah mencapai tahap akhir penyelesaian dan seluruh tugas telah selesai diulas.
- **Postconditions:** Persentase Final Contribution % disahkan oleh Supervisor sebagai acuan mutlak pembagian porsi bounty.
- **User Story:** Sebagai Supervisor dan Project Manager, saya ingin meninjau perbandingan kontribusi terencana versus kontribusi aktual berbasis data, sehingga saya dapat mengesahkan persentase kontribusi akhir yang menjadi dasar alokasi pencairan bounty proyek.
- **User Flow:** Sistem menghitung Actual Contribution % berdasarkan tugas selesai dan bobot -> Supervisor meninjau rekomendasi sistem -> Supervisor menyesuaikan dan mengonfirmasi Final Contribution %.
- **Functional Requirements:**
  1. Menerapkan model perhitungan kontribusi tiga lapis yang independen:
     - **Layer 1 (Planned Contribution):** Persentase kesepakatan awal saat pembentukan tim proyek (misal: Anggota A = 40%, B = 35%, C = 25%).
     - **Layer 2 (Actual Contribution):** Dihitung secara algoritmik oleh sistem berdasarkan kumpulan metrik kerja: jumlah tugas yang diselesaikan, penyelesaian milestone tepat waktu, bobot kesulitan tugas, rating evaluasi peninjau pada submission, dan validitas bukti kerja.
     - **Layer 3 (Final Contribution):** Peninjauan dan penyesuaian manual oleh Supervisor terhadap rekomendasi sistem dari Layer 2, menghasilkan persentase akhir yang dikonfirmasi secara resmi.
  2. Menyediakan antarmuka perbandingan visual ketiga layer kontribusi pada Supervisor Portal sebelum pengesahan final.
  3. Memastikan bahwa total persentase pada masing-masing layer selalu berjumlah tepat 100,00%.
  4. Menyalurkan persentase Final Contribution yang telah disahkan ke Project Bounty Distribution (FR-032).
- **Business Rules:**
  - BR-016: 3-Tier Contribution Model: Bagian bounty dihitung melalui alur `Planned → Actual (sistem) → Final (konfirmasi supervisor)`.
- **Validation:**
  - Supervisor wajib memasukkan justifikasi tertulis jika terjadi deviasi signifikan (>10%) antara Actual Contribution hasil sistem dan Final Contribution yang ditetapkan secara manual.
  - Akumulasi total `Final Contribution %` seluruh anggota tim proyek wajib berjumlah tepat 100,00%.
- **Error States:**
  - `422 Unprocessable Entity`: Total persentase Final Contribution tidak berjumlah 100,00%.
  - `400 Bad Request`: Mencoba mengesahkan Final Contribution saat proyek belum mencapai status penyelesaian tugas minimal.
- **Empty States:**
  - Tabel evaluasi kontribusi menampilkan placeholder: "Kalkulasi kontribusi aktual akan tampil setelah tugas pertama disetujui."
- **Acceptance Criteria:**
  - **Given** proyek selesai dengan data Layer 2 (Actual) Anggota A = 45% dan Anggota B = 55%, **When** Supervisor mengonfirmasi nilai tersebut sebagai Final Contribution, **Then** sistem mengunci nilai persentase tersebut dan mengalirkannya ke mesin kalkulasi bounty finansial.
- **Dependencies:** FR-019 (Project Team Management), FR-021 (Task Management), FR-022 (Work Report / Submission), FR-032 (Project Bounty Distribution).
- **Audit Events:** `ACTUAL_CONTRIBUTION_CALCULATED`, `FINAL_CONTRIBUTION_CONFIRMED`, `CONTRIBUTION_OVERRIDE_LOGGED`.
- **Notifications:** Notifikasi hasil penetapan kontribusi akhir dan rincian evaluasi ke seluruh anggota tim proyek.
- **Data Entities:** `Project Team`, `Project` (DATA-003), `Performance Evaluation`, `Audit Log`.

---

### 8.6 PERFORMANCE & GAMIFICATION

#### FR-025: XP Rules Engine
- **Feature ID:** FR-025
- **Feature Name:** XP Rules Engine
- **Priority:** P0
- **Actor:** Intern, Alumni, System Event Engine, Admin
- **Purpose:** Menyediakan mesin perhitungan poin gamifikasi (Experience Points) yang transparan, konsisten, dan terkonfigurasi secara dinamis untuk mengapresiasi kebiasaan kerja positif dan mendisiplinkan pelanggaran.
- **Preconditions:** Peristiwa operasional (kehadiran, penyelesaian tugas, lembur, evaluasi) terjadi dalam sistem.
- **Postconditions:** Poin pengalaman (XP) dikreditkan atau dipotong sesuai aturan kebijakan aktif dan saldo XP terbarui.
- **User Story:** Sebagai Intern, saya ingin memperoleh XP atas ketepatan kehadiran dan penyelesaian tugas berkualitas tinggi, serta mengetahui konsekuensi penalti XP atas pelanggaran, sehingga saya termotivasi meningkatkan disiplin dan kinerja harian.
- **User Flow:** Event presensi/tugas terjadi -> XP Rules Engine mencocokkan event ke policy aktif -> Menghitung nominal poin -> Mencatat transaksi mutasi XP -> Menambah/mengurangi saldo XP pengguna.
- **Functional Requirements:**
  1. Menghitung perolehan XP positif dari berbagai trigger aktivitas:
     - **Attendance:** Check-in tepat waktu (`ON_TIME`), streak kehadiran tanpa putus, kehadiran penuh dalam seminggu, dan eksekusi lembur yang disetujui.
     - **Work Session:** Menyelesaikan target durasi sesi kerja terjadwal, konsistensi fokus kerja.
     - **Project:** Penyelesaian tugas proyek (`Task Completed`), penuntasan milestone, penyelesaian tugas lebih awal dari tenggat, dan keberhasilan proyek.
     - **Quality:** Rating bintang evaluasi kerja dari supervisor, kelengkapan bukti kerja, umpan balik positif peninjau.
     - **Learning:** Penyelesaian modul onboarding/pelatihan, pembukaan achievement tertentu, dan peningkatan level skill.
  2. Menerapkan matriks penalti deduksi XP yang dapat dikonfigurasi melalui Policy Engine (FR-045):
     - Terlambat 1–15 menit: **-1 XP**
     - Terlambat 16–30 menit: **-2 XP**
     - Terlambat >30 menit: **-3 XP**
     - Istirahat tidak sah (Unauthorized break): **-2 XP**
     - Lupa / tidak melakukan checkout (Missing checkout): **-2 XP**
     - Mangkir tanpa keterangan (Absent): **-5 XP**
     - Melewati tenggat waktu tugas (Missed deadline): Dapat dikonfigurasi (`TBD` default policy)
     - Cuti yang disetujui (Approved leave): **0 XP** (bebas penalti)
  3. Mempartisi akumulasi XP ke dalam 3 skema terisolasi yang tidak saling mencampuradukkan:
     - **Internship XP:** Mengukur progresi dan standing magang aktif.
     - **Project XP:** Penghargaan atas penyelesaian proyek teknis nyata.
     - **Alumni Contribution:** Keterlibatan komunitas dan pengerjaan proyek pasca-magang.
  4. Mencatat setiap perubahan saldo XP ke dalam buku besar transaksi XP (`XP Transaction`) yang memuat referensi event pemicu.
- **Business Rules:**
  - BR-004: XP Scheme Separation: 3 jalur XP terpisah. XP Alumni tidak memengaruhi standing magang aktif.
  - BR-012: Configurable XP Penalties: Semua besaran penalti XP digerakkan oleh policy engine yang dapat dikonfigurasi, bukan angka statis di kode.
- **Validation:**
  - Saldo Internship XP untuk perhitungan standing tidak boleh bernilai negatif secara kumulatif (floor limit = 0 XP jika kebijakan menentukan demikian, atau tercatat sebagai nilai penalti riil).
- **Error States:**
  - `400 Bad Request`: Payload transaksi XP tidak memiliki referensi event pemicu yang valid.
- **Empty States:**
  - Riwayat log XP menampilkan informasi: "Belum ada transaksi XP yang tercatat hari ini."
- **Acceptance Criteria:**
  - **Given** intern terlambat check-in 18 menit pada hari kerja normal, **When** Attendance Engine mencatat event, **Then** XP Rules Engine memproses penalti sebesar -2 XP pada skema `Internship XP` dengan referensi ID event kehadiran terkait.
  - **Given** intern mengambil cuti sakit yang telah disetujui HR, **When** jam kerja harian berakhir, **Then** sistem memberikan 0 XP penalti (bukan mangkir).
- **Dependencies:** FR-007 (Work Schedule Engine), FR-008 (Attendance Event Engine), FR-021 (Task Management), FR-045 (Unified Policy Engine).
- **Audit Events:** `XP_CREDITED`, `XP_DEDUCTED`, `XP_RULE_UPDATED`.
- **Notifications:** Toast notification visual instan di UI saat XP diperoleh ("+10 XP: Task Approved!") atau notifikasi peringatan saat penalti terjadi ("-2 XP: Keterlambatan 18 Menit").
- **Data Entities:** `XP Rule`, `XP Transaction`, `User`, `Policy`.

---

#### FR-026: Rank Progression System
- **Feature ID:** FR-026
- **Feature Name:** Rank Progression System
- **Priority:** P0
- **Actor:** Intern, Alumni, System Engine
- **Purpose:** Mengonversi akumulasi poin XP menjadi tingkat peringkat (Rank) gamifikasi yang mencerminkan jam terbang dan dedikasi pengguna.
- **Preconditions:** Pengguna mengumpulkan akumulasi Internship XP yang melampaui ambang batas kenaikan rank berikutnya.
- **Postconditions:** Tingkatan Rank pengguna naik resmi dan paket reward terkait otomatis terbuka.
- **User Story:** Sebagai Intern, saya ingin melihat progres peningkatan level peringkat saya dari waktu ke waktu berdasarkan pencapaian XP, sehingga saya termotivasi untuk mencapai tingkatan rank yang lebih tinggi dan membuka reward eksklusif.
- **User Flow:** Saldo XP mencapai ambang batas -> Sistem memicu event kenaikan rank -> Menampilkan modal perayaan kenaikan level -> Menerbitkan paket reward klaim.
- **Functional Requirements:**
  1. Mengelola tingkatan rank berjenjang (misal: Rank F, E, D, C, B, A, S atau sebutan pangkat magang yang dikonfigurasi pada sistem).
  2. Menghitung progres persentase menuju rank berikutnya secara dinamis berdasarkan ambang batas XP (`XP Threshold`).  
     *(Catatan: Nilai numerik ambang batas pasti dikonfigurasi melalui konfigurasi bisnis; status OQ-003)*.
  3. Memperbarui status rank pengguna secara otomatis segera setelah transaksi XP baru melewati ambang batas rank berikutnya (Promosi Rank).
  4. Memicu pembukaan paket hadiah promosi rank (`Multi-Component Rank Rewards`, FR-031) pada setiap kenaikan level rank.
  5. Menegakkan pemisahan tegas antara Rank gamifikasi dan Skor Kinerja Formal: Rank adalah ukuran progresi pengalaman, sedangkan Performance Score adalah ukuran kualitas hasil evaluasi kerja.
- **Business Rules:**
  - BR-017: Rank vs Performance Score: $\text{Rank} \neq \text{Performance Score}$. Rank adalah progresi gamified; Performance Score adalah evaluasi kualitas formal.
- **Validation:**
  - Pengguna tidak dapat diturunkan rank-nya (demotion) secara otomatis oleh fluktuasi penalti harian biasa, kecuali melalui keputusan disipliner administratif terotorisasi.
- **Error States:**
  - `400 Bad Request`: Kesalahan penentuan jenjang rank karena skema ambang batas tidak konsisten.
- **Empty States:**
  - Pengguna baru memulai pada Rank terendah (Level 1 / Tier Pemula) dengan progres 0%.
- **Acceptance Criteria:**
  - **Given** intern berada pada Rank D dengan batas XP berikutnya 500 XP dan saldo saat ini 490 XP, **When** intern memperoleh +20 XP dari penyelesaian tugas, **Then** sistem menaikkan rank intern menjadi Rank C, mencatat event promosi rank, dan menerbitkan hak klaim reward promosi terkait.
- **Dependencies:** FR-025 (XP Rules Engine), FR-031 (Multi-Component Rank Rewards).
- **Audit Events:** `RANK_PROMOTION_ACHIEVED`, `RANK_THRESHOLD_UPDATED`.
- **Notifications:** Visual fanfare modal promosi rank di antarmuka "Selamat! Anda naik ke Rank C!", disertai rangkuman reward yang terbuka.
- **Data Entities:** `Rank`, `User`, `XP Transaction`, `Reward`.

---

#### FR-027: Performance Evaluation Engine
- **Feature ID:** FR-027
- **Feature Name:** Performance Evaluation Engine
- **Priority:** P0
- **Actor:** Supervisor, HR Admin, Project Manager, Reviewer
- **Purpose:** Menyediakan perhitungan skor performa profesional yang objektif, formal, dan komposit, terpisah dari mekanik gamifikasi XP.
- **Preconditions:** Periode evaluasi berkala (mingguan/bulanan/akhir batch) tiba dan supervisor siap menilai timnya.
- **Postconditions:** Performance Score formal tersimpan dan terpisah dari skor rank gamifikasi (BR-017).
- **User Story:** Sebagai Supervisor dan HR Admin, saya ingin mengevaluasi kinerja peserta magang secara komprehensif berdasarkan kehadiran, ketepatan penyelesaian tugas, kualitas kerja, dan penilaian perilaku, sehingga keputusan sertifikasi dan predikat kelulusan didasarkan pada data evaluasi yang teruji.
- **User Flow:** Supervisor membuka form evaluasi kinerja anggota -> Memberikan nilai kriteria kualitas, disiplin, dan inisiatif -> Mengesahkan evaluasi -> Skor tersimpan.
- **Functional Requirements:**
  1. Menghitung Skor Kinerja Formal (`Performance Score`) pengguna menggunakan formula komposit berbobot yang dapat dikonfigurasi:
     - Komponen **Attendance:** Rasio ketepatan waktu dan tingkat kehadiran penuh.
     - Komponen **Task Delivery:** Rasio tugas selesai tepat waktu terhadap total tugas yang ditugaskan.
     - Komponen **Work Quality:** Rata-rata nilai evaluasi reviewer terhadap laporan kerja dan evidence.
     - Komponen **Supervisor Evaluation:** Skor evaluasi rubrik berkala (integritas, komunikasi, kerja sama tim, inisiatif).
     - Komponen **Project Contribution:** Bobot kontribusi akhir yang telah disahkan pada proyek nyata.
  2. Menyediakan formulir penilaian rubrik berkala bagi Supervisor untuk memberikan skor kualitatif dan kuantitatif.
  3. Memastikan kalkulasi skor performa berjalan secara independen dari jumlah akumulasi XP mentah pengguna.
  4. Menjadi data rujukan utama bagi penentuan Top Performer per Batch (FR-028) dan penerbitan sertifikat kelulusan.
- **Business Rules:**
  - BR-017: Rank vs Performance Score: $\text{Rank} \neq \text{Performance Score}$.
  - Pembobotan matematis detail dikonfigurasi melalui Unified Policy Engine `[TBD — OQ-004]`.
- **Validation:**
  - Seluruh bobot variabel dalam formula komposit evaluasi wajib berjumlah tepat 100%.
  - Nilai masukan rubrik evaluasi supervisor berada pada rentang skala 1 hingga 5 atau 0 hingga 100.
- **Error States:**
  - `422 Unprocessable Entity`: Data masukan evaluasi supervisor belum lengkap saat mencoba menerbitkan evaluasi final.
- **Empty States:**
  - Kartu evaluasi kinerja menampilkan status: "Belum ada evaluasi berkala yang diterbitkan untuk periode ini."
- **Acceptance Criteria:**
  - **Given** data metrik kerja intern selama 1 bulan telah terkumpul, **When** Supervisor mengirimkan evaluasi bulanan, **Then** Performance Evaluation Engine menghitung nilai komposit terbobot secara instan dan memperbarui ringkasan kartu performa intern.
- **Dependencies:** FR-008 (Attendance), FR-021 (Task), FR-022 (Work Report), FR-024 (Contribution Engine).
- **Audit Events:** `PERFORMANCE_EVALUATION_SUBMITTED`, `PERFORMANCE_SCORE_CALCULATED`.
- **Notifications:** Notifikasi hasil evaluasi berkala kepada Intern; laporan rekap performa kohort kepada HR Admin.
- **Data Entities:** `Performance Evaluation`, `User`, `Work Schedule`, `Project Team`.

---

#### FR-028: Top Performer per Batch (Prioritas: P1)
- **Aktor:** HR Admin, Super Admin, Supervisor, Sistem Otomatis.
- **Tujuan:** Menetapkan dan memberikan apresiasi tertinggi kepada tepat satu peserta magang terbaik dalam setiap kohort batch pada setiap siklus evaluasi.
- **Kebutuhan Utama:** Algoritma pemeringkat kohort berdasarkan kalkulasi komposit: Performance Score + Final Evaluation + Project Contribution + Attendance Rate. Menetapkan pemenang tunggal per periode (Mingguan, Bulanan, Akhir Batch). Menampilkan badge khusus Top Performer pada profil dan direktori.
- **Aturan Bisnis & Catatan:** Sesuai BR-018: Tepat SATU Top Performer diakui per kohort batch per siklus. Terpisah dari Global Leaderboard (yang menampilkan urutan XP bebas).
- **Entitas:** `Performance Evaluation`, `Batch`, `User`, `Achievement`.

---

#### FR-029: Achievement System (Prioritas: P1)
- **Aktor:** Intern, Alumni, Sistem Event Engine.
- **Tujuan:** Memberikan penghargaan visual dan lencana pencapaian (badges) atas penyelesaian tonggak pencapaian khusus (milestone achievements).
- **Kebutuhan Utama:** Pengelolaan katalog achievement (Judul, Deskripsi, Ikon Badge, Syarat Kriteria: misal "10 Hari On-Time Berturut-turut", "Penyelesaian Proyek Pertama", "Zero Defect Review"). Pemicuan otomatis pembukaan badge saat event pemicu terpenuhi. Opsi pemberian bonus XP atau reward saat achievement terbuka.
- **Aturan Bisnis & Catatan:** Dapat dibuka oleh Intern maupun Alumni (BR-004).
- **Entitas:** `Achievement`, `User`, `XP Transaction`.

---

#### FR-030: Skill Growth Matrix (Prioritas: P1)
- **Aktor:** Intern, Alumni, Supervisor.
- **Tujuan:** Memvisualisasikan dan melacak pertumbuhan level kemahiran keahlian individu seiring berjalannya program magang dan penugasan proyek.
- **Kebutuhan Utama:** Pemetaan level keahlian (Level 1–5). Peningkatan level dihitung dari frekuensi penggunaan skill pada tugas proyek yang berhasil disetujui, akumulasi XP pada domain skill terkait, dan verifikasi kualitatif oleh Supervisor. Grafik radar kompetensi pada profil pengguna.
- **Aturan Bisnis & Catatan:** Mendukung penyusunan portofolio digital lulusan (FR-042).
- **Entitas:** `Skill`, `User`, `Task`, `Performance Evaluation`.

---

### 8.7 INCENTIVES & COMPENSATION

#### FR-031: Multi-Component Rank Rewards
- **Feature ID:** FR-031
- **Feature Name:** Multi-Component Rank Rewards
- **Priority:** P0
- **Actor:** Intern, Admin, Finance
- **Purpose:** Menyediakan paket apresiasi beragam komponen (moneter dan non-moneter) yang dapat diperoleh pengguna setiap kali mencapai kenaikan peringkat (Rank Promotion).
- **Preconditions:** Pengguna mencapai promosi Rank baru yang memiliki konfigurasi paket hadiah aktif.
- **Postconditions:** Paket reward multi-komponen (Cash, XP, Merchandise, Badge, Voucher) tercatat siap diklaim.
- **User Story:** Sebagai Intern, saya ingin menerima paket reward yang menarik (seperti bonus tunai, barang fisik, voucher, atau keistimewaan platform) ketika berhasil naik rank, sehingga pencapaian rank terasa nyata dan berharga.
- **User Flow:** Rank tercapai -> Sistem mengaitkan paket hadiah sesuai konfigurasi policy -> Pengguna melihat daftar hadiah yang didapat pada menu Rewards -> Mengajukan klaim.
- **Functional Requirements:**
  1. Mengelola konfigurasi paket reward multi-komponen untuk setiap jenjang rank, yang mencakup kombinasi dari jenis reward berikut:
     - `CASH`: Bonus insentif uang tunai yang langsung dikreditkan ke Personal Wallet.
     - `BONUS_XP`: Tambahan poin XP akselerasi.
     - `PHYSICAL_ITEM`: Barang merchandise fisik resmi (misal: Jaket, Tumbler, Kaos, ID Badge Eksklusif).
     - `BADGE`: Lencana digital khusus pada profil.
     - `VOUCHER`: Voucher belanja, makanan, atau kursus.
     - `PLATFORM_PRIVILEGE`: Hak istimewa sistem (misal: prioritas pelamaran proyek eksklusif).
  2. Menerbitkan tiket hak klaim reward (`Reward Claim`) secara otomatis saat pengguna dipromosikan ke rank baru.
  3. Mengarahkan komponen bertipe moneter (`CASH`) langsung melalui Deduction & Tax Engine (FR-034) sebelum masuk ke Personal Wallet.
  4. Menyediakan pencatatan status pemenuhan untuk komponen fisik/voucher melalui Reward Claim Management (FR-033).
- **Business Rules:**
  - BR-019: Multi-Component Reward Packages: Promosi rank dapat menggabungkan Cash, XP, Merchandise, Badges, Vouchers, dan Hak Istimewa.
- **Validation:**
  - Setiap jenjang rank minimal memiliki satu komponen reward aktif.
  - Komponen cash harus bernilai $\ge 0$.
- **Error States:**
  - `400 Bad Request`: Kegagalan inisialisasi paket reward karena data komponen tidak lengkap.
- **Empty States:**
  - Jika suatu rank tidak memiliki reward uang tunai, komponen cash tidak ditampilkan dalam paket.
- **Acceptance Criteria:**
  - **Given** pengguna dipromosikan ke Rank B yang memiliki reward: Uang Tunai Rp200.000 + Kaos Eksklusif + Badge "Rank B Achiever", **When** kenaikan rank terjadi, **Then** sistem otomatis mengkreditkan dana Rp200.000 (setelah kalkulasi pajak) ke wallet, menerbitkan tiket klaim kaos ke modul rewards, dan menyematkan badge ke profil pengguna.
- **Dependencies:** FR-026 (Rank Progression System), FR-034 (Deduction Engine), FR-035 (Personal Wallet).
- **Audit Events:** `RANK_REWARD_ISSUED`, `CASH_REWARD_DISBURSED`, `MERCHANDISE_CLAIM_CREATED`.
- **Notifications:** Notifikasi rincian paket hadiah yang terbuka saat kenaikan peringkat terjadi.
- **Data Entities:** `Reward`, `Reward Claim`, `Wallet Transaction` (DATA-006), `Rank`, `User`.

---

#### FR-032: Project Bounty Distribution
- **Feature ID:** FR-032
- **Feature Name:** Project Bounty Distribution
- **Priority:** P0
- **Actor:** Project Manager, Supervisor, Finance, Anggota Tim Proyek
- **Purpose:** Menghitung dan mendistribusikan alokasi imbalan finansial (bounty) proyek kepada setiap anggota tim secara adil berdasarkan persentase kontribusi akhir yang telah disahkan.
- **Preconditions:** Proyek berstatus COMPLETED, Final Contribution % telah disahkan, dan total bounty pool terdefinisi.
- **Postconditions:** Porsi kotor bounty dihitung per anggota tim dan dialirkan ke Deduction & Tax Engine.
- **User Story:** Sebagai Anggota Tim Proyek, saya ingin menerima pembagian bounty proyek yang sesuai dengan persentase kontribusi akhir saya, sehingga jerih payah saya terbayar secara adil dan transparan.
- **User Flow:** PM mengesahkan penyelesaian proyek -> Sistem mengalikan Bounty Pool x Final Contribution % tiap anggota -> Menghasilkan Gross Bounty -> Mengalirkan ke pemotongan pajak dan kas batch.
- **Functional Requirements:**
  1. Mengambil nilai total kumpulan bounty proyek (`bounty_pool`) saat proyek mencapai status `COMPLETED`.
  2. Menghitung nilai imbalan kotor (`Gross Bounty`) untuk setiap anggota tim berdasarkan persentase Final Contribution (FR-024):
     $$\text{Gross Bounty}_i = \text{Bounty Pool} \times \text{Final Contribution}_i\%$$
  3. Mengalirkan nilai Gross Bounty setiap anggota tim ke Financial Flow Orchestration (FR-039) untuk dievaluasi oleh Deduction & Tax Engine (FR-034).
  4. Menerbitkan transaksi kredit bersih (`Net Distributable`) ke Personal Wallet masing-masing anggota tim.
  5. Mengunci seluruh data pembagian bounty agar tidak dapat diubah kembali setelah proses distribusi selesai.
- **Business Rules:**
  - BR-016: 3-Tier Contribution Model: Alokasi bounty bersandar pada Final Contribution % yang dikonfirmasi supervisor.
  - BR-020: Dynamic Deduction Engine: Seluruh distribusi finansial wajib melewati mesin deduksi.
- **Validation:**
  - Akumulasi Gross Bounty seluruh anggota tim tidak boleh melebihi nilai total `bounty_pool` proyek.
  - Proyek wajib berstatus `COMPLETED` dan seluruh persentase kontribusi anggota telah berstatus terkunci (`LOCKED`).
- **Error States:**
  - `400 Bad Request`: Mencoba membagikan bounty pada proyek yang belum disetujui selesai.
  - `422 Unprocessable Entity`: Selisih pembulatan desimal menyebabkan total bounty terbagi melebihi kapasitas pool.
- **Empty States:**
  - Status pembagian bounty menampilkan: "Menunggu proyek selesai dan verifikasi kontribusi tim."
- **Acceptance Criteria:**
  - **Given** proyek bernilai bounty pool Rp10.000.000 dengan anggota A memiliki Final Contribution 40%, **When** Project Manager memicu distribusi bounty, **Then** Gross Bounty anggota A dihitung tepat Rp4.000.000 dan dialirkan ke mesin pemotongan pajak dan dana kebersamaan.
- **Dependencies:** FR-016 (Project Marketplace), FR-024 (Contribution Engine), FR-034 (Deduction Engine), FR-039 (Financial Flow Orchestration).
- **Audit Events:** `BOUNTY_DISTRIBUTION_INITIALIZED`, `GROSS_BOUNTY_CALCULATED`, `BOUNTY_CREDITED_TO_WALLET`.
- **Notifications:** Notifikasi penerimaan bounty proyek ke seluruh anggota tim penerima.
- **Data Entities:** `Project` (DATA-003), `Project Team`, `Wallet Transaction` (DATA-006), `Financial Ledger`.

---

#### FR-033: Reward Claim Management (Prioritas: P1)
- **Aktor:** Intern, Alumni, HR Admin, Finance.
- **Tujuan:** Mengelola siklus operasional klaim reward fisik, merchandise, voucher, atau uang tunai yang membutuhkan persetujuan manual atau pengiriman fisik.
- **Kebutuhan Utama:** Antarmuka pengajuan klaim oleh penerima (alamat pengiriman, konfirmasi penukaran). Siklus status klaim: `ISSUED → CLAIMED → PROCESSING → FULFILLED → REJECTED`. Verifikasi nomor resi pengiriman untuk barang fisik atau nomor referensi voucher.
- **Aturan Bisnis & Catatan:** Sesuai BR-019. Status claim diturunkan dari widget ATTENTION Command Center.
- **Entitas:** `Reward Claim`, `Reward`, `User`.

---

### 8.8 FINANCE & TAXATION

#### FR-034: Deduction & Tax Engine
- **Feature ID:** FR-034
- **Feature Name:** Deduction & Tax Engine
- **Priority:** P0
- **Actor:** Finance, Super Admin, Sistem Kalkulasi Finansial
- **Purpose:** Menghitung potongan pajak dan biaya potongan lainnya secara terpusat, fleksibel, dan dinamis berdasarkan konfigurasi kebijakan tanpa hardcoding aturan perpajakan di kode sumber.
- **Preconditions:** Terdapat peristiwa moneter (bounty proyek atau penarikan dana) yang memerlukan perhitungan pemotongan resmi.
- **Postconditions:** Rincian pemotongan pajak dan kas perpisahan batch fund dihitung akurat sesuai aturan dinamis.
- **User Story:** Sebagai Finance Admin, saya ingin membuat dan memperbarui aturan pemotongan pajak penghasilan, pajak proyek, iuran perpisahan, atau biaya administrasi, sehingga platform selalu patuh terhadap regulasi dan kebijakan internal perusahaan yang berlaku.
- **User Flow:** Gross amount masuk ke Deduction Engine -> Sistem mencocokkan aturan pajak aktif (PPh 21) dan aturan kas batch -> Menghitung nominal potongan -> Menghasilkan Net Distributable.
- **Functional Requirements:**
  1. Mengelola aturan deduksi/pajak dengan skema data resmi (DATA-005):
     - `id` (UUID)
     - `name` (String, misal: "PPh 21 Kompensasi Magang", "Iuran Perpisahan Batch")
     - `type` (Enum: `INCOME_TAX`, `PROJECT_TAX`, `GRADUATION_TAX`, `WITHDRAWAL_TAX`, `ADMINISTRATIVE_FEE`, `PENALTY`, `OTHER_DEDUCTION`)
     - `rate_type` (Enum: `PERCENTAGE`, `FIXED_AMOUNT`)
     - `rate_value` (Decimal, persentase misal 5,00% atau nominal tetap)
     - `calculation_basis` (String/Enum, basis dasar pengenaan pajak)
     - `minimum_amount` (Decimal, batas minimal pemotongan)
     - `maximum_amount` (Decimal, batas plafon maksimal pemotongan)
     - `effective_date` (Date, tanggal mulai berlakunya aturan)
     - `is_active` (Boolean, toggle aktifasi)
  2. Menerapkan evaluasi berantai terhadap setiap pendapatan kotor: menghitung potongan pajak regulasi, biaya penarikan, dan iuran perpisahan batch secara berurutan.
  3. Memastikan pemisahan konseptual mutlak: Iuran Dana Perpisahan (`Farewell Fund`) diklasifikasikan sebagai *Batch Contribution*, terpisah dari *Taxes* perusahaan/pemerintah.
  4. Menghasilkan rincian transparansi potongan (`deduction breakdown`) pada setiap transaksi finansial.
- **Business Rules:**
  - BR-020: Dynamic Deduction Engine: Seluruh formula pajak dan potongan wajib dapat dikonfigurasi tanpa modifikasi kode program.
  - BR-021: Farewell Fund Classification: Deduksi farewell fund diklasifikasikan sebagai Batch Contributions, bukan pajak pemerintah.
- **Validation:**
  - `rate_value` bertipe persentase harus berada dalam rentang 0,00% hingga 100,00%.
  - Aturan yang sedang aktif wajib memiliki `effective_date` yang sah.
- **Error States:**
  - `422 Unprocessable Entity`: Konflik aturan deduksi ganda yang saling tumpang tindih untuk basis perhitungan yang sama.
- **Empty States:**
  - Jika tidak ada aturan deduksi aktif, sistem mengenakan potongan Rp0 (Gross = Net) disertai catatan log audit.
- **Acceptance Criteria:**
  - **Given** aturan pajak proyek aktif 5% dan iuran batch 2%, **When** Gross Bounty Rp1.000.000 diproses, **Then** sistem menghitung pajak proyek Rp50.000, iuran batch Rp20.000, total deduksi Rp70.000, dan menghasilkan nilai bersih Rp930.000.
- **Dependencies:** FR-045 (Unified Policy Engine).
- **Audit Events:** `DEDUCTION_RULE_CREATED`, `DEDUCTION_RULE_UPDATED`, `DEDUCTION_RULE_DEACTIVATED`.
- **Notifications:** Pemberitahuan ke bagian Finance jika terjadi perubahan kebijakan pajak sistem.
- **Data Entities:** `Deduction / Tax Rule` (DATA-005), `Policy`, `Wallet Transaction` (DATA-006).

---

#### FR-035: Personal Wallet & Source Tracking
- **Feature ID:** FR-035
- **Feature Name:** Personal Wallet & Source Tracking
- **Priority:** P0
- **Actor:** Intern, Alumni, Finance, Sistem Finansial
- **Purpose:** Menyediakan buku tabungan digital multi-sumber bagi setiap pengguna untuk menampung seluruh perolehan hak finansial secara transparan dan terperinci.
- **Preconditions:** Pengguna memiliki akun aktif dalam platform DCISP.
- **Postconditions:** Buku kas mutasi dompet mencatat penambahan saldo bersih dan riwayat asal sumber dana secara transparan.
- **User Story:** Sebagai Intern atau Alumni, saya ingin melihat saldo dompet pribadi saya beserta rincian asal muasal dana (proyek, lembur, reward rank) dan detail potongan yang dikenakan, sehingga catatan keuangan saya transparan dan dapat dipertanggungjawabkan.
- **User Flow:** Net Distributable disahkan -> Saldo dompet pribadi bertambah -> Pengguna menerima notifikasi kredit -> Meninjau mutasi itemized di menu Wallet.
- **Functional Requirements:**
  1. Membuat satu entitas dompet pribadi (`Wallet`) untuk setiap pengguna yang terdaftar di platform.
  2. Mencatat setiap mutasi dana dalam skema transaksi terperinci (DATA-006):
     - `id` (UUID)
     - `wallet_id` (UUID)
     - `source` (Enum: `PROJECT_BOUNTY`, `OVERTIME_BONUS`, `RANK_REWARD`, `TAX_DEDUCTION`, `BATCH_CONTRIBUTION`, `PAYOUT`, `OTHER`)
     - `reference_id` (String/UUID, ID proyek/tugas/reward pemicu)
     - `gross_amount` (Decimal)
     - `deduction_amount` (Decimal)
     - `net_amount` (Decimal)
     - `status` (Enum: `PENDING`, `COMPLETED`, `FAILED`, `REVERSED`)
     - `created_at` (Timestamp)
     - `approved_by` (UUID, nullable)
  3. Memelihara saldo bersih riil (`current_balance`) yang diperbarui secara atomik bersamaan dengan pencatatan transaksi.
  4. Mencegah mutasi saldo negatif yang tidak sah: saldo wallet tidak boleh kurang dari Rp0.
- **Business Rules:**
  - BR-022: Itemized Wallet Ledger: Wallet harus menampilkan jejak audit lengkap dan sumber asal untuk setiap mutasi kredit/debit.
  - BR-023: Mandatory Double-Entry Ledger: Setiap mutasi wallet wajib memiliki entri pasangan dalam Financial Ledger.
- **Validation:**
  - Setiap mutasi kredit wajib menyertakan `source` dan `reference_id` yang sah.
  - Nilai `net_amount` harus sama persis dengan `gross_amount - deduction_amount`.
- **Error States:**
  - `400 Bad Request`: Penarikan dana atau mutasi debit melebihi saldo aktif saat ini (insufficient balance).
- **Empty States:**
  - Halaman wallet pengguna baru menampilkan saldo Rp0 dengan pesan: "Belum ada transaksi pada dompet Anda. Selesaikan proyek atau tugas untuk mendapatkan kompensasi."
- **Acceptance Criteria:**
  - **Given** pengguna menerima hak bounty proyek, **When** transaksi berhasil diproses, **Then** riwayat transaksi wallet menampilkan jumlah kotor, rincian potongan, jumlah bersih, sumber proyek terkait, dan saldo bertambah sebesar jumlah bersih.
- **Dependencies:** FR-034 (Deduction Engine), FR-037 (Financial Ledger).
- **Audit Events:** `WALLET_CREDITED`, `WALLET_DEBITED`, `WALLET_BALANCE_RECONCILED`.
- **Notifications:** Notifikasi instan saat saldo masuk ke dompet pengguna ("Dana sebesar RpX berhasil masuk ke dompet Anda dari Proyek Y").
- **Data Entities:** `Wallet`, `Personal Wallet Transaction` (DATA-006), `Financial Ledger`, `User`.

---

#### FR-036: Batch Fund (Prioritas: P1)
- **Aktor:** Finance, HR Admin, Perwakilan Batch Intern.
- **Tujuan:** Mengelola kumpulan dana bersama (Farewell Fund) yang diakumulasikan dari kontribusi potongan batch untuk mendanai kegiatan kebersamaan/perpisahan kohort.
- **Kebutuhan Utama:** Buku besar khusus per batch yang menampung aliran dana `BATCH_CONTRIBUTION`. Tampilan saldo akumulasi per batch transparan untuk seluruh anggota batch. Alur otorisasi pencairan dana untuk kegiatan resmi kohort `[TBD — OQ-007]`.
- **Aturan Bisnis & Catatan:** Sesuai BR-021: Dana batch terpisah total dari pajak perusahaan.
- **Entitas:** `Batch Fund`, `Batch`, `Wallet Transaction` (DATA-006), `Financial Ledger`.

---

#### FR-037: Financial Ledger
- **Feature ID:** FR-037
- **Feature Name:** Financial Ledger
- **Priority:** P0
- **Actor:** Finance Admin, Super Admin, Sistem Finansial
- **Purpose:** Menyediakan buku besar akuntansi ganda (double-entry ledger) yang tidak dapat diubah (immutable) untuk mencatat setiap peristiwa pergerakan moneter di seluruh ekosistem platform.
- **Preconditions:** Setiap transaksi moneter (kredit atau debit) yang terjadi di platform.
- **Postconditions:** Entri jurnal ganda (double-entry) tercatat secara permanen dan tidak dapat diubah atau dihapus (append-only).
- **User Story:** Sebagai Finance Admin dan Auditor, saya ingin setiap perpindahan uang tercatat dalam ledger berpasangan debit-kredit yang tidak dapat dimanipulasi atau dihapus, sehingga integritas keuangan platform terjamin secara absolut.
- **User Flow:** Transaksi keuangan terjadi -> Sistem menulis baris debit dan kredit berpasangan ke Financial Ledger -> Memverifikasi selisih = 0 -> Mengunci record.
- **Functional Requirements:**
  1. Mencatat transaksi menggunakan prinsip pembukuan berpasangan (double-entry): setiap transaksi memiliki minimal satu akun debit dan satu akun kredit dengan total nilai yang seimbang ($\sum \text{Debit} = \sum \text{Kredit}$).
  2. Menegakkan sifat keabadian (immutability): DILARANG menyediakan fungsi UPDATE atau DELETE pada tabel ledger; setiap koreksi kesalahan wajib dilakukan melalui entri jurnal pembalik (`REVERSAL`).
  3. Mengelompokkan kode akun standar: Kas/Bank Penampung, Pool Bounty Proyek, Utang Kompensasi Intern, Utang Pajak PPh, Rekening Penampung Batch Fund, dan Pendapatan Operasional.
  4. Menyediakan antarmuka audit log finansial untuk keperluan rekonsiliasi kas keluar dan masuk.
- **Business Rules:**
  - BR-023: Mandatory Double-Entry Ledger: Semua transaksi finansial harus dicatat dalam ledger yang tidak dapat diubah.
- **Validation:**
  - Total nilai entri debit wajib sama persis dengan total nilai entri kredit dalam satu transaksi ID.
- **Error States:**
  - `500 Internal Server Error / Data Integrity Fault`: Upaya penyimpanan entri ledger yang tidak seimbang (unbalanced transaction) dibatalkan seketika oleh database constraint.
- **Empty States:**
  - Laporan ledger menampilkan saldo nol jika belum ada transaksi finansial yang dibukukan.
- **Acceptance Criteria:**
  - **Given** transaksi pencairan bounty sebesar Rp1.000.000 (bersih Rp930.000, pajak Rp50.000, batch fund Rp20.000), **When** transaksi dibukukan, **Then** ledger mencatat: Debit Pool Bounty Proyek Rp1.000.000; Kredit Utang Intern Rp930.000; Kredit Utang Pajak Rp50.000; Kredit Batch Fund Rp20.000.
- **Dependencies:** FR-035 (Personal Wallet), FR-039 (Financial Flow Orchestration).
- **Audit Events:** `LEDGER_ENTRY_RECORDED`, `LEDGER_REVERSAL_EXECUTED`, `LEDGER_AUDIT_EXPORTED`.
- **Notifications:** Peringatan otomatis ke Finance jika terjadi kegagalan rekonsiliasi neraca saldo ledger.
- **Data Entities:** `Financial Ledger`, `Personal Wallet Transaction` (DATA-006), `Wallet`, `Batch Fund`.

---

#### FR-038: Payout Engine (Prioritas: P1)
- **Aktor:** Intern, Alumni, Finance Admin.
- **Tujuan:** Memproses pencairan dana dari saldo Personal Wallet ke rekening bank atau e-wallet eksternal pengguna.
- **Kebutuhan Utama:** Permohonan pencairan dana (nominal, nama bank/e-wallet, nomor rekening, nama pemilik). Validasi saldo mencukupi. Penahanan saldo sementara (`HOLD`) saat proses pencairan berjalan. Siklus status pencairan: `REQUESTED → APPROVED → PROCESSING → SETTLED → FAILED → REJECTED`. Integrasi payment gateway `[TBD — OQ-005]`.
- **Aturan Bisnis & Catatan:** Sesuai BR-022. Minimal nominal pencairan dikonfigurasi melalui sistem settings.
- **Entitas:** `Payout`, `Wallet`, `Wallet Transaction` (DATA-006), `Financial Ledger`.

---

#### FR-039: Financial Flow Orchestration
- **Feature ID:** FR-039
- **Feature Name:** Financial Flow Orchestration
- **Priority:** P0
- **Actor:** Sistem Finansial Otomatis, Finance Admin
- **Purpose:** Mengorkestrasi rantai aliran pemrosesan finansial secara berurutan dan transparan dari nilai kotor hingga masuk ke dompet pengguna dan pembukuan ledger.
- **Preconditions:** Alokasi bounty proyek kotor telah disahkan untuk seluruh anggota tim.
- **Postconditions:** Orkestrasi formula Gross - Tax - Farewell = Net dieksekusi tuntas dan dana terdistribusi ke dompet serta batch fund.
- **User Story:** Sebagai Sistem dan Tim Manajemen, saya ingin aliran dana terdistribusi mengikuti formula baku secara otomatis tanpa intervensi manual yang rawan kesalahan, sehingga integritas keuangan terjamin.
- **User Flow:** Trigger penyelesaian bounty -> Sistem menghitung Gross Bounty -> Memotong Pajak -> Memotong Kontribusi Farewell -> Menyetor ke Batch Fund -> Mengkreditkan Net ke Dompet Pengguna.
- **Functional Requirements:**
  1. Mengeksekusi formula aliran finansial platform resmi:
     $$\text{Gross Bounty} - \text{Tax/Deductions} - \text{Farewell Contribution} = \text{Net Distributable} \rightarrow \text{Personal Wallet}$$
  2. Memastikan pemrosesan transaksi berlangsung secara transaksional (ACID): jika salah satu langkah gagal (misal pencatatan ke wallet gagal), seluruh mutasi dibatalkan kembali (rollback).
  3. Menghubungkan secara otomatis output potongan pajak ke akun kewajiban perpajakan, potongan iuran ke Batch Fund, dan sisa bersih ke Personal Wallet pengguna.
  4. Memicu pembukuan ganda secara serentak ke Financial Ledger (FR-037).
- **Business Rules:**
  - BR-020: Dynamic Deduction Engine
  - BR-021: Farewell Fund Classification
  - BR-022: Itemized Wallet Ledger
  - BR-023: Mandatory Double-Entry Ledger
- **Validation:**
  - Formula matematika wajib konsisten: $\text{Gross} = \text{Deductions} + \text{Farewell} + \text{Net}$.
- **Error States:**
  - `500 Financial Orchestration Error`: Kegagalan atomik pada salah satu entitas penampung menyebabkan seluruh alur dibatalkan dan memicu flag peringatan sistem.
- **Empty States:**
  - Antarmuka orkestrasi menampilkan status antrean "Idle" saat tidak ada transaksi yang menunggu pemrosesan.
- **Acceptance Criteria:**
  - **Given** dana kotor proyek Rp5.000.000 siap cair, **When** orkestrator finansial dijalankan, **Then** sistem memecah nilai sesuai aturan deduksi aktif, mengkreditkan bagian batch ke Batch Fund, mengkreditkan bagian bersih ke dompet penerima, dan membukukan seluruh jurnal ledger dalam satu transaksi database tunggal.
- **Dependencies:** FR-032, FR-034, FR-035, FR-036, FR-037.
- **Audit Events:** `FINANCIAL_FLOW_EXECUTED`, `ORCHESTRATION_FAILED_ROLLBACK`.
- **Notifications:** Notifikasi rekap eksekusi distribusi ke tim Finance.
- **Data Entities:** `Wallet Transaction` (DATA-006), `Financial Ledger`, `Wallet`, `Batch Fund`.

---

### 8.9 DOCUMENTS & STORAGE

#### FR-040: ID Card Generation (Prioritas: P1)
- **Aktor:** HR Admin, Intern, Sistem Dokumen.
- **Tujuan:** Menghasilkan kartu identitas magang digital dan siap cetak yang dilengkapi kode identifikasi, data diri, foto, dan QR Code/NFC token.
- **Kebutuhan Utama:** Pembuatan dokumen ID Card otomatis saat status intern aktif. Template desain standar kartu. Integrasi QR Code terenkripsi untuk kebutuhan pemindaian kehadiran fisik di kantor. Ekspor PDF/gambar ke Cloudflare R2 bucket `id-cards/`.
- **Aturan Bisnis & Catatan:** Menggunakan aset foto profil resmi yang diunggah ke storage.
- **Entitas:** `User`, `Intern`, `File Metadata` (DATA-007).

---

#### FR-041: Certificate Engine (Prioritas: P1)
- **Aktor:** HR Admin, Super Admin, Lulusan (Alumni).
- **Tujuan:** Menerbitkan sertifikat kelulusan program magang digital resmi dengan tautan verifikasi keaslian publik berbasis kode QR.
- **Kebutuhan Utama:** Pembuatan sertifikat PDF otomatis saat intern lulus (`GRADUATED`). Penyertaan nomor registrasi sertifikat unik, tanda tangan digital pejabat berwenang, dan kode QR menuju URL validasi publik platform.
- **Aturan Bisnis & Catatan:** Keputusan Bisnis Terbuka `[PERLU KONFIRMASI — OQ-001, OQ-002]`: Desain template, layout, tanda tangan digital, skema penomoran, penandatangan yang berwenang, isi teks kelulusan, dan penyertaan Rank/XP menunggu konfirmasi Reihan/Business Owner.
- **Entitas:** `Certificate`, `User`, `Alumni`, `File Metadata` (DATA-007).

---

#### FR-042: Digital Portfolio Builder (Prioritas: P1)
- **Aktor:** Alumni, Intern, Rekruter Eksternal (Tampilan Publik).
- **Tujuan:** Mengompilasi rekam jejak deliverable proyek, tugas yang diselesaikan, keahlian yang terbukti, dan bukti kerja terverifikasi menjadi portofolio profesional digital yang dapat dibagikan secara publik.
- **Kebutuhan Utama:** Halaman portofolio digital dinamis per lulusan. Tampilan proyek nyata yang pernah diselesaikan, peran dalam tim, evidence tautan/screenshot, badge achievement, dan verifikasi kompetensi. Opsi visibilitas publik/privat dengan tautan slug unik (misal: `/portfolio/nama-pengguna`).
- **Aturan Bisnis & Catatan:** Menegakkan retensi nilai jangka panjang bagi alumni (BR-003).
- **Entitas:** `Portfolio`, `Alumni`, `Project`, `Task`, `Skill`.

---

#### FR-043: Document Reports Generator (Prioritas: P1)
- **Aktor:** Intern, Supervisor, HR Admin, Institusi Pendidikan.
- **Tujuan:** Menghasilkan dokumen laporan formal magang, logbook harian, dan ringkasan kehadiran dalam format PDF standar untuk keperluan laporan akademik sekolah/universitas.
- **Kebutuhan Utama:** Ekspor laporan berkala terfilter (mingguan, bulanan, akhir masa magang). Format mencakup data kehadiran resmi, rekap jam sesi kerja, ringkasan laporan tugas harian (`what_i_did`), kendala, dan tanda tangan digital supervisor.
- **Aturan Bisnis & Catatan:** Data laporan bersumber langsung dari logbook dan absensi resmi yang tidak dapat diubah secara sepihak.
- **Entitas:** `Work Report & Submission` (DATA-004), `Attendance Event Log` (DATA-001), `User`.

---

#### FR-044: Integrasi Cloudflare R2 Object Storage
- **Feature ID:** FR-044
- **Feature Name:** Integrasi Cloudflare R2 Object Storage
- **Priority:** P0
- **Actor:** Sistem Layanan Penyimpanan, Seluruh Pengguna Pengunggah Berkas
- **Purpose:** Menyediakan infrastruktur penyimpanan objek berkas berskala besar yang aman, hemat biaya, dan terisolasi dari database operasional melalui penyedia Cloudflare R2.
- **Preconditions:** Akun Cloudflare R2 terkonfigurasi dengan bucket, kredensial S3-compatible, dan policy akses yang sah.
- **Postconditions:** Berkas fisik tersimpan pada hierarki bucket R2 dan metadata berkas tercatat pada basis data relasional.
- **User Story:** Sebagai Pengembang Sistem dan Pengguna, saya ingin berkas dokumen, foto, kartu identitas, dan bukti tugas tersimpan di object storage terdedikasi dengan URL akses bertanda tangan sementara (presigned URL), sehingga database utama tetap ringan dan performan.
- **User Flow:** Klien meminta presigned upload URL -> Backend menerbitkan URL berbatas waktu -> Klien mengunggah langsung ke R2 -> Backend mencatat metadata berkas.
- **Functional Requirements:**
  1. Mengintegrasikan API penyimpanan kompatibel S3 menggunakan Cloudflare R2 sebagai repositori penyimpanan berkas utama.
  2. Menerapkan struktur partisi direktori bucket terstandarisasi:
     - `profile/` (foto profil pengguna)
     - `id-cards/` (dokumen kartu identitas digital)
     - `project/` (aset deskripsi proyek)
     - `task-evidence/` (tangkapan layar, dokumen, berkas tugas)
     - `logbooks/` (ekspor logbook cetak)
     - `rewards/` (gambar aset merchandise/voucher)
     - `certificates/` (dokumen sertifikat kelulusan digital)
     - `gallery/` (foto/media kegiatan program magang)
     - `reports/` (laporan performa dan analitik dokumen)
     - `attachments/` (berkas lampiran pendukung umum)
  3. Menegakkan batasan integritas database mutlak: Database SQL HANYA menyimpan metadata berkas resmi (DATA-007: `id`, `disk`, `path_key`, `filename`, `mime_type`, `size_bytes`, `metadata`). Penyimpanan biner BLOB di dalam database SQL DILARANG KERAS.
  4. Menerapkan arsitektur upload langsung menggunakan Presigned PUT URL dan pembatasan akses baca privat menggunakan Presigned GET URL bertanda tangan sementara.
- **Business Rules:**
  - BR-024: Cloudflare R2 Storage Separation: Database hanya menyimpan metadata berkas; objek fisik berada di R2.
- **Validation:**
  - File upload wajib melewati validasi ukuran maksimal dan pembatasan MIME type di sisi backend sebelum presigned upload URL diterbitkan.
- **Error States:**
  - `502 Bad Gateway / Storage Unavailable`: Gangguan konektivitas ke layanan Cloudflare R2.
  - `403 Forbidden`: Presigned URL kedaluwarsa atau signature tidak valid.
- **Empty States:**
  - Folder atau lampiran kosong menampilkan placeholder visual tanpa ada tautan storage yang aktif.
- **Acceptance Criteria:**
  - **Given** pengguna mengunggah berkas bukti tugas tugas berukuran 2 MB, **When** upload berhasil, **Then** berkas tersimpan di Cloudflare R2 pada direktori `task-evidence/`, dan database SQL menyimpan record pada tabel `file_metadata` tanpa adanya data biner di tabel.
- **Dependencies:** Cloudflare R2 Bucket Infrastructure, IAM API Credentials.
- **Audit Events:** `FILE_UPLOADED_METADATA_RECORDED`, `FILE_DELETED_FROM_R2`.
- **Notifications:** Tidak ada notifikasi pengguna; pencatatan galat penyimpanan ke sistem monitoring.
- **Data Entities:** `File Metadata Record` (DATA-007).

---

### 8.10 SYSTEM & INTEGRITY

#### FR-045: Unified Policy Engine
- **Feature ID:** FR-045
- **Feature Name:** Unified Policy Engine
- **Priority:** P0
- **Actor:** Super Admin, Sistem Internal Platform
- **Purpose:** Menyediakan mesin konfigurasi aturan bisnis runtime terpusat untuk mengatur kebijakan seluruh modul tanpa memerlukan perubahan kode program atau redeployment sistem.
- **Preconditions:** Super Admin memiliki wewenang konfigurasi kebijakan runtime sistem.
- **Postconditions:** Aturan kebijakan runtime aktif terbarui di database dan langsung berlaku bagi seluruh modul sistem.
- **User Story:** Sebagai Super Admin, saya ingin mengatur kebijakan jam toleransi kehadiran, matriks penalti deduksi XP, aturan pajak, dan batas lembur melalui tabel kebijakan terpusat, sehingga aturan operasional platform dapat beradaptasi secara fleksibel dan cepat.
- **User Flow:** Admin membuka Unified Policy Engine -> Memilih kategori kebijakan (XP, Penalti, Jadwal, Pajak) -> Mengubah parameter -> Menyimpan versi kebijakan baru -> Sistem memuat aturan baru.
- **Functional Requirements:**
  1. Mengelola repositori kebijakan terpusat yang mencakup domain kebijakan operasional:
     - `Attendance Policy` (jadwal, grace period, aturan absensi)
     - `Work Session Policy` (batas toleransi idle, heartbeat sesi)
     - `Break Policy` (jendela istirahat, toleransi keterlambatan)
     - `Overtime Policy` (ambang persetujuan, rasio kompensasi)
     - `XP Policy & Penalty Policy` (matriks perolehan poin dan potongan penalti)
     - `Rank Policy & Reward Policy` (ambang batas XP rank, isi paket reward)
     - `Tax / Deduction Policy` (formula pemotongan pajak dan kas bersama)
     - `Project Policy` (visibilitas, batas kapasitas, aturan marketplace)
  2. Menyimpan skema rekaman kebijakan resmi: `id` (UUID), `policy_domain` (String), `effective_date` (Date), `condition_rules` (JSONB, predikat evaluasi), `action_definitions` (JSONB, aksi eksekusi), `priority_order` (Integer), `is_active` (Boolean).
  3. Menyediakan fungsi resolver kebijakan pada waktu runtime yang mengevaluasi kondisi pengguna/transaksi terhadap aturan kebijakan aktif dengan prioritas tertinggi.
  4. Mencegah manipulasi aturan langsung di kode: seluruh logika bisnis variabel wajib merujuk pada Policy Engine.
- **Business Rules:**
  - BR-001: Dynamic RBAC: Roles dan permissions digerakkan oleh policy engine.
  - BR-012: Configurable XP Penalties: Seluruh deduksi XP digerakkan oleh konfigurasi policy.
  - BR-020: Dynamic Deduction Engine: Pajak dan potongan digerakkan oleh policy.
- **Validation:**
  - Struktur `condition_rules` dan `action_definitions` JSONB harus valid sesuai skema JSON Schema domain bersangkutan.
  - `effective_date` tidak boleh tumpang tindih secara ambigu untuk domain kebijakan yang sama tanpa pembeda urutan prioritas.
- **Error States:**
  - `422 Unprocessable Entity`: Format skema kondisi atau aksi kebijakan tidak valid.
- **Empty States:**
  - Jika belum ada kebijakan kustom dibuat, sistem menerapkan konfigurasi dasar bawaan platform (default fallback policy).
- **Acceptance Criteria:**
  - **Given** Super Admin memperbarui aturan penalti keterlambatan >30 menit dari -3 XP menjadi -4 XP pada tabel policy, **When** peristiwa keterlambatan terjadi berikutnya, **Then** XP Engine memproses penalti sebesar -4 XP secara otomatis tanpa modifikasi kode aplikasi.
- **Dependencies:** FR-001 (RBAC).
- **Audit Events:** `POLICY_CREATED`, `POLICY_VERSION_UPDATED`, `POLICY_TOGGLED`.
- **Notifications:** Peringatan audit kepada Super Admin saat kebijakan inti mengalami perubahan.
- **Data Entities:** `Policy`, `Audit Log`.

---

#### FR-046: Event-Driven Notification Center (Prioritas: P1)
- **Aktor:** Seluruh Peran Pengguna, Sistem Notifikasi.
- **Tujuan:** Mengirimkan notifikasi dan pengingat yang dipicu oleh peristiwa sistem secara tepat waktu kepada pihak-pihak terkait.
- **Kebutuhan Utama:** Pengelolaan pusat notifikasi dalam aplikasi (in-app notification center). Pengiriman peringatan real-time (tenggat tugas, persetujuan lembur, check-in terlambat, reward terbuka, aplikasi proyek diterima/ditolak). Status baca/belum dibaca (`is_read`). Saluran notifikasi eksternal (email/WhatsApp) berstatus `[TBD]`.
- **Aturan Bisnis & Catatan:** Notifikasi digerakkan oleh event bus terdistribusi.
- **Entitas:** `Notification`, `User`.

---

#### FR-047: Audit Logging & Security
- **Feature ID:** FR-047
- **Feature Name:** Audit Logging & Security
- **Priority:** P0
- **Actor:** Super Admin, Petugas Keamanan Sistem, Sistem Internal
- **Purpose:** Menyediakan pencatatan jejak audit yang komprehensif, terstruktur, dan tidak dapat diubah (append-only) untuk seluruh aktivitas kritis demi keamanan dan kepatuhan sistem.
- **Preconditions:** Peristiwa penting terkait keamanan, otorisasi, finansial, atau presensi terjadi di platform.
- **Postconditions:** Log audit dicatat pada tabel terisolasi yang hanya mengizinkan penambahan data (append-only).
- **User Story:** Sebagai Super Admin dan Auditor Keamanan, saya ingin setiap perubahan data sensitif, eksekusi hak akses, koreksi absensi manual, dan mutasi finansial tercatat dengan identitas pelaku dan stempel waktu akurat, sehingga investigasi keamanan dan audit kepatuhan dapat dilakukan secara presisi.
- **User Flow:** Aksi administratif/kritis dieksekusi -> Interceptor menangkap detail konteks (IP, user ID, payload, timestamp) -> Menulis ke Audit Log -> Mengunci entri.
- **Functional Requirements:**
  1. Merekam seluruh peristiwa keamanan dan mutasi data kritis:
     - Percobaan login, kegagalan autentikasi, eskalasi hak akses.
     - Perubahan konfigurasi peran dan hak akses pada RBAC Engine.
     - Pengajuan dan persetujuan permohonan koreksi kehadiran manual.
     - Perubahan parameter kebijakan pada Unified Policy Engine.
     - Seluruh transaksi moneter, penyesuaian saldo, dan pembukuan ledger.
  2. Menyimpan skema rekaman log audit lengkap: `id` (UUID), `user_id` (UUID aktor, nullable jika sistem), `event_name` (String), `resource_type` (String), `resource_id` (String), `ip_address` (String), `user_agent` (String), `old_state` (JSONB), `new_state` (JSONB), `timestamp` (Timestamp with timezone).
  3. Menegakkan prinsip immutability mutlak: tabel log audit DILARANG KERAS menyediakan operasi UPDATE, DELETE, atau TRUNCATE kepada peran apa pun (termasuk Super Admin).
  4. Menyediakan antarmuka pencarian dan pemfilteran log audit bagi Super Admin.
- **Business Rules:**
  - BR-023: Mandatory Double-Entry Ledger didukung oleh jejak audit finansial.
  - BR-025: Audit Trail Preservation on Corrections: Log historis asli tidak boleh diubah saat terjadi koreksi manual.
- **Validation:**
  - Nilai payload audit wajib mencatat `ip_address` dan `user_agent` yang valid pada setiap panggilan HTTP.
- **Error States:**
  - `500 Internal Server Error`: Kegagalan penulisan rekaman audit kritis wajib menggagalkan seluruh transaksi bisnis utama (audit failure blocks critical transaction).
- **Empty States:**
  - Antarmuka pencarian menampilkan: "Tidak ditemukan rekaman log audit yang sesuai dengan filter pencarian."
- **Acceptance Criteria:**
  - **Given** Supervisor menyetujui koreksi kehadiran intern, **When** aksi dieksekusi, **Then** satu entri baru tercatat pada tabel `Audit Log` memuat identitas supervisor, timestamp server, alamat IP, nilai data sebelum koreksi, dan nilai data sesudah koreksi.
- **Dependencies:** FR-001 (RBAC).
- **Audit Events:** `AUDIT_LOG_EXPORTED`, `SECURITY_BREACH_ATTEMPT_LOGGED`.
- **Notifications:** Peringatan darurat instan ke Super Admin jika terdeteksi aktivitas mencurigakan berulang (misal kegagalan login 10x berturut-turut).
- **Data Entities:** `Audit Log`, `User`.

---

#### FR-048: System Settings & Configurations (Prioritas: P1)
- **Aktor:** Super Admin, Admin.
- **Tujuan:** Mengelola konfigurasi operasional global aplikasi yang tidak tercakup dalam kebijakan dinamis.
- **Kebutuhan Utama:** Pengaturan nama platform, logo perusahaan, zona waktu default (Asia/Jakarta), batas retensi sesi login, konfigurasi kunci integrasi API pihak ketiga, dan mode pemeliharaan (maintenance mode).
- **Aturan Bisnis & Catatan:** Nilai sensitif wajib dienkripsi saat tersimpan di database.
- **Entitas:** `System Settings`, `Audit Log`.

---

### 8.11 INTELLIGENCE & INSIGHTS

#### FR-049: Analytics & Reporting (Prioritas: P1)
- **Aktor:** HR Admin, Super Admin, Finance, Supervisor.
- **Tujuan:** Menyediakan analitik visual organisasi mengenai tingkat kepatuhan kehadiran, produktivitas tugas proyek, laju serapan bounty, dan performa kohort batch.
- **Kebutuhan Utama:** Agregasi metrik organisasi (Attendance Rate, Work Session Accuracy, Project Velocity, Budget Burn Rate, Rasio Top Performer). Grafik visual perbandingan performa antar-batch. Ekspor laporan analitik ke format berkas Excel dan PDF.
- **Aturan Bisnis & Catatan:** Sumber data berbasis data historis yang telah terkonsolidasi.
- **Entitas:** `Performance Evaluation`, `Batch`, `Attendance Event Log`, `Financial Ledger`.

---

#### FR-050: AI Insights & Activity Intelligence (Prioritas: P2)
- **Aktor:** HR Admin, Supervisor, Project Manager, Super Admin.
- **Tujuan:** Menyediakan analisis prediktif berbasis kecerdasan buatan untuk mendeteksi anomali kinerja lebih dini, memberikan wawasan produktivitas objektif, dan memberikan rekomendasi peningkatan kompetensi.
- **Kebutuhan Utama:** Deteksi dini potensi kejenuhan kerja (burnout) atau penurunan disiplin peserta. Rekomendasi otomatis anggota tim terbaik untuk proyek baru berdasarkan riwayat keahlian. Ringkasan otomatis performa mingguan individu.
- **Aturan Bisnis & Catatan:** Sesuai BR-015: Seluruh saran dan analisis berbasis AI bersifat murni advisory/rekomendasi bagi penilai manusia. Provider AI dan arsitektur model berstatus `[TBD — OQ-009]`.
- **Entitas:** `Performance Evaluation`, `Work Session`, `User`, `Project`.

---

### 8.12 USER INTERFACE PORTALS & NAVIGATION

#### FR-051: Command Center (Dashboard Admin / Eksekutif) (Prioritas: P1)
- **Aktor:** Super Admin, Admin, HR Admin, Finance.
- **Tujuan:** Menyediakan kokpit monitoring operasional dan strategis tingkat tinggi bagi pemangku kepentingan utama melalui empat kuadran widget terdedikasi.
- **Kebutuhan Utama:**
  - **Widget TODAY:** Rincian headcount langsung: Jumlah Present, Working, On Break, Overtime, Absent.
  - **Widget ATTENTION:** Antrean butuh tindakan: Anomali kehadiran, Permohonan Cuti, Permohonan Lembur, Lamaran Proyek Baru, Peninjauan Tugas Tertunda, Laporan Kerja Menunggu Evaluasi, Tiket Klaim Reward.
  - **Widget PERFORMANCE:** Kurva pertumbuhan XP, Tingkat Ketepatan Kehadiran Organisasi, Laju Penyelesaian Proyek, Top Performer per Batch aktif.
  - **Widget FINANCE:** Ringkasan keuangan: Total Bounty Proyek Berjalan, Akumulasi Pajak/Deduksi Terkumpul, Total Saldo Personal Wallets, Saldo Terkini Batch Fund, Antrean Pembayaran Payout Tertunda.
- **Aturan Bisnis & Catatan:** Data diperbarui secara periodik mendekati real-time. Visibilitas widget dibatasi oleh scope RBAC pengguna (BR-001).
- **Entitas:** `Attendance Event Log`, `Project`, `Wallet`, `Performance Evaluation`.

---

#### FR-052: Portal "My Day" (Tampilan Harian Intern & Alumni)
- **Feature ID:** FR-052
- **Feature Name:** Portal "My Day" (Tampilan Harian Intern & Alumni)
- **Priority:** P0
- **Actor:** Intern, Alumni
- **Purpose:** Menyediakan antarmuka kerja layar tunggal (single-screen cockpit) yang terfokus bagi pelaksana untuk mengelola seluruh aktivitas kerja harian, pencatatan waktu, pemilihan tugas, dan pelaporan tanpa navigasi yang membingungkan.
- **Preconditions:** Pengguna (Intern atau Alumni) berhasil masuk ke platform pada hari kerja aktif.
- **Postconditions:** Pengguna dapat mengelola seluruh aktivitas harian (presensi, tugas, sesi kerja, jeda istirahat) dari layar tunggal.
- **User Story:** Sebagai Intern atau Alumni yang bekerja hari ini, saya ingin mengakses satu layar kerja utama yang menampilkan status kehadiran, tombol kontrol mulai/jeda/selesai kerja, tugas aktif, dan perolehan XP hari ini, sehingga saya dapat fokus bekerja secara produktif.
- **User Flow:** Intern membuka portal My Day -> Memeriksa status presensi -> Menekan [START WORK] -> Memilih tugas hari ini -> Menjalankan aktivitas kerja -> Menyerahkan ringkasan harian -> Sesi selesai.
- **Functional Requirements:**
  1. Menyediakan tampilan layar tunggal terintegrasi yang memuat:
     - **Status Kehadiran Live:** Indikator check-in hari ini, waktu kedatangan, dan status ketepatan waktu.
     - **Kontrol Sesi Kerja:** Tombol aksi dinamis `[START WORK]`, `[TAKE BREAK]`, `[RESUME WORK]`, dan `[END WORK]` yang terhubung langsung ke Work Session Engine (FR-009) dan Break State Engine (FR-010).
     - **Timer Sesi Berjalan:** Penghitung waktu digital durasi aktif sesi kerja saat ini.
     - **Pemilih Tugas Aktif (Active Task Selector):** Dropdown/pemilih tugas proyek atau tugas harian yang sedang dikerjakan secara riil.
     - **Timer Hitung Mundur Istirahat:** Tampilan durasi istirahat dan sisa waktu menuju batas akhir jendela istirahat terjadwal.
     - **Jadwal Kerja Acuan:** Garis waktu jadwal kerja normal dan jendela istirahat hari ini.
     - **Pelacak XP Harian:** Rangkuman perolehan XP dan penalti yang diperoleh pada hari berjalan.
     - **Modal Ringkasan Akhir Hari:** Modal pop-up yang muncul saat menekan tombol `[END WORK]` untuk mengisi pengiriman laporan kerja ringkas harian (FR-022) sebelum melakukan checkout fisik.
  2. Mengunci antarmuka kerja jika pengguna belum melakukan `CHECK_IN` kehadiran pada hari tersebut.
  3. Mengintegrasikan script Session Integrity Tracking (FR-011) di latar belakang halaman untuk memantau fokus tab dan idle.
- **Business Rules:**
  - BR-005: Work Type Partition: Daily Work dan Project Tasks dapat dipilih dari antarmuka ini.
  - BR-007: 5-Way Time Differentiation: Jam sesi kerja My Day terpisah dari jam kehadiran scanner.
  - BR-008: Work Session State Machine: Tombol aksi mematuhi siklus status kerja-istirahat.
- **Validation:**
  - Tombol `[START WORK]` hanya dapat ditekan jika status kehadiran `CHECK_IN` aktif dan belum ada sesi kerja lain yang berjalan.
  - Pengguna wajib memilih tugas atau menuliskan deskripsi pekerjaan sebelum sesi kerja dapat dimulai.
- **Error States:**
  - `400 Bad Request`: Menekan tombol aksi kerja di luar urutan status yang diizinkan.
- **Empty States:**
  - Jika pengguna belum memiliki tugas yang ditugaskan, pemilih tugas menampilkan opsi "Pekerjaan Harian / Belajar Mandiri (General Daily Work)".
- **Acceptance Criteria:**
  - **Given** intern telah berhasil scan check-in di kantor, **When** membuka portal My Day, **Then** tombol [START WORK] aktif, dan saat ditekan, timer berjalan, status berubah menjadi WORKING, dan opsi pause break tersedia.
- **Dependencies:** FR-008 (Attendance), FR-009 (Work Session), FR-010 (Break), FR-011 (Integrity), FR-021 (Task).
- **Audit Events:** `MY_DAY_SESSION_STARTED`, `MY_DAY_BREAK_TOGGLED`, `MY_DAY_SESSION_ENDED`.
- **Notifications:** Pengingat modal pengiriman laporan kerja harian saat sesi kerja diakhiri.
- **Data Entities:** `Work Session`, `Attendance Event Log` (DATA-001), `Task`, `Break`.

---

#### FR-053: Portal "Team Today" (Tampilan Supervisor) (Prioritas: P1)
- **Aktor:** Supervisor, Project Manager.
- **Tujuan:** Menyediakan dasbor operasional langsung bagi penyelia untuk memantau status keberadaan, keaktifan kerja, dan anomali anggota tim yang berada di bawah tanggung jawabnya secara real-time.
- **Kebutuhan Utama:**
  - **Rincian Headcount Tim:** Jumlah dan daftar anggota: Sedang Bekerja (`Working`), Sedang Istirahat (`On Break`), Belum Mulai (`Not Started`), Mangkir/Tidak Hadir (`Absent`), Sedang Lembur (`Overtime`).
  - **Peringatan Anomali Real-Time:** Peringatan kedatangan terlambat, idle berlebihan (>15 menit), istirahat tidak sah melampaui jadwal, permohonan lembur yang menunggu persetujuan, dan tugas yang melewati tenggat waktu.
  - **Aksi Cepat:** Tautan langsung untuk menyetujui lembur atau memvalidasi status kehadiran anggota tim.
- **Aturan Bisnis & Catatan:** Menampilkan data terfilter khusus untuk anggota tim yang dibawahi supervisor bersangkutan.
- **Entitas:** `Attendance Event Log`, `Work Session`, `Project Team`, `Overtime Request`.

---

#### FR-054: Portal Khusus Supervisor (Prioritas: P1)
- **Aktor:** Supervisor, HR Admin.
- **Tujuan:** Menyediakan ruang kerja administrasi komprehensif bagi supervisor untuk mengelola seluruh siklus evaluasi, persetujuan administratif, dan pemantauan deliverable tim.
- **Kebutuhan Utama:** Modul navigasi terpusat mencakup: Kehadiran Tim Saya (`My Team Attendance`), Pemantauan Proyek Aktif (`Active Projects`), Verifikasi Tugas (`Task Reviews`), Peninjauan Laporan Kerja (`Work Reports Approval`), Persetujuan Cuti (`Leave Approvals`), Persetujuan Lembur (`Overtime Approvals`), Penetapan Kontribusi Akhir Tim Proyek (FR-024), dan Pengisian Rubrik Evaluasi Kinerja (FR-027).
- **Aturan Bisnis & Catatan:** Mengonsolidasikan wewenang evaluasi berkala supervisor dalam satu ruang kerja.
- **Entitas:** `Project Team`, `Task`, `Work Report & Submission`, `Performance Evaluation`, `Leave Request`.

---

#### FR-055: Skema Navigasi Sidebar Global
- **Feature ID:** FR-055
- **Feature Name:** Skema Navigasi Sidebar Global
- **Priority:** P0
- **Actor:** Seluruh Peran Pengguna (Disesuaikan berdasarkan hak akses RBAC)
- **Purpose:** Menyediakan tata letak navigasi samping (sidebar-based navigation) terstandarisasi, konsisten, dan dinamis yang mencakup seluruh domain kapabilitas platform sesuai izin peran masing-masing.
- **Preconditions:** Pengguna berhasil melakukan otentikasi ke dalam sistem.
- **Postconditions:** Menu navigasi sidebar menampilkan daftar modul yang sesuai secara ketat dengan peran dan izin pengguna.
- **User Story:** Sebagai Pengguna Sistem, saya ingin memiliki menu navigasi samping yang terstruktur rapi dan hanya menampilkan menu yang relevan dengan hak akses saya, sehingga saya dapat bernavigasi dengan mudah dan cepat antar-modul.
- **User Flow:** Pengguna membuka aplikasi -> Sistem membaca permissions pengguna dari token/basis data -> Menyaring pohon menu sitemap -> Menampilkan menu yang diizinkan.
- **Functional Requirements:**
  1. Menerapkan struktur navigasi hierarkis 11 modul utama platform:
     - `COMMAND CENTER`
     - `PEOPLE` (Interns, Alumni, Batches, Institutions, Skills)
     - `ATTENDANCE` (Live Attendance, Scanner, Attendance Records, Work Sessions, Corrections, Leave, Overtime)
     - `PROJECTS` (Project Marketplace, My Projects, Applications, Projects, Tasks, Milestones, Submissions, Reports)
     - `PERFORMANCE` (Performance Overview, Evaluations, XP, Ranks, Achievements, Leaderboard, Top Performers)
     - `REWARDS` (Rank Rewards, Reward Claims, Reward History)
     - `FINANCE` (Wallets, Project Bounty, Batch Fund, Tax & Deductions, Payouts, Ledger)
     - `DOCUMENTS` (ID Cards, Certificates, Reports, Portfolio)
     - `MEDIA` (Gallery)
     - `INSIGHTS` (Analytics, AI Insights, Reports)
     - `SYSTEM` (Users, Roles & Permissions, Notifications, Audit Logs, Security, Storage, Settings)
  2. Merender item navigasi secara dinamis di sisi klien berdasarkan daftar permission dan scope pengguna yang diterbitkan oleh RBAC Engine (FR-001).
  3. Menegakkan aturan isolasi ketat bagi peran khusus:
     - Peran **Scanner Operator** HANYA menampilkan menu tunggal: `Attendance → Scanner`. Seluruh modul lainnya disembunyikan dan diisolasi.
     - Peran **Intern** difokuskan pada: `My Day`, `Attendance (Own)`, `Projects (Own + Marketplace)`, `Performance (Own)`, `Rewards (Own)`, `Finance (Wallet Own)`, `Documents (Own)`, `Media`.
     - Peran **Alumni** difokuskan pada: `My Day`, `Projects (Public Marketplace + Own)`, `Performance (Own)`, `Rewards (Own)`, `Finance (Wallet Own)`, `Documents (Own + Portfolio)`, `Media`.
- **Business Rules:**
  - BR-001: Dynamic RBAC: Visibilitas sidebar ditentukan oleh konfigurasi permission dinamis, bukan hardcode role.
  - BR-002: Scanner Operator Isolation: Operator pemindai hanya mengakses antarmuka pemindai kehadiran.
- **Validation:**
  - Setiap rute URL yang diakses melalui sidebar wajib divalidasi ulang oleh middleware otorisasi backend (tidak hanya menyembunyikan elemen UI).
- **Error States:**
  - `403 Forbidden`: Pengguna memanipulasi rute URL untuk membuka menu sidebar yang tidak menjadi haknya.
- **Empty States:**
  - Tidak ada menu yang kosong tanpa hak akses; sistem hanya menampilkan item yang memiliki izin baca minimal.
- **Acceptance Criteria:**
  - **Given** pengguna login dengan peran Scanner Operator, **When** sidebar dimuat, **Then** menu yang tampil hanya menu Scanner kehadiran, dan modul lain tidak ada dalam struktur DOM navigasi.
- **Dependencies:** FR-001 (RBAC & Permission Engine).
- **Audit Events:** `UNAUTHORIZED_ROUTE_ACCESS_ATTEMPT`.
- **Notifications:** Tidak ada notifikasi pengguna.
- **Data Entities:** `Role`, `Permission`, `User`.

---

---

## 9. Desain Pengalaman Pengguna (UX/UI Requirements)

> **BATASAN ARSITEKTURAL & ATURAN DESAIN:**  
> Bagian ini secara eksklusif mengatur **filosofi tema, karakter visual, arah estetika, hierarki informasi, representasi progresi gamifikasi, batasan etika gamifikasi, nada komunikasi, pemetaan atmosfer visual antar-peran, serta arketipe layar, komponen, interaksi, dan responsivitas**.  
> Sesuai arahan tata kelola produk fase ini, **TIDAK DIBUAT** rancangan wireframe detail, tata letak piksel spesifik, atau spesifikasi komponen kode front-end prematur. Bagian ini mendefinisikan prinsip visual dan arketipe sistemik yang menjadi fondasi fase desain UI/UX berikutnya.

### 9.1 Filosofi Tema: "Game-Inspired Professional"
Platform DCISP menerapkan tema visual **"Game-Inspired Professional"**, sebuah perpaduan cermat antara ketelitian, kejelasan, dan kepercayaan perangkat lunak enterprise SaaS kelas atas dengan kepuasan visual, umpan balik langsung, dan keterlibatan emosional dari sistem progresi permainan peran (Role-Playing Game).

1. **Bukan Permainan Anak, Melainkan Arena Pertumbuhan:** Desain menolak elemen kartun kekanak-kanakan, warna neon mencolok tanpa tujuan, atau ilustrasi kekanak-kanakan yang menurunkan martabat kerja profesional. Sebaliknya, antarmuka menghadirkan atmosfer ruang kendali (*cockpit*) modern yang memberikan rasa pencapaian, kehormatan, dan pengakuan formal atas setiap kontribusi kerja riil.
2. **Keseimbangan Dualitas (Enterprise Rigor + Gaming Delight):**
   - *Di sisi Enterprise:* Presisi data numerik, kepatuhan akuntansi ganda, tata kelola audit yang kokoh, keterbacaan tipografi tinggi, dan kejelasan status operasional.
   - *Di sisi Gaming:* Rasa petualangan yang jelas, indikator progresi yang memuaskan saat target tercapai, lencana kehormatan yang elegan, dan perayaan visual yang proporsional atas dedikasi kerja.

### 9.2 Konsep Pengalaman Pengguna: "Your Internship Is a Progression Journey"
Seluruh pengalaman pengguna dirancang berlandaskan narasi utama: **"Your Internship Is a Progression Journey"** (Masa Magang Anda adalah Perjalanan Progresi Diri).

```text
                      [ LEVELING UP ]
              Mastery / Rank Promotion (S / A)
                         ▲
                         │  • High-Quality Deliverables
                         │  • Verified Evidence
                         │  • Peer & Supervisor Trust
                         │
              [ QUESTS & SPECIALIZATIONS ]
               Project Delivery & Milestones
                         ▲
                         │  • Real-World Problem Solving
                         │  • Bounty Distribution
                         │  • Multi-Tier Contribution
                         │
             [ DAILY DISCIPLINE & DRILLS ]
             Attendance & Work Session Mastery
                         ▲
                         │  • Punctual Check-In Streaks
                         │  • Focused Work Sessions
                         │  • Clean State Transitions
                         │
                    [ THE NOVICE ]
                  Intern Onboarding
```

1. **Metafora Tingkatan Progresi:**
   - **The Novice (Intern Baru):** Masuk dengan potensi dasar, mempelajari ritme kerja, membangun disiplin kehadiran, dan memahami batasan profesional.
   - **The Practitioner (Daily Discipline):** Menguasai sesi kerja harian, menjaga streak kehadiran tepat waktu, bekerja fokus tanpa anomali, dan menyelesaikan tugas-tugas terpandu.
   - **The Specialist (Project Delivery):** Mengambil tanggung jawab proyek nyata di marketplace, bekerja sama dalam tim lintas disiplin, menghasilkan bukti deliverable berkualitas, dan menerima pembagian bounty yang adil.
   - **The Master / Elite (Kelulusan & Alumni):** Mencapai peringkat tertinggi, meraih predikat Top Performer kohort, memperoleh sertifikat dengan verifikasi digital, dan bertransisi menjadi Alumni tepercaya yang terus berkontribusi di ekosistem.
2. **Umpan Balik Tanpa Jeda:** Setiap tindakan kerja positif (hadir tepat waktu, menyelesaikan tugas dengan bukti terverifikasi) memberikan umpan balik progresi langsung yang terlihat nyata, menumbuhkan rasa kepemilikan atas perkembangan karier pribadi.

### 9.3 Karakter & Arah Visual (Visual Direction)
1. **Atmosfer Visual (Visual Mood):** Menampilkan karakter *Sleek, Authoritative, Focused, and Rewarding*. Terasa seperti antarmuka workstation rekayasa perangkat lunak modern dipadukan dengan panel instrumen wahana eksplorasi presisi tinggi.
2. **Palet Nuansa & Warna Simbolik:**
   - *Dominan Netral (The Canvas):* Warna latar abu-abu gelap terstruktur atau putih/slate bersih pada mode terang yang menjamin rasio kontras prima, menghilangkan kelelahan mata saat bekerja seharian.
   - *Aksen Keberhasilan (Achievement Gold / Amber):* Digunakan khusus untuk menunjukkan perolehan XP, kenaikan rank, pembukaan achievement, dan penghargaan Top Performer.
   - *Aksen Fokus & Produktivitas (Precision Cyan / Slate Blue):* Menandai status kerja aktif, penghitung waktu sesi kerja berjalan, dan indikator tugas prioritas.
   - *Aksen Kepatuhan & Disiplin (Status Emerald & Crimson):* Warna hijau tegas untuk ketepatan waktu/kehadiran sah; warna merah bata untuk pelanggaran keterlambatan/anomali tanpa memberikan kesan menghukum secara kasar.
3. **Tipografi Berkarakter Fungsional:**
   - Tipografi sans-serif modern yang dirancang untuk antarmuka pengguna teknis, memastikan angka-angka metrik waktu, jam lembur, nilai finansial, dan persentase kontribusi terbaca secara jernih pada berbagai resolusi layar.

### 9.4 Hierarki Visual (Visual Hierarchy & User Questions)
Tata letak visual dirancang untuk menjawab 7 pertanyaan kritis pengguna dalam hitungan detik:
1. **Apa yang harus saya kerjakan hari ini?** Dijawab oleh daftar tugas prioritas di portal "My Day".
2. **Bagaimana status pekerjaan saya saat ini?** Dijawab oleh widget timer status sesi (Working, Break, Idle).
3. **Bagaimana progres capaian saya?** Dijawab oleh indikator persentase milestone proyek.
4. **Berapa akumulasi XP dan performa saya?** Dijawab oleh XP Counter dan Rank Badge di header.
5. **Apa achievement atau reward yang saya peroleh?** Dijawab oleh notifikasi modal klaim reward baru.
6. **Apa hal penting yang membutuhkan perhatian saya?** Dijawab oleh widget ATTENTION di Command Center / Team Today.
7. **Bagaimana posisi progresi saya terhadap target kelulusan?** Dijawab oleh bilah progresi masa magang dan kurva evaluasi kinerja.

### 9.5 Representasi Visual Progresi (Progression Visuals) & Pemetaan Game-to-Product
Sistem memetakan konsep mekanik game ke dalam padanan profesional formal tanpa mengubah terminologi bisnis yang sah:

| Konsep Gamifikasi | Padanan Profesional Produk | Representasi Visual yang Digunakan |
|---|---|---|
| **XP (Experience Points)** | Progresi Dedikasi & Kinerja Kerja | Bilah Pengalaman (XP Bar) presisi dengan pecahan angka numerik (misal: 450/600 XP). |
| **Rank** | Jenjang Reputasi & Karier Magang | Lencana Logam Bertingkat (Rank Badges: D, C, B, A, S) dengan aksen elegan. |
| **Achievement** | Capaian Milestone Khusus | Lencana Prestasi Visual (Badge Cards) berstatus Unlocked / Locked. |
| **Quest** | Deliverable Proyek Terstruktur | Kartu Proyek Marketplace dengan tag keahlian dan batas tenggat. |
| **Daily Mission** | Pelaporan Kerja Harian (Daily Work) | Checklist Tugas Hari Ini pada portal My Day. |
| **Reward** | Paket Insentif Promosi | Kartu Paket Hadiah Multi-Komponen (Uang Kas, Merchandise, Hak Istimewa). |
| **Bounty** | Bagi Hasil Nilai Proyek Riil | Indikator Alokasi Finansial Berbasis Kontribusi Akhir (Rp). |
| **Leaderboard** | Peringkat Benchmark Sehat | Tabel Perbandingan Kinerja Sehat Global (tanpa mempermalukan individu). |
| **Streak** | Konsistensi Kedisiplinan Kehadiran | Indikator Hari Hadir Berturut-turut Tepat Waktu. |

### 9.6 Batasan Gamifikasi (Gamification Boundaries)
Gamifikasi dalam DCISP adalah **lapisan kinerja (performance layer)** pendorong motivasi kerja, bukan sistem manipulatif. Batasan etika berikut ditegakkan secara arsitektural:
1. **Dilarang Mekanik Adiktif Eksploitatif:** TIDAK ADA elemen *loot boxes*, *gacha*, roda keberuntungan untung-untungan (*gambling wheels*), atau mekanisme ketergantungan semu. Setiap reward dan poin XP berakar pada jam kerja riil dan deliverable terverifikasi.
2. **Pemisahan Mutlak Rank vs Performance Score:**
   - Sistem antarmuka dilarang menyamakan Rank gamifikasi dengan Skor Kinerja Evaluasi Formal ($	ext{Rank} 
eq 	ext{Performance Score}$).
   - Rank mencerminkan akumulasi aktivitas dan dedikasi waktu; Performance Score mencerminkan kualitas teknis, akurasi pekerjaan, dan penilaian profesional supervisor.
3. **Non-Toxic Competition:** Leaderboard dan peringkat disajikan untuk menginspirasi pencapaian standar keunggulan, bukan untuk mempermalukan peserta yang sedang belajar. Informasi penalti XP disajikan secara privat kepada individu terkait dan supervisor, tidak diumbar di ruang publik antarmuka.

### 9.7 Nada & Suara Komunikasi (Tone & Voice)
1. **Bahasa Profesional, Menghargai & Memberdayakan:** Komunikasi sistem menggunakan bahasa Indonesia baku yang formal namun hangat, objektif, dan mendorong perbaikan diri (contoh: *"Check-in Anda tercatat pada pukul 08:32 (Tepat Waktu)"* atau *"Terdeteksi jeda istirahat melebihi jadwal sebesar 12 menit. Harap tinjau kembali jadwal kerja Anda"*).
2. **Transparan & Tanpa Ambiguitas:** Setiap pemotongan penalti atau deduksi pajak finansial wajib disertai penjelasan penyebab dan formula perhitungannya secara lugas tanpa menyembunyikan rincian.
3. **Bebas Kalimat Menghakimi:** Kegagalan atau pelanggaran kehadiran disajikan sebagai fakta data operasional, bukan teguran personal atau bahasa yang menjatuhkan martabat peserta magang.

### 9.8 Pemetaan Visual Berdasarkan Domain & Peran
Atmosfer visual antarmuka beradaptasi secara dinamis sesuai dengan peran pengguna yang sedang aktif:

| Peran Pengguna | Atmosfer & Fokus Visual | Nada Pengalaman Pengguna |
|---|---|---|
| **Intern** | *Progression Cockpit & Focus.* Fokus pada portal My Day, visual bar XP yang tumbuh, visibilitas tugas aktif, dan pencapaian lencana. | Memberdayakan, terstruktur, membimbing, dan memotivasi pencapaian harian. |
| **Alumni** | *Professional Hub & Opportunity.* Fokus pada bursa proyek publik, portofolio digital yang dapat dibagikan, dan perolehan kompensasi bounty. | Kolegial, profesional mandiri, berorientasi peluang pasar kerja. |
| **Supervisor** | *Operational Radar & Review Deck.* Fokus pada pemantauan tim (Team Today), anomali kehadiran langsung, antrean persetujuan, dan evaluasi kontribusi. | Tajam, ringkas, berorientasi pengambilan keputusan cepat dan pembinaan tim. |
| **Project Manager** | *Delivery Board & Allocation Matrix.* Fokus pada kesehatan milestone, pembagian tugas kanban, keterisian kuota tim, dan verifikasi bukti kerja. | Terstruktur, berbasis hasil, analitis, dan kolaboratif. |
| **Finance** | *Audit Ledger & Financial Stream.* Fokus pada kejelasan neraca berpasangan, verifikasi pemotongan pajak, pengawasan kas bersama, dan persetujuan pencairan. | Presisi tinggi, konservatif, transparan, dan patuh kepatuhan audit. |
| **Scanner Operator** | *High-Contrast Terminal.* Antarmuka pemindai satu fungsi berkontras tinggi dengan umpan balik visual dan audio instan saat scan berhasil/gagal. | Minimalis, utilitarian murni, cepat, dan bebas distraksi menu lain. |
| **Super Admin** | *Command Center & Governance.* Kendali penuh konfigurasi kebijakan sistem, integritas log audit, pengaturan peran dinamis, dan analitik makro. | Otoritatif, menyeluruh, sistemik, dan terkontrol ketat. |

### 9.9 Prinsip Antarmuka Layar (Screen Archetypes)
Pengembangan UI fase berikutnya wajib mengacu pada 6 arketipe layar standar:
1. **Executive Command Deck (Command Center — FR-051):** Berorientasi widget modular dengan kepadatan data terukur (*dense but breathable*). Fokus pada 4 kuadran: Hari Ini (Today), Perhatian (Attention), Kinerja (Performance), dan Keuangan (Finance).
2. **Focused Single-Screen Cockpit ("My Day" — FR-052):** Layar tunggal tanpa scrolling berlebihan. Menempatkan kontrol aksi kerja utama (Start, Break, End) di area yang mudah dijangkau dengan visual status yang dominan.
3. **Operational Monitoring Grid ("Team Today" — FR-053):** Tata letak kartu status tim berbasis grid dengan kode warna tegas untuk memantau status anggota tim dalam satu pandangan mata.
4. **Bazaar / Marketplace Catalog (Project Marketplace — FR-016):** Tata letak katalog dua kolom/tiga kolom dengan kartu proyek informatif yang menonjolkan batas kuota, keahlian yang dibutuhkan, dan total bounty pool.
5. **Interactive Kanban / Task Flow (Tasks — FR-021):** Kolom status tugas terstruktur dengan dukungan drag-and-drop intuitif dan indikator bukti deliverable terlampir.
6. **Financial Ledger & Itemized Stream (Personal Wallet — FR-035):** Tampilan neraca saldo bersih di bagian atas diikuti aliran mutasi transaksi terperinci (*itemized transaction stream*) dengan penanda warna debit (merah/oranye) dan kredit (hijau).

### 9.10 Prinsip Komponen Visual (Component Archetypes & Tokens)
1. **Progression Bar Component:** Indikator kemajuan horizontal berbasis token variabel. Memiliki penanda target milestone, teks rasio numerik, dan animasi pengisian yang halus saat perolehan XP baru dicatat.
2. **Status Pill / Badge Tokens:** Label status berbasis warna semantik baku:
   - *Emerald (Success / On-Time / Active / Approved)*
   - *Amber (Warning / Pending / On-Break / Under-Review)*
   - *Crimson (Danger / Late / Rejected / Unauthorized)*
   - *Slate / Cyan (Neutral / Informational / Draft)*
3. **Precision Timer Component:** Penghitung waktu digital berbasis monospace font untuk memastikan angka detik dan menit tidak bergoyang saat waktu berjalan.
4. **Action Confirmation Modals:** Dialog konfirmasi tindakan kritis (seperti Check-out Akhir Hari atau Pengesahan Kontribusi) wajib menampilkan ringkasan konsekuensi data secara transparan sebelum tombol eksekusi ditekan.

### 9.11 Prinsip Interaksi & Umpan Balik (Interaction Patterns)
1. **Non-Intrusive Ambient Feedback:** Notifikasi sistem rutin (seperti check-in berhasil atau timer sesi aktif) disajikan melalui toast message non-intrusif atau indikator status di sudut layar tanpa memblokir alur kerja pengguna.
2. **Contextual Attention Alerts:** Notifikasi yang membutuhkan tindakan mendesak (peringatan jadwal istirahat berakhir, permohonan lembur masuk) disajikan dengan penanda visual berdenyut halus (*subtle pulse animation*) pada ikon navigasi terkait.
3. **Zero State / Empty State Guidance:** Setiap tampilan data yang belum memiliki rekam jejak (misal: dompet baru tanpa transaksi atau tugas kosong) wajib menyertakan ilustrasi profesional minimalis dan instruksi tindakan yang harus dilakukan pengguna berikutnya.

### 9.12 Perilaku Responsif & Batasan Perangkat (Responsive Behavior)
1. **Desktop Primary Workstation (1440px – 1920px+):** Layout utama untuk peserta magang, supervisor, PM, admin, dan finance. Sidebar navigasi terbuka penuh, dasbor multi-kolom, dan tabel data luas dengan pagination cepat.
2. **Compact Laptop Display (1024px – 1366px):** Sidebar navigasi dapat diciutkan secara otomatis menjadi mode ikon (*collapsed icon-only mode*), tabel data menggunakan scroll horizontal terkontrol.
3. **Dedicated Scanner Tablet (768px – 1024px):** Antarmuka pemindai presensi terminal dioptimalkan untuk layar sentuh tablet, dengan tombol sentuh berukuran minimal 48x48px untuk operasional cepat Scanner Operator.
4. **Batasan Tampilan Mobile (Mobile Constraint):** Pada layar ponsel cerdas (< 768px), antarmuka difokuskan untuk tugas esensial (melihat status hari ini di My Day, memindai QR code kehadiran, dan menerima notifikasi darurat). Layar manajemen kompleks (seperti konfigurasi Policy Engine atau evaluasi kontribusi multi-anggota) diarahkan untuk dibuka melalui komputer desktop/laptop.

---

---

## 10. Kebutuhan Data (Data Requirements)

### 10.1 Prinsip Desain Data & Distingsi Entitas Produk vs Tabel Basis Data
Dokumen ini mendefinisikan kebutuhan data pada tingkat **Entitas Produk (Product Entities)** yang berfokus pada kebutuhan domain bisnis, atribut esensial, relasi logis, dan aturan siklus hidup data. Hal ini secara tegas dibedakan dari perancangan teknis fisik basis data (*Physical Database Tables/DDL*):
1. **Entitas Produk (Product Entity):** Merepresentasikan konsep domain bisnis nyata (misal: *Project Application*, *Work Report*, *Attendance Event*, *Personal Wallet*) yang memuat data esensial, relasi, dan aturan validasi untuk memenuhi alur kerja pengguna tanpa memaksakan struktur tabel database fisik.
2. **Tabel Basis Data (Database Table):** Merupakan rancangan implementasi teknis tingkat rendah (seperti indexing B-tree, partitioning tabel PostgreSQL, query tuning, tipe DDL spesifik) yang dikelola oleh tim rekayasa perangkat lunak backend.
3. **Penyimpanan Berkas Fisik vs Metadata (BR-024):** Database hanya menyimpan metadata berkas (DATA-007); seluruh berkas fisik disimpan secara terpisah pada Cloudflare R2 Object Storage. Penyimpanan biner BLOB di dalam database SQL sangat dilarang.

### 10.2 Inventaris Entitas Komprehensif
Tabel berikut merangkum seluruh entitas data yang membentuk model informasi DCISP:

| No | Nama Entitas | Domain Sistem | Deskripsi Singkat | Tingkat Kepentingan | Relasi Utama |
|---|---|---|---|---|---|
| 1 | User | IDENTITY | Kredensial autentikasi dan identitas dasar pengguna platform. | Wajib (Kritis) | Role, Scope, Intern, Alumni |
| 2 | Role | IDENTITY | Definisi peran dinamis sistem. | Wajib (Kritis) | Permission, User |
| 3 | Permission | IDENTITY | Hak akses granular (Resource + Action + Scope). | Wajib (Kritis) | Role |
| 4 | Scope | IDENTITY | Batasan konteks otorisasi (Global, Batch, Project, Own, dll). | Wajib (Kritis) | Permission, User |
| 5 | Intern | PEOPLE | Profil spesifik dan status siklus hidup peserta magang aktif. | Wajib (Kritis) | User, Batch, Institution |
| 6 | Alumni | PEOPLE | Data profil permanen peserta magang yang telah lulus. | Wajib (Kritis) | User, Portfolio |
| 7 | Batch | PEOPLE | Kohort pengelompokan peserta magang. | Wajib (Kritis) | Intern, Batch Fund |
| 8 | Institution | PEOPLE | Data master institusi pendidikan asal peserta magang. | Pendukung | Intern |
| 9 | Skill | PEOPLE | Repositori katalog keahlian teknis dan non-teknis. | Pendukung | Project, User, Task |
| 10 | Work Schedule | WORKFORCE | Definisi spesifikasi jadwal acuan kerja dan toleransi (DATA-002). | Wajib (Kritis) | Policy, Attendance Event |
| 11 | Holiday | WORKFORCE | Kalender hari libur resmi pengecualian kehadiran. | Wajib (Kritis) | Work Schedule |
| 12 | Attendance Event | WORKFORCE | Catatan log peristiwa kehadiran append-only (DATA-001). | Wajib (Kritis) | User, Work Session, Device |
| 13 | Work Session | WORKFORCE | Sesi pelacakan durasi kerja nyata dan integritas fokus. | Wajib (Kritis) | User, Attendance Event, Task |
| 14 | Break | WORKFORCE | Rekaman status istirahat dan pelacakan anomali. | Wajib (Kritis) | Work Session |
| 15 | Overtime Request | WORKFORCE | Permohonan dan persetujuan jam lembur resmi. | Wajib (Kritis) | User, Work Schedule, Task |
| 16 | Leave Request | WORKFORCE | Permohonan dan persetujuan cuti/izin resmi. | Wajib (Kritis) | User, Attendance Event |
| 17 | Attendance Correction | WORKFORCE | Entitas rekonsiliasi koreksi absensi manual. | Wajib (Kritis) | Attendance Event, User |
| 18 | Device | WORKFORCE | Registry perangkat terminal pemindai resmi. | Wajib (Kritis) | Attendance Event |
| 19 | Project | PROJECTS | Inisiatif proyek kerja bursa dan penugasan (DATA-003). | Wajib (Kritis) | Project Team, Task, Bounty |
| 20 | Project Application | PROJECTS | Berkas lamaran kandidat untuk bergabung ke proyek. | Wajib (Kritis) | Project, User |
| 21 | Project Team | PROJECTS | Struktur keanggotaan dan proporsi kontribusi proyek. | Wajib (Kritis) | Project, User, Bounty |
| 22 | Milestone | PROJECTS | Target capaian perantara pekerjaan proyek. | Pendukung | Project, Task |
| 23 | Task | PROJECTS | Unit tugas pekerjaan spesifik yang ditugaskan. | Wajib (Kritis) | Milestone, User, Work Report |
| 24 | Assignment | PROJECTS | Relasi penugasan anggota tim ke tugas tertentu. | Pendukung | Task, User |
| 25 | Work Report & Submission | PROJECTS | Laporan kemajuan dan penyerahan tugas resmi (DATA-004). | Wajib (Kritis) | Task, User, Evidence |
| 26 | Submission | PROJECTS | Sub-entitas penyerahan deliverable tugas proyek. | Pendukung | Work Report, Task |
| 27 | Evidence | PROJECTS | Berkas bukti hasil kerja terverifikasi yang dilampirkan. | Wajib (Kritis) | Work Report, File Metadata |
| 28 | XP Rule | PERFORMANCE | Konfigurasi aturan perolehan dan penalti poin XP. | Wajib (Kritis) | Policy, XP Transaction |
| 29 | XP Transaction | PERFORMANCE | Buku besar mutasi perubahan poin pengalaman (XP). | Wajib (Kritis) | User, XP Rule |
| 30 | Rank | PERFORMANCE | Definisi jenjang peringkat gamifikasi pengguna. | Wajib (Kritis) | User, Reward |
| 31 | Achievement | PERFORMANCE | Katalog lencana penghargaan pencapaian khusus. | Pendukung | User, XP Transaction |
| 32 | Performance Evaluation | PERFORMANCE | Rekaman evaluasi kinerja formal komposit berkala. | Wajib (Kritis) | User, Supervisor, Batch |
| 33 | Reward | INCENTIVES | Definisi paket reward promosi rank dan pencapaian. | Wajib (Kritis) | Rank, Reward Claim |
| 34 | Reward Claim | INCENTIVES | Tiket operasional klaim reward fisik/moneter. | Wajib (Kritis) | Reward, User, Wallet |
| 35 | Wallet | FINANCE | Dompet digital penampung hak finansial pengguna. | Wajib (Kritis) | User, Wallet Transaction |
| 36 | Wallet Transaction | FINANCE | Mutasi transaksi dompet digital multi-sumber (DATA-006). | Wajib (Kritis) | Wallet, Financial Ledger |
| 37 | Batch Fund | FINANCE | Rekening penampung dana perpisahan bersama per kohort. | Wajib (Kritis) | Batch, Financial Ledger |
| 38 | Deduction / Tax Rule | FINANCE | Konfigurasi aturan pemotongan pajak dan biaya (DATA-005). | Wajib (Kritis) | Policy, Wallet Transaction |
| 39 | Payout | FINANCE | Permohonan pencairan saldo dompet ke rekening bank. | Wajib (Kritis) | Wallet, Financial Ledger |
| 40 | Financial Ledger | FINANCE | Buku besar akuntansi ganda abadi (double-entry). | Wajib (Kritis) | Wallet Transaction, Batch Fund |
| 41 | Notification | SYSTEM | Rekaman pesan notifikasi sistem kepada pengguna. | Pendukung | User |
| 42 | Audit Log | SYSTEM | Jejak rekam audit append-only peristiwa keamanan/data. | Wajib (Kritis) | User |
| 43 | Policy | SYSTEM | Master data aturan konfigurasi runtime sistem. | Wajib (Kritis) | XP Rule, Deduction Rule |
| 44 | File Metadata | STORAGE | Catatan metadata berkas di Cloudflare R2 (DATA-007). | Wajib (Kritis) | User, Evidence, Document |
| 45 | Certificate | DOCUMENTS | Berkas sertifikat kelulusan digital resmi terverifikasi. | Pendukung | Alumni, File Metadata |
| 46 | Portfolio | DOCUMENTS | Kompilasi portofolio profesional publik alumni. | Pendukung | Alumni, Project, Task |

---

### 10.3 Spesifikasi Detail Entitas Produk

#### Preservasi Skema DATA-001 hingga DATA-007 (Verbatim)

```
### DATA-001: Attendance Event Log
- id: UUID
- user_id: UUID
- event_type: Enum (ARRIVED, CHECK_IN, WORK_STARTED, BREAK_STARTED, BREAK_ENDED, WORK_RESUMED, OVERTIME_STARTED, OVERTIME_ENDED, CHECK_OUT)
- timestamp: Timestamp with timezone
- method: Enum (QR_CODE, NFC, ADMIN_SCANNER, MANUAL_CORRECTION)
- device_id: String / Reference
- location_context: String / JSON
- session_id: UUID (Nullable)
- metadata: JSONB
```

```
### DATA-002: Work Schedule Specification
- id: UUID
- name: String
- working_days: Array of Integers (mis. 1-5 untuk Sen-Jum)
- start_time: Time (mis. 08:30)
- end_time: Time (mis. 17:00)
- break_start: Time (mis. 12:00)
- break_end: Time (mis. 13:00)
- grace_period_minutes: Integer (mis. 10)
- overtime_policy_id: UUID
- holiday_calendar_id: UUID
```

```
### DATA-003: Project Entity
- id: UUID
- title: String
- description: Text
- owner_id: UUID
- visibility: Enum (INTERN_ONLY, PUBLIC, PRIVATE)
- required_skills: Array of UUIDs / Strings
- capacity: Integer
- accepted_count: Integer
- deadline: Timestamp
- bounty_pool: Decimal / Currency Amount
- status: Enum (DRAFT, PUBLISHED, IN_PROGRESS, COMPLETED, CANCELLED)
```

```
### DATA-004: Work Report & Submission
- id: UUID
- task_id: UUID
- user_id: UUID
- date: Date
- progress_percentage: Integer (0-100)
- what_i_did: Text
- evidence_type: Enum (GIT_COMMIT, SCREENSHOT, URL, DOCUMENT, ATTACHMENT)
- evidence_url_or_key: String
- problems: Text
- next_actions: Text
- status: Enum (SUBMITTED, UNDER_REVIEW, APPROVED, REVISION_REQUIRED)
```

```
### DATA-005: Deduction / Tax Rule
- id: UUID
- name: String
- type: Enum (INCOME_TAX, PROJECT_TAX, GRADUATION_TAX, WITHDRAWAL_TAX, ADMINISTRATIVE_FEE, PENALTY, OTHER_DEDUCTION)
- rate_type: Enum (PERCENTAGE, FIXED_AMOUNT)
- rate_value: Decimal
- calculation_basis: String / Enum
- minimum_amount: Decimal (Nullable)
- maximum_amount: Decimal (Nullable)
- effective_date: Date
- is_active: Boolean
```

```
### DATA-006: Personal Wallet Transaction
- id: UUID
- wallet_id: UUID
- source: Enum (PROJECT_BOUNTY, OVERTIME_BONUS, RANK_REWARD, TAX_DEDUCTION, BATCH_CONTRIBUTION, PAYOUT, OTHER)
- reference_id: String / UUID
- gross_amount: Decimal
- deduction_amount: Decimal
- net_amount: Decimal
- status: Enum (PENDING, COMPLETED, FAILED, REVERSED)
- created_at: Timestamp
- approved_by: UUID (Nullable)
```

```
### DATA-007: File Metadata Record (R2 Storage)
- id: UUID
- disk: String (mis. r2)
- path_key: String (mis. task-evidence/2026/09/uuid.png)
- filename: String
- mime_type: String
- size_bytes: BigInt
- metadata: JSONB
```

---

#### Spesifikasi Detail Entitas Tambahan

##### 1. User
- `id`: UUID (Primary Key)
- `email`: String (Unique, Indexed)
- `password_hash`: String
- `full_name`: String
- `avatar_file_id`: UUID (Nullable, FK ke File Metadata)
- `status`: Enum (`PENDING`, `ACTIVE`, `SUSPENDED`, `INACTIVE`)
- `created_at`: Timestamp with timezone
- `updated_at`: Timestamp with timezone

##### 2. Role
- `id`: UUID (Primary Key)
- `name`: String (Unique, misal: `SUPER_ADMIN`, `INTERN`, `SUPERVISOR`)
- `description`: String
- `is_system`: Boolean (Default: false)

##### 3. Permission
- `id`: UUID (Primary Key)
- `role_id`: UUID (FK ke Role)
- `resource`: String (misal: `projects`, `attendance`, `finance`)
- `action`: String (misal: `create`, `read`, `update`, `delete`, `approve`)
- `scope_type`: Enum (`GLOBAL`, `ORGANIZATION`, `BATCH`, `PROJECT`, `TEAM`, `SCANNER_ONLY`, `OWN`)

##### 4. Scope
- `id`: UUID (Primary Key)
- `name`: String
- `scope_type`: Enum (`GLOBAL`, `ORGANIZATION`, `BATCH`, `PROJECT`, `TEAM`, `SCANNER_ONLY`, `OWN`)
- `context_id`: UUID (Nullable, ID entitas spesifik seperti batch_id atau project_id)

##### 5. Intern
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User, Unique)
- `batch_id`: UUID (FK ke Batch)
- `institution_id`: UUID (FK ke Institution)
- `id_number`: String (NIM/NISN)
- `mentor_id`: UUID (Nullable, FK ke User / Supervisor)
- `status`: Enum (`APPLICANT`, `ONBOARDING`, `ACTIVE`, `ON_LEAVE`, `SUSPENDED`, `GRADUATED`, `TERMINATED`)
- `join_date`: Date
- `end_date`: Date
- `current_rank_id`: UUID (FK ke Rank)
- `internship_xp`: Integer (Default: 0)

##### 6. Alumni
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User, Unique)
- `batch_id`: UUID (FK ke Batch asal)
- `graduation_date`: Date
- `alumni_xp`: Integer (Default: 0)
- `certificate_id`: UUID (Nullable, FK ke Certificate)
- `is_public_profile`: Boolean (Default: true)

##### 7. Batch
- `id`: UUID (Primary Key)
- `batch_code`: String (Unique, misal: `BATCH-2026-01`)
- `name`: String
- `start_date`: Date
- `end_date`: Date
- `quota`: Integer
- `status`: Enum (`DRAFT`, `ACTIVE`, `COMPLETED`, `ARCHIVED`)
- `created_at`: Timestamp with timezone

##### 8. Institution
- `id`: UUID (Primary Key)
- `name`: String
- `address`: Text `[TBD]`
- `contact_person`: String `[TBD]`
- `email`: String `[TBD]`
- `phone`: String `[TBD]`

##### 9. Skill
- `id`: UUID (Primary Key)
- `name`: String (Unique)
- `category`: String (misal: Frontend, Backend, UI/UX, QA, Management)
- `description`: Text (Nullable)

##### 10. Holiday
- `id`: UUID (Primary Key)
- `holiday_calendar_id`: UUID
- `date`: Date
- `name`: String
- `is_national`: Boolean

##### 11. Work Session
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User)
- `date`: Date
- `start_time`: Timestamp with timezone
- `end_time`: Timestamp with timezone (Nullable)
- `gross_duration_seconds`: Integer (Default: 0)
- `active_duration_seconds`: Integer (Default: 0)
- `idle_duration_seconds`: Integer (Default: 0)
- `active_task_id`: UUID (Nullable, FK ke Task)
- `status`: Enum (`WORKING`, `BREAK`, `PAUSED`, `ENDED`)

##### 12. Break
- `id`: UUID (Primary Key)
- `session_id`: UUID (FK ke Work Session)
- `start_time`: Timestamp with timezone
- `end_time`: Timestamp with timezone (Nullable)
- `duration_seconds`: Integer (Default: 0)
- `is_anomaly_early`: Boolean (Default: false)
- `is_violation_late`: Boolean (Default: false)

##### 13. Overtime Request
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User)
- `project_id`: UUID (Nullable, FK ke Project)
- `task_id`: UUID (Nullable, FK ke Task)
- `date`: Date
- `requested_start`: Time
- `requested_end`: Time
- `approved_start`: Time (Nullable)
- `approved_end`: Time (Nullable)
- `actual_worked_minutes`: Integer (Default: 0)
- `reason`: Text
- `status`: Enum (`PENDING`, `APPROVED`, `REJECTED`, `COMPLETED`, `CANCELLED`)
- `reviewer_id`: UUID (Nullable, FK ke User)
- `review_notes`: Text (Nullable)

##### 14. Leave Request
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User)
- `leave_type`: Enum (`SICK`, `ACADEMIC`, `URGENT_PERSONAL`, `OTHER`)
- `start_date`: Date
- `end_date`: Date
- `reason`: Text
- `evidence_file_id`: UUID (Nullable, FK ke File Metadata)
- `status`: Enum (`PENDING`, `APPROVED`, `REJECTED`, `CANCELLED`)
- `reviewer_id`: UUID (Nullable, FK ke User)

##### 15. Attendance Correction
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User)
- `attendance_event_id`: UUID (Nullable, FK ke Attendance Event Log jika koreksi data existing)
- `target_date`: Date
- `proposed_event_type`: Enum (`ARRIVED`, `CHECK_IN`, `WORK_STARTED`, `BREAK_STARTED`, `BREAK_ENDED`, `WORK_RESUMED`, `OVERTIME_STARTED`, `OVERTIME_ENDED`, `CHECK_OUT`)
- `proposed_timestamp`: Timestamp with timezone
- `reason`: Text
- `evidence_file_id`: UUID (Nullable, FK ke File Metadata)
- `status`: Enum (`PENDING`, `APPROVED`, `REJECTED`)
- `reviewer_id`: UUID (Nullable, FK ke User)
- `reviewed_at`: Timestamp with timezone (Nullable)

##### 16. Device
- `id`: UUID (Primary Key)
- `device_identifier`: String (Unique, MAC / Serial / Machine UUID)
- `device_type`: Enum (`NFC_READER`, `BARCODE_CAMERA`, `ADMIN_TERMINAL`, `MOBILE_OPERATOR`)
- `location_name`: String
- `api_key_hash`: String
- `is_active`: Boolean (Default: true)
- `last_seen_at`: Timestamp with timezone

##### 17. Project Application
- `id`: UUID (Primary Key)
- `project_id`: UUID (FK ke Project)
- `user_id`: UUID (FK ke User)
- `cover_letter`: Text
- `skill_match_percentage`: Decimal (Nullable, komputasi informatif AI)
- `status`: Enum (`APPLIED`, `UNDER_REVIEW`, `SHORTLISTED`, `ACCEPTED`, `REJECTED`, `WITHDRAWN`)
- `applied_at`: Timestamp with timezone

##### 18. Project Team
- `id`: UUID (Primary Key)
- `project_id`: UUID (FK ke Project)
- `user_id`: UUID (FK ke User)
- `project_role`: Enum (`OWNER`, `MANAGER`, `SUPERVISOR`, `MEMBER`)
- `responsibility`: Text
- `planned_contribution_pct`: Decimal (Precision 5,2)
- `actual_contribution_pct`: Decimal (Precision 5,2, Nullable)
- `final_contribution_pct`: Decimal (Precision 5,2, Nullable)
- `is_locked`: Boolean (Default: false)

##### 19. Milestone
- `id`: UUID (Primary Key)
- `project_id`: UUID (FK ke Project)
- `title`: String
- `description`: Text
- `deadline`: Timestamp with timezone
- `weight_pct`: Decimal (Precision 5,2)
- `status`: Enum (`PENDING`, `IN_PROGRESS`, `COMPLETED`)

##### 20. Task
- `id`: UUID (Primary Key)
- `project_id`: UUID (FK ke Project)
- `milestone_id`: UUID (Nullable, FK ke Milestone)
- `assignee_id`: UUID (Nullable, FK ke User)
- `title`: String
- `description`: Text
- `estimated_hours`: Decimal (Precision 4,1)
- `difficulty_weight`: Integer (1-5)
- `priority`: Enum (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`)
- `deadline`: Timestamp with timezone
- `status`: Enum (`TODO`, `IN_PROGRESS`, `IN_REVIEW`, `COMPLETED`, `BLOCKED`, `CANCELLED`)

##### 21. Evidence
- `id`: UUID (Primary Key)
- `submission_id`: UUID (FK ke Work Report & Submission)
- `file_metadata_id`: UUID (Nullable, FK ke File Metadata jika berkas)
- `evidence_type`: Enum (`GIT_COMMIT`, `SCREENSHOT`, `URL`, `DOCUMENT`, `ATTACHMENT`)
- `external_url`: String (Nullable)
- `description`: Text (Nullable)

##### 22. XP Rule
- `id`: UUID (Primary Key)
- `policy_id`: UUID (FK ke Policy)
- `event_trigger`: String (misal: `CHECK_IN_ON_TIME`, `LATE_1_15`, `TASK_APPROVED`)
- `xp_value`: Integer (Bisa positif atau negatif)
- `xp_scheme`: Enum (`INTERNSHIP_XP`, `PROJECT_XP`, `ALUMNI_CONTRIBUTION`)
- `description`: String

##### 23. XP Transaction
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User)
- `xp_rule_id`: UUID (Nullable, FK ke XP Rule)
- `scheme`: Enum (`INTERNSHIP_XP`, `PROJECT_XP`, `ALUMNI_CONTRIBUTION`)
- `points`: Integer
- `running_balance`: Integer
- `reference_event`: String (misal: ID attendance event atau ID submission)
- `created_at`: Timestamp with timezone

##### 24. Rank
- `id`: UUID (Primary Key)
- `name`: String (misal: Rank F, Rank E, Rank D, Rank C, Rank B, Rank A, Rank S)
- `min_xp`: Integer (Ambang bawah XP)
- `level_order`: Integer (Unique sequence)
- `badge_icon_url`: String

##### 25. Achievement
- `id`: UUID (Primary Key)
- `code`: String (Unique)
- `title`: String
- `description`: Text
- `badge_icon_url`: String
- `reward_xp`: Integer (Default: 0)

##### 26. Performance Evaluation
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User)
- `evaluator_id`: UUID (FK ke User / Supervisor)
- `batch_id`: UUID (FK ke Batch)
- `period_type`: Enum (`WEEKLY`, `MONTHLY`, `END_OF_BATCH`, `CUSTOM`)
- `attendance_score`: Decimal
- `task_delivery_score`: Decimal
- `work_quality_score`: Decimal
- `supervisor_rubric_score`: Decimal
- `composite_performance_score`: Decimal
- `is_top_performer`: Boolean (Default: false)
- `notes`: Text
- `evaluated_at`: Timestamp with timezone

##### 27. Reward
- `id`: UUID (Primary Key)
- `rank_id`: UUID (Nullable, FK ke Rank)
- `achievement_id`: UUID (Nullable, FK ke Achievement)
- `component_type`: Enum (`CASH`, `BONUS_XP`, `PHYSICAL_ITEM`, `BADGE`, `VOUCHER`, `PLATFORM_PRIVILEGE`)
- `title`: String
- `monetary_value`: Decimal (Default: 0)
- `description`: Text

##### 28. Reward Claim
- `id`: UUID (Primary Key)
- `reward_id`: UUID (FK ke Reward)
- `user_id`: UUID (FK ke User)
- `status`: Enum (`ISSUED`, `CLAIMED`, `PROCESSING`, `FULFILLED`, `REJECTED`)
- `shipping_address`: Text (Nullable)
- `tracking_number`: String (Nullable)
- `claimed_at`: Timestamp with timezone (Nullable)
- `fulfilled_at`: Timestamp with timezone (Nullable)

##### 29. Wallet
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User, Unique)
- `current_balance`: Decimal (Default: 0, Constraint: >= 0)
- `created_at`: Timestamp with timezone
- `updated_at`: Timestamp with timezone

##### 30. Batch Fund
- `id`: UUID (Primary Key)
- `batch_id`: UUID (FK ke Batch, Unique)
- `total_accumulated`: Decimal (Default: 0)
- `current_balance`: Decimal (Default: 0)
- `created_at`: Timestamp with timezone

##### 31. Payout
- `id`: UUID (Primary Key)
- `wallet_id`: UUID (FK ke Wallet)
- `amount`: Decimal
- `bank_code`: String
- `account_number`: String
- `account_holder_name`: String
- `status`: Enum (`REQUESTED`, `APPROVED`, `PROCESSING`, `SETTLED`, `FAILED`, `REJECTED`)
- `settlement_reference`: String (Nullable)
- `created_at`: Timestamp with timezone

##### 32. Financial Ledger
- `id`: UUID (Primary Key)
- `transaction_id`: UUID (ID pengelompokan seimbang debit-kredit)
- `account_code`: String (misal: `1001-CASH`, `2001-LIABILITY-INTERN`, `2002-TAX-PAYABLE`, `3001-BATCH-FUND`)
- `direction`: Enum (`DEBIT`, `CREDIT`)
- `amount`: Decimal (Harus positif)
- `reference_table`: String
- `reference_id`: UUID
- `narration`: Text
- `created_at`: Timestamp with timezone (Immutable)

##### 33. Notification
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User)
- `title`: String
- `body`: Text
- `event_type`: String
- `action_url`: String (Nullable)
- `is_read`: Boolean (Default: false)
- `created_at`: Timestamp with timezone

##### 34. Audit Log
- `id`: UUID (Primary Key)
- `user_id`: UUID (Nullable, FK ke User)
- `event_name`: String
- `resource_type`: String
- `resource_id`: String
- `ip_address`: String
- `user_agent`: String
- `old_state`: JSONB (Nullable)
- `new_state`: JSONB (Nullable)
- `timestamp`: Timestamp with timezone (Immutable)

##### 35. Policy
- `id`: UUID (Primary Key)
- `policy_domain`: String (misal: `ATTENDANCE`, `XP`, `TAX`, `OVERTIME`)
- `name`: String
- `effective_date`: Date
- `condition_rules`: JSONB
- `action_definitions`: JSONB
- `priority_order`: Integer
- `is_active`: Boolean (Default: true)

##### 36. Certificate
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User)
- `certificate_number`: String (Unique)
- `template_version`: String `[PERLU KONFIRMASI — OQ-001]`
- `signer_name`: String `[PERLU KONFIRMASI — OQ-002]`
- `signer_title`: String `[PERLU KONFIRMASI — OQ-002]`
- `file_metadata_id`: UUID (FK ke File Metadata)
- `issued_at`: Date
- `verification_hash`: String (Unique)

##### 37. Portfolio
- `id`: UUID (Primary Key)
- `user_id`: UUID (FK ke User, Unique)
- `public_slug`: String (Unique)
- `bio`: Text
- `is_published`: Boolean (Default: false)
- `featured_projects`: Array of UUIDs
- `custom_theme`: JSONB (Nullable)

---

### 10.4 Relasi Entitas & Kardinalitas (Relationships & Foreign Keys)
1. **User $\rightarrow$ Intern (1 : 0..1):** Satu pengguna dapat terhubung ke tepat satu catatan profil intern aktif.
2. **User $\rightarrow$ Alumni (1 : 0..1):** Pengguna yang lulus bertransisi memiliki satu catatan alumni permanen.
3. **Batch $\rightarrow$ Intern (1 : N):** Satu kohort batch menampung banyak peserta magang; setiap intern terikat pada tepat satu batch.
4. **Batch $\rightarrow$ Batch Fund (1 : 1):** Setiap batch memiliki tepat satu entitas Batch Fund rekening bersama.
5. **Work Schedule $\rightarrow$ Attendance Event (1 : N):** Jadwal kerja menjadi dasar validasi seluruh peristiwa kehadiran terkait.
6. **User $\rightarrow$ Attendance Event Log (1 : N):** Satu pengguna memiliki banyak log peristiwa kehadiran append-only.
7. **Work Session $\rightarrow$ Break (1 : N):** Satu sesi kerja dapat memiliki banyak rentang status istirahat.
8. **Project $\rightarrow$ Project Team (1 : N):** Proyek memiliki banyak anggota tim dengan alokasi peran spesifik.
9. **Project $\rightarrow$ Milestone $\rightarrow$ Task (1 : N : M):** Proyek dibagi menjadi beberapa milestone, dan setiap milestone membawahi sekumpulan tugas.
10. **Task $\rightarrow$ Work Report & Submission (1 : N):** Satu tugas dapat menerima beberapa laporan kemajuan/penyerahan hingga disetujui.
11. **Work Report & Submission $\rightarrow$ Evidence (1 : N):** Setiap penyerahan tugas wajib menyertakan minimal satu artefak bukti.
12. **User $\rightarrow$ Wallet $\rightarrow$ Wallet Transaction (1 : 1 : N):** Setiap pengguna memiliki tepat satu dompet digital dengan banyak mutasi transaksi.
13. **Wallet Transaction $\rightarrow$ Financial Ledger (1 : N):** Setiap transaksi finansial dompet menghasilkan minimal sepasang entri debit dan kredit pada buku besar akuntansi ganda.
14. **File Metadata $\rightarrow$ R2 Object (1 : 1):** Satu record metadata merepresentasikan tepat satu objek biner fisik di bucket Cloudflare R2.

---

### 10.5 Aturan Integritas Data (Data Rules / BR Mapping)
1. **BR-001 (Dynamic RBAC Enforcement):** Seluruh evaluasi otorisasi wajib membaca relasi tabel `Role`, `Permission`, dan `Scope`. Tidak boleh ada pengecekan nama role string statis di controller/service.
2. **BR-003 & BR-004 (XP Scheme Isolation):** Saldo XP wajib disimpan terpisah pada 3 skema: `internship_xp`, `project_xp`, dan `alumni_xp`. Transaksi alumni dilarang mengubah nilai `internship_xp`.
3. **BR-006 (Central Attendance Ingestion):** Seluruh metode (QR, NFC, Manual) wajib menulis ke tabel tunggal `Attendance Event Log` (DATA-001).
4. **BR-016 (Three-Layer Contribution Balance):** Penjumlahan persentase pada masing-masing layer (`planned_contribution_pct`, `actual_contribution_pct`, `final_contribution_pct`) untuk seluruh anggota tim proyek wajib berjumlah tepat 100,00%.
5. **BR-020 & BR-021 (Farewell vs Tax Separation):** Pemotongan untuk iuran perpisahan wajib dimasukkan ke akun kredit `Batch Fund`, bukan ke akun `Tax Payable`.
6. **BR-022 & BR-023 (Immutable Double-Entry Balance):** Untuk setiap `transaction_id` pada tabel `Financial Ledger`, $\sum \text{Debit} - \sum \text{Kredit} = 0$. Operasi UPDATE dan DELETE dinonaktifkan di tingkat database trigger.
7. **BR-024 (R2 Binary Isolation):** Kolom database SQL dilarang bertipe BLOB biner. Seluruh objek file diarahkan ke `File Metadata` (DATA-007) yang menyimpan referensi path key Cloudflare R2.
8. **BR-025 (Audit Preservation on Corrections):** Persetujuan koreksi absensi manual wajib membuat event baru bertipe `MANUAL_CORRECTION` tanpa menghapus record kehadiran historis yang salah.

---

### 10.6 Model Status & Siklus Hidup Entitas (State Machines)

```
INTERN LIFECYCLE:
[APPLICANT] ──► [ONBOARDING] ──► [ACTIVE] ──► [GRADUATED] (Transisi ke Alumni)
                                   │   ▲
                                   ▼   │
                              [ON_LEAVE] / [SUSPENDED]
                                   │
                                   ▼
                              [TERMINATED]

WORK SESSION STATE:
[IDLE] ──► [WORKING] ◄───────────────┐
               │                     │
               ▼                     │
           [BREAK] ──────────────────┘
               │
               ▼
            [ENDED]

PROJECT STATUS:
[DRAFT] ──► [PUBLISHED] ──► [IN_PROGRESS] ──► [COMPLETED]
                │                  │
                ▼                  ▼
           [CANCELLED]        [CANCELLED]

TASK STATUS:
[TODO] ──► [IN_PROGRESS] ──► [IN_REVIEW] ──► [COMPLETED]
                ▲                 │
                │                 ▼
                └── [REVISION] ◄──┘

FINANCIAL MUTATION STATUS:
[PENDING] ──► [COMPLETED]
    │
    ├──► [FAILED]
    └──► [REVERSED] (Via Reversal Entry)
```

---

### 10.7 Retensi Data, Pengarsipan & Kebijakan Siklus Hidup Berkas
1. **Keberlanjutan Akun Pengguna:** Akun pengguna, catatan evaluasi, transaksi wallet, sertifikat, dan logbook kelulusan disimpan secara **permanen** (retensi tanpa batas waktu) guna mendukung retensi alumni dan validasi portofolio seumur hidup.
2. **Jejak Audit & Ledger Finansial:** Log audit keamanan (`Audit Log`) dan buku besar moneter (`Financial Ledger`) disimpan permanen dan dilarang di-purge atau dihapus.
3. **Penyimpanan Berkas di Cloudflare R2:**
   - Direktori `profile/`, `certificates/`, `id-cards/`, dan `task-evidence/` disimpan tanpa kedaluwarsa (permanent storage class).
   - Berkas ekspor laporan sementara pada direktori `reports/` dan berkas lampiran pendukung transien dapat menerapkan lifecycle rule R2: otomatis dihapus setelah 180 hari jika tidak ditautkan ke submission terverifikasi.

---

## 11. Kebutuhan Integrasi (Integration Requirements)

### 11.1 Arsitektur Layanan & Kontrak API Internal
1. **Prinsip Layanan Modular:** Sistem dibangun menggunakan arsitektur modular yang rapi (Modular Monolith atau Microservices berbasis domain) dengan pemisahan domain yang jelas: Identity, Workforce, Projects, Performance, Finance, Documents, dan Intelligence.
2. **Protokol Komunikasi:** Seluruh komunikasi antar-layanan dan klien frontend menggunakan RESTful API berbasis JSON dengan standar kepatuhan OpenAPI 3.0.
3. **Autentikasi & Otorisasi API:**
   - Autentikasi menggunakan standar Bearer Token (JSON Web Token - JWT) dengan masa berlaku akses singkat (access token 15 menit) dan rotasi refresh token yang aman.
   - Setiap panggilan endpoint API wajib melewati filter otorisasi gateway yang mengevaluasi permission dan context scope secara dinamis (FR-001).
4. **Rate Limiting & Throttling:** Menerapkan pembatasan laju panggilan (misal: 120 request/menit per IP/pengguna untuk API umum; 10 request/menit untuk endpoint autentikasi) guna mencegah serangan Denial of Service (DoS) dan brute force.

---

### 11.2 Integrasi Perangkat Kehadiran Fisik (Attendance Devices)

#### End-to-End Event Pipeline
Alur data kehadiran fisik dari penempelan kartu/pemindaian hingga pembaruan dasbor:

```
[ Reader Hardware (NFC / QR / Barcode) ]
                   │
                   ▼ (1) Serial / USB / TCP Payload Ingestion
[ Device Gateway Service ]
                   │
                   ▼ (2) Signed HTTPS POST Payload (Device Token)
[ Attendance API (/api/v1/attendance/events) ]
                   │
                   ▼ (3) Verify Signature & Debounce
[ Attendance Event Engine ]
                   │
                   ▼ (4) Evaluate Work Schedule & Tolerances
[ Schedule Validation Service ]
                   │
                   ▼ (5) Atomic Append-Only Storage (DATA-001)
[ Event Persistence (Database) ]
                   │
       ┌───────────┴───────────────────────────┐
       ▼                                       ▼
[ Real-Time Notification Bus ]          [ XP Rules Engine ]
       │                                       │
       ├──► Live Dashboard (My Day)            ├──► Calculate Points / Penalties
       └──► Team Today (Supervisor)            └──► Write to XP Ledger
```

#### Spesifikasi Penanganan Skenario Lapangan:
1. **Mode Offline & Buffer Sinkronisasi (Offline Buffer & Sync):**
   - Terminal scanner wajib memiliki memori penyimpanan lokal sementara (Flash/SQLite lokal).
   - Jika koneksi internet terputus, terminal tetap dapat membaca kartu NFC dan menyimpan antrean pemindaian secara lokal lengkap dengan timestamp perangkat saat kartu ditempelkan.
   - Saat koneksi pulih, terminal mengirimkan antrean data pemindaian secara batch dengan menyertakan flag `buffered_offline = true`.
2. **Mekanisme Retry dengan Exponential Backoff:**
   - Pengiriman data dari Device Gateway ke backend menerapkan retry otomatis dengan jeda eksponensial dan penambahan jitter acak jika menerima respon HTTP 5xx atau timeout.
3. **Pencegahan Pemindaian Ganda (Duplicate Scan Debouncing):**
   - Sistem menerapkan jendela debouncing waktu (threshold) minimal **30 detik**: jika UID kartu yang sama dipindai berulang kali dalam kurun waktu $\le 30$ detik di terminal yang sama, event kedua diabaikan dan dianggap sebagai pembacaan duplikat tanpa menghasilkan log ganda di database.
4. **Penanganan Kartu Tidak Valid (Invalid Card):**
   - Kartu dengan struktur payload rusak atau checksum tidak valid ditolak di tingkat hardware/gateway dengan sinyal indikator visual merah dan bunyi bip peringatan (buzzer error). Log error dicatat di terminal.
5. **Protokol Kartu Tidak Terdaftar (Unregistered Card):**
   - Jika kartu NFC valid secara format namun UID-nya tidak terdaftar pada basis data akun pengguna aktif, gateway mengirimkan event ke endpoint khusus yang mencatat event `UNKNOWN_TAG_SCANNED` pada audit keamanan, dan layar operator menampilkan pesan: "Kartu Belum Terdaftar. Hubungi Administrator".
6. **Konsistensi Timestamp & Penanganan Drift Waktu:**
   - Seluruh terminal pemindai wajib mengaktifkan sinkronisasi waktu jaringan standar (Network Time Protocol - NTP) dengan server waktu acuan setiap 1 jam.
   - Toleransi perbedaan waktu (time drift) maksimal antara perangkat dan server backend adalah 60 detik. Jika data offline disinkronkan, sistem memvalidasi urutan logis timestamp offline terhadap histori kehadiran yang telah ada.
7. **Idempotensi Event (Event Idempotency):**
   - Setiap payload pemindaian yang dikirim oleh perangkat wajib menyertakan kunci idempotensi unik:
     $$\text{Idempotency Key} = \text{SHA-256}(\text{device\_id} + \text{user\_id} + \text{event\_type} + \text{timestamp\_minute})$$
   - Backend menolak pemrosesan ulang jika kunci idempotensi yang sama telah berhasil dicatat sebelumnya.

---

### 11.3 Protokol Ingestion QR Code & NFC
1. **Dynamic Rotating QR Code:**
   - Check-in via QR Code menggunakan QR Code dinamis yang di-render pada layar aplikasi pengguna atau terminal kantor.
   - Token QR Code berisi string bertanda tangan kriptografis dengan batas masa berlaku (TTL) maksimal **30 detik** untuk mencegah pemalsuan presensi melalui tangkapan layar (anti-screenshot).
2. **NFC Card Payload Protocol:**
   - Membaca UID fisik kartu NFC (standar MIFARE / NTAG213 / NTAG215) melalui terminal reader terenkripsi.
   - UID kartu dipetakan ke identitas pengguna melalui tabel `Device Registry` atau relasi profil pengguna terotentikasi.

---

### 11.4 Integrasi Cloudflare R2 Object Storage
1. **Protokol Akses:** Menggunakan library klien resmi AWS S3 SDK yang diarahkan ke endpoint unik Cloudflare R2 (`https://<account_id>.r2.cloudflarestorage.com/<bucket_name>`).
2. **Arsitektur Presigned URL:**
   - Klien meminta presigned upload URL ke backend dengan menyertakan ukuran dan tipe MIME berkas.
   - Backend memvalidasi hak akses dan menerbitkan URL presigned PUT dengan masa kedaluwarsa maksimal **15 menit**.
   - Klien mengunggah langsung berkas biner ke Cloudflare R2.
   - Setelah upload selesai, klien mengonfirmasi ke backend untuk menyimpan rekaman metadata resmi (DATA-007).
3. **Keamanan Berkas:** Seluruh bucket dikonfigurasi sebagai bucket privat (private access). Berkas sensitif (evidence, ID card, laporan) hanya dapat diunduh melalui presigned GET URL bertanda tangan dengan TTL singkat (maksimal 60 menit).

---

### 11.5 Integrasi Payment & Payout Gateway
- **Status:** Kebutuhan Terbuka `[TBD — OQ-005]`.
- **Spesifikasi Antarmuka yang Diantisipasi:**
  - Integrasi API pencairan dana pihak ketiga (seperti Midtrans Iris, Xendit Payouts, atau Oy! Indonesia) untuk eksekusi transfer dana dari rekening operasional ke rekening bank atau e-wallet pengguna.
  - Penanganan Webhook Masuk: Endpoint `/api/v1/finance/payouts/callback` yang memproses status akhir transfer (`SUCCESS`, `FAILED`).
  - Transaksional Berpasangan: Callback keberhasilan pencairan secara otomatis mengubah status record `Payout` menjadi `SETTLED` dan mengunci pencatatan jurnal pengeluaran kas pada Financial Ledger.

---

### 11.6 Layanan Notifikasi Terdistribusi
1. **Pola Arsitektur (Pub/Sub):** Menggunakan antrean pesan asinkron (Message Queue / Event Bus) untuk mendistribusikan notifikasi tanpa membebani thread transaksi utama.
2. **Saluran Notifikasi (Multi-Channel):**
   - *In-App Notification:* Penyimpanan persisten ke tabel `Notification` dengan pembaruan UI real-time via WebSocket / Server-Sent Events (SSE).
   - *Email / Push Notification:* Diantisipasi untuk peringatan penting (persetujuan lembur, kelulusan, dan ringkasan pembayaran) `[TBD]`.

---

### 11.7 Integrasi AI & Intelligence Engine
- **Status:** Kebutuhan Terbuka `[TBD — OQ-009]`.
- **Spesifikasi Arsitektur yang Diantisipasi:**
  - Konektor layanan eksternal berbasis LLM / AI API (OpenAI / Claude / model lokal terdedikasi) untuk fitur FR-018 (Smart Skill Matching) dan FR-050 (AI Insights).
  - Skema Payload Permintaan: Pengiriman konteks anonim terkurasi (tanpa data pribadi sensitif) untuk menghasilkan saran persentase kecocokan keahlian atau ringkasan anomali produktivitas mingguan.
  - Sifat Layanan: Kegagalan koneksi ke penyedia AI dilarang menghentikan alur operasional utama platform (graceful degradation / fallback to algorithmic heuristics).

---

### 11.8 Integrasi Repositori Git (Evidence Tracking)
1. **Cakupan Penyedia:** Mendukung verifikasi tautan commit dan Pull Request dari penyedia repositori Git resmi (GitHub, GitLab).
2. **Metode Validasi:**
   - Parsing struktur URL commit/PR secara aman.
   - Opsi integrasi Webhook repositori proyek: Menerima event webhook commit/push dari GitHub/GitLab untuk mencocokkan hash commit yang dilampirkan intern pada submission tugas secara otomatis.

---

### 11.9 Sistem Webhook Masuk & Keluar (Webhooks)
1. **Keamanan Webhook (HMAC Signature):**
   - Seluruh payload webhook keluar (outbound webhooks) ditandatangani menggunakan algoritma HMAC-SHA256 dengan secret key yang dibagikan.
   - Header request menyertakan `X-DCISP-Signature` dan stempel waktu untuk mencegah replay attacks.
2. **Kebijakan Retry:** Webhook keluar yang gagal menerima respon HTTP 200 OK akan dicoba kembali secara otomatis dengan interval: 1 menit, 5 menit, 15 menit, dan 1 jam sebelum ditandai sebagai gagal permanen.

---

### 11.10 Penanganan Kegagalan, Ketahanan & Circuit Breakers
1. **Pola Circuit Breaker:** Menerapkan pola circuit breaker pada seluruh integrasi pihak ketiga (Cloudflare R2, Gateway Payout, Layanan AI). Jika tingkat kegagalan pemanggilan pihak ketiga melampaui 50% dalam jendela 1 menit, sirkuit terbuka (open circuit) dan sistem langsung mengembalikan status fallback yang aman tanpa membiarkan request klien menggantung (hang).
2. **Pencatatan Dead-Letter Queue (DLQ):** Event pesan asynchronous yang gagal diproses setelah batas maksimal retry dipindahkan ke Dead-Letter Queue untuk analisis dan intervensi manual tim rekayasa perangkat lunak.

---

## 12. Kebutuhan Non-Fungsional (Non-Functional Requirements)

### 12.1 Kinerja & Skalabilitas (Performance & Scalability)
1. **Waktu Respon API:**
   - Endpoint pemindaian kehadiran (`/api/v1/attendance/events`): Waktu respon rata-rata $\le 200\text{ ms}$ (95th percentile $\le 500\text{ ms}$) guna mencegah antrean penumpukan saat jam sibuk masuk kantor.
   - Panggilan API transaksional standar: Waktu respon rata-rata $\le 300\text{ ms}$.
   - Query analitik dan agregasi laporan kompleks: Waktu respon maksimal $\le 2000\text{ ms}$.
2. **Kapasitas Beban Bersamaan (Concurrency):** Sistem dirancang untuk mampu menangani minimal 500 pengguna aktif bersamaan (concurrent users) dan puncak lonjakan pemindaian kehadiran hingga 50 transaksi presensi per detik tanpa degradasi performa.
3. **Beban Data & Pengambilan Aset:** Waktu muat halaman pertama antarmuka pengguna $\le 2,0\text{ detik}$ pada koneksi internet broadband standar, dengan pemanfaatan CDN Cloudflare untuk caching aset statis antarmuka.

---

### 12.2 Keamanan Sistem (System Security & Access Control)
1. **Enkripsi Data Komprehensif:**
   - *Data in Transit:* Seluruh komunikasi jaringan eksternal dan internal wajib menggunakan enkripsi TLS 1.3 (HTTPS / WSS). Panggilan tanpa enkripsi langsung ditolak.
   - *Data at Rest:* Enkripsi basis data menggunakan standar industri AES-256 pada media penyimpanan database dan bucket Cloudflare R2.
2. **Penyimpanan Kredensial:** Kata sandi pengguna wajib di-hash menggunakan algoritma modern yang lambat terhadap serangan brute force (Argon2id atau BCrypt dengan work factor minimal 12). Kunci API perangkat dan token rahasia di-hash menggunakan SHA-256.
3. **Pertahanan Aplikasi Web (OWASP Top 10):**
   - Sanitasi data masukan secara ketat di tingkat controller backend untuk mencegah kerentanan SQL Injection, Cross-Site Scripting (XSS), dan Remote Code Execution.
   - Penerapan proteksi Cross-Site Request Forgery (CSRF) pada seluruh operasi state-changing berbasis cookie/session.
   - Penerapan header keamanan HTTP modern: Content Security Policy (CSP), X-Frame-Options: DENY, Strict-Transport-Security (HSTS), dan X-Content-Type-Options: nosniff.
4. **Isolasi RBAC Ketat:** Setiap endpoint controller backend wajib memvalidasi token otorisasi dan cakupan permission pengguna; validasi keamanan DILARANG hanya bersandar pada penyembunyian elemen UI frontend.

---

### 12.3 Privasi Pengguna & Integritas Kerja (Privacy-First Tracking Boundary)
Platform DCISP menjunjung tinggi etika privasi dan martabat pekerja. Batasan arsitektural berikut ditegakkan secara permanen pada level kode sumber:

1. **LARANGAN MUTLAK Alat Pengawasan Invasif:**
   - **TIDAK ADA Perekaman Layar (NO Screenshot Capture):** Sistem dilarang mengambil tangkapan layar desktop atau browser pengguna secara diam-diam.
   - **TIDAK ADA Perekam Papan Ketik (NO Keylogger):** Sistem dilarang merekam urutan ketukan tombol papan tik atau teks yang diketikkan pengguna di luar form isian resmi platform.
   - **TIDAK ADA Inspeksi Konten Tab (NO Tab Content Inspection):** Sistem dilarang membaca URL, DOM, riwayat pencarian, atau konten yang sedang dibuka pengguna pada tab browser lain.
   - **TIDAK ADA Pemindaian Aplikasi Latar Belakang (NO Desktop Process Scanning):** Sistem dilarang memindai daftar proses aplikasi yang sedang berjalan pada perangkat sistem operasi pengguna.
   - **TIDAK ADA Perekaman Papan Klip (NO Clipboard Recording):** Sistem dilarang membaca atau menyalin teks/gambar yang ada pada clipboard pengguna.
   - **TIDAK ADA Pembacaan Percakapan Pribadi (NO Chat Reading):** Sistem dilarang mengakses atau membaca pesan instan pengguna pada aplikasi perpesanan lain.
2. **Cakupan Pemantauan Integritas yang Diizinkan:** Sistem HANYA diizinkan memanfaatkan event standar peramban web:
   - Event `document.visibilityState` (mendeteksi apakah tab aktif DCISP sedang terlihat atau tersembunyi).
   - Event `window.onblur` dan `window.onfocus` (mendeteksi jendela aktif).
   - Pengatur waktu tidak aktif lokal (Idle Timer) berdasarkan ketiadaan aktivitas mouse/keyboard di dalam jendela platform DCISP melebihi 15 menit.
3. **Prinsip Non-Punitive Activity Tracking:** Kehilangan fokus jendela atau perpindahan tab tidak boleh memicu pengurangan poin XP secara otomatis; data fokus hanya disajikan sebagai metrik komputasi durasi aktif sesi kerja harian.

---

### 12.4 Keandalan & Ketersediaan (Reliability & High Availability)
1. **Target Ketersediaan Sistem (Uptime):** Ketersediaan operasional platform ditargetkan mencapai tingkat ketersediaan tinggi minimal **99,5%** setiap bulan (di luar jadwal pemeliharaan terencana).
2. **Toleransi Kegagalan (Fault Tolerance):** Kegagalan pada modul analitik non-kritis atau gangguan pada integrasi pihak ketiga (misal gateway payout atau AI) dilarang melumpuhkan fungsi dasar pencatatan kehadiran dan sesi kerja My Day.
3. **Integritas Transaksi Atomik (ACID):** Seluruh mutasi moneter dompet dan pencatatan buku besar akuntansi ganda wajib dibungkus dalam blok transaksi database ACID untuk mencegah terjadinya saldo gantung atau ketidakseimbangan neraca.

---

### 12.5 Observabilitas, Telemetri & Log Audit (Observability & Logging)
1. **Structured Logging:** Seluruh log aplikasi dicatat dalam format JSON terstruktur yang memuat atribut standar: `timestamp`, `level` (INFO, WARN, ERROR), `service_name`, `trace_id`, `user_id`, `path`, dan `message`.
2. **Log Audit Immutability:** Log audit keamanan (FR-047) disimpan pada tabel terisolasi dengan hak akses tingkat database yang hanya mengizinkan operasi `INSERT` (append-only), memblokir modifikasi atau penghapusan data audit oleh pengguna mana pun.
3. **Health Check Endpoints:** Menyediakan rute pemantauan `/health/liveness` dan `/health/readiness` yang memvalidasi kesiapan konektivitas database, cache, dan koneksi storage Cloudflare R2 secara real-time.

---

### 12.6 Kemudahan Pemeliharaan & Evolusi Sistem (Maintainability & Extensibility)
1. **Arsitektur Berbasis Kebijakan (Policy-Driven Architecture):** Mencegah penulisan logika ambang batas, persentase pajak, formula penalti XP, dan aturan jadwal kerja yang di-hardcode di dalam kode program; seluruh aturan wajib dapat dimodifikasi melalui antarmuka Unified Policy Engine (FR-045).
2. **Strategi Migrasi Basis Data:** Seluruh perubahan skema tabel database wajib dikelola melalui skrip migrasi berurutan (database migration versioning) yang terdokumentasi dan dapat diputar kembali (reversible).
3. **Modularitas Kode:** Pemisahan kode yang tegas antara lapisan presentasi (Controllers/Handlers), lapisan logika bisnis (Services/Use Cases), dan lapisan akses data (Repositories/Data Access Layer).

---

### 12.7 Aksesibilitas (Accessibility & Usability Standards)
1. **Kepatuhan Standar Aksesibilitas:** Antarmuka web mengacu pada pedoman standar WCAG 2.1 level AA, mencakup kontras warna teks terhadap latar belakang minimal rasio 4,5:1 untuk keterbacaan prima.
2. **Dukungan Navigasi Papan Ketik (Keyboard Navigation):** Seluruh elemen interaktif utama (tombol aksi My Day, menu navigasi sidebar, form pelaporan kerja) dapat dioperasikan secara penuh menggunakan tombol papan ketik (Tab, Enter, Escape, Arrow keys).
3. **Kompatibilitas Perangkat & Responsivitas:** Antarmuka dirancang responsif dan dapat dioperasikan secara optimal pada berbagai resolusi layar desktop dan tablet operator (lebar layar minimal 768px hingga 1920px+).

---

### 12.8 Pencadangan & Pemulihan Bencana (Backup & Disaster Recovery)
1. **Kebijakan Pencadangan Otomatis:**
   - Pencadangan penuh (full backup) database dilakukan secara otomatis setiap hari pada jam sepi beban operasional (misal pukul 02:00 WIB).
   - Pencadangan log transaksi (point-in-time recovery WAL logs) berjalan berkelanjutan setiap 15 menit.
   - Salinan berkas cadangan disimpan terenkripsi di luar lokasi infrastruktur utama (off-site cross-region storage).
2. **Target Metrik Pemulihan (RTO & RPO):**
   - **RPO (Recovery Point Objective):** `[PERLU KONFIRMASI — TBD]` (Target arsitektural yang diantisipasi: maksimal 15 menit kehilangan data transaksi).
   - **RTO (Recovery Time Objective):** `[PERLU KONFIRMASI — TBD]` (Target arsitektural yang diantisipasi: sistem dapat dipulihkan beroperasi kembali dalam waktu kurang dari 4 jam setelah insiden bencana fatal).
3. **Service Level Agreement (SLA) Resmi:** `[PERLU KONFIRMASI — TBD]` (Menunggu pengesahan formal tingkat manajemen Dagang Creative).

---

---

## 13. Analitik & Metrik Produk (Product & Operational Analytics)

### 13.1 Ikhtisar Arsitektur Analitik
Subsistem analitik DCISP dibangun di atas pemisahan operasional transaksional (OLTP) dan pembacaan analitik (Read-Model / Reporting Query Layer). Komponen pelaporan mengonsumsi log event transaksi (misal: *Attendance Event Log*, *Work Session Log*, *Project Ledger*, *Financial Transaction Ledger*) untuk menghasilkan wawasan teragregasi yang disajikan pada portal Command Center (FR-051), modul Analitik & Laporan (FR-049), dan modul AI Insights (FR-050).

Seluruh target numerik KPI spesifik tidak diada-adakan dan diberi label eksplisit `TBD` sesuai arahan tata kelola produk, menunggu ketetapan resmi dari Product Owner dan manajemen PT. Aplikasi Dagang Teknologi.

---

### 13.2 Analitik Produk (Product Analytics)
Analitik produk memantau adopsi, keterlibatan (*engagement*), dan retensi seluruh persona pengguna dalam ekosistem platform.

1. **Adopsi & Pertumbuhan Pengguna (User Adoption):**
   - **Metrik:** Total Pengguna Terdaftar, Pengguna Aktif Harian (*Daily Active Users* / DAU), Pengguna Aktif Bulanan (*Monthly Active Users* / MAU), Rasio DAU/MAU.
   - **Segmentasi Persona:** Intern, Alumni, Supervisor, Project Manager, Reviewer, Admin/HR, Finance, Scanner Operator.
   - **Target KPI:** `TBD`.
2. **Retensi & Keterlibatan Kohort (Cohort Retention & Engagement):**
   - **Metrik:** Kurva retensi mingguan peserta magang selama durasi program berjalan, durasi rata-rata sesi per pengguna di portal "My Day" (FR-052).
   - **Target KPI:** `TBD`.
3. **Pemanfaatan Fitur (Feature Adoption Rate):**
   - **Metrik:** Persentase pengguna yang mengakses fitur pelaporan kerja harian (Daily Work), Project Marketplace, klaim reward rank, dan dompet personal.
   - **Target KPI:** `TBD`.

---

### 13.3 Analitik Operasional (Operational Analytics)
Analitik operasional mengukur efisiensi alur kerja administratif dan ketepatan tata kelola supervisi.

1. **Kecepatan Tindak Lanjut Supervisi (Supervisor Turnaround Time):**
   - **Metrik:** Durasi waktu rata-rata (*mean time to resolve*) persetujuan lembur (*Overtime Approval* — FR-012), pengajuan cuti (*Leave Request* — FR-014), dan verifikasi koreksi absensi (*Attendance Correction* — FR-015).
   - **Target KPI:** `TBD`.
2. **Kepadatan Antrean & Beban Kerja (Backlog & Workload Distribution):**
   - **Metrik:** Jumlah item pending approval pada widget ATTENTION di Command Center (FR-051) per supervisor dan per batch.
   - **Target KPI:** `TBD`.
3. **Efisiensi Ingesti Presensi Scanner (Scanner Operator Throughput):**
   - **Metrik:** Kecepatan pemindaian rata-rata per peserta pada terminal fisik (detik/tap), rasio kegagalan pemindaian (*scan failure rate*).
   - **Target KPI:** `TBD`.

---

### 13.4 Analitik Kehadiran & Sesi Kerja (Attendance & Workforce Analytics)
Analitik kehadiran menyajikan data kepatuhan kedisiplinan kerja tanpa melanggar batasan etika privasi.

1. **Tingkat Kehadiran & Ketepatan Waktu (Attendance & Punctuality):**
   - **Metrik:** *Punctuality Rate* (% kehadiran tepat waktu sebelum jam masuk reguler + grace period), *Late Arrival Frequency*, *Absenteeism Rate*.
   - **Formula Ketepatan Waktu:**  
     $$\text{Punctuality Rate} = \left( \frac{\text{Total Sesi Tepat Waktu}}{\text{Total Sesi Jadwal Wajib}} \right) \times 100\%$$
   - **Target KPI:** `TBD`.
2. **Rasio Integritas Sesi Kerja (Work Session Integrity Ratio):**
   - **Metrik:** Rasio perbandingan antara Waktu Kerja Aktif (*Active Session Time*) terhadap Total Durasi Sesi Kerja Kotor (*Gross Work Session Time*).
   - **Formula Integritas Sesi:**  
     $$\text{Integrity Ratio} = \frac{\sum \text{Active Session Duration}}{\sum \text{Gross Session Duration}}$$
   - **Target KPI:** `TBD`.
3. **Kepatuhan Waktu Istirahat (Break Compliance):**
   - **Metrik:** Frekuensi anomali istirahat awal (*Early Break Anomaly*) dan pelanggaran istirahat berlebih (*Unauthorized Break Violation*).
   - **Target KPI:** `TBD`.
4. **Pemanfaatan Lembur (Overtime Utilization):**
   - **Metrik:** Total jam lembur aktual yang disetujui per peserta, rasio lembur terhadap jam kerja reguler.
   - **Target KPI:** `TBD`.

---

### 13.5 Analitik Proyek & Tugas (Project Analytics)
Analitik proyek memantau kecepatan pengiriman deliverable kerja dan kolaborasi tim.

1. **Kecepatan Penyelesaian Milestone (Milestone Velocity):**
   - **Metrik:** Persentase milestone proyek yang diselesaikan sebelum atau tepat tenggat waktu (*On-Time Milestone Delivery Rate*).
   - **Target KPI:** `TBD`.
2. **Konversi Pelamar Proyek (Application-to-Acceptance Ratio):**
   - **Metrik:** Jumlah lamaran per lowongan proyek di Marketplace, rata-rata waktu peninjauan lamaran oleh Project Manager.
   - **Target KPI:** `TBD`.
3. **Penyebaran Beban Tugas (Task Distribution & Cycle Time):**
   - **Metrik:** Rata-rata waktu penyelesaian tugas (*task cycle time*) dari status ASSIGNED hingga APPROVED, rasio tugas yang memerlukan revisi (*Revision Rate*).
   - **Target KPI:** `TBD`.

---

### 13.6 Analitik Kinerja & Gamifikasi (Performance & Gamification Analytics)
Analitik kinerja memantau efektivitas lapisan motivasi kerja dan pemerataan kontribusi.

1. **Distribusi Poin Pengalaman (XP Distribution by Track):**
   - **Metrik:** Total XP yang diterbitkan per track terpisah (*Internship XP*, *Project XP*, *Alumni Contribution*).
   - **Target KPI:** `TBD`.
2. **Kurva Progresi Peringkat (Rank Progression Velocity):**
   - **Metrik:** Waktu rata-rata yang dibutuhkan peserta magang untuk naik dari satu tingkatan Rank ke Rank berikutnya (misal: Rank D ke Rank C).
   - **Target KPI:** `TBD`.
3. **Pencapaian Lencana (Achievement Unlock Rate):**
   - **Metrik:** Persentase peserta yang berhasil membuka lencana prestasi tertentu, identifikasi lencana yang terlalu mudah atau terlalu sulit diraih.
   - **Target KPI:** `TBD`.
4. **Korelasi Evaluasi Kinerja (Performance vs XP Correlation):**
   - **Metrik:** Analisis korelasi antara skor evaluasi formal supervisor (*Performance Score*) dengan total akumulasi *Internship XP*, memastikan tidak terjadi deviasi ekstrem antara dedikasi waktu dan mutu hasil kerja nyata.
   - **Target KPI:** `TBD`.

---

### 13.7 Analitik Finansial & Perpajakan (Financial Analytics)
Analitik finansial menyediakan auditibilitas arus kas insentif dan kepatuhan penyisihan dana.

1. **Total Perputaran Kas Proyek (Bounty Flow Volume):**
   - **Metrik:** Akumulasi Gross Bounty yang dialokasikan, total Net Distributable yang dikreditkan ke Personal Wallet, dan saldo outstanding dompet pengguna.
   - **Target KPI:** `TBD`.
2. **Realisasi Pemotongan Pajak (Tax Withholding Summary):**
   - **Metrik:** Total pemotongan pajak penghasilan yang terkumpul per periode pelaporan perpajakan.
   - **Target KPI:** `TBD`.
3. **Pertumbuhan Dana Kas Angkatan (Batch Fund Growth):**
   - **Metrik:** Saldo terkumpul dana perpisahan angkatan (*Farewell Contributions*) per batch magang aktif.
   - **Target KPI:** `TBD`.
4. **Metrik Pencairan Dana (Payout Processing Time):**
   - **Metrik:** Rata-rata durasi penyelesaian permohonan pencairan dana (*payout request*) hingga transfer bank sukses dilakukan.
   - **Target KPI:** `TBD`.

---

### 13.8 Analitik Alumni (Alumni Analytics)
Analitik alumni mengukur kesinambungan hubungan talenta pasca-magang.

1. **Tingkat Keaktifan Alumni (Alumni Engagement Index):**
   - **Metrik:** Persentase akun alumni yang tetap aktif masuk ke platform minimal satu kali dalam 90 hari pasca-kelulusan.
   - **Target KPI:** `TBD`.
2. **Partisipasi Proyek Publik Alumni (Alumni Project Participation):**
   - **Metrik:** Jumlah proyek publik yang dikerjakan oleh alumni, total bounty yang berhasil diperoleh alumni.
   - **Target KPI:** `TBD`.
3. **Pemanfaatan Portofolio Digital (Portfolio Showcase Views):**
   - **Metrik:** Jumlah kunjungan publik ke tautan portofolio digital alumni yang diterbitkan oleh sistem DCISP.
   - **Target KPI:** `TBD`.

---

### 13.9 Taksonomi & Spesifikasi Event Tracking
Sistem mengimplementasikan skema pelacakan event terstruktur yang memancarkan payload JSON standar ke message broker/event engine dengan skema formal: `Event`, `Trigger`, `Actor`, `Properties`, `Purpose`.

| Event | Trigger | Actor | Properties | Purpose |
|---|---|---|---|---|
| `attendance.scanned` | Pemindaian kartu NFC / kode QR pada terminal | Intern, Scanner Operator | `user_id`, `method`, `scanner_device_id`, `timestamp`, `schedule_id`, `result` | Mencatat keberadaan fisik di kantor untuk audit kehadiran formal. |
| `session.work_started` | Pengguna mengklik tombol "Start Work" di My Day | Intern, Alumni | `user_id`, `session_id`, `timestamp`, `device_info` | Memulai perhitungan durasi sesi kerja resmi hari ini. |
| `session.break_started` | Pengguna mengklik tombol "Take Break" di My Day | Intern | `user_id`, `session_id`, `timestamp`, `is_early_anomaly` | Mencatat status istirahat dan mendeteksi anomali jeda lebih awal. |
| `session.break_ended` | Pengguna mengklik tombol "Resume Work" di My Day | Intern | `user_id`, `session_id`, `timestamp`, `duration_minutes`, `is_late_penalty` | Menutup status jeda dan memicu penalti jika melewati batas waktu. |
| `session.work_ended` | Pengguna mengklik tombol "End Work" di My Day | Intern, Alumni | `user_id`, `session_id`, `timestamp`, `work_session_duration`, `active_duration` | Mengunci ringkasan jam kerja harian dan memicu kalkulasi agregat. |
| `session.focus_lost` | Event peramban window onblur / tab hidden | Intern (Client) | `user_id`, `session_id`, `timestamp`, `unfocused_duration_sec` | Mengukur integritas sesi aktif secara non-invasif (tanpa screenshot). |
| `project.applied` | Pengguna mengirimkan formulir lamaran proyek | Intern, Alumni | `project_id`, `applicant_id`, `applied_at`, `quota_standing` | Memantau volume minat pelamar terhadap bursa proyek. |
| `project.submission` | Pengguna mengunggah laporan kerja dan bukti tugas | Intern, Alumni | `task_id`, `user_id`, `file_keys`, `report_text`, `submitted_at` | Menyerahkan hasil kerja ke antrean penelaahan Reviewer. |
| `contribution.confirmed` | Supervisor mengesahkan proporsi kontribusi akhir | Supervisor | `project_id`, `member_id`, `planned_pct`, `actual_pct`, `final_pct`, `approver_id` | Menetapkan dasar persentase pembagian insentif bounty proyek. |
| `xp.mutated` | Pemicu aturan XP (kehadiran, tugas, evaluasi, penalti) | System Event Engine | `user_id`, `track`, `amount`, `source_type`, `source_id`, `balance_after` | Mencatat mutasi perolehan/deduksi XP pada buku besar XP immutabel. |
| `rank.promoted` | Akumulasi XP melampaui ambang batas rank berikutnya | System Event Engine | `user_id`, `old_rank`, `new_rank`, `reward_package_id`, `timestamp` | Memicu perayaan kenaikan level dan penerbitan reward rank. |
| `bounty.distributed` | Eksekusi penyelesaian bagi hasil proyek selesai | Project Manager, Finance | `project_id`, `recipient_id`, `gross`, `tax_cut`, `batch_cut`, `net_amount` | Mengalirkan dana bounty ke Deduction Engine dan Personal Wallet. |
| `wallet.transaction` | Terjadi mutasi kredit atau debit pada dompet | Finance, System Engine | `wallet_id`, `user_id`, `direction`, `amount`, `source_type`, `ledger_entry_id` | Merekam mutasi saldo personal wallet dengan referensi buku besar. |
| `payout.settled` | Eksekusi transfer penarikan dana dompet berhasil | Finance Administrator | `payout_id`, `user_id`, `amount`, `destination_account`, `settled_at` | Menyelesaikan permohonan pencairan saldo kas pengguna. |
| `certificate.issued` | Penerbitan sertifikat kelulusan peserta magang | HR Admin, System | `certificate_id`, `user_id`, `batch_id`, `verification_hash`, `signer_id` | Menerbitkan dokumen kelulusan digital resmi berverifikasi QR. |

---

### 13.10 Kamus & Definisi Formal Metrik (Metric Dictionary)

| ID Metrik | Nama Metrik | Definisi Konseptual & Formula | Satuan | Target KPI |
|---|---|---|---|---|
| **MTR-ATT-01** | Attendance Punctuality Rate | Persentase kehadiran tepat waktu terhadap total sesi wajib: $(\sum \text{OnTime} / \sum \text{Required}) \times 100\%$ | Persen (%) | `TBD` |
| **MTR-ATT-02** | Work Session Integrity Ratio | Rasio waktu aktif terhadap waktu sesi kotor: $\sum \text{ActiveDuration} / \sum \text{GrossDuration}$ | Rasio (0.0 – 1.0) | `TBD` |
| **MTR-ATT-03** | Break Anomaly Frequency | Total kemunculan flag Early Break dan Unauthorized Break per pengguna per bulan | Jumlah Kejadian | `TBD` |
| **MTR-PRJ-01** | Milestone On-Time Delivery | Persentase pengiriman milestone sebelum batas waktu: $(\sum \text{OnTimeMilestones} / \sum \text{Milestones}) \times 100\%$ | Persen (%) | `TBD` |
| **MTR-PRJ-02** | Task Cycle Time | Durasi waktu dari tugas berstatus `IN_PROGRESS` hingga `APPROVED` oleh reviewer | Jam / Hari | `TBD` |
| **MTR-PRF-01** | XP Growth Velocity | Rata-rata akumulasi penambahan XP mingguan per peserta aktif | XP / Minggu | `TBD` |
| **MTR-PRF-02** | Evaluation Quality Index | Rata-rata skor evaluasi kinerja formal berkala supervisor pada skala 0 – 100 | Angka Indeks | `TBD` |
| **MTR-FIN-01** | Net Payout Turnaround Time | Durasi waktu dari pengajuan payout dompet hingga dana ditransfer oleh bagian keuangan | Jam Kerja | `TBD` |
| **MTR-ALU-01** | Alumni Engagement Ratio | Persentase alumni aktif berpartisipasi dalam bursa proyek publik: $(\text{Active Alumni} / \text{Total Alumni}) \times 100\%$ | Persen (%) | `TBD` |

---

## 14. Ketergantungan Sistem (System Dependencies)

### 14.1 Peta Ketergantungan Alur Bisnis Inti (Core Business Flow Dependency Map)

```text
[ 1. IDENTITY & RBAC ]
          │
          ▼
   [ 2. PEOPLE ] (Batches, Interns, Institutions)
          │
          ▼
 [ 3. WORKFORCE ] (Schedules, Attendance Engine, Work Sessions, Breaks)
          │
          ▼
  [ 4. PROJECTS ] (Marketplace, Teams, Milestones, Tasks, Work Reports, Evidence)
          │
          ▼
[ 5. PERFORMANCE ] (XP Rules Engine, Ranks, Formal Evaluation, Top Performer)
          │
          ▼
 [ 6. INCENTIVES ] (Rank Rewards, Project Bounty Distribution, Reward Claims)
          │
          ▼
   [ 7. FINANCE ] (Deduction Engine, Batch Fund, Personal Wallets, Double-Entry Ledger, Payout)
          │
          ▼
    [ 8. ALUMNI ] (Alumni Accounts, Public Marketplace, Digital Portfolio Builder)
```

### 14.2 Ketergantungan Produk (Product Dependencies)
1. **Otorisasi Terhadap Seluruh Modul:** Modul Workforce, Projects, Performance, dan Finance bergantung mutlak pada kestabilan Identity & RBAC (FR-001) untuk evaluasi izin runtime dan pembatasan scope.
2. **Kompensasi Terhadap Pengesahan Kontribusi:** Distribusi bounty proyek (FR-032) bergantung mutlak pada penyelesaian seluruh tugas dan pengesahan Final Contribution % (FR-024) oleh supervisor.
3. **Progresi Rank Terhadap Log Kerja Terverifikasi:** Promosi rank (FR-026) bergantung pada akumulasi Internship XP yang bersumber dari kehadiran sah (FR-008) dan tugas yang disetujui (FR-022).
4. **Portofolio Alumni Terhadap Rekam Jejak Deliverable:** Digital Portfolio Builder (FR-042) bergantung pada riwayat proyek yang selesai dan bukti kerja fisik yang tersimpan di R2 (FR-023, FR-044).

### 14.3 Ketergantungan Fitur (Feature Dependencies)
- FR-008 (Attendance Ingestion) $\longrightarrow$ FR-007 (Work Schedule Engine), FR-013 (Device Registry).
- FR-009 (Work Session) $\longrightarrow$ FR-008 (Attendance Event CHECK_IN).
- FR-010 (Break State Engine) $\longrightarrow$ FR-009 (Work Session Active State).
- FR-012 (Overtime Tracking) $\longrightarrow$ FR-007 (Work Schedule), Persetujuan Supervisor.
- FR-017 (Project Application) $\longrightarrow$ FR-016 (Project Marketplace Published).
- FR-024 (Contribution Engine) $\longrightarrow$ FR-021 (Tasks Completed), FR-022 (Reports Approved).
- FR-026 (Rank Progression) $\longrightarrow$ FR-025 (XP Rules Engine).
- FR-032 (Bounty Distribution) $\longrightarrow$ FR-024 (Final Contribution confirmed), FR-034 (Tax Engine).
- FR-035 (Personal Wallet) $\longrightarrow$ FR-037 (Financial Ledger immutabel).
- FR-039 (Financial Flow) $\longrightarrow$ FR-034 (Tax Engine), FR-035 (Wallet), FR-036 (Batch Fund).

### 14.4 Ketergantungan Eksternal (External & Environment Dependencies)
1. **Lingkungan Peramban Klien (Client Browser Environment):** Sistem bergantung pada peramban modern yang mendukung W3C Page Visibility API (`document.visibilityState`) dan Window Focus API untuk pelacakan integritas sesi kerja etis (FR-011).
2. **Perangkat Keras Pembaca Terminal (Scanner Hardware):** Sistem bergantung pada ketersediaan reader NFC/RFID dan kamera pemindai barcode/QR yang terhubung ke Device Gateway via protokol USB/Serial/WebSocket (FR-013).
3. **Repositori Git Publik/Organisasi:** Bukti kerja berupa tautan commit/PR bergantung pada ketersediaan layanan hosting Git eksternal (misal: GitHub / GitLab).

### 14.5 Ketergantungan Infrastruktur (Infrastructure Dependencies)
1. **Penyimpanan Objek Cloudflare R2 (BR-024, FR-044):** Ketergantungan primer untuk penyimpanan berkas fisik bukti deliverable, foto profil, dan dokumen laporan. Kegagalan API R2 memengaruhi fitur unggah/unduh berkas.
2. **Basis Data Relasional Transaksional (ACID RDBMS):** Ketergantungan mutlak untuk penegakan transaksi atomik pada buku kas ganda immutabel (FR-037) dan evaluasi kebijakan otorisasi dinamis.
3. **In-Memory Cache & Message Queue (Redis / Broker):** Dibutuhkan untuk penanganan debouncing pemindaian presensi ganda, pemrosesan event asinkron notifikasi, dan pengelolaan sesi terdistribusi.

### 14.6 Ketergantungan Operasional & Pihak Ketiga (Third-Party Dependencies)
1. **Mitra Gateway Pembayaran (Payout Gateway — OQ-005):** `[Status: TBD]` Menunggu keputusan formal pemilihan penyedia layanan transfer bank/e-wallet untuk otomasi pencairan dana dompet.
2. **Penyedia Layanan Kecerdasan Buatan (AI Provider — OQ-009):** `[Status: TBD]` Ketergantungan fitur P2 (FR-050) terhadap penyedia model LLM eksternal (OpenAI, Claude, atau model open-source lokal).
3. **Kepatuhan Kebijakan Perusahaan:** Keabsahan formula pemotongan pajak (FR-034) dan tata kelola kas angkatan (FR-036) bergantung pada keputusan legal/manajemen PT. Aplikasi Dagang Teknologi.

---

## 15. Risiko & Asumsi (Risks & Assumptions)

### 15.1 Register Risiko Komprehensif (Comprehensive Risk Register)

| ID | Risiko | Probabilitas | Dampak | Rencana Mitigasi | Pemilik (Owner) | Status |
|---|---|---|---|---|---|---|
| **RSK-001** | Kompleksitas Unified Policy Engine (FR-045) memicu kegagalan runtime jika terjadi kesalahan konfigurasi aturan. | Sedang | Tinggi | Terapkan sandbox validator, simulasi dry-run execution, skema versioning aturan, dan fallback otomatis ke konfigurasi aman. | Lead Architect | Open |
| **RSK-002** | Kerancuan antara Rank XP dengan Performance Score formal (BR-017) membingungkan evaluasi mutu kerja. | Tinggi | Sedang | Pisahkan antarmuka visual secara mutlak; cantumkan definisi eksplisit pada tooltip dan laporan evaluasi formal. | Product Owner | Open |
| **RSK-003** | Kerentanan pembulatan nilai finansial atau selisih pembukuan pada Financial Ledger (FR-034, FR-037, BR-023). | Rendah | Tinggi | Gunakan tipe fixed-point integer (satuan Rupiah utuh tanpa desimal mengambang), wajibkan transaksi ACID ganda selisih Rp 0. | Finance Lead | Open |
| **RSK-004** | Inkonsistensi data log kehadiran akibat multi-method ingestion (NFC, QR, Manual) yang masuk bersamaan. | Sedang | Sedang | Rute seluruh masukan melalui SATU Central Attendance Engine (BR-006) dengan timestamp server dan kunci idempotensi unik. | Tech Lead | Open |
| **RSK-005** | Putusnya konektivitas jaringan terminal scanner fisik di lobi kantor saat jam sibuk kedatangan. | Sedang | Sedang | Implementasikan local storage queue pada Device Gateway dengan penandatanganan payload offline dan sinkronisasi otomatis saat online. | Infrastructure | Open |
| **RSK-006** | Potensi manipulasi absensi offline saat koneksi pulih (pemalsuan timestamp lokal perangkat). | Rendah | Tinggi | Wajibkan verifikasi timestamp kriptografis perangkat keras atau flag manual review jika selisih waktu offline melebihi batas toleransi. | Security Lead | Open |
| **RSK-007** | Pelanggaran batas privasi pekerja akibat pelacakan fokus peramban (BR-010, BR-011) yang memicu resistensi pengguna. | Sedang | Tinggi | Tegaskan pembatasan kode: dilarang screenshot, keylogger, dan pembacaan chat. Publikasikan piagam transparansi privasi sistem. | Legal & HR | Open |
| **RSK-008** | Sengketa pembagian kontribusi proyek (FR-024, BR-016) jika anggota merasa Actual Contribution tidak adil. | Tinggi | Sedang | Sediakan ruang mediasi melalui alur 3-lapis di mana Supervisor memiliki kewenangan penyesuaian manual (Final Contribution %). | PM Lead | Open |
| **RSK-009** | Keterlambatan integrasi gateway pembayaran eksternal (FR-038, OQ-005) menahan pencairan dompet. | Sedang | Sedang | Siapkan alur pencairan manual (Manual Bank Transfer Batch) dengan konfirmasi bukti upload slip transfer sebelum otomasi gateway aktif. | Finance Lead | Open |
| **RSK-010** | Ketidakpastian spesifikasi tata kelola sertifikat kelulusan digital (FR-041, OQ-001, OQ-002). | Sedang | Rendah | Sediakan template generator berbasis HTML-to-PDF fleksibel yang parameternya dapat diperbarui dinamis melalui konfigurasi sistem. | Business Owner | Open |
| **RSK-011** | Kegagalan atau biaya tinggi API penyedia AI pihak ketiga (FR-050, OQ-009) mengganggu operasional. | Sedang | Rendah | Posisikan fitur AI sebagai modul P2 non-kritis murni informatif (BR-015); sistem inti wajib berjalan 100% tanpa dependensi AI. | Tech Lead | Open |
| **RSK-012** | Inkonsistensi pencatatan event `WORK_ENDED` antara diagram alur WF-001 dengan enum event FR-008 (OQ-014). | Rendah | Sedang | Segera lakukan pengesahan arsitektural apakah `WORK_ENDED` ditambahkan ke enum resmi atau diwakili oleh transisi status sesi kerja. | System Architect | Open |
| **RSK-013** | Beban kueri pelaporan analitik berat menurunkan performa transaksi operasional harian. | Sedang | Sedang | Pisahkan basis data transaksional (OLTP) dengan basis data replika analitik pembacaan agregat (Read-Replica / Materialized Views). | Database Admin | Open |

---

### 15.2 Register Asumsi Produk (Product Assumptions)

| ID | Asumsi | Dasar Pemikiran & Konteks | Dampak Jika Asumsi Tidak Terpenuhi | Status Validasi |
|---|---|---|---|---|
| **ASM-001** | Akses Internet Memadai | Platform berbasis web membutuhkan konektivitas internet stabil di lingkungan kantor operasional PT. ADT. | Peserta tidak dapat memperbarui laporan tugas dan live dashboard tertunda. | Terkonfirmasi |
| **ASM-002** | Ketersediaan Terminal Scanner | Perusahaan menyediakan minimal satu perangkat pemindai (tablet/laptop) berdedikasi di pintu masuk kantor. | Terjadi antrean presensi fisik dan penurunan akurasi waktu kedatangan. | Terkonfirmasi |
| **ASM-003** | Penerbitan Kartu Fisik NFC/QR | Setiap peserta magang baru menerima kartu identitas fisik atau kode QR unik saat hari pertama onboarding. | Metode presensi fisik terhambat dan bergantung penuh pada koreksi manual. | Terkonfirmasi |
| **ASM-004** | Supervisi Harian Aktif `[Inferred]` | Supervisor secara teratur memantau portal Team Today dan meninjau pengajuan minimal sekali dalam 24 jam. | Penumpukan antrean lembur/koreksi dan keterlambatan pembagian bounty proyek. | Asumsi Terbuka |
| **ASM-005** | Kepatuhan Regulasi Pajak RI `[Inferred]` | Rumus pemotongan pajak mengikuti aturan perpajakan Indonesia (PPh 21 atas imbalan magang/tenaga lepas). | Perlu perombakan formula perhitungan pada Deduction Engine. | Asumsi Terbuka |
| **ASM-006** | Akun Cloudflare R2 Tersedia | Kuota dan bandwidth Cloudflare R2 dialokasikan memadai untuk seluruh berkas bukti kerja dan dokumen. | Unggah bukti tugas gagal dan dokumen tidak dapat diunduh. | Terkonfirmasi |
| **ASM-007** | Zona Waktu Seragam (WIB / UTC+7) | Seluruh jadwal kantor mengacu pada Waktu Indonesia Barat (UTC+7) dengan penyimpanan basis data berbasis UTC. | Kerancuan perhitungan keterlambatan dan catatan sesi kerja lintas zona waktu. | Terkonfirmasi |
| **ASM-008** | Tidak Ada Gaji Pokok Karyawan Tetap | Platform tidak digunakan untuk memproses gaji bulanan karyawan tetap PT. ADT (Non-Goals). | Ruang lingkup membengkak dan membutuhkan modul payroll korporat terpisah. | Terkonfirmasi |
| **ASM-009** | Kepatuhan Peramban Modern W3C | Pengguna mengoperasikan peramban web modern yang mendukung Page Visibility API dan standar web terkini. | Pelacakan integritas sesi aktif tidak berfungsi akurat. | Terkonfirmasi |
| **ASM-010** | Retensi Hubungan Alumni Berkelanjutan | Alumni memiliki minat untuk tetap terhubung dan mengambil proyek publik perusahaan. | Modul pasar proyek publik dan portofolio digital minim pemanfaatan. | Asumsi Terbuka |

---

## 16. Pertanyaan Terbuka (Open Questions)

Katalog pertanyaan terbuka berikut mencakup seluruh isu yang belum memiliki keputusan bisnis final dari para pemangku kepentingan:

#### OQ-001: Spesifikasi Desain & Template Sertifikat Digital
- **Question:** Bagaimana spesifikasi formal rancangan visual, skema penomoran unik, format tanda tangan digital, dan URL verifikasi QR untuk sertifikat kelulusan magang?
- **Context:** FR-041 (Certificate Engine).
- **Impact:** Mempengaruhi implementasi template generator berkas PDF dan infrastruktur validasi publik.
- **Status:** Open `[PERLU KONFIRMASI]`
- **Owner:** Reihan / Business Owner
- **Decision:** `TBD`

#### OQ-002: Otoritas Penandatangan Sah Sertifikat
- **Question:** Siapa saja pejabat berwenang yang berhak menandatangani sertifikat kelulusan secara digital (Direksi, HR Head, atau Program Lead)?
- **Context:** FR-041 (Certificate Engine) & Tata Kelola Dokumen.
- **Impact:** Menentukan relasi foreign key penandatangan sah dan hierarki persetujuan penerbitan sertifikat.
- **Status:** Open `[PERLU KONFIRMASI]`
- **Owner:** Reihan
- **Decision:** `TBD`

#### OQ-003: Nilai Ambang Batas XP untuk Progresi Rank
- **Question:** Berapa nilai ambang batas akumulasi XP numerik yang pasti untuk kenaikan tingkatan Rank (misal: Rank D $	o$ C $	o$ B $	o$ A $	o$ S)?
- **Context:** FR-026 (Rank Progression System).
- **Impact:** Mempengaruhi konfigurasi awal data master pada Unified Policy Engine.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** Business Configuration / Product Owner
- **Decision:** `TBD`

#### OQ-004: Bobot Matematis Formula Performance Score
- **Question:** Berapa pembobotan matematis eksak antara komponen Presensi, Kualitas Tugas, Evaluasi Supervisor, dan Kontribusi Proyek dalam menghasilkan Skor Kinerja Formal?
- **Context:** FR-027 (Performance Evaluation Engine).
- **Impact:** Menentukan algoritma kalkulasi komposit kinerja di luar sistem poin XP gamifikasi.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** HR / Program Lead
- **Decision:** `TBD`

#### OQ-005: Mitra Penyedia Gateway Pembayaran Payout
- **Question:** Layanan payment gateway / disbursement API apa yang akan diintegrasikan untuk memproses pencairan saldo dompet personal ke rekening bank/e-wallet pengguna?
- **Context:** FR-038 (Payout Engine).
- **Impact:** Menentukan spesifikasi kontrak integrasi teknis eksternal dan biaya transaksi penarikan.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** Finance Team
- **Decision:** `TBD`

#### OQ-006: Kuota & Ketentuan Kebijakan Cuti Magang
- **Question:** Berapa jatah kuota cuti resmi per peserta selama durasi program magang dan apa saja persyaratan verifikasi dokumen surat sakit?
- **Context:** FR-014 (Leave Management).
- **Impact:** Menentukan batasan validasi form pengajuan cuti dan dampaknya terhadap perhitungan kehadiran penuh.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** HR Team
- **Decision:** `TBD`

#### OQ-007: Mekanisme Otorisasi Pencairan Kas Angkatan (Batch Fund)
- **Question:** Bagaimana tata kelola dan alur persetujuan resmi untuk pencairan dan pemanfaatan akumulasi dana Batch Fund untuk kegiatan kohort angkatan?
- **Context:** FR-036 (Batch Fund).
- **Impact:** Menentukan perancangan alur kerja persetujuan finansial multi-pihak sebelum dana dapat ditarik.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** Program Management / Finance
- **Decision:** `TBD`

#### OQ-008: Kebijakan Moderasi & Unggah Galeri Kegiatan
- **Question:** Peran mana yang berhak mengunggah dokumentasi media ke Galeri dan apakah memerlukan mekanisme persetujuan moderasi sebelum tampil ke seluruh pengguna?
- **Context:** FR-044 (Media & Gallery).
- **Impact:** Menentukan permission model dan alur kerja publikasi media kegiatan.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** Marketing / Admin
- **Decision:** `TBD`

#### OQ-009: Penyedia Model & Batasan Cakupan Kecerdasan Buatan (AI)
- **Question:** Penyedia model AI apa (OpenAI, Claude, atau model lokal) yang akan digunakan dan bagaimana batasan parameter biaya serta privasi data operasional?
- **Context:** FR-050 (AI Insights & Activity Intelligence).
- **Impact:** Menentukan arsitektur perutean payload eksternal dan kepatuhan kerahasiaan data perusahaan.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** Technical Architecture Lead
- **Decision:** `TBD`

#### OQ-010: Skema Pencatatan Formal Daily Work
- **Question:** Apakah pelaporan aktivitas harian (*Daily Work*) memerlukan persetujuan pra-penugasan dari supervisor atau murni berupa pelaporan mandiri pasca-fakta (*self-reporting*)?
- **Context:** FR-022 (Work Report) & BR-005.
- **Impact:** Mempengaruhi alur validasi pada portal My Day dan beban kerja peninjauan supervisor.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** Program Lead
- **Decision:** `TBD`

#### OQ-011: Standar Kompensasi Moneter Jam Lembur
- **Question:** Berapa tarif nominal kompensasi finansial per jam lembur resmi yang disetujui (apakah bernilai tetap atau proporsional terhadap peran)?
- **Context:** FR-012 (Overtime Tracking) & BR-014.
- **Impact:** Menentukan formula perhitungan bonus moneter lembur pada Deduction & Tax Engine.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** Finance / HR
- **Decision:** `TBD`

#### OQ-012: Standardisasi Nomenklatur Visibilitas Proyek
- **Question:** Istilah baku mana yang akan disepakati untuk opsi visibilitas proyek internal: apakah "Internal" atau "Intern Only"?
- **Context:** FR-016 (Project Marketplace).
- **Impact:** Mempengaruhi konsistensi nilai enum basis data dan penamaan label antarmuka pengguna.
- **Status:** Open `[PERLU KONFIRMASI]`
- **Owner:** Product Owner
- **Decision:** `TBD`

#### OQ-013: Batasan Hak Akses Super Admin vs Admin
- **Question:** Apa batasan sistemik spesifik yang membedakan kewenangan Super Admin dengan Admin operasional biasa?
- **Context:** FR-001 (RBAC Engine) & Modul 3.
- **Impact:** Menentukan pemisahan permission master pada basis data otorisasi.
- **Status:** Open `[TIDAK DISEBUTKAN DALAM SUMBER]`
- **Owner:** System Architect
- **Decision:** `TBD`

#### OQ-014: Resolusi Event `WORK_ENDED` pada Enum Presensi
- **Question:** Apakah event `WORK_ENDED` (yang digunakan pada diagram alur WF-001 saat jam kerja selesai) akan ditambahkan secara resmi ke dalam daftar Attendance Event Types (FR-008 & DATA-001), atau diwakili oleh transisi status sesi kerja?
- **Context:** WF-001, FR-008, DATA-001.
- **Impact:** Menghilangkan diskrepansi arsitektur antara dokumentasi proses bisnis dan skema basis data event.
- **Status:** Open `[INCONSISTENCY — PERLU RESOLUSI]`
- **Owner:** System Architect / Product Owner
- **Decision:** `TBD`

---

## 17. Rencana Rilis (Release Plan)

### 17.1 Strategi Rilis (Release Strategy)
Pengembangan platform DCISP mengadopsi strategi rilis bertahap berbasis ketergantungan arsitektur (*phased dependency-driven evolutionary delivery*). Setiap fase rilis membangun fondasi kokoh bagi fase berikutnya, memastikan stabilitas integritas data sebelum fitur berbasis kecerdasan buatan dan integrasi eksternal diterapkan.

> **Catatan Tata Kelola:** Rencana fase rilis berikut berstatus **DRAFT** dan disusun sebagai rancangan acuan rekayasa perangkat lunak yang menunggu persetujuan komite pengarah produk.

### 17.2 Rencana Rilis Bertahap (Phased Release Plan)

| ID Fitur | Nama Fitur | Prioritas | Fase Rilis | Ketergantungan Utama | Status Kesiapan |
|---|---|---|---|---|---|
| **FR-001** | RBAC & Permission Engine | P0 | **Fase 1 — Foundation** | Basis Data & Auth Core | Ready for Dev |
| **FR-002** | Intern Lifecycle Management | P0 | **Fase 1 — Foundation** | FR-001 | Ready for Dev |
| **FR-004** | Batch Management | P0 | **Fase 1 — Foundation** | FR-001 | Ready for Dev |
| **FR-044** | Cloudflare R2 Object Storage | P0 | **Fase 1 — Foundation** | Cloudflare S3 API | Ready for Dev |
| **FR-045** | Unified Policy Engine | P0 | **Fase 1 — Foundation** | FR-001 | Ready for Dev |
| **FR-047** | Audit Logging & Security | P0 | **Fase 1 — Foundation** | DB Schema | Ready for Dev |
| **FR-048** | System Settings & Configurations | P1 | **Fase 1 — Foundation** | FR-001 | Ready for Dev |
| **FR-007** | Work Schedule Engine | P0 | **Fase 2 — Workforce** | FR-045 | Ready for Dev |
| **FR-008** | Attendance Event Engine | P0 | **Fase 2 — Workforce** | FR-007, FR-013 | Ready for Dev |
| **FR-009** | Work Session Tracking | P0 | **Fase 2 — Workforce** | FR-008 | Ready for Dev |
| **FR-010** | Break State Engine | P0 | **Fase 2 — Workforce** | FR-009 | Ready for Dev |
| **FR-011** | Session Integrity Tracking | P0 | **Fase 2 — Workforce** | FR-009 | Ready for Dev |
| **FR-012** | Overtime Request & Approval | P0 | **Fase 2 — Workforce** | FR-007 | Ready for Dev |
| **FR-013** | Device Registry | P1 | **Fase 2 — Workforce** | Hardware Readers | Ready for Dev |
| **FR-014** | Leave Management | P1 | **Fase 2 — Workforce** | FR-007 | Ready for Dev |
| **FR-015** | Attendance Correction Management | P0 | **Fase 2 — Workforce** | FR-008, FR-047 | Ready for Dev |
| **FR-016** | Project Marketplace | P0 | **Fase 3 — Projects** | FR-001, FR-002 | Ready for Dev |
| **FR-017** | Project Application & Quota | P0 | **Fase 3 — Projects** | FR-016 | Ready for Dev |
| **FR-019** | Project Team Management | P0 | **Fase 3 — Projects** | FR-017 | Ready for Dev |
| **FR-020** | Milestone Management | P1 | **Fase 3 — Projects** | FR-016 | Ready for Dev |
| **FR-021** | Task Management | P0 | **Fase 3 — Projects** | FR-020 | Ready for Dev |
| **FR-022** | Work Report / Submission | P0 | **Fase 3 — Projects** | FR-021, FR-044 | Ready for Dev |
| **FR-023** | Evidence Management | P0 | **Fase 3 — Projects** | FR-022, FR-044 | Ready for Dev |
| **FR-024** | Three-Layer Contribution Engine | P0 | **Fase 3 — Projects** | FR-021, FR-022 | Ready for Dev |
| **FR-025** | XP Rules Engine | P0 | **Fase 4 — Performance** | FR-008, FR-022, FR-045 | Ready for Dev |
| **FR-026** | Rank Progression System | P0 | **Fase 4 — Performance** | FR-025 | Ready for Dev |
| **FR-027** | Performance Evaluation Engine | P0 | **Fase 4 — Performance** | FR-008, FR-024 | Ready for Dev |
| **FR-028** | Top Performer per Batch | P1 | **Fase 4 — Performance** | FR-027 | Ready for Dev |
| **FR-029** | Achievement System | P1 | **Fase 4 — Performance** | FR-025 | Ready for Dev |
| **FR-030** | Skill Growth Matrix | P1 | **Fase 4 — Performance** | FR-027 | Ready for Dev |
| **FR-031** | Multi-Component Rank Rewards | P0 | **Fase 5 — Finance** | FR-026 | Ready for Dev |
| **FR-032** | Project Bounty Distribution | P0 | **Fase 5 — Finance** | FR-024 | Ready for Dev |
| **FR-033** | Reward Claim Management | P1 | **Fase 5 — Finance** | FR-031 | Ready for Dev |
| **FR-034** | Deduction & Tax Engine | P0 | **Fase 5 — Finance** | FR-045 | Ready for Dev |
| **FR-035** | Personal Wallet & Ledger | P0 | **Fase 5 — Finance** | FR-037 | Ready for Dev |
| **FR-036** | Batch Fund | P1 | **Fase 5 — Finance** | FR-004, FR-034 | Ready for Dev |
| **FR-037** | Financial Ledger (Double-Entry) | P0 | **Fase 5 — Finance** | DB ACID Schema | Ready for Dev |
| **FR-038** | Payout Engine | P1 | **Fase 5 — Finance** | FR-035, OQ-005 | Ready for Dev |
| **FR-039** | Financial Flow Orchestration | P0 | **Fase 5 — Finance** | FR-032, FR-034, FR-035 | Ready for Dev |
| **FR-051** | Command Center Dashboard | P1 | **Fase 6 — Portals** | Multi-Domain APIs | Ready for Dev |
| **FR-052** | "My Day" Intern Portal | P0 | **Fase 6 — Portals** | FR-009, FR-021, FR-025 | Ready for Dev |
| **FR-053** | "Team Today" Supervisor Portal | P1 | **Fase 6 — Portals** | FR-008, FR-009 | Ready for Dev |
| **FR-054** | Dedicated Supervisor Workspace | P1 | **Fase 6 — Portals** | FR-012, FR-015, FR-027 | Ready for Dev |
| **FR-055** | Global Sidebar Navigation | P0 | **Fase 6 — Portals** | FR-001 | Ready for Dev |
| **FR-046** | Notification Center | P1 | **Fase 6 — Portals** | Event Engine | Ready for Dev |
| **FR-003** | Alumni Lifecycle & Capabilities | P0 | **Fase 7 — Alumni & Docs** | FR-002 | Ready for Dev |
| **FR-005** | Institution Management | P1 | **Fase 7 — Alumni & Docs** | FR-002 | Ready for Dev |
| **FR-006** | Skill Matrix & Profiles | P1 | **Fase 7 — Alumni & Docs** | DB Master | Ready for Dev |
| **FR-040** | ID Card Generation | P1 | **Fase 7 — Alumni & Docs** | FR-002, FR-044 | Ready for Dev |
| **FR-041** | Certificate Engine | P1 | **Fase 7 — Alumni & Docs** | FR-002, OQ-001 | Needs Decision |
| **FR-042** | Digital Portfolio Builder | P1 | **Fase 7 — Alumni & Docs** | FR-003, FR-023 | Ready for Dev |
| **FR-043** | Document Reports Generator | P1 | **Fase 7 — Alumni & Docs** | FR-022, FR-044 | Ready for Dev |
| **FR-018** | Skill Matching Engine | P2 | **Fase 8 — Intelligence** | FR-006, FR-016 | Ready for Dev |
| **FR-049** | Analytics & Reporting | P1 | **Fase 8 — Intelligence** | Read-Model Store | Ready for Dev |
| **FR-050** | AI Insights & Intelligence | P2 | **Fase 8 — Intelligence** | OQ-009, Event Logs | Needs Decision |

### 17.3 Pemetaan Cakupan Scope ke Rilis (MVP vs Progressive Scope)
- **Minimum Viable Product (MVP Core):** Mencakup seluruh fitur Fase 1 (Foundation), Fase 2 (Workforce), Fase 3 (Projects), dan portal esensial My Day (FR-052). Memungkinkan peserta magang melakukan presensi, mencatat sesi kerja, mengerjakan tugas proyek, dan mengunggah bukti kerja.
- **Commercial & Governance Release (P0 Full):** Melengkapi seluruh mesin performa (Fase 4), orkestrasi finansial, pajak, dompet, dan buku besar (Fase 5), serta kedaulatan navigasi RBAC (FR-055).
- **Enrichment Release (P1 Scope):** Menghadirkan portal supervisor lengkap, pengakuan Top Performer, kartu ID digital, portofolio alumni, dan sistem klaim reward fisik.
- **Intelligent Operations (P2 Scope):** Integrasi AI Insights, smart skill matching algoritmik, dan analitik makro organisasi.

---

## 18. Kriteria Keberterimaan & Definition of Done (Acceptance & DoD)

### 18.1 Distingsi Kriteria Keberterimaan (AC) vs Definition of Done (DoD)
Untuk menjaga ketepatan tata kelola rekayasa perangkat lunak, sistem membedakan secara tegas:
- **Acceptance Criteria (Kriteria Keberterimaan Fitur):** Kriteria spesifik yang menguji apakah fungsionalitas dan aturan bisnis suatu fitur tertentu telah terpenuhi sesuai ekspektasi produk (diuji menggunakan skenario formal *Given / When / Then*).
- **Definition of Done (Definisi Selesai Komprehensif):** Standar mutu rekayasa perangkat lunak tingkat produk secara keseluruhan yang wajib dipenuhi oleh setiap fitur sebelum kode dapat dinyatakan selesai (*production-ready*) dan dirilis ke pengguna akhir.

### 18.2 Kerangka Keberterimaan Produk (Product Acceptance Framework)
Sistem DCISP dinyatakan dapat diterima secara produk apabila:
1. Seluruh 32 fitur prioritas P0 telah beroperasi stabil tanpa ada pelanggaran aturan bisnis (BR-001 s/d BR-025).
2. Tidak ada kebocoran batas privasi: pelacakan sesi kerja terbukti 100% mematuhi BR-010 (tanpa screenshot, keylogger, atau inspeksi chat).
3. Integritas buku kas ganda immutabel (BR-023) terbukti seimbang dengan selisih Rp 0 pada setiap siklus transaksi moneter.
4. Isolasi peran Scanner Operator (BR-002) dan partisi 3 skema XP (BR-004) terverifikasi secara ketat pada pengujian penetrasi keamanan.

### 18.3 Standar Pengujian Kualitas (QA Criteria)
1. **Unit Testing Coverage:** Logika bisnis krusial (formula XP Rules Engine, Three-Layer Contribution, Deduction & Tax Engine, dan State Machines) memiliki cakupan uji unit minimal 85%.
2. **Integration Testing:** Seluruh alur end-to-end (misal: Tap NFC $	o$ Attendance Record $	o$ Work Session $	o$ XP Mutation) lolos uji integrasi otomatis.
3. **Security Audit:** Tidak ada celah kerentanan kritis (OWASP Top 10) pada autentikasi JWT, evaluasi otorisasi RBAC, dan sanitasi berkas upload R2.

### 18.4 Definition of Done (DoD) Komprehensif — 15 Butir Standar Selesai
Setiap fitur dalam platform DCISP wajib memenuhi checklist 15 butir kriteria selesai berikut sebelum dinyatakan tuntas:

- [ ] **1. Requirements Implemented:** Seluruh kebutuhan fungsional telah diimplementasikan secara utuh sesuai klausul spesifikasi PRD.
- [ ] **2. Business Rules Implemented:** Seluruh aturan bisnis terkait (BR-001 s/d BR-025) terintegrasi dan terbukti mencegah pelanggaran aturan.
- [ ] **3. Authorization Verified:** Hak akses RBAC terverifikasi secara ketat pada tingkat API gateway dan database; peran terisolasi (misal Scanner Operator) terbukti tidak dapat mengakses resource lain.
- [ ] **4. Validation Implemented:** Seluruh validasi masukan (input sanitization, batas kuota, format payload, ukuran berkas) aktif pada front-end dan back-end.
- [ ] **5. Error States Handled:** Seluruh skenario kegagalan dan status HTTP error (400, 401, 403, 404, 409, 500) ditangani dengan pesan informatif yang ramah pengguna.
- [ ] **6. Empty States Handled:** Kondisi ketiadaan data menyajikan tampilan antarmuka terpandu (*zero state guidance*) dengan ilustrasi dan tombol aksi yang tepat.
- [ ] **7. Loading States Handled:** Masa tunggu pemrosesan data menampilkan skeleton loader atau spinner non-intrusif tanpa menyebabkan antarmuka freeze.
- [ ] **8. Edge Cases Tested:** Seluruh kasus batas (duplicate tap, race conditions, offline sync, koneksi terputus) teruji dan tertangani dengan aman.
- [ ] **9. Audit Events Recorded:** Seluruh peristiwa audit keamanan dan mutasi data penting tercatat secara permanen pada log audit append-only (FR-047).
- [ ] **10. Relevant Notifications Active:** Pemicu notifikasi transaksional terdistribusi dengan benar ke pihak yang berwenang (misal notifikasi izin lembur ke supervisor).
- [ ] **11. Data Persisted Correctly:** Integritas relasi entitas, tipe data fixed-point finansial, dan pemisahan metadata database vs berkas fisik Cloudflare R2 terverifikasi.
- [ ] **12. Integration Tested:** Titik integrasi internal dan eksternal (Device Gateway, R2 API) lolos pengujian integrasi dengan penanganan kegagalan (*retry & circuit breaker*).
- [ ] **13. Security & Privacy Reviewed:** Kepatuhan privasi kerja (BR-010) terverifikasi; tidak ada celah keamanan otorisasi atau kebocoran data pribadi.
- [ ] **14. Acceptance Criteria Passed:** Seluruh skenario pengujian berbasis *Given / When / Then* pada kriteria keberterimaan fitur dinyatakan lulus pengujian QA.
- [ ] **15. Regression Passed & Documentation Updated:** Pengujian regresi otomatis tidak mendeteksi kerusakan pada fitur yang telah ada sebelumnya, dan dokumentasi teknis API telah diperbarui.

---

## Appendix A — Requirement Traceability Matrix

### A.1 Matriks Ketertelusuran Fungsional Penuh (FR to BR, AC, Priority & Status)

Tabel berikut menyajikan pemetaan komprehensif seluruh 55 Kebutuhan Fungsional (FR-001 hingga FR-055) terhadap Aturan Bisnis terkait, Kriteria Keberterimaan, Prioritas Rilis, dan Status Validasi:

| ID Kebutuhan | Nama Fitur | Pemetaan Aturan Bisnis | Kriteria Keberterimaan (AC) | Prioritas | Status Validasi |
|---|---|---|---|---|---|
| **FR-001** | RBAC & Permission Engine | BR-001, BR-002 | AC-RBAC-001 (Scanner isolation 403, Dynamic policy reload) | P0 | Preserved & Lossless |
| **FR-002** | Manajemen Siklus Hidup Intern | BR-003, BR-004 | AC-PEO-001 (Transisi Applicant ke Intern aktif) | P0 | Preserved & Lossless |
| **FR-003** | Siklus Hidup Alumni & Retensi Akun | BR-003, BR-004 | AC-ALU-001 (Retensi akun permanen, isolasi XP alumni) | P0 | Preserved & Lossless |
| **FR-004** | Manajemen Batch | BR-018, BR-021 | AC-BAT-001 (Kohort batch aktif, pembentukan Batch Fund) | P0 | Preserved & Lossless |
| **FR-005** | Manajemen Institusi | `[TBD]` | AC-INS-001 (Master data kampus mitra tersimpan) | P1 | Preserved & Lossless |
| **FR-006** | Skill Matrix & Profil Skill | BR-015 | AC-SKL-001 (Pencatatan repositori keahlian pengguna) | P1 | Preserved & Lossless |
| **FR-007** | Work Schedule Engine | BR-007, BR-013 | AC-SCH-001 (Validasi batas jadwal, kalender libur & grace period) | P0 | Preserved & Lossless |
| **FR-008** | Attendance Event Engine | BR-006, BR-007 | AC-ATT-001 (Ingestion multi-metode, debouncing tap ganda) | P0 | Preserved & Lossless |
| **FR-009** | Work Session Tracking | BR-005, BR-007 | AC-SES-001 (Pencatatan sesi kerja, pemisahan 5 dimensi waktu) | P0 | Preserved & Lossless |
| **FR-010** | Break State Engine | BR-008, BR-009 | AC-BRK-001 (Transisi WORKING-BREAK, deteksi anomali istirahat) | P0 | Preserved & Lossless |
| **FR-011** | Session Integrity Tracking | BR-010, BR-011 | AC-INT-001 (Pelacakan visibilitas etis tanpa screenshot) | P0 | Preserved & Lossless |
| **FR-012** | Permintaan & Persetujuan Overtime | BR-013, BR-014 | AC-OVT-001 (Persetujuan supervisor, hitung jam aktual) | P0 | Preserved & Lossless |
| **FR-013** | Device Registry | BR-002 | AC-DEV-001 (Pendaftaran terminal fisik, otentikasi device) | P1 | Preserved & Lossless |
| **FR-014** | Leave Management | BR-012 | AC-LEV-001 (Pengajuan cuti, 0 penalti XP saat disetujui) | P1 | Preserved & Lossless |
| **FR-015** | Attendance Correction Management | BR-025 | AC-COR-001 (Koreksi absensi manual mempertahankan log asli) | P0 | Preserved & Lossless |
| **FR-016** | Project Marketplace | BR-003, BR-005 | AC-PRJ-001 (Publikasi bursa, pembatasan visibilitas proyek) | P0 | Preserved & Lossless |
| **FR-017** | Project Application & Quota | BR-015 | AC-APP-001 (Alur lamaran proyek, penutupan otomatis saat kuota penuh) | P0 | Preserved & Lossless |
| **FR-018** | Skill Matching Engine | BR-015 | AC-SKM-001 (Indikator kecocokan skill informatif, larangan auto-reject) | P2 | Preserved & Lossless |
| **FR-019** | Project Team Management | BR-016 | AC-TEM-001 (Penugasan peran tim, penetapan Planned Contribution %) | P0 | Preserved & Lossless |
| **FR-020** | Milestone Management | BR-005 | AC-MLS-001 (Pembagian deliverable milestone proyek terstruktur) | P1 | Preserved & Lossless |
| **FR-021** | Task Management | BR-005 | AC-TSK-001 (Manajemen tugas kanban, penugasan anggota tim) | P0 | Preserved & Lossless |
| **FR-022** | Work Report / Submission | BR-005 | AC-REP-001 (Pelaporan tugas, pencatatan kendala & progres kerja) | P0 | Preserved & Lossless |
| **FR-023** | Evidence Management | BR-024 | AC-EVD-001 (Penyimpanan tautan commit git & berkas fisik di R2) | P0 | Preserved & Lossless |
| **FR-024** | Three-Layer Contribution Engine | BR-016 | AC-CON-001 (Alur kontribusi Planned -> Actual -> Final % disahkan) | P0 | Preserved & Lossless |
| **FR-025** | XP Rules Engine | BR-004, BR-012 | AC-XPR-001 (Kalkulasi poin XP dinamis, isolasi 3 jalur XP) | P0 | Preserved & Lossless |
| **FR-026** | Rank Progression System | BR-017, BR-019 | AC-RNK-001 (Kenaikan level otomatis saat ambang XP tercapai; Rank != Score) | P0 | Preserved & Lossless |
| **FR-027** | Performance Evaluation Engine | BR-017, BR-018 | AC-EVL-001 (Evaluasi formal supervisor menghasilkan Performance Score) | P0 | Preserved & Lossless |
| **FR-028** | Top Performer per Batch | BR-018 | AC-TOP-001 (Pengakuan tepat SATU juara per batch per periode evaluasi) | P1 | Preserved & Lossless |
| **FR-029** | Achievement System | BR-004 | AC-ACH-001 (Pencapaian lencana kehormatan berdasarkan capaian kerja) | P1 | Preserved & Lossless |
| **FR-030** | Skill Growth Matrix | BR-015 | AC-SGM-001 (Pembaruan tingkat kemahiran skill berbasis deliverable) | P1 | Preserved & Lossless |
| **FR-031** | Multi-Component Rank Rewards | BR-019 | AC-RWD-001 (Penerbitan paket hadiah promosi rank: Cash, XP, Merchandise) | P0 | Preserved & Lossless |
| **FR-032** | Project Bounty Distribution | BR-016, BR-020 | AC-BNT-001 (Alokasi porsi bounty kotor berbasis Final Contribution %) | P0 | Preserved & Lossless |
| **FR-033** | Reward Claim Management | BR-019 | AC-CLM-001 (Siklus pengajuan klaim hadiah hingga status terpenuhi) | P1 | Preserved & Lossless |
| **FR-034** | Deduction & Tax Engine | BR-020, BR-021 | AC-TAX-001 (Perhitungan pemotongan pajak dinamis & kas angkatan) | P0 | Preserved & Lossless |
| **FR-035** | Personal Wallet & Source Tracking | BR-022, BR-023 | AC-WLT-001 (Buku mutasi dompet personal dengan jejak asal sumber dana) | P0 | Preserved & Lossless |
| **FR-036** | Batch Fund | BR-021 | AC-FND-001 (Akumulasi dana kas perpisahan angkatan terpisah dari pajak) | P1 | Preserved & Lossless |
| **FR-037** | Financial Ledger (Double-Entry) | BR-022, BR-023 | AC-LDG-001 (Pencatatan jurnal ganda immutabel dengan selisih Rp 0) | P0 | Preserved & Lossless |
| **FR-038** | Payout Engine | BR-022 | AC-PAY-001 (Alur penarikan saldo dompet menuju rekening bank/e-wallet) | P1 | Preserved & Lossless |
| **FR-039** | Financial Flow Orchestration | BR-021, BR-022 | AC-FLO-001 (Orkestrasi formula: Gross - Tax - Farewell = Net) | P0 | Preserved & Lossless |
| **FR-040** | ID Card Generation | `[TBD]` | AC-IDC-001 (Penerbitan kartu identitas digital ber-QR/NFC) | P1 | Preserved & Lossless |
| **FR-041** | Certificate Engine | OQ-001, OQ-002 | AC-CRT-001 (Penerbitan sertifikat kelulusan berverifikasi kode QR) | P1 | Marked `[Needs Confirmation]` |
| **FR-042** | Digital Portfolio Builder | BR-003 | AC-PTF-001 (Kompilasi otomatis portofolio karya alumni terverifikasi) | P1 | Preserved & Lossless |
| **FR-043** | Document Reports Generator | `[TBD]` | AC-DOC-001 (Ekspor laporan formal PDF dan logbook kegiatan) | P1 | Preserved & Lossless |
| **FR-044** | Cloudflare R2 Object Storage | BR-024 | AC-STR-001 (Penyimpanan berkas fisik di R2, metadata di database SQL) | P0 | Preserved & Lossless |
| **FR-045** | Unified Policy Engine | BR-001, BR-012, BR-020 | AC-POL-001 (Konfigurasi terpusat aturan runtime tanpa deploy ulang kode) | P0 | Preserved & Lossless |
| **FR-046** | Notification Center | `[TBD]` | AC-NOT-001 (Pengiriman peringatan sistem reaktif secara real-time) | P1 | Preserved & Lossless |
| **FR-047** | Audit Logging & Security | BR-023, BR-025 | AC-AUD-001 (Pencatatan jejak audit append-only yang tidak dapat dihapus) | P0 | Preserved & Lossless |
| **FR-048** | System Settings & Configurations | `[TBD]` | AC-SET-001 (Pengaturan parameter operasional global platform) | P1 | Preserved & Lossless |
| **FR-049** | Analytics & Reporting | `[TBD]` | AC-ANL-001 (Dasbor analitik kehadiran, proyek, dan utilisasi anggaran) | P1 | Preserved & Lossless |
| **FR-050** | AI Insights & Intelligence | BR-015, OQ-009 | AC-AIX-001 (Wawasan prediktif dan rekomendasi informatif non-punitive) | P2 | Preserved & Lossless |
| **FR-051** | Command Center Dashboard | BR-001 | AC-CMD-001 (Dasbor eksekutif: TODAY, ATTENTION, PERFORMANCE, FINANCE) | P1 | Preserved & Lossless |
| **FR-052** | Portal "My Day" | BR-005, BR-007, BR-008 | AC-MYD-001 (Layar tunggal kerja harian intern: timer, tugas, break) | P0 | Preserved & Lossless |
| **FR-053** | Portal "Team Today" | BR-007, BR-009 | AC-TTD-001 (Dasbor monitoring langsung kehadiran dan anomali tim supervisor) | P1 | Preserved & Lossless |
| **FR-054** | Dedicated Supervisor Workspace | BR-013, BR-016 | AC-SPV-001 (Ruang kerja persetujuan lembur, koreksi, dan evaluasi tim) | P1 | Preserved & Lossless |
| **FR-055** | Global Sidebar Navigation | BR-001, BR-002 | AC-NAV-001 (Penyajian menu hierarkis terisolasi berdasarkan peran dinamis) | P0 | Preserved & Lossless |

### A.2 Pemetaan Tujuan Bisnis ke Fitur (Business Goal to Feature Mapping)
- **Tujuan Efisiensi Operasional:** FR-007, FR-008, FR-009, FR-010, FR-012, FR-015, FR-051, FR-052, FR-053, FR-054.
- **Tujuan Integritas & Akuntabilitas Waktu Kerja:** FR-007, FR-008, FR-009, FR-010, FR-011, FR-015, FR-047.
- **Tujuan Peningkatan Kualitas Deliverable Proyek:** FR-016, FR-017, FR-019, FR-020, FR-021, FR-022, FR-023, FR-024.
- **Tujuan Kepatuhan Finansial & Perpajakan:** FR-031, FR-032, FR-034, FR-035, FR-036, FR-037, FR-038, FR-039.
- **Tujuan Tata Kelola Talenta & Retensi Alumni:** FR-002, FR-003, FR-004, FR-005, FR-006, FR-040, FR-041, FR-042.
- **Tujuan Motivasi & Gamifikasi Terstandar:** FR-025, FR-026, FR-027, FR-028, FR-029, FR-030, FR-031.

### A.3 Pemetaan Fitur ke Entitas Data (Feature to Data Mapping)
- FR-001 $\longrightarrow$ `User`, `Role`, `Permission`, `Scope`
- FR-007, FR-008, FR-013 $\longrightarrow$ `Work Schedule`, `Holiday`, `Attendance Event`, `Device`
- FR-009, FR-010, FR-011 $\longrightarrow$ `Work Session`, `Break`, `Attendance Event`
- FR-012, FR-014, FR-015 $\longrightarrow$ `Overtime Request`, `Leave Request`, `Attendance Correction`
- FR-016, FR-017, FR-019 $\longrightarrow$ `Project`, `Project Application`, `Project Team`
- FR-020, FR-021, FR-022, FR-023 $\longrightarrow$ `Milestone`, `Task`, `Work Report`, `Evidence`, `File Metadata`
- FR-024, FR-032 $\longrightarrow$ `Project Team`, `Project`, `Wallet Transaction`
- FR-025, FR-026, FR-027, FR-028 $\longrightarrow$ `XP Rule`, `XP Transaction`, `Rank`, `Achievement`, `Performance Evaluation`
- FR-031, FR-033 $\longrightarrow$ `Reward`, `Reward Claim`, `Rank`
- FR-034, FR-035, FR-036, FR-037, FR-038, FR-039 $\longrightarrow$ `Deduction / Tax Rule`, `Personal Wallet`, `Wallet Transaction`, `Batch Fund`, `Financial Ledger`, `Payout`
- FR-040, FR-041, FR-042, FR-044 $\longrightarrow$ `ID Card`, `Certificate`, `Portfolio`, `File Metadata (R2)`
- FR-045, FR-046, FR-047 $\longrightarrow$ `Policy`, `Notification`, `Audit Log`

### A.4 Pemetaan Fitur ke Antarmuka Pengguna (Feature to UI Mapping)
- **Command Center (FR-051):** Widget TODAY (FR-008, FR-009), ATTENTION (FR-012, FR-014, FR-015, FR-017, FR-022), PERFORMANCE (FR-027, FR-028), FINANCE (FR-032, FR-034, FR-035, FR-036).
- **Portal "My Day" (FR-052):** Kontrol Sesi Kerja (FR-009, FR-010), Daftar Tugas Hari Ini (FR-021), Pelaporan Harian (FR-022), XP Counter (FR-025).
- **Portal "Team Today" (FR-053):** Headcount Live Tim (FR-008, FR-009), Peringatan Anomali Istirahat & Keterlambatan (FR-010, FR-011).
- **Supervisor Workspace (FR-054):** Persetujuan Lembur (FR-012), Koreksi Absensi (FR-015), Verifikasi Laporan Tugas (FR-022, FR-023), Finalisasi Kontribusi (FR-024), Evaluasi Kinerja (FR-027).
- **Global Sidebar Navigation (FR-055):** Penegakan visibilitas seluruh 11 modul utama berbasis peran dinamis (FR-001, BR-001, BR-002).

---

## Appendix B — Glosarium Istilah (Glossary)

Daftar istilah standar yang digunakan secara konsisten dalam seluruh dokumen spesifikasi DCISP:

- **Active Session Time:** Durasi waktu riil di mana pengguna berinteraksi aktif dengan aplikasi/tugas di peramban, tidak termasuk waktu jeda ketiadaan aktivitas (*idle* > 15 menit).
- **Actual Contribution %:** Persentase kontribusi anggota tim proyek yang dihitung secara otomatis oleh algoritma sistem berbasis tugas yang diselesaikan, bobot tugas, dan hasil verifikasi bukti deliverable.
- **Alumni:** Status permanen bagi peserta magang yang telah menyelesaikan program secara resmi, mempertahankan hak akses ke bursa proyek publik dan portofolio digital tanpa memengaruhi rank magang aktif.
- **Alumni Contribution:** Skema partisi poin pengalaman (XP) khusus bagi alumni yang diperoleh dari keterlibatan pada proyek publik pasca-magang; terisolasi penuh dari *Internship XP*.
- **Attendance Time:** Total durasi keberadaan fisik pengguna di kantor, diukur dari timestamp check-in fisik pertama hingga timestamp check-out fisik akhir hari.
- **Batch:** Kohort pengelompokan peserta magang yang menjalani masa program bersamaan dalam periode kalender tertentu.
- **Batch Fund:** Kas kolektif angkatan yang terakumulasi dari kontribusi perpisahan peserta magang (*Farewell Contributions*), didedikasikan untuk mendanai kegiatan kebersamaan kohort; terpisah secara hukum dari pajak pemerintah.
- **Bounty Pool:** Total alokasi dana imbalan moneter yang disediakan untuk suatu inisiatif proyek, yang nantinya dibagikan kepada anggota tim berbasis persentase kontribusi akhir (*Final Contribution %*).
- **Break Time:** Total durasi waktu saat pengguna berada dalam status istirahat resmi (`BREAK`), baik yang terjadwal maupun permohonan jeda.
- **Cloudflare R2:** Layanan penyimpanan objek (*object storage*) nir-egress yang kompatibel dengan protokol S3, digunakan sebagai repositori fisik berkas bukti kerja, foto profil, dan dokumen sertifikat.
- **Daily Work:** Aktivitas kerja operasional rutin harian selama jam kantor yang tidak selalu merupakan bagian dari tugas proyek marketplace terstruktur.
- **Device Gateway:** Layanan perantara yang menjembatani komunikasi data perangkat keras terminal scanner (NFC reader / barcode scanner) dengan Attendance Event Engine melalui HTTPS terenkripsi.
- **Double-Entry Ledger:** Sistem pencatatan akuntansi keuangan berpasangan di mana setiap transaksi dicatat pada sisi debit dan kredit dengan total nilai yang sama persis (selisih Rp 0), bersifat immutabel (*append-only*).
- **Evidence:** Berkas bukti hasil kerja terverifikasi (berupa tautan commit/PR repositori Git, berkas dokumen, atau tangkapan layar antarmuka) yang wajib dilampirkan pada penyerahan tugas proyek.
- **Farewell Contribution:** Potongan dana kontribusi sukarela/kesepakatan batch yang disisihkan dari hasil bounty proyek untuk dialirkan ke dalam Batch Fund.
- **Final Contribution %:** Persentase kontribusi akhir anggota tim proyek yang disahkan secara manual oleh Supervisor setelah meninjau rekomendasi Actual Contribution sistem; menjadi dasar mutlak perhitungan pembagian porsi bounty.
- **Grace Period:** Toleransi waktu keterlambatan kedatangan (misal: 10 menit setelah jam masuk jadwal resmi) yang tidak dikenakan penalti poin XP.
- **Idle Duration:** Akumulasi durasi waktu saat sesi kerja berjalan namun peramban pengguna tidak mendeteksi adanya aktivitas interaksi selama lebih dari 15 menit berturut-turut.
- **Intern:** Peserta aktif yang sedang menjalani program magang resmi di PT. Aplikasi Dagang Teknologi.
- **Internship XP:** Skema partisi poin pengalaman yang mencerminkan kedisiplinan kehadiran, keaktifan kerja, dan dedikasi magang peserta aktif; digunakan sebagai dasar progresi kenaikan Rank.
- **Milestone:** Titik capaian fase pekerjaan perantara dalam suatu proyek yang memayungi sejumlah unit tugas spesifik.
- **Overtime Time:** Jam kerja lembur aktual yang dijalankan pengguna di luar jam kerja reguler dan berada di dalam jendela waktu lembur yang telah disetujui supervisor di muka.
- **Performance Score:** Skor penilaian kualitas formal profesional (skala 0 – 100) yang diberikan oleh atasan/supervisor melalui instrumen evaluasi berkala; terpisah secara mutlak dari akumulasi Rank gamifikasi ($	ext{Rank} 
eq 	ext{Performance Score}$).
- **Personal Wallet:** Buku kas dompet digital pribadi setiap pengguna yang mencatat saldo bersih (*Net Distributable*), mutasi penambahan insentif, dan riwayat pencairan dana (*payout*).
- **Planned Contribution %:** Kesepakatan awal pembagian porsi kontribusi anggota tim proyek saat tim pertama kali dibentuk oleh Project Manager.
- **Policy Engine:** Mesin terpusat yang mengatur evaluasi aturan bisnis runtime (kebijakan jadwal, batas penalti XP, rumus deduksi pajak, hak akses) secara dinamis tanpa mengubah kode program.
- **Project Task:** Unit pekerjaan terstruktur yang terikat pada hierarki proyek: $	ext{Project} 	o 	ext{Milestone} 	o 	ext{Task} 	o 	ext{Assignment} 	o 	ext{Submission}$.
- **Project XP:** Skema partisi poin pengalaman yang diperoleh murni dari penyelesaian deliverable proyek teknis di Marketplace.
- **Rank:** Tingkatan status kehormatan gamifikasi pengguna (Rank D, C, B, A, S) yang naik secara bertahap saat akumulasi poin pengalaman mencapai ambang batas tertentu.
- **Scanner Operator:** Peran pengguna khusus dengan hak akses terbatas murni (`scope = scanner_only`) untuk mengoperasikan terminal pemindai presensi fisik di lobi kantor tanpa akses ke modul data lain.
- **Session Integrity:** Pemantauan keterlibatan kerja aktif berbasis visibilitas tab peramban secara etis dan menjaga privasi tanpa menggunakan perangkat mata-mata invasif (*non-invasive tracking*).
- **Skill Matrix:** Repositori pemetaan profil keahlian teknis dan kemahiran kerja pengguna yang digunakan untuk rekomendasi pencocokan proyek di marketplace.
- **Three-Layer Contribution:** Metodologi penentuan alokasi insentif proyek 3 lapis: *Planned Contribution* (kesepakatan awal) $	o$ *Actual Contribution* (rekomendasi algoritmik sistem) $	o$ *Final Contribution* (pengesahan manual supervisor).
- **Top Performer:** Gelar pengakuan prestasi tertinggi yang dianugerahkan kepada tepat SATU orang peserta magang terbaik per kohort batch per periode evaluasi berbasis formula komposit terstandar.
- **Work Report:** Laporan kemajuan kerja berkala yang diserahkan pengguna untuk mendokumentasikan progres, kendala, dan bukti deliverable hasil kerja harian atau penugasan proyek.
- **Work Session Time:** Total durasi sesi kerja resmi pengguna, dihitung dari penekanan tombol "Start Work" hingga sesi diakhiri ("End Work") pada portal My Day.

---

## Appendix C — Log Keputusan Terbuka (Open Decisions Log)

Tabel ini merekam riwayat keputusan arsitektur dan bisnis formal yang telah disahkan, serta menyediakan kerangka kerja untuk pencatatan resolusi pertanyaan terbuka (*Open Questions*) di masa mendatang:

| ID Keputusan | Tanggal | Topik / Ref OQ | Status | Ringkasan Keputusan yang Diambil | Dasar Pemikiran & Konsekuensi Arsitektural | Pengambil Keputusan |
|---|---|---|---|---|---|---|
| **DEC-001** | 2026-09-19 | Paradigma Produk | **DISETUJUI** | DCISP ditetapkan sebagai platform pengelolaan individu dan project-based work dengan internship lifecycle sebagai core business dan gamification sebagai performance layer. | Menghindari reduksi sistem menjadi sekadar aplikasi absensi atau game. Arsitektur memisahkan modul inti dengan lapisan gamifikasi. | Management & Core Team |
| **DEC-002** | 2026-09-19 | Kebijakan Arsitektur (BR-001) | **DISETUJUI** | Seluruh aturan bisnis, formula penalti, ambang XP, dan jadwal kerja wajib digerakkan oleh Unified Policy Engine (FR-045), tidak boleh di-hardcode. | Mencegah kerapuhan kode sumber dan memungkinkan penyesuaian aturan tanpa perlu melakukan build dan deployment ulang. | Lead Systems Architect |
| **DEC-003** | 2026-09-19 | Batasan Privasi (BR-010) | **DISETUJUI** | Pelarangan mutlak terhadap alat surveilans invasif (screenshot layar, keylogger, pemindaian proses, intip clipboard, baca chat). | Menjaga martabat privasi pekerja, kepatuhan etika, dan kepercayaan pengguna terhadap platform. | Product Owner & Legal |
| **DEC-004** | 2026-09-19 | Penyimpanan Berkas (BR-024) | **DISETUJUI** | Seluruh objek berkas fisik disimpan pada Cloudflare R2; database relasional hanya menyimpan metadata berkas (DATA-007). | Mengoptimalkan performa basis data, menekan biaya penyimpanan, dan mencegah pembengkakan ukuran tabel relasional. | Infrastructure Lead |
| **DEC-005** | 2026-09-19 | Pembukuan Finansial (BR-023) | **DISETUJUI** | Mengadopsi prinsip pembukuan ganda (*double-entry ledger*) yang tidak dapat diubah (*immutable*) untuk seluruh transaksi dompet personal dan bounty. | Menjamin auditabilitas moneter 100%, mencegah saldo gantung, dan memenuhi standar kepatuhan akuntansi. | Finance & Systems Architect |
| **DEC-006** | *Pending* | OQ-001 & OQ-002 | *Menunggu Keputusan* | Penetapan template resmi sertifikat digital dan tanda tangan berwenang. | Menentukan implementasi rendering SVG/PDF pada FR-041. | Reihan / Business Owner |
| **DEC-007** | *Pending* | OQ-005 | *Menunggu Keputusan* | Pemilihan mitra gateway pembayaran untuk otomasi penarikan dana personal wallet. | Menentukan endpoint integrasi API pada FR-038 (Payout Engine). | Finance Team |
| **DEC-008** | *Pending* | OQ-014 | *Menunggu Keputusan* | Resolusi penambahan `WORK_ENDED` ke dalam enum Attendance Event Type (FR-008 & DATA-001). | Menyelaraskan alur presensi WF-001 dengan skema database backend. | System Architect |

---

> **Pernyataan Penutup Single Source of Truth:**  
> Dokumen Product Requirements Document (PRD) Dagang Creative Intern Solutions Program (DCISP) Versi 1.0 ini merupakan **satu-satunya sumber kebenaran resmi (*Single Source of Truth*)** bagi seluruh pemangku kepentingan, Product Owner, Business Analyst, Software Architect, UI/UX Designer, QA Engineer, dan AI Coding Agent. Seluruh implementasi rekayasa perangkat lunak wajib mematuhi seluruh batasan, aturan bisnis, dan spesifikasi yang tercantum di dalam dokumen ini tanpa pengecualian.
