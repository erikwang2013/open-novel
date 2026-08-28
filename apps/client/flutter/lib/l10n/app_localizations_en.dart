// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'Sign in';

  @override
  String get register => 'Sign up';

  @override
  String get username => 'Username';

  @override
  String get password => 'Password';

  @override
  String get email => 'Email';

  @override
  String get nickname => 'Nickname';

  @override
  String get logout => 'Sign out';

  @override
  String get home => 'Home';

  @override
  String get allBooks => 'All Books';

  @override
  String get mine => 'Me';

  @override
  String get searchHint => 'Search title / author';

  @override
  String get recommend => 'Recommended';

  @override
  String get searchResult => 'Search Results';

  @override
  String get bookDetail => 'Book Details';

  @override
  String get chapters => 'Chapters';

  @override
  String get read => 'Read';

  @override
  String get prevChapter => 'Prev';

  @override
  String get nextChapter => 'Next';

  @override
  String get comments => 'Comments';

  @override
  String get commentHint => 'Write a comment…';

  @override
  String get post => 'Post';

  @override
  String get like => 'Like';

  @override
  String get loading => 'Loading…';

  @override
  String get empty => 'Nothing here';

  @override
  String get emptyComment => 'No comments yet. Be the first!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'Free';

  @override
  String get loginRequired => 'Please sign in first';

  @override
  String get errorNetwork => 'Network error, please retry';

  @override
  String errorServer(Object msg) {
    return 'Server error: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'Welcome, $name';
  }

  @override
  String get notLoggedIn => 'Not signed in';

  @override
  String get retry => 'Retry';

  @override
  String bookCount(num count) {
    return '$count books';
  }

  @override
  String get settings => 'Reading settings';

  @override
  String get theme => 'Theme';

  @override
  String get themeSystem => 'Follow system';

  @override
  String get themeLight => 'Light';

  @override
  String get themeDark => 'Dark';

  @override
  String get fontSize => 'Font size';

  @override
  String get lineHeight => 'Line height';

  @override
  String get pageMode => 'Page mode';

  @override
  String get scrollMode => 'Vertical scroll';

  @override
  String get pagedMode => 'Page flip';

  @override
  String get offline => 'Offline cache';

  @override
  String get vipActive => 'VIP Member';

  @override
  String get vipNotActive => 'VIP not activated';

  @override
  String vipExpiresAt(Object date) {
    return 'Expires: $date';
  }

  @override
  String get vipRenew => 'Renew';

  @override
  String get vipOpen => 'Get VIP';

  @override
  String get payNow => 'Pay Now';

  @override
  String get paymentResult => 'Payment Result';

  @override
  String get paymentSuccess => 'Payment Successful';

  @override
  String get paymentFailed => 'Payment Failed';

  @override
  String get paymentPending => 'Awaiting payment';

  @override
  String get paymentChecking => 'Confirming payment…';

  @override
  String get paymentNoMethod => 'No payment method available';

  @override
  String get vipChapterLocked => 'This is a VIP chapter. Get VIP to read.';

  @override
  String get openVipToRead => 'Get VIP to read';

  @override
  String get retryPay => 'Pay Again';
}
