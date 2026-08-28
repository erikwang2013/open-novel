import 'package:flutter/material.dart';

import '../api/api_client.dart';
import 'books_page.dart';
import 'categories_page.dart';
import 'comments_page.dart';
import 'dashboard_page.dart';
import 'login_page.dart';
import 'reports_page.dart';
import 'users_page.dart';

/// 主框架：左侧导航 + 内容区，IndexedStack 保持各页状态。
class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _index = 0;

  static const _titles = ['仪表盘', '书籍', '评论', '举报', '用户', '分类标签'];
  static const _icons = [
    Icons.dashboard_outlined,
    Icons.menu_book_outlined,
    Icons.comment_outlined,
    Icons.report_outlined,
    Icons.people_outline,
    Icons.category_outlined,
  ];

  static const _pages = [
    DashboardPage(),
    BooksPage(),
    CommentsPage(),
    ReportsPage(),
    UsersPage(),
    CategoriesPage(),
  ];

  void _logout() {
    ApiClient.instance.logout();
    Navigator.of(context).pushAndRemoveUntil(
        MaterialPageRoute(builder: (_) => const LoginPage()),
        (route) => false);
  }

  @override
  Widget build(BuildContext context) {
    final u = ApiClient.instance.currentUser;
    final nickname = u?.nickname ?? '';
    final name = nickname.isNotEmpty ? nickname : (u?.username ?? '管理员');
    return Scaffold(
      appBar: AppBar(
        title: Text(_titles[_index]),
        actions: [
          Center(
              child: Text(name,
                  style: Theme.of(context).textTheme.titleMedium)),
          IconButton(
              tooltip: '退出登录',
              onPressed: _logout,
              icon: const Icon(Icons.logout)),
        ],
      ),
      body: Row(
        children: [
          NavigationRail(
            selectedIndex: _index,
            onDestinationSelected: (i) => setState(() => _index = i),
            labelType: NavigationRailLabelType.all,
            destinations: [
              for (var i = 0; i < _titles.length; i++)
                NavigationRailDestination(
                    icon: Icon(_icons[i]), label: Text(_titles[i])),
            ],
          ),
          const VerticalDivider(thickness: 1, width: 1),
          Expanded(
            child: IndexedStack(index: _index, children: _pages),
          ),
        ],
      ),
    );
  }
}

