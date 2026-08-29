# Open Novel 项目规划（2026-08 修订）

> 全球多语言小说平台：**Go-Kratos** 后端 + **Flutter / HarmonyOS** 多端客户端 + **Flutter Web 管理端**，支持 12+ 语种。
> 本文档基于当前已实现功能与 [novel-project-planning.md](novel-project-planning.md)（架构设计）补充，聚焦「现状盘点 + 后续执行规划」。

---

## 一、项目定位

| 维度 | 定位 |
| :--- | :--- |
| 产品 | 全球多语言小说阅读平台，面向全球读者 |
| 客户端（C 端） | Flutter（Web / Desktop / Mobile）+ HarmonyOS NEXT，共用一套后端 API |
| 管理端（B 端） | Flutter Web（`apps/admin/`），运营/内容/用户/数据管理 |
| 后端 | Go-Kratos v2.9.2 微服务，gRPC 内部 + HTTP/JSON 对外（`/api` 前缀，`X-Api-Version: v1` 协商） |
| 存储 | MySQL 8（读写分离）+ Redis 7（热点缓存/会话）+ OpenSearch 2（多语言搜索） |
| 发布 | GitHub Actions 自动发版（push main → v* tag 递增 → Release） |

---

## 二、现状盘点（2026-08-29 已实现）

### 后端（`kratos/backend/`，8 个服务领域）

| 领域 | 已实现端点 | 状态 |
| :--- | :--- | :--- |
| 用户 | register / login / refresh / me / 列表 / 状态封禁 / 角色调整（管理） | ✅ |
| 书籍 | 列表 / 详情 / 创建 / 翻译更新 / 上下架（管理）/ categories / tags | ✅ |
| 章节 | 章节 CRUD / 正文 / 进度（段落级 position）/ 书架 / 禁用（管理） | ✅ |
| 评论 | 列表 / 发布 / 点赞 / 取消点赞 / 举报 / 举报列表与处理（管理）/ 收藏 | ✅ |
| 搜索 | 搜索 / 热搜词榜（/api/search/hot-keywords）/ 搜索建议（/api/search/suggest）/ 索引同步 / 索引删除 | ✅ |
| 推荐 | recommend（strategy=hot / new） | ✅ |
| 支付 | 下单 / 查单 / 公开套餐 / VIP 状态 / 方式列表 / Webhook（11 渠道：Stripe / PayPal / NOWPayments-USDT / Razorpay / KOMOJU / PortOne / Mercado Pago / Xendit / Alipay-RSA2 / WeChat Pay Global / UnionPay-RSA）/ 管理端流水·方式·套餐 | ✅ |
| 管理 | 仪表盘统计（GET /api/stats/overview）/ 分类标签 CRUD / 审计日志查询（GET /api/admin/audit-logs，分页 + user_id/action/target_type/target_id/时间范围筛选） | ✅ |

基础设施：JWT + Refresh 轮换（Redis GETDEL 防重放）、可选鉴权中间件 + requireAdmin（RBAC，role=3）、按路径限流（登录 10/min、评论发布 10/min、举报 5/min、搜索 10/min、支付回调 30/min）、业务错误码（11xxxx 用户 / 12xxxx 书籍 / 14xxxx 章节 / 15xxxx 评论 / 16xxxx 搜索 / 17xxxx 推荐 / 18xxxx 管理 / 19xxxx 支付，HTTP 200 + `{code,reason,message}`）、Redis 缓存（5min±30s 抖动 / 空值防穿透 / SETNX 单飞）+ ristretto 进程内 L1 二级缓存（128MB / 30s TTL / 写路径双删）、读写分离（FORCE_MASTER）、OpenSearch 多语言索引（zh/en/ja kuromoji / ko nori）、API 版本协商、支付渠道密钥 AES-GCM 加密存储、回调验签（Stripe / NOWPayments HMAC / Alipay RSA2 / WeChat Pay Global 平台公钥 + AES-GCM 解密）+ 金额强校验 + 幂等 settle。

### 客户端（`apps/client/`）

| 功能 | Flutter | HarmonyOS |
| :--- | :---: | :---: |
| 登录 / 注册 / 退出 | ✅ | ✅ |
| 首页推荐 + 语言切换（12+ 语种） | ✅ | ✅ |
| 书库列表 + 搜索 | ✅ | ✅（含分页加载更多） |
| 书籍详情（收藏 / 书架 / 评论入口） | ✅ | ✅ |
| 阅读器（正文 + 上/下一章 + 字号行距 13~26 / 1.2~2.6 + 深浅主题 + 上下滚动/左右翻页 + 离线缓存 + 段落级进度 + VIP 引导） | ✅ | ✅ |
| 评论区（列表 / 发布 / 点赞） | ✅ | ✅ |
| 我的（书架 / 收藏 / 最近进度 / VIP 卡片） | ✅ | ✅ |
| VIP 购买 / 支付结果页（套餐 + 按语言支付方式 + 3 秒轮询三态） | ✅ | ✅ |
| i18n | 13 ARB（zh-CN / en / ja / ko / fr / de / es / ru / pt / hi / ar / bn / id） | 13 语种资源 |

### 管理端（`apps/admin/`）

| 模块 | 状态 |
| :--- | :--- |
| 登录（role==3 管理员校验） | ✅ |
| 主框架（NavigationRail 10 tab：仪表盘 / 书籍 / 章节 / 评论 / 举报 / 用户 / 支付方式 / 流水 / 套餐 / 审计日志） | ✅ |
| 内容审核（书籍 / 章节 / 评论 / 举报 4 页） | ✅ |
| 用户管理（列表 / 封禁 / 角色） | ✅ |
| 仪表盘 + 分类标签配置 | ✅ |
| 支付方式 / 流水 / 套餐 3 页 | ✅ |
| 审计日志查询（分页 + user_id/action/target_type/target_id/时间范围筛选） | ✅ |
| token 持久化 | ✅（T-A-17） |
| 管理 API（后端） | ✅ admin 领域（stats / categories / tags / audit-logs） |

### CI / 脚本

- `.github/workflows/release.yml`：push main 自动递增 v* tag + 增量 changelog 发 Release（幂等）
- `scripts/post-push.sh`：本地手动兜底；`scripts/smoke.sh`：核心接口冒烟测试

---

## 三、差距分析（规划 vs 现状）

| 规划项（原规划文档） | 现状 | 优先级 |
| :--- | :--- | :--- |
| 12+ 语种 i18n | ✅ 完成：后端 lang 规范（zh-CN/en/ja/ko…13 语种）、OpenSearch ko nori 分词、Flutter 13 ARB、HarmonyOS 13 语种资源 | ✅ |
| 管理端 | ✅ 完成：审核 / 用户 / 统计 / 配置 / 支付管理 / 审计日志（10 tab），后端 admin 领域 + requireAdmin | ✅ |
| VIP / 支付链路 | ✅ 完成：10 支付渠道（Stripe / PayPal / NOWPayments-USDT / Razorpay / KOMOJU / PortOne / Mercado Pago / Xendit / Alipay，支付宝 RSA2 验签 / WeChat Pay Global，平台公钥 RSA-SHA256 验签）、下单/查单/Webhook、VIP 套餐订阅续期、管理端流水/方式/套餐（T-P-01~21）；国内微信支付未接入（需大陆商户资质） | ✅ |
| 字体字号 / 深浅主题 / 翻页动画 | ✅ 完成：字号 13~26 / 行距 1.2~2.6 持久化、Material3 深浅主题（跟随系统/浅/深）+ HarmonyOS dark、上下滚动/左右翻页双模式 | ✅ |
| 离线缓存（flutter_cache_manager） | ✅ 完成：Flutter flutter_cache_manager LRU；HarmonyOS 沙箱 50 篇 | ✅ |
| 阅读进度精确化 | ✅ 完成：进度精确到段落/滚动位置（复用后端 progress 端点 position 字段） | ✅ |
| 桌面端自适应布局 | ✅ 已实现（T-C-20/21） | ✅ |
| 分页策略统一 / token 持久化 | ✅ 已实现（T-C-16~19、T-A-17） | ✅ |
| 评论增强 / 分类浏览 / 热搜词榜 / 搜索建议 | ✅ 已实现（T-C-13~15） | ✅ |
| 本地支付 / PayPal / 微信国际版 | ✅ 已实现（T-P-19/20：Razorpay / KOMOJU / PortOne / Mercado Pago / Xendit / Alipay + PayPal；T-P-21：WeChat Pay Global 国际版）；国内微信支付未接入（需大陆商户资质） | ✅ |
| ristretto 本地二级缓存 | ✅ 已实现（128MB / 30s TTL / 写路径双删） | ✅ |

技术债：~~分类/标签路由遮蔽疑点~~ 已查实非问题（admin 用 /api/admin/* 独立路径，2026-08-29）。

---

## 四、管理端规划（B 端，M1~M5 均已交付 ✅）

### 阶段 M1：登录 + 框架（✅ 已完成）
登录（复用 `/api/users/login`，role==3 校验）→ NavigationRail 主框架。

### 阶段 M2：内容审核模块（✅ 已完成，T-A-01~11）
| 子功能 | 说明 | 后端配套 |
| :--- | :--- | :--- |
| 书籍管理 | 列表（分页/搜索）、状态上下架、编辑元数据 | 复用 GET/POST `/api/books` + PATCH `/api/books/{id}/status` |
| 章节管理 | 章节列表、正文查看、禁用/恢复 | PATCH `/api/chapters/{id}/status` |
| 评论审核 | 评论列表（按书籍/章节筛选）、下架/恢复 | PUT `/api/comments/{id}/status` |
| 举报中心 | 举报列表（status=2 待审核）、处理（通过/驳回） | GET `/api/comments/reports` + POST `/api/comments/{id}/report-handle` + `novel_audit_log` 审计 |

### 阶段 M3：用户管理模块（✅ 已完成，T-A-12~13）
用户列表（分页/搜索）、封禁/解封、角色调整。**后端配套**：admin 错误码段 18xxxx + requireAdmin（role=3 检查），替换前端硬编码。

### 阶段 M4：数据统计模块（✅ 已完成，T-A-14~15）
仪表盘（书籍数 / 用户数 / 评论数 / DAU 近似：当日登录 ∪ 当日搜索去重用户）、榜单（热门书籍 / 热门搜索词，复用 `/api/search/hot` + 搜索日志聚合）。端点：GET `/api/stats/overview`。

### 阶段 M5：配置管理（✅ 已完成，T-A-16）
分类 / 标签 CRUD：POST / PUT / DELETE `/api/categories{/id}`、`/api/tags{/id}`（requireAdmin）。

---

## 五、客户端规划（C 端）

### 阶段 C1：多语言补齐（✅ 已完成，T-C-01~04）
- Flutter：ARB 13 语种（zh-CN / en / ja / ko / fr / de / es / ru / pt / hi / ar / bn / id）
- HarmonyOS：13 语种资源目录
- 后端：`lang` 参数规范统一（`zh-CN` / `en` / `ja`…），`langCode()` 映射对齐；OpenSearch 补 ko（nori 分词）

### 阶段 C2：阅读体验增强（✅ 已完成，T-C-05~12）
- 字号 13~26 / 行距 1.2~2.6 调节并持久化（两端）
- 深浅主题（Material 3 ThemeMode：跟随系统 / 浅 / 深 + HarmonyOS dark 资源）
- 翻页方向（上下滚动 / 左右翻页双模式）；章节离线缓存（Flutter `flutter_cache_manager` LRU、HarmonyOS 沙箱 50 篇）
- 阅读进度精确到段落/滚动位置（复用 progress 端点 `position` 字段，后端未改）

### 阶段 C3：社区与发现（✅ 已完成，T-C-13~15）
- 评论增强：取消点赞、举报入口（两端对齐）
- 分类浏览 Tab（`/api/categories`）+ 热搜词榜（`/api/search/hot-keywords`）
- 搜索建议（`/api/search/suggest`，本地历史 20 条可清空 + 200ms 防抖）

### 阶段 C4：多端体验统一（✅ 已完成，T-C-16~23）
- 桌面端自适应布局（LayoutBuilder 宽屏多列 / 移动端底部导航）
- 分页策略统一（按 total 循环拉取，消除 500 vs 5000 章上限差异）
- token 持久化（shared_preferences / preferences），刷新不丢登录

---

## 六、后端规划（支撑端）

| 项 | 说明 | 优先级 |
| :--- | :--- | :--- |
| Admin API 领域 | `api/admin/v1`：统计（/api/stats/overview）、分类/标签 CRUD；admin 错误码段（18xxxx） | ✅ 已完成 |
| RBAC | requireAdmin（role=3 检查，service 层 helper），替换前端 role==3 硬编码 | ✅ 已完成 |
| 审计 | 管理操作 / 审核操作 / 用户状态变更写 `novel_audit_log`；查询页 GET /api/admin/audit-logs（分页 + user_id/action/target_type/target_id/时间范围筛选） | ✅ 已完成 |
| VIP / 支付 | `novel_payment_provider`（config AES-GCM 加密）/ `novel_payment_order` / `novel_vip_order` / `novel_vip_plan` 表 + Provider 抽象（11 渠道：Stripe / PayPal / NOWPayments-USDT / Razorpay / KOMOJU / PortOne / Mercado Pago / Xendit / Alipay-RSA2 / WeChat Pay Global / UnionPay-RSA）+ Webhook 验签 + 幂等 settle + 15min 超时关闭 + VIP 叠加续期；后台控制展示与流水 | ✅ 已完成（T-P-01~22） |
| AI 推荐 | 基于阅读/搜索行为埋点（`novel_search_log` 已有）的启发式画像推荐（`strategy=ai`，候选 <5 回退 hot）；数据达标后换第三方模型 | ✅ 已实现（2026-08-29，配置门控） |
| 本地二级缓存 | ristretto 进程内 L1（128MB / 30s TTL / 写路径双删），Redis 之上 | ✅ 已完成 |

---

## 七、阶段路线图

| 阶段 | 周期 | 内容 | 状态 |
| :--- | :--- | :--- | :--- |
| M2 内容审核 | 1-2 周 | 管理端书籍/章节/评论审核 + 后端审核端点 + RBAC | ✅ 已完成 |
| C1 多语言补齐 | 1-2 周 | ARB 13 语种、HarmonyOS 资源、后端 lang 对齐、ko nori | ✅ 已完成 |
| M3 用户管理 | 1 周 | 管理端用户模块 + admin API + 审计 | ✅ 已完成 |
| C2 阅读体验 | 1-2 周 | 字号/主题/翻页/离线缓存/进度精确化 | ✅ 已完成 |
| M4/M5 统计配置 | 1 周 | 仪表盘、分类/标签管理 | ✅ 已完成 |
| 支付链路 | 2-3 周 | Provider 底座 + 11 渠道（Stripe / NOWPayments-USDT / PayPal / Razorpay / KOMOJU / PortOne / Mercado Pago / Xendit / Alipay / WeChat Pay Global / UnionPay）+ 语言路由 + 后台流水 + 客户端 VIP（T-P-01~22） | ✅ 已完成 |
| C3 社区发现 | 1 周 | 评论增强、分类 Tab、热搜词榜、搜索建议（T-C-13~15） | ✅ 已完成 |
| 多端统一 / 收尾 | 1-2 周 | 分页统一、token 持久化、桌面布局（T-C-16~23、T-A-17） | ✅ 已完成 |
| 二期支付 | 按资质 | 本地支付逐语言（Razorpay/KOMOJU/PortOne/Mercado Pago/Xendit/Alipay）、PayPal（T-P-19~20） | ✅ 已完成 |

**里程碑**：M2+C1+支付链完成后平台形态完整（管理端可用 + 多语言达标 + 商业化闭环）；当前全部任务链（T-A-01~17 / T-C-01~23 / T-P-01~20）已完成，可选方向 7.4/7.5 已实现（2026-08-29，均配置门控默认关闭），仅剩 T-C-23 人工回归按需执行。

---

## 八、风险与技术债

1. **多语言内容供给**：翻译表（`novel_book_translation`）有结构但缺内容生产流程——需定义翻译工作流（人工/第三方翻译 API）。
2. ~~路由重叠疑点~~：已查实为非问题（2026-08-29）——admin 用独立路径 `GET /api/admin/categories` / `/api/admin/tags`（requireAdmin，admin.proto:25 注释记录设计决策），与 book 公开 `GET /api/categories` / `/api/tags`（无鉴权浏览用）路径不冲突、无遮蔽。
3. **支付资质**：微信支付国际版 2026-08-29 已接入（wechatpay_global，HK API v3 H5，平台公钥验签 + AES-GCM 解密，全球可用）；国内微信支付未接入（需中国大陆商户资质）；支付宝 2026-08-29 已接入（RSA2 验签，沙箱可用）；各渠道上线前需配置生产密钥。
