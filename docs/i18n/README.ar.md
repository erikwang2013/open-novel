# Open Novel — منصة روايات عالمية متعددة اللغات

<div align="center">

[中文](../../README.md) · [English](README.en.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Русский](README.ru.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · **العربية** · [বাংলা](README.bn.md) · [Bahasa Indonesia](README.id.md)

</div>

> منصة عالمية لقراءة الروايات متعددة اللغات، مبنية على معمارية الخدمات المصغرة **Go-Kratos** مع واجهات أمامية متعددة المنصات **Flutter / HarmonyOS**، تدعم **أكثر من 12 لغة رئيسية**، وتوفر لمستخدمي العالم أجمع قدرات القراءة والتفاعل والبحث والتوصيات الشخصية.

---

## مقدمة المشروع

Open Novel منصة عالمية للروايات متعددة اللغات بمعمارية خدمات مصغرة سحابية:

- **الواجهة الخلفية**: Go-Kratos v2 (بروتوكول مزدوج gRPC / HTTP)، خدمات مصغرة مقسمة حسب المجالات (المستخدمون، الكتب، الفصول، التعليقات، البحث، التوصيات)
- **الواجهة الأمامية**: Flutter لجميع المنصات (Web / Desktop / Mobile) + تطبيق HarmonyOS NEXT الأصلي، تشترك جميعها في نفس واجهات API الخلفية
- **تعدد اللغات**: تحميل ديناميكي لموارد i18n، يدعم أكثر من 12 لغة (الصينية، الإنجليزية، اليابانية، الكورية، الفرنسية، الألمانية، الإسبانية، الروسية، العربية وغيرها)
- **التخزين**: MySQL 8 (فصل القراءة والكتابة بين الرئيسي والتابع) + Redis (الذاكرة المؤقتة السريعة / الجلسات) + OpenSearch (بحث متعدد اللغات)
- **التشغيل**: نشر بضغطة واحدة عبر Docker Compose، مراقبة عبر Prometheus + Grafana، تكامل مستمر عبر GitHub Actions


## الميزات

<p align="center"><img src="images/ar/features.svg" alt="مخطط معمارية الميزات" width="860"/></p>

- **مركز المستخدم**: التسجيل وتسجيل الدخول (JWT)، رف الكتب الشخصي، مزامنة تقدم القراءة عبر الأجهزة، ملف شخصي متعدد اللغات
- **تجربة القراءة**: قراءة فصلًا بفصل، تبديل الخط وحجمه، وضعان فاتح وداكن، تخزين مؤقت دون اتصال، حركات قلب الصفحات
- **محتوى الكتب**: بيانات وصفية للكتب، إدارة الفصول، تصنيفات ووسوم، تحديثات متسلسلة، ترجمة متعددة اللغات
- **المجتمع التفاعلي**: تعليقات ومراجعات، إعجابات، مفضلة، بلاغات ومراجعة المحتوى
- **البحث والاكتشاف**: بحث بتقطيع النصوص متعدد اللغات، قوائم الأكثر شهرة، توصيات AI، تصفح حسب التصنيف
- **لوحة الإدارة**: مراجعة المحتوى، إدارة المستخدمين، إحصائيات البيانات، إدارة الإعدادات

## معمارية النظام

<p align="center"><img src="images/ar/architecture.svg" alt="مخطط معمارية النظام" width="860"/></p>

## نظرة عامة على المشروع

<p align="center"><img src="images/ar/project.svg" alt="مخطط نظرة عامة على المشروع" width="860"/></p>

## دورة الطلب

<p align="center"><img src="images/ar/request-cycle.svg" alt="مخطط دورة الطلب" width="860"/></p>

## المعمارية الأمنية

<p align="center"><img src="images/ar/security.svg" alt="مخطط المعمارية الأمنية" width="860"/></p>

## هيكل المشروع

<p align="center"><img src="images/ar/structure.svg" alt="مخطط هيكل المشروع" width="860"/></p>

---

## التقنيات المستخدمة

| الطبقة | التقنية |
| :--- | :--- |
| العميل | Flutter (Web / Desktop / Mobile)، HarmonyOS NEXT (ArkTS / ArkUI) |
| البوابة | Nginx + CDN، بوابة API الخاصة بـ Go-Kratos (بروتوكول مزدوج gRPC / HTTP) |
| الخادم | Go 1.22+، Kratos v2، protobuf / gRPC |
| التخزين | MySQL 8.0 (رئيسي/تابع)، Redis 7.x (Cluster)، OpenSearch 2.x |
| المراقبة | Prometheus، Grafana، ELK، تتبع الروابط OpenTelemetry |
| التشغيل | Docker Compose، GitHub Actions CI/CD |

## قاعدة البيانات

- اسم قاعدة البيانات: `novel`
- بادئة الجداول: `novel_` (مثل `novel_user` و `novel_book` و `novel_chapter` و `novel_comment` وغيرها)

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

تفاصيل تصميم الجداول واستراتيجية فصل القراءة والكتابة موجودة في [docs/novel-project-planning.md](../novel-project-planning.md).

## مجلدات التطبيقات المتعددة

```
apps/
├─ flutter/     # Flutter لجميع المنصات (Web / Desktop / Mobile)، تعدد لغات i18n
└─ harmonyos/   # تطبيق HarmonyOS NEXT الأصلي (ArkTS / ArkUI)
```

انظر [apps/README.md](../../apps/README.md) لمزيد من التفاصيل.

## خارطة الطريق

| المرحلة | المدة | محور المهام |
| :--- | :--- | :--- |
| Phase 1 | 2-3 أسابيع | خدمات Kratos الأساسية + تكامل MySQL / Redis / OpenSearch |
| Phase 2 | 3-4 أسابيع | واجهات Flutter / HarmonyOS متعددة المنصات + كتابة ملفات ARB متعددة اللغات |
| Phase 3 | أسبوعان | تقوية الأمان (JWT / RBAC / تحديد المعدل) + اختبارات الضغط |
| Phase 4 | 1-2 أسبوع | اختبار التكامل الشامل عبر المسار الكامل + إعداد تسريع CDN |
| Phase 5 | مستمر | دمج خوارزمية التوصيات AI، تتبع تحليلات سلوك المستخدم |

## التطوير المحلي

```bash
# تشغيل التبعيات (MySQL / Redis / OpenSearch)
docker compose up -d

# الخدمات الخلفية (مساحة عمل Kratos)
cd kratos/backend && go mod tidy && go run ./cmd/server

# طرف Flutter
cd apps/flutter && flutter pub get && flutter run

# طرف HarmonyOS
cd apps/harmonyos && hvigorw assembleHap
```

---

## الدعم والتبرعات

إذا كان هذا المشروع مفيدًا لك، فلا تتردد في دعمه عبر **Star** و **Fork**؛ كما نرحب بالتبرع عبر مسح رمز الاستجابة السريعة. كل دعم منك هو حافز لي للاستمرار في الصيانة والتحديث، شكرًا لتشجيعك!

<div align="center">

**تبرع WeChat** ｜ **تبرع Alipay**

<img src="../weixinpay.png" width="130" height="130" alt="رمز تبرع WeChat" />　<img src="../alipay.png" width="130" height="130" alt="رمز تبرع Alipay" />

</div>

### تبرع عبر التحويل البنكي العالمي (حوالة عبر الحدود)

【معلومات المستفيد】

- اسم المستفيد: WANG KEXUN
- رقم حساب المستفيد: 881015918251

【البنك المستلم】

- رمز ZA Bank SWIFT: AABLHKHHXXX
- اسم البنك: ZA Bank Limited
- رقم البنك: 387
- عنوان البنك: Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【البنك الوكيل للتحويل عبر الحدود (عند الحاجة)】

> يرجى الانتباه: هذه معلومات البنك الوكيل (البنك الوسيط) للتحويل عبر الحدود، وليست معلومات البنك المستلم. يرجى الاستفسار من البنك المُحوِّل عما إذا كان يلزم تقديم معلومات البنك الوكيل للتحويل عبر الحدود.

**البنك الوكيل لاستلام دولار هونغ كونغ واليوان والدولار الأمريكي هو Citibank**

- اسم البنك: Citibank N.A. Hong Kong
- رمز SWIFT: CITIHKHXXXX
- رقم البنك: 006
- اسم الفرع: Hong Kong Branch
- رقم الفرع: 391
- عنوان البنك: Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**البنك الوكيل للعملات الأخرى هو BNY Mellon**

- اسم البنك: THE BANK OF NEW YORK MELLON
- رمز SWIFT: IRVTUS3NXXX
- عنوان البنك: THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States
