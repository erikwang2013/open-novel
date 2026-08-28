import 'package:flutter/material.dart';

import '../api/api_client.dart';
import 'login_page.dart';

/// 主框架：左侧导航 + 内容区占位，各模块页面后续实现。
class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _index = 0;

  static const _titles = ['仪表盘', '书籍', '评论', '用户'];
  static const _icons = [
    Icons.dashboard_outlined,
    Icons.menu_book_outlined,
    Icons.comment_outlined,
    Icons.people_outline,
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
            child: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(_titles[_index],
                      style: Theme.of(context).textTheme.headlineSmall),
                  const SizedBox(height: 8),
                  Text('待实现',
                      style: TextStyle(
                          color: Theme.of(context).colorScheme.outline)),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
