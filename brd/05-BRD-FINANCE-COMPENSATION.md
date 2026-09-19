# BUSINESS REQUIREMENTS DOCUMENT (BRD)
## Domain: Finance, Taxation, Compensation & Double-Entry Ledger
### Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0

---

# PART A — BUSINESS DOMAIN ANALYSIS

### 1. Business Domain yang Dipilih
**Finance, Taxation, Compensation & Double-Entry Ledger (Tata Kelola Keuangan, Perpajakan Dinamis, Dompet Personal, Kas Perpisahan Angkatan, Pembukuan Ganda Immutabel, dan Pencairan Dana)**.

### 2. Alasan Domain Ini Dianggap Satu Kesatuan Bisnis
Domain ini mengatur seluruh pergerakan moneter, kepatuhan audit finansial, dan distribusi kompensasi hak peserta di dalam platform DCISP:
- Setiap peristiwa moneter (distribusi bounty proyek, kompensasi lembur, klaim hadiah uang tunai kenaikan rank) wajib dievaluasi oleh mesin pemotongan pajak dan iuran bersama (*Deduction & Tax Engine*).
- Mengorkestrasi formula pembagian keuangan resmi (*Financial Flow Orchestration*):
  $$\text{Gross Bounty} - \text{Tax/Deductions} - \text{Farewell Contribution} = \text{Net Distributable} \rightarrow \text{Personal Wallet}$$
- Memisahkan secara tegas dan berlandaskan hukum antara pemotongan pajak resmi negara (*Income Tax / PPh*) dengan iuran sukarela perpisahan angkatan (*Batch Fund / Farewell Contribution* - *BR-021*).
- Menjamin integritas audit finansial 100% melalui pencatatan buku besar akuntansi ganda berpasangan (*Immutable Double-Entry Ledger* - *BR-023*) yang tidak dapat dimodifikasi atau dihapus (*append-only*).
- Mengatur tata kelola penarikan dana dompet (*Payout Engine*) ke rekening bank/e-wallet eksternal pengguna.

### 3. Batasan Domain Bisnis (Boundary)
* **Business Trigger:** Pengesahan deliverable proyek selesai (bounty pool siap bagi hasil), kenaikan rank yang membuka reward uang tunai, eksekusi lembur yang disetujui, atau pengajuan permohonan pencairan saldo dompet (*payout request*).
* **Input Bisnis:** Alokasi total bounty pool proyek, persentase *Final Contribution %* anggota tim, aturan tarif pajak aktif (*tax rules*), aturan potongan iuran kas batch (*batch contribution rate*), saldo dompet aktif, dan rekening bank tujuan pencairan.
* **Proses Utama Bisnis:**
  1. Perhitungan bagi hasil kotor (*Gross Bounty*) per anggota tim berbasis persentase kontribusi akhir yang disahkan (*BR-016*).
  2. Evaluasi aturan pemotongan pajak dan biaya administrasi secara dinamis tanpa hardcode (*BR-020*).
  3. Pemisahan dana iuran kas angkatan (*Batch Fund*) dari pos kewajiban pajak perusahaan/negara (*BR-021*).
  4. Pengkreditan saldo bersih (*Net Distributable*) ke dompet digital personal (*Personal Wallet*) dengan pencatatan mutasi transparan (*BR-022*).
  5. Pembukuan ganda berpasangan immutabel (*Double-Entry Ledger*) dengan keseimbangan mutlak debit dan kredit ($\sum \text{Debit} = \sum \text{Kredit}$, selisih Rp 0 - *BR-023*).
  6. Alur verifikasi dan pemrosesan permohonan pencairan saldo (*Payout Engine*) ke rekening bank pengguna.
* **Keputusan Bisnis yang Dibuat:**
  - Penetapan keabsahan formula dan tarif pemotongan pajak penghasilan aktif.
  - Persetujuan/penolakan permohonan pencairan dana dompet oleh Finance Admin.
  - Otorisasi pencairan dana bersama untuk kegiatan perpisahan kohort (*Batch Fund*).
* **Output Bisnis:** Saldo bersih di dompet pengguna, saldo terakumulasi pada rekening bersama Batch Fund, catatan entri jurnal ganda abadi (*Ledger Entries*), dan status transfer pencairan dana (*Settled Payouts*).
* **Kapan Selesai:** Dana bersih masuk ke rekening bank pengguna dan transaksi kas keluar dibukukan tuntas pada buku besar immutabel.
* **Domain Konsumen Output:**
  - *Domain Executive & Command Center* (mengonsumsi agregasi total kas keluar, pajak terkumpul, saldo dompet, dan saldo batch fund).
  - *Domain People* (mengonsumsi saldo Batch Fund untuk penyelenggaraan upacara kelulusan batch).

### 4. Hubungan dengan Domain Lain (Related Domains)
* **Domain Projects & Tasks (Upstream Input):** Menyediakan alokasi *Bounty Pool* dan *Final Contribution %*.
* **Domain Workforce & Attendance (Upstream Input):** Menyediakan data jam lembur aktual yang disahkan.
* **Domain Performance & Gamification (Upstream Input):** Menyediakan event kenaikan rank yang memicu reward uang tunai.
* **Domain Identity & RBAC (Security Boundary):** Menyediakan identitas pengguna terotentikasi dan isolasi akses perbendaharaan khusus role *Finance*.

---

# PART B — BUSINESS REQUIREMENTS DOCUMENT (BRD)

## 1. Document Control

| Atribut | Detail |
|---|---|
| **Document Name** | Business Requirements Document (BRD) — Finance, Taxation & Double-Entry Ledger |
| **Business Domain** | Finance & Taxation Management (Domain 7 PRD) |
| **Document Version** | 1.0 |
| **Document Status** | Final Draft / Ready for Business Sign-Off |
| **Business Owner** | Faisal (Finance Administrator Lead, PT. Aplikasi Dagang Teknologi) |
| **Prepared By** | Lead Requirements Engineer & Business Analyst |
| **Date** | 2026-09-19 |
| **Related PRD** | Product Requirements Document (PRD) Platform DCISP v1.0 (Section 2, 3, 5, 6.2.6, 8.7, 8.8, 10) |

---

## 2. Executive Summary

Dokumen ini mendefinisikan kebutuhan bisnis untuk domain **Finance, Taxation, Compensation & Double-Entry Ledger**. Tujuannya adalah menjamin integritas finansial, keadilan kompensasi insentif magang, kepatuhan regulasi perpajakan di Indonesia, transparansi pembukuan dompet personal, serta pemisahan tegas antara kas perpisahan angkatan (*Batch Fund*) dengan pos pajak resmi melalui buku besar akuntansi ganda immutabel (*BR-020 s/d BR-023*).

---

## 3. Business Context

Sebelum platform DCISP dirancang:
1. Perhitungan bagi hasil proyek (*bounty*) tidak memiliki tahapan audit yang jelas, memicu sengketa antar-anggota tim.
2. Pemotongan pajak penghasilan dan iuran kas angkatan bercampur aduk dalam satu pos tidak resmi, menyulitkan pelaporan pajak korporat.
3. Pencatatan keuangan bersifat single-entry pada spreadsheet biasa yang rentan manipulasi data dan penghapusan transaksi tanpa jejak audit.

---

## 4. Business Problem

1. **Kerentanan Inkonsistensi Saldo:** Risiko ketidakseimbangan neraca keuangan akibat transaksi sepihak tanpa pembukuan berpasangan.
2. **Kerapuhan Aturan Pajak Statis:** Kebijakan tarif pajak yang di-hardcode menyulitkan adaptasi regulasi perpajakan yang berubah.
3. **Pencampuran Dana Iuran vs Pajak:** Risiko hukum akibat pencampuran kas bersama angkatan ke dalam pos pajak resmi perusahaan.

---

## 5. Business Objectives

| ID Objective | Business Problem yang Diselesaikan | Desired Business Outcome | Business Value | Success Indicator |
|---|---|---|---|---|
| **OBJ-FIN-01** | Inkonsistensi & manipulasi pencatatan kas | 100% mutasi moneter dicatat dalam buku besar ganda immutabel ($\sum \text{Debit} = \sum \text{Kredit}$) | Integritas keuangan mutlak & kepatuhan audit akuntansi | Selisih neraca buku besar tepat Rp 0 pada setiap transaksi |
| **OBJ-FIN-02** | Ketidaktransparanan potongan insentif | Rincian potongan (*Gross, Tax, Batch Fund, Net*) transparan di dompet personal | Kepercayaan peserta magang & kejelasan hak finansial | 100% mutasi dompet memiliki jejak sumber dana terperinci |
| **OBJ-FIN-03** | Risiko hukum pencampuran dana | Pemisahan konseptual dan rekening mutlak antara pajak negara dengan kas angkatan | Kepatuhan regulasi hukum & perlindungan dana bersama | Saldo kas Batch Fund terisolasi 100% per kohort batch |

---

## 6. Business Scope

### 6.1 In Scope
1. **Multi-Component Rank Rewards (FR-031, BR-019):** Paket hadiah promosi rank mencakup uang tunai (*Cash*), bonus XP, merchandise, badge, voucher, dan hak istimewa.
2. **Project Bounty Distribution (FR-032, BR-016):** Alokasi porsi bounty kotor berbasis *Final Contribution %* yang disahkan.
3. **Reward Claim Management (FR-033):** Siklus operasional klaim barang merchandise fisik dan voucher.
4. **Deduction & Tax Engine Dinamis (FR-034, BR-020):** Perhitungan pemotongan pajak penghasilan, biaya penarikan, dan iuran batch berbasis kebijakan runtime.
5. **Personal Wallet & Source Tracking (FR-035, BR-022):** Dompet digital pribadi dengan pencatatan mutasi itemized multi-sumber (DATA-006).
6. **Manajemen Kas Bersama Angkatan / Batch Fund (FR-036, BR-021):** Akumulasi dan pengelolaan dana perpisahan kohort magang.
7. **Buku Kas Ganda Immutabel / Financial Ledger (FR-037, BR-023):** Pencatatan akuntansi ganda berpasangan abadi (*append-only, no update/delete*).
8. **Payout Engine (FR-038):** Permohonan, validasi, dan penyelesaian penarikan saldo ke rekening bank eksternal.
9. **Financial Flow Orchestration (FR-039):** Orkestrasi transaksi atomik: $\text{Gross} - \text{Tax} - \text{Farewell} = \text{Net Distributable}$.

### 6.2 Out of Scope
1. **Penggajian Gaji Pokok Karyawan Tetap (Corporate Payroll):** Penggajian bulanan dan tunjangan hari raya karyawan tetap di luar cakupan sistem.
2. **Penerbitan Faktur Pajak Pajak Pertambahan Nilai (PPN) Korporat:** Platform mengelola pemotongan insentif individu, bukan faktur PPN B2B.

---

## 7. Stakeholders

| Stakeholder | Responsibility | Business Interest | Decision Authority |
|---|---|---|---|
| **Finance Administrator (The Merchant)** | Mengonfigurasi aturan pajak, memantau buku besar immutabel, menyetujui payout | Kepatuhan perpajakan, keakuratan buku kas, dan keamanan distribusi dana | Otorisasi pencairan dana (*Payout Approval*) & konfigurasi tarif deduksi |
| **Program Lead / Operations Owner** | Meninjau akumulasi Batch Fund dan pengesahan pencairan dana kegiatan kohort | Transparansi kas perpisahan angkatan dan efisiensi insentif | Otorisasi penggunaan kas bersama angkatan (*Batch Fund*) |
| **Intern / Alumni (Adventurer)** | Menerima kompensasi bounty, reward kenaikan rank, mengajukan penarikan dana | Kecepatan penerimaan insentif dan kejelasan rincian potongan | Mengajukan permohonan pencairan saldo dompet pribadi |
| **Auditor Keuangan Eksternal** | Meninjau kepatuhan rekonsiliasi buku besar akuntansi ganda | Kepatuhan standar akuntansi dan auditabilitas jejak transaksi | Menilai kewajaran laporan keuangan platform |

---

## 8. Business Actors

| Business Role | System Role Terkait | Tindakan Utama | Batasan Akses Data |
|---|---|---|---|
| **Treasury Officer** | Finance / Super Admin | Kelola tarif deduksi, eksekusi payout, audit ledger | `scope = financial_data` |
| **Reward Beneficiary** | Intern / Alumni | Memantau saldo dompet, mengajukan payout, klaim hadiah | `scope = own_data` |

---

## 9. Current Business Process (AS-IS)

* **AS-IS:** Bagi hasil insentif proyek dihitung manual di Excel. Pemotongan pajak dilakukan tanpa aturan tertulis dan dana kas perpisahan angkatan ditampung di rekening pribadi salah satu staf HR/Finance tanpa pembukuan ganda resmi.

---

## 10. Target Business Process (TO-BE)

```text
[ Proyek Disahkan Selesai (Bounty Pool Siap Bagi Hasil) ]
                           │
                           ▼
[ Hitung Gross Bounty Tiap Anggota (Pool x Final Contribution %) ]
                           │
                           ▼
[ Deduction & Tax Engine Memproses Pemotongan Dinamis ]
                           │
                           ├──► Hitung Pajak Penghasilan (PPh) ──► Kredit Akun Kewajiban Pajak
                           └──► Hitung Iuran Kas Angkatan    ──► Kredit Akun Batch Fund
                           │
                           ▼
[ Hitung Net Distributable = Gross - Tax - Farewell ]
                           │
                           ▼
[ Kreditkan Net Distributable ke Personal Wallet Pengguna ]
                           │
                           ▼
[ Catat Baris Debit & Kredit Berpasangan ke Financial Ledger (Immutable) ]
                           │
                           ▼
[ Pengguna Mengajukan Payout ──► Verifikasi Finance ──► Transfer Bank Selesai ]
```

---

## 11. Business Process Scenarios

### Scenario 1: Distribusi Bagi Hasil Bounty Proyek
* **Trigger:** Proyek selesai dengan Bounty Pool Rp10.000.000 dan Final Contribution Anggota A = 40%.
* **Actor:** Sistem Finansial, Finance Admin.
* **Main Flow:**
  1. Sistem menghitung Gross Bounty Anggota A = Rp4.000.000.
  2. Deduction Engine mengevaluasi aturan aktif: Pajak Penghasilan 5% (Rp200.000) dan Iuran Batch Fund 2% (Rp80.000).
  3. Total deduksi = Rp280.000; Nilai bersih (Net Distributable) = Rp3.720.000.
  4. Sistem mengkreditkan Rp3.720.000 ke Personal Wallet Anggota A.
  5. Sistem mengkreditkan Rp80.000 ke rekening bersama Batch Fund.
  6. Financial Ledger mencatat jurnal berpasangan seimbang: Debit Pool Proyek Rp4.000.000; Kredit Utang Anggota A Rp3.720.000; Kredit Utang Pajak Rp200.000; Kredit Batch Fund Rp80.000 ($\sum \text{Debit} - \sum \text{Kredit} = 0$).
* **Output:** Saldo dompet personal bertambah Rp3.720.000, mutasi itemized tercatat, buku besar terkunci permanen.

### Scenario 2: Pencairan Saldo Dompet (*Payout Processing*)
* **Trigger:** Intern mengajukan penarikan saldo dompet sebesar Rp1.500.000 ke rekening BCA miliknya.
* **Actor:** Intern (Pemohon), Finance Admin (Penyetujui).
* **Main Flow:**
  1. Intern mengisi form payout (nominal Rp1.500.000, bank BCA, no rekening terverifikasi).
  2. Sistem memvalidasi saldo aktif mencukupi dan menahan saldo sebesar Rp1.500.000 (status: `REQUESTED / HOLD`).
  3. Finance Admin meninjau antrean payout di modul Finance dan menyetujui transaksi.
  4. Finance Admin melakukan transfer perbankan dan memasukkan nomor referensi transfer.
  5. Status payout berubah menjadi `SETTLED`; Financial Ledger mencatat pengeluaran kas bank.
* **Output:** Saldo bersih dompet terpotong Rp1.500.000; dana masuk ke rekening bank pengguna.

---

## 12. Business Requirements

| ID Kebutuhan | Deskripsi Kebutuhan Bisnis | Rasional Bisnis | Prioritas | Sumber PRD |
|---|---|---|---|---|
| **BRQ-FIN-001** | Bisnis mewajibkan promosi kenaikan rank membuka paket hadiah multi-komponen (Uang Tunai, XP, Barang Fisik, Badge, Voucher) | Memberikan insentif prestasi nyata yang beragam bagi peserta | P0 | FR-031, BR-019 |
| **BRQ-FIN-002** | Bisnis mewajibkan pembagian bounty proyek bersandar pada persentase kontribusi akhir (*Final Contribution %*) | Menjamin keadilan imbalan finansial berbasis kontribusi riil yang disahkan | P0 | FR-032, BR-016 |
| **BRQ-FIN-003** | Bisnis mewajibkan seluruh formula pemotongan pajak dan biaya dikendalikan secara dinamis via policy engine | Memastikan kepatuhan perpajakan yang adaptif terhadap regulasi | P0 | FR-034, BR-020 |
| **BRQ-FIN-004** | Bisnis mewajibkan pemisahan tegas antara iuran dana kas angkatan (*Batch Fund*) dengan pos pajak resmi negara | Menjaga kepatuhan hukum perpajakan dan perlindungan hak dana kebersamaan | P0 | FR-034, FR-036, BR-021 |
| **BRQ-FIN-005** | Bisnis mewajibkan dompet personal mencatat riwayat mutasi transparan (*gross*, deduksi, *net*, dan referensi sumber dana) | Menghilangkan keraguan peserta atas rincian potongan finansial | P0 | FR-035, BR-022 |
| **BRQ-FIN-006** | Bisnis mewajibkan seluruh transaksi moneter dicatat dalam buku kas ganda immutabel berpasangan ($\sum \text{Debit} = \sum \text{Kredit}$) | Menjamin auditabilitas moneter 100% dan mencegah saldo gantung (*zero discrepancy*) | P0 | FR-037, BR-023 |
| **BRQ-FIN-007** | Bisnis mewajibkan orkestrasi aliran finansial dieksekusi secara atomik ($\text{Gross} - \text{Tax} - \text{Farewell} = \text{Net}$) | Mencegah kegagalan parsial yang merusak konsistensi pembukuan kas | P0 | FR-039 |
| **BRQ-FIN-008** | Bisnis mewajibkan tersedianya alur pencairan dana dompet (*Payout Engine*) yang aman dengan validasi saldo | Memfasilitasi penerimaan hak finansial peserta ke rekening perbankan nyata | P1 | FR-038 |

---

## 13. Business Rules

| ID Aturan | Nama Aturan Bisnis | Kondisi Bisnis | Tindakan Sistem | Pengecualian |
|---|---|---|---|---|
| **BRULE-FIN-001** | *Dynamic Deduction Engine* | Pendapatan kotor diproses | Evaluasi seluruh aturan pajak/deduksi aktif; hitung nominal potongan secara dinamis | Aturan pajak Rp0 jika dibebaskan kebijakan |
| **BRULE-FIN-002** | *Farewell Fund Isolation* | Iuran perpisahan batch dipotong | Setor strictly ke akun kredit `Batch Fund`; dilarang masuk ke akun `Tax Payable` | Tidak ada pengecualian (BR-021) |
| **BRULE-FIN-003** | *Itemized Wallet Transparency* | Mutasi saldo dompet bertambah/berkurang | Wajib mencatat `gross_amount`, `deduction_amount`, `net_amount`, dan `source` referensi ID | Tidak ada pengecualian (BR-022) |
| **BRULE-FIN-004** | *Immutable Double-Entry Ledger* | Transaksi moneter dicatat ke ledger | Wajib seimbang $\sum \text{Debit} = \sum \text{Kredit}$; dilarang operasi UPDATE dan DELETE (append-only) | Koreksi kesalahan wajib melalui entri pembalik `REVERSAL` (BR-023) |
| **BRULE-FIN-005** | *Non-Negative Wallet Balance* | Penarikan dana dompet diajukan | Saldo dompet tidak boleh kurang dari Rp0; tolak penarikan melebihi saldo aktif | Tidak ada |
| **BRULE-FIN-006** | *Atomic Orchestration Rollback* | Terjadi kegagalan jaringan saat distribusi dana | Batalkan seluruh rangkaian mutasi seketika (*atomic rollback*) dan tandai transaksi `FAILED` | Tidak ada |

---

## 14. Business Entities

1. **Reward:** Master paket hadiah promosi rank (rank_id, component_type, title, monetary_value, description).
2. **Reward Claim:** Tiket operasional klaim hadiah (reward_id, user_id, status, shipping_address, tracking_number).
3. **Deduction / Tax Rule (DATA-005):** Master aturan pemotongan (name, type, rate_type, rate_value, calculation_basis, effective_date, is_active).
4. **Wallet:** Dompet digital personal pengguna (user_id, current_balance).
5. **Personal Wallet Transaction (DATA-006):** Mutasi transaksi dompet (wallet_id, source, reference_id, gross_amount, deduction_amount, net_amount, status).
6. **Batch Fund:** Rekening penampung dana bersama per angkatan (batch_id, total_accumulated, current_balance).
7. **Financial Ledger:** Buku besar akuntansi ganda immutabel (transaction_id, account_code, direction, amount, reference_table, narration).
8. **Payout:** Permohonan pencairan dana (wallet_id, amount, bank_code, account_number, account_holder_name, status).

---

## 15. Data Flow

```text
[ Gross Bounty Proyek / Reward Rank / Bonus Lembur ]
                          │
                          ▼
             [ Deduction & Tax Engine ]
                          │
            ┌─────────────┴─────────────┐
            ▼                           ▼
[ Akun Kewajiban Pajak ]    [ Rekening Batch Fund ]
            │                           │
            └─────────────┬─────────────┘
                          ▼
            [ Net Distributable Dihasilkan ]
                          │
                          ▼
             [ Personal Wallet Pengguna ]
                          │
                          ▼
   [ The Immutable Double-Entry Ledger (PostgreSQL) ]
   (Pencatatan Debit & Kredit Berpasangan Seimbang Rp 0)
                          │
                          ▼
        [ Permohonan Payout ──► Rekening Bank ]
```

---

## 16. Status & Lifecycle

```text
1. WALLET TRANSACTION LIFECYCLE:
   [ PENDING ] ──(Validasi Buku Besar Berhasil)──► [ COMPLETED ]
        │
        ├──(Kegagalan Transaksi)───────────────► [ FAILED ]
        └──(Koreksi Pembalik)──────────────────► [ REVERSED ]

2. PAYOUT LIFECYCLE:
   [ REQUESTED ] ──(Persetujuan Finance)──► [ PROCESSING ] ──(Transfer Berhasil)──► [ SETTLED ]
         │                                       │
         └──(Ditolak)──► [ REJECTED ]            └──(Transfer Gagal)──► [ FAILED ]
```

---

## 17. Approval & Decision Flow

* **Persetujuan Payout:** Requester: Intern / Alumni $\rightarrow$ Approver: Finance Administrator $\rightarrow$ Eksekusi transfer perbankan.
* **Pencairan Kas Angkatan (Batch Fund):** Requester: Koordinator Batch / HR $\rightarrow$ Approver: Program Lead & Finance Lead $\rightarrow$ Pencairan untuk acara perpisahan.

---

## 18. Exception & Failure Handling

* **Transaksi Ledger Tidak Seimbang:** Database trigger menolak eksekusi jika debit $\neq$ kredit; seluruh transaksi dibatalkan seketika.
* **Saldo Dompet Tidak Cukup:** Pengajuan payout otomatis ditolak dengan pesan *Insufficient Balance*.

---

## 19. Business Reporting

1. **Laporan Rekapitulasi Pemotongan Pajak (*Tax Withholding Report*):** Total pemotongan PPh terkumpul per periode untuk pelaporan kepatuhan pajak perusahaan.
2. **Laporan Neraca Saldo Buku Besar (*Ledger Balance Sheet*):** Audit rekonsiliasi kas keluar, utang kompensasi, saldo dompet, dan kas bersama.

---

## 20. KPI & Success Metrics

* **Bounty Settlement Turnaround Time (MET-04):** Durasi waktu dari pengesahan proyek hingga dana bersih masuk ke dompet (`Target: TBD`).
* **Net Payout Turnaround Time (MTR-FIN-01):** Durasi waktu dari pengajuan payout hingga transfer bank sukses (`Target: TBD`).

---

## 21. Business Integration

* **Dari Domain Projects:** Menerima *Final Contribution %* dan alokasi bounty pool.
* **Dari Domain Workforce:** Menerima data jam lembur aktual yang disahkan.
* **Dari Domain Performance:** Menerima event kenaikan rank untuk reward moneter.

---

## 22. Business Dependencies

* **Upstream:** Projects Module (kontribusi), Workforce Module (lembur), Performance Module (rank).
* **External:** Integrasi Payment Gateway / Bank API.

---

## 23. Business Constraints

* **Immutabilitas Akuntansi Mutlak (BR-023):** Dilarang menghapus atau mengubah baris buku besar; koreksi wajib via jurnal pembalik.
* **Presisi Moneter:** Perhitungan moneter menggunakan bilangan bulat fixed-point (Rupiah utuh tanpa selisih desimal).

---

## 24. Business Assumptions

* Peserta memiliki rekening bank nasional atau e-wallet aktif atas nama sendiri untuk menerima pencairan dana insentif.

---

## 25. Business Gaps

* Pemilihan mitra vendor Payment Gateway spesifik dan tata kelola otorisasi kas angkatan berstatus `[TBD]`.

---

## 26. Ambiguities

* *Iuran Batch Fund diklasifikasikan sebagai Batch Contribution, terpisah mutlak dari pajak negara (BR-021).*

---

## 27. Conflicting Requirements

* *Tidak ada konflik requirement.*

---

## 28. Traceability Matrix

| Kebutuhan PRD | Kebutuhan BRD | Aturan Bisnis Terkait | Entitas Terkait | Acceptance Criteria |
|---|---|---|---|---|
| FR-031 (Rank Rewards) | BRQ-FIN-001 | — | `Reward`, `Reward Claim` | AC-FIN-001 |
| FR-032 (Bounty Dist) | BRQ-FIN-002 | BRULE-FIN-001 | `Wallet Transaction` | AC-FIN-002 |
| FR-034 (Tax Engine) | BRQ-FIN-003, BRQ-FIN-004 | BRULE-FIN-001, BRULE-FIN-002 | `Deduction / Tax Rule` | AC-FIN-003 |
| FR-035 (Personal Wallet)| BRQ-FIN-005 | BRULE-FIN-003, BRULE-FIN-005 | `Wallet`, `Wallet Transaction` | AC-FIN-004 |
| FR-036 (Batch Fund) | BRQ-FIN-004 | BRULE-FIN-002 | `Batch Fund` | AC-FIN-003 |
| FR-037 (Financial Ledger)| BRQ-FIN-006 | BRULE-FIN-004 | `Financial Ledger` | AC-FIN-005 |
| FR-038 (Payout Engine) | BRQ-FIN-008 | BRULE-FIN-005 | `Payout`, `Wallet` | AC-FIN-006 |
| FR-039 (Orchestration) | BRQ-FIN-007 | BRULE-FIN-006 | `Financial Ledger`, `Wallet` | AC-FIN-002, AC-FIN-005 |

---

## 29. Acceptance Criteria

### AC-FIN-001: Orkestrasi Aliran Finansial & Pemisahan Pajak vs Batch Fund
* **Given:** Gross Bounty sebesar Rp1.000.000, aturan pajak PPh aktif 5%, dan iuran Batch Fund 2%.
* **When:** Transaksi distribusi bounty dieksekusi sistem.
* **Then:** Sistem menghitung pemotongan pajak Rp50.000 (disetor ke pos utang pajak), iuran batch Rp20.000 (disetor ke pos kas angkatan), dan mengkreditkan Rp930.000 ke dompet pengguna.

### AC-FIN-002: Keseimbangan Mutlak Buku Kas Ganda
* **Given:** Transaksi finansial sebesar Rp1.000.000 diproses.
* **When:** Sistem menuliskan baris jurnal ke tabel `Financial Ledger`.
* **Then:** Total nilai sisi Debit sama persis dengan total nilai sisi Kredit ($\sum \text{Debit} - \sum \text{Kredit} = 0$), dan record terkunci secara permanen (*append-only*).

### AC-FIN-003: Penahanan Saldo pada Pengajuan Payout
* **Given:** Intern memiliki saldo dompet Rp2.000.000 dan mengajukan payout sebesar Rp1.500.000.
* **When:** Form payout dikirimkan.
* **Then:** Saldo aktif yang tersedia untuk penarikan berikutnya menjadi Rp500.000, status payout menjadi `REQUESTED`, dan dana ditahan hingga diverifikasi Finance.

---

## 30. Open Questions

* `OQ-FIN-01 (Ref OQ-005)`: Pemilihan vendor spesifik Payment / Disbursement Gateway (`TBD`).
* `OQ-FIN-02 (Ref OQ-007)`: Mekanisme otorisasi dan pencairan kas bersama angkatan Batch Fund (`TBD`).

---

## 31. BRD Completion Checklist

- [x] Seluruh kebutuhan domain Finance, Taxation, Compensation, dan Ledger terdefinisi lengkap.
- [x] Aturan Double-Entry Ledger (BR-023) dan pemisahan Batch Fund vs Pajak (BR-021) terpetakan eksplisit.

---
*Dokumen ini disimpan permanen di `brd/05-BRD-FINANCE-COMPENSATION.md`.*
