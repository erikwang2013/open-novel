# Open Novel 详细任务分解（2026-08 修订）

> 基于 [roadmap.md](roadmap.md) 的阶段规划拆解为可执行任务链。管理端链（T-A-01~17）对应 M2~M5，客户端链（T-C-01~23）对应 C1~C4。
> 约定：任务编号稳定不变；后端配套（🔧）先于前端页面（🖥）；每任务完成后跑 `scripts/smoke.sh` 冒烟回归。
> 状态盘点时间：2026-08-29（✅ 已完成 / ⏳ 进行中 / ⬜ 未开始）。

---

## 一、任务总览

| 链 | 编号 | 内容 | 阶段 | 状态 |
| :--- | :--- | :--- | :--- | :--- |
| 管理端 | T-A-01~07 | 后端配套：RBAC / 审核端点 / 审计 | M2 | ✅ 已完成 |
| 管理端 | T-A-08~11 | 内容审核前端（书籍 / 章节 / 评论 / 举报） | M2 | ✅ 已完成 |
| 管理端 | T-A-12~13 | 用户管理（API + 页面） | M3 | ✅ 已完成 |
| 管理端 | T-A-14~15 | 数据统计（API + 仪表盘） | M4 | ✅ 已完成 |
| 管理端 | T-A-16 | 配置管理（分类 / 标签） | M5 | ✅ 已完成 |
| 管理端 | T-A-17 | token 持久化（对齐 C4） | 收尾 | ✅ 已完成 |
| 客户端 | T-C-01~04 | 多语言补齐（后端 lang 对齐 + 两端资源） | C1 | ✅ 已完成 |
| 客户端 | T-C-05~12 | 阅读体验（字号 / 主题 / 翻页 / 缓存 / 进度） | C2 | ✅ 已完成 |
| 客户端 | T-C-13~15 | 社区与发现（评论增强 / 分类 / 热搜） | C3 | ✅ 已完成 |
| 客户端 | T-C-16~21 | 多端统一（分页 / token / 桌面布局） | C4 | ✅ 已完成 |
| 客户端 | T-C-22~23 | 加载空态统一 + 全链路回归 | 收尾 | ✅ 已完成 |
| 支付 | T-P-01~08 | 支付底座（表 / 错误码 / Stripe / NOWPayments / 下单查单 / VIP 激活） | 新增需求 | ✅ 已完成 |
| 支付 | T-P-09~13 | 支付管理（流水 / 方式 / 套餐 + 管理端 3 页） | 新增需求 | ✅ 已完成 |
| 支付 | T-P-14~17 | 客户端 VIP（购买页 / 结果页 / 阅读器引导 / 我的-VIP） | 新增需求 | ✅ 已完成 |
| 支付 | T-P-18 | 安全与合规（回调验签 / 金额校验 / 幂等 / 审计，随各批内嵌） | 新增需求 | ✅ 已完成 |
| 支付 | T-P-19~20 | 本地支付补强 / PayPal（二期） | 新增需求 | ✅ 已完成（国内微信支付需 CN 资质除外） |
| 支付 | T-P-21 | 微信支付国际版 wechatpay_global | 新增需求 | ✅ 已完成（2026-08-29，HK API v3） |
| 支付 | T-P-22 | 银联在线支付 unionpay | 新增需求 | ✅ 已完成（2026-08-29，UPOP 网关 RSA-SHA256） |

## 二、管理端链（T-A-01~17）

### 2.1 后端配套（M2 前置，🔧）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-A-01 | RBAC 校验 helper | service 层 `requireRole(...)` 中间件等价物（auth.go 无 per-route middleware），定义 role 常量（1 读者 / 2 作者 / 3 管理员），替换管理端前端 `role==3` 硬编码 | 无 |
| T-A-02 | admin 错误码段 | `errcode.go` 注册 18xxxx 段（如 180401 无权限 / 180402 不存在 / 180403 状态非法） | T-A-01 |
| T-A-03 | 书籍状态端点 | `PATCH /api/books/{id}/status`（上架 / 下架，状态机 1 上架 / 0 下架），复用 book biz | T-A-01 |
| T-A-04 | 章节状态端点 | `PATCH /api/chapters/{id}/status`（禁用 / 恢复，**禁用而非删除**，正文接口对禁用章节返回业务错误） | T-A-01 |
| T-A-05 | 评论状态端点 | `PUT /api/comments/{id}/status`（下架 / 恢复），修复 ListComments 硬编码 `status=1` 过滤（status 参数化） | T-A-01 |
| T-A-06 | 举报审核端点 | 举报列表（`status=2` 待审核 / 分页）+ 处理（通过 → 恢复评论 / 驳回 → 关闭举报），写审计 | T-A-05 |
| T-A-07 | 审计表与写入 | `novel_audit_log` 表（admin_id / action / target_type / target_id / detail / created_at）+ 所有审核 / 管理操作写入 | T-A-03~06 |

### 2.2 内容审核前端（M2，🖥）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-A-08 | 书籍管理页 | 列表（分页 / 搜索）、状态上下架、编辑元数据（复用 GET/POST `/api/books` + T-A-03） | T-A-03 |
| T-A-09 | 章节管理页 | 章节列表、正文查看、禁用 / 恢复（T-A-04） | T-A-04 |
| T-A-10 | 评论审核页 | 评论列表（按书籍 / 章节筛选）、下架 / 恢复（T-A-05） | T-A-05 |
| T-A-11 | 举报中心页 | 待审核列表、通过 / 驳回（T-A-06），操作结果展示 | T-A-06 |

### 2.3 用户管理（M3）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-A-12 | 用户管理 API | 用户列表（分页 / 搜索）、封禁 / 解封（status）、角色调整，全部 requireRole(3) | T-A-01 |
| T-A-13 | 用户管理页 | 列表 / 搜索 / 封禁 / 角色操作 + 确认弹窗 | T-A-12 |

### 2.4 数据统计（M4）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-A-14 | 统计端点 | 书籍数 / 用户数 / 评论数 / DAU 近似（audit_log 或用户活跃表）、热门书籍 / 热门搜索词（复用 `/api/search/hot`） | T-A-07 |
| T-A-15 | 仪表盘页 | 统计卡片 + 榜单表格 | T-A-14 |

### 2.5 配置管理（M5）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-A-16 | 分类 / 标签管理 | `POST /api/categories`、`PUT /api/categories/{id}`、`DELETE`（标签同），前端管理页 | T-A-01 |

### 2.6 收尾

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-A-17 | 管理端 token 持久化 | shared_preferences 持久化 access/refresh token，刷新页面免重登（对齐 T-C-18） | T-A-13 后 |

## 三、客户端链（T-C-01~23）

### 3.1 多语言补齐（C1，🔴 高）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-C-01 | 后端 lang 规范统一 | `lang` 参数规范（`zh-CN` / `en` / `ja`…），`langCode()` 映射与两端对齐，文档同步 api.md | 无 |
| T-C-02 | OpenSearch 分词补齐 | ko（nori）、通用 ICU 分词，索引模板 / 映射更新 + 重建脚本 | 无 |
| T-C-03 | Flutter ARB 扩展 12+ 语种 | zh / en / ja / ko / fr / de / es / ru / ar / pt / hi / bn / id，每语种 ~40 键起步，L10n 生成接入 | 无 |
| T-C-04 | HarmonyOS 资源按语种补齐 | `resources/<lang>/`（base/en_US 已有，补其余语种 string），运行时按 `lang` 切换 | 无 |

### 3.2 阅读体验（C2）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-C-05 | Flutter 字号 / 行距 | 设置项替换硬编码 `fontSize 17`，阅读页即时生效 + 持久化 | 无 |
| T-C-06 | HarmonyOS 字号 / 行距 | 同 T-C-05，preferences 持久化 | 无 |
| T-C-07 | 深浅主题 | Flutter Material 3 `ThemeMode`（跟随系统 / 浅 / 深）+ 阅读页适配 | 无 |
| T-C-08 | HarmonyOS 深浅色 | 深浅色资源目录 + 跟随系统切换 | 无 |
| T-C-09 | 翻页方向 | 上下滚动 / 左右翻页两种模式，两端对齐 | T-C-05/06 后 |
| T-C-10 | Flutter 章节离线缓存 | 正文本地缓存（flutter_cache_manager），离线可读，缓存淘汰策略 | 无 |
| T-C-11 | HarmonyOS 章节离线缓存 | 同 T-C-10（沙箱文件 / preferences） | 无 |
| T-C-12 | 阅读进度精确化 | 进度精确到滚动位置（当前 `position=0` 章节粒度 → 段落粒度），两端对齐 | T-C-09 |

### 3.3 社区与发现（C3）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-C-13 | 评论增强 | 取消点赞（后端已有 `/api/comments/{id}/unlike`，两端补齐入口）+ 举报入口（对接 T-A-06 的 status=2 举报表） | T-A-06 |
| T-C-14 | 分类浏览 Tab | 首页 / 书库接入 `/api/categories` 分类筛选，两端对齐 | 无 |
| T-C-15 | 热门搜索词 | 搜索页展示 `/api/search/hot` 热搜词，点击即搜 | 无 |

### 3.4 多端统一（C4）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-C-16 | Flutter 分页统一 | 书架批量拉章节按 `total` 循环（消除 500 章上限） | 无 |
| T-C-17 | HarmonyOS 分页统一 | 同 T-C-16（5000 章上限 → total 循环） | 无 |
| T-C-18 | Flutter token 持久化 | shared_preferences 持久化，刷新页面免重登 | 无 |
| T-C-19 | HarmonyOS token 恢复 | 启动时从 preferences 恢复会话（已有读写，补启动恢复流程） | 无 |
| T-C-20 | 桌面端自适应 | Flutter `LayoutBuilder` 宽屏多列（书库 / 详情 / 阅读），移动端保持底部导航 | 无 |
| T-C-21 | HarmonyOS 宽屏适配 | 平板 / 折叠屏布局检查与适配 | 无 |

### 3.5 收尾

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-C-22 | 加载 / 空态统一 | 列表加载态、错误重试、空数据文案两端对齐 | 无 |
| T-C-23 | 全链路回归 | `smoke.sh` 扩展覆盖新端点（status / audit / 举报 / 统计），两端手测清单 | T-A/T-C 主体 |

#### T-C-23 两端手测清单

- [ ] **登录**：Flutter 与 HarmonyOS 各自完成 注册→登录→退出→重新登录 全流程；token 重启 App 后仍有效
- [ ] **书库分类筛选**：全部书籍 tab 顶部分类 chip 点击切换，列表随之刷新；宽屏（≥900vp）多列不回归
- [ ] **热搜**：首页/书库热搜词点击触发搜索，结果正确
- [ ] **评论点赞/取消**：评论区点赞后数字 +1 且图标变实心，再点取消 -1 恢复空心；重复请求不报错
- [ ] **举报**：评论 flag 图标 → 确认弹窗 → 成功 toast；管理端可在举报列表看到该条（status=2）
- [ ] **书架分页**：我的页书架加载正常（>20 本时翻页完整），移出书架后列表即时更新
- [ ] **阅读器**：打开书详情 → 章节列表 → 进入阅读；章节翻页、进度保存（重新进入跳转上次章节）
- [ ] **VIP 购买**：VIP 页套餐列表展示 → 下单跳转支付 → 返回后状态刷新为已开通；到期时间展示正确
- [ ] **加载/错误/空态**：列表页三态两端一致——加载居中转圈、错误有重试按钮、空数据有文案（评论空态用 emptyComment）

## 四、并行与调度建议

| 批次 | 任务 | 说明 | 状态 |
| :--- | :--- | :--- | :--- |
| 第一批 | T-A-01~07 ‖ T-C-01~04 | 后端审核底座与 C1 多语言无依赖，可并行 | ✅ 已完成 |
| 第二批 | T-A-08~11（M2 前端）‖ T-C-05~12（C2） | M2 前端依赖第一批后端；C2 独立 | ✅ 已完成 |
| 第三批 | T-A-12~13（M3）‖ T-C-13~15（C3） | C3 的举报入口依赖 T-A-06 已就绪 | ✅ 已完成 |
| 第四批 | T-A-14~16（M4/M5）‖ T-C-16~21（C4） | 均无强依赖 | ✅ 已完成 |
| 收尾 | T-A-17、T-C-22~23 | 统一收尾 + 回归 | ✅ 已完成 |

**里程碑**：第一批 + 第二批完成后平台形态完整（管理端可用 + 多语言达标），对应 roadmap「M2+C1 完成」节点。

## 五、不在本期范围

| 项 | 说明 | 触发条件 |
| :--- | :--- | :--- |
| AI 推荐 | 埋点（`novel_search_log` 已有）后算法推荐替代策略推荐 | 埋点数据量达标 |
| ristretto 本地二级缓存 | 超热 key 兜底 | 压测发现缓存瓶颈 |
| 离线缓存（后端 CDN 侧） | 章节静态化 / CDN 预热 | 流量规模达阈值 |

---

## 六、支付链（T-P-01~20，2026-08-29 用户立项）

定位：平台商业化支撑。国际主流支付 + 按语言优先本地支付 + USDT 等主流币；后台可控制支付方式展示与流水对账。

**调研结论（2026-08）**：
- **Stripe**：官方 Go SDK（stripe-go），135+ 币种、120+ 本地方式、>45 国收单 —— 一期首选（欧美卡 + iDEAL/SEPA/Apple Pay/Google Pay 全覆盖；韩日本地卡弱）
- **本地聚合网关**（二期按语言接入，实现同一 Provider 接口，语言路由表优先本地）：hi→Razorpay(UPI)、ja→KOMOJU、ko→PortOne/Toss、pt-BR→Mercado Pago、id/th/vn→Xendit、zh-CN→支付宝/微信（需 CN 企业资质或经 Adyen）
- **USDT/主流币**：NOWPayments（350+ 币、0.5~1% 费率、REST + HMAC-SHA512 webhook + 沙箱，TRON/ERC20/BSC）—— 一期；BitPay 备选
- **架构**：Provider 抽象接口（CreateCheckout / Verify / VerifyWebhook）+ 语言→方式路由表（DB 驱动，admin 控制 enabled / region / sort）

| 编号 | 任务 | 说明 | 依赖 |
| :--- | :--- | :--- | :--- |
| T-P-01 | 支付表结构 | `novel_payment_provider`（code/lang/region/enabled/sort/config 密钥加密）、`novel_payment_order`（order_no/user_id/amount/currency/provider/status/tx_id/paid_at）、`novel_vip_order`（套餐/时长/到期）；`novel_user` 加 `vip_expires_at`；init.sql | 无 |
| T-P-02 | 支付错误码 | 19xxxx 段：190401 下单失败 / 190402 订单不存在 / 190403 金额不匹配 / 190404 支付未完成 / 190405 方式未启用 | T-P-01 |
| T-P-03 | Provider 抽象 + payment.v1 | proto（下单 / 查单 / 方式列表 / 回调）+ Provider 接口 + 注册表 | T-P-01/02 |
| T-P-04 | Stripe 实现 | 官方 stripe-go SDK，Checkout Session + webhook 签名验签，沙箱先行 | T-P-03 |
| T-P-05 | NOWPayments 实现 | USDT（TRON/ERC20/BSC）支付 + IPN webhook 验签 | T-P-03 |
| T-P-06 | 语言→方式路由 | `GET /api/payments/methods?lang=`：本地支付优先、enabled/sort/region 过滤；后台可控 | T-P-01 |
| T-P-07 | 下单/查单 | 幂等、15min 超时未付、webhook 幂等处理、回调金额=订单金额强校验、FORCE_MASTER + Redis | T-P-04/05 |
| T-P-08 | VIP 激活链路 | 支付成功 → `vip_expires_at` 续期（可叠加）、VIP 章节/书籍校验打通（IsVip 已有） | T-P-07 |
| T-P-09 | 流水账单 API | 订单分页/按用户/方式/状态/时间筛选 + 汇总统计，requireAdmin | T-P-07 |
| T-P-10 | 支付方式管理 API | provider CRUD + enabled + region + sort，密钥加密存储，requireAdmin | T-P-01 |
| T-P-11 | 管理端支付方式页 | 列表 / 启停 / 排序 / 区域 / 密钥配置 | T-P-10 |
| T-P-12 | 管理端流水页 | 分页 / 筛选 / 详情 / 汇总卡片 | T-P-09 |
| T-P-13 | 管理端 VIP 套餐配置 | 套餐 CRUD（时长 / 价格 / 币种 / 标签） | T-P-01 |
| T-P-14 | 客户端 VIP 购买页 | 套餐列表 + 按语言推荐支付方式 + 下单 / 跳转 / 回跳 | T-P-06/07 |
| T-P-15 | 支付结果页 | 成功 / 失败 / 待确认态 + 订单轮询 | T-P-14 |
| T-P-16 | 阅读器 VIP 引导 | VIP 章节未订阅 → 引导购买（两端对齐） | T-P-08 |
| T-P-17 | 我的-VIP 状态 | 到期时间 / 续费入口（两端） | T-P-08 |
| T-P-18 | 安全与合规 | 回调验签、金额防篡改、幂等、审计日志、日志脱敏 | T-P-04~07 |
| T-P-19 | 本地支付补强（二期） | 按语言：hi→Razorpay、ja→KOMOJU、ko→PortOne、pt-BR→Mercado Pago、id/th/vn→Xendit、zh-CN→支付宝/微信（需 CN 资质） | T-P-03 后逐语言 |
| T-P-20 | PayPal 接入（备选） | Stripe 未覆盖市场兜底 | T-P-03 |
| T-P-21 | 微信支付国际版（wechatpay_global） | HK API v3 H5 支付；WECHATPAY2 请求签名 + 平台公钥 RSA-SHA256 验签 + apiv3_key AES-GCM 解密 resource；全球可用不需 CN 资质 | T-P-03 |
| T-P-22 | 银联在线支付（unionpay） | UPOP 网关支付 5.1（zh-CN）；字典序 RSA-SHA256 签名验签（notify 表单，应答 success）；base_url 沙箱覆盖；需银联商户资质 | T-P-03 |

**调度**：第一批 T-P-01~08（后端支付底座）→ T-P-09~13（管理端）→ T-P-14~17（客户端），2026-08-29 前已全部提交（✅）；T-P-18 随各批内嵌完成；T-P-19/20 二期已实现（✅，razorpay/komoju/portone/mercadopago/xendit/paypal 后端全部就绪）；T-P-21 微信支付国际版 2026-08-29 已实现（✅，wechatpay_global，请求签名/平台公钥验签/AES-GCM 解密均有单测）；T-P-22 银联 UPOP 2026-08-29 已实现（✅，unionpay，字典序 RSA-SHA256 验签有单测）；zh-CN 支付宝/微信/银联需 CN 企业资质，上线前配真实商户密钥联调。
**前置条件**：Stripe / NOWPayments 商户密钥（沙箱即可开发，当前配置为沙箱、enabled 渠道为空）；zh-CN 支付宝/微信需中国大陆企业资质（或 Adyen 渠道），T-P-19 立项时确认。

---

## 七、可选方向规划（2026-08-29）

定位：主链全部 ✅ 后的增量优化，独立立项、可并行、互不依赖。方向 1/2/3 为可执行级，4/5 为门槛项（只给触发条件 + 方案草图），6 为回归引用。涉及后端均只改 `kratos/backend/`，不碰 `kratos/` 框架源码。

### 7.1 搜索历史 / 搜索建议（方向 1）

**范围**：客户端搜索框本地历史（近 20 条、可清空）+ 输入建议（search_log 聚合热词补全）。客户端为主 + 轻后端。搜索框位于 `apps/client/flutter/lib/pages/books_tab.dart`（搜索页内嵌于该 tab），建议页同文件或独立 `search_suggest.dart`；token 持久化已有先例（shared_preferences），本地历史复用同一存储即可，无需新依赖。

**后端触点（1 个新 API）**：
- `kratos/backend/api/search/v1/search.proto`：新增 `rpc Suggest(SuggestReq) returns (SuggestReply)`，`get: "/api/search/suggest?q="`，返回 `repeated string keywords`（匹配 `keyword LIKE 'q%'`，按 search_log 聚合 count 降序，LIMIT 10）
- `kratos/backend/internal/biz/search.go`：参照 `HotKeywords`（已按 `Model(&data.SearchLog{}).Group("keyword")` 聚合 TOP 10，口径已验证），新增 `Suggest` 查询（LIKE 前缀 + count 排序），补 1s 级别 Redis 缓存（key `suggest:{q}`）
- `kratos/backend/internal/service/search.go`：新增适配层方法（参照 `HotKeywords`，protojson camelCase 自动处理，int64 无涉及）

**实现步骤**：① proto + 重新生成 → ② biz Suggest 查询（复用 HotKeywords 聚合口径）→ ③ service 适配 + 测试（参照 `search_test.go`）→ ④ 客户端本地历史（shared_preferences，输入时写入、按时间倒序去重、上限 20、清空按钮）→ ⑤ 建议接口接入（防抖 200ms，`asStr` 取关键字）。

**工作量**：后端 S（一个查询 + 一个端点）；客户端 S（本地存储 + UI 集成），合计 S~M。
**风险**：低。Suggest LIKE 前缀查询无索引时可能慢——`novel_search_log.keyword` 量级大后需加索引或换 ES 前缀查询（当前量级可忽略，`ponytail:` 标注：数据量大时在 search_log 加 keyword 前缀索引即可）。

### 7.2 管理端审计日志查询页（方向 2）

**范围**：后端 audit_log 查询 API（requireAdmin）+ 管理端列表页（分页 / 筛选 by admin / action / target / 时间）。

**后端触点**：
- `kratos/backend/internal/data/models.go`：`AuditLog` 模型已存在（`novel_audit_log`，字段 ID/UserID/Action/TargetType/TargetID/Detail/IP/UserAgent/CreatedAt），无需改表
- `kratos/backend/api/admin/v1/admin.proto`：新增 `rpc ListAuditLogs(ListAuditLogsReq) returns (ListAuditLogsReply)`，`get: "/api/admin/audit-logs"`，query 参数 `page/page_size/user_id/action/target_type/target_id/start_time/end_time`，返回分页列表（复用 `pkg.Page`）
- `kratos/backend/internal/biz/admin.go`：新增 `ListAuditLogs`，`WHERE` 条件动态拼接（各字段可选），`ORDER BY created_at DESC`，时间范围用 `>= start_time AND <= end_time`；参照现有管理端 stats 查询口径
- `kratos/backend/internal/service/admin.go`：requireAdmin 守卫（参照 `GetStats` 已有 `requireAdmin(ctx)` 模式，T-A-14/16 已建立），返回 protojson camelCase + int64 string
- 写入侧已存在：`kratos/backend/internal/biz/user.go:135` 等（login 等动作写 AuditLog），本方向只做查询

**管理端触点**：`apps/admin/lib/pages/` 新增 `audit_logs_page.dart`（参照 `reports_page.dart` / `orders_page.dart` 的分页表格 + 筛选模式），路由注册在 `apps/admin/lib/pages/home_page.dart`（或对应导航处）。

**实现步骤**：① proto + 生成 → ② biz 查询（条件拼接）→ ③ service 适配 + requireAdmin + 测试 → ④ 管理端页面（分页表 + 筛选表单）→ ⑤ 导航接入。

**工作量**：后端 S；管理端 S（已有页面模板可复制），合计 S~M。
**风险**：低。条件拼接注意 SQL 注入（全部走 GORM 参数绑定，不拼字符串）；audit_log 无 created_at 索引时按时间筛选会全表扫——量级大后补索引（`ponytail:` 标注）。

### 7.3 ristretto 进程内二级缓存（方向 3）

**范围**：纯后端性能项，超热 key 兜底 Redis。不改业务代码——`data/cache.go` 的 `Cache` 封装（`Get/Set/Del/DelPattern/GetOrLoad`）是唯一读写 Redis 的咽喉，L1 拦截点收敛于此，单文件改动。

**后端触点**：仅 `kratos/backend/internal/data/cache.go`。`Get` 先查 ristretto，未命中回源 Redis 并回填 L1（`GetOrLoad` 同路线）；`Set` 同步写 L1 + Redis；`Del/DelPattern` 双删。ristretto 需新增依赖 `github.com/dgraph-io/ristretto`（当前 `go.mod` 无此依赖）——这是本项目极少见的加依赖场景，属合理例外，但必须确认压测已证明瓶颈（懒人原则：Redis 命中率 >95% 时不值得加）。

**实现步骤**：① 压测确认 Redis 是热点瓶颈 → ② go.mod 加 ristretto + `data.go` 初始化 L1（容量按可用内存配置，如 128MB，TTL 30s 短过期兜底一致性）→ ③ `cache.go` Get/Set/Del 接 L1 → ④ 回归压测对比命中率与 p99。

**工作量**：S（单文件 + 初始化，改动面 <50 行）。
**风险**：中。L1 数据一致性——短 TTL（30s）兜底，写路径双删避免脏读；内存占用需监控（ristretto 超出容量会逐出，无溢出风险但有命中率下降）。**触发条件**：压测/线上 Redis 读 QPS 达到瓶颈（如 >10k QPS 且热点集中），否则不做。

### 7.4 AI 推荐（方向 4，门槛项）

**状态**：✅ 已实现（2026-08-29，`strategy=ai` 启发式画像排序，候选 <5 回退 hot；行为数据量达标后换第三方模型，替换点在 `rankAI`）。

**触发条件**：`novel_recommend_log` + `novel_search_log` 累积 ≥ 阈值（如单语言 ≥ 10 万条有效行为），且业务侧确认第三方 API 预算。当前推荐为 `recommend.go` 仅 hot/new 策略，AI 属替代策略，不是补丁。

**方案草图（≤10 行）**：① 现有 `recommend.go` 的 `Log()` 已在写 impression 日志（`novel_recommend_log`，含 UserID/BookID/Strategy/RankNo），无需新埋点 → ② 批量导出行为数据 → ③ 第三方（如 OpenAI embeddings 或推荐 SaaS）离线训练/计算 → ④ `recommend.go` 新增 strategy 分支（如 `ai`），按语言路由 + 预热到 Redis → ⑤ `recommendation.proto` 无接口变更，策略参数走现有 `GetRecommendationsReq`；未达标前不接第三方 API（无账单、无数据泄露面）。数据量达标前，hot/new 策略 + 热搜词已够用。

### 7.5 CDN 章节静态化（方向 5，门槛项）

**状态**：✅ 已实现（2026-08-29，`CDN_BASE_URL`/`CDN_PURGE_URL` 门控默认关闭；免费章节 `Cache-Control: public, s-maxage=3600`，VIP `no-store`；章节创建/状态变更 fire-and-forget purge，key 约定 `chapter/{id}?lang={lang}`）。

**触发条件**：章节读流量达阈值（如单章日读 > 1 万次或 CDN 成本超过源站 CPU 成本），且章节变更频率低（连载期间不适用——章节每日更新，静态化收益被失效成本抵消）。

**方案草图（≤10 行）**：① 章节内容为纯文本，天然可静态化 → ② 后端生成 HTML/纯文本静态文件推 CDN（或对象存储 + CDN 域名）→ ③ `chapter.proto` 读章节接口加 CDN 回退：CDN 未命中 → 源站 → 回源写 CDN（cache miss 回源模式，不阻塞主流程）→ ④ 章节发布时主动失效 CDN key（发布侧已有更新链路，加一个 invalidate 调用即可）→ ⑤ 需评估：付费章节鉴权与静态化冲突（VIP 章节无法走公开 CDN，只静态化免费章节或加签名 URL）。当前量级下直接做属于过度设计。

### 7.6 人工回归 T-C-23（方向 6）

引用 `docs/tasks.md` 上文 T-C-23（9 项两端手测清单，含搜索、推荐、评论、支付等两端对齐验证），按需执行，无开发工作。

**调度建议**：7.1 + 7.2 可并行（后端两个独立 proto 文件）；7.3 独立、等压测信号；7.4/7.5 已实现（2026-08-29，均配置门控默认关闭，仅启用后才生效）；7.6 人工回归按需执行。
