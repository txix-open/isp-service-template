package assembly

import (
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/log"
	"msp-service-template/controller"
	"msp-service-template/repository"
	"msp-service-template/routes"
	"msp-service-template/service"
)

type Locator struct {
	db     db.DB
	logger log.Logger
}

func NewLocator(db db.DB, logger log.Logger) Locator {
	return Locator{
		db:     db,
		logger: logger,
	}
}

func (l Locator) Handler() isp.BackendServiceServer {
	objectRepo := repository.NewObject(l.db)
	objectService := service.NewObject(objectRepo)
	objectController := controller.NewObject(objectService)
	c := routes.Controllers{
		Object: objectController,
	}
	mapper := endpoint.DefaultWrapper(l.logger)
	handler := routes.Handler(mapper, c)
	return handler
}
