// 与 kratos backend proto JSON 对齐的模型。
// 注意：后端 int64 字段序列化为 JSON 字符串，统一用 asStr 兼容 string/num。

String asStr(dynamic v) {
  if (v == null) return '';
  if (v is String) return v;
  return v.toString();
}

int asInt(dynamic v, [int fallback = 0]) {
  if (v == null) return fallback;
  if (v is int) return v;
  return int.tryParse(v.toString()) ?? fallback;
}

class AdminUser {
  final String id;
  final String username;
  final String nickname;
  final int role;
  final int status;

  AdminUser.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        username = asStr(j['username']),
        nickname = asStr(j['nickname']),
        role = asInt(j['role']),
        status = asInt(j['status']);
}

class LoginResult {
  final String accessToken;
  final String refreshToken;
  final AdminUser user;

  LoginResult.fromJson(Map<String, dynamic> j)
      : accessToken = asStr(j['accessToken']),
        refreshToken = asStr(j['refreshToken']),
        user = AdminUser.fromJson((j['user'] ?? {}) as Map<String, dynamic>);
}

/// 书籍（BookReply）。status: 1 上架 0 下架。
class Book {
  final String id;
  final String lang;
  final String title;
  final String author;
  final String summary;
  final String cover;
  final int isVip;
  final int status;

  Book.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        lang = asStr(j['lang']),
        title = asStr(j['title']),
        author = asStr(j['author']),
        summary = asStr(j['summary']),
        cover = asStr(j['cover']),
        isVip = asInt(j['isVip']),
        status = asInt(j['status']);
}

/// 章节（ChapterReply）。status: 1 启用 0 禁用。
class Chapter {
  final String id;
  final String bookId;
  final int chapterNo;
  final String title;
  final int wordCount;
  final int isVip;
  final int status;
  final String createdAt;

  Chapter.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        bookId = asStr(j['bookId']),
        chapterNo = asInt(j['chapterNo']),
        title = asStr(j['title']),
        wordCount = asInt(j['wordCount']),
        isVip = asInt(j['isVip']),
        status = asInt(j['status']),
        createdAt = asStr(j['createdAt']);
}

/// 用户管理（UserReply）。status: 1 正常 0 封禁；role: 1 读者 2 作者 3 管理员。
class User {
  final String id;
  final String username;
  final String nickname;
  final String email;
  final int role;
  final int status;
  final String createdAt;

  User.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        username = asStr(j['username']),
        nickname = asStr(j['nickname']),
        email = asStr(j['email']),
        role = asInt(j['role']),
        status = asInt(j['status']),
        createdAt = asStr(j['createdAt']);
}

/// 分类（CategoryReply）。status: 1 启用 0 禁用；parentId 0=一级。
class Category {
  final String id;
  final String name;
  final String parentId;
  final int sortOrder;
  final int status;

  Category.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        name = asStr(j['name']),
        parentId = asStr(j['parentId']),
        sortOrder = asInt(j['sortOrder']),
        status = asInt(j['status']);
}

/// 标签（TagReply）。status: 1 启用 0 禁用。
class Tag {
  final String id;
  final String name;
  final String lang;
  final int status;

  Tag.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        name = asStr(j['name']),
        lang = asStr(j['lang']),
        status = asInt(j['status']);
}

/// 仪表盘统计（GetStatsReply）。
class StatsData {
  final int bookCount;
  final int userCount;
  final int commentCount;
  final int dau;
  final List<HotBook> hotBooks;
  final List<HotKeyword> hotKeywords;

  StatsData.fromJson(Map<String, dynamic> j)
      : bookCount = asInt(j['bookCount']),
        userCount = asInt(j['userCount']),
        commentCount = asInt(j['commentCount']),
        dau = asInt(j['dau']),
        hotBooks = ((j['hotBooks'] ?? []) as List)
            .map((e) => HotBook.fromJson((e as Map).cast<String, dynamic>()))
            .toList(),
        hotKeywords = ((j['hotKeywords'] ?? []) as List)
            .map((e) =>
                HotKeyword.fromJson((e as Map).cast<String, dynamic>()))
            .toList();
}

class HotBook {
  final String bookId;
  final String title;
  final int hot;

  HotBook.fromJson(Map<String, dynamic> j)
      : bookId = asStr(j['bookId']),
        title = asStr(j['title']),
        hot = asInt(j['hot']);
}

class HotKeyword {
  final String keyword;
  final int count;

  HotKeyword.fromJson(Map<String, dynamic> j)
      : keyword = asStr(j['keyword']),
        count = asInt(j['count']);
}

/// 支付方式（ProviderReply）。configConfigured: 密钥是否已配置（不返回明文）。
class PaymentProvider {
  final String id;
  final String code;
  final String lang;
  final String region;
  final int enabled;
  final int sort;
  final bool configConfigured;

  PaymentProvider.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        code = asStr(j['code']),
        lang = asStr(j['lang']),
        region = asStr(j['region']),
        enabled = asInt(j['enabled']),
        sort = asInt(j['sort']),
        configConfigured = j['configConfigured'] == true;
}

/// 流水（OrderItem）。amount 单位分；status: 0待支付 1已付 2失败 3关闭。
class PaymentOrder {
  final String orderNo;
  final String userId;
  final int amount;
  final String currency;
  final String provider;
  final int status;
  final String txId;
  final String paidAt;
  final String createdAt;

  PaymentOrder.fromJson(Map<String, dynamic> j)
      : orderNo = asStr(j['orderNo']),
        userId = asStr(j['userId']),
        amount = asInt(j['amount']),
        currency = asStr(j['currency']),
        provider = asStr(j['provider']),
        status = asInt(j['status']),
        txId = asStr(j['txId']),
        paidAt = asStr(j['paidAt']),
        createdAt = asStr(j['createdAt']);
}

/// 流水汇总（OrderStatsReply）。amount 已付总金额（分）。
class OrderStats {
  final int total;
  final int paid;
  final int pending;
  final int failed;
  final int closed;
  final int amount;

  OrderStats.fromJson(Map<String, dynamic> j)
      : total = asInt(j['total']),
        paid = asInt(j['paid']),
        pending = asInt(j['pending']),
        failed = asInt(j['failed']),
        closed = asInt(j['closed']),
        amount = asInt(j['amount']);
}

/// VIP 套餐（PlanReply）。amount 单位分；status: 1 启用 0 禁用。
class VipPlan {
  final String id;
  final String planCode;
  final int days;
  final int amount;
  final String currency;
  final String label;
  final int sort;
  final int status;

  VipPlan.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        planCode = asStr(j['planCode']),
        days = asInt(j['days']),
        amount = asInt(j['amount']),
        currency = asStr(j['currency']),
        label = asStr(j['label']),
        sort = asInt(j['sort']),
        status = asInt(j['status']);
}

/// 审计日志（AuditLogItem）。
class AuditLog {
  final String id;
  final String userId;
  final String action;
  final String targetType;
  final String targetId;
  final String detail;
  final String ip;
  final String userAgent;
  final String createdAt;

  AuditLog.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        userId = asStr(j['userId']),
        action = asStr(j['action']),
        targetType = asStr(j['targetType']),
        targetId = asStr(j['targetId']),
        detail = asStr(j['detail']),
        ip = asStr(j['ip']),
        userAgent = asStr(j['userAgent']),
        createdAt = asStr(j['createdAt']);
}

/// 评论（CommentReply）。status: 1 正常 0 下架 2 举报待审。
class Comment {
  final String id;
  final String bookId;
  final String chapterId;
  final String userId;
  final String parentId;
  final String content;
  final int likeCount;
  final int reportCount;
  final int status;
  final String createdAt;

  Comment.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        bookId = asStr(j['bookId']),
        chapterId = asStr(j['chapterId']),
        userId = asStr(j['userId']),
        parentId = asStr(j['parentId']),
        content = asStr(j['content']),
        likeCount = asInt(j['likeCount']),
        reportCount = asInt(j['reportCount']),
        status = asInt(j['status']),
        createdAt = asStr(j['createdAt']);
}
