# Open Novel — 글로벌 다국어 소설 플랫폼

<div align="center">

[中文](../../README.md) · [English](README.en.md) · [日本語](README.ja.md) · **한국어** · [Русский](README.ru.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [বাংলা](README.bn.md) · [Bahasa Indonesia](README.id.md)

</div>

> **Go-Kratos** 마이크로서비스 아키텍처 + **Flutter / HarmonyOS** 멀티 클라이언트 프런트엔드 기반의 글로벌 다국어 소설 읽기 플랫폼으로, **12개 이상의 주요 언어**를 지원하며 전 세계 사용자에게 읽기, 상호작용, 검색 및 개인화 추천 기능을 제공합니다.

---

## 프로젝트 소개

Open Novel은 클라우드 네이티브 마이크로서비스 아키텍처의 글로벌 다국어 소설 플랫폼입니다:

- **백엔드**: Go-Kratos v2(gRPC / HTTP 듀얼 프로토콜). 마이크로서비스는 도메인별로 분리(사용자, 도서, 챕터, 댓글, 검색, 추천)
- **프런트엔드**: Flutter 올 플랫폼(Web / Desktop / Mobile) + HarmonyOS NEXT 네이티브 앱. 동일한 백엔드 API 공유
- **다국어 지원**: i18n 리소스 동적 로드, 12개 이상의 언어 지원(중국어, 영어, 일본어, 한국어, 프랑스어, 독일어, 스페인어, 러시아어, 아랍어 등)
- **스토리지**: MySQL 8(마스터-슬레이브 읽기/쓰기 분리) + Redis(핫 캐시 / 세션) + OpenSearch(다국어 검색)
- **운영**: Docker Compose 원클릭 배포, Prometheus + Grafana 모니터링, GitHub Actions 지속적 통합

## 기능 및 특징

<p align="center"><img src="images/ko/features.svg" alt="기능 아키텍처 다이어그램" width="860"/></p>

- **사용자 센터**: 회원가입·로그인(JWT), 개인 서재, 기기 간 독서 진행률 동기화, 다국어 프로필
- **독서 경험**: 챕터별 독서, 글꼴·크기 전환, 라이트/다크 테마, 오프라인 캐시, 페이지 넘김 애니메이션
- **도서 콘텐츠**: 도서 메타데이터, 챕터 관리, 카테고리 태그, 연재 업데이트, 다국어 번역
- **소통 커뮤니티**: 댓글·서평, 좋아요, 즐겨찾기, 신고·검수
- **검색 및 발견**: 다국어 형태소 분석 검색, 인기 랭킹, AI 추천, 카테고리 탐색
- **관리 백엔드**: 콘텐츠 검수, 사용자 관리, 데이터 통계, 설정 관리

## 시스템 아키텍처

<p align="center"><img src="images/ko/architecture.svg" alt="시스템 아키텍처 다이어그램" width="860"/></p>

## 프로젝트 전체 구조

<p align="center"><img src="images/ko/project.svg" alt="프로젝트 전체 구조 다이어그램" width="860"/></p>

## 요청 주기

<p align="center"><img src="images/ko/request-cycle.svg" alt="요청 주기 다이어그램" width="860"/></p>

## 보안 아키텍처

<p align="center"><img src="images/ko/security.svg" alt="보안 아키텍처 다이어그램" width="860"/></p>

## 프로젝트 구조

<p align="center"><img src="images/ko/structure.svg" alt="프로젝트 구조 다이어그램" width="860"/></p>

---

## 기술 스택

| 계층 | 기술 선정 |
| :--- | :--- |
| 클라이언트 | Flutter(Web / Desktop / Mobile), HarmonyOS NEXT(ArkTS / ArkUI) |
| 게이트웨이 | Nginx + CDN, Go-Kratos API 게이트웨이(gRPC / HTTP 듀얼 프로토콜) |
| 서버 | Go 1.22+, Kratos v2, protobuf / gRPC |
| 스토리지 | MySQL 8.0(마스터-슬레이브), Redis 7.x(Cluster), OpenSearch 2.x |
| 관측성 | Prometheus, Grafana, ELK, OpenTelemetry 링크 추적 |
| 운영 | Docker Compose, GitHub Actions CI/CD |

## 데이터베이스

- 데이터베이스 이름: `novel`
- 테이블 접두사: `novel_`(예: `novel_user`, `novel_book`, `novel_chapter`, `novel_comment` 등)

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

자세한 테이블 설계 및 읽기/쓰기 분리 전략은 [docs/novel-project-planning.md](../novel-project-planning.md)를 참조하세요.

## 멀티 클라이언트 디렉터리

```
apps/
├─ flutter/     # Flutter 올 플랫폼(Web / Desktop / Mobile), i18n 다국어 지원
└─ harmonyos/   # HarmonyOS NEXT 네이티브 앱(ArkTS / ArkUI)
```

자세한 내용은 [apps/README.md](../../apps/README.md)를 참조하세요.

## 로드맵

| 단계 | 기간 | 주요 작업 |
| :--- | :--- | :--- |
| Phase 1 | 2~3주 | Kratos 백엔드 기반 서비스 + MySQL / Redis / OpenSearch 통합 |
| Phase 2 | 3~4주 | Flutter / HarmonyOS 멀티 클라이언트 프런트엔드 + 다국어 ARB 작성 |
| Phase 3 | 2주 | 보안 강화(JWT / RBAC / 속도 제한) + 스트레스 테스트 |
| Phase 4 | 1~2주 | 전체 연동 테스트 + CDN 가속 설정 |
| Phase 5 | 지속 | AI 추천 알고리즘 도입, 사용자 행동 분석 트래킹 |

## 로컬 개발

```bash
# 의존 서비스 시작(MySQL / Redis / OpenSearch)
docker compose up -d

# 백엔드 서비스(Kratos 워크스페이스)
cd kratos/backend && go mod tidy && go run ./cmd/server

# Flutter 클라이언트
cd apps/flutter && flutter pub get && flutter run

# HarmonyOS 클라이언트
cd apps/harmonyos && hvigorw assembleHap
```

---

## 후원 및 기부

이 프로젝트가 도움이 되셨다면 **Star**와 **Fork**로 지원해 주세요. 또한 QR 코드를 스캔하여 기부해 주셔도 환영합니다. 여러분의 지원 하나하나가 제가 계속 유지보수하고 업데이트하는 원동력입니다. 감사합니다!

<div align="center">

**위챗 후원** ｜ **알리페이 후원**

<img src="../weixinpay.png" width="130" height="130" alt="위챗 후원 QR 코드" />　<img src="../alipay.png" width="130" height="130" alt="알리페이 후원 QR 코드" />

</div>

### 글로벌 송금 후원(국제 송금)

【수취인 정보】

- 수취인 이름: WANG KEXUN
- 수취 계좌번호: 881015918251

【수취 은행】

- ZA Bank SWIFT Code: AABLHKHHXXX
- 은행명: ZA Bank Limited
- 은행 번호: 387
- 은행 주소: Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【국제 송금 중개 은행(필요한 경우)】

> 유의하시기 바랍니다. 이는 국제 송금 중개 은행(중개 은행) 정보이며 수취 은행 정보가 아닙니다. 송금 은행에 국제 송금 중개 은행 정보 제공이 필요한지 문의하시기 바랍니다.

**홍콩 달러, 위안화, 미국 달러 입금 시 중개 은행은 Citibank입니다**

- 은행명: Citibank N.A. Hong Kong
- SWIFT Code: CITIHKHXXXX
- 은행 번호: 006
- 지점명: Hong Kong Branch
- 지점 번호: 391
- 은행 주소: Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**기타 통화 입금 시 중개 은행은 BNY Mellon입니다**

- 은행명: THE BANK OF NEW YORK MELLON
- SWIFT Code: IRVTUS3NXXX
- 은행 주소: THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States
