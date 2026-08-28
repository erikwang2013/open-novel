import 'package:dio/dio.dart';

import '../models/models.dart';

/// 管理端 API 客户端：dio + JWT + 401 自动刷新（模式同客户端）。
/// baseUrl 通过 --dart-define=API_BASE_URL=... 注入，默认本地开发地址。
class ApiClient {
  ApiClient._() {
    _dio = Dio(BaseOptions(
      baseUrl: const String.fromEnvironment('API_BASE_URL',
          defaultValue: 'http://localhost:8000'),
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 15),
    ));
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (o, h) {
        o.headers['X-Api-Version'] = 'v1';
        if (_accessToken != null) {
          o.headers['Authorization'] = 'Bearer $_accessToken';
        }
        h.next(o);
      },
      onError: (e, h) async {
        if (e.response?.statusCode == 401 && _refreshToken != null) {
          try {
            final ok = await _refresh();
            if (ok) {
              final opts = e.requestOptions
                ..headers['Authorization'] = 'Bearer $_accessToken';
              h.resolve(await _dio.fetch(opts));
              return;
            }
          } catch (_) {
            // 刷新失败走下方统一错误
          }
          _accessToken = null;
          _refreshToken = null;
        }
        h.next(e);
      },
    ));
  }

  static final ApiClient instance = ApiClient._();
  late final Dio _dio;

  String? _accessToken;
  String? _refreshToken;
  AdminUser? currentUser;

  bool get loggedIn => _accessToken != null;

  Future<LoginResult> login(String username, String password) async {
    final r = await _dio.post('/api/users/login',
        data: {'username': username, 'password': password});
    final data = r.data as Map<String, dynamic>;
    // 业务错误（密码错误等）返回 HTTP 200 + {code,message}，需显式抛出
    if (data['code'] != null) {
      throw DioException(
          requestOptions: r.requestOptions,
          response: r,
          message: data['message']?.toString() ?? '登录失败');
    }
    return _saveSession(LoginResult.fromJson(data));
  }

  Future<bool> _refresh() async {
    final r = await _dio.post('/api/users/refresh',
        data: {'refresh_token': _refreshToken});
    _saveSession(LoginResult.fromJson(r.data as Map<String, dynamic>));
    return true;
  }

  LoginResult _saveSession(LoginResult r) {
    _accessToken = r.accessToken;
    _refreshToken = r.refreshToken;
    currentUser = r.user;
    return r;
  }

  void logout() {
    _accessToken = null;
    _refreshToken = null;
    currentUser = null;
  }

  /// 统一错误文案：优先取后端 {code,message,detail} 的 message；网络类返回 'network'。
  String errorMessage(Object e) {
    if (e is DioException) {
      final data = e.response?.data;
      if (data is Map && data['message'] != null) {
        return data['message'].toString();
      }
      if (e.type == DioExceptionType.connectionError ||
          e.type == DioExceptionType.connectionTimeout) {
        return 'network';
      }
    }
    return e.toString();
  }
}
