# 全球多语言小说平台 — 项目规划

全球多语言小说平台：**Go-Kratos** 后端 + **Flutter / HarmonyOS** 多端前端，支持 12+ 语种独立适配，云原生微服务架构。

---

## 一、系统整体架构（概念）

```
[客户端层]
├─ Flutter Web App（Chrome）
├─ Flutter Desktop App（Windows / Mac / Linux）
├─ Flutter Mobile App（iOS / Android）
├─ HarmonyOS NEXT App（ArkTS / ArkUI）
└── [本地化引擎: i18n 路由 + 资源动态加载]

[接入层]
├─ Nginx（TLS 终结 / 静态资源 / CDN 加速 / 反向代理）
└─ HTTP 网关 → gRPC 服务（Go-Kratos v2.x，双协议）

[服务层]
├─ User Service（用户 / 会员 / 书架）
├─ Book Service（书籍 / 章节 / 多语言翻译）
├─ Comment Service（评论 / 点赞 / 收藏）
├─ Recommendation Service（推荐引擎接口）
└─ Payment Service（支付 / 会员订单）

[存储层]
├─ MySQL（主从 + 读写分离）— 业务数据，库名 novel，表前缀 novel_
├─ Redis Cluster — 热点缓存 / 会话 / 分布式锁
└─ OpenSearch — 多语言全文检索（kuromoji / nori / ICU 分词）

[运维层]
├─ Prometheus + Grafana（监控）
├─ Loki / ELK（日志聚合）
└─ CI/CD（GitHub Actions / Jenkins）
```

---

## 二、数据库设计

**规范**：数据库名 `novel`，所有表前缀 `novel_`，统一 utf8mb4。

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 核心表清单

| 表名 | 职责 |
| :--- | :--- |
| `novel_user` | 用户表（账号、多语言昵称、头像、密码哈希、状态） |
| `novel_book` | 书籍表（书名、作者、简介、封面、状态） |
| `novel_book_translation` | 书籍多语言翻译表（`book_id + lang` 联合唯一） |
| `novel_chapter` | 章节表（book_id、序号、字数、状态） |
| `novel_chapter_content` | 章节正文表（分块/全文存储、多语言） |
| `novel_category` | 分类表 |
| `novel_book_category` | 书籍-分类关联表 |
| `novel_tag` | 标签表 |
| `novel_book_tag` | 书籍-标签关联表 |
| `novel_comment` | 评论表（书籍/章节评论） |
| `novel_like` | 点赞表 |
| `novel_favorite` | 收藏表 |
| `novel_bookshelf` | 书架表 |
| `novel_reading_progress` | 阅读进度表（book_id + chapter_id + user_id） |
| `novel_search_log` | 搜索日志表 |
| `novel_recommend_log` | 推荐日志表 |
| `novel_payment_order` | 支付订单表 |
| `novel_vip_order` | 会员订单表 |
| `novel_audit_log` | 审计日志表（登录 / 管理操作 / 支付） |

示例 DDL（关键索引与约束）：

```sql
CREATE TABLE novel_book_translation (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  book_id BIGINT UNSIGNED NOT NULL,
  lang CHAR(5) NOT NULL COMMENT 'zh-CN / en / ja ...',
  title VARCHAR(255) NOT NULL,
  summary TEXT,
  UNIQUE KEY uk_book_lang (book_id, lang)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 主从读写分离

- **写主库**：注册、下单、上传章节、评论等写操作走主库。
- **读从库**：书籍列表、搜索、阅读进度等读操作走从库（Kratos `master-replica` 数据源配置）。
- 关键读写命令须带 `;FORCE_MASTER` 或短事务直连主库，避免主从延迟导致读不到刚写入的数据。

---

## 三、核心模块设计

### 1. 多端前端目录 `apps/`

```
apps/
├─ flutter/        # Flutter 全平台（Web / Desktop / Mobile）
│  ├─ lib/
│  │  ├─ services/        # dio 请求、i18n 服务
│  │  ├─ screens/         # 书籍详情 / 阅读器 / 书架 ...
│  │  └─ l10n/            # ARB 多语言文件（12+ 语种）
│  └─ pubspec.yaml
└─ harmonyos/      # HarmonyOS NEXT 原生应用（ArkTS / ArkUI）
   ├─ entry/src/main/ets/ # 页面与组件
   └─ build-profile.json5 # hvigor 构建配置
```

两端共享同一套后端 API（HTTP 网关 + gRPC），客户端只走 HTTP/JSON。

### 2. Go-Kratos 后端服务

```go
// 多语言内容读取：缓存 → 数据库 → 回填缓存
func (s *BookService) GetBook(ctx context.Context, req *types.GetBookRequest) (*types.BookResponse, error) {
    cacheKey := fmt.Sprintf("book:%d:%s", req.Id, req.Lang)
    if v, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
        return unmarshalBook(v)
    }
    var t bookModel.BookTranslation
    if err := s.db.WithContext(ctx).Where("book_id = ? AND lang = ?", req.Id, req.Lang).First(&t).Error; err != nil {
        return nil, errors.NotFound("book.translation", "translation not found")
    }
    s.cache.Set(ctx, cacheKey, marshalBook(t), 5*time.Minute)
    return &types.BookResponse{Data: t}, nil
}
```

### 3. OpenSearch 多语言索引

```yaml
# 多语言分词器：日语 kuromoji、韩语 nori、通用 ICU normalization
index_settings:
  analysis:
    analyzer:
      multi_lang_analyzer:
        tokenizer: standard
        filter: [lowercase, kuromoji_stemmer, icu_normalization]
```

按用户语言路由对应 analyzer，同一文档以多字段（zh / en / ja...）存储便于分语种搜索。

### 4. Flutter 多语言与多端适配（要点）

- **i18n**：`flutter_localizations` + `intl`，资源为 `assets/i18n/<lang>.arb`，运行时按 locale 动态加载并切换主题/字体。
- **HTTP**：`dio` 统一封装，自动携带 `Accept-Language` 与 JWT。
- **自适应布局**：桌面端侧边栏 + 宽屏多列；移动端底部导航 + 卡片流；通过 `LayoutBuilder` / 平台判断选择布局。

```dart
class BookDetailScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(builder: (context, constraints) {
      if (constraints.maxWidth >= 900) {
        return DesktopLayout(sidebar: Sidebar(), mainContent: ContentArea());
      }
      return MobileLayout(navBar: NavBar(), content: BookInfo());
    });
  }
}
```

---

## 四、API 设计规范

- **双协议**：内部服务间 gRPC；对外 REST over HTTP 网关，统一前缀 `/api/`。
- **proto 风格**：服务/方法命名采用 PascalCase：

```proto
service BookService {
  rpc GetBook(GetBookRequest) returns (BookResponse);
}

message GetBookRequest {
  int64 id = 1;
  string lang = 2;   // 可选，缺省按 Accept-Language
}
```

- **REST 映射**：`GET /api/books/{id}?lang=zh-CN` → `BookService.GetBook`。
- **错误码规范**：业务错误使用 Kratos 错误码语义（`InvalidArgument` / `NotFound` / `Unauthenticated`），对外输出 `code + message + detail` 结构；HTTP 状态与业务码分离（`200 + code=140404`）。
- **分页约定**：所有列表接口统一 `page`（从 1 开始）+ `page_size`（默认 20，上限 100），响应携带 `total / page / page_size` 字段。

---

## 五、缓存策略

- **热点缓存**：书籍元数据、章节列表/正文入 Redis，TTL = 5 分钟；按 `book_id + lang` 分 key。
- **穿透防护**：查询不存在的数据时缓存空值（TTL 较短，如 60s）；高价值场景可选布隆过滤器前置拦截。
- **击穿防护**：热点 key 过期瞬间用分布式锁（Redis `SETNX`）仅放行一个请求回源，其余等待后读缓存。
- **雪崩防护**：缓存 TTL 加随机抖动（±30s），避免同一时间批量过期。
- **本地缓存**（可选）：单服务内 `ristretto` 缓存超高热数据（如首页榜单），配合 Redis 失效订阅清理。

---

## 六、安全体系

采用 Kratos 官方中间件组合，无第三方安全插件依赖：

| 中间件 | 职责 |
| :--- | :--- |
| `middleware/jwt` | JWT 认证 + RefreshToken 续期 |
| `middleware/ratelimit` | 令牌桶限流，防刷接口（登录/搜索/评论） |
| `middleware/recovery` | 全局 panic 恢复，防止单点故障拖垮进程 |
| `middleware/validate` | protovalidate 入参校验（信任边界） |
| `middleware/tracing` | OpenTelemetry 链路追踪 |

```go
import (
    "github.com/go-kratos/kratos/v2/middleware/jwt"
    "github.com/go-kratos/kratos/v2/middleware/ratelimit"
    "github.com/go-kratos/kratos/v2/middleware/recovery"
    "github.com/go-kratos/kratos/v2/middleware/validate"
)

httpSrv := http.NewServer(
    http.Middleware(
        recovery.Recovery(),
        ratelimit.Server(),
        jwt.Server(func(ctx context.Context) (jwt.Claims, error) { /* 解析 */ }),
    ),
)
```

**防护目标与手段**：

- **认证授权**：JWT（短时 access + RefreshToken 轮换）；RBAC 角色权限（普通用户 / 作者 / 管理员 / 运营）。
- **Web 攻击面**：SQL 注入 → 全部参数化查询 / ORM；XSS → 输出转义 + CSP 响应头；CSRF → 网关校验 `Origin/Referer` + SameSite Cookie，写操作走 Authorization Header。
- **敏感数据**：密码 bcrypt/argon2 哈希存储；手机号等字段 AES 加密；密钥托管 KMS。
- **传输安全**：全链路 HTTPS/TLS（Nginx 终结）。
- **审计日志**：管理操作、登录、支付写审计表（`novel_audit_log`），留存可追溯。

---

## 七、部署与运维

### Docker Compose 示例

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=secret123
      - MYSQL_DATABASE=novel
    ports:
      - "3307:3307"

  redis:
    image: redis:alpine
    ports:
      - "6380:6380"

  opensearch:
    image: opensearchproject/opensearch:latest
    environment:
      - discovery.type=single-node
    ports:
      - "9200:9200"

  kratos-api:
    build: ./kratos/backend
    depends_on: [mysql, redis, opensearch]
    ports:
      - "8000:8000"

  flutter-web:
    build: ./apps/client/flutter
    volumes:
      - /var/www/flutter-web:/var/www/html
```

### CI/CD（GitHub Actions）

1. Flutter Web 构建 → 静态产物 + HarmonyOS 签名打包。
2. Go-Kratos 单测 + golangci-lint + 镜像构建。
3. 部署至 Kubernetes / 阿里云 ACK，镜像滚动更新。

---

## 八、项目阶段规划与资源估算

| 阶段 | 周期 | 任务重点 | 负责人角色 |
| :--- | :--- | :--- | :--- |
| **Phase 1** | 2-3 周 | Kratos 基础服务 + MySQL 主从 / Redis / OpenSearch 集成，数据库建表 | 后端开发（Go） |
| **Phase 2** | 3-4 周 | Flutter 多语言 ARB + 多端布局；HarmonyOS 基础壳工程 | 前端开发（Flutter / ArkTS） |
| **Phase 3** | 2 周 | 安全体系落地（JWT / 限流 / 校验 / 追踪）+ 压力测试 | 安全工程师 |
| **Phase 4** | 1-2 周 | 全链路联调 + CDN（CloudFront / 阿里云 OSS）+ 监控告警 | 运维工程师 |
| **Phase 5** | 持续迭代 | AI 推荐、用户行为埋点分析、搜索优化 | 数据工程师 |

### 关键资源清单

- **开发环境**：Go 1.22+、Flutter SDK、DevEco Studio（HarmonyOS）、MySQL 8.0、Redis 7.x、OpenSearch 2.x。
- **云服务建议**：AWS EC2 (t3.large) + RDS + ElastiCache；或阿里云 ECS + RDS + Redis。
- **成本估算（首年）**：约 $5000–$10000，初期可先使用云免费额度。

---

## 九、总结与建议

1. **多语言资源管理**：ARB 文件建议用 Git LFS 管理，避免仓库膨胀。
2. **安全**：JWT 密钥与字段加密密钥一律走环境变量/KMS，不落代码仓库。
3. **缓存**：上线初期先只做 Redis 热点缓存，穿透/击穿防护按实际 QPS 渐进引入。
4. **性能**：阅读器图片/章节正文用 `flutter_cache_manager` 本地缓存，减少重复请求。
