// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Portuguese (`pt`).
class AppLocalizationsPt extends AppLocalizations {
  AppLocalizationsPt([String locale = 'pt']) : super(locale);

  @override
  String get appTitle => 'Open Novel';

  @override
  String get login => 'Entrar';

  @override
  String get register => 'Cadastrar';

  @override
  String get username => 'Nome de usuário';

  @override
  String get password => 'Senha';

  @override
  String get email => 'E-mail';

  @override
  String get nickname => 'Apelido';

  @override
  String get logout => 'Sair';

  @override
  String get home => 'Início';

  @override
  String get allBooks => 'Biblioteca';

  @override
  String get mine => 'Minha conta';

  @override
  String get searchHint => 'Buscar título / autor';

  @override
  String get recommend => 'Recomendados';

  @override
  String get searchResult => 'Resultados da busca';

  @override
  String get bookDetail => 'Detalhes do livro';

  @override
  String get chapters => 'Capítulos';

  @override
  String get read => 'Ler';

  @override
  String get prevChapter => 'Capítulo anterior';

  @override
  String get nextChapter => 'Próximo capítulo';

  @override
  String get comments => 'Comentários';

  @override
  String get commentHint => 'Escreva um comentário…';

  @override
  String get post => 'Publicar';

  @override
  String get like => 'Curtir';

  @override
  String get unlike => 'Descurtir';

  @override
  String get report => 'Denunciar';

  @override
  String get reportSuccess => 'Denúncia enviada';

  @override
  String get reportConfirm => 'Denunciar este comentário?';

  @override
  String get cancel => 'Cancelar';

  @override
  String get confirm => 'Confirmar';

  @override
  String get all => 'Todos';

  @override
  String get hotSearch => 'Pesquisas populares';

  @override
  String get searchHistory => 'Histórico de busca';

  @override
  String get searchSuggest => 'Sugestões';

  @override
  String get clearHistory => 'Limpar';

  @override
  String get loading => 'Carregando…';

  @override
  String get empty => 'Sem dados';

  @override
  String get emptyComment => 'Ainda sem comentários. Seja o primeiro!';

  @override
  String get vip => 'VIP';

  @override
  String get free => 'Grátis';

  @override
  String get loginRequired => 'Entre primeiro';

  @override
  String get errorNetwork => 'Erro de rede, tente novamente';

  @override
  String errorServer(Object msg) {
    return 'Erro do servidor: $msg';
  }

  @override
  String errorMsg(Object msg) {
    return '$msg';
  }

  @override
  String welcome(Object name) {
    return 'Bem-vindo, $name';
  }

  @override
  String get notLoggedIn => 'Não conectado';

  @override
  String get retry => 'Tentar novamente';

  @override
  String bookCount(num count) {
    return '$count livros';
  }

  @override
  String get settings => 'Configurações de leitura';

  @override
  String get theme => 'Tema';

  @override
  String get themeSystem => 'Seguir sistema';

  @override
  String get themeLight => 'Claro';

  @override
  String get themeDark => 'Escuro';

  @override
  String get fontSize => 'Tamanho da fonte';

  @override
  String get lineHeight => 'Espaçamento';

  @override
  String get pageMode => 'Modo de página';

  @override
  String get scrollMode => 'Rolagem vertical';

  @override
  String get pagedMode => 'Virar página';

  @override
  String get offline => 'Cache offline';

  @override
  String get vipActive => 'Membro VIP';

  @override
  String get vipNotActive => 'VIP não ativado';

  @override
  String vipExpiresAt(Object date) {
    return 'Expira em: $date';
  }

  @override
  String get vipRenew => 'Renovar';

  @override
  String get vipOpen => 'Obter VIP';

  @override
  String get payNow => 'Pagar agora';

  @override
  String get paymentResult => 'Resultado do pagamento';

  @override
  String get paymentSuccess => 'Pagamento realizado';

  @override
  String get paymentFailed => 'Falha no pagamento';

  @override
  String get paymentPending => 'Aguardando pagamento';

  @override
  String get paymentChecking => 'Confirmando pagamento…';

  @override
  String get paymentNoMethod => 'Nenhum método de pagamento disponível';

  @override
  String get vipChapterLocked => 'Este capítulo é exclusivo para membros VIP.';

  @override
  String get openVipToRead => 'Torne-se membro para ler';

  @override
  String get retryPay => 'Pagar novamente';
}
