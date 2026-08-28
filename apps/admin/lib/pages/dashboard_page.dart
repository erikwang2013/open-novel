import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 仪表盘：统计卡片 + 热门书籍 / 热门搜索词榜单（T-A-15）。
class DashboardPage extends StatefulWidget {
  const DashboardPage({super.key});

  @override
  State<DashboardPage> createState() => _DashboardPageState();
}

class _DashboardPageState extends State<DashboardPage> {
  StatsData? _stats;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final s = await ApiClient.instance.stats();
      if (!mounted) return;
      setState(() => _stats = s);
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = _stats;
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              const Spacer(),
              IconButton(
                  tooltip: '刷新',
                  onPressed: _loading ? null : _load,
                  icon: const Icon(Icons.refresh)),
            ],
          ),
        ),
        Expanded(
          child: _loading && s == null
              ? const Center(child: CircularProgressIndicator())
              : ListView(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  children: [
                    Row(
                      children: [
                        _StatCard('书籍', s?.bookCount ?? 0, Icons.menu_book),
                        _StatCard('用户', s?.userCount ?? 0, Icons.people),
                        _StatCard('评论', s?.commentCount ?? 0, Icons.comment),
                        _StatCard('DAU（近似）', s?.dau ?? 0, Icons.trending_up),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Text('热门书籍',
                        style: Theme.of(context).textTheme.titleMedium),
                    _RankTable(
                      header: const ['书名', '热度'],
                      rows: [
                        for (final b in s?.hotBooks ?? [])
                          [b.title, '${b.hot}'],
                      ],
                      empty: '暂无热门数据',
                    ),
                    const SizedBox(height: 16),
                    Text('热门搜索词',
                        style: Theme.of(context).textTheme.titleMedium),
                    _RankTable(
                      header: const ['关键词', '次数'],
                      rows: [
                        for (final w in s?.hotKeywords ?? [])
                          [w.keyword, '${w.count}'],
                      ],
                      empty: '暂无搜索记录',
                    ),
                    const SizedBox(height: 16),
                  ],
                ),
        ),
      ],
    );
  }
}

class _StatCard extends StatelessWidget {
  const _StatCard(this.title, this.value, this.icon);

  final String title;
  final int value;
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
              Text('$value',
                  style: Theme.of(context).textTheme.headlineSmall),
              Text(title, style: Theme.of(context).textTheme.bodySmall),
            ],
          ),
        ),
      ),
    );
  }
}

/// 榜单小表格。
class _RankTable extends StatelessWidget {
  const _RankTable(
      {required this.header, required this.rows, required this.empty});

  final List<String> header;
  final List<List<String>> rows;
  final String empty;

  @override
  Widget build(BuildContext context) {
    if (rows.isEmpty) {
      return Padding(
        padding: const EdgeInsets.all(16),
        child: Text(empty,
            style: TextStyle(color: Theme.of(context).colorScheme.outline)),
      );
    }
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: DataTable(
        columns: [for (final h in header) DataColumn(label: Text(h))],
        rows: [
          for (final r in rows)
            DataRow(cells: [for (final c in r) DataCell(Text(c))]),
        ],
      ),
    );
  }
}
