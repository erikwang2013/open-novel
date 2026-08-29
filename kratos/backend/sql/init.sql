-- ============================================================
-- Open Novel 平台建库建表脚本（Phase 1）
-- 库名 novel，表前缀 novel_，统一 utf8mb4 / InnoDB
-- 说明：不加外键约束，关联完整性由应用层保证（便于分库与多写）
-- ============================================================

CREATE DATABASE IF NOT EXISTS novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE novel;

-- ------------------------------------------------------------
-- 用户表：账号、多语言昵称、头像、密码哈希(bcrypt)、状态
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_user (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username      VARCHAR(64)  NOT NULL COMMENT '登录账号',
  email         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
  password_hash VARCHAR(255) NOT NULL COMMENT 'bcrypt 密码哈希',
  nickname      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '默认昵称',
  nickname_i18n JSON         NULL COMMENT '多语言昵称覆盖 {"en":"..","ja":".."}',
  avatar        VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像 URL',
  status        TINYINT      NOT NULL DEFAULT 1 COMMENT '0禁用 1正常',
  role          TINYINT      NOT NULL DEFAULT 1 COMMENT '1普通用户 2作者 3管理员 4运营',
  last_login_at DATETIME     NULL COMMENT '最近登录时间',
  vip_expires_at DATETIME    NULL COMMENT 'VIP 到期时间（NULL=非会员；支付成功后激活）',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_username (username),
  UNIQUE KEY uk_email (email),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 存量开发库升级（新库已含 vip_expires_at 列，旧库手动执行）：
-- ALTER TABLE novel_user ADD COLUMN vip_expires_at DATETIME NULL COMMENT 'VIP 到期时间' AFTER last_login_at;

-- ------------------------------------------------------------
-- 书籍表：书名、作者、简介、封面、状态
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_book (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title      VARCHAR(255) NOT NULL COMMENT '书名（原语言）',
  author     VARCHAR(128) NOT NULL DEFAULT '' COMMENT '作者',
  summary    TEXT         NULL COMMENT '简介',
  cover      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '封面 URL',
  lang       CHAR(5)      NOT NULL DEFAULT 'zh-CN' COMMENT '原语言',
  is_vip     TINYINT      NOT NULL DEFAULT 0 COMMENT '0免费 1付费',
  status     TINYINT      NOT NULL DEFAULT 1 COMMENT '0下架 1连载 2完结',
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_status (status),
  KEY idx_author (author)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='书籍表';

-- ------------------------------------------------------------
-- 书籍多语言翻译表：每本书每种语言一条
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_book_translation (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  lang       CHAR(5)         NOT NULL COMMENT 'zh-CN / en / ja ...',
  title      VARCHAR(255)    NOT NULL COMMENT '本地化书名',
  summary    TEXT            NULL COMMENT '本地化简介',
  cover      VARCHAR(512)    NOT NULL DEFAULT '' COMMENT '本地化封面',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_book_lang (book_id, lang)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='书籍多语言翻译表';

-- ------------------------------------------------------------
-- 章节表：book_id + 序号、字数、状态
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_chapter (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  chapter_no INT UNSIGNED    NOT NULL COMMENT '章序号（从 1 开始）',
  title      VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '章节标题',
  word_count INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '字数',
  is_vip     TINYINT         NOT NULL DEFAULT 0 COMMENT '0免费 1付费',
  status     TINYINT         NOT NULL DEFAULT 1 COMMENT '0删除 1发布 2草稿',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_book_chapter (book_id, chapter_no),
  KEY idx_book_status (book_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='章节表';

-- ------------------------------------------------------------
-- 章节正文表：每章每语言一条
-- ponytail: 先用 MEDIUMTEXT 全文存储，章节超大需分块时再引入 chunk 表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_chapter_content (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  chapter_id BIGINT UNSIGNED NOT NULL COMMENT '章节 ID',
  lang       CHAR(5)         NOT NULL COMMENT 'zh-CN / en / ja ...',
  content    MEDIUMTEXT      NOT NULL COMMENT '正文全文',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_chapter_lang (chapter_id, lang)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='章节正文表';

-- ------------------------------------------------------------
-- 分类表：支持二级分类（parent_id）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_category (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name       VARCHAR(64)     NOT NULL COMMENT '分类名',
  parent_id  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父分类 ID，0=一级',
  sort_order INT             NOT NULL DEFAULT 0 COMMENT '排序',
  status     TINYINT         NOT NULL DEFAULT 1 COMMENT '0禁用 1启用',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_parent (parent_id, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分类表';

-- ------------------------------------------------------------
-- 书籍-分类关联表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_book_category (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  book_id     BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  category_id BIGINT UNSIGNED NOT NULL COMMENT '分类 ID',
  created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_book_cat (book_id, category_id),
  KEY idx_category (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='书籍-分类关联表';

-- ------------------------------------------------------------
-- 标签表：多语言标签（name+lang 唯一）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_tag (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name       VARCHAR(64) NOT NULL COMMENT '标签名',
  lang       CHAR(5)     NOT NULL DEFAULT 'zh-CN' COMMENT '语言',
  status     TINYINT     NOT NULL DEFAULT 1 COMMENT '0禁用 1启用',
  created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_name_lang (name, lang)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签表';

-- ------------------------------------------------------------
-- 书籍-标签关联表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_book_tag (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  tag_id     BIGINT UNSIGNED NOT NULL COMMENT '标签 ID',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_book_tag (book_id, tag_id),
  KEY idx_tag (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='书籍-标签关联表';

-- ------------------------------------------------------------
-- 评论表：书籍/章节评论，支持楼中楼回复
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_comment (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  chapter_id BIGINT UNSIGNED NULL COMMENT '章节 ID，NULL=书籍评论',
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '评论人',
  parent_id  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '回复的父评论 ID，0=根评论',
  content    TEXT            NOT NULL COMMENT '评论内容',
  like_count INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '点赞数',
  report_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '举报次数',
  status     TINYINT         NOT NULL DEFAULT 1 COMMENT '0删除 1正常 2举报待审',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_book_chapter (book_id, chapter_id),
  KEY idx_user (user_id),
  KEY idx_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- ------------------------------------------------------------
-- 点赞表：多态目标（书/评论/章节）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_like (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     BIGINT UNSIGNED NOT NULL COMMENT '点赞用户',
  target_type TINYINT         NOT NULL COMMENT '1书 2评论 3章节',
  target_id   BIGINT UNSIGNED NOT NULL COMMENT '目标 ID',
  created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_target (user_id, target_type, target_id),
  KEY idx_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞表';

-- ------------------------------------------------------------
-- 收藏表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_favorite (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_book (user_id, book_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏表';

-- ------------------------------------------------------------
-- 书架表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_bookshelf (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  sort_order INT             NOT NULL DEFAULT 0 COMMENT '书架内排序',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_book (user_id, book_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='书架表';

-- ------------------------------------------------------------
-- 阅读进度表：book_id + chapter_id + user_id 唯一
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_reading_progress (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  chapter_id BIGINT UNSIGNED NOT NULL COMMENT '章节 ID',
  position   INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '章内位置（字符偏移）',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_book_chapter_user (book_id, chapter_id, user_id),
  KEY idx_user_book (user_id, book_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='阅读进度表';

-- ------------------------------------------------------------
-- 搜索日志表：搜索词分析 / 热词
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_search_log (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id      BIGINT UNSIGNED NULL COMMENT '用户 ID（未登录为 NULL）',
  keyword      VARCHAR(255) NOT NULL COMMENT '搜索词',
  lang         CHAR(5)      NOT NULL DEFAULT 'zh-CN' COMMENT '搜索语言',
  result_count INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '结果数',
  ip           VARCHAR(45)  NOT NULL DEFAULT '' COMMENT '客户端 IP',
  created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_keyword (keyword),
  KEY idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='搜索日志表';

-- ------------------------------------------------------------
-- 推荐日志表：推荐策略效果分析
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_recommend_log (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '推荐的书',
  strategy   VARCHAR(32)     NOT NULL COMMENT '策略：ai/hot/new/rule',
  rank_no    INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '推荐位次',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_user_time (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='推荐日志表';

-- ------------------------------------------------------------
-- 阅读事件日志表：保存进度时顺带记录，行为分析数据源
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_reading_log (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  book_id    BIGINT UNSIGNED NOT NULL COMMENT '书籍 ID',
  chapter_id BIGINT UNSIGNED NOT NULL COMMENT '章节 ID',
  lang       CHAR(5)         NOT NULL DEFAULT 'zh-CN' COMMENT '阅读语言',
  position   INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '章内位置（字符偏移）',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_user_time (user_id, created_at),
  KEY idx_book_time (book_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='阅读事件日志表';

-- ------------------------------------------------------------
-- 支付订单表：单章/整书购买
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_payment_order (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_no   VARCHAR(64)    NOT NULL COMMENT '业务订单号',
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  book_id    BIGINT UNSIGNED NULL COMMENT '购买的书（非书籍购买为 NULL）',
  amount     DECIMAL(10,2)  NOT NULL COMMENT '金额',
  currency   CHAR(3)        NOT NULL DEFAULT 'CNY' COMMENT '币种',
  channel    VARCHAR(32)    NOT NULL DEFAULT '' COMMENT '渠道：wechat/alipay/stripe',
  status     TINYINT        NOT NULL DEFAULT 0 COMMENT '0待支付 1已支付 2失败 3已退款',
  paid_at    DATETIME       NULL COMMENT '支付时间',
  created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_user (user_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单表';

-- ------------------------------------------------------------
-- 会员订单表：套餐购买与有效期
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_vip_order (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_no   VARCHAR(64)    NOT NULL COMMENT '业务订单号',
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  plan       VARCHAR(32)    NOT NULL COMMENT '套餐：monthly/quarterly/yearly',
  amount     DECIMAL(10,2)  NOT NULL COMMENT '金额',
  currency   CHAR(3)        NOT NULL DEFAULT 'CNY' COMMENT '币种',
  status     TINYINT        NOT NULL DEFAULT 0 COMMENT '0待支付 1已支付 2失败 3已退款',
  start_at   DATETIME       NULL COMMENT '会员生效时间',
  end_at     DATETIME       NULL COMMENT '会员到期时间',
  paid_at    DATETIME       NULL COMMENT '支付时间',
  created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_user (user_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员订单表';

-- ------------------------------------------------------------
-- 支付方式表：可用支付渠道（stripe/np_usdt...），按语言/地区路由
-- config 为 AES-GCM 加密的 JSON（密钥见 config.yaml payment.encrypt_key）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_payment_provider (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code       VARCHAR(32)    NOT NULL COMMENT '渠道码：stripe/np_usdt',
  lang       VARCHAR(16)    NOT NULL DEFAULT '*' COMMENT '适用语言: en 或 * 全局',
  region     VARCHAR(16)    NOT NULL DEFAULT '*' COMMENT '适用地区: US/CN 或 * 全局',
  enabled    TINYINT        NOT NULL DEFAULT 1 COMMENT '0禁用 1启用',
  sort       INT            NOT NULL DEFAULT 0 COMMENT '排序（升序）',
  config     TEXT           NULL COMMENT '加密 JSON 配置（API key/webhook secret/币种等）',
  created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_code (code),
  KEY idx_enabled_sort (enabled, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付方式表';

-- ------------------------------------------------------------
-- 支付订单表：一次支付对应一行（VIP 套餐无 plan 列，plan 由金额反查）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_payment_order (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_no   VARCHAR(64)    NOT NULL COMMENT '业务订单号',
  user_id    BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  amount     DECIMAL(10,2)  NOT NULL COMMENT '金额（分转元存储）',
  currency   CHAR(3)        NOT NULL DEFAULT 'USD' COMMENT '币种',
  provider   VARCHAR(32)    NOT NULL DEFAULT '' COMMENT '渠道码：stripe/np_usdt',
  status     TINYINT        NOT NULL DEFAULT 0 COMMENT '0待支付 1已支付 2失败 3已关闭',
  tx_id      VARCHAR(128)   NOT NULL DEFAULT '' COMMENT '支付渠道交易号',
  paid_at    DATETIME       NULL COMMENT '支付时间',
  created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_user (user_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单表';

-- ------------------------------------------------------------
-- VIP 套餐表：支付流程生效金额/天数数据源（T-P-13），表空/缺行回退内置默认
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_vip_plan (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  plan_code   VARCHAR(32)    NOT NULL COMMENT '套餐码：monthly/quarterly/yearly',
  days        INT            NOT NULL COMMENT '有效天数',
  amount_cents BIGINT       NOT NULL COMMENT '金额（分）',
  currency    CHAR(3)        NOT NULL DEFAULT 'USD' COMMENT '币种',
  label       VARCHAR(64)    NOT NULL DEFAULT '' COMMENT '展示标签',
  sort        INT            NOT NULL DEFAULT 0 COMMENT '排序（升序）',
  status      TINYINT        NOT NULL DEFAULT 1 COMMENT '0禁用 1启用',
  created_at  DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_plan_code (plan_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='VIP 套餐表';

-- ------------------------------------------------------------
-- 审计日志表：登录 / 管理操作 / 支付审计
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_audit_log (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NULL COMMENT '操作用户 ID（匿名为 NULL）',
  action     VARCHAR(64)     NOT NULL COMMENT '动作：login/pay/book_update/...',
  target_type VARCHAR(32)    NOT NULL DEFAULT '' COMMENT '对象类型：user/book/order',
  target_id  VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '对象 ID',
  detail     TEXT            NULL COMMENT '变更详情（JSON）',
  ip         VARCHAR(45)     NOT NULL DEFAULT '' COMMENT '客户端 IP',
  user_agent VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'UA',
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_user (user_id),
  KEY idx_action (action),
  KEY idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';
