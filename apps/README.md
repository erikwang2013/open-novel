# apps — 多端应用目录

本目录存放 Open Novel 平台的全部前端实现，各端共用同一套后端 API（Go-Kratos gRPC / HTTP 网关）。

```
apps/
├─ client/     # 客户端多端（C 端）
│  ├─ flutter/     # Flutter 全平台（Web / Desktop / Mobile）
│  ├─ harmonyos/   # HarmonyOS NEXT 原生应用（ArkTS / ArkUI）
│  └─ README.md    # 客户端目录说明
└─ admin/      # 管理端（Flutter Web，B 端后台）
```

## client — 客户端

面向读者的多端应用，详见 [client/README.md](client/README.md)。

## admin — 管理端

面向运营/管理员的 Flutter Web 后台，详见 [admin/README.md](admin/README.md)。

## 目录约定

- 每个端独立成目录，自带依赖锁文件（`pubspec.lock` / `oh-package-lock.json5`）
- 多端共用后端 API（前缀 `/api`），接口定义（proto）维护在 `kratos/backend/api/`
