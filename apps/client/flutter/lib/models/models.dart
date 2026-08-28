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

class Chapter {
  final String id;
  final String bookId;
  final int chapterNo;
  final String title;
  final int wordCount;
  final int isVip;
  final int status;

  Chapter.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        bookId = asStr(j['bookId']),
        chapterNo = asInt(j['chapterNo']),
        title = asStr(j['title']),
        wordCount = asInt(j['wordCount']),
        isVip = asInt(j['isVip']),
        status = asInt(j['status']);
}

class ChapterContent {
  final String id;
  final String chapterId;
  final String lang;
  final String content;

  /// 本次内容是否来自本地缓存（离线命中）。
  final bool fromCache;

  ChapterContent({
    required this.id,
    required this.chapterId,
    required this.lang,
    required this.content,
    this.fromCache = false,
  });

  ChapterContent.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        chapterId = asStr(j['chapterId']),
        lang = asStr(j['lang']),
        content = asStr(j['content']),
        fromCache = false;
}

class Comment {
  final String id;
  final String userId;
  final String content;
  final int likeCount;
  final String createdAt;

  Comment.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        userId = asStr(j['userId']),
        content = asStr(j['content']),
        likeCount = asInt(j['likeCount']),
        createdAt = asStr(j['createdAt']);
}

/// 书架条目（GET/POST/DELETE /api/bookshelf）。
class ShelfItem {
  final String id;
  final String userId;
  final String bookId;
  final int sortOrder;

  ShelfItem.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        userId = asStr(j['userId']),
        bookId = asStr(j['bookId']),
        sortOrder = asInt(j['sortOrder']);
}

/// 收藏条目（GET /api/favorites）。
class FavoriteItem {
  final String id;
  final String userId;
  final String bookId;

  FavoriteItem.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        userId = asStr(j['userId']),
        bookId = asStr(j['bookId']);
}

/// 阅读进度（GET/PUT /api/progress）。position 为章节内位置（uint32）。
class ReadingProgress {
  final String id;
  final String userId;
  final String bookId;
  final String chapterId;
  final int position;
  final String updatedAt;

  ReadingProgress.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        userId = asStr(j['userId']),
        bookId = asStr(j['bookId']),
        chapterId = asStr(j['chapterId']),
        position = asInt(j['position']),
        updatedAt = asStr(j['updatedAt']);
}

class User {
  final String id;
  final String username;
  final String nickname;
  final int role;

  User.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        username = asStr(j['username']),
        nickname = asStr(j['nickname']),
        role = asInt(j['role']);
}

class LoginResult {
  final String accessToken;
  final String refreshToken;
  final User user;

  LoginResult.fromJson(Map<String, dynamic> j)
      : accessToken = asStr(j['accessToken']),
        refreshToken = asStr(j['refreshToken']),
        user = User.fromJson((j['user'] ?? {}) as Map<String, dynamic>);
}

class RecommendItem {
  final String bookId;
  final String title;
  final String author;
  final String summary;

  RecommendItem.fromJson(Map<String, dynamic> j)
      : bookId = asStr(j['bookId']),
        title = asStr(j['title']),
        author = asStr(j['author']),
        summary = asStr(j['summary']);
}

/// OpenSearch 搜索结果文档：按 lang 取对应语言字段。
class SearchDoc {
  final String bookId;
  final String lang;
  final String titleZh, titleEn, titleJa;
  final String summaryZh, summaryEn, summaryJa;
  final String authorZh, authorEn, authorJa;

  SearchDoc.fromJson(Map<String, dynamic> j)
      : bookId = asStr(j['bookId']),
        lang = asStr(j['lang']),
        titleZh = asStr(j['titleZh']),
        titleEn = asStr(j['titleEn']),
        titleJa = asStr(j['titleJa']),
        summaryZh = asStr(j['summaryZh']),
        summaryEn = asStr(j['summaryEn']),
        summaryJa = asStr(j['summaryJa']),
        authorZh = asStr(j['authorZh']),
        authorEn = asStr(j['authorEn']),
        authorJa = asStr(j['authorJa']);

  String title(String lang) => switch (lang) {
        'en' => titleEn.isNotEmpty ? titleEn : titleZh,
        'ja' => titleJa.isNotEmpty ? titleJa : titleZh,
        _ => titleZh,
      };

  String summary(String lang) => switch (lang) {
        'en' => summaryEn.isNotEmpty ? summaryEn : summaryZh,
        'ja' => summaryJa.isNotEmpty ? summaryJa : summaryZh,
        _ => summaryZh,
      };

  String author(String lang) => switch (lang) {
        'en' => authorEn.isNotEmpty ? authorEn : authorZh,
        'ja' => authorJa.isNotEmpty ? authorJa : authorZh,
        _ => authorZh,
      };
}
