import { router, promptAction } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { PaymentMethod, Plan } from '../common/models/Models';

/// VIP 购买页（T-P-14）：公开套餐列表 → 支付方式提示 → 下单 → 系统浏览器跳转 → 结果页轮询。
@Entry
@Component
struct VipPage {
  @State plans: Plan[] = [];
  @State methods: PaymentMethod[] = [];
  @State loading: boolean = true;
  @State errorMsg: string = '';
  @State ordering: boolean = false;

  aboutToAppear(): void {
    this.load();
  }

  private async load(): Promise<void> {
    this.loading = true;
    this.errorMsg = '';
    try {
      const plans = await ApiClient.instance.listPublicPlans();
      const methods = await ApiClient.instance.listMethods(ApiClient.LANG);
      this.plans = plans;
      this.methods = methods;
    } catch (e) {
      this.errorMsg = (e as Error).message;
    }
    this.loading = false;
  }

  /// 金额分 → 元展示，currency 标出。
  private price(p: Plan): string {
    return (p.amountCents / 100).toFixed(2) + ' ' + p.currency;
  }

  private async buy(p: Plan): Promise<void> {
    if (this.ordering) {
      return;
    }
    this.ordering = true;
    try {
      const o = await ApiClient.instance.createOrder(p.planCode, ApiClient.LANG);
      const ok = await ApiClient.instance.openCheckoutUrl(o.checkoutUrl);
      if (!ok) {
        this.toast(this.t($r('app.string.errorServer'), 'launch failed'));
        this.ordering = false;
        return;
      }
      if (o.orderNo !== '') {
        await router.pushUrl({ url: 'pages/PaymentResultPage', params: { 'orderNo': o.orderNo } });
        this.load(); // 回跳后刷新（可能已开通）
      }
    } catch (e) {
      this.toast((e as Error).message);
    }
    this.ordering = false;
  }

  private toast(msg: string): void {
    promptAction.showToast({ message: msg });
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  build() {
    Column() {
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
          Button($r('app.string.retry')).onClick(() => {
            this.load();
          })
        }
        .layoutWeight(1)
        .width('100%')
        .justifyContent(FlexAlign.Center)
      } else {
        List() {
          ForEach(this.plans, (p: Plan) => {
            ListItem() {
              Column({ space: 6 }) {
                Row() {
                  Text(p.label !== '' ? p.label : p.planCode)
                    .fontSize(16)
                    .fontWeight(FontWeight.Bold)
                    .layoutWeight(1)
                  Button(this.t($r('app.string.payNow')))
                    .fontSize(14)
                    .height(32)
                    .enabled(!this.ordering)
                    .onClick(() => {
                      this.buy(p);
                    })
                }
                Row() {
                  Text(`${p.days} days · ${this.price(p)}`)
                    .fontSize(13)
                    .fontColor('#888888')
                }
              }
              .width('100%')
              .padding(14)
              .backgroundColor('#F5F5F5')
              .borderRadius(8)
            }
          }, (p: Plan) => p.planCode)
          ListItem() {
            Text(this.methods.length === 0
              ? this.t($r('app.string.paymentNoMethod'))
              : this.methods.map((m: PaymentMethod) => m.code).join(' / '))
              .fontSize(12)
              .fontColor('#999999')
              .margin({ top: 8 })
          }
        }
        .width('100%')
        .layoutWeight(1)
      }
    }
    .width('100%')
    .height('100%')
    .padding({ left: 16, right: 16, top: 16 })
  }
}
