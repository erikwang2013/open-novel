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
