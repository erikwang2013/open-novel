import { router } from '@kit.ArkUI';
import { ApiClient } from '../common/api/ApiClient';

/// 登录 / 注册页（POST /api/users/login、/api/users/register）。
@Entry
@Component
struct LoginPage {
  @State username: string = '';
  @State password: string = '';
  @State email: string = '';
  @State nickname: string = '';
  @State isRegister: boolean = false;
  @State busy: boolean = false;
  @State errorMsg: string = '';

  private async submit(): Promise<void> {
    this.busy = true;
    this.errorMsg = '';
    try {
      if (this.isRegister) {
        const nick = this.nickname.trim() === '' ? this.username.trim() : this.nickname.trim();
        await ApiClient.instance.register(this.username.trim(), this.password, this.email.trim(), nick);
      } else {
        await ApiClient.instance.login(this.username.trim(), this.password);
      }
      router.back();
    } catch (e) {
      this.errorMsg = (e as Error).message;
    }
    this.busy = false;
  }

  build() {
    Column() {
      Text(this.isRegister ? $r('app.string.register') : $r('app.string.login'))
        .fontSize(22)
        .fontWeight(FontWeight.Bold)
        .margin({ top: 32, bottom: 24 })

      TextInput({ placeholder: $r('app.string.username'), text: this.username })
        .onChange((v: string) => {
          this.username = v;
        })
        .margin({ left: 24, right: 24, bottom: 12 })

      TextInput({ placeholder: $r('app.string.password'), text: this.password })
        .type(InputType.Password)
        .onChange((v: string) => {
          this.password = v;
        })
        .margin({ left: 24, right: 24, bottom: 12 })

      if (this.isRegister) {
        TextInput({ placeholder: $r('app.string.email'), text: this.email })
          .onChange((v: string) => {
            this.email = v;
          })
          .margin({ left: 24, right: 24, bottom: 12 })

        TextInput({ placeholder: $r('app.string.nickname'), text: this.nickname })
          .onChange((v: string) => {
            this.nickname = v;
          })
          .margin({ left: 24, right: 24, bottom: 12 })
      }

      if (this.errorMsg !== '') {
        Text(this.errorMsg === 'network' ? $r('app.string.errorNetwork') : this.errorMsg)
          .fontSize(14)
          .fontColor('#E84026')
          .margin({ left: 24, right: 24, bottom: 8 })
      }

      Button(this.isRegister ? $r('app.string.register') : $r('app.string.login'))
        .width('80%')
        .margin({ top: 8 })
        .enabled(!this.busy)
        .onClick(() => {
          this.submit();
        })

      Button(this.isRegister ? $r('app.string.login') : $r('app.string.register'))
        .width('80%')
        .backgroundColor('#F0F0F0')
        .fontColor('#333333')
        .enabled(!this.busy)
        .onClick(() => {
          this.isRegister = !this.isRegister;
          this.errorMsg = '';
        })
    }
    .width('100%')
    .height('100%')
  }
}
