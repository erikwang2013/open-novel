// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Spanish Castilian (`es`).
class AppLocalizationsEs extends AppLocalizations {
  AppLocalizationsEs([String locale = 'es']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'Iniciar sesión';

  @override
  String get register => 'Registrarse';

  @override
  String get username => 'Nombre de usuario';

  @override
  String get password => 'Contraseña';

  @override
  String get email => 'Correo electrónico';

  @override
  String get nickname => 'Apodo';

  @override
  String get logout => 'Cerrar sesión';

  @override
  String get home => 'Inicio';

  @override
  String get allBooks => 'Biblioteca';

  @override
  String get mine => 'Mi cuenta';

  @override
  String get searchHint => 'Buscar título / autor';

  @override
  String get recommend => 'Recomendados';

  @override
  String get searchResult => 'Resultados de búsqueda';

  @override
  String get bookDetail => 'Detalles del libro';

  @override
  String get chapters => 'Capítulos';

  @override
  String get read => 'Leer';

  @override
  String get prevChapter => 'Capítulo anterior';

  @override
  String get nextChapter => 'Siguiente capítulo';

  @override
  String get comments => 'Comentarios';

  @override
  String get commentHint => 'Escribe un comentario…';

  @override
  String get post => 'Publicar';

  @override
  String get like => 'Me gusta';

  @override
  String get unlike => 'Quitar Me Gusta';

  @override
  String get report => 'Denunciar';

  @override
  String get reportSuccess => 'Denuncia enviada';

  @override
  String get reportConfirm => '¿Denunciar este comentario?';

  @override
  String get cancel => 'Cancelar';

  @override
  String get confirm => 'Confirmar';

  @override
  String get all => 'Todos';

  @override
  String get hotSearch => 'Búsquedas populares';

  @override
  String get searchHistory => 'Historial de búsqueda';

  @override
  String get searchSuggest => 'Sugerencias';

  @override
  String get clearHistory => 'Borrar';

  @override
  String get loading => 'Cargando…';

  @override
  String get empty => 'Sin datos';

  @override
  String get emptyComment => 'Aún no hay comentarios. ¡Sé el primero!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'Gratis';

  @override
  String get loginRequired => 'Inicia sesión primero';

  @override
  String get errorNetwork => 'Error de red, inténtalo de nuevo';

  @override
  String errorServer(Object msg) {
    return 'Error del servidor: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'Bienvenido, $name';
  }

  @override
  String get notLoggedIn => 'Sin iniciar sesión';

  @override
  String get retry => 'Reintentar';

  @override
  String bookCount(num count) {
    return '$count libros';
  }

  @override
  String get settings => 'Ajustes de lectura';

  @override
  String get theme => 'Tema';

  @override
  String get themeSystem => 'Seguir sistema';

  @override
  String get themeLight => 'Claro';

  @override
  String get themeDark => 'Oscuro';

  @override
  String get fontSize => 'Tamaño de fuente';

  @override
  String get lineHeight => 'Interlineado';

  @override
  String get pageMode => 'Modo de página';

  @override
  String get scrollMode => 'Desplazamiento vertical';

  @override
  String get pagedMode => 'Pasar página';

  @override
  String get offline => 'Caché sin conexión';

  @override
  String get vipActive => 'Miembro VIP';

  @override
  String get vipNotActive => 'VIP no activado';

  @override
  String vipExpiresAt(Object date) {
    return 'Expira: $date';
  }

  @override
  String get vipRenew => 'Renovar';

  @override
  String get vipOpen => 'Obtener VIP';

  @override
  String get payNow => 'Pagar ahora';

  @override
  String get paymentResult => 'Resultado del pago';

  @override
  String get paymentSuccess => 'Pago exitoso';

  @override
  String get paymentFailed => 'Pago fallido';

  @override
  String get paymentPending => 'Esperando pago';

  @override
  String get paymentChecking => 'Confirmando pago…';

  @override
  String get paymentNoMethod => 'No hay métodos de pago disponibles';

  @override
  String get vipChapterLocked => 'Este capítulo es solo para miembros VIP.';

  @override
  String get openVipToRead => 'Hazte miembro para leer';

  @override
  String get retryPay => 'Pagar de nuevo';
}
