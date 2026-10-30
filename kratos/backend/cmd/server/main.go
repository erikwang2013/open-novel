package main

// open-novel 后端入口（Kratos v2 标准布局）。
// 启动：cd backend && go run ./cmd/server -conf config/config.yaml

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
	"open-novel/backend/internal/server"
	"open-novel/backend/internal/service"
)

var (
	confPath = flag.String("conf", "config/config.yaml", "config file path")
)

func main() {
	flag.Parse()
	logger := log.NewStdLogger(os.Stdout)

	cfg, err := conf.Load(*confPath)
	if err != nil {
		panic(err)
	}
	d, err := data.NewData(cfg.Data)
	if err != nil {
		panic(err)
	}
	defer d.Cache.Close()
	// JWT 密钥：生产必须注入非默认值；开发缺省用 DefaultJWTSecret 并告警。
	secret := cfg.Auth.JwtSecret
	if os.Getenv("APP_ENV") == "production" && (secret == "" || secret == pkg.DefaultJWTSecret) {
		logger.Log(log.LevelError, "msg", "JWT_SECRET must be set to a non-default value in production")
		os.Exit(1)
	}
	if secret == "" || secret == pkg.DefaultJWTSecret {
		logger.Log(log.LevelWarn, "msg", "using default JWT secret; set JWT_SECRET in production")
	}
	am := pkg.NewAuthManager(d.RDB, secret,
		time.Duration(cfg.Auth.JwtAccessTtl)*time.Second,
		time.Duration(cfg.Auth.JwtRefreshTtl)*time.Second)

	// 启动时幂等建搜索索引（best-effort，失败不影响启动）
	searchUc := biz.NewSearchUsecase(d)
	if err := searchUc.EnsureIndex(context.Background()); err != nil {
		logger.Log(log.LevelWarn, "msg", "ensure search index failed", "err", err.Error())
	}

	userSvc := service.NewUserService(biz.NewUserUsecase(d, am))
	bookSvc := service.NewBookService(biz.NewBookUsecase(d, searchUc))
	payUc := biz.NewPaymentUsecase(d, cfg.Payment)
	chapterSvc := service.NewChapterService(biz.NewChapterUsecase(d, payUc))
	commentSvc := service.NewCommentService(biz.NewCommentUsecase(d))
	searchSvc := service.NewSearchService(searchUc)
	recSvc := service.NewRecommendationService(biz.NewRecommendUsecase(d))
	paySvc := service.NewPaymentService(payUc)
	adminSvc := service.NewAdminService(biz.NewAdminUsecase(d, searchUc))

	httpSrv := server.NewHTTPServer(cfg.Server, am, logger,
		userSvc, bookSvc, chapterSvc, commentSvc, searchSvc, recSvc, paySvc, adminSvc)
	grpcSrv := server.NewGRPCServer(cfg.Server, am, logger,
		userSvc, bookSvc, chapterSvc, commentSvc, searchSvc, recSvc, paySvc, adminSvc)

	app := kratos.New(
		kratos.Name("open-novel"),
		kratos.Version("0.1.0"),
		kratos.Logger(logger),
		kratos.Server(httpSrv, grpcSrv),
	)
	if err := app.Run(); err != nil {
		logger.Log(log.LevelError, "msg", "app run failed", "err", err)
		os.Exit(1)
	}
}
