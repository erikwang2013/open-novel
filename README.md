# Open Novel — 全球多语言小说平台

<div align="center">

**中文** · [English](docs/i18n/README.en.md) · [日本語](docs/i18n/README.ja.md) · [한국어](docs/i18n/README.ko.md) · [Русский](docs/i18n/README.ru.md) · [Deutsch](docs/i18n/README.de.md) · [Français](docs/i18n/README.fr.md) · [Español](docs/i18n/README.es.md) · [Português](docs/i18n/README.pt.md) · [हिन्दी](docs/i18n/README.hi.md) · [العربية](docs/i18n/README.ar.md) · [বাংলা](docs/i18n/README.bn.md) · [Bahasa Indonesia](docs/i18n/README.id.md)

</div>

> 基于 **Go-Kratos** 微服务架构 + **Flutter / HarmonyOS** 多端前端的全球多语言小说阅读平台，支持 **12+ 种主要语种**，面向全球用户提供阅读、互动、搜索与个性化推荐能力。

<div align="center"><img src="docs/pet.svg" alt="Open Novel 项目宠物 Novi" width="150"/></div>

---

## 项目简介

Open Novel 是一个云原生微服务架构的全球多语言小说平台：

- **后端**：Go-Kratos v2（gRPC / HTTP 双协议），微服务按领域拆分（用户、书籍、章节、评论、搜索、推荐）
- **前端**：Flutter 全平台（Web / Desktop / Mobile）+ HarmonyOS NEXT 原生应用，共用同一套后端 API
- **多语言**：i18n 资源动态加载，支持 12+ 语种（中文、英文、日文、韩文、法文、德文、西班牙文、俄文、阿拉伯文等）
- **存储**：MySQL 8（主从读写分离）+ Redis（热点缓存 / 会话）+ ristretto（进程内 L1 二级缓存）+ OpenSearch（多语言搜索）
- **运维**：Docker Compose 一键部署，Prometheus + Grafana 监控，GitHub Actions 持续集成
- **项目宠物**：Novi（[docs/pet.svg](docs/pet.svg)，会眨眼 / 甩尾凝字 / 墨身挤压）—— 终端 `bash scripts/pet.sh` 看仓库心情，后端 `GET /api/pet.svg` 按依赖健康度换表情，客户端与管理端空状态、两端网页外壳（标签页图标 / PWA 图标）也用它

## 功能特性

<p align="center"><img src="docs/features.svg" alt="功能架构图" width="860"/></p>

- **用户中心**：注册登录（JWT）、个人书架、阅读进度跨端同步、多语言个人资料
- **阅读体验**：分章阅读、字体字号切换、深浅主题、离线缓存、翻页动画
- **书籍内容**：书籍元数据、章节管理、分类标签、连载更新、多语言翻译
- **互动社区**：评论书评、点赞、收藏、举报审核
- **搜索发现**：多语言分词搜索、热搜词榜（/api/search/hot-keywords）、搜索建议（/api/search/suggest，本地历史 20 条可清空 + 200ms 防抖）、热门榜单、AI 推荐、分类浏览
- **管理后台**：内容审核、用户管理、数据统计（仪表盘 / DAU / 榜单 / 行为分析 /api/stats/behavior）、配置管理（分类标签）、机器翻译工作流（DeepL，/api/admin/translate/*，管理端「翻译」页 + 人工编辑）、审计日志查询（/api/admin/audit-logs）、CDN 厂商管理（多厂商配置 / 启停 / 排序，指纹热更新即时生效）
- **支付与会员**：11 个支付渠道（国际卡 Stripe / PayPal + USDT NOWPayments + 本地 Razorpay(hi) / KOMOJU(ja) / PortOne(ko) / Mercado Pago(pt-BR) / Xendit(id/th/vn) / Alipay(zh-CN) / WeChat Pay Global(国际版) / UnionPay(zh-CN)）、VIP 套餐订阅与续期、支付方式多语言路由

## 系统架构

<p align="center"><img src="docs/architecture.svg" alt="系统架构图" width="860"/></p>

整体为 Go-Kratos 微服务架构：Flutter / HarmonyOS 客户端经 Nginx + 多厂商 CDN（Cloudflare / CloudFront 全球线，阿里云 / 腾讯云中国线；管理端可配置、指纹热更新即时生效）与 API 网关交互，网关按领域路由到用户、书籍、章节、评论、搜索、推荐等后端服务；数据层为 MySQL 主从（读写分离）+ Redis 缓存（其上叠加 ristretto 进程内 L1 二级缓存）+ OpenSearch 搜索索引。服务间 gRPC 通信，对外 HTTP 接口统一前缀 `/api`。

其余设计图：项目全景 [docs/project.svg](docs/project.svg) · 请求周期 [docs/request-cycle.svg](docs/request-cycle.svg) · 安全架构 [docs/security.svg](docs/security.svg) · 项目结构 [docs/structure.svg](docs/structure.svg) · 项目生命周期 [docs/lifecycle.svg](docs/lifecycle.svg)。

## 项目生命周期

<p align="center"><img src="docs/lifecycle.svg" alt="项目生命周期图" width="860"/></p>

从**规划 → 架构设计 → 编码 → 测试 → 构建 → 发布 → 部署 → 监控 → 迭代**九个环节串成闭环：规划落地在 `planning / roadmap / tasks` 三件套；发布环节由 GitHub Actions 在 push `main` 后自动递增 `v*` tag 并以增量 changelog 创建 Release（幂等）；迭代环节把行为埋点、报表统计与搜索日志回流到需求池，进入下一轮规划。底部为已交付里程碑（后端基础 / 多端前端 / 安全加固 / CDN 全站加速 / 管理后台 / VIP 与支付）。

## 项目宠物 Novi

<p align="center"><img src="docs/pet.svg" alt="Open Novel 项目宠物 Novi" width="220"/></p>

**Novi** 是本项目的宠物（静止版为 [docs/mascot.svg](docs/mascot.svg)）：一滴活过来的墨，悬在摊开的书页上方——尾巴甩出的墨迹落到纸面，凝成 文 / A / あ 三个字，对应 12+ 语种与「阅读」两件事。源图是**会动**的 SVG（SMIL：呼吸浮动、墨身挤压、眨眼、尾墨摇摆、墨点飞溅），不需要 JS，README 与浏览器里直接就能看到它在动。

它有四个住处，其中三个能反映项目当前状态：

| 在哪见 | 怎么看 | 状态由什么决定 |
| :--- | :--- | :--- |
| 文档 / README | `docs/pet.svg`（12 语言 README 共用，与语言无关） | 固定形象 |
| 终端 | `bash scripts/pet.sh`（`--line` 只输出一行，可挂 pre-commit / CI 日志） | 工作区是否干净、领先 / 落后远端多少提交 |
| 后端 | `GET /api/pet.svg`（请求头 `X-Api-Version: v1`，见 `internal/pkg/pet.go`） | MySQL / Redis / OpenSearch 探测结果 → `happy` / `degraded` / `down` 三态表情 + 状态牌（含运行时长与不可用依赖） |
| 浏览器 / PWA | 客户端与管理端的 `web/favicon.svg`（矢量版，标签页里会眨眼）、`web/icons/Icon-*.png` | 固定形象（全部位图由 `bash scripts/gen-assets.sh` 从 `docs/pet.svg` 生成） |

客户端与管理端也把它接进了产品：客户端「暂无数据」与加载态（`lib/widgets/novi_pet.dart`）、管理端登录页（`apps/admin/assets/pet.png`）显示同一只 Novi，两端的网页外壳（标签页图标 / PWA 图标 / 应用名）也一并换成了它——`docs/pet-icon.svg` 是裁剪到脸部的图标版，`favicon.png` 由它生成。`GET /api/pet.svg` 是活的——依赖探测失败的次数越多，它的表情越困。


<!-- pet-expressions:start -->
## 项目宠物表情形态 · Novi

<table align="center">
<tr><td align="center"><img src="docs/pet-expressions/smile.svg" width="88" alt="笑"><br><b>笑</b><br><code>:emoji-smile:</code></td><td align="center"><img src="docs/pet-expressions/joy.svg" width="88" alt="欢乐"><br><b>欢乐</b><br><code>:emoji-joy:</code></td><td align="center"><img src="docs/pet-expressions/cry.svg" width="88" alt="哭"><br><b>哭</b><br><code>:emoji-cry:</code></td><td align="center"><img src="docs/pet-expressions/depressed.svg" width="88" alt="沮丧"><br><b>沮丧</b><br><code>:emoji-depressed:</code></td><td align="center"><img src="docs/pet-expressions/terrified.svg" width="88" alt="惊恐"><br><b>惊恐</b><br><code>:emoji-terrified:</code></td><td align="center"><img src="docs/pet-expressions/angry.svg" width="88" alt="愤怒"><br><b>愤怒</b><br><code>:emoji-angry:</code></td></tr>
<tr><td align="center"><img src="docs/pet-expressions/coquettish.svg" width="88" alt="撒娇"><br><b>撒娇</b><br><code>:emoji-coquettish:</code></td><td align="center"><img src="docs/pet-expressions/cute.svg" width="88" alt="卖萌"><br><b>卖萌</b><br><code>:emoji-cute:</code></td><td align="center"><img src="docs/pet-expressions/pleasing.svg" width="88" alt="讨好"><br><b>讨好</b><br><code>:emoji-pleasing:</code></td><td align="center"><img src="docs/pet-expressions/knocking.svg" width="88" alt="敲打"><br><b>敲打</b><br><code>:emoji-knocking:</code></td><td align="center"><img src="docs/pet-expressions/facepalm.svg" width="88" alt="捂脸"><br><b>捂脸</b><br><code>:emoji-facepalm:</code></td><td align="center"><img src="docs/pet-expressions/tongue.svg" width="88" alt="吐舌头"><br><b>吐舌头</b><br><code>:emoji-tongue:</code></td></tr>
</table>

> 12 种彩色动态形态：笑 / 欢乐 / 哭 / 沮丧 / 惊恐 / 愤怒 / 撒娇 / 卖萌 / 讨好 / 敲打 / 捂脸 / 吐舌头。每个 SVG 自带表情、情绪特效与整体动效，并遵循 `prefers-reduced-motion`。
<!-- pet-expressions:end -->

<!-- assets-license:start -->
## 视觉资产授权 / Visual Assets License

本项目的项目宠物、图片、图标、表情（包括 `docs/pet-expressions/` 等目录下的
SVG / PNG）**不是开源资产**，版权归作者所有。未经书面许可不得复制、修改、再分发或
用于商业用途；商业使用需取得商业授权，联系 **erik@erik.xyz**。详见
[ASSETS-LICENSE.md](ASSETS-LICENSE.md)。

> 代码部分仍遵循仓库原有 `LICENSE`；第三方素材（如明确标注的 Go Gopher 等）仍遵循其原许可证。
<!-- assets-license:end -->


## 项目规划

架构设计 [docs/novel-project-planning.md](docs/novel-project-planning.md) · 现状盘点与阶段路线 [docs/roadmap.md](docs/roadmap.md) · 详细任务分解 [docs/tasks.md](docs/tasks.md) · 支付渠道上线配置 [docs/merchant-config.md](docs/merchant-config.md)。

## 目录结构

```
open-novel/
├─ apps/                     # 多端前端
│  ├─ client/                #   客户端（C 端）
│  │  ├─ flutter/            #     Flutter 全平台（Web / Desktop / Mobile），i18n 多语言
│  │  │  └─ assets/pet.png   #       项目宠物贴图（源图 docs/pet.svg，空 / 加载态使用）
│  │  └─ harmonyos/          #     HarmonyOS NEXT 原生应用（ArkTS / ArkUI）
│  └─ admin/                 #   管理端（Flutter Web，B 端后台）
├─ .github/                  # CI/CD：GitHub Actions 发布工作流（release.yml）
├─ kratos/                   # Go-Kratos 框架源码（上游框架，原样保留，勿改）
│  └─ backend/               #   本项目业务后端：cmd/server 入口 + api/ + internal/ + sql/ + opensearch/
├─ docs/                     # 项目文档（规划、架构 / 生命周期 / 宠物等 SVG 图、i18n README、打赏码）
├─ scripts/                  # 构建与部署脚本（install.sh 一键安装、post-push.sh 自动发布、smoke.sh、pet.sh 终端宠物、gen-assets.sh 宠物图生成）
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
| 网关 | Nginx + 多厂商 CDN（Cloudflare / CloudFront / 阿里云 / 腾讯云）、Go-Kratos API 网关（gRPC / HTTP 双协议） |
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
| 项目宠物 | `/api/pet.svg`（活体状态图，非 proto 路由，需 `X-Api-Version: v1`） | `kratos/backend/internal/pkg/pet.go` |

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

## 一键安装

```bash
bash scripts/install.sh
```

一条命令完成环境检查 + 依赖栈启动 + 启动提示：检查 Docker / Go ≥ 1.22 / Flutter 是否安装（缺失时给出安装提示）、`docker compose up -d` 启动依赖栈、等待 MySQL 就绪（最多 60 秒）、最后打印后端与三个前端的启动命令和访问地址。可重复执行（幂等）；`bash scripts/install.sh --skip-deps` 跳过依赖栈步骤。

## 安装说明

- **前置依赖**：Docker（含 Compose 插件）、Go 1.22+、Flutter 3.x
- **一键方式**：运行 `bash scripts/install.sh`，按脚本提示完成环境检查与依赖栈启动
- **手动方式**（与脚本步骤等价，亦见上文「[快速开始](#快速开始)」）：

  ```bash
  docker compose up -d                                     # 依赖栈：MySQL / Redis / OpenSearch（首次启动自动执行 init.sql 建表）
  cd kratos/backend && go mod tidy && go run ./cmd/server  # 后端 HTTP :8000 / gRPC :9000
  cd apps/client/flutter && flutter pub get && flutter run -d chrome  # 客户端
  ```

## 使用说明

- **后端**：HTTP `http://localhost:8000`（gRPC `:9000`），接口统一前缀 `/api`，版本经请求头 `X-Api-Version: v1` 协商
- **项目宠物**：`curl -H 'X-Api-Version: v1' http://localhost:8000/api/pet.svg > pet.svg` 取活体状态图；终端看仓库心情用 `bash scripts/pet.sh`（`--line` 只输出一行）
- **客户端**：`cd apps/client/flutter && flutter pub get && flutter run -d chrome`（默认连 localhost:8000）
- **管理端**：`cd apps/admin && flutter pub get && flutter run -d chrome`，端口由 Flutter 随机分配并打印在控制台（可用 `--web-port` 固定）
- **依赖栈端口映射**：MySQL `3307`、Redis `6380`、OpenSearch `9200`
- **默认配置**：`kratos/backend/config/`（数据库连接、密钥、端口），支持环境变量覆盖
- **常见问题**：
  - 端口占用：`lsof -i :8000` 定位占用进程，或修改 `kratos/backend/config/` 中端口后重启后端
  - 数据库连接失败：`docker compose ps` 确认 mysql 为 healthy；首次启动需等待 `init.sql` 建表完成

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
