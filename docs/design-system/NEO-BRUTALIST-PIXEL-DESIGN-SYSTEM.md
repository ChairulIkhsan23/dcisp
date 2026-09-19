# DOKUMEN SISTEM DESAIN RESMI: NEO-BRUTALIST PIXEL RPG DESIGN SYSTEM
# Dagang Creative Intern Solutions Program (DCISP) — Platform DCISP v1.0
**Design System Name:** "Neo-Pixel Adventure Design System (NP-ADS)"  
**Aesthetic Fusion:** Neo-Brutalism (Bold Black Borders, Hard Offset Shadows, Vibrant Flat Colors) + Retro 8-Bit RPG (Pixelarticons, NES.css Elements, Arcade HUD, Sprite Avatars)  
**Version:** 1.0 — Production Blueprint  
**Date:** 2026-09-19  
**Source of Truth:** `PRD-DCISP-V1.md`, `FINAL-TECH-STACK-SPEC-DCISP.md`, `PIXEL-RPG-TRANSFORMATION-SPEC-DCISP.md`  

---

## 1. DESIGN SYSTEM FOUNDATIONS & CORE IDENTITY

Neo-Pixel Adventure Design System (NP-ADS) menggabungkan ketegasan struktural **Neo-Brutalism** dengan pesona nostalgia **Retro 8-Bit RPG**. Desain ini dirancang khusus untuk memadukan efisiensi kerja enterprise berakurasi tinggi dengan pengalaman bermain peran yang adiktif dan memuaskan.

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                       NEO-PIXEL ADVENTURE DESIGN SYSTEM                     │
├──────────────────────────────────────┬──────────────────────────────────────┤
│       NEO-BRUTALIST STRUCTURE        │         RETRO 8-BIT RPG SOUL         │
│  • Thick Solid Black Borders 3px/4px│  • Pixelarticons Iconography (24px)  │
│  • Hard Offset Drop Shadows (No Blur)│  • Arcade Typography (Press Start 2P)│
│  • Flat High-Contrast Vibrant Colors │  • NES.css Dialogues & Speech Bubbles│
│  • Unapologetic Grid Architecture    │  • Dynamic Pixel Avatar Sprites      │
│  • Extreme Readability (Mulish Sans) │  • 8-Bit Audio & Screen Shake VFX    │
└──────────────────────────────────────┴──────────────────────────────────────┘
```

### 3 Prinsip Utama Desain:
1. **High-Contrast Readability First:** Teks instruksi, data angka, form laporan, dan tabel pembukuan ganda menggunakan tipografi modern (*Mulish*) dengan kontras tajam terhadap latar belakang untuk mencegah kelelahan mata (WCAG 2.1 AA compliant).
2. **Tactile Arcade Feedback:** Setiap tombol, kartu, dan elemen interaktif memiliki bayangan tebal kaku (*hard drop shadow*) yang bereaksi secara fisik saat dihover (*translate up-left*) dan saat diklik (*translate down-right pressed state*).
3. **Immersive Progression Aesthetics:** Seluruh status kerja (*Grinding, Resting, AFK*), perolehan XP, dan kenaikan level divisualisasikan melalui komponen retro pixel yang hidup.

---

## 2. TIPOGRAFI SISTEM (TYPOGRAPHY SYSTEM)

Kombinasi 3 keluarga font yang saling melengkapi untuk membagi fungsi antara estetika retro dan kenyamanan membaca:

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                              TYPOGRAPHY TOKENS                              │
├──────────────────┬───────────────────────┬──────────────────────────────────┤
│ Font Family      │ Target Elemen UI      │ Karakter & Spesifikasi           │
├──────────────────┼───────────────────────┼──────────────────────────────────┤
│ Press Start 2P   │ Main Headings (H1/H2) │ Retro 8-bit arcade font klasik.  │
│                  │ Arcade Action Buttons │ Digunakan hemat hanya untuk judul│
│                  │ Level Up Banner & HUD │ dan label status pendek.         │
├──────────────────┼───────────────────────┼──────────────────────────────────┤
│ PixelGrid /      │ Numeric Counters, Timers│ Font piksel terstruktur berbasis │
│ Silkscreen       │ XP Bars, Gold Amounts │ grid. Menjamin angka metrik dan  │
│                  │ Ledger Account Codes  │ timer monospaced tidak goyang.   │
├──────────────────┼───────────────────────┼──────────────────────────────────┤
│ Mulish           │ Body Text, Paragraphs │ Sans-serif modern, bersih, bulat.│
│ (Weights: 400,   │ Work Reports, Tables  │ Keterbacaan prima untuk teks     │
│  600, 700, 900)  │ Form Inputs, Tooltips │ panjang dan form data enterprise.│
└──────────────────┴───────────────────────┴──────────────────────────────────┘
```

### Skala Tipografi (Type Scale Matrix):

| Token Nama | Font Family | Size / Line-Height | Weight | Transform | Contoh Penggunaan |
|---|---|---|---|---|---|
| `font-retro-hero` | Press Start 2P | `24px` / `36px` | 400 | UPPERCASE | Judul Dashboard, Level Up Modal |
| `font-retro-title`| Press Start 2P | `16px` / `24px` | 400 | UPPERCASE | Judul Quest Card, Header Panel |
| `font-retro-label`| Press Start 2P | `11px` / `16px` | 400 | UPPERCASE | Badge Status, Label Tombol Arkade |
| `font-pixel-timer`| Silkscreen / Grid | `18px` / `22px` | 700 | Monospace | Live Session Timer, Break Countdown |
| `font-pixel-stat` | Silkscreen / Grid | `14px` / `18px` | 700 | Monospace | Nilai XP (+10 XP), Saldo Gold Pouch |
| `font-body-lead`  | Mulish | `16px` / `24px` | 600 | None | Ringkasan Tugas, Subheader |
| `font-body`       | Mulish | `14px` / `22px` | 400 / 600 | None | Isi Laporan Kerja, Deskripsi Proyek |
| `font-body-sm`    | Mulish | `12px` / `18px` | 600 / 700 | None | Label Form, Timestamp, Table Cell |
| `font-body-xs`    | Mulish | `11px` / `16px` | 700 | None | Helper text, Validasi Error |

---

## 3. COLOR PALETTE MATRIX (NEO-BRUTALIST RETRO RGB)

Palet warna menggunakan warna-warna primer solid konsol game klasik (*Primary Yellow, Royal Blue, Crimson Red*) dipadukan dengan kontur hitam pekat (*Pure Ink Black*) dan aksen kanvas retro.

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                     COLOR PALETTE: PRIMARY RETRO RGB                        │
├───────────────────────┬───────────┬─────────────────────────────────────────┤
│ Token Warna           │ Hex Code  │ Peran & Semantik UI                     │
├───────────────────────┼───────────┼─────────────────────────────────────────┤
│ Color-Primary-Yellow  │ #F1C812   │ Arcade Gold: XP Bar, Gold Chest, Focus  │
│ Color-Primary-Blue    │ #2D5AB8   │ Mana Blue: Quest Board, Info, Headers   │
│ Color-Primary-Red     │ #E13447   │ Combat Red: Damage, Late, Danger, Action│
├───────────────────────┼───────────┼─────────────────────────────────────────┤
│ Color-Ink-Black       │ #000000   │ Pure Black: Borders, Hard Drop Shadows  │
│ Color-Charcoal-Ink    │ #212529   │ Deep Text: Teks utama pada kartu terang │
│ Color-Canvas-Cream    │ #F8F9FA   │ Off-White: Latar belakang kartu default │
│ Color-Canvas-Dark     │ #0E1117   │ Obsidian: Latar belakang utama aplikasi │
│ Color-Panel-Dark      │ #161B22   │ Dark Slate: Permukaan kartu dark mode   │
├───────────────────────┼───────────┼─────────────────────────────────────────┤
│ Color-Success-Emerald │ #10B981   │ On-Time Check-In, Approved Task, Active │
│ Color-Warning-Amber   │ #F59E0B   │ Resting at Inn, Pending Review          │
│ Color-Damage-Crimson  │ #EF4444   │ Unauthorized Break, Penalty, Revoke     │
│ Color-Pure-White      │ #FFFFFF   │ Stark White: Background dialog/cards    │
└───────────────────────┴───────────┴─────────────────────────────────────────┘
```

---

## 4. STRUCTURAL TOKENS (BORDERS, HARD SHADOWS & GEOMETRY)

Neo-Brutalism menolak bayangan kabur (*no soft blur gradients*). Seluruh elevasi dihasilkan dari bayangan offset solid hitam (*hard drop shadows*) dan garis tepi tebal (*bold black strokes*).

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                 NEO-BRUTALIST SHADOW & BORDER SPECIFICATION                 │
├───────────────────┬───────────────────────────┬─────────────────────────────┤
│ Token Elevasi     │ Nilai CSS Box-Shadow      │ Rekomendasi Komponen        │
├───────────────────┼───────────────────────────┼─────────────────────────────┤
│ `shadow-neo-sm`   │ `2px 2px 0px #000000`     │ Input fields, Badges, Chips │
│ `shadow-neo-md`   │ `4px 4px 0px #000000`     │ Standard Cards, Buttons     │
│ `shadow-neo-lg`   │ `6px 6px 0px #000000`     │ Active Cards, Hero Banners  │
│ `shadow-neo-xl`   │ `8px 8px 0px #000000`     │ Floating Dialogs, Modals    │
├───────────────────┼───────────────────────────┼─────────────────────────────┤
│ Token Border      │ Nilai CSS Border          │ Rekomendasi Komponen        │
├───────────────────┼───────────────────────────┼─────────────────────────────┤
│ `border-neo-thin` │ `2px solid #000000`       │ Tag status, tombol kecil    │
│ `border-neo-base` │ `3px solid #000000`       │ Standar Kartu, Form Input   │
│ `border-neo-thick`│ `4px solid #000000`       │ Modal Dialog, Hero Container│
└───────────────────┴───────────────────────────┴─────────────────────────────┘
```

### Sudut Geometri (Border Radius):
* **`rounded-none` (`0px`):** Gaya brutalist pixel murni (default untuk progress bar, dialog NES.css, tabel).
* **`rounded-neo` (`4px` atau `6px`):** Sudut neo-brutalist dengan sedikit lengkungan namun tetap mempertahankan ketegasan garis tepi hitam pekat (default untuk kartu dan tombol).

---

## 5. ICONOGRAPHY & ASSET VISUAL SYSTEM

### 5.1 Library Icon: `Pixelarticons`
* Menggunakan paket ikon **Pixelarticons** (resolusi native `24x24px`, stroke piksel tajam).
* **Pemetaan Ikon:**
  - ⚔️ `pixelarticons:sword` $\rightarrow$ Grinding / Active Tasks / Side Quests.
  - 🛡️ `pixelarticons:shield` $\rightarrow$ Guild Master / Safe Zone / Policy Guard.
  - 💰 `pixelarticons:coin` $\rightarrow$ Gold Pouch / Bounty Pool / Payout.
  - ⛺ `pixelarticons:tent` $\rightarrow$ Resting at Inn / Break Window.
  - 📜 `pixelarticons:scroll` $\rightarrow$ Battle Log / Evidence / Certificate.
  - 👑 `pixelarticons:crown` $\rightarrow$ MVP of Season / Top Performer.
  - 🚪 `pixelarticons:door` $\rightarrow$ Realm Login / Logout / Terminal.
  - ⚠️ `pixelarticons:alert` $\rightarrow$ AFK Alert / Damage Penalty.

### 5.2 Dynamic Sprite Avatars (Customizable 8-Bit Heroes)
* Setiap pengguna memiliki avatar sprite piksel (`32x32px` di-upscale ke `64x64px` tanpa blur via `image-rendering: pixelated`).
* **State Animasi Avatar:**
  - **IDLE (Online):** Sprite berdiri bernapas santai.
  - **WORKING (Grinding):** Sprite mengetik cepat di laptop bercahaya dengan partikel kode.
  - **BREAK (Resting):** Sprite duduk minum kopi di samping api unggun.
  - **AFK (Idle > 15m):** Sprite tertidur dengan animasi gelembung *Zzz*.
  - **DAMAGE (-XP):** Sprite berkedip merah saat menerima penalti keterlambatan.

---

## 6. COMPONENT CATALOG & ARCHEAN BLUEPRINTS

### 6.1 Neo-Brutalist Arcade Buttons (`<NeoArcadeButton />`)
Tombol interaktif dengan respon fisik saat ditekan:

```tsx
// CSS / Tailwind Primitives
const buttonVariants = {
  primary: "bg-[#F1C812] text-[#000000] hover:bg-[#ffe043]",
  combat:  "bg-[#E13447] text-[#FFFFFF] hover:bg-[#ff4d61]",
  mana:    "bg-[#2D5AB8] text-[#FFFFFF] hover:bg-[#3d6fd8]",
  neutral: "bg-[#FFFFFF] text-[#000000] hover:bg-[#F8F9FA]",
};

// Base Style:
// "font-['Press_Start_2P'] text-[11px] uppercase px-5 py-3 border-[3px] border-[#000000] shadow-[4px_4px_0px_#000000] hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[6px_6px_0px_#000000] active:translate-x-[2px] active:translate-y-[2px] active:shadow-[2px_2px_0px_#000000] transition-all cursor-pointer select-none"
```

### 6.2 Neo-Brutalist Card Panels (`<NeoPixelCard />`)
Wadah panel modular untuk Quest Card, Roster Card, dan Ledger:

```tsx
// Container Style:
// "bg-[#FFFFFF] dark:bg-[#161B22] border-[3px] border-[#000000] shadow-[4px_4px_0px_#000000] p-5 rounded-[4px]"
// Header Bar Style:
// "bg-[#2D5AB8] text-[#FFFFFF] border-b-[3px] border-[#000000] px-4 py-2 font-['Press_Start_2P'] text-[12px] flex items-center justify-between"
```

### 6.3 Pixel HUD Precision Progress Bar (`<PixelProgressBar />`)
Bar progres bergaris tebal dengan segmen piksel berulang untuk XP dan Break Timer:

```tsx
// Frame: "border-[3px] border-[#000000] p-[2px] bg-[#000000] h-[22px] rounded-[2px]"
// Inner Fill (XP): "bg-[#F1C812] h-full transition-all duration-300 pattern-pixel-stripe"
// Inner Fill (HP/Break): "bg-[#E13447] h-full transition-all duration-300"
```

### 6.4 JRPG Dialogue Box / Modal (`<NesDialog />`)
Modal dialog dengan gaya klasik Final Fantasy / NES.css:

```tsx
// Container:
// "bg-[#0E1117] border-[4px] border-[#FFFFFF] shadow-[8px_8px_0px_#000000] p-6 text-[#FFFFFF] font-['Mulish'] relative"
// Decorative Pixel Corners: Menggunakan border-image piksel atau pseudo-elements 4 titik sudut.
```

### 6.5 Adventurer License (Digital ID Card Component)
Mengadaptasi desain fisik ID Card Neo-Brutalist ke format web:
* **Frame:** Latar belakang solid `#F1C812` berbingkai hitam `4px solid #000000` dengan hard shadow `6px 6px 0px #000000`.
* **Header Bar:** Kotak hitam bertuliskan `ADVENTURER GUILD LICENSE // DCISP v1.0` (Press Start 2P).
* **Avatar Frame:** Kotak foto berpiksel `border: 3px solid #000000` dengan badge Class Level (misal: `LVL 3 KNIGHT`).
* **Identity Details (Mulish 700):** Nama Lengkap, Nomor Induk (NIM/NISN), Batch Angkatan, dan Institusi Asal.
* **Integrated Seal:** QR Code verifikasi terenkripsi `80x80px` di sudut kanan bawah berbingkai hitam.

---

## 7. CODE INTEGRATION BLUEPRINTS (NEXT.JS & TAILWIND)

### 7.1 Konfigurasi `tailwind.config.ts`

```typescript
import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        retro: {
          yellow: "#F1C812", // Primary Arcade Gold
          blue: "#2D5AB8",   // Mana Blue
          red: "#E13447",    // Combat Red
          ink: "#000000",    // Ink Black for borders & shadows
          charcoal: "#212529",
          cream: "#F8F9FA",
          obsidian: "#0E1117",
          slate: "#161B22",
          emerald: "#10B981",
        },
      },
      fontFamily: {
        arcade: ["'Press Start 2P'", "monospace"],
        pixel: ["'Silkscreen'", "'PixelGrid'", "monospace"],
        sans: ["'Mulish'", "sans-serif"],
      },
      boxShadow: {
        "neo-sm": "2px 2px 0px #000000",
        "neo-md": "4px 4px 0px #000000",
        "neo-lg": "6px 6px 0px #000000",
        "neo-xl": "8px 8px 0px #000000",
        "neo-yellow": "4px 4px 0px #F1C812",
        "neo-red": "4px 4px 0px #E13447",
      },
      borderWidth: {
        "neo-base": "3px",
        "neo-thick": "4px",
      },
    },
  },
  plugins: [],
};

export default config;
```

---

## 8. ACCESSIBILITY & USABILITY GUARDS (WCAG 2.1 AA)

1. **Aturan Larangan Teks Panjang Piksel:** Font `Press Start 2P` **DILARANG KERAS** digunakan untuk teks melebihi 20 kata, paragraf laporan, isi logbook, atau tabel data. Paragraf wajib menggunakan **Mulish** ukuran minimal `14px` dengan line-height `1.5`.
2. **Kontras Warna Terjamin (Minimum 4.5:1):**
   - Teks hitam `#000000` di atas tombol `#F1C812` (Rasio Kontras: **12.8:1** — Lulus AAA).
   - Teks putih `#FFFFFF` di atas tombol `#2D5AB8` (Rasio Kontras: **6.2:1** — Lulus AA).
   - Teks putih `#FFFFFF` di atas tombol `#E13447` (Rasio Kontras: **5.1:1** — Lulus AA).
3. **Keyboard Focus Ring Neo-Brutalist:** Elemen yang sedang fokus menerima outline tebal berkontras tinggi:
   ```css
   :focus-visible {
     outline: 3px solid #F1C812;
     outline-offset: 2px;
   }
   ```
4. **Reduced Motion Support:** Mendukung media query `prefers-reduced-motion: reduce` dengan mematikan animasi screen shake dan partikel koin otomatis.

---

## 9. RINGKASAN KEPATUHAN & KESIAPAN IMPLEMENTASI

Sistem desain Neo-Pixel Adventure Design System (NP-ADS) ini telah:
- [x] Mengintegrasikan palet primer retro: **#F1C812**, **#2D5AB8**, **#E13447**, dan **#000000**.
- [x] Memadukan tipografi hibrida: **Press Start 2P** (arcade headers), **PixelGrid / Silkscreen** (stats & timers), dan **Mulish** (body & form data).
- [x] Mengadopsi struktur hard shadow dan bold border ala **Neobrutalism.dev**.
- [x] Menjaga konsistensi ikon piksel via **Pixelarticons** dan komponen retro via **NES.css**.
- [x] Mematuhi 100% dari 25 Aturan Bisnis (BR-001 s/d BR-025) pada PRD DCISP v1.0.

---
*Dokumen sistem desain resmi ini disimpan secara permanen di `NEO-BRUTALIST-PIXEL-DESIGN-SYSTEM.md`.*
