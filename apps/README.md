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

## 快速开始

### 一键安装

```bash
bash scripts/install.sh
```

在项目根目录运行上述命令，即可完成环境检查（Docker / Go 1.22+ / Flutter，缺失时给出安装提示）、依赖栈启动（MySQL / Redis / OpenSearch，等待 MySQL 就绪）与后端 / 前端启动提示。

### 安装说明

前置依赖：Docker（含 Compose 插件）、Go 1.22+、Flutter 3.x。

- 一键方式：`bash scripts/install.sh`
- 手动方式：

  ```bash
  docker compose up -d
  cd kratos/backend && go mod tidy && go run ./cmd/server   # 后端 HTTP :8000 / gRPC :9000
  cd apps/client/flutter && flutter pub get && flutter run -d chrome   # 客户端
  ```

### 使用说明

- 客户端：`cd apps/client/flutter && flutter pub get && flutter run -d chrome`
- 管理端：`cd apps/admin && flutter pub get && flutter run -d chrome`（端口由 Flutter 随机分配并打印在控制台，可用 `--web-port` 固定）
- HarmonyOS：用 DevEco Studio 打开 `apps/client/harmonyos` 运行
- 三个端共用同一套后端 API（默认连 `http://localhost:8000`，接口前缀 `/api`，版本头 `X-Api-Version: v1`）；连其他后端：`flutter run -d chrome --dart-define=API_BASE_URL=http://<host>:8000`
- 常见问题：端口占用用 `lsof -i :8000` 排查；数据库连接失败先 `docker compose ps` 确认 mysql 为 healthy

## 目录约定

- 每个端独立成目录，自带依赖锁文件（`pubspec.lock` / `oh-package-lock.json5`）
- 多端共用后端 API（前缀 `/api`），接口定义（proto）维护在 `kratos/backend/api/`
