import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 举报中心页：待审核举报列表（status=2），通过（下架评论）/驳回。
class ReportsPage extends StatefulWidget {
  const ReportsPage({super.key});

  @override
  State<ReportsPage> createState() => _ReportsPageState();
}

class _ReportsPageState extends State<ReportsPage> {
  static const _pageSize = 20;

  int _page = 1;
  List<Comment> _list = [];
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
      final (list, total) =
          await ApiClient.instance.reports(page: _page, pageSize: _pageSize);
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

  Future<void> _handle(Comment c, bool approved) async {
    final ok = await confirmDialog(
        context,
        approved ? '通过举报' : '驳回举报',
        approved ? '通过后该评论将被下架，确定？' : '驳回举报，评论保持正常，确定？',
        confirmText: approved ? '通过' : '驳回');
    if (!ok) return;
    try {
      await ApiClient.instance.handleReport(c.id, approved);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(approved ? '已通过，评论已下架' : '已驳回')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
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
                          DataColumn(label: Text('内容')),
                          DataColumn(label: Text('用户')),
                          DataColumn(label: Text('书籍/章节')),
                          DataColumn(label: Text('被举报次数')),
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
                              DataCell(Text('${c.reportCount}')),
                              DataCell(Text(c.createdAt)),
                              DataCell(Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  TextButton(
                                      onPressed: () => _handle(c, true),
                                      child: const Text('通过')),
                                  TextButton(
                                      onPressed: () => _handle(c, false),
                                      child: const Text('驳回')),
                                ],
                              )),
                            ]),
                        ],
                      ),
                    ),
                    if (_list.isEmpty && !_loading)
                      const Padding(
                          padding: EdgeInsets.all(24),
                          child: Center(child: Text('暂无待审核举报'))),
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
}
