import 'package:flutter/material.dart';

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
  List<SearchDoc> _hot = [];
  String? _categoryId;
  String? _error;
  bool _searching = false;

  @override
  void initState() {
    super.initState();
    _loadBooks();
    // 分类 / 热搜为增强功能，失败静默不影响书单浏览
    _loadMeta();
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
    _searchCtrl.dispose();
    super.dispose();
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
    setState(() {
      _searching = true;
      _error = null;
    });
    try {
      final lang = langCode(localeNotifier.value);
      final results = await ApiClient.instance.search(q, lang: lang);
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
            TextButton(
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
            for (final d in _hot)
              ActionChip(
                label: Text(d.title(langCode(localeNotifier.value))),
                onPressed: () => _search(d.title(langCode(localeNotifier.value))),
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
