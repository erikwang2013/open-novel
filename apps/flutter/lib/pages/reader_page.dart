import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import 'comments_page.dart';

/// 阅读器：章节正文 + 上一章 / 下一章（GET /api/v1/chapters/{id}/content）。
class ReaderPage extends StatefulWidget {
  const ReaderPage({
    super.key,
    required this.chapter,
    required this.chapters,
    this.bookId = '',
  });

  final Chapter chapter;
  final List<Chapter> chapters;
  final String bookId;

  @override
  State<ReaderPage> createState() => _ReaderPageState();
}

class _ReaderPageState extends State<ReaderPage> {
  late Chapter _chapter = widget.chapter;
  ChapterContent? _content;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _error = null;
      _content = null;
    });
    try {
      final lang = langCode(localeNotifier.value);
      final c =
          await ApiClient.instance.getChapterContent(_chapter.id, lang: lang);
      setState(() => _content = c);
    } catch (e) {
      setState(() => _error = ApiClient.instance.errorMessage(e));
    }
  }

  void _goto(int index) {
    setState(() {
      _chapter = widget.chapters[index];
      _content = null;
      _error = null;
    });
    _load();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final idx = widget.chapters.indexWhere((c) => c.id == _chapter.id);
    return Scaffold(
      appBar: AppBar(
        title: Text('${_chapter.chapterNo}. ${_chapter.title}',
            overflow: TextOverflow.ellipsis),
        actions: [
          IconButton(
            icon: const Icon(Icons.chat_bubble_outline),
            tooltip: l10n.comments,
            onPressed: () => Navigator.of(context).push(MaterialPageRoute(
              builder: (_) =>
                  CommentsPage(bookId: widget.bookId, chapterId: _chapter.id),
            )),
          ),
        ],
      ),
      body: _buildBody(context, l10n, idx),
    );
  }

  Widget _buildBody(BuildContext context, AppLocalizations l10n, int idx) {
    final Widget body;
    if (_error != null) {
      body = Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error == 'network' ? l10n.errorNetwork : l10n.errorServer(_error!)),
            TextButton(onPressed: _load, child: Text(l10n.retry)),
          ],
        ),
      );
    } else if (_content == null) {
      body = const Center(child: CircularProgressIndicator());
    } else {
      body = SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: SelectableText(
          _content!.content,
          style: const TextStyle(fontSize: 17, height: 1.8),
        ),
      );
    }
    return Column(
      children: [
        Expanded(child: body),
        SafeArea(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Row(
              children: [
                Expanded(
                  child: FilledButton.tonal(
                    onPressed: idx > 0 ? () => _goto(idx - 1) : null,
                    child: Text(l10n.prevChapter),
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: FilledButton.tonal(
                    onPressed: idx >= 0 && idx < widget.chapters.length - 1
                        ? () => _goto(idx + 1)
                        : null,
                    child: Text(l10n.nextChapter),
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}
