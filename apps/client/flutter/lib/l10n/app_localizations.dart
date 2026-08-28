import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_ar.dart';
import 'app_localizations_bn.dart';
import 'app_localizations_de.dart';
import 'app_localizations_en.dart';
import 'app_localizations_es.dart';
import 'app_localizations_fr.dart';
import 'app_localizations_hi.dart';
import 'app_localizations_id.dart';
import 'app_localizations_ja.dart';
import 'app_localizations_ko.dart';
import 'app_localizations_pt.dart';
import 'app_localizations_ru.dart';
import 'app_localizations_zh.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('ar'),
    Locale('bn'),
    Locale('de'),
    Locale('en'),
    Locale('es'),
    Locale('fr'),
    Locale('hi'),
    Locale('id'),
    Locale('ja'),
    Locale('ko'),
    Locale('pt'),
    Locale('ru'),
    Locale('zh'),
  ];

  /// No description provided for @appTitle.
  ///
  /// In zh, this message translates to:
  /// **'Open Novel'**
  String get appTitle;

  /// No description provided for @login.
  ///
  /// In zh, this message translates to:
  /// **'登录'**
  String get login;

  /// No description provided for @register.
  ///
  /// In zh, this message translates to:
  /// **'注册'**
  String get register;

  /// No description provided for @username.
  ///
  /// In zh, this message translates to:
  /// **'用户名'**
  String get username;

  /// No description provided for @password.
  ///
  /// In zh, this message translates to:
  /// **'密码'**
  String get password;

  /// No description provided for @email.
  ///
  /// In zh, this message translates to:
  /// **'邮箱'**
  String get email;

  /// No description provided for @nickname.
  ///
  /// In zh, this message translates to:
  /// **'昵称'**
  String get nickname;

  /// No description provided for @logout.
  ///
  /// In zh, this message translates to:
  /// **'退出登录'**
  String get logout;

  /// No description provided for @home.
  ///
  /// In zh, this message translates to:
  /// **'首页'**
  String get home;

  /// No description provided for @allBooks.
  ///
  /// In zh, this message translates to:
  /// **'全部书籍'**
  String get allBooks;

  /// No description provided for @mine.
  ///
  /// In zh, this message translates to:
  /// **'我的'**
  String get mine;

  /// No description provided for @searchHint.
  ///
  /// In zh, this message translates to:
  /// **'搜索书名 / 作者'**
  String get searchHint;

  /// No description provided for @recommend.
  ///
  /// In zh, this message translates to:
  /// **'推荐'**
  String get recommend;

  /// No description provided for @searchResult.
  ///
  /// In zh, this message translates to:
  /// **'搜索结果'**
  String get searchResult;

  /// No description provided for @bookDetail.
  ///
  /// In zh, this message translates to:
  /// **'书籍详情'**
  String get bookDetail;

  /// No description provided for @chapters.
  ///
  /// In zh, this message translates to:
  /// **'章节'**
  String get chapters;

  /// No description provided for @read.
  ///
  /// In zh, this message translates to:
  /// **'阅读'**
  String get read;

  /// No description provided for @prevChapter.
  ///
  /// In zh, this message translates to:
  /// **'上一章'**
  String get prevChapter;

  /// No description provided for @nextChapter.
  ///
  /// In zh, this message translates to:
  /// **'下一章'**
  String get nextChapter;

  /// No description provided for @comments.
  ///
  /// In zh, this message translates to:
  /// **'评论'**
  String get comments;

  /// No description provided for @commentHint.
  ///
  /// In zh, this message translates to:
  /// **'写下你的评论…'**
  String get commentHint;

  /// No description provided for @post.
  ///
  /// In zh, this message translates to:
  /// **'发布'**
  String get post;

  /// No description provided for @like.
  ///
  /// In zh, this message translates to:
  /// **'点赞'**
  String get like;

  /// No description provided for @loading.
  ///
  /// In zh, this message translates to:
  /// **'加载中…'**
  String get loading;

  /// No description provided for @empty.
  ///
  /// In zh, this message translates to:
  /// **'暂无数据'**
  String get empty;

  /// No description provided for @emptyComment.
  ///
  /// In zh, this message translates to:
  /// **'还没有评论，来抢沙发'**
  String get emptyComment;

  /// No description provided for @vip.
  ///
  /// In zh, this message translates to:
  /// **'VIP'**
  String get vip;

  /// No description provided for @free.
  ///
  /// In zh, this message translates to:
  /// **'免费'**
  String get free;

  /// No description provided for @loginRequired.
  ///
  /// In zh, this message translates to:
  /// **'请先登录'**
  String get loginRequired;

  /// No description provided for @errorNetwork.
  ///
  /// In zh, this message translates to:
  /// **'网络错误，请重试'**
  String get errorNetwork;

  /// No description provided for @errorServer.
  ///
  /// In zh, this message translates to:
  /// **'服务异常：{msg}'**
  String errorServer(Object msg);

  /// No description provided for @errorMsg.
  ///
  /// In zh, this message translates to:
  /// **'{msg}'**
  String errorMsg(Object msg);

  /// No description provided for @welcome.
  ///
  /// In zh, this message translates to:
  /// **'欢迎，{name}'**
  String welcome(Object name);

  /// No description provided for @notLoggedIn.
  ///
  /// In zh, this message translates to:
  /// **'未登录'**
  String get notLoggedIn;

  /// No description provided for @retry.
  ///
  /// In zh, this message translates to:
  /// **'重试'**
  String get retry;

  /// No description provided for @bookCount.
  ///
  /// In zh, this message translates to:
  /// **'共 {count} 本'**
  String bookCount(num count);

  /// No description provided for @settings.
  ///
  /// In zh, this message translates to:
  /// **'阅读设置'**
  String get settings;

  /// No description provided for @theme.
  ///
  /// In zh, this message translates to:
  /// **'主题'**
  String get theme;

  /// No description provided for @themeSystem.
  ///
  /// In zh, this message translates to:
  /// **'跟随系统'**
  String get themeSystem;

  /// No description provided for @themeLight.
  ///
  /// In zh, this message translates to:
  /// **'浅色'**
  String get themeLight;

  /// No description provided for @themeDark.
  ///
  /// In zh, this message translates to:
  /// **'深色'**
  String get themeDark;

  /// No description provided for @fontSize.
  ///
  /// In zh, this message translates to:
  /// **'字号'**
  String get fontSize;

  /// No description provided for @lineHeight.
  ///
  /// In zh, this message translates to:
  /// **'行距'**
  String get lineHeight;

  /// No description provided for @pageMode.
  ///
  /// In zh, this message translates to:
  /// **'翻页方式'**
  String get pageMode;

  /// No description provided for @scrollMode.
  ///
  /// In zh, this message translates to:
  /// **'上下滚动'**
  String get scrollMode;

  /// No description provided for @pagedMode.
  ///
  /// In zh, this message translates to:
  /// **'左右翻页'**
  String get pagedMode;

  /// No description provided for @offline.
  ///
  /// In zh, this message translates to:
  /// **'离线缓存'**
  String get offline;

  /// No description provided for @vipActive.
  ///
  /// In zh, this message translates to:
  /// **'VIP 会员'**
  String get vipActive;

  /// No description provided for @vipNotActive.
  ///
  /// In zh, this message translates to:
  /// **'尚未开通 VIP'**
  String get vipNotActive;

  /// No description provided for @vipExpiresAt.
  ///
  /// In zh, this message translates to:
  /// **'到期时间：{date}'**
  String vipExpiresAt(Object date);

  /// No description provided for @vipRenew.
  ///
  /// In zh, this message translates to:
  /// **'续费'**
  String get vipRenew;

  /// No description provided for @vipOpen.
  ///
  /// In zh, this message translates to:
  /// **'开通 VIP'**
  String get vipOpen;

  /// No description provided for @payNow.
  ///
  /// In zh, this message translates to:
  /// **'立即支付'**
  String get payNow;

  /// No description provided for @paymentResult.
  ///
  /// In zh, this message translates to:
  /// **'支付结果'**
  String get paymentResult;

  /// No description provided for @paymentSuccess.
  ///
  /// In zh, this message translates to:
  /// **'支付成功'**
  String get paymentSuccess;

  /// No description provided for @paymentFailed.
  ///
  /// In zh, this message translates to:
  /// **'支付失败'**
  String get paymentFailed;

  /// No description provided for @paymentPending.
  ///
  /// In zh, this message translates to:
  /// **'等待支付确认'**
  String get paymentPending;

  /// No description provided for @paymentChecking.
  ///
  /// In zh, this message translates to:
  /// **'正在确认支付结果…'**
  String get paymentChecking;

  /// No description provided for @paymentNoMethod.
  ///
  /// In zh, this message translates to:
  /// **'暂无可用支付方式'**
  String get paymentNoMethod;

  /// No description provided for @vipChapterLocked.
  ///
  /// In zh, this message translates to:
  /// **'该章节为 VIP 章节，开通会员后可阅读'**
  String get vipChapterLocked;

  /// No description provided for @openVipToRead.
  ///
  /// In zh, this message translates to:
  /// **'开通会员阅读'**
  String get openVipToRead;

  /// No description provided for @retryPay.
  ///
  /// In zh, this message translates to:
  /// **'重新支付'**
  String get retryPay;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) => <String>[
    'ar',
    'bn',
    'de',
    'en',
    'es',
    'fr',
    'hi',
    'id',
    'ja',
    'ko',
    'pt',
    'ru',
    'zh',
  ].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'ar':
      return AppLocalizationsAr();
    case 'bn':
      return AppLocalizationsBn();
    case 'de':
      return AppLocalizationsDe();
    case 'en':
      return AppLocalizationsEn();
    case 'es':
      return AppLocalizationsEs();
    case 'fr':
      return AppLocalizationsFr();
    case 'hi':
      return AppLocalizationsHi();
    case 'id':
      return AppLocalizationsId();
    case 'ja':
      return AppLocalizationsJa();
    case 'ko':
      return AppLocalizationsKo();
    case 'pt':
      return AppLocalizationsPt();
    case 'ru':
      return AppLocalizationsRu();
    case 'zh':
      return AppLocalizationsZh();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
