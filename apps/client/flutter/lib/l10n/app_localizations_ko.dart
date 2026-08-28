// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Korean (`ko`).
class AppLocalizationsKo extends AppLocalizations {
  AppLocalizationsKo([String locale = 'ko']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => '로그인';

  @override
  String get register => '회원가입';

  @override
  String get username => '사용자 이름';

  @override
  String get password => '비밀번호';

  @override
  String get email => '이메일';

  @override
  String get nickname => '닉네임';

  @override
  String get logout => '로그아웃';

  @override
  String get home => '홈';

  @override
  String get allBooks => '서재';

  @override
  String get mine => '마이페이지';

  @override
  String get searchHint => '제목 / 작가 검색';

  @override
  String get recommend => '추천';

  @override
  String get searchResult => '검색 결과';

  @override
  String get bookDetail => '작품 정보';

  @override
  String get chapters => '목차';

  @override
  String get read => '읽기';

  @override
  String get prevChapter => '이전 장';

  @override
  String get nextChapter => '다음 장';

  @override
  String get comments => '댓글';

  @override
  String get commentHint => '댓글을 남겨보세요…';

  @override
  String get post => '등록';

  @override
  String get like => '좋아요';

  @override
  String get loading => '불러오는 중…';

  @override
  String get empty => '데이터가 없습니다';

  @override
  String get emptyComment => '아직 댓글이 없습니다. 첫 댓글을 남겨보세요';

  @override
  String get vip => 'VIP';

  @override
  String get free => '무료';

  @override
  String get loginRequired => '먼저 로그인해 주세요';

  @override
  String get errorNetwork => '네트워크 오류가 발생했습니다. 다시 시도해 주세요';

  @override
  String errorServer(Object msg) {
    return '서버 오류: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return '환영합니다, $name';
  }

  @override
  String get notLoggedIn => '로그인되지 않음';

  @override
  String get retry => '다시 시도';

  @override
  String bookCount(Object count) {
    return '총 $count권';
  }
}
