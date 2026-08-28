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

## 二、现状盘点（2026-08 已实现）

### 后端（`kratos/backend/`，6 个服务领域）

| 领域 | 已实现端点 | 状态 |
| :--- | :--- | :--- |
| 用户 | register / login / refresh / me | ✅ |
| 书籍 | 列表 / 详情 / 创建 / 翻译更新 / categories / tags | ✅ |
| 章节 | 章节 CRUD / 正文 / 进度 / 书架 | ✅ |
| 评论 | 列表 / 发布 / 点赞 / 取消点赞 / 举报 / 收藏 / 收藏列表 | ✅ |
| 搜索 | 搜索 / 热门词 / 索引同步 / 索引删除 | ✅ |
| 推荐 | recommend（strategy=hot / new） | ✅ |

基础设施：JWT + Refresh 轮换（Redis GETDEL 防重放）、可选鉴权中间件、按路径限流、业务错误码（11xxxx 用户 / 12xxxx 书籍 / 14xxxx 章节 / 15xxxx 评论 / 16xxxx 搜索 / 17xxxx 推荐，HTTP 200 + `{code,reason,message}`）、Redis 缓存（5min±30s 抖动 / 空值防穿透 / SETNX 单飞）、读写分离（FORCE_MASTER）、OpenSearch 多语言索引（zh/en/ja kuromoji）、API 版本协商。

### 客户端（`apps/client/`）

| 功能 | Flutter | HarmonyOS |
| :--- | :---: | :---: |
| 登录 / 注册 / 退出 | ✅ | ✅ |
| 首页推荐 + 语言切换（zh/en） | ✅ | ✅ |
| 书库列表 + 搜索 | ✅ | ✅（含分页加载更多） |
| 书籍详情（收藏 / 书架 / 评论入口） | ✅ | ✅ |
| 阅读器（章节正文 + 上/下一章 + 进度保存） | ✅ | ✅ |
| 评论区（列表 / 发布 / 点赞） | ✅ | ✅ |
| 我的（书架 / 收藏 / 最近进度） | ✅ | ✅ |
| i18n | zh/en ARB（37 键） | base/en_US |

### 管理端（`apps/admin/`）

| 模块 | 状态 |
| :--- | :--- |
| 登录（role==3 管理员校验） | ✅ |
| 主框架（NavigationRail：仪表盘 / 书籍 / 评论 / 用户） | ✅ 骨架 |
| 业务模块 | ❌ 全部「待实现」占位 |
| 管理 API（后端） | ❌ 后端无 admin 领域 |

### CI / 脚本

- `.github/workflows/release.yml`：push main 自动递增 v* tag + 增量 changelog 发 Release（幂等）
- `scripts/post-push.sh`：本地手动兜底；`scripts/smoke.sh`：核心接口冒烟测试

---

## 三、差距分析（规划 vs 现状）

| 规划项（原规划文档） | 现状 | 优先级 |
| :--- | :--- | :--- |
| 12+ 语种 i18n | 客户端仅 zh/en（README 12 语种仅文档级） | 🔴 高 |
| 管理端 | 仅骨架 | 🔴 高（本期重点） |
| VIP / 支付链路 | 完全缺失（表 / 接口均无） | 🔴 高（2026-08-29 用户立项，见 [tasks.md](tasks.md) 支付链） |
| 字体字号 / 深浅主题 / 翻页动画 | 未实现（字号硬编码） | 🟡 中 |
| 离线缓存（flutter_cache_manager） | 未实现 | 🟡 中 |
| 桌面端自适应布局 | 未实现 | 🟢 低 |
| ristretto 本地二级缓存 / nori 分词 | 未实现 | 🟢 低 |

技术债：Flutter 书架批量拉章节上限 500 章 vs HarmonyOS 5000 章（分页策略不一致，均未按 total 循环）；管理端/客户端 token 内存态（刷新页面需重登）；后端无角色检查中间件（管理员仅前端 role==3 判定）。

---

## 四、管理端规划（B 端）

### 阶段 M1：登录 + 框架（已完成 ✅）
登录（复用 `/api/users/login`，role==3 校验）→ NavigationRail 主框架。

### 阶段 M2：内容审核模块（下一期）
| 子功能 | 说明 | 后端配套 |
| :--- | :--- | :--- |
| 书籍管理 | 列表（分页/搜索）、状态上下架、编辑元数据 | 复用 GET/POST `/api/books` |
| 章节管理 | 章节列表、正文查看、禁用/删除 | 复用 chapter 端点 |
| 评论审核 | 评论列表（按书籍/章节筛选）、下架/恢复 | 新增评论状态管理端点 |
| 举报中心 | 举报列表（status=2 待审核）、处理（通过/驳回） | 新增审核端点 + `novel_audit_log` 审计 |

### 阶段 M3：用户管理模块
用户列表（分页/搜索）、封禁/解封、角色调整。**后端配套**：新增 admin 领域（`POST /api/admin/users` 等）+ **RBAC 中间件**（role 检查，定义 role=3 管理员常量，替换前端硬编码）。

### 阶段 M4：数据统计模块
仪表盘（书籍数 / 用户数 / 评论数 / 日活）、榜单（热门书籍 / 热门搜索词，复用 `/api/search/hot`）。

### 阶段 M5：配置管理
分类 / 标签管理（复用 `/api/categories`、`/api/tags` 增删改）。

---

## 五、客户端规划（C 端）

### 阶段 C1：多语言补齐（🔴 高，与平台定位直接相关）
- Flutter：ARB 扩展至 12+ 语种（zh / en / ja / ko / fr / de / es / ru / ar / pt / hi / bn / id），每语种 ~40 键起步
- HarmonyOS：资源目录按语种补齐（`resources/<lang>/`）
- 后端：`lang` 参数规范统一（`zh-CN` / `en` / `ja`…），`langCode()` 映射与后端对齐；OpenSearch 补 ko（nori）、通用 ICU 分词

### 阶段 C2：阅读体验增强
- 字号 / 行距调节（替换硬编码 fontSize 17）＋ 深浅主题（Material 3 ThemeMode + HarmonyOS 深浅色资源）
- 翻页方向（上下滚动 / 左右翻页）；章节缓存（正文本地缓存，离线可读，`flutter_cache_manager`）
- 阅读进度精确到滚动位置（当前章节粒度 position=0）

### 阶段 C3：社区与发现
- 评论增强：取消点赞、举报入口（两端对齐）
- 分类浏览 Tab（`/api/categories`）+ 热门搜索词（`/api/search/hot`）

### 阶段 C4：多端体验统一
- 桌面端自适应布局（LayoutBuilder 宽屏多列 / 移动端底部导航）
- 分页策略统一（按 total 循环拉取，消除 500 vs 5000 章上限差异）
- token 持久化（shared_preferences / preferences），刷新不丢登录

---

## 六、后端规划（支撑端）

| 项 | 说明 | 优先级 |
| :--- | :--- | :--- |
| Admin API 领域 | `api/admin/v1`：用户管理、评论审核、统计端点；admin 错误码段（18xxxx） | 🔴 随管理端 |
| RBAC 中间件 | role 检查（1 读者 / 2 作者 / 3 管理员），替换前端 role==3 硬编码 | 🔴 随管理端 |
| 审计完善 | 管理操作 / 审核操作写 `novel_audit_log` | 🟡 |
| VIP / 支付 | `novel_payment_provider` / `novel_payment_order` / `novel_vip_order` 表 + Provider 抽象（Stripe / NOWPayments-USDT / 本地聚合），VIP 章节校验链路；后台控制展示与流水 | 🔴 高（2026-08-29 立项，任务分解见 [tasks.md](tasks.md) T-P-01~20） |
| AI 推荐 | 基于阅读/搜索行为埋点（`novel_search_log` 已有），算法推荐替代策略推荐 | 🟢 |
| 本地二级缓存 | ristretto 兜底超热 key | 🟢 |

---

## 七、阶段路线图

| 阶段 | 周期 | 内容 | 依赖 |
| :--- | :--- | :--- | :--- |
| M2 内容审核 | 1-2 周 | 管理端书籍/章节/评论审核 + 后端审核端点 + RBAC | 无 |
| C1 多语言补齐 | 1-2 周 | ARB 12+ 语种、HarmonyOS 资源、后端 lang 对齐 | 无（可与 M2 并行） |
| M3 用户管理 | 1 周 | 管理端用户模块 + admin API + 审计 | M2 |
| C2 阅读体验 | 1-2 周 | 字号/主题/翻页/离线缓存 | 无 |
| C3 社区发现 | 1 周 | 评论增强、分类 Tab | 无 |
| M4/M5 统计配置 | 1 周 | 仪表盘、分类/标签管理 | M3 |
| 支付链路 | 2-3 周 | Provider 底座 + Stripe/USDT + 语言路由 + 后台流水（T-P-01~18） | 商户资质确认 |
| 多端统一 / 收尾 | 1-2 周 | 分页统一、token 持久化、桌面布局 | C2/C3 后 |

**里程碑**：M2+C1 完成后平台形态完整（管理端可用 + 多语言达标）；VIP 之后进入商业化阶段。

---

## 八、风险与技术债

1. **多语言内容供给**：翻译表（`novel_book_translation`）有结构但缺内容生产流程——需定义翻译工作流（人工/第三方翻译 API）。
2. **RBAC 后置**：管理端功能上线的安全前提是后端角色校验，勿仅依赖前端判断。
3. **分页不一致**：Flutter 500 章 vs HarmonyOS 5000 章上限，需统一按 total 循环。
4. **token 内存态**：管理端/客户端刷新页面需重登，体验债，C4 处理。
5. **支付资质**：支付宝/微信等本地支付需当地企业资质（zh-CN 尤甚），本地支付逐语言启用受商户开通周期限制；开发期全部走沙箱，T-P-19 依赖资质确认。
