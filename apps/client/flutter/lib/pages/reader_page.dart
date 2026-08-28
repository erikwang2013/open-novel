import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import '../reader_settings.dart';
import 'comments_page.dart';
import 'vip_page.dart';

/// 阅读器：章节正文 + 上一章 / 下一章（GET /api/chapters/{id}/content）。
/// 设置入口：AppBar ⋮ 菜单（字号 / 行距 / 主题 / 翻页方式），持久化见 ReaderSettings。
/// 进度：滚动模式存滚动偏移、翻页模式存页码（T-C-12，position 复用后端 uint32）。
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
  bool _fromCache = false;
  bool _vipLocked = false; // VIP 章节未订阅 → 引导购买

  final ScrollController _scroll = ScrollController();
  final PageController _page = PageController();
  List<String> _pages = const [];
  int _pageIndex = 0;
  int _restore = -1; // 待恢复的滚动偏移；-1 表示无
  String _pageKey = '';

  @override
  void initState() {
    super.initState();
    _load();
    _checkProgress();
  }

  @override
  void dispose() {
    _scroll.dispose();
    _page.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _error = null;
      _content = null;
      _vipLocked = false;
      _pages = const [];
      _pageKey = '';
      _pageIndex = 0;
      _restore = -1;
    });
    // VIP 章节未订阅 → 引导购买，不拉正文（后端未做 VIP 拦截，客户端自行判断）
    if (_chapter.isVip == 1 && !await _isVipActive()) {
      if (!mounted) return;
      setState(() => _vipLocked = true);
      return;
    }
    try {
      final lang = langCode(localeNotifier.value);
      final c = await ApiClient.instance
          .getChapterContentCached(_chapter.id, lang: lang);
      if (!mounted) return;
      setState(() {
        _content = c;
        _fromCache = c.fromCache;
      });
      _saveProgress();
      _restorePosition();
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = ApiClient.instance.errorMessage(e));
    }
  }

  /// 是否 VIP 会员：未登录视为非会员；接口失败也视为非会员（保守引导）。
  Future<bool> _isVipActive() async {
    final api = ApiClient.instance;
    if (!api.loggedIn) return false;
    try {
      return (await api.vipStatus()).active;
    } catch (_) {
      return false;
    }
  }

  /// 打开 VIP 购买页；返回后重新检查（可能已开通）。
  Future<void> _openVipGuide() async {
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => const VipPage()),
    );
    if (mounted) _load();
  }

  /// best-effort 保存进度（已登录才调，失败静默不打扰阅读）。
  /// 滚动模式存像素偏移，翻页模式存页码（T-C-12）。
  void _saveProgress([int? pos]) {
    final api = ApiClient.instance;
    if (!api.loggedIn || widget.bookId.isEmpty) return;
    final position = pos ??
        (ReaderSettings.pageMode.value == 1
            ? _pageIndex
            : (_scroll.hasClients ? _scroll.offset.round() : 0));
    api
        .updateProgress(widget.bookId, _chapter.id, position: position)
        .catchError((_) {
      // 保存失败静默
    });
  }

  /// 恢复同章节的精确位置（滚动偏移 / 页码）。
  Future<void> _restorePosition() async {
    final api = ApiClient.instance;
    if (!api.loggedIn || widget.bookId.isEmpty) return;
    final p = await api.getProgress(widget.bookId);
    if (!mounted || p == null || p.chapterId != _chapter.id) return;
    if (ReaderSettings.pageMode.value == 1) {
      _pageIndex = p.position;
    } else {
      _restore = p.position;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) => _applyRestore());
  }

  void _applyRestore() {
    if (!mounted) return;
    if (ReaderSettings.pageMode.value == 1) {
      if (_page.hasClients && _pages.isNotEmpty) {
        final max = _pages.length - 1;
        _page.jumpToPage(_pageIndex > max ? max : _pageIndex);
      }
    } else if (_scroll.hasClients && _restore >= 0) {
      final max = _scroll.position.maxScrollExtent;
      _scroll.jumpTo(_restore > max ? max : _restore.toDouble());
      _restore = -1;
    }
  }

  /// 进入阅读器后有已存进度则提示可跳转（仅章节不同时提示）。
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
      _pages = const [];
      _pageKey = '';
      _pageIndex = 0;
      _restore = -1;
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
          PopupMenuButton<String>(
            icon: const Icon(Icons.tune),
            tooltip: l10n.settings,
            onSelected: _onSetting,
            itemBuilder: (context) => _settingItems(l10n),
          ),
        ],
      ),
      body: _buildBody(context, l10n, idx),
    );
  }

  List<PopupMenuEntry<String>> _settingItems(AppLocalizations l10n) {
    final fs = ReaderSettings.fontSize.value;
    final lh = ReaderSettings.lineHeight.value;
    final paged = ReaderSettings.pageMode.value == 1;
    final tm = ReaderSettings.themeMode.value;
    return [
      PopupMenuItem<String>(
          enabled: false,
          child:
              Text('${l10n.fontSize}：${fs.toStringAsFixed(0)}')),
      PopupMenuItem<String>(value: 'font-', child: Text('${l10n.fontSize} −')),
      PopupMenuItem<String>(value: 'font+', child: Text('${l10n.fontSize} +')),
      const PopupMenuDivider(),
      PopupMenuItem<String>(
          enabled: false,
          child: Text('${l10n.lineHeight}：${lh.toStringAsFixed(1)}')),
      PopupMenuItem<String>(value: 'lh-', child: Text('${l10n.lineHeight} −')),
      PopupMenuItem<String>(value: 'lh+', child: Text('${l10n.lineHeight} +')),
      const PopupMenuDivider(),
      CheckedPopupMenuItem<String>(
          value: 'mode-scroll',
          checked: !paged,
          child: Text(l10n.scrollMode)),
      CheckedPopupMenuItem<String>(
          value: 'mode-paged',
          checked: paged,
          child: Text(l10n.pagedMode)),
      const PopupMenuDivider(),
      CheckedPopupMenuItem<String>(
          value: 'theme-system',
          checked: tm == ThemeMode.system,
          child: Text(l10n.themeSystem)),
      CheckedPopupMenuItem<String>(
          value: 'theme-light',
          checked: tm == ThemeMode.light,
          child: Text(l10n.themeLight)),
      CheckedPopupMenuItem<String>(
          value: 'theme-dark',
          checked: tm == ThemeMode.dark,
          child: Text(l10n.themeDark)),
    ];
  }

  void _onSetting(String v) {
    switch (v) {
      case 'font-':
        ReaderSettings.setFontSize((ReaderSettings.fontSize.value - 1)
            .clamp(ReaderSettings.fontSizeMin, ReaderSettings.fontSizeMax)
            .toDouble());
        break;
      case 'font+':
        ReaderSettings.setFontSize((ReaderSettings.fontSize.value + 1)
            .clamp(ReaderSettings.fontSizeMin, ReaderSettings.fontSizeMax)
            .toDouble());
        break;
      case 'lh-':
        ReaderSettings.setLineHeight((ReaderSettings.lineHeight.value - 0.2)
            .clamp(ReaderSettings.lineHeightMin, ReaderSettings.lineHeightMax)
            .toDouble());
        break;
      case 'lh+':
        ReaderSettings.setLineHeight((ReaderSettings.lineHeight.value + 0.2)
            .clamp(ReaderSettings.lineHeightMin, ReaderSettings.lineHeightMax)
            .toDouble());
        break;
      case 'mode-scroll':
        ReaderSettings.setPageMode(0);
        break;
      case 'mode-paged':
        ReaderSettings.setPageMode(1);
        break;
      case 'theme-system':
        ReaderSettings.setThemeMode(ThemeMode.system);
        break;
      case 'theme-light':
        ReaderSettings.setThemeMode(ThemeMode.light);
        break;
      case 'theme-dark':
        ReaderSettings.setThemeMode(ThemeMode.dark);
        break;
    }
  }

  Widget _buildBody(BuildContext context, AppLocalizations l10n, int idx) {
    final Widget body;
    if (_vipLocked) {
      body = Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.lock_outline,
                size: 64, color: Theme.of(context).colorScheme.primary),
            const SizedBox(height: 16),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Text(l10n.vipChapterLocked,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.titleMedium),
            ),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: _openVipGuide,
              child: Text(l10n.openVipToRead),
            ),
          ],
        ),
      );
    } else if (_error != null) {
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
      body = Column(
        children: [
          if (_fromCache)
            Container(
              width: double.infinity,
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              padding: const EdgeInsets.symmetric(vertical: 4),
              child: Center(
                child: Text(l10n.offline,
                    style: Theme.of(context).textTheme.bodySmall),
              ),
            ),
          Expanded(child: _readerBody()),
        ],
      );
    }
    // 宽屏居中限宽 820，保持适宜行宽（移动端宽度 < 820 时约束不生效，零回归）。
    // ponytail: 未做左章节目录右正文的桌面布局，宽屏阅读器仅限行宽；需要时再拆列。
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 820),
        child: Column(
          children: [
            Expanded(child: body),
            SafeArea(
              child: Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
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
        ),
      ),
    );
  }

  /// 正文渲染：跟随 字号/行距/翻页方式 设置即时生效。
  Widget _readerBody() {
    return ListenableBuilder(
      listenable: Listenable.merge([
        ReaderSettings.fontSize,
        ReaderSettings.lineHeight,
        ReaderSettings.pageMode,
      ]),
      builder: (context, _) {
        final style = TextStyle(
          fontSize: ReaderSettings.fontSize.value,
          height: ReaderSettings.lineHeight.value,
        );
        if (ReaderSettings.pageMode.value == 1) {
          return LayoutBuilder(
            builder: (context, c) {
              final key =
                  '${_content!.content.length}:${style.fontSize}x${style.height}:'
                  '${c.maxWidth}x${c.maxHeight}';
              if (key != _pageKey) {
                _pageKey = key;
                _pages = _paginate(_content!.content, style, c.maxWidth, c.maxHeight);
                if (_pageIndex >= _pages.length) {
                  _pageIndex = _pages.isEmpty ? 0 : _pages.length - 1;
                }
                WidgetsBinding.instance.addPostFrameCallback((_) {
                  if (mounted && _page.hasClients && _pages.isNotEmpty) {
                    final max = _pages.length - 1;
                    _page.jumpToPage(_pageIndex > max ? max : _pageIndex);
                  }
                });
              }
              return PageView.builder(
                controller: _page,
                itemCount: _pages.length,
                onPageChanged: (i) {
                  setState(() => _pageIndex = i);
                  _saveProgress(i);
                },
                itemBuilder: (context, i) => Padding(
                  padding: const EdgeInsets.all(20),
                  child: SelectableText(_pages[i], style: style),
                ),
              );
            },
          );
        }
        return NotificationListener<ScrollEndNotification>(
          onNotification: (n) {
            if (n.depth == 0) _saveProgress();
            return false;
          },
          child: SingleChildScrollView(
            controller: _scroll,
            padding: const EdgeInsets.all(20),
            child: SelectableText(_content!.content, style: style),
          ),
        );
      },
    );
  }

  /// 按可用尺寸用 TextPainter 精确分页（每页换行回绕，页间可切断单词）。
  /// ponytail: 每页重建 TextPainter O(n²)，章节 <1 万字无感；超长再优化。
  List<String> _paginate(String text, TextStyle style, double w, double h) {
    final pages = <String>[];
    var start = 0;
    while (start < text.length) {
      final tp = TextPainter(
        text: TextSpan(text: text.substring(start), style: style),
        textDirection: Directionality.of(context),
      )..layout(maxWidth: w);
      if (tp.height <= h) {
        pages.add(text.substring(start));
        break;
      }
      final end = tp.getPositionForOffset(Offset(0, h - 1)).offset;
      final cut = end > start ? end : start + 1;
      pages.add(text.substring(start, cut));
      start = cut;
    }
    return pages;
  }
}
