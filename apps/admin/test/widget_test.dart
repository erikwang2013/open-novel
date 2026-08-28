import 'package:flutter_test/flutter_test.dart';

import 'package:open_novel_admin/main.dart';

void main() {
  testWidgets('shows login page when not logged in', (tester) async {
    await tester.pumpWidget(const AdminApp());
    expect(find.text('管理员登录'), findsOneWidget);
  });
}
