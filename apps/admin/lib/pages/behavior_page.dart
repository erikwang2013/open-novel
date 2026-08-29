import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 行为分析：当日/7 天活跃阅读用户、热门阅读书籍 TOP10、24 小时活跃分布。
class BehaviorPage extends StatefulWidget {
  const BehaviorPage({super.key});

  @override
  State<BehaviorPage> createState() => _BehaviorPageState();
}

class _BehaviorPageState extends State<BehaviorPage> {
  BehaviorStats? _stats;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final s = await ApiClient.instance.behaviorStats();
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
                        _StatCard('当日阅读用户', s?.activeReaders ?? 0, Icons.person),
                        _StatCard('近 7 天阅读用户', s?.readers7d ?? 0, Icons.groups),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Text('热门阅读书籍 TOP10',
                        style: Theme.of(context).textTheme.titleMedium),
                    _RankTable(
                      header: const ['书名', '阅读次数'],
                      rows: [
                        for (final b in s?.hotBooks ?? [])
                          [b.title, '${b.count}'],
                      ],
                      empty: '暂无阅读数据',
                    ),
                    const SizedBox(height: 16),
                    Text('活跃时段（当日，0-23 时）',
                        style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 8),
                    _HourlyBars(hourly: s?.hourly ?? const []),
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

/// 24 小时分布：按最大值等比宽度的横向条形（无图表库，基础组件拼接）。
class _HourlyBars extends StatelessWidget {
  const _HourlyBars({required this.hourly});

  final List<int> hourly;

  @override
  Widget build(BuildContext context) {
    final list = hourly.length == 24 ? hourly : List<int>.filled(24, 0);
    final max = list.reduce((a, b) => a > b ? a : b);
    if (max == 0) {
      return Text('当日暂无阅读事件',
          style: TextStyle(color: Theme.of(context).colorScheme.outline));
    }
    return Column(
      children: [
        for (var h = 0; h < 24; h++)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 1),
            child: Row(
              children: [
                SizedBox(
                    width: 42,
                    child: Text('${h.toString().padLeft(2, '0')} 时',
                        style: Theme.of(context).textTheme.bodySmall)),
                Expanded(
                  child: FractionallySizedBox(
                    alignment: Alignment.centerLeft,
                    widthFactor: list[h] / max,
                    child: Container(
                      height: 14,
                      decoration: BoxDecoration(
                        color: Theme.of(context).colorScheme.primary
                            .withValues(alpha: 0.75),
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  ),
                ),
                SizedBox(
                    width: 56,
                    child: Text('${list[h]}',
                        textAlign: TextAlign.right,
                        style: Theme.of(context).textTheme.bodySmall)),
              ],
            ),
          ),
      ],
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
