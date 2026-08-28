import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';

/// 评论区：列表 + 发布 + 点赞（GET/POST /api/v1/comments，POST /api/v1/comments/{id}/like）。
class CommentsPage extends StatefulWidget {
  const CommentsPage({super.key, required this.bookId, this.chapterId});

  final String bookId;
  final String? chapterId;

  @override
  State<CommentsPage> createState() => _CommentsPageState();
}

class _CommentsPageState extends State<CommentsPage> {
  final _input = TextEditingController();
  List<Comment>? _comments;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _input.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _error = null;
      _comments = null;
    });
    try {
      final list = await ApiClient.instance.listComments(widget.bookId,
          chapterId: widget.chapterId);
      setState(() => _comments = list);
    } catch (e) {
      setState(() => _error = ApiClient.instance.errorMessage(e));
    }
  }

  Future<void> _post() async {
    final l10n = AppLocalizations.of(context)!;
    final text = _input.text.trim();
    if (text.isEmpty) return;
    if (!await ensureLogin(context)) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(l10n.loginRequired)));
      }
      return;
    }
    try {
      await ApiClient.instance.postComment(widget.bookId, text,
          chapterId: widget.chapterId);
      _input.clear();
      await _load();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(ApiClient.instance.errorMessage(e))));
      }
    }
  }

  Future<void> _toggleLike(Comment c) async {
    try {
      await ApiClient.instance.likeComment(c.id);
      await _load();
    } catch (_) {
      // 点赞失败忽略
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.comments)),
      body: _buildBody(context, l10n),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _input,
                  decoration: InputDecoration(
                    hintText: l10n.commentHint,
                    isDense: true,
                    border: const OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              FilledButton(onPressed: _post, child: Text(l10n.post)),
            ],
          ),
        ),
      ),
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
    final list = _comments;
    if (list == null) return const Center(child: CircularProgressIndicator());
    if (list.isEmpty) return Center(child: Text(l10n.emptyComment));
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.builder(
        padding: const EdgeInsets.all(8),
        itemCount: list.length,
        itemBuilder: (_, i) {
          final c = list[i];
          return ListTile(
            title: Text(c.content),
            subtitle: Text(c.createdAt,
                style: Theme.of(context).textTheme.bodySmall),
            trailing: IconButton(
              icon: Icon(
                c.likeCount > 0 ? Icons.thumb_up : Icons.thumb_up_outlined,
                size: 18,
              ),
              tooltip: '${l10n.like} ${c.likeCount}',
              onPressed: () => _toggleLike(c),
            ),
          );
        },
      ),
    );
  }
}
