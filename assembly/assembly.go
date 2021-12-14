package assembly

import (
	"context"

	"github.com/integration-system/isp-kit/app"
	"github.com/integration-system/isp-kit/bootstrap"
	"github.com/integration-system/isp-kit/cluster"
	"github.com/integration-system/isp-kit/dbrx"
	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/client"
	"github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/rc"
	"github.com/pkg/errors"
	"msp-service-template/conf"
	"msp-service-template/routes"
)

type Assembly struct {
	rc             *rc.Config
	db             *dbrx.Client
	server         *grpc.Server
	clusterCli     *cluster.Client
	mdmCli         *client.Client
	logger         *log.Adapter
	bindingAddress string
}

func New(boot *bootstrap.Bootstrap) (*Assembly, error) {
	db := dbrx.New(dbx.WithMigration(boot.MigrationsDir))
	server := grpc.NewServer()
	mdmCli, err := client.Default()
	if err != nil {
		return nil, errors.WithMessage(err, "create mdm client")
	}
	return &Assembly{
		db:             db,
		server:         server,
		clusterCli:     boot.ClusterCli,
		rc:             boot.RemoteConfig,
		mdmCli:         mdmCli,
		logger:         boot.App.Logger(),
		bindingAddress: boot.BindingAddress,
	}, nil
}

func (a *Assembly) ReceiveConfig(ctx context.Context, remoteConfig []byte) error {
	var (
		newCfg  conf.Remote
		prevCfg conf.Remote
	)
	err := a.rc.Upgrade(remoteConfig, &newCfg, &prevCfg)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade remote config"))
	}

	err = a.db.Upgrade(ctx, newCfg.Database)
	if err != nil {
		a.logger.Fatal(ctx, errors.WithMessage(err, "upgrade db client"), log.Any("config", newCfg.Database))
	}

	mapper := endpoint.Default(a.logger)
	c := NewLocator(a.db).Controllers()
	a.server.Upgrade(routes.Handler(mapper, c))
	return nil
}

func (a *Assembly) Runners() []app.Runner {
	eventHandler := cluster.NewEventHandler().
		RequireModule("mdm", a.mdmCli).
		RemoteConfigReceiver(a)
	return []app.Runner{
		app.RunnerFunc(func(ctx context.Context) error {
			return a.server.ListenAndServe(a.bindingAddress)
		}),
		app.RunnerFunc(func(ctx context.Context) error {
			return a.clusterCli.Run(ctx, eventHandler)
		}),
	}
}

func (a *Assembly) Closers() []app.Closer {
	return []app.Closer{
		a.clusterCli,
		app.CloserFunc(func() error {
			a.server.Shutdown()
			return nil
		}),
		a.db,
		a.mdmCli,
	}
}
