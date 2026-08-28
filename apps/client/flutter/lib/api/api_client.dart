import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_cache_manager/flutter_cache_manager.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:url_launcher/url_launcher.dart';

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
          logout();
        }
        h.next(e);
      },
    ));
  }

  static final ApiClient instance = ApiClient._();
  late final Dio _dio;

  static const _kAccessToken = 'access_token';
  static const _kRefreshToken = 'refresh_token';

  String? _accessToken;
  String? _refreshToken;
  User? currentUser;

  bool get loggedIn => _accessToken != null;

  /// main() 启动时调用一次，从 shared_preferences 恢复 token，免重登（T-C-18）。
  /// key 无前缀与 HarmonyOS 端（open_novel prefs）隔离，与管理端 admin_ 前缀也互不影响。
  Future<void> init() async {
    final p = await SharedPreferences.getInstance();
    _accessToken = p.getString(_kAccessToken);
    _refreshToken = p.getString(_kRefreshToken);
  }

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
    // fire-and-forget 写入：与 T-A-17 管理端一致，写入失败不阻塞请求
    SharedPreferences.getInstance().then((p) {
      p.setString(_kAccessToken, r.accessToken);
      p.setString(_kRefreshToken, r.refreshToken);
    });
    return r;
  }

  void logout() {
    _accessToken = null;
    _refreshToken = null;
    currentUser = null;
    SharedPreferences.getInstance().then((p) {
      p.remove(_kAccessToken);
      p.remove(_kRefreshToken);
    });
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
      {int page = 1,
      int pageSize = 20,
      String lang = 'zh',
      String? categoryId}) async {
    final r = await _dio.get('/api/books', queryParameters: {
      'page': page,
      'page_size': pageSize,
      'lang': lang,
      'category_id': ?categoryId,
    });
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

  Future<(List<Chapter>, int)> listChapters(String bookId,
      {int page = 1, int pageSize = 100, String lang = 'zh'}) async {
    final r = await _dio.get('/api/books/$bookId/chapters',
        queryParameters: {'page': page, 'page_size': pageSize, 'lang': lang});
    final d = r.data as Map<String, dynamic>;
    return (_parseList<Chapter>(d, Chapter.fromJson), asInt(d['total']));
  }

  /// 分页拉取书籍全部章节：后端返回 total，按 total 循环分页拉全（T-C-16）。
  /// page_size 沿用后端上限 100；total 缺失/为 0 时以「返回不足一页」兜底退出。
  Future<List<Chapter>> fetchAllChapters(String bookId,
      {String lang = 'zh'}) async {
    const pageSize = 100;
    final all = <Chapter>[];
    var page = 1;
    while (true) {
      final (part, total) = await listChapters(bookId,
          page: page, pageSize: pageSize, lang: lang);
      all.addAll(part);
      if (all.length >= total || part.length < pageSize) break;
      page++;
    }
    return all;
  }

  Future<ChapterContent> getChapterContent(String chapterId,
      {String lang = 'zh'}) async {
    final r = await _dio.get('/api/chapters/$chapterId/content',
        queryParameters: {'lang': lang});
    return ChapterContent.fromJson(r.data as Map<String, dynamic>);
  }

  /// 章节正文（带本地缓存，T-C-10）：在线成功写缓存，网络失败读缓存，离线可读。
  /// 缓存键 novel://chapter/{id}?lang=，淘汰用 flutter_cache_manager 默认策略
  /// （LRU，最多 200 个对象、7 天过期）。离线且未命中缓存时抛错。
  Future<ChapterContent> getChapterContentCached(String chapterId,
      {String lang = 'zh'}) async {
    final cacheUrl = 'https://novel.invalid/chapter/$chapterId?lang=$lang';
    final cm = DefaultCacheManager();
    try {
      final c = await getChapterContent(chapterId, lang: lang);
      try {
        final tmp = File(
            '${Directory.systemTemp.path}/novel_ch_$chapterId.json');
        await tmp.writeAsString(
            jsonEncode({'content': c.content, 'lang': c.lang}));
        await cm.putFile(cacheUrl, tmp.readAsBytesSync(),
            maxAge: const Duration(days: 7));
      } catch (_) {
        // 写缓存失败不影响在线阅读
      }
      return c;
    } catch (e) {
      final f = await cm.getSingleFile(cacheUrl); // 离线未命中时在此抛错
      final j = jsonDecode(await f.readAsString()) as Map<String, dynamic>;
      return ChapterContent(
        id: '',
        chapterId: chapterId,
        lang: asStr(j['lang']),
        content: asStr(j['content']),
        fromCache: true,
      );
    }
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

  Future<void> reportComment(String id) =>
      _dio.post('/api/comments/$id/report');

  /// 一级分类（GET /api/categories 公开路由）。
  Future<List<Category>> listCategories() async {
    final r = await _dio.get('/api/categories');
    return _parseList<Category>(r.data, Category.fromJson);
  }

  /// 热门搜索（GET /api/search/hot）：后端返回热门书籍 BookDoc，字段与 SearchDoc 一致。
  Future<List<SearchDoc>> hotSearches() async {
    final r = await _dio.get('/api/search/hot');
    return _parseList<SearchDoc>(r.data, SearchDoc.fromJson);
  }

  Future<void> favoriteBook(String bookId) =>
      _dio.post('/api/books/$bookId/favorite');

  Future<void> unfavoriteBook(String bookId) =>
      _dio.delete('/api/books/$bookId/favorite');

  Future<List<ShelfItem>> listBookshelf({int page = 1, int pageSize = 20}) async {
    final r = await _dio.get('/api/bookshelf',
        queryParameters: {'page': page, 'page_size': pageSize});
    return _parseList<ShelfItem>(r.data, ShelfItem.fromJson);
  }

  Future<void> addBookshelf(String bookId) =>
      _dio.post('/api/bookshelf', data: {'book_id': bookId});

  Future<void> removeBookshelf(String bookId) =>
      _dio.delete('/api/bookshelf/$bookId');

  Future<List<FavoriteItem>> listFavorites(
      {int page = 1, int pageSize = 20}) async {
    final r = await _dio.get('/api/favorites',
        queryParameters: {'page': page, 'page_size': pageSize});
    return _parseList<FavoriteItem>(r.data, FavoriteItem.fromJson);
  }

  /// 查询阅读进度；无进度 / 请求失败返回 null（调用方 best-effort 使用）。
  Future<ReadingProgress?> getProgress(String bookId) async {
    try {
      final r = await _dio
          .get('/api/progress', queryParameters: {'book_id': bookId});
      final data = r.data;
      if (data is! Map || data.isEmpty) return null;
      return ReadingProgress.fromJson(data as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  Future<void> updateProgress(String bookId, String chapterId,
          {int position = 0}) =>
      _dio.put('/api/progress', data: {
        'book_id': bookId,
        'chapter_id': chapterId,
        'position': position,
      });

  // ---- VIP / 支付（T-P-14~17）----

  Future<List<Plan>> listPublicPlans() async {
    final r = await _dio.get('/api/payments/plans');
    return _parseList<Plan>(r.data, Plan.fromJson);
  }

  Future<List<PaymentMethod>> listMethods({String lang = 'zh'}) async {
    final r = await _dio
        .get('/api/payments/methods', queryParameters: {'lang': lang});
    return _parseList<PaymentMethod>(r.data, PaymentMethod.fromJson);
  }

  Future<CreateOrderResult> createOrder(String plan, {String lang = 'zh'}) async {
    final r = await _dio
        .post('/api/payments/order', data: {'plan': plan, 'lang': lang});
    return CreateOrderResult.fromJson(r.data as Map<String, dynamic>);
  }

  Future<OrderStatus> getOrder(String orderNo) async {
    final r = await _dio.get('/api/payments/order/$orderNo');
    return OrderStatus.fromJson(r.data as Map<String, dynamic>);
  }

  Future<VipStatus> vipStatus() async {
    final r = await _dio.get('/api/payments/vip-status');
    return VipStatus.fromJson(r.data as Map<String, dynamic>);
  }

  /// 打开支付跳转 URL（系统浏览器，外部应用模式）。
  Future<bool> openCheckoutUrl(String url) {
    return launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
  }

  List<T> _parseList<T>(
      dynamic data, T Function(Map<String, dynamic>) fromJson) {
    final raw = data is Map ? (data['list'] ?? []) : const [];
    return (raw as List)
        .map((e) => fromJson(e as Map<String, dynamic>))
        .toList();
  }
}

/// 客户端语言码 → 后端语言参数（见 kratos/backend/internal/pkg/lang.go NormalizeLang：
/// zh*→zh-CN、en/ja/ko 保留，未知语言原样透传）。
String langCode(String locale) {
  final l = locale.toLowerCase();
  if (l.startsWith('zh')) return 'zh-CN';
  if (l.startsWith('en')) return 'en';
  if (l.startsWith('ja')) return 'ja';
  if (l.startsWith('ko')) return 'ko';
  return locale; // fr/de/es/ru/ar/pt/hi/bn/id 等直接透传
}
