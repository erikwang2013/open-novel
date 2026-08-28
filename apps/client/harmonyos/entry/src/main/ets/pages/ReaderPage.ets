import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import resourceManager from '@ohos.resourceManager';
import preferences from '@ohos.data.preferences';
import fs from '@ohos.file.fs';
import { ApiClient } from '../common/api/ApiClient';
import { Chapter, asInt, asStr } from '../common/models/Models';

interface CacheEntry {
  n: string;
  t: number;
}

/// 阅读器：章节正文 + 上一章 / 下一章。
/// 设置入口：标题栏「设置」按钮菜单（字号 / 行距 / 主题 / 翻页方式），preferences 持久化。
/// 主题：深浅色资源目录（base=浅、dark=深），跟随系统或手动指定。
/// 离线缓存：沙箱 cacheDir 存 ch_{id}.txt，最多 50 篇，超限删 mtime 最旧。
/// 阅读进度：滚动模式存 y 偏移、翻页模式存页码（复用后端 position uint32）。
@Entry
@Component
struct ReaderPage {
  @State chapters: Chapter[] = [];
  @State index: number = 0;
  @State content: string = '';
  @State loading: boolean = true;
  @State errorMsg: string = '';
  @State fontSize: number = 17;
  @State lineHeight: number = 30;
  @State pageMode: number = 0; // 0 上下滚动 1 左右翻页
  @State themePref: string = 'system';
  @State bgColor: Resource = $r('app.color.reader_bg');
  @State textColor: Resource = $r('app.color.reader_text');
  @State pageList: string[] = [];
  @State currentPage: number = 0;
  @State fromCache: boolean = false;
  private bookId: string = '';
  private loadSeq: number = 0;
  private scroller: Scroller = new Scroller();
  private viewW: number = 0;
  private viewH: number = 0;
  private restorePos: number = -1;
  private pageKey: string = '';

  aboutToAppear(): void {
    const p = router.getParams() as Record<string, Object>;
    this.bookId = asStr(p['bookId']);
    this.index = asInt(p['index']);
    this.loadPrefs();
    this.load();
  }

  /// 页面每次可见时刷新深浅色（跟随系统在切系统深色后生效）。
  onPageShow(): void {
    this.updateTheme();
  }

  private loadPrefs(): void {
    const prefs = preferences.getPreferencesSync(getContext(this) as common.Context, { name: 'reader' });
    this.fontSize = prefs.getSync('fontSize', 17) as number;
    this.lineHeight = prefs.getSync('lineHeight', 30) as number;
    this.pageMode = prefs.getSync('pageMode', 0) as number;
    this.themePref = prefs.getSync('theme', 'system') as string;
    this.updateTheme();
  }

  private savePref(key: string, value: number | string): void {
    const prefs = preferences.getPreferencesSync(getContext(this) as common.Context, { name: 'reader' });
    prefs.putSync(key, value);
    prefs.flush();
  }

  private updateTheme(): void {
    const cm = getContext(this).resourceManager.getConfigurationSync().colorMode;
    const sysDark = cm === resourceManager.ColorMode.DARK;
    if (this.themePref === 'system') {
      // $r('app.color.reader_bg') 由 dark/ 资源目录自动跟随系统深浅色
      this.bgColor = $r('app.color.reader_bg');
      this.textColor = $r('app.color.reader_text');
    } else if (this.themePref === 'dark' || sysDark) {
      this.bgColor = $r('app.color.reader_bg_dark');
      this.textColor = $r('app.color.reader_text_dark');
    } else {
      this.bgColor = $r('app.color.reader_bg_light');
      this.textColor = $r('app.color.reader_text_light');
    }
  }

  private async load(): Promise<void> {
    const seq = ++this.loadSeq;
    this.loading = true;
    this.errorMsg = '';
    this.pageKey = '';
    try {
      if (this.chapters.length === 0) {
        const list = await ApiClient.instance.listChapters(this.bookId);
        if (seq !== this.loadSeq) {
          return;
        }
        if (list.length === 0) {
          this.errorMsg = 'server';
          this.loading = false;
          return;
        }
        this.chapters = list;
        if (this.index >= this.chapters.length) {
          this.index = this.chapters.length - 1; // 越界钳制到末章
        }
      }
      const id = this.chapters[this.index].id;
      try {
        const c = await ApiClient.instance.getChapterContent(id);
        if (seq !== this.loadSeq) {
          return;
        }
        this.content = c.content;
        this.fromCache = false;
        this.cacheContent(id, c.content);
      } catch (e) {
        if (seq !== this.loadSeq) {
          return;
        }
        const cached = this.readCache(id);
        if (cached !== '') {
          this.content = cached;
          this.fromCache = true;
        } else {
          this.errorMsg = (e as Error).message;
          this.loading = false;
          return;
        }
      }
      this.currentPage = 0;
      this.saveProgress(0);
      this.restorePosition();
    } catch (e) {
      if (seq !== this.loadSeq) {
        return;
      }
      this.errorMsg = (e as Error).message;
    }
    this.loading = false;
  }

  /// best-effort 保存进度：滚动模式存 y 偏移，翻页模式存页码。
  private saveProgress(pos?: number): void {
    if (ApiClient.instance.accessToken === '') {
      return;
    }
    const c = this.chapters[this.index];
    if (c === undefined || c.id === '') {
      return;
    }
    const position = pos ?? (this.pageMode === 1
      ? this.currentPage
      : this.scroller.currentOffset().yOffset);
    ApiClient.instance.updateProgress(this.bookId, c.id, position).catch(() => {
      // 忽略
    });
  }

  private async restorePosition(): Promise<void> {
    try {
      const p = await ApiClient.instance.getProgress(this.bookId);
      if (p.chapterId !== this.chapters[this.index].id) {
        return;
      }
      this.restorePos = p.position;
      if (this.pageMode === 1) {
        // 页码直接应用；若页面尚未构建，buildPages 里会消费 restorePos
        if (this.pageList.length > 0) {
          this.currentPage = Math.max(0, Math.min(this.restorePos, this.pageList.length - 1));
          this.restorePos = -1;
        }
      } else if (this.restorePos >= 0) {
        // 兜底：若 onAreaChange 已错过首次布局，延迟滚动一次
        setTimeout(() => {
          if (this.restorePos >= 0) {
            this.scroller.scrollTo({
              xOffset: 0,
              yOffset: this.restorePos,
              animation: false
            });
            this.restorePos = -1;
          }
        }, 100);
      }
    } catch (e) {
      // 忽略
    }
  }

  /// 离线缓存：写 ch_{id}.txt；超 50 篇删最旧（mtime）。
  private cacheContent(id: string, content: string): void {
    try {
      const dir = (getContext(this) as common.UIAbilityContext).cacheDir;
      const names = fs.listFileSync(dir).filter((n: string) => n.startsWith('ch_'));
      if (names.length >= 50) {
        const entries: CacheEntry[] = names.map((n: string) => {
          const e: CacheEntry = { n: n, t: fs.statSync(`${dir}/${n}`).mtime };
          return e;
        });
        entries.sort((a, b) => a.t - b.t);
        fs.unlinkSync(`${dir}/${entries[0].n}`);
      }
      const path = `${dir}/ch_${id}.txt`;
      const f = fs.openSync(path, fs.OpenMode.READ_WRITE | fs.OpenMode.CREATE | fs.OpenMode.TRUNC);
      try {
        fs.writeSync(f.fd, content);
      } finally {
        fs.closeSync(f);
      }
    } catch (e) {
      // 缓存失败不影响在线阅读
    }
  }

  private readCache(id: string): string {
    try {
      const p = `${(getContext(this) as common.UIAbilityContext).cacheDir}/ch_${id}.txt`;
      return fs.readTextSync(p);
    } catch (e) {
      return '';
    }
  }

  private goto(delta: number): void {
    const n = this.index + delta;
    if (n < 0 || n >= this.chapters.length) {
      return;
    }
    this.index = n;
    this.content = '';
    this.restorePos = -1;
    this.load();
  }

  /// 近似分页：CJK 字符按宽 ≈ fontSize 估算每行字数。
  /// ponytail: 未用 Paragraph 精确测量；对混排文本页边界略偏，后续可换 Paragraph API。
  private buildPages(): void {
    const key = `${this.content.length}|${this.fontSize}|${this.lineHeight}|${this.viewW}|${this.viewH}`;
    if (key === this.pageKey || this.content === '' || this.viewW <= 0 || this.viewH <= 0) {
      return;
    }
    this.pageKey = key;
    const charsPerLine = Math.max(1, Math.floor((this.viewW - 40) / this.fontSize));
    const linesPerPage = Math.max(1, Math.floor((this.viewH - 16) / this.lineHeight));
    const pageSize = charsPerLine * linesPerPage;
    const pages: string[] = [];
    for (let i = 0; i < this.content.length; i += pageSize) {
      pages.push(this.content.substring(i, i + pageSize));
    }
    this.pageList = pages;
    if (this.restorePos >= 0) {
      this.currentPage = Math.max(0, Math.min(this.restorePos, Math.max(0, this.pageList.length - 1)));
      this.restorePos = -1;
    } else {
      this.currentPage = Math.min(this.currentPage, Math.max(0, this.pageList.length - 1));
    }
  }

  private adjustFont(delta: number): void {
    this.fontSize = Math.max(13, Math.min(26, this.fontSize + delta));
    this.savePref('fontSize', this.fontSize);
  }

  private adjustLineHeight(delta: number): void {
    this.lineHeight = Math.max(20, Math.min(48, this.lineHeight + delta));
    this.savePref('lineHeight', this.lineHeight);
  }

  private setPageMode(m: number): void {
    this.pageMode = m;
    this.savePref('pageMode', m);
  }

  private setTheme(t: string): void {
    this.themePref = t;
    this.savePref('theme', t);
    this.updateTheme();
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  private buildMenu(): Array<MenuElement> {
    const cur = this.pageMode === 1 ? this.t($r('app.string.pagedMode')) : this.t($r('app.string.scrollMode'));
    const themeLabel = this.themePref === 'system' ? this.t($r('app.string.themeSystem'))
      : this.themePref === 'dark' ? this.t($r('app.string.themeDark')) : this.t($r('app.string.themeLight'));
    return [
      { value: `${this.t($r('app.string.fontSize'))}: ${this.fontSize}`, action: () => {}, enabled: false },
      { value: `${this.t($r('app.string.fontSize'))} -`, action: () => this.adjustFont(-1) },
      { value: `${this.t($r('app.string.fontSize'))} +`, action: () => this.adjustFont(1) },
      { value: `${this.t($r('app.string.lineHeight'))}: ${this.lineHeight}`, action: () => {}, enabled: false },
      { value: `${this.t($r('app.string.lineHeight'))} -`, action: () => this.adjustLineHeight(-2) },
      { value: `${this.t($r('app.string.lineHeight'))} +`, action: () => this.adjustLineHeight(2) },
      { value: `${this.t($r('app.string.pageMode'))}: ${cur}`, action: () => {}, enabled: false },
      { value: this.t($r('app.string.scrollMode')), action: () => this.setPageMode(0) },
      { value: this.t($r('app.string.pagedMode')), action: () => this.setPageMode(1) },
      { value: `${this.t($r('app.string.theme'))}: ${themeLabel}`, action: () => {}, enabled: false },
      { value: this.t($r('app.string.themeSystem')), action: () => this.setTheme('system') },
      { value: this.t($r('app.string.themeLight')), action: () => this.setTheme('light') },
      { value: this.t($r('app.string.themeDark')), action: () => this.setTheme('dark') },
    ];
  }

  build() {
    Column() {
      if (this.chapters.length > 0) {
        Row() {
          Text(`${this.chapters[this.index].chapterNo}. ${this.chapters[this.index].title}`)
            .fontSize(17)
            .fontWeight(FontWeight.Bold)
            .fontColor(this.textColor)
            .maxLines(1)
            .textOverflow({ overflow: TextOverflow.Ellipsis })
            .layoutWeight(1)
          Button(this.t($r('app.string.settings')))
            .fontSize(14)
            .backgroundColor(Color.Transparent)
            .fontColor(this.textColor)
            .bindMenu(this.buildMenu())
        }
        .padding({ left: 16, right: 16, top: 12, bottom: 8 })
      }

      if (this.loading) {
        Column() {
          LoadingProgress().width(50).height(50)
        }
        .layoutWeight(1)
        .width('100%')
        .justifyContent(FlexAlign.Center)
      } else if (this.errorMsg !== '') {
        Column({ space: 12 }) {
          Text(this.errorMsg === 'network'
            ? $r('app.string.errorNetwork')
            : this.t($r('app.string.errorServer'), this.errorMsg))
            .fontSize(14)
            .fontColor('#666666')
          Button($r('app.string.retry'))
            .onClick(() => {
              this.load();
            })
        }
        .layoutWeight(1)
        .width('100%')
        .justifyContent(FlexAlign.Center)
      } else {
        Column() {
          if (this.fromCache) {
            Text($r('app.string.offline'))
              .fontSize(12)
              .fontColor(this.textColor)
              .width('100%')
              .textAlign(TextAlign.Center)
              .padding({ top: 4, bottom: 4 })
          }
          if (this.pageMode === 1) {
            Swiper() {
              ForEach(this.pageList, (pg: string, i: number) => {
                Text(pg)
                  .fontSize(this.fontSize)
                  .lineHeight(this.lineHeight)
                  .fontColor(this.textColor)
                  .width('100%')
                  .height('100%')
                  .padding({ left: 20, right: 20, top: 8, bottom: 8 })
              }, (pg: string, i: number) => `${i}`)
            }
            .layoutWeight(1)
            .width('100%')
            .loop(false)
            .index(this.currentPage)
            .onChange((idx: number) => {
              this.currentPage = idx;
              this.saveProgress(idx);
            })
            .onAreaChange((oldV: Area, newV: Area) => {
              this.viewW = Number(newV.width);
              this.viewH = Number(newV.height);
              this.buildPages();
            })
          } else {
            Scroll(this.scroller) {
              Column() {
                Text(this.content)
                  .fontSize(this.fontSize)
                  .lineHeight(this.lineHeight)
                  .fontColor(this.textColor)
              }
              .width('100%')
              .alignItems(HorizontalAlign.Start)
              .onAreaChange((oldV: Area, newV: Area) => {
                if (this.restorePos > 0) {
                  this.scroller.scrollTo({
                    xOffset: 0,
                    yOffset: this.restorePos,
                    animation: false
                  });
                  this.restorePos = -1;
                }
              })
            }
            .layoutWeight(1)
            .padding({ left: 20, right: 20, top: 8, bottom: 8 })
            .scrollBar(BarState.Off)
            .onScrollEnd(() => {
              this.saveProgress();
            })
          }
        }
        .layoutWeight(1)
        .width('100%')
      }

      Row({ space: 16 }) {
        Button($r('app.string.prevChapter'))
          .layoutWeight(1)
          .enabled(this.index > 0)
          .onClick(() => {
            this.goto(-1);
          })
        Button($r('app.string.nextChapter'))
          .layoutWeight(1)
          .enabled(this.index < this.chapters.length - 1)
          .onClick(() => {
            this.goto(1);
          })
      }
      .padding({ left: 16, right: 16, top: 8, bottom: 12 })
    }
    .width('100%')
    .height('100%')
    .backgroundColor(this.bgColor)
  }
}
