package model

import (
	"github.com/go-pg/pg/v9/orm"
	"github.com/integration-system/isp-lib/v2/database"
	"msp-service-template/entity"
)

type ObjectRepository struct {
	DB     orm.DB
	client *database.RxDbClient
}

func (r ObjectRepository) GetAll() ([]entity.Object, error) {
	objects := make([]entity.Object, 0)
	err := r.getDb().Model(&objects).Select()
	if err != nil {
		return nil, err
	}
	return objects, nil
}

func (r *ObjectRepository) getDb() orm.DB {
	if r.DB != nil {
		return r.DB
	}
	return r.client.Unsafe()
}
