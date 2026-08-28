import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../l10n/app_localizations.dart';

/// 登录 / 注册页（POST /api/v1/users/login、/api/v1/users/register）。
class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _username = TextEditingController();
  final _password = TextEditingController();
  final _email = TextEditingController();
  final _nickname = TextEditingController();
  bool _register = false;
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _username.dispose();
    _password.dispose();
    _email.dispose();
    _nickname.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      if (_register) {
        await ApiClient.instance.register(
            _username.text.trim(),
            _password.text,
            _email.text.trim(),
            _nickname.text.trim().isEmpty
                ? _username.text.trim()
                : _nickname.text.trim());
      } else {
        await ApiClient.instance.login(_username.text.trim(), _password.text);
      }
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      setState(() => _error = ApiClient.instance.errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(_register ? l10n.register : l10n.login)),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          TextField(
            controller: _username,
            decoration: InputDecoration(
                labelText: l10n.username, prefixIcon: const Icon(Icons.person)),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _password,
            obscureText: true,
            decoration: InputDecoration(
                labelText: l10n.password,
                prefixIcon: const Icon(Icons.lock_outline)),
          ),
          if (_register) ...[
            const SizedBox(height: 12),
            TextField(
              controller: _email,
              keyboardType: TextInputType.emailAddress,
              decoration: InputDecoration(
                  labelText: l10n.email,
                  prefixIcon: const Icon(Icons.email_outlined)),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _nickname,
              decoration: InputDecoration(
                  labelText: l10n.nickname,
                  prefixIcon: const Icon(Icons.badge_outlined)),
            ),
          ],
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(
                _error == 'network'
                    ? l10n.errorNetwork
                    : l10n.errorMsg(_error!),
                style: TextStyle(color: Theme.of(context).colorScheme.error)),
          ],
          const SizedBox(height: 20),
          FilledButton(
            onPressed: _busy ? null : _submit,
            child: _busy
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : Text(_register ? l10n.register : l10n.login),
          ),
          TextButton(
            onPressed: () => setState(() {
              _register = !_register;
              _error = null;
            }),
            child: Text(_register ? l10n.login : l10n.register),
          ),
        ],
      ),
    );
  }
}
