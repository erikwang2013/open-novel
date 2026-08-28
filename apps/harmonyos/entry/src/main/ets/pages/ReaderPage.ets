import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { Chapter, asInt, asStr } from '../common/models/Models';

/// 阅读器：章节正文 + 上一章 / 下一章（GET /api/chapters/{id}/content）。
/// 章节列表由 ReaderPage 按 bookId 自行拉取（GET /api/books/{id}/chapters），
/// BookDetailPage 仅传 bookId 与初始序号，避免 router 传大 JSON 串。
@Entry
@Component
struct ReaderPage {
  @State chapters: Chapter[] = [];
  @State index: number = 0;
  @State content: string = '';
  @State loading: boolean = true;
  @State errorMsg: string = '';
  private bookId: string = '';
  /// 请求序号：上一/下一章连点时只应用最后一次请求的结果，慢响应不覆盖新章节。
  private loadSeq: number = 0;

  aboutToAppear(): void {
    const p = router.getParams() as Record<string, Object>;
    this.bookId = asStr(p['bookId']);
    this.index = asInt(p['index']);
    this.load();
  }

  private async load(): Promise<void> {
    const seq = ++this.loadSeq;
    this.loading = true;
    this.errorMsg = '';
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
          this.index = 0;
        }
      }
      const c = await ApiClient.instance.getChapterContent(this.chapters[this.index].id);
      if (seq !== this.loadSeq) {
        return;
      }
      this.content = c.content;
    } catch (e) {
      if (seq !== this.loadSeq) {
        return;
      }
      this.errorMsg = (e as Error).message;
    }
    this.loading = false;
  }

  private goto(delta: number): void {
    const n = this.index + delta;
    if (n < 0 || n >= this.chapters.length) {
      return;
    }
    this.index = n;
    this.content = '';
    this.load();
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  build() {
    Column() {
      if (this.chapters.length > 0) {
        Text(`${this.chapters[this.index].chapterNo}. ${this.chapters[this.index].title}`)
          .fontSize(17)
          .fontWeight(FontWeight.Bold)
          .maxLines(1)
          .textOverflow({ overflow: TextOverflow.Ellipsis })
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
        Scroll() {
          Column() {
            Text(this.content)
              .fontSize(17)
              .lineHeight(30)
              .fontColor('#333333')
          }
          .width('100%')
          .alignItems(HorizontalAlign.Start)
        }
        .layoutWeight(1)
        .padding({ left: 20, right: 20, top: 8, bottom: 8 })
        .scrollBar(BarState.Off)
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
  }
}
