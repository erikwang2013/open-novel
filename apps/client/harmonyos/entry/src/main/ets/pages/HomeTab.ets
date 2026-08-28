import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { RecommendItem } from '../common/models/Models';

/// 首页 tab：推荐列表（GET /api/recommend?strategy=hot）。
@Component
export struct HomeTab {
  @State items: RecommendItem[] = [];
  @State loading: boolean = true;
  @State errorMsg: string = '';

  aboutToAppear(): void {
    this.load();
  }

  private async load(): Promise<void> {
    this.loading = true;
    this.errorMsg = '';
    try {
      this.items = await ApiClient.instance.recommend();
    } catch (e) {
      this.errorMsg = (e as Error).message;
    }
    this.loading = false;
  }

  private openBook(item: RecommendItem): void {
    router.pushUrl({
      url: 'pages/BookDetailPage',
      params: {
        'bookId': item.bookId,
        'title': item.title,
        'author': item.author,
        'summary': item.summary
      }
    });
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  @Builder
  errorView() {
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
  }

  @Builder
  bookRow(title: string, author: string, summary: string) {
    Column() {
      Text(title)
        .fontSize(16)
        .fontWeight(FontWeight.Medium)
        .maxLines(1)
        .textOverflow({ overflow: TextOverflow.Ellipsis })
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
      Text($r('app.string.recommend'))
        .fontSize(20)
        .fontWeight(FontWeight.Bold)
        .margin({ top: 12, bottom: 4 })

      if (this.loading) {
        Column() {
          LoadingProgress().width(50).height(50)
        }
        .layoutWeight(1)
        .width('100%')
        .justifyContent(FlexAlign.Center)
      } else if (this.errorMsg !== '') {
        this.errorView()
      } else if (this.items.length === 0) {
        Column() {
          Text($r('app.string.empty'))
            .fontSize(14)
            .fontColor('#999999')
        }
        .layoutWeight(1)
        .width('100%')
        .justifyContent(FlexAlign.Center)
      } else {
        List({ space: 8 }) {
          ForEach(this.items, (item: RecommendItem) => {
            ListItem() {
              this.bookRow(item.title, item.author, item.summary)
            }
            .onClick(() => {
              this.openBook(item);
            })
          }, (item: RecommendItem) => item.bookId)
        }
        .layoutWeight(1)
        .padding({ left: 12, right: 12, top: 8, bottom: 8 })
      }
    }
    .width('100%')
    .height('100%')
  }
}
