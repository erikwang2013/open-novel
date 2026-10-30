# open-novel backend 构建与生成入口。
# proto 重新生成：用 protoc-gen-go-http（kratos 官方插件，v2.9.2 或 v3 需与 go.mod 一致）。

GOBIN = /home/wwwroot/go/bin
THIRD_PARTY = ../third_party
PROTOS = api/user/v1/user.proto api/book/v1/book.proto api/chapter/v1/chapter.proto \
	api/comment/v1/comment.proto api/search/v1/search.proto api/recommendation/v1/recommendation.proto \
	api/payment/v1/payment.proto api/admin/v1/admin.proto

.PHONY: proto build vet test

proto:
	@mkdir -p /tmp/open-novel-gen
	@for p in $(PROTOS); do \
		protoc --plugin=protoc-gen-go_http=$(GOBIN)/protoc-gen-go-http \
			-I api -I $(THIRD_PARTY) \
			--go_out=/tmp/open-novel-gen --go_http_out=/tmp/open-novel-gen $$p || exit 1; \
	done
	@for s in user book chapter comment search recommendation payment admin; do \
		cp /tmp/open-novel-gen/open-novel/backend/api/$$s/v1/$$s.pb.go \
		   /tmp/open-novel-gen/open-novel/backend/api/$$s/v1/$${s}_http.pb.go \
		   api/$$s/v1/; \
	done
	@echo "proto regenerated"

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./internal/service/
