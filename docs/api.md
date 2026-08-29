# Open Novel API 文档

后端 HTTP 接口统一前缀 `/api`，**版本不写在 URL**，通过请求头协商：

| Header | 值 | 说明 |
| :--- | :--- | :--- |
| `X-Api-Version` | `v1` | 必填；缺失或不支持的值返回 `140426 API_VERSION_MISMATCH` |
| `Authorization` | `Bearer <accessToken>` | 登录/注册后获取；受保护接口必填 |

- Base URL：`http://<host>:8000`（开发默认 `http://localhost:8000`）
- 序列化：JSON 驼峰字段；**int64 字段（id 等）序列化为字符串**
- 分页：`page`（从 1 起）+ `page_size`，响应 `{list, total, page, pageSize}`
- 多语言：查询参数 `lang=zh-CN|en|ja|ko`（`zh-CN` 默认，`zh` 为 `zh-CN` 别名），影响书籍标题/简介/章节内容返回的语言版本

## 错误格式

统一 JSON：`{"code": <int>, "reason": <string>, "message": <string>, "detail": null}`

- 业务错误码 ≥ 100000 时 **HTTP 状态码为 200**，客户端须解析 `code`
- 格式：`1 + 服务号 + 3 位标准码`，如 `110401` = User 服务未认证、`140404` = Chapter 未找到
- 标准码：400 参数 / 401 未认证 / 403 无权限 / 404 未找到 / 409 冲突 / 426 版本不符 / 429 限流 / 500 内部

## 用户

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| POST | `/api/users/register` | 注册 `{username, password, email, nickname}` | 无 |
| POST | `/api/users/login` | 登录 `{username, password}` → `{accessToken, refreshToken, user}` | 无 |
| POST | `/api/users/refresh` | 刷新 `{refresh_token}` → 新 token 对 | 无 |
| GET | `/api/users/me` | 当前用户信息 | Bearer |
| GET | `/api/users` | 用户列表（`page`/`page_size`/`search` 模糊匹配 username/nickname/email） | Bearer（管理） |
| PATCH | `/api/users/{id}/status` | 封禁/解封 `{status: 0|1}`（0 封禁 1 解封；禁止操作自己） | Bearer（管理） |
| PATCH | `/api/users/{id}/role` | 角色调整 `{role: 1|2|3}`（读者/作者/管理员；禁止操作自己） | Bearer（管理） |

用户管理错误码：`180401` 无权限 / `180402` 目标不存在 / `180403` 非法状态变更（非法 status/role 值或操作自己）。每次状态/角色变更写入审计日志。

## 书籍

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/books` | 书籍列表（分页 + `lang`） | 无 |
| GET | `/api/books/{id}` | 书籍详情 | 无 |
| POST | `/api/books` | 创建书籍 | Bearer（管理） |
| PUT | `/api/books/{id}/translation` | 创建/更新书籍翻译（多语言标题/简介） | Bearer（管理） |
| POST | `/api/books/{book_id}/favorite` | 收藏书籍 | Bearer |
| DELETE | `/api/books/{book_id}/favorite` | 取消收藏 | Bearer |
| GET | `/api/categories` | 分类列表（一级分类，客户端浏览用） | 无 |
| GET | `/api/tags` | 标签列表（`lang` 过滤） | 无 |

> 分类/标签写操作（POST / PUT / DELETE）见下方「管理端统计与配置」；管理端列表接口为 `GET /api/admin/categories`、`GET /api/admin/tags`（全量含 status/sort_order，见下方表格）。

## 章节

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/books/{book_id}/chapters` | 章节列表（分页 + `lang`） | 无 |
| POST | `/api/books/{book_id}/chapters` | 创建章节 | Bearer（管理） |
| GET | `/api/chapters/{id}/content` | 章节正文（`lang`） | VIP 章节需 Bearer |
| GET | `/api/progress` | 阅读进度 | Bearer |
| PUT | `/api/progress` | 保存阅读进度（`position` 支持段落/滚动位置粒度） | Bearer |
| GET | `/api/bookshelf` | 书架 | Bearer |
| POST | `/api/bookshelf` | 加入书架 `{book_id}` | Bearer |
| DELETE | `/api/bookshelf/{book_id}` | 移出书架 | Bearer |

## 评论

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/comments` | 评论列表（`book_id`、可选 `chapter_id`） | 无 |
| POST | `/api/comments` | 发布评论 `{book_id, chapter_id?, content}` | Bearer |
| POST | `/api/comments/{id}/like` | 点赞 | Bearer |
| DELETE | `/api/comments/{id}/like` | 取消点赞 | Bearer |
| POST | `/api/comments/{id}/report` | 举报 | Bearer |
| GET | `/api/favorites` | 收藏列表 | Bearer |

## 管理审核

仅管理员（role=3）可用；越权返回 `180401`。

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| PATCH | `/api/books/{id}/status` | 书籍上下架 `{status: 0|1}` | Bearer（管理） |
| PATCH | `/api/chapters/{id}/status` | 章节启用/禁用 `{status: 0|1}`（禁用章节正文不可读） | Bearer（管理） |
| PUT | `/api/comments/{id}/status` | 评论状态 `{status: 0|1}` | Bearer（管理） |
| GET | `/api/comments/reports` | 举报列表（status=2 待审核，分页） | Bearer（管理） |
| POST | `/api/comments/{id}/report-handle` | 举报处理 `{approved: bool}`（true 下架评论，false 驳回恢复） | Bearer（管理） |

管理错误码：`180401` 无权限 / `180402` 目标不存在 / `180403` 非法状态变更 / `140403` 章节已禁用。

### 管理端统计与配置（T-A-14~16）

仅管理员（role=3）可用；越权返回 `180401`。分类/标签**写操作**（POST / PUT / DELETE）与**列表**（GET `/api/admin/categories`、`GET /api/admin/tags`）均由 admin 服务 requireAdmin 处理；公开 GET `/api/categories`、`/api/tags`（book 服务，客户端浏览用）无需鉴权且不返回 status/sort_order。

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/stats/overview` | 仪表盘统计：书籍/用户/评论数、DAU 近似（当日登录 ∪ 当日搜索的去重用户）、热门书籍（复用 `/api/search/hot`）、热门搜索词（搜索日志聚合） | Bearer（管理） |
| GET | `/api/admin/categories` | 分类列表（全量，含 status/sort_order） | Bearer（管理） |
| GET | `/api/admin/tags` | 标签列表（全量，含 status） | Bearer（管理） |
| GET | `/api/admin/audit-logs` | 审计日志分页查询（筛选 user_id/action/target_type/target_id/start_time/end_time，按时间倒序） | Bearer（管理） |
| POST | `/api/categories` | 创建分类 `{name, parent_id?, sort_order?}` | Bearer（管理） |
| PUT | `/api/categories/{id}` | 更新分类（可选字段，仅更新非空项；`status` 0 禁用 1 启用） | Bearer（管理） |
| DELETE | `/api/categories/{id}` | 删除分类 | Bearer（管理） |
| POST | `/api/tags` | 创建标签 `{name, lang?}`（缺省 `zh-CN`） | Bearer（管理） |
| PUT | `/api/tags/{id}` | 更新标签（`name`/`lang`/`status`） | Bearer（管理） |
| DELETE | `/api/tags/{id}` | 删除标签 | Bearer（管理） |

## 搜索与推荐

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/search` | 搜索 `q` + 分页 + `lang` | 无 |
| GET | `/api/search/hot` | 热门书籍榜（BookDoc 列表） | 无 |
| GET | `/api/search/hot-keywords` | 热搜词榜 `{list:[{keyword,count}]}`，搜索日志聚合 TOP 10 | 无 |
| GET | `/api/search/suggest` | 搜索建议 `q` 前缀补全 `{keywords:[...]}`，搜索日志聚合 TOP 10，缓存 1s | 无 |
| POST | `/api/search/index/{book_id}` | 重建单本书搜索索引 | Bearer（管理） |
| DELETE | `/api/search/index/{book_id}` | 删除单本书搜索索引 | Bearer（管理） |
| GET | `/api/recommend` | 推荐 `strategy=hot|new` + `page_size` + `lang` | 无 |

## 支付（T-P-01~18 已完成）

业务码段 19xxxx；金额一律整数分（`amount`）。订单状态：`0 待支付 1 已支付 2 失败 3 已关闭`。支付成功 → `novel_user.vip_expires_at` 续期（可叠加）。

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| POST | `/api/payments/order` | 创建 VIP 订单 `{plan, lang}` → `{order_no, amount, currency, checkout_url, provider}`；同 user+套餐未支付订单幂等复用 | Bearer |
| GET | `/api/payments/order/{order_no}` | 查订单状态（仅本人；未支付超 15 分钟自动关闭为 3） | Bearer |
| GET | `/api/payments/plans` | 公开 VIP 套餐列表（仅 status=1，sort 升序；DB 套餐表优先，回退内置默认） | 无 |
| GET | `/api/payments/vip-status` | 当前用户 VIP 状态 → `{active, vip_expires_at?}` | Bearer |
| GET | `/api/payments/methods?lang=` | 支付方式列表（enabled 且 lang/region 匹配，sort 升序） | 无 |
| POST | `/api/payments/webhook/{provider}` | 渠道回调（stripe / nowpayments / razorpay / komoju / portone / mercadopago / xendit / paypal / alipay / wechatpay_global / unionpay，验签在内部；回调金额=订单金额强校验，幂等 settle） | 无 |

### 渠道与 config 键（T-P-19~20）

provider 行由 admin「支付方式」页创建，config 键 AES-GCM 加密存储；语言路由由行的 lang/region 决定。

| 渠道码 | 语言路由 | config 键 | 验签方式 |
| :--- | :--- | :--- | :--- |
| stripe | * | 全局 env STRIPE_SECRET_KEY / STRIPE_WEBHOOK_SECRET（config 键 `currency` 可覆盖币种） | Stripe-Signature |
| np_usdt | * | 全局 env NP_API_KEY / NP_IPN_SECRET（config 键 `coin`） | X-Nowpayments-Sig HMAC-SHA512 |
| razorpay | hi | `key_id` `key_secret` `webhook_secret` | X-Razorpay-Signature HMAC-SHA256 |
| komoju | ja | `api_key` `webhook_secret` | X-KOMOJU-SIGNATURE HMAC-SHA256 |
| portone | ko | `api_secret` `webhook_token` | X-IAMPORT-TOKEN 比对 |
| mercadopago | pt-BR | `access_token` | IPN 无签名，access_token 回查 GET /v1/payments/{id} |
| xendit | id / th / vn | `api_key` `callback_token` | X-CALLBACK-TOKEN 比对 |
| paypal | * | `client_id` `client_secret` `webhook_id`（可选 `base_url` 沙箱） | 官方 verify-webhook-signature API 验签（复用 OAuth access_token） |
| alipay | zh-CN | `app_id` `merchant_private_key` `alipay_public_key` `notify_url`（可选 `base_url` 沙箱） | 表单 notify RSA2 验签（沙箱可测） |
| wechatpay_global | *（国际版） | `app_id` `mch_id` `merchant_serial_no` `merchant_private_key` `platform_public_key` `apiv3_key` `notify_url`（可选 `base_url` 区域/沙箱覆盖） | 平台公钥 RSA-SHA256 验签 + apiv3_key AES-GCM 解密 resource |
| unionpay | zh-CN | `mer_id` `sign_cert_id` `merchant_private_key` `unionpay_public_key` `notify_url`（可选 `front_url` `base_url` 沙箱覆盖） | 字典序 RSA-SHA256 签名验签（notify 表单，应答 `success`） |

zh-CN 支付宝已实现（沙箱可配置）；微信支付国际版（wechatpay_global，HK API v3 H5）已实现（2026-08-29，商户申请后配置 4 个密钥 + apiv3_key + 真实 notify_url 即可用）；国内微信支付需企业商户号，未实现。

### 支付管理（T-P-09~13，全部 requireAdmin → 180401）

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/payments/admin/orders` | 订单分页/按用户/方式/状态/时间筛选 | Bearer（管理） |
| GET | `/api/payments/admin/order-stats` | 流水汇总统计 | Bearer（管理） |
| GET | `/api/payments/admin/providers` | 支付方式列表（含密钥配置） | Bearer（管理） |
| POST | `/api/payments/admin/providers` | 创建支付方式（config 密钥 AES-GCM 加密存储） | Bearer（管理） |
| PUT | `/api/payments/admin/providers/{id}` | 更新支付方式 | Bearer（管理） |
| DELETE | `/api/payments/admin/providers/{id}` | 删除支付方式 | Bearer（管理） |
| PATCH | `/api/payments/admin/providers/{id}/toggle` | 启用/停用支付方式 | Bearer（管理） |
| GET | `/api/payments/admin/plans` | VIP 套餐列表 | Bearer（管理） |
| POST | `/api/payments/admin/plans` | 创建 VIP 套餐（时长/价格/币种/标签） | Bearer（管理） |
| PUT | `/api/payments/admin/plans/{id}` | 更新 VIP 套餐 | Bearer（管理） |
| DELETE | `/api/payments/admin/plans/{id}` | 删除 VIP 套餐 | Bearer（管理） |

错误码：`190401 PAYMENT_CREATE_FAILED`、`190402 ORDER_NOT_FOUND`、`190403 AMOUNT_MISMATCH`、`190404 PAYMENT_PENDING`、`190405 PROVIDER_DISABLED`。

## 限流

按 IP 固定窗口（`X-Forwarded-For`）：

| 路径 | 限额 |
| :--- | :--- |
| `/api/users/login` | 10 次/分钟 |
| `/api/comments`（发布） | 10 次/分钟 |
| `/api/comments/{id}/report` | 5 次/分钟 |
| `/api/search` | 10 次/分钟 |
| `/api/payments/webhook/{provider}` | 30 次/分钟 |

超限返回 `140429 TOO_MANY_REQUESTS`。

## 客户端示例

```bash
curl -H "X-Api-Version: v1" http://localhost:8000/api/books?page=1&page_size=20
curl -X POST -H "X-Api-Version: v1" -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"demo123"}' http://localhost:8000/api/users/login
```
