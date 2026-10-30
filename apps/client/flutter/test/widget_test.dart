import 'package:flutter_test/flutter_test.dart';

import 'package:open_novel/main.dart';

void main() {
  testWidgets('App boots smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(const OpenNovelApp());
    await tester.pump();
    // 首页 initState 会发起推荐请求；测试 HttpClient 立即返回错误，
    // 需推进时钟让 dio 的超时 timer 结束，否则触发 !timersPending。
    await tester.pump(const Duration(seconds: 1));
    await tester.pumpAndSettle();
    // 主壳底部导航渲染（首页/全部/我的）
    expect(find.byType(MainShell), findsOneWidget);
  });
}
