import 'package:flutter_test/flutter_test.dart';

import 'package:open_novel/main.dart';

void main() {
  testWidgets('App boots smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(const OpenNovelApp());
    await tester.pump();
    // 主壳底部导航渲染（首页/全部/我的）
    expect(find.byType(MainShell), findsOneWidget);
  });
}
