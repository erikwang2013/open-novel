// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Indonesian (`id`).
class AppLocalizationsId extends AppLocalizations {
  AppLocalizationsId([String locale = 'id']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'Masuk';

  @override
  String get register => 'Daftar';

  @override
  String get username => 'Nama pengguna';

  @override
  String get password => 'Kata sandi';

  @override
  String get email => 'Email';

  @override
  String get nickname => 'Nama panggilan';

  @override
  String get logout => 'Keluar';

  @override
  String get home => 'Beranda';

  @override
  String get allBooks => 'Perpustakaan';

  @override
  String get mine => 'Akun saya';

  @override
  String get searchHint => 'Cari judul / penulis';

  @override
  String get recommend => 'Rekomendasi';

  @override
  String get searchResult => 'Hasil pencarian';

  @override
  String get bookDetail => 'Detail buku';

  @override
  String get chapters => 'Bab';

  @override
  String get read => 'Baca';

  @override
  String get prevChapter => 'Bab sebelumnya';

  @override
  String get nextChapter => 'Bab berikutnya';

  @override
  String get comments => 'Komentar';

  @override
  String get commentHint => 'Tulis komentar…';

  @override
  String get post => 'Kirim';

  @override
  String get like => 'Suka';

  @override
  String get loading => 'Memuat…';

  @override
  String get empty => 'Tidak ada data';

  @override
  String get emptyComment => 'Belum ada komentar. Jadilah yang pertama!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'Gratis';

  @override
  String get loginRequired => 'Silakan masuk dulu';

  @override
  String get errorNetwork => 'Kesalahan jaringan, coba lagi';

  @override
  String errorServer(Object msg) {
    return 'Kesalahan server: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'Selamat datang, $name';
  }

  @override
  String get notLoggedIn => 'Belum masuk';

  @override
  String get retry => 'Coba lagi';

  @override
  String bookCount(num count) {
    return '$count buku';
  }
}
