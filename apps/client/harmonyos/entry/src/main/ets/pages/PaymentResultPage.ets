import { router } from '@kit.ArkUI';
import { common } from '@kit.AbilityKit';
import { ApiClient } from '../common/api/ApiClient';
import { asStr } from '../common/models/Models';

/// 支付结果页（T-P-15）：每 3 秒轮询 GetOrder。
/// status 0 待确认 / 1 已付 / 2 失败或关闭；终态停止轮询。
@Entry
@Component
struct PaymentResultPage {
  @State status: number = 0; // 0 待确认 1 已付 2 失败/关闭
  @State errorMsg: string = '';
  private orderNo: string = '';
  private timer: number = -1;

  aboutToAppear(): void {
    const p = router.getParams() as Record<string, Object>;
    this.orderNo = asStr(p['orderNo']);
    this.poll();
    this.timer = setInterval(() => {
      this.poll();
    }, 3000);
  }

  aboutToDisappear(): void {
    if (this.timer >= 0) {
      clearInterval(this.timer);
      this.timer = -1;
    }
  }

  private async poll(): Promise<void> {
    try {
      const o = await ApiClient.instance.getOrder(this.orderNo);
      this.status = o.status === 1 ? 1 : (o.status === 0 ? 0 : 2);
      if (o.status !== 0 && this.timer >= 0) {
        clearInterval(this.timer); // 终态停止轮询
        this.timer = -1;
      }
    } catch (e) {
      this.errorMsg = (e as Error).message;
    }
  }

  /// 失败/关闭：返回购买页重试。
  private retry(): void {
    router.back();
  }

  private t(res: Resource, ...args: string[]): string {
    return (getContext(this) as common.UIAbilityContext).resourceManager.getStringSync(res, ...args);
  }

  build() {
    Column({ space: 16 }) {
      if (this.status === 1) {
        Text('✓')
          .fontSize(72)
          .fontColor('#4CAF50')
      } else if (this.status === 2) {
        Text('✕')
          .fontSize(72)
          .fontColor('#F44336')
      } else {
        LoadingProgress().width(72).height(72)
      }
      Text(this.status === 1 ? this.t($r('app.string.paymentSuccess'))
        : this.status === 2 ? this.t($r('app.string.paymentFailed'))
          : this.t($r('app.string.paymentPending')))
        .fontSize(20)
        .fontWeight(FontWeight.Bold)
      if (this.status === 0) {
        Text($r('app.string.paymentChecking'))
          .fontSize(13)
          .fontColor('#999999')
      }
      if (this.errorMsg !== '') {
        Text(this.t($r('app.string.errorServer'), this.errorMsg))
          .fontSize(12)
          .fontColor('#999999')
      }
      if (this.status === 2) {
        Button($r('app.string.retryPay')).onClick(() => {
          this.retry();
        })
      }
    }
    .width('100%')
    .height('100%')
    .justifyContent(FlexAlign.Center)
  }
}
