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
  String bookCount(Object count) {
    return '$count Bücher';
  }
}
