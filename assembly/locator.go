package assembly

import (
	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/log"
	"msp-service-template/routes"
	"msp-service-template/usecase/object"
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
	repo := object.NewRepository(l.db)
	service := object.NewService(repo)
	controller := object.NewController(service)
	c := routes.Controllers{
		Object: controller,
	}
	mapper := endpoint.DefaultWrapper(l.logger)
	handler := routes.Handler(mapper, c)
	return handler
}
