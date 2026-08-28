# harmonyos — HarmonyOS NEXT 原生应用

Open Novel 的 HarmonyOS NEXT 客户端，使用 ArkTS + ArkUI 开发，与 Flutter 端共用同一套后端 API（前缀 `/api`，版本经请求头 `X-Api-Version: v1` 协商）。

## 目录结构

```
apps/client/harmonyos/
├─ AppScope/                 # 应用级配置（图标、名称）
├─ entry/                    # 主模块（HAP）
│  └─ src/main/ets/          #   ArkTS 源码
│     ├─ pages/              #   页面（书架、书籍详情、阅读器、评论、登录等）
│     └─ common/             #   公共代码（Config 配置、ApiClient 网络封装）
├─ build-profile.json5       # 构建配置（签名、产品）
├─ hvigorfile.ts             # hvigor 构建脚本
└─ oh-package.json5          # 依赖声明
```

## 构建

```bash
cd apps/client/harmonyos && ./hvigorw assembleHap
```

产物为 `.hap` 包，位于 `entry/build/` 下。

## 配置后端地址

后端地址在 `entry/src/main/ets/common/Config.ets` 的 `Config.BASE_URL` 中修改：

- HarmonyOS 模拟器：`http://10.0.2.2:8000`（需宿主防火墙放行 8000）
- 真机调试：宿主机局域网 IP，如 `http://192.168.1.100:8000`
- 发布：线上 https 域名

> 注意：unsigned hap 无发布签名，无法使用网络，真机安装需在 DevEco Studio 中自行配置签名（Project Structure → Signing Configs）。
