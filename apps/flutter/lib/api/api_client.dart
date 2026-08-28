import 'package:dio/dio.dart';

import '../models/models.dart';

/// 后端 API 客户端：dio + JWT + 401 自动刷新。
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
  User? currentUser;

  bool get loggedIn => _accessToken != null;

  Future<LoginResult> login(String username, String password) async {
    final r = await _dio.post('/api/users/login',
        data: {'username': username, 'password': password});
    return _saveSession(LoginResult.fromJson(r.data as Map<String, dynamic>));
  }

  Future<LoginResult> register(
      String username, String password, String email, String nickname) async {
    final r = await _dio.post('/api/users/register', data: {
      'username': username,
      'password': password,
      'email': email,
      'nickname': nickname,
    });
    return _saveSession(LoginResult.fromJson(r.data as Map<String, dynamic>));
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

  /// 统一错误文案：优先取后端 {code,message,detail} 的 message。
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

  Future<List<Book>> listBooks(
      {int page = 1, int pageSize = 20, String lang = 'zh'}) async {
    final r = await _dio.get('/api/books',
        queryParameters: {'page': page, 'page_size': pageSize, 'lang': lang});
    return _parseList<Book>(r.data, Book.fromJson);
  }

  Future<Book> getBook(String id, {String lang = 'zh'}) async {
    final r =
        await _dio.get('/api/books/$id', queryParameters: {'lang': lang});
    return Book.fromJson(r.data as Map<String, dynamic>);
  }

  Future<List<RecommendItem>> recommend(
      {String strategy = 'hot', int pageSize = 10, String lang = 'zh'}) async {
    final r = await _dio.get('/api/recommend', queryParameters: {
      'strategy': strategy,
      'page_size': pageSize,
      'lang': lang,
    });
    return _parseList<RecommendItem>(r.data, RecommendItem.fromJson);
  }

  Future<List<SearchDoc>> search(String q,
      {int page = 1, int pageSize = 20, String lang = 'zh'}) async {
    final r = await _dio.get('/api/search', queryParameters: {
      'q': q,
      'page': page,
      'page_size': pageSize,
      'lang': lang,
    });
    return _parseList<SearchDoc>(r.data, SearchDoc.fromJson);
  }

  Future<List<Chapter>> listChapters(String bookId,
      {int page = 1, int pageSize = 100, String lang = 'zh'}) async {
    final r = await _dio.get('/api/books/$bookId/chapters',
        queryParameters: {'page': page, 'page_size': pageSize, 'lang': lang});
    return _parseList<Chapter>(r.data, Chapter.fromJson);
  }

  Future<ChapterContent> getChapterContent(String chapterId,
      {String lang = 'zh'}) async {
    final r = await _dio.get('/api/chapters/$chapterId/content',
        queryParameters: {'lang': lang});
    return ChapterContent.fromJson(r.data as Map<String, dynamic>);
  }

  Future<List<Comment>> listComments(String bookId,
      {String? chapterId, int page = 1, int pageSize = 20}) async {
    final q = <String, dynamic>{
      'book_id': bookId,
      'page': page,
      'page_size': pageSize,
      'chapter_id': ?chapterId,
    };
    final r = await _dio.get('/api/comments', queryParameters: q);
    return _parseList<Comment>(r.data, Comment.fromJson);
  }

  Future<void> postComment(String bookId, String content,
      {String? chapterId}) async {
    await _dio.post('/api/comments', data: {
      'book_id': bookId,
      'chapter_id': ?chapterId,
      'content': content,
    });
  }

  Future<void> likeComment(String id) =>
      _dio.post('/api/comments/$id/like');

  Future<void> unlikeComment(String id) =>
      _dio.delete('/api/comments/$id/like');

  Future<void> favoriteBook(String bookId) =>
      _dio.post('/api/books/$bookId/favorite');

  Future<void> unfavoriteBook(String bookId) =>
      _dio.delete('/api/books/$bookId/favorite');

  List<T> _parseList<T>(
      dynamic data, T Function(Map<String, dynamic>) fromJson) {
    final raw = data is Map ? (data['list'] ?? []) : const [];
    return (raw as List)
        .map((e) => fromJson(e as Map<String, dynamic>))
        .toList();
  }
}

String langCode(String locale) =>
    locale.startsWith('en') ? 'en' : (locale.startsWith('ja') ? 'ja' : 'zh');
