import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import 'login_page.dart';
import 'reader_page.dart';
import 'vip_page.dart';

/// 我的 tab：未登录显示登录入口；已登录显示书架、收藏与最近阅读进度。
class MineTab extends StatefulWidget {
  const MineTab({super.key});

  @override
  State<MineTab> createState() => _MineTabState();
}

class _MineTabState extends State<MineTab> {
  List<ShelfItem>? _shelf;
  List<FavoriteItem>? _favorites;
  final Map<String, String> _titles = {}; // bookId -> 书名
  final Map<String, List<Chapter>> _chapters = {}; // bookId -> 章节列表
  final Map<String, Chapter?> _progressChapter = {}; // bookId -> 进度章节
  final Map<String, String> _progressAt = {}; // bookId -> 进度更新时间
  VipStatus? _vip;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final api = ApiClient.instance;
    setState(() {
      _error = null;
      _shelf = null;
      _favorites = null;
      _vip = null;
    });
    if (!api.loggedIn) return;
    try {
      final shelf = await api.listBookshelf(pageSize: 20);
      final favorites = await api.listFavorites(pageSize: 20);
      await _fillTitles({
        ...shelf.map((e) => e.bookId),
        ...favorites.map((e) => e.bookId),
      });
      await _fillProgress(shelf);
      final vip = await api.vipStatus();
      if (!mounted) return;
      setState(() {
        _shelf = shelf;
        _favorites = favorites;
        _vip = vip;
      });
    } catch (e) {
      setState(() => _error = api.errorMessage(e));
    }
  }

  /// 批量补齐书名（书架 + 收藏去重），失败留空不阻塞。
  Future<void> _fillTitles(Set<String> ids) async {
    final api = ApiClient.instance;
    for (final id in ids.where((id) => !_titles.containsKey(id))) {
      try {
        _titles[id] = (await api.getBook(id)).title;
      } catch (_) {
        // 单本失败静默，重试时会再取
      }
    }
  }

  /// 每本书架书拉取章节 + 进度，用于显示 chapterNo 与「阅读」跳转。
  Future<void> _fillProgress(List<ShelfItem> shelf) async {
    final api = ApiClient.instance;
    for (final s in shelf) {
      try {
        final chapters = await api.fetchAllChapters(s.bookId);
        _chapters[s.bookId] = chapters;
        final p = await api.getProgress(s.bookId);
        if (p != null && p.chapterId.isNotEmpty) {
          final idx = chapters.indexWhere((c) => c.id == p.chapterId);
          _progressChapter[s.bookId] = idx >= 0 ? chapters[idx] : null;
          _progressAt[s.bookId] = p.updatedAt;
        }
      } catch (_) {
        // 进度 / 章节失败静默
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final user = ApiClient.instance.currentUser;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.mine)),
      body: user == null ? _loginView(context, l10n) : _loggedInView(context),
    );
  }

  Widget _loginView(BuildContext context, AppLocalizations l10n) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          CircleAvatar(
            radius: 40,
            child: Icon(Icons.person,
                size: 44, color: Theme.of(context).colorScheme.primary),
          ),
          const SizedBox(height: 16),
          Text(l10n.notLoggedIn,
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: () async {
              await Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const LoginPage()),
              );
              if (mounted) _load();
            },
            child: Text(l10n.login),
          ),
        ],
      ),
    );
  }

  Widget _loggedInView(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error == 'network'
                ? l10n.errorNetwork
                : l10n.errorServer(_error!)),
            TextButton(onPressed: _load, child: Text(l10n.retry)),
          ],
        ),
      );
    }
    final shelf = _shelf;
    final favorites = _favorites;
    if (shelf == null || favorites == null) {
      return const Center(child: CircularProgressIndicator());
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(8),
        children: [
          _profileHeader(context, l10n),
          _vipCard(context, l10n),
          _sectionTitle(context, '我的书架'),
          if (shelf.isEmpty)
            ListTile(title: Text(l10n.empty))
          else
            ...shelf.map((s) => _shelfTile(context, s)),
          _sectionTitle(context, '我的收藏'),
          if (favorites.isEmpty)
            ListTile(title: Text(l10n.empty))
          else
            ...favorites.map((f) => ListTile(
                  leading: const Icon(Icons.menu_book),
                  title: Text(_titles[f.bookId] ?? ''),
                  onTap: () => openBook(context,
                      id: f.bookId, title: _titles[f.bookId] ?? ''),
                )),
        ],
      ),
    );
  }

  Widget _profileHeader(BuildContext context, AppLocalizations l10n) {
    final user = ApiClient.instance.currentUser;
    return ListTile(
      leading: CircleAvatar(
        child: Icon(Icons.person,
            size: 32, color: Theme.of(context).colorScheme.primary),
      ),
      title: Text(l10n.welcome(
          user!.nickname.isNotEmpty ? user.nickname : user.username)),
      subtitle: Text(user.username),
      trailing: OutlinedButton(
        onPressed: () {
          ApiClient.instance.logout();
          _load();
        },
        child: Text(l10n.logout),
      ),
    );
  }

  /// VIP 状态卡（T-P-17）：active/到期时间 + 开通/续费入口。
  Widget _vipCard(BuildContext context, AppLocalizations l10n) {
    final vip = _vip;
    final active = vip?.active == true;
    return Card(
      child: ListTile(
        leading: Icon(Icons.workspace_premium,
            color: active ? Colors.amber : Theme.of(context).disabledColor),
        title: Text(active ? l10n.vipActive : l10n.vipNotActive),
        subtitle: vip != null && vip.vipExpiresAt.isNotEmpty
            ? Text(l10n.vipExpiresAt(_shortDate(vip.vipExpiresAt)))
            : null,
        trailing: FilledButton.tonal(
          onPressed: () async {
            await Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => const VipPage()),
            );
            if (mounted) _load();
          },
          child: Text(active ? l10n.vipRenew : l10n.vipOpen),
        ),
      ),
    );
  }

  Widget _sectionTitle(BuildContext context, String title) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
      child: Text(title,
          style: Theme.of(context)
              .textTheme
              .titleMedium
              ?.copyWith(fontWeight: FontWeight.bold)),
    );
  }

  Widget _shelfTile(BuildContext context, ShelfItem s) {
    final api = ApiClient.instance;
    final chapter = _progressChapter[s.bookId];
    final chapters = _chapters[s.bookId];
    final title = _titles[s.bookId] ?? '';
    final progressText = chapter != null
        ? '第 ${chapter.chapterNo} 章 · ${_shortDate(_progressAt[s.bookId] ?? '')}'
        : '未开始阅读';
    return ListTile(
      leading: const Icon(Icons.bookmark),
      title: Text(title),
      subtitle: Text(progressText),
      onTap: () => openBook(context, id: s.bookId, title: title),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            icon: const Icon(Icons.play_circle_outline),
            tooltip: '阅读',
            onPressed: chapters == null || chapters.isEmpty
                ? null
                : () => Navigator.of(context).push(MaterialPageRoute(
                      builder: (_) => ReaderPage(
                        chapter: chapter ?? chapters.first,
                        chapters: chapters,
                        bookId: s.bookId,
                      ),
                    )),
          ),
          IconButton(
            icon: const Icon(Icons.delete_outline),
            tooltip: '移出书架',
            onPressed: () async {
              try {
                await api.removeBookshelf(s.bookId);
                if (!mounted) return;
                setState(() {
                  _shelf?.removeWhere((x) => x.bookId == s.bookId);
                });
              } catch (e) {
                if (!context.mounted) return;
                ScaffoldMessenger.of(context)
                  ..hideCurrentSnackBar()
                  ..showSnackBar(SnackBar(content: Text(api.errorMessage(e))));
              }
            },
          ),
        ],
      ),
    );
  }

  /// 截取日期部分（后端时间串含时分秒时只留 yyyy-MM-dd）。
  String _shortDate(String s) => s.length >= 10 ? s.substring(0, 10) : s;
}
