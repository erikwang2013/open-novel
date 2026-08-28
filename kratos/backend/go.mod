module open-novel/backend

go 1.25.0

replace github.com/go-kratos/kratos/v2 => ..

require (
	github.com/go-kratos/kratos/v2 v2.9.2
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/redis/go-redis/v9 v9.14.1
	golang.org/x/crypto v0.55.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260825221802-da73d73af1c5
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.31.2
	gorm.io/plugin/dbresolver v1.6.2
)

require github.com/rogpeppe/go-internal v1.14.1 // indirect

require (
	dario.cat/mergo v1.0.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-playground/form/v4 v4.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/text v0.2.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260819154853-08b0e4226688 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
