import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 用户管理页：搜索、分页列表、封禁/解封、角色调整。
class UsersPage extends StatefulWidget {
  const UsersPage({super.key});

  @override
  State<UsersPage> createState() => _UsersPageState();
}

class _UsersPageState extends State<UsersPage> {
  static const _pageSize = 20;

  final _search = TextEditingController();
  int _page = 1;
  List<User> _list = [];
  int _total = 0;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (list, total) = await ApiClient.instance.users(
          search: _search.text.trim(), page: _page, pageSize: _pageSize);
      if (!mounted) return;
      setState(() {
        _list = list;
        _total = total;
      });
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _toggleStatus(User u) async {
    final ban = u.status == 1;
    final ok = await confirmDialog(
        context,
        ban ? '封禁用户' : '解封用户',
        ban ? '确定封禁 ${u.username}？封禁后无法登录。' : '确定解封 ${u.username}？',
        confirmText: ban ? '封禁' : '解封');
    if (!ok) return;
    try {
      await ApiClient.instance.updateUserStatus(u.id, ban ? 0 : 1);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(ban ? '已封禁' : '已解封')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  Future<void> _changeRole(User u, int role) async {
    if (role == u.role) return;
    final ok = await confirmDialog(context, '调整角色',
        '确定将 ${u.username} 的角色改为「${_roleText(role)}」？',
        confirmText: '确定');
    if (!ok) return;
    try {
      await ApiClient.instance.updateUserRole(u.id, role);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('角色已更新')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  void _applyFilter() {
    setState(() => _page = 1);
    _load();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                    controller: _search,
                    decoration: const InputDecoration(
                        labelText: '搜索用户名/昵称/邮箱', isDense: true),
                    onSubmitted: (_) => _applyFilter()),
              ),
              const SizedBox(width: 12),
              FilledButton(onPressed: _applyFilter, child: const Text('查询')),
            ],
          ),
        ),
        Expanded(
          child: _loading && _list.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : ListView(
                  children: [
                    SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      child: DataTable(
                        columns: const [
                          DataColumn(label: Text('ID')),
                          DataColumn(label: Text('用户名')),
                          DataColumn(label: Text('昵称')),
                          DataColumn(label: Text('邮箱')),
                          DataColumn(label: Text('角色')),
                          DataColumn(label: Text('状态')),
                          DataColumn(label: Text('注册时间')),
                          DataColumn(label: Text('操作')),
                        ],
                        rows: [
                          for (final u in _list)
                            DataRow(cells: [
                              DataCell(Text(u.id)),
                              DataCell(Text(u.username)),
                              DataCell(Text(u.nickname)),
                              DataCell(Text(u.email)),
                              DataCell(Text(_roleText(u.role))),
                              DataCell(_StatusTag(u.status)),
                              DataCell(Text(u.createdAt)),
                              DataCell(Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  TextButton(
                                      onPressed: () => _toggleStatus(u),
                                      child: Text(u.status == 1 ? '封禁' : '解封')),
                                  DropdownButton<int>(
                                    value: u.role,
                                    underline: const SizedBox.shrink(),
                                    items: const [
                                      DropdownMenuItem(
                                          value: 1, child: Text('读者')),
                                      DropdownMenuItem(
                                          value: 2, child: Text('作者')),
                                      DropdownMenuItem(
                                          value: 3, child: Text('管理员')),
                                    ],
                                    onChanged: (v) {
                                      if (v != null) _changeRole(u, v);
                                    },
                                  ),
                                ],
                              )),
                            ]),
                        ],
                      ),
                    ),
                    if (_list.isEmpty && !_loading)
                      const Padding(
                          padding: EdgeInsets.all(24),
                          child: Center(child: Text('暂无用户'))),
                  ],
                ),
        ),
        PagingBar(
            page: _page,
            pageSize: _pageSize,
            total: _total,
            onChanged: (p) {
              setState(() => _page = p);
              _load();
            }),
      ],
    );
  }

  static String _roleText(int r) =>
      r == 3 ? '管理员' : (r == 2 ? '作者' : '读者');
}

/// 封禁状态标签。
class _StatusTag extends StatelessWidget {
  const _StatusTag(this.status);

  final int status;

  @override
  Widget build(BuildContext context) {
    final on = status == 1;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: on ? Colors.green.shade100 : Colors.red.shade100,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(on ? '正常' : '已封禁',
          style: TextStyle(
              fontSize: 12,
              color: on ? Colors.green.shade900 : Colors.red.shade900)),
    );
  }
}
