// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for German (`de`).
class AppLocalizationsDe extends AppLocalizations {
  AppLocalizationsDe([String locale = 'de']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'Anmelden';

  @override
  String get register => 'Registrieren';

  @override
  String get username => 'Benutzername';

  @override
  String get password => 'Passwort';

  @override
  String get email => 'E-Mail';

  @override
  String get nickname => 'Spitzname';

  @override
  String get logout => 'Abmelden';

  @override
  String get home => 'Start';

  @override
  String get allBooks => 'Bibliothek';

  @override
  String get mine => 'Mein Bereich';

  @override
  String get searchHint => 'Titel / Autor suchen';

  @override
  String get recommend => 'Empfehlungen';

  @override
  String get searchResult => 'Suchergebnisse';

  @override
  String get bookDetail => 'Buchdetails';

  @override
  String get chapters => 'Kapitel';

  @override
  String get read => 'Lesen';

  @override
  String get prevChapter => 'Vorheriges Kapitel';

  @override
  String get nextChapter => 'Nächstes Kapitel';

  @override
  String get comments => 'Kommentare';

  @override
  String get commentHint => 'Kommentar schreiben…';

  @override
  String get post => 'Veröffentlichen';

  @override
  String get like => 'Gefällt mir';

  @override
  String get unlike => 'Like entfernen';

  @override
  String get report => 'Melden';

  @override
  String get reportSuccess => 'Meldung gesendet';

  @override
  String get reportConfirm => 'Diesen Kommentar melden?';

  @override
  String get cancel => 'Abbrechen';

  @override
  String get confirm => 'Bestätigen';

  @override
  String get all => 'Alle';

  @override
  String get hotSearch => 'Beliebte Suchanfragen';

  @override
  String get loading => 'Wird geladen…';

  @override
  String get empty => 'Keine Daten';

  @override
  String get emptyComment => 'Noch keine Kommentare. Sei der Erste!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'Kostenlos';

  @override
  String get loginRequired => 'Bitte zuerst anmelden';

  @override
  String get errorNetwork => 'Netzwerkfehler, bitte erneut versuchen';

  @override
  String errorServer(Object msg) {
    return 'Serverfehler: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'Willkommen, $name';
  }

  @override
  String get notLoggedIn => 'Nicht angemeldet';

  @override
  String get retry => 'Erneut versuchen';

  @override
  String bookCount(num count) {
    return '$count Bücher';
  }

  @override
  String get settings => 'Leseeinstellungen';

  @override
  String get theme => 'Thema';

  @override
  String get themeSystem => 'System folgen';

  @override
  String get themeLight => 'Hell';

  @override
  String get themeDark => 'Dunkel';

  @override
  String get fontSize => 'Schriftgröße';

  @override
  String get lineHeight => 'Zeilenabstand';

  @override
  String get pageMode => 'Seitenmodus';

  @override
  String get scrollMode => 'Vertikal scrollen';

  @override
  String get pagedMode => 'Blättern';

  @override
  String get offline => 'Offline-Cache';

  @override
  String get vipActive => 'VIP-Mitglied';

  @override
  String get vipNotActive => 'VIP nicht aktiviert';

  @override
  String vipExpiresAt(Object date) {
    return 'Gültig bis $date';
  }

  @override
  String get vipRenew => 'Verlängern';

  @override
  String get vipOpen => 'VIP erhalten';

  @override
  String get payNow => 'Jetzt zahlen';

  @override
  String get paymentResult => 'Zahlungsergebnis';

  @override
  String get paymentSuccess => 'Zahlung erfolgreich';

  @override
  String get paymentFailed => 'Zahlung fehlgeschlagen';

  @override
  String get paymentPending => 'Warte auf Zahlung';

  @override
  String get paymentChecking => 'Zahlung wird geprüft…';

  @override
  String get paymentNoMethod => 'Keine Zahlungsmethode verfügbar';

  @override
  String get vipChapterLocked => 'Dieses Kapitel ist nur für VIP-Mitglieder.';

  @override
  String get openVipToRead => 'Mitglied werden';

  @override
  String get retryPay => 'Erneut zahlen';
}
