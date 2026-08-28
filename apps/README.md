# apps — 多端前端目录

本目录存放 Open Novel 平台的全部客户端实现，各端共用同一套后端 API（Go-Kratos gRPC / HTTP 网关）。

```
apps/
├─ flutter/     # Flutter 全平台应用
└─ harmonyos/   # HarmonyOS NEXT 原生应用
```

## flutter — Flutter 全平台

- 目标平台：Web（浏览器）、Desktop（Windows / macOS / Linux）、Mobile（iOS / Android）
- 多语言：`intl` + ARB 资源文件（`assets/i18n/*.arb`），支持 12+ 语种动态切换
- 网络：`dio` 请求库，统一封装鉴权、错误码与重试
- 状态管理 / 路由：按 Flutter 生态标准方案（Riverpod / Bloc 等，见各端内说明）

```bash
cd flutter && flutter pub get && flutter run
```

## harmonyos — HarmonyOS NEXT 原生应用

- 目标平台：HarmonyOS NEXT（手机 / 平板 / 智慧屏等）
- 语言框架：ArkTS + ArkUI，`hvigor` 构建
- 多语言：`ohos i18n` 资源目录，`base/element/string.json` + 各语言目录

```bash
cd harmonyos && hvigorw assembleHap
```

## 目录约定

- 每个端独立成目录，自带依赖锁文件（`pubspec.lock` / `oh-package-lock.json5`）
- 多端共用后端 API（前缀 `/api`），接口定义（proto）维护在 `kratos/backend/api/`
- 各端内部结构说明由各端自己的 README 负责：
  - [flutter/README.md](flutter/README.md) — 运行、构建、API_BASE_URL、l10n 使用说明
  - [harmonyos/README.md](harmonyos/README.md) — 构建、BASE_URL（`common/Config.ets`）、unsigned hap 说明
