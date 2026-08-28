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
