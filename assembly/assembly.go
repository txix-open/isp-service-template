package assembly

import (
	"context"

	"github.com/integration-system/isp-kit/app"
	"github.com/integration-system/isp-kit/bootstrap"
	"github.com/integration-system/isp-kit/cluster"
	"github.com/integration-system/isp-kit/dbrx"
	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grmqx"
	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/client"
	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"msp-service-template/conf"
)

type Assembly struct {
	boot   *bootstrap.Bootstrap
	db     *dbrx.Client
	server *grpc.Server
	mdmCli *client.Client
	logger *log.Adapter
	mqCli  *grmqx.Client
}

func New(boot *bootstrap.Bootstrap) (*Assembly, error) {
	db := dbrx.New(dbx.WithMigration(boot.MigrationsDir))
	server := grpc.NewServer()
	mdmCli, err := client.Default()
	if err != nil {
		return nil, errors.WithMessage(err, "create mdm client")
	}
	mqCli := grmqx.New(boot.App.Logger())
	return &Assembly{
		boot:   boot,
		db:     db,
		server: server,
		mdmCli: mdmCli,
		logger: boot.App.Logger(),
		mqCli:  mqCli,
	}, nil
}

func (a *Assembly) ReceiveConfig(ctx context.Context, remoteConfig []byte) error {
	var (
		newCfg  conf.Remote
		prevCfg conf.Remote
	)
	err := a.boot.RemoteConfig.Upgrade(remoteConfig, &newCfg, &prevCfg)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade remote config"))
	}

	a.logger.SetLevel(newCfg.LogLevel)

	err = a.db.Upgrade(ctx, newCfg.Database)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade db client"), log.Any("config", newCfg.Database))
	}

	locator := NewLocator(a.db, a.logger)
	handler := locator.Handler()
	a.server.Upgrade(handler)

	brokerConfig := locator.BrokerConfig(newCfg.Consumer)
	err = a.mqCli.Upgrade(ctx, brokerConfig)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade mq client"))
	}
	return nil
}

func (a *Assembly) Runners() []app.Runner {
	eventHandler := cluster.NewEventHandler().
		RequireModule("mdm", a.mdmCli).
		RemoteConfigReceiver(a)
	return []app.Runner{
		app.RunnerFunc(func(ctx context.Context) error {
			return a.server.ListenAndServe(a.boot.BindingAddress)
		}),
		app.RunnerFunc(func(ctx context.Context) error {
			return a.boot.ClusterCli.Run(ctx, eventHandler)
		}),
	}
}

func (a *Assembly) Closers() []app.Closer {
	return []app.Closer{
		a.boot.ClusterCli,
		app.CloserFunc(func() error {
			a.server.Shutdown()
			return nil
		}),
		a.db,
		a.mdmCli,
	}
}
