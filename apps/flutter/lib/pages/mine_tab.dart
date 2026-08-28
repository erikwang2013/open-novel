import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';
import 'login_page.dart';

/// 我的 tab：登录态展示 / 登录入口 / 退出。
class MineTab extends StatefulWidget {
  const MineTab({super.key});

  @override
  State<MineTab> createState() => _MineTabState();
}

class _MineTabState extends State<MineTab> {
  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final user = ApiClient.instance.currentUser;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.mine)),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            CircleAvatar(
              radius: 40,
              child: Icon(Icons.person,
                  size: 44, color: Theme.of(context).colorScheme.primary),
            ),
            const SizedBox(height: 16),
            Text(
              user != null
                  ? l10n.welcome(
                      user.nickname.isNotEmpty ? user.nickname : user.username)
                  : l10n.notLoggedIn,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(user?.username ?? '',
                style: Theme.of(context).textTheme.bodySmall),
            const SizedBox(height: 24),
            if (user == null)
              FilledButton(
                onPressed: () => Navigator.of(context).push(
                  MaterialPageRoute(builder: (_) => const LoginPage()),
                ),
                child: Text(l10n.login),
              )
            else
              OutlinedButton(
                onPressed: () {
                  ApiClient.instance.logout();
                  setState(() {});
                },
                child: Text(l10n.logout),
              ),
          ],
        ),
      ),
    );
  }
}
