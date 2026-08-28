// 全局配置。
// 当前工具链（hvigor 5.19.2 / build-profile.json5）无 buildConfig 注入机制，BASE_URL 统一在此处修改：
// - HarmonyOS 模拟器：改 http://10.0.2.2:8000（需宿主防火墙放行 8000）
// - 真机调试：改宿主机局域网 IP，如 http://192.168.1.100:8000
// - 发布：改线上 https 域名。注意发布签名（unsigned hap）无法使用网络，真机需自行配置签名。
export class Config {
  static readonly BASE_URL: string = 'http://localhost:8000';
}
