import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 分类 / 标签管理页（T-A-16）：TabBar 切换，DataTable + 新建/编辑/删除。
class CategoriesPage extends StatefulWidget {
  const CategoriesPage({super.key});

  @override
  State<CategoriesPage> createState() => _CategoriesPageState();
}

class _CategoriesPageState extends State<CategoriesPage>
    with SingleTickerProviderStateMixin {
  late final TabController _tab = TabController(length: 2, vsync: this);

  List<Category> _cats = [];
  List<Tag> _tags = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _tab.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (cats, _) = await ApiClient.instance.categories();
      final (tags, _) = await ApiClient.instance.tags();
      if (!mounted) return;
      setState(() {
        _cats = cats;
        _tags = tags;
      });
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _createCategory() async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => const _CategoryDialog());
    if (ok == true) _load();
  }

  Future<void> _editCategory(Category c) async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => _CategoryDialog(category: c));
    if (ok == true) _load();
  }

  Future<void> _deleteCategory(Category c) async {
    final ok = await confirmDialog(
        context, '删除分类', '确定删除分类「${c.name}」？关联将被一并清理。',
        confirmText: '删除');
    if (!ok) return;
    try {
      await ApiClient.instance.deleteCategory(c.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('已删除')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  Future<void> _createTag() async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => const _TagDialog());
    if (ok == true) _load();
  }

  Future<void> _editTag(Tag t) async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => _TagDialog(tag: t));
    if (ok == true) _load();
  }

  Future<void> _deleteTag(Tag t) async {
    final ok = await confirmDialog(
        context, '删除标签', '确定删除标签「${t.name}」？关联将被一并清理。',
        confirmText: '删除');
    if (!ok) return;
    try {
      await ApiClient.instance.deleteTag(t.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('已删除')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        TabBar(
          controller: _tab,
          tabs: const [Tab(text: '分类'), Tab(text: '标签')],
        ),
        Expanded(
          child: _loading && _cats.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : TabBarView(
                  controller: _tab,
                  children: [_buildCategories(), _buildTags()],
                ),
        ),
      ],
    );
  }

  Widget _buildCategories() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              Text('共 ${_cats.length} 个分类'),
              const Spacer(),
              FilledButton.icon(
                  onPressed: _createCategory,
                  icon: const Icon(Icons.add),
                  label: const Text('新建分类')),
            ],
          ),
        ),
        Expanded(
          child: _cats.isEmpty
              ? const Center(child: Text('暂无分类'))
              : SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: DataTable(
                    columns: const [
                      DataColumn(label: Text('ID')),
                      DataColumn(label: Text('名称')),
                      DataColumn(label: Text('父分类')),
                      DataColumn(label: Text('排序')),
                      DataColumn(label: Text('状态')),
                      DataColumn(label: Text('操作')),
                    ],
                    rows: [
                      for (final c in _cats)
                        DataRow(cells: [
                          DataCell(Text(c.id)),
                          DataCell(Text(c.name)),
                          DataCell(Text(c.parentId == '0' ? '—' : c.parentId)),
                          DataCell(Text('${c.sortOrder}')),
                          DataCell(_StatusTag(c.status == 1)),
                          DataCell(Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              TextButton(
                                  onPressed: () => _editCategory(c),
                                  child: const Text('编辑')),
                              TextButton(
                                  onPressed: () => _deleteCategory(c),
                                  child: const Text('删除')),
                            ],
                          )),
                        ]),
                    ],
                  ),
                ),
        ),
      ],
    );
  }

  Widget _buildTags() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              Text('共 ${_tags.length} 个标签'),
              const Spacer(),
              FilledButton.icon(
                  onPressed: _createTag,
                  icon: const Icon(Icons.add),
                  label: const Text('新建标签')),
            ],
          ),
        ),
        Expanded(
          child: _tags.isEmpty
              ? const Center(child: Text('暂无标签'))
              : SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: DataTable(
                    columns: const [
                      DataColumn(label: Text('ID')),
                      DataColumn(label: Text('名称')),
                      DataColumn(label: Text('语言')),
                      DataColumn(label: Text('状态')),
                      DataColumn(label: Text('操作')),
                    ],
                    rows: [
                      for (final t in _tags)
                        DataRow(cells: [
                          DataCell(Text(t.id)),
                          DataCell(Text(t.name)),
                          DataCell(Text(t.lang)),
                          DataCell(_StatusTag(t.status == 1)),
                          DataCell(Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              TextButton(
                                  onPressed: () => _editTag(t),
                                  child: const Text('编辑')),
                              TextButton(
                                  onPressed: () => _deleteTag(t),
                                  child: const Text('删除')),
                            ],
                          )),
                        ]),
                    ],
                  ),
                ),
        ),
      ],
    );
  }
}

/// 启用/禁用状态标签。
class _StatusTag extends StatelessWidget {
  const _StatusTag(this.on);

  final bool on;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: on ? Colors.green.shade100 : Colors.red.shade100,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(on ? '启用' : '禁用',
          style: TextStyle(
              fontSize: 12,
              color: on ? Colors.green.shade900 : Colors.red.shade900)),
    );
  }
}

/// 新建/编辑分类弹窗；编辑态仅提交变更字段。
class _CategoryDialog extends StatefulWidget {
  const _CategoryDialog({this.category});

  final Category? category;

  @override
  State<_CategoryDialog> createState() => _CategoryDialogState();
}

class _CategoryDialogState extends State<_CategoryDialog> {
  late final _name = TextEditingController(text: widget.category?.name ?? '');
  late final _parent =
      TextEditingController(text: widget.category?.parentId ?? '0');
  late final _sort = TextEditingController(
      text: widget.category == null ? '0' : '${widget.category!.sortOrder}');
  late bool _enabled = widget.category?.status != 0;
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _name.dispose();
    _parent.dispose();
    _sort.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final c = widget.category;
      if (c == null) {
        await ApiClient.instance.createCategory(
          name: _name.text.trim(),
          parentId: _parent.text.trim().isEmpty ? '0' : _parent.text.trim(),
          sortOrder: int.tryParse(_sort.text.trim()) ?? 0,
        );
      } else {
        final patch = <String, dynamic>{
          if (_name.text.trim() != c.name) 'name': _name.text.trim(),
          if ((_parent.text.trim().isEmpty ? '0' : _parent.text.trim()) !=
              c.parentId)
            'parent_id': _parent.text.trim().isEmpty ? '0' : _parent.text.trim(),
          if ((int.tryParse(_sort.text.trim()) ?? 0) != c.sortOrder)
            'sort_order': int.tryParse(_sort.text.trim()) ?? 0,
          if ((_enabled ? 1 : 0) != c.status) 'status': _enabled ? 1 : 0,
        };
        if (patch.isEmpty) {
          if (mounted) Navigator.pop(context, true);
          return;
        }
        await ApiClient.instance.updateCategory(c.id, patch);
      }
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
      title: Text(widget.category == null ? '新建分类' : '编辑分类'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
                controller: _name,
                decoration: const InputDecoration(labelText: '分类名 *')),
            TextField(
                controller: _parent,
                decoration:
                    const InputDecoration(labelText: '父分类 ID（0=一级）')),
            TextField(
                controller: _sort,
                decoration:
                    const InputDecoration(labelText: '排序（升序）')),
            SwitchListTile(
              title: const Text('启用'),
              value: _enabled,
              onChanged: (v) => setState(() => _enabled = v),
            ),
            if (_error != null)
              Text(_error!,
                  style: TextStyle(
                      color: Theme.of(context).colorScheme.error)),
          ],
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

/// 新建/编辑标签弹窗。
class _TagDialog extends StatefulWidget {
  const _TagDialog({this.tag});

  final Tag? tag;

  @override
  State<_TagDialog> createState() => _TagDialogState();
}

class _TagDialogState extends State<_TagDialog> {
  late final _name = TextEditingController(text: widget.tag?.name ?? '');
  late final _lang = TextEditingController(text: widget.tag?.lang ?? 'zh-CN');
  late bool _enabled = widget.tag?.status != 0;
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _name.dispose();
    _lang.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final t = widget.tag;
      if (t == null) {
        await ApiClient.instance.createTag(
            name: _name.text.trim(),
            lang: _lang.text.trim().isEmpty ? 'zh-CN' : _lang.text.trim());
      } else {
        final patch = <String, dynamic>{
          if (_name.text.trim() != t.name) 'name': _name.text.trim(),
          if ((_lang.text.trim().isEmpty ? 'zh-CN' : _lang.text.trim()) !=
              t.lang)
            'lang': _lang.text.trim().isEmpty ? 'zh-CN' : _lang.text.trim(),
          if ((_enabled ? 1 : 0) != t.status) 'status': _enabled ? 1 : 0,
        };
        if (patch.isEmpty) {
          if (mounted) Navigator.pop(context, true);
          return;
        }
        await ApiClient.instance.updateTag(t.id, patch);
      }
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
      title: Text(widget.tag == null ? '新建标签' : '编辑标签'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
                controller: _name,
                decoration: const InputDecoration(labelText: '标签名 *')),
            TextField(
                controller: _lang,
                decoration: const InputDecoration(labelText: '语言')),
            SwitchListTile(
              title: const Text('启用'),
              value: _enabled,
              onChanged: (v) => setState(() => _enabled = v),
            ),
            if (_error != null)
              Text(_error!,
                  style: TextStyle(
                      color: Theme.of(context).colorScheme.error)),
          ],
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
