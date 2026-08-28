import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { Chapter, asInt, asStr } from '../common/models/Models';

/// 阅读器：章节正文 + 上一章 / 下一章（GET /api/chapters/{id}/content）。
/// 章节列表与当前 index 由 BookDetailPage 经 router 参数传入（chapters 为 JSON 字符串）。
@Entry
@Component
struct ReaderPage {
  @State chapters: Chapter[] = [];
  @State index: number = 0;
  @State content: string = '';
  @State loading: boolean = true;
  @State errorMsg: string = '';

  aboutToAppear(): void {
    const p = router.getParams() as Record<string, Object>;
    this.index = asInt(p['index']);
    const chaptersStr = asStr(p['chapters']);
    if (chaptersStr !== '') {
      const raw = JSON.parse(chaptersStr) as Array<Record<string, Object>>;
      this.chapters = [];
      for (let i = 0; i < raw.length; i++) {
        this.chapters.push(Chapter.fromJson(raw[i]));
      }
    }
    if (this.index >= this.chapters.length) {
      this.index = 0;
    }
    this.load();
  }

  private async load(): Promise<void> {
    if (this.chapters.length === 0) {
      return;
    }
    this.loading = true;
    this.errorMsg = '';
    try {
      const c = await ApiClient.instance.getChapterContent(this.chapters[this.index].id);
      this.content = c.content;
    } catch (e) {
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
