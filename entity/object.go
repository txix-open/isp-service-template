package entity

import (
	"msp-service-template/shared"
)

type Object struct {
	tableName string `sql:"?db_schema.profiles" json:"-"`
	shared.ObjectDomain
}
