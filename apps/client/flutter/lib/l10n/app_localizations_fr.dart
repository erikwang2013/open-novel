// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for French (`fr`).
class AppLocalizationsFr extends AppLocalizations {
  AppLocalizationsFr([String locale = 'fr']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'Connexion';

  @override
  String get register => 'Inscription';

  @override
  String get username => 'Nom d\'utilisateur';

  @override
  String get password => 'Mot de passe';

  @override
  String get email => 'E-mail';

  @override
  String get nickname => 'Pseudo';

  @override
  String get logout => 'Déconnexion';

  @override
  String get home => 'Accueil';

  @override
  String get allBooks => 'Bibliothèque';

  @override
  String get mine => 'Mon espace';

  @override
  String get searchHint => 'Rechercher un titre / un auteur';

  @override
  String get recommend => 'Recommandations';

  @override
  String get searchResult => 'Résultats de recherche';

  @override
  String get bookDetail => 'Détails du livre';

  @override
  String get chapters => 'Chapitres';

  @override
  String get read => 'Lire';

  @override
  String get prevChapter => 'Chapitre préc.';

  @override
  String get nextChapter => 'Chap. suivant';

  @override
  String get comments => 'Commentaires';

  @override
  String get commentHint => 'Écrire un commentaire…';

  @override
  String get post => 'Publier';

  @override
  String get like => 'J\'aime';

  @override
  String get loading => 'Chargement…';

  @override
  String get empty => 'Aucune donnée';

  @override
  String get emptyComment =>
      'Aucun commentaire pour l\'instant. Soyez le premier !';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'Gratuit';

  @override
  String get loginRequired => 'Veuillez vous connecter';

  @override
  String get errorNetwork => 'Erreur réseau, veuillez réessayer';

  @override
  String errorServer(Object msg) {
    return 'Erreur serveur : $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'Bienvenue, $name';
  }

  @override
  String get notLoggedIn => 'Non connecté';

  @override
  String get retry => 'Réessayer';

  @override
  String bookCount(num count) {
    return '$count livres';
  }
}
