# Open Novel — বিশ্বব্যাপী বহুভাষিক উপন্যাস প্ল্যাটফর্ম

<div align="center">

[中文](../../README.md) · [English](README.en.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [Русский](README.ru.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · **বাংলা** · [Bahasa Indonesia](README.id.md)

</div>

> **Go-Kratos** মাইক্রোসার্ভিস আর্কিটেকচার + **Flutter / HarmonyOS** মাল্টি-প্ল্যাটফর্ম ফ্রন্টএন্ডের উপর ভিত্তি করে গড়ে ওঠা বিশ্বব্যাপী বহুভাষিক উপন্যাস পড়ার প্ল্যাটফর্ম, যা **১২+টি প্রধান ভাষা** সমর্থন করে এবং বিশ্বজুড়ে ব্যবহারকারীদের পড়া, ইন্টারঅ্যাকশন, সার্চ ও ব্যক্তিগতকৃত রিকমেন্ডেশন সুবিধা প্রদান করে।

---

## প্রকল্প পরিচিতি

Open Novel একটি ক্লাউড-নেটিভ মাইক্রোসার্ভিস আর্কিটেকচারভিত্তিক বিশ্বব্যাপী বহুভাষিক উপন্যাস প্ল্যাটফর্ম:

- **ব্যাকএন্ড**: Go-Kratos v2 (gRPC / HTTP দ্বৈত প্রোটোকল), ডোমেইনভিত্তিতে বিভক্ত মাইক্রোসার্ভিস (ব্যবহারকারী, বই, অধ্যায়, মন্তব্য, সার্চ, রিকমেন্ডেশন)
- **ফ্রন্টএন্ড**: Flutter সর্ব-প্ল্যাটফর্ম (Web / Desktop / Mobile) + HarmonyOS NEXT নেটিভ অ্যাপ, সবগুলো একই ব্যাকএন্ড API ব্যবহার করে
- **বহুভাষিকতা**: i18n রিসোর্স ডায়নামিকভাবে লোড হয়, ১২+ ভাষা সমর্থন করে (চীনা, ইংরেজি, জাপানি, কোরিয়ান, ফরাসি, জার্মান, স্প্যানিশ, রুশ, আরবি ইত্যাদি)
- **স্টোরেজ**: MySQL 8 (মাস্টার-স্লেভ রিড-রাইট সেপারেশন) + Redis (হট ক্যাশ / সেশন) + OpenSearch (বহুভাষিক সার্চ)
- **অপারেশন**: Docker Compose এক-ক্লিক ডিপ্লয়মেন্ট, Prometheus + Grafana মনিটরিং, GitHub Actions কন্টিনিউয়াস ইন্টিগ্রেশন


## বৈশিষ্ট্যসমূহ

<p align="center"><img src="images/features.svg" alt="ফিচার আর্কিটেকচার ডায়াগ্রাম" width="860"/></p>

- **ব্যবহারকারী সেন্টার**: রেজিস্ট্রেশন/লগইন (JWT), ব্যক্তিগত বুকশেলফ, ডিভাইস জুড়ে পড়ার অগ্রগতি সিঙ্ক, বহুভাষিক প্রোফাইল
- **পড়ার অভিজ্ঞতা**: অধ্যায়ভিত্তিক পড়া, ফন্ট ও সাইজ পরিবর্তন, হালকা/গাঢ় থিম, অফলাইন ক্যাশ, পেজ-ফ্লিপ অ্যানিমেশন
- **বইয়ের বিষয়বস্তু**: বইয়ের মেটাডেটা, অধ্যায় ব্যবস্থাপনা, ক্যাটাগরি ট্যাগ, সিরিয়াল আপডেট, বহুভাষিক অনুবাদ
- **ইন্টারঅ্যাক্টিভ কমিউনিটি**: মন্তব্য ও রিভিউ, লাইক, বুকমার্ক, রিপোর্ট ও মডারেশন
- **সার্চ ও আবিষ্কার**: বহুভাষিক টোকেনাইজেশন সার্চ, জনপ্রিয় র্যাঙ্কিং, AI রিকমেন্ডেশন, ক্যাটাগরি ব্রাউজিং
- **অ্যাডমিন প্যানেল**: কন্টেন্ট মডারেশন, ইউজার ম্যানেজমেন্ট, ডেটা পরিসংখ্যান, কনফিগারেশন ম্যানেজমেন্ট

## সিস্টেম আর্কিটেকচার

<p align="center"><img src="images/architecture.svg" alt="সিস্টেম আর্কিটেকচার ডায়াগ্রাম" width="860"/></p>

## প্রকল্পের সার্বিক চিত্র

<p align="center"><img src="images/project.svg" alt="প্রকল্পের সার্বিক চিত্র ডায়াগ্রাম" width="860"/></p>

## রিকোয়েস্ট সাইকেল

<p align="center"><img src="images/request-cycle.svg" alt="রিকোয়েস্ট সাইকেল ডায়াগ্রাম" width="860"/></p>

## নিরাপত্তা আর্কিটেকচার

<p align="center"><img src="images/security.svg" alt="নিরাপত্তা আর্কিটেকচার ডায়াগ্রাম" width="860"/></p>

## প্রকল্প কাঠামো

<p align="center"><img src="images/structure.svg" alt="প্রকল্প কাঠামো ডায়াগ্রাম" width="860"/></p>

---

## টেকনোলজি স্ট্যাক

| স্তর | প্রযুক্তি নির্বাচন |
| :--- | :--- |
| ক্লায়েন্ট | Flutter (Web / Desktop / Mobile), HarmonyOS NEXT (ArkTS / ArkUI) |
| গেটওয়ে | Nginx + CDN, Go-Kratos API গেটওয়ে (gRPC / HTTP দ্বৈত প্রোটোকল) |
| সার্ভার | Go 1.22+, Kratos v2, protobuf / gRPC |
| স্টোরেজ | MySQL 8.0 (মাস্টার-স্লেভ), Redis 7.x (Cluster), OpenSearch 2.x |
| অবজারভেবিলিটি | Prometheus, Grafana, ELK, OpenTelemetry ট্রেসিং |
| অপারেশন | Docker Compose, GitHub Actions CI/CD |

## ডেটাবেস

- ডেটাবেসের নাম: `novel`
- টেবিল প্রিফিক্স: `novel_` (যেমন `novel_user`, `novel_book`, `novel_chapter`, `novel_comment` ইত্যাদি)

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

বিস্তারিত টেবিল ডিজাইন ও রিড-রাইট সেপারেশন কৌশলের জন্য [docs/novel-project-planning.md](../novel-project-planning.md) দেখুন।

## মাল্টি-এন্ড ডিরেক্টরি

```
apps/
├─ flutter/     # Flutter সর্ব-প্ল্যাটফর্ম (Web / Desktop / Mobile), i18n বহুভাষিক
└─ harmonyos/   # HarmonyOS NEXT নেটিভ অ্যাপ (ArkTS / ArkUI)
```

বিস্তারিত জানতে দেখুন [apps/README.md](../../apps/README.md)।

## রোডম্যাপ

| পর্যায় | সময়কাল | কাজের মূল ফোকাস |
| :--- | :--- | :--- |
| Phase 1 | ২-৩ সপ্তাহ | Kratos ব্যাকএন্ড বেস সার্ভিস + MySQL / Redis / OpenSearch ইন্টিগ্রেশন |
| Phase 2 | ৩-৪ সপ্তাহ | Flutter / HarmonyOS মাল্টি-প্ল্যাটফর্ম ফ্রন্টএন্ড + বহুভাষিক ARB লেখা |
| Phase 3 | ২ সপ্তাহ | সিকিউরিটি হার্ডেনিং (JWT / RBAC / রেট লিমিটিং) + স্ট্রেস টেস্ট |
| Phase 4 | ১-২ সপ্তাহ | ফুল-পাইপলাইন ইন্টিগ্রেশন টেস্ট + CDN এক্সিলারেশন কনফিগারেশন |
| Phase 5 | চলমান | AI রিকমেন্ডেশন অ্যালগরিদম ইন্টিগ্রেশন, ইউজার বিহেভিয়ার অ্যানালিটিক্স ট্র্যাকিং |

## লোকাল ডেভেলপমেন্ট

```bash
# ডিপেন্ডেন্সি চালু করুন (MySQL / Redis / OpenSearch)
docker compose up -d

# ব্যাকএন্ড সার্ভিস (Kratos ওয়ার্কস্পেস)
cd backend && go mod tidy && go run ./cmd/server

# Flutter এন্ড
cd apps/flutter && flutter pub get && flutter run

# HarmonyOS এন্ড
cd apps/harmonyos && hvigorw assembleHap
```

---

## সাপোর্ট ও দান

যদি এই প্রকল্পটি আপনার কাজে লাগে, তাহলে **Star** ও **Fork** দিয়ে সাপোর্ট করতে স্বাগতম; স্ক্যান করে দান করেও সাপোর্ট করতে পারেন। আপনার প্রতিটি সমর্থনই আমার নিয়মিত মেইনটেন্যান্স ও আপডেটের চালিকাশক্তি, আপনার উৎসাহের জন্য ধন্যবাদ!

<div align="center">

**WeChat দান** ｜ **Alipay দান**

<img src="../weixinpay.png" width="130" height="130" alt="WeChat দান QR কোড" />　<img src="../alipay.png" width="130" height="130" alt="Alipay দান QR কোড" />

</div>

### বৈশ্বিক ব্যাংক ট্রান্সফার দান (ক্রস-বর্ডার রেমিট্যান্স)

【প্রাপকের তথ্য】

- প্রাপকের নাম: WANG KEXUN
- প্রাপকের অ্যাকাউন্ট নম্বর: 881015918251

【প্রাপক ব্যাংক】

- ZA Bank SWIFT কোড: AABLHKHHXXX
- ব্যাংকের নাম: ZA Bank Limited
- ব্যাংক কোড: 387
- ব্যাংকের ঠিকানা: Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【ক্রস-বর্ডার রেমিট্যান্স এজেন্ট ব্যাংক (প্রয়োজন হলে)】

> অনুগ্রহ করে লক্ষ্য করুন: এটি ক্রস-বর্ডার রেমিট্যান্স এজেন্ট ব্যাংক (মধ্যস্থতাকারী ব্যাংক) এর তথ্য, প্রাপক ব্যাংকের তথ্য নয়। রেমিট্যান্স পাঠানোর ব্যাংককে জিজ্ঞাসা করুন ক্রস-বর্ডার এজেন্ট ব্যাংকের তথ্য প্রদান করা প্রয়োজন কিনা।

**হংকং ডলার, রেনমিনবি ও মার্কিন ডলারের জন্য এজেন্ট ব্যাংক Citibank**

- ব্যাংকের নাম: Citibank N.A. Hong Kong
- SWIFT কোড: CITIHKHXXXX
- ব্যাংক কোড: 006
- শাখার নাম: Hong Kong Branch
- শাখা কোড: 391
- ব্যাংকের ঠিকানা: Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**অন্যান্য মুদ্রার জন্য এজেন্ট ব্যাংক BNY Mellon**

- ব্যাংকের নাম: THE BANK OF NEW YORK MELLON
- SWIFT কোড: IRVTUS3NXXX
- ব্যাংকের ঠিকানা: THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States
