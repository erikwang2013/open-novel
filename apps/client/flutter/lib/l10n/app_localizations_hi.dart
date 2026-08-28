// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Hindi (`hi`).
class AppLocalizationsHi extends AppLocalizations {
  AppLocalizationsHi([String locale = 'hi']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'लॉग इन करें';

  @override
  String get register => 'साइन अप करें';

  @override
  String get username => 'उपयोगकर्ता नाम';

  @override
  String get password => 'पासवर्ड';

  @override
  String get email => 'ईमेल';

  @override
  String get nickname => 'उपनाम';

  @override
  String get logout => 'लॉग आउट करें';

  @override
  String get home => 'होम';

  @override
  String get allBooks => 'लाइब्रेरी';

  @override
  String get mine => 'मेरा';

  @override
  String get searchHint => 'शीर्षक / लेखक खोजें';

  @override
  String get recommend => 'अनुशंसित';

  @override
  String get searchResult => 'खोज परिणाम';

  @override
  String get bookDetail => 'पुस्तक विवरण';

  @override
  String get chapters => 'अध्याय';

  @override
  String get read => 'पढ़ें';

  @override
  String get prevChapter => 'पिछला अध्याय';

  @override
  String get nextChapter => 'अगला अध्याय';

  @override
  String get comments => 'टिप्पणियाँ';

  @override
  String get commentHint => 'टिप्पणी लिखें…';

  @override
  String get post => 'पोस्ट करें';

  @override
  String get like => 'पसंद';

  @override
  String get loading => 'लोड हो रहा है…';

  @override
  String get empty => 'कोई डेटा नहीं';

  @override
  String get emptyComment => 'अभी कोई टिप्पणी नहीं है। पहली टिप्पणी करें!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'मुफ़्त';

  @override
  String get loginRequired => 'कृपया पहले लॉग इन करें';

  @override
  String get errorNetwork => 'नेटवर्क त्रुटि, कृपया पुनः प्रयास करें';

  @override
  String errorServer(Object msg) {
    return 'सर्वर त्रुटि: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'स्वागत है, $name';
  }

  @override
  String get notLoggedIn => 'लॉग इन नहीं किया गया';

  @override
  String get retry => 'पुनः प्रयास करें';

  @override
  String bookCount(Object count) {
    return '$count पुस्तकें';
  }
}
