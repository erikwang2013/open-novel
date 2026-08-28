// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Chinese (`zh`).
class AppLocalizationsZh extends AppLocalizations {
  AppLocalizationsZh([String locale = 'zh']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => '登录';

  @override
  String get register => '注册';

  @override
  String get username => '用户名';

  @override
  String get password => '密码';

  @override
  String get email => '邮箱';

  @override
  String get nickname => '昵称';

  @override
  String get logout => '退出登录';

  @override
  String get home => '首页';

  @override
  String get allBooks => '全部书籍';

  @override
  String get mine => '我的';

  @override
  String get searchHint => '搜索书名 / 作者';

  @override
  String get recommend => '推荐';

  @override
  String get searchResult => '搜索结果';

  @override
  String get bookDetail => '书籍详情';

  @override
  String get chapters => '章节';

  @override
  String get read => '阅读';

  @override
  String get prevChapter => '上一章';

  @override
  String get nextChapter => '下一章';

  @override
  String get comments => '评论';

  @override
  String get commentHint => '写下你的评论…';

  @override
  String get post => '发布';

  @override
  String get like => '点赞';

  @override
  String get loading => '加载中…';

  @override
  String get empty => '暂无数据';

  @override
  String get emptyComment => '还没有评论，来抢沙发';

  @override
  String get vip => 'VIP';

  @override
  String get free => '免费';

  @override
  String get loginRequired => '请先登录';

  @override
  String get errorNetwork => '网络错误，请重试';

  @override
  String errorServer(Object msg) {
    return '服务异常：$msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return '欢迎，$name';
  }

  @override
  String get notLoggedIn => '未登录';

  @override
  String get retry => '重试';

  @override
  String bookCount(num count) {
    return '共 $count 本';
  }

  @override
  String get settings => '阅读设置';

  @override
  String get theme => '主题';

  @override
  String get themeSystem => '跟随系统';

  @override
  String get themeLight => '浅色';

  @override
  String get themeDark => '深色';

  @override
  String get fontSize => '字号';

  @override
  String get lineHeight => '行距';

  @override
  String get pageMode => '翻页方式';

  @override
  String get scrollMode => '上下滚动';

  @override
  String get pagedMode => '左右翻页';

  @override
  String get offline => '离线缓存';

  @override
  String get vipActive => 'VIP 会员';

  @override
  String get vipNotActive => '尚未开通 VIP';

  @override
  String vipExpiresAt(Object date) {
    return '到期时间：$date';
  }

  @override
  String get vipRenew => '续费';

  @override
  String get vipOpen => '开通 VIP';

  @override
  String get payNow => '立即支付';

  @override
  String get paymentResult => '支付结果';

  @override
  String get paymentSuccess => '支付成功';

  @override
  String get paymentFailed => '支付失败';

  @override
  String get paymentPending => '等待支付确认';

  @override
  String get paymentChecking => '正在确认支付结果…';

  @override
  String get paymentNoMethod => '暂无可用支付方式';

  @override
  String get vipChapterLocked => '该章节为 VIP 章节，开通会员后可阅读';

  @override
  String get openVipToRead => '开通会员阅读';

  @override
  String get retryPay => '重新支付';
}
