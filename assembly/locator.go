package assembly

import (
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
	"msp-service-template/conf"
	"msp-service-template/controller"
	"msp-service-template/repository"
	"msp-service-template/routes"
	"msp-service-template/service"
	"msp-service-template/transaction"
)

type DB interface {
	db.DB
	db.Transactional
}

type Locator struct {
	db     DB
	logger log.Logger
}

func NewLocator(db DB, logger log.Logger) Locator {
	return Locator{
		db:     db,
		logger: logger,
	}
}

func (l Locator) Handler() *grpc.Mux {
	objectRepo := repository.NewObject(l.db)
	objectService := service.NewObject(objectRepo)
	objectController := controller.NewObject(objectService)
	c := routes.Controllers{
		Object: objectController,
	}
	mapper := endpoint.DefaultWrapper(l.logger, endpoint.Log(l.logger, true))
	handler := routes.Handler(mapper, c)
	return handler
}

func (l Locator) BrokerConfig(consumerCfg conf.Consumer) grmqx.Config {
	txManager := transaction.NewManager(l.db)
	msgService := service.NewMessage(l.logger, txManager)
	msgController := controller.NewMessage(msgService)

	handler := grmqx.NewResultHandler(l.logger, msgController)
	consumer := consumerCfg.Config.DefaultConsumer(handler, grmqx.ConsumerLog(l.logger, true))

	brokerConfig := grmqx.NewConfig(
		consumerCfg.Client.Url(),
		grmqx.WithConsumers(consumer),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.Config)),
	)
	return brokerConfig
}
