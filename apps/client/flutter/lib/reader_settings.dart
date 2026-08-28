import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// 阅读器设置：字号 / 行距 / 主题 / 翻页方式。
/// ValueNotifier + shared_preferences 持久化（模式同 localeNotifier）。
class ReaderSettings {
  static const double fontSizeMin = 13;
  static const double fontSizeMax = 26;
  static const double lineHeightMin = 1.2;
  static const double lineHeightMax = 2.6;

  static final ValueNotifier<double> fontSize = ValueNotifier(17);
  static final ValueNotifier<double> lineHeight = ValueNotifier(1.8);
  static final ValueNotifier<ThemeMode> themeMode =
      ValueNotifier(ThemeMode.system);
  static final ValueNotifier<int> pageMode = ValueNotifier(0); // 0 上下滚动 1 左右翻页

  /// main() 启动时调用一次，从磁盘恢复设置。
  static Future<void> init() async {
    final p = await SharedPreferences.getInstance();
    fontSize.value =
        (p.getDouble('fontSize') ?? 17).clamp(fontSizeMin, fontSizeMax).toDouble();
    lineHeight.value = (p.getDouble('lineHeight') ?? 1.8)
        .clamp(lineHeightMin, lineHeightMax)
        .toDouble();
    themeMode.value = ThemeMode.values[p.getInt('themeMode') ?? 0];
    pageMode.value = p.getInt('pageMode') ?? 0;
  }

  static void setFontSize(double v) {
    fontSize.value = v;
    SharedPreferences.getInstance()
        .then((p) => p.setDouble('fontSize', v));
  }

  static void setLineHeight(double v) {
    lineHeight.value = v;
    SharedPreferences.getInstance()
        .then((p) => p.setDouble('lineHeight', v));
  }

  static void setThemeMode(ThemeMode m) {
    themeMode.value = m;
    SharedPreferences.getInstance().then((p) => p.setInt('themeMode', m.index));
  }

  static void setPageMode(int m) {
    pageMode.value = m;
    SharedPreferences.getInstance().then((p) => p.setInt('pageMode', m));
  }
}
