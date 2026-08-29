# admin — 管理端

Open Novel 运营/管理后台（Flutter Web，B 端），与客户端共用同一套后端 API（`kratos/backend/api/` 的 proto 定义）。

```bash
cd apps/admin && flutter pub get && flutter run -d chrome
```

- 构建：`flutter build web`
- 后端地址配置同 Flutter 客户端（`--dart-define=API_BASE_URL=...`）
- 管理端功能：内容审核、用户管理、数据统计（仪表盘 / DAU / 榜单 / 行为分析 /api/stats/behavior）、配置管理（分类标签）、机器翻译工作流（DeepL，/api/admin/translate/*，管理端「翻译」页 + 人工编辑）、审计日志查询（/api/admin/audit-logs）
