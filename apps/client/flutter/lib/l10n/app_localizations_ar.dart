// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Arabic (`ar`).
class AppLocalizationsAr extends AppLocalizations {
  AppLocalizationsAr([String locale = 'ar']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'تسجيل الدخول';

  @override
  String get register => 'إنشاء حساب';

  @override
  String get username => 'اسم المستخدم';

  @override
  String get password => 'كلمة المرور';

  @override
  String get email => 'البريد الإلكتروني';

  @override
  String get nickname => 'الاسم المستعار';

  @override
  String get logout => 'تسجيل الخروج';

  @override
  String get home => 'الرئيسية';

  @override
  String get allBooks => 'المكتبة';

  @override
  String get mine => 'حسابي';

  @override
  String get searchHint => 'ابحث عن العنوان / المؤلف';

  @override
  String get recommend => 'الموصى بها';

  @override
  String get searchResult => 'نتائج البحث';

  @override
  String get bookDetail => 'تفاصيل الكتاب';

  @override
  String get chapters => 'الفصول';

  @override
  String get read => 'اقرأ';

  @override
  String get prevChapter => 'الفصل السابق';

  @override
  String get nextChapter => 'الفصل التالي';

  @override
  String get comments => 'التعليقات';

  @override
  String get commentHint => 'اكتب تعليقًا…';

  @override
  String get post => 'نشر';

  @override
  String get like => 'إعجاب';

  @override
  String get loading => 'جارٍ التحميل…';

  @override
  String get empty => 'لا توجد بيانات';

  @override
  String get emptyComment => 'لا توجد تعليقات بعد. كن أول من يعلق!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'مجاني';

  @override
  String get loginRequired => 'الرجاء تسجيل الدخول أولاً';

  @override
  String get errorNetwork => 'خطأ في الشبكة، حاول مرة أخرى';

  @override
  String errorServer(Object msg) {
    return 'خطأ في الخادم: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'مرحبًا، $name';
  }

  @override
  String get notLoggedIn => 'غير مسجل الدخول';

  @override
  String get retry => 'إعادة المحاولة';

  @override
  String bookCount(num count) {
    return '$count كتب';
  }
}
