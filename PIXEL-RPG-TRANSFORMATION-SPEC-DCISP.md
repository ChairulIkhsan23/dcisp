# DOKUMEN SPESIFIKASI TRANSFORMASI PIXEL RPG & GAMIFIED DESIGN SYSTEM
# Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0
**Document Version:** 1.0 — Pixel Adventure & Gamified System Architecture  
**Theme:** "The 8-Bit Professional Realm: Work As An Epic Progression Adventure"  
**Date:** 2026-09-19  
**Source of Truth:** `PRD-DCISP-V1.md` & `FINAL-TECH-STACK-SPEC-DCISP.md`  

---

## 1. VISI TRANSFORMASI & PRINSIP TEMA (DESIGN PHILOSOPHY)

Transformasi tema ini mengubah seluruh terminologi, antarmuka pengguna (UI/UX), mikro-interaksi, dan visual asset platform DCISP menjadi **Dunia Petualangan Pixel RPG Retro-Modern** (terinspirasi dari estetika *Codedex.io*, *Chrono Trigger*, dan *Final Fantasy Classic*). 

Transformasi ini **TIDAK MENGUBAH ATAU MENGURANGI INTEGRITAS BISNIS**:
1. Seluruh 25 Aturan Bisnis (**BR-001 s/d BR-025**) tetap berjalan 100% di backend Golang.
2. Skema database PostgreSQL (DATA-001 s/d DATA-007), audit trail, dan buku besar berpasangan tetap menggunakan prinsip akuntansi dan keamanan formal.
3. Batasan etika dan privasi kerja (**BR-010**: No invasive surveillance, No screenshot, No keylogger) tetap ditegakkan mutlak.

---

## 2. MASTER DICTIONARY: TRANSFORMASI TERMINOLOGI LENGKAP

| Domain Sistem | Terminologi Bisnis / PRD Asli | Terminologi Pixel RPG Theme | Deskripsi Konseptual & Representasi Visual |
|---|---|---|---|
| **Identity** | Super Admin | **The Creator (Sang Pencipta)** | Penguasa arsitektur tertinggi pengatur hukum dunia (*Unified Policy Engine*). |
| **Identity** | Admin | **Realm Warden (Pengawas Ranah)** | Pengelola operasional harian ranah, memantau pemain dan data umum. |
| **Identity** | HR / Internship Admin | **Game Master (GM) - Operations** | Pengatur server kohort angkatan (*Batch*), izin cuti, dan penentu MVP. |
| **Identity** | Project Manager | **Quest Giver (Pemberi Misi)** | Tokoh yang mempublikasikan misi bursa dan merekrut anggota party. |
| **Identity** | Supervisor | **Guild Master / Party Captain** | Pemimpin regu bimbingan, pemberi battle rating, dan pengesah kontribusi tim. |
| **Identity** | Reviewer | **The Oracle / Loot Validator** | Penilai kelayakan barang bukti kerja (*evidence*) dan deliverable tugas. |
| **Identity** | Finance Administrator | **The Merchant / Vault Keeper** | Penjaga perbendaharaan koin emas, pemungut upeti pajak, dan kasir pencairan. |
| **Identity** | Scanner Operator | **Gatekeeper (Penjaga Gerbang)** | Penjaga terminal fisik monolit scanner yang memvalidasi tap kartu jimat. |
| **Identity** | Intern | **Adventurer (Petualang / Hero Initiate)** | Pemain utama yang menyelesaikan misi harian dan menaikkan level rank. |
| **Identity** | Alumni | **Legendary Hero (Pahlawan Legendaris)** | Pemain veteran lulusan magang yang berhak mengambil misi publik ranah. |
| **Workforce** | Live Attendance (Check-in/out)| **Realm Login / Realm Logout** | Menempelkan jimat NFC (*Adventurer Amulet*) ke Monolit Terminal lobi. |
| **Workforce** | Work Schedule & Grace Period | **Realm Curfew & Grace Window** | Jam rotasi matahari ranah (08:30–17:00) dengan toleransi 10 menit. |
| **Workforce** | Work Session (Start / End Work)| **Grinding Mode (Enter / Exit Combat)** | Sesi bertarung melawan tugas kerja produktif harian. |
| **Workforce** | Break Time (Istirahat) | **Resting at the Inn / Campfire** | Mode istirahat; avatar duduk santai di depan api unggun dengan animasi tidur/makan. |
| **Workforce** | Early Break / Unauthorized Break| **Early Camp / Safe Zone Deserter** | Anomali istirahat di luar jadwal yang memicu kutukan damage penalti XP. |
| **Workforce** | Overtime (Lembur) | **Night Raid / Extra Grind** | Ekspedisi malam di luar jam kerja yang memerlukan restu Guild Master. |
| **Workforce** | Leave Management (Cuti) | **Safe Zone / Hibernation Potion** | Status perlindungan resmi dari kewajiban login tanpa penalti HP/XP (0 XP). |
| **Workforce** | Attendance Correction | **Chrono-Correction / Time Scroll** | Permohonan pemulihan data presensi dengan mempertahankan catatan sejarah asli. |
| **Workforce** | Session Integrity Tracking | **AFK Detector (Anti-Idling Spirit)** | Pendeteksi hilangnya fokus tab atau pemain tidur (>15 menit) dengan ikon *Zzz*. |
| **Projects** | Project Marketplace | **Bounty Board / Quest Board** | Papan pengumuman kayu di alun-alun desa tempat misi dipajang. |
| **Projects** | Project (Proyek) | **Epic Quest / Campaign** | Ekspedisi besar berhadiah koin emas dengan batas kapasitas party. |
| **Projects** | Milestone | **Checkpoint / Boss Battle** | Titik pencapaian fase misi perantara yang harus ditaklukkan regu. |
| **Projects** | Task (Tugas) | **Side Quest / Daily Mission** | Kartu tugas spesifik yang harus diselesaikan anggota party. |
| **Projects** | Work Report & Submission | **Battle Log (Catatan Pertarungan)** | Formulir ringkasan pencapaian progres %, kendala, dan rencana aksi. |
| **Projects** | Evidence (Bukti Deliverable) | **Loot Drop / Artifact Proof** | Tautan commit Git, screenshot UI, atau artefak dokumen di R2 Vault. |
| **Projects** | 3-Layer Contribution | **Three-Tier Honor Split** | Rekonsiliasi bagi hasil: Party Agreement $\rightarrow$ Battle Stats $\rightarrow$ Guild Master Verdict. |
| **Performance**| Internship XP | **Main Storyline XP** | Poin pengalaman progresi kedisiplinan dan standing magang aktif. |
| **Performance**| Project XP | **Guild XP** | Poin pengalaman dari hasil penuntasan deliverable proyek nyata. |
| **Performance**| Alumni XP | **Veteran XP** | Poin pengalaman kontribusi alumni (terisolasi dari ranking magang aktif). |
| **Performance**| Rank Progression | **Class Tier / Character Level** | Jenjang kelas petualang (*Novice $\rightarrow$ Apprentice $\rightarrow$ Knight $\rightarrow$ Paladin $\rightarrow$ Grandmaster*). |
| **Performance**| Performance Score | **Battle Honor Rating (0–100)** | Skor evaluasi formal kualitas mutu supervisor ($\text{Level} \neq \text{Honor Rating}$). |
| **Performance**| Top Performer per Batch | **MVP of the Season / Realm Champion** | Tepat SATU juara angkatan yang dipajang di *Hall of Fame / Hall of Heroes*. |
| **Performance**| Achievement Badges | **Epic Feats & Pixel Trophies** | Lencana prestasi unik berpiksel emas/perak yang dapat dipamerkan. |
| **Performance**| Penalty (-XP) | **HP Damage / Curse Debuff** | Efek layar bergetar (*screen shake*) dan teks damage merah saat terlambat. |
| **Finance** | Gross Bounty | **Gross Loot / Gold Chest** | Total peti harta koin emas yang dialokasikan untuk suatu Quest. |
| **Finance** | Tax & Deductions | **Realm Tribute (Pajak Kerajaan)** | Potongan resmi kas negara/perusahaan berbasis formula dinamis. |
| **Finance** | Batch Fund (Kas Perpisahan) | **Guild Treasury (Kas Serikat)** | Iuran sukarela bersama per angkatan untuk pesta perpisahan kohort. |
| **Finance** | Personal Wallet | **Coin Pouch / Inventory Wallet** | Kantong koin emas pribadi petualang dengan riwayat mutasi itemized. |
| **Finance** | Financial Ledger | **The Ancient Stone Tablet** | Buku besar akuntansi ganda abadi yang dipahat permanen (*append-only*). |
| **Finance** | Payout Engine | **Gold Vault Extraction / Cashout** | Pencairan koin emas dompet ke rekening perbankan nyata. |
| **Documents** | Digital ID Card | **Adventurer License / Guild Pass** | Kartu identitas petualang berornamen pixel art lengkap dengan segel QR code. |
| **Documents** | Digital Certificate | **Legendary Scroll of Graduation** | Gulungan sertifikat kelulusan digital resmi berstempel verifikasi publik. |
| **Documents** | Digital Portfolio | **Hero’s Chronicle (Buku Riwayat)** | Galeri karya dan pencapaian publik alumni yang dapat dibagikan ke dunia luar. |
| **Documents** | Document Reports / Logbook | **Adventurer’s Journal (Buku Harian)**| Ekspor berkas logbook dan laporan berkala magang format PDF. |
| **Storage** | Cloudflare R2 | **The Cloud Vault (Gudang Pusaka)** | Ruang penyimpanan artefak berkas tanpa batas terisolasi dari database. |
| **System** | Unified Policy Engine | **Codex of Creation (Kitab Hukum)** | Pusat pengaturan aturan main runtime ranah oleh The Creator. |
| **System** | Audit Log | **Akasha Chronicles (Buku Sejarah)** | Catatan sejarah abadi yang merekam setiap peristiwa di ranah tanpa bisa dihapus. |

---

## 3. TRANSFORMASI PERAN & IDENTITAS (10 HERO ARCHETYPES)

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                      THE REALM HEROES & ROLES                           │
├──────────────────────────┬──────────────────────────┬───────────────────┤
│ The Creator (SuperAdmin) │ Realm Warden (Admin)     │ GM - Ops (HR)     │
│ [Master World Architect] │ [Guardian of the Realm]  │ [Server Arbiter]  │
├──────────────────────────┼──────────────────────────┼───────────────────┤
│ Quest Giver (PM)         │ Guild Master (Supervisor)│ The Oracle (QA)   │
│ [Mission Publisher]      │ [Party Leader & Mentor]  │ [Loot Inspector]  │
├──────────────────────────┼──────────────────────────┼───────────────────┤
│ The Merchant (Finance)   │ Gatekeeper (Scanner Op)  │ Adventurer (Intern│
│ [Vault & Gold Keeper]    │ [Monolith Guardian]      │ [Hero Initiate]   │
├──────────────────────────┴──────────────────────────┴───────────────────┤
│ Legendary Hero (Alumni)                                                 │
│ [Veteran of the Realm]                                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

1. **Adventurer (Intern):** Karakter pemula bersenjatakan pena & keyboard kayu. Memulai dari Class Tier *Novice*, bertualang mengumpulkan XP, menaikkan level, menuntaskan Side Quest, dan mengklaim koin emas di dompetnya.
2. **Legendary Hero (Alumni):** Petualang veteran yang telah menuntaskan Main Storyline. Mengenakan jubah emas, memiliki akses ke Bounty Board khusus publik, mengumpulkan *Veteran XP*, dan membangun *Hero’s Chronicle* seumur hidup.
3. **Guild Master / Captain (Supervisor):** Pemimpin barak petualang yang mengawasi *Party Roster*, menyetujui izin Night Raid (lembur), menyembuhkan anomali presensi via Chrono-Correction, dan memberikan *Battle Honor Rating*.
4. **Quest Giver (Project Manager):** Papan misi master yang merancang *Epic Quests*, menentukan ukuran party, mengunci slot pelamar, membagi misi menjadi Boss Checkpoints, dan mengalokasikan peti *Gold Chest*.
5. **The Oracle (Reviewer):** Penjaga standar mutu yang memeriksa artefak *Loot Drop* (commit git, PR, screenshot). Memberikan status *Blessed (Approved)* atau *Cursed with Rework (Revision Required)*.
6. **Game Master - Operations (HR Admin):** Pengelola siklus hidup petualang dari pendaftaran guild hingga upacara kelulusan, penata jadwal *Realm Curfew*, dan pemilih 1 *MVP of the Season*.
7. **The Merchant (Finance Administrator):** Pengelola brankas kerajaan, pemungut upeti pajak *Realm Tribute*, pemelihara dana kas serikat *Guild Treasury*, dan pemverifikasi pencairan koin emas ke rekening bank.
8. **Gatekeeper (Scanner Operator):** Penjaga portal Monolit fisik di pintu gerbang kantor, strictly bertugas mengawasi tap jimat NFC petualang tanpa akses ke ruang rahasia guild lainnya (*scope = scanner_only*).
9. **The Creator (Super Admin):** Penguasa arsitektur sistem yang memegang *Codex of Creation*, mengonfigurasi master RBAC dinamis, dan memantau *Akasha Chronicles*.
10. **Realm Warden (Admin):** Operator keamanan dan operasional harian yang membantu administrasi akun, manajemen institusi akademi, dan pemeliharaan ranah.

---

## 4. WORKFORCE & IOT TERMINAL RITUALS (THE MONOLITH & NFC AMULET)

### 4.1 Ritual Kedatangan (Realm Login Ceremony)
* **Perangkat Monolit (Terminal ESP32-S3 + ACR1552U):** Terminal fisik di lobi gerbang didesain dengan casing retro pixel bertuliskan *"The DCISP Monolith"*.
* **Jimat Petualang (Adventurer Amulet):** Kartu NFC fisik yang dipegang setiap Adventurer.
* **Alur Ritual & Output Audio Monolit:**
  1. Petualang menempelkan jimat ke Monolit $\rightarrow$ Monolit berbunyi instan: **“Kartu terbaca. Stand by.”** (`card_read`).
  2. Monolit mengirimkan payload bertanda tangan ke Go Backend $\rightarrow$ Backend memvalidasi jadwal *Realm Curfew* (08:30 WIB + toleransi 10 menit).
  3. **Jika Tepat Waktu:** Monolit menyala hijau dan memutar: **“Good Morning! Absensi berhasil.”** (`check_in`) $\rightarrow$ Adventurer memperoleh +10 Main XP.
  4. **Jika Terlambat (>08:40 WIB):** Layar Monolit berkedip merah, efek suara petir 8-bit, dan audio bersuara: **“Perhatian. Anda terlambat.”** (`late`) $\rightarrow$ Adventurer terkena debuff HP Damage (-1 s/d -3 XP).
  5. **Jika Kartu Tidak Terdaftar:** Audio bersuara: **“Kartu belum terdaftar. Contact admin.”** (`card_unregistered`).

### 4.2 Mode Pertarungan Kerja (Grinding Mode / Enter Combat)
* Di portal *Player HUD (My Day)*, petualang menekan tombol arkade berdenyut **`[ENTER COMBAT]`** untuk memulai sesi kerja.
* **Avatar Sprite:** Karakter pixel Adventurer di layar mulai bergerak dalam animasi bertarung (misal: mengetik cepat di laptop bercahaya atau menebaskan pena sihir).
* **AFK Detector (BR-010):** Jika petualang meninggalkan tab browser atau tidak berinteraksi selama $>15$ menit, avatar berubah menjadi posisi tidur dengan balon animasi *Zzz*, dan status berubah menjadi `Idle (AFK)`.

### 4.3 Istirahat di Api Unggun (Resting at the Inn / Campfire)
* Pukul 12:00, petualang menekan tombol **`[REST AT INN]`** $\rightarrow$ Status berubah menjadi `Resting`.
* **Visual:** Layar beralih ke animasi api unggun (*campfire*), musik latar chiptune santai, dan penghitung waktu mundur 60 menit.
* **Audio Voice Monolit (jika via scanner):** **“Enjoy your break!.”** (`break_start`).
* **Selesai Istirahat Tepat Waktu:** Petualang menekan **`[RESUME COMBAT]`** $\rightarrow$ Audio: **“Welcome back! Silakan lanjut bekerja.”** (`break_end`).
* **Pelanggaran Terlambat Istirahat (>13:00 WIB):** Efek layar retak, audio *late penalty*, dan penalti -2 XP (*Unauthorized Break*).

### 4.4 Ritual Kepulangan (Realm Logout)
* Petualang menekan tombol **`[EXIT COMBAT]`** di Player HUD, mengisi formulir *Battle Log (Laporan Harian)*, lalu men-tap jimat NFC di Monolit keluar $\rightarrow$ Audio Monolit: **“Thank you, See you tomorrow.”** (`check_out`).

---

## 5. PROYEK, TUGAS & PENYELESAIAN MISI (QUESTS & THREE-TIER HONOR SPLIT)

```text
[ BOUNTY BOARD (Marketplace) ]
               │
               ▼
[ EPIC QUEST (Project) ] ──(Party Size / Quota Lock)
               │
               ├──► [ CHECKPOINT 1 (Milestone) ] ──► [ SIDE QUEST (Task) ]
               ├──► [ CHECKPOINT 2 (Milestone) ] ──► [ SIDE QUEST (Task) ]
               └──► [ FINAL BOSS (Completion) ]  ──► [ BATTLE LOG & LOOT DROP (Submission) ]
                                                              │
                                                              ▼ (Review by The Oracle)
                                               [ THREE-TIER HONOR SPLIT (FR-024) ]
                                               1. Party Agreement %  (Planned)
                                               2. Battle Stats %     (Actual)
                                               3. Guild Master Verdict % (Final)
                                                              │
                                                              ▼
                                                   [ GOLD CHEST UNLOCKED ]
```

### 5.1 Alur Rekrutmen Party (Quest Application)
1. Quest Giver (PM) menempelkan poster misi baru di *Bounty Board* dengan label:
   - `INTERN_ONLY`: Khusus Adventurer magang aktif.
   - `PUBLIC`: Terbuka untuk Adventurer & Legendary Heroes (Alumni).
   - `PRIVATE`: Misi rahasia dengan undangan khusus.
2. Adventurer melamar $\rightarrow$ The Oracle menghitung *Skill Compatibility Divination* (misal: 85% Affinity) murni sebagai saran bagi Quest Giver (*BR-015*).
3. Setelah kuota party penuh, slot terkunci otomatis (*Party Full*).

### 5.2 Pembagian Harta Tiga Lapis (Three-Tier Honor Split - BR-016)
Saat Final Boss tertaklukkan (Proyek selesai):
* **Layer 1 (Party Agreement %):** Alokasi awal yang disepakati saat party dibentuk (misal: Hero A: 40%, Hero B: 35%, Hero C: 25%).
* **Layer 2 (Battle Stats %):** Algoritma sistem menghitung kontribusi riil dari jumlah Side Quest yang diselesaikan, bobot kesulitan misi, dan rating bukti *Loot Drop*.
* **Layer 3 (Guild Master Verdict %):** Guild Master mengulas statistik dan mengesahkan persentase akhir resmi yang menjadi dasar pencairan peti emas.

---

## 6. GAMIFIKASI & FINANSIAL (LEVEL PROGRESSION & THE SACRED LEDGER)

### 6.1 Jenjang Kelas Petualang (Class Tiers & Level Progression)

$$\text{Class Tier (Peringkat)} \neq \text{Battle Honor Rating (Evaluasi Formal)}$$

| Level Tier | Nama Gelar Pixel RPG | Ambang Syarat (Default Policy) | Paket Hadiah Promosi (*Multi-Component Loot Package*) |
|---|---|---|---|
| **Tier 1** | **Novice (Petualang Pemula)** | 0 XP | Starter Wooden Badge + Adventurer License Digital. |
| **Tier 2** | **Apprentice (Magang Terampil)** | 250 Main XP | +50 Bonus XP + 50.000 Gold + Bronze Dagger Badge. |
| **Tier 3** | **Knight (Ksatria Kode & Tugas)** | 750 Main XP | +100 Bonus XP + 150.000 Gold + T-Shirt Guild Eksklusif + Silver Sword Badge. |
| **Tier 4** | **Paladin (Penjaga Keandalan Ranah)**| 1.500 Main XP | +250 Bonus XP + 300.000 Gold + Hoodie Eksklusif + Golden Shield Badge. |
| **Tier 5** | **Grandmaster / Legendary Hero** | 3.000 Main XP + Evaluasi | Master Cloak Merchandise + Crown Badge + Akses Misi Publik Abadi. |

### 6.2 Formula Aliran Emas Kerajaan (The Sacred Gold Flow - BR-021, BR-023)

$$\text{Gross Loot (Peti Emas)} - \text{Realm Tribute (Pajak)} - \text{Guild Treasury (Kas Bersama)} = \text{Pure Gold Pouch (Saldo Bersih)}$$

```text
[ GROSS LOOT: 1.000.000 Gold ]
               │
   ┌───────────┴───────────────────────────┐
   ▼                                       ▼
[ REALM TRIBUTE (Tax 5%): 50.000 Gold ]   [ GUILD TREASURY (Farewell 2%): 20.000 Gold ]
(Kas Resmi Negara / Perusahaan)           (Kas Serikat Angkatan Bersama)
   │                                       │
   └───────────────────┬───────────────────┘
                       ▼
           [ PURE GOLD POUCH: 930.000 Gold ]
           (Dikreditkan ke Coin Pouch Adventurer)
                       │
                       ▼
   [ THE IMMUTABLE ANCIENT STONE TABLET (PostgreSQL) ]
   • Debit:  Quest Bounty Pool Account  = 1.000.000
   • Credit: Adventurer Liability       =   930.000
   • Credit: Realm Tax Payable          =    50.000
   • Credit: Guild Treasury Fund        =    20.000
   (Keseimbangan Runes: Total Debit - Total Credit = 0)
```

---

## 7. UI/UX DESIGN SYSTEM: RETRO-MODERN PIXEL WORKSTATION

### 7.1 Tipografi & Desain Font
* **Header & Stat Numbers (Pixel Typography):** `Press Start 2P`, `VT323`, atau `Silkscreen` untuk level angka, XP counter, saldo emas, judul quest, dan status badge.
* **Isi Konten & Form Deskripsi (Clean Typography):** `Inter` atau `Geist Sans` dengan kontras tinggi (WCAG 2.1 AA) untuk deskripsi teknis panjang, form laporan *Battle Log*, dan kode pemrograman agar mata tidak cepat lelah saat bekerja seharian.

### 7.2 Palet Warna & Semantic Tokens

```text
┌────────────────────────────────────────────────────────────────────────┐
│                        PIXEL REALM COLOR TOKENS                        │
├───────────────────┬────────────────────────────────────────────────────┤
│ Obsidian Black    │ #0D1117 (Latar Belakang Ruang Kerja Gelap)         │
│ Retro Slate       │ #161B22 (Permukaan Kotak Dialog & Kartu Panel)     │
│ Pixel Border Cyan │ #38BDF8 (Garis Pinggir Panel 2px Solid Klasik)     │
├───────────────────┼────────────────────────────────────────────────────┤
│ Emerald Green     │ #10B981 (Status Success, On-Time, Active Grinding) │
│ Amber Sun         │ #F59E0B (Status Warning, Resting at Inn, Pending)  │
│ Crimson Dragon    │ #EF4444 (Status Late Penalty, Damage, Rejected)    │
│ Quest Gold        │ #FBBF24 (XP Progress Bar, Bounty Chest, MVP Crown) │
└───────────────────┴────────────────────────────────────────────────────┘
```

### 7.3 Arketipe Layar & Portal UI (Screen Archetypes)

```text
1. THE OVERWORLD MAP (Command Center — FR-051):
   • Peta dunia 8-bit interaktif membagi ranah menjadi 4 Region:
     - Region 1 (Today's Vanguard): Headcount pemain (Grinding, Resting, AFK).
     - Region 2 (Attention Volcano): Titik wilayah berkedip merah (anomali presensi, antrean approval).
     - Region 3 (Performance Citadel): Kurva XP ranah dan papan pengumuman MVP Season.
     - Region 4 (Treasury Vault): Aliran koin emas dan saldo Guild Treasury.

2. PLAYER HUD (Portal "My Day" — FR-052):
   ┌──────────────────────────────────────────────────────────────────┐
   │ [AVATAR SPRITE]  HERO: Andi (Lvl 3 Knight)    [XP BAR: 850/1500] │
   │ Status: GRINDING (Active: 02:45:12)           GOLD: 450.000 G    │
   ├──────────────────────────────────────────────────────────────────┤
   │ ACTIVE QUEST: [Side Quest #42: Build OAuth Gateway]              │
   │ CONTROLS: [ENTER COMBAT] [REST AT INN] [RESUME] [EXIT COMBAT]    │
   ├──────────────────────────────────────────────────────────────────┤
   │ TODAY'S LOGBOOK & QUEST TASKS:                                   │
   │  [x] Standup Ceremony (+5 XP)                                    │
   │  [ ] Implement JWT Guard (Due: 17:00)                            │
   └──────────────────────────────────────────────────────────────────┘

3. PARTY ROSTER (Portal "Team Today" — FR-053):
   • Deretan kartu profil anggota party bergaya pixel art cards.
   • Setiap kartu menampilkan mini-sprite animasi: pedang bergerak (Grinding), cangkir kopi (Resting), atau balon Zzz (AFK).
   • Tombol restu instan: [Approve Night Raid], [Cast Chrono-Correction].

4. INVENTORY MENU (Global Sidebar — FR-055):
   • Kotak navigasi samping biru klasik bergaya pause menu JRPG dengan border putih 2px.
   • Menu item:
     - 🗺️ Overworld Map (Command Center)
     - ⚔️ Player HUD (My Day)
     - 📜 Bounty Board (Projects Marketplace)
     - 🛡️ Party Roster (Team Today - Supervisor)
     - 🏆 Hall of Heroes (Leaderboard & MVP)
     - 💰 Coin Pouch (Personal Wallet)
     - 🏛️ Guild Treasury (Batch Fund)
     - 📜 Adventurer License (ID Cards & Certs)
     - ⚙️ Codex of Creation (Policy Engine - Creator)
```

### 7.4 Mikro-Interaksi, Animasi Sprite & Efek Suara (Audio & VFX)
1. **Screen Shake Damage:** Saat Adventurer terlambat check-in atau terkena penalti break (-XP), jendela browser bergetar ringan 300ms dengan efek suara damage 8-bit (*"Crunch/Hit"*).
2. **Level Up Fanfare:** Saat akumulasi XP mencapai ambang class tier baru, muncul pop-up modal dialog pixel emas bertuliskan **"LEVEL UP! YOU ARE NOW A KNIGHT!"** dengan animasi kembang api pixel dan fanfare melodi 8-bit pendek.
3. **Gold Drop Cascade:** Saat bagi hasil bounty disahkan, partikel koin emas berpiksel jatuh bergemerincing ke sudut widget *Coin Pouch*.
4. **Customizable Sprites:** Adventurer dapat memilih 1 dari 8 avatar pixel awal (Warrior Coder, Mage Designer, Rogue Tester, Paladin DevOps, dll.) yang mengekspresikan status kerja secara real-time.

---

## 8. MATRIX PEMETAAN LENGKAP KEBUTUHAN PRD (TRACEABILITY TABLE)

| ID Fitur PRD | Nama Fitur Asli PRD | Nama Elemen Pixel RPG | Status Kepatuhan Aturan Bisnis |
|---|---|---|---|
| **FR-001** | Dynamic RBAC Engine | **Guild Hierarchy & Permission Spells** | BR-001 (Zero Hardcode), BR-002 (Gatekeeper Isolation) |
| **FR-002** | Intern Lifecycle Management | **Adventurer Journey (Applicant $\rightarrow$ Hero)**| BR-003 (Permanent Account Retained) |
| **FR-003** | Alumni Lifecycle & Capabilities | **Legendary Hero Realm Access** | BR-003 (Public Quests), BR-004 (Veteran XP Partitions) |
| **FR-004** | Batch Management | **Guild Season / Cohort Server** | BR-018 (Singular MVP), BR-021 (Guild Treasury Bound) |
| **FR-007** | Work Schedule Engine | **Realm Curfew & Sun Cycles** | BR-007 (5-Way Time), BR-013 (Curfew Rules) |
| **FR-008** | Attendance Event Engine & NFC | **The Monolith Terminal & Amulet Ingestion** | BR-006 (Central Gate), Exact 15 Audio Scripts |
| **FR-009** | Work Session Tracking | **Grinding Combat Session** | BR-005 (Task Separation), BR-007 (Active vs Gross) |
| **FR-010** | Break State Engine | **Campfire Rest & Inn Sanctuary** | BR-008 (Break State), BR-009 (Camp Deserter Penalty) |
| **FR-011** | Session Integrity (Privacy) | **AFK Detector (No Spy Spells)** | BR-010 (Strict Privacy), BR-011 (Non-Punitive Blur) |
| **FR-012** | Overtime Tracking | **Night Raid Expeditions** | BR-013 (Captain's Approval), BR-014 (Actual Battle Time) |
| **FR-015** | Attendance Corrections | **Chrono-Correction Scroll** | BR-025 (Preserve Ancient Chronicle Log) |
| **FR-016** | Project Marketplace | **Bounty Board of the Realm** | BR-003 (Public Access), BR-005 (Quest vs Routine) |
| **FR-017** | Project Quota Workflow | **Party Enrollment & Slot Locking** | BR-015 (Human Captain Verdict, No Auto-Reject) |
| **FR-019** | Project Team Management | **Party Formation & Initial Honor Split** | BR-016 (Party Agreement Sum = 100%) |
| **FR-021** | Task Management | **Side Quests & Objective Cards** | BR-005 (Separation of Deliverables) |
| **FR-022** | Work Report / Submission | **Battle Log Dispatch** | BR-005 (Accountability Report) |
| **FR-023** | Evidence Management | **Loot Drop & Artifact Storage (R2)** | BR-024 (R2 Cloud Vault Storage, No DB BLOBs) |
| **FR-024** | Three-Layer Contribution | **Three-Tier Honor Split Engine** | BR-016 (Agreement $\rightarrow$ Stats $\rightarrow$ Verdict = 100%) |
| **FR-025** | XP Rules Engine | **Experience Points System (3 Pools)** | BR-004 (Main XP, Guild XP, Veteran XP), BR-012 |
| **FR-026** | Rank Progression System | **Class Tier & Level Promotions** | BR-017 ($\text{Class Level} \neq \text{Battle Honor Rating}$) |
| **FR-027** | Performance Evaluation Engine | **Battle Honor Rating (Formal 0–100)** | BR-017 (Quality Metric Distinct From Level) |
| **FR-028** | Top Performer per Batch | **MVP of the Season (Realm Champion)** | BR-018 (Strictly ONE Champion in Hall of Heroes) |
| **FR-031** | Multi-Component Rank Rewards | **Multi-Tier Loot Packages** | BR-019 (Gold + Bonus XP + Merchandise + Badges) |
| **FR-032** | Project Bounty Distribution | **Quest Gold Chest Split** | BR-016 (Tied to Final Honor Split), BR-020 |
| **FR-034** | Deduction & Tax Engine | **Realm Tribute Engine (Taxes)** | BR-020 (Dynamic Policy), BR-021 (Tribute vs Treasury) |
| **FR-035** | Personal Wallet | **Coin Pouch & Itemized Audit** | BR-022 (Itemized Gold Stream), BR-023 |
| **FR-036** | Batch Fund | **Guild Treasury (Farewell Fund)** | BR-021 (Isolated from Realm Taxes) |
| **FR-037** | Financial Double-Entry Ledger | **The Ancient Stone Tablet** | BR-023 (Immutable $\sum \text{Debit} = \sum \text{Credit}$, Append-Only) |
| **FR-040** | ID Card Generation | **Adventurer License (Guild Pass)** | Digital Pixel ID Card with NFC/QR Seal |
| **FR-041** | Certificate Engine | **Legendary Graduation Scroll** | Official Digital Scroll with Verification Hash |
| **FR-042** | Digital Portfolio Builder | **Hero’s Chronicle Showcase** | BR-003 (Public Legacy Showcase) |
| **FR-044** | Cloudflare R2 Integration | **The Cloud Vault Object Storage** | BR-024 (Physical Assets in R2, Metadata in SQL) |
| **FR-045** | Unified Policy Engine | **Codex of Creation** | BR-001, BR-012, BR-020 (Dynamic Runtime Laws) |
| **FR-047** | Audit Logging & Security | **Akasha Chronicles (History Ledger)** | BR-023, BR-025 (Append-Only Event Ledger) |
| **FR-051** | Command Center | **The Overworld Map Cockpit** | 4-Region World Monitoring (Vanguard, Volcano, etc.) |
| **FR-052** | Portal "My Day" | **Player HUD (Single-Screen Cockpit)**| Precision Monospace Timer + Combat Controls |
| **FR-053** | Portal "Team Today" | **Party Roster Operational HUD** | Real-time Member Status Cards (Sprites) |
| **FR-055** | Global Sidebar Navigation | **JRPG Inventory & Pause Menu** | Dynamic RBAC Route Isolation |

---

## 9. RINGKASAN KESIAPAN DESAIN & IMPLEMENTASI

1. **Selaras 100% dengan PRD & Tech Stack:**
   * Backend tetap menggunakan **Golang (Gin)** dengan arsitektur Modular Monolith.
   * Frontend tetap menggunakan **Next.js 15+ (App Router) + TypeScript + Tailwind CSS + Shadcn UI**.
   * Database tetap menggunakan **PostgreSQL 16+** dengan trigger append-only mutlak untuk pembukuan ganda dan audit log.
   * Hardware IoT presensi tetap menggunakan **ACR1552U + ESP32-S3 + MAX98357A** dengan 15 file audio global terstandar di Cloudflare R2 dan LittleFS cache.
2. **Karakter Visual yang Unik & Menyenangkan:**
   * Menghilangkan kesan membosankan dari software presensi dan enterprise task manager tradisional.
   * Memotivasi peserta magang melalui kepuasan leveling RPG, animasi sprite pixel yang hidup, visual audio retro, dan transparansi bagi hasil koin emas yang adil.

---
*Dokumen spesifikasi transformasi Pixel RPG ini disimpan secara permanen di `PIXEL-RPG-TRANSFORMATION-SPEC-DCISP.md`.*
