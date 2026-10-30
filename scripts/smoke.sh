#!/usr/bin/env bash
# Smoke test for chapter + comment services (tasks #16/#19).
# Usage: JWT_SECRET=xxx PORT=18080 scripts/smoke.sh
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:18080/api}"
JWT_SECRET="${JWT_SECRET:-smoke-test-secret}"
BOOK="${BOOK:-$((700000 + RANDOM))}" # random book id: repeat runs stay idempotent
BACKEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../kratos/backend" && pwd)"

# mint tokens via the backend's own JWT helper (same claims as AuthManager)
token_bin=$(mktemp)
trap 'rm -f "$token_bin"' EXIT
(cd "$BACKEND_DIR" && go build -o "$token_bin" ./cmd/smoketoken)
mint() { JWT_SECRET="$JWT_SECRET" SMOKE_UID=1001 ROLE="$1" "$token_bin"; }
token=$(mint 2)        # author
reader_token=$(mint 1) # reader: must be rejected by the author guard
admin_token=$(mint 3)  # admin: required by admin stats endpoint

pass=0; fail=0
# check <desc> <want> <json>: success body has no "code" key (0); error body code == want.
check() {
  local got
  got=$(echo "$3" | python3 -c "import sys,json;d=json.load(sys.stdin);print(d['code'] if 'code' in d else '0')" 2>/dev/null || echo ERR)
  if [ "$got" = "$2" ]; then pass=$((pass+1)); echo "PASS: $1"; else fail=$((fail+1)); echo "FAIL: $1 (want $2 got $got): $3"; fi
}
# field <desc> <json> <python-expr> <want>
field() {
  local got
  got=$(echo "$2" | python3 -c "import sys,json;print(json.load(sys.stdin)$3)" 2>/dev/null || echo ERR)
  if [ "$got" = "$4" ]; then pass=$((pass+1)); echo "PASS: $1"; else fail=$((fail+1)); echo "FAIL: $1 (want $4 got $got)"; fi
}

hdr=(-H "Authorization: Bearer $token" -H 'Content-Type: application/json' -H 'X-Api-Version: v1')

# 正文接口校验书籍状态：先建一本测试书，章节挂在该书下（随机 BOOK 仅作覆盖前的兜底）
book_json=$(curl -s -X POST "$BASE/books" "${hdr[@]}" -d '{"title":"smoke book","lang":"zh-CN"}')
check "create book" 0 "$book_json"
BOOK=$(echo "$book_json" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

check "reader cannot create chapter" 140403 "$(curl -s -X POST "$BASE/books/$BOOK/chapters" -H "Authorization: Bearer $reader_token" -H 'Content-Type: application/json' -H 'X-Api-Version: v1' -d '{"chapter_no":99,"title":"X"}')"
ch1_json=$(curl -s -X POST "$BASE/books/$BOOK/chapters" "${hdr[@]}" -d '{"chapter_no":1,"title":"T1","content":"正文内容","lang":"zh-CN"}')
check "create chapter" 0 "$ch1_json"
field "chapter no" "$ch1_json" "['chapterNo']" "1"
field "chapter wordCount" "$ch1_json" "['wordCount']" "4"
chid=$(echo "$ch1_json" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")
check "create chapter 2" 0 "$(curl -s -X POST "$BASE/books/$BOOK/chapters" "${hdr[@]}" -d '{"chapter_no":2,"title":"T2","content":"第二节内容"}')"
check "list chapters" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/books/$BOOK/chapters?page=1&page_size=20")"
content_json=$(curl -s -H 'X-Api-Version: v1' "$BASE/chapters/$chid/content")
check "chapter content" 0 "$content_json"
field "content text" "$content_json" "['content']" "正文内容"
check "content miss -> 404" 140404 "$(curl -s -H 'X-Api-Version: v1' "$BASE/chapters/$chid/content?lang=en")"
check "put progress" 0 "$(curl -s -X PUT "$BASE/progress" "${hdr[@]}" -d "{\"book_id\":$BOOK,\"chapter_id\":1,\"position\":10}")"
check "get progress" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/progress?book_id=$BOOK" "${hdr[@]}")"
check "add bookshelf" 0 "$(curl -s -X POST "$BASE/bookshelf" "${hdr[@]}" -d "{\"book_id\":$BOOK}")"
check "list bookshelf" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/bookshelf" "${hdr[@]}")"
check "del bookshelf" 0 "$(curl -s -X DELETE "$BASE/bookshelf/$BOOK" "${hdr[@]}")"
comment_json=$(curl -s -X POST "$BASE/comments" "${hdr[@]}" -d "{\"book_id\":$BOOK,\"content\":\"smoke comment\"}")
check "post comment" 0 "$comment_json"
cid=$(echo "$comment_json" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")
check "list comments" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/comments?book_id=$BOOK")"
check "like" 0 "$(curl -s -X POST "$BASE/comments/$cid/like" "${hdr[@]}")"
check "like dup" 150409 "$(curl -s -X POST "$BASE/comments/$cid/like" "${hdr[@]}")"
check "report" 0 "$(curl -s -X POST "$BASE/comments/$cid/report" "${hdr[@]}")"
check "favorite" 0 "$(curl -s -X POST "$BASE/books/$BOOK/favorite" "${hdr[@]}")"
check "favorites list" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/favorites" "${hdr[@]}")"
check "no token rejected" 140401 "$(curl -s -X POST "$BASE/comments" -H 'Content-Type: application/json' -H 'X-Api-Version: v1' -d "{\"book_id\":$BOOK,\"content\":\"x\"}")"

# --- T-C-23: 新端点回归（分类/标签、admin 统计、支付公开端点、VIP 状态）---
# 评论举报 POST /api/comments/{id}/report 已由上方 "report" 用例覆盖（登录态、真实评论数据）
check "list categories" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/categories")"
check "list tags" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/tags")"
check "payments plans" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/payments/plans")"
check "vip status" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/payments/vip-status" "${hdr[@]}")"
check "admin stats" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/stats/overview" -H "Authorization: Bearer $admin_token")"
check "admin stats forbidden for reader" 180401 "$(curl -s -H 'X-Api-Version: v1' "$BASE/stats/overview" -H "Authorization: Bearer $reader_token")"
check "hot keywords" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/search/hot-keywords")"
check "admin categories" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/admin/categories" -H "Authorization: Bearer $admin_token")"
check "admin categories forbidden for reader" 180401 "$(curl -s -H 'X-Api-Version: v1' "$BASE/admin/categories" -H "Authorization: Bearer $reader_token")"
check "admin tags" 0 "$(curl -s -H 'X-Api-Version: v1' "$BASE/admin/tags" -H "Authorization: Bearer $admin_token")"
check "admin tags forbidden for reader" 180401 "$(curl -s -H 'X-Api-Version: v1' "$BASE/admin/tags" -H "Authorization: Bearer $reader_token")"

echo "---"
echo "passed=$pass failed=$fail"
[ "$fail" = 0 ]
