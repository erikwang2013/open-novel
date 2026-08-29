import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 审计日志页：分页 / 筛选日志列表（只读）。
class AuditLogsPage extends StatefulWidget {
  const AuditLogsPage({super.key});

  @override
  State<AuditLogsPage> createState() => _AuditLogsPageState();
}

class _AuditLogsPageState extends State<AuditLogsPage> {
  static const _pageSize = 20;

  final _userId = TextEditingController();
  final _action = TextEditingController();
  final _targetType = TextEditingController();
  final _targetId = TextEditingController();
  final _start = TextEditingController();
  final _end = TextEditingController();

  int _page = 1;
  int _total = 0;
  List<AuditLog> _items = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _userId.dispose();
    _action.dispose();
    _targetType.dispose();
    _targetId.dispose();
    _start.dispose();
    _end.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (items, total) = await ApiClient.instance.auditLogs(
        userId: _userId.text.trim(),
        action: _action.text.trim(),
        targetType: _targetType.text.trim(),
        targetId: _targetId.text.trim(),
        startTime: _start.text.trim(),
        endTime: _end.text.trim(),
        page: _page,
        pageSize: _pageSize,
      );
      if (!mounted) return;
      setState(() {
        _items = items;
        _total = total;
      });
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _search() {
    setState(() => _page = 1);
    _load();
  }

  static String _dash(String v) => v.isEmpty ? '—' : v;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        _buildFilters(),
        const Divider(height: 1),
        Expanded(
          child: _loading && _items.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : Column(
                  children: [
                    Expanded(
                      child: _items.isEmpty
                          ? const Center(child: Text('暂无审计日志'))
                          : SingleChildScrollView(
                              scrollDirection: Axis.horizontal,
                              child: DataTable(
                                columns: const [
                                  DataColumn(label: Text('ID')),
                                  DataColumn(label: Text('用户 ID')),
                                  DataColumn(label: Text('动作')),
                                  DataColumn(label: Text('目标类型')),
                                  DataColumn(label: Text('目标 ID')),
                                  DataColumn(label: Text('详情')),
                                  DataColumn(label: Text('IP')),
                                  DataColumn(label: Text('UA')),
                                  DataColumn(label: Text('时间')),
                                ],
                                rows: [
                                  for (final l in _items)
                                    DataRow(cells: [
                                      DataCell(Text(l.id)),
                                      DataCell(Text(l.userId)),
                                      DataCell(Text(l.action)),
                                      DataCell(Text(_dash(l.targetType))),
                                      DataCell(Text(_dash(l.targetId))),
                                      DataCell(ConstrainedBox(
                                        constraints:
                                            const BoxConstraints(maxWidth: 320),
                                        child: Text(l.detail,
                                            maxLines: 2,
                                            overflow: TextOverflow.ellipsis),
                                      )),
                                      DataCell(Text(_dash(l.ip))),
                                      DataCell(ConstrainedBox(
                                        constraints:
                                            const BoxConstraints(maxWidth: 200),
                                        child: Text(_dash(l.userAgent),
                                            maxLines: 1,
                                            overflow: TextOverflow.ellipsis),
                                      )),
                                      DataCell(Text(l.createdAt)),
                                    ]),
                                ],
                              ),
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
        ),
      ],
    );
  }

  Widget _buildFilters() {
    return Padding(
      padding: const EdgeInsets.all(8),
      child: Wrap(
        spacing: 8,
        runSpacing: 8,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: [
          SizedBox(
              width: 120,
              child: TextField(
                  controller: _userId,
                  decoration:
                      const InputDecoration(labelText: '用户 ID', isDense: true))),
          SizedBox(
              width: 120,
              child: TextField(
                  controller: _action,
                  decoration: const InputDecoration(
                      labelText: '动作', isDense: true, hintText: 'login'))),
          SizedBox(
              width: 120,
              child: TextField(
                  controller: _targetType,
                  decoration: const InputDecoration(
                      labelText: '目标类型', isDense: true, hintText: 'book'))),
          SizedBox(
              width: 120,
              child: TextField(
                  controller: _targetId,
                  decoration: const InputDecoration(
                      labelText: '目标 ID', isDense: true))),
          SizedBox(
              width: 140,
              child: TextField(
                  controller: _start,
                  decoration: const InputDecoration(
                      labelText: '开始时间', isDense: true,
                      hintText: '2026-08-01'))),
          SizedBox(
              width: 140,
              child: TextField(
                  controller: _end,
                  decoration: const InputDecoration(
                      labelText: '结束时间', isDense: true,
                      hintText: '2026-08-29'))),
          FilledButton.icon(
              onPressed: _search,
              icon: const Icon(Icons.search),
              label: const Text('查询')),
        ],
      ),
    );
  }
}
