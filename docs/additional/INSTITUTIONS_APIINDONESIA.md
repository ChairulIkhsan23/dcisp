# Integrasi Katalog Institusi — Public API API Indonesia

> Data katalog institusi (kampus + sekolah) diperoleh dari Public API API Indonesia
> (`https://use.apiindonesia.id`) dan **bukan** daftar statis/manual sebagai source of truth utama.

## 1. Data Source

- **Provider:** API Indonesia
- **Base URL:** `https://use.apiindonesia.id` (dapat dioverride via `API_INDONESIA_BASE_URL`)
- **Endpoint terverifikasi (docs https://docs.apiindonesia.id):**
  - `GET /api/v1/kampus?q=&province=&regency=&type=&group=&page=&per_page=` — direktori 2.100+ PT (universitas, institut, politeknik, akademi, sekolah tinggi)
  - `GET /api/v1/kampus/search?q=` — pencarian cepat (maks 50 hasil, `q` min 2 karakter)
  - `GET /api/v1/kampus/:id` — detail satu kampus (`pt_001`, dst.)
  - `GET /api/v1/sekolah?...&q=&provinsi_id=&kabupaten_id=&jenis=&status=` — direktori 200.000+ sekolah (NPSN)
  - `GET /api/v1/sekolah/search?q=` — pencarian sekolah (maks 50 hasil, `q` min 3 karakter)
  - `GET /api/v1/sekolah/:npsn` — detail satu sekolah
- **Autentikasi:** header `x-api-key: aip_live_...` untuk seluruh endpoint `/api/v1/*`
- **Envelope sukses:** `{ "data": [...], "meta": { "total", "page", "per_page", "total_pages" } }`
- **Envelope gagal:** `{ "error": { "code", "message" } }`
- **Rate limit:** 20 req/detik per key → `429 RATE_LIMIT_EXCEEDED`; kuota habis → `402 QUOTA_EXCEEDED`

## 2. Security — Secret Handling

- Kredensial **HANYA** melalui environment:
  ```env
  API_INDONESIA_BASE_URL=https://use.apiindonesia.id
  API_INDONESIA_KEY=
  ```
- Nilai asli key **JANGAN** di-commit, ditulis di README/docs/test/migration/seed/log, atau dikirim ke frontend.
- File contoh (`.env.example`) hanya berisi placeholder kosong.
- Audit log (`IMPORT_EXTERNAL_INSTITUTION`) hanya mencatat `name`, `external_id`, `source` — tidak pernah mencatat key.

## 3. Arsitektur

```text
API Indonesia (kampus/sekolah)
      ↓  x-api-key, timeout 10s
APIIndonesiaClient (external DTO, error mapping aman)
      ↓
Service (cache Redis 30 mnt, mapper, idempotensi)
      ↓
Internal Institution Domain (PostgreSQL institutions)
      ↓
Application (interns.institution_id FK RESTRICT)
```

- **External DTO** (`APIIndonesiaKampus`, `APIIndonesiaSekolah`) terpisah dari model internal.
- **Mapper** (`MapKampusToInstitutionDraft`, `MapSekolahToInstitutionDraft`) mengonversi ke draf internal.
- Business logic tetap di service layer; client hanya request/auth/timeout/parse/pagination.

## 4. Identifier Mapping & Database

Migrasi `000004_add_institution_external_fields`:

- `institutions.external_id VARCHAR(128)` — format stabil `kampus:<id>` atau `sekolah:<npsn>`
- `institutions.source VARCHAR(32)` — `MANUAL`, `API_KAMPUS`, `API_SEKOLAH`
- `institutions.synced_at TIMESTAMPTZ`
- Constraint: `UNIQUE(external_id) WHERE external_id IS NOT NULL` (mencegah duplikasi sinkronisasi + aman terhadap race)
- Index: `idx_institutions_source`

Relasi existing (`interns.institution_id → institutions.id ON DELETE RESTRICT`, guard `BRULE-PEO-005`) **tidak berubah**.

## 5. Endpoint Internal Baru (backward compatible)

CRUD lama (`POST/GET/PUT/DELETE /api/v1/people/institutions`) dipertahankan apa adanya.

- `GET /api/v1/people/institutions/search-external?q=&source=ALL|KAMPUS|SEKOLAH&province=&regency=&page=&per_page=`
- `GET /api/v1/people/institutions/external/:source/:external_id` (`source` = `API_KAMPUS`/`API_SEKOLAH`)
- `POST /api/v1/people/institutions/import-external` body `{ "source": "API_KAMPUS", "external_id": "pt_001" }` → `201` + `isNew` implisit via pesan

## 6. Failure Handling & Cache

- Timeout HTTP 10 detik via `http.Client`; context request diteruskan.
- Error eksternal dipetakan ke pesan aman (tanpa `dial tcp`, SQL, stack trace, atau secret).
- Cache Redis 30 menit dengan key `apiindonesia:institutions:<source>:<q>:<prov>:<reg>:<page>:<perPage>`; kegagalan cache tidak menggagalkan request.
- Fallback: database internal tetap menjadi sumber relasi; bila API down, pencarian eksternal mengembalikan error informatif, data lokal tetap dapat dibaca.
- Tidak ada fallback berupa data palsu.

## 7. Data Ownership

- **Data katalog eksternal** (nama, alamat, wilayah, akreditasi) → milik API Indonesia.
- **Data relasi internal** (penugasan `interns.institution_id`, kontak koordinator manual, audit) → milik database aplikasi dan tidak dikirim ke API eksternal.
