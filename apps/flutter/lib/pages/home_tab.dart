import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import '../widgets/book_card.dart';

/// 首页 tab：推荐列表（GET /api/v1/recommend?strategy=hot）。
class HomeTab extends StatefulWidget {
  const HomeTab({super.key});

  @override
  State<HomeTab> createState() => _HomeTabState();
}

class _HomeTabState extends State<HomeTab> {
  List<RecommendItem>? _items;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _error = null;
      _items = null;
    });
    try {
      final lang = langCode(localeNotifier.value);
      final items =
          await ApiClient.instance.recommend(strategy: 'hot', lang: lang);
      setState(() => _items = items);
    } catch (e) {
      setState(() => _error = ApiClient.instance.errorMessage(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.recommend),
        actions: const [_LangSwitch()],
      ),
      body: _buildBody(context, l10n),
    );
  }

  Widget _buildBody(BuildContext context, AppLocalizations l10n) {
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error == 'network' ? l10n.errorNetwork : l10n.errorServer(_error!)),
            TextButton(onPressed: _load, child: Text(l10n.retry)),
          ],
        ),
      );
    }
    final items = _items;
    if (items == null) return const Center(child: CircularProgressIndicator());
    if (items.isEmpty) return Center(child: Text(l10n.empty));
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.builder(
        padding: const EdgeInsets.all(8),
        itemCount: items.length,
        itemBuilder: (_, i) => BookCard(
          bookId: items[i].bookId,
          title: items[i].title,
          author: items[i].author,
          summary: items[i].summary,
        ),
      ),
    );
  }
}

/// 顶部语言切换（zh/en）。
class _LangSwitch extends StatelessWidget {
  const _LangSwitch();

  @override
  Widget build(BuildContext context) {
    final cur = localeNotifier.value;
    return TextButton(
      onPressed: () => localeNotifier.value = cur == 'zh' ? 'en' : 'zh',
      child: Text(cur == 'zh' ? 'EN' : '中文'),
    );
  }
}
