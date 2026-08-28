# Open Novel Flutter 客户端

Open Novel 的多端前端之一：Flutter 全平台应用（Web / Desktop / Mobile），多语言支持，共用后端 Go-Kratos API（前缀 `/api/v1`，默认 `http://localhost:8000`）。

## 环境要求

- Flutter SDK 3.x（含 Dart 3.x），见 [Flutter 安装文档](https://docs.flutter.dev/get-started/install)
- 目标平台工具链（按需）：Chrome（Web 调试）、Android Studio / SDK（Android）、Linux/macOS/Windows 桌面工具链
- 后端依赖栈：`docker compose up -d`（MySQL / Redis / OpenSearch，见根 README）

## 安装依赖

```bash
cd apps/flutter
flutter pub get
```

## 本地运行

默认直接连本机后端 `http://localhost:8000`，无需任何配置（`lib/api/api_client.dart` 中 `String.fromEnvironment('API_BASE_URL', defaultValue: 'http://localhost:8000')`）。

```bash
# 先启动后端（根目录）
cd kratos/backend && go run ./cmd/server

# 再跑 Flutter（Web 调试）
cd apps/flutter && flutter pub get && flutter run -d chrome
```

### 修改 API_BASE_URL

通过 `--dart-define` 覆盖，指向局域网/远程后端：

```bash
flutter run -d chrome --dart-define=API_BASE_URL=http://192.168.1.10:8000
flutter build web --dart-define=API_BASE_URL=https://api.example.com
```

## 构建各平台

```bash
flutter build web                 # Web 静态站点（输出 build/web）
flutter build apk                 # Android APK
flutter build linux               # Linux 桌面（输出 build/linux/x64/release/bundle）
flutter build macos               # macOS（需 macOS 环境）
flutter build windows             # Windows（需 Windows 环境）
flutter build ipa                 # iOS（需 macOS + Xcode）
```

## 多语言（l10n）

- ARB 资源位于 `lib/l10n/`：`app_zh.arb`（模板）与 `app_en.arb` 等。
- `pubspec.yaml` 已启用 `generate: true`，`l10n.yaml` 配置 `arb-dir: lib/l10n`。
- 新增/修改翻译后运行 `flutter gen-l10n` 重新生成 `app_localizations*.dart`（`flutter run` / `flutter build` 时也会自动生成），并重启应用生效。新增语种时复制 `app_zh.arb` 为对应 `app_xx.arb` 并在 `MaterialApp` 的 `supportedLocales` 中登记。
