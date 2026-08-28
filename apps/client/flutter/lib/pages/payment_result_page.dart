import 'dart:async';

import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';

/// 支付结果页（T-P-15）：每 3 秒轮询 GetOrder。
/// status 0 待确认 / 1 已付 / 2 失败 / 3 关闭；已付或失败/关闭即停止轮询。
class PaymentResultPage extends StatefulWidget {
  const PaymentResultPage({super.key, required this.orderNo});

  final String orderNo;

  @override
  State<PaymentResultPage> createState() => _PaymentResultPageState();
}

class _PaymentResultPageState extends State<PaymentResultPage> {
  Timer? _timer;
  int _status = 0; // 0 待确认 1 已付 2 失败/关闭
  String? _error;

  @override
  void initState() {
    super.initState();
    _poll();
    _timer = Timer.periodic(const Duration(seconds: 3), (_) => _poll());
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  Future<void> _poll() async {
    final api = ApiClient.instance;
    try {
      final o = await api.getOrder(widget.orderNo);
      if (!mounted) return;
      setState(() {
        _error = null;
        _status = o.status == 1 ? 1 : (o.status == 0 ? 0 : 2);
      });
      if (o.status != 0) _timer?.cancel(); // 终态停止轮询
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = api.errorMessage(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final (icon, title, color) = switch (_status) {
      1 => (Icons.check_circle, l10n.paymentSuccess, Colors.green),
      2 => (Icons.cancel, l10n.paymentFailed, Colors.red),
      _ => (Icons.hourglass_top, l10n.paymentPending, Colors.orange),
    };
    return Scaffold(
      appBar: AppBar(title: Text(l10n.paymentResult)),
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 72, color: color),
            const SizedBox(height: 16),
            Text(title, style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 8),
            if (_status == 0)
              Text(l10n.paymentChecking,
                  style: Theme.of(context).textTheme.bodySmall),
            if (_error != null && _error != 'network')
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(l10n.errorServer(_error!),
                    style: Theme.of(context).textTheme.bodySmall),
              ),
            if (_status == 2) ...[
              const SizedBox(height: 24),
              FilledButton.tonal(
                onPressed: () => Navigator.of(context).pop(), // 返回购买页重试
                child: Text(l10n.retryPay),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
