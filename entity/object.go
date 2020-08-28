package entity

import (
	"msp-service-template/shared"
)

type Object struct {
	//nolint
	tableName string `pq:"?db_schema.objects" json:"-"`
	shared.ObjectDomain
}
