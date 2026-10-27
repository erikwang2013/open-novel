# apps — 多端前端目录

本目录存放 Open Novel 平台的全部客户端实现，各端共用同一套后端 API（Go-Kratos gRPC / HTTP 网关）。

```
apps/
└─ client/        # 客户端（C 端），即本目录
   └─ flutter/    # Flutter 全平台应用
```

## flutter — Flutter 全平台

- 目标平台：Web（浏览器）、Desktop（Windows / macOS / Linux）、Mobile（iOS / Android）
- 多语言：`intl` + ARB 资源文件（`assets/i18n/*.arb`），支持 12+ 语种动态切换
- 网络：`dio` 请求库，统一封装鉴权、错误码与重试
- 状态管理 / 路由：按 Flutter 生态标准方案（Riverpod / Bloc 等，见各端内说明）

```bash
cd flutter && flutter pub get && flutter run
```

## 目录约定

- 每个端独立成目录，自带依赖锁文件（`pubspec.lock`）
- 多端共用后端 API（前缀 `/api`），接口定义（proto）维护在 `kratos/backend/api/`
- 各端内部结构说明由各端自己的 README 负责：
  - [flutter/README.md](flutter/README.md) — 运行、构建、API_BASE_URL、l10n 使用说明
