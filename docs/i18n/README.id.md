# Open Novel — Platform novel multibahasa global

<div align="center">

[中文](../../README.md) · [English](README.en.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Русский](README.ru.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [বাংলা](README.bn.md) · **Bahasa Indonesia**

</div>

> Platform membaca novel multibahasa global berbasis arsitektur mikroservis **Go-Kratos** dengan frontend multi-platform **Flutter / HarmonyOS**, mendukung **lebih dari 12 bahasa utama**, dan dirancang untuk memberikan kemampuan membaca, interaksi, pencarian, dan rekomendasi personal kepada pengguna di seluruh dunia.

---

## Pengenalan proyek

Open Novel adalah platform novel multibahasa global dengan arsitektur mikroservis cloud-native:

- **Backend**: Go-Kratos v2 (protokol ganda gRPC / HTTP), mikroservis yang dibagi per domain (pengguna, buku, bab, komentar, pencarian, rekomendasi)
- **Frontend**: Flutter lintas platform (Web / Desktop / Mobile) + aplikasi native HarmonyOS NEXT, berbagi satu set API backend yang sama
- **Multibahasa**: pemuatan dinamis sumber daya i18n, mendukung lebih dari 12 bahasa (Mandarin, Inggris, Jepang, Korea, Prancis, Jerman, Spanyol, Rusia, Arab, dll.)
- **Penyimpanan**: MySQL 8 (master-slave dengan pemisahan baca/tulis) + Redis (cache data panas / sesi) + OpenSearch (pencarian multibahasa)
- **Operasional**: deployment satu klik dengan Docker Compose, pemantauan Prometheus + Grafana, integrasi berkelanjutan GitHub Actions


## Fitur

<p align="center"><img src="images/features.svg" alt="Diagram arsitektur fitur" width="860"/></p>

- **Pusat pengguna**: registrasi dan login (JWT), rak buku pribadi, sinkronisasi progres baca lintas perangkat, profil multibahasa
- **Pengalaman membaca**: membaca per bab, pergantian font dan ukuran, tema terang/gelap, cache offline, animasi ganti halaman
- **Konten buku**: metadata buku, manajemen bab, tag kategori, pembaruan berseri, terjemahan multibahasa
- **Komunitas interaktif**: komentar dan ulasan, suka, favorit, pelaporan dan moderasi
- **Pencarian dan penemuan**: pencarian dengan segmentasi multibahasa, peringkat populer, rekomendasi AI, penjelajahan kategori
- **Panel admin**: moderasi konten, manajemen pengguna, statistik data, manajemen konfigurasi

## Arsitektur sistem

<p align="center"><img src="images/architecture.svg" alt="Diagram arsitektur sistem" width="860"/></p>

## Gambaran umum proyek

<p align="center"><img src="images/project.svg" alt="Diagram gambaran umum proyek" width="860"/></p>

## Siklus permintaan

<p align="center"><img src="images/request-cycle.svg" alt="Diagram siklus permintaan" width="860"/></p>

## Arsitektur keamanan

<p align="center"><img src="images/security.svg" alt="Diagram arsitektur keamanan" width="860"/></p>

## Struktur proyek

<p align="center"><img src="images/structure.svg" alt="Diagram struktur proyek" width="860"/></p>

---

## Tumpukan teknologi

| Lapisan | Teknologi |
| :--- | :--- |
| Klien | Flutter（Web / Desktop / Mobile）、HarmonyOS NEXT（ArkTS / ArkUI） |
| Gateway | Nginx + CDN、Go-Kratos API Gateway（protokol ganda gRPC / HTTP） |
| Server | Go 1.22+、Kratos v2、protobuf / gRPC |
| Penyimpanan | MySQL 8.0（master-slave）、Redis 7.x（Cluster）、OpenSearch 2.x |
| Observabilitas | Prometheus、Grafana、ELK、penelusuran rantai OpenTelemetry |
| Operasional | Docker Compose、GitHub Actions CI/CD |

## Basis data

- Nama basis data: `novel`
- Prefiks tabel: `novel_` (misalnya `novel_user`, `novel_book`, `novel_chapter`, `novel_comment`, dll.)

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Lihat desain tabel terperinci dan strategi pemisahan baca/tulis di [docs/novel-project-planning.md](../novel-project-planning.md).

## Direktori multi-platform

```
apps/
├─ flutter/     # Flutter 全平台（Web / Desktop / Mobile），i18n 多语言
└─ harmonyos/   # HarmonyOS NEXT 原生应用（ArkTS / ArkUI）
```

Lihat [apps/README.md](../../apps/README.md) untuk detailnya.

## Peta jalan

| Fase | Periode | Fokus tugas |
| :--- | :--- | :--- |
| Phase 1 | 2-3 minggu | Layanan dasar backend Kratos + integrasi MySQL / Redis / OpenSearch |
| Phase 2 | 3-4 minggu | Frontend multi-platform Flutter / HarmonyOS + penulisan ARB multibahasa |
| Phase 3 | 2 minggu | Penguatan keamanan (JWT / RBAC / pembatasan laju) + uji beban |
| Phase 4 | 1-2 minggu | Integrasi seluruh alur + konfigurasi akselerasi CDN |
| Phase 5 | Berkelanjutan | Integrasi algoritme rekomendasi AI, pelacakan analisis perilaku pengguna |

## Pengembangan lokal

```bash
# 启动依赖（MySQL / Redis / OpenSearch）
docker compose up -d

# 后端服务（Kratos 工作区）
cd backend && go mod tidy && go run ./cmd/server

# Flutter 端
cd apps/flutter && flutter pub get && flutter run

# HarmonyOS 端
cd apps/harmonyos && hvigorw assembleHap
```

---

## Dukungan dan donasi

Jika proyek ini bermanfaat bagi Anda, silakan dukung dengan **Star** atau **Fork**; donasi dengan memindai kode QR juga sangat kami hargai. Setiap dukungan Anda adalah motivasi saya untuk terus memelihara dan memperbarui proyek ini. Terima kasih atas dukungannya!

<div align="center">

**Donasi WeChat** ｜ **Donasi Alipay**

<img src="../weixinpay.png" width="130" height="130" alt="Kode donasi WeChat" />　<img src="../alipay.png" width="130" height="130" alt="Kode donasi Alipay" />

</div>

### Donasi transfer global (remitansi lintas negara)

【Informasi penerima】

- Nama penerima: WANG KEXUN
- Nomor akun penerima: 881015918251

【Bank penerima】

- ZA Bank SWIFT Code: AABLHKHHXXX
- Nama bank: ZA Bank Limited
- Kode bank: 387
- Alamat bank: Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【Bank koresponden remitansi lintas negara (jika diperlukan)】

> Perlu diperhatikan bahwa ini adalah informasi bank koresponden (bank perantara) untuk remitansi lintas negara, bukan informasi bank penerima. Tanyakan kepada bank pengirim apakah diperlukan informasi bank koresponden untuk remitansi lintas negara.

**Bank koresponden untuk setoran HKD, CNY, dan USD adalah Citibank**

- Nama bank: Citibank N.A. Hong Kong
- SWIFT Code: CITIHKHXXXX
- Kode bank: 006
- Nama cabang: Hong Kong Branch
- Kode cabang: 391
- Alamat bank: Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**Bank koresponden untuk setoran mata uang lainnya adalah BNY Mellon**

- Nama bank: THE BANK OF NEW YORK MELLON
- SWIFT Code: IRVTUS3NXXX
- Alamat bank: THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States
