#!/usr/bin/env bash
# 推送规则：main 推送成功后，基于最新 v* tag 递增 patch 版本，
# 创建 tag、推送 tag、以增量 changelog 创建 GitHub release。
# 只响应 refs/heads/main 的推送；tag 推送被忽略，避免递归。
# 验证：echo "x y refs/heads/main z" | DRY_RUN=1 scripts/post-push.sh
set -euo pipefail

DRY_RUN=${DRY_RUN:-0}
BASE_VERSION=v1.0.0 # 无 tag 时的初始版本（用户裁定）

triggered=0
while read -r _ _ rref _; do
  [[ "$rref" == "refs/heads/main" ]] && triggered=1
done
[[ $triggered == 0 ]] && exit 0

prev_tag=$(git tag --list 'v*' --sort=-v:refname | head -n1)
if [[ -n "$prev_tag" ]] && [[ "$(git rev-list -n1 "$prev_tag")" == "$(git rev-parse HEAD)" ]]; then
  exit 0 # HEAD 已带版本 tag，无新内容
fi

if [[ -n "$prev_tag" ]]; then
  IFS=. read -r maj min pat <<< "${prev_tag#v}"
  new_tag="v$maj.$min.$((pat + 1))"
else
  new_tag=$BASE_VERSION
fi

command -v gh >/dev/null || { echo "post-push: gh 未安装，跳过发布" >&2; exit 1; }

{
  echo "prev: ${prev_tag:-无}  new: $new_tag"
  if [[ -n "$prev_tag" ]]; then
    git log --oneline --no-merges "$prev_tag..HEAD"
  else
    git log --oneline --no-merges
  fi
}
[[ $DRY_RUN == 1 ]] && exit 0

git tag -a "$new_tag" -m "release $new_tag"
git push origin "$new_tag"

notes=$(mktemp)
trap 'rm -f "$notes"' EXIT
if [[ -n "$prev_tag" ]]; then
  git log --oneline --no-merges "$prev_tag..HEAD" >"$notes"
else
  git log --oneline --no-merges >"$notes"
fi
gh release create "$new_tag" --title "$new_tag" --notes-file "$notes"
echo "post-push: created $new_tag"
