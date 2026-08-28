# Open Novel API 文档

后端 HTTP 接口统一前缀 `/api`，**版本不写在 URL**，通过请求头协商：

| Header | 值 | 说明 |
| :--- | :--- | :--- |
| `X-Api-Version` | `v1` | 必填；缺失或不支持的值返回 `140426 API_VERSION_MISMATCH` |
| `Authorization` | `Bearer <accessToken>` | 登录/注册后获取；受保护接口必填 |

- Base URL：`http://<host>:8000`（开发默认 `http://localhost:8000`）
- 序列化：JSON 驼峰字段；**int64 字段（id 等）序列化为字符串**
- 分页：`page`（从 1 起）+ `page_size`，响应 `{list, total, page, pageSize}`
- 多语言：查询参数 `lang=zh|en|ja`（`zh` 默认），影响书籍标题/简介/章节内容返回的语言版本

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

## 书籍

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/books` | 书籍列表（分页 + `lang`） | 无 |
| GET | `/api/books/{id}` | 书籍详情 | 无 |
| POST | `/api/books` | 创建书籍 | Bearer（管理） |
| POST | `/api/books/{book_id}/favorite` | 收藏书籍 | Bearer |
| DELETE | `/api/books/{book_id}/favorite` | 取消收藏 | Bearer |
| GET | `/api/categories` | 分类列表 | 无 |
| GET | `/api/tags` | 标签列表 | 无 |

## 章节

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/books/{book_id}/chapters` | 章节列表（分页 + `lang`） | 无 |
| GET | `/api/chapters/{id}/content` | 章节正文（`lang`） | VIP 章节需 Bearer |
| GET | `/api/progress` | 阅读进度 | Bearer |
| GET | `/api/bookshelf` | 书架 | Bearer |
| POST | `/api/bookshelf/{book_id}` | 加入书架 | Bearer |
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

## 搜索与推荐

| 方法 | 路径 | 说明 | 鉴权 |
| :--- | :--- | :--- | :--- |
| GET | `/api/search` | 搜索 `q` + 分页 + `lang` | 无 |
| GET | `/api/search/hot` | 热门搜索词 | 无 |
| POST | `/api/search/index/{book_id}` | 重建单本书搜索索引 | Bearer（管理） |
| GET | `/api/recommend` | 推荐 `strategy=hot|new` + `page_size` + `lang` | 无 |

## 限流

按 IP 固定窗口（`X-Forwarded-For`）：

| 路径 | 限额 |
| :--- | :--- |
| `/api/users/login` | 10 次/分钟 |
| `/api/comments`（发布） | 10 次/分钟 |
| `/api/comments/{id}/report` | 5 次/分钟 |
| `/api/search` | 10 次/分钟 |

超限返回 `140429 TOO_MANY_REQUESTS`。

## 客户端示例

```bash
curl -H "X-Api-Version: v1" http://localhost:8000/api/books?page=1&page_size=20
curl -X POST -H "X-Api-Version: v1" -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"demo123"}' http://localhost:8000/api/users/login
```
