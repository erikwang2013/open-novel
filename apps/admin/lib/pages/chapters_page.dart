import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 章节管理页：从书籍列表钻取进入，分页列表 + 正文查看 + 禁用/恢复。
class ChaptersPage extends StatefulWidget {
  const ChaptersPage({super.key, required this.book});

  final Book book;

  @override
  State<ChaptersPage> createState() => _ChaptersPageState();
}

class _ChaptersPageState extends State<ChaptersPage> {
  static const _pageSize = 20;

  int _page = 1;
  List<Chapter> _list = [];
  int _total = 0;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (list, total) = await ApiClient.instance
          .chapters(widget.book.id, page: _page, pageSize: _pageSize);
      if (!mounted) return;
      setState(() {
        _list = list;
        _total = total;
      });
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _toggleStatus(Chapter c) async {
    final on = c.status == 1;
    final ok = await confirmDialog(
        context,
        on ? '禁用章节' : '恢复章节',
        on ? '确定禁用「${c.title}」？禁用后正文不可访问。' : '确定恢复「${c.title}」？',
        confirmText: on ? '禁用' : '恢复');
    if (!ok) return;
    try {
      await ApiClient.instance.updateChapterStatus(c.id, on ? 0 : 1);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(on ? '已禁用' : '已恢复')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  void _showContent(Chapter c) {
    showDialog(context: context, builder: (_) => _ContentDialog(chapter: c));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('章节管理 · ${widget.book.title}')),
      body: Column(
        children: [
          Expanded(
            child: _loading && _list.isEmpty
                ? const Center(child: CircularProgressIndicator())
                : ListView(
                    children: [
                      SingleChildScrollView(
                        scrollDirection: Axis.horizontal,
                        child: DataTable(
                          columns: const [
                            DataColumn(label: Text('章节号')),
                            DataColumn(label: Text('标题')),
                            DataColumn(label: Text('字数')),
                            DataColumn(label: Text('VIP')),
                            DataColumn(label: Text('状态')),
                            DataColumn(label: Text('创建时间')),
                            DataColumn(label: Text('操作')),
                          ],
                          rows: [
                            for (final c in _list)
                              DataRow(cells: [
                                DataCell(Text('${c.chapterNo}')),
                                DataCell(Text(c.title)),
                                DataCell(Text('${c.wordCount}')),
                                DataCell(Text(c.isVip == 1 ? '是' : '否')),
                                DataCell(Text(c.status == 1 ? '启用' : '禁用')),
                                DataCell(Text(c.createdAt)),
                                DataCell(Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    TextButton(
                                        onPressed: () => _showContent(c),
                                        child: const Text('正文')),
                                    TextButton(
                                        onPressed: () => _toggleStatus(c),
                                        child: Text(
                                            c.status == 1 ? '禁用' : '恢复')),
                                  ],
                                )),
                              ]),
                          ],
                        ),
                      ),
                      if (_list.isEmpty && !_loading)
                        const Padding(
                            padding: EdgeInsets.all(24),
                            child: Center(child: Text('暂无章节'))),
                    ],
                  ),
          ),
          PagingBar(
              page: _page,
              pageSize: _pageSize,
              total: _total,
              onChanged: (p) {
                setState(() => _page = p);
                _load();
              }),
        ],
      ),
    );
  }
}

/// 正文查看弹窗。
class _ContentDialog extends StatefulWidget {
  const _ContentDialog({required this.chapter});

  final Chapter chapter;

  @override
  State<_ContentDialog> createState() => _ContentDialogState();
}

class _ContentDialogState extends State<_ContentDialog> {
  String? _content;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final c = await ApiClient.instance.chapterContent(widget.chapter.id);
      if (mounted) setState(() => _content = c);
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = ApiClient.instance.errorMessage(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('正文 · ${widget.chapter.title}'),
      content: SizedBox(
        width: 520,
        height: 400,
        child: _content != null
            ? SingleChildScrollView(child: Text(_content!))
            : Center(
                child: _error != null
                    ? Text('加载失败：$_error')
                    : const CircularProgressIndicator()),
      ),
      actions: [
        TextButton(
            onPressed: () => Navigator.pop(context), child: const Text('关闭')),
      ],
    );
  }
}
