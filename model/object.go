package model

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/database"
	"msp-service-template/entity"
)

type ObjectRepository struct {
	db *database.RxDbClient
}

func (r ObjectRepository) GetAll() ([]entity.Object, error) {
	objects := make([]entity.Object, 0)
	err := r.db.Visit(func(db *pg.DB) error {
		return db.Model(&objects).Select()
	})
	if err != nil {
		return nil, err
	}
	return objects, nil
}
