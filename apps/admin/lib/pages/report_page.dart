import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 报表：订单/收入、用户增长、VIP 订阅、内容/互动四类运营报表，按日期范围查询。
class ReportPage extends StatefulWidget {
  const ReportPage({super.key});

  @override
  State<ReportPage> createState() => _ReportPageState();
}

class _ReportPageState extends State<ReportPage> {
  late DateTime _start;
  late DateTime _end;
  ReportsData? _data;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    final now = DateTime.now();
    _end = now;
    _start = now.subtract(const Duration(days: 29)); // 默认近 30 天
    _load();
  }

  String _fmt(DateTime d) =>
      '${d.year}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';

  Future<void> _pickStart() async {
    final picked = await showDatePicker(
        context: context,
        initialDate: _start,
        firstDate: DateTime(2020),
        lastDate: _end);
    if (picked == null) return;
    setState(() => _start = picked);
  }

  Future<void> _pickEnd() async {
    final picked = await showDatePicker(
        context: context,
        initialDate: _end,
        firstDate: _start,
        lastDate: DateTime.now());
    if (picked == null) return;
    setState(() => _end = picked);
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final d = await ApiClient.instance
          .reportsData(startDate: _fmt(_start), endDate: _fmt(_end));
      if (!mounted) return;
      setState(() => _data = d);
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final d = _data;
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              OutlinedButton(
                  onPressed: _pickStart, child: Text(_fmt(_start))),
              const Text(' 至 '),
              OutlinedButton(onPressed: _pickEnd, child: Text(_fmt(_end))),
              const SizedBox(width: 8),
              FilledButton(onPressed: _loading ? null : _load, child: const Text('查询')),
              const Spacer(),
              IconButton(
                  tooltip: '刷新',
                  onPressed: _loading ? null : _load,
                  icon: const Icon(Icons.refresh)),
            ],
          ),
        ),
        Expanded(
          child: _loading && d == null
              ? const Center(child: CircularProgressIndicator())
              : DefaultTabController(
                  length: 4,
                  child: Column(
                    children: [
                      const TabBar(tabs: [
                        Tab(text: '订单/收入'),
                        Tab(text: '用户增长'),
                        Tab(text: 'VIP 订阅'),
                        Tab(text: '内容/互动'),
                      ]),
                      Expanded(
                        child: TabBarView(
                          children: [
                            _OrderTab(r: d?.orderReport),
                            _UserTab(r: d?.userReport),
                            _VipTab(r: d?.vipReport),
                            _ContentTab(r: d?.contentReport),
                          ],
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

class _OrderTab extends StatelessWidget {
  const _OrderTab({this.r});

  final OrderReport? r;

  @override
  Widget build(BuildContext context) {
    final d = r;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Row(children: [
          _SumCard('订单数', '${d?.totalCount ?? 0}', Icons.receipt_long),
          _SumCard('总金额', fenToYuan(d?.totalAmount ?? 0), Icons.payments),
        ]),
        const SizedBox(height: 16),
        _Section('按日统计', header: const ['日期', '单数', '金额'], rows: [
          for (final x in d?.byDate ?? const <DateAmount>[])
            [x.date, '${x.count}', fenToYuan(x.amount)],
        ]),
        _Section('按渠道统计', header: const ['渠道', '单数', '金额'], rows: [
          for (final x in d?.byChannel ?? const <ChannelAmount>[])
            [x.channel, '${x.count}', fenToYuan(x.amount)],
        ]),
      ],
    );
  }
}

class _UserTab extends StatelessWidget {
  const _UserTab({this.r});

  final UserReport? r;

  @override
  Widget build(BuildContext context) {
    final d = r;
    final days = [...?d?.byDate]
      ..sort((a, b) => a.date.compareTo(b.date));
    var cum = 0;
    final rows = <List<String>>[];
    for (final x in days) {
      cum += x.count;
      rows.add([x.date, '${x.count}', '$cum']);
    }
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Row(children: [
          _SumCard('累计用户', '${d?.totalUsers ?? 0}', Icons.people),
        ]),
        const SizedBox(height: 16),
        _Section('按日统计', header: const ['日期', '新增', '累计'], rows: rows),
      ],
    );
  }
}

class _VipTab extends StatelessWidget {
  const _VipTab({this.r});

  final VipReport? r;

  @override
  Widget build(BuildContext context) {
    final d = r;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Row(children: [
          _SumCard('订阅数', '${d?.totalCount ?? 0}', Icons.workspace_premium),
          _SumCard('收入', fenToYuan(d?.totalAmount ?? 0), Icons.payments),
        ]),
        const SizedBox(height: 16),
        _Section('按日统计', header: const ['日期', '订阅数', '金额'], rows: [
          for (final x in d?.byDate ?? const <DateAmount>[])
            [x.date, '${x.count}', fenToYuan(x.amount)],
        ]),
        _Section('按套餐统计', header: const ['套餐', '订阅数', '金额'], rows: [
          for (final x in d?.byPlan ?? const <PlanStat>[])
            [x.planName.isEmpty ? '套餐#${x.planId}' : x.planName, '${x.count}', fenToYuan(x.amount)],
        ]),
      ],
    );
  }
}

class _ContentTab extends StatelessWidget {
  const _ContentTab({this.r});

  final ContentReport? r;

  @override
  Widget build(BuildContext context) {
    final d = r;
    final books = {for (final x in d?.booksByDate ?? const <DateCount>[]) x.date: x.count};
    final chapters = {for (final x in d?.chaptersByDate ?? const <DateCount>[]) x.date: x.count};
    final dates = {...books.keys, ...chapters.keys}.toList()..sort();
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Row(children: [
          _SumCard('评论数', '${d?.commentCount ?? 0}', Icons.comment),
          _SumCard('举报数', '${d?.reportCount ?? 0}', Icons.report),
        ]),
        const SizedBox(height: 16),
        _Section('新增内容按日统计', header: const ['日期', '新增书籍', '新增章节'], rows: [
          for (final date in dates)
            [date, '${books[date] ?? 0}', '${chapters[date] ?? 0}'],
        ]),
      ],
    );
  }
}

/// 汇总小卡片。
class _SumCard extends StatelessWidget {
  const _SumCard(this.title, this.value, this.icon);

  final String title;
  final String value;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Card(
        margin: const EdgeInsets.symmetric(horizontal: 4),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              Icon(icon, color: Theme.of(context).colorScheme.primary),
              const SizedBox(height: 8),
              Text(value, style: Theme.of(context).textTheme.headlineSmall),
              Text(title, style: Theme.of(context).textTheme.bodySmall),
            ],
          ),
        ),
      ),
    );
  }
}

/// 小节标题 + 表格（空数据给占位文案）。
class _Section extends StatelessWidget {
  const _Section(this.title, {required this.header, required this.rows});

  final String title;
  final List<String> header;
  final List<List<String>> rows;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title, style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        if (rows.isEmpty)
          Padding(
            padding: const EdgeInsets.all(8),
            child: Text('暂无数据',
                style: TextStyle(color: Theme.of(context).colorScheme.outline)),
          )
        else
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: DataTable(
              columns: [for (final h in header) DataColumn(label: Text(h))],
              rows: [
                for (final row in rows)
                  DataRow(cells: [for (final c in row) DataCell(Text(c))]),
              ],
            ),
          ),
        const SizedBox(height: 16),
      ],
    );
  }
}
