// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Bengali Bangla (`bn`).
class AppLocalizationsBn extends AppLocalizations {
  AppLocalizationsBn([String locale = 'bn']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'লগ ইন করুন';

  @override
  String get register => 'নিবন্ধন করুন';

  @override
  String get username => 'ব্যবহারকারীর নাম';

  @override
  String get password => 'পাসওয়ার্ড';

  @override
  String get email => 'ইমেইল';

  @override
  String get nickname => 'ডাকনাম';

  @override
  String get logout => 'লগ আউট করুন';

  @override
  String get home => 'হোম';

  @override
  String get allBooks => 'লাইব্রেরি';

  @override
  String get mine => 'আমার';

  @override
  String get searchHint => 'শিরোনাম / লেখক খুঁজুন';

  @override
  String get recommend => 'সুপারিশকৃত';

  @override
  String get searchResult => 'অনুসন্ধান ফলাফল';

  @override
  String get bookDetail => 'বইয়ের বিবরণ';

  @override
  String get chapters => 'অধ্যায়';

  @override
  String get read => 'পড়ুন';

  @override
  String get prevChapter => 'আগের অধ্যায়';

  @override
  String get nextChapter => 'পরের অধ্যায়';

  @override
  String get comments => 'মন্তব্য';

  @override
  String get commentHint => 'মন্তব্য লিখুন…';

  @override
  String get post => 'পোস্ট করুন';

  @override
  String get like => 'পছন্দ';

  @override
  String get unlike => 'লাইক বাতিল';

  @override
  String get report => 'রিপোর্ট';

  @override
  String get reportSuccess => 'রিপোর্ট সফল';

  @override
  String get reportConfirm => 'এই মন্তব্যটি রিপোর্ট করবেন?';

  @override
  String get cancel => 'বাতিল';

  @override
  String get confirm => 'নিশ্চিত';

  @override
  String get all => 'সব';

  @override
  String get hotSearch => 'জনপ্রিয় অনুসন্ধান';

  @override
  String get searchHistory => 'অনুসন্ধান ইতিহাস';

  @override
  String get searchSuggest => 'পরামর্শ';

  @override
  String get clearHistory => 'মুছুন';

  @override
  String get loading => 'লোড হচ্ছে…';

  @override
  String get empty => 'কোনো তথ্য নেই';

  @override
  String get emptyComment => 'এখনো কোনো মন্তব্য নেই। প্রথম মন্তব্য করুন!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'বিনামূল্য';

  @override
  String get loginRequired => 'অনুগ্রহ করে প্রথমে লগ ইন করুন';

  @override
  String get errorNetwork => 'নেটওয়ার্ক ত্রুটি, আবার চেষ্টা করুন';

  @override
  String errorServer(Object msg) {
    return 'সার্ভার ত্রুটি: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'স্বাগতম, $name';
  }

  @override
  String get notLoggedIn => 'লগ ইন করা হয়নি';

  @override
  String get retry => 'আবার চেষ্টা করুন';

  @override
  String bookCount(num count) {
    return '$countটি বই';
  }

  @override
  String get settings => 'পড়ার সেটিংস';

  @override
  String get theme => 'থিম';

  @override
  String get themeSystem => 'সিস্টেম অনুসারে';

  @override
  String get themeLight => 'হালকা';

  @override
  String get themeDark => 'গাঢ়';

  @override
  String get fontSize => 'ফন্ট সাইজ';

  @override
  String get lineHeight => 'লাইন স্পেসিং';

  @override
  String get pageMode => 'পৃষ্ঠা মোড';

  @override
  String get scrollMode => 'উপর-নিচে স্ক্রল';

  @override
  String get pagedMode => 'বাম-ডান পাতা';

  @override
  String get offline => 'অফলাইন ক্যাশে';

  @override
  String get vipActive => 'VIP সদস্য';

  @override
  String get vipNotActive => 'VIP সক্রিয় নয়';

  @override
  String vipExpiresAt(Object date) {
    return 'মেয়াদ শেষ: $date';
  }

  @override
  String get vipRenew => 'নবায়ন';

  @override
  String get vipOpen => 'VIP নিন';

  @override
  String get payNow => 'এখনই পেমেন্ট';

  @override
  String get paymentResult => 'পেমেন্ট ফলাফল';

  @override
  String get paymentSuccess => 'পেমেন্ট সফল';

  @override
  String get paymentFailed => 'পেমেন্ট ব্যর্থ';

  @override
  String get paymentPending => 'পেমেন্টের অপেক্ষায়';

  @override
  String get paymentChecking => 'পেমেন্ট নিশ্চিত হচ্ছে…';

  @override
  String get paymentNoMethod => 'কোনো পেমেন্ট পদ্ধতি নেই';

  @override
  String get vipChapterLocked => 'এই অধ্যায়টি শুধু VIP সদস্যদের জন্য।';

  @override
  String get openVipToRead => 'সদস্য হয়ে পড়ুন';

  @override
  String get retryPay => 'আবার পেমেন্ট করুন';
}
