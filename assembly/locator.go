package assembly

import (
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/grpc/endpoint/grpclog"
	httpEndpoint "github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/log"
	"isp-service-template/conf"
	"isp-service-template/controller"
	"isp-service-template/repository"
	"isp-service-template/routes"
	"isp-service-template/service"
	"isp-service-template/transaction"
	"net/http"
)

type DB interface {
	db.DB
	db.Transactional
}

type Locator struct {
	db     DB
	logger log.Logger
}

type LocatorConfig struct {
	HttpHandler http.Handler
	GrpcHandler *grpc.Mux
	RmqHandler  consumer.Consumer
}

func NewLocator(db DB, logger log.Logger) Locator {
	return Locator{
		db:     db,
		logger: logger,
	}
}

func (l Locator) Handlers(conf conf.Remote) LocatorConfig {
	objectRepo := repository.NewObject(l.db)
	objectService := service.NewObject(objectRepo)
	objectController := controller.NewObject(objectService)
	c := routes.Controllers{
		Object: objectController,
	}
	mapper := endpoint.DefaultWrapper(l.logger, grpclog.Log(l.logger, true))
	grpcHandler := routes.Handler(mapper, c)

	wrapper := httpEndpoint.DefaultWrapper(l.logger, httplog.Log(l.logger, true))
	httpHandler := routes.HttpHandler(wrapper, c)

	txManager := transaction.NewManager(l.db)
	msgService := service.NewMessage(l.logger, txManager)
	msgController := controller.NewMessage(msgService)

	handler := grmqx.NewResultHandler(l.logger, msgController)
	rmqHandler := conf.Consumer.Config.DefaultConsumer(handler, grmqx.ConsumerLog(l.logger, true))

	return LocatorConfig{
		HttpHandler: httpHandler,
		GrpcHandler: grpcHandler,
		RmqHandler:  rmqHandler,
	}
}
