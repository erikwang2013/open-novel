import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import '../main.dart';
import '../models/models.dart';
import 'payment_result_page.dart';

/// VIP 购买页（T-P-14）：公开套餐列表 → 支付方式提示 → 下单 → 跳转支付 → 结果页轮询。
class VipPage extends StatefulWidget {
  const VipPage({super.key});

  @override
  State<VipPage> createState() => _VipPageState();
}

class _VipPageState extends State<VipPage> {
  List<Plan>? _plans;
  List<PaymentMethod>? _methods;
  String? _error;
  bool _ordering = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final api = ApiClient.instance;
    setState(() {
      _error = null;
      _plans = null;
      _methods = null;
    });
    try {
      final lang = langCode(localeNotifier.value);
      final plans = await api.listPublicPlans();
      final methods = await api.listMethods(lang: lang);
      if (!mounted) return;
      setState(() {
        _plans = plans;
        _methods = methods;
      });
    } catch (e) {
      setState(() => _error = api.errorMessage(e));
    }
  }

  Future<void> _buy(Plan plan) async {
    final api = ApiClient.instance;
    setState(() => _ordering = true);
    try {
      final lang = langCode(localeNotifier.value);
      final o = await api.createOrder(plan.planCode, lang: lang);
      final ok = await api.openCheckoutUrl(o.checkoutUrl);
      if (!mounted) return;
      if (!ok) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.errorServer('launch failed'))),
        );
        return;
      }
      // 跳转外部支付后进结果页轮询（用户回跳 App 时页面仍在轮询）
      await Navigator.of(context).push(MaterialPageRoute(
        builder: (_) => PaymentResultPage(orderNo: o.orderNo),
      ));
      if (mounted) _load(); // 回到购买页时刷新（可能已开通）
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(SnackBar(content: Text(api.errorMessage(e))));
    } finally {
      if (mounted) setState(() => _ordering = false);
    }
  }

  AppLocalizations get l10n => AppLocalizations.of(context)!;

  String _price(Plan p) => '${(p.amountCents / 100).toStringAsFixed(2)} ${p.currency}';

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(l10n.vipOpen)),
      body: _buildBody(context),
    );
  }

  Widget _buildBody(BuildContext context) {
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error == 'network' ? l10n.errorNetwork : l10n.errorServer(_error!)),
            OutlinedButton(onPressed: _load, child: Text(l10n.retry)),
          ],
        ),
      );
    }
    final plans = _plans;
    final methods = _methods;
    if (plans == null || methods == null) {
      return const Center(child: CircularProgressIndicator());
    }
    final methodText = methods.isEmpty
        ? l10n.paymentNoMethod
        : methods.map((m) => m.code).join(' / ');
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(l10n.vipNotActive,
              style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 12),
          for (final p in plans)
            Card(
              child: ListTile(
                title: Text(
                    p.label.isNotEmpty ? p.label : p.planCode,
                    style: const TextStyle(fontWeight: FontWeight.bold)),
                subtitle: Text('${p.days} days · ${_price(p)}'),
                trailing: FilledButton.tonal(
                  onPressed: _ordering ? null : () => _buy(p),
                  child: Text(l10n.payNow),
                ),
              ),
            ),
          const SizedBox(height: 8),
          Text(methodText, style: Theme.of(context).textTheme.bodySmall),
        ],
      ),
    );
  }
}
