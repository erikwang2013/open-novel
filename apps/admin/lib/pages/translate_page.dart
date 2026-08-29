import 'package:dio/dio.dart';
import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// 翻译管理页：选中书籍 → 目标语言 → 机器翻译（标题/简介、全部章节）+ 人工编辑。
/// 已翻译语言为页面内跟踪（书籍详情接口无 translations 字段）。
class TranslatePage extends StatefulWidget {
  const TranslatePage({super.key});

  @override
  State<TranslatePage> createState() => _TranslatePageState();
}

class _TranslatePageState extends State<TranslatePage> {
  static const _pageSize = 20;
  static const _langs = [
    'zh-CN', 'en', 'ja', 'ko', 'fr', 'de', 'es', 'ru', 'pt', 'hi', 'ar', 'bn', 'id',
  ];
  static const _langNames = {
    'zh-CN': '简体中文', 'en': 'English', 'ja': '日本語', 'ko': '한국어',
    'fr': 'Français', 'de': 'Deutsch', 'es': 'Español', 'ru': 'Русский',
    'pt': 'Português', 'hi': 'हिन्दी', 'ar': 'العربية', 'bn': 'বাংলা',
    'id': 'Bahasa Indonesia',
  };

  int _page = 1;
  List<Book> _list = [];
  int _total = 0;
  bool _loading = false;

  Book? _sel;
  String _lang = 'en';
  final Set<String> _doneLangs = {};
  bool _busy = false;
  String? _result;

  final _titleCtrl = TextEditingController();
  final _summaryCtrl = TextEditingController();
  bool _editBusy = false;
  String? _editError;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _titleCtrl.dispose();
    _summaryCtrl.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (list, total) = await ApiClient.instance.books(
          page: _page, pageSize: _pageSize);
      if (!mounted) return;
      setState(() {
        _list = list;
        _total = total;
        if (_sel != null) {
          _sel = list.firstWhere((b) => b.id == _sel!.id,
              orElse: () => _sel!);
        }
      });
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _select(Book b) {
    setState(() {
      _sel = b;
      _result = null;
      _editError = null;
      _titleCtrl.text = '';
      _summaryCtrl.text = '';
    });
  }

  /// 未配置翻译 key（180405）→ 专门提示；其余走统一错误。
  void _handleError(Object e) {
    if (e is DioException) {
      final data = e.response?.data;
      if (data is Map && data['code'] == 180405) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text('未配置翻译 API key（env TRANSLATE_API_KEY）')));
        return;
      }
    }
    showErr(context, e);
  }

  Future<void> _translateBook() async {
    final b = _sel!;
    setState(() {
      _busy = true;
      _result = null;
    });
    try {
      final d = await ApiClient.instance.translateBook(b.id, _lang);
      if (!mounted) return;
      setState(() {
        _doneLangs.add(_lang);
        _titleCtrl.text = asStr(d['title']);
        _summaryCtrl.text = asStr(d['summary']);
        _result = '标题与简介已翻译为 ${_langNames[_lang]}，可下方人工编辑微调';
      });
    } catch (e) {
      if (!mounted) return;
      _handleError(e);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _translateChapters() async {
    final b = _sel!;
    setState(() {
      _busy = true;
      _result = null;
    });
    try {
      final d = await ApiClient.instance.translateBookChapters(b.id, _lang);
      if (!mounted) return;
      final failed = ((d['failed_chapters'] as List? ?? []))
          .map((e) => e.toString())
          .join(', ');
      setState(() {
        _doneLangs.add(_lang);
        _result = '章节翻译完成：共 ${asInt(d['total'])} 章，'
            '成功 ${asInt(d['succeeded'])}，失败 ${asInt(d['failed'])}'
            '${failed.isEmpty ? '' : '，失败章节号：$failed'}';
      });
    } catch (e) {
      if (!mounted) return;
      _handleError(e);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _saveEdit() async {
    final b = _sel!;
    setState(() {
      _editBusy = true;
      _editError = null;
    });
    try {
      await ApiClient.instance.updateBookTranslation(
          b.id, _lang, _titleCtrl.text.trim(), _summaryCtrl.text.trim());
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('翻译已保存')));
      setState(() => _doneLangs.add(_lang));
    } catch (e) {
      if (!mounted) return;
      setState(() => _editError = ApiClient.instance.errorMessage(e));
    } finally {
      if (mounted) setState(() => _editBusy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(child: _buildList()),
        VerticalDivider(thickness: 1, width: 1),
        SizedBox(width: 380, child: _buildPanel()),
      ],
    );
  }

  Widget _buildList() {
    return Column(
      children: [
        const Padding(
          padding: EdgeInsets.all(8),
          child: Align(
              alignment: Alignment.centerLeft,
              child: Text('选择书籍后翻译到目标语言；bn（孟加拉语）DeepL 不支持。')),
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
                          DataColumn(label: Text('原语言')),
                          DataColumn(label: Text('操作')),
                        ],
                        rows: [
                          for (final b in _list)
                            DataRow(
                                selected: _sel?.id == b.id,
                                onSelectChanged: (_) => _select(b),
                                cells: [
                                  DataCell(Text(b.title)),
                                  DataCell(Text(b.author)),
                                  DataCell(Text(b.lang)),
                                  DataCell(FilledButton.tonal(
                                      onPressed: () => _select(b),
                                      child: const Text('翻译'))),
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

  Widget _buildPanel() {
    final b = _sel;
    if (b == null) {
      return const Center(child: Text('左侧选择书籍'));
    }
    return SingleChildScrollView(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('《${b.title}》', style: Theme.of(context).textTheme.titleMedium),
          Text('作者：${b.author}    原语言：${b.lang}'),
          const SizedBox(height: 8),
          Text('已翻译语言：${_doneLangs.isEmpty ? '（无）' : _doneLangs.join('、')}'),
          const SizedBox(height: 12),
          DropdownButton<String>(
            value: _lang,
            items: [
              for (final l in _langs)
                DropdownMenuItem(
                  value: l,
                  enabled: l != 'bn',
                  child: Text(
                      l == 'bn' ? '${_langNames[l]}（不支持）' : '${_langNames[l]} · $l'),
                ),
            ],
            onChanged: (v) {
              if (v != null) setState(() => _lang = v);
            },
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              FilledButton.icon(
                  onPressed: _busy ? null : _translateBook,
                  icon: const Icon(Icons.translate),
                  label: const Text('翻译标题与简介')),
              const SizedBox(width: 8),
              FilledButton.tonalIcon(
                  onPressed: _busy ? null : _translateChapters,
                  icon: const Icon(Icons.article_outlined),
                  label: const Text('翻译全部章节')),
            ],
          ),
          if (_busy) const Padding(
              padding: EdgeInsets.all(12),
              child: Center(child: CircularProgressIndicator())),
          if (_result != null)
            Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: Text(_result!,
                    style: TextStyle(color: Theme.of(context).colorScheme.primary))),
          const Divider(),
          const Text('人工编辑（保存到该语言翻译）'),
          TextField(
              controller: _titleCtrl,
              decoration: const InputDecoration(labelText: '标题')),
          TextField(
              controller: _summaryCtrl,
              maxLines: 4,
              decoration: const InputDecoration(labelText: '简介')),
          if (_editError != null)
            Text(_editError!,
                style: TextStyle(color: Theme.of(context).colorScheme.error)),
          const SizedBox(height: 8),
          FilledButton(
              onPressed: _editBusy ? null : _saveEdit,
              child: _editBusy
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('保存编辑')),
        ],
      ),
    );
  }
}
