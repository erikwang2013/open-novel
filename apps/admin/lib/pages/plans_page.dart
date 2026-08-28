import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// VIP 套餐配置页（T-P-13）：套餐 CRUD（时长 / 价格 / 币种 / 标签）。
/// 金额输入元，提交转分；删除 = 禁用（软删，历史订单仍引用 plan_code）。
class PlansPage extends StatefulWidget {
  const PlansPage({super.key});

  @override
  State<PlansPage> createState() => _PlansPageState();
}

class _PlansPageState extends State<PlansPage> {
  List<VipPlan> _items = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (items, _) = await ApiClient.instance.plans();
      if (!mounted) return;
      setState(() => _items = items);
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _create() async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => const _PlanDialog());
    if (ok == true) _load();
  }

  Future<void> _edit(VipPlan p) async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => _PlanDialog(plan: p));
    if (ok == true) _load();
  }

  Future<void> _toggle(VipPlan p) async {
    try {
      await ApiClient.instance.updatePlan(
          p.id, {'status': p.status == 1 ? 0 : 1});
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(p.status == 1 ? '已禁用' : '已启用')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  Future<void> _delete(VipPlan p) async {
    final ok = await confirmDialog(
        context, '删除套餐', '确定删除套餐「${p.planCode}」？删除后该套餐不再参与定价。',
        confirmText: '删除');
    if (!ok) return;
    try {
      await ApiClient.instance.deletePlan(p.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('已删除')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  static String _yuan(int cents) => (cents / 100).toStringAsFixed(2);

  static const _codes = {'monthly': '月卡', 'quarterly': '季卡', 'yearly': '年卡'};

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              Text('共 ${_items.length} 个套餐'),
              const Spacer(),
              FilledButton.icon(
                  onPressed: _create,
                  icon: const Icon(Icons.add),
                  label: const Text('新建套餐')),
            ],
          ),
        ),
        Expanded(
          child: _loading && _items.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : _items.isEmpty
                  ? const Center(child: Text('暂无套餐（支付走内置默认价）'))
                  : SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      child: DataTable(
                        columns: const [
                          DataColumn(label: Text('代码')),
                          DataColumn(label: Text('时长')),
                          DataColumn(label: Text('价格')),
                          DataColumn(label: Text('币种')),
                          DataColumn(label: Text('标签')),
                          DataColumn(label: Text('排序')),
                          DataColumn(label: Text('状态')),
                          DataColumn(label: Text('操作')),
                        ],
                        rows: [
                          for (final p in _items)
                            DataRow(cells: [
                              DataCell(Text(
                                  '${p.planCode}${_codes.containsKey(p.planCode) ? '（${_codes[p.planCode]}）' : ''}')),
                              DataCell(Text('${p.days} 天')),
                              DataCell(Text(_yuan(p.amount))),
                              DataCell(Text(p.currency)),
                              DataCell(Text(p.label.isEmpty ? '—' : p.label)),
                              DataCell(Text('${p.sort}')),
                              DataCell(Text(p.status == 1 ? '启用' : '禁用')),
                              DataCell(Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  TextButton(
                                      onPressed: () => _edit(p),
                                      child: const Text('编辑')),
                                  TextButton(
                                      onPressed: () => _toggle(p),
                                      child: Text(p.status == 1 ? '禁用' : '启用')),
                                  TextButton(
                                      onPressed: () => _delete(p),
                                      child: const Text('删除')),
                                ],
                              )),
                            ]),
                        ],
                      ),
                    ),
        ),
      ],
    );
  }
}

/// 新建/编辑套餐弹窗；编辑态仅提交变更字段。
class _PlanDialog extends StatefulWidget {
  const _PlanDialog({this.plan});

  final VipPlan? plan;

  @override
  State<_PlanDialog> createState() => _PlanDialogState();
}

class _PlanDialogState extends State<_PlanDialog> {
  late final _code =
      TextEditingController(text: widget.plan?.planCode ?? 'monthly');
  late final _days = TextEditingController(text: '${widget.plan?.days ?? 30}');
  late final _price = TextEditingController(
      text: widget.plan == null
          ? '3.00'
          : (widget.plan!.amount / 100).toStringAsFixed(2));
  late final _currency =
      TextEditingController(text: widget.plan?.currency ?? 'USD');
  late final _label = TextEditingController(text: widget.plan?.label ?? '');
  late final _sort = TextEditingController(text: '${widget.plan?.sort ?? 0}');
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _code.dispose();
    _days.dispose();
    _price.dispose();
    _currency.dispose();
    _label.dispose();
    _sort.dispose();
    super.dispose();
  }

  int? _parseCents(String s) {
    final v = double.tryParse(s.trim());
    if (v == null || v <= 0) return null;
    return (v * 100).round();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final p = widget.plan;
      if (p == null) {
        final cents = _parseCents(_price.text);
        final days = int.tryParse(_days.text.trim());
        if (cents == null || days == null || days <= 0) {
          throw StateError('时长或价格无效');
        }
        await ApiClient.instance.createPlan(
          planCode: _code.text.trim(),
          days: days,
          amount: cents,
          currency: _currency.text.trim().isEmpty
              ? 'USD'
              : _currency.text.trim().toUpperCase(),
          label: _label.text.trim(),
          sort: int.tryParse(_sort.text.trim()) ?? 0,
        );
      } else {
        final cents = _parseCents(_price.text);
        final days = int.tryParse(_days.text.trim());
        if (cents == null || days == null || days <= 0) {
          throw StateError('时长或价格无效');
        }
        final patch = <String, dynamic>{
          if (days != p.days) 'days': days,
          if (cents != p.amount) 'amount': cents,
          if (_currency.text.trim().toUpperCase() != p.currency)
            'currency': _currency.text.trim().toUpperCase(),
          if (_label.text.trim() != p.label) 'label': _label.text.trim(),
          if ((int.tryParse(_sort.text.trim()) ?? 0) != p.sort)
            'sort': int.tryParse(_sort.text.trim()) ?? 0,
        };
        if (patch.isEmpty) {
          if (mounted) Navigator.pop(context, true);
          return;
        }
        await ApiClient.instance.updatePlan(p.id, patch);
      }
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = ApiClient.instance.errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(widget.plan == null ? '新建套餐' : '编辑套餐'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            DropdownButtonFormField<String>(
              initialValue: _code.text,
              decoration: const InputDecoration(labelText: '套餐代码 *'),
              items: const [
                DropdownMenuItem(value: 'monthly', child: Text('monthly（月卡）')),
                DropdownMenuItem(value: 'quarterly', child: Text('quarterly（季卡）')),
                DropdownMenuItem(value: 'yearly', child: Text('yearly（年卡）')),
              ],
              onChanged: widget.plan == null
                  ? (v) => setState(() => _code.text = v ?? 'monthly')
                  : null,
            ),
            TextField(
                controller: _days,
                decoration: const InputDecoration(labelText: '有效天数 *')),
            TextField(
                controller: _price,
                decoration: const InputDecoration(labelText: '价格（元）*')),
            TextField(
                controller: _currency,
                decoration: const InputDecoration(labelText: '币种（USD）')),
            TextField(
                controller: _label,
                decoration: const InputDecoration(labelText: '标签')),
            TextField(
                controller: _sort,
                decoration: const InputDecoration(labelText: '排序（升序）')),
            if (_error != null)
              Text(_error!,
                  style: TextStyle(
                      color: Theme.of(context).colorScheme.error)),
          ],
        ),
      ),
      actions: [
        TextButton(
            onPressed: _busy ? null : () => Navigator.pop(context, false),
            child: const Text('取消')),
        FilledButton(
            onPressed: _busy ? null : _submit,
            child: _busy
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('保存')),
      ],
    );
  }
}
