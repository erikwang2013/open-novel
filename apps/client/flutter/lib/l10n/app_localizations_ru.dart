// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Russian (`ru`).
class AppLocalizationsRu extends AppLocalizations {
  AppLocalizationsRu([String locale = 'ru']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'Войти';

  @override
  String get register => 'Регистрация';

  @override
  String get username => 'Имя пользователя';

  @override
  String get password => 'Пароль';

  @override
  String get email => 'Эл. почта';

  @override
  String get nickname => 'Никнейм';

  @override
  String get logout => 'Выйти';

  @override
  String get home => 'Главная';

  @override
  String get allBooks => 'Библиотека';

  @override
  String get mine => 'Моё';

  @override
  String get searchHint => 'Поиск по названию / автору';

  @override
  String get recommend => 'Рекомендации';

  @override
  String get searchResult => 'Результаты поиска';

  @override
  String get bookDetail => 'О книге';

  @override
  String get chapters => 'Главы';

  @override
  String get read => 'Читать';

  @override
  String get prevChapter => 'Предыдущая глава';

  @override
  String get nextChapter => 'Следующая глава';

  @override
  String get comments => 'Комментарии';

  @override
  String get commentHint => 'Напишите комментарий…';

  @override
  String get post => 'Опубликовать';

  @override
  String get like => 'Нравится';

  @override
  String get unlike => 'Убрать лайк';

  @override
  String get report => 'Пожаловаться';

  @override
  String get reportSuccess => 'Жалоба отправлена';

  @override
  String get reportConfirm => 'Пожаловаться на комментарий?';

  @override
  String get cancel => 'Отмена';

  @override
  String get confirm => 'Подтвердить';

  @override
  String get all => 'Все';

  @override
  String get hotSearch => 'Популярные запросы';

  @override
  String get searchHistory => 'История поиска';

  @override
  String get searchSuggest => 'Подсказки';

  @override
  String get clearHistory => 'Очистить';

  @override
  String get loading => 'Загрузка…';

  @override
  String get empty => 'Нет данных';

  @override
  String get emptyComment => 'Комментариев пока нет. Будьте первым!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'Бесплатно';

  @override
  String get loginRequired => 'Сначала войдите';

  @override
  String get errorNetwork => 'Ошибка сети, попробуйте ещё раз';

  @override
  String errorServer(Object msg) {
    return 'Ошибка сервера: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'Добро пожаловать, $name';
  }

  @override
  String get notLoggedIn => 'Вы не вошли';

  @override
  String get retry => 'Повторить';

  @override
  String bookCount(num count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count книг',
      few: '$count книги',
      one: '1 книга',
    );
    return '$_temp0';
  }

  @override
  String get settings => 'Настройки чтения';

  @override
  String get theme => 'Тема';

  @override
  String get themeSystem => 'Как в системе';

  @override
  String get themeLight => 'Светлая';

  @override
  String get themeDark => 'Тёмная';

  @override
  String get fontSize => 'Размер шрифта';

  @override
  String get lineHeight => 'Межстрочный интервал';

  @override
  String get pageMode => 'Режим страниц';

  @override
  String get scrollMode => 'Вертикальная прокрутка';

  @override
  String get pagedMode => 'Перелистывание';

  @override
  String get offline => 'Офлайн-кэш';

  @override
  String get vipActive => 'VIP-участник';

  @override
  String get vipNotActive => 'VIP не активирован';

  @override
  String vipExpiresAt(Object date) {
    return 'Истекает: $date';
  }

  @override
  String get vipRenew => 'Продлить';

  @override
  String get vipOpen => 'Получить VIP';

  @override
  String get payNow => 'Оплатить';

  @override
  String get paymentResult => 'Результат оплаты';

  @override
  String get paymentSuccess => 'Оплата успешна';

  @override
  String get paymentFailed => 'Оплата не удалась';

  @override
  String get paymentPending => 'Ожидание оплаты';

  @override
  String get paymentChecking => 'Проверка оплаты…';

  @override
  String get paymentNoMethod => 'Нет доступных способов оплаты';

  @override
  String get vipChapterLocked =>
      'Эта глава доступна только для VIP-участников.';

  @override
  String get openVipToRead => 'Стать участником';

  @override
  String get retryPay => 'Оплатить снова';
}
