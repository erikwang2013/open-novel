import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';

import 'api/api_client.dart';
import 'l10n/app_localizations.dart';
import 'pages/book_detail_page.dart';
import 'pages/books_tab.dart';
import 'pages/home_tab.dart';
import 'pages/login_page.dart';
import 'pages/mine_tab.dart';
import 'reader_settings.dart';

/// 全局语言切换（zh / en），随 MaterialApp 重建生效。
final ValueNotifier<String> localeNotifier = ValueNotifier('zh');

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await ReaderSettings.init();
  runApp(const OpenNovelApp());
}

class OpenNovelApp extends StatelessWidget {
  const OpenNovelApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ValueListenableBuilder<String>(
      valueListenable: localeNotifier,
      builder: (context, locale, _) {
        return ValueListenableBuilder<ThemeMode>(
          valueListenable: ReaderSettings.themeMode,
          builder: (context, mode, _) => MaterialApp(
            title: 'Open Novel',
            theme: ThemeData(
              colorScheme: ColorScheme.fromSeed(seedColor: Colors.indigo),
              useMaterial3: true,
            ),
            darkTheme: ThemeData(
              colorScheme: ColorScheme.fromSeed(
                seedColor: Colors.indigo,
                brightness: Brightness.dark,
              ),
              useMaterial3: true,
            ),
            themeMode: mode,
          locale: Locale(locale),
          supportedLocales: const [
            Locale('zh'),
            Locale('en'),
            Locale('ja'),
            Locale('ko'),
            Locale('fr'),
            Locale('de'),
            Locale('es'),
            Locale('ru'),
            Locale('ar'),
            Locale('pt'),
            Locale('hi'),
            Locale('bn'),
            Locale('id'),
          ],
          localizationsDelegates: const [
            AppLocalizations.delegate,
            GlobalMaterialLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
          ],
            home: const MainShell(),
          ),
        );
      },
    );
  }
}

class MainShell extends StatefulWidget {
  const MainShell({super.key});

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      body: IndexedStack(
        index: _index,
        children: const [HomeTab(), BooksTab(), MineTab()],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: [
          NavigationDestination(
              icon: const Icon(Icons.home_outlined),
              selectedIcon: const Icon(Icons.home),
              label: l10n.home),
          NavigationDestination(
              icon: const Icon(Icons.menu_book_outlined),
              selectedIcon: const Icon(Icons.menu_book),
              label: l10n.allBooks),
          NavigationDestination(
              icon: const Icon(Icons.person_outline),
              selectedIcon: const Icon(Icons.person),
              label: l10n.mine),
        ],
      ),
    );
  }
}

/// 全局帮助函数：校验登录态，未登录跳登录页。
Future<bool> ensureLogin(BuildContext context) async {
  final api = ApiClient.instance;
  if (api.loggedIn) return true;
  if (!context.mounted) return false;
  await Navigator.of(context).push(MaterialPageRoute(
    builder: (_) => const LoginPage(),
  ));
  return api.loggedIn;
}

/// 跳转书籍详情。
void openBook(BuildContext context,
    {String? id, String? title, String? author, String? summary}) {
  Navigator.of(context).push(MaterialPageRoute(
    builder: (_) => BookDetailPage(
      bookId: id ?? '',
      title: title ?? '',
      author: author ?? '',
      summary: summary ?? '',
    ),
  ));
}
