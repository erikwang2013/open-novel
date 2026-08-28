import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 支付方式管理页（T-P-11）：列表 / 启停 / 排序 / 区域 / 密钥配置。
/// 密钥只显示「已配置/未配置」，绝不回显明文。
class ProvidersPage extends StatefulWidget {
  const ProvidersPage({super.key});

  @override
  State<ProvidersPage> createState() => _ProvidersPageState();
}

class _ProvidersPageState extends State<ProvidersPage> {
  List<PaymentProvider> _items = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (items, _) = await ApiClient.instance.providers();
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
        context: context, builder: (_) => const _ProviderDialog());
    if (ok == true) _load();
  }

  Future<void> _edit(PaymentProvider p) async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => _ProviderDialog(provider: p));
    if (ok == true) _load();
  }

  Future<void> _toggle(PaymentProvider p) async {
    try {
      await ApiClient.instance.toggleProvider(p.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(p.enabled == 1 ? '已禁用' : '已启用')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  Future<void> _delete(PaymentProvider p) async {
    final ok = await confirmDialog(
        context, '删除支付方式', '确定删除支付方式「${p.code}」？', confirmText: '删除');
    if (!ok) return;
    try {
      await ApiClient.instance.deleteProvider(p.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('已删除')));
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
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              Text('共 ${_items.length} 个支付方式'),
              const Spacer(),
              FilledButton.icon(
                  onPressed: _create,
                  icon: const Icon(Icons.add),
                  label: const Text('新建支付方式')),
            ],
          ),
        ),
        Expanded(
          child: _loading && _items.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : _items.isEmpty
                  ? const Center(child: Text('暂无支付方式'))
                  : SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      child: DataTable(
                        columns: const [
                          DataColumn(label: Text('代码')),
                          DataColumn(label: Text('语言')),
                          DataColumn(label: Text('地区')),
                          DataColumn(label: Text('排序')),
                          DataColumn(label: Text('密钥')),
                          DataColumn(label: Text('状态')),
                          DataColumn(label: Text('操作')),
                        ],
                        rows: [
                          for (final p in _items)
                            DataRow(cells: [
                              DataCell(Text(p.code)),
                              DataCell(Text(p.lang)),
                              DataCell(Text(p.region)),
                              DataCell(Text('${p.sort}')),
                              DataCell(Text(p.configConfigured ? '已配置' : '未配置')),
                              DataCell(_StatusTag(p.enabled == 1)),
                              DataCell(Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  TextButton(
                                      onPressed: () => _edit(p),
                                      child: const Text('编辑')),
                                  TextButton(
                                      onPressed: () => _toggle(p),
                                      child: Text(p.enabled == 1 ? '禁用' : '启用')),
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

class _StatusTag extends StatelessWidget {
  const _StatusTag(this.on);

  final bool on;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: on ? Colors.green.shade100 : Colors.red.shade100,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(on ? '启用' : '禁用',
          style: TextStyle(
              fontSize: 12,
              color: on ? Colors.green.shade900 : Colors.red.shade900)),
    );
  }
}

/// 新建/编辑支付方式弹窗。密钥字段（币种/币种链）加密存储，编辑时留空 = 保留原值。
class _ProviderDialog extends StatefulWidget {
  const _ProviderDialog({this.provider});

  final PaymentProvider? provider;

  @override
  State<_ProviderDialog> createState() => _ProviderDialogState();
}

class _ProviderDialogState extends State<_ProviderDialog> {
  late final _code = TextEditingController(text: widget.provider?.code ?? '');
  late final _lang = TextEditingController(text: widget.provider?.lang ?? '*');
  late final _region =
      TextEditingController(text: widget.provider?.region ?? '*');
  late final _sort =
      TextEditingController(text: '${widget.provider?.sort ?? 0}');
  late final _currency = TextEditingController();
  late final _coin = TextEditingController();
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _code.dispose();
    _lang.dispose();
    _region.dispose();
    _sort.dispose();
    _currency.dispose();
    _coin.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final p = widget.provider;
      if (p == null) {
        final config = <String, String>{
          if (_currency.text.trim().isNotEmpty) 'currency': _currency.text.trim(),
          if (_coin.text.trim().isNotEmpty) 'coin': _coin.text.trim(),
        };
        await ApiClient.instance.createProvider(
          code: _code.text.trim(),
          lang: _lang.text.trim().isEmpty ? '*' : _lang.text.trim(),
          region: _region.text.trim().isEmpty ? '*' : _region.text.trim(),
          sort: int.tryParse(_sort.text.trim()) ?? 0,
          config: config,
        );
      } else {
        final patch = <String, dynamic>{
          if (_lang.text.trim().isNotEmpty && _lang.text.trim() != p.lang)
            'lang': _lang.text.trim(),
          if (_region.text.trim().isNotEmpty && _region.text.trim() != p.region)
            'region': _region.text.trim(),
          if ((int.tryParse(_sort.text.trim()) ?? 0) != p.sort)
            'sort': int.tryParse(_sort.text.trim()) ?? 0,
        };
        final config = <String, String>{
          if (_currency.text.trim().isNotEmpty) 'currency': _currency.text.trim(),
          if (_coin.text.trim().isNotEmpty) 'coin': _coin.text.trim(),
        };
        if (config.isNotEmpty) patch['config'] = config;
        if (patch.isEmpty) {
          if (mounted) Navigator.pop(context, true);
          return;
        }
        await ApiClient.instance.updateProvider(p.id, patch);
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
      title: Text(widget.provider == null ? '新建支付方式' : '编辑支付方式'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: _code,
              enabled: widget.provider == null,
              decoration: const InputDecoration(labelText: '渠道码 *（stripe/np_usdt）'),
            ),
            TextField(
                controller: _lang,
                decoration: const InputDecoration(labelText: '语言（* 或 en）')),
            TextField(
                controller: _region,
                decoration: const InputDecoration(labelText: '地区（* 或 US）')),
            TextField(
                controller: _sort,
                decoration: const InputDecoration(labelText: '排序（升序）')),
            TextField(
              controller: _currency,
              decoration: InputDecoration(
                labelText: '下单币种（USD）',
                hintText: widget.provider?.configConfigured == true ? '留空保留原值' : '默认 USD',
              ),
            ),
            TextField(
              controller: _coin,
              decoration: InputDecoration(
                labelText: 'USDT 链（usdttrc20/usdterc20）',
                hintText: widget.provider?.configConfigured == true ? '留空保留原值' : '默认 usdttrc20',
              ),
            ),
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
