import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// CDN 厂商管理页（T-CDN-11）：列表 / 启停 / 排序 / 密钥配置。
/// 密钥只显示「已配置/未配置」，绝不回显明文；config 明文输入，后端校验键名并加密落库。
class CdnProvidersPage extends StatefulWidget {
  const CdnProvidersPage({super.key});

  @override
  State<CdnProvidersPage> createState() => _CdnProvidersPageState();
}

class _CdnProvidersPageState extends State<CdnProvidersPage> {
  List<CdnProvider> _items = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (items, _) = await ApiClient.instance.cdnProviders();
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
        context: context, builder: (_) => const _CdnProviderDialog());
    if (ok == true) _load();
  }

  Future<void> _edit(CdnProvider p) async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => _CdnProviderDialog(provider: p));
    if (ok == true) _load();
  }

  Future<void> _toggle(CdnProvider p) async {
    try {
      await ApiClient.instance.toggleCdnProvider(p.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(p.enabled == 1 ? '已禁用' : '已启用')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  Future<void> _delete(CdnProvider p) async {
    final ok = await confirmDialog(
        context, '删除 CDN 厂商', '确定删除 CDN 厂商「${p.code}」？', confirmText: '删除');
    if (!ok) return;
    try {
      await ApiClient.instance.deleteCdnProvider(p.id);
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
              Text('共 ${_items.length} 个 CDN 厂商'),
              const Spacer(),
              FilledButton.icon(
                  onPressed: _create,
                  icon: const Icon(Icons.add),
                  label: const Text('新建 CDN 厂商')),
            ],
          ),
        ),
        Expanded(
          child: _loading && _items.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : _items.isEmpty
                  ? const Center(child: Text('暂无 CDN 厂商'))
                  : SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      child: DataTable(
                        columns: const [
                          DataColumn(label: Text('厂商码')),
                          DataColumn(label: Text('排序')),
                          DataColumn(label: Text('密钥')),
                          DataColumn(label: Text('状态')),
                          DataColumn(label: Text('操作')),
                        ],
                        rows: [
                          for (final p in _items)
                            DataRow(cells: [
                              DataCell(Text(p.code)),
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

/// 厂商配置字段（§3.3）：key → (label, hint)。编辑时留空 = 保留原值（后端合并重加密）。
const _vendorConfigFields = <String, List<(String, String, String)>>{
  'cloudflare': [
    ('zone_id', 'Zone ID', 'Cloudflare 区域 ID'),
    ('api_token', 'API Token', 'Cloudflare API Token'),
    ('batch_size', 'Batch Size', '默认 30'),
  ],
  'cloudfront': [
    ('access_key_id', 'Access Key ID', 'AWS Access Key ID'),
    ('secret_access_key', 'Secret Access Key', 'AWS Secret Access Key'),
    ('distribution_id', 'Distribution ID', 'CloudFront 分发 ID'),
    ('batch_size', 'Batch Size', '默认 3000'),
  ],
  'aliyun': [
    ('access_key_id', 'Access Key ID', '阿里云 AccessKey ID'),
    ('access_key_secret', 'Access Key Secret', '阿里云 AccessKey Secret'),
    ('batch_size', 'Batch Size', '默认 1000'),
    ('rate_limit_qps', 'Rate Limit QPS', '默认 50'),
  ],
  'tencent': [
    ('secret_id', 'Secret ID', '腾讯云 SecretId'),
    ('secret_key', 'Secret Key', '腾讯云 SecretKey'),
    ('batch_size', 'Batch Size', '默认 1000'),
    ('rate_limit_qps', 'Rate Limit QPS', '默认 20'),
  ],
};

/// 新建/编辑 CDN 厂商弹窗。密钥字段加密存储，编辑时留空 = 保留原值。
class _CdnProviderDialog extends StatefulWidget {
  const _CdnProviderDialog({this.provider});

  final CdnProvider? provider;

  @override
  State<_CdnProviderDialog> createState() => _CdnProviderDialogState();
}

class _CdnProviderDialogState extends State<_CdnProviderDialog> {
  String _vendor = '';
  late final _sort = TextEditingController(text: '0');
  final _config = <String, TextEditingController>{};
  bool _busy = false;
  String? _error;

  List<(String, String, String)> get _fields =>
      _vendorConfigFields[_vendor] ?? const [];

  @override
  void initState() {
    super.initState();
    _vendor = widget.provider?.code ?? 'cloudflare';
    _sort.text = '${widget.provider?.sort ?? 0}';
    _reinitConfig();
  }

  void _reinitConfig() {
    _config.clear();
    for (final (key, _, _) in _fields) {
      _config[key] = TextEditingController();
    }
  }

  void _onVendor(String v) {
    setState(() {
      _vendor = v;
      _reinitConfig();
    });
  }

  @override
  void dispose() {
    _sort.dispose();
    for (final c in _config.values) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final p = widget.provider;
      final config = <String, String>{
        for (final (key, _, _) in _fields)
          if (_config[key]!.text.trim().isNotEmpty)
            key: _config[key]!.text.trim(),
      };
      if (p == null) {
        await ApiClient.instance.createCdnProvider(
          code: _vendor,
          sort: int.tryParse(_sort.text.trim()) ?? 0,
          config: config,
        );
      } else {
        final patch = <String, dynamic>{};
        if ((int.tryParse(_sort.text.trim()) ?? 0) != p.sort) {
          patch['sort'] = int.tryParse(_sort.text.trim()) ?? 0;
        }
        if (config.isNotEmpty) patch['config'] = config;
        if (patch.isEmpty) {
          if (mounted) Navigator.pop(context, true);
          return;
        }
        await ApiClient.instance.updateCdnProvider(p.id, patch);
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
      title: Text(widget.provider == null ? '新建 CDN 厂商' : '编辑 CDN 厂商'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (widget.provider == null)
              DropdownButtonFormField<String>(
                initialValue: _vendor,
                items: [
                  for (final v in _vendorConfigFields.keys)
                    DropdownMenuItem(value: v, child: Text(v)),
                ],
                onChanged: (v) {
                  if (v != null) _onVendor(v);
                },
                decoration: const InputDecoration(labelText: '厂商码 *'),
              )
            else
              TextField(
                controller: TextEditingController(text: _vendor),
                enabled: false,
                decoration: const InputDecoration(labelText: '厂商码 *'),
              ),
            TextField(
                controller: _sort,
                decoration: const InputDecoration(labelText: '排序（升序）')),
            for (final (key, label, hint) in _fields)
              TextField(
                controller: _config[key],
                decoration: InputDecoration(
                  labelText: label,
                  hintText: widget.provider?.configConfigured == true
                      ? '留空保留原值'
                      : hint,
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
