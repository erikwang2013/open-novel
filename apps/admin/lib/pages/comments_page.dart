import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 评论审核页：按 book_id/chapter_id/status 筛选、分页，下架/恢复。
class CommentsPage extends StatefulWidget {
  const CommentsPage({super.key});

  @override
  State<CommentsPage> createState() => _CommentsPageState();
}

class _CommentsPageState extends State<CommentsPage> {
  static const _pageSize = 20;

  final _bookId = TextEditingController();
  final _chapterId = TextEditingController();
  int? _status; // null=全部（后端默认正常） 2=举报待审
  int _page = 1;
  List<Comment> _list = [];
  int _total = 0;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _bookId.dispose();
    _chapterId.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (list, total) = await ApiClient.instance.comments(
          bookId: _bookId.text.trim(),
          chapterId: _chapterId.text.trim(),
          status: _status,
          page: _page,
          pageSize: _pageSize);
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

  Future<void> _toggleStatus(Comment c) async {
    final on = c.status == 1;
    final ok = await confirmDialog(
        context,
        on ? '下架评论' : '恢复评论',
        on ? '确定下架这条评论？' : '确定恢复这条评论？',
        confirmText: on ? '下架' : '恢复');
    if (!ok) return;
    try {
      await ApiClient.instance.updateCommentStatus(c.id, on ? 0 : 1);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(on ? '已下架' : '已恢复')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  void _applyFilter() {
    setState(() => _page = 1);
    _load();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Wrap(
            spacing: 12,
            runSpacing: 8,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              SizedBox(
                  width: 120,
                  child: TextField(
                      controller: _bookId,
                      decoration: const InputDecoration(
                          labelText: 'book_id', isDense: true))),
              SizedBox(
                  width: 120,
                  child: TextField(
                      controller: _chapterId,
                      decoration: const InputDecoration(
                          labelText: 'chapter_id', isDense: true))),
              DropdownButton<int?>(
                value: _status,
                hint: const Text('状态'),
                items: const [
                  // 后端 biz 对 status<=0 强制为 1，故「全部」实为正常评论
                  DropdownMenuItem<int?>(value: null, child: Text('正常')),
                  DropdownMenuItem<int?>(value: 2, child: Text('待审核')),
                ],
                onChanged: (v) => setState(() => _status = v),
              ),
              FilledButton(onPressed: _applyFilter, child: const Text('查询')),
            ],
          ),
        ),
        Expanded(
          child: _loading && _list.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : ListView(
                  children: [
                    SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      child: DataTable(
                        columns: const [
                          DataColumn(label: Text('内容')),
                          DataColumn(label: Text('用户')),
                          DataColumn(label: Text('书籍/章节')),
                          DataColumn(label: Text('点赞/举报')),
                          DataColumn(label: Text('状态')),
                          DataColumn(label: Text('时间')),
                          DataColumn(label: Text('操作')),
                        ],
                        rows: [
                          for (final c in _list)
                            DataRow(cells: [
                              DataCell(ConstrainedBox(
                                  constraints: const BoxConstraints(
                                      maxWidth: 320),
                                  child: Text(c.content,
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis))),
                              DataCell(Text(c.userId)),
                              DataCell(Text(
                                  '${c.bookId} / ${c.chapterId == '0' || c.chapterId.isEmpty ? '-' : c.chapterId}')),
                              DataCell(Text('${c.likeCount} / ${c.reportCount}')),
                              DataCell(Text(_statusText(c.status))),
                              DataCell(Text(c.createdAt)),
                              DataCell(Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  if (c.status != 1)
                                    TextButton(
                                        onPressed: () => _toggleStatus(c),
                                        child: const Text('恢复')),
                                  if (c.status == 1)
                                    TextButton(
                                        onPressed: () => _toggleStatus(c),
                                        child: const Text('下架')),
                                ],
                              )),
                            ]),
                        ],
                      ),
                    ),
                    if (_list.isEmpty && !_loading)
                      const Padding(
                          padding: EdgeInsets.all(24),
                          child: Center(child: Text('暂无评论'))),
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
    );
  }

  static String _statusText(int s) =>
      s == 2 ? '待审核' : (s == 1 ? '正常' : '下架');
}
