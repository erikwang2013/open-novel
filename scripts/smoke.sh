#!/usr/bin/env bash
# Smoke test for chapter + comment services (tasks #16/#19).
# Usage: JWT_SECRET=xxx PORT=18080 scripts/smoke.sh
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:18080/api/v1}"
JWT_SECRET="${JWT_SECRET:-smoke-test-secret}"
BOOK="${BOOK:-2}" # use a fresh book id to stay idempotent

# mint an author token (role=2) for user 1001
token=$(cd /tmp/novel-smoke && JWT_SECRET="$JWT_SECRET" go run token.go 2>/dev/null || \
  (cd /tmp/novel-smoke && go mod init smoke >/dev/null 2>&1; GOPROXY=https://goproxy.cn,direct GOFLAGS=-mod=mod go mod tidy >/dev/null 2>&1 && JWT_SECRET="$JWT_SECRET" go run token.go))
# role-1 (reader) token: must be rejected by the author guard
reader_token=$(cd /tmp/novel-smoke && ROLE=1 JWT_SECRET="$JWT_SECRET" go run token.go 2>/dev/null)

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

hdr=(-H "Authorization: Bearer $token" -H 'Content-Type: application/json')

check "reader cannot create chapter" 140403 "$(curl -s -X POST "$BASE/books/$BOOK/chapters" -H "Authorization: Bearer $reader_token" -H 'Content-Type: application/json' -d '{"chapter_no":99,"title":"X"}')"
ch1_json=$(curl -s -X POST "$BASE/books/$BOOK/chapters" "${hdr[@]}" -d '{"chapter_no":1,"title":"T1","content":"正文内容","lang":"zh-CN"}')
check "create chapter" 0 "$ch1_json"
field "chapter no" "$ch1_json" "['chapterNo']" "1"
field "chapter wordCount" "$ch1_json" "['wordCount']" "4"
chid=$(echo "$ch1_json" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")
check "create chapter 2" 0 "$(curl -s -X POST "$BASE/books/$BOOK/chapters" "${hdr[@]}" -d '{"chapter_no":2,"title":"T2","content":"第二节内容"}')"
check "list chapters" 0 "$(curl -s "$BASE/books/$BOOK/chapters?page=1&page_size=20")"
content_json=$(curl -s "$BASE/chapters/$chid/content")
check "chapter content" 0 "$content_json"
field "content text" "$content_json" "['content']" "正文内容"
check "content miss -> 404" 140404 "$(curl -s "$BASE/chapters/$chid/content?lang=en")"
check "put progress" 0 "$(curl -s -X PUT "$BASE/progress" "${hdr[@]}" -d "{\"book_id\":$BOOK,\"chapter_id\":1,\"position\":10}")"
check "get progress" 0 "$(curl -s "$BASE/progress?book_id=$BOOK" "${hdr[@]}")"
check "add bookshelf" 0 "$(curl -s -X POST "$BASE/bookshelf" "${hdr[@]}" -d "{\"book_id\":$BOOK}")"
check "list bookshelf" 0 "$(curl -s "$BASE/bookshelf" "${hdr[@]}")"
check "del bookshelf" 0 "$(curl -s -X DELETE "$BASE/bookshelf/$BOOK" "${hdr[@]}")"
comment_json=$(curl -s -X POST "$BASE/comments" "${hdr[@]}" -d "{\"book_id\":$BOOK,\"content\":\"smoke comment\"}")
check "post comment" 0 "$comment_json"
cid=$(echo "$comment_json" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")
check "list comments" 0 "$(curl -s "$BASE/comments?book_id=$BOOK")"
check "like" 0 "$(curl -s -X POST "$BASE/comments/$cid/like" "${hdr[@]}")"
check "like dup" 150409 "$(curl -s -X POST "$BASE/comments/$cid/like" "${hdr[@]}")"
check "report" 0 "$(curl -s -X POST "$BASE/comments/$cid/report" "${hdr[@]}")"
check "favorite" 0 "$(curl -s -X POST "$BASE/books/$BOOK/favorite" "${hdr[@]}")"
check "favorites list" 0 "$(curl -s "$BASE/favorites" "${hdr[@]}")"
check "no token rejected" 140401 "$(curl -s -X POST "$BASE/comments" -H 'Content-Type: application/json' -d "{\"book_id\":$BOOK,\"content\":\"x\"}")"

echo "---"
echo "passed=$pass failed=$fail"
[ "$fail" = 0 ]
