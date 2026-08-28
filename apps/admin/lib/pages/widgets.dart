import 'package:flutter/material.dart';

import '../api/api_client.dart';

/// 确认弹窗；用户确认返回 true。
Future<bool> confirmDialog(BuildContext context, String title, String message,
    {String confirmText = '确认'}) async {
  final ok = await showDialog<bool>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: Text(title),
      content: Text(message),
      actions: [
        TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('取消')),
        FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(confirmText)),
      ],
    ),
  );
  return ok ?? false;
}

/// 统一错误提示（snackbar）。
void showErr(BuildContext context, Object e) {
  final m = ApiClient.instance.errorMessage(e);
  ScaffoldMessenger.of(context)
      .showSnackBar(SnackBar(content: Text(m == 'network' ? '网络错误，请稍后重试' : m)));
}

/// 上一页 / 下一页分页条。
class PagingBar extends StatelessWidget {
  const PagingBar({
    super.key,
    required this.page,
    required this.pageSize,
    required this.total,
    required this.onChanged,
  });

  final int page;
  final int pageSize;
  final int total;
  final ValueChanged<int> onChanged;

  int get _pages =>
      pageSize <= 0 || total <= 0 ? 1 : (total + pageSize - 1) ~/ pageSize;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          IconButton(
            tooltip: '上一页',
            onPressed: page > 1 ? () => onChanged(page - 1) : null,
            icon: const Icon(Icons.chevron_left),
          ),
          Text('第 $page / $_pages 页 · 共 $total 条'),
          IconButton(
            tooltip: '下一页',
            onPressed: page < _pages ? () => onChanged(page + 1) : null,
            icon: const Icon(Icons.chevron_right),
          ),
        ],
      ),
    );
  }
}
