# 支付渠道上线运营 Checklist（2026-08-29）

> 11 个渠道已接入后端（T-P-01~22），但**全部渠道上线前必须**：① 配置真实商户密钥 ② 替换 notify_url 占位符 ③ 在管理端创建 provider 行并启用 ④ 真实环境联调。本清单按渠道列出商户后台操作与验证点。

## 一、通用上线步骤（每个渠道）

1. 商户平台申请/配置密钥 → 填入管理端「支付方式」页创建 provider 行（config 键见下表，AES-GCM 加密存储）
2. `notify_url` 改为真实回调地址：`https://<你的域名>/api/payments/webhook/{provider}`（当前 alipay / wechatpay_global 为 `https://example.com/...` 占位符，**必须替换**）
3. 管理端启用渠道（enabled=1，按 lang/region/sort 控制客户端展示）
4. 沙箱/测试环境联调：下单 → 支付 → 回调 → 验签 → settle → VIP 到期时间变化
5. 验证幂等：重复回调不重复续期；金额不符回调被拒（190403）

## 二、渠道配置键与商户后台开关

| 渠道 | config 键 | 商户后台要做的事 |
| :--- | :--- | :--- |
| stripe | 全局 env `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET`（config `currency` 可覆盖） | **开启 Apple Pay / Google Pay**（Dashboard → Payment methods → Wallets）；按需开启 Klarna / iDEAL / SEPA / Bancontact 等（Payments → Payment methods，逐一激活）；Webhook 端点注册到 `/api/payments/webhook/stripe`，取 signing secret |
| np_usdt | 全局 env `NP_API_KEY` / `NP_IPN_SECRET`（config `coin` 选币种） | NOWPayments 后台开启 IPN 回调；确认币种与订单币种一致 |
| razorpay | `key_id` `key_secret` `webhook_secret` | 印度商户：激活卡支付；Webhook 注册 |
| komoju | `api_key` `webhook_secret` | 日本商户：按需开启 PayPay / 银行转账 / 便利店（KOMOJU 收银台选项） |
| portone | `api_secret` `webhook_token` | 韩国商户：按需开启 KakaoPay / Naver Pay / Toss（PortOne 收银台选项） |
| mercadopago | `access_token` | 拉美商户：按需开启 **Pix** / Boleto / OXXO（Mercado Pago 收银台选项） |
| xendit | `api_key` `callback_token` | 东南亚商户：按需开启 **GCash / GrabPay** / 银行转账（Xendit 收银台选项） |
| paypal | `client_id` `client_secret` `webhook_id`（可选 `base_url` 沙箱） | PayPal 后台创建 Webhook 到 `/api/payments/webhook/paypal`，取 Webhook ID |
| alipay | `app_id` `merchant_private_key` `alipay_public_key` `notify_url`（可选 `base_url` 沙箱） | CN 商户资质；沙箱用 openapi.alipaydev.com；**替换 notify_url 占位符** |
| wechatpay_global | `app_id` `mch_id` `merchant_serial_no` `merchant_private_key` `platform_public_key` `apiv3_key` `notify_url`（可选 `base_url` 区域覆盖） | 国际商户（HK API v3 H5）；配置 4 个密钥 + apiv3_key；**替换 notify_url 占位符** |
| unionpay | `mer_id` `sign_cert_id` `merchant_private_key` `unionpay_public_key` `notify_url`（可选 `front_url` `base_url` 沙箱） | CN 商户资质；签名证书 + 银联验签公钥；开发联调环境 `http://58.246.226.99/UpopWeb/api/Pay.action` |

## 三、上线联调验证点（代码已知不确定项）

以下为实现时标注的技术不确定点，**联调时必须逐一验证**，不符则需小改后端（均有 ponytail 注释定位）：

| 渠道 | 验证点 | 不符时改哪里 |
| :--- | :--- | :--- |
| unionpay | ① 网关是否接受 GET query 跳转（银联规范为表单 POST）② 空值参数是否参与验签（当前按「所有数据元」含空值）③ `signMethod=01` 摘要为 SHA-256 | ① 需在 CreateOrder 流程改 form HTML 下发（provider_unionpay.go CreateCheckout）② unionpaySignString 空值跳过 ③ 常量一处 |
| wechatpay_global | 平台公钥验签 + apiv3_key AES-GCM 解密 resource；notify_url 必须为公网 HTTPS | provider_wechatpay_global.go |
| alipay | RSA2 验签（沙箱可测）；notify_url 必须为公网 HTTPS | provider_alipay.go |

## 四、客户端验证

- 支付方式列表随 lang 返回（`GET /api/payments/methods?lang=`），enabled 渠道才出现
- 下单 → checkout_url 跳转 → 支付 → 支付结果页三态轮询（3s）
- VIP 生效：`GET /api/payments/vip-status` → active=true，`vip_expires_at` 顺延
- 重复支付同一套餐：未支付订单幂等复用，已支付后续期叠加

## 五、未接入清单（按需开启，无需代码）

| 能力 | 开启方式 |
| :--- | :--- |
| Apple Pay / Google Pay | Stripe 后台 Payment methods → Wallets（**推荐优先做，全球转化率收益最大**） |
| Klarna / iDEAL / SEPA | Stripe 后台 Payment methods |
| Pix / Boleto / OXXO | Mercado Pago 后台收银台选项 |
| KakaoPay / Naver Pay | PortOne 后台收银台选项 |
| PayPay | KOMOJU 后台收银台选项 |
| GCash / GrabPay | Xendit 后台收银台选项 |
| 国内微信支付 | 需 CN 企业商户号，当前未接入（需新增 provider 开发） |
