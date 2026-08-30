# Open Novel — Global Multilingual Novel Platform

<div align="center">

[中文](../../README.md) · **English** · [日本語](README.ja.md) · [한국어](README.ko.md) · [Русский](README.ru.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [বাংলা](README.bn.md) · [Bahasa Indonesia](README.id.md)

</div>

> A global multilingual novel reading platform built on the **Go-Kratos** microservice architecture with **Flutter / HarmonyOS** multi-client frontends, supporting **12+ major languages** and delivering reading, interaction, search and personalized recommendation capabilities to users worldwide.

<div align="center"><img src="../mascot.svg" alt="Open Novel mascot Novi" width="150"/></div>

---

## Project Overview

Open Novel is a cloud-native, microservice-architecture global multilingual novel platform:

- **Backend**: Go-Kratos v2 (gRPC / HTTP dual protocol), microservices split by domain (users, books, chapters, comments, search, recommendations)
- **Frontend**: Flutter all-platform (Web / Desktop / Mobile) + HarmonyOS NEXT native app, sharing the same set of backend APIs
- **Multilingual**: Dynamically loaded i18n resources, supporting 12+ languages (Chinese, English, Japanese, Korean, French, German, Spanish, Russian, Arabic, etc.)
- **Storage**: MySQL 8 (master-slave read/write separation) + Redis (hot cache / sessions) + OpenSearch (multilingual search)
- **Operations**: Docker Compose one-click deployment, Prometheus + Grafana monitoring, GitHub Actions continuous integration


## Features

<p align="center"><img src="images/en/features.svg" alt="Feature architecture diagram" width="860"/></p>

- **User Center**: Registration and login (JWT), personal bookshelf, cross-device reading progress sync, multilingual profiles
- **Reading Experience**: Chapter-by-chapter reading, font and size switching, light/dark themes, offline caching, page-turn animations
- **Book Content**: Book metadata, chapter management, category tags, serialized updates, multilingual translation
- **Interactive Community**: Comments and reviews, likes, favorites, reporting and moderation
- **Search & Discovery**: Multilingual tokenized search, hot keyword rankings, search suggestions (client-side local history of 20 entries + 200 ms debounced suggestions), AI recommendations, category browsing
- **Admin Console**: Content moderation, user management, data statistics (dashboard / DAU / rankings / behavior analytics /api/stats/behavior), configuration management (category tags), machine translation workflow (DeepL, /api/admin/translate/*, admin "Translation" page + manual editing), audit log query (/api/admin/audit-logs), CDN provider management (multi-vendor configuration / enable-disable / ordering, instant effect via hot reload)
- **Payments & VIP**: Multi-channel payments via 11 providers (Stripe, NOWPayments (USDT), Razorpay, KOMOJU, PortOne, Mercado Pago, Xendit, PayPal, Alipay, WeChat Pay Global, UnionPay), VIP plan subscription and renewal, language-based payment method routing (WeChat Pay Global integrated; domestic WeChat Pay not integrated, requires CN merchant qualification)

## System Architecture

<p align="center"><img src="images/en/architecture.svg" alt="System architecture diagram" width="860"/></p>

Overall Go-Kratos microservice architecture: Flutter / HarmonyOS clients interact with the API gateway via Nginx + a multi-vendor CDN (Cloudflare / CloudFront on the global route, Aliyun / Tencent Cloud on the China route; admin-configurable, config changes apply instantly via hot reload); the gateway routes by domain to backend services such as users, books, chapters, comments, search and recommendations; the data layer is MySQL master-slave (read/write separation) + Redis cache + OpenSearch search index. Services communicate via gRPC, and external HTTP APIs uniformly use the `/api` prefix.

Other diagrams: project panorama [docs/project.svg](../../docs/project.svg) · request cycle [docs/request-cycle.svg](../../docs/request-cycle.svg) · security architecture [docs/security.svg](../../docs/security.svg) · project structure [docs/structure.svg](../../docs/structure.svg).

## Project Panorama

<p align="center"><img src="images/en/project.svg" alt="Project panorama diagram" width="860"/></p>

## Request Cycle

<p align="center"><img src="images/en/request-cycle.svg" alt="Request cycle diagram" width="860"/></p>

## Security Architecture

<p align="center"><img src="images/en/security.svg" alt="Security architecture diagram" width="860"/></p>

---

## Directory Structure

```
open-novel/
├─ apps/                     # Multi-client frontends
│  ├─ flutter/               #   Flutter all-platform (Web / Desktop / Mobile), i18n multilingual
│  └─ harmonyos/             #   HarmonyOS NEXT native app (ArkTS / ArkUI)
├─ kratos/                   # Go-Kratos framework source (upstream framework, keep as-is, do not modify)
│  └─ backend/               #   Business backend of this project: cmd/server entry + api/ + internal/ + sql/ + opensearch/
├─ docs/                     # Project documentation (planning, architecture diagrams, i18n READMEs, reward QR codes)
├─ scripts/                  # Build and deployment scripts (post-push.sh auto release, smoke.sh)
├─ docker-compose.yml        # Local dependency stack: MySQL 8 + Redis 7 + OpenSearch 2
├─ CLAUDE.md                 # Project collaboration conventions
└─ README.md                 # Project documentation
```

<p align="center"><img src="images/en/structure.svg" alt="Project structure diagram" width="860"/></p>

> Note: `kratos/` is the Kratos framework source (with its own README / LICENSE); all business code lives in `kratos/backend/`.

## Tech Stack

| Layer | Technology |
| :--- | :--- |
| Client | Flutter (Web / Desktop / Mobile), HarmonyOS NEXT (ArkTS / ArkUI) |
| Gateway | Nginx + multi-vendor CDN (Cloudflare / CloudFront / Aliyun / Tencent Cloud), Go-Kratos API Gateway (gRPC / HTTP dual protocol) |
| Server | Go 1.22+, Kratos v2, protobuf / gRPC |
| Storage | MySQL 8.0 (master-slave), Redis 7.x (Cluster), OpenSearch 2.x, ristretto in-process L1 cache on top of Redis (30 s TTL) |
| Observability | Prometheus, Grafana, ELK, OpenTelemetry tracing |
| Operations | Docker Compose, GitHub Actions CI/CD |

## Database

- Database name: `novel`
- Table prefix: `novel_` (e.g. `novel_user`, `novel_book`, `novel_chapter`, `novel_comment`, etc.)

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Table creation script: `kratos/backend/sql/init.sql` (automatically executed on the first Docker Compose startup). For detailed table design and the read/write separation strategy, see [docs/novel-project-planning.md](../novel-project-planning.md).

## API Prefix

Backend HTTP APIs uniformly start with `/api`; the version is negotiated via the `X-Api-Version: v1` request header (not in the URL). Endpoints are grouped by domain:

| Domain | Example routes | proto definition |
| :--- | :--- | :--- |
| User | `/api/users`, etc. | `kratos/backend/api/user/v1` |
| Book | `/api/books`, `/api/books/{id}`, `/api/categories`, `/api/tags` | `kratos/backend/api/book/v1` |
| Chapter | `/api/...` | `kratos/backend/api/chapter/v1` |
| Comment | `/api/...` | `kratos/backend/api/comment/v1` |
| Search | `/api/...` | `kratos/backend/api/search/v1` |
| Recommendation | `/api/...` | `kratos/backend/api/recommendation/v1` |

For detailed routes, see the `option (google.api.http)` declarations in each proto file.

## Quick Start

```bash
# 1. Start the dependency stack (MySQL / Redis / OpenSearch; the first startup automatically executes kratos/backend/sql/init.sql to create tables)
docker compose up -d

# 2. Start the backend service (Kratos business directory, HTTP :8000 / gRPC :9000)
cd kratos/backend && go mod tidy && go run ./cmd/server

# 3. Start the Flutter client (defaults to localhost:8000, no extra configuration needed)
cd apps/client/flutter && flutter pub get && flutter run -d chrome
```

- Dependency stack port mapping: MySQL `3307`, Redis `6380`, OpenSearch `9200` (host ports 3306/6379 are occupied by local services, see docker-compose.yml comments).
- Backend address and secrets are configured in `kratos/backend/config/`, with environment variable overrides (e.g. `PORT`, `OPENSEARCH_ADDR`).
- To connect Flutter to another backend: `flutter run -d chrome --dart-define=API_BASE_URL=http://<host>:8000`.

See [apps/README.md](../../apps/README.md) and [apps/client/flutter/README.md](../../apps/client/flutter/README.md).

## One-Click Installation

```bash
bash scripts/install.sh
```

A single command completes the environment check, dependency stack startup and launch hints: it checks whether Docker / Go ≥ 1.22 / Flutter are installed (printing installation hints if anything is missing), runs `docker compose up -d` to start the dependency stack, waits for MySQL to become ready (up to 60 seconds), then prints the startup commands and access addresses for the backend and all three frontends. The script is idempotent and safe to re-run; `bash scripts/install.sh --skip-deps` skips the dependency stack step.

## Installation

- **Prerequisites**: Docker (with the Compose plugin), Go 1.22+, Flutter 3.x
- **One-click**: run `bash scripts/install.sh` and follow the printed hints for the environment check and dependency stack startup
- **Manual** (equivalent to the script steps):

  ```bash
  docker compose up -d                                     # dependency stack: MySQL / Redis / OpenSearch (init.sql runs on first start)
  cd kratos/backend && go mod tidy && go run ./cmd/server  # backend HTTP :8000 / gRPC :9000
  cd apps/client/flutter && flutter pub get && flutter run -d chrome  # client
  ```

## Usage

- **Backend**: HTTP `http://localhost:8000` (gRPC `:9000`), all APIs under the `/api` prefix, version negotiated via the `X-Api-Version: v1` header
- **Client**: `cd apps/client/flutter && flutter pub get && flutter run -d chrome` (defaults to localhost:8000)
- **Admin console**: `cd apps/admin && flutter pub get && flutter run -d chrome`; Flutter assigns a random port and prints it in the console (fix it with `--web-port`)
- **Dependency stack ports**: MySQL `3307`, Redis `6380`, OpenSearch `9200`
- **Default config**: `kratos/backend/config/` (database connection, secrets, ports), overridable via environment variables
- **FAQ**:
  - Port already in use: locate the process with `lsof -i :8000`, or change the port in `kratos/backend/config/` and restart the backend
  - Database connection failure: run `docker compose ps` and make sure mysql is healthy; the first startup needs to wait for `init.sql` to create the tables

## Release Process

- **Automatic**: after pushing `main`, GitHub Actions ([.github/workflows/release.yml](../../.github/workflows/release.yml)) automatically bumps the patch version based on the latest `v*` tag, creates and pushes the tag, then creates a GitHub Release with an incremental changelog; skipped if HEAD already carries a version tag. The first release starts from `v1.0.0`.
- **Manual fallback**: run [scripts/post-push.sh](../../scripts/post-push.sh) (requires authenticated `gh`): `echo "x y refs/heads/main z" | scripts/post-push.sh`.
- **Manual**:

  ```bash
  git tag -a v1.0.1 -m "release v1.0.1"
  git push origin v1.0.1
  gh release create v1.0.1 --generate-notes
  ```

## Roadmap

| Phase | Duration | Focus |
| :--- | :--- | :--- |
| Phase 1 | 2-3 weeks | Kratos backend base services + MySQL / Redis / OpenSearch integration |
| Phase 2 | 3-4 weeks | Flutter / HarmonyOS multi-client frontends + multilingual ARB authoring |
| Phase 3 | 2 weeks | Security hardening (JWT / RBAC / rate limiting) + stress testing |
| Phase 4 | 1-2 weeks | Full-link integration testing + CDN acceleration configuration |
| Phase 5 | Ongoing | AI recommendation algorithms, user behavior analytics tracking |

All task chains have been completed.

## Support and Donations

If this project has helped you, feel free to **Star** and **Fork** it; you are also welcome to scan the QR codes to donate. Every bit of your support is my motivation to keep maintaining and updating the project. Thank you for your encouragement!

<div align="center">

**WeChat Reward** ｜ **Alipay Reward**

<img src="../weixinpay.png" width="130" height="130" alt="WeChat reward QR code" />　<img src="../alipay.png" width="130" height="130" alt="Alipay reward QR code" />

</div>

### Global Transfer Donations (Cross-border Remittance)

【Payee Information】

- Payee Name: WANG KEXUN
- Payee Account Number: 881015918251

【Payee Bank】

- ZA Bank SWIFT Code: AABLHKHHXXX
- Bank Name: ZA Bank Limited
- Bank Code: 387
- Bank Address: Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【Cross-border Remittance Intermediary Bank (If Required)】

> Please note that this is the information of the cross-border remittance intermediary bank (correspondent bank), not the payee's bank. Please check with your remitting bank whether the cross-border remittance intermediary bank information is required.

**The intermediary bank for HKD, CNY and USD remittances is Citibank**

- Bank Name: Citibank N.A. Hong Kong
- SWIFT Code: CITIHKHXXXX
- Bank Code: 006
- Branch Name: Hong Kong Branch
- Branch Code: 391
- Bank Address: Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**The intermediary bank for other currencies is BNY Mellon**

- Bank Name: THE BANK OF NEW YORK MELLON
- SWIFT Code: IRVTUS3NXXX
- Bank Address: THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States

### Crypto Donation

If this project helps you, scan the QR code to donate, thank you!

| Network | QR Code | Wallet Address |
|---|---|---|
| BNB Smart Chain (BEP20) | [<img src="../coin/1.jpg" width="150" alt="BNB Smart Chain (BEP20)">](../coin/1.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Tron (TRC20) | [<img src="../coin/2.jpg" width="150" alt="Tron (TRC20)">](../coin/2.jpg) | `TEdDHWLajt1XvqtPDWmQctdrJaC3pzZZzz` |
| Ethereum (ERC20) | [<img src="../coin/3.jpg" width="150" alt="Ethereum (ERC20)">](../coin/3.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Aptos | [<img src="../coin/4.jpg" width="150" alt="Aptos">](../coin/4.jpg) | `0x836e3780edfc3f7b2372b39e2a1a3a5d7adfaccd96c726f21cfde1b50dd68030` |
| Plasma | [<img src="../coin/5.jpg" width="150" alt="Plasma">](../coin/5.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Polygon POS | [<img src="../coin/6.jpg" width="150" alt="Polygon POS">](../coin/6.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Solana | [<img src="../coin/7.jpg" width="150" alt="Solana">](../coin/7.jpg) | `2hfhboHdmdrYsY25XfQSsEWxq5ip4EQsR7f4AzSRMUyr` |
| The Open Network (TON) | [<img src="../coin/8.jpg" width="150" alt="The Open Network (TON)">](../coin/8.jpg) | `UQB9kFQohzmXUir9QSSZq01iwl9aQZIDdBpNmDklljRtCoGK` |
| Arbitrum One | [<img src="../coin/9.jpg" width="150" alt="Arbitrum One">](../coin/9.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| AVAX C-Chain | [<img src="../coin/10.jpg" width="150" alt="AVAX C-Chain">](../coin/10.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |

---

## License and Contact

- **License**: there is no standalone LICENSE at the repository root; `kratos/` is the upstream Kratos framework source, governed by its [MIT License](../../kratos/LICENSE). The licensing of the business code is subject to subsequent project announcements.
- **Contact**: communicate via GitHub Issues / PRs; donations see "Support and Donations" above.
