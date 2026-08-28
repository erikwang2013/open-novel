import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import 'comments_page.dart';
import 'reader_page.dart';

/// 书籍详情页：书籍信息 + 章节列表 + 评论入口。
class BookDetailPage extends StatefulWidget {
  const BookDetailPage({
    super.key,
    required this.bookId,
    this.title = '',
    this.author = '',
    this.summary = '',
  });

  final String bookId;
  final String title;
  final String author;
  final String summary;

  @override
  State<BookDetailPage> createState() => _BookDetailPageState();
}

class _BookDetailPageState extends State<BookDetailPage> {
  Book? _book;
  List<Chapter>? _chapters;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _error = null;
      _chapters = null;
    });
    try {
      final lang = langCode(localeNotifier.value);
      final book = await ApiClient.instance.getBook(widget.bookId, lang: lang);
      final chapters =
          await ApiClient.instance.listChapters(widget.bookId, lang: lang);
      setState(() {
        _book = book;
        _chapters = chapters;
      });
    } catch (e) {
      setState(() => _error = ApiClient.instance.errorMessage(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final book = _book;
    return Scaffold(
      appBar: AppBar(
        title: Text(book?.title ?? widget.title,
            overflow: TextOverflow.ellipsis),
        actions: [
          IconButton(
            icon: const Icon(Icons.chat_bubble_outline),
            tooltip: l10n.comments,
            onPressed: () => Navigator.of(context).push(MaterialPageRoute(
              builder: (_) => CommentsPage(bookId: widget.bookId),
            )),
          ),
        ],
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
    final book = _book;
    final chapters = _chapters;
    if (book == null || chapters == null) {
      return const Center(child: CircularProgressIndicator());
    }
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 96,
              height: 128,
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primaryContainer,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(Icons.menu_book,
                  size: 48, color: Theme.of(context).colorScheme.primary),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(book.title,
                      style: Theme.of(context)
                          .textTheme
                          .titleLarge
                          ?.copyWith(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  Text(book.author,
                      style: Theme.of(context).textTheme.bodyMedium),
                  const SizedBox(height: 8),
                  Text(book.summary,
                      maxLines: 4,
                      overflow: TextOverflow.ellipsis,
                      style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
          ],
        ),
        const SizedBox(height: 20),
        Text(l10n.chapters,
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(fontWeight: FontWeight.bold)),
        const SizedBox(height: 8),
        if (chapters.isEmpty)
          Center(child: Text(l10n.empty))
        else
          ...chapters.map((c) => ListTile(
                dense: true,
                leading: Text('${c.chapterNo}',
                    style: Theme.of(context).textTheme.bodySmall),
                title: Text(c.title,
                    maxLines: 1, overflow: TextOverflow.ellipsis),
                trailing: c.isVip == 1 ? Text(l10n.vip) : null,
                onTap: () => Navigator.of(context).push(MaterialPageRoute(
                  builder: (_) => ReaderPage(
                      chapter: c, chapters: chapters, bookId: widget.bookId),
                )),
              )),
      ],
    );
  }
}
