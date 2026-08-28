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

  @override
  String get settings => 'Réglages de lecture';

  @override
  String get theme => 'Thème';

  @override
  String get themeSystem => 'Suivre le système';

  @override
  String get themeLight => 'Clair';

  @override
  String get themeDark => 'Sombre';

  @override
  String get fontSize => 'Taille de police';

  @override
  String get lineHeight => 'Interligne';

  @override
  String get pageMode => 'Mode de page';

  @override
  String get scrollMode => 'Défilement vertical';

  @override
  String get pagedMode => 'Feuilletage';

  @override
  String get offline => 'Cache hors ligne';

  @override
  String get vipActive => 'Membre VIP';

  @override
  String get vipNotActive => 'VIP non activé';

  @override
  String vipExpiresAt(Object date) {
    return 'Expire le $date';
  }

  @override
  String get vipRenew => 'Renouveler';

  @override
  String get vipOpen => 'Obtenir VIP';

  @override
  String get payNow => 'Payer';

  @override
  String get paymentResult => 'Résultat du paiement';

  @override
  String get paymentSuccess => 'Paiement réussi';

  @override
  String get paymentFailed => 'Échec du paiement';

  @override
  String get paymentPending => 'En attente de paiement';

  @override
  String get paymentChecking => 'Vérification du paiement…';

  @override
  String get paymentNoMethod => 'Aucun moyen de paiement disponible';

  @override
  String get vipChapterLocked => 'Ce chapitre est réservé aux membres VIP.';

  @override
  String get openVipToRead => 'Devenir membre pour lire';

  @override
  String get retryPay => 'Repayer';
}
