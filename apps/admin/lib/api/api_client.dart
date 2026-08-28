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
    return _saveSession(LoginResult.fromJson(_data(r)));
  }

  // ---------- 书籍 ----------

  /// 书籍列表。status: 0 不过滤（管理员） 1 仅上架。
  Future<(List<Book>, int)> books(
      {int page = 1, int pageSize = 20, int status = 0}) async {
    final r = await _dio.get('/api/books', queryParameters: {
      'page': page,
      'page_size': pageSize,
      'status': status,
    });
    final d = _data(r);
    return (_listOf(d, Book.fromJson), asInt(d['total']));
  }

  Future<void> updateBookStatus(String id, int status) async {
    _data(await _dio.patch('/api/books/$id/status', data: {'status': status}));
  }

  /// 建书/编辑元数据（复用 POST /api/books，后端总是新建）。
  Future<void> createBook(Map<String, dynamic> req) async {
    _data(await _dio.post('/api/books', data: req));
  }

  // ---------- 章节 ----------

  Future<(List<Chapter>, int)> chapters(String bookId,
      {int page = 1, int pageSize = 20}) async {
    final r = await _dio.get('/api/books/$bookId/chapters', queryParameters: {
      'page': page,
      'page_size': pageSize,
    });
    final d = _data(r);
    return (_listOf(d, Chapter.fromJson), asInt(d['total']));
  }

  Future<String> chapterContent(String id) async {
    final d = _data(await _dio.get('/api/chapters/$id/content'));
    return asStr(d['content']);
  }

  Future<void> updateChapterStatus(String id, int status) async {
    _data(
        await _dio.patch('/api/chapters/$id/status', data: {'status': status}));
  }

  // ---------- 评论 ----------

  /// 评论列表。status: null=全部 1 正常 2 举报待审（管理员）。
  Future<(List<Comment>, int)> comments(
      {String? bookId,
      String? chapterId,
      int? status,
      int page = 1,
      int pageSize = 20}) async {
    final r = await _dio.get('/api/comments', queryParameters: {
      if (bookId != null && bookId.isNotEmpty) 'book_id': bookId,
      if (chapterId != null && chapterId.isNotEmpty) 'chapter_id': chapterId,
      'status': ?status,
      'page': page,
      'page_size': pageSize,
    });
    final d = _data(r);
    return (_listOf(d, Comment.fromJson), asInt(d['total']));
  }

  Future<void> updateCommentStatus(String id, int status) async {
    _data(
        await _dio.put('/api/comments/$id/status', data: {'status': status}));
  }

  // ---------- 举报 ----------

  /// 待审核举报列表（status=2）。
  Future<(List<Comment>, int)> reports(
      {int page = 1, int pageSize = 20}) async {
    final r = await _dio.get('/api/comments/reports', queryParameters: {
      'page': page,
      'page_size': pageSize,
    });
    final d = _data(r);
    return (_listOf(d, Comment.fromJson), asInt(d['total']));
  }

  /// 举报处理：approved=true 下架评论，false 驳回恢复。
  Future<void> handleReport(String id, bool approved) async {
    _data(
        await _dio.post('/api/comments/$id/report-handle', data: {'approved': approved}));
  }

  // ---------- 用户 ----------

  /// 用户列表（管理员）。search 模糊匹配 username/nickname/email。
  Future<(List<User>, int)> users(
      {String search = '', int page = 1, int pageSize = 20}) async {
    final r = await _dio.get('/api/users', queryParameters: {
      if (search.isNotEmpty) 'search': search,
      'page': page,
      'page_size': pageSize,
    });
    final d = _data(r);
    return (_listOf(d, User.fromJson), asInt(d['total']));
  }

  Future<void> updateUserStatus(String id, int status) async {
    _data(await _dio.patch('/api/users/$id/status', data: {'status': status}));
  }

  Future<void> updateUserRole(String id, int role) async {
    _data(await _dio.patch('/api/users/$id/role', data: {'role': role}));
  }

  // ---------- 仪表盘 ----------

  /// 仪表盘统计（管理员）。
  Future<StatsData> stats() async {
    return StatsData.fromJson(_data(await _dio.get('/api/stats/overview')));
  }

  // ---------- 分类 / 标签 ----------

  Future<(List<Category>, int)> categories() async {
    final d = _data(await _dio.get('/api/categories'));
    return (_listOf(d, Category.fromJson), asInt(d['total']));
  }

  Future<void> createCategory(
      {required String name,
      String parentId = '0',
      int sortOrder = 0}) async {
    _data(await _dio.post('/api/categories',
        data: {'name': name, 'parent_id': parentId, 'sort_order': sortOrder}));
  }

  Future<void> updateCategory(String id, Map<String, dynamic> patch) async {
    _data(await _dio.put('/api/categories/$id', data: patch));
  }

  Future<void> deleteCategory(String id) async {
    _data(await _dio.delete('/api/categories/$id'));
  }

  Future<(List<Tag>, int)> tags() async {
    final d = _data(await _dio.get('/api/tags'));
    return (_listOf(d, Tag.fromJson), asInt(d['total']));
  }

  Future<void> createTag({required String name, String lang = 'zh-CN'}) async {
    _data(await _dio.post('/api/tags', data: {'name': name, 'lang': lang}));
  }

  Future<void> updateTag(String id, Map<String, dynamic> patch) async {
    _data(await _dio.put('/api/tags/$id', data: patch));
  }

  Future<void> deleteTag(String id) async {
    _data(await _dio.delete('/api/tags/$id'));
  }

  // ---------- 支付管理 ----------

  /// 支付方式列表（含禁用）。config 仅返回是否已配置。
  Future<(List<PaymentProvider>, int)> providers() async {
    final d = _data(await _dio.get('/api/payments/admin/providers'));
    return (_listOf(d, PaymentProvider.fromJson), asInt(d['total']));
  }

  Future<void> createProvider({
    required String code,
    String lang = '*',
    String region = '*',
    int sort = 0,
    Map<String, String> config = const {},
  }) async {
    _data(await _dio.post('/api/payments/admin/providers', data: {
      'code': code,
      'lang': lang,
      'region': region,
      'sort': sort,
      'config': config,
    }));
  }

  Future<void> updateProvider(String id, Map<String, dynamic> patch) async {
    _data(await _dio.put('/api/payments/admin/providers/$id', data: patch));
  }

  Future<void> toggleProvider(String id) async {
    _data(await _dio.patch('/api/payments/admin/providers/$id/toggle'));
  }

  Future<void> deleteProvider(String id) async {
    _data(await _dio.delete('/api/payments/admin/providers/$id'));
  }

  /// 流水分页。status: -1 全部；userId/provider 空=全部。
  Future<(List<PaymentOrder>, int)> orders({
    String userId = '',
    String provider = '',
    int status = -1,
    String startAt = '',
    String endAt = '',
    int page = 1,
    int pageSize = 20,
  }) async {
    final r = await _dio.get('/api/payments/admin/orders', queryParameters: {
      if (userId.isNotEmpty) 'user_id': userId,
      if (provider.isNotEmpty) 'provider': provider,
      'status': status,
      if (startAt.isNotEmpty) 'start_at': startAt,
      if (endAt.isNotEmpty) 'end_at': endAt,
      'page': page,
      'page_size': pageSize,
    });
    final d = _data(r);
    return (_listOf(d, PaymentOrder.fromJson), asInt(d['total']));
  }

  Future<OrderStats> orderStats({String startAt = '', String endAt = ''}) async {
    final d = _data(await _dio.get('/api/payments/admin/order-stats',
        queryParameters: {
          if (startAt.isNotEmpty) 'start_at': startAt,
          if (endAt.isNotEmpty) 'end_at': endAt,
        }));
    return OrderStats.fromJson(d);
  }

  /// VIP 套餐列表（含禁用）。
  Future<(List<VipPlan>, int)> plans() async {
    final d = _data(await _dio.get('/api/payments/admin/plans'));
    return (_listOf(d, VipPlan.fromJson), asInt(d['total']));
  }

  Future<void> createPlan({
    required String planCode,
    required int days,
    required int amount,
    String currency = 'USD',
    String label = '',
    int sort = 0,
  }) async {
    _data(await _dio.post('/api/payments/admin/plans', data: {
      'plan_code': planCode,
      'days': days,
      'amount': amount,
      'currency': currency,
      'label': label,
      'sort': sort,
    }));
  }

  Future<void> updatePlan(String id, Map<String, dynamic> patch) async {
    _data(await _dio.put('/api/payments/admin/plans/$id', data: patch));
  }

  Future<void> deletePlan(String id) async {
    _data(await _dio.delete('/api/payments/admin/plans/$id'));
  }

  // ---------- 内部工具 ----------

  /// 解析响应体；业务错误（HTTP 200 + code != null）显式抛出，与 login 一致。
  Map<String, dynamic> _data(Response r) {
    final data = r.data;
    if (data is Map && data['code'] != null) {
      throw DioException(
          requestOptions: r.requestOptions,
          response: r,
          message: data['message']?.toString() ?? '请求失败');
    }
    return (data as Map?)?.cast<String, dynamic>() ?? <String, dynamic>{};
  }

  static List<T> _listOf<T>(
      Map<String, dynamic> d, T Function(Map<String, dynamic>) fromJson) {
    return (d['list'] as List<dynamic>? ?? [])
        .map((e) => fromJson((e as Map).cast<String, dynamic>()))
        .toList();
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
