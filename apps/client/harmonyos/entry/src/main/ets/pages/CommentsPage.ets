import { router, promptAction } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { Comment, asStr } from '../common/models/Models';

/// 评论区：列表 + 发布 + 点赞（GET/POST /api/comments，POST /api/comments/{id}/like）。
/// 未登录发布跳登录页；onPageShow 在从登录页返回时刷新列表。
@Entry
@Component
struct CommentsPage {
  @State comments: Comment[] = [];
  @State loading: boolean = true;
  @State errorMsg: string = '';
  @State content: string = '';
  @State chapterTitle: string = '';
  private bookId: string = '';
  private chapterId: string = '';
  /// 请求序号：连点发布/点赞时只应用最后一次请求的结果。
  private loadSeq: number = 0;

  aboutToAppear(): void {
    const p = router.getParams() as Record<string, Object>;
    this.bookId = asStr(p['bookId']);
    this.chapterId = asStr(p['chapterId']);
    this.loadChapterTitle();
  }

  onPageShow(): void {
    this.load();
  }

  private async load(): Promise<void> {
    const seq = ++this.loadSeq;
    this.loading = true;
    this.errorMsg = '';
    try {
      const list = await ApiClient.instance.listComments(this.bookId, 1, 20, this.chapterId);
      if (seq !== this.loadSeq) {
        return;
      }
      this.comments = list;
    } catch (e) {
      if (seq !== this.loadSeq) {
        return;
      }
      this.errorMsg = (e as Error).message;
    }
    this.loading = false;
  }

  /// 章节号+标题（如有）：解析失败不影响评论区。
  private async loadChapterTitle(): Promise<void> {
    if (this.chapterId === '') {
      return;
    }
    try {
      const list = await ApiClient.instance.listChapters(this.bookId);
      const c = list.find((x) => x.id === this.chapterId);
      if (c !== undefined) {
        this.chapterTitle = `${c.chapterNo}. ${c.title}`;
      }
    } catch (e) {
      // 忽略
    }
  }

  private async post(): Promise<void> {
    const text = this.content.trim();
    if (text === '') {
      return;
    }
    if (ApiClient.instance.accessToken === '') {
      router.pushUrl({ url: 'pages/LoginPage' });
      return;
    }
    try {
      await ApiClient.instance.postComment(this.bookId, text, this.chapterId);
      this.content = '';
      this.load();
    } catch (e) {
      promptAction.showToast({ message: (e as Error).message });
    }
  }

  private like(c: Comment): void {
    const req = c.likeCount > 0
      ? ApiClient.instance.unlikeComment(c.id)
      : ApiClient.instance.likeComment(c.id);
    req.then(() => {
      this.load();
    }).catch(() => {
      // 点赞/取消点赞失败静默（含重复点赞业务错误 150409）
    });
  }

  private report(c: Comment): void {
    if (ApiClient.instance.accessToken === '') {
      router.pushUrl({ url: 'pages/LoginPage' });
      return;
    }
    AlertDialog.show({
      title: $r('app.string.report'),
      message: this.t($r('app.string.reportConfirm')),
      primaryButton: {
        value: $r('app.string.cancel'),
        action: () => {
        }
      },
      secondaryButton: {
        value: $r('app.string.confirm'),
        action: () => {
          ApiClient.instance.reportComment(c.id).then(() => {
            promptAction.showToast({ message: this.t($r('app.string.reportSuccess')) });
          }).catch((e: Error) => {
            promptAction.showToast({ message: e.message });
          });
        }
      }
    });
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  build() {
    Column() {
      Text($r('app.string.comments'))
        .fontSize(18)
        .fontWeight(FontWeight.Bold)
        .padding({ left: 16, right: 16, top: 12, bottom: 4 })
      if (this.chapterTitle !== '') {
        Text(this.chapterTitle)
          .fontSize(13)
          .fontColor('#888888')
          .padding({ left: 16, right: 16, bottom: 8 })
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
      } else if (this.comments.length === 0) {
        Column() {
          Text($r('app.string.empty'))
            .fontSize(14)
            .fontColor('#999999')
        }
        .layoutWeight(1)
        .width('100%')
        .justifyContent(FlexAlign.Center)
      } else {
        Scroll() {
          Column() {
            ForEach(this.comments, (c: Comment) => {
              Row() {
                Column() {
                  Text(c.content)
                    .fontSize(15)
                    .fontColor('#333333')
                  Text(c.createdAt)
                    .fontSize(12)
                    .fontColor('#999999')
                    .margin({ top: 4 })
                }
                .alignItems(HorizontalAlign.Start)
                .layoutWeight(1)

                Row({ space: 12 }) {
                  Row({ space: 2 }) {
                    SymbolGlyph(c.likeCount > 0 ? $r('sys.symbol.hand_thumbsup_fill') : $r('sys.symbol.hand_thumbsup'))
                      .fontSize(18)
                      .fontColor([c.likeCount > 0 ? '#E84026' : '#999999'])
                    Text(`${c.likeCount}`)
                      .fontSize(12)
                      .fontColor(c.likeCount > 0 ? '#E84026' : '#999999')
                  }
                  .onClick(() => {
                    this.like(c);
                  })

                  SymbolGlyph($r('sys.symbol.flag'))
                    .fontSize(16)
                    .fontColor(['#999999'])
                    .onClick(() => {
                      this.report(c);
                    })
                }
              }
              .width('100%')
              .padding({ left: 16, right: 16, top: 12, bottom: 12 })
            }, (c: Comment) => c.id)
          }
          .width('100%')
        }
        .layoutWeight(1)
        .scrollBar(BarState.Off)
      }

      Row({ space: 8 }) {
        TextInput({ placeholder: this.t($r('app.string.commentHint')) })
          .layoutWeight(1)
          .height(40)
          .backgroundColor('#F5F5F5')
          .borderRadius(8)
          .onChange((v: string) => {
            this.content = v;
          })
        Button($r('app.string.post'))
          .onClick(() => {
            this.post();
          })
      }
      .padding({ left: 12, right: 12, top: 8, bottom: 12 })
    }
    .width('100%')
    .height('100%')
  }
}
