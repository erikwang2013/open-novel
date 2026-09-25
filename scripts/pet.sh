#!/usr/bin/env bash
# 项目宠物 Novi（终端版）：按仓库当前状态显示心情。
# 用法：scripts/pet.sh            看心情
#       scripts/pet.sh --line     只输出一行状态（挂 pre-commit / CI 日志用）
# 心情取自：工作区是否干净、是否领先/落后远端分支。
set -euo pipefail
cd "$(dirname "$0")/.."

dirty=$(git status --porcelain | wc -l | tr -d ' ')
ahead=$(git rev-list --count '@{u}..HEAD' 2>/dev/null || echo 0)
behind=$(git rev-list --count 'HEAD..@{u}' 2>/dev/null || echo 0)

if (( dirty > 0 )); then
  mood="dirty";  face='( o  o )'; line="工作区有 ${dirty} 处未提交改动 —— Novi 盯着你的 diff"
elif (( ahead > 0 )); then
  mood="ahead";  face='( ^  ^ )'; line="领先远端 ${ahead} 个提交 —— Novi 在门口等着推送"
elif (( behind > 0 )); then
  mood="behind"; face='( .  . )'; line="落后远端 ${behind} 个提交 —— Novi 饿了，想拉取"
else
  mood="happy";  face='( ^  ^ )'; line="工作区干净、与远端同步 —— Novi 神清气爽"
fi

if [[ "${1:-}" == "--line" ]]; then
  echo "Novi[$mood] $line"
  exit 0
fi

color=32 # 绿
case "$mood" in dirty) color=33 ;; behind) color=31 ;; esac # 黄 / 红

printf '\n'
printf '   \033[33m/\\   /\\\033[0m\n'
printf '  \033[33m/  \\_/  \\\033[0m      \033[1;%smNovi\033[0m  \033[%sm%s\033[0m\n' "$color" "$color" "$mood"
printf '  \033[1;37m%s\033[0m      \033[2m%s\033[0m\n' "$face" "$line"
printf '   \033[33m\\  %s  /\033[0m\n' '▽'
printf '  \033[1;37m__/_______\\__\033[0m\n'
printf '   \033[36m^  ^  ^  ^  ^\033[0m\n'
printf '\n'
