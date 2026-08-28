# Open Novel — Global Multilingual Novel Platform

<div align="center">

[中文](../../README.md) · **English** · [日本語](README.ja.md) · [한국어](README.ko.md) · [Русский](README.ru.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [বাংলা](README.bn.md) · [Bahasa Indonesia](README.id.md)

</div>

> A global multilingual novel reading platform built on the **Go-Kratos** microservice architecture with **Flutter / HarmonyOS** multi-client frontends, supporting **12+ major languages** and delivering reading, interaction, search and personalized recommendation capabilities to users worldwide.

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
- **Search & Discovery**: Multilingual tokenized search, trending rankings, AI recommendations, category browsing
- **Admin Console**: Content moderation, user management, data statistics, configuration management

## System Architecture

<p align="center"><img src="images/en/architecture.svg" alt="System architecture diagram" width="860"/></p>

Overall Go-Kratos microservice architecture: Flutter / HarmonyOS clients interact with the API gateway via Nginx + CDN; the gateway routes by domain to backend services such as users, books, chapters, comments, search and recommendations; the data layer is MySQL master-slave (read/write separation) + Redis cache + OpenSearch search index. Services communicate via gRPC, and external HTTP APIs uniformly use the `/api/v1` prefix.

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
| Gateway | Nginx + CDN, Go-Kratos API Gateway (gRPC / HTTP dual protocol) |
| Server | Go 1.22+, Kratos v2, protobuf / gRPC |
| Storage | MySQL 8.0 (master-slave), Redis 7.x (Cluster), OpenSearch 2.x |
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

Backend HTTP APIs uniformly start with `/api/v1`, grouped by domain:

| Domain | Example routes | proto definition |
| :--- | :--- | :--- |
| User | `/api/v1/users`, etc. | `kratos/backend/api/user/v1` |
| Book | `/api/v1/books`, `/api/v1/books/{id}`, `/api/v1/categories`, `/api/v1/tags` | `kratos/backend/api/book/v1` |
| Chapter | `/api/v1/...` | `kratos/backend/api/chapter/v1` |
| Comment | `/api/v1/...` | `kratos/backend/api/comment/v1` |
| Search | `/api/v1/...` | `kratos/backend/api/search/v1` |
| Recommendation | `/api/v1/...` | `kratos/backend/api/recommendation/v1` |

For detailed routes, see the `option (google.api.http)` declarations in each proto file.

## Quick Start

```bash
# 1. Start the dependency stack (MySQL / Redis / OpenSearch; the first startup automatically executes kratos/backend/sql/init.sql to create tables)
docker compose up -d

# 2. Start the backend service (Kratos business directory, HTTP :8000 / gRPC :9000)
cd kratos/backend && go mod tidy && go run ./cmd/server

# 3. Start the Flutter client (defaults to localhost:8000, no extra configuration needed)
cd apps/flutter && flutter pub get && flutter run -d chrome
```

- Dependency stack port mapping: MySQL `3307`, Redis `6380`, OpenSearch `9200` (host ports 3306/6379 are occupied by local services, see docker-compose.yml comments).
- Backend address and secrets are configured in `kratos/backend/config/`, with environment variable overrides (e.g. `PORT`, `OPENSEARCH_ADDR`).
- To connect Flutter to another backend: `flutter run -d chrome --dart-define=API_BASE_URL=http://<host>:8000`.

See [apps/README.md](../../apps/README.md) and [apps/flutter/README.md](../../apps/flutter/README.md).

## Release Process

- **Automatic**: after pushing `main`, run [scripts/post-push.sh](../../scripts/post-push.sh) (either as a git push hook or manually). The script bumps the patch version based on the latest `v*` tag, creates and pushes the tag, then creates a GitHub Release with an incremental changelog; `gh` must be authenticated. The first release starts from `v1.0.0`.
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

---

## License and Contact

- **License**: there is no standalone LICENSE at the repository root; `kratos/` is the upstream Kratos framework source, governed by its [MIT License](../../kratos/LICENSE). The licensing of the business code is subject to subsequent project announcements.
- **Contact**: communicate via GitHub Issues / PRs; donations see "Support and Donations" above.
