import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import 'comments_page.dart';

/// 阅读器：章节正文 + 上一章 / 下一章（GET /api/chapters/{id}/content）。
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
    _checkProgress();
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
      _saveProgress();
    } catch (e) {
      setState(() => _error = ApiClient.instance.errorMessage(e));
    }
  }

  /// best-effort 保存进度（已登录才调，失败静默不打扰阅读）。
  /// ponytail: 只存章节粒度（position=0），不做滚动位置跟踪。
  Future<void> _saveProgress() async {
    final api = ApiClient.instance;
    if (!api.loggedIn || widget.bookId.isEmpty) return;
    try {
      await api.updateProgress(widget.bookId, _chapter.id);
    } catch (_) {
      // 保存失败静默
    }
  }

  /// 进入阅读器后有已存进度则提示可跳转。
  Future<void> _checkProgress() async {
    final api = ApiClient.instance;
    if (!api.loggedIn || widget.bookId.isEmpty) return;
    final p = await api.getProgress(widget.bookId);
    if (p == null || p.chapterId.isEmpty || p.chapterId == _chapter.id) return;
    final idx = widget.chapters.indexWhere((c) => c.id == p.chapterId);
    if (idx < 0 || !mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(
      content: Text('上次读到《${widget.chapters[idx].title}》，是否继续？'),
      duration: const Duration(seconds: 6),
      action: SnackBarAction(label: '继续', onPressed: () => _goto(idx)),
    ));
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
