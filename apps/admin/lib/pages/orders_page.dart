import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 流水账单页（T-P-12）：汇总卡片 + 分页 / 筛选 / 详情。
/// 金额单位分，展示层转元。
class OrdersPage extends StatefulWidget {
  const OrdersPage({super.key});

  @override
  State<OrdersPage> createState() => _OrdersPageState();
}

class _OrdersPageState extends State<OrdersPage> {
  final _userId = TextEditingController();
  final _provider = TextEditingController();
  final _start = TextEditingController();
  final _end = TextEditingController();

  int _status = -1;
  int _page = 1;
  int _total = 0;
  List<PaymentOrder> _items = [];
  OrderStats _stats = OrderStats.fromJson({});
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _userId.dispose();
    _provider.dispose();
    _start.dispose();
    _end.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (items, total) = await ApiClient.instance.orders(
        userId: _userId.text.trim(),
        provider: _provider.text.trim(),
        status: _status,
        startAt: _start.text.trim(),
        endAt: _end.text.trim(),
        page: _page,
      );
      final stats = await ApiClient.instance.orderStats(
        startAt: _start.text.trim(),
        endAt: _end.text.trim(),
      );
      if (!mounted) return;
      setState(() {
        _items = items;
        _total = total;
        _stats = stats;
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

  Future<void> _detail(PaymentOrder o) async {
    await showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('订单详情'),
        content: SizedBox(
          width: 360,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _kv('订单号', o.orderNo),
              _kv('用户 ID', o.userId),
              _kv('金额', '${_yuan(o.amount)} ${o.currency}'),
              _kv('支付方式', o.provider),
              _kv('状态', _statusText(o.status)),
              _kv('交易号', o.txId.isEmpty ? '—' : o.txId),
              _kv('支付时间', o.paidAt.isEmpty ? '—' : o.paidAt),
              _kv('创建时间', o.createdAt),
            ],
          ),
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx), child: const Text('关闭')),
        ],
      ),
    );
  }

  static Widget _kv(String k, String v) => Text('$k：$v');

  static String _yuan(int cents) => (cents / 100).toStringAsFixed(2);

  static String _statusText(int s) => switch (s) {
        0 => '待支付',
        1 => '已付',
        2 => '失败',
        3 => '已关闭',
        _ => '未知',
      };

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        _buildFilters(),
        _buildStats(),
        const Divider(height: 1),
        Expanded(
          child: _loading && _items.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : Column(
                  children: [
                    Expanded(
                      child: _items.isEmpty
                          ? const Center(child: Text('暂无流水'))
                          : SingleChildScrollView(
                              scrollDirection: Axis.horizontal,
                              child: DataTable(
                                columns: const [
                                  DataColumn(label: Text('订单号')),
                                  DataColumn(label: Text('用户')),
                                  DataColumn(label: Text('金额')),
                                  DataColumn(label: Text('方式')),
                                  DataColumn(label: Text('状态')),
                                  DataColumn(label: Text('支付时间')),
                                  DataColumn(label: Text('操作')),
                                ],
                                rows: [
                                  for (final o in _items)
                                    DataRow(cells: [
                                      DataCell(Text(o.orderNo)),
                                      DataCell(Text(o.userId)),
                                      DataCell(Text(
                                          '${_yuan(o.amount)} ${o.currency}')),
                                      DataCell(Text(o.provider)),
                                      DataCell(Text(_statusText(o.status))),
                                      DataCell(Text(
                                          o.paidAt.isEmpty ? '—' : o.paidAt)),
                                      DataCell(TextButton(
                                          onPressed: () => _detail(o),
                                          child: const Text('详情'))),
                                    ]),
                                ],
                              ),
                            ),
                    ),
                    PagingBar(
                        page: _page,
                        pageSize: 20,
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
              width: 140,
              child: TextField(
                  controller: _provider,
                  decoration: const InputDecoration(
                      labelText: '支付方式', isDense: true))),
          SizedBox(
            width: 130,
            child: DropdownButtonFormField<int>(
              initialValue: _status,
              decoration: const InputDecoration(labelText: '状态', isDense: true),
              items: const [
                DropdownMenuItem(value: -1, child: Text('全部')),
                DropdownMenuItem(value: 0, child: Text('待支付')),
                DropdownMenuItem(value: 1, child: Text('已付')),
                DropdownMenuItem(value: 2, child: Text('失败')),
                DropdownMenuItem(value: 3, child: Text('已关闭')),
              ],
              onChanged: (v) => setState(() => _status = v ?? -1),
            ),
          ),
          SizedBox(
              width: 140,
              child: TextField(
                  controller: _start,
                  decoration: const InputDecoration(
                      labelText: '开始日期', isDense: true, hintText: '2026-08-01'))),
          SizedBox(
              width: 140,
              child: TextField(
                  controller: _end,
                  decoration: const InputDecoration(
                      labelText: '结束日期', isDense: true, hintText: '2026-08-29'))),
          FilledButton.icon(
              onPressed: _search,
              icon: const Icon(Icons.search),
              label: const Text('查询')),
        ],
      ),
    );
  }

  Widget _buildStats() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(8, 0, 8, 8),
      child: Row(
        children: [
          _card('总单数', _stats.total.toString(), Colors.blue),
          _card('已付', _stats.paid.toString(), Colors.green),
          _card('待支付', _stats.pending.toString(), Colors.orange),
          _card('已付金额', _yuan(_stats.amount), Colors.purple),
        ],
      ),
    );
  }

  Widget _card(String label, String value, MaterialColor color) {
    return Container(
      margin: const EdgeInsets.only(right: 12),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: color.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontSize: 12)),
          Text(value,
              style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: color.shade900)),
        ],
      ),
    );
  }
}
