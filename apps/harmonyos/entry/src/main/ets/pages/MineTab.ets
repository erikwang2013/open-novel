import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { User } from '../common/models/Models';

/// 我的 tab：登录态展示 / 登录入口 / 退出（onPageShow 时刷新，登录页返回后自动更新）。
@Component
export struct MineTab {
  @State user: User | null = null;

  aboutToAppear(): void {
    this.refresh();
  }

  onPageShow(): void {
    this.refresh();
  }

  private refresh(): void {
    this.user = ApiClient.instance.currentUser;
  }

  private logout(): void {
    ApiClient.instance.logout();
    this.refresh();
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  build() {
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
    .height('100%')
  }
}
