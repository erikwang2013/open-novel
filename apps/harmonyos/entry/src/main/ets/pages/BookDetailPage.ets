import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { Book, Chapter, asStr } from '../common/models/Models';

/// 书籍详情页：书籍信息 + 章节列表（GET /api/books/{id}、/api/books/{id}/chapters）。
@Entry
@Component
struct BookDetailPage {
  @State book: Book | null = null;
  @State chapters: Chapter[] = [];
  @State loading: boolean = true;
  @State errorMsg: string = '';
  @State title: string = '';
  @State author: string = '';
  @State summary: string = '';
  private bookId: string = '';

  aboutToAppear(): void {
    const p = router.getParams() as Record<string, Object>;
    this.bookId = asStr(p['bookId']);
    this.title = asStr(p['title']);
    this.author = asStr(p['author']);
    this.summary = asStr(p['summary']);
    this.load();
  }

  private async load(): Promise<void> {
    this.loading = true;
    this.errorMsg = '';
    try {
      this.book = await ApiClient.instance.getBook(this.bookId);
      this.chapters = await ApiClient.instance.listChapters(this.bookId);
    } catch (e) {
      this.errorMsg = (e as Error).message;
    }
    this.loading = false;
  }

  private openReader(index: number): void {
    router.pushUrl({
      url: 'pages/ReaderPage',
      params: {
        'bookId': this.bookId,
        'chapters': JSON.stringify(this.chapters),
        'index': index
      }
    });
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  build() {
    Column() {
      Text(this.title !== '' ? this.title : $r('app.string.bookDetail'))
        .fontSize(18)
        .fontWeight(FontWeight.Bold)
        .maxLines(1)
        .textOverflow({ overflow: TextOverflow.Ellipsis })
        .padding({ left: 16, right: 16, top: 12, bottom: 4 })

      if (this.loading) {
        LoadingProgress().width(50).height(50)
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
      } else if (this.book !== null) {
        Column() {
          Text(this.book.title)
            .fontSize(20)
            .fontWeight(FontWeight.Bold)
          Text(this.book.author)
            .fontSize(14)
            .fontColor('#888888')
            .margin({ top: 6 })
          Text(this.book.summary)
            .fontSize(14)
            .fontColor('#666666')
            .margin({ top: 8 })
            .maxLines(4)
            .textOverflow({ overflow: TextOverflow.Ellipsis })
          Text(this.t($r('app.string.bookCount'), `${this.chapters.length}`))
            .fontSize(13)
            .fontColor('#999999')
            .margin({ top: 12 })
        }
        .alignItems(HorizontalAlign.Start)
        .width('100%')
        .padding(16)

        List() {
          ForEach(this.chapters, (c: Chapter, idx: number) => {
            ListItem() {
              Row() {
                Text(`${c.chapterNo}`)
                  .fontSize(14)
                  .fontColor('#999999')
                  .width(40)
                Text(c.title)
                  .fontSize(15)
                  .maxLines(1)
                  .textOverflow({ overflow: TextOverflow.Ellipsis })
                  .layoutWeight(1)
                if (c.isVip === 1) {
                  Text($r('app.string.vip'))
                    .fontSize(12)
                    .fontColor('#E84026')
                } else {
                  Text($r('app.string.free'))
                    .fontSize(12)
                    .fontColor('#999999')
                }
              }
              .width('100%')
              .padding({ left: 16, right: 16, top: 12, bottom: 12 })
            }
            .onClick(() => {
              this.openReader(idx);
            })
          }, (c: Chapter) => c.id)
        }
        .layoutWeight(1)
      }
    }
    .width('100%')
    .height('100%')
  }
}
