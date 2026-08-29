# Open Novel — globale mehrsprachige Plattform für Online-Romane

<div align="center">

[中文](../../README.md) · [English](README.en.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Русский](README.ru.md) · **Deutsch** · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [বাংলা](README.bn.md) · [Bahasa Indonesia](README.id.md)

</div>

> Eine globale mehrsprachige Plattform für Online-Romane auf Basis der **Go-Kratos**-Mikroservicearchitektur mit **Flutter / HarmonyOS**-Multiplattform-Frontends, die **12+ Hauptsprachen** unterstützt und Nutzern weltweit Lesen, Interaktion, Suche und personalisierte Empfehlungen bietet.

<div align="center"><img src="../mascot.svg" alt="Open-Novel-Maskottchen Novi" width="150"/></div>

---

## Projektübersicht

Open Novel ist eine globale mehrsprachige Roman-Plattform mit cloudnativer Mikroservicearchitektur:

- **Backend**: Go-Kratos v2 (Doppelprotokoll gRPC / HTTP), Mikroservices nach Domänen aufgeteilt (Benutzer, Bücher, Kapitel, Kommentare, Suche, Empfehlungen)
- **Frontend**: Flutter auf allen Plattformen (Web / Desktop / Mobile) + native HarmonyOS-NEXT-App, alle mit derselben Backend-API
- **Mehrsprachigkeit**: dynamisches Laden von i18n-Ressourcen, Unterstützung für 12+ Sprachen (Chinesisch, Englisch, Japanisch, Koreanisch, Französisch, Deutsch, Spanisch, Russisch, Arabisch usw.)
- **Speicherung**: MySQL 8 (Master-Replica mit Lese-/Schreibtrennung) + Redis (Hotspot-Cache / Sessions) + OpenSearch (mehrsprachige Suche)
- **Betrieb**: Ein-Klick-Bereitstellung per Docker Compose, Überwachung mit Prometheus + Grafana, kontinuierliche Integration mit GitHub Actions

## Funktionen

<p align="center"><img src="../features.svg" alt="Diagramm der Funktionsarchitektur" width="860"/></p>

- **Benutzerkonto**: Registrierung/Login (JWT), persönliches Bücherregal, Synchronisierung des Lesefortschritts über Geräte hinweg, mehrsprachiges Profil
- **Leseerlebnis**: kapitelweises Lesen, Schriftart- und Schriftgrößenwechsel, helles/dunkles Design, Offline-Cache, Blätteranimationen
- **Buchinhalte**: Buch-Metadaten, Kapitelverwaltung, Kategorien und Tags, Updates laufender Serien, mehrsprachige Übersetzungen
- **Interaktive Community**: Kommentare und Rezensionen, Likes, Favoriten, Meldung und Moderation
- **Suche & Entdecken**: mehrsprachige Tokensuche, Ranking heißer Suchbegriffe, Suchvorschläge (lokale Client-Historie mit 20 Einträgen + Vorschläge mit 200 ms Debounce), KI-Empfehlungen, Stöbern nach Kategorien
- **Admin-Backend**: Inhaltsmoderation, Benutzerverwaltung, Statistiken (Dashboard / DAU / Ranglisten / Verhaltensanalyse /api/stats/behavior), Konfigurationsverwaltung (Kategorietags), Workflow für maschinelle Übersetzung (DeepL, /api/admin/translate/*, „Übersetzung“-Seite im Admin + manuelle Bearbeitung), Abfrage von Audit-Logs (/api/admin/audit-logs)
- **Zahlungen & VIP**: Mehrkanal-Zahlungen über 11 Anbieter (Stripe, NOWPayments (USDT), Razorpay, KOMOJU, PortOne, Mercado Pago, Xendit, PayPal, Alipay, WeChat Pay Global, UnionPay), VIP-Abo und Verlängerung, sprachbasiertes Zahlungsmethoden-Routing (WeChat Pay Global angebunden; inländisches WeChat Pay nicht angebunden, erfordert chinesische Händlerqualifikation)

## Systemarchitektur

<p align="center"><img src="../architecture.svg" alt="Diagramm der Systemarchitektur" width="860"/></p>

Die Gesamtarchitektur ist eine Go-Kratos-Mikroservice-Architektur: Die Flutter- / HarmonyOS-Clients interagieren über Nginx + CDN mit dem API-Gateway; das Gateway routet domänenbasiert zu Backend-Diensten wie Benutzer, Bücher, Kapitel, Kommentare, Suche und Empfehlungen. Die Datenschicht besteht aus MySQL-Master-Replica (Lese-/Schreibtrennung) + Redis-Cache + OpenSearch-Suchindex. Die Dienste kommunizieren über gRPC; die externen HTTP-Schnittstellen verwenden einheitlich das Präfix `/api`.

Weitere Diagramme: Projektübersicht [../project.svg](../project.svg) · Anfragezyklus [../request-cycle.svg](../request-cycle.svg) · Sicherheitsarchitektur [../security.svg](../security.svg) · Projektstruktur [../structure.svg](../structure.svg).

## Projektübersicht

<p align="center"><img src="images/de/project.svg" alt="Projektübersicht" width="860"/></p>

## Anfragezyklus

<p align="center"><img src="images/de/request-cycle.svg" alt="Anfragezyklus" width="860"/></p>

## Sicherheitsarchitektur

<p align="center"><img src="images/de/security.svg" alt="Sicherheitsarchitektur" width="860"/></p>

## Verzeichnisstruktur

```
open-novel/
├─ apps/                     # Multiplattform-Frontends
│  ├─ flutter/               #   Flutter auf allen Plattformen (Web / Desktop / Mobile), i18n-Mehrsprachigkeit
│  └─ harmonyos/             #   HarmonyOS-NEXT-Native-App (ArkTS / ArkUI)
├─ kratos/                   # Go-Kratos-Framework-Quellcode (Upstream-Framework, unverändert übernommen, nicht ändern)
│  └─ backend/               #   Projekt-Backend: cmd/server-Einstieg + api/ + internal/ + sql/ + opensearch/
├─ docs/                     # Projektdokumentation (Planung, Architekturdiagramme, i18n-READMEs, Spenden-QR-Codes)
├─ scripts/                  # Build- und Deployment-Skripte (post-push.sh für automatische Releases, smoke.sh)
├─ docker-compose.yml        # Lokaler Abhängigkeits-Stack: MySQL 8 + Redis 7 + OpenSearch 2
├─ CLAUDE.md                 # Projekt-Kooperationsrichtlinien
└─ README.md                 # Projektdokumentation
```

<p align="center"><img src="../structure.svg" alt="Diagramm der Projektstruktur" width="860"/></p>

> Hinweis: `kratos/` ist der Quellcode des Kratos-Frameworks (mit eigenem README / LICENSE); der gesamte Geschäftscode befindet sich in `kratos/backend/`.

## Technologie-Stack

| Ebene | Technologien |
| :--- | :--- |
| Client | Flutter (Web / Desktop / Mobile), HarmonyOS NEXT (ArkTS / ArkUI) |
| Gateway | Nginx + CDN, Go-Kratos-API-Gateway (Doppelprotokoll gRPC / HTTP) |
| Server | Go 1.22+, Kratos v2, protobuf / gRPC |
| Speicher | MySQL 8.0 (Master-Replica), Redis 7.x (Cluster), OpenSearch 2.x, In-Prozess-L1-Cache ristretto über Redis (30 s TTL) |
| Observability | Prometheus, Grafana, ELK, OpenTelemetry-Tracing |
| Betrieb | Docker Compose, GitHub Actions CI/CD |

## Datenbank

- Datenbankname: `novel`
- Tabellenpräfix: `novel_` (z. B. `novel_user`, `novel_book`, `novel_chapter`, `novel_comment` usw.)

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Tabellen-Skript: `kratos/backend/sql/init.sql` (wird beim ersten Start von Docker Compose automatisch ausgeführt). Detailliertes Tabellendesign und die Strategie zur Lese-/Schreibtrennung finden Sie in [docs/novel-project-planning.md](../novel-project-planning.md).

## API-Präfix

Die HTTP-Schnittstellen des Backends beginnen einheitlich mit `/api`; die Version wird über den Request-Header `X-Api-Version: v1` ausgehandelt (nicht in der URL). Sie sind nach Domänen gruppiert:

| Domäne | Beispielrouten | proto-Definition |
| :--- | :--- | :--- |
| Benutzer | `/api/users` usw. | `kratos/backend/api/user/v1` |
| Bücher | `/api/books`、`/api/books/{id}`、`/api/categories`、`/api/tags` | `kratos/backend/api/book/v1` |
| Kapitel | `/api/...` | `kratos/backend/api/chapter/v1` |
| Kommentare | `/api/...` | `kratos/backend/api/comment/v1` |
| Suche | `/api/...` | `kratos/backend/api/search/v1` |
| Empfehlungen | `/api/...` | `kratos/backend/api/recommendation/v1` |

Detaillierte Routen finden Sie in den Deklarationen `option (google.api.http)` der jeweiligen proto-Dateien.

## Schnellstart

```bash
# 1. Abhängigkeits-Stack starten (MySQL / Redis / OpenSearch; führt beim ersten Start automatisch kratos/backend/sql/init.sql aus)
docker compose up -d

# 2. Backend-Dienst starten (Kratos-Geschäftsverzeichnis, HTTP :8000 / gRPC :9000)
cd kratos/backend && go mod tidy && go run ./cmd/server

# 3. Flutter-Frontend starten (verbindet sich standardmäßig mit localhost:8000, keine weitere Konfiguration nötig)
cd apps/client/flutter && flutter pub get && flutter run -d chrome
```

- Port-Zuordnung des Abhängigkeits-Stacks: MySQL `3307`、Redis `6380`、OpenSearch `9200` (die Host-Ports 3306/6379 sind durch lokale Dienste belegt, siehe Kommentar in docker-compose.yml).
- Backend-Adresse und Schlüssel werden in `kratos/backend/config/` konfiguriert und unterstützen Überschreibung über Umgebungsvariablen (z. B. `PORT`, `OPENSEARCH_ADDR`).
- Flutter mit anderem Backend verbinden: `flutter run -d chrome --dart-define=API_BASE_URL=http://<host>:8000`.

Siehe [apps/README.md](../../apps/README.md) und [apps/client/flutter/README.md](../../apps/client/flutter/README.md).

## Release-Prozess

- **Automatisch**: Nach dem Push auf `main` erhöht GitHub Actions ([.github/workflows/release.yml](../../.github/workflows/release.yml)) automatisch die Patch-Version auf Basis des neuesten `v*`-Tags, erstellt und pusht den Tag und erstellt anschließend mit einem inkrementellen Changelog ein GitHub Release; übersprungen, wenn HEAD bereits einen Versions-Tag trägt. Das erste Release startet bei `v1.0.0`.
- **Manueller Fallback**: Führen Sie [scripts/post-push.sh](../../scripts/post-push.sh) aus (`gh` muss authentifiziert sein): `echo "x y refs/heads/main z" | scripts/post-push.sh`.
- **Manuell**:

  ```bash
  git tag -a v1.0.1 -m "release v1.0.1"
  git push origin v1.0.1
  gh release create v1.0.1 --generate-notes
  ```

## Roadmap

| Phase | Zeitraum | Schwerpunkte |
| :--- | :--- | :--- |
| Phase 1 | 2-3 Wochen | Kratos-Backend-Basisdienste + Integration von MySQL / Redis / OpenSearch |
| Phase 2 | 3-4 Wochen | Flutter / HarmonyOS-Multi-Endgeräte-Frontends + Erstellung mehrsprachiger ARB-Dateien |
| Phase 3 | 2 Wochen | Sicherheitshärtung (JWT / RBAC / Ratenbegrenzung) + Stresstests |
| Phase 4 | 1-2 Wochen | End-to-End-Integration aller Komponenten + Konfiguration der CDN-Beschleunigung |
| Phase 5 | laufend | Integration von KI-Empfehlungsalgorithmen, Tracking zur Analyse des Nutzerverhaltens |

Alle Aufgabenketten sind abgeschlossen.

## Unterstützung und Spenden

Wenn Ihnen dieses Projekt hilft, unterstützen Sie es gern mit **Star** und **Fork**; auch Spenden per QR-Code sind willkommen. Jede Unterstützung motiviert mich, das Projekt weiter zu pflegen und zu aktualisieren. Danke für Ihre Ermutigung!

<div align="center">

**WeChat-Spende** ｜ **Alipay-Spende**

<img src="../weixinpay.png" width="130" height="130" alt="WeChat-Spenden-QR-Code" />　<img src="../alipay.png" width="130" height="130" alt="Alipay-Spenden-QR-Code" />

</div>

### Globale Spenden per Überweisung (internationale Überweisung)

【Informationen zum Empfänger】

- Name des Empfängers: WANG KEXUN
- Kontonummer des Empfängers: 881015918251

【Empfängerbank】

- SWIFT Code der ZA Bank: AABLHKHHXXX
- Bankname: ZA Bank Limited
- Bankleitzahl: 387
- Bankadresse: Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【Korrespondenzbank für internationale Überweisungen (falls erforderlich)】

> Bitte beachten Sie: Dies sind die Angaben der Korrespondenzbank (Zwischenbank) für internationale Überweisungen, nicht die der Empfängerbank. Erkundigen Sie sich bei Ihrer Bank, ob die Angabe der Korrespondenzbank erforderlich ist.

**Für Überweisungen in Hongkong-Dollar, chinesischen Yuan und US-Dollar ist Citibank die Korrespondenzbank**

- Bankname: Citibank N.A. Hong Kong
- SWIFT Code: CITIHKHXXXX
- Bankleitzahl: 006
- Filialname: Hong Kong Branch
- Filialnummer: 391
- Bankadresse: Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**Für Überweisungen in anderen Währungen ist BNY Mellon die Korrespondenzbank**

- Bankname: THE BANK OF NEW YORK MELLON
- SWIFT Code: IRVTUS3NXXX
- Bankadresse: THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States

### Krypto-Spenden (Crypto Donation)

Wenn dieses Projekt Ihnen hilft, scannen Sie gerne den QR-Code, um zu spenden. Vielen Dank!

| Netzwerk (Network) | QR-Code (QR Code) | Wallet-Adresse (Wallet Address) |
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

## Lizenz und Kontakt

- **Lizenz**: Im Repository-Stammverzeichnis gibt es keine eigenständige LICENSE-Datei; `kratos/` ist der Upstream-Quellcode des Kratos-Frameworks und folgt dessen [MIT-Lizenz](../../kratos/LICENSE). Die Lizenzierung des Geschäftscodes wird durch spätere Bekanntmachungen des Projekts festgelegt.
- **Kontakt**: Kommunikation über GitHub Issues / PR; Spenden siehe oben unter „Unterstützung und Spenden“.
