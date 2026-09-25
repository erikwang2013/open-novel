import 'package:flutter/material.dart';

/// 项目宠物 Novi —— 加载态与空状态插图。
///
/// 素材 `assets/pet.png` 由 `docs/pet.svg` 光栅化而来（源图会动，此处取静止首帧）。
/// 后端 `/api/pet.svg` 是同一位宠物的活体版本：按 MySQL / Redis / OpenSearch
/// 探测结果换表情（happy / degraded / down）。
class NoviPet extends StatelessWidget {
  const NoviPet({super.key, this.size = 120, this.busy = false, this.text});

  /// 图片边长（逻辑像素）。
  final double size;

  /// 加载中：宠物下方补一个进度圈，避免只剩静态图的“假死”观感。
  final bool busy;

  /// 空状态文案（通常传 `l10n.empty`）。
  final String? text;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Image.asset('assets/pet.png',
              width: size, height: size, filterQuality: FilterQuality.medium),
          if (busy) ...[
            const SizedBox(height: 12),
            const SizedBox(
                width: 22, height: 22,
                child: CircularProgressIndicator(strokeWidth: 2.4)),
          ],
          if (text != null) ...[
            const SizedBox(height: 10),
            Text(text!, style: Theme.of(context).textTheme.bodyMedium),
          ],
        ],
      ),
    );
  }
}
