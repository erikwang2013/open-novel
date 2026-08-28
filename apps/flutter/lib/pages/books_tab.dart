import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import '../widgets/book_card.dart';

/// 全部书籍 tab：搜索框 + 书籍列表（GET /api/v1/books、GET /api/v1/search）。
class BooksTab extends StatefulWidget {
  const BooksTab({super.key});

  @override
  State<BooksTab> createState() => _BooksTabState();
}

class _BooksTabState extends State<BooksTab> {
  final _searchCtrl = TextEditingController();
  List<Book>? _books;
  List<SearchDoc>? _searchResults;
  String? _error;
  bool _searching = false;

  @override
  void initState() {
    super.initState();
    _loadBooks();
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
      final books =
          await ApiClient.instance.listBooks(pageSize: 20, lang: lang);
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
    return RefreshIndicator(
      onRefresh: _loadBooks,
      child: ListView.builder(
        padding: const EdgeInsets.all(8),
        itemCount: books.length,
        itemBuilder: (_, i) => BookCard(
          bookId: books[i].id,
          title: books[i].title,
          author: books[i].author,
          summary: books[i].summary,
          vip: books[i].isVip == 1,
        ),
      ),
    );
  }
}
