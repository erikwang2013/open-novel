import 'package:flutter/material.dart';

import '../api/api_client.dart';
import 'home_page.dart';

/// 管理端登录页（POST /api/users/login）。
/// 后端未定义角色常量，管理端以 role==3 判定管理员。
class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _username = TextEditingController();
  final _password = TextEditingController();
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _username.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final r = await ApiClient.instance
          .login(_username.text.trim(), _password.text);
      if (r.user.role != 3) {
        ApiClient.instance.logout();
        setState(() => _error = '无管理员权限，请联系管理员');
      } else if (mounted) {
        Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const HomePage()));
      }
    } catch (e) {
      setState(() {
        final m = ApiClient.instance.errorMessage(e);
        _error = m == 'network' ? '网络错误，请稍后重试' : m;
      });
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('管理员登录')),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 360),
          child: ListView(
            shrinkWrap: true,
            padding: const EdgeInsets.all(24),
            children: [
              TextField(
                controller: _username,
                decoration: const InputDecoration(
                    labelText: '用户名', prefixIcon: Icon(Icons.person)),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _password,
                obscureText: true,
                onSubmitted: (_) => _submit(),
                decoration: const InputDecoration(
                    labelText: '密码', prefixIcon: Icon(Icons.lock_outline)),
              ),
              if (_error != null) ...[
                const SizedBox(height: 12),
                Text(_error!,
                    style:
                        TextStyle(color: Theme.of(context).colorScheme.error)),
              ],
              const SizedBox(height: 20),
              FilledButton(
                onPressed: _busy ? null : _submit,
                child: _busy
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2))
                    : const Text('登录'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
