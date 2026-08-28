// 与 kratos backend proto JSON 对齐的模型。
// 注意：后端 int64 字段序列化为 JSON 字符串，统一用 asStr/asInt 兼容 string/num。

export function asStr(v: Object | undefined | null): string {
  if (v === undefined || v === null) {
    return '';
  }
  if (typeof v === 'string') {
    return v as string;
  }
  return v.toString();
}

export function asInt(v: Object | undefined | null, fallback: number = 0): number {
  if (v === undefined || v === null) {
    return fallback;
  }
  if (typeof v === 'number') {
    return v as number;
  }
  const n = parseInt(v.toString(), 10);
  return Number.isNaN(n) ? fallback : n;
}

export class Book {
  id: string = '';
  lang: string = '';
  title: string = '';
  author: string = '';
  summary: string = '';
  cover: string = '';
  isVip: number = 0;
  status: number = 0;

  static fromJson(j: Record<string, Object>): Book {
    const b = new Book();
    b.id = asStr(j['id']);
    b.lang = asStr(j['lang']);
    b.title = asStr(j['title']);
    b.author = asStr(j['author']);
    b.summary = asStr(j['summary']);
    b.cover = asStr(j['cover']);
    b.isVip = asInt(j['isVip']);
    b.status = asInt(j['status']);
    return b;
  }
}

export class Chapter {
  id: string = '';
  bookId: string = '';
  chapterNo: number = 0;
  title: string = '';
  wordCount: number = 0;
  isVip: number = 0;
  status: number = 0;

  static fromJson(j: Record<string, Object>): Chapter {
    const c = new Chapter();
    c.id = asStr(j['id']);
    c.bookId = asStr(j['bookId']);
    c.chapterNo = asInt(j['chapterNo']);
    c.title = asStr(j['title']);
    c.wordCount = asInt(j['wordCount']);
    c.isVip = asInt(j['isVip']);
    c.status = asInt(j['status']);
    return c;
  }
}

export class ChapterContent {
  id: string = '';
  chapterId: string = '';
  lang: string = '';
  content: string = '';

  static fromJson(j: Record<string, Object>): ChapterContent {
    const c = new ChapterContent();
    c.id = asStr(j['id']);
    c.chapterId = asStr(j['chapterId']);
    c.lang = asStr(j['lang']);
    c.content = asStr(j['content']);
    return c;
  }
}

export class User {
  id: string = '';
  username: string = '';
  nickname: string = '';
  role: number = 0;

  static fromJson(j: Record<string, Object>): User {
    const u = new User();
    u.id = asStr(j['id']);
    u.username = asStr(j['username']);
    u.nickname = asStr(j['nickname']);
    u.role = asInt(j['role']);
    return u;
  }
}

export class LoginResult {
  accessToken: string = '';
  refreshToken: string = '';
  user: User = new User();

  static fromJson(j: Record<string, Object>): LoginResult {
    const r = new LoginResult();
    r.accessToken = asStr(j['accessToken']);
    r.refreshToken = asStr(j['refreshToken']);
    const u = j['user'];
    if (u !== undefined) {
      r.user = User.fromJson(u as Record<string, Object>);
    }
    return r;
  }
}

export class RecommendItem {
  bookId: string = '';
  title: string = '';
  author: string = '';
  summary: string = '';

  static fromJson(j: Record<string, Object>): RecommendItem {
    const r = new RecommendItem();
    r.bookId = asStr(j['bookId']);
    r.title = asStr(j['title']);
    r.author = asStr(j['author']);
    r.summary = asStr(j['summary']);
    return r;
  }
}

export class BookshelfItem {
  id: string = '';
  userId: string = '';
  bookId: string = '';
  sortOrder: number = 0;

  static fromJson(j: Record<string, Object>): BookshelfItem {
    const b = new BookshelfItem();
    b.id = asStr(j['id']);
    b.userId = asStr(j['userId']);
    b.bookId = asStr(j['bookId']);
    b.sortOrder = asInt(j['sortOrder']);
    return b;
  }
}

export class FavoriteItem {
  id: string = '';
  userId: string = '';
  bookId: string = '';

  static fromJson(j: Record<string, Object>): FavoriteItem {
    const f = new FavoriteItem();
    f.id = asStr(j['id']);
    f.userId = asStr(j['userId']);
    f.bookId = asStr(j['bookId']);
    return f;
  }
}

export class Progress {
  id: string = '';
  userId: string = '';
  bookId: string = '';
  chapterId: string = '';
  position: number = 0;
  updatedAt: string = '';

  static fromJson(j: Record<string, Object>): Progress {
    const p = new Progress();
    p.id = asStr(j['id']);
    p.userId = asStr(j['userId']);
    p.bookId = asStr(j['bookId']);
    p.chapterId = asStr(j['chapterId']);
    p.position = asInt(j['position']);
    p.updatedAt = asStr(j['updatedAt']);
    return p;
  }
}

export class Comment {
  id: string = '';
  content: string = '';
  likeCount: number = 0;
  createdAt: string = '';

  static fromJson(j: Record<string, Object>): Comment {
    const c = new Comment();
    c.id = asStr(j['id']);
    c.content = asStr(j['content']);
    c.likeCount = asInt(j['likeCount']);
    c.createdAt = asStr(j['createdAt']);
    return c;
  }
}

/// VIP 套餐（GET /api/payments/plans 公开）。amount 为分。
export class Plan {
  planCode: string = '';
  days: number = 0;
  amountCents: number = 0;
  currency: string = '';
  label: string = '';

  static fromJson(j: Record<string, Object>): Plan {
    const p = new Plan();
    p.planCode = asStr(j['planCode']);
    p.days = asInt(j['days']);
    p.amountCents = asInt(j['amount']);
    p.currency = asStr(j['currency']);
    p.label = asStr(j['label']);
    return p;
  }
}

/// 创建订单结果（POST /api/payments/order）。
export class CreateOrderResult {
  orderNo: string = '';
  amount: number = 0;
  currency: string = '';
  checkoutUrl: string = '';
  provider: string = '';

  static fromJson(j: Record<string, Object>): CreateOrderResult {
    const o = new CreateOrderResult();
    o.orderNo = asStr(j['orderNo']);
    o.amount = asInt(j['amount']);
    o.currency = asStr(j['currency']);
    o.checkoutUrl = asStr(j['checkoutUrl']);
    o.provider = asStr(j['provider']);
    return o;
  }
}

/// 支付方式（GET /api/payments/methods）。
export class PaymentMethod {
  code: string = '';
  lang: string = '';
  region: string = '';

  static fromJson(j: Record<string, Object>): PaymentMethod {
    const m = new PaymentMethod();
    m.code = asStr(j['code']);
    m.lang = asStr(j['lang']);
    m.region = asStr(j['region']);
    return m;
  }
}

/// 订单状态（GET /api/payments/order/{order_no}）。status 0待支付 1已付 2失败 3关闭。
export class OrderStatus {
  orderNo: string = '';
  status: number = 0;
  amount: number = 0;
  currency: string = '';
  provider: string = '';
  txId: string = '';
  paidAt: string = '';

  static fromJson(j: Record<string, Object>): OrderStatus {
    const o = new OrderStatus();
    o.orderNo = asStr(j['orderNo']);
    o.status = asInt(j['status']);
    o.amount = asInt(j['amount']);
    o.currency = asStr(j['currency']);
    o.provider = asStr(j['provider']);
    o.txId = asStr(j['txId']);
    o.paidAt = asStr(j['paidAt']);
    return o;
  }
}

/// VIP 状态（GET /api/payments/vip-status）。
export class VipStatus {
  active: boolean = false;
  vipExpiresAt: string = '';

  static fromJson(j: Record<string, Object>): VipStatus {
    const v = new VipStatus();
    v.active = j['active'] === true;
    v.vipExpiresAt = asStr(j['vipExpiresAt']);
    return v;
  }
}

/// OpenSearch 搜索结果文档：按 lang 取对应语言字段（同 Flutter 实现）。
export class SearchDoc {
  bookId: string = '';
  lang: string = '';
  titleZh: string = '';
  titleEn: string = '';
  titleJa: string = '';
  summaryZh: string = '';
  summaryEn: string = '';
  summaryJa: string = '';
  authorZh: string = '';
  authorEn: string = '';
  authorJa: string = '';

  static fromJson(j: Record<string, Object>): SearchDoc {
    const d = new SearchDoc();
    d.bookId = asStr(j['bookId']);
    d.lang = asStr(j['lang']);
    d.titleZh = asStr(j['titleZh']);
    d.titleEn = asStr(j['titleEn']);
    d.titleJa = asStr(j['titleJa']);
    d.summaryZh = asStr(j['summaryZh']);
    d.summaryEn = asStr(j['summaryEn']);
    d.summaryJa = asStr(j['summaryJa']);
    d.authorZh = asStr(j['authorZh']);
    d.authorEn = asStr(j['authorEn']);
    d.authorJa = asStr(j['authorJa']);
    return d;
  }

  title(lang: string): string {
    if (lang === 'en') return this.titleEn !== '' ? this.titleEn : this.titleZh;
    if (lang === 'ja') return this.titleJa !== '' ? this.titleJa : this.titleZh;
    return this.titleZh;
  }

  summary(lang: string): string {
    if (lang === 'en') return this.summaryEn !== '' ? this.summaryEn : this.summaryZh;
    if (lang === 'ja') return this.summaryJa !== '' ? this.summaryJa : this.summaryZh;
    return this.summaryZh;
  }

  author(lang: string): string {
    if (lang === 'en') return this.authorEn !== '' ? this.authorEn : this.authorZh;
    if (lang === 'ja') return this.authorJa !== '' ? this.authorJa : this.authorZh;
    return this.authorZh;
  }
}
