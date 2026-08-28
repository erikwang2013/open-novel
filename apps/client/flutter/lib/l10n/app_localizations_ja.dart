// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Japanese (`ja`).
class AppLocalizationsJa extends AppLocalizations {
  AppLocalizationsJa([String locale = 'ja']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'ログイン';

  @override
  String get register => '新規登録';

  @override
  String get username => 'ユーザー名';

  @override
  String get password => 'パスワード';

  @override
  String get email => 'メールアドレス';

  @override
  String get nickname => 'ニックネーム';

  @override
  String get logout => 'ログアウト';

  @override
  String get home => 'ホーム';

  @override
  String get allBooks => '本棚';

  @override
  String get mine => 'マイページ';

  @override
  String get searchHint => 'タイトル / 作者で検索';

  @override
  String get recommend => 'おすすめ';

  @override
  String get searchResult => '検索結果';

  @override
  String get bookDetail => '書籍詳細';

  @override
  String get chapters => '章';

  @override
  String get read => '読む';

  @override
  String get prevChapter => '前の章';

  @override
  String get nextChapter => '次の章';

  @override
  String get comments => 'コメント';

  @override
  String get commentHint => 'コメントを書く…';

  @override
  String get post => '投稿';

  @override
  String get like => 'いいね';

  @override
  String get loading => '読み込み中…';

  @override
  String get empty => 'データがありません';

  @override
  String get emptyComment => 'まだコメントがありません。最初のコメントを投稿しましょう';

  @override
  String get vip => 'VIP';

  @override
  String get free => '無料';

  @override
  String get loginRequired => 'ログインしてください';

  @override
  String get errorNetwork => 'ネットワークエラーです。もう一度お試しください';

  @override
  String errorServer(Object msg) {
    return 'サーバーエラー：$msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'ようこそ、$name';
  }

  @override
  String get notLoggedIn => '未ログイン';

  @override
  String get retry => '再試行';

  @override
  String bookCount(num count) {
    return '全$count冊';
  }

  @override
  String get settings => '読書設定';

  @override
  String get theme => 'テーマ';

  @override
  String get themeSystem => 'システムに従う';

  @override
  String get themeLight => 'ライト';

  @override
  String get themeDark => 'ダーク';

  @override
  String get fontSize => '文字サイズ';

  @override
  String get lineHeight => '行間';

  @override
  String get pageMode => 'ページモード';

  @override
  String get scrollMode => '上下スクロール';

  @override
  String get pagedMode => '左右ページめくり';

  @override
  String get offline => 'オフラインキャッシュ';
}
