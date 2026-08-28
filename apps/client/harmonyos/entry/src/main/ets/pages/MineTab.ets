import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { Book, Chapter, Progress, User } from '../common/models/Models';

/// 我的 tab：登录态展示 / 登录入口 / 退出（onPageShow 时刷新，登录页返回后自动更新）。
/// 已登录时展示书架、收藏、最近进度三个区块；各接口失败均静默，不阻塞页面。
@Component
export struct MineTab {
  @State user: User | null = null;
  @State shelfRows: ShelfRow[] = [];
  @State favRows: FavRow[] = [];
  @State progressRows: ProgressRow[] = [];

  aboutToAppear(): void {
    this.refresh();
  }

  onPageShow(): void {
    this.refresh();
  }

  private refresh(): void {
    this.user = ApiClient.instance.currentUser;
    this.shelfRows = [];
    this.favRows = [];
    this.progressRows = [];
    if (this.user === null) {
      return;
    }
    this.loadSections();
  }

  private async loadSections(): Promise<void> {
    try {
      const shelf = await ApiClient.instance.listBookshelf();
      const favs = await ApiClient.instance.listFavorites();
      const shelfRows: ShelfRow[] = [];
      const favRows: FavRow[] = [];
      const progressRows: ProgressRow[] = [];
      for (const it of shelf) {
        const book = await this.getBookSafe(it.bookId);
        if (book === null) {
          continue;
        }
        shelfRows.push({ bookId: it.bookId, title: book.title });
        const pr = await this.loadProgress(it.bookId, book.title);
        if (pr !== null) {
          progressRows.push(pr);
        }
      }
      for (const it of favs) {
        const book = await this.getBookSafe(it.bookId);
        if (book !== null) {
          favRows.push({ bookId: it.bookId, title: book.title });
        }
      }
      this.shelfRows = shelfRows;
      this.favRows = favRows;
      this.progressRows = progressRows;
    } catch (e) {
      // 整体失败（未登录 401 等）静默
    }
  }

  private async getBookSafe(bookId: string): Promise<Book | null> {
    try {
      return await ApiClient.instance.getBook(bookId);
    } catch (e) {
      return null;
    }
  }

  private async loadProgress(bookId: string, title: string): Promise<ProgressRow | null> {
    let p: Progress;
    try {
      p = await ApiClient.instance.getProgress(bookId);
    } catch (e) {
      return null; // 无进度记录
    }
    if (p.chapterId === '') {
      return null;
    }
    const row: ProgressRow = {
      bookId: bookId,
      title: title,
      chapterTitle: '',
      index: 0,
      updatedAt: this.fmtTime(p.updatedAt)
    };
    try {
      const chapters = await ApiClient.instance.listChapters(bookId);
      const idx = chapters.findIndex((c: Chapter) => c.id === p.chapterId);
      if (idx >= 0) {
        row.index = idx;
        row.chapterTitle = chapters[idx].title;
      }
    } catch (e) {
      // 章节列表失败仅显示书名
    }
    return row;
  }

  /// updatedAt 兼容 unix 秒/毫秒/日期字符串三种形态。
  private fmtTime(s: string): string {
    if (s === '') {
      return '';
    }
    const n = Number(s);
    if (Number.isNaN(n)) {
      return s;
    }
    const d = new Date(n < 1e12 ? n * 1000 : n);
    return `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()} ${d.getHours()}:${d.getMinutes()}`;
  }

  private async removeFromShelf(bookId: string): Promise<void> {
    try {
      await ApiClient.instance.removeBookshelf(bookId);
    } catch (e) {
      // 静默
    }
    this.refresh();
  }

  private async removeFavorite(bookId: string): Promise<void> {
    try {
      await ApiClient.instance.removeFavorite(bookId);
    } catch (e) {
      // 静默
    }
    this.refresh();
  }

  private openBook(bookId: string, title: string): void {
    router.pushUrl({
      url: 'pages/BookDetailPage',
      params: { 'bookId': bookId, 'title': title, 'author': '', 'summary': '' }
    });
  }

  private openReader(bookId: string, index: number): void {
    router.pushUrl({
      url: 'pages/ReaderPage',
      params: { 'bookId': bookId, 'index': index }
    });
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  @Builder
  sectionTitle(text: string) {
    Text(text)
      .fontSize(16)
      .fontWeight(FontWeight.Bold)
      .width('100%')
      .margin({ top: 8 })
  }

  @Builder
  shelfRowView(r: ShelfRow) {
    Row({ space: 8 }) {
      Text(r.title)
        .fontSize(15)
        .maxLines(1)
        .textOverflow({ overflow: TextOverflow.Ellipsis })
        .layoutWeight(1)
      Button('移出')
        .fontSize(12)
        .height(28)
        .backgroundColor('#F0F0F0')
        .fontColor('#333333')
        .onClick(() => {
          this.removeFromShelf(r.bookId);
        })
    }
    .width('100%')
    .padding({ left: 12, right: 12, top: 10, bottom: 10 })
    .backgroundColor('#F5F5F5')
    .borderRadius(8)
    .onClick(() => {
      this.openBook(r.bookId, r.title);
    })
  }

  @Builder
  favRowView(r: FavRow) {
    Row({ space: 8 }) {
      Text(r.title)
        .fontSize(15)
        .maxLines(1)
        .textOverflow({ overflow: TextOverflow.Ellipsis })
        .layoutWeight(1)
      Button('取消收藏')
        .fontSize(12)
        .height(28)
        .backgroundColor('#F0F0F0')
        .fontColor('#333333')
        .onClick(() => {
          this.removeFavorite(r.bookId);
        })
    }
    .width('100%')
    .padding({ left: 12, right: 12, top: 10, bottom: 10 })
    .backgroundColor('#F5F5F5')
    .borderRadius(8)
    .onClick(() => {
      this.openBook(r.bookId, r.title);
    })
  }

  @Builder
  progressRowView(r: ProgressRow) {
    Column() {
      Text(r.title)
        .fontSize(15)
        .maxLines(1)
        .textOverflow({ overflow: TextOverflow.Ellipsis })
      Text(r.chapterTitle !== '' ? r.chapterTitle : '')
        .fontSize(13)
        .fontColor('#888888')
        .maxLines(1)
        .textOverflow({ overflow: TextOverflow.Ellipsis })
        .margin({ top: 4 })
      Text(r.updatedAt)
        .fontSize(12)
        .fontColor('#999999')
        .margin({ top: 4 })
    }
    .alignItems(HorizontalAlign.Start)
    .width('100%')
    .padding(12)
    .backgroundColor('#F5F5F5')
    .borderRadius(8)
    .onClick(() => {
      this.openReader(r.bookId, r.index);
    })
  }

  build() {
    Scroll() {
      Column({ space: 16 }) {
        Circle({ width: 72, height: 72 })
          .fill('#E0E0E0')
          .margin({ top: 48 })

        if (this.user !== null) {
          Text(this.t($r('app.string.welcome'),
            this.user.nickname !== '' ? this.user.nickname : this.user.username))
            .fontSize(20)
            .fontWeight(FontWeight.Bold)
          Text(this.user.username)
            .fontSize(14)
            .fontColor('#888888')
          Button($r('app.string.logout'))
            .margin({ top: 12 })
            .backgroundColor('#F0F0F0')
            .fontColor('#333333')
            .onClick(() => {
              this.logout();
            })

          this.sectionTitle('书架')
          if (this.shelfRows.length === 0) {
            Text($r('app.string.empty'))
              .fontSize(13)
              .fontColor('#999999')
          } else {
            ForEach(this.shelfRows, (r: ShelfRow) => {
              this.shelfRowView(r)
            }, (r: ShelfRow) => r.bookId)
          }

          this.sectionTitle('收藏')
          if (this.favRows.length === 0) {
            Text($r('app.string.empty'))
              .fontSize(13)
              .fontColor('#999999')
          } else {
            ForEach(this.favRows, (r: FavRow) => {
              this.favRowView(r)
            }, (r: FavRow) => r.bookId)
          }

          this.sectionTitle('最近进度')
          if (this.progressRows.length === 0) {
            Text($r('app.string.empty'))
              .fontSize(13)
              .fontColor('#999999')
          } else {
            ForEach(this.progressRows, (r: ProgressRow) => {
              this.progressRowView(r)
            }, (r: ProgressRow) => r.bookId)
          }
        } else {
          Text($r('app.string.notLoggedIn'))
            .fontSize(20)
            .fontWeight(FontWeight.Bold)
          Button($r('app.string.login'))
            .margin({ top: 12 })
            .onClick(() => {
              router.pushUrl({ url: 'pages/LoginPage' });
            })
        }
      }
      .width('100%')
      .padding({ left: 16, right: 16 })
    }
    .width('100%')
    .height('100%')
  }

  private async logout(): Promise<void> {
    await ApiClient.instance.logout();
    this.refresh();
  }
}

class ShelfRow {
  bookId: string = '';
  title: string = '';
}

class FavRow {
  bookId: string = '';
  title: string = '';
}

class ProgressRow {
  bookId: string = '';
  title: string = '';
  chapterTitle: string = '';
  index: number = 0;
  updatedAt: string = '';
}
