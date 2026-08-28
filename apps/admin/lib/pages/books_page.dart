import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'chapters_page.dart';
import 'widgets.dart';

/// 书籍管理页：分页列表 + 状态过滤，上下架、编辑元数据、钻取章节管理。
class BooksPage extends StatefulWidget {
  const BooksPage({super.key});

  @override
  State<BooksPage> createState() => _BooksPageState();
}

class _BooksPageState extends State<BooksPage> {
  static const _pageSize = 20;

  int _page = 1;
  int _status = 0; // 0 全部（管理员不过滤） 1 仅上架
  List<Book> _list = [];
  int _total = 0;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (list, total) = await ApiClient.instance.books(
          page: _page, pageSize: _pageSize, status: _status);
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

  Future<void> _toggleStatus(Book b) async {
    final on = b.status == 1;
    final ok = await confirmDialog(
        context,
        on ? '下架书籍' : '上架书籍',
        on ? '确定下架《${b.title}》？下架后 C 端不可见。' : '确定上架《${b.title}》？',
        confirmText: on ? '下架' : '上架');
    if (!ok) return;
    try {
      await ApiClient.instance.updateBookStatus(b.id, on ? 0 : 1);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(on ? '已下架' : '已上架')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  Future<void> _create() async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => const _BookEditDialog());
    if (ok == true) _load();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              const Text('状态：'),
              DropdownButton<int>(
                value: _status,
                items: const [
                  DropdownMenuItem(value: 0, child: Text('全部')),
                  DropdownMenuItem(value: 1, child: Text('上架')),
                ],
                onChanged: (v) {
                  if (v == null) return;
                  setState(() {
                    _status = v;
                    _page = 1;
                  });
                  _load();
                },
              ),
              const Spacer(),
              FilledButton.icon(
                  onPressed: () => _create(),
                  icon: const Icon(Icons.add),
                  label: const Text('新建书籍')),
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
                          DataColumn(label: Text('标题')),
                          DataColumn(label: Text('作者')),
                          DataColumn(label: Text('语言')),
                          DataColumn(label: Text('VIP')),
                          DataColumn(label: Text('状态')),
                          DataColumn(label: Text('操作')),
                        ],
                        rows: [
                          for (final b in _list)
                            DataRow(cells: [
                              DataCell(Text(b.title)),
                              DataCell(Text(b.author)),
                              DataCell(Text(b.lang)),
                              DataCell(Text(b.isVip == 1 ? '是' : '否')),
                              DataCell(Text(b.status == 1 ? '上架' : '下架')),
                              DataCell(Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  TextButton(
                                      onPressed: () => _toggleStatus(b),
                                      child: Text(
                                          b.status == 1 ? '下架' : '上架')),
                                  TextButton(
                                      onPressed: () => Navigator.of(context)
                                          .push(MaterialPageRoute(
                                              builder: (_) =>
                                                  ChaptersPage(book: b))),
                                      child: const Text('管理章节')),
                                ],
                              )),
                            ]),
                        ],
                      ),
                    ),
                    if (_list.isEmpty && !_loading)
                      const Padding(
                          padding: EdgeInsets.all(24),
                          child: Center(child: Text('暂无书籍'))),
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
}

/// 新建书籍弹窗（后端无更新书籍端点，复用 POST /api/books）。
class _BookEditDialog extends StatefulWidget {
  const _BookEditDialog();

  @override
  State<_BookEditDialog> createState() => _BookEditDialogState();
}

class _BookEditDialogState extends State<_BookEditDialog> {
  late final _title = TextEditingController();
  late final _author = TextEditingController();
  late final _summary = TextEditingController();
  late final _cover = TextEditingController();
  late final _lang = TextEditingController(text: 'zh-CN');
  bool _isVip = false;
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _title.dispose();
    _author.dispose();
    _summary.dispose();
    _cover.dispose();
    _lang.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await ApiClient.instance.createBook({
        'title': _title.text.trim(),
        'author': _author.text.trim(),
        'summary': _summary.text.trim(),
        'cover': _cover.text.trim(),
        'lang': _lang.text.trim().isEmpty ? 'zh-CN' : _lang.text.trim(),
        'isVip': _isVip ? 1 : 0,
      });
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = ApiClient.instance.errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('新建书籍'),
      content: SizedBox(
        width: 420,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                  controller: _title,
                  decoration: const InputDecoration(labelText: '标题 *')),
              TextField(
                  controller: _author,
                  decoration: const InputDecoration(labelText: '作者')),
              TextField(
                  controller: _summary,
                  maxLines: 3,
                  decoration: const InputDecoration(labelText: '简介')),
              TextField(
                  controller: _cover,
                  decoration: const InputDecoration(labelText: '封面 URL')),
              TextField(
                  controller: _lang,
                  decoration: const InputDecoration(labelText: '语言')),
              SwitchListTile(
                title: const Text('VIP 书籍'),
                value: _isVip,
                onChanged: (v) => setState(() => _isVip = v),
              ),
              if (_error != null)
                Text(_error!,
                    style:
                        TextStyle(color: Theme.of(context).colorScheme.error)),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
            onPressed: _busy ? null : () => Navigator.pop(context, false),
            child: const Text('取消')),
        FilledButton(
            onPressed: _busy ? null : _submit,
            child: _busy
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('保存')),
      ],
    );
  }
}
