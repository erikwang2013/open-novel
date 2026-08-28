import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { Book, SearchDoc } from '../common/models/Models';

/// 全部书籍 tab：搜索框 + 书籍列表（GET /api/books 分页、GET /api/search）。
@Component
export struct BooksTab {
  @State books: Book[] = [];
  @State results: SearchDoc[] = [];
  @State searching: boolean = false;
  @State loading: boolean = true;
  @State errorMsg: string = '';
  @State keyword: string = '';
  @State page: number = 1;
  @State hasMore: boolean = true;

  aboutToAppear(): void {
    this.loadBooks();
  }

  private async loadBooks(): Promise<void> {
    this.loading = true;
    this.errorMsg = '';
    this.searching = false;
    this.results = [];
    this.page = 1;
    this.hasMore = true;
    try {
      this.books = await ApiClient.instance.listBooks(1, 20);
    } catch (e) {
      this.errorMsg = (e as Error).message;
    }
    this.loading = false;
  }

  private async loadMore(): Promise<void> {
    if (!this.hasMore || this.loading || this.searching) {
      return;
    }
    try {
      const more = await ApiClient.instance.listBooks(this.page + 1, 20);
      if (more.length === 0) {
        this.hasMore = false;
      } else {
        this.page = this.page + 1;
        this.books = this.books.concat(more);
      }
    } catch (e) {
      // 加载更多失败静默，下次滑到底再试
    }
  }

  private async doSearch(q: string): Promise<void> {
    if (q === '') {
      this.loadBooks();
      return;
    }
    this.searching = true;
    this.errorMsg = '';
    this.keyword = q;
    try {
      this.results = await ApiClient.instance.search(q);
    } catch (e) {
      this.errorMsg = (e as Error).message;
    }
    this.searching = false;
  }

  private openBook(id: string, title: string, author: string, summary: string): void {
    router.pushUrl({
      url: 'pages/BookDetailPage',
      params: {
        'bookId': id,
        'title': title,
        'author': author,
        'summary': summary
      }
    });
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  @Builder
  bookRow(title: string, author: string, summary: string, isVip: boolean) {
    Column() {
      Row() {
        Text(title)
          .fontSize(16)
          .fontWeight(FontWeight.Medium)
          .maxLines(1)
          .textOverflow({ overflow: TextOverflow.Ellipsis })
          .layoutWeight(1)
        if (isVip) {
          Text($r('app.string.vip'))
            .fontSize(11)
            .fontColor('#E84026')
            .margin({ left: 8 })
        }
      }
      .width('100%')
      Text(author)
        .fontSize(13)
        .fontColor('#888888')
        .maxLines(1)
        .textOverflow({ overflow: TextOverflow.Ellipsis })
        .margin({ top: 4 })
      Text(summary)
        .fontSize(13)
        .fontColor('#666666')
        .maxLines(2)
        .textOverflow({ overflow: TextOverflow.Ellipsis })
        .margin({ top: 4 })
    }
    .alignItems(HorizontalAlign.Start)
    .width('100%')
    .padding(12)
    .backgroundColor('#F5F5F5')
    .borderRadius(8)
  }

  build() {
    Column() {
      Row({ space: 8 }) {
        TextInput({ placeholder: $r('app.string.searchHint'), text: this.keyword })
          .onChange((v: string) => {
            this.keyword = v;
          })
          .onSubmit(() => {
            this.doSearch(this.keyword);
          })
          .layoutWeight(1)
        Button($r('app.string.searchResult'))
          .onClick(() => {
            this.doSearch(this.keyword);
          })
      }
      .padding(12)

      if (this.searching) {
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
              if (this.keyword === '') {
                this.loadBooks();
              } else {
                this.doSearch(this.keyword);
              }
            })
        }
      } else if (this.results.length > 0) {
        List({ space: 8 }) {
          ForEach(this.results, (d: SearchDoc) => {
            ListItem() {
              this.bookRow(d.title(ApiClient.LANG), d.author(ApiClient.LANG), d.summary(ApiClient.LANG), false)
            }
            .onClick(() => {
              this.openBook(d.bookId, d.title(ApiClient.LANG), d.author(ApiClient.LANG), d.summary(ApiClient.LANG));
            })
          }, (d: SearchDoc) => d.bookId)
        }
        .layoutWeight(1)
        .padding({ left: 12, right: 12, top: 8, bottom: 8 })
      } else if (this.keyword !== '') {
        Text($r('app.string.empty'))
          .fontSize(14)
          .fontColor('#999999')
      } else if (this.loading) {
        LoadingProgress().width(50).height(50)
      } else if (this.books.length === 0) {
        Text($r('app.string.empty'))
          .fontSize(14)
          .fontColor('#999999')
      } else {
        List({ space: 8 }) {
          ForEach(this.books, (b: Book) => {
            ListItem() {
              this.bookRow(b.title, b.author, b.summary, b.isVip === 1)
            }
            .onClick(() => {
              this.openBook(b.id, b.title, b.author, b.summary);
            })
          }, (b: Book) => b.id)
        }
        .layoutWeight(1)
        .padding({ left: 12, right: 12, top: 8, bottom: 8 })
        .onReachEnd(() => {
          this.loadMore();
        })
      }
    }
    .width('100%')
    .height('100%')
  }
}
