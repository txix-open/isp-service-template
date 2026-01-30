package assembly

import (
	"context"

	"isp-service-template/conf"

	"github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/observability/sentry"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/dbrx"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/log"
)

type Assembly struct {
	boot       *bootstrap.StandaloneBootstrap
	db         *dbrx.Client
	grpcServer *grpc.Server
	httpServer *http.Server
	mdmCli     *client.Client
	logger     *log.Adapter
	mqCli      *grmqx.Client
}

func New(boot *bootstrap.StandaloneBootstrap) (*Assembly, error) {
	logger := boot.App.Logger()

	db := dbrx.New(logger, dbx.WithMigrationRunner(boot.MigrationsDir, logger))
	mdmCli, err := client.Default()
	if err != nil {
		return nil, errors.WithMessage(err, "create mdm client")
	}
	mqCli := grmqx.New(sentry.WrapErrorLogger(logger, boot.SentryHub))
	boot.HealthcheckRegistry.Register("db", db)
	boot.HealthcheckRegistry.Register("mq", mqCli)
	return &Assembly{
		boot:       boot,
		db:         db,
		grpcServer: grpc.DefaultServer(),
		httpServer: http.NewServer(logger),
		mdmCli:     mdmCli,
		logger:     logger,
		mqCli:      mqCli,
	}, nil
}

func (a *Assembly) ReceiveConfig(shortTtlCtx context.Context, cfg conf.Config) error {
	a.logger.SetLevel(cfg.LogLevel)

	err := a.db.Upgrade(shortTtlCtx, cfg.Database)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "upgrade db client"))
	}

	locator := NewLocator(a.db, sentry.WrapErrorLogger(a.logger, a.boot.SentryHub))
	handlers := locator.Handlers(cfg)

	a.grpcServer.Upgrade(handlers.GrpcHandler)

	err = a.mqCli.Upgrade(a.boot.App.Context(),
		grmqx.NewConfig(
			cfg.Consumer.Client.Url(),
			grmqx.WithConsumers(handlers.RmqHandler),
			grmqx.WithDeclarations(grmqx.TopologyFromConsumers(cfg.Consumer.Config)),
		),
	)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "upgrade mq client"))
	}

	a.httpServer.Upgrade(handlers.HttpHandler)

	return nil
}

func (a *Assembly) Runners() []app.Runner {
	return []app.Runner{
		app.RunnerFunc(func(ctx context.Context) error {
			err := a.grpcServer.ListenAndServe(a.boot.BindingAddress)
			if err != nil {
				return errors.WithMessage(err, "listen ans serve grpc grpcServer")
			}
			return nil
		}),
		// app.RunnerFunc(func(ctx context.Context) error {
		// 	err := a.httpServer.ListenAndServe(a.boot.BindingAddress)
		// 	if err != nil {
		// 		return errors.WithMessage(err, "listen ans serve http httpServer")
		// 	}
		// 	return nil
		// }),
	}
}

func (a *Assembly) Closers() []app.Closer {
	return []app.Closer{
		app.CloserFunc(func() error {
			a.grpcServer.Shutdown()
			return nil
		}),
		app.CloserFunc(func() error {
			_ = a.httpServer.Shutdown(context.Background())
			return nil
		}),
		app.CloserFunc(func() error {
			a.mqCli.Close()
			return nil
		}),
		a.db,
		a.mdmCli,
	}
}
