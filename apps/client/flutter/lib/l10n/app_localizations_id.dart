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

  @override
  String get settings => 'Pengaturan baca';

  @override
  String get theme => 'Tema';

  @override
  String get themeSystem => 'Ikuti sistem';

  @override
  String get themeLight => 'Terang';

  @override
  String get themeDark => 'Gelap';

  @override
  String get fontSize => 'Ukuran font';

  @override
  String get lineHeight => 'Spasi baris';

  @override
  String get pageMode => 'Mode halaman';

  @override
  String get scrollMode => 'Gulir vertikal';

  @override
  String get pagedMode => 'Balik halaman';

  @override
  String get offline => 'Cache offline';

  @override
  String get vipActive => 'Anggota VIP';

  @override
  String get vipNotActive => 'VIP belum aktif';

  @override
  String vipExpiresAt(Object date) {
    return 'Berakhir: $date';
  }

  @override
  String get vipRenew => 'Perpanjang';

  @override
  String get vipOpen => 'Dapatkan VIP';

  @override
  String get payNow => 'Bayar Sekarang';

  @override
  String get paymentResult => 'Hasil Pembayaran';

  @override
  String get paymentSuccess => 'Pembayaran Berhasil';

  @override
  String get paymentFailed => 'Pembayaran Gagal';

  @override
  String get paymentPending => 'Menunggu Pembayaran';

  @override
  String get paymentChecking => 'Memverifikasi pembayaran…';

  @override
  String get paymentNoMethod => 'Tidak ada metode pembayaran tersedia';

  @override
  String get vipChapterLocked => 'Bab ini khusus anggota VIP.';

  @override
  String get openVipToRead => 'Jadi anggota untuk membaca';

  @override
  String get retryPay => 'Bayar Lagi';
}
