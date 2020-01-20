package entity

import (
	"msp-service-template/shared"
)

type Object struct {
	//nolint
	tableName string `sql:"?db_schema.profiles"`
	shared.ObjectDomain
}
