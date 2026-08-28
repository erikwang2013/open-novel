// API 客户端：@ohos.net.http 封装，Bearer JWT，401 自动刷新（/api/users/refresh）并重放一次。
// 错误消息优先取后端 {code,message,detail} 的 message，否则返回 'network'（网络错误）或 'server'（服务异常）。
// 注意：HttpResponse 每请求需 createHttp + destroy，不可复用。
// 注意：accessToken/refreshToken 明文存于 preferences。@ohos.security.asset 可做加密存储，
// 但改动面大（异步迁移全部读写），暂不采用，如有安全要求再升级。
import http from '@ohos.net.http';
import preferences from '@ohos.data.preferences';
import { common } from '@kit.AbilityKit';
import { Config } from '../Config';
import {
  Book,
  Chapter,
  ChapterContent,
  LoginResult,
  RecommendItem,
  SearchDoc,
  User,
  asStr
} from '../models/Models';

interface HttpResult {
  code: number;
  data: string;
  message: string;
  networkError: boolean;
}

export class ApiClient {
  static readonly instance: ApiClient = new ApiClient();
  static readonly BASE_URL: string = Config.BASE_URL;
  static readonly PREFS_NAME: string = 'open_novel';
  static LANG: string = 'zh';

  accessToken: string = '';
  refreshToken: string = '';
  currentUser: User | null = null;
  private context: common.Context | null = null;

  /** 由 EntryAbility 启动时调用：恢复持久化登录态。API lang v1 固定 zh。 */
  init(context: common.Context): void {
    this.context = context;
    ApiClient.LANG = 'zh';
    const prefs = preferences.getPreferencesSync(context, { name: ApiClient.PREFS_NAME });
    this.accessToken = prefs.getSync('access_token', '') as string;
    this.refreshToken = prefs.getSync('refresh_token', '') as string;
    const userJson = prefs.getSync('user_json', '') as string;
    if (userJson !== '') {
      try {
        this.currentUser = User.fromJson(JSON.parse(userJson) as Record<string, Object>);
      } catch (e) {
        this.currentUser = null;
      }
    }
  }

  async login(username: string, password: string): Promise<LoginResult> {
    const body: Record<string, Object> = { 'username': username, 'password': password };
    const obj = await this.post<Record<string, Object>>('/api/users/login', body);
    return await this.saveSession(LoginResult.fromJson(obj));
  }

  async register(username: string, password: string, email: string, nickname: string): Promise<LoginResult> {
    const body: Record<string, Object> = {
      'username': username,
      'password': password,
      'email': email,
      'nickname': nickname
    };
    const obj = await this.post<Record<string, Object>>('/api/users/register', body);
    return await this.saveSession(LoginResult.fromJson(obj));
  }

  async listBooks(page: number = 1, pageSize: number = 20): Promise<Book[]> {
    const obj = await this.get<Record<string, Array<Record<string, Object>> | undefined>>(
      `/api/books?page=${page}&page_size=${pageSize}&lang=${ApiClient.LANG}`);
    return this.parseList(obj, Book.fromJson);
  }

  async getBook(id: string): Promise<Book> {
    const obj = await this.get<Record<string, Object>>(`/api/books/${id}?lang=${ApiClient.LANG}`);
    return Book.fromJson(obj);
  }

  async recommend(pageSize: number = 10): Promise<RecommendItem[]> {
    const obj = await this.get<Record<string, Array<Record<string, Object>> | undefined>>(
      `/api/recommend?strategy=hot&page_size=${pageSize}&lang=${ApiClient.LANG}`);
    return this.parseList(obj, RecommendItem.fromJson);
  }

  async search(q: string, page: number = 1, pageSize: number = 20): Promise<SearchDoc[]> {
    const obj = await this.get<Record<string, Array<Record<string, Object>> | undefined>>(
      `/api/search?q=${encodeURIComponent(q)}&page=${page}&page_size=${pageSize}&lang=${ApiClient.LANG}`);
    return this.parseList(obj, SearchDoc.fromJson);
  }

  async listChapters(bookId: string): Promise<Chapter[]> {
    const obj = await this.get<Record<string, Array<Record<string, Object>> | undefined>>(
      `/api/books/${bookId}/chapters?page=1&page_size=100&lang=${ApiClient.LANG}`);
    return this.parseList(obj, Chapter.fromJson);
  }

  async getChapterContent(chapterId: string): Promise<ChapterContent> {
    const obj = await this.get<Record<string, Object>>(`/api/chapters/${chapterId}/content?lang=${ApiClient.LANG}`);
    return ChapterContent.fromJson(obj);
  }

  async logout(): Promise<void> {
    this.accessToken = '';
    this.refreshToken = '';
    this.currentUser = null;
    if (this.context !== null) {
      const prefs = preferences.getPreferencesSync(this.context, { name: ApiClient.PREFS_NAME });
      prefs.deleteSync('access_token');
      prefs.deleteSync('refresh_token');
      prefs.deleteSync('user_json');
      await prefs.flush();
    }
  }

  private async get<T>(path: string): Promise<T> {
    return this.request<T>(http.RequestMethod.GET, path, '', 1);
  }

  private async post<T>(path: string, body: Record<string, Object>): Promise<T> {
    return this.request<T>(http.RequestMethod.POST, path, JSON.stringify(body), 1);
  }

  /** 发起请求；401 且有 refresh token 时刷新一次并重放（retried 控制只重试一次）。 */
  private async request<T>(method: http.RequestMethod, path: string, body: string, retried: number): Promise<T> {
    const resp = await this.doRequest(method, path, body);
    if (resp.networkError) {
      throw new Error('network');
    }
    if (resp.code === 401 && this.refreshToken !== '' && retried > 0) {
      if (await this.doRefresh()) {
        return this.request<T>(method, path, body, 0);
      }
      await this.logout();
    }
    if (resp.code < 200 || resp.code >= 300) {
      throw new Error(resp.message);
    }
    try {
      return JSON.parse(resp.data) as T;
    } catch (e) {
      // 保留状态码与原始响应体，便于定位后端返回非 JSON（网关错误页等）
      const raw = resp.data.length > 200 ? `${resp.data.substring(0, 200)}...` : resp.data;
      throw new Error(`server(${resp.code}): ${raw}`);
    }
  }

  private async doRequest(method: http.RequestMethod, path: string, body: string): Promise<HttpResult> {
    const client = http.createHttp();
    const header: Record<string, string> = { 'Content-Type': 'application/json', 'X-Api-Version': 'v1' };
    if (this.accessToken !== '') {
      header['Authorization'] = `Bearer ${this.accessToken}`;
    }
    const options: http.HttpRequestOptions = {
      method: method,
      header: header,
      connectTimeout: 10000,
      readTimeout: 15000
    };
    if (body !== '') {
      options.extraData = body;
    }
    const result: HttpResult = { code: 0, data: '', message: 'network', networkError: true };
    try {
      const resp = await client.request(`${ApiClient.BASE_URL}${path}`, options);
      result.code = resp.responseCode;
      result.data = `${resp.result}`;
      result.networkError = false;
      result.message = this.extractMessage(resp.responseCode, result.data);
      return result;
    } catch (e) {
      return result;
    } finally {
      client.destroy();
    }
  }

  private async doRefresh(): Promise<boolean> {
    const client = http.createHttp();
    try {
      const header: Record<string, string> = { 'Content-Type': 'application/json', 'X-Api-Version': 'v1' };
      const body: Record<string, Object> = { 'refresh_token': this.refreshToken };
      const resp = await client.request(`${ApiClient.BASE_URL}/api/users/refresh`, {
        method: http.RequestMethod.POST,
        header: header,
        connectTimeout: 10000,
        readTimeout: 15000,
        extraData: JSON.stringify(body)
      });
      if (resp.responseCode === 200) {
        const obj = JSON.parse(`${resp.result}`) as Record<string, Object>;
        const r = LoginResult.fromJson(obj);
        if (r.accessToken === '' || r.refreshToken === '') {
          return false;
        }
        await this.saveSession(r);
        return true;
      }
      return false;
    } catch (e) {
      return false;
    } finally {
      client.destroy();
    }
  }

  private async saveSession(r: LoginResult): Promise<LoginResult> {
    this.accessToken = r.accessToken;
    this.refreshToken = r.refreshToken;
    this.currentUser = r.user;
    await this.persist();
    return r;
  }

  private async persist(): Promise<void> {
    if (this.context === null) {
      return;
    }
    const prefs = preferences.getPreferencesSync(this.context, { name: ApiClient.PREFS_NAME });
    prefs.putSync('access_token', this.accessToken);
    prefs.putSync('refresh_token', this.refreshToken);
    prefs.putSync('user_json', JSON.stringify(this.currentUser));
    await prefs.flush();
  }

  private extractMessage(code: number, data: string): string {
    if (code >= 200 && code < 300) {
      return '';
    }
    try {
      const obj = JSON.parse(data) as Record<string, Object>;
      const m = asStr(obj['message']);
      if (m !== '') {
        return m;
      }
    } catch (e) {
      // 非 JSON 响应体，走统一 server 错误
    }
    return 'server';
  }

  private parseList<T>(obj: Record<string, Array<Record<string, Object>> | undefined>,
    fromJson: (j: Record<string, Object>) => T): T[] {
    const result: T[] = [];
    const list = obj['list'];
    if (list === undefined || list === null) {
      return result;
    }
    for (let i = 0; i < list.length; i++) {
      result.push(fromJson(list[i]));
    }
    return result;
  }
}
