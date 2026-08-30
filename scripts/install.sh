#!/usr/bin/env bash
# Open Novel 一键安装脚本：环境检查 -> 依赖栈启动 -> 后端/前端启动提示。
# 幂等，可重复执行。用法: bash scripts/install.sh [--skip-deps]
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # 项目根目录

SKIP_DEPS=0
for arg in "$@"; do
  case "$arg" in
    --skip-deps) SKIP_DEPS=1 ;;
    -h|--help)
      echo "用法: bash scripts/install.sh [--skip-deps]" >&2
      echo "  --skip-deps  跳过依赖栈启动与 MySQL 就绪等待（仅打印启动提示）" >&2
      exit 0 ;;
    *) echo "install: 未知参数 $arg（可用 --skip-deps 跳过依赖栈）" >&2; exit 2 ;;
  esac
done

missing=() # 缺失或版本过低的组件

note_missing() { # note_missing <名称> <提示行...>
  local name=$1; shift
  printf '[install] 缺少 %s：\n' "$name"
  printf '          %s\n' "$@"
  missing+=("$name")
}

echo "[install] ① 环境检查"
command -v docker >/dev/null 2>&1 \
  || note_missing docker "安装 Docker: https://docs.docker.com/get-docker/" "（Linux 用发行版包管理器，Windows/macOS 用 Docker Desktop）"
if command -v docker >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
  note_missing "docker compose" "Docker 需要 Compose 插件（docker compose 子命令），请更新 Docker 或安装 docker-compose-plugin"
fi
if ! command -v go >/dev/null 2>&1; then
  note_missing go "安装 Go >= 1.22: https://go.dev/dl/"
else
  gover=$(go version | sed -nE 's/.*go([0-9]+\.[0-9]+).*/\1/p')
  go_ok() { # go_ok <主版本.次版本>：纯 bash 比较，避免依赖 sort -V
    local major minor
    IFS=. read -r major minor _ <<< "$1"
    [[ -n "$major" && -n "$minor" ]] || return 1
    (( major > 1 )) || (( major == 1 && minor >= 22 ))
  }
  if ! go_ok "${gover:-}"; then
    printf '[install] go 版本过低（当前 %s，需要 >= 1.22）: https://go.dev/dl/\n' "${gover:-未知}"
    missing+=("go")
  fi
fi
command -v flutter >/dev/null 2>&1 \
  || note_missing flutter "安装 Flutter 3.x: https://docs.flutter.dev/get-started/install"

# go/flutter 缺失只提示不阻塞：脚本只打印其启动命令，不代为执行；docker 缺失则无法启动依赖栈
if [[ "$SKIP_DEPS" != 1 ]] && ! command -v docker >/dev/null 2>&1; then
  echo "[install] Docker 未安装，无法启动依赖栈。请先安装 Docker 后重试，或使用 --skip-deps 仅打印启动提示。" >&2
  exit 1
fi
if [[ "$SKIP_DEPS" != 1 ]] && command -v docker >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
  echo "[install] Docker Compose 插件不可用，无法启动依赖栈。请安装 docker-compose-plugin 后重试。" >&2
  exit 1
fi

if [[ "$SKIP_DEPS" != 1 ]]; then
  if ! docker info >/dev/null 2>&1; then
    echo "[install] Docker 已安装但守护进程未运行，请先启动 Docker 后重试。" >&2
    exit 1
  fi

  echo "[install] ② 启动依赖栈（MySQL / Redis / OpenSearch）"
  docker compose up -d

  echo "[install] ③ 等待 MySQL 就绪（127.0.0.1:3307，最长 60s）"
  probe_mysql() { # bash /dev/tcp 探测端口；bash 3.2（macOS）无 /dev/tcp 时退回 nc
    (exec 3<>/dev/tcp/127.0.0.1/3307) 2>/dev/null || nc -z 127.0.0.1 3307 2>/dev/null
  }
  ready=0
  for _ in $(seq 1 60); do
    if probe_mysql; then ready=1; break; fi
    sleep 1
  done
  if [[ "$ready" != 1 ]]; then
    echo "[install] 60s 内 MySQL 未就绪，请检查: docker compose logs mysql" >&2
    exit 1
  fi
  echo "[install] MySQL 已就绪（OpenSearch 首次构建镜像较慢，后端启动后如搜索报错请稍候重试）"
fi

echo "[install] ④ 启动后端（前台运行，请另开终端执行）"
echo "    cd kratos/backend && go mod tidy && go run ./cmd/server"
echo "    （HTTP http://localhost:8000，gRPC :9000；配置在 kratos/backend/config/，支持环境变量覆盖）"

echo "[install] ⑤ 启动前端与访问地址"
echo "    客户端: cd apps/client/flutter && flutter pub get && flutter run -d chrome"
echo "    管理端: cd apps/admin && flutter pub get && flutter run -d chrome"
echo "    HarmonyOS: 用 DevEco Studio 打开 apps/client/harmonyos 运行"
echo "    后端 HTTP: http://localhost:8000（接口前缀 /api，版本头 X-Api-Version: v1）"
echo "    客户端/管理端端口由 flutter run 随机分配并打印在控制台（可用 --web-port 固定）"
echo "    依赖栈端口映射: MySQL 3307 / Redis 6380 / OpenSearch 9200"

if [[ "${#missing[@]}" -gt 0 ]]; then
  echo "[install] 注意：以下组件缺失或版本过低，安装完成后按上面命令启动即可：${missing[*]}"
fi
echo "[install] 完成。"
