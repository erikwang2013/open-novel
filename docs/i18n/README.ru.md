# Open Novel — глобальная мультиязычная платформа для чтения романов

<div align="center">

[中文](../../README.md) · [English](README.en.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · **Русский** · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [বাংলা](README.bn.md) · [Bahasa Indonesia](README.id.md)

</div>

> Мультиязычная платформа для чтения романов на микросервисной архитектуре **Go-Kratos** + мультиплатформенных фронтендах **Flutter / HarmonyOS**, поддерживающая **12+ основных языков** и предоставляющая пользователям по всему миру чтение, интерактив, поиск и персонализированные рекомендации.

<div align="center"><img src="../mascot.svg" alt="Талисман Open Novel Novi" width="150"/></div>

---

## Обзор проекта

Open Novel — это глобальная мультиязычная платформа для чтения романов в облачной микросервисной архитектуре:

- **Бэкенд**: Go-Kratos v2 (двойные протоколы gRPC / HTTP), микросервисы разделены по доменам (пользователи, книги, главы, комментарии, поиск, рекомендации)
- **Фронтенд**: Flutter на всех платформах (Web / Desktop / Mobile) + нативное приложение HarmonyOS NEXT, использующие единый набор API бэкенда
- **Мультиязычность**: динамическая загрузка i18n-ресурсов, поддержка 12+ языков (китайский, английский, японский, корейский, французский, немецкий, испанский, русский, арабский и др.)
- **Хранилище**: MySQL 8 (мастер-реплика с разделением чтения и записи) + Redis (кэш горячих данных / сессии) + OpenSearch (мультиязычный поиск)
- **Эксплуатация**: развёртывание одной командой через Docker Compose, мониторинг Prometheus + Grafana, непрерывная интеграция GitHub Actions

## Возможности

<p align="center"><img src="images/ru/features.svg" alt="Схема функциональной архитектуры" width="860"/></p>

- **Личный кабинет**: регистрация и вход (JWT), личная книжная полка, синхронизация прогресса чтения между устройствами, мультиязычный профиль
- **Опыт чтения**: чтение по главам, настройка шрифта и его размера, светлая / тёмная тема, офлайн-кэш, анимация перелистывания страниц
- **Контент книг**: метаданные книг, управление главами, категории и теги, обновление серий, мультиязычные переводы
- **Интерактивное сообщество**: комментарии и рецензии, лайки, избранное, модерация жалоб
- **Поиск и открытия**: мультиязычный поиск по токенам, топ горячих ключевых слов, подсказки поиска (локальная история на клиенте — 20 записей + подсказки с дебаунсом 200 мс), AI-рекомендации, просмотр по категориям
- **Админ-панель**: модерация контента, управление пользователями, статистика (дашборд / DAU / рейтинги / поведенческий анализ /api/stats/behavior), управление конфигурацией (теги категорий), конвейер машинного перевода (DeepL, /api/admin/translate/*, страница «Перевод» в админке + ручное редактирование), запрос журнала аудита (/api/admin/audit-logs), управление CDN-провайдерами (мультивендорная настройка / вкл-выкл / сортировка, мгновенное применение через горячую перезагрузку)
- **Платежи и VIP**: мультиканальные платежи через 11 провайдеров (Stripe, NOWPayments (USDT), Razorpay, KOMOJU, PortOne, Mercado Pago, Xendit, PayPal, Alipay, WeChat Pay Global, UnionPay), подписка и продление VIP-планов, маршрутизация способов оплаты по языку (WeChat Pay Global подключён; китайский WeChat Pay не подключён, требуется статус мерчанта в КНР)

## Системная архитектура

<p align="center"><img src="images/ru/architecture.svg" alt="Схема системной архитектуры" width="860"/></p>

В целом это микросервисная архитектура Go-Kratos: клиенты Flutter / HarmonyOS взаимодействуют с API-шлюзом через Nginx + мультивендорный CDN (Cloudflare / CloudFront — глобальная линия, Aliyun / Tencent Cloud — китайская линия; настраивается в админке, горячая перезагрузка конфигурации применяется мгновенно); шлюз маршрутизирует по доменам к бэкенд-сервисам — пользователи, книги, главы, комментарии, поиск, рекомендации и т. д.; уровень данных — MySQL мастер-реплика (разделение чтения и записи) + кэш Redis + поисковый индекс OpenSearch. Сервисы общаются по gRPC, внешние HTTP-интерфейсы используют единый префикс `/api`.

Остальные схемы: общая схема проекта [docs/project.svg](../../docs/project.svg) · цикл запросов [docs/request-cycle.svg](../../docs/request-cycle.svg) · архитектура безопасности [docs/security.svg](../../docs/security.svg) · структура проекта [docs/structure.svg](../../docs/structure.svg).

## Общая схема проекта

<p align="center"><img src="images/ru/project.svg" alt="Общая схема проекта" width="860"/></p>

## Цикл запросов

<p align="center"><img src="images/ru/request-cycle.svg" alt="Схема цикла запросов" width="860"/></p>

## Архитектура безопасности

<p align="center"><img src="images/ru/security.svg" alt="Схема архитектуры безопасности" width="860"/></p>

---

## Структура каталогов

```
open-novel/
├─ apps/                     # Мультиплатформенные фронтенды
│  ├─ flutter/               #   Flutter на всех платформах (Web / Desktop / Mobile), i18n мультиязычность
│  └─ harmonyos/             #   Нативное приложение HarmonyOS NEXT (ArkTS / ArkUI)
├─ kratos/                   # Исходный код фреймворка Go-Kratos (вышестоящий фреймворк, сохранять как есть, не изменять)
│  └─ backend/               #   Бизнес-бэкенд проекта: точка входа cmd/server + api/ + internal/ + sql/ + opensearch/
├─ docs/                     # Документация проекта (планирование, схемы архитектуры, i18n README, QR-коды для донатов)
├─ scripts/                  # Скрипты сборки и развёртывания (авторелиз post-push.sh, smoke.sh)
├─ docker-compose.yml        # Локальный стек зависимостей: MySQL 8 + Redis 7 + OpenSearch 2
├─ CLAUDE.md                 # Правила совместной работы над проектом
└─ README.md                 # Документация проекта
```

<p align="center"><img src="images/ru/structure.svg" alt="Схема структуры проекта" width="860"/></p>

> Примечание: `kratos/` — это исходный код фреймворка Kratos (со своим README / LICENSE); весь бизнес-код находится в `kratos/backend/`.

## Технологический стек

| Уровень | Технологии |
| :--- | :--- |
| Клиент | Flutter (Web / Desktop / Mobile), HarmonyOS NEXT (ArkTS / ArkUI) |
| Шлюз | Nginx + мультивендорный CDN (Cloudflare / CloudFront / Aliyun / Tencent Cloud), API-шлюз Go-Kratos (двойные протоколы gRPC / HTTP) |
| Серверная часть | Go 1.22+, Kratos v2, protobuf / gRPC |
| Хранилище | MySQL 8.0 (мастер-реплика), Redis 7.x (Cluster), OpenSearch 2.x, внутрипроцессный L1-кэш ristretto поверх Redis (TTL 30 с) |
| Наблюдаемость | Prometheus, Grafana, ELK, трассировка OpenTelemetry |
| Эксплуатация | Docker Compose, CI/CD GitHub Actions |

## База данных

- Имя базы данных: `novel`
- Префикс таблиц: `novel_` (например, `novel_user`, `novel_book`, `novel_chapter`, `novel_comment` и т. д.)

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Скрипт создания таблиц: `kratos/backend/sql/init.sql` (автоматически выполняется при первом запуске Docker Compose). Подробное описание структуры таблиц и стратегии разделения чтения и записи см. в [docs/novel-project-planning.md](../novel-project-planning.md).

## Префикс API

Все HTTP-интерфейсы бэкенда начинаются с `/api`; версия согласуется через заголовок запроса `X-Api-Version: v1` (не указывается в URL). Интерфейсы сгруппированы по доменам:

| Домен | Примеры маршрутов | Определение proto |
| :--- | :--- | :--- |
| Пользователи | `/api/users` и т. д. | `kratos/backend/api/user/v1` |
| Книги | `/api/books`, `/api/books/{id}`, `/api/categories`, `/api/tags` | `kratos/backend/api/book/v1` |
| Главы | `/api/...` | `kratos/backend/api/chapter/v1` |
| Комментарии | `/api/...` | `kratos/backend/api/comment/v1` |
| Поиск | `/api/...` | `kratos/backend/api/search/v1` |
| Рекомендации | `/api/...` | `kratos/backend/api/recommendation/v1` |

Подробные маршруты — в объявлениях `option (google.api.http)` в каждом proto-файле.

## Быстрый старт

```bash
# 1. Запустите стек зависимостей (MySQL / Redis / OpenSearch; при первом запуске автоматически выполняется kratos/backend/sql/init.sql для создания таблиц)
docker compose up -d

# 2. Запустите бэкенд-сервис (бизнес-каталог Kratos, HTTP :8000 / gRPC :9000)
cd kratos/backend && go mod tidy && go run ./cmd/server

# 3. Запустите клиент Flutter (по умолчанию подключается к localhost:8000, дополнительная настройка не требуется)
cd apps/client/flutter && flutter pub get && flutter run -d chrome
```

- Проброс портов стека зависимостей: MySQL `3307`, Redis `6380`, OpenSearch `9200` (порты хоста 3306/6379 заняты локальными сервисами, см. комментарии в docker-compose.yml).
- Адрес бэкенда и секреты настраиваются в `kratos/backend/config/`, поддерживается переопределение через переменные окружения (например, `PORT`, `OPENSEARCH_ADDR`).
- Для подключения Flutter к другому бэкенду: `flutter run -d chrome --dart-define=API_BASE_URL=http://<host>:8000`.

Подробности — в [apps/README.md](../../apps/README.md) и [apps/client/flutter/README.md](../../apps/client/flutter/README.md).

## Процесс релиза

- **Автоматически**: после пуша в `main` GitHub Actions ([.github/workflows/release.yml](../../.github/workflows/release.yml)) автоматически увеличивает patch-версию на основе последнего тега `v*`, создаёт и пушит тег, затем создаёт GitHub Release с инкрементальным changelog; пропускается, если HEAD уже содержит тег версии. Первый релиз начинается с `v1.0.0`.
- **Ручной резервный вариант**: запустите [scripts/post-push.sh](../../scripts/post-push.sh) (требуется авторизованный `gh`): `echo "x y refs/heads/main z" | scripts/post-push.sh`.
- **Вручную**:

  ```bash
  git tag -a v1.0.1 -m "release v1.0.1"
  git push origin v1.0.1
  gh release create v1.0.1 --generate-notes
  ```

## Дорожная карта

| Этап | Период | Основные задачи |
| :--- | :--- | :--- |
| Phase 1 | 2-3 недели | Базовые сервисы бэкенда на Kratos + интеграция MySQL / Redis / OpenSearch |
| Phase 2 | 3-4 недели | Мультиплатформенные фронтенды Flutter / HarmonyOS + написание мультиязычных ARB-файлов |
| Phase 3 | 2 недели | Усиление безопасности (JWT / RBAC / ограничение частоты запросов) + нагрузочное тестирование |
| Phase 4 | 1-2 недели | Сквозная интеграция всех звеньев + настройка ускорения CDN |
| Phase 5 | постоянно | Внедрение AI-алгоритмов рекомендаций, сбор данных аналитики поведения пользователей |

Все цепочки задач завершены.

## Поддержка и благодарность

Если этот проект оказался для вас полезным, поддержите его **Star** и **Fork**; также можно отблагодарить автора донатом по QR-коду. Каждая ваша поддержка — моя мотивация продолжать развитие и обновление проекта. Спасибо за поддержку!

<div align="center">

**Донат WeChat** ｜ **Донат Alipay**

<img src="../weixinpay.png" width="130" height="130" alt="QR-код доната WeChat" />　<img src="../alipay.png" width="130" height="130" alt="QR-код доната Alipay" />

</div>

### Глобальные донаты банковским переводом (международный перевод)

【Информация о получателе】

- Имя получателя: WANG KEXUN
- Номер счёта получателя: 881015918251

【Банк получателя】

- SWIFT Code банка ZA Bank: AABLHKHHXXX
- Название банка: ZA Bank Limited
- Код банка: 387
- Адрес банка: Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【Банк-посредник для международных переводов (при необходимости)】

> Обратите внимание: это информация о банке-посреднике (промежуточном банке) для международных переводов, а не о банке получателя. Уточните в своём банке, требуется ли предоставление информации о банке-посреднике.

**Для переводов в гонконгских долларах, юанях и долларах США банком-посредником является Citibank**

- Название банка: Citibank N.A. Hong Kong
- SWIFT Code: CITIHKHXXXX
- Код банка: 006
- Название отделения: Hong Kong Branch
- Номер отделения: 391
- Адрес банка: Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**Для переводов в других валютах банком-посредником является BNY Mellon**

- Название банка: THE BANK OF NEW YORK MELLON
- SWIFT Code: IRVTUS3NXXX
- Адрес банка: THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States

### Пожертвование в криптовалюте (Crypto Donation)

Если этот проект помог вам, отсканируйте QR-код, чтобы сделать пожертвование, спасибо!

| Сеть (Network) | QR-код (QR Code) | Адрес кошелька (Wallet Address) |
|---|---|---|
| BNB Smart Chain (BEP20) | [<img src="../coin/1.jpg" width="150" alt="BNB Smart Chain (BEP20)">](../coin/1.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Tron (TRC20) | [<img src="../coin/2.jpg" width="150" alt="Tron (TRC20)">](../coin/2.jpg) | `TEdDHWLajt1XvqtPDWmQctdrJaC3pzZZzz` |
| Ethereum (ERC20) | [<img src="../coin/3.jpg" width="150" alt="Ethereum (ERC20)">](../coin/3.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Aptos | [<img src="../coin/4.jpg" width="150" alt="Aptos">](../coin/4.jpg) | `0x836e3780edfc3f7b2372b39e2a1a3a5d7adfaccd96c726f21cfde1b50dd68030` |
| Plasma | [<img src="../coin/5.jpg" width="150" alt="Plasma">](../coin/5.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Polygon POS | [<img src="../coin/6.jpg" width="150" alt="Polygon POS">](../coin/6.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Solana | [<img src="../coin/7.jpg" width="150" alt="Solana">](../coin/7.jpg) | `2hfhboHdmdrYsY25XfQSsEWxq5ip4EQsR7f4AzSRMUyr` |
| The Open Network (TON) | [<img src="../coin/8.jpg" width="150" alt="The Open Network (TON)">](../coin/8.jpg) | `UQB9kFQohzmXUir9QSSZq01iwl9aQZIDdBpNmDklljRtCoGK` |
| Arbitrum One | [<img src="../coin/9.jpg" width="150" alt="Arbitrum One">](../coin/9.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| AVAX C-Chain | [<img src="../coin/10.jpg" width="150" alt="AVAX C-Chain">](../coin/10.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |

---

## Лицензия и контакты

- **Лицензия**: в корне репозитория нет отдельного LICENSE; `kratos/` — это вышестоящий исходный код фреймворка Kratos, который распространяется по его [MIT License](../../kratos/LICENSE). Способ лицензирования бизнес-кода будет объявлен в последующих публикациях проекта.
- **Контакты**: общение через GitHub Issues / PR; донаты — см. раздел «Поддержка и благодарность» выше.
