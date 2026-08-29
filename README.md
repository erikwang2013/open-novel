# Open Novel — 全球多语言小说平台

<div align="center">

**中文** · [English](docs/i18n/README.en.md) · [日本語](docs/i18n/README.ja.md) · [한국어](docs/i18n/README.ko.md) · [Русский](docs/i18n/README.ru.md) · [Deutsch](docs/i18n/README.de.md) · [Français](docs/i18n/README.fr.md) · [Español](docs/i18n/README.es.md) · [Português](docs/i18n/README.pt.md) · [हिन्दी](docs/i18n/README.hi.md) · [العربية](docs/i18n/README.ar.md) · [বাংলা](docs/i18n/README.bn.md) · [Bahasa Indonesia](docs/i18n/README.id.md)

</div>

> 基于 **Go-Kratos** 微服务架构 + **Flutter / HarmonyOS** 多端前端的全球多语言小说阅读平台，支持 **12+ 种主要语种**，面向全球用户提供阅读、互动、搜索与个性化推荐能力。

---

## 项目简介

Open Novel 是一个云原生微服务架构的全球多语言小说平台：

- **后端**：Go-Kratos v2（gRPC / HTTP 双协议），微服务按领域拆分（用户、书籍、章节、评论、搜索、推荐）
- **前端**：Flutter 全平台（Web / Desktop / Mobile）+ HarmonyOS NEXT 原生应用，共用同一套后端 API
- **多语言**：i18n 资源动态加载，支持 12+ 语种（中文、英文、日文、韩文、法文、德文、西班牙文、俄文、阿拉伯文等）
- **存储**：MySQL 8（主从读写分离）+ Redis（热点缓存 / 会话）+ ristretto（进程内 L1 二级缓存）+ OpenSearch（多语言搜索）
- **运维**：Docker Compose 一键部署，Prometheus + Grafana 监控，GitHub Actions 持续集成

## 功能特性

<p align="center"><img src="docs/features.svg" alt="功能架构图" width="860"/></p>

- **用户中心**：注册登录（JWT）、个人书架、阅读进度跨端同步、多语言个人资料
- **阅读体验**：分章阅读、字体字号切换、深浅主题、离线缓存、翻页动画
- **书籍内容**：书籍元数据、章节管理、分类标签、连载更新、多语言翻译
- **互动社区**：评论书评、点赞、收藏、举报审核
- **搜索发现**：多语言分词搜索、热搜词榜（/api/search/hot-keywords）、搜索建议（/api/search/suggest，本地历史 20 条可清空 + 200ms 防抖）、热门榜单、AI 推荐、分类浏览
- **管理后台**：内容审核、用户管理、数据统计（仪表盘 / DAU / 榜单）、配置管理（分类标签）、审计日志查询（/api/admin/audit-logs）
- **支付与会员**：11 个支付渠道（国际卡 Stripe / PayPal + USDT NOWPayments + 本地 Razorpay(hi) / KOMOJU(ja) / PortOne(ko) / Mercado Pago(pt-BR) / Xendit(id/th/vn) / Alipay(zh-CN) / WeChat Pay Global(国际版) / UnionPay(zh-CN)）、VIP 套餐订阅与续期、支付方式多语言路由

## 系统架构

<p align="center"><img src="docs/architecture.svg" alt="系统架构图" width="860"/></p>

整体为 Go-Kratos 微服务架构：Flutter / HarmonyOS 客户端经 Nginx + CDN 与 API 网关交互，网关按领域路由到用户、书籍、章节、评论、搜索、推荐等后端服务；数据层为 MySQL 主从（读写分离）+ Redis 缓存（其上叠加 ristretto 进程内 L1 二级缓存）+ OpenSearch 搜索索引。服务间 gRPC 通信，对外 HTTP 接口统一前缀 `/api`。

其余设计图：项目全景 [docs/project.svg](docs/project.svg) · 请求周期 [docs/request-cycle.svg](docs/request-cycle.svg) · 安全架构 [docs/security.svg](docs/security.svg) · 项目结构 [docs/structure.svg](docs/structure.svg)。

项目规划：架构设计 [docs/novel-project-planning.md](docs/novel-project-planning.md) · 现状盘点与阶段路线 [docs/roadmap.md](docs/roadmap.md) · 详细任务分解 [docs/tasks.md](docs/tasks.md)。

## 目录结构

```
open-novel/
├─ apps/                     # 多端前端
│  ├─ client/                #   客户端（C 端）
│  │  ├─ flutter/            #     Flutter 全平台（Web / Desktop / Mobile），i18n 多语言
│  │  └─ harmonyos/          #     HarmonyOS NEXT 原生应用（ArkTS / ArkUI）
│  └─ admin/                 #   管理端（Flutter Web，B 端后台）
├─ .github/                  # CI/CD：GitHub Actions 发布工作流（release.yml）
├─ kratos/                   # Go-Kratos 框架源码（上游框架，原样保留，勿改）
│  └─ backend/               #   本项目业务后端：cmd/server 入口 + api/ + internal/ + sql/ + opensearch/
├─ docs/                     # 项目文档（规划、架构图、i18n README、打赏码）
├─ scripts/                  # 构建与部署脚本（post-push.sh 自动发布、smoke.sh）
├─ docker-compose.yml        # 本地依赖栈：MySQL 8 + Redis 7 + OpenSearch 2
├─ CLAUDE.md                 # 项目协作规范
└─ README.md                 # 项目说明文档
```

<p align="center"><img src="docs/structure.svg" alt="项目结构图" width="860"/></p>

> 注意：`kratos/` 是 Kratos 框架源码（自带 README / LICENSE），业务代码全部在 `kratos/backend/`。

## 技术栈

| 层次 | 技术选型 |
| :--- | :--- |
| 客户端 | Flutter（Web / Desktop / Mobile，url_launcher / flutter_cache_manager）、HarmonyOS NEXT（ArkTS / ArkUI） |
| 管理端 | Flutter Web（`apps/admin/`，B 端后台） |
| 网关 | Nginx + CDN、Go-Kratos API 网关（gRPC / HTTP 双协议） |
| 服务端 | Go 1.22+、Kratos v2、protobuf / gRPC、stripe-go（Stripe 支付 SDK） |
| 存储 | MySQL 8.0（主从）、Redis 7.x（Cluster）、ristretto（进程内 L1 二级缓存，128MB / 30s TTL）、OpenSearch 2.x |
| 可观测 | Prometheus、Grafana、ELK、OpenTelemetry 链路追踪 |
| 运维 | Docker Compose、GitHub Actions CI/CD |

## 数据库

- 数据库名：`novel`
- 表前缀：`novel_`（如 `novel_user`、`novel_book`、`novel_chapter`、`novel_comment`、`novel_payment_order`、`novel_payment_provider`、`novel_vip_order`、`novel_vip_plan`、`novel_audit_log` 等）

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

建表脚本：`kratos/backend/sql/init.sql`（Docker Compose 首次启动自动执行）。详细表设计与读写分离策略见 [docs/novel-project-planning.md](docs/novel-project-planning.md)。

## API 文档

接口统一前缀 `/api`，版本经请求头 `X-Api-Version: v1` 协商（不写在 URL 中）。完整端点、参数、错误码与限流说明见 **[docs/api.md](docs/api.md)**。按领域分组：

| 领域 | 示例路由 | proto 定义 |
| :--- | :--- | :--- |
| 用户 | `/api/users` 等 | `kratos/backend/api/user/v1` |
| 书籍 | `/api/books`、`/api/books/{id}`、`/api/categories`、`/api/tags` | `kratos/backend/api/book/v1` |
| 章节 | `/api/...` | `kratos/backend/api/chapter/v1` |
| 评论 | `/api/...` | `kratos/backend/api/comment/v1` |
| 搜索 | `/api/search/hot-keywords`（热搜词榜）、`/api/search/suggest`（搜索建议）、`/api/search/hot` 等 | `kratos/backend/api/search/v1` |
| 推荐 | `/api/...` | `kratos/backend/api/recommendation/v1` |
| 支付 | `/api/payments/...` | `kratos/backend/api/payment/v1` |
| 管理 | `/api/stats/overview`、`/api/categories`、`/api/tags`、`/api/admin/audit-logs`（审计日志，分页 + 多条件筛选） | `kratos/backend/api/admin/v1` |

详细路由见各 proto 文件 `option (google.api.http)` 声明。

## 快速开始

```bash
# 1. 启动依赖栈（MySQL / Redis / OpenSearch，首次启动自动执行 kratos/backend/sql/init.sql 建表）
docker compose up -d

# 2. 启动后端服务（Kratos 业务目录，HTTP :8000 / gRPC :9000）
cd kratos/backend && go mod tidy && go run ./cmd/server

# 3. 启动 Flutter 端（默认连 localhost:8000，无需额外配置）
cd apps/client/flutter && flutter pub get && flutter run -d chrome
```

- 依赖栈端口映射：MySQL `3307`、Redis `6380`、OpenSearch `9200`（宿主机 3306/6379 被本机服务占用，见 docker-compose.yml 注释）。
- 后端地址与密钥在 `kratos/backend/config/` 配置，支持环境变量覆盖（如 `PORT`、`OPENSEARCH_ADDR`）。
- Flutter 连其他后端：`flutter run -d chrome --dart-define=API_BASE_URL=http://<host>:8000`。

详见 [apps/README.md](apps/README.md) 与 [apps/client/flutter/README.md](apps/client/flutter/README.md)。

## 发布流程

- **自动**：推送 `main` 后 GitHub Actions（[.github/workflows/release.yml](.github/workflows/release.yml)）自动基于最新 `v*` tag 递增 patch 版本，创建 tag 并推送，再以增量 changelog 创建 GitHub Release；HEAD 已带版本 tag 时跳过。首次发布从 `v1.0.0` 起。
- **手动兜底**：运行 [scripts/post-push.sh](scripts/post-push.sh)（需 `gh` 已认证）：`echo "x y refs/heads/main z" | scripts/post-push.sh`。
- **手动**：

  ```bash
  git tag -a v1.0.1 -m "release v1.0.1"
  git push origin v1.0.1
  gh release create v1.0.1 --generate-notes
  ```

## 路线图

| 阶段 | 周期 | 任务重点 | 状态 |
| :--- | :--- | :--- | :--- |
| Phase 1 | 2-3 周 | Kratos 后端基础服务 + MySQL / Redis / OpenSearch 集成 | ✅ 已完成 |
| Phase 2 | 3-4 周 | Flutter / HarmonyOS 多端前端 + 多语言 ARB 编写 | ✅ 已完成 |
| Phase 3 | 2 周 | 安全加固（JWT / RBAC / 限流）+ 压力测试 | ✅ 已完成 |
| Phase 4 | 1-2 周 | 全链路联调 + CDN 加速配置 | ✅ 已完成 |
| Phase 5 | 持续 | AI 推荐算法接入、用户行为分析埋点 | ⏳ 进行中 |
| 商业化 | 2026-08 | 管理后台（审核/用户/统计/配置/审计日志）、多语言与阅读体验、VIP 与支付链（11 渠道，支付宝 / 微信支付国际版 / 银联 2026-08-29 接入） | ✅ 已完成（T-A-01~17 / T-C-01~23 / T-P-01~22） |

## 支持与打赏

如果这个项目对你有帮助，欢迎 **Star**、**Fork** 支持；也欢迎扫码打赏，你的每一份支持都是我持续维护与更新的动力，感谢你的鼓励！

<div align="center">

**微信赞赏** ｜ **支付宝赞赏**

<img src="docs/weixinpay.png" width="130" height="130" alt="微信赞赏码" />　<img src="docs/alipay.png" width="130" height="130" alt="支付宝赞赏码" />

</div>

### 全球转账打赏（跨境汇款）

【收款人信息】

- 收款人姓名：WANG KEXUN
- 收款账户号码：881015918251

【收款银行】

- ZA Bank SWIFT Code：AABLHKHHXXX
- 银行名称：ZA Bank Limited
- 银行编号：387
- 银行地址：Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【跨境汇款代理银行（如需）】

> 请留意，此为跨境汇款代理银行（中转银行）信息，非收款银行信息。请向汇款银行查询是否需要提供跨境汇款代理银行信息。

**汇入港元、人民币及美元的代理银行为 Citibank**

- 银行名称：Citibank N.A. Hong Kong
- SWIFT Code：CITIHKHXXXX
- 银行编号：006
- 分行名称：Hong Kong Branch
- 分行编号：391
- 银行地址：Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**汇入其他币种时的代理银行为 BNY Mellon**

- 银行名称：THE BANK OF NEW YORK MELLON
- SWIFT Code：IRVTUS3NXXX
- 银行地址：THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States

### 虚拟币打赏 (Crypto Donation)

如果这个项目对你有帮助，欢迎扫描二维码打赏支持，谢谢！

| 主网 (Network) | 二维码 (QR Code) | 钱包地址 (Wallet Address) |
|---|---|---|
| BNB Smart Chain (BEP20) | [<img src="./docs/coin/1.jpg" width="150" alt="BNB Smart Chain (BEP20)">](./docs/coin/1.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Tron (TRC20) | [<img src="./docs/coin/2.jpg" width="150" alt="Tron (TRC20)">](./docs/coin/2.jpg) | `TEdDHWLajt1XvqtPDWmQctdrJaC3pzZZzz` |
| Ethereum (ERC20) | [<img src="./docs/coin/3.jpg" width="150" alt="Ethereum (ERC20)">](./docs/coin/3.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Aptos | [<img src="./docs/coin/4.jpg" width="150" alt="Aptos">](./docs/coin/4.jpg) | `0x836e3780edfc3f7b2372b39e2a1a3a5d7adfaccd96c726f21cfde1b50dd68030` |
| Plasma | [<img src="./docs/coin/5.jpg" width="150" alt="Plasma">](./docs/coin/5.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Polygon POS | [<img src="./docs/coin/6.jpg" width="150" alt="Polygon POS">](./docs/coin/6.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Solana | [<img src="./docs/coin/7.jpg" width="150" alt="Solana">](./docs/coin/7.jpg) | `2hfhboHdmdrYsY25XfQSsEWxq5ip4EQsR7f4AzSRMUyr` |
| The Open Network (TON) | [<img src="./docs/coin/8.jpg" width="150" alt="The Open Network (TON)">](./docs/coin/8.jpg) | `UQB9kFQohzmXUir9QSSZq01iwl9aQZIDdBpNmDklljRtCoGK` |
| Arbitrum One | [<img src="./docs/coin/9.jpg" width="150" alt="Arbitrum One">](./docs/coin/9.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| AVAX C-Chain | [<img src="./docs/coin/10.jpg" width="150" alt="AVAX C-Chain">](./docs/coin/10.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |

---

## License 与联系方式

- **License**：仓库根目录无独立 LICENSE；`kratos/` 为 Kratos 框架上游源码，遵循其 [MIT License](kratos/LICENSE)。业务代码授权方式以项目后续声明为准。
- **联系方式**：GitHub Issues / PR 交流；捐赠见上方「支持与打赏」。
