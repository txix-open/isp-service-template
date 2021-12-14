package assembly

import (
	"github.com/integration-system/isp-kit/db"
	"msp-service-template/routes"
	"msp-service-template/usecase/object"
)

type Locator struct {
	db db.DB
}

func NewLocator(db db.DB) Locator {
	return Locator{
		db: db,
	}
}

func (l Locator) Controllers() routes.Controllers {
	repo := object.NewRepository(l.db)
	service := object.NewService(repo)
	controller := object.NewController(service)
	return routes.Controllers{
		Object: controller,
	}
}
