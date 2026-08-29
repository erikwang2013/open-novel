import 'dart:async';

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import '../widgets/book_card.dart';

/// 全部书籍 tab：搜索框 + 书籍列表（GET /api/books、GET /api/search）。
class BooksTab extends StatefulWidget {
  const BooksTab({super.key});

  @override
  State<BooksTab> createState() => _BooksTabState();
}

class _BooksTabState extends State<BooksTab> {
  final _searchCtrl = TextEditingController();
  List<Book>? _books;
  List<SearchDoc>? _searchResults;
  List<Category> _categories = [];
  List<HotKeyword> _hot = [];
  String? _categoryId;
  String? _error;
  bool _searching = false;

  static const _kSearchHistory = 'search_history';
  List<String> _history = [];
  List<String> _suggestions = [];
  Timer? _debounce;

  @override
  void initState() {
    super.initState();
    _loadBooks();
    // 分类 / 热搜为增强功能，失败静默不影响书单浏览
    _loadMeta();
    _loadHistory();
  }

  Future<void> _loadMeta() async {
    try {
      final cats = await ApiClient.instance.listCategories();
      final hot = await ApiClient.instance.hotSearches();
      if (!mounted) return;
      setState(() {
        _categories = cats;
        _hot = hot;
      });
    } catch (_) {
      // 静默
    }
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _searchCtrl.dispose();
    super.dispose();
  }

  Future<void> _loadHistory() async {
    final p = await SharedPreferences.getInstance();
    final h = p.getStringList(_kSearchHistory) ?? const [];
    if (mounted && h.isNotEmpty) setState(() => _history = h);
  }

  /// 新搜索插入头部、去重，最多 20 条（最新在前）。
  void _addHistory(String q) {
    final h = [q, ..._history.where((e) => e != q)];
    if (h.length > 20) h.removeRange(20, h.length);
    setState(() => _history = h);
    SharedPreferences.getInstance()
        .then((p) => p.setStringList(_kSearchHistory, h));
  }

  void _clearHistory() {
    setState(() => _history = []);
    SharedPreferences.getInstance().then((p) => p.remove(_kSearchHistory));
  }

  /// 输入防抖 200ms 后拉搜索建议，空输入取消。
  void _onQueryChanged(String q) {
    _debounce?.cancel();
    final query = q.trim();
    if (query.isEmpty) {
      if (_suggestions.isNotEmpty) setState(() => _suggestions = []);
      return;
    }
    _debounce = Timer(const Duration(milliseconds: 200),
        () => _loadSuggestions(query));
  }

  Future<void> _loadSuggestions(String q) async {
    try {
      final list = await ApiClient.instance.suggest(q);
      // 响应过期（输入已变/已提交搜索）则丢弃
      if (!mounted || _searchCtrl.text.trim() != q) return;
      setState(() => _suggestions = list);
    } catch (_) {
      // 建议失败静默，不影响搜索
    }
  }

  Future<void> _loadBooks() async {
    setState(() {
      _error = null;
      _books = null;
      _searching = false;
      _searchResults = null;
    });
    try {
      final lang = langCode(localeNotifier.value);
      final categoryId = _categoryId;
      final books = await ApiClient.instance
          .listBooks(pageSize: 20, lang: lang, categoryId: categoryId);
      setState(() => _books = books);
    } catch (e) {
      setState(() => _error = ApiClient.instance.errorMessage(e));
    }
  }

  Future<void> _search(String q) async {
    final query = q.trim();
    if (query.isEmpty) return;
    _addHistory(query);
    setState(() {
      _searching = true;
      _error = null;
    });
    try {
      final lang = langCode(localeNotifier.value);
      final results = await ApiClient.instance.search(query, lang: lang);
      setState(() {
        _searchResults = results;
        _searching = false;
      });
    } catch (e) {
      setState(() {
        _error = ApiClient.instance.errorMessage(e);
        _searching = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: TextField(
          controller: _searchCtrl,
          textInputAction: TextInputAction.search,
          decoration: InputDecoration(
            hintText: l10n.searchHint,
            border: InputBorder.none,
            suffixIcon: IconButton(
              icon: const Icon(Icons.search),
              onPressed: () => _search(_searchCtrl.text.trim()),
            ),
          ),
          onChanged: _onQueryChanged,
          onSubmitted: _search,
        ),
      ),
      body: _buildBody(context, l10n),
    );
  }

  Widget _buildBody(BuildContext context, AppLocalizations l10n) {
    if (_searching) return const Center(child: CircularProgressIndicator());
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error == 'network' ? l10n.errorNetwork : l10n.errorServer(_error!)),
            OutlinedButton(
                onPressed: () => _searchCtrl.text.isEmpty
                    ? _loadBooks()
                    : _search(_searchCtrl.text.trim()),
                child: Text(l10n.retry)),
          ],
        ),
      );
    }
    final results = _searchResults;
    if (results != null) {
      if (results.isEmpty) return Center(child: Text(l10n.empty));
      return ListView.builder(
        padding: const EdgeInsets.all(8),
        itemCount: results.length,
        itemBuilder: (_, i) {
          final d = results[i];
          final lang = langCode(localeNotifier.value);
          return BookCard(
            bookId: d.bookId,
            title: d.title(lang),
            author: d.author(lang),
            summary: d.summary(lang),
          );
        },
      );
    }
    // 输入中（未提交）：搜索框下方展示「搜索建议 + 搜索历史」
    if (_searchCtrl.text.trim().isNotEmpty) {
      return _buildSearchPanel(context, l10n);
    }
    final books = _books;
    if (books == null) return const Center(child: CircularProgressIndicator());
    if (books.isEmpty) return Center(child: Text(l10n.empty));
    final topCats = _categories.where((c) => c.parentId == 0).toList();
    final header = <Widget>[
      if (topCats.isNotEmpty)
        SizedBox(
          height: 40,
          child: ListView(
            scrollDirection: Axis.horizontal,
            children: [
              _catChip(l10n.all, null),
              for (final c in topCats) _catChip(c.name, c.id),
            ],
          ),
        ),
      if (_hot.isNotEmpty) ...[
        Padding(
          padding: const EdgeInsets.only(top: 8, bottom: 4),
          child: Text(l10n.hotSearch,
              style: Theme.of(context).textTheme.titleSmall),
        ),
        Wrap(
          spacing: 8,
          runSpacing: 4,
          children: [
            for (final w in _hot)
              ActionChip(
                label: Text(w.keyword),
                onPressed: () => _search(w.keyword),
              ),
          ],
        ),
      ],
      const SizedBox(height: 8),
    ];
    Widget bookCard(Book b) => BookCard(
          bookId: b.id,
          title: b.title,
          author: b.author,
          summary: b.summary,
          vip: b.isVip == 1,
        );
    return LayoutBuilder(
      builder: (context, c) {
        if (c.maxWidth >= 900) {
          // 宽屏（>=900）：分类/热搜头部 + 多列网格（T-C-20）
          return RefreshIndicator(
            onRefresh: _loadBooks,
            child: CustomScrollView(
              slivers: [
                SliverToBoxAdapter(child: Column(children: header)),
                SliverPadding(
                  padding: const EdgeInsets.fromLTRB(8, 0, 8, 8),
                  sliver: SliverGrid(
                    gridDelegate:
                        const SliverGridDelegateWithMaxCrossAxisExtent(
                      maxCrossAxisExtent: 280,
                      mainAxisExtent: 130,
                      crossAxisSpacing: 8,
                      mainAxisSpacing: 8,
                    ),
                    delegate: SliverChildBuilderDelegate(
                      (_, i) => bookCard(books[i]),
                      childCount: books.length,
                    ),
                  ),
                ),
              ],
            ),
          );
        }
        // 移动端：单列列表（原布局，不得回归）
        return RefreshIndicator(
          onRefresh: _loadBooks,
          child: ListView(
            padding: const EdgeInsets.all(8),
            children: [...header, ...books.map(bookCard)],
          ),
        );
      },
    );
  }

  /// 建议优先、历史其次；无内容返回空（不显示区块）。
  Widget _buildSearchPanel(BuildContext context, AppLocalizations l10n) {
    final suggest = _suggestions;
    final history = _history;
    if (suggest.isEmpty && history.isEmpty) {
      return const SizedBox.shrink();
    }
    Widget section(String title, List<String> words, {Widget? trailing}) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.only(top: 12, bottom: 4),
            child: Row(
              children: [
                Expanded(
                  child: Text(title,
                      style: Theme.of(context).textTheme.titleSmall),
                ),
                ?trailing,
              ],
            ),
          ),
          Wrap(
            spacing: 8,
            runSpacing: 4,
            children: [
              for (final w in words)
                ActionChip(
                  label: Text(w),
                  onPressed: () {
                    _searchCtrl.text = w;
                    _search(w);
                  },
                ),
            ],
          ),
        ],
      );
    }

    return ListView(
      padding: const EdgeInsets.all(8),
      children: [
        if (suggest.isNotEmpty) section(l10n.searchSuggest, suggest),
        if (history.isNotEmpty)
          section(
            l10n.searchHistory,
            history,
            trailing: IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: l10n.clearHistory,
              onPressed: _clearHistory,
            ),
          ),
      ],
    );
  }

  Widget _catChip(String label, String? id) {
    final sel = _categoryId == id;
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: ChoiceChip(
        label: Text(label),
        selected: sel,
        onSelected: (_) {
          setState(() => _categoryId = id);
          _loadBooks();
        },
      ),
    );
  }
}
