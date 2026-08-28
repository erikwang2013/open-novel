# OpenSearch 定制镜像：预装多语言分词插件（kuromoji=日语, nori=韩语, icu=通用 Unicode）
# 基础镜像 = 本机当前运行版本 2.19.5（digest 锁定，避免 tag 漂移；升级时改 digest 并重编）
FROM opensearchproject/opensearch@sha256:e225ca1a828ef054e0335a0c234198a7b8f80abcc90f1159273ca7be5d8bbdff
USER root
RUN /usr/share/opensearch/bin/opensearch-plugin install --batch analysis-kuromoji analysis-nori analysis-icu \
  && chown -R opensearch:opensearch /usr/share/opensearch/plugins
USER opensearch
