# Dokumen Tambahan (Supplemental Docs)

Folder ini dikhususkan untuk **dokumen tambahan** di luar dokumen inti project
(`docs/prd/`, `docs/brd/`, `docs/architecture/`, `docs/guide/`, `docs/design-system/`).

Isi folder ini **dilacak oleh Git** (pengecualian terhadap aturan ignore `docs/*`
pada `.gitignore`), sehingga cocok untuk:

- Panduan integrasi pihak ketiga (contoh: API Indonesia, Cloudflare R2, ESP32).
- Catatan migrasi / operasional tambahan.
- ADRs (Architecture Decision Records) tambahan.

Aturan:

- Jangan menduplikasi dokumen inti; cukup rujuk path dokumen inti yang relevan.
- Jangan menyimpan secret / API key / kredensial dalam bentuk apa pun.
- Gunakan Bahasa Indonesia yang baku untuk penjelasan operasional.
